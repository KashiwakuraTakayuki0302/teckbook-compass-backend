-- ISBN-10インデックスを削除
DROP INDEX IF EXISTS idx_books_isbn10;

-- ISBN-10カラムを削除
ALTER TABLE books DROP COLUMN IF EXISTS isbn10;
