package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"wmc/backend/internal/platform/apperr"
	"wmc/backend/internal/platform/httpx"
	"wmc/backend/internal/platform/identity"
)

// Handler 后台接口。
type Handler struct{ svc *Service }

// NewHandler 构造处理器。
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Register 注册后台路由（需管理员身份）。
func (h *Handler) Register(r chi.Router) {
	r.Get("/admin/users", h.listUsers)
	r.Get("/admin/users/{id}", h.getUser)
	r.Get("/admin/users/{id}/reports", h.userReports)
	r.Post("/admin/users/{id}/freeze", h.freeze)
	r.Post("/admin/users/{id}/unfreeze", h.unfreeze)
	r.Get("/admin/audit-logs", h.auditLogs)

	r.Get("/admin/dashboard", h.dashboard)
	r.Get("/admin/funnel", h.funnel)

	r.Get("/admin/reports", h.listReports)
	r.Post("/admin/reports/{id}/resolve", h.resolveReport)

	r.Get("/admin/photos", h.listPhotos)
	r.Post("/admin/photos/{id}/moderate", h.moderatePhoto)
	r.Post("/admin/photos/{id}/approve", h.approvePhoto)
	r.Post("/admin/photos/{id}/reject", h.rejectPhoto)

	r.Get("/admin/settings", h.getSettings)
	r.Put("/admin/settings", h.updateSetting)
}

func (h *Handler) dashboard(w http.ResponseWriter, r *http.Request) {
	res, err := h.svc.GetDashboard(r.Context())
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, res)
}

func (h *Handler) funnel(w http.ResponseWriter, r *http.Request) {
	res, err := h.svc.GetDashboard(r.Context())
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, map[string]any{"steps": res.Funnel, "note": res.Note})
}

func (h *Handler) userReports(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Fail(w, r, apperr.Invalid("用户 ID 非法"))
		return
	}
	items, err := h.svc.UserReports(r.Context(), id)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, items)
}

func (h *Handler) listReports(w http.ResponseWriter, r *http.Request) {
	status, _ := strconv.Atoi(r.URL.Query().Get("status"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, err := h.svc.ListReports(r.Context(), int16(status), int32(limit), int32(offset))
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, items)
}

func (h *Handler) resolveReport(w http.ResponseWriter, r *http.Request) {
	p, ok := identity.From(r.Context())
	if !ok {
		httpx.Fail(w, r, apperr.ErrUnauthorized)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Fail(w, r, apperr.Invalid("举报 ID 非法"))
		return
	}
	var in struct {
		Status int16  `json:"status"`
		Note   string `json:"note"`
	}
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	if err := h.svc.ResolveReport(r.Context(), p.UserID, id, in.Status, in.Note, httpx.ClientIP(r)); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, map[string]string{"status": "resolved"})
}

func (h *Handler) listPhotos(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, err := h.svc.ListPendingPhotos(r.Context(), int32(limit), int32(offset))
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, items)
}

func (h *Handler) moderatePhoto(w http.ResponseWriter, r *http.Request) {
	p, ok := identity.From(r.Context())
	if !ok {
		httpx.Fail(w, r, apperr.ErrUnauthorized)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Fail(w, r, apperr.Invalid("照片 ID 非法"))
		return
	}
	var in struct {
		Status int16  `json:"status"`
		Reason string `json:"reason"`
	}
	_ = httpx.DecodeJSON(w, r, &in)
	if err := h.svc.ModeratePhoto(r.Context(), p.UserID, id, in.Status, in.Reason, httpx.ClientIP(r)); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, map[string]string{"status": "moderated"})
}

func (h *Handler) approvePhoto(w http.ResponseWriter, r *http.Request) {
	h.moderatePhotoFixed(w, r, 2)
}

func (h *Handler) rejectPhoto(w http.ResponseWriter, r *http.Request) {
	h.moderatePhotoFixed(w, r, 3)
}

func (h *Handler) moderatePhotoFixed(w http.ResponseWriter, r *http.Request, status int16) {
	p, ok := identity.From(r.Context())
	if !ok {
		httpx.Fail(w, r, apperr.ErrUnauthorized)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Fail(w, r, apperr.Invalid("照片 ID 非法"))
		return
	}
	if err := h.svc.ModeratePhoto(r.Context(), p.UserID, id, status, "", httpx.ClientIP(r)); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, map[string]string{"status": "moderated"})
}

func (h *Handler) getSettings(w http.ResponseWriter, r *http.Request) {
	bundle, err := h.svc.GetSettings(r.Context())
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, bundle)
}

func (h *Handler) updateSetting(w http.ResponseWriter, r *http.Request) {
	p, ok := identity.From(r.Context())
	if !ok {
		httpx.Fail(w, r, apperr.ErrUnauthorized)
		return
	}
	var in struct {
		Key   string          `json:"key"`
		Value json.RawMessage `json:"value"`
	}
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	if err := h.svc.UpdateSetting(r.Context(), p.UserID, in.Key, in.Value, httpx.ClientIP(r)); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	bundle, err := h.svc.GetSettings(r.Context())
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, bundle)
}

func (h *Handler) listUsers(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	statusStr := r.URL.Query().Get("status")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	var status int64
	if statusStr != "" {
		status, _ = strconv.ParseInt(statusStr, 10, 16)
	}
	res, err := h.svc.ListUsers(r.Context(), keyword, int16(status), page, pageSize)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, res)
}

func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Fail(w, r, apperr.Invalid("用户 ID 非法"))
		return
	}
	detail, err := h.svc.GetUser(r.Context(), id)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, detail)
}

func (h *Handler) freeze(w http.ResponseWriter, r *http.Request) {
	p, ok := identity.From(r.Context())
	if !ok {
		httpx.Fail(w, r, apperr.ErrUnauthorized)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Fail(w, r, apperr.Invalid("用户 ID 非法"))
		return
	}
	var in struct {
		Reason string `json:"reason"`
	}
	_ = httpx.DecodeJSON(w, r, &in)
	if err := h.svc.Freeze(r.Context(), p.UserID, id, in.Reason, httpx.ClientIP(r)); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, map[string]string{"status": "suspended"})
}

func (h *Handler) unfreeze(w http.ResponseWriter, r *http.Request) {
	p, ok := identity.From(r.Context())
	if !ok {
		httpx.Fail(w, r, apperr.ErrUnauthorized)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Fail(w, r, apperr.Invalid("用户 ID 非法"))
		return
	}
	if err := h.svc.Unfreeze(r.Context(), p.UserID, id, httpx.ClientIP(r)); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, map[string]string{"status": "active"})
}

func (h *Handler) auditLogs(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, err := h.svc.ListAuditLogs(r.Context(), int32(limit), int32(offset))
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, items)
}