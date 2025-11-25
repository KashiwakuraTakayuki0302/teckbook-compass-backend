package config

import "os"

// Config アプリケーション設定
type Config struct {
	ServerPort string
	Env        string
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
	}
}
