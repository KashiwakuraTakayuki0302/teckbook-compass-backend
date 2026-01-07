BEGIN;

ALTER TABLE book_scores_daily DROP CONSTRAINT IF EXISTS book_scores_daily_book_id_key;
ALTER TABLE book_scores_daily ADD CONSTRAINT book_scores_daily_book_id_date_key UNIQUE (book_id, date);

COMMIT;
