# Devin.md

## 概要
このドキュメントは、本プロジェクト（Teckbook Compass Backend）の開発に参加するAIエージェント（Devin等）のためのガイドラインです。

## 1. 環境・ランタイム
- **言語**: Go
- **バージョン**: 1.25.4
- **パッケージ管理**: Go Modules (`go.mod`)

## 2. アーキテクチャ方針
**クリーンアーキテクチャ**を採用しています。依存方向は `Infrastructure` -> `Interface` -> `Use Case` -> `Domain` となり、内側のレイヤーは外側のレイヤーを知りません。

### ディレクトリ構造と責務
- **cmd/**: アプリケーションのエントリポイント。
    - `api/`: APIサーバー（Lambda + Gin）のエントリポイント。DI（依存性の注入）はここで行います。
    - `batch/`: バッチ処理のエントリポイント。
- **internal/**: 外部からインポートされないアプリケーション固有のコード。
    - **domain/**: ドメイン層。ビジネスロジックの中核。
        - `entity/`: ドメインモデル（構造体 struct）。JSONタグを含みます。
        - `repository/`: リポジトリのインターフェース定義。
    - **usecase/**: ユースケース層。アプリケーションのビジネスロジック。ドメイン層のインターフェースを利用して処理を組み立てます。
    - **interface/**: インターフェース層。
        - `handler/`: HTTPリクエストを受け取り、ユースケースを呼び出してレスポンスを返します。
        - `router/`: ルーティング定義。
    - **infrastructure/**: インフラストラクチャ層。詳細な実装。
        - `database/`: データベース接続、リポジトリの実装（PostgreSQL）。
        - `config/`: 設定値の管理。
        - `secrets/`: AWS Secrets Manager 等の外部サービス連携。
- **pkg/**: 外部からも利用可能な共有パッケージ。
    - `response/`: APIレスポンスの統一フォーマット定義。
- **api/**: OpenAPI 定義 (`openapi.yaml`)。
- **migrations/**: データベースマイグレーションファイル。

### _core/ の役割
（現状のコードベースには `_core/` ディレクトリは存在しません。共通処理は `pkg/` または `internal/` 内の各層に配置されています。）

## 3. API 設計ルール
- **API 定義**: `api/openapi.yaml` (OpenAPI 3.0) に仕様を記述します。
- **URL 命名規則**: ケバブケース（例: `/categories/with-books`）を使用します。
- **フレームワーク**: Gin (`github.com/gin-gonic/gin`) を使用していますが、AWS Lambda 上で動作させるために `ginv2` アダプタを使用しています。
- **レスポンス形式**:
    - 成功時: JSON形式。`pkg/response` パッケージの `Success` 関数を使用。
    - エラー時: `{"error": "メッセージ"}` 形式。`pkg/response` パッケージの `Error` 関数を使用。
- **エラーハンドリング**:
    - ハンドラ層でエラーをキャッチし、適切なHTTPステータスコードと共にレスポンスを返します。
    - 500エラーはサーバー内部エラーとして扱います。

## 4. DB / 永続化レイヤ
- **DB 種別**: PostgreSQL
- **アクセス方法**: **生SQL** (`database/sql`) を使用します。ORM（GORM等）は使用していません。
- **実装場所**: `internal/infrastructure/database/postgres/`
- **マイグレーション**: `golang-migrate` を使用。
    - ファイル場所: `migrations/`
    - コマンド: `make db-migrate` (Up), `make db-rollback` (Down)
- **トランザクション**: 必要な場合は `database/sql` のトランザクション機能を使用します。

## 5. 外部サービス連携
- **AWS**:
    - Lambda (実行環境)
    - Secrets Manager (DB認証情報など)
    - SDK: `github.com/aws/aws-sdk-go-v2`
- **その他**:
    - Amazon Product Advertising API (書籍情報)
    - Rakuten Books API (書籍情報)

## 6. 実行・開発コマンド
`Makefile` に定義されています。
- `make run`: ローカルでサーバーを起動 (`go run cmd/api/main.go`)
- `make build`: APIバイナリをビルド
- `make build-batch`: バッチバイナリをビルド
- `make test`: テスト実行
- `make db-test`: DB接続テスト
- `make db-migrate`: DBマイグレーション実行

## 7. Devin（AIエージェント）の禁止事項
1.  **アーキテクチャ違反**:
    - 下位レイヤー（例: Domain）から上位レイヤー（例: Infrastructure）をインポートしてはいけません。
    - ビジネスロジックを Handler に書かないでください（Usecase に記述すること）。
2.  **ORM の導入**:
    - 現在の方針（生SQL）に従ってください。勝手に GORM 等を導入しないでください。
3.  **破壊的な変更**:
    - 既存の API コントラクト (`openapi.yaml`) を勝手に変更しないでください。変更が必要な場合はユーザーに確認してください。
4.  **不必要な依存関係の追加**:
    - `go.mod` に不要なライブラリを追加しないでください。
5.  **テスト無視**:
    - 変更を加えた際は既存のテストが通ることを確認してください。
