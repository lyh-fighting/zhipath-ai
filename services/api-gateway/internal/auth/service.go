package auth

import (
	"context"
	"fmt"
)

// Service 认证服务，绑定 token 签发与微信登录。
type Service struct {
	tokens   *TokenService
	wechat   WechatProvider
	tenantID string
}

// NewService 构造认证服务。
func NewService(tokens *TokenService, wechat WechatProvider, tenantID string) *Service {
	return &Service{tokens: tokens, wechat: wechat, tenantID: tenantID}
}

// LoginByWechat 用微信 code 登录，返回 access/refresh token。
// userID 由调用方根据 openid 查找或创建用户后传入（Task 7 接入持久化）。
func (s *Service) LoginByWechat(ctx context.Context, code, userID string) (access, refresh string, err error) {
	if _, err = s.wechat.Code2Session(ctx, code); err != nil {
		return "", "", fmt.Errorf("wechat login: %w", err)
	}
	if access, err = s.tokens.IssueAccess(s.tenantID, userID); err != nil {
		return "", "", err
	}
	refresh, err = s.tokens.IssueRefresh(s.tenantID, userID)
	return access, refresh, err
}

// Refresh 用 refresh token 换发新 access/refresh token。
func (s *Service) Refresh(refreshToken string) (access, refresh string, err error) {
	claims, err := s.tokens.Parse(refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("refresh token 无效: %w", err)
	}
	if access, err = s.tokens.IssueAccess(claims.TenantID, claims.UserID); err != nil {
		return "", "", err
	}
	refresh, err = s.tokens.IssueRefresh(claims.TenantID, claims.UserID)
	return access, refresh, err
}
