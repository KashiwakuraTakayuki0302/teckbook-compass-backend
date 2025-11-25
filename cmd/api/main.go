package main

import (
	"fmt"
	"log"
	"teckbook-compass-backend/internal/infrastructure/config"
	"teckbook-compass-backend/internal/infrastructure/database/mock"
	"teckbook-compass-backend/internal/interface/handler"
	"teckbook-compass-backend/internal/interface/router"
	"teckbook-compass-backend/internal/usecase"
)

func main() {
	// 設定の初期化
	cfg := config.NewConfig()

	// リポジトリの初期化（モック）
	categoryRepo := mock.NewCategoryRepositoryMock()
	bookRepo := mock.NewBookRepositoryMock()

	// ユースケースの初期化
	categoryUsecase := usecase.NewCategoryUsecase(categoryRepo, bookRepo)

	// ハンドラの初期化
	categoryHandler := handler.NewCategoryHandler(categoryUsecase)

	// ルーターのセットアップ
	r := router.SetupRouter(categoryHandler)

	// サーバー起動
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
