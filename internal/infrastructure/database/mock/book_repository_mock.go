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
			BookID:       "9784297125967",
			Title:        "良いコード/悪いコードで学ぶ設計入門",
			Author:       "仙塲大也",
			Rating:       4.9,
			ReviewCount:  580,
			PublishedAt:  parseDate("2022-04-30"),
			Thumbnail:    "https://example.com/books/101.jpg",
			Tags:         []string{"設計", "ベストプラクティス", "Web"},
			ArticleCount: 12,
			AmazonURL:    "https://amazon.co.jp/dp/B09Y1MFB4K",
			RakutenURL:   "https://books.rakuten.co.jp/rb/17199622/",
			Rank:         1,
			CategoryID:   "web",
		},
		{
			BookID:       "9784873117584",
			Title:        "ゼロから作るDeep Learning",
			Author:       "斎藤康毅",
			Rating:       4.7,
			ReviewCount:  892,
			PublishedAt:  parseDate("2016-09-24"),
			Thumbnail:    "https://example.com/books/001.jpg",
			Tags:         []string{"AI", "機械学習", "Python"},
			ArticleCount: 45,
			AmazonURL:    "https://amazon.co.jp/dp/4873117585",
			RakutenURL:   "https://books.rakuten.co.jp/rb/14258520/",
			Rank:         2,
			CategoryID:   "ai-ml",
		},
		{
			BookID:       "9784873115658",
			Title:        "リーダブルコード",
			Author:       "Dustin Boswell",
			Rating:       4.8,
			ReviewCount:  1203,
			PublishedAt:  parseDate("2012-06-23"),
			Thumbnail:    "https://example.com/books/102.jpg",
			Tags:         []string{"コーディング", "可読性", "ベストプラクティス"},
			ArticleCount: 78,
			AmazonURL:    "https://amazon.co.jp/dp/4873115655",
			RakutenURL:   "https://books.rakuten.co.jp/rb/11753651/",
			Rank:         3,
			CategoryID:   "web",
		},
		{
			BookID:       "9784798163444",
			Title:        "AWSではじめるインフラ構築入門",
			Author:       "中垣健志",
			Rating:       4.5,
			ReviewCount:  324,
			PublishedAt:  parseDate("2021-03-16"),
			Thumbnail:    "https://example.com/books/201.jpg",
			Tags:         []string{"AWS", "インフラ", "クラウド"},
			ArticleCount: 23,
			AmazonURL:    "https://amazon.co.jp/dp/4798163449",
			RakutenURL:   "https://books.rakuten.co.jp/rb/16610598/",
			Rank:         4,
			CategoryID:   "cloud",
		},
		{
			BookID:       "9784297118372",
			Title:        "機械学習エンジニアのための本",
			Author:       "有賀康顕",
			Rating:       4.6,
			ReviewCount:  412,
			PublishedAt:  parseDate("2020-10-26"),
			Thumbnail:    "https://example.com/books/002.jpg",
			Tags:         []string{"機械学習", "MLOps", "実践"},
			ArticleCount: 31,
			AmazonURL:    "https://amazon.co.jp/dp/4297118378",
			RakutenURL:   "https://books.rakuten.co.jp/rb/16444083/",
			Rank:         5,
			CategoryID:   "ai-ml",
		},
		{
			BookID:       "9784873116860",
			Title:        "Web API: The Good Parts",
			Author:       "水野貴明",
			Rating:       4.4,
			ReviewCount:  267,
			PublishedAt:  parseDate("2014-11-21"),
			Thumbnail:    "https://example.com/books/103.jpg",
			Tags:         []string{"API", "設計", "REST"},
			ArticleCount: 19,
			AmazonURL:    "https://amazon.co.jp/dp/4873116864",
			RakutenURL:   "https://books.rakuten.co.jp/rb/12963679/",
			Rank:         6,
			CategoryID:   "web",
		},
		{
			BookID:       "9784295005643",
			Title:        "Kubernetes実践ガイド",
			Author:       "北山晋吾",
			Rating:       4.3,
			ReviewCount:  198,
			PublishedAt:  parseDate("2019-03-15"),
			Thumbnail:    "https://example.com/books/202.jpg",
			Tags:         []string{"Kubernetes", "コンテナ", "DevOps"},
			ArticleCount: 27,
			AmazonURL:    "https://amazon.co.jp/dp/4295005649",
			RakutenURL:   "https://books.rakuten.co.jp/rb/15791097/",
			Rank:         7,
			CategoryID:   "cloud",
		},
		{
			BookID:       "9784295003373",
			Title:        "Python機械学習プログラミング",
			Author:       "Sebastian Raschka",
			Rating:       4.5,
			ReviewCount:  534,
			PublishedAt:  parseDate("2018-03-21"),
			Thumbnail:    "https://example.com/books/003.jpg",
			Tags:         []string{"Python", "機械学習", "scikit-learn"},
			ArticleCount: 38,
			AmazonURL:    "https://amazon.co.jp/dp/4295003379",
			RakutenURL:   "https://books.rakuten.co.jp/rb/15365304/",
			Rank:         8,
			CategoryID:   "ai-ml",
		},
		{
			BookID:       "9784297113513",
			Title:        "インフラエンジニアの教科書",
			Author:       "佐野裕",
			Rating:       4.2,
			ReviewCount:  156,
			PublishedAt:  parseDate("2020-06-18"),
			Thumbnail:    "https://example.com/books/203.jpg",
			Tags:         []string{"インフラ", "ネットワーク", "サーバー"},
			ArticleCount: 14,
			AmazonURL:    "https://amazon.co.jp/dp/4297113511",
			RakutenURL:   "https://books.rakuten.co.jp/rb/16315789/",
			Rank:         9,
			CategoryID:   "cloud",
		},
		{
			BookID:       "9784873114798",
			Title:        "プログラマが知るべき97のこと",
			Author:       "和田卓人",
			Rating:       4.6,
			ReviewCount:  445,
			PublishedAt:  parseDate("2010-12-18"),
			Thumbnail:    "https://example.com/books/104.jpg",
			Tags:         []string{"プログラミング", "知識", "ベストプラクティス"},
			ArticleCount: 52,
			AmazonURL:    "https://amazon.co.jp/dp/4873114799",
			RakutenURL:   "https://books.rakuten.co.jp/rb/6598823/",
			Rank:         10,
			CategoryID:   "web",
		},
	}

	// 期間によってランキングを調整（モック用）
	switch rangeType {
	case "all":
		// 全期間ランキング（デフォルト）- そのままの順位
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

// parseDate 日付文字列を*time.Timeに変換
func parseDate(dateStr string) *time.Time {
	t, _ := time.Parse("2006-01-02", dateStr)
	return &t
}

// GetBookByID 書籍IDで書籍詳細を取得（モックデータ）
func (r *BookRepositoryMock) GetBookByID(ctx context.Context, bookID string) (*entity.BookDetail, error) {
	// モック書籍詳細データ
	mockBookDetails := map[string]*entity.BookDetail{
		"9784297125967": {
			BookID:        "9784297125967",
			Title:         "良いコード／悪いコードで学ぶ設計入門 〜保守しやすい成長し続けるコードの書き方〜",
			Author:        "仙塲 大也",
			PublishedDate: parseDate("2022-04-30"),
			Price:         3080,
			ISBN:          "978-4297125967",
			BookImage:     "https://example.com/books/9784297125967.jpg",
			Tags:          []string{"設計", "初学者", "初級者", "クリーンコード"},
			Overview:      "本書は、設計の基本から実務的な観点をチェックし、保守しやすく成長し続けるコードの書き方を学べる入門書です。設計の原則や実務的なテクニックまで幅広く学べます。",
			QiitaArticles: []entity.QiitaArticle{
				{
					Title:    "良いコード/悪いコードを読んで学んだ設計の基礎",
					URL:      "https://qiita.com/engineer_taro/items/abc123",
					Likes:    42,
					Stocks:   15,
					Comments: 3,
				},
				{
					Title:    "【書評】良いコード／悪いコードで学ぶ設計入門を読んでみた",
					URL:      "https://qiita.com/dev_hanako/items/def456",
					Likes:    128,
					Stocks:   67,
					Comments: 12,
				},
				{
					Title:    "設計入門書として「良いコード/悪いコード」がおすすめな理由",
					URL:      "https://qiita.com/programmer_jiro/items/ghi789",
					Likes:    85,
					Stocks:   32,
					Comments: 5,
				},
			},
			RakutenReviewSummary: entity.RakutenReviewSummary{
				AverageRating: 4.4,
				TotalReviews:  128,
			},
			PurchaseLinks: entity.PurchaseLinks{
				Amazon:  "https://www.amazon.co.jp/dp/4297125966",
				Rakuten: "https://books.rakuten.co.jp/",
			},
		},
		"9784873117584": {
			BookID:        "9784873117584",
			Title:         "ゼロから作るDeep Learning ―Pythonで学ぶディープラーニングの理論と実装",
			Author:        "斎藤 康毅",
			PublishedDate: parseDate("2016-09-24"),
			Price:         3740,
			ISBN:          "978-4873117584",
			BookImage:     "https://example.com/books/9784873117584.jpg",
			Tags:          []string{"AI", "機械学習", "Python", "ディープラーニング"},
			Overview:      "ディープラーニングの本格的な入門書。実際にPythonでディープラーニングを実装することで、ディープラーニングの原理を理解できます。",
			QiitaArticles: []entity.QiitaArticle{
				{
					Title:    "ゼロから作るDeep Learningで学んだニューラルネットワークの基礎",
					URL:      "https://qiita.com/ai_beginner/items/xyz111",
					Likes:    256,
					Stocks:   189,
					Comments: 24,
				},
				{
					Title:    "【読書メモ】ゼロから作るDeep Learning 第3章まとめ",
					URL:      "https://qiita.com/data_scientist/items/xyz222",
					Likes:    78,
					Stocks:   42,
					Comments: 8,
				},
				{
					Title:    "ゼロから作るDeep Learningで学んだニューラルネットワークの基礎",
					URL:      "https://qiita.com/ai_beginner/items/xyz111",
					Likes:    256,
					Stocks:   189,
					Comments: 24,
				},
				{
					Title:    "【読書メモ】ゼロから作るDeep Learning 第3章まとめ",
					URL:      "https://qiita.com/data_scientist/items/xyz222",
					Likes:    78,
					Stocks:   42,
					Comments: 8,
				},
				{
					Title:    "ゼロから作るDeep Learningで学んだニューラルネットワークの基礎",
					URL:      "https://qiita.com/ai_beginner/items/xyz111",
					Likes:    256,
					Stocks:   189,
					Comments: 24,
				},
				{
					Title:    "【読書メモ】ゼロから作るDeep Learning 第3章まとめ",
					URL:      "https://qiita.com/data_scientist/items/xyz222",
					Likes:    78,
					Stocks:   42,
					Comments: 8,
				},
				{
					Title:    "ゼロから作るDeep Learningで学んだニューラルネットワークの基礎",
					URL:      "https://qiita.com/ai_beginner/items/xyz111",
					Likes:    256,
					Stocks:   189,
					Comments: 24,
				},
				{
					Title:    "【読書メモ】ゼロから作るDeep Learning 第3章まとめ",
					URL:      "https://qiita.com/data_scientist/items/xyz222",
					Likes:    78,
					Stocks:   42,
					Comments: 8,
				},
			},
			RakutenReviewSummary: entity.RakutenReviewSummary{
				AverageRating: 4.6,
				TotalReviews:  312,
			},
			PurchaseLinks: entity.PurchaseLinks{
				Amazon:  "https://www.amazon.co.jp/dp/4873117585",
				Rakuten: "https://books.rakuten.co.jp/rb/14258520/",
			},
		},
		"9784873115658": {
			BookID:        "9784873115658",
			Title:         "リーダブルコード ―より良いコードを書くためのシンプルで実践的なテクニック",
			Author:        "Dustin Boswell, Trevor Foucher",
			PublishedDate: parseDate("2012-06-23"),
			Price:         2640,
			ISBN:          "978-4873115658",
			BookImage:     "https://example.com/books/9784873115658.jpg",
			Tags:          []string{"コーディング", "可読性", "ベストプラクティス"},
			Overview:      "コードは理解しやすくなければならない。本書はこの原則を日常のコーディングの様々な場面に適用する方法を紹介します。",
			QiitaArticles: []entity.QiitaArticle{
				{
					Title:    "リーダブルコードを読んで実践した命名規則のポイント",
					URL:      "https://qiita.com/senior_engineer/items/aaa111",
					Likes:    312,
					Stocks:   245,
					Comments: 18,
				},
				{
					Title:    "新人エンジニアがリーダブルコードを読んで変わったこと",
					URL:      "https://qiita.com/new_grad/items/bbb222",
					Likes:    189,
					Stocks:   98,
					Comments: 15,
				},
				{
					Title:    "【書評】リーダブルコードはなぜ名著と呼ばれるのか",
					URL:      "https://qiita.com/tech_lead/items/ccc333",
					Likes:    421,
					Stocks:   312,
					Comments: 32,
				},
			},
			RakutenReviewSummary: entity.RakutenReviewSummary{
				AverageRating: 4.7,
				TotalReviews:  567,
			},
			PurchaseLinks: entity.PurchaseLinks{
				Amazon:  "https://www.amazon.co.jp/dp/4873115655",
				Rakuten: "https://books.rakuten.co.jp/rb/11753651/",
			},
		},
	}

	// 指定されたIDの書籍詳細を検索
	if bookDetail, exists := mockBookDetails[bookID]; exists {
		return bookDetail, nil
	}

	// 存在しない場合はnilを返す
	return nil, nil
}
