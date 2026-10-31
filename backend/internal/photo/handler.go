package photo

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"wmc/backend/internal/platform/apperr"
	"wmc/backend/internal/platform/httpx"
	"wmc/backend/internal/platform/identity"
	"wmc/backend/internal/storage"
)

// Handler 照片接口。
type Handler struct {
	svc *Service
}

// NewHandler 构造处理器。
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Register 注册需要登录的路由。
func (h *Handler) Register(r chi.Router) {
	r.Get("/me/photos", h.list)
	r.Post("/me/photos", h.upload)
	r.Delete("/me/photos/{id}", h.remove)
}

// RegisterMedia 注册媒体代理路由（仅本地存储驱动需要）。
func (h *Handler) RegisterMedia(r chi.Router, fs *storage.FSStorage) {
	if fs == nil {
		return
	}
	r.Get("/media/*", func(w http.ResponseWriter, req *http.Request) {
		key := chi.URLParam(req, "*")
		exp := req.URL.Query().Get("exp")
		sig := req.URL.Query().Get("sig")
		var expUnix int64
		if _, err := fmtSscan(exp, &expUnix); err != nil || !fs.Verify(key, sig, expUnix) {
			http.NotFound(w, req)
			return
		}
		data, err := fs.Read(key)
		if err != nil {
			http.NotFound(w, req)
			return
		}
		w.Header().Set("Content-Type", storage.ContentTypeOf(key))
		w.Header().Set("Cache-Control", "private, max-age=3600")
		_, _ = w.Write(data)
	})
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	p, ok := identity.From(r.Context())
	if !ok {
		httpx.Fail(w, r, apperr.ErrUnauthorized)
		return
	}
	items, err := h.svc.List(r.Context(), p.UserID)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, items)
}

func (h *Handler) upload(w http.ResponseWriter, r *http.Request) {
	p, ok := identity.From(r.Context())
	if !ok {
		httpx.Fail(w, r, apperr.ErrUnauthorized)
		return
	}
	if err := r.ParseMultipartForm(MaxFileSize + (1 << 20)); err != nil {
		httpx.Fail(w, r, apperr.Invalid("上传内容解析失败"))
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		httpx.Fail(w, r, apperr.Invalid("请选择要上传的图片"))
		return
	}
	defer file.Close()
	makePrimary := r.FormValue("make_primary") == "true"
	photo, err := h.svc.Upload(r.Context(), p.UserID, file, header.Size, makePrimary)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.Created(w, r, photo)
}

func (h *Handler) remove(w http.ResponseWriter, r *http.Request) {
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
	if err := h.svc.Delete(r.Context(), p.UserID, id); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, map[string]string{"status": "deleted"})
}
