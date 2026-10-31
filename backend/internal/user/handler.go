package user

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"wmc/backend/internal/platform/apperr"
	"wmc/backend/internal/platform/httpx"
	"wmc/backend/internal/platform/identity"
)

// Handler 当前用户接口。
type Handler struct{ svc *Service }

// NewHandler 构造处理器。
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Register 注册需要登录的路由。
func (h *Handler) Register(r chi.Router) {
	r.Get("/me/profile", h.getProfile)
	r.Put("/me/profile", h.saveProfile)
	r.Get("/me/preferences", h.getPreferences)
	r.Put("/me/preferences", h.savePreferences)
	r.Get("/me/compatibility", h.getRules)
	r.Put("/me/compatibility", h.saveRules)
	r.Get("/me/consents", h.listConsents)
	r.Post("/me/consents", h.recordConsent)
	r.Get("/me/market/status", h.marketStatus)
	r.Post("/me/market/join", h.joinMarket)
	r.Delete("/me/market/join", h.leaveMarket)
	r.Delete("/me/account", h.deleteAccount)
}

func principal(w http.ResponseWriter, r *http.Request) (identity.Principal, bool) {
	p, ok := identity.From(r.Context())
	if !ok {
		httpx.Fail(w, r, apperr.ErrUnauthorized)
		return identity.Principal{}, false
	}
	return p, true
}

func (h *Handler) getProfile(w http.ResponseWriter, r *http.Request) {
	p, ok := principal(w, r)
	if !ok {
		return
	}
	prof, err := h.svc.GetProfile(r.Context(), p.UserID)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, prof)
}

func (h *Handler) saveProfile(w http.ResponseWriter, r *http.Request) {
	p, ok := principal(w, r)
	if !ok {
		return
	}
	var in ProfileInput
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	prof, err := h.svc.SaveProfile(r.Context(), p.UserID, in)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, prof)
}

func (h *Handler) getPreferences(w http.ResponseWriter, r *http.Request) {
	p, ok := principal(w, r)
	if !ok {
		return
	}
	pref, err := h.svc.GetPreferences(r.Context(), p.UserID)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, pref)
}

func (h *Handler) savePreferences(w http.ResponseWriter, r *http.Request) {
	p, ok := principal(w, r)
	if !ok {
		return
	}
	var in Preferences
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	pref, err := h.svc.SavePreferences(r.Context(), p.UserID, in)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, pref)
}

func (h *Handler) getRules(w http.ResponseWriter, r *http.Request) {
	p, ok := principal(w, r)
	if !ok {
		return
	}
	rules, err := h.svc.GetRules(r.Context(), p.UserID)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, rules)
}

func (h *Handler) saveRules(w http.ResponseWriter, r *http.Request) {
	p, ok := principal(w, r)
	if !ok {
		return
	}
	var in struct {
		Rules []CompatibilityRuleInput `json:"rules"`
	}
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	if err := h.svc.SaveRules(r.Context(), p.UserID, in.Rules); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	rules, err := h.svc.GetRules(r.Context(), p.UserID)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, rules)
}

func (h *Handler) listConsents(w http.ResponseWriter, r *http.Request) {
	p, ok := principal(w, r)
	if !ok {
		return
	}
	items, err := h.svc.ListConsents(r.Context(), p.UserID)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, items)
}

func (h *Handler) recordConsent(w http.ResponseWriter, r *http.Request) {
	p, ok := principal(w, r)
	if !ok {
		return
	}
	var in ConsentInput
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	if err := h.svc.RecordConsent(r.Context(), p.UserID, in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, map[string]string{"status": "recorded"})
}

func (h *Handler) marketStatus(w http.ResponseWriter, r *http.Request) {
	p, ok := principal(w, r)
	if !ok {
		return
	}
	status, err := h.svc.MarketStatus(r.Context(), p.UserID)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, status)
}

func (h *Handler) joinMarket(w http.ResponseWriter, r *http.Request) {
	p, ok := principal(w, r)
	if !ok {
		return
	}
	if err := h.svc.JoinMarket(r.Context(), p.UserID); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, map[string]bool{"discovery_enabled": true})
}

func (h *Handler) leaveMarket(w http.ResponseWriter, r *http.Request) {
	p, ok := principal(w, r)
	if !ok {
		return
	}
	if err := h.svc.LeaveMarket(r.Context(), p.UserID); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, map[string]bool{"discovery_enabled": false})
}

func (h *Handler) deleteAccount(w http.ResponseWriter, r *http.Request) {
	p, ok := principal(w, r)
	if !ok {
		return
	}
	if err := h.svc.RequestDeletion(r.Context(), p.UserID); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, map[string]string{"status": "deletion_pending"})
}
