package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"teckbook-compass-backend/internal/domain/entity"
	"teckbook-compass-backend/internal/domain/repository"
)

// BatchRepositoryImpl バッチ処理用リポジトリ実装
type BatchRepositoryImpl struct {
	db *sql.DB
}

// NewBatchRepository BatchRepositoryを生成
func NewBatchRepository(db *sql.DB) repository.BatchRepository {
	return &BatchRepositoryImpl{db: db}
}

// ArticleExists 記事が既に存在するか確認
func (r *BatchRepositoryImpl) ArticleExists(ctx context.Context, articleID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM articles WHERE id = $1)`
	err := r.db.QueryRowContext(ctx, query, articleID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check article existence: %w", err)
	}
	return exists, nil
}

// SaveArticle 記事を保存
func (r *BatchRepositoryImpl) SaveArticle(ctx context.Context, article *entity.Article) error {
	query := `
		INSERT INTO articles (id, title, url, likes, stocks, comments, published_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			likes = EXCLUDED.likes,
			stocks = EXCLUDED.stocks,
			comments = EXCLUDED.comments,
			updated_at = NOW()
	`
	_, err := r.db.ExecContext(ctx, query,
		article.ID,
		article.Title,
		article.URL,
		article.Likes,
		article.Stocks,
		article.Comments,
		article.PublishedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save article: %w", err)
	}
	return nil
}

// SaveArticleTags 記事タグを保存
func (r *BatchRepositoryImpl) SaveArticleTags(ctx context.Context, articleID string, tags []string) error {
	// 既存のタグを削除
	_, err := r.db.ExecContext(ctx, `DELETE FROM article_tags WHERE article_id = $1`, articleID)
	if err != nil {
		return fmt.Errorf("failed to delete existing tags: %w", err)
	}

	// 新しいタグを挿入
	for _, tag := range tags {
		_, err := r.db.ExecContext(ctx, `
			INSERT INTO article_tags (article_id, tag_name, created_at)
			VALUES ($1, $2, NOW())
		`, articleID, tag)
		if err != nil {
			return fmt.Errorf("failed to save tag: %w", err)
		}
	}
	return nil
}

// SaveArticleBook 記事と書籍の紐付けを保存
func (r *BatchRepositoryImpl) SaveArticleBook(ctx context.Context, articleID string, bookID string) error {
	query := `
		INSERT INTO article_books (article_id, book_id, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (article_id, book_id) DO NOTHING
	`
	_, err := r.db.ExecContext(ctx, query, articleID, bookID)
	if err != nil {
		return fmt.Errorf("failed to save article_book: %w", err)
	}
	return nil
}

// BookExists 書籍が既に存在するか確認（ISBN-10とISBN-13の両方でチェック）
func (r *BatchRepositoryImpl) BookExists(ctx context.Context, bookID string) (bool, error) {
	var exists bool
	// ISBN-13（id）またはISBN-10でチェック
	query := `SELECT EXISTS(SELECT 1 FROM books WHERE id = $1 OR isbn10 = $1)`
	err := r.db.QueryRowContext(ctx, query, bookID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check book existence: %w", err)
	}
	return exists, nil
}

// GetBookIDByISBN ISBNから書籍IDを取得（ISBN-10でもISBN-13でも検索可能）
func (r *BatchRepositoryImpl) GetBookIDByISBN(ctx context.Context, isbn string) (string, error) {
	var bookID string
	query := `SELECT id FROM books WHERE id = $1 OR isbn10 = $1 LIMIT 1`
	err := r.db.QueryRowContext(ctx, query, isbn).Scan(&bookID)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get book ID: %w", err)
	}
	return bookID, nil
}

// SaveBook 書籍を保存
func (r *BatchRepositoryImpl) SaveBook(ctx context.Context, book *entity.RakutenBook) error {
	// 出版日をパース
	var publishedDate *time.Time
	if book.SalesDate != "" {
		patterns := []string{
			"2006年01月02日",
			"2006年01月",
			"2006年1月2日",
			"2006年1月",
		}
		for _, pattern := range patterns {
			if t, err := time.Parse(pattern, book.SalesDate); err == nil {
				publishedDate = &t
				break
			}
		}
	}

	// 評価をパース
	var rating float64
	if book.ReviewAverage != "" {
		fmt.Sscanf(book.ReviewAverage, "%f", &rating)
	}

	// ISBN-13からISBN-10を計算
	isbn10 := convertISBN13to10(book.ISBN)

	query := `
		INSERT INTO books (
			id, isbn10, title, author, publisher, published_date, price,
			thumbnail_url, rakuten_url, rakuten_average_rating, rakuten_review_count,
			overview, latest_mentioned_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), NOW(), NOW())
		ON CONFLICT (id) DO UPDATE SET
			isbn10 = COALESCE(EXCLUDED.isbn10, books.isbn10),
			title = EXCLUDED.title,
			author = EXCLUDED.author,
			publisher = EXCLUDED.publisher,
			price = EXCLUDED.price,
			thumbnail_url = COALESCE(EXCLUDED.thumbnail_url, books.thumbnail_url),
			rakuten_url = COALESCE(EXCLUDED.rakuten_url, books.rakuten_url),
			rakuten_average_rating = EXCLUDED.rakuten_average_rating,
			rakuten_review_count = EXCLUDED.rakuten_review_count,
			overview = COALESCE(EXCLUDED.overview, books.overview),
			latest_mentioned_at = NOW(),
			updated_at = NOW()
	`
	_, err := r.db.ExecContext(ctx, query,
		book.ISBN,
		isbn10,
		book.Title,
		book.Author,
		book.PublisherName,
		publishedDate,
		book.ItemPrice,
		book.LargeImageURL,
		book.AffiliateURL,
		rating,
		book.ReviewCount,
		book.ItemCaption,
	)
	if err != nil {
		return fmt.Errorf("failed to save book: %w", err)
	}
	return nil
}

// UpdateBookScore 書籍のスコアを更新（latest_mentioned_atも更新）
func (r *BatchRepositoryImpl) UpdateBookScore(ctx context.Context, bookID string, score float64) error {
	query := `
		UPDATE books SET latest_mentioned_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, bookID)
	if err != nil {
		return fmt.Errorf("failed to update book score: %w", err)
	}
	return nil
}

