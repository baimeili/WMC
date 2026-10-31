package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"wmc/backend/internal/db/sqlcdb"
	"wmc/backend/internal/platform/apperr"
	"wmc/backend/internal/platform/httpx"
	"wmc/backend/internal/platform/identity"
)

// UserLoader 按 ID 加载用户，用于校验账号状态。
type UserLoader interface {
	GetUser(ctx context.Context, id uuid.UUID) (sqlcdb.User, error)
}

// Middleware 解析 Bearer Token 并写入身份。
func Middleware(tokens *TokenManager, users UserLoader) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := bearerToken(r)
			if raw == "" {
				httpx.Fail(w, r, apperr.ErrUnauthorized)
				return
			}
			claims, err := tokens.ParseAccess(raw)
			if err != nil {
				httpx.Fail(w, r, apperr.ErrUnauthorized.WithMessage("登录已过期，请重新登录"))
				return
			}
			uid, err := claims.UserID()
			if err != nil {
				httpx.Fail(w, r, apperr.ErrUnauthorized)
				return
			}
			u, err := users.GetUser(r.Context(), uid)
			if err != nil {
				httpx.Fail(w, r, apperr.ErrUnauthorized.WithMessage("账户不存在或已被删除"))
				return
			}
			if u.Status != 2 {
				httpx.Fail(w, r, apperr.ErrForbidden.WithMessage("账户状态异常，请联系管理员"))
				return
			}
			ctx := identity.With(r.Context(), identity.Principal{UserID: u.ID, Role: u.Role, Status: u.Status})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole 要求最低角色等级。
func RequireRole(minRole int16) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, ok := identity.From(r.Context())
			if !ok {
				httpx.Fail(w, r, apperr.ErrUnauthorized)
				return
			}
			if p.Role < minRole {
				httpx.Fail(w, r, apperr.ErrForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// Optional 允许游客访问，但解析到身份时写入 context。
func Optional(tokens *TokenManager, users UserLoader) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := bearerToken(r)
			if raw != "" {
				if claims, err := tokens.ParseAccess(raw); err == nil {
					if uid, err := claims.UserID(); err == nil {
						if u, err := users.GetUser(r.Context(), uid); err == nil && u.Status == 2 {
							ctx := identity.With(r.Context(), identity.Principal{UserID: u.ID, Role: u.Role, Status: u.Status})
							next.ServeHTTP(w, r.WithContext(ctx))
							return
						}
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if h == "" {
		return ""
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
