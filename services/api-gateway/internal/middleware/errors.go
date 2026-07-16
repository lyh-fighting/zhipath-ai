package middleware

import (
	"context"
	"encoding/json"

	"github.com/cloudwego/hertz/pkg/app"
)

// 统一错误响应结构（与 OpenAPI 契约一致）
type errorBody struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	TraceID   string `json:"trace_id"`
	Retryable bool   `json:"retryable"`
}

// WriteError 写统一错误响应，便于 handler 复用。
func WriteError(ctx context.Context, c *app.RequestContext, status int, code, msg string, retryable bool) {
	traceID, _ := ctx.Value(TraceIDKey).(string)
	body, _ := json.Marshal(errorBody{Code: code, Message: msg, TraceID: traceID, Retryable: retryable})
	c.SetStatusCode(status)
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.Write(body)
}

// WriteOK 写统一成功响应 {code:SUCCESS, message, data, trace_id}。
func WriteOK(ctx context.Context, c *app.RequestContext, data any) {
	traceID, _ := ctx.Value(TraceIDKey).(string)
	body, _ := json.Marshal(map[string]any{
		"code":     "SUCCESS",
		"message":  "ok",
		"data":     data,
		"trace_id": traceID,
	})
	c.SetStatusCode(200)
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.Write(body)
}
