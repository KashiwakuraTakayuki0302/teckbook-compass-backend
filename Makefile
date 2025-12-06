.PHONY: help run build test lint swagger-ui validate-api db-test db-migrate db-rollback build-batch

help: ## ヘルプを表示
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

run: ## サーバーを起動
	go run cmd/api/main.go

build: ## APIバイナリをビルド
	go build -o bin/api cmd/api/main.go

build-batch: ## バッチバイナリをビルド
	go build -o bin/batch cmd/batch/main.go

build-all: build build-batch ## 全てのバイナリをビルド

test: ## テストを実行
	go test -v ./...

test-coverage: ## カバレッジ付きでテストを実行
	go test -v -cover ./...

lint: ## コードをリント
	golangci-lint run

swagger-ui: ## Swagger UIを起動（Docker必要）
	@echo "Swagger UIを起動中... http://localhost:8081 でアクセスできます"
	docker run -p 8081:8080 -e SWAGGER_JSON=/api/openapi.yaml -v $(PWD)/api:/api swaggerapi/swagger-ui

validate-api: ## OpenAPI定義をバリデーション
	@command -v spectral >/dev/null 2>&1 || { echo "Spectralがインストールされていません。npm install -g @stoplight/spectral-cli でインストールしてください"; exit 1; }
	spectral lint api/openapi.yaml

# データベース関連コマンド
db-test: ## データベース接続をテスト
	go run cmd/batch/main.go -test-connection

db-migrate: ## データベースマイグレーションを実行
	go run cmd/batch/main.go -migrate-up

db-rollback: ## データベースマイグレーションをロールバック（1ステップ）
	go run cmd/batch/main.go -migrate-down

db-rollback-all: ## データベースマイグレーションを全てロールバック
	go run cmd/batch/main.go -migrate-down -migrate-steps=999

clean: ## ビルド成果物を削除
	rm -rf bin/
