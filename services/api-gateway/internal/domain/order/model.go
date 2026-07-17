package order

import "time"

// Product 产品（会员/深度报告两类）。
type Product struct {
	ProductID   string `json:"product_id"`
	ProductType string `json:"product_type"` // membership|deep_report|consultation
	Name        string `json:"name"`
	PriceCents  int64  `json:"price_cents"`
	Currency    string `json:"currency"`
	Status      string `json:"status"`
}

// Order 订单。
type Order struct {
	OrderID     string     `json:"order_id"`
	TenantID    string     `json:"tenant_id"`
	UserID      string     `json:"user_id"`
	ProductID   string     `json:"product_id"`
	OrderType   string     `json:"order_type"`
	AmountCents int64      `json:"amount_cents"`
	Currency    string     `json:"currency"`
	Status      string     `json:"status"` // pending|paid|cancelled|refunded
	PaidAt      *time.Time `json:"paid_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// CreateRequest 创建订单请求。
type CreateRequest struct {
	ProductID string `json:"product_id" vd:"required"`
}

// 产品目录（MVP 内置，后续入库）
var Catalog = map[string]Product{
	"prod_membership_monthly": {ProductID: "prod_membership_monthly", ProductType: "membership", Name: "月度会员", PriceCents: 2900, Currency: "CNY", Status: "active"},
	"prod_deep_report":        {ProductID: "prod_deep_report", ProductType: "deep_report", Name: "深度报告", PriceCents: 9900, Currency: "CNY", Status: "active"},
}
