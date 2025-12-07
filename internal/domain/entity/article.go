package entity

import "time"

// Article Qiita記事エンティティ
type Article struct {
	ID          string    // Qiita記事ID
	Title       string    // 記事タイトル
	URL         string    // 記事URL
	Body        string    // 記事本文（HTML or Markdown）
	Likes       int       // いいね数
	Stocks      int       // ストック数
	Comments    int       // コメント数
	Tags        []string  // タグ名配列
	PublishedAt time.Time // 公開日時
	CreatedAt   time.Time // DB登録日時
	UpdatedAt   time.Time // DB更新日時
}

// ArticleTag 記事タグエンティティ
type ArticleTag struct {
	ID        int64     // タグID
	ArticleID string    // 記事ID
	TagName   string    // タグ名
	CreatedAt time.Time // 作成日時
}

// ArticleBook 記事と書籍の紐付けエンティティ
type ArticleBook struct {
	ID        int64     // ID
	ArticleID string    // 記事ID
	BookID    string    // 書籍ID（ISBN）
	CreatedAt time.Time // 作成日時
}

// QiitaAPIArticle Qiita APIから取得した記事
type QiitaAPIArticle struct {
	ID            string     `json:"id"`
	Title         string     `json:"title"`
	URL           string     `json:"url"`
	Body          string     `json:"body"`
	RenderedBody  string     `json:"rendered_body"`
	LikesCount    int        `json:"likes_count"`
	StocksCount   int        `json:"stocks_count"`
	CommentsCount int        `json:"comments_count"`
	Tags          []QiitaTag `json:"tags"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// QiitaTag Qiita APIから取得したタグ
type QiitaTag struct {
	Name string `json:"name"`
}

// GetTagNames タグ名の配列を取得
func (a *QiitaAPIArticle) GetTagNames() []string {
	tags := make([]string, len(a.Tags))
	for i, t := range a.Tags {
		tags[i] = t.Name
	}
	return tags
}

// ToArticle QiitaAPIArticleをArticleエンティティに変換
func (a *QiitaAPIArticle) ToArticle() *Article {
	return &Article{
		ID:          a.ID,
		Title:       a.Title,
		URL:         a.URL,
		Body:        a.Body,
		Likes:       a.LikesCount,
		Stocks:      a.StocksCount,
		Comments:    a.CommentsCount,
		Tags:        a.GetTagNames(),
		PublishedAt: a.CreatedAt,
	}
}
