# 日次バッチ処理 実装レポート

## 概要

TechBook Compassの技術書データを自動収集・更新する日次バッチ処理を実装しました。Qiita APIから技術書に関連する記事を収集し、楽天ブックスAPIで書籍情報を取得、スコアリングを行います。

## アーキテクチャ

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         Daily Batch Process                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐                  │
│  │  Qiita API   │───▶│  Book        │───▶│  Rakuten     │                  │
│  │  Client      │    │  Extractor   │    │  API Client  │                  │
│  └──────────────┘    └──────────────┘    └──────────────┘                  │
│         │                   │                   │                            │
│         ▼                   ▼                   ▼                            │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                      Batch Usecase                                   │   │
│  │  - 記事取得・保存                                                     │   │
│  │  - 書籍抽出・保存                                                     │   │
│  │  - スコア計算・保存                                                   │   │
│  │  - カテゴリ振り分け                                                   │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│         │                                                                    │
│         ▼                                                                    │
│  ┌──────────────┐    ┌──────────────┐                                      │
│  │  PostgreSQL  │    │    Slack     │                                      │
│  │  Database    │    │  通知        │                                      │
│  └──────────────┘    └──────────────┘                                      │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

## ディレクトリ構造

```
teckbook-compass-backend/
├── cmd/batch/
│   └── main.go                              # バッチエントリーポイント
├── internal/
│   ├── domain/
│   │   ├── entity/
│   │   │   ├── article.go                   # Qiita記事エンティティ
│   │   │   ├── batch_status.go              # バッチ状態管理エンティティ
│   │   │   ├── book_score.go                # 書籍スコアエンティティ
│   │   │   └── rakuten_book.go              # 楽天書籍エンティティ
│   │   └── repository/
│   │       └── batch_repository.go          # バッチ用リポジトリIF
│   ├── usecase/
│   │   └── batch_usecase.go                 # バッチ処理ユースケース
│   └── infrastructure/
│       ├── database/postgres/
│       │   └── batch_repository.go          # バッチ用リポジトリ実装
│       ├── external/
│       │   ├── qiita_client.go              # Qiita APIクライアント
│       │   ├── rakuten_client.go            # 楽天ブックスAPIクライアント
│       │   └── slack_client.go              # Slack通知クライアント
│       └── extractor/
│           └── book_extractor.go            # 書籍情報抽出器
└── migrations/
    ├── 000003_add_isbn10_to_books.*.sql     # ISBN-10カラム追加
    └── 000004_create_batch_statuses.*.sql   # バッチ状態テーブル
```

## 処理フロー

### 1. 取得モードの判定

バッチは2つのモードで動作します：

| モード | 説明 | 実行条件 |
|--------|------|----------|
| **最新記事取得** | 前回取得以降の新しい記事を取得 | 24時間以上経過した場合（1日1回） |
| **過去記事取得** | ページネーションで過去記事を遡って取得 | 最新記事取得を行わない場合 |

```go
// バッチ状態に基づいてモードを自動判定
func (bs *BatchStatus) GetFetchMode() FetchMode {
    if bs.ShouldFetchNewArticles() {
        return FetchModeNew
    }
    return FetchModeHistorical
}
```

### 2. Qiita記事の取得

以下の検索クエリで技術書関連記事を収集：

```go
var SearchQueries = []string{
    "技術書",
    "書籍",
    "本の紹介",
    "おすすめ本",
    "読んだ本",
    "入門書",
    "書評",
    "レビュー",
    "読書",
}
```

### 3. 書籍情報の抽出

記事本文から以下のパターンで書籍情報を抽出：

- **ISBN-13**: `978-4-XXXX-XXXX-X` 形式
- **ISBN-10**: `4-XXXX-XXXX-X` 形式
- **ASIN**: Amazonリンクから抽出
- **書籍タイトル**: 特定のパターンから抽出

### 4. 楽天ブックスAPIで書籍情報取得

抽出したISBNを使用して楽天ブックスAPIから詳細情報を取得：

- 書籍タイトル
- 著者名
- 出版社
- 価格
- 書影URL
- 商品URL

### 5. スコア計算

記事の反響に基づいて書籍スコアを算出：

```go
// スコア計算ロジック
func (bs *BookScore) AddScore(likes, stocks int, publishedAt time.Time) {
    bs.Score += likes*2 + stocks*3  // いいね×2 + ストック×3
    bs.ArticleCount++
    // 最新の投稿日を保持
    if publishedAt.After(bs.LatestArticleDate) {
        bs.LatestArticleDate = publishedAt
    }
}
```

### 6. カテゴリ自動振り分け

記事に付けられたタグに基づいて書籍をカテゴリに分類。タグとカテゴリのマッピングはデータベースで管理。

### 7. バッチ状態の更新

処理完了後、次回実行のための状態を更新：

- 最新記事取得モード: `last_fetched_at` を更新
- 過去記事取得モード: `next_page` を更新

## コマンドラインオプション

```bash
# バッチ処理を実行（自動モード判定）
go run cmd/batch/main.go -run-batch

# 最新記事取得モードを強制
go run cmd/batch/main.go -run-batch -fetch-new

# 過去記事取得モードを強制
go run cmd/batch/main.go -run-batch -fetch-historical

# ドライラン（DBへの書き込みなし）
go run cmd/batch/main.go -run-batch -dry-run

# データベース接続テスト
go run cmd/batch/main.go -test-connection

# マイグレーション実行
go run cmd/batch/main.go -migrate-up

# マイグレーションロールバック
go run cmd/batch/main.go -migrate-down
go run cmd/batch/main.go -migrate-down -migrate-steps=2
```

