package calculator

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"wmc/backend/internal/platform/apperr"
	"wmc/backend/internal/platform/httpx"
	"wmc/backend/internal/platform/identity"
)

// Handler 计算器 HTTP 处理器。
type Handler struct {
	svc *Service
}

// NewHandler 构造处理器。
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Register 注册路由（游客可访问）。
func (h *Handler) Register(r chi.Router) {
	r.Post("/calculator/estimate", h.estimate)
	r.Post("/calculator/share", h.share)
	r.Get("/calculator/share/{token}", h.getShare)
}

type shareRequest struct {
	RunID              string `json:"run_id"`
	AnonymousSessionID string `json:"anonymous_session_id"`
}

func (h *Handler) estimate(w http.ResponseWriter, r *http.Request) {
	var req EstimateRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	var userID *uuid.UUID
	if p, ok := identity.From(r.Context()); ok {
		userID = &p.UserID
	}
	res, err := h.svc.Estimate(r.Context(), req, userID)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, res)
}

func (h *Handler) share(w http.ResponseWriter, r *http.Request) {
	var req shareRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	runID, err := uuid.Parse(req.RunID)
	if err != nil {
		httpx.Fail(w, r, apperr.Invalid("run_id 格式不正确"))
		return
	}
	var userID *uuid.UUID
	if p, ok := identity.From(r.Context()); ok {
		userID = &p.UserID
	}
	token, err := h.svc.CreateShare(r.Context(), runID, req.AnonymousSessionID, userID)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, map[string]string{"share_token": token})
}

func (h *Handler) getShare(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	res, err := h.svc.GetShare(r.Context(), token)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, res)
}
