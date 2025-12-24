package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"teckbook-compass-backend/internal/domain/repository"
	"teckbook-compass-backend/internal/infrastructure/external"
)

// CategorizeBatchUsecase カテゴライズバッチ処理ユースケース
type CategorizeBatchUsecase struct {
	repo          repository.BatchRepository
	chatGPTClient *external.ChatGPTClient
	slackClient   *external.SlackClient
}

// NewCategorizeBatchUsecase CategorizeBatchUsecaseを生成
func NewCategorizeBatchUsecase(
	repo repository.BatchRepository,
	chatGPTClient *external.ChatGPTClient,
	slackClient *external.SlackClient,
) *CategorizeBatchUsecase {
	return &CategorizeBatchUsecase{
		repo:          repo,
		chatGPTClient: chatGPTClient,
		slackClient:   slackClient,
	}
}

// CategorizeBatchResult カテゴライズバッチ結果
type CategorizeBatchResult struct {
	ProcessedBooks   int
	CategorizedBooks int
	Errors           int
	ErrorMessage     string
	StartTime        time.Time
	EndTime          time.Time
	// トークン使用量
	TotalPromptTokens     int
	TotalCompletionTokens int
	TotalTokens           int
}

// カテゴリIDとコードのマッピング（DB順序に対応）
var categoryCodeMapping = []string{
	"ai-ml",        // 1
	"frontend",     // 2
	"mobile",       // 3
	"cloud",        // 4
	"infra-devops", // 5
	"backend",      // 6
	"database",     // 7
	"security",     // 8
	"beginner-cs",  // 9
	"pm-business",  // 10
}

