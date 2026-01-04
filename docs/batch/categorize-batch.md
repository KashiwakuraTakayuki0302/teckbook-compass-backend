# カテゴライズバッチ処理

## 概要

TechBook Compassの技術書に対して、ChatGPT APIを使用してカテゴリを自動分類するバッチ処理です。
書籍のタイトル、概要、ISBN13を元に、10個のカテゴリの中から最も適切な1つを選択し、`book_categories`テーブルに保存します。

## アーキテクチャ

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      Categorize Batch Process                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐                  │
│  │  PostgreSQL  │───▶│  Categorize  │───▶│  ChatGPT     │                  │
│  │  (books)     │    │  Usecase     │    │  API Client  │                  │
│  └──────────────┘    └──────────────┘    └──────────────┘                  │
│         │                   │                   │                            │
│         │                   ▼                   │                            │
│         │           ┌──────────────┐            │                            │
│         │           │  カテゴリ判定  │◀──────────┘                            │
│         │           │  結果をDBに保存 │                                       │
│         │           └──────────────┘                                        │
│         │                   │                                                │
│         ▼                   ▼                                                │
│  ┌──────────────┐    ┌──────────────┐                                      │
│  │ book_        │    │    Slack     │                                      │
│  │ categories   │    │  通知        │                                      │
│  └──────────────┘    └──────────────┘                                      │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

## 処理フロー

1. **カテゴリ未設定の書籍を取得**
   - `book_categories`にエントリがない書籍を、スコアが高い順に取得
   - 取得件数は環境変数`CATEGORIZE_LIMIT`で制御（デフォルト: 40件）

2. **ChatGPT APIでカテゴリを判定**
   - 書籍のタイトル、概要、ISBN13をプロンプトに含めて送信
   - 10カテゴリの中から最も適切な1つを返答

3. **結果をDBに保存**
   - `book_categories`テーブルに書籍とカテゴリの紐付けを保存

4. **Slack通知**
   - 処理結果とトークン使用量を通知

### 書籍取得の優先順位（スコア順）

カテゴリ未設定の書籍は、`book_scores_daily`テーブルのスコア合計が高い順に処理されます。
これにより、Qiita記事で多く言及されている人気の技術書から優先的にカテゴライズされます。

```sql
SELECT b.id, b.title, COALESCE(b.overview, '') as overview, 
       COALESCE(SUM(bsd.score), 0) as total_score
FROM books b
LEFT JOIN book_categories bc ON b.id = bc.book_id
LEFT JOIN book_scores_daily bsd ON b.id = bsd.book_id
WHERE bc.id IS NULL
GROUP BY b.id, b.title, b.overview
ORDER BY total_score DESC, 
         b.latest_mentioned_at DESC NULLS LAST, 
         b.created_at DESC
LIMIT $1
```

**ソート順序:**
1. スコア合計（高い順）
2. 最終言及日時（新しい順）
3. 作成日時（新しい順）

## カテゴリ一覧

| ID | カテゴリコード | カテゴリ名 |
|----|---------------|-----------|
| 1 | `ai-ml` | AI / 機械学習 |
| 2 | `frontend` | Web / フロントエンド |
| 3 | `mobile` | モバイル / アプリ開発 |
| 4 | `cloud` | クラウド（AWS / GCP / Azure） |
| 5 | `infra-devops` | インフラ / DevOps |
| 6 | `backend` | バックエンド / API / Webアーキテクチャ |
| 7 | `database` | データベース / データエンジニアリング |
| 8 | `security` | セキュリティ |
| 9 | `beginner-cs` | プログラミング入門 / CS基礎 |
| 10 | `pm-business` | PM / プロダクト / ビジネス・キャリア |

## 環境変数

### ChatGPT API設定（必須）

| 変数名 | 説明 | デフォルト値 |
|--------|------|-------------|
| `CHATGPT_API_KEY` | OpenAI APIキー | なし（必須） |
| `CHATGPT_ENABLED` | APIを有効にするか | `false` |
| `CHATGPT_MODEL` | 使用するモデル | `gpt-4o-mini` |
| `CHATGPT_BASE_URL` | API URL | `https://api.openai.com/v1` |

