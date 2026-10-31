package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"wmc/backend/internal/platform/apperr"
)

func newTestManager() *TokenManager { return NewTokenManager("test-secret-0123456789abcdef0123456789", 15*time.Minute, 30*24*time.Hour) }

func TestIssueAndParseAccess(t *testing.T) {
	m := newTestManager()
	uid := uuid.New()
	token, exp, err := m.IssueAccess(uid, 1, 2)
	if err != nil {
		t.Fatalf("签发失败: %v", err)
	}
	if time.Until(exp) <= 0 {
		t.Fatal("过期时间应在未来")
	}
	claims, err := m.ParseAccess(token)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	got, err := claims.UserID()
	if err != nil || got != uid {
		t.Fatalf("用户 ID 不一致: %v != %v", got, uid)
	}
	if claims.Role != 1 || claims.Status != 2 {
		t.Fatalf("角色/状态不正确: %+v", claims)
	}
}

func TestParseAccessWrongSecret(t *testing.T) {
	m := newTestManager()
	token, _, _ := m.IssueAccess(uuid.New(), 1, 2)
	other := NewTokenManager("another-secret-0123456789abcdef0123456789", time.Minute, time.Hour)
	if _, err := other.ParseAccess(token); err == nil {
		t.Fatal("不同密钥签发的 Token 不应通过校验")
	}
}

func TestParseAccessExpired(t *testing.T) {
	m := NewTokenManager("test-secret-0123456789abcdef0123456789", -time.Minute, time.Hour)
	token, _, _ := m.IssueAccess(uuid.New(), 1, 2)
	_, err := m.ParseAccess(token)
	if err == nil {
		t.Fatal("过期 Token 必须拒绝")
	}
	if e := apperr.From(err); e.Code != apperr.CodeUnauthorized {
		t.Fatalf("错误码应为 UNAUTHORIZED，实际 %s", e.Code)
	}
}

func TestRefreshTokenIsRandomAndHashed(t *testing.T) {
	m := newTestManager()
	raw1, hash1, err := m.NewRefreshToken()
	if err != nil {
		t.Fatalf("生成失败: %v", err)
	}
	raw2, hash2, _ := m.NewRefreshToken()
	if raw1 == raw2 {
		t.Fatal("Refresh Token 必须随机")
	}
	if string(hash1) == raw1 {
		t.Fatal("数据库中只保存摘要")
	}
	if string(hash1) == string(hash2) {
		t.Fatal("不同 Token 摘要不应相同")
	}
	if string(HashToken(raw1)) != string(hash1) {
		t.Fatal("摘要计算不一致")
	}
}
