package dto

// BookDetailResponse 書籍詳細取得APIのレスポンス
type BookDetailResponse struct {
	BookID               string                  `json:"bookId"`
	Title                string                  `json:"title"`
	Author               string                  `json:"author"`
	PublishedDate        *string                 `json:"publishedDate,omitempty"`
	Price                int                     `json:"price"`
	ISBN                 string                  `json:"isbn"`
	BookImage            string                  `json:"bookImage"`
	Tags                 []string                `json:"tags"`
	Overview             string                  `json:"overview"`
	QiitaArticles        []QiitaArticleDTO       `json:"qiitaArticles"`
	RakutenReviewSummary RakutenReviewSummaryDTO `json:"rakutenReviewSummary"`
	PurchaseLinks        PurchaseLinksDTO        `json:"purchaseLinks"`
}

// QiitaArticleDTO Qiita記事の紹介情報
type QiitaArticleDTO struct {
	Title    string `json:"title"`
	URL      string `json:"url"`
	Likes    int    `json:"likes"`
	Stocks   int    `json:"stocks"`
	Comments int    `json:"comments"`
}

// RakutenReviewSummaryDTO 楽天レビューサマリー
type RakutenReviewSummaryDTO struct {
	AverageRating float64 `json:"averageRating"`
	TotalReviews  int     `json:"totalReviews"`
}

// PurchaseLinksDTO 購入リンク
type PurchaseLinksDTO struct {
	Amazon  string `json:"amazon"`
	Rakuten string `json:"rakuten"`
}
