//go:build !lambda

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"teckbook-compass-backend/internal/usecase"
)

// ============================================================================
// CLI フラグ定義
// ============================================================================

type cliFlags struct {
	migrateUp        bool
	migrateDown      bool
	migrateSteps     int
	testConnection   bool
	runBatch         bool
	fetchNew         bool
	fetchHistorical  bool
	runAmazonBatch   bool
	amazonBatchLimit int
}

func parseFlags() *cliFlags {
	f := &cliFlags{}

	flag.BoolVar(&f.migrateUp, "migrate-up", false, "Run database migrations up")
	flag.BoolVar(&f.migrateDown, "migrate-down", false, "Rollback database migrations (1 step)")
	flag.IntVar(&f.migrateSteps, "migrate-steps", 1, "Number of migration steps to rollback")
	flag.BoolVar(&f.testConnection, "test-connection", false, "Test database connection")
	flag.BoolVar(&f.runBatch, "run-batch", false, "Run the batch process to fetch and process articles")
	flag.BoolVar(&f.fetchNew, "fetch-new", false, "Force fetch new articles mode (use with -run-batch)")
	flag.BoolVar(&f.fetchHistorical, "fetch-historical", false, "Force fetch historical articles mode (use with -run-batch)")
	flag.BoolVar(&f.runAmazonBatch, "run-amazon-batch", false, "Run Amazon URL fetch batch")
	flag.IntVar(&f.amazonBatchLimit, "amazon-limit", 50, "Number of books to process in Amazon batch (default: 50)")

	flag.Parse()
	return f
}

// ============================================================================
// メイン処理
// ============================================================================

func main() {
	// 環境変数 BATCH_TYPE が設定されている場合はそれを使用
	if os.Getenv("BATCH_TYPE") != "" {
		runBatchByEnvVar()
		return
	}

	// コマンドライン引数の解析
	flags := parseFlags()

	// コマンドの実行
	switch {
	case flags.testConnection:
		runTestConnection()

	case flags.migrateUp:
		runMigrateUp()

	case flags.migrateDown:
		runMigrateDown(flags.migrateSteps)

	case flags.runBatch:
		runArticleBatch(flags)

	case flags.runAmazonBatch:
		runAmazonBatch(flags.amazonBatchLimit)

	default:
		printUsage()
	}
}

// ============================================================================
// コマンド実行関数
// ============================================================================

// runTestConnection データベース接続テスト
func runTestConnection() {
	app, err := NewApp()
	if err != nil {
		log.Fatalf("初期化失敗: %v", err)
	}
	defer app.Close()

	log.Println("Database connection test successful!")
}

// runMigrateUp マイグレーション実行
func runMigrateUp() {
	migrationsPath, err := getMigrationsPath()
	if err != nil {
		log.Fatalf("Failed to get migrations path: %v", err)
	}

	app, err := NewApp()
	if err != nil {
		log.Fatalf("初期化失敗: %v", err)
	}
	defer app.Close()

	log.Println("Running database migrations...")
	if err := app.DB.RunMigrations(migrationsPath); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
}

// runMigrateDown マイグレーションロールバック
func runMigrateDown(steps int) {
	migrationsPath, err := getMigrationsPath()
	if err != nil {
		log.Fatalf("Failed to get migrations path: %v", err)
	}

	app, err := NewApp()
	if err != nil {
		log.Fatalf("初期化失敗: %v", err)
	}
	defer app.Close()

	log.Printf("Rolling back database migrations (%d steps)...", steps)
	if err := app.DB.RollbackMigrations(migrationsPath, steps); err != nil {
		log.Fatalf("Rollback failed: %v", err)
	}
}

