package usecase

import (
	"context"
	"teckbook-compass-backend/internal/domain/repository"
	"teckbook-compass-backend/internal/usecase/dto"
)

// CategoryUsecase カテゴリユースケース
type CategoryUsecase struct {
	categoryRepo repository.CategoryRepository
	bookRepo     repository.BookRepository
}

// NewCategoryUsecase カテゴリユースケースのコンストラクタ
func NewCategoryUsecase(
	categoryRepo repository.CategoryRepository,
	bookRepo repository.BookRepository,
) *CategoryUsecase {
	return &CategoryUsecase{
		categoryRepo: categoryRepo,
		bookRepo:     bookRepo,
	}
}

// GetCategoriesWithBooksParams カテゴリと書籍取得のパラメータ
type GetCategoriesWithBooksParams struct {
	MaxCategories int // 返却するカテゴリ数の上限（0は全件）
	Limit         int // 各カテゴリ内の書籍数の上限（0は全件）
}

// GetCategoriesWithBooks カテゴリと関連書籍を取得
func (uc *CategoryUsecase) GetCategoriesWithBooks(ctx context.Context, params GetCategoriesWithBooksParams) (*dto.CategoryWithBooksResponse, error) {
	// カテゴリと書籍を取得
	categories, err := uc.categoryRepo.GetCategoriesWithBooks(ctx, params.MaxCategories, params.Limit)
	if err != nil {
		return nil, err
	}

	// エンティティをDTOに変換
	items := make([]dto.CategoryItem, 0, len(categories))
	for _, category := range categories {
		books := make([]dto.BookItem, 0, len(category.Books))
		for _, book := range category.Books {
			books = append(books, dto.BookItem{
				Rank:      book.Rank,
				BookID:    book.BookID,
				Title:     book.Title,
				Thumbnail: book.Thumbnail,
				Score:     book.Score,
			})
		}

		items = append(items, dto.CategoryItem{
			ID:       category.ID,
			Name:     category.Name,
			Icon:     category.Icon,
			TrendTag: category.TrendTag,
			Score:    category.Score,
			Books:    books,
		})
	}

	return &dto.CategoryWithBooksResponse{
		Items: items,
	}, nil
}
