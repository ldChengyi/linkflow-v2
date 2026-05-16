package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/ldchengyi/linkflow-v2/service/backend/internal/httperror"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/middleware"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/response"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/service"
)

type ThingsModelHandler struct {
	service *service.ThingsModelService
	log     *slog.Logger
}

type thingsModelCreateRequest struct {
	TenantID     string                    `json:"tenant_id"`
	ProductID    string                    `json:"product_id"`
	ModelVersion int                       `json:"model_version"`
	ModelName    string                    `json:"model_name"`
	Description  string                    `json:"description"`
	Status       string                    `json:"status"`
	IsCurrent    bool                      `json:"is_current"`
	Properties   service.ThingsModelObject `json:"properties"`
	Events       service.ThingsModelObject `json:"events"`
	Services     service.ThingsModelObject `json:"services"`
}

type thingsModelUpdateRequest struct {
	ModelName   string                    `json:"model_name"`
	Description string                    `json:"description"`
	Status      string                    `json:"status"`
	IsCurrent   bool                      `json:"is_current"`
	Properties  service.ThingsModelObject `json:"properties"`
	Events      service.ThingsModelObject `json:"events"`
	Services    service.ThingsModelObject `json:"services"`
}

func NewThingsModelHandler(service *service.ThingsModelService, log *slog.Logger) (*ThingsModelHandler, error) {
	if service == nil {
		return nil, errors.New("thingsmodel service is nil")
	}
	if log == nil {
		log = slog.Default()
	}
	return &ThingsModelHandler{service: service, log: log}, nil
}

func (h *ThingsModelHandler) RegisterRoutes(mux RouteRegistrar, authenticate func(http.Handler) http.Handler, audit func(http.Handler) http.Handler) {
	if authenticate == nil {
		return
	}
	mux.Handle("POST /api/v1/thingsmodels", authenticatedBusinessHandler(http.HandlerFunc(h.create), authenticate, audit))
	mux.Handle("GET /api/v1/thingsmodels", authenticate(http.HandlerFunc(h.list)))
	mux.Handle("GET /api/v1/thingsmodels/{thingsmodel_id}", authenticate(http.HandlerFunc(h.get)))
	mux.Handle("PUT /api/v1/thingsmodels/{thingsmodel_id}", authenticatedBusinessHandler(http.HandlerFunc(h.update), authenticate, audit))
	mux.Handle("DELETE /api/v1/thingsmodels/{thingsmodel_id}", authenticatedBusinessHandler(http.HandlerFunc(h.delete), authenticate, audit))
}

func (h *ThingsModelHandler) create(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		writeHTTPError(w, httperror.ErrUnauthorized)
		return
	}

	var req thingsModelCreateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeHTTPError(w, err)
		return
	}
	auditRecorder, _ := middleware.AuditFromContext(r.Context())
	auditRecorder.Set("thingsmodel.create", "thingsmodel", "", map[string]any{
		"tenant_id":     req.TenantID,
		"product_id":    req.ProductID,
		"model_version": req.ModelVersion,
		"model_name":    req.ModelName,
		"status":        req.Status,
		"is_current":    req.IsCurrent,
	})

	model, err := h.service.Create(r.Context(), service.ThingsModelCreateInput{
		UserID:       principal.UserID,
		TenantID:     req.TenantID,
		ProductID:    req.ProductID,
		ModelVersion: req.ModelVersion,
		ModelName:    req.ModelName,
		Description:  req.Description,
		Status:       req.Status,
		IsCurrent:    req.IsCurrent,
		Properties:   req.Properties,
		Events:       req.Events,
		Services:     req.Services,
	})
	if err != nil {
		auditRecorder.SetErrorCode(mapThingsModelError(err).Code)
		h.writeThingsModelError(w, err)
		return
	}
	auditRecorder.SetResourceID(model.ID)

	writeJSON(w, http.StatusCreated, response.SuccessData("thingsmodel created", http.StatusCreated, model))
}

