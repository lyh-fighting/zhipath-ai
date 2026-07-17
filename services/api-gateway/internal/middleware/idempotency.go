package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
)

// Idempotency 基于 request_id 的幂等控制。
// 同一 request_id 重试返回相同结果（完整实现需 Redis 缓存响应体，此处先做 header 透传与校验）。
// TODO Task 8 后续：接入 Redis 缓存 request_id -> response，重复请求命中缓存直接返回。
func Idempotency() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		reqID := string(c.GetHeader("X-Request-Id"))
		if reqID == "" {
			// 也接受请求体内的 request_id（由 handler 解析后写入 header）
			reqID = string(c.GetHeader("request_id"))
		}
		if reqID != "" {
			ctx = context.WithValue(ctx, RequestIDKey, reqID)
			c.Header("X-Request-Id", reqID)
		}
		c.Next(ctx)
	}
}
