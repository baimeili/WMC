package ratelimit

import (
	"net/http"
	"time"

	"wmc/backend/internal/platform/apperr"
	"wmc/backend/internal/platform/httpx"
)

// Middleware 基于客户端 IP 的限流中间件。
func (l *Limiter) Middleware(scope string, limit int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := scope + ":" + httpx.ClientIP(r)
			ok, _, err := l.Allow(r.Context(), key, limit, window)
			if err == nil && !ok {
				httpx.Fail(w, r, apperr.ErrRateLimited)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
