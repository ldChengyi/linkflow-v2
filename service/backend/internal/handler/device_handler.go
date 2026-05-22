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

type DeviceHandler struct {
	service *service.DeviceService
	log     *slog.Logger
}

type deviceCreateRequest struct {
	TenantID        string `json:"tenant_id"`
	ProductID       string `json:"product_id"`
	DeviceSlug      string `json:"device_slug"`
	DeviceName      string `json:"device_name"`
	Description     string `json:"description"`
	GatewayDeviceID string `json:"gateway_device_id"`
}

type deviceUpdateRequest struct {
	DeviceName      string `json:"device_name"`
	Description     string `json:"description"`
	Status          string `json:"status"`
	GatewayDeviceID string `json:"gateway_device_id"`
}

type deviceServiceCallRequest struct {
	Input map[string]any `json:"input"`
}

func NewDeviceHandler(service *service.DeviceService, log *slog.Logger) (*DeviceHandler, error) {
	if service == nil {
		return nil, errors.New("device service is nil")
	}
	if log == nil {
		log = slog.Default()
	}
	return &DeviceHandler{service: service, log: log}, nil
}

func (h *DeviceHandler) RegisterRoutes(mux RouteRegistrar, authenticate func(http.Handler) http.Handler, audit func(http.Handler) http.Handler) {
	if authenticate == nil {
		return
	}
	mux.Handle("POST /api/v1/devices", authenticatedBusinessHandler(http.HandlerFunc(h.create), authenticate, audit))
	mux.Handle("GET /api/v1/devices", authenticate(http.HandlerFunc(h.list)))
	mux.Handle("GET /api/v1/devices/{device_id}/properties/latest", authenticate(http.HandlerFunc(h.latestProperties)))
	mux.Handle("GET /api/v1/devices/{device_id}/events", authenticate(http.HandlerFunc(h.eventHistory)))
	mux.Handle("GET /api/v1/devices/{device_id}/services/history", authenticate(http.HandlerFunc(h.serviceCallHistory)))
	mux.Handle("GET /api/v1/devices/{device_id}", authenticate(http.HandlerFunc(h.get)))
	mux.Handle("POST /api/v1/devices/{device_id}/services/{service_name}/call", authenticatedBusinessHandler(http.HandlerFunc(h.callService), authenticate, audit))
	mux.Handle("PUT /api/v1/devices/{device_id}", authenticatedBusinessHandler(http.HandlerFunc(h.update), authenticate, audit))
	mux.Handle("DELETE /api/v1/devices/{device_id}", authenticatedBusinessHandler(http.HandlerFunc(h.delete), authenticate, audit))
}

func (h *DeviceHandler) create(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		writeHTTPError(w, httperror.ErrUnauthorized)
		return
	}

	var req deviceCreateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeHTTPError(w, err)
		return
	}
	auditRecorder, _ := middleware.AuditFromContext(r.Context())
	auditRecorder.Set("device.create", "device", "", map[string]any{
		"tenant_id":         req.TenantID,
		"product_id":        req.ProductID,
		"device_slug":       req.DeviceSlug,
		"device_name":       req.DeviceName,
		"gateway_device_id": req.GatewayDeviceID,
	})

	result, err := h.service.Create(r.Context(), service.DeviceCreateInput{
		UserID:          principal.UserID,
		TenantID:        req.TenantID,
		ProductID:       req.ProductID,
		DeviceSlug:      req.DeviceSlug,
		DeviceName:      req.DeviceName,
		Description:     req.Description,
		GatewayDeviceID: req.GatewayDeviceID,
	})
	if err != nil {
		auditRecorder.SetErrorCode(mapDeviceError(err).Code)
		h.writeDeviceError(w, err)
		return
	}
	device := result.Device
	auditRecorder.SetResourceID(device.ID)

	writeJSON(w, http.StatusCreated, response.SuccessData("device created", http.StatusCreated, result))
}

func (h *DeviceHandler) list(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		writeHTTPError(w, httperror.ErrUnauthorized)
		return
	}

	devices, err := h.service.List(r.Context(), service.DeviceListInput{
		UserID:    principal.UserID,
		TenantID:  r.URL.Query().Get("tenant_id"),
		ProductID: r.URL.Query().Get("product_id"),
		PageInput: pageInputFromRequest(r),
	})
	if err != nil {
		h.writeDeviceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response.SuccessData("ok", http.StatusOK, devices))
}

func (h *DeviceHandler) get(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		writeHTTPError(w, httperror.ErrUnauthorized)
		return
	}

	device, err := h.service.Get(r.Context(), service.DeviceGetInput{
		UserID:   principal.UserID,
		DeviceID: r.PathValue("device_id"),
	})
	if err != nil {
		h.writeDeviceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response.SuccessData("ok", http.StatusOK, device))
}

func (h *DeviceHandler) latestProperties(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		writeHTTPError(w, httperror.ErrUnauthorized)
		return
	}

	latest, err := h.service.LatestProperties(r.Context(), service.DeviceLatestPropertiesInput{
		UserID:   principal.UserID,
		DeviceID: r.PathValue("device_id"),
	})
	if err != nil {
		h.writeDeviceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response.SuccessData("ok", http.StatusOK, latest))
}

