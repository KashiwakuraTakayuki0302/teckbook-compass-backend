package entity

import "time"

// BookDetail 書籍詳細エンティティ
type BookDetail struct {
	BookID               string               // 書籍ID（ISBN形式）
	Title                string               // 書籍タイトル
	Author               string               // 著者名
	PublishedDate        time.Time            // 出版日
	Price                int                  // 価格
	ISBN                 string               // ISBN
	BookImage            string               // 書籍画像URL
	Tags                 []string             // タグ配列
	Overview             string               // 概要
	QiitaArticles        []QiitaArticle       // Qiita紹介記事一覧
	RakutenReviewSummary RakutenReviewSummary // 楽天レビューサマリー
	PurchaseLinks        PurchaseLinks        // 購入リンク
}

// QiitaArticle Qiita記事の紹介情報
type QiitaArticle struct {
	Title    string // 記事タイトル
	URL      string // 記事URL
	Likes    int    // いいね（LGTM）数
	Stocks   int    // ストック数
	Comments int    // コメント数
}

// RakutenReviewSummary 楽天レビューサマリー
type RakutenReviewSummary struct {
	AverageRating float64 // 平均評価
	TotalReviews  int     // レビュー件数
}

// PurchaseLinks 購入リンク
type PurchaseLinks struct {
	Amazon  string // Amazon URL
	Rakuten string // 楽天 URL
}
