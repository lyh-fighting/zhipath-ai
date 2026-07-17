package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/lyh-fighting/zhipath-ai/services/api-gateway/internal/domain/order"
	"github.com/lyh-fighting/zhipath-ai/services/api-gateway/internal/domain/payment"
	"github.com/lyh-fighting/zhipath-ai/services/api-gateway/internal/middleware"
)

// OrderHandler 处理订单与支付。
type OrderHandler struct {
	repo     *order.Repository
	provider payment.PaymentProvider
}

// NewOrderHandler 构造 OrderHandler。
func NewOrderHandler(repo *order.Repository, provider payment.PaymentProvider) *OrderHandler {
	return &OrderHandler{repo: repo, provider: provider}
}

// Create POST /api/v1/orders
func (h *OrderHandler) Create(ctx context.Context, c *app.RequestContext) {
	tenantID := middleware.TenantFromContext(ctx)
	userID := middleware.UserFromContext(ctx)
	var req order.CreateRequest
	if err := c.BindAndValidate(&req); err != nil {
		middleware.WriteError(ctx, c, 400, "INVALID_PARAM", "参数错误: product_id 必填", false)
		return
	}
	o, err := h.repo.Create(ctx, tenantID, userID, &req)
	if err != nil {
		middleware.WriteError(ctx, c, 400, "INVALID_PARAM", err.Error(), false)
		return
	}
	payID, payURL, err := h.provider.CreatePayment(o.OrderID, o.AmountCents, o.OrderType)
	if err != nil {
		middleware.WriteError(ctx, c, 500, "INTERNAL_ERROR", "创建支付失败", false)
		return
	}
	c.SetStatusCode(201)
	middleware.WriteOK(ctx, c, map[string]any{
		"order_id":   o.OrderID,
		"payment_id": payID,
		"pay_url":    payURL,
		"provider":   h.provider.Name(),
	})
}

// Callback POST /api/v1/payments/callback（微信回调，验签+金额+幂等）
func (h *OrderHandler) Callback(ctx context.Context, c *app.RequestContext) {
	var payload map[string]any
	if err := c.Bind(&payload); err != nil {
		c.SetStatusCode(400)
		return
	}
	orderID, paid, err := h.provider.VerifyCallback(payload)
	if err != nil || !paid {
		c.SetStatusCode(400)
		return
	}
	h.repo.MarkPaid(ctx, "default_consumer", orderID) // 幂等 + 发 Outbox
	c.SetStatusCode(200)
}
