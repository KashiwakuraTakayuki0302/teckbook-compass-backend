package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"teckbook-compass-backend/internal/infrastructure/config"
	"teckbook-compass-backend/internal/infrastructure/database/postgres"
	"teckbook-compass-backend/internal/infrastructure/external"
	"teckbook-compass-backend/internal/infrastructure/secrets"
	"teckbook-compass-backend/internal/usecase"
)

const (
	lockFilePath = "/tmp/teckbook-compass-batch.lock"
)

// ============================================================================
// 型定義
// ============================================================================

// BatchType バッチの種類
type BatchType string

const (
	BatchTypeArticle BatchType = "article" // 記事取得バッチ
	BatchTypeAmazon  BatchType = "amazon"  // Amazon URL取得バッチ
)

// IsValid バッチタイプが有効かどうかを判定
func (b BatchType) IsValid() bool {
	return b == BatchTypeArticle || b == BatchTypeAmazon
}

// String バッチタイプの日本語名を返す
func (b BatchType) String() string {
	switch b {
	case BatchTypeArticle:
		return "記事取得バッチ"
	case BatchTypeAmazon:
		return "Amazon URL取得バッチ"
	default:
		return string(b)
	}
}

// BatchParams バッチ実行パラメータ
type BatchParams struct {
	Type  BatchType // バッチの種類
	Mode  string    // 取得モード ("new", "historical", "auto"/空)
	Limit int       // 処理上限（amazonバッチ用）
}

// NewBatchParamsFromEnv 環境変数からBatchParamsを生成
func NewBatchParamsFromEnv() BatchParams {
	return BatchParams{
		Type:  BatchType(os.Getenv("BATCH_TYPE")),
		Mode:  os.Getenv("FETCH_MODE"),
		Limit: getAmazonLimitFromEnv(),
	}
}

// BatchResult バッチ実行結果
type BatchResult struct {
	Success bool
	Message string
}

// ============================================================================
// アプリケーション構造体（共通初期化）
// ============================================================================

// App バッチアプリケーションの共通構造体
type App struct {
	Config *config.Config
	DB     *postgres.DB
	lock   *BatchLock
}

// NewApp 新しいAppを作成（設定読み込み・DB接続）
func NewApp() (*App, error) {
	cfg := config.NewConfig()

	// Secrets Managerからusername/passwordを取得
	if err := secrets.LoadDatabaseCredentials(cfg); err != nil {
		log.Printf("警告: Secrets Managerからの認証情報取得に失敗しました（環境変数を使用）: %v", err)
		// エラーが発生しても環境変数の値で続行
	}

	db, err := postgres.NewConnection(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("データベース接続失敗: %w", err)
	}

	return &App{
		Config: cfg,
		DB:     db,
	}, nil
}

// Close リソースを解放
func (a *App) Close() {
	if a.lock != nil {
		a.lock.Release()
	}
	if a.DB != nil {
		a.DB.Close()
	}
}

// AcquireLock ロックを取得（エラー時はSlack通知）
func (a *App) AcquireLock(skip bool) error {
	lock, err := acquireLockInternal(skip)
	if err != nil {
		a.notifyError("バッチ起動失敗", err.Error())
		return err
	}
	a.lock = lock

	if !skip {
		log.Println("排他ロックを取得しました")
	}
	return nil
}

// notifyError Slack通知を送信（有効な場合のみ）
func (a *App) notifyError(title, message string) {
	slackClient := external.NewSlackClient(a.Config.Slack)
	if slackClient.IsEnabled() {
		slackClient.SendError(title, message)
	}
}

// ExecuteBatch パラメータに基づいてバッチを実行
func (a *App) ExecuteBatch(params BatchParams) BatchResult {
	// バッチタイプのバリデーション
	if !params.Type.IsValid() {
		errMsg := fmt.Sprintf("不明なバッチタイプ: %s (使用可能: article, amazon)", params.Type)
		log.Println(errMsg)
		return BatchResult{Success: false, Message: errMsg}
	}

	log.Printf("Starting %s...", params.Type.String())

	var err error
	switch params.Type {
	case BatchTypeArticle:
		fetchMode := parseFetchMode(params.Mode)
		err = runBatchProcess(a.Config, a.DB, fetchMode)
	case BatchTypeAmazon:
		limit := params.Limit
		if limit <= 0 {
			limit = getAmazonLimitFromEnv()
		}
		err = runAmazonBatchProcess(a.Config, a.DB, limit)
	}

	if err != nil {
		return BatchResult{
			Success: false,
			Message: fmt.Sprintf("%s失敗: %v", params.Type.String(), err),
		}
	}

	return BatchResult{
		Success: true,
		Message: fmt.Sprintf("%sが完了しました", params.Type.String()),
	}
}

// ============================================================================
// ロック処理
// ============================================================================

// BatchLock バッチの排他制御用ロック
type BatchLock struct {
	file *os.File
}

