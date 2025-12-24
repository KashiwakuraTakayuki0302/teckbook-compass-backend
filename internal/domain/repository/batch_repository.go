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

	// Amazon API用
	// GetBooksWithoutAmazonURLByScore スコアが高い順にamazon_urlがない書籍を取得
	GetBooksWithoutAmazonURLByScore(ctx context.Context, limit int) ([]*BookForAmazonUpdate, error)
	// UpdateBookAmazonURL 書籍のAmazon URLを更新
	UpdateBookAmazonURL(ctx context.Context, bookID string, amazonURL string) error

	// カテゴライズバッチ用
	// GetBooksWithoutCategory カテゴリが未設定の書籍を取得（スコア順）
	GetBooksWithoutCategory(ctx context.Context, limit int) ([]*BookForCategorize, error)
	// GetAllCategories 全カテゴリを取得
	GetAllCategories(ctx context.Context) ([]*CategoryInfo, error)
	// SaveBookCategory 書籍のカテゴリを保存
	SaveBookCategory(ctx context.Context, bookID string, categoryID string) error
<<<<<<< HEAD

	// Category関連
	// UpdateCategoryScores 全カテゴリのスコアを更新（book_categoriesに紐づくbook_scores_dailyの合計）
	UpdateCategoryScores(ctx context.Context) error
=======
>>>>>>> 7290383 (CPG-31 技術書カテゴライズバッチ実装)
}

// BookForAmazonUpdate Amazon URL更新用の書籍情報
type BookForAmazonUpdate struct {
	ID     string  // ISBN-13
	ISBN10 *string // ISBN-10
	Title  string
	Score  float64
}

// BookForCategorize カテゴライズ用の書籍情報
type BookForCategorize struct {
	ID       string // ISBN-13
	Title    string
	Overview string
}

// CategoryInfo カテゴリ情報
type CategoryInfo struct {
	ID   string // カテゴリID（例: "ai-ml"）
	Name string // カテゴリ名（例: "AI / 機械学習"）
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
