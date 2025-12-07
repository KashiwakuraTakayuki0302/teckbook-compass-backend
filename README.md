# TechBook Compass Backend

技術書レコメンデーションサービス「TechBook Compass」のBFF (Backend for Frontend) API

## 📖 概要

TechBook Compassは、Qiitaのトレンド情報とAmazonの書籍データを組み合わせて、今注目の技術分野と関連する技術書をレコメンデーションするサービスです。このリポジトリは、フロントエンド向けにデータを提供するBFF APIをGo言語で実装しています。

### 主な機能

- 📚 カテゴリ別技術書の取得
- 🔥 トレンドタグ付きカテゴリ表示（急上昇中、人気上昇、注目）
- 📊 技術書ランキング（全体、月間、日間、年間）※今後実装予定
- 🔍 技術書のキーワード検索 ※今後実装予定
- ⚙️ 日次バッチ処理（Qiita記事収集・書籍情報取得・スコアリング）

## 🏗️ アーキテクチャ

クリーンアーキテクチャを採用し、保守性とテスタビリティを重視した設計になっています。

```
teckbook-compass-backend/
├── api/                    # OpenAPI仕様
├── cmd/
│   ├── api/               # APIサーバーエントリーポイント
│   └── batch/             # バッチ処理エントリーポイント
├── internal/
│   ├── domain/            # ドメイン層（エンティティ、リポジトリIF）
│   ├── usecase/           # ユースケース層（ビジネスロジック）
│   ├── infrastructure/    # インフラ層（DB、外部API）
│   └── interface/         # インターフェース層（HTTPハンドラ）
├── migrations/            # データベースマイグレーション
├── pkg/                   # 共有ユーティリティ
└── docs/                  # ドキュメント
```

詳細は[アーキテクチャドキュメント](./docs/ImplementationPlan/bff-api-clean-architecture.md)を参照してください。

## 🚀 クイックスタート

### 必要な環境

- Go 1.25.4以上

### インストール

```bash
# リポジトリをクローン
git clone <repository-url>
cd teckbook-compass-backend

# 依存関係をインストール
go mod download
```

### サーバー起動

```bash
# Makefileを使用（推奨）
make run

# または直接実行
go run cmd/api/main.go
```

サーバーは http://localhost:8080 で起動します。

### 動作確認

```bash
# ヘルスチェック
curl http://localhost:8080/health

# カテゴリ別書籍取得
curl http://localhost:8080/categories/with-books | jq
```


## 🛠️ 開発

### Makefileコマンド

```bash
# 利用可能なコマンド一覧
make help

# サーバー起動
make run

# テスト実行
make test

# カバレッジ付きテスト
make test-coverage

# ビルド
make build

# バッチビルド
make build-batch

# クリーンアップ
make clean
```

### バッチ処理

```bash
# バッチ処理を実行（自動モード判定）
go run cmd/batch/main.go -run-batch

# 最新記事取得モードを強制
go run cmd/batch/main.go -run-batch -fetch-new

# 過去記事取得モードを強制
go run cmd/batch/main.go -run-batch -fetch-historical

# データベースマイグレーション
make db-migrate

# マイグレーションロールバック
make db-rollback
```

詳細は[日次バッチ処理ドキュメント](./docs/Walkthrough/daily-batch-walkthrough.md)を参照してください。

### 環境変数

#### サーバー設定

| 変数名 | 説明 | デフォルト値 |
|--------|------|-------------|
| `PORT` | サーバーポート番号 | `8080` |
| `ENV` | 実行環境 | `development` |

#### データベース設定

| 変数名 | 説明 | デフォルト値 |
|--------|------|-------------|
| `DB_HOST` | データベースホスト | `localhost` |
| `DB_PORT` | データベースポート | `5432` |
| `DB_USER` | データベースユーザー | `test` |
| `DB_PASSWORD` | データベースパスワード | `password` |
| `DB_NAME` | データベース名 | `teckbook` |

#### 外部API設定

| 変数名 | 説明 |
|--------|------|
| `QIITA_ACCESS_TOKEN` | Qiita APIアクセストークン |
| `RAKUTEN_APPLICATION_ID` | 楽天アプリケーションID |
| `RAKUTEN_APPLICATION_SECRET` | 楽天アプリケーションシークレット |
| `SLACK_WEBHOOK_URL` | Slack Webhook URL（通知用） |

```bash
# 環境変数の設定例
export PORT=3000
export ENV=production
export DB_HOST=your-db-host
export QIITA_ACCESS_TOKEN=your-token
```

### ディレクトリ構造

```
teckbook-compass-backend/
├── api/                            # API仕様
│   ├── openapi.yaml               # OpenAPI 3.0定義
│   └── README.md                  # API仕様の使い方
├── cmd/
│   └── api/
│       └── main.go                 # エントリーポイント
├── internal/
│   ├── domain/                     # ドメイン層
│   │   ├── entity/                # エンティティ
│   │   └── repository/            # リポジトリIF
│   ├── usecase/                    # ユースケース層
│   │   ├── category_usecase.go
│   │   └── dto/                   # レスポンスDTO
│   ├── infrastructure/             # インフラ層
│   │   ├── database/mock/         # モックリポジトリ
│   │   └── config/                # 設定管理
│   └── interface/                  # インターフェース層
│       ├── handler/               # HTTPハンドラ
│       └── router/                # ルーター
├── pkg/                            # 共有ユーティリティ
│   └── response/                  # レスポンスヘルパー
├── docs/                           # ドキュメント
│   ├── ImplementationPlan/        # 実装計画
│   └── Walkthrough/               # 実装レポート
├── Makefile                        # 開発用コマンド
├── go.mod
└── README.md
```

## 📚 ドキュメント

- [実装計画](./docs/ImplementationPlan/bff-api-clean-architecture.md) - クリーンアーキテクチャ設計書
- [実装完了レポート](./docs/Walkthrough/bff-api-implementation-report.md) - 検証結果と使用方法
- [日次バッチ処理](./docs/Walkthrough/daily-batch-walkthrough.md) - バッチ処理の詳細ドキュメント

## 🧪 テスト

```bash
# すべてのテストを実行
make test

# カバレッジ付きで実行
make test-coverage

# 特定のパッケージをテスト
go test ./internal/usecase/... -v
```

## 🔧 技術スタック

- **言語**: Go 1.25.4
- **Webフレームワーク**: Gin
- **データベース**: PostgreSQL（予定）
- **アーキテクチャ**: Clean Architecture

## 🗺️ ロードマップ

### 完了済み ✅

- [x] クリーンアーキテクチャの基盤構築
- [x] カテゴリ別技術書取得API
- [x] Makefileによる開発環境整備
- [x] モックデータでの動作確認
- [x] PostgreSQLデータベース接続
- [x] データベースマイグレーション
- [x] Qiita API統合バッチジョブ
- [x] 楽天ブックスAPI統合
- [x] 書籍スコアリング機能
- [x] Slack通知機能

### 今後の予定 📋

- [ ] Amazon API統合バッチジョブ
- [ ] 技術書ランキングAPI
- [ ] キーワード検索API
- [ ] 技術書詳細情報API
- [ ] キャッシュ層（Redis）
- [ ] ロギング・モニタリング
- [ ] CI/CDパイプライン

## 🤝 コントリビューション

1. このリポジトリをフォーク
2. フィーチャーブランチを作成 (`git checkout -b feature/amazing-feature`)
3. 変更をコミット (`git commit -m 'Add some amazing feature'`)
4. ブランチにプッシュ (`git push origin feature/amazing-feature`)
5. プルリクエストを作成

---
