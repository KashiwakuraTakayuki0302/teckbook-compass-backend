package mock

import (
	"context"
	"teckbook-compass-backend/internal/domain/entity"
)

// BookRepositoryMock 書籍リポジトリのモック実装
type BookRepositoryMock struct{}

// NewBookRepositoryMock 書籍リポジトリモックのコンストラクタ
func NewBookRepositoryMock() *BookRepositoryMock {
	return &BookRepositoryMock{}
}

// GetTopBooksByCategory カテゴリ別のトップ書籍を取得（モックデータ）
func (r *BookRepositoryMock) GetTopBooksByCategory(ctx context.Context, categoryID string, limit int) ([]*entity.Book, error) {
	// 現在はGetCategoriesWithBooksで書籍も含めて取得しているため、
	// このメソッドは将来の拡張用として空実装
	return []*entity.Book{}, nil
}
