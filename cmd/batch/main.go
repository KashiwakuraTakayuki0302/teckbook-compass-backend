package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"teckbook-compass-backend/internal/infrastructure/config"
	"teckbook-compass-backend/internal/infrastructure/database/postgres"
)

func main() {
	// コマンドライン引数の解析
	migrateUp := flag.Bool("migrate-up", false, "Run database migrations up")
	migrateDown := flag.Bool("migrate-down", false, "Rollback database migrations (1 step)")
	migrateSteps := flag.Int("migrate-steps", 1, "Number of migration steps to rollback")
	testConnection := flag.Bool("test-connection", false, "Test database connection")

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

	default:
		flag.Usage()
		os.Exit(1)
	}
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
