package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/lyh-fighting/zhipath-ai/services/api-gateway/internal/config"
)

// HistoryItem 拼装给 AI 的上下文历史单条消息。
type HistoryItem struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AgentResult AI Agent 调用结果。
type AgentResult struct {
	ContentSummary   string // AI 回复文本（来自 data.content_summary）
	NeedHumanHandoff bool   // 是否需转人工
	QualityScore     int    // 质量评分
}

// AgentClient 调用 ai-agent 服务的内部接口。
// 网关作为内部调用方，携带 X-Internal-Token，超时必须足够大以容纳真实模型推理（约 45-60s）。
type AgentClient struct {
	baseURL string
	token   string
	client  *http.Client
}

// NewAgentClient 构造客户端，固定 120s 超时应对真实模型慢响应。
func NewAgentClient(cfg *config.Config) *AgentClient {
	return &AgentClient{
		baseURL: cfg.AIServiceURL,
		token:   cfg.InternalToken,
		client:  &http.Client{Timeout: 120 * time.Second},
	}
}

// invokeRequest 对应 ai-agent /internal/v1/agent/invoke 的请求体。
type invokeRequest struct {
	Message          string        `json:"message"`
	ConversationType string        `json:"conversation_type"`
	History          []HistoryItem `json:"history"`
	TraceID          string        `json:"trace_id"`
}

// invokeResponse 对应 ai-agent 的响应体（仅取需要的字段）。
type invokeResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		ContentSummary   string `json:"content_summary"`
		NeedHumanHandoff bool   `json:"need_human_handoff"`
		QualityScore     int    `json:"quality_score"`
	} `json:"data"`
	TraceID string `json:"trace_id"`
}

// Invoke 调用 AI Agent Service，返回 AI 文本与转人工标志。
// 任何失败（网络错误、超时、非 2xx、业务 code != SUCCESS）都以 error 返回，由调用方优雅降级。
func (c *AgentClient) Invoke(ctx context.Context, message, conversationType string, history []HistoryItem, traceID string) (*AgentResult, error) {
	if conversationType == "" {
		conversationType = "career"
	}
	body, err := json.Marshal(invokeRequest{
		Message:          message,
		ConversationType: conversationType,
		History:          history,
		TraceID:          traceID,
	})
	if err != nil {
		return nil, fmt.Errorf("ai: 序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/internal/v1/agent/invoke", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ai: 构造请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Token", c.token)
	if traceID != "" {
		req.Header.Set("X-Trace-Id", traceID)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ai: HTTP 调用失败: %w", err)
	}
	defer resp.Body.Close()

	var parsed invokeResponse
	if decErr := json.NewDecoder(resp.Body).Decode(&parsed); decErr != nil {
		return nil, fmt.Errorf("ai: 解析响应失败 (status=%d): %w", resp.StatusCode, decErr)
	}
	if resp.StatusCode != http.StatusOK || parsed.Code != "SUCCESS" {
		return nil, fmt.Errorf("ai: 调用未成功 (status=%d, code=%s, msg=%s)",
			resp.StatusCode, parsed.Code, parsed.Message)
	}

	return &AgentResult{
		ContentSummary:   parsed.Data.ContentSummary,
		NeedHumanHandoff: parsed.Data.NeedHumanHandoff,
		QualityScore:     parsed.Data.QualityScore,
	}, nil
}
