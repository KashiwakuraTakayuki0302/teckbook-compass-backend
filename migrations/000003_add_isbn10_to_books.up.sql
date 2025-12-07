-- booksテーブルにISBN-10カラムを追加
ALTER TABLE books ADD COLUMN IF NOT EXISTS isbn10 VARCHAR(10);

-- ISBN-10用のインデックスを追加
CREATE INDEX IF NOT EXISTS idx_books_isbn10 ON books(isbn10);

-- 既存のISBN-13からISBN-10を生成して更新（978で始まるもののみ）
UPDATE books
SET isbn10 = (
    CASE
        WHEN id LIKE '978%' AND LENGTH(id) = 13 THEN
            -- ISBN-13からISBN-10への変換（チェックディジット再計算）
            SUBSTRING(id FROM 4 FOR 9) || (
                CASE
                    WHEN (
                        (CAST(SUBSTRING(id FROM 4 FOR 1) AS INTEGER) * 10 +
                         CAST(SUBSTRING(id FROM 5 FOR 1) AS INTEGER) * 9 +
                         CAST(SUBSTRING(id FROM 6 FOR 1) AS INTEGER) * 8 +
                         CAST(SUBSTRING(id FROM 7 FOR 1) AS INTEGER) * 7 +
                         CAST(SUBSTRING(id FROM 8 FOR 1) AS INTEGER) * 6 +
                         CAST(SUBSTRING(id FROM 9 FOR 1) AS INTEGER) * 5 +
                         CAST(SUBSTRING(id FROM 10 FOR 1) AS INTEGER) * 4 +
                         CAST(SUBSTRING(id FROM 11 FOR 1) AS INTEGER) * 3 +
                         CAST(SUBSTRING(id FROM 12 FOR 1) AS INTEGER) * 2) % 11
                    ) = 0 THEN '0'
                    WHEN (
                        11 - (CAST(SUBSTRING(id FROM 4 FOR 1) AS INTEGER) * 10 +
                              CAST(SUBSTRING(id FROM 5 FOR 1) AS INTEGER) * 9 +
                              CAST(SUBSTRING(id FROM 6 FOR 1) AS INTEGER) * 8 +
                              CAST(SUBSTRING(id FROM 7 FOR 1) AS INTEGER) * 7 +
                              CAST(SUBSTRING(id FROM 8 FOR 1) AS INTEGER) * 6 +
                              CAST(SUBSTRING(id FROM 9 FOR 1) AS INTEGER) * 5 +
                              CAST(SUBSTRING(id FROM 10 FOR 1) AS INTEGER) * 4 +
                              CAST(SUBSTRING(id FROM 11 FOR 1) AS INTEGER) * 3 +
                              CAST(SUBSTRING(id FROM 12 FOR 1) AS INTEGER) * 2) % 11
                    ) = 10 THEN 'X'
                    ELSE CAST(
                        11 - (CAST(SUBSTRING(id FROM 4 FOR 1) AS INTEGER) * 10 +
                              CAST(SUBSTRING(id FROM 5 FOR 1) AS INTEGER) * 9 +
                              CAST(SUBSTRING(id FROM 6 FOR 1) AS INTEGER) * 8 +
                              CAST(SUBSTRING(id FROM 7 FOR 1) AS INTEGER) * 7 +
                              CAST(SUBSTRING(id FROM 8 FOR 1) AS INTEGER) * 6 +
                              CAST(SUBSTRING(id FROM 9 FOR 1) AS INTEGER) * 5 +
                              CAST(SUBSTRING(id FROM 10 FOR 1) AS INTEGER) * 4 +
                              CAST(SUBSTRING(id FROM 11 FOR 1) AS INTEGER) * 3 +
                              CAST(SUBSTRING(id FROM 12 FOR 1) AS INTEGER) * 2) % 11
                        AS VARCHAR)
                END
            )
        ELSE NULL
    END
)
WHERE isbn10 IS NULL;
