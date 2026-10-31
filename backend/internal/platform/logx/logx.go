// Package logx 提供结构化 JSON 日志。
// 严禁在日志中输出手机号、JWT、验证码与完整个人资料。
package logx

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

// New 创建 JSON 日志器。
func New(level string) *slog.Logger {
	var lv slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lv = slog.LevelDebug
	case "warn":
		lv = slog.LevelWarn
	case "error":
		lv = slog.LevelError
	default:
		lv = slog.LevelInfo
	}
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lv})
	return slog.New(h)
}

// 敏感字段集合：任何日志属性命中即被替换。
var sensitiveKeys = map[string]struct{}{
	"phone": {}, "phone_number": {}, "mobile": {}, "token": {}, "access_token": {},
	"refresh_token": {}, "password": {}, "code": {}, "sms_code": {}, "jwt": {},
	"secret": {}, "authorization": {},
}

// Redact 对属性做脱敏（仅用于必须记录的情况）。
func Redact(attrs map[string]any) map[string]any {
	out := make(map[string]any, len(attrs))
	for k, v := range attrs {
		if _, bad := sensitiveKeys[strings.ToLower(k)]; bad {
			out[k] = "***"
			continue
		}
		out[k] = v
	}
	return out
}

type ctxKey struct{}

// WithLogger 将日志器放入 context。
func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

// From 取出日志器，缺省返回默认日志器。
func From(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok && l != nil {
		return l
	}
	return slog.Default()
}
