package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"

	"github.com/lyh-fighting/zhipath-ai/services/api-gateway/internal/config"
	"github.com/lyh-fighting/zhipath-ai/services/api-gateway/internal/middleware"
	"github.com/lyh-fighting/zhipath-ai/services/api-gateway/internal/platform"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		hlog.Fatalf("加载配置失败: %v", err)
	}

	rel, err := platform.NewMySQLStore(cfg.DatabaseURL)
	if err != nil {
		hlog.Fatalf("初始化 MySQL 失败: %v", err)
	}
	cache, err := platform.NewRedisStore(cfg.RedisURL)
	if err != nil {
		hlog.Fatalf("初始化 Redis 失败: %v", err)
	}
	objStore := platform.NewObjectStore(
		cfg.ObjectStorage.Endpoint,
		cfg.ObjectStorage.AccessKey,
		cfg.ObjectStorage.SecretKey,
		cfg.ObjectStorage.UseSSL,
	)
	_ = objStore

	h := server.Default(server.WithHostPorts(":" + cfg.HTTPPort))
	h.Use(middleware.Recovery(), middleware.Trace())

	deps := &Deps{cfg: cfg, rel: rel, cache: cache}

	h.GET("/healthz", healthz)
	h.GET("/readyz", deps.readyz)

	hlog.Infof("api-gateway 启动于 :%s (env=%s)", cfg.HTTPPort, cfg.AppEnv)
	h.Spin()
}

// Deps 持有运行时依赖，供 handler 使用。
type Deps struct {
	cfg   *config.Config
	rel   platform.RelationalStore
	cache platform.CacheStore
}

// healthz 存活探针，不检查依赖。
func healthz(ctx context.Context, c *app.RequestContext) {
	middleware.WriteOK(ctx, c, map[string]any{"status": "ok"})
}

// readyz 就绪探针，检查 MySQL / Redis / AI Service。
func (d *Deps) readyz(ctx context.Context, c *app.RequestContext) {
	checks := map[string]string{}
	ok := true

	if err := d.rel.Ping(ctx); err != nil {
		checks["mysql"] = "down"
		ok = false
	} else {
		checks["mysql"] = "ok"
	}

	if err := d.cache.Ping(ctx); err != nil {
		checks["redis"] = "down"
		ok = false
	} else {
		checks["redis"] = "ok"
	}

	checks["ai_agent_service"] = checkUpstream(d.cfg.AIServiceURL + "/healthz")
	if checks["ai_agent_service"] != "ok" {
		ok = false
	}

	status := "ok"
	code := 200
	if !ok {
		status = "degraded"
		code = 503
	}
	c.SetStatusCode(code)
	c.Header("Content-Type", "application/json; charset=utf-8")
	body, _ := json.Marshal(map[string]any{"status": status, "checks": checks})
	c.Write(body)
}

// checkUpstream 检查上游服务健康（2s 超时）。
func checkUpstream(url string) string {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "down"
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("down:%d", resp.StatusCode)
	}
	return "ok"
}