// GetExistingBookScore 既存の書籍スコアを取得
func (r *BatchRepositoryImpl) GetExistingBookScore(ctx context.Context, bookID string) (float64, error) {
	var score float64
	query := `
		SELECT COALESCE(SUM(score), 0)
		FROM book_scores_daily
		WHERE book_id = $1
	`
	err := r.db.QueryRowContext(ctx, query, bookID).Scan(&score)
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to get existing book score: %w", err)
	}
	return score, nil
}

// SaveBookScoreDaily 書籍スコア日次集計を保存
func (r *BatchRepositoryImpl) SaveBookScoreDaily(ctx context.Context, bookID string, date time.Time, score float64, articleCount int) error {
	query := `
		INSERT INTO book_scores_daily (book_id, date, score, article_count, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (book_id, date) DO UPDATE SET
			score = book_scores_daily.score + EXCLUDED.score,
			article_count = book_scores_daily.article_count + EXCLUDED.article_count
	`
	_, err := r.db.ExecContext(ctx, query, bookID, date, score, articleCount)
	if err != nil {
		return fmt.Errorf("failed to save book score daily: %w", err)
	}
	return nil
}

// GetCategoryIDsByTags タグ名からカテゴリIDを取得
func (r *BatchRepositoryImpl) GetCategoryIDsByTags(ctx context.Context, tags []string) ([]string, error) {
	if len(tags) == 0 {
		return nil, nil
	}

	// プレースホルダーを生成
	placeholders := make([]string, len(tags))
	args := make([]interface{}, len(tags))
	for i, tag := range tags {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = tag
	}

	query := fmt.Sprintf(`
		SELECT DISTINCT category_id
		FROM tag_category_map
		WHERE tag_name IN (%s)
	`, joinStrings(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get category IDs: %w", err)
	}
	defer rows.Close()

	var categoryIDs []string
	for rows.Next() {
		var categoryID string
		if err := rows.Scan(&categoryID); err != nil {
			return nil, fmt.Errorf("failed to scan category ID: %w", err)
		}
		categoryIDs = append(categoryIDs, categoryID)
	}

	return categoryIDs, nil
}

// SaveBookCategories 書籍とカテゴリの紐付けを保存（最大3つ）
func (r *BatchRepositoryImpl) SaveBookCategories(ctx context.Context, bookID string, categoryIDs []string) error {
	// 最大3つまで
	maxCategories := 3
	if len(categoryIDs) > maxCategories {
		categoryIDs = categoryIDs[:maxCategories]
	}

	for i, categoryID := range categoryIDs {
		query := `
			INSERT INTO book_categories (book_id, category_id, score, rank, created_at)
			VALUES ($1, $2, 0, $3, NOW())
			ON CONFLICT (book_id, category_id) DO NOTHING
		`
		_, err := r.db.ExecContext(ctx, query, bookID, categoryID, i+1)
		if err != nil {
			return fmt.Errorf("failed to save book category: %w", err)
		}
	}
	return nil
}

// SaveErrorLog エラーログを保存
func (r *BatchRepositoryImpl) SaveErrorLog(ctx context.Context, log *repository.ErrorLog) error {
	requestPayloadJSON, _ := json.Marshal(log.RequestPayload)
	responseBodyJSON, _ := json.Marshal(log.ResponseBody)

	query := `
		INSERT INTO error_logs (
			batch_process, error_type, level, api_name, endpoint,
			status_code, request_payload, response_body, related_id, message, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
	`
	_, err := r.db.ExecContext(ctx, query,
		log.BatchProcess,
		log.ErrorType,
		log.Level,
		log.APIName,
		log.Endpoint,
		log.StatusCode,
		requestPayloadJSON,
		responseBodyJSON,
		log.RelatedID,
		log.Message,
	)
	if err != nil {
		return fmt.Errorf("failed to save error log: %w", err)
	}
	return nil
}

// GetBatchStatus バッチ状態を取得
func (r *BatchRepositoryImpl) GetBatchStatus(ctx context.Context, id string) (*entity.BatchStatus, error) {
	query := `
		SELECT id, last_fetched_at, next_page, last_run_at, last_new_fetch_at, created_at, updated_at
		FROM batch_statuses
		WHERE id = $1
	`
	var status entity.BatchStatus
	var lastFetchedAt, lastRunAt, lastNewFetchAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&status.ID,
		&lastFetchedAt,
		&status.NextPage,
		&lastRunAt,
		&lastNewFetchAt,
		&status.CreatedAt,
		&status.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		// 存在しない場合は新規作成
		return &entity.BatchStatus{
			ID:       id,
			NextPage: 1,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get batch status: %w", err)
	}

	if lastFetchedAt.Valid {
		status.LastFetchedAt = &lastFetchedAt.Time
	}
	if lastRunAt.Valid {
		status.LastRunAt = &lastRunAt.Time
	}
	if lastNewFetchAt.Valid {
		status.LastNewFetchAt = &lastNewFetchAt.Time
	}

	return &status, nil
}

// UpdateBatchStatusForNewFetch 最新記事取得後のバッチ状態を更新
func (r *BatchRepositoryImpl) UpdateBatchStatusForNewFetch(ctx context.Context, id string, lastFetchedAt time.Time) error {
	query := `
		INSERT INTO batch_statuses (id, last_fetched_at, next_page, last_run_at, last_new_fetch_at)
		VALUES ($1, $2, 1, NOW(), NOW())
		ON CONFLICT (id) DO UPDATE SET
			last_fetched_at = EXCLUDED.last_fetched_at,
			last_run_at = NOW(),
			last_new_fetch_at = NOW()
	`
	_, err := r.db.ExecContext(ctx, query, id, lastFetchedAt)
	if err != nil {
		return fmt.Errorf("failed to update batch status for new fetch: %w", err)
	}
	return nil
}

// UpdateBatchStatusForHistoricalFetch 過去記事取得後のバッチ状態を更新
func (r *BatchRepositoryImpl) UpdateBatchStatusForHistoricalFetch(ctx context.Context, id string, nextPage int) error {
	query := `
		UPDATE batch_statuses
		SET next_page = $2, last_run_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id, nextPage)
	if err != nil {
		return fmt.Errorf("failed to update batch status for historical fetch: %w", err)
	}
	return nil
}

// joinStrings 文字列スライスを結合
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// convertISBN13to10 ISBN-13をISBN-10に変換
// 978で始まるISBN-13のみ変換可能（979で始まるものはISBN-10に対応がない）
func convertISBN13to10(isbn13 string) *string {
	// 978で始まる13桁のみ変換可能
	if len(isbn13) != 13 || isbn13[:3] != "978" {
		return nil
	}

	// ISBN-13の4〜12桁目を取得（9桁）
	body := isbn13[3:12]

	// チェックディジットを計算
	sum := 0
	for i := 0; i < 9; i++ {
		digit := int(body[i] - '0')
		sum += digit * (10 - i)
	}

	checkDigit := (11 - (sum % 11)) % 11
	var checkChar string
	if checkDigit == 10 {
		checkChar = "X"
	} else {
		checkChar = fmt.Sprintf("%d", checkDigit)
	}

	isbn10 := body + checkChar
	return &isbn10
}
