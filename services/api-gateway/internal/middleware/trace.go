package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
)

const (
	TraceIDKey   = "trace_id"
	RequestIDKey = "request_id"
)

// Trace 为每个请求生成或透传 trace_id / request_id，串联 Go 与 Python 服务日志。
func Trace() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		traceID := string(c.GetHeader("X-Trace-Id"))
		if traceID == "" {
			traceID = "trace_" + uuid.NewString()
		}
		reqID := string(c.GetHeader("X-Request-Id"))
		if reqID == "" {
			reqID = uuid.NewString()
		}
		ctx = context.WithValue(ctx, TraceIDKey, traceID)
		ctx = context.WithValue(ctx, RequestIDKey, reqID)
		c.Header("X-Trace-Id", traceID)
		c.Header("X-Request-Id", reqID)
		c.Next(ctx)
	}
}
