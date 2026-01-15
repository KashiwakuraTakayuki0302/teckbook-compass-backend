package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"teckbook-compass-backend/internal/domain/entity"
	"teckbook-compass-backend/internal/domain/repository"
)

// CategoryRepositoryImpl カテゴリリポジトリ実装
type CategoryRepositoryImpl struct {
	db *sql.DB
}

// NewCategoryRepository カテゴリリポジトリを生成
func NewCategoryRepository(db *sql.DB) repository.CategoryRepository {
	return &CategoryRepositoryImpl{db: db}
}

// GetCategoriesWithBooks カテゴリと関連する書籍を取得
// maxCategories: 返却するカテゴリ数の上限（0は全件）
// bookLimit: 各カテゴリ内の書籍数の上限（0は全件）
func (r *CategoryRepositoryImpl) GetCategoriesWithBooks(ctx context.Context, maxCategories int, bookLimit int) ([]*entity.Category, error) {
	// カテゴリをスコアの降順で取得
	categoryQuery := `
		SELECT id, name, COALESCE(icon, '') as icon, COALESCE(score, 0) as score
		FROM categories
		ORDER BY score DESC, id
	`
	if maxCategories > 0 {
		categoryQuery += fmt.Sprintf(" LIMIT %d", maxCategories)
	}

	rows, err := r.db.QueryContext(ctx, categoryQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	defer rows.Close()

	var categories []*entity.Category
	for rows.Next() {
		var cat entity.Category
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.Icon, &cat.Score); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		// TrendTagを決定（スコアに基づく）
		cat.TrendTag = determineTrendTag(cat.Score)
		categories = append(categories, &cat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating categories: %w", err)
	}

	// 各カテゴリに書籍を取得
	for _, cat := range categories {
		books, err := r.getBooksByCategory(ctx, cat.ID, bookLimit)
		if err != nil {
			return nil, fmt.Errorf("failed to get books for category %s: %w", cat.ID, err)
		}
		cat.Books = books
	}

	return categories, nil
}

// getBooksByCategory カテゴリに紐づく書籍をスコア順に取得
func (r *CategoryRepositoryImpl) getBooksByCategory(ctx context.Context, categoryID string, limit int) ([]*entity.Book, error) {
	// book_categoriesとbook_scores_dailyをJOINしてスコア順に取得
	query := `
		SELECT
			b.id,
			b.title,
			COALESCE(b.thumbnail_url, '') as thumbnail,
			COALESCE(SUM(bsd.score), 0) as total_score
		FROM books b
		INNER JOIN book_categories bc ON b.id = bc.book_id
		LEFT JOIN book_scores_daily bsd ON b.id = bsd.book_id
		WHERE bc.category_id = $1
		GROUP BY b.id, b.title, b.thumbnail_url
		ORDER BY total_score DESC, b.id
	`
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.db.QueryContext(ctx, query, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get books: %w", err)
	}
	defer rows.Close()

	var books []*entity.Book
	rank := 1
	for rows.Next() {
		var book entity.Book
		if err := rows.Scan(&book.BookID, &book.Title, &book.Thumbnail, &book.Score); err != nil {
			return nil, fmt.Errorf("failed to scan book: %w", err)
		}
		book.Rank = rank
		book.CategoryID = categoryID
		books = append(books, &book)
		rank++
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating books: %w", err)
	}

	return books, nil
}

// determineTrendTag スコアに基づいてトレンドタグを決定
func determineTrendTag(score float64) string {
	switch {
	case score >= 1000:
		return "hot"
	case score >= 500:
		return "popular"
	default:
		return "attention"
	}
}
