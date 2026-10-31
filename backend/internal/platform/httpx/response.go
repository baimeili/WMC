// Package httpx 封装统一响应、请求上下文与通用中间件。
package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"wmc/backend/internal/platform/apperr"
	"wmc/backend/internal/platform/logx"
)

// Envelope 统一响应结构。
type Envelope struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Data      any    `json:"data,omitempty"`
	RequestID string `json:"request_id"`
}

// Page 游标分页响应。
type Page[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// 请求体大小上限：1MB（图片走独立上传接口）。
const maxJSONBody = 1 << 20

// OK 返回成功响应。
func OK(w http.ResponseWriter, r *http.Request, data any) {
	write(w, r, http.StatusOK, Envelope{Code: apperr.CodeOK, Data: data})
}

// Created 返回 201。
func Created(w http.ResponseWriter, r *http.Request, data any) {
	write(w, r, http.StatusCreated, Envelope{Code: apperr.CodeOK, Data: data})
}

// NoContent 返回 204。
func NoContent(w http.ResponseWriter) { w.WriteHeader(http.StatusNoContent) }

// Fail 返回统一错误响应。
func Fail(w http.ResponseWriter, r *http.Request, err error) {
	e := apperr.From(err)
	if e.Status >= 500 {
		logx.From(r.Context()).ErrorContext(r.Context(), "请求处理失败",
			"error_code", e.Code, "path", r.URL.Path, "error", e.Error())
	}
	write(w, r, e.Status, Envelope{Code: e.Code, Message: e.Message})
}

// DecodeJSON 解析 JSON 请求体，限制大小并拒绝未知字段。
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBody)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return apperr.Invalid("请求体不能为空")
		}
		return apperr.Invalid("请求体格式不正确").Wrap(err)
	}
	return nil
}

func write(w http.ResponseWriter, r *http.Request, status int, env Envelope) {
	env.RequestID = RequestIDFrom(r.Context())
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Request-Id", env.RequestID)
	w.WriteHeader(status)
	if status == http.StatusNoContent {
		return
	}
	if err := json.NewEncoder(w).Encode(env); err != nil {
		logx.From(r.Context()).ErrorContext(r.Context(), "写出响应失败", "error", err)
	}
}
