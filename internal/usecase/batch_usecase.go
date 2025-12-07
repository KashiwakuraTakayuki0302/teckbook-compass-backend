package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"teckbook-compass-backend/internal/domain/entity"
	"teckbook-compass-backend/internal/domain/repository"
	"teckbook-compass-backend/internal/infrastructure/external"
	"teckbook-compass-backend/internal/infrastructure/extractor"
)

// BatchUsecase バッチ処理ユースケース
type BatchUsecase struct {
	repo          repository.BatchRepository
	qiitaClient   *external.QiitaClient
	rakutenClient *external.RakutenClient
	slackClient   *external.SlackClient
	bookExtractor *extractor.BookExtractor
}

// NewBatchUsecase BatchUsecaseを生成
func NewBatchUsecase(
	repo repository.BatchRepository,
	qiitaClient *external.QiitaClient,
	rakutenClient *external.RakutenClient,
	slackClient *external.SlackClient,
) *BatchUsecase {
	return &BatchUsecase{
		repo:          repo,
		qiitaClient:   qiitaClient,
		rakutenClient: rakutenClient,
		slackClient:   slackClient,
		bookExtractor: extractor.NewBookExtractor(),
	}
}

// BookScoreMap バッチ処理中のスコアを管理
type BookScoreMap map[string]*entity.BookScore

// BatchResult バッチ処理結果
type BatchResult struct {
	FetchMode         string
	ProcessedArticles int
	NewArticles       int
	ProcessedBooks    int
	NewBooks          int
	Errors            int
	NextPage          int
	FetchStats        *external.FetchStats
	StartTime         time.Time
	EndTime           time.Time
}

// FetchModeOption 取得モードオプション（コマンドラインから指定）
type FetchModeOption int

const (
	FetchModeOptionNew        FetchModeOption = iota // 最新記事取得モード
	FetchModeOptionHistorical                        // 過去記事取得モード
)