// Run カテゴライズバッチを実行
// limit: 処理する書籍の最大数
func (u *CategorizeBatchUsecase) Run(ctx context.Context, limit int) (*CategorizeBatchResult, error) {
	result := &CategorizeBatchResult{
		StartTime: time.Now(),
	}

	log.Println("カテゴライズバッチを開始します...")

	// ChatGPT APIが有効かチェック
	if !u.chatGPTClient.IsEnabled() {
		return nil, fmt.Errorf("chatgpt api is disabled")
	}

	// 全カテゴリを取得
	dbCategories, err := u.repo.GetAllCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	// カテゴリ情報をChatGPT用に変換
	categories := make([]external.CategoryInfo, len(dbCategories))
	categoryIDMap := make(map[int]string) // 数値ID -> カテゴリコード
	for i, cat := range dbCategories {
		numericID := i + 1
		categories[i] = external.CategoryInfo{
			ID:   numericID,
			Code: cat.ID,
			Name: cat.Name,
		}
		categoryIDMap[numericID] = cat.ID
	}

	log.Printf("取得したカテゴリ数: %d", len(categories))

	// カテゴリ未設定の書籍を取得
	books, err := u.repo.GetBooksWithoutCategory(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get books without category: %w", err)
	}

	log.Printf("処理対象の書籍数: %d", len(books))

	// Slack通知: 開始メッセージ
	if u.slackClient != nil && u.slackClient.IsEnabled() {
		_ = u.slackClient.SendCategorizeBatchStartMessage(len(books))
	}

	if len(books) == 0 {
		log.Println("処理対象の書籍がありません")
		result.EndTime = time.Now()
		return result, nil
	}

	// 推定トークン数を計算（プロンプト基本部分 約600トークン + 書籍情報）
	estimateTokensPerBook := 700 // 1冊あたりの推定トークン数
	totalEstimatedTokens := len(books) * estimateTokensPerBook
	log.Printf("推定トークン使用量: 約 %d トークン（%d冊 × 約%dトークン/冊）",
		totalEstimatedTokens, len(books), estimateTokensPerBook)

	// 各書籍に対してChatGPT APIを呼び出し
	for i, book := range books {
		result.ProcessedBooks++

		log.Printf("処理中 [%d/%d]: %s (ISBN: %s)", i+1, len(books), book.Title, book.ID)

		// ChatGPTでカテゴリを判定
		catResult, err := u.chatGPTClient.CategorizeBook(ctx, book.Title, book.Overview, book.ID, categories)
		if err != nil {
			// レート制限エラーの場合は即座に終了
			if errors.Is(err, external.ErrChatGPTRateLimited) {
				remainingBooks := len(books) - i
				remainingEstimatedTokens := remainingBooks * estimateTokensPerBook
				log.Printf("ChatGPT APIレート制限に達したため終了します: %v", err)
				log.Printf("未処理: %d冊", remainingBooks)
				log.Printf("使用済みトークン: prompt=%d, completion=%d, total=%d",
					result.TotalPromptTokens, result.TotalCompletionTokens, result.TotalTokens)
				result.ErrorMessage = err.Error()
				result.Errors++
				result.EndTime = time.Now()

				if u.slackClient != nil && u.slackClient.IsEnabled() {
					errMsg := fmt.Sprintf("レート制限で終了: %v\n未処理: %d冊, 推定必要トークン: 約%d\n\n使用済みトークン: %d (prompt: %d, completion: %d)",
						err, remainingBooks, remainingEstimatedTokens,
						result.TotalTokens, result.TotalPromptTokens, result.TotalCompletionTokens)
					_ = u.slackClient.SendError("カテゴライズバッチエラー", errMsg)
				}

				u.logError(ctx, "chatgpt_rate_limit", err, book.ID)
				return result, fmt.Errorf("chatgpt api rate limited: %w", err)
			}

			// 認証エラーの場合は即座に終了
			if errors.Is(err, external.ErrChatGPTUnauthorized) {
				log.Printf("ChatGPT API認証エラーのため終了します: %v", err)
				log.Printf("使用済みトークン: prompt=%d, completion=%d, total=%d",
					result.TotalPromptTokens, result.TotalCompletionTokens, result.TotalTokens)
				result.ErrorMessage = err.Error()
				result.Errors++
				result.EndTime = time.Now()

				if u.slackClient != nil && u.slackClient.IsEnabled() {
					errMsg := fmt.Sprintf("認証エラーで終了: %v\n\n使用済みトークン: %d (prompt: %d, completion: %d)",
						err, result.TotalTokens, result.TotalPromptTokens, result.TotalCompletionTokens)
					_ = u.slackClient.SendError("カテゴライズバッチエラー", errMsg)
				}

				u.logError(ctx, "chatgpt_unauthorized", err, book.ID)
				return result, fmt.Errorf("chatgpt api unauthorized: %w", err)
			}

			// その他のエラーはスキップして続行
			log.Printf("Warning: ChatGPT API エラー (BookID: %s): %v", book.ID, err)
			result.Errors++

			u.logError(ctx, "chatgpt_api_error", err, book.ID)
			continue
		}

		// トークン使用量を集計
		result.TotalPromptTokens += catResult.PromptTokens
		result.TotalCompletionTokens += catResult.CompletionTokens
		result.TotalTokens += catResult.TotalTokens

		// カテゴリIDを変換
		categoryCode, ok := categoryIDMap[catResult.CategoryID]
		if !ok {
			log.Printf("Warning: 不明なカテゴリID (BookID: %s, CategoryID: %d)", book.ID, catResult.CategoryID)
			result.Errors++
			continue
		}

		// カテゴリを保存
		if err := u.repo.SaveBookCategory(ctx, book.ID, categoryCode); err != nil {
			log.Printf("Warning: カテゴリ保存エラー (BookID: %s): %v", book.ID, err)
			result.Errors++
			continue
		}

		result.CategorizedBooks++
		log.Printf("カテゴリ設定成功: %s -> %s (tokens: %d)", book.Title, categoryCode, catResult.TotalTokens)

		// レートリミット対策（1分に3回までの制限があるため、2冊処理したら1分待機）
		// 最後の書籍の場合は待機不要
		if i < len(books)-1 {
			if (i+1)%2 == 0 {
				log.Printf("レート制限対策: 1分間待機します...")
				time.Sleep(1 * time.Minute)
			} else {
				// 2冊目以外は短い間隔で次へ
				time.Sleep(500 * time.Millisecond)
			}
		}
	}

	result.EndTime = time.Now()

	log.Printf("カテゴライズバッチ完了: 処理=%d, 成功=%d, エラー=%d, トークン合計=%d (prompt=%d, completion=%d)",
		result.ProcessedBooks, result.CategorizedBooks, result.Errors,
		result.TotalTokens, result.TotalPromptTokens, result.TotalCompletionTokens)

	// Slack通知: 結果メッセージ
	if u.slackClient != nil && u.slackClient.IsEnabled() {
		_ = u.slackClient.SendCategorizeBatchResultMessage(
			result.ProcessedBooks,
			result.CategorizedBooks,
			result.Errors,
			result.EndTime.Sub(result.StartTime),
			result.TotalPromptTokens,
			result.TotalCompletionTokens,
			result.TotalTokens,
		)
	}

	return result, nil
}

// logError エラーをログに記録
func (u *CategorizeBatchUsecase) logError(ctx context.Context, errorType string, err error, relatedID string) {
	errLog := &repository.ErrorLog{
		BatchProcess: "categorize_batch",
		ErrorType:    errorType,
		Level:        "ERROR",
		APIName:      "ChatGPT API",
		RelatedID:    relatedID,
		Message:      err.Error(),
	}
	if saveErr := u.repo.SaveErrorLog(ctx, errLog); saveErr != nil {
		log.Printf("Failed to save error log: %v\n", saveErr)
	}
}
