package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/ldchengyi/linkflow-v2/service/backend/internal/httperror"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/middleware"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/response"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/service"
)

type TenantHandler struct {
	service *service.TenantService
	log     *slog.Logger
}

type tenantCreateRequest struct {
	TenantSlug string `json:"tenant_slug"`
	TenantName string `json:"tenant_name"`
}

type tenantUpdateRequest struct {
	TenantName string `json:"tenant_name"`
	Status     string `json:"status"`
}

func NewTenantHandler(service *service.TenantService, log *slog.Logger) (*TenantHandler, error) {
	if service == nil {
		return nil, errors.New("tenant service is nil")
	}
	if log == nil {
		log = slog.Default()
	}
	return &TenantHandler{service: service, log: log}, nil
}

func (h *TenantHandler) RegisterRoutes(mux RouteRegistrar, authenticate func(http.Handler) http.Handler, audit func(http.Handler) http.Handler) {
	if authenticate == nil {
		return
	}
	mux.Handle("POST /api/v1/tenants", authenticatedBusinessHandler(http.HandlerFunc(h.create), authenticate, audit))
	mux.Handle("GET /api/v1/tenants", authenticate(http.HandlerFunc(h.list)))
	mux.Handle("GET /api/v1/tenants/{tenant_id}", authenticate(http.HandlerFunc(h.get)))
	mux.Handle("PUT /api/v1/tenants/{tenant_id}", authenticatedBusinessHandler(http.HandlerFunc(h.update), authenticate, audit))
	mux.Handle("DELETE /api/v1/tenants/{tenant_id}", authenticatedBusinessHandler(http.HandlerFunc(h.delete), authenticate, audit))
}

func authenticatedBusinessHandler(h http.Handler, authenticate func(http.Handler) http.Handler, audit func(http.Handler) http.Handler) http.Handler {
	if audit == nil {
		return authenticate(h)
	}
	return authenticate(audit(h))
}

func (h *TenantHandler) create(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		writeHTTPError(w, httperror.ErrUnauthorized)
		return
	}

	var req tenantCreateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeHTTPError(w, err)
		return
	}
	auditRecorder, _ := middleware.AuditFromContext(r.Context())
	auditRecorder.Set("tenant.create", "tenant", "", map[string]any{
		"tenant_slug": req.TenantSlug,
		"tenant_name": req.TenantName,
	})

	tenant, err := h.service.Create(r.Context(), service.TenantCreateInput{
		UserID:     principal.UserID,
		TenantSlug: req.TenantSlug,
		TenantName: req.TenantName,
	})
	if err != nil {
		auditRecorder.SetErrorCode(mapTenantError(err).Code)
		h.writeTenantError(w, err)
		return
	}
	auditRecorder.SetResourceID(tenant.ID)

	writeJSON(w, http.StatusCreated, response.SuccessData("tenant created", http.StatusCreated, tenant))
}

func (h *TenantHandler) list(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		writeHTTPError(w, httperror.ErrUnauthorized)
		return
	}

	tenants, err := h.service.List(r.Context(), service.TenantListInput{
		UserID:    principal.UserID,
		PageInput: pageInputFromRequest(r),
	})
	if err != nil {
		h.writeTenantError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response.SuccessData("ok", http.StatusOK, tenants))
}

func (h *TenantHandler) get(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		writeHTTPError(w, httperror.ErrUnauthorized)
		return
	}

	tenant, err := h.service.Get(r.Context(), service.TenantGetInput{
		UserID:   principal.UserID,
		TenantID: r.PathValue("tenant_id"),
	})
	if err != nil {
		h.writeTenantError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response.SuccessData("ok", http.StatusOK, tenant))
}

func (h *TenantHandler) update(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		writeHTTPError(w, httperror.ErrUnauthorized)
		return
	}

	var req tenantUpdateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeHTTPError(w, err)
		return
	}
	auditRecorder, _ := middleware.AuditFromContext(r.Context())
	auditRecorder.Set("tenant.update", "tenant", r.PathValue("tenant_id"), map[string]any{
		"tenant_name": req.TenantName,
		"status":      req.Status,
	})

	tenant, err := h.service.Update(r.Context(), service.TenantUpdateInput{
		UserID:     principal.UserID,
		TenantID:   r.PathValue("tenant_id"),
		TenantName: req.TenantName,
		Status:     req.Status,
	})
	if err != nil {
		auditRecorder.SetErrorCode(mapTenantError(err).Code)
		h.writeTenantError(w, err)
		return
	}
	auditRecorder.AddMetadata(map[string]any{
		"tenant_slug": tenant.TenantSlug,
	})

	writeJSON(w, http.StatusOK, response.SuccessData("tenant updated", http.StatusOK, tenant))
}

func pageInputFromRequest(r *http.Request) service.PageInput {
	query := r.URL.Query()
	return service.PageInput{
		Page:     parsePositiveInt(query.Get("page")),
		PageSize: parsePositiveInt(query.Get("page_size")),
	}
}

func parsePositiveInt(raw string) int {
	if raw == "" {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return 0
	}
	return n
}

func (h *TenantHandler) delete(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		writeHTTPError(w, httperror.ErrUnauthorized)
		return
	}

	auditRecorder, _ := middleware.AuditFromContext(r.Context())
	auditRecorder.Set("tenant.delete", "tenant", r.PathValue("tenant_id"), nil)

	if err := h.service.Delete(r.Context(), service.TenantDeleteInput{
		UserID:   principal.UserID,
		TenantID: r.PathValue("tenant_id"),
	}); err != nil {
		auditRecorder.SetErrorCode(mapTenantError(err).Code)
		h.writeTenantError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response.SuccessData("tenant deleted", http.StatusOK, map[string]bool{"deleted": true}))
}

func (h *TenantHandler) writeTenantError(w http.ResponseWriter, err error) {
	httpErr := mapTenantError(err)
	if httpErr.HTTPStatus >= http.StatusInternalServerError {
		h.log.Error("tenant request failed", "err", err)
	}
	writeHTTPError(w, httpErr)
}

func mapTenantError(err error) *httperror.Error {
	switch {
	case errors.Is(err, service.ErrInvalidTenantInput):
		return httperror.ErrInvalidRequest
	case errors.Is(err, service.ErrTenantAlreadyExists):
		return httperror.ErrTenantAlreadyExists
	case errors.Is(err, service.ErrTenantNotFound):
		return httperror.ErrNotFound
	default:
		return httperror.From(err)
	}
}
