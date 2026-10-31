package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"wmc/backend/internal/platform/apperr"
	"wmc/backend/internal/platform/redisx"
	"wmc/backend/internal/platform/secure"
	"wmc/backend/internal/settings"
	"wmc/backend/internal/sms"
	"wmc/backend/internal/store"
)

// SMSService 短信验证码登录。
type SMSService struct {
	auth     *Service
	store    *store.Store
	sms      sms.Provider
	kv       redisx.Store
	settings *settings.Service
}

// NewSMSService 构造短信登录服务。
func NewSMSService(authSvc *Service, st *store.Store, provider sms.Provider, kv redisx.Store, set *settings.Service) *SMSService {
	return &SMSService{auth: authSvc, store: st, sms: provider, kv: kv, settings: set}
}

// 验证码有效期与错误次数上限。
const (
	codeTTL        = 5 * time.Minute
	maxVerifyTries = 10
)

// NormalizePhone 校验并规范化中国大陆手机号。
func NormalizePhone(phone string) (string, error) {
	p := strings.TrimSpace(strings.ReplaceAll(phone, " ", ""))
	p = strings.TrimPrefix(p, "+86")
	if len(p) != 11 || !strings.HasPrefix(p, "1") {
		return "", apperr.Invalid("请输入正确的手机号")
	}
	for _, ch := range p {
		if ch < '0' || ch > '9' {
			return "", apperr.Invalid("请输入正确的手机号")
		}
	}
	return p, nil
}

func (s *SMSService) codeKey(phone string) string {
	return "sms:code:" + string(secure.PhoneHash(s.auth.hashKey, phone)[:16])
}

// SendCode 发送验证码（含冷却与每日上限）。
func (s *SMSService) SendCode(ctx context.Context, phone, ip string) error {
	normalized, err := NormalizePhone(phone)
	if err != nil {
		return err
	}
	// 同一手机号 60 秒冷却
	resend := time.Duration(s.settings.Int(ctx, settings.KeySMSResendSeconds, 60)) * time.Second
	ok, err := s.kv.SetNX(ctx, "sms:cooldown:"+normalized, "1", resend)
	if err == nil && !ok {
		return apperr.ErrRateLimited.WithMessage("验证码发送过于频繁，请稍后再试")
	}
	// 每日次数上限
	dailyLimit := s.settings.Int(ctx, settings.KeySMSDailyLimit, 5)
	dayKey := "sms:daily:" + normalized + ":" + time.Now().Format("20060102")
	count, err := s.kv.Incr(ctx, dayKey, 24*time.Hour)
	if err == nil && count > int64(dailyLimit) {
		return apperr.ErrRateLimited.WithMessage("今日验证码次数已达上限")
	}

	code := sms.GenerateCode(6)
	if mock, ok := s.sms.(*sms.MockProvider); ok && mock.FixedCode() != "" {
		code = mock.FixedCode()
	}
	if err := s.kv.Set(ctx, s.codeKey(normalized), code, codeTTL); err != nil {
		return apperr.ErrInternal.Wrap(err)
	}
	// IP 维度限流
	ipKey := "sms:ip:" + ip + ":" + time.Now().Format("2006010215")
	if n, err := s.kv.Incr(ctx, ipKey, time.Hour); err == nil && n > 30 {
		return apperr.ErrRateLimited.WithMessage("请求过于频繁")
	}
	return s.sms.SendVerificationCode(ctx, normalized, code)
}

// Verify 校验验证码；首次登录自动注册。
func (s *SMSService) Verify(ctx context.Context, phone, code, userAgent, ip string) (*TokenPair, error) {
	normalized, err := NormalizePhone(phone)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(code) == "" {
		return nil, apperr.Invalid("请输入验证码")
	}
	// 错误次数限制
	if n, err := s.kv.Incr(ctx, "sms:tries:"+normalized, 10*time.Minute); err == nil && n > maxVerifyTries {
		return nil, apperr.ErrRateLimited.WithMessage("验证码尝试次数过多，请稍后重试")
	}

	stored, ok, err := s.kv.Get(ctx, s.codeKey(normalized))
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	if !ok || stored == "" {
		return nil, apperr.Invalid("验证码已过期，请重新获取")
	}
	if stored != code {
		return nil, apperr.Invalid("验证码不正确")
	}
	_ = s.kv.Del(ctx, s.codeKey(normalized))

	user, err := s.auth.FindByPhone(ctx, normalized)
	if errors.Is(err, apperr.ErrNotFound) {
		if !s.settings.Bool(ctx, settings.KeyRegistrationEnabled, true) {
			return nil, apperr.ErrFeatureDisabled.WithMessage("注册暂未开放")
		}
		created, createErr := s.auth.CreateUser(ctx, normalized, 1, false)
		if createErr != nil {
			return nil, createErr
		}
		user = created
	} else if err != nil {
		return nil, err
	}
	if user.Status == 3 {
		return nil, apperr.ErrForbidden.WithMessage("账号已被冻结，请联系管理员")
	}
	if user.Status == 4 || user.Status == 5 {
		return nil, apperr.ErrForbidden.WithMessage("账号已注销")
	}
	if user.Status == 1 {
		// 首次验证通过即激活
		if err := s.store.Q.UpdateUserStatus(ctx, updateStatusParams(user.ID, 2)); err != nil {
			return nil, apperr.ErrInternal.Wrap(err)
		}
		user.Status = 2
	}
	if _, err := s.store.Q.GetUserByID(ctx, user.ID); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	_ = s.store.Q.UpdateUserLastLogin(ctx, user.ID)
	_ = s.auth.VerifyPhone(ctx, user.ID)
	return s.auth.IssueForUser(ctx, user, userAgent, ip)
}

// FormatPhoneTail 输出手机号尾号（日志使用）。
func FormatPhoneTail(phone string) string {
	if len(phone) <= 4 {
		return "***"
	}
	return fmt.Sprintf("****%s", phone[len(phone)-4:])
}
