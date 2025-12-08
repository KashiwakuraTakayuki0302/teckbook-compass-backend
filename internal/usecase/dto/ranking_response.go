package dto

// RankingResponse 総合ランキング取得APIのレスポンス
type RankingResponse struct {
	Range string           `json:"range"`
	Items []RankedBookItem `json:"items"`
}

// RankedBookItem ランキング書籍アイテム
type RankedBookItem struct {
	Rank         int      `json:"rank"`
	BookID       string   `json:"bookId"`
	Title        string   `json:"title"`
	Author       string   `json:"author"`
	Rating       float64  `json:"rating"`
	ReviewCount  int      `json:"reviewCount"`
	PublishedAt  *string  `json:"publishedAt,omitempty"`
	Thumbnail    string   `json:"thumbnail"`
	Tags         []string `json:"tags"`
	ArticleCount int      `json:"articleCount"`
	AmazonURL    string   `json:"amazonUrl"`
	RakutenURL   string   `json:"rakutenUrl"`
}