### バッチ処理設定

| 変数名 | 説明 | デフォルト値 |
|--------|------|-------------|
| `BATCH_TYPE` | バッチタイプ | なし |
| `CATEGORIZE_LIMIT` | 1回の処理件数上限 | `40` |

### データベース設定

| 変数名 | 説明 | デフォルト値 |
|--------|------|-------------|
| `DB_HOST` | データベースホスト | `localhost` |
| `DB_PORT` | データベースポート | `5432` |
| `DB_USER` | データベースユーザー | `test` |
| `DB_PASSWORD` | データベースパスワード | `password` |
| `DB_NAME` | データベース名 | `teckbook` |
| `DB_SSLMODE` | SSLモード | `disable` |

### Slack通知設定（オプション）

| 変数名 | 説明 | 備考 |
|--------|------|------|
| `SLACK_WEBHOOK_URL` | Slack Webhook URL | 通知用 |
| `SLACK_BOT_TOKEN` | Slack Bot Token | スレッド返信用 |
| `SLACK_CHANNEL_ID` | 通知先チャンネルID | Bot Token使用時に必要 |

## 起動方法

### ローカル環境

```bash
# 基本的な起動（デフォルト50冊処理）
BATCH_TYPE=categorize CHATGPT_ENABLED=true CHATGPT_API_KEY=sk-xxx go run ./cmd/batch/...

# 処理件数を指定
BATCH_TYPE=categorize CHATGPT_ENABLED=true CHATGPT_API_KEY=sk-xxx CATEGORIZE_LIMIT=10 go run ./cmd/batch/...
```

### .envファイルを使用する場合

```bash
# .envファイルに設定を記述
cat << 'EOF' > .env
DB_HOST=localhost
DB_PORT=5432
DB_USER=test
DB_PASSWORD=password
DB_NAME=teckbook
DB_SSLMODE=disable

CHATGPT_API_KEY=sk-xxx
CHATGPT_ENABLED=true
CHATGPT_MODEL=gpt-4o-mini

CATEGORIZE_LIMIT=40
EOF

# バッチを実行
BATCH_TYPE=categorize go run ./cmd/batch/...
```

### AWS Lambda（EventBridge）

EventBridgeから以下のJSONイベントを送信：

```json
{
  "type": "categorize",
  "limit": 5
}
```

## レートリミット対策

ChatGPT APIには1分あたり3回のリクエスト制限があるため、以下の対策を実装しています：

| タイミング | 待機時間 | 説明 |
|-----------|---------|------|
| 2冊処理後 | 1分 | レート制限回避 |
| 1冊目処理後 | 0.5秒 | 基本的な間隔 |
| 最後の書籍 | なし | 不要なため |

**処理フロー例（5冊の場合）:**
```
1冊目処理 → 0.5秒待機
2冊目処理 → 1分待機（ログ: "レート制限対策: 1分間待機します..."）
3冊目処理 → 0.5秒待機
4冊目処理 → 1分待機
5冊目処理 → 終了
```

## 実行結果サンプル

```
===========================================
  TeckBook Compass Categorize Batch
  開始時刻: 2024-12-19 14:00:00
  処理上限: 5 件
===========================================
カテゴライズバッチを開始します...
取得したカテゴリ数: 10
処理対象の書籍数: 5
推定トークン使用量: 約 3500 トークン（5冊 × 約700トークン/冊）
処理中 [1/5]: リーダブルコード (ISBN: 9784873115658)
ChatGPT response: {"category_id": 9} (tokens: prompt=720, completion=15, total=735)
カテゴリ設定成功: リーダブルコード -> beginner-cs (tokens: 735)
処理中 [2/5]: 実践ドメイン駆動設計 (ISBN: 9784798131610)
ChatGPT response: {"category_id": 6} (tokens: prompt=680, completion=15, total=695)
カテゴリ設定成功: 実践ドメイン駆動設計 -> backend (tokens: 695)
レート制限対策: 1分間待機します...
...
===========================================
  カテゴライズバッチ結果
===========================================
  処理した書籍数:   5
  カテゴライズ成功: 5
  エラー数:         0
  処理時間:         2m35s
===========================================
  終了時刻: 2024-12-19 14:02:35
===========================================
```

