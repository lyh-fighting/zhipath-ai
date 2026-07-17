package payment

import "fmt"

// PaymentProvider 支付 provider 接口。
// 本地用 Mock（显式标记 provider=mock）；生产用 Wechat（验签/金额/幂等）。
type PaymentProvider interface {
	Name() string // mock|wechat
	CreatePayment(orderID string, amountCents int64, description string) (paymentID, payURL string, err error)
	VerifyCallback(payload map[string]any) (orderID string, paid bool, err error)
}

// MockPayment 本地 mock，显式标记 provider=mock，不静默伪造成功。
type MockPayment struct{}

func (m *MockPayment) Name() string { return "mock" }

func (m *MockPayment) CreatePayment(orderID string, amountCents int64, description string) (string, string, error) {
	return "pay_mock_" + orderID, "http://localhost/mock-pay?order=" + orderID, nil
}

func (m *MockPayment) VerifyCallback(payload map[string]any) (string, bool, error) {
	orderID, _ := payload["order_id"].(string)
	return orderID, true, nil
}

// WechatPayment 微信支付（生产 adapter，验签 + 金额校验 + 幂等更新）。
type WechatPayment struct {
	MchID    string
	APIV3Key string
	CertPath string
}

func (w *WechatPayment) Name() string { return "wechat" }

func (w *WechatPayment) CreatePayment(orderID string, amountCents int64, description string) (string, string, error) {
	// TODO: 调微信统一下单 API
	return "", "", fmt.Errorf("wechat payment 未配置真实凭证")
}

func (w *WechatPayment) VerifyCallback(payload map[string]any) (string, bool, error) {
	// TODO: 验签 + 金额校验 + 幂等
	orderID, _ := payload["order_id"].(string)
	return orderID, false, fmt.Errorf("wechat callback 验签未实现")
}

// IsMock 判断是否为 mock provider。
func IsMock(p PaymentProvider) bool {
	return p.Name() == "mock"
}
