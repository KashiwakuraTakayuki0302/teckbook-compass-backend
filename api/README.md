# API ドキュメント

このディレクトリにはTechBook Compass APIのOpenAPI定義が含まれています。

## ファイル

- `openapi.yaml` - OpenAPI 3.0仕様に基づくAPI定義

## 使い方

### 1. Swagger UIでドキュメントを表示

#### オンラインエディタで確認
[Swagger Editor](https://editor.swagger.io/)にアクセスして、`openapi.yaml`の内容を貼り付けます。

#### ローカルでSwagger UIを起動

```bash
# Dockerを使用する場合
docker run -p 8081:8080 -e SWAGGER_JSON=/api/openapi.yaml -v $(pwd)/api:/api swaggerapi/swagger-ui

# ブラウザで http://localhost:8081 にアクセス
```

### 2. OpenAPI定義のバリデーション

```bash
# openapi-generatorをインストール
brew install openapi-generator

# バリデーション実行
openapi-generator validate -i api/openapi.yaml
```

または、npmを使用：

```bash
# Spectralをインストール（OpenAPIリンター）
npm install -g @stoplight/spectral-cli

# バリデーション実行
spectral lint api/openapi.yaml
```

### 3. TypeScriptクライアントコード生成（フロントエンド用）

```bash
# TypeScript Axiosクライアントを生成
openapi-generator generate \
  -i api/openapi.yaml \
  -g typescript-axios \
  -o ./generated/client
```

### 4. Goのサーバースタブ生成

```bash
# oapi-codegenをインストール
go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest

# サーバーコード生成
oapi-codegen -package api -generate types,server api/openapi.yaml > internal/generated/api.go
```

### 5. Postmanコレクション生成

```bash
# OpenAPIからPostmanコレクションを生成
openapi-generator generate \
  -i api/openapi.yaml \
  -g postman-collection \
  -o ./generated/postman
```

## API仕様の更新

1. `openapi.yaml`を編集
2. バリデーションを実行して構文エラーがないか確認
3. 必要に応じてクライアントコードを再生成
4. 変更をコミット

## 参考リンク

- [OpenAPI Specification](https://swagger.io/specification/)
- [Swagger Editor](https://editor.swagger.io/)
- [OpenAPI Generator](https://openapi-generator.tech/)
