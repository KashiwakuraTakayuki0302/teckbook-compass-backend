//go:build lambda

package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
)

// ============================================================================
// Lambda イベント/レスポンス定義
// ============================================================================

// LambdaEvent EventBridgeから受け取るイベント構造体
type LambdaEvent struct {
	Type  string `json:"type"`  // バッチの種類 ("article" or "amazon")
	Mode  string `json:"mode"`  // 取得モード ("new", "historical", "auto") - articleバッチ用
	Limit int    `json:"limit"` // 処理上限 - amazonバッチ用
}

// LambdaResponse Lambda用のレスポンス構造体
type LambdaResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ============================================================================
// エントリポイント
// ============================================================================

func main() {
	log.Println("Lambda環境で起動しています")
	lambda.Start(handleLambdaEvent)
}

// ============================================================================
// ハンドラー
// ============================================================================

// handleLambdaEvent Lambdaイベントをハンドリング
func handleLambdaEvent(ctx context.Context, event LambdaEvent) (LambdaResponse, error) {
	log.Printf("Lambda event received: type=%s, mode=%s, limit=%d",
		event.Type, event.Mode, event.Limit)

	// イベントからパラメータを構築
	params := buildBatchParams(event)

	// アプリケーション初期化
	app, err := NewApp()
	if err != nil {
		log.Printf("初期化失敗: %v", err)
		return LambdaResponse{Success: false, Message: err.Error()}, err
	}
	defer app.Close()

	// Lambda環境ではファイルロックをスキップ
	if err := app.AcquireLock(true); err != nil {
		return LambdaResponse{Success: false, Message: err.Error()}, err
	}

	// バッチ実行
	result := app.ExecuteBatch(params)

	if !result.Success {
		return LambdaResponse{Success: false, Message: result.Message}, nil
	}

	return LambdaResponse{Success: true, Message: result.Message}, nil
}

// ============================================================================
// ヘルパー関数
// ============================================================================

// buildBatchParams イベントからBatchParamsを構築
func buildBatchParams(event LambdaEvent) BatchParams {
	// バッチタイプ: イベント → 環境変数 → デフォルト の優先順位
	batchType := BatchType(event.Type)
	if batchType == "" {
		batchType = BatchType(os.Getenv("BATCH_TYPE"))
	}
	if batchType == "" {
		batchType = BatchTypeArticle // デフォルト
	}

	return BatchParams{
		Type:  batchType,
		Mode:  event.Mode,
		Limit: event.Limit,
	}
}

// ============================================================================
// ログ初期化
// ============================================================================

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
