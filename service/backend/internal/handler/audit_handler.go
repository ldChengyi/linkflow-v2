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

type AuditHandler struct {
	service *service.AuditService
	log     *slog.Logger
}

func NewAuditHandler(service *service.AuditService, log *slog.Logger) (*AuditHandler, error) {
	if service == nil {
		return nil, errors.New("audit service is nil")
	}
	if log == nil {
		log = slog.Default()
	}
	return &AuditHandler{service: service, log: log}, nil
}

func (h *AuditHandler) RegisterRoutes(mux RouteRegistrar, authenticate func(http.Handler) http.Handler) {
	if authenticate == nil {
		return
	}
	mux.Handle("GET /api/v1/audit-logs", authenticate(http.HandlerFunc(h.list)))
}

func (h *AuditHandler) list(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		writeHTTPError(w, httperror.ErrUnauthorized)
		return
	}

	logs, err := h.service.List(r.Context(), service.AuditListInput{
		UserID:    principal.UserID,
		PageInput: pageInputFromRequest(r),
	})
	if err != nil {
		h.writeAuditError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response.SuccessData("ok", http.StatusOK, logs))
}

func (h *AuditHandler) writeAuditError(w http.ResponseWriter, err error) {
	httpErr := mapAuditError(err)
	if httpErr.HTTPStatus >= http.StatusInternalServerError {
		h.log.Error("audit request failed", "err", err)
	}
	writeHTTPError(w, httpErr)
}

func mapAuditError(err error) *httperror.Error {
	switch {
	case errors.Is(err, service.ErrInvalidAuditInput):
		return httperror.ErrInvalidRequest
	default:
		return httperror.From(err)
	}
}