## Makefileコマンド

```bash
# バッチをビルド
make build-batch

# データベース接続テスト
make db-test

# マイグレーション実行
make db-migrate

# マイグレーションロールバック
make db-rollback
make db-rollback-all
```

## 環境変数

### データベース設定

| 変数名 | 説明 | デフォルト値 |
|--------|------|-------------|
| `DB_HOST` | データベースホスト | `localhost` |
| `DB_PORT` | データベースポート | `5432` |
| `DB_USER` | データベースユーザー | `test` |
| `DB_PASSWORD` | データベースパスワード | `password` |
| `DB_NAME` | データベース名 | `teckbook` |
| `DB_SSLMODE` | SSLモード | `disable` |

### Qiita API設定

| 変数名 | 説明 | 備考 |
|--------|------|------|
| `QIITA_ACCESS_TOKEN` | Qiita APIアクセストークン | 必須 |
| `QIITA_BASE_URL` | Qiita API URL | デフォルト: `https://qiita.com/api/v2` |

### 楽天ブックスAPI設定

| 変数名 | 説明 | 備考 |
|--------|------|------|
| `RAKUTEN_APPLICATION_ID` | 楽天アプリケーションID | 必須 |
| `RAKUTEN_APPLICATION_SECRET` | 楽天アプリケーションシークレット | 必須 |
| `RAKUTEN_AFFILIATE_ID` | 楽天アフィリエイトID | 任意 |
| `RAKUTEN_BASE_URL` | 楽天API URL | デフォルト: `https://app.rakuten.co.jp/services/api/BooksBook/Search/20170404` |

### Slack通知設定

| 変数名 | 説明 | 備考 |
|--------|------|------|
| `SLACK_WEBHOOK_URL` | Slack Webhook URL | 簡易通知用 |
| `SLACK_BOT_TOKEN` | Slack Bot Token | スレッド返信用（`xoxb-...`形式） |
| `SLACK_CHANNEL_ID` | 通知先チャンネルID | Bot Token使用時に必要 |

## 実行結果サンプル

```
===========================================
  TeckBook Compass Daily Batch
  開始時刻: 2024-01-15 09:00:00
===========================================
バッチ処理を開始します...
取得モード: 最新記事取得（自動判定）
Step 1: Qiita APIから記事を取得中...
取得した記事数: 150
Step 2-4: 各記事から技術書を抽出中...
進捗: 50/150 記事を処理済み
進捗: 100/150 記事を処理済み
Step 5: 書籍スコアを保存中...
Step 6: Amazon API処理はスキップ（後で追加）
Step 7: カテゴリ振り分けは記事処理時に完了済み
Step 8: バッチ状態を更新中...
最新記事取得完了 - 次回まで過去記事取得モードに移行
バッチ処理完了: 処理時間 5m30s
===========================================
  バッチ処理結果
===========================================
  取得モード:       最新記事取得
  処理した記事数:   150
  新規記事数:       45
  処理した書籍数:   23
  エラー数:         2
  処理時間:         5m30.123s
===========================================
  終了時刻: 2024-01-15 09:05:30
===========================================
```

## Slack通知

バッチ処理の開始・終了時にSlackへ通知が送信されます：

### 開始メッセージ
```
🚀 TechBook Compass バッチ処理開始
取得モード: 最新記事取得
```

### 結果メッセージ
```
✅ TechBook Compass バッチ処理完了
━━━━━━━━━━━━━━━━━━━━
📊 処理結果
  取得モード: 最新記事取得
  処理記事数: 150件
  新規記事数: 45件
  処理書籍数: 23件
  エラー数: 2件
  処理時間: 5m30s
```

## データベーススキーマ

### batch_statuses テーブル

バッチ処理の状態を管理：

```sql
CREATE TABLE batch_statuses (
    id VARCHAR(50) PRIMARY KEY,
    last_fetched_at TIMESTAMP,        -- 最新記事取得時の基準日時
    next_page INTEGER DEFAULT 1,       -- 過去記事取得用の次ページ番号
    last_run_at TIMESTAMP,             -- 最後にバッチを実行した日時
    last_new_fetch_at TIMESTAMP,       -- 最後に最新記事取得を実行した日時
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

### 関連テーブル

- `articles`: Qiita記事情報
- `article_tags`: 記事とタグの紐付け
- `article_books`: 記事と書籍の紐付け
- `books`: 書籍情報
- `book_scores_daily`: 日次スコア
- `book_categories`: 書籍とカテゴリの紐付け
- `batch_error_logs`: エラーログ

## エラーハンドリング

エラーは `batch_error_logs` テーブルに記録され、処理は継続されます：

```go
type ErrorLog struct {
    BatchProcess string    // "daily_batch"
    ErrorType    string    // "article_processing", "book_fetch" など
    Level        string    // "ERROR", "WARNING"
    RelatedID    string    // 関連するID（記事ID、ISBN等）
    Message      string    // エラーメッセージ
}
```

## レートリミット対策

外部APIのレートリミットに対応：

| API | 待機時間 | 備考 |
|-----|---------|------|
| Qiita API | 1秒/リクエスト、2秒/クエリ | ページ間、クエリ間で待機 |
| 楽天ブックスAPI | 300ms/リクエスト | 書籍情報取得時 |
| 記事処理間 | 500ms | 全体の処理速度調整 |

## 今後の拡張予定

- [ ] Amazon Product Advertising API連携
- [ ] Zenn記事の取得対応
- [ ] エラーリトライ機構
- [ ] 並列処理による高速化
- [ ] バッチ実行スケジューリング（cron）

