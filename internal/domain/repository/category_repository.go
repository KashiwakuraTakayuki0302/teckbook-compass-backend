package repository

import (
	"context"
	"teckbook-compass-backend/internal/domain/entity"
)

// CategoryRepository カテゴリリポジトリインターフェース
type CategoryRepository interface {
	// GetCategoriesWithBooks カテゴリと関連する書籍を取得
	GetCategoriesWithBooks(ctx context.Context, limit int) ([]*entity.Category, error)
}
