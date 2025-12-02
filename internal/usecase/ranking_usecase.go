package usecase

import (
	"context"
	"teckbook-compass-backend/internal/domain/repository"
	"teckbook-compass-backend/internal/usecase/dto"
)

// RankingUsecase ランキングユースケース
type RankingUsecase struct {
	bookRepo repository.BookRepository
}

// NewRankingUsecase ランキングユースケースのコンストラクタ
func NewRankingUsecase(bookRepo repository.BookRepository) *RankingUsecase {
	return &RankingUsecase{
		bookRepo: bookRepo,
	}
}

// GetRankings 総合ランキングを取得
func (uc *RankingUsecase) GetRankings(ctx context.Context, rangeType string, limit int, offset int, categoryID string) (*dto.RankingResponse, error) {
	// リポジトリから書籍ランキングを取得
	books, err := uc.bookRepo.GetRankings(ctx, rangeType, limit, offset, categoryID)
	if err != nil {
		return nil, err
	}

	// エンティティをDTOに変換
	items := make([]dto.RankedBookItem, 0, len(books))
	for _, book := range books {
		items = append(items, dto.RankedBookItem{
			Rank:          book.Rank,
			BookID:        book.BookID,
			Title:         book.Title,
			Author:        book.Author,
			Rating:        book.Rating,
			ReviewCount:   book.ReviewCount,
			PublishedAt:   book.PublishedAt.Format("2006-01-02"),
			Thumbnail:     book.Thumbnail,
			Tags:          book.Tags,
			QiitaMentions: book.QiitaMentions,
			AmazonURL:     book.AmazonURL,
			RakutenURL:    book.RakutenURL,
		})
	}

	return &dto.RankingResponse{
		Range: rangeType,
		Items: items,
	}, nil
}
