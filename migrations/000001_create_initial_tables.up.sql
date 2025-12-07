-- 1. categories（技術書カテゴリ 10個固定）
-- 他のテーブルから参照されるため先に作成
CREATE TABLE IF NOT EXISTS categories (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    icon VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 2. books（技術書のマスタ）
CREATE TABLE IF NOT EXISTS books (
    id VARCHAR(20) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    author VARCHAR(255),
    publisher VARCHAR(255),
    published_date DATE,
    price INT,
    thumbnail_url TEXT,
    amazon_url TEXT,
    rakuten_url TEXT,
    rakuten_average_rating REAL,
    rakuten_review_count INT,
    overview TEXT,
    latest_mentioned_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 3. articles（Qiita記事）
CREATE TABLE IF NOT EXISTS articles (
    id VARCHAR(50) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    url TEXT NOT NULL,
    likes INT DEFAULT 0,
    stocks INT DEFAULT 0,
    comments INT DEFAULT 0,
    published_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 4. article_books（記事 × 技術書 の対応）
CREATE TABLE IF NOT EXISTS article_books (
    id BIGSERIAL PRIMARY KEY,
    article_id VARCHAR(50) NOT NULL,
    book_id VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE CASCADE,
    FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE,
    UNIQUE(article_id, book_id)
);

-- 5. article_tags（Qiitaタグ）
CREATE TABLE IF NOT EXISTS article_tags (
    id BIGSERIAL PRIMARY KEY,
    article_id VARCHAR(50) NOT NULL,
    tag_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE CASCADE
);

-- 6. book_scores_daily（書籍スコア日次集計）
CREATE TABLE IF NOT EXISTS book_scores_daily (
    id BIGSERIAL PRIMARY KEY,
    book_id VARCHAR(20) NOT NULL,
    date DATE NOT NULL,
    score REAL NOT NULL DEFAULT 0,
    article_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE,
    UNIQUE(book_id, date)
);

-- 7. tag_category_map（Qiitaタグ → カテゴリ）
CREATE TABLE IF NOT EXISTS tag_category_map (
    id BIGSERIAL PRIMARY KEY,
    tag_name VARCHAR(100) NOT NULL,
    category_id VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE,
    UNIQUE(tag_name, category_id)
);

-- 8. book_categories（技術書 → カテゴリの結果）
CREATE TABLE IF NOT EXISTS book_categories (
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

-- 9. error_logs（エラーログ）
CREATE TABLE IF NOT EXISTS error_logs (
    id BIGSERIAL PRIMARY KEY,
    batch_process VARCHAR(50) NOT NULL,
    error_type VARCHAR(50) NOT NULL,
    level VARCHAR(20) NOT NULL,
    api_name VARCHAR(50),
    endpoint TEXT,
    status_code INT,
    request_payload JSONB,
    response_body JSONB,
    related_id VARCHAR(100),
    message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- インデックスの作成
CREATE INDEX IF NOT EXISTS idx_books_published_date ON books(published_date DESC);
CREATE INDEX IF NOT EXISTS idx_books_latest_mentioned_at ON books(latest_mentioned_at DESC);
CREATE INDEX IF NOT EXISTS idx_articles_published_at ON articles(published_at DESC);
CREATE INDEX IF NOT EXISTS idx_article_books_article_id ON article_books(article_id);
CREATE INDEX IF NOT EXISTS idx_article_books_book_id ON article_books(book_id);
CREATE INDEX IF NOT EXISTS idx_article_tags_article_id ON article_tags(article_id);
CREATE INDEX IF NOT EXISTS idx_article_tags_tag_name ON article_tags(tag_name);
CREATE INDEX IF NOT EXISTS idx_book_scores_daily_book_id ON book_scores_daily(book_id);
CREATE INDEX IF NOT EXISTS idx_book_scores_daily_date ON book_scores_daily(date DESC);
CREATE INDEX IF NOT EXISTS idx_tag_category_map_tag_name ON tag_category_map(tag_name);
CREATE INDEX IF NOT EXISTS idx_tag_category_map_category_id ON tag_category_map(category_id);
CREATE INDEX IF NOT EXISTS idx_book_categories_book_id ON book_categories(book_id);
CREATE INDEX IF NOT EXISTS idx_book_categories_category_id ON book_categories(category_id);
CREATE INDEX IF NOT EXISTS idx_book_categories_rank ON book_categories(rank);
CREATE INDEX IF NOT EXISTS idx_error_logs_batch_process ON error_logs(batch_process);
CREATE INDEX IF NOT EXISTS idx_error_logs_level ON error_logs(level);
CREATE INDEX IF NOT EXISTS idx_error_logs_created_at ON error_logs(created_at DESC);

-- updated_atを自動更新するトリガー関数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 各テーブルにトリガーを設定
CREATE TRIGGER update_categories_updated_at
    BEFORE UPDATE ON categories
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_books_updated_at
    BEFORE UPDATE ON books
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_articles_updated_at
    BEFORE UPDATE ON articles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