// Run バッチ処理を実行
// fetchModeOption: nilの場合は自動判定、指定された場合は強制的にそのモードで実行
func (u *BatchUsecase) Run(ctx context.Context, fetchModeOption *FetchModeOption) (*BatchResult, error) {
	result := &BatchResult{
		StartTime: time.Now(),
	}

	log.Println("バッチ処理を開始します...")

	// バッチ状態を取得
	batchStatus, err := u.repo.GetBatchStatus(ctx, entity.BatchStatusIDQiitaFetch)
	if err != nil {
		return nil, fmt.Errorf("failed to get batch status: %w", err)
	}

	// 取得モードを判定
	var fetchMode entity.FetchMode
	if fetchModeOption != nil {
		// コマンドラインから指定された場合は強制
		if *fetchModeOption == FetchModeOptionNew {
			fetchMode = entity.FetchModeNew
			log.Println("取得モード: 最新記事取得（強制指定）")
		} else {
			fetchMode = entity.FetchModeHistorical
			log.Println("取得モード: 過去記事取得（強制指定）")
		}
	} else {
		// 自動判定
		fetchMode = batchStatus.GetFetchMode()
		log.Printf("取得モード: %s（自動判定）\n", fetchMode.String())
	}
	result.FetchMode = fetchMode.String()

	// Slack通知: 開始メッセージ
	if u.slackClient != nil {
		if err := u.slackClient.SendStartMessage(result.FetchMode); err != nil {
			log.Printf("Warning: Slack通知エラー: %v\n", err)
		}
	}

	var articles []*entity.QiitaAPIArticle
	var fetchStats *external.FetchStats

	// 1. Qiita APIから記事を取得
	log.Println("Step 1: Qiita APIから記事を取得中...")
	u.slackLog("Step 1: Qiita APIから記事を取得中...")

	if fetchMode == entity.FetchModeNew {
		// 最新記事取得モード
		articles, fetchStats, err = u.qiitaClient.FetchNewArticles(ctx, external.SearchQueries, batchStatus.LastFetchedAt, 8)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch new articles: %w", err)
		}
	} else {
		// 過去記事取得モード
		var nextPage int
		articles, nextPage, fetchStats, err = u.qiitaClient.FetchHistoricalArticles(ctx, external.SearchQueries, batchStatus.NextPage, 1)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch historical articles: %w", err)
		}
		result.NextPage = nextPage
	}

	log.Printf("取得した記事数: %d\n", len(articles))
	u.slackLogf("取得した記事数: %d", len(articles))
	result.ProcessedArticles = len(articles)
	result.FetchStats = fetchStats

	// 書籍スコアを管理するマップ
	bookScores := make(BookScoreMap)

	// 2-4. 各記事を処理
	log.Println("Step 2-4: 各記事から技術書を抽出中...")
	u.slackLog("Step 2-4: 各記事から技術書を抽出中...")

	for i, article := range articles {
		if i > 0 && i%50 == 0 {
			log.Printf("進捗: %d/%d 記事を処理済み\n", i, len(articles))
			u.slackLogf("進捗: %d/%d 記事を処理済み", i, len(articles))
		}

		isNew, err := u.processArticle(ctx, article, bookScores)
		if err != nil {
			log.Printf("Warning: 記事処理エラー (ID: %s): %v\n", article.ID, err)
			u.logError(ctx, "article_processing", err, article.ID)
			result.Errors++
			continue
		}

		if isNew {
			result.NewArticles++
		}

		// レートリミット対策
		time.Sleep(500 * time.Millisecond)
	}

	// 5. スコアを保存
	log.Println("Step 5: 書籍スコアを保存中...")
	u.slackLog("Step 5: 書籍スコアを保存中...")

	for bookID, score := range bookScores {
		// 日付は紐づく記事の最新投稿日（日付のみ）
		scoreDate := score.LatestArticleDate.Truncate(24 * time.Hour)
		if err := u.repo.SaveBookScoreDaily(ctx, bookID, scoreDate, score.Score, score.ArticleCount); err != nil {
			log.Printf("Warning: スコア保存エラー (BookID: %s): %v\n", bookID, err)
			result.Errors++
			continue
		}
		result.ProcessedBooks++
	}

	// 6. Amazon API処理（後で追加するためスキップ）
	log.Println("Step 6: Amazon API処理はスキップ（後で追加）")

	// 7. カテゴリ振り分けは記事処理時に実行済み
	log.Println("Step 7: カテゴリ振り分けは記事処理時に完了済み")

	// 8. バッチ状態を更新
	log.Println("Step 8: バッチ状態を更新中...")
	u.slackLog("Step 8: バッチ状態を更新中...")

	if fetchMode == entity.FetchModeNew {
		// 最新記事取得モードの場合、last_fetched_atを更新
		if err := u.repo.UpdateBatchStatusForNewFetch(ctx, entity.BatchStatusIDQiitaFetch, time.Now()); err != nil {
			log.Printf("Warning: バッチ状態更新エラー: %v\n", err)
		}
		log.Println("最新記事取得完了 - 次回まで過去記事取得モードに移行")
	} else {
		// 過去記事取得モードの場合、next_pageを更新
		if err := u.repo.UpdateBatchStatusForHistoricalFetch(ctx, entity.BatchStatusIDQiitaFetch, result.NextPage); err != nil {
			log.Printf("Warning: バッチ状態更新エラー: %v\n", err)
		}
		log.Printf("過去記事取得完了 - 次回開始ページ: %d\n", result.NextPage)
	}

	result.EndTime = time.Now()
	log.Printf("バッチ処理完了: 処理時間 %v\n", result.EndTime.Sub(result.StartTime))

	// Slack通知: 結果メッセージ
	if u.slackClient != nil {
		if err := u.slackClient.SendResultMessage(
			result.FetchMode,
			result.ProcessedArticles,
			result.NewArticles,
			result.ProcessedBooks,
			result.Errors,
			result.NextPage,
			result.EndTime.Sub(result.StartTime),
			result.FetchStats,
		); err != nil {
			log.Printf("Warning: Slack結果通知エラー: %v\n", err)
		}
	}

	return result, nil
}

// slackLog Slackにログを送信
func (u *BatchUsecase) slackLog(message string) {
	if u.slackClient != nil {
		_ = u.slackClient.SendLog(message)
	}
}

// slackLogf フォーマット付きでSlackにログを送信
func (u *BatchUsecase) slackLogf(format string, args ...interface{}) {
	if u.slackClient != nil {
		_ = u.slackClient.SendLogf(format, args...)
	}
}

// processArticle 記事を処理
func (u *BatchUsecase) processArticle(ctx context.Context, qiitaArticle *entity.QiitaAPIArticle, bookScores BookScoreMap) (bool, error) {
	// 既に処理済みかチェック
	exists, err := u.repo.ArticleExists(ctx, qiitaArticle.ID)
	if err != nil {
		return false, fmt.Errorf("failed to check article existence: %w", err)
	}

	// 既存の記事でも更新する（スコア計算のため）
	article := qiitaArticle.ToArticle()

	// 記事を保存
	if err := u.repo.SaveArticle(ctx, article); err != nil {
		return false, fmt.Errorf("failed to save article: %w", err)
	}

	// タグを保存
	if err := u.repo.SaveArticleTags(ctx, qiitaArticle.ID, qiitaArticle.GetTagNames()); err != nil {
		return false, fmt.Errorf("failed to save article tags: %w", err)
	}

	// 記事本文から書籍を抽出
	extractedBooks := u.bookExtractor.ExtractFromText(qiitaArticle.Body)
	if len(extractedBooks) == 0 {
		// HTMLからも試す
		extractedBooks = u.bookExtractor.ExtractFromHTML(qiitaArticle.RenderedBody)
	}

	// 抽出した書籍を処理
	for _, extracted := range extractedBooks {
		bookID, err := u.processExtractedBook(ctx, extracted)
		if err != nil {
			// 書籍取得に失敗した場合はスキップ
			continue
		}

		if bookID != "" {
			// 記事と書籍を紐付け
			if err := u.repo.SaveArticleBook(ctx, article.ID, bookID); err != nil {
				log.Printf("Warning: 記事-書籍紐付けエラー: %v\n", err)
			}

			// スコアを加算
			if score, ok := bookScores[bookID]; ok {
				score.AddScore(article.Likes, article.Stocks, article.PublishedAt)
			} else {
				// 既存のスコアを取得
				existingScore, _ := u.repo.GetExistingBookScore(ctx, bookID)
				bookScores[bookID] = &entity.BookScore{
					BookID:       bookID,
					Score:        existingScore,
					ArticleCount: 0,
				}
				bookScores[bookID].AddScore(article.Likes, article.Stocks, article.PublishedAt)
			}

			// カテゴリを振り分け
			u.assignBookCategories(ctx, bookID, article.Tags)
		}
	}

	return !exists, nil
}

