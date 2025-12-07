package entity

import "time"

// BookScoreDaily 書籍スコア日次集計エンティティ
type BookScoreDaily struct {
	ID           int64     // ID
	BookID       string    // 書籍ID
	Date         time.Time // 集計日
	Score        float64   // スコア
	ArticleCount int       // 記事数
	CreatedAt    time.Time // 作成日時
}

// BookScore バッチ処理中に使用する書籍スコア
type BookScore struct {
	BookID            string    // 書籍ID（ISBN）
	Score             float64   // 累積スコア
	ArticleCount      int       // 記事数
	LatestArticleDate time.Time // 紐づく記事の最新投稿日
}

// AddScore スコアを加算し、最新記事投稿日を更新
func (bs *BookScore) AddScore(likes int, stocks int, articleCreatedAt time.Time) {
	// スコア計算: いいね数 + ストック数 * 1.5
	bs.Score += float64(likes) + float64(stocks)*1.5
	bs.ArticleCount++

	// 最新記事投稿日を更新
	if articleCreatedAt.After(bs.LatestArticleDate) {
		bs.LatestArticleDate = articleCreatedAt
	}
}

// TagCategoryMap タグとカテゴリのマッピング
type TagCategoryMap struct {
	ID         int64     // ID
	TagName    string    // タグ名
	CategoryID string    // カテゴリID
	CreatedAt  time.Time // 作成日時
}

// BookCategory 書籍とカテゴリの紐付け
type BookCategory struct {
	ID         int64     // ID
	BookID     string    // 書籍ID
	CategoryID string    // カテゴリID
	Score      float64   // カテゴリスコア
	Rank       int       // カテゴリ内ランク
	CreatedAt  time.Time // 作成日時
}
