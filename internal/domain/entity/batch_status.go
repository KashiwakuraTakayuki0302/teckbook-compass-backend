package entity

import "time"

// BatchStatus バッチ状態管理エンティティ
type BatchStatus struct {
	ID             string     // バッチ識別子
	LastFetchedAt  *time.Time // 最新記事取得時の基準日時
	NextPage       int        // 過去記事取得用の次ページ番号
	LastRunAt      *time.Time // 最後にバッチを実行した日時
	LastNewFetchAt *time.Time // 最後に最新記事取得を実行した日時
	CreatedAt      time.Time  // 作成日時
	UpdatedAt      time.Time  // 更新日時
}

// BatchStatusID バッチ状態のID定数
const (
	BatchStatusIDQiitaFetch = "qiita_fetch"
)

// ShouldFetchNewArticles 最新記事を取得すべきかどうかを判定
// 1日1回（最後の最新記事取得から24時間以上経過している場合）に実行
func (bs *BatchStatus) ShouldFetchNewArticles() bool {
	if bs.LastNewFetchAt == nil {
		return true
	}
	return time.Since(*bs.LastNewFetchAt) >= 24*time.Hour
}

// GetFetchMode 取得モードを判定
func (bs *BatchStatus) GetFetchMode() FetchMode {
	if bs.ShouldFetchNewArticles() {
		return FetchModeNew
	}
	return FetchModeHistorical
}

// FetchMode 取得モード
type FetchMode int

const (
	FetchModeNew        FetchMode = iota // 最新記事取得モード
	FetchModeHistorical                  // 過去記事取得モード
)

func (fm FetchMode) String() string {
	switch fm {
	case FetchModeNew:
		return "最新記事取得"
	case FetchModeHistorical:
		return "過去記事取得"
	default:
		return "不明"
	}
}
