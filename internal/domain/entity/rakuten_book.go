package entity

import (
	"fmt"
	"time"
)

// RakutenBookResponse 楽天ブックスAPI レスポンス
type RakutenBookResponse struct {
	Count     int               `json:"count"`
	Page      int               `json:"page"`
	First     int               `json:"first"`
	Last      int               `json:"last"`
	Hits      int               `json:"hits"`
	Carrier   int               `json:"carrier"`
	PageCount int               `json:"pageCount"`
	Items     []RakutenBookItem `json:"Items"`
}

// RakutenBookItem 楽天ブックスAPI 書籍アイテム
type RakutenBookItem struct {
	Item RakutenBook `json:"Item"`
}

// RakutenBook 楽天ブックスAPI 書籍情報
type RakutenBook struct {
	Title          string `json:"title"`
	TitleKana      string `json:"titleKana"`
	SubTitle       string `json:"subTitle"`
	SubTitleKana   string `json:"subTitleKana"`
	SeriesName     string `json:"seriesName"`
	SeriesNameKana string `json:"seriesNameKana"`
	Contents       string `json:"contents"`
	Author         string `json:"author"`
	AuthorKana     string `json:"authorKana"`
	PublisherName  string `json:"publisherName"`
	Size           string `json:"size"`
	ISBN           string `json:"isbn"`
	ItemCaption    string `json:"itemCaption"`
	SalesDate      string `json:"salesDate"`
	ItemPrice      int    `json:"itemPrice"`
	ListPrice      int    `json:"listPrice"`
	DiscountRate   int    `json:"discountRate"`
	DiscountPrice  int    `json:"discountPrice"`
	ItemURL        string `json:"itemUrl"`
	AffiliateURL   string `json:"affiliateUrl"`
	SmallImageURL  string `json:"smallImageUrl"`
	MediumImageURL string `json:"mediumImageUrl"`
	LargeImageURL  string `json:"largeImageUrl"`
	Chirayomiurl   string `json:"chipirayomiUrl"`
	Availability   string `json:"availability"`
	PostageFlag    int    `json:"postageFlag"`
	LimitedFlag    int    `json:"limitedFlag"`
	ReviewCount    int    `json:"reviewCount"`
	ReviewAverage  string `json:"reviewAverage"`
	BooksGenreID   string `json:"booksGenreId"`
}

// ToBook RakutenBookをBookエンティティに変換
func (rb *RakutenBook) ToBook() *Book {
	// 評価を文字列から数値に変換
	var rating float64
	if rb.ReviewAverage != "" {
		// 文字列を数値に変換（エラーは無視）
		_, _ = fmt.Sscanf(rb.ReviewAverage, "%f", &rating)
	}

	// 出版日をパース
	var publishedAt time.Time
	if rb.SalesDate != "" {
		// "2024年01月01日" 形式をパース
		publishedAt, _ = parseJapaneseDate(rb.SalesDate)
	}

	return &Book{
		BookID:      rb.ISBN,
		Title:       rb.Title,
		Author:      rb.Author,
		Rating:      rating,
		ReviewCount: rb.ReviewCount,
		PublishedAt: publishedAt,
		Thumbnail:   rb.LargeImageURL,
		RakutenURL:  rb.AffiliateURL,
	}
}

// parseJapaneseDate 日本語形式の日付をパース（例: "2024年01月01日"）
func parseJapaneseDate(dateStr string) (time.Time, error) {
	// パターンに応じてパース
	patterns := []string{
		"2006年01月02日",
		"2006年01月",
		"2006年1月2日",
		"2006年1月",
	}

	for _, pattern := range patterns {
		if t, err := time.Parse(pattern, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("failed to parse date: %s", dateStr)
}