// acquireLockInternal ロックを取得（内部用）
func acquireLockInternal(skip bool) (*BatchLock, error) {
	if skip {
		log.Println("ファイルロックをスキップします")
		return &BatchLock{file: nil}, nil
	}

	// ロックファイルを開く（なければ作成）
	file, err := os.OpenFile(lockFilePath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open lock file: %w", err)
	}

	// 排他ロックを取得（ノンブロッキング）
	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		file.Close()
		if err == syscall.EWOULDBLOCK {
			return nil, fmt.Errorf("別のバッチプロセスが実行中です")
		}
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}

	// PIDをファイルに書き込み（デバッグ用）
	file.Truncate(0)
	file.Seek(0, 0)
	fmt.Fprintf(file, "%d\n", os.Getpid())

	return &BatchLock{file: file}, nil
}

// Release ロックを解放
func (l *BatchLock) Release() error {
	if l.file == nil {
		return nil
	}

	// ロックを解放
	if err := syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN); err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	// ファイルを閉じる
	if err := l.file.Close(); err != nil {
		return fmt.Errorf("failed to close lock file: %w", err)
	}

	// ロックファイルを削除
	os.Remove(lockFilePath)

	return nil
}

// ============================================================================
// Secrets Manager関連
// ============================================================================
// (共有パッケージ secrets に移動しました)

// ============================================================================
// ヘルパー関数
// ============================================================================

// parseFetchMode 文字列から取得モードを解析
func parseFetchMode(mode string) *usecase.FetchModeOption {
	switch mode {
	case "new":
		m := usecase.FetchModeOptionNew
		log.Println("Fetch mode: 最新記事取得")
		return &m
	case "historical":
		m := usecase.FetchModeOptionHistorical
		log.Println("Fetch mode: 過去記事取得")
		return &m
	default:
		log.Println("Fetch mode: 自動判定")
		return nil
	}
}

// getAmazonLimitFromEnv 環境変数からAmazon処理上限を取得
func getAmazonLimitFromEnv() int {
	if limitStr := os.Getenv("AMAZON_LIMIT"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			return limit
		}
	}
	return 50 // デフォルト値
}

// getMigrationsPath マイグレーションファイルのパスを取得
func getMigrationsPath() (string, error) {
	if path := os.Getenv("MIGRATIONS_PATH"); path != "" {
		return path, nil
	}

	execPath, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	return filepath.Join(execPath, "migrations"), nil
}

// ============================================================================
// バッチ処理実行（既存ロジック維持）
// ============================================================================

// runBatchProcess 記事取得バッチ処理を実行
func runBatchProcess(cfg *config.Config, db *postgres.DB, fetchMode *usecase.FetchModeOption) error {
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

// runAmazonBatchProcess Amazon URL取得バッチ処理を実行
func runAmazonBatchProcess(cfg *config.Config, db *postgres.DB, limit int) error {
	log.Println("===========================================")
	log.Println("  TeckBook Compass Amazon URL Fetch Batch")
	log.Printf("  開始時刻: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	log.Printf("  処理上限: %d 件\n", limit)
	log.Println("===========================================")

	// コンテキストを作成（タイムアウト付き）
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// リポジトリを初期化
	batchRepo := postgres.NewBatchRepository(db.DB)

	// 外部APIクライアントを初期化
	amazonClient := external.NewAmazonClient(cfg.Amazon)
	slackClient := external.NewSlackClient(cfg.Slack)

	if !amazonClient.IsEnabled() {
		log.Println("Amazon API is disabled. Please set AMAZON_ENABLED=true in your environment.")
		return fmt.Errorf("amazon api is disabled")
	}

	if slackClient.IsEnabled() {
		log.Println("Slack通知: 有効")
	} else {
		log.Println("Slack通知: 無効")
	}

	// ユースケースを初期化
	amazonBatchUsecase := usecase.NewAmazonBatchUsecase(batchRepo, amazonClient, slackClient)

	// バッチ処理を実行
	result, err := amazonBatchUsecase.Run(ctx, limit)
	if err != nil {
		// エラーが発生しても結果は出力する
		if result != nil {
			log.Println("===========================================")
			log.Println("  Amazon URL取得バッチ結果 (エラーで終了)")
			log.Println("===========================================")
			log.Printf("  処理した書籍数:   %d\n", result.ProcessedBooks)
			log.Printf("  更新した書籍数:   %d\n", result.UpdatedBooks)
			log.Printf("  未発見書籍数:     %d\n", result.NotFoundBooks)
			log.Printf("  エラー数:         %d\n", result.Errors)
			log.Printf("  エラー内容:       %s\n", result.ErrorMessage)
			log.Printf("  処理時間:         %v\n", result.EndTime.Sub(result.StartTime))
			log.Println("===========================================")
		}
		return err
	}

	// 結果を出力
	log.Println("===========================================")
	log.Println("  Amazon URL取得バッチ結果")
	log.Println("===========================================")
	log.Printf("  処理した書籍数:   %d\n", result.ProcessedBooks)
	log.Printf("  更新した書籍数:   %d\n", result.UpdatedBooks)
	log.Printf("  未発見書籍数:     %d\n", result.NotFoundBooks)
	log.Printf("  エラー数:         %d\n", result.Errors)
	log.Printf("  処理時間:         %v\n", result.EndTime.Sub(result.StartTime))
	log.Println("===========================================")
	log.Printf("  終了時刻: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	log.Println("===========================================")

	return nil
}
