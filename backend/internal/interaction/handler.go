package interaction

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"wmc/backend/internal/events"
	"wmc/backend/internal/platform/apperr"
	"wmc/backend/internal/platform/httpx"
	"wmc/backend/internal/platform/identity"
	"wmc/backend/internal/store"
)

// Handler 互动接口。
type Handler struct {
	svc   *Service
	store *store.Store
}

// NewHandler 构造处理器。
func NewHandler(svc *Service, st *store.Store) *Handler { return &Handler{svc: svc, store: st} }

// RegisterPublic 注册游客可用的路由。
func (h *Handler) RegisterPublic(r chi.Router) {
	r.Post("/events", h.recordEvent)
}

// Register 注册需要登录的路由。
func (h *Handler) Register(r chi.Router) {
	r.Put("/interactions/{user_id}", h.act)
	r.Get("/matches", h.listMatches)
	r.Get("/matches/{id}", h.getMatch)
	r.Delete("/matches/{id}", h.cancelMatch)
	r.Post("/matches/{id}/outcomes", h.addOutcome)
	r.Post("/users/{id}/block", h.block)
	r.Delete("/users/{id}/block", h.unblock)
	r.Post("/users/{id}/report", h.report)
}

func (h *Handler) actor(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	p, ok := identity.From(r.Context())
	if !ok {
		httpx.Fail(w, r, apperr.ErrUnauthorized)
		return uuid.UUID{}, false
	}
	return p.UserID, true
}

func pathID(r *http.Request, key string) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, key))
	if err != nil {
		return uuid.UUID{}, apperr.Invalid("ID 格式非法")
	}
	return id, nil
}

func (h *Handler) act(w http.ResponseWriter, r *http.Request) {
	uid, ok := h.actor(w, r)
	if !ok {
		return
	}
	targetID, err := pathID(r, "user_id")
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	var in struct {
		Action int16 `json:"action"`
		Source int16 `json:"source"`
	}
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	res, err := h.svc.Act(r.Context(), uid, targetID, in.Action, in.Source)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, res)
}

func (h *Handler) listMatches(w http.ResponseWriter, r *http.Request) {
	uid, ok := h.actor(w, r)
	if !ok {
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, err := h.svc.ListMatches(r.Context(), uid, int32(limit), int32(offset))
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, map[string]any{"items": items})
}

func (h *Handler) getMatch(w http.ResponseWriter, r *http.Request) {
	uid, ok := h.actor(w, r)
	if !ok {
		return
	}
	matchID, err := pathID(r, "id")
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	detail, err := h.svc.GetMatch(r.Context(), uid, matchID)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, detail)
}

func (h *Handler) cancelMatch(w http.ResponseWriter, r *http.Request) {
	uid, ok := h.actor(w, r)
	if !ok {
		return
	}
	matchID, err := pathID(r, "id")
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	if err := h.svc.CancelMatch(r.Context(), uid, matchID); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, map[string]string{"status": "cancelled"})
}

func (h *Handler) addOutcome(w http.ResponseWriter, r *http.Request) {
	uid, ok := h.actor(w, r)
	if !ok {
		return
	}
	matchID, err := pathID(r, "id")
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	var in struct {
		OutcomeType int16  `json:"outcome_type"`
		OccurredAt  string `json:"occurred_at"`
		Note        string `json:"note"`
	}
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	var occurred *time.Time
	if in.OccurredAt != "" {
		if t, err := time.Parse(time.RFC3339, in.OccurredAt); err == nil {
			occurred = &t
		}
	}
	if err := h.svc.AddOutcome(r.Context(), uid, matchID, in.OutcomeType, occurred, in.Note); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, map[string]string{"status": "recorded"})
}

func (h *Handler) block(w http.ResponseWriter, r *http.Request) {
	uid, ok := h.actor(w, r)
	if !ok {
		return
	}
	targetID, err := pathID(r, "id")
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	if err := h.svc.Block(r.Context(), uid, targetID); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, map[string]string{"status": "blocked"})
}

func (h *Handler) unblock(w http.ResponseWriter, r *http.Request) {
	uid, ok := h.actor(w, r)
	if !ok {
		return
	}
	targetID, err := pathID(r, "id")
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	if err := h.svc.Unblock(r.Context(), uid, targetID); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, map[string]string{"status": "unblocked"})
}

func (h *Handler) report(w http.ResponseWriter, r *http.Request) {
	uid, ok := h.actor(w, r)
	if !ok {
		return
	}
	targetID, err := pathID(r, "id")
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	var in struct {
		Reason       int16    `json:"reason"`
		Description  string   `json:"description"`
		EvidenceKeys []string `json:"evidence_keys"`
	}
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	if err := h.svc.Report(r.Context(), uid, targetID, in.Reason, in.Description, in.EvidenceKeys); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, map[string]string{"status": "reported"})
}

func (h *Handler) recordEvent(w http.ResponseWriter, r *http.Request) {
	var in struct {
		EventName   string         `json:"event_name"`
		AnonymousID string         `json:"anonymous_id"`
		Properties  map[string]any `json:"properties"`
	}
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	if in.EventName == "" {
		httpx.Fail(w, r, apperr.Invalid("event_name 不能为空"))
		return
	}
	var userID *uuid.UUID
	if p, ok := identity.From(r.Context()); ok {
		userID = &p.UserID
	}
	var anon *string
	if in.AnonymousID != "" {
		anon = &in.AnonymousID
	}
	events.Record(r.Context(), h.store, userID, anon, in.EventName, in.Properties)
	httpx.OK(w, r, map[string]string{"status": "recorded"})
}