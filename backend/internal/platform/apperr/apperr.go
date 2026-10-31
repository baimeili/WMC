// Package apperr 定义统一业务错误。
// 每个错误都携带稳定的错误码与 HTTP 状态码，供 API 统一输出。
package apperr

import (
	"errors"
	"fmt"
	"net/http"
)

// 错误码常量（对外契约，前端依赖这些字符串）。
const (
	CodeOK                 = "OK"
	CodeInvalidArgument    = "INVALID_ARGUMENT"
	CodeUnauthorized       = "UNAUTHORIZED"
	CodeForbidden          = "FORBIDDEN"
	CodeNotFound           = "NOT_FOUND"
	CodeConflict           = "CONFLICT"
	CodeProfileIncomplete  = "PROFILE_INCOMPLETE"
	CodeDatasetUnavailable = "DATASET_UNAVAILABLE"
	CodeRateLimited        = "RATE_LIMITED"
	CodeFeatureDisabled    = "FEATURE_DISABLED"
	CodeInternal           = "INTERNAL"
)

// Error 业务错误。
type Error struct {
	Code    string
	Message string
	Status  int
	cause   error
}

func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap 支持 errors.Is / errors.As。
func (e *Error) Unwrap() error { return e.cause }

// Wrap 附加底层原因（不对外输出）。
func (e *Error) Wrap(err error) *Error {
	return &Error{Code: e.Code, Message: e.Message, Status: e.Status, cause: err}
}

// WithMessage 替换对外文案。
func (e *Error) WithMessage(msg string) *Error {
	return &Error{Code: e.Code, Message: msg, Status: e.Status, cause: e.cause}
}

// New 构造错误。
func New(code, message string, status int) *Error {
	return &Error{Code: code, Message: message, Status: status}
}

// 预定义错误。
var (
	ErrUnauthorized       = New(CodeUnauthorized, "请先登录", http.StatusUnauthorized)
	ErrForbidden          = New(CodeForbidden, "没有权限执行该操作", http.StatusForbidden)
	ErrNotFound           = New(CodeNotFound, "资源不存在", http.StatusNotFound)
	ErrConflict           = New(CodeConflict, "当前状态不允许该操作", http.StatusConflict)
	ErrProfileIncomplete  = New(CodeProfileIncomplete, "请先完善个人资料", http.StatusConflict)
	ErrDatasetUnavailable = New(CodeDatasetUnavailable, "暂无官方基准数据集", http.StatusConflict)
	ErrRateLimited        = New(CodeRateLimited, "操作过于频繁，请稍后再试", http.StatusTooManyRequests)
	ErrInternal           = New(CodeInternal, "服务暂时不可用", http.StatusInternalServerError)
	ErrFeatureDisabled    = New(CodeFeatureDisabled, "该功能暂未开放", http.StatusForbidden)
)

// Invalid 参数错误。
func Invalid(message string) *Error {
	return New(CodeInvalidArgument, message, http.StatusBadRequest)
}

// From 将任意错误转换为 *Error。
func From(err error) *Error {
	if err == nil {
		return nil
	}
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	return ErrInternal.Wrap(err)
}
