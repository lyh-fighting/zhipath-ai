package middleware

import (
	"context"
	"runtime/debug"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// Recovery 捕获 panic，返回 500 INTERNAL_ERROR，避免进程崩溃。
func Recovery() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		defer func() {
			if r := recover(); r != nil {
				hlog.Errorf("panic recovered: %v\n%s", r, debug.Stack())
				WriteError(ctx, c, 500, "INTERNAL_ERROR", "服务内部错误，请稍后重试", false)
				c.Abort()
			}
		}()
		c.Next(ctx)
	}
}
