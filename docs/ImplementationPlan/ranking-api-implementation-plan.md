# 総合ランキング取得API実装計画

## 概要

総合ランキング取得API (`GET /rankings`) を実装します。このAPIは、技術書の総合ランキングを期間（全期間/月次/年次）、カテゴリ、ページネーション対応で取得できるようにします。

既存の `/categories/with-books` エンドポイントと同様のアーキテクチャパターン（Handler → Usecase → Repository）を踏襲し、現在はモックデータで実装します。

## 提案する変更内容

### Domain層

#### [MODIFY] [book.go](file:///Users/kashiwakura/develop/teckbook-compass-backend/internal/domain/entity/book.go)

`Book` エンティティを拡張し、OpenAPI定義の `RankedBook` スキーマに対応するフィールドを追加します：

- `Author` (著者名)
- `Rating` (評価)
- `ReviewCount` (レビュー数)
- `PublishedAt` (出版日)
- `Tags` (タグ配列)
- `QiitaMentions` (Qiita言及数)
- `AmazonURL` (Amazon URL)
- `RakutenURL` (楽天 URL)

#### [MODIFY] [book_repository.go](file:///Users/kashiwakura/develop/teckbook-compass-backend/internal/domain/repository/book_repository.go)

`BookRepository` インターフェースに新しいメソッドを追加します：

```go
GetRankings(ctx context.Context, range string, limit int, offset int, categoryID string) ([]*entity.Book, error)
```

---

### Infrastructure層

#### [MODIFY] [book_repository_mock.go](file:///Users/kashiwakura/develop/teckbook-compass-backend/internal/infrastructure/database/mock/book_repository_mock.go)

`GetRankings` メソッドのモック実装を追加します。以下のモックデータを返します：

- 複数の書籍データ（著者、評価、レビュー数、タグなど含む）
- `range` パラメータ（all/monthly/yearly）に応じた異なるランキング
- `categoryID` が指定された場合のフィルタリング
- `limit` と `offset` によるページネーション

---

### Usecase層

#### [NEW] [ranking_usecase.go](file:///Users/kashiwakura/develop/teckbook-compass-backend/internal/usecase/ranking_usecase.go)

新しいユースケースを作成します：

- `BookRepository` への依存
- `GetRankings(ctx, range, limit, offset, categoryID)` メソッド
- エンティティからDTOへの変換ロジック

#### [NEW] [ranking_response.go](file:///Users/kashiwakura/develop/teckbook-compass-backend/internal/usecase/dto/ranking_response.go)

レスポンスDTOを作成します：

```go
type RankingResponse struct {
    Range string          `json:"range"`
    Items []RankedBookDTO `json:"items"`
}

type RankedBookDTO struct {
    Rank          int      `json:"rank"`
    ID            int      `json:"id"`
    Title         string   `json:"title"`
    Author        string   `json:"author"`
    Rating        float64  `json:"rating"`
    ReviewCount   int      `json:"reviewCount"`
    PublishedAt   string   `json:"publishedAt"`
    Thumbnail     string   `json:"thumbnail"`
    Tags          []string `json:"tags"`
    QiitaMentions int      `json:"qiitaMentions"`
    AmazonURL     string   `json:"amazonUrl"`
    RakutenURL    string   `json:"rakutenUrl"`
}
```

---

### Interface層

#### [NEW] [ranking_handler.go](file:///Users/kashiwakura/develop/teckbook-compass-backend/internal/interface/handler/ranking_handler.go)

新しいハンドラを作成します：

- `RankingUsecase` への依存
- `GetRankings(c *gin.Context)` メソッド
- クエリパラメータのバリデーション（range, limit, offset, category）
- デフォルト値の設定（range=all, limit=5, offset=0）

#### [MODIFY] [router.go](file:///Users/kashiwakura/develop/teckbook-compass-backend/internal/interface/router/router.go)

新しいエンドポイントを追加します：

```go
r.GET("/rankings", rankingHandler.GetRankings)
```

`SetupRouter` 関数のシグネチャを変更し、`rankingHandler` パラメータを追加します。

---

### Application層

#### [MODIFY] [main.go](file:///Users/kashiwakura/develop/teckbook-compass-backend/cmd/api/main.go)

依存性注入を更新します：

1. `RankingUsecase` のインスタンス化
2. `RankingHandler` のインスタンス化
3. `SetupRouter` に `rankingHandler` を渡す

---

### API定義

#### [MODIFY] [openapi.yaml](file:///Users/kashiwakura/develop/teckbook-compass-backend/api/openapi.yaml)

提供されたOpenAPI定義を追加します：

- `/rankings` エンドポイント
- クエリパラメータ（range: all/monthly/yearly, limit, offset, category）
- `RankedBook` スキーマ（既存のものを拡張）

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
# 1. 基本的なランキング取得（デフォルト: all, limit=5）
curl -X GET "http://localhost:8080/rankings"

# 2. 月次ランキング取得
curl -X GET "http://localhost:8080/rankings?range=monthly"

# 3. 年次ランキング取得
curl -X GET "http://localhost:8080/rankings?range=yearly"

# 4. limit指定
curl -X GET "http://localhost:8080/rankings?limit=10"

# 5. ページネーション
curl -X GET "http://localhost:8080/rankings?limit=5&offset=5"

# 6. カテゴリフィルタ
curl -X GET "http://localhost:8080/rankings?category=ai-ml"

# 7. 複合条件
curl -X GET "http://localhost:8080/rankings?range=yearly&limit=3&category=web"
```

期待される結果：
- ステータスコード200
- JSON形式のレスポンス
- `range` と `items` フィールドを含む
- 各書籍に必要なフィールド（rank, id, title, author, rating等）が含まれる
- パラメータに応じた適切なデータが返される
