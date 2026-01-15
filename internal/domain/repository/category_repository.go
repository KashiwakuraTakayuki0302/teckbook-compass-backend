package repository

import (
	"context"
	"teckbook-compass-backend/internal/domain/entity"
)

// CategoryRepository カテゴリリポジトリインターフェース
type CategoryRepository interface {
	// GetCategoriesWithBooks カテゴリと関連する書籍を取得
	// maxCategories: 返却するカテゴリ数の上限（0は全件）
	// bookLimit: 各カテゴリ内の書籍数の上限（0は全件）
	GetCategoriesWithBooks(ctx context.Context, maxCategories int, bookLimit int) ([]*entity.Category, error)
}
