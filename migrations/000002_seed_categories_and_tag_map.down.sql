-- タグ → カテゴリマッピングのシードデータを削除
DELETE FROM tag_category_map WHERE tag_name IN (
    -- AI / 機械学習
    'ai', 'machine-learning', 'ml', 'deep-learning', 'neural-network',
    'pytorch', 'tensorflow', 'nlp', 'llm',
    -- Web / フロントエンド
    'javascript', 'typescript', 'react', 'nextjs', 'vue', 'nuxt', 'svelte', 'css', 'html',
    -- モバイル / アプリ開発
    'android', 'kotlin', 'ios', 'swift', 'flutter', 'reactnative',
    -- クラウド
    'aws', 'ec2', 'lambda', 'gcp', 'cloudrun', 'azure',
    -- インフラ / DevOps
    'docker', 'kubernetes', 'k8s', 'terraform', 'ansible', 'linux', 'devops', 'cicd',
    -- バックエンド / API
    'go', 'golang', 'python', 'java', 'spring', 'nodejs', 'fastapi', 'django', 'restapi',
    -- データベース
    'mysql', 'postgresql', 'sqlite', 'redis', 'bigquery', 'data-engineering',
    -- セキュリティ
    'security', 'vulnerability', 'ctf', 'owasp',
    -- プログラミング入門 / CS基礎
    'beginner', '入門', '新人教育', 'cs', 'algorithm',
    -- PM / ビジネス / キャリア
    'pm', 'product-management', 'startup', 'career', 'management'
);

-- カテゴリのシードデータを削除
DELETE FROM categories WHERE id IN (
    'ai-ml',
    'frontend',
    'mobile',
    'cloud',
    'infra-devops',
    'backend',
    'database',
    'security',
    'beginner-cs',
    'pm-business'
);
