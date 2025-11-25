# クリーンアーキテクチャによるBFF API実装

## 概要

Go言語でクリーンアーキテクチャの原則に従ったRESTful BFF（Backend for Frontend）APIを実装します。このサービスは、QiitaのAPIデータとAmazonの購入情報に基づいて技術書のレコメンデーションを提供し、まずTOPページ用の`GET /categories/with-books`エンドポイントから開始します。

## ユーザー確認事項

> [!IMPORTANT]
> **データベース選択**: PostgreSQLを使用します（ユーザー確認済み）。

> [!IMPORTANT]
> **モックデータ戦略**: 初期実装ではインメモリのモックデータを使用します。QiitaとAmazonのデータを取得するバッチジョブはまだ実装されていないため、これによりAPI構造をすぐに構築・テストできます（ユーザー確認済み）。

> [!NOTE]
> **今後のバッチジョブ**: 計画には、後でバッチジョブによって入力される実際のデータベーステーブルに接続するリポジトリインターフェースが含まれています。現在のモック実装は簡単に置き換え可能です。

## 提案する変更内容

### クリーンアーキテクチャのディレクトリ構造

クリーンアーキテクチャの層に従ってコードベースを整理します:

```
teckbook-compass-backend/
├── api/                            # API仕様
│   ├── openapi.yaml               # OpenAPI 3.0定義
│   └── README.md                  # API仕様の使い方
├── cmd/
│   └── api/
│       └── main.go                 # アプリケーションエントリーポイント
├── internal/
│   ├── domain/                     # エンタープライズビジネスルール
│   │   ├── entity/
│   │   │   ├── category.go
│   │   │   └── book.go
│   │   └── repository/
│   │       ├── category_repository.go
│   │       └── book_repository.go
│   ├── usecase/                    # アプリケーションビジネスルール
│   │   ├── category_usecase.go
│   │   └── dto/
│   │       └── category_response.go
│   ├── infrastructure/             # フレームワーク＆ドライバー
│   │   ├── database/
│   │   │   └── mock/
│   │   │       ├── category_repository_mock.go
│   │   │       └── book_repository_mock.go
│   │   └── config/
│   │       └── config.go
│   └── interface/                  # インターフェースアダプター
│       ├── handler/
│       │   └── category_handler.go
│       └── router/
│           └── router.go
├── pkg/                            # 共有ユーティリティ
│   └── response/
│       └── response.go
└── Makefile                        # 便利コマンド集
```

---

### ドメイン層（エンタープライズビジネスルール）

#### [NEW] [entity/category.go](file:///Users/kashiwakura/develop/teckbook-compass-backend/internal/domain/entity/category.go)

`Category`エンティティを定義。フィールド: ID、Name、Icon、TrendTag、関連Books。

#### [NEW] [entity/book.go](file:///Users/kashiwakura/develop/teckbook-compass-backend/internal/domain/entity/book.go)

`Book`エンティティを定義。フィールド: ID、Title、Thumbnail、Rank、CategoryID、ランキング用メタデータ。

#### [NEW] [repository/category_repository.go](file:///Users/kashiwakura/develop/teckbook-compass-backend/internal/domain/repository/category_repository.go)

リポジトリメソッドを定義するインターフェース:
- `GetCategoriesWithBooks(ctx context.Context, limit int) ([]*entity.Category, error)`

#### [NEW] [repository/book_repository.go](file:///Users/kashiwakura/develop/teckbook-compass-backend/internal/domain/repository/book_repository.go)

リポジトリメソッドを定義するインターフェース:
- `GetTopBooksByCategory(ctx context.Context, categoryID string, limit int) ([]*entity.Book, error)`

---

### ユースケース層（アプリケーションビジネスルール）

#### [NEW] [usecase/category_usecase.go](file:///Users/kashiwakura/develop/teckbook-compass-backend/internal/usecase/category_usecase.go)

カテゴリ操作のビジネスロジックを実装:
- `GetCategoriesWithBooks(ctx context.Context) (*dto.CategoryWithBooksResponse, error)`

リポジトリへの呼び出しを調整し、ドメインエンティティをDTOに変換します。

#### [NEW] [usecase/dto/category_response.go](file:///Users/kashiwakura/develop/teckbook-compass-backend/internal/usecase/dto/category_response.go)

API仕様に一致するレスポンスDTOを定義:
- `CategoryWithBooksResponse`
- `CategoryItem`
- `BookItem`

---

### インフラストラクチャ層（フレームワーク＆ドライバー）

