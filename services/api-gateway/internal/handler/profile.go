package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/lyh-fighting/zhipath-ai/services/api-gateway/internal/domain/profile"
	"github.com/lyh-fighting/zhipath-ai/services/api-gateway/internal/middleware"
)

// ProfileHandler 处理用户画像 HTTP 请求。
type ProfileHandler struct {
	repo *profile.Repository
}

// NewProfileHandler 构造 ProfileHandler。
func NewProfileHandler(repo *profile.Repository) *ProfileHandler {
	return &ProfileHandler{repo: repo}
}

// Get GET /api/v1/me/profile（user_id 由 token 推导）。
func (h *ProfileHandler) Get(ctx context.Context, c *app.RequestContext) {
	tenantID := middleware.TenantFromContext(ctx)
	userID := middleware.UserFromContext(ctx)
	p, err := h.repo.Get(ctx, tenantID, userID)
	if err != nil {
		middleware.WriteError(ctx, c, 500, "INTERNAL_ERROR", "查询画像失败", false)
		return
	}
	if p == nil {
		p = &profile.Profile{TenantID: tenantID, UserID: userID}
	}
	middleware.WriteOK(ctx, c, p)
}

// Update PUT /api/v1/me/profile（仅更新非 nil 字段）。
func (h *ProfileHandler) Update(ctx context.Context, c *app.RequestContext) {
	tenantID := middleware.TenantFromContext(ctx)
	userID := middleware.UserFromContext(ctx)
	var req profile.UpdateRequest
	if err := c.BindAndValidate(&req); err != nil {
		middleware.WriteError(ctx, c, 400, "INVALID_PARAM", "参数错误", false)
		return
	}
	existing, err := h.repo.Get(ctx, tenantID, userID)
	if err != nil {
		middleware.WriteError(ctx, c, 500, "INTERNAL_ERROR", "查询画像失败", false)
		return
	}
	p := &profile.Profile{TenantID: tenantID, UserID: userID}
	if existing != nil {
		p = existing
	}
	if req.AgeRange != nil {
		p.AgeRange = *req.AgeRange
	}
	if req.City != nil {
		p.City = *req.City
	}
	if req.Education != nil {
		p.Education = *req.Education
	}
	if req.Occupation != nil {
		p.Occupation = *req.Occupation
	}
	if req.Industry != nil {
		p.Industry = *req.Industry
	}
	if req.WorkYears != nil {
		p.WorkYears = req.WorkYears
	}
	if req.IncomeRange != nil {
		p.IncomeRange = *req.IncomeRange
	}
	if req.RelationshipStatus != nil {
		p.RelationshipStatus = *req.RelationshipStatus
	}
	p.ProfileCompleteness = computeCompleteness(p)
	if err := h.repo.Upsert(ctx, p); err != nil {
		middleware.WriteError(ctx, c, 500, "INTERNAL_ERROR", "更新画像失败", false)
		return
	}
	middleware.WriteOK(ctx, c, p)
}

// computeCompleteness 计算画像完整度（0-100）。
func computeCompleteness(p *profile.Profile) int {
	score := 0
	if p.AgeRange != "" {
		score += 10
	}
	if p.City != "" {
		score += 10
	}
	if p.Education != "" {
		score += 10
	}
	if p.Occupation != "" {
		score += 15
	}
	if p.Industry != "" {
		score += 10
	}
	if p.WorkYears != nil {
		score += 10
	}
	if p.IncomeRange != "" {
		score += 10
	}
	if p.RelationshipStatus != "" {
		score += 10
	}
	if p.CurrentMBTIResultID != nil {
		score += 15
	}
	return score
}
