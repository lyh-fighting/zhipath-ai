package config

import (
	"fmt"
	"os"
)

// Config 是 api-gateway 的运行配置。
// 加载失败必须返回明确错误，禁止默认连接生产资源。
type Config struct {
	AppEnv        string
	HTTPPort      string
	DatabaseURL   string // Go MySQL DSN
	RedisURL      string
	AIServiceURL  string
	OCRServiceURL string
	InternalToken string // 服务间内部鉴权 token

	ObjectStorage ObjectStorageConfig
}

type ObjectStorageConfig struct {
	Endpoint  string
	Bucket    string
	AccessKey string
	SecretKey string
	UseSSL    bool
}

// Load 从环境变量加载配置。关键凭证缺失时返回错误。
func Load() (*Config, error) {
	cfg := &Config{
		AppEnv:        getenv("APP_ENV", "local"),
		HTTPPort:      getenv("API_GATEWAY_PORT", "8080"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		RedisURL:      getenv("REDIS_URL", "redis://localhost:6379/0"),
		AIServiceURL:  getenv("AI_AGENT_SERVICE_URL", "http://localhost:8001"),
		OCRServiceURL: getenv("OCR_SERVICE_URL", "http://localhost:8002"),
		InternalToken: os.Getenv("INTERNAL_SERVICE_TOKEN"),
		ObjectStorage: ObjectStorageConfig{
			Endpoint:  getenv("OBJECT_STORAGE_ENDPOINT", "localhost:9000"),
			Bucket:    getenv("OBJECT_STORAGE_BUCKET", "zhipath-ai"),
			AccessKey: getenv("OBJECT_STORAGE_ACCESS_KEY", "minioadmin"),
			SecretKey: getenv("OBJECT_STORAGE_SECRET_KEY", "minioadmin"),
			UseSSL:    getenv("OBJECT_STORAGE_USE_SSL", "false") == "true",
		},
	}
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("config: DATABASE_URL 未配置，拒绝启动")
	}
	if cfg.InternalToken == "" {
		return nil, fmt.Errorf("config: INTERNAL_SERVICE_TOKEN 未配置，拒绝启动")
	}
	return cfg, nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
