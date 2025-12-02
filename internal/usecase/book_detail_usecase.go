package usecase

import (
	"context"
	"teckbook-compass-backend/internal/domain/repository"
	"teckbook-compass-backend/internal/usecase/dto"
)

// BookDetailUsecase 書籍詳細ユースケース
type BookDetailUsecase struct {
	bookRepo repository.BookRepository
}

// NewBookDetailUsecase 書籍詳細ユースケースのコンストラクタ
func NewBookDetailUsecase(bookRepo repository.BookRepository) *BookDetailUsecase {
	return &BookDetailUsecase{
		bookRepo: bookRepo,
	}
}

// GetBookDetail 書籍詳細を取得
func (uc *BookDetailUsecase) GetBookDetail(ctx context.Context, bookID string) (*dto.BookDetailResponse, error) {
	// リポジトリから書籍詳細を取得
	bookDetail, err := uc.bookRepo.GetBookByID(ctx, bookID)
	if err != nil {
		return nil, err
	}

	// 書籍が見つからない場合
	if bookDetail == nil {
		return nil, nil
	}

	// エンティティをDTOに変換
	// レビューの変換
	reviews := make([]dto.ReviewDTO, 0, len(bookDetail.FeaturedReviews))
	for _, review := range bookDetail.FeaturedReviews {
		reviews = append(reviews, dto.ReviewDTO{
			Reviewer: review.Reviewer,
			Date:     review.Date.Format("2006-01-02"),
			Rating:   review.Rating,
			Comment:  review.Comment,
		})
	}

	return &dto.BookDetailResponse{
		BookID:         bookDetail.BookID,
		Title:          bookDetail.Title,
		Author:         bookDetail.Author,
		PublishedDate:  bookDetail.PublishedDate.Format("2006-01-02"),
		Price:          bookDetail.Price,
		ISBN:           bookDetail.ISBN,
		BookImage:      bookDetail.BookImage,
		Tags:           bookDetail.Tags,
		Overview:       bookDetail.Overview,
		AboutThisBook:  bookDetail.AboutThisBook,
		TrendingPoints: bookDetail.TrendingPoints,
		AmazonReviewSummary: dto.AmazonReviewSummaryDTO{
			AverageRating: bookDetail.AmazonReviewSummary.AverageRating,
			TotalReviews:  bookDetail.AmazonReviewSummary.TotalReviews,
		},
		FeaturedReviews: reviews,
		PurchaseLinks: dto.PurchaseLinksDTO{
			Amazon:  bookDetail.PurchaseLinks.Amazon,
			Rakuten: bookDetail.PurchaseLinks.Rakuten,
		},
	}, nil
}
