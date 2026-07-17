package payment

import "testing"

func TestMockPaymentName(t *testing.T) {
	m := &MockPayment{}
	if m.Name() != "mock" {
		t.Error("mock provider 应显式标记 name=mock")
	}
	if !IsMock(m) {
		t.Error("IsMock 应返回 true")
	}
}

func TestMockPaymentCreate(t *testing.T) {
	m := &MockPayment{}
	pid, url, err := m.CreatePayment("o_001", 9900, "深度报告")
	if err != nil {
		t.Fatal(err)
	}
	if pid == "" || url == "" {
		t.Error("mock 应返回 payment_id 和 pay_url")
	}
}

func TestMockCallback(t *testing.T) {
	m := &MockPayment{}
	orderID, paid, err := m.VerifyCallback(map[string]any{"order_id": "o_001"})
	if err != nil || !paid {
		t.Fatal("mock callback 应返回 paid=true")
	}
	if orderID != "o_001" {
		t.Error("orderID 不匹配")
	}
}

func TestWechatNotMock(t *testing.T) {
	w := &WechatPayment{}
	if IsMock(w) {
		t.Error("wechat 不应是 mock")
	}
	_, _, err := w.CreatePayment("o_1", 100, "x")
	if err == nil {
		t.Error("未配置凭证的 wechat 应返回错误")
	}
}
