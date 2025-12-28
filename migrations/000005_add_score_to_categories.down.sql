-- idx_categories_scoreインデックスを削除
DROP INDEX IF EXISTS idx_categories_score;

-- categoriesテーブルからscoreカラムを削除
ALTER TABLE categories DROP COLUMN IF EXISTS score;