func (h *ThingsModelHandler) list(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		writeHTTPError(w, httperror.ErrUnauthorized)
		return
	}

	models, err := h.service.List(r.Context(), service.ThingsModelListInput{
		UserID:    principal.UserID,
		TenantID:  r.URL.Query().Get("tenant_id"),
		ProductID: r.URL.Query().Get("product_id"),
		PageInput: pageInputFromRequest(r),
	})
	if err != nil {
		h.writeThingsModelError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response.SuccessData("ok", http.StatusOK, models))
}

func (h *ThingsModelHandler) get(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		writeHTTPError(w, httperror.ErrUnauthorized)
		return
	}

	model, err := h.service.Get(r.Context(), service.ThingsModelGetInput{
		UserID:        principal.UserID,
		ThingsModelID: r.PathValue("thingsmodel_id"),
	})
	if err != nil {
		h.writeThingsModelError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response.SuccessData("ok", http.StatusOK, model))
}

func (h *ThingsModelHandler) update(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		writeHTTPError(w, httperror.ErrUnauthorized)
		return
	}

	var req thingsModelUpdateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeHTTPError(w, err)
		return
	}
	auditRecorder, _ := middleware.AuditFromContext(r.Context())
	auditRecorder.Set("thingsmodel.update", "thingsmodel", r.PathValue("thingsmodel_id"), map[string]any{
		"model_name": req.ModelName,
		"status":     req.Status,
		"is_current": req.IsCurrent,
	})

	model, err := h.service.Update(r.Context(), service.ThingsModelUpdateInput{
		UserID:        principal.UserID,
		ThingsModelID: r.PathValue("thingsmodel_id"),
		ModelName:     req.ModelName,
		Description:   req.Description,
		Status:        req.Status,
		IsCurrent:     req.IsCurrent,
		Properties:    req.Properties,
		Events:        req.Events,
		Services:      req.Services,
	})
	if err != nil {
		auditRecorder.SetErrorCode(mapThingsModelError(err).Code)
		h.writeThingsModelError(w, err)
		return
	}
	auditRecorder.AddMetadata(map[string]any{
		"tenant_id":     model.TenantID,
		"product_id":    model.ProductID,
		"model_version": model.ModelVersion,
	})

	writeJSON(w, http.StatusOK, response.SuccessData("thingsmodel updated", http.StatusOK, model))
}

func (h *ThingsModelHandler) delete(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		writeHTTPError(w, httperror.ErrUnauthorized)
		return
	}

	auditRecorder, _ := middleware.AuditFromContext(r.Context())
	auditRecorder.Set("thingsmodel.delete", "thingsmodel", r.PathValue("thingsmodel_id"), nil)

	if err := h.service.Delete(r.Context(), service.ThingsModelDeleteInput{
		UserID:        principal.UserID,
		ThingsModelID: r.PathValue("thingsmodel_id"),
	}); err != nil {
		auditRecorder.SetErrorCode(mapThingsModelError(err).Code)
		h.writeThingsModelError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response.SuccessData("thingsmodel deleted", http.StatusOK, map[string]bool{"deleted": true}))
}

func (h *ThingsModelHandler) writeThingsModelError(w http.ResponseWriter, err error) {
	httpErr := mapThingsModelError(err)
	if httpErr.HTTPStatus >= http.StatusInternalServerError {
		h.log.Error("thingsmodel request failed", "err", err)
	}
	writeHTTPError(w, httpErr)
}

func mapThingsModelError(err error) *httperror.Error {
	switch {
	case errors.Is(err, service.ErrInvalidThingsModelInput):
		return httperror.ErrInvalidRequest
	case errors.Is(err, service.ErrThingsModelAlreadyExists):
		return httperror.ErrThingsModelAlreadyExists
	case errors.Is(err, service.ErrThingsModelNotFound):
		return httperror.ErrNotFound
	default:
		return httperror.From(err)
	}
}
