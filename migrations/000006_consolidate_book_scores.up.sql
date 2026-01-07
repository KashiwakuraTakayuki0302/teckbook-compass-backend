BEGIN;

-- 1. Create temporary table with aggregated data
CREATE TEMP TABLE temp_consolidated_scores AS
SELECT 
    book_id, 
    MAX(date) as date, 
    SUM(score) as score,
    SUM(article_count) as article_count,
    MIN(created_at) as created_at
FROM book_scores_daily
GROUP BY book_id;

-- 2. Truncate original table
TRUNCATE TABLE book_scores_daily;

-- 3. Drop old unique constraint
ALTER TABLE book_scores_daily DROP CONSTRAINT IF EXISTS book_scores_daily_book_id_date_key;

-- 4. Add new unique constraint
ALTER TABLE book_scores_daily ADD CONSTRAINT book_scores_daily_book_id_key UNIQUE (book_id);

-- 5. Insert consolidated data
INSERT INTO book_scores_daily (book_id, date, score, article_count, created_at)
SELECT book_id, date, score, article_count, created_at FROM temp_consolidated_scores;

-- 6. Drop temp table
DROP TABLE temp_consolidated_scores;

COMMIT;
