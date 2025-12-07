package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	// .envファイルを読み込み（存在しない場合はスキップ）
	if err := godotenv.Load(); err != nil {
		log.Println("Info: .env file not found, using environment variables")
	}
}

// Config アプリケーション設定
type Config struct {
	ServerPort string
	Env        string
	Database   DatabaseConfig
	Qiita      QiitaConfig
	Rakuten    RakutenConfig
	Amazon     AmazonConfig
	Slack      SlackConfig
}

// SlackConfig Slack通知設定
type SlackConfig struct {
	WebhookURL string // Incoming Webhook URL（簡易通知用）
	BotToken   string // Bot Token（スレッド返信用、xoxb-...）
	ChannelID  string // 通知先チャンネルID
	Enabled    bool   // Slack通知を有効にするか
}

// QiitaConfig Qiita API設定
type QiitaConfig struct {
	AccessToken string
	BaseURL     string
}

// RakutenConfig 楽天ブックスAPI設定
type RakutenConfig struct {
	ApplicationID     string
	ApplicationSecret string
	AffiliateID       string
	BaseURL           string
}

// AmazonConfig Amazon Product Advertising API設定
type AmazonConfig struct {
	AccessKey  string
	SecretKey  string
	PartnerTag string
	Region     string
	BaseURL    string
	Enabled    bool // Amazon APIを有効にするかどうか（後で追加するため）
}

// DatabaseConfig データベース接続設定
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// DSN データベース接続文字列を生成
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode,
	)
}

// NewConfig 設定を初期化
func NewConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	return &Config{
		ServerPort: port,
		Env:        env,
		Database:   newDatabaseConfig(),
		Qiita:      newQiitaConfig(),
		Rakuten:    newRakutenConfig(),
		Amazon:     newAmazonConfig(),
		Slack:      newSlackConfig(),
	}
}

// newSlackConfig Slack通知設定を初期化
func newSlackConfig() SlackConfig {
	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	botToken := os.Getenv("SLACK_BOT_TOKEN")
	channelID := os.Getenv("SLACK_CHANNEL_ID")

	// WebhookURLまたはBotTokenが設定されていれば有効
	enabled := webhookURL != "" || botToken != ""

	return SlackConfig{
		WebhookURL: webhookURL,
		BotToken:   botToken,
		ChannelID:  channelID,
		Enabled:    enabled,
	}
}

// newQiitaConfig Qiita API設定を初期化
func newQiitaConfig() QiitaConfig {
	baseURL := os.Getenv("QIITA_BASE_URL")
	if baseURL == "" {
		baseURL = "https://qiita.com/api/v2"
	}

	return QiitaConfig{
		AccessToken: os.Getenv("QIITA_ACCESS_TOKEN"),
		BaseURL:     baseURL,
	}
}

// newRakutenConfig 楽天ブックスAPI設定を初期化
func newRakutenConfig() RakutenConfig {
	baseURL := os.Getenv("RAKUTEN_BASE_URL")
	if baseURL == "" {
		baseURL = "https://app.rakuten.co.jp/services/api/BooksBook/Search/20170404"
	}

	return RakutenConfig{
		ApplicationID:     os.Getenv("RAKUTEN_APPLICATION_ID"),
		ApplicationSecret: os.Getenv("RAKUTEN_APPLICATION_SECRET"),
		AffiliateID:       os.Getenv("RAKUTEN_AFFILIATE_ID"),
		BaseURL:           baseURL,
	}
}

// newAmazonConfig Amazon Product Advertising API設定を初期化
func newAmazonConfig() AmazonConfig {
	return AmazonConfig{
		AccessKey:  os.Getenv("AMAZON_ACCESS_KEY"),
		SecretKey:  os.Getenv("AMAZON_SECRET_KEY"),
		PartnerTag: os.Getenv("AMAZON_PARTNER_TAG"),
		Region:     "us-west-2",
		BaseURL:    "webservices.amazon.co.jp",
		Enabled:    false, // 後で追加するため、デフォルトは無効
	}
}

// newDatabaseConfig データベース設定を初期化
func newDatabaseConfig() DatabaseConfig {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		user = "test"
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "password"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "teckbook"
	}

	sslMode := os.Getenv("DB_SSLMODE")
	if sslMode == "" {
		sslMode = "disable"
	}

	return DatabaseConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbName,
		SSLMode:  sslMode,
	}
}
