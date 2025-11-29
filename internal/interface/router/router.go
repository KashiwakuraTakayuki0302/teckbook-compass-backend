package router

import (
	"teckbook-compass-backend/internal/interface/handler"

	"github.com/gin-gonic/gin"
)

// SetupRouter ルーターをセットアップ
func SetupRouter(categoryHandler *handler.CategoryHandler, rankingHandler *handler.RankingHandler) *gin.Engine {
	r := gin.Default()

	// CORSミドルウェア
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// ヘルスチェック
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// カテゴリエンドポイント
	r.GET("/categories/with-books", categoryHandler.GetCategoriesWithBooks)

	// ランキングエンドポイント
	r.GET("/rankings", rankingHandler.GetRankings)

	// OpenAPI仕様ファイルの提供
	r.StaticFile("/api/openapi.yaml", "./api/openapi.yaml")

	return r
}
