package entity

// Category カテゴリエンティティ
type Category struct {
	ID       string  // カテゴリID（例: "ai-ml", "web", "cloud"）
	Name     string  // カテゴリ名（例: "AI・機械学習"）
	Icon     string  // アイコン識別子（例: "ai-robot"）
	TrendTag string  // トレンドタグ（"hot", "popular", "attention"）
	Books    []*Book // カテゴリに属する書籍リスト
}
