package auth

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"wmc/backend/internal/platform/apperr"
	"wmc/backend/internal/platform/httpx"
)

// Handler 认证处理器。
type Handler struct {
	svc        *Service
	sms        *SMSService
	devLoginOK bool
}

// NewHandler 构造处理器。
func NewHandler(svc *Service, smsSvc *SMSService, devLoginEnabled bool) *Handler {
	return &Handler{svc: svc, sms: smsSvc, devLoginOK: devLoginEnabled}
}

// Register 注册路由。
func (h *Handler) Register(r chi.Router) {
	r.Post("/admin/auth/dev-login", h.devLogin)
	r.Post("/auth/refresh", h.refresh)
	r.Post("/auth/logout", h.logout)
	r.Post("/auth/sms/send", h.sendCode)
	r.Post("/auth/sms/verify", h.verifyCode)
}

type sendCodeRequest struct {
	Phone string `json:"phone"`
}

func (h *Handler) sendCode(w http.ResponseWriter, r *http.Request) {
	var req sendCodeRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	if err := h.sms.SendCode(r.Context(), req.Phone, httpx.ClientIP(r)); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, map[string]string{"status": "sent"})
}

type verifyCodeRequest struct {
	Phone              string `json:"phone"`
	Code               string `json:"code"`
	AnonymousSessionID string `json:"anonymous_session_id"`
}

func (h *Handler) verifyCode(w http.ResponseWriter, r *http.Request) {
	var req verifyCodeRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	pair, err := h.sms.Verify(r.Context(), req.Phone, req.Code, r.UserAgent(), httpx.ClientIP(r))
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, pair)
}

type devLoginRequest struct {
	Phone string `json:"phone"`
}

func (h *Handler) devLogin(w http.ResponseWriter, r *http.Request) {
	if !h.devLoginOK {
		httpx.Fail(w, r, apperr.ErrNotFound)
		return
	}
	var req devLoginRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	if req.Phone == "" {
		httpx.Fail(w, r, apperr.Invalid("手机号不能为空"))
		return
	}
	pair, err := h.svc.DevLogin(r.Context(), req.Phone, r.UserAgent(), httpx.ClientIP(r))
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, pair)
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	if req.RefreshToken == "" {
		httpx.Fail(w, r, apperr.Invalid("refresh_token 不能为空"))
		return
	}
	pair, err := h.svc.Refresh(r.Context(), req.RefreshToken, r.UserAgent(), httpx.ClientIP(r))
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, pair)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	_ = httpx.DecodeJSON(w, r, &req)
	if err := h.svc.Logout(r.Context(), req.RefreshToken); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, map[string]string{"status": "logged_out"})
}