// processExtractedBook 抽出した書籍情報を処理
func (u *BatchUsecase) processExtractedBook(ctx context.Context, extracted extractor.ExtractedBook) (string, error) {
	var rakutenBook *entity.RakutenBook
	var err error

	// ISBNがある場合
	if extracted.ISBN != "" {
		// まずDBで存在チェック（ISBN-10/13両方で検索）
		existingBookID, err := u.repo.GetBookIDByISBN(ctx, extracted.ISBN)
		if err != nil {
			return "", fmt.Errorf("failed to check book existence: %w", err)
		}
		if existingBookID != "" {
			// 既存の書籍が見つかった場合はそのIDを返す（楽天API呼び出し不要）
			return existingBookID, nil
		}

		// 楽天APIで書籍情報を取得
		rakutenBook, err = u.rakutenClient.SearchByISBN(ctx, extracted.ISBN)
		if err != nil {
			// ISBNで見つからない場合はスキップ
			return "", fmt.Errorf("failed to fetch book by ISBN: %w", err)
		}
	} else if extracted.Title != "" {
		// タイトルで検索
		books, err := u.rakutenClient.SearchByTitle(ctx, extracted.Title)
		if err != nil || len(books) == 0 {
			return "", fmt.Errorf("failed to fetch book by title: %w", err)
		}
		rakutenBook = books[0]
	} else if extracted.ASIN != "" {
		// ASINの場合は楽天APIでは直接検索できないのでスキップ
		// 将来的にはAmazon APIで対応
		return "", fmt.Errorf("ASIN lookup not supported yet")
	}

	if rakutenBook == nil || rakutenBook.ISBN == "" {
		return "", fmt.Errorf("no valid book data")
	}

	// 楽天APIで取得したISBN（正規化済み）で再度存在チェック
	existingBookID, err := u.repo.GetBookIDByISBN(ctx, rakutenBook.ISBN)
	if err != nil {
		return "", fmt.Errorf("failed to check book existence: %w", err)
	}

	if existingBookID != "" {
		// 既存の書籍の場合はそのIDを返す
		return existingBookID, nil
	}

	// 書籍を保存
	if err = u.repo.SaveBook(ctx, rakutenBook); err != nil {
		return "", fmt.Errorf("failed to save book: %w", err)
	}

	// レートリミット対策
	time.Sleep(300 * time.Millisecond)

	// 楽天APIから取得したISBNを返す（保存したIDと一致させる）
	return rakutenBook.ISBN, nil
}

// assignBookCategories 書籍にカテゴリを割り当て
func (u *BatchUsecase) assignBookCategories(ctx context.Context, bookID string, tags []string) {
	categoryIDs, err := u.repo.GetCategoryIDsByTags(ctx, tags)
	if err != nil {
		log.Printf("Warning: カテゴリ取得エラー: %v\n", err)
		return
	}

	if len(categoryIDs) > 0 {
		if err := u.repo.SaveBookCategories(ctx, bookID, categoryIDs); err != nil {
			log.Printf("Warning: カテゴリ保存エラー: %v\n", err)
		}
	}
}

// logError エラーをログに記録
func (u *BatchUsecase) logError(ctx context.Context, errorType string, err error, relatedID string) {
	errLog := &repository.ErrorLog{
		BatchProcess: "daily_batch",
		ErrorType:    errorType,
		Level:        "ERROR",
		RelatedID:    relatedID,
		Message:      err.Error(),
	}
	if saveErr := u.repo.SaveErrorLog(ctx, errLog); saveErr != nil {
		log.Printf("Failed to save error log: %v\n", saveErr)
	}
}

// GetStats バッチ処理の統計情報を取得（デバッグ用）
func (u *BatchUsecase) GetStats(ctx context.Context) (map[string]interface{}, error) {
	// 統計情報を返す（将来の拡張用）
	return map[string]interface{}{
		"search_queries": external.SearchQueries,
	}, nil
}
