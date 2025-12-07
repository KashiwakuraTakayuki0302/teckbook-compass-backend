-- バッチ状態管理テーブル
CREATE TABLE IF NOT EXISTS batch_statuses (
    id VARCHAR(50) PRIMARY KEY,                    -- バッチ識別子（例: 'qiita_fetch'）
    last_fetched_at TIMESTAMP,                     -- 最新記事取得時の基準日時
    next_page INT NOT NULL DEFAULT 1,              -- 過去記事取得用の次ページ番号
    last_run_at TIMESTAMP,                         -- 最後にバッチを実行した日時
    last_new_fetch_at TIMESTAMP,                   -- 最後に最新記事取得を実行した日時
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 初期データを挿入
INSERT INTO batch_statuses (id, last_fetched_at, next_page, last_run_at, last_new_fetch_at)
VALUES ('qiita_fetch', NULL, 1, NULL, NULL)
ON CONFLICT (id) DO NOTHING;

-- updated_atを自動更新するトリガー
CREATE TRIGGER update_batch_statuses_updated_at
    BEFORE UPDATE ON batch_statuses
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