func (h *DeviceHandler) eventHistory(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		writeHTTPError(w, httperror.ErrUnauthorized)
		return
	}

	events, err := h.service.EventHistory(r.Context(), service.DeviceEventHistoryInput{
		UserID:    principal.UserID,
		DeviceID:  r.PathValue("device_id"),
		EventName: r.URL.Query().Get("event_name"),
		PageInput: pageInputFromRequest(r),
	})
	if err != nil {
		h.writeDeviceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response.SuccessData("ok", http.StatusOK, events))
}

func (h *DeviceHandler) serviceCallHistory(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		writeHTTPError(w, httperror.ErrUnauthorized)
		return
	}

	calls, err := h.service.ServiceCallHistory(r.Context(), service.DeviceServiceCallHistoryInput{
		UserID:             principal.UserID,
		DeviceID:           r.PathValue("device_id"),
		ServiceName:        r.URL.Query().Get("service_name"),
		AckDeadlineSeconds: parsePositiveInt(r.URL.Query().Get("ack_deadline_seconds")),
		PageInput:          pageInputFromRequest(r),
	})
	if err != nil {
		h.writeDeviceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response.SuccessData("ok", http.StatusOK, calls))
}

func (h *DeviceHandler) callService(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		writeHTTPError(w, httperror.ErrUnauthorized)
		return
	}

	var req deviceServiceCallRequest
	if err := decodeJSON(r, &req); err != nil {
		writeHTTPError(w, err)
		return
	}

	deviceID := r.PathValue("device_id")
	serviceName := r.PathValue("service_name")
	auditRecorder, _ := middleware.AuditFromContext(r.Context())
	auditRecorder.Set("device.service.call", "device", deviceID, map[string]any{
		"service_name": serviceName,
	})

	result, err := h.service.CallService(r.Context(), service.DeviceServiceCallInput{
		UserID:      principal.UserID,
		DeviceID:    deviceID,
		ServiceName: serviceName,
		Input:       req.Input,
	})
	if err != nil {
		auditRecorder.SetErrorCode(mapDeviceError(err).Code)
		h.writeDeviceError(w, err)
		return
	}
	auditRecorder.AddMetadata(map[string]any{
		"command_id": result.CommandID,
		"topic":      result.Topic,
	})

	writeJSON(w, http.StatusAccepted, response.SuccessData("service call dispatched", http.StatusAccepted, result))
}

func (h *DeviceHandler) update(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		writeHTTPError(w, httperror.ErrUnauthorized)
		return
	}

	var req deviceUpdateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeHTTPError(w, err)
		return
	}
	auditRecorder, _ := middleware.AuditFromContext(r.Context())
	auditRecorder.Set("device.update", "device", r.PathValue("device_id"), map[string]any{
		"device_name":       req.DeviceName,
		"status":            req.Status,
		"gateway_device_id": req.GatewayDeviceID,
	})

	device, err := h.service.Update(r.Context(), service.DeviceUpdateInput{
		UserID:          principal.UserID,
		DeviceID:        r.PathValue("device_id"),
		DeviceName:      req.DeviceName,
		Description:     req.Description,
		Status:          req.Status,
		GatewayDeviceID: req.GatewayDeviceID,
	})
	if err != nil {
		auditRecorder.SetErrorCode(mapDeviceError(err).Code)
		h.writeDeviceError(w, err)
		return
	}
	auditRecorder.AddMetadata(map[string]any{
		"tenant_id":   device.TenantID,
		"product_id":  device.ProductID,
		"device_slug": device.DeviceSlug,
		"device_name": device.DeviceName,
		"gateway_id":  device.GatewayDeviceID,
	})

	writeJSON(w, http.StatusOK, response.SuccessData("device updated", http.StatusOK, device))
}

func (h *DeviceHandler) delete(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		writeHTTPError(w, httperror.ErrUnauthorized)
		return
	}

	auditRecorder, _ := middleware.AuditFromContext(r.Context())
	auditRecorder.Set("device.delete", "device", r.PathValue("device_id"), nil)

	if err := h.service.Delete(r.Context(), service.DeviceDeleteInput{
		UserID:   principal.UserID,
		DeviceID: r.PathValue("device_id"),
	}); err != nil {
		auditRecorder.SetErrorCode(mapDeviceError(err).Code)
		h.writeDeviceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response.SuccessData("device deleted", http.StatusOK, map[string]bool{"deleted": true}))
}

func (h *DeviceHandler) writeDeviceError(w http.ResponseWriter, err error) {
	httpErr := mapDeviceError(err)
	if httpErr.HTTPStatus >= http.StatusInternalServerError {
		h.log.Error("device request failed", "err", err)
	}
	writeHTTPError(w, httpErr)
}

func mapDeviceError(err error) *httperror.Error {
	switch {
	case errors.Is(err, service.ErrInvalidDeviceInput):
		return httperror.ErrInvalidRequest
	case errors.Is(err, service.ErrDeviceAlreadyExists):
		return httperror.ErrDeviceAlreadyExists
	case errors.Is(err, service.ErrDeviceNotFound):
		return httperror.ErrNotFound
	case errors.Is(err, service.ErrDeviceServiceNotFound):
		return httperror.ErrNotFound
	case errors.Is(err, service.ErrDeviceOffline):
		return httperror.ErrServiceUnavailable
	case errors.Is(err, service.ErrDeviceCommandPublisherUnavailable):
		return httperror.ErrServiceUnavailable
	default:
		return httperror.From(err)
	}
}
