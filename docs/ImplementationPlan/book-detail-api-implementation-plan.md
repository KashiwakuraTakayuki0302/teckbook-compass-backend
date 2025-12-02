# 書籍詳細取得API実装計画

## 概要

書籍詳細取得API (`GET /books/{bookId}`) を実装します。このAPIは、指定された書籍IDの詳細情報（書籍基本情報、レビュー、購入リンク）を1回のリクエストで取得できるようにします。

既存の `/rankings` や `/categories/with-books` エンドポイントと同様のアーキテクチャパターン（Handler → Usecase → Repository）を踏襲し、現在はモックデータで実装します。

## 提案する変更内容

### Domain層

#### [NEW] [book_detail.go](file:///Users/kashiwakura/develop/teckbook-compass-backend/internal/domain/entity/book_detail.go)

書籍詳細用のエンティティを新規作成します：

```go
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

type AmazonReviewSummary struct {
    AverageRating float64 // 平均評価
    TotalReviews  int     // レビュー総数
}

type Review struct {
    Reviewer string    // レビュアー名
    Date     time.Time // レビュー日付
    Rating   float64   // 評価
    Comment  string    // コメント
}

type PurchaseLinks struct {
    Amazon  string // Amazon URL
    Rakuten string // 楽天 URL
}
```

#### [MODIFY] [book_repository.go](file:///Users/kashiwakura/develop/teckbook-compass-backend/internal/domain/repository/book_repository.go)

`BookRepository` インターフェースに新しいメソッドを追加します：

```go
GetBookByID(ctx context.Context, bookID string) (*entity.BookDetail, error)
```

---

### Infrastructure層

#### [MODIFY] [book_repository_mock.go](file:///Users/kashiwakura/develop/teckbook-compass-backend/internal/infrastructure/database/mock/book_repository_mock.go)

`GetBookByID` メソッドのモック実装を追加します：

- 複数の書籍詳細データ（ISBN形式のID、価格、概要、レビューなど）
- 存在しないIDの場合は `nil` を返す

---

### Usecase層

#### [NEW] [book_detail_usecase.go](file:///Users/kashiwakura/develop/teckbook-compass-backend/internal/usecase/book_detail_usecase.go)

新しいユースケースを作成します：

- `BookRepository` への依存
- `GetBookDetail(ctx, bookID)` メソッド
- エンティティからDTOへの変換ロジック

#### [NEW] [book_detail_response.go](file:///Users/kashiwakura/develop/teckbook-compass-backend/internal/usecase/dto/book_detail_response.go)

レスポンスDTOを作成します：

```go
type BookDetailResponse struct {
    BookID              string                 `json:"bookId"`
    Title               string                 `json:"title"`
    Author              string                 `json:"author"`
    PublishedDate       string                 `json:"publishedDate"`
    Price               int                    `json:"price"`
    ISBN                string                 `json:"isbn"`
    BookImage           string                 `json:"bookImage"`
    Tags                []string               `json:"tags"`
    Overview            string                 `json:"overview"`
    AboutThisBook       []string               `json:"aboutThisBook"`
    TrendingPoints      []string               `json:"trendingPoints"`
    AmazonReviewSummary AmazonReviewSummaryDTO `json:"amazonReviewSummary"`
    FeaturedReviews     []ReviewDTO            `json:"featuredReviews"`
    PurchaseLinks       PurchaseLinksDTO       `json:"purchaseLinks"`
}

type AmazonReviewSummaryDTO struct {
    AverageRating float64 `json:"averageRating"`
    TotalReviews  int     `json:"totalReviews"`
}

type ReviewDTO struct {
    Reviewer string  `json:"reviewer"`
    Date     string  `json:"date"`
    Rating   float64 `json:"rating"`
    Comment  string  `json:"comment"`
}

type PurchaseLinksDTO struct {
    Amazon  string `json:"amazon"`
    Rakuten string `json:"rakuten"`
}
```

---

### Interface層

#### [NEW] [book_detail_handler.go](file:///Users/kashiwakura/develop/teckbook-compass-backend/internal/interface/handler/book_detail_handler.go)

新しいハンドラを作成します：

- `BookDetailUsecase` への依存
- `GetBookDetail(c *gin.Context)` メソッド
- パスパラメータ `bookId` の取得とバリデーション
- 404エラーハンドリング（書籍が見つからない場合）

#### [MODIFY] [router.go](file:///Users/kashiwakura/develop/teckbook-compass-backend/internal/interface/router/router.go)

新しいエンドポイントを追加します：

```go
r.GET("/books/:bookId", bookDetailHandler.GetBookDetail)
```

`SetupRouter` 関数のシグネチャを変更し、`bookDetailHandler` パラメータを追加します。

---

### Application層

#### [MODIFY] [main.go](file:///Users/kashiwakura/develop/teckbook-compass-backend/cmd/api/main.go)

依存性注入を更新します：

1. `BookDetailUsecase` のインスタンス化
2. `BookDetailHandler` のインスタンス化
3. `SetupRouter` に `bookDetailHandler` を渡す

---

### API定義

#### [MODIFY] [openapi.yaml](file:///Users/kashiwakura/develop/teckbook-compass-backend/api/openapi.yaml)

提供されたOpenAPI定義を追加します：

- `/books/{bookId}` エンドポイント
- パスパラメータ（bookId）
- `BookDetail` スキーマ
- `AmazonReviewSummary` スキーマ
- `Review` スキーマ
- `PurchaseLinks` スキーマ

## 検証計画

### 自動テスト

現在、プロジェクトには自動テストが存在しないため、手動検証のみを行います。

### 手動検証

以下のコマンドでサーバーを起動し、curlでAPIをテストします：

```bash
# サーバー起動
cd /Users/kashiwakura/develop/teckbook-compass-backend
make run
```

別のターミナルで以下のテストを実行：

```bash
# 1. 存在する書籍の詳細取得
curl -X GET "http://localhost:8080/books/9784297125967"

# 2. 別の書籍の詳細取得
curl -X GET "http://localhost:8080/books/9784873117584"

# 3. リーダブルコードの詳細取得
curl -X GET "http://localhost:8080/books/9784873115658"

# 4. 存在しない書籍ID（404エラー）
curl -X GET "http://localhost:8080/books/notexist"
```

期待される結果：
- 存在する書籍：ステータスコード200、書籍詳細JSONを返却
- 存在しない書籍：ステータスコード404、エラーメッセージを返却
- 全フィールド（bookId, title, author, price, overview, reviews等）が含まれる

