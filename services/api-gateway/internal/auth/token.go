package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT 载荷，携带 tenant_id 与 user_id。
// 公网接口的身份来自此 token，不信任请求体中的 user_id。
type Claims struct {
	TenantID string `json:"tenant_id"`
	UserID   string `json:"user_id"`
	jwt.RegisteredClaims
}

// TokenService 签发与解析 access/refresh token。
type TokenService struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewTokenService 用 secret 构造 token 服务。
func NewTokenService(secret string) *TokenService {
	return &TokenService{
		secret:     []byte(secret),
		accessTTL:  2 * time.Hour,
		refreshTTL: 14 * 24 * time.Hour,
	}
}

// IssueAccess 签发 access token。
func (s *TokenService) IssueAccess(tenantID, userID string) (string, error) {
	return s.issue(tenantID, userID, s.accessTTL)
}

// IssueRefresh 签发 refresh token。
func (s *TokenService) IssueRefresh(tenantID, userID string) (string, error) {
	return s.issue(tenantID, userID, s.refreshTTL)
}

func (s *TokenService) issue(tenantID, userID string, ttl time.Duration) (string, error) {
	claims := Claims{
		TenantID: tenantID,
		UserID:   userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(s.secret)
}

// Parse 解析并校验 token。
func (s *TokenService) Parse(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	tok, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("非预期签名方法: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !tok.Valid {
		return nil, fmt.Errorf("token 无效")
	}
	return claims, nil
}