#### [NEW] [infrastructure/database/mock/category_repository_mock.go](file:///Users/kashiwakura/develop/teckbook-compass-backend/internal/infrastructure/database/mock/category_repository_mock.go)

`CategoryRepository`のモック実装。APIサンプルに一致するハードコードされたデータを返します。

#### [NEW] [infrastructure/database/mock/book_repository_mock.go](file:///Users/kashiwakura/develop/teckbook-compass-backend/internal/infrastructure/database/mock/book_repository_mock.go)

`BookRepository`のモック実装。ハードコードされた書籍データを返します。

#### [NEW] [infrastructure/config/config.go](file:///Users/kashiwakura/develop/teckbook-compass-backend/internal/infrastructure/config/config.go)

サーバーポート、データベース接続（将来）、環境設定の構成管理。

---

### インターフェース層（インターフェースアダプター）

#### [NEW] [interface/handler/category_handler.go](file:///Users/kashiwakura/develop/teckbook-compass-backend/internal/interface/handler/category_handler.go)

カテゴリエンドポイント用のHTTPハンドラ:
- `GetCategoriesWithBooks(c *gin.Context)` - `GET /categories/with-books`を処理

#### [NEW] [interface/router/router.go](file:///Users/kashiwakura/develop/teckbook-compass-backend/internal/interface/router/router.go)

Ginルーターのセットアップ:
- CORSミドルウェア
- ルート定義
- ハンドラ登録

---

### 共有ユーティリティ

#### [NEW] [pkg/response/response.go](file:///Users/kashiwakura/develop/teckbook-compass-backend/pkg/response/response.go)

標準化されたレスポンスヘルパー:
- `Success(c *gin.Context, data interface{})`
- `Error(c *gin.Context, code int, message string)`

---

### アプリケーションエントリーポイント

#### [NEW] [cmd/api/main.go](file:///Users/kashiwakura/develop/teckbook-compass-backend/cmd/api/main.go)

アプリケーションブートストラップ:
- 構成の初期化
- 依存性注入のセットアップ
- モックリポジトリの作成
- ユースケースとハンドラの初期化
- HTTPサーバーの起動

---

## 検証計画

### 自動テスト

#### ユニットテスト
```bash
# すべてのユニットテストを実行
go test ./internal/usecase/... -v

# カバレッジ付きで実行
go test ./internal/usecase/... -cover
```

作成するテスト:
- `internal/usecase/category_usecase_test.go` - モックリポジトリでビジネスロジックをテスト

#### 統合テスト
```bash
# 統合テストを実行
go test ./internal/interface/handler/... -v
```

作成するテスト:
- `internal/interface/handler/category_handler_test.go` - モックユースケースでHTTPハンドラをテスト

### 手動検証

#### 1. サーバーを起動
```bash
cd /Users/kashiwakura/develop/teckbook-compass-backend
go run cmd/api/main.go
```

期待される出力:
```
Server starting on :8080
```

#### 2. エンドポイントをテスト
```bash
curl -X GET http://localhost:8080/categories/with-books | jq
```

期待されるレスポンス構造:
```json
{
  "items": [
    {
      "id": "ai-ml",
      "name": "AI・機械学習",
      "icon": "ai-robot",
      "trendTag": "hot",
      "books": [
        {
          "rank": 1,
          "id": "book_001",
          "title": "ゼロから作るDeep Learning",
          "thumbnail": "https://example.com/books/001.jpg"
        }
      ]
    }
  ]
}
```

#### 3. レスポンス形式を検証
- すべての必須フィールドが存在することを確認
- データ型が仕様と一致することを検証
- 書籍がランク順に並んでいることを確認
- トレンドタグが正しいことを検証

#### 4. エラーハンドリングをテスト
```bash
# 無効なルートでテスト
curl -X GET http://localhost:8080/invalid-route
```

期待される結果: 404エラーレスポンス

---

## 実装メモ

### 依存性注入
コンストラクタベースの依存性注入を使用します（この初期実装ではDIフレームワークは不要）。Google Wireは既に`go.mod`にあり、後で追加可能です。

### エラーハンドリング
- ドメイン層でカスタムエラー型を使用
- 各層でコンテキストを含めてエラーをラップ
- ハンドラで適切なHTTPステータスコードを返す

### 今後の拡張
- データベース接続プールの追加（PostgreSQL/MySQL）
- Qiita API統合用バッチジョブの実装
- Amazon API統合用バッチジョブの実装
- キャッシュ層の追加（Redis）
- ロギングミドルウェアの追加
- リクエストバリデーションの追加
- ✅ **完了**: APIドキュメント（OpenAPI 3.0仕様、Makefile、Swagger UI対応）
