package repository

import (
	"context"
	"teckbook-compass-backend/internal/domain/entity"
)

// BookRepository 書籍リポジトリインターフェース
type BookRepository interface {
	// GetTopBooksByCategory カテゴリ別のトップ書籍を取得
	GetTopBooksByCategory(ctx context.Context, categoryID string, limit int) ([]*entity.Book, error)
}
