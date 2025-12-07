package repository

import (
	"context"
	"time"

	"teckbook-compass-backend/internal/domain/entity"
)

// BatchRepository バッチ処理用リポジトリインターフェース
type BatchRepository interface {
	// Article関連
	ArticleExists(ctx context.Context, articleID string) (bool, error)
	SaveArticle(ctx context.Context, article *entity.Article) error
	SaveArticleTags(ctx context.Context, articleID string, tags []string) error
	SaveArticleBook(ctx context.Context, articleID string, bookID string) error

	// Book関連
	BookExists(ctx context.Context, bookID string) (bool, error)
	GetBookIDByISBN(ctx context.Context, isbn string) (string, error) // ISBN-10/13どちらでも検索可能
	SaveBook(ctx context.Context, book *entity.RakutenBook) error
	UpdateBookScore(ctx context.Context, bookID string, score float64) error
	GetExistingBookScore(ctx context.Context, bookID string) (float64, error)

	// BookScoreDaily関連
	SaveBookScoreDaily(ctx context.Context, bookID string, date time.Time, score float64, articleCount int) error

	// TagCategoryMap関連
	GetCategoryIDsByTags(ctx context.Context, tags []string) ([]string, error)

	// BookCategory関連
	SaveBookCategories(ctx context.Context, bookID string, categoryIDs []string) error

	// BatchStatus関連
	GetBatchStatus(ctx context.Context, id string) (*entity.BatchStatus, error)
	UpdateBatchStatusForNewFetch(ctx context.Context, id string, lastFetchedAt time.Time) error
	UpdateBatchStatusForHistoricalFetch(ctx context.Context, id string, nextPage int) error

	// ErrorLog関連
	SaveErrorLog(ctx context.Context, log *ErrorLog) error
}

// ErrorLog エラーログ
type ErrorLog struct {
	BatchProcess   string
	ErrorType      string
	Level          string
	APIName        string
	Endpoint       string
	StatusCode     int
	RequestPayload interface{}
	ResponseBody   interface{}
	RelatedID      string
	Message        string
}
