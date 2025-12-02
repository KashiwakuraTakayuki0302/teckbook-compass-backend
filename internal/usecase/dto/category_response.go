package dto

// CategoryWithBooksResponse カテゴリ別書籍取得APIのレスポンス
type CategoryWithBooksResponse struct {
	Items []CategoryItem `json:"items"`
}

// CategoryItem カテゴリアイテム
type CategoryItem struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	Icon     string     `json:"icon"`
	TrendTag string     `json:"trendTag"`
	Books    []BookItem `json:"books"`
}

// BookItem 書籍アイテム
type BookItem struct {
	Rank      int    `json:"rank"`
	BookID    string `json:"bookId"`
	Title     string `json:"title"`
	Thumbnail string `json:"thumbnail"`
}
