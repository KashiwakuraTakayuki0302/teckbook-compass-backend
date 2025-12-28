-- categoriesテーブルにscoreカラムを追加
-- book_categoriesに紐づく書籍のbook_scores_dailyスコア合計値を保存

ALTER TABLE categories ADD COLUMN IF NOT EXISTS score REAL DEFAULT 0;

-- スコアによるソート用インデックス
CREATE INDEX IF NOT EXISTS idx_categories_score ON categories(score DESC);

