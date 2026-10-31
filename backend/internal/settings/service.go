// Package settings 提供系统参数与 Feature Flag 的读取。
// 参数全部来自 system_settings 表，修改后无需重新发版。
package settings

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"
)

// 系统参数键名常量。
const (
	KeyCalculatorEnabled        = "calculator_enabled"
	KeyRegistrationEnabled      = "registration_enabled"
	KeyDiscoveryEnabled         = "discovery_enabled"
	KeyLikesEnabled             = "likes_enabled"
	KeyMatchesEnabled           = "matches_enabled"
	KeyChatEnabled              = "chat_enabled"
	KeyPaymentsEnabled          = "payments_enabled"
	KeyVerificationEnabled      = "verification_enabled"
	KeyMinProfileCompletion     = "min_profile_completion"
	KeyKAnonymityThreshold      = "k_anonymity_threshold"
	KeyPreferenceWeight         = "preference_score_weight"
	KeyCompatibilityWeight      = "compatibility_score_weight"
	KeyMinValidQuestions        = "min_valid_compatibility_questions"
	KeyDeletionRetentionDays    = "account_deletion_retention_days"
	KeyPolicyTerms              = "policy_version_terms"
	KeyPolicyPrivacy            = "policy_version_privacy"
	KeyPolicyMarket             = "policy_version_market"
	KeySMSResendSeconds         = "sms_resend_seconds"
	KeySMSDailyLimit            = "sms_daily_limit"
	KeyCandidatePageSize        = "candidate_page_size"
	KeyMaxCandidatePageSize     = "max_candidate_page_size"
	KeyDemoDatasetEnabled       = "demo_dataset_enabled"
)

// FeatureKeys 对外暴露的 Feature Flag 列表。
var FeatureKeys = []string{
	KeyCalculatorEnabled, KeyRegistrationEnabled, KeyDiscoveryEnabled, KeyLikesEnabled,
	KeyMatchesEnabled, KeyChatEnabled, KeyPaymentsEnabled, KeyVerificationEnabled,
}

// 默认值（数据库缺省时兜底）。
var defaults = map[string]any{
	KeyCalculatorEnabled:     true,
	KeyRegistrationEnabled:   true,
	KeyDiscoveryEnabled:      true,
	KeyLikesEnabled:          true,
	KeyMatchesEnabled:        true,
	KeyChatEnabled:           false,
	KeyPaymentsEnabled:       false,
	KeyVerificationEnabled:   false,
	KeyMinProfileCompletion:  80,
	KeyKAnonymityThreshold:   10,
	KeyPreferenceWeight:      0.4,
	KeyCompatibilityWeight:   0.6,
	KeyMinValidQuestions:     6,
	KeyDeletionRetentionDays: 30,
	KeySMSResendSeconds:      60,
	KeySMSDailyLimit:         5,
	KeyCandidatePageSize:     20,
	KeyMaxCandidatePageSize:  50,
	KeyDemoDatasetEnabled:    true,
}

// Loader 从存储加载全部参数。
type Loader interface {
	LoadSettings(ctx context.Context) (map[string]json.RawMessage, error)
}

// Service 带缓存的参数服务。
type Service struct {
	loader Loader
	ttl    time.Duration
	now    func() time.Time

	mu      sync.RWMutex
	cache   map[string]json.RawMessage
	loadedAt time.Time
}

// NewService 构造参数服务。
func NewService(loader Loader, ttl time.Duration) *Service {
	if ttl <= 0 {
		ttl = 15 * time.Second
	}
	return &Service{loader: loader, ttl: ttl, now: time.Now}
}

// SetClock 测试注入时钟。
func (s *Service) SetClock(now func() time.Time) { s.now = now }

// Invalidate 使缓存失效（后台修改参数后调用）。
func (s *Service) Invalidate() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache = nil
	s.loadedAt = time.Time{}
}

