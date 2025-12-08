package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"teckbook-compass-backend/internal/domain/entity"
	"teckbook-compass-backend/internal/domain/repository"
)

// BookRepositoryImpl 書籍リポジトリ実装
type BookRepositoryImpl struct {
	db *sql.DB
}

// NewBookRepository 書籍リポジトリを生成
func NewBookRepository(db *sql.DB) repository.BookRepository {
	return &BookRepositoryImpl{db: db}
}

// GetTopBooksByCategory カテゴリ別のトップ書籍を取得
func (r *BookRepositoryImpl) GetTopBooksByCategory(ctx context.Context, categoryID string, limit int) ([]*entity.Book, error) {
	// 現在はGetCategoriesWithBooksで書籍も含めて取得しているため、
	// このメソッドは将来の拡張用として空実装
	return []*entity.Book{}, nil
}

// GetRankings 総合ランキングを取得
// book_scores_dailyテーブルからスコアが高い順に書籍を取得
func (r *BookRepositoryImpl) GetRankings(ctx context.Context, rangeType string, limit int, offset int, categoryID string) ([]*entity.Book, error) {
	// 日付範囲を決定
	var dateCondition string
	args := []interface{}{}
	argIndex := 1

	switch rangeType {
	case "monthly":
		dateCondition = fmt.Sprintf("AND bsd.date >= $%d", argIndex)
		args = append(args, time.Now().AddDate(0, -1, 0))
		argIndex++
	case "yearly":
		dateCondition = fmt.Sprintf("AND bsd.date >= $%d", argIndex)
		args = append(args, time.Now().AddDate(-1, 0, 0))
		argIndex++
	default: // "all"
		dateCondition = ""
	}

	// カテゴリフィルタ
	var categoryJoin string
	if categoryID != "" {
		categoryJoin = fmt.Sprintf("INNER JOIN book_categories bc ON b.id = bc.book_id AND bc.category_id = $%d", argIndex)
		args = append(args, categoryID)
		argIndex++
	}

	// 集計クエリを構築
	query := fmt.Sprintf(`
		SELECT
			b.id,
			b.title,
			COALESCE(b.author, '') as author,
			COALESCE(b.rakuten_average_rating, 0) as rating,
			COALESCE(b.rakuten_review_count, 0) as review_count,
			b.published_date,
			COALESCE(b.thumbnail_url, '') as thumbnail,
			COALESCE(b.amazon_url, '') as amazon_url,
			COALESCE(b.rakuten_url, '') as rakuten_url,
			COALESCE(SUM(bsd.score), 0) as total_score,
			COALESCE(SUM(bsd.article_count), 0) as total_article_count
		FROM books b
		INNER JOIN book_scores_daily bsd ON b.id = bsd.book_id
		%s
		WHERE 1=1 %s
		GROUP BY b.id, b.title, b.author, b.rakuten_average_rating, b.rakuten_review_count, b.published_date, b.thumbnail_url, b.amazon_url, b.rakuten_url
		ORDER BY total_score DESC, total_article_count DESC, b.id
	`, categoryJoin, dateCondition)

	// limitとoffsetを追加（limit=0は全件取得）
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
		argIndex++
	}
	query += fmt.Sprintf(" OFFSET $%d", argIndex)
	args = append(args, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get rankings: %w", err)
	}
	defer rows.Close()

	var books []*entity.Book
	rank := offset + 1
	for rows.Next() {
		var book entity.Book
		var totalScore float64
		var articleCount int
		var publishedAt sql.NullTime
		err := rows.Scan(
			&book.BookID,
			&book.Title,
			&book.Author,
			&book.Rating,
			&book.ReviewCount,
			&publishedAt,
			&book.Thumbnail,
			&book.AmazonURL,
			&book.RakutenURL,
			&totalScore,
			&articleCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan book: %w", err)
		}
		book.Rank = rank
		book.ArticleCount = articleCount
		// PublishedAtがNULLでない場合のみ設定
		if publishedAt.Valid {
			book.PublishedAt = &publishedAt.Time
		}
		// タグは別途取得が必要な場合は追加実装
		book.Tags = []string{}
		books = append(books, &book)
		rank++
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// 各書籍のタグを取得
	for _, book := range books {
		tags, err := r.getBookTags(ctx, book.BookID)
		if err != nil {
			// タグ取得失敗はログに出すが、処理は継続
			continue
		}
		book.Tags = tags
	}

	return books, nil
}

