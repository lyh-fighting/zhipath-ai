package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/lyh-fighting/zhipath-ai/services/api-gateway/internal/auth"
	"github.com/lyh-fighting/zhipath-ai/services/api-gateway/internal/middleware"
)

// AuthHandler 处理认证相关 HTTP 请求。
type AuthHandler struct {
	svc *auth.Service
}

// NewAuthHandler 构造 AuthHandler。
func NewAuthHandler(svc *auth.Service) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type wechatLoginReq struct {
	Code string `json:"code" vd:"required"`
}

// WechatLogin POST /api/v1/auth/wechat/login
// userID 暂由 code 派生，Task 7 接入用户持久化后改为查找/创建。
func (h *AuthHandler) WechatLogin(ctx context.Context, c *app.RequestContext) {
	var req wechatLoginReq
	if err := c.BindAndValidate(&req); err != nil {
		middleware.WriteError(ctx, c, 400, "INVALID_PARAM", "参数错误: code 必填", false)
		return
	}
	userID := "u_" + req.Code
	access, refresh, err := h.svc.LoginByWechat(ctx, req.Code, userID)
	if err != nil {
		middleware.WriteError(ctx, c, 500, "INTERNAL_ERROR", "登录失败", false)
		return
	}
	middleware.WriteOK(ctx, c, map[string]any{
		"access_token":  access,
		"refresh_token": refresh,
		"expires_in":    7200,
		"user_id":       userID,
	})
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token" vd:"required"`
}

// Refresh POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(ctx context.Context, c *app.RequestContext) {
	var req refreshReq
	if err := c.BindAndValidate(&req); err != nil {
		middleware.WriteError(ctx, c, 400, "INVALID_PARAM", "参数错误: refresh_token 必填", false)
		return
	}
	access, refresh, err := h.svc.Refresh(req.RefreshToken)
	if err != nil {
		middleware.WriteError(ctx, c, 401, "AUTH_FAILED", "refresh token 无效", false)
		return
	}
	middleware.WriteOK(ctx, c, map[string]any{
		"access_token":  access,
		"refresh_token": refresh,
		"expires_in":    7200,
	})
}
