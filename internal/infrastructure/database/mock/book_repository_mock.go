package mock

import (
	"context"
	"teckbook-compass-backend/internal/domain/entity"
	"time"
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

// GetRankings 総合ランキングを取得（モックデータ）
func (r *BookRepositoryMock) GetRankings(ctx context.Context, rangeType string, limit int, offset int, categoryID string) ([]*entity.Book, error) {
	// モックデータを作成
	allBooks := r.createMockRankingData(rangeType)

	// カテゴリフィルタリング
	var filteredBooks []*entity.Book
	if categoryID != "" {
		for _, book := range allBooks {
			if book.CategoryID == categoryID {
				filteredBooks = append(filteredBooks, book)
			}
		}
	} else {
		filteredBooks = allBooks
	}

	// ページネーション
	start := offset
	end := offset + limit

	if start >= len(filteredBooks) {
		return []*entity.Book{}, nil
	}

	if end > len(filteredBooks) {
		end = len(filteredBooks)
	}

	return filteredBooks[start:end], nil
}

// createMockRankingData 期間別のモックランキングデータを作成
func (r *BookRepositoryMock) createMockRankingData(rangeType string) []*entity.Book {
	// 基本的な書籍データ
	books := []*entity.Book{
		{
			ID:            101,
			Title:         "良いコード/悪いコードで学ぶ設計入門",
			Author:        "仙塲大也",
			Rating:        4.9,
			ReviewCount:   580,
			PublishedAt:   parseDate("2022-04-30"),
			Thumbnail:     "https://example.com/books/101.jpg",
			Tags:          []string{"設計", "ベストプラクティス", "Web"},
			QiitaMentions: 12,
			AmazonURL:     "https://amazon.co.jp/dp/B09Y1MFB4K",
			RakutenURL:    "https://books.rakuten.co.jp/rb/17199622/",
			Rank:          1,
			CategoryID:    "web",
		},
		{
			ID:            1,
			Title:         "ゼロから作るDeep Learning",
			Author:        "斎藤康毅",
			Rating:        4.7,
			ReviewCount:   892,
			PublishedAt:   parseDate("2016-09-24"),
			Thumbnail:     "https://example.com/books/001.jpg",
			Tags:          []string{"AI", "機械学習", "Python"},
			QiitaMentions: 45,
			AmazonURL:     "https://amazon.co.jp/dp/4873117585",
			RakutenURL:    "https://books.rakuten.co.jp/rb/14258520/",
			Rank:          2,
			CategoryID:    "ai-ml",
		},
		{
			ID:            102,
			Title:         "リーダブルコード",
			Author:        "Dustin Boswell",
			Rating:        4.8,
			ReviewCount:   1203,
			PublishedAt:   parseDate("2012-06-23"),
			Thumbnail:     "https://example.com/books/102.jpg",
			Tags:          []string{"コーディング", "可読性", "ベストプラクティス"},
			QiitaMentions: 78,
			AmazonURL:     "https://amazon.co.jp/dp/4873115655",
			RakutenURL:    "https://books.rakuten.co.jp/rb/11753651/",
			Rank:          3,
			CategoryID:    "web",
		},
		{
			ID:            201,
			Title:         "AWSではじめるインフラ構築入門",
			Author:        "中垣健志",
			Rating:        4.5,
			ReviewCount:   324,
			PublishedAt:   parseDate("2021-03-16"),
			Thumbnail:     "https://example.com/books/201.jpg",
			Tags:          []string{"AWS", "インフラ", "クラウド"},
			QiitaMentions: 23,
			AmazonURL:     "https://amazon.co.jp/dp/4798163449",
			RakutenURL:    "https://books.rakuten.co.jp/rb/16610598/",
			Rank:          4,
			CategoryID:    "cloud",
		},
		{
			ID:            2,
			Title:         "機械学習エンジニアのための本",
			Author:        "有賀康顕",
			Rating:        4.6,
			ReviewCount:   412,
			PublishedAt:   parseDate("2020-10-26"),
			Thumbnail:     "https://example.com/books/002.jpg",
			Tags:          []string{"機械学習", "MLOps", "実践"},
			QiitaMentions: 31,
			AmazonURL:     "https://amazon.co.jp/dp/4297118378",
			RakutenURL:    "https://books.rakuten.co.jp/rb/16444083/",
			Rank:          5,
			CategoryID:    "ai-ml",
		},
		{
			ID:            103,
			Title:         "Web API: The Good Parts",
			Author:        "水野貴明",
			Rating:        4.4,
			ReviewCount:   267,
			PublishedAt:   parseDate("2014-11-21"),
			Thumbnail:     "https://example.com/books/103.jpg",
			Tags:          []string{"API", "設計", "REST"},
			QiitaMentions: 19,
			AmazonURL:     "https://amazon.co.jp/dp/4873116864",
			RakutenURL:    "https://books.rakuten.co.jp/rb/12963679/",
			Rank:          6,
			CategoryID:    "web",
		},
		{
			ID:            202,
			Title:         "Kubernetes実践ガイド",
			Author:        "北山晋吾",
			Rating:        4.3,
			ReviewCount:   198,
			PublishedAt:   parseDate("2019-03-15"),
			Thumbnail:     "https://example.com/books/202.jpg",
			Tags:          []string{"Kubernetes", "コンテナ", "DevOps"},
			QiitaMentions: 27,
			AmazonURL:     "https://amazon.co.jp/dp/4295005649",
			RakutenURL:    "https://books.rakuten.co.jp/rb/15791097/",
			Rank:          7,
			CategoryID:    "cloud",
		},
		{
			ID:            3,
			Title:         "Python機械学習プログラミング",
			Author:        "Sebastian Raschka",
			Rating:        4.5,
			ReviewCount:   534,
			PublishedAt:   parseDate("2018-03-21"),
			Thumbnail:     "https://example.com/books/003.jpg",
			Tags:          []string{"Python", "機械学習", "scikit-learn"},
			QiitaMentions: 38,
			AmazonURL:     "https://amazon.co.jp/dp/4295003379",
			RakutenURL:    "https://books.rakuten.co.jp/rb/15365304/",
			Rank:          8,
			CategoryID:    "ai-ml",
		},
		{
			ID:            203,
			Title:         "インフラエンジニアの教科書",
			Author:        "佐野裕",
			Rating:        4.2,
			ReviewCount:   156,
			PublishedAt:   parseDate("2020-06-18"),
			Thumbnail:     "https://example.com/books/203.jpg",
			Tags:          []string{"インフラ", "ネットワーク", "サーバー"},
			QiitaMentions: 14,
			AmazonURL:     "https://amazon.co.jp/dp/4297113511",
			RakutenURL:    "https://books.rakuten.co.jp/rb/16315789/",
			Rank:          9,
			CategoryID:    "cloud",
		},
		{
			ID:            104,
			Title:         "プログラマが知るべき97のこと",
			Author:        "和田卓人",
			Rating:        4.6,
			ReviewCount:   445,
			PublishedAt:   parseDate("2010-12-18"),
			Thumbnail:     "https://example.com/books/104.jpg",
			Tags:          []string{"プログラミング", "知識", "ベストプラクティス"},
			QiitaMentions: 52,
			AmazonURL:     "https://amazon.co.jp/dp/4873114799",
			RakutenURL:    "https://books.rakuten.co.jp/rb/6598823/",
			Rank:          10,
			CategoryID:    "web",
		},
	}

	// 期間によってランキングを調整（モック用）
	switch rangeType {
	case "monthly":
		// 月次ランキングは少し順位を変更
		books[0].Rank = 2
		books[1].Rank = 1
		books[2].Rank = 3
	case "yearly":
		// 年次ランキングはさらに順位を変更
		books[0].Rank = 3
		books[1].Rank = 2
		books[2].Rank = 1
	}

	return books
}

// parseDate 日付文字列をtime.Timeに変換
func parseDate(dateStr string) time.Time {
	t, _ := time.Parse("2006-01-02", dateStr)
	return t
}
