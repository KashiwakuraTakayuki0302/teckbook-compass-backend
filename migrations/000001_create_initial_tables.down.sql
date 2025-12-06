-- トリガーの削除
DROP TRIGGER IF EXISTS update_articles_updated_at ON articles;
DROP TRIGGER IF EXISTS update_books_updated_at ON books;
DROP TRIGGER IF EXISTS update_categories_updated_at ON categories;

-- トリガー関数の削除
DROP FUNCTION IF EXISTS update_updated_at_column();

-- インデックスの削除
DROP INDEX IF EXISTS idx_error_logs_created_at;
DROP INDEX IF EXISTS idx_error_logs_level;
DROP INDEX IF EXISTS idx_error_logs_batch_process;
DROP INDEX IF EXISTS idx_book_categories_rank;
DROP INDEX IF EXISTS idx_book_categories_category_id;
DROP INDEX IF EXISTS idx_book_categories_book_id;
DROP INDEX IF EXISTS idx_tag_category_map_category_id;
DROP INDEX IF EXISTS idx_tag_category_map_tag_name;
DROP INDEX IF EXISTS idx_book_scores_daily_date;
DROP INDEX IF EXISTS idx_book_scores_daily_book_id;
DROP INDEX IF EXISTS idx_article_tags_tag_name;
DROP INDEX IF EXISTS idx_article_tags_article_id;
DROP INDEX IF EXISTS idx_article_books_book_id;
DROP INDEX IF EXISTS idx_article_books_article_id;
DROP INDEX IF EXISTS idx_articles_published_at;
DROP INDEX IF EXISTS idx_books_latest_mentioned_at;
DROP INDEX IF EXISTS idx_books_published_date;

-- テーブルの削除（依存関係の逆順）
DROP TABLE IF EXISTS error_logs;
DROP TABLE IF EXISTS book_categories;
DROP TABLE IF EXISTS tag_category_map;
DROP TABLE IF EXISTS book_scores_daily;
DROP TABLE IF EXISTS article_tags;
DROP TABLE IF EXISTS article_books;
DROP TABLE IF EXISTS articles;
DROP TABLE IF EXISTS books;
DROP TABLE IF EXISTS categories;

