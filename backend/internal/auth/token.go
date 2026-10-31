// Package auth 实现手机号登录、JWT 签发与 Refresh Token 轮换。
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"wmc/backend/internal/platform/apperr"
	"wmc/backend/internal/platform/secure"
)

// Claims Access Token 载荷。
type Claims struct {
	jwt.RegisteredClaims
	Role   int16 `json:"role"`
	Status int16 `json:"status"`
}

// TokenManager 负责签发与校验 Token。
type TokenManager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
	issuer     string
	now        func() time.Time
}

// NewTokenManager 构造 Token 管理器。
func NewTokenManager(secret string, accessTTL, refreshTTL time.Duration) *TokenManager {
	return &TokenManager{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		issuer:     "wmc",
		now:        time.Now,
	}
}

// AccessTTL 返回 Access Token 有效期。
func (m *TokenManager) AccessTTL() time.Duration { return m.accessTTL }

// RefreshTTL 返回 Refresh Token 有效期。
func (m *TokenManager) RefreshTTL() time.Duration { return m.refreshTTL }

// IssueAccess 签发 Access Token。
func (m *TokenManager) IssueAccess(userID uuid.UUID, role, status int16) (string, time.Time, error) {
	now := m.now().UTC()
	exp := now.Add(m.accessTTL)
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-time.Second)),
			ExpiresAt: jwt.NewNumericDate(exp),
			ID:        uuid.NewString(),
		},
		Role:   role,
		Status: status,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("签发 Access Token 失败: %w", err)
	}
	return signed, exp, nil
}

// ParseAccess 校验并解析 Access Token。
func (m *TokenManager) ParseAccess(token string) (*Claims, error) {
	claims := &Claims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("非法的签名算法")
		}
		return m.secret, nil
	}, jwt.WithIssuer(m.issuer), jwt.WithExpirationRequired())
	if err != nil || !parsed.Valid {
		return nil, apperr.ErrUnauthorized.Wrap(err)
	}
	return claims, nil
}

// UserID 从 Claims 解析用户 ID。
func (c *Claims) UserID() (uuid.UUID, error) {
	return uuid.Parse(c.Subject)
}

// NewRefreshToken 生成不透明 Refresh Token 与其摘要。
func (m *TokenManager) NewRefreshToken() (raw string, hash []byte, err error) {
	raw, err = secure.RandomToken(32)
	if err != nil {
		return "", nil, fmt.Errorf("生成 Refresh Token 失败: %w", err)
	}
	return raw, secure.TokenHash(raw), nil
}

// HashToken 计算给定 Token 的摘要。
func HashToken(raw string) []byte { return secure.TokenHash(raw) }
