package secrets

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	"teckbook-compass-backend/internal/infrastructure/config"
)

const (
	// RDSSecretName RDSのシークレット名
	RDSSecretName = "rds!db-b4c4e561-b94f-45be-bc58-48bc79bcee6f"
)

// RDSSecret RDSシークレットの構造体
type RDSSecret struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoadDatabaseCredentials Secrets ManagerからDB認証情報を取得して設定に反映
func LoadDatabaseCredentials(cfg *config.Config) error {
	ctx := context.Background()

	// AWS設定を読み込み
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("AWS設定の読み込み失敗: %w", err)
	}

	// Secrets Managerクライアントを作成
	client := secretsmanager.NewFromConfig(awsCfg)

	// シークレットを取得
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(RDSSecretName),
	}

	result, err := client.GetSecretValue(ctx, input)
	if err != nil {
		return fmt.Errorf("シークレット取得失敗: %w", err)
	}

	// JSONをパース
	var rdsSecret RDSSecret
	if err := json.Unmarshal([]byte(*result.SecretString), &rdsSecret); err != nil {
		return fmt.Errorf("シークレットJSONのパース失敗: %w", err)
	}

	// 設定に反映
	if rdsSecret.Username != "" {
		cfg.Database.User = rdsSecret.Username
		log.Println("Secrets Managerからusernameを取得しました")
	}
	if rdsSecret.Password != "" {
		cfg.Database.Password = rdsSecret.Password
		log.Println("Secrets Managerからpasswordを取得しました")
	}

	return nil
}
