package meta

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"wmc/backend/internal/platform/apperr"
	"wmc/backend/internal/platform/httpx"
)

// Handler 元数据处理器。
type Handler struct{ svc *Service }

// NewHandler 构造处理器。
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// RegisterPublic 注册公开路由。
func (h *Handler) RegisterPublic(r chi.Router) {
	r.Get("/meta/options", h.options)
	r.Get("/meta/regions", h.regions)
	r.Get("/meta/compatibility-questions", h.questions)
	r.Get("/meta/app-config", h.appConfig)
}

// RegisterAdmin 注册后台路由。
func (h *Handler) RegisterAdmin(r chi.Router) {
	r.Get("/admin/questions", h.adminQuestions)
	r.Post("/admin/questions", h.createQuestion)
	r.Put("/admin/questions/{id}", h.updateQuestion)
}

func (h *Handler) options(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, r, h.svc.Options())
}

func (h *Handler) regions(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.Regions(r.Context())
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, items)
}

func (h *Handler) questions(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ActiveQuestions(r.Context())
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, items)
}

func (h *Handler) appConfig(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, r, h.svc.AppConfig(r.Context()))
}

func (h *Handler) adminQuestions(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.AllQuestions(r.Context())
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, items)
}

func (h *Handler) createQuestion(w http.ResponseWriter, r *http.Request) {
	var in QuestionInput
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	q, err := h.svc.CreateQuestion(r.Context(), in)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.Created(w, r, q)
}

func (h *Handler) updateQuestion(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Fail(w, r, apperr.Invalid("id 格式不正确"))
		return
	}
	var in QuestionInput
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	q, err := h.svc.UpdateQuestion(r.Context(), id, in)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.OK(w, r, q)
}
