package handler

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/lyh-fighting/zhipath-ai/services/api-gateway/internal/domain/mbti"
	"github.com/lyh-fighting/zhipath-ai/services/api-gateway/internal/domain/profile"
	"github.com/lyh-fighting/zhipath-ai/services/api-gateway/internal/middleware"
)

// MBTIHandler 处理 MBTI HTTP 请求。
type MBTIHandler struct {
	mbtiRepo    *mbti.Repository
	profileRepo *profile.Repository
}

// NewMBTIHandler 构造 MBTIHandler。
func NewMBTIHandler(mr *mbti.Repository, pr *profile.Repository) *MBTIHandler {
	return &MBTIHandler{mbtiRepo: mr, profileRepo: pr}
}

// List GET /api/v1/me/mbti（历史 + 当前结果）。
func (h *MBTIHandler) List(ctx context.Context, c *app.RequestContext) {
	tenantID := middleware.TenantFromContext(ctx)
	userID := middleware.UserFromContext(ctx)
	results, err := h.mbtiRepo.ListByUser(ctx, tenantID, userID, 20)
	if err != nil {
		middleware.WriteError(ctx, c, 500, "INTERNAL_ERROR", "查询 MBTI 失败", false)
		return
	}
	p, _ := h.profileRepo.Get(ctx, tenantID, userID)
	var current *mbti.Result
	if p != nil && p.CurrentMBTIResultID != nil {
		current, _ = h.mbtiRepo.Get(ctx, tenantID, *p.CurrentMBTIResultID)
	}
	for _, r := range results {
		if current != nil && r.MBTIResultID == current.MBTIResultID {
			r.IsCurrent = true
		}
	}
	middleware.WriteOK(ctx, c, map[string]any{
		"current": current,
		"history": results,
	})
}

type mbtiSubmitReq struct {
	ResultType    string          `json:"result_type" vd:"required"`
	Assertiveness string          `json:"assertiveness"`
	Dimensions    mbti.Dimensions `json:"dimensions"`
	Source        string          `json:"source"`
	TestURL       string          `json:"test_url"`
	TestedAt      *time.Time      `json:"tested_at"`
}

// Submit POST /api/v1/me/mbti（手动提交，默认已确认并设为当前）。
func (h *MBTIHandler) Submit(ctx context.Context, c *app.RequestContext) {
	tenantID := middleware.TenantFromContext(ctx)
	userID := middleware.UserFromContext(ctx)
	var req mbtiSubmitReq
	if err := c.BindAndValidate(&req); err != nil {
		middleware.WriteError(ctx, c, 400, "INVALID_PARAM", "参数错误: result_type 必填", false)
		return
	}
	if req.Source == "" {
		req.Source = "manual"
	}
	res := &mbti.Result{
		UserID:          userID,
		Source:          req.Source,
		TestURL:         req.TestURL,
		ResultType:      req.ResultType,
		Assertiveness:   req.Assertiveness,
		Dimensions:      req.Dimensions,
		TestedAt:        req.TestedAt,
		ConfidenceScore: 1.0,
		ConfirmedByUser: true,
	}
	if err := h.mbtiRepo.Create(ctx, tenantID, res); err != nil {
		middleware.WriteError(ctx, c, 500, "INTERNAL_ERROR", "保存 MBTI 失败", false)
		return
	}
	if err := h.profileRepo.SetCurrentMBTI(ctx, tenantID, userID, res.MBTIResultID); err != nil {
		middleware.WriteError(ctx, c, 500, "INTERNAL_ERROR", "设置当前 MBTI 失败", false)
		return
	}
	res.IsCurrent = true
	middleware.WriteOK(ctx, c, res)
}

// Confirm POST /api/v1/me/mbti/{mbti_result_id}/confirm
// 截图识别结果必须经用户确认才成当前。校验越权返回 403。
func (h *MBTIHandler) Confirm(ctx context.Context, c *app.RequestContext) {
	tenantID := middleware.TenantFromContext(ctx)
	userID := middleware.UserFromContext(ctx)
	mbtiResultID := c.Param("mbti_result_id")
	res, err := h.mbtiRepo.Get(ctx, tenantID, mbtiResultID)
	if err != nil {
		middleware.WriteError(ctx, c, 500, "INTERNAL_ERROR", "查询失败", false)
		return
	}
	if res == nil {
		middleware.WriteError(ctx, c, 404, "NOT_FOUND", "MBTI 结果不存在", false)
		return
	}
	if res.UserID != userID {
		middleware.WriteError(ctx, c, 403, "RESOURCE_FORBIDDEN", "越权访问他人 MBTI 结果", false)
		return
	}
	if err := h.mbtiRepo.Confirm(ctx, tenantID, mbtiResultID); err != nil {
		middleware.WriteError(ctx, c, 500, "INTERNAL_ERROR", "确认失败", false)
		return
	}
	if err := h.profileRepo.SetCurrentMBTI(ctx, tenantID, userID, mbtiResultID); err != nil {
		middleware.WriteError(ctx, c, 500, "INTERNAL_ERROR", "设置当前 MBTI 失败", false)
		return
	}
	res.ConfirmedByUser = true
	res.IsCurrent = true
	middleware.WriteOK(ctx, c, res)
}
