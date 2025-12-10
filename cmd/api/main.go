package main

import (
	"fmt"
	"log"
	"teckbook-compass-backend/internal/infrastructure/config"
	"teckbook-compass-backend/internal/infrastructure/database/mock"
	"teckbook-compass-backend/internal/infrastructure/database/postgres"
	"teckbook-compass-backend/internal/infrastructure/secrets"
	"teckbook-compass-backend/internal/interface/handler"
	"teckbook-compass-backend/internal/interface/router"
	"teckbook-compass-backend/internal/usecase"
)

func main() {
	// 設定の初期化
	cfg := config.NewConfig()

	// Secrets Managerからusername/passwordを取得
	if err := secrets.LoadDatabaseCredentials(cfg); err != nil {
		log.Printf("警告: Secrets Managerからの認証情報取得に失敗しました（環境変数を使用）: %v", err)
		// エラーが発生しても環境変数の値で続行
	}

	// データベース接続
	db, err := postgres.NewConnection(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// リポジトリの初期化
	categoryRepo := mock.NewCategoryRepositoryMock() // カテゴリはまだモック
	bookRepo := postgres.NewBookRepository(db.DB)    // 書籍はPostgreSQL

	// ユースケースの初期化
	categoryUsecase := usecase.NewCategoryUsecase(categoryRepo, bookRepo)
	rankingUsecase := usecase.NewRankingUsecase(bookRepo)
	bookDetailUsecase := usecase.NewBookDetailUsecase(bookRepo)

	// ハンドラの初期化
	categoryHandler := handler.NewCategoryHandler(categoryUsecase)
	rankingHandler := handler.NewRankingHandler(rankingUsecase)
	bookDetailHandler := handler.NewBookDetailHandler(bookDetailUsecase)

	// ルーターのセットアップ
	r := router.SetupRouter(categoryHandler, rankingHandler, bookDetailHandler)

	// サーバー起動
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
