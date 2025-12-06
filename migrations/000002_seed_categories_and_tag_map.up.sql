-- カテゴリのシードデータ（10カテゴリ）
INSERT INTO categories (id, name, icon) VALUES
    ('ai-ml', 'AI / 機械学習', 'robot'),
    ('frontend', 'Web / フロントエンド', 'web'),
    ('mobile', 'モバイル / アプリ開発', 'mobile'),
    ('cloud', 'クラウド（AWS / GCP / Azure）', 'cloud'),
    ('infra-devops', 'インフラ / DevOps', 'devops'),
    ('backend', 'バックエンド / API / Webアーキ', 'backend'),
    ('database', 'データベース / データエンジニアリング', 'database'),
    ('security', 'セキュリティ', 'shield'),
    ('beginner-cs', 'プログラミング入門 / CS基礎', 'book'),
    ('pm-business', 'PM / プロダクト / ビジネス・キャリア', 'briefcase')
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    icon = EXCLUDED.icon;

-- タグ → カテゴリマッピングのシードデータ

-- AI / 機械学習
INSERT INTO tag_category_map (tag_name, category_id) VALUES
    ('ai', 'ai-ml'),
    ('machine-learning', 'ai-ml'),
    ('ml', 'ai-ml'),
    ('deep-learning', 'ai-ml'),
    ('neural-network', 'ai-ml'),
    ('pytorch', 'ai-ml'),
    ('tensorflow', 'ai-ml'),
    ('nlp', 'ai-ml'),
    ('llm', 'ai-ml')
ON CONFLICT (tag_name, category_id) DO NOTHING;

-- Web / フロントエンド
INSERT INTO tag_category_map (tag_name, category_id) VALUES
    ('javascript', 'frontend'),
    ('typescript', 'frontend'),
    ('react', 'frontend'),
    ('nextjs', 'frontend'),
    ('vue', 'frontend'),
    ('nuxt', 'frontend'),
    ('svelte', 'frontend'),
    ('css', 'frontend'),
    ('html', 'frontend')
ON CONFLICT (tag_name, category_id) DO NOTHING;

-- モバイル / アプリ開発
INSERT INTO tag_category_map (tag_name, category_id) VALUES
    ('android', 'mobile'),
    ('kotlin', 'mobile'),
    ('ios', 'mobile'),
    ('swift', 'mobile'),
    ('flutter', 'mobile'),
    ('reactnative', 'mobile')
ON CONFLICT (tag_name, category_id) DO NOTHING;

-- クラウド（AWS / GCP / Azure）
INSERT INTO tag_category_map (tag_name, category_id) VALUES
    ('aws', 'cloud'),
    ('ec2', 'cloud'),
    ('lambda', 'cloud'),
    ('gcp', 'cloud'),
    ('cloudrun', 'cloud'),
    ('azure', 'cloud')
ON CONFLICT (tag_name, category_id) DO NOTHING;

-- インフラ / DevOps
INSERT INTO tag_category_map (tag_name, category_id) VALUES
    ('docker', 'infra-devops'),
    ('kubernetes', 'infra-devops'),
    ('k8s', 'infra-devops'),
    ('terraform', 'infra-devops'),
    ('ansible', 'infra-devops'),
    ('linux', 'infra-devops'),
    ('devops', 'infra-devops'),
    ('cicd', 'infra-devops')
ON CONFLICT (tag_name, category_id) DO NOTHING;

-- バックエンド / API / Webアーキ
INSERT INTO tag_category_map (tag_name, category_id) VALUES
    ('go', 'backend'),
    ('golang', 'backend'),
    ('python', 'backend'),
    ('java', 'backend'),
    ('spring', 'backend'),
    ('nodejs', 'backend'),
    ('fastapi', 'backend'),
    ('django', 'backend'),
    ('restapi', 'backend')
ON CONFLICT (tag_name, category_id) DO NOTHING;

-- データベース / データエンジニアリング
INSERT INTO tag_category_map (tag_name, category_id) VALUES
    ('mysql', 'database'),
    ('postgresql', 'database'),
    ('sqlite', 'database'),
    ('redis', 'database'),
    ('bigquery', 'database'),
    ('data-engineering', 'database')
ON CONFLICT (tag_name, category_id) DO NOTHING;

-- セキュリティ
INSERT INTO tag_category_map (tag_name, category_id) VALUES
    ('security', 'security'),
    ('vulnerability', 'security'),
    ('ctf', 'security'),
    ('owasp', 'security')
ON CONFLICT (tag_name, category_id) DO NOTHING;

-- プログラミング入門 / CS基礎
INSERT INTO tag_category_map (tag_name, category_id) VALUES
    ('beginner', 'beginner-cs'),
    ('入門', 'beginner-cs'),
    ('新人教育', 'beginner-cs'),
    ('cs', 'beginner-cs'),
    ('algorithm', 'beginner-cs')
ON CONFLICT (tag_name, category_id) DO NOTHING;

-- PM / プロダクト / ビジネス・キャリア
INSERT INTO tag_category_map (tag_name, category_id) VALUES
    ('pm', 'pm-business'),
    ('product-management', 'pm-business'),
    ('startup', 'pm-business'),
    ('career', 'pm-business'),
    ('management', 'pm-business')
ON CONFLICT (tag_name, category_id) DO NOTHING;
