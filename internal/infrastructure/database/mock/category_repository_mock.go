package mock

import (
	"context"
	"teckbook-compass-backend/internal/domain/entity"
)

// CategoryRepositoryMock カテゴリリポジトリのモック実装
type CategoryRepositoryMock struct{}

// NewCategoryRepositoryMock カテゴリリポジトリモックのコンストラクタ
func NewCategoryRepositoryMock() *CategoryRepositoryMock {
	return &CategoryRepositoryMock{}
}

// GetCategoriesWithBooks カテゴリと関連書籍を取得（モックデータ）
func (r *CategoryRepositoryMock) GetCategoriesWithBooks(ctx context.Context, limit int) ([]*entity.Category, error) {
	// モックデータ: AI・機械学習カテゴリ
	aiCategory := &entity.Category{
		ID:       "ai-ml",
		Name:     "AI・機械学習",
		Icon:     "robot",
		TrendTag: "hot",
		Books: []*entity.Book{
			{
				BookID:     "9784873117584",
				Title:      "ゼロから作るDeep Learning",
				Thumbnail:  "https://example.com/books/001.jpg",
				Rank:       1,
				CategoryID: "ai-ml",
			},
			{
				BookID:     "9784297118372",
				Title:      "機械学習エンジニアのための本",
				Thumbnail:  "https://example.com/books/002.jpg",
				Rank:       2,
				CategoryID: "ai-ml",
			},
			{
				BookID:     "9784295003373",
				Title:      "Python機械学習プログラミング",
				Thumbnail:  "https://example.com/books/003.jpg",
				Rank:       3,
				CategoryID: "ai-ml",
			},
		},
	}

	// モックデータ: Web開発カテゴリ
	webCategory := &entity.Category{
		ID:       "web",
		Name:     "Web開発",
		Icon:     "pc",
		TrendTag: "popular",
		Books: []*entity.Book{
			{
				BookID:     "9784873115658",
				Title:      "リーダブルコード",
				Thumbnail:  "https://example.com/books/101.jpg",
				Rank:       1,
				CategoryID: "web",
			},
			{
				BookID:     "9784297125967",
				Title:      "良いコード/悪いコードで学ぶ設計入門",
				Thumbnail:  "https://example.com/books/102.jpg",
				Rank:       2,
				CategoryID: "web",
			},
			{
				BookID:     "9784873116860",
				Title:      "Web API: The Good Parts",
				Thumbnail:  "https://example.com/books/103.jpg",
				Rank:       3,
				CategoryID: "web",
			},
		},
	}

	// モックデータ: クラウド・インフラカテゴリ
	cloudCategory := &entity.Category{
		ID:       "cloud",
		Name:     "クラウド・インフラ",
		Icon:     "cloud",
		TrendTag: "attention",
		Books: []*entity.Book{
			{
				BookID:     "9784798163444",
				Title:      "AWSではじめるインフラ構築入門",
				Thumbnail:  "https://example.com/books/201.jpg",
				Rank:       1,
				CategoryID: "cloud",
			},
			{
				BookID:     "9784295005643",
				Title:      "Kubernetes実践ガイド",
				Thumbnail:  "https://example.com/books/202.jpg",
				Rank:       2,
				CategoryID: "cloud",
			},
			{
				BookID:     "9784297113513",
				Title:      "インフラエンジニアの教科書",
				Thumbnail:  "https://example.com/books/203.jpg",
				Rank:       3,
				CategoryID: "cloud",
			},
		},
	}

	categories := []*entity.Category{aiCategory, webCategory, cloudCategory}

	// limitが指定されている場合は制限
	if limit > 0 && limit < len(categories) {
		categories = categories[:limit]
	}

	return categories, nil
}
