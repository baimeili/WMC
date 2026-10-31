// Package identity 在 context 中传递当前登录身份。
package identity

import (
	"context"

	"github.com/google/uuid"
)

// Principal 当前请求主体。
type Principal struct {
	UserID uuid.UUID
	Role   int16
	Status int16
}

type ctxKey struct{}

// With 写入 context。
func With(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, ctxKey{}, p)
}

// From 读取 context。
func From(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(ctxKey{}).(Principal)
	return p, ok
}
