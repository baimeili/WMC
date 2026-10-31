// Package ratelimit 提供基于 Store 的计数限流。
package ratelimit

import (
	"context"
	"time"

	"wmc/backend/internal/platform/apperr"
	"wmc/backend/internal/platform/redisx"
)

// Limiter 固定窗口计数器限流器。
type Limiter struct {
	store redisx.Store
}

// New 创建限流器。
func New(store redisx.Store) *Limiter { return &Limiter{store: store} }

// Allow 在 window 内最多允许 limit 次；返回是否放行与当前计数。
func (l *Limiter) Allow(ctx context.Context, scope string, limit int, window time.Duration) (bool, int64, error) {
	if limit <= 0 {
		return true, 0, nil
	}
	key := "rl:" + scope + ":" + time.Now().UTC().Format("200601021504")
	n, err := l.store.Incr(ctx, key, window+time.Minute)
	if err != nil {
		// 限流存储不可用时放行，避免影响主流程
		return true, 0, err
	}
	return n <= int64(limit), n, nil
}

// MustAllow 超限时返回 RATE_LIMITED 错误。
func (l *Limiter) MustAllow(ctx context.Context, scope string, limit int, window time.Duration) error {
	ok, _, err := l.Allow(ctx, scope, limit, window)
	if err != nil {
		return nil
	}
	if !ok {
		return apperr.ErrRateLimited
	}
	return nil
}

// Cooldown 判断某 scope+subject 是否仍在冷却期；若可用则占用一次。
func (l *Limiter) Cooldown(ctx context.Context, scope, subject string, d time.Duration) (bool, error) {
	if d <= 0 {
		return true, nil
	}
	ok, err := l.store.SetNX(ctx, "cd:"+scope+":"+subject, "1", d)
	if err != nil {
		return true, err
	}
	return ok, nil
}
