package baseline

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"wmc/backend/internal/platform/apperr"
	"wmc/backend/internal/platform/httpx"
	"wmc/backend/internal/platform/identity"
)

// Handler 数据集后台处理器。
type Handler struct{ svc *Service }

// NewHandler 构造处理器。
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Register 注册后台路由。
func (h *Handler) Register(r chi.Router) {
	r.Get("/admin/baseline/datasets", h.list)
	r.Post("/admin/baseline/datasets", h.create)
	r.Get("/admin/baseline/datasets/{id}", h.get)
	r.Post("/admin/baseline/datasets/{id}/import", h.importCSV)
	r.Post("/admin/baseline/datasets/{id}/activate", h.activate)
	r.Post("/admin/baseline/datasets/{id}/rollback", h.rollback)
	r.Post("/admin/baseline/datasets/demo", h.createDemo)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.List(r.Context())
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, items)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	ds, err := h.svc.Get(r.Context(), id)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, ds)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var in CreateInput
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	var createdBy *uuid.UUID
	if p, ok := identity.From(r.Context()); ok {
		createdBy = &p.UserID
	}
	ds, err := h.svc.Create(r.Context(), in, createdBy)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.Created(w, r, ds)
}

func (h *Handler) importCSV(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		httpx.Fail(w, r, apperr.Invalid("上传文件解析失败"))
		return
	}
	dryRun := r.FormValue("dry_run") == "true"
	file, _, err := r.FormFile("file")
	if err != nil {
		httpx.Fail(w, r, apperr.Invalid("请上传 CSV 文件"))
		return
	}
	defer file.Close()
	report, err := h.svc.ImportCSV(r.Context(), id, file, dryRun)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, report)
}

func (h *Handler) activate(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	if err := h.svc.Activate(r.Context(), id); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, map[string]string{"status": "active"})
}

func (h *Handler) rollback(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	if err := h.svc.Rollback(r.Context(), id); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, map[string]string{"status": "archived"})
}

func (h *Handler) createDemo(w http.ResponseWriter, r *http.Request) {
	var createdBy *uuid.UUID
	if p, ok := identity.From(r.Context()); ok {
		createdBy = &p.UserID
	}
	ds, err := h.svc.CreateDemoDataset(r.Context(), 3000, createdBy)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	if err := h.svc.Activate(r.Context(), ds.ID); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.Created(w, r, ds)
}

func parseID(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return uuid.UUID{}, apperr.Invalid("id 格式不正确")
	}
	return id, nil
}
