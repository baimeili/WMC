package discovery

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"wmc/backend/internal/platform/apperr"
	"wmc/backend/internal/platform/httpx"
	"wmc/backend/internal/platform/identity"
)

// Handler 候选池接口。
type Handler struct{ svc *Service }

// NewHandler 构造处理器。
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Register 注册需要登录的路由。
func (h *Handler) Register(r chi.Router) {
	r.Get("/discovery/summary", h.summary)
	r.Get("/discovery/funnel", h.funnel)
	r.Get("/discovery/sensitivity", h.sensitivity)
	r.Get("/discovery/candidates", h.candidates)
	r.Get("/discovery/candidates/{id}", h.detail)
}

func (h *Handler) user(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	p, ok := identity.From(r.Context())
	if !ok {
		httpx.Fail(w, r, apperr.ErrUnauthorized)
		return uuid.UUID{}, false
	}
	return p.UserID, true
}

func (h *Handler) summary(w http.ResponseWriter, r *http.Request) {
	uid, ok := h.user(w, r)
	if !ok {
		return
	}
	res, err := h.svc.GetSummary(r.Context(), uid)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, res)
}

func (h *Handler) funnel(w http.ResponseWriter, r *http.Request) {
	uid, ok := h.user(w, r)
	if !ok {
		return
	}
	res, err := h.svc.GetFunnel(r.Context(), uid)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, res)
}

func (h *Handler) sensitivity(w http.ResponseWriter, r *http.Request) {
	uid, ok := h.user(w, r)
	if !ok {
		return
	}
	res, err := h.svc.GetSensitivity(r.Context(), uid)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, res)
}

func (h *Handler) candidates(w http.ResponseWriter, r *http.Request) {
	uid, ok := h.user(w, r)
	if !ok {
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	res, err := h.svc.Candidates(r.Context(), uid, r.URL.Query().Get("cursor"), limit)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, res)
}

func (h *Handler) detail(w http.ResponseWriter, r *http.Request) {
	uid, ok := h.user(w, r)
	if !ok {
		return
	}
	candidateID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Fail(w, r, apperr.Invalid("候选人 ID 非法"))
		return
	}
	res, err := h.svc.CandidateDetail(r.Context(), uid, candidateID)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, res)
}