// getBookTags 書籍に紐づくタグを取得（article_tagsから集計）
func (r *BookRepositoryImpl) getBookTags(ctx context.Context, bookID string) ([]string, error) {
	query := `
		SELECT DISTINCT at.tag_name
		FROM article_tags at
		INNER JOIN article_books ab ON at.article_id = ab.article_id
		WHERE ab.book_id = $1
		LIMIT 5
	`
	rows, err := r.db.QueryContext(ctx, query, bookID)
	if err != nil {
		return nil, fmt.Errorf("failed to get book tags: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// GetBookByID 書籍IDで書籍詳細を取得
func (r *BookRepositoryImpl) GetBookByID(ctx context.Context, bookID string) (*entity.BookDetail, error) {
	// 書籍基本情報を取得
	query := `
		SELECT
			b.id,
			b.title,
			COALESCE(b.author, '') as author,
			b.published_date,
			COALESCE(b.price, 0) as price,
			b.id as isbn,
			COALESCE(b.thumbnail_url, '') as book_image,
			COALESCE(b.overview, '') as overview,
			COALESCE(b.rakuten_average_rating, 0) as average_rating,
			COALESCE(b.rakuten_review_count, 0) as total_reviews,
			COALESCE(b.amazon_url, '') as amazon_url,
			COALESCE(b.rakuten_url, '') as rakuten_url
		FROM books b
		WHERE b.id = $1
	`
	var bookDetail entity.BookDetail
	var publishedDate sql.NullTime
	err := r.db.QueryRowContext(ctx, query, bookID).Scan(
		&bookDetail.BookID,
		&bookDetail.Title,
		&bookDetail.Author,
		&publishedDate,
		&bookDetail.Price,
		&bookDetail.ISBN,
		&bookDetail.BookImage,
		&bookDetail.Overview,
		&bookDetail.RakutenReviewSummary.AverageRating,
		&bookDetail.RakutenReviewSummary.TotalReviews,
		&bookDetail.PurchaseLinks.Amazon,
		&bookDetail.PurchaseLinks.Rakuten,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get book by ID: %w", err)
	}

	// PublishedDateがNULLでない場合のみ設定
	if publishedDate.Valid {
		bookDetail.PublishedDate = &publishedDate.Time
	}

	// ISBNフォーマット調整（978-xxxxxxxxxx形式）
	if len(bookDetail.ISBN) == 13 {
		bookDetail.ISBN = bookDetail.ISBN[:3] + "-" + bookDetail.ISBN[3:]
	}

	// タグを取得
	tags, err := r.getBookTags(ctx, bookID)
	if err == nil {
		bookDetail.Tags = tags
	} else {
		bookDetail.Tags = []string{}
	}

	// Qiita記事を取得
	qiitaArticles, err := r.getQiitaArticles(ctx, bookID)
	if err == nil {
		bookDetail.QiitaArticles = qiitaArticles
	} else {
		bookDetail.QiitaArticles = []entity.QiitaArticle{}
	}

	return &bookDetail, nil
}

// getQiitaArticles 書籍に紐づくQiita記事を取得
func (r *BookRepositoryImpl) getQiitaArticles(ctx context.Context, bookID string) ([]entity.QiitaArticle, error) {
	query := `
		SELECT
			a.title,
			a.url,
			a.likes,
			a.stocks,
			a.comments
		FROM articles a
		INNER JOIN article_books ab ON a.id = ab.article_id
		WHERE ab.book_id = $1
		ORDER BY a.likes DESC
		LIMIT 10
	`
	rows, err := r.db.QueryContext(ctx, query, bookID)
	if err != nil {
		return nil, fmt.Errorf("failed to get qiita articles: %w", err)
	}
	defer rows.Close()

	var articles []entity.QiitaArticle
	for rows.Next() {
		var article entity.QiitaArticle
		if err := rows.Scan(&article.Title, &article.URL, &article.Likes, &article.Stocks, &article.Comments); err != nil {
			return nil, fmt.Errorf("failed to scan article: %w", err)
		}
		articles = append(articles, article)
	}

	return articles, nil
}
