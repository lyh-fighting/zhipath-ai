package middleware

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/lyh-fighting/zhipath-ai/services/api-gateway/internal/auth"
)

// 上下文键，handler 通过 FromContext 取身份。
type ctxKey string

const (
	CtxTenantIDKey ctxKey = "ctx_tenant_id"
	CtxUserIDKey   ctxKey = "ctx_user_id"
)

// TenantFromContext 从 ctx 取 tenant_id。
func TenantFromContext(ctx context.Context) string {
	v, _ := ctx.Value(CtxTenantIDKey).(string)
	return v
}

// UserFromContext 从 ctx 取 user_id（来自 token，非请求体）。
func UserFromContext(ctx context.Context) string {
	v, _ := ctx.Value(CtxUserIDKey).(string)
	return v
}

// Auth Bearer token 鉴权。
// 身份由 token 推导，忽略请求体中的 tenant_id / user_id 字段。
func Auth(svc *auth.TokenService) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		authz := string(c.GetHeader("Authorization"))
		if !strings.HasPrefix(authz, "Bearer ") {
			WriteError(ctx, c, 401, "AUTH_FAILED", "缺少 Authorization Bearer token", false)
			c.Abort()
			return
		}
		tokenStr := strings.TrimPrefix(authz, "Bearer ")
		claims, err := svc.Parse(tokenStr)
		if err != nil {
			WriteError(ctx, c, 401, "TOKEN_EXPIRED", "token 无效或已过期", false)
			c.Abort()
			return
		}
		ctx = context.WithValue(ctx, CtxTenantIDKey, claims.TenantID)
		ctx = context.WithValue(ctx, CtxUserIDKey, claims.UserID)
		c.Next(ctx)
	}
}
