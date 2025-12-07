-- トリガーを削除
DROP TRIGGER IF EXISTS update_batch_statuses_updated_at ON batch_statuses;

-- テーブルを削除
DROP TABLE IF EXISTS batch_statuses;
