package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"wmc/backend/internal/db/sqlcdb"
	"wmc/backend/internal/platform/apperr"
	"wmc/backend/internal/platform/secure"
	"wmc/backend/internal/store"
)

// TokenPair 返回给客户端的令牌。
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int64     `json:"expires_in"`
	UserID       uuid.UUID `json:"user_id"`
	Role         int16     `json:"role"`
	Status       int16     `json:"status"`
}

// Service 认证服务。
type Service struct {
	store    *store.Store
	tokens   *TokenManager
	phoneKey []byte
	hashKey  string
}

// NewService 构造认证服务。
func NewService(st *store.Store, tokens *TokenManager, phoneHashSecret string, phoneKey []byte) *Service {
	return &Service{store: st, tokens: tokens, phoneKey: phoneKey, hashKey: phoneHashSecret}
}

// PhoneHash 计算手机号摘要。
func (s *Service) PhoneHash(phone string) []byte { return secure.PhoneHash(s.hashKey, phone) }

// FindByPhone 按手机号查找用户（不存在返回 apperr.ErrNotFound）。
func (s *Service) FindByPhone(ctx context.Context, phone string) (sqlcdb.User, error) {
	u, err := s.store.Q.GetUserByPhoneHash(ctx, s.PhoneHash(phone))
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlcdb.User{}, apperr.ErrNotFound
	}
	if err != nil {
		return sqlcdb.User{}, apperr.ErrInternal.Wrap(err)
	}
	return u, nil
}

// IssueForUser 为指定用户签发令牌对（含 Refresh Token 落库）。
func (s *Service) IssueForUser(ctx context.Context, u sqlcdb.User, userAgent, ip string) (*TokenPair, error) {
	access, exp, err := s.tokens.IssueAccess(u.ID, u.Role, u.Status)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	raw, hash, err := s.tokens.NewRefreshToken()
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	_, err = s.store.Q.CreateRefreshToken(ctx, sqlcdb.CreateRefreshTokenParams{
		UserID:    u.ID,
		TokenHash: hash,
		FamilyID:  uuid.New(),
		UserAgent: userAgent,
		IpHash:    secure.TokenHash(ip),
		ExpiresAt: toTimestamptz(time.Now().Add(s.tokens.RefreshTTL())),
	})
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	return &TokenPair{
		AccessToken:  access,
		RefreshToken: raw,
		ExpiresIn:    int64(time.Until(exp).Seconds()),
		UserID:       u.ID,
		Role:         u.Role,
		Status:       u.Status,
	}, nil
}

// Refresh 轮换 Refresh Token。检测到重用时吊销整个家族。
func (s *Service) Refresh(ctx context.Context, rawToken, userAgent, ip string) (*TokenPair, error) {
	hash := HashToken(rawToken)
	rec, err := s.store.Q.GetRefreshTokenByHash(ctx, hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperr.ErrUnauthorized.WithMessage("登录状态已失效，请重新登录")
	}
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	if rec.RevokedAt.Valid {
		// Token 已被使用或吊销：判定为重放攻击，吊销整个家族
		if revokeErr := s.store.Q.RevokeRefreshFamily(ctx, rec.FamilyID); revokeErr != nil {
			return nil, apperr.ErrInternal.Wrap(revokeErr)
		}
		return nil, apperr.ErrUnauthorized.WithMessage("检测到重复使用的登录凭证，请重新登录")
	}
	if time.Now().After(rec.ExpiresAt.Time) {
		return nil, apperr.ErrUnauthorized.WithMessage("登录已过期，请重新登录")
	}

	u, err := s.store.Q.GetUserByID(ctx, rec.UserID)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	if u.Status != 2 {
		return nil, apperr.ErrForbidden.WithMessage("账户状态异常，无法继续登录")
	}

	access, exp, err := s.tokens.IssueAccess(u.ID, u.Role, u.Status)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	newRaw, newHash, err := s.tokens.NewRefreshToken()
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	created, err := s.store.Q.CreateRefreshToken(ctx, sqlcdb.CreateRefreshTokenParams{
		UserID:    u.ID,
		TokenHash: newHash,
		FamilyID:  rec.FamilyID,
		UserAgent: userAgent,
		IpHash:    secure.TokenHash(ip),
		ExpiresAt: toTimestamptz(time.Now().Add(s.tokens.RefreshTTL())),
	})
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	if err := s.store.Q.MarkRefreshTokenReplaced(ctx, sqlcdb.MarkRefreshTokenReplacedParams{
		ID: rec.ID, ReplacedBy: &created.ID,
	}); err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	return &TokenPair{
		AccessToken:  access,
		RefreshToken: newRaw,
		ExpiresIn:    int64(time.Until(exp).Seconds()),
		UserID:       u.ID,
		Role:         u.Role,
		Status:       u.Status,
	}, nil
}

