package order

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Repository 订单数据访问。
type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create 创建订单。校验产品存在且 active。
func (r *Repository) Create(ctx context.Context, tenantID, userID string, req *CreateRequest) (*Order, error) {
	product, ok := Catalog[req.ProductID]
	if !ok {
		return nil, fmt.Errorf("产品不存在: %s", req.ProductID)
	}
	if product.Status != "active" {
		return nil, fmt.Errorf("产品已下架")
	}
	o := &Order{
		OrderID:     "o_" + uuid.NewString(),
		TenantID:    tenantID,
		UserID:      userID,
		ProductID:   product.ProductID,
		OrderType:   product.ProductType,
		AmountCents: product.PriceCents,
		Currency:    product.Currency,
		Status:      "pending",
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO orders (tenant_id, order_id, user_id, product_id, order_type, amount_cents, currency, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'pending')`,
		o.TenantID, o.OrderID, o.UserID, o.ProductID, o.OrderType, o.AmountCents, o.Currency,
	)
	if err != nil {
		return nil, fmt.Errorf("order create: %w", err)
	}
	return o, nil
}

// Get 查询订单。
func (r *Repository) Get(ctx context.Context, tenantID, orderID string) (*Order, error) {
	o := &Order{}
	err := r.db.QueryRowContext(ctx, `
		SELECT order_id, tenant_id, user_id, product_id, order_type, amount_cents, currency, status, paid_at, created_at
		FROM orders WHERE tenant_id = ? AND order_id = ?`,
		tenantID, orderID,
	).Scan(&o.OrderID, &o.TenantID, &o.UserID, &o.ProductID, &o.OrderType, &o.AmountCents, &o.Currency, &o.Status, &o.PaidAt, &o.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("order get: %w", err)
	}
	return o, nil
}

// MarkPaid 幂等更新：已 paid 则跳过，不重复发 Outbox 事件。
// 付款成功后发 Outbox 事件，不直接调下游。
func (r *Repository) MarkPaid(ctx context.Context, tenantID, orderID string) (bool, error) {
	o, err := r.Get(ctx, tenantID, orderID)
	if err != nil || o == nil {
		return false, fmt.Errorf("订单不存在")
	}
	if o.Status == "paid" {
		return false, nil // 幂等
	}
	now := time.Now()
	_, err = r.db.ExecContext(ctx, `UPDATE orders SET status='paid', paid_at=? WHERE order_id=? AND status='pending'`, now, orderID)
	if err != nil {
		return false, err
	}
	// TODO: INSERT INTO outbox_events (event_type='payment_succeeded', aggregate_id=orderID)
	return true, nil
}
