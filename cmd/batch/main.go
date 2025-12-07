package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"teckbook-compass-backend/internal/infrastructure/config"
	"teckbook-compass-backend/internal/infrastructure/database/postgres"
	"teckbook-compass-backend/internal/infrastructure/external"
	"teckbook-compass-backend/internal/usecase"
)

func main() {
	// コマンドライン引数の解析
	migrateUp := flag.Bool("migrate-up", false, "Run database migrations up")
	migrateDown := flag.Bool("migrate-down", false, "Rollback database migrations (1 step)")
	migrateSteps := flag.Int("migrate-steps", 1, "Number of migration steps to rollback")
	testConnection := flag.Bool("test-connection", false, "Test database connection")
	runBatch := flag.Bool("run-batch", false, "Run the batch process to fetch and process articles")
	dryRun := flag.Bool("dry-run", false, "Run batch in dry-run mode (no DB writes)")
	fetchNew := flag.Bool("fetch-new", false, "Force fetch new articles mode (use with -run-batch)")
	fetchHistorical := flag.Bool("fetch-historical", false, "Force fetch historical articles mode (use with -run-batch)")

	flag.Parse()

	// 設定の読み込み
	cfg := config.NewConfig()

	// マイグレーションパスの取得（DB接続前に実行してリソースリークを防ぐ）
	migrationsPath, err := getMigrationsPath()
	if err != nil {
		log.Fatalf("Failed to get migrations path: %v", err)
	}

	// データベース接続
	db, err := postgres.NewConnection(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// コマンドの実行
	switch {
	case *testConnection:
		log.Println("Database connection test successful!")

	case *migrateUp:
		log.Println("Running database migrations...")
		if err := db.RunMigrations(migrationsPath); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}

	case *migrateDown:
		log.Printf("Rolling back database migrations (%d steps)...", *migrateSteps)
		if err := db.RollbackMigrations(migrationsPath, *migrateSteps); err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}

	case *runBatch:
		log.Println("Starting batch process...")
		if *dryRun {
			log.Println("Running in dry-run mode")
		}

		// 取得モードを決定
		var fetchMode *usecase.FetchModeOption
		if *fetchNew && *fetchHistorical {
			log.Fatal("Cannot specify both -fetch-new and -fetch-historical")
		} else if *fetchNew {
			mode := usecase.FetchModeOptionNew
			fetchMode = &mode
			log.Println("Forced mode: 最新記事取得")
		} else if *fetchHistorical {
			mode := usecase.FetchModeOptionHistorical
			fetchMode = &mode
			log.Println("Forced mode: 過去記事取得")
		}

		if err := runBatchProcess(cfg, db, *dryRun, fetchMode); err != nil {
			log.Fatalf("Batch process failed: %v", err)
		}

	default:
		flag.Usage()
		fmt.Println("\nAvailable commands:")
		fmt.Println("  -migrate-up        Run database migrations")
		fmt.Println("  -migrate-down      Rollback database migrations")
		fmt.Println("  -test-connection   Test database connection")
		fmt.Println("  -run-batch         Run the batch process")
		fmt.Println("  -fetch-new         Force fetch new articles mode (use with -run-batch)")
		fmt.Println("  -fetch-historical  Force fetch historical articles mode (use with -run-batch)")
		fmt.Println("  -dry-run           Run batch in dry-run mode (use with -run-batch)")
		os.Exit(1)
	}
}

// runBatchProcess バッチ処理を実行
func runBatchProcess(cfg *config.Config, db *postgres.DB, dryRun bool, fetchMode *usecase.FetchModeOption) error {
	log.Println("===========================================")
	log.Println("  TeckBook Compass Daily Batch")
	log.Printf("  開始時刻: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	log.Println("===========================================")

	// コンテキストを作成（タイムアウト付き）
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
	defer cancel()

	// リポジトリを初期化
	batchRepo := postgres.NewBatchRepository(db.DB)

	// 外部APIクライアントを初期化
	qiitaClient := external.NewQiitaClient(cfg.Qiita)
	rakutenClient := external.NewRakutenClient(cfg.Rakuten)
	slackClient := external.NewSlackClient(cfg.Slack)

	if slackClient.IsEnabled() {
		log.Println("Slack通知: 有効")
	} else {
		log.Println("Slack通知: 無効")
	}

	// ユースケースを初期化
	batchUsecase := usecase.NewBatchUsecase(batchRepo, qiitaClient, rakutenClient, slackClient)

	// バッチ処理を実行
	result, err := batchUsecase.Run(ctx, fetchMode)
	if err != nil {
		return fmt.Errorf("batch process error: %w", err)
	}

	// 結果を出力
	log.Println("===========================================")
	log.Println("  バッチ処理結果")
	log.Println("===========================================")
	log.Printf("  取得モード:       %s\n", result.FetchMode)
	log.Printf("  処理した記事数:   %d\n", result.ProcessedArticles)
	log.Printf("  新規記事数:       %d\n", result.NewArticles)
	log.Printf("  処理した書籍数:   %d\n", result.ProcessedBooks)
	log.Printf("  エラー数:         %d\n", result.Errors)
	if result.NextPage > 0 {
		log.Printf("  次回開始ページ:   %d\n", result.NextPage)
	}
	log.Printf("  処理時間:         %v\n", result.EndTime.Sub(result.StartTime))
	log.Println("===========================================")
	log.Printf("  終了時刻: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	log.Println("===========================================")

	return nil
}

// getMigrationsPath マイグレーションファイルのパスを取得
func getMigrationsPath() (string, error) {
	// 環境変数からパスを取得
	if path := os.Getenv("MIGRATIONS_PATH"); path != "" {
		return path, nil
	}

	// デフォルトは実行ディレクトリからの相対パス
	execPath, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	return filepath.Join(execPath, "migrations"), nil
}