## Slack通知

### 開始メッセージ

```
🏷️ カテゴライズバッチ開始
処理対象: 5件
開始時刻: 2024-12-19 14:00:00
```

### 結果メッセージ（正常終了時）

```
✅ カテゴライズバッチ完了

処理結果
• 処理した書籍数: 5
• カテゴライズ成功: 5
• エラー数: 0
• 処理時間: 2m35s
終了時刻: 2024-12-19 14:02:35

🤖 ChatGPT トークン使用量
• プロンプトトークン: 3500
• 完了トークン: 75
• 合計トークン: 3575
```

### エラーメッセージ（レート制限時）

```
🚨 TeckBook Compass エラー

カテゴライズバッチエラー
レート制限で終了: chatgpt api rate limited
未処理: 4冊

使用済みトークン: 735 (prompt: 720, completion: 15)
```

## エラーハンドリング

| エラー種別 | 動作 | Slack通知 |
|-----------|------|----------|
| レート制限（429） | 即座に終了 | 未処理冊数と使用済みトークンを通知 |
| 認証エラー（401） | 即座に終了 | 使用済みトークンを通知 |
| タイムアウト | スキップして続行 | なし |
| JSONパースエラー | スキップして続行 | なし |
| DB保存エラー | スキップして続行 | なし |

## データベーススキーマ

### book_categories テーブル

```sql
CREATE TABLE book_categories (
    id BIGSERIAL PRIMARY KEY,
    book_id VARCHAR(20) NOT NULL,
    category_id VARCHAR(50) NOT NULL,
    score REAL DEFAULT 0,
    rank SMALLINT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE,
    UNIQUE(book_id, category_id)
);
```

### categories テーブル

```sql
CREATE TABLE categories (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    icon VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

## ファイル構成

```
teckbook-compass-backend/
├── cmd/batch/
│   ├── batch.go                              # バッチ共通処理（BatchTypeCategorize追加）
│   ├── handler.go                            # Lambda用ハンドラー
│   └── main.go                               # エントリーポイント
├── internal/
│   ├── infrastructure/
│   │   ├── config/
│   │   │   └── config.go                     # ChatGPTConfig追加
│   │   ├── database/postgres/
│   │   │   └── batch_repository.go           # GetBooksWithoutCategory等追加
│   │   └── external/
│   │       ├── chatgpt_client.go             # ChatGPT APIクライアント（新規）
│   │       └── slack_client.go               # カテゴライズ用Slack通知メソッド追加
│   ├── domain/repository/
│   │   └── batch_repository.go               # カテゴライズ用インターフェース追加
│   └── usecase/
│       └── categorize_batch_usecase.go       # カテゴライズバッチユースケース（新規）
└── docs/batch/
    └── categorize-batch.md                   # このドキュメント
```

## 注意事項

1. **APIキーの管理**: `CHATGPT_API_KEY`は機密情報のため、環境変数またはSecrets Managerで管理してください

2. **コスト管理**: ChatGPT APIは従量課金のため、`CATEGORIZE_LIMIT`で処理件数を適切に制限してください

3. **一度カテゴライズした書籍は再処理されません**: `book_categories`に既にエントリがある書籍はスキップされます

4. **処理時間の目安**: 5冊処理で約2〜3分（レート制限対策の待機時間を含む）

## トラブルシューティング

### ChatGPT APIが無効と表示される

```
ChatGPT API is disabled. Please set CHATGPT_ENABLED=true and CHATGPT_API_KEY in your environment.
```

**対策**: 環境変数を確認してください
```bash
export CHATGPT_ENABLED=true
export CHATGPT_API_KEY=sk-xxx
```

### レート制限エラーが頻発する

**対策**: 
- `CATEGORIZE_LIMIT`を小さくする（例: 2〜3件）
- OpenAIのプランをアップグレードする

### タイムアウトエラーが発生する

```
failed to send request: Post "https://api.openai.com/v1/chat/completions": context deadline exceeded
```

**対策**: 
- 現在のタイムアウトは120秒に設定されています
- ネットワーク状況を確認してください
