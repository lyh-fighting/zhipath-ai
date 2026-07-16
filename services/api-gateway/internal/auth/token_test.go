package auth_test

import (
	"testing"

	"github.com/lyh-fighting/zhipath-ai/services/api-gateway/internal/auth"
)

func TestTokenIssueAndParse(t *testing.T) {
	svc := auth.NewTokenService("test-secret")
	access, err := svc.IssueAccess("default_consumer", "u_001")
	if err != nil {
		t.Fatalf("签发 access token 失败: %v", err)
	}
	claims, err := svc.Parse(access)
	if err != nil {
		t.Fatalf("解析 access token 失败: %v", err)
	}
	if claims.UserID != "u_001" {
		t.Errorf("UserID 不匹配: got %s", claims.UserID)
	}
	if claims.TenantID != "default_consumer" {
		t.Errorf("TenantID 不匹配: got %s", claims.TenantID)
	}
}

func TestTokenInvalid(t *testing.T) {
	svc := auth.NewTokenService("test-secret")
	if _, err := svc.Parse("invalid.token.here"); err == nil {
		t.Fatal("无效 token 应返回错误")
	}
}

func TestTokenWrongSecret(t *testing.T) {
	svc1 := auth.NewTokenService("secret-a")
	svc2 := auth.NewTokenService("secret-b")
	tok, _ := svc1.IssueAccess("t", "u")
	if _, err := svc2.Parse(tok); err == nil {
		t.Fatal("不同 secret 签发的 token 应解析失败")
	}
}

func TestRefreshTokenRoundTrip(t *testing.T) {
	svc := auth.NewTokenService("s")
	refresh, _ := svc.IssueRefresh("default_consumer", "u_002")
	claims, err := svc.Parse(refresh)
	if err != nil {
		t.Fatalf("refresh token 解析失败: %v", err)
	}
	if claims.UserID != "u_002" {
		t.Errorf("UserID 不匹配: got %s", claims.UserID)
	}
}