// Logout 吊销指定 Refresh Token。
func (s *Service) Logout(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}
	rec, err := s.store.Q.GetRefreshTokenByHash(ctx, HashToken(rawToken))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return apperr.ErrInternal.Wrap(err)
	}
	return s.store.Q.RevokeRefreshFamily(ctx, rec.FamilyID)
}

// DevLogin 开发环境管理员登录：要求该手机号已是 moderator 及以上角色。
func (s *Service) DevLogin(ctx context.Context, phone, userAgent, ip string) (*TokenPair, error) {
	u, err := s.FindByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}
	if u.Role < 2 {
		return nil, apperr.ErrForbidden.WithMessage("该账号不是管理员")
	}
	if u.Status != 2 {
		return nil, apperr.ErrForbidden.WithMessage("账号状态不可用")
	}
	_ = s.store.Q.UpdateUserLastLogin(ctx, u.ID)
	pair, err := s.IssueForUser(ctx, u, userAgent, ip)
	if err != nil {
		return nil, err
	}
	return pair, nil
}

// CreateUser 创建手机号用户。
func (s *Service) CreateUser(ctx context.Context, phone string, role int16, isSeed bool) (sqlcdb.User, error) {
	enc, err := secure.EncryptPhone(s.phoneKey, phone)
	if err != nil {
		return sqlcdb.User{}, apperr.ErrInternal.Wrap(err)
	}
	u, err := s.store.Q.CreateUser(ctx, sqlcdb.CreateUserParams{
		PhoneHash:      s.PhoneHash(phone),
		PhoneEncrypted: enc,
		Status:         2,
		Role:           role,
		IsSeed:         isSeed,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return sqlcdb.User{}, apperr.ErrConflict.WithMessage("该手机号已注册")
		}
		return sqlcdb.User{}, apperr.ErrInternal.Wrap(err)
	}
	return u, nil
}

// MaskPhone 手机号脱敏。
func MaskPhone(phone string) string { return secure.MaskPhone(phone) }

func isUniqueViolation(err error) bool {
	return err != nil && (contains(err.Error(), "duplicate key") || contains(err.Error(), "23505"))
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})()
}

// GetUser 读取用户。
func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (sqlcdb.User, error) {
	u, err := s.store.Q.GetUserByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlcdb.User{}, apperr.ErrNotFound.WithMessage("用户不存在")
	}
	if err != nil {
		return sqlcdb.User{}, apperr.ErrInternal.Wrap(err)
	}
	return u, nil
}

// VerifyPhone 记录手机号验证通过。
func (s *Service) VerifyPhone(ctx context.Context, userID uuid.UUID) error {
	_, err := s.store.Q.UpsertVerification(ctx, sqlcdb.UpsertVerificationParams{
		UserID:     userID,
		Type:       1,
		Status:     3,
		Evidence:   []byte("{}"),
		VerifiedAt: toTimestamptz(time.Now()),
	})
	if err != nil {
		return apperr.ErrInternal.Wrap(err)
	}
	return nil
}

// FormatError 便于日志输出（不包含敏感信息）。
func FormatError(err error) string {
	if err == nil {
		return ""
	}
	var e *apperr.Error
	if errors.As(err, &e) {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	}
	return "internal error"
}