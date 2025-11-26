package repository

import (
	"context"
	"teckbook-compass-backend/internal/domain/entity"
)

// BookRepository 書籍リポジトリインターフェース
type BookRepository interface {
	// GetTopBooksByCategory カテゴリ別のトップ書籍を取得
	GetTopBooksByCategory(ctx context.Context, categoryID string, limit int) ([]*entity.Book, error)

	// GetRankings 総合ランキングを取得
	GetRankings(ctx context.Context, rangeType string, limit int, offset int, categoryID string) ([]*entity.Book, error)
}
