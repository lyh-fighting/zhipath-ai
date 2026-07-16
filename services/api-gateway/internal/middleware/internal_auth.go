package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
)

// InternalAuth 服务间鉴权，校验 X-Internal-Token。
// 服务拒绝无凭证调用；Go 调 AI/OCR 必须携带此 token。
func InternalAuth(expected string) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		got := string(c.GetHeader("X-Internal-Token"))
		if got == "" || got != expected {
			WriteError(ctx, c, 401, "AUTH_FAILED", "内部服务鉴权失败", false)
			c.Abort()
			return
		}
		c.Next(ctx)
	}
}
