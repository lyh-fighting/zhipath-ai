package auth

import "context"

// WechatLoginResult 微信 code2session 返回。
type WechatLoginResult struct {
	OpenID  string
	UnionID string
}

// WechatProvider 微信登录 provider。
// 生产环境调微信 API；本地使用 MockWechatProvider，显式标记，不静默伪造。
type WechatProvider interface {
	Code2Session(ctx context.Context, code string) (*WechatLoginResult, error)
}

// MockWechatProvider 本地 mock，openid 由 code 派生（确定性，便于测试）。
type MockWechatProvider struct{}

func NewMockWechatProvider() *MockWechatProvider { return &MockWechatProvider{} }

func (m *MockWechatProvider) Code2Session(_ context.Context, code string) (*WechatLoginResult, error) {
	return &WechatLoginResult{
		OpenID:  "mock_openid_" + code,
		UnionID: "mock_unionid_" + code,
	}, nil
}
