// Package sms 抽象短信发送能力。
// 开发环境使用 mock，生产环境必须接入真实供应商（本包预留统一接口）。
package sms

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"strings"
)

// Provider 短信供应商抽象。
type Provider interface {
	// Name 返回供应商名称。
	Name() string
	// SendVerificationCode 发送验证码。
	SendVerificationCode(ctx context.Context, phone, code string) error
}

// MockProvider 开发环境实现：不真正发送，只记录脱敏日志。
type MockProvider struct {
	fixedCode string
	logger    *slog.Logger
}

// NewMock 创建 mock 供应商。
func NewMock(fixedCode string, logger *slog.Logger) *MockProvider {
	return &MockProvider{fixedCode: fixedCode, logger: logger}
}

// Name 实现 Provider。
func (m *MockProvider) Name() string { return "mock" }

// SendVerificationCode 实现 Provider。日志中禁止出现完整手机号与验证码。
func (m *MockProvider) SendVerificationCode(_ context.Context, phone, _ string) error {
	if m.logger != nil {
		m.logger.Info("模拟短信已发送", "phone_suffix", tail(phone, 4), "provider", "mock")
	}
	return nil
}

// FixedCode 返回 mock 固定验证码（仅开发环境可用）。
func (m *MockProvider) FixedCode() string { return m.fixedCode }

func tail(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}

// UnsupportedProvider 尚未接入的供应商占位实现。
type UnsupportedProvider struct{ name string }

// Name 实现 Provider。
func (u *UnsupportedProvider) Name() string { return u.name }

// SendVerificationCode 实现 Provider。
func (u *UnsupportedProvider) SendVerificationCode(context.Context, string, string) error {
	return fmt.Errorf("短信供应商 %s 尚未接入", u.name)
}

// New 按配置创建供应商。生产环境禁止 mock（由 config 校验兜底）。
func New(name, mockCode string, logger *slog.Logger) (Provider, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "mock":
		return NewMock(mockCode, logger), nil
	case "":
		return nil, fmt.Errorf("SMS_PROVIDER 不能为空")
	default:
		return &UnsupportedProvider{name: name}, nil
	}
}

const codeAlphabet = "0123456789"

// GenerateCode 生成 n 位数字验证码。
func GenerateCode(n int) string {
	if n <= 0 {
		n = 6
	}
	b := make([]byte, n)
	for i := range b {
		b[i] = codeAlphabet[rand.Intn(len(codeAlphabet))]
	}
	return string(b)
}
