.PHONY: help run build test lint swagger-ui validate-api

help: ## ヘルプを表示
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

run: ## サーバーを起動
	go run cmd/api/main.go

build: ## バイナリをビルド
	go build -o bin/api cmd/api/main.go

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

clean: ## ビルド成果物を削除
	rm -rf bin/
