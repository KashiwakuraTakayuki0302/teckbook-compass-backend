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

// GetCategoriesWithBooks カテゴリと関連書籍を取得
func (uc *CategoryUsecase) GetCategoriesWithBooks(ctx context.Context) (*dto.CategoryWithBooksResponse, error) {
	// カテゴリと書籍を取得
	categories, err := uc.categoryRepo.GetCategoriesWithBooks(ctx, 10)
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
				ID:        book.ID,
				Title:     book.Title,
				Thumbnail: book.Thumbnail,
			})
		}

		items = append(items, dto.CategoryItem{
			ID:       category.ID,
			Name:     category.Name,
			Icon:     category.Icon,
			TrendTag: category.TrendTag,
			Books:    books,
		})
	}

	return &dto.CategoryWithBooksResponse{
		Items: items,
	}, nil
}