// Snapshot 返回当前参数快照。
func (s *Service) Snapshot(ctx context.Context) (map[string]json.RawMessage, error) {
	s.mu.RLock()
	if s.cache != nil && s.now().Sub(s.loadedAt) < s.ttl {
		out := s.cache
		s.mu.RUnlock()
		return out, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cache != nil && s.now().Sub(s.loadedAt) < s.ttl {
		return s.cache, nil
	}
	loaded, err := s.loader.LoadSettings(ctx)
	if err != nil {
		if s.cache != nil {
			return s.cache, nil // 加载失败时使用旧缓存，保证核心流程可用
		}
		return nil, fmt.Errorf("加载系统参数失败: %w", err)
	}
	s.cache = loaded
	s.loadedAt = s.now()
	return s.cache, nil
}

func (s *Service) raw(ctx context.Context, key string) any {
	snap, err := s.Snapshot(ctx)
	if err != nil || snap == nil {
		return nil
	}
	v, ok := snap[key]
	if !ok {
		return nil
	}
	var out any
	if err := json.Unmarshal(v, &out); err != nil {
		return nil
	}
	return out
}

// Bool 读取布尔参数。
func (s *Service) Bool(ctx context.Context, key string, def bool) bool {
	v := s.raw(ctx, key)
	switch x := v.(type) {
	case bool:
		return x
	case string:
		if b, err := strconv.ParseBool(x); err == nil {
			return b
		}
	case float64:
		return x != 0
	}
	if d, ok := defaults[key].(bool); ok {
		return d
	}
	return def
}

// Int 读取整数参数。
func (s *Service) Int(ctx context.Context, key string, def int) int {
	v := s.raw(ctx, key)
	switch x := v.(type) {
	case float64:
		return int(x)
	case string:
		if n, err := strconv.Atoi(x); err == nil {
			return n
		}
	}
	if d, ok := defaults[key].(int); ok {
		return d
	}
	return def
}

// Float 读取浮点参数。
func (s *Service) Float(ctx context.Context, key string, def float64) float64 {
	v := s.raw(ctx, key)
	switch x := v.(type) {
	case float64:
		return x
	case string:
		if f, err := strconv.ParseFloat(x, 64); err == nil {
			return f
		}
	}
	if d, ok := defaults[key].(float64); ok {
		return d
	}
	return def
}

// String 读取字符串参数。
func (s *Service) String(ctx context.Context, key, def string) string {
	if v := s.raw(ctx, key); v != nil {
		if str, ok := v.(string); ok {
			return str
		}
	}
	return def
}

// Features 返回全部 Feature Flag。
func (s *Service) Features(ctx context.Context) map[string]bool {
	out := make(map[string]bool, len(FeatureKeys))
	for _, k := range FeatureKeys {
		out[k] = s.Bool(ctx, k, true)
	}
	return out
}

// PublicConfig 返回给前端的公开配置。
func (s *Service) PublicConfig(ctx context.Context) map[string]any {
	settings := map[string]any{
		"k_anonymity_threshold":      s.Int(ctx, KeyKAnonymityThreshold, 10),
		"min_profile_completion":     s.Int(ctx, KeyMinProfileCompletion, 80),
		"candidate_page_size":        s.Int(ctx, KeyCandidatePageSize, 20),
		"max_candidate_page_size":    s.Int(ctx, KeyMaxCandidatePageSize, 50),
		"sms_resend_seconds":         s.Int(ctx, KeySMSResendSeconds, 60),
		"min_valid_compatibility_questions": s.Int(ctx, KeyMinValidQuestions, 6),
	}
	return map[string]any{
		"features": s.Features(ctx),
		"settings": settings,
		"policy_versions": map[string]string{
			"terms_of_service": s.String(ctx, KeyPolicyTerms, "2026.01"),
			"privacy_policy":   s.String(ctx, KeyPolicyPrivacy, "2026.01"),
			"market_participation": s.String(ctx, KeyPolicyMarket, "2026.01"),
		},
	}
}