// runArticleBatch 記事取得バッチ実行
func runArticleBatch(flags *cliFlags) {
	// 取得モードのバリデーション
	if flags.fetchNew && flags.fetchHistorical {
		log.Fatal("Cannot specify both -fetch-new and -fetch-historical")
	}

	app, err := NewApp()
	if err != nil {
		log.Fatalf("初期化失敗: %v", err)
	}
	defer app.Close()

	// 排他ロック取得
	if err := app.AcquireLock(false); err != nil {
		log.Fatalf("ロック取得失敗: %v", err)
	}

	// 取得モードを決定
	mode := determineFetchMode(flags)

	// バッチ実行
	result := app.ExecuteBatch(BatchParams{
		Type: BatchTypeArticle,
		Mode: mode,
	})

	if !result.Success {
		log.Fatalf("Batch process failed: %s", result.Message)
	}
}

// runAmazonBatch Amazon URL取得バッチ実行
func runAmazonBatch(limit int) {
	app, err := NewApp()
	if err != nil {
		log.Fatalf("初期化失敗: %v", err)
	}
	defer app.Close()

	// 排他ロック取得
	if err := app.AcquireLock(false); err != nil {
		log.Fatalf("ロック取得失敗: %v", err)
	}

	// バッチ実行
	result := app.ExecuteBatch(BatchParams{
		Type:  BatchTypeAmazon,
		Limit: limit,
	})

	if !result.Success {
		log.Fatalf("Amazon batch process failed: %s", result.Message)
	}
}

// runBatchByEnvVar 環境変数からバッチを実行
func runBatchByEnvVar() {
	params := NewBatchParamsFromEnv()
	log.Printf("環境変数BATCH_TYPE=%sでバッチを起動します", params.Type)

	app, err := NewApp()
	if err != nil {
		log.Fatalf("初期化失敗: %v", err)
	}
	defer app.Close()

	// 排他ロック取得
	if err := app.AcquireLock(false); err != nil {
		log.Fatalf("ロック取得失敗: %v", err)
	}

	// バッチ実行
	result := app.ExecuteBatch(params)

	if !result.Success {
		log.Fatalf("Batch process failed: %s", result.Message)
	}
}

// ============================================================================
// ヘルパー関数
// ============================================================================

// determineFetchMode フラグから取得モードを決定
func determineFetchMode(flags *cliFlags) string {
	switch {
	case flags.fetchNew:
		log.Println("Forced mode: 最新記事取得")
		return "new"
	case flags.fetchHistorical:
		log.Println("Forced mode: 過去記事取得")
		return "historical"
	default:
		return "" // 自動判定
	}
}

// printUsage ヘルプを表示
func printUsage() {
	flag.Usage()
	fmt.Println("\nAvailable commands:")
	fmt.Println("  -migrate-up        Run database migrations")
	fmt.Println("  -migrate-down      Rollback database migrations")
	fmt.Println("  -test-connection   Test database connection")
	fmt.Println("  -run-batch         Run the batch process")
	fmt.Println("  -fetch-new         Force fetch new articles mode (use with -run-batch)")
	fmt.Println("  -fetch-historical  Force fetch historical articles mode (use with -run-batch)")
	fmt.Println("  -run-amazon-batch  Run Amazon URL fetch batch")
	fmt.Println("  -amazon-limit      Number of books to process in Amazon batch (default: 50)")
	fmt.Println("\nEnvironment variables:")
	fmt.Println("  BATCH_TYPE=article|amazon  Run batch directly without flags")
	fmt.Println("  FETCH_MODE=new|historical  Fetch mode for article batch")
	fmt.Println("  AMAZON_LIMIT=50            Limit for amazon batch")
	os.Exit(1)
}

// ============================================================================
// ログ初期化
// ============================================================================

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// ============================================================================
// 下位互換性のための公開関数（batch.go から移動した関数の互換性維持）
// ============================================================================

// getFetchModeOption CLIフラグからFetchModeOptionを生成（テスト用途など）
func getFetchModeOption(fetchNew, fetchHistorical bool) *usecase.FetchModeOption {
	switch {
	case fetchNew:
		mode := usecase.FetchModeOptionNew
		return &mode
	case fetchHistorical:
		mode := usecase.FetchModeOptionHistorical
		return &mode
	default:
		return nil
	}
}
