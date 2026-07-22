package handler

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"

	"github.com/lyh-fighting/zhipath-ai/services/api-gateway/internal/ai"
	"github.com/lyh-fighting/zhipath-ai/services/api-gateway/internal/domain/conversation"
	"github.com/lyh-fighting/zhipath-ai/services/api-gateway/internal/middleware"
)

// ConversationHandler 处理会话与消息请求。
type ConversationHandler struct {
	repo     *conversation.Repository
	aiClient *ai.AgentClient
}

// NewConversationHandler 构造 ConversationHandler。
func NewConversationHandler(repo *conversation.Repository, aiClient *ai.AgentClient) *ConversationHandler {
	return &ConversationHandler{repo: repo, aiClient: aiClient}
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
// 落库用户消息后调用 AI Agent Service，并把返回的 assistant 文本落库再一并返回。
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

	traceID, _ := ctx.Value(middleware.TraceIDKey).(string)

	userMsg, err := h.repo.SendMessage(ctx, tenantID, userID, convID, &req)
	if err != nil {
		hlog.Errorf("发送用户消息失败 (conv=%s, trace=%s): %v", convID, traceID, err)
		middleware.WriteError(ctx, c, 500, "INTERNAL_ERROR", "发送消息失败", false)
		return
	}

	// conversation_type 传会话 domain；auto/空按 ai-agent 默认 career 处理。
	convType := conv.Domain
	if convType == "" || convType == "auto" {
		convType = "career"
	}

	// 拼装最近历史（role/content），供 AI 维持上下文。
	var history []ai.HistoryItem
	if hist, herr := h.repo.History(ctx, tenantID, convID, 20); herr == nil {
		for _, m := range hist {
			if m.ContentSummary == "" {
				continue
			}
			history = append(history, ai.HistoryItem{Role: m.Role, Content: m.ContentSummary})
		}
	} else {
		hlog.Warnf("获取会话历史失败，将不带历史调用 AI: %v", herr)
	}

	// 调用 AI Agent Service。失败不 panic、不整体 500，仅标记 AI 不可用并返回已保存的用户消息。
	if h.aiClient == nil {
		middleware.WriteOK(ctx, c, map[string]any{
			"user_message":      userMsg,
			"assistant_message": nil,
			"ai_available":      false,
			"ai_error":          "AI 服务未配置",
		})
		return
	}

	aiResult, aiErr := h.aiClient.Invoke(ctx, req.Message, convType, history, traceID)
	if aiErr != nil {
		hlog.Errorf("AI 调用失败，仅返回用户消息 (conv=%s, trace=%s): %v", convID, traceID, aiErr)
		middleware.WriteOK(ctx, c, map[string]any{
			"user_message":      userMsg,
			"assistant_message": nil,
			"ai_available":      false,
			"ai_error":          "AI 服务暂不可用，您的消息已保存，请稍后重试",
		})
		return
	}

	assistantMsg, sErr := h.repo.SaveAssistantMessage(ctx, tenantID, userID, convID, aiResult.ContentSummary)
	if sErr != nil {
		hlog.Errorf("落库 assistant 消息失败 (conv=%s, trace=%s): %v", convID, traceID, sErr)
		middleware.WriteOK(ctx, c, map[string]any{
			"user_message":      userMsg,
			"assistant_message": nil,
			"ai_available":      true,
			"ai_error":          "AI 已生成回复，但保存失败，请刷新查看",
		})
		return
	}

	middleware.WriteOK(ctx, c, map[string]any{
		"user_message":       userMsg,
		"assistant_message":  assistantMsg,
		"ai_available":       true,
		"need_human_handoff": aiResult.NeedHumanHandoff,
	})
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
