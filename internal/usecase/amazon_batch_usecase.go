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

// AmazonBatchUsecase Amazon URL取得バッチ処理ユースケース
type AmazonBatchUsecase struct {
	repo         repository.BatchRepository
	amazonClient *external.AmazonClient
	slackClient  *external.SlackClient
}

// NewAmazonBatchUsecase AmazonBatchUsecaseを生成
func NewAmazonBatchUsecase(
	repo repository.BatchRepository,
	amazonClient *external.AmazonClient,
	slackClient *external.SlackClient,
) *AmazonBatchUsecase {
	return &AmazonBatchUsecase{
		repo:         repo,
		amazonClient: amazonClient,
		slackClient:  slackClient,
	}
}

// AmazonBatchResult Amazon URL取得バッチ結果
type AmazonBatchResult struct {
	ProcessedBooks int
	UpdatedBooks   int
	NotFoundBooks  int
	Errors         int
	ErrorMessage   string // API エラー時のメッセージ
	StartTime      time.Time
	EndTime        time.Time
}

// Run Amazon URL取得バッチを実行
// limit: 処理する書籍の最大数
func (u *AmazonBatchUsecase) Run(ctx context.Context, limit int) (*AmazonBatchResult, error) {
	result := &AmazonBatchResult{
		StartTime: time.Now(),
	}

	log.Println("Amazon URL取得バッチを開始します...")

	// Amazon APIが有効かチェック
	if !u.amazonClient.IsEnabled() {
		return nil, fmt.Errorf("amazon api is disabled")
	}

	// スコアが高い順にamazon_urlがない書籍を取得
	books, err := u.repo.GetBooksWithoutAmazonURLByScore(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get books without amazon url: %w", err)
	}

	log.Printf("処理対象の書籍数: %d", len(books))

	// Slack通知: 開始メッセージ
	if u.slackClient != nil && u.slackClient.IsEnabled() {
		_ = u.slackClient.SendAmazonBatchStartMessage(len(books))
	}

	if len(books) == 0 {
		log.Println("処理対象の書籍がありません")
		result.EndTime = time.Now()
		return result, nil
	}

	// 各書籍に対してAmazon APIを呼び出し
	for i, book := range books {
		result.ProcessedBooks++

		if i > 0 && i%10 == 0 {
			log.Printf("進捗: %d/%d 書籍を処理済み", i, len(books))
		}

		// ISBN-10がある場合はそれを使用、なければISBN-13を使用
		searchISBN := book.ID
		if book.ISBN10 != nil && *book.ISBN10 != "" {
			searchISBN = *book.ISBN10
		}

		// タイトルで検索
		amazonBook, err := u.amazonClient.SearchByTitle(ctx, book.Title)
		if err != nil {
			// API エラーの場合は即座に終了
			if errors.Is(err, external.ErrAmazonAPIError) {
				log.Printf("Amazon API エラーが発生したため終了します: %v", err)
				result.ErrorMessage = err.Error()
				result.Errors++
				result.EndTime = time.Now()

				// Slack通知: エラーメッセージ
				if u.slackClient != nil && u.slackClient.IsEnabled() {
					_ = u.slackClient.SendError("Amazon URL取得バッチエラー", fmt.Sprintf("APIエラーで終了: %v", err))
				}

				// エラーログを保存
				u.logError(ctx, "amazon_api_error", err, searchISBN)

				return result, fmt.Errorf("amazon api error: %w", err)
			}

			// 商品が見つからない場合はスキップして続行
			if errors.Is(err, external.ErrAmazonNotFound) {
				log.Printf("書籍が見つかりません: %s (%s)", book.Title, searchISBN)
				result.NotFoundBooks++
				continue
			}

			// その他のエラーはスキップして続行
			log.Printf("Warning: Amazon API エラー (ISBN: %s): %v", searchISBN, err)
			result.Errors++
			continue
		}

		// Amazon URLを更新
		if err := u.repo.UpdateBookAmazonURL(ctx, book.ID, amazonBook.URL); err != nil {
			log.Printf("Warning: Amazon URL更新エラー (BookID: %s): %v", book.ID, err)
			result.Errors++
			continue
		}

		result.UpdatedBooks++
		log.Printf("Amazon URL更新成功: %s -> %s", book.Title, amazonBook.URL)

		// レートリミット対策（Amazon PA-APIは1秒あたり1リクエストの制限がある）
		time.Sleep(1100 * time.Millisecond)
	}

	result.EndTime = time.Now()

	log.Printf("Amazon URL取得バッチ完了: 処理=%d, 更新=%d, 未発見=%d, エラー=%d",
		result.ProcessedBooks, result.UpdatedBooks, result.NotFoundBooks, result.Errors)

	// Slack通知: 結果メッセージ
	if u.slackClient != nil && u.slackClient.IsEnabled() {
		_ = u.slackClient.SendAmazonBatchResultMessage(
			result.ProcessedBooks,
			result.UpdatedBooks,
			result.NotFoundBooks,
			result.Errors,
			result.EndTime.Sub(result.StartTime),
		)
	}

	return result, nil
}

// logError エラーをログに記録
func (u *AmazonBatchUsecase) logError(ctx context.Context, errorType string, err error, relatedID string) {
	errLog := &repository.ErrorLog{
		BatchProcess: "amazon_url_batch",
		ErrorType:    errorType,
		Level:        "ERROR",
		APIName:      "Amazon PA-API",
		RelatedID:    relatedID,
		Message:      err.Error(),
	}
	if saveErr := u.repo.SaveErrorLog(ctx, errLog); saveErr != nil {
		log.Printf("Failed to save error log: %v\n", saveErr)
	}
}
