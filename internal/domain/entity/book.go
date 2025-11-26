package entity

import "time"

// Book 書籍エンティティ
type Book struct {
	ID            int       // 書籍ID（主キー）
	Title         string    // 書籍タイトル
	Author        string    // 著者名
	Rating        float64   // 評価（0.0-5.0）
	ReviewCount   int       // レビュー数
	PublishedAt   time.Time // 出版日
	Thumbnail     string    // サムネイル画像URL
	Tags          []string  // タグ配列
	QiitaMentions int       // Qiita言及数
	AmazonURL     string    // Amazon URL
	RakutenURL    string    // 楽天 URL
	Rank          int       // カテゴリ内のランク
	CategoryID    string    // 所属カテゴリID
	CreatedAt     time.Time // 作成日時
	UpdatedAt     time.Time // 更新日時
}
