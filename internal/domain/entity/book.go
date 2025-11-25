package entity

import "time"

// Book 書籍エンティティ
type Book struct {
	ID         string    // 書籍ID（例: "book_001"）
	Title      string    // 書籍タイトル
	Thumbnail  string    // サムネイル画像URL
	Rank       int       // カテゴリ内のランク
	CategoryID string    // 所属カテゴリID
	CreatedAt  time.Time // 作成日時
	UpdatedAt  time.Time // 更新日時
}
