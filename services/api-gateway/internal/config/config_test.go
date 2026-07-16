package config_test

import (
	"os"
	"testing"

	"github.com/lyh-fighting/zhipath-ai/services/api-gateway/internal/config"
)

func TestLoadMissingDatabaseURL(t *testing.T) {
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("INTERNAL_SERVICE_TOKEN")
	if _, err := config.Load(); err == nil {
		t.Fatal("DATABASE_URL 缺失时应返回错误，禁止默认连接生产")
	}
}

func TestLoadMissingInternalToken(t *testing.T) {
	t.Setenv("DATABASE_URL", "u:p@tcp(localhost:3306)/x")
	os.Unsetenv("INTERNAL_SERVICE_TOKEN")
	if _, err := config.Load(); err == nil {
		t.Fatal("INTERNAL_SERVICE_TOKEN 缺失时应返回错误")
	}
}

func TestLoadOK(t *testing.T) {
	t.Setenv("DATABASE_URL", "u:p@tcp(localhost:3306)/x")
	t.Setenv("INTERNAL_SERVICE_TOKEN", "tok")
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("配置完整时应成功: %v", err)
	}
	if cfg.InternalToken != "tok" {
		t.Errorf("InternalToken 不匹配: %s", cfg.InternalToken)
	}
	if cfg.HTTPPort != "8080" {
		t.Errorf("默认端口应为 8080: %s", cfg.HTTPPort)
	}
}
