package dto

// BookDetailResponse 書籍詳細取得APIのレスポンス
type BookDetailResponse struct {
	BookID              string                 `json:"bookId"`
	Title               string                 `json:"title"`
	Author              string                 `json:"author"`
	PublishedDate       string                 `json:"publishedDate"`
	Price               int                    `json:"price"`
	ISBN                string                 `json:"isbn"`
	BookImage           string                 `json:"bookImage"`
	Tags                []string               `json:"tags"`
	Overview            string                 `json:"overview"`
	AboutThisBook       []string               `json:"aboutThisBook"`
	TrendingPoints      []string               `json:"trendingPoints"`
	AmazonReviewSummary AmazonReviewSummaryDTO `json:"amazonReviewSummary"`
	FeaturedReviews     []ReviewDTO            `json:"featuredReviews"`
	PurchaseLinks       PurchaseLinksDTO       `json:"purchaseLinks"`
}

// AmazonReviewSummaryDTO Amazonレビューサマリー
type AmazonReviewSummaryDTO struct {
	AverageRating float64 `json:"averageRating"`
	TotalReviews  int     `json:"totalReviews"`
}

// ReviewDTO レビュー
type ReviewDTO struct {
	Reviewer string  `json:"reviewer"`
	Date     string  `json:"date"`
	Rating   float64 `json:"rating"`
	Comment  string  `json:"comment"`
}

// PurchaseLinksDTO 購入リンク
type PurchaseLinksDTO struct {
	Amazon  string `json:"amazon"`
	Rakuten string `json:"rakuten"`
}
