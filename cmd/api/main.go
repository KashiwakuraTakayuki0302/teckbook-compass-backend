package main

import (
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"

	"teckbook-compass-backend/internal/infrastructure/config"
	"teckbook-compass-backend/internal/infrastructure/database/mock"
	"teckbook-compass-backend/internal/infrastructure/database/postgres"
	"teckbook-compass-backend/internal/infrastructure/secrets"
	"teckbook-compass-backend/internal/interface/handler"
	"teckbook-compass-backend/internal/interface/router"
	"teckbook-compass-backend/internal/usecase"
)

var (
	ginLambda   *ginadapter.GinLambdaV2
	routerEngine *gin.Engine
	appConfig   *config.Config
)

// 🔹 cold start 時に1回だけ実行される
func init() {
	// 設定の初期化
	appConfig = config.NewConfig()
	cfg := appConfig

	// Secrets Manager
	if err := secrets.LoadDatabaseCredentials(cfg); err != nil {
		log.Printf("Secrets Manager warning: %v", err)
	}

	// DB接続（使い回される）
	db, err := postgres.NewConnection(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect DB: %v", err)
	}

	// Repository
	categoryRepo := mock.NewCategoryRepositoryMock()
	bookRepo := postgres.NewBookRepository(db.DB)

	// Usecase
	categoryUsecase := usecase.NewCategoryUsecase(categoryRepo, bookRepo)
	rankingUsecase := usecase.NewRankingUsecase(bookRepo)
	bookDetailUsecase := usecase.NewBookDetailUsecase(bookRepo)

	// Handler
	categoryHandler := handler.NewCategoryHandler(categoryUsecase)
	rankingHandler := handler.NewRankingHandler(rankingUsecase)
	bookDetailHandler := handler.NewBookDetailHandler(bookDetailUsecase)

	// Router
	routerEngine = router.SetupRouter(categoryHandler, rankingHandler, bookDetailHandler)

	// Lambda Adapter (API Gateway v2 HTTP API用)
	ginLambda = ginadapter.NewV2(routerEngine)
}

// 🔹 API Gateway v2 (HTTP API) → Lambda → Gin
func lambdaHandler(req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return ginLambda.Proxy(req)
}

func main() {
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		lambda.Start(lambdaHandler)
	} else {
		log.Printf("Starting server on port %s", appConfig.ServerPort)
		if err := routerEngine.Run(":" + appConfig.ServerPort); err != nil {
			log.Fatalf("Failed to run server: %v", err)
		}
	}
}
