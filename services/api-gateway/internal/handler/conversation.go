package handler

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/lyh-fighting/zhipath-ai/services/api-gateway/internal/domain/conversation"
	"github.com/lyh-fighting/zhipath-ai/services/api-gateway/internal/middleware"
)

// ConversationHandler 处理会话与消息请求。
type ConversationHandler struct {
	repo *conversation.Repository
}

// NewConversationHandler 构造 ConversationHandler。
func NewConversationHandler(repo *conversation.Repository) *ConversationHandler {
	return &ConversationHandler{repo: repo}
}

// Create POST /api/v1/conversations
func (h *ConversationHandler) Create(ctx context.Context, c *app.RequestContext) {
	tenantID := middleware.TenantFromContext(ctx)
	userID := middleware.UserFromContext(ctx)
	var req conversation.CreateRequest
	if err := c.BindAndValidate(&req); err != nil {
		middleware.WriteError(ctx, c, 400, "INVALID_PARAM", "参数错误", false)
		return
	}
	conv, err := h.repo.Create(ctx, tenantID, userID, &req)
	if err != nil {
		middleware.WriteError(ctx, c, 500, "INTERNAL_ERROR", "创建会话失败", false)
		return
	}
	c.SetStatusCode(201)
	middleware.WriteOK(ctx, c, conv)
}

// List GET /api/v1/conversations（游标分页）。
func (h *ConversationHandler) List(ctx context.Context, c *app.RequestContext) {
	tenantID := middleware.TenantFromContext(ctx)
	userID := middleware.UserFromContext(ctx)
	limit, _ := strconv.Atoi(string(c.Query("limit")))
	cursor := string(c.Query("cursor"))
	var cursorPtr *string
	if cursor != "" {
		cursorPtr = &cursor
	}
	convs, err := h.repo.ListByUser(ctx, tenantID, userID, limit, cursorPtr)
	if err != nil {
		middleware.WriteError(ctx, c, 500, "INTERNAL_ERROR", "查询会话失败", false)
		return
	}
	middleware.WriteOK(ctx, c, map[string]any{"items": convs})
}

// SendMessage POST /api/v1/conversations/{conversation_id}/messages
func (h *ConversationHandler) SendMessage(ctx context.Context, c *app.RequestContext) {
	tenantID := middleware.TenantFromContext(ctx)
	userID := middleware.UserFromContext(ctx)
	convID := c.Param("conversation_id")
	conv, err := h.repo.Get(ctx, tenantID, convID)
	if err != nil {
		middleware.WriteError(ctx, c, 500, "INTERNAL_ERROR", "查询会话失败", false)
		return
	}
	if conv == nil {
		middleware.WriteError(ctx, c, 404, "NOT_FOUND", "会话不存在", false)
		return
	}
	if conv.UserID != userID {
		middleware.WriteError(ctx, c, 403, "RESOURCE_FORBIDDEN", "越权访问他人会话", false)
		return
	}
	var req conversation.SendMessageRequest
	if err := c.BindAndValidate(&req); err != nil {
		middleware.WriteError(ctx, c, 400, "INVALID_PARAM", "参数错误: message 必填", false)
		return
	}
	msg, err := h.repo.SendMessage(ctx, tenantID, userID, convID, &req)
	if err != nil {
		middleware.WriteError(ctx, c, 500, "INTERNAL_ERROR", "发送消息失败", false)
		return
	}
	// TODO Task 9+: 调用 AI Agent Service 写入 assistant 消息
	middleware.WriteOK(ctx, c, msg)
}

// ListMessages GET /api/v1/conversations/{conversation_id}/messages（游标分页）。
func (h *ConversationHandler) ListMessages(ctx context.Context, c *app.RequestContext) {
	tenantID := middleware.TenantFromContext(ctx)
	userID := middleware.UserFromContext(ctx)
	convID := c.Param("conversation_id")
	conv, err := h.repo.Get(ctx, tenantID, convID)
	if err != nil {
		middleware.WriteError(ctx, c, 500, "INTERNAL_ERROR", "查询会话失败", false)
		return
	}
	if conv == nil {
		middleware.WriteError(ctx, c, 404, "NOT_FOUND", "会话不存在", false)
		return
	}
	if conv.UserID != userID {
		middleware.WriteError(ctx, c, 403, "RESOURCE_FORBIDDEN", "越权访问他人会话", false)
		return
	}
	limit, _ := strconv.Atoi(string(c.Query("limit")))
	before := string(c.Query("before"))
	var beforePtr *string
	if before != "" {
		beforePtr = &before
	}
	msgs, err := h.repo.ListMessages(ctx, tenantID, convID, limit, beforePtr)
	if err != nil {
		middleware.WriteError(ctx, c, 500, "INTERNAL_ERROR", "查询消息失败", false)
		return
	}
	middleware.WriteOK(ctx, c, msgs)
}
