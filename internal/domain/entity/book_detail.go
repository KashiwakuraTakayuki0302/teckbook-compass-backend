package entity

import "time"

// BookDetail 書籍詳細エンティティ
type BookDetail struct {
	BookID              string              // 書籍ID（ISBN形式）
	Title               string              // 書籍タイトル
	Author              string              // 著者名
	PublishedDate       time.Time           // 出版日
	Price               int                 // 価格
	ISBN                string              // ISBN
	BookImage           string              // 書籍画像URL
	Tags                []string            // タグ配列
	Overview            string              // 概要
	AboutThisBook       []string            // この本について（ポイント）
	TrendingPoints      []string            // 注目ポイント
	AmazonReviewSummary AmazonReviewSummary // Amazonレビューサマリー
	FeaturedReviews     []Review            // 注目レビュー
	PurchaseLinks       PurchaseLinks       // 購入リンク
}

// AmazonReviewSummary Amazonレビューサマリー
type AmazonReviewSummary struct {
	AverageRating float64 // 平均評価
	TotalReviews  int     // レビュー総数
}

// Review レビューエンティティ
type Review struct {
	Reviewer string    // レビュアー名
	Date     time.Time // レビュー日付
	Rating   float64   // 評価
	Comment  string    // コメント
}

// PurchaseLinks 購入リンク
type PurchaseLinks struct {
	Amazon  string // Amazon URL
	Rakuten string // 楽天 URL
}
