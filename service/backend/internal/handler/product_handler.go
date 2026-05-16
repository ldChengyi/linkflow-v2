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

type ProductHandler struct {
	service *service.ProductService
	log     *slog.Logger
}

type productCreateRequest struct {
	TenantID     string `json:"tenant_id"`
	ProductKey   string `json:"product_key"`
	ProductName  string `json:"product_name"`
	Description  string `json:"description"`
	NodeType     string `json:"node_type"`
	AuthType     string `json:"auth_type"`
	ProtocolType string `json:"protocol_type"`
}

type productUpdateRequest struct {
	ProductName  string `json:"product_name"`
	Description  string `json:"description"`
	NodeType     string `json:"node_type"`
	AuthType     string `json:"auth_type"`
	ProtocolType string `json:"protocol_type"`
	Status       string `json:"status"`
}

func NewProductHandler(service *service.ProductService, log *slog.Logger) (*ProductHandler, error) {
	if service == nil {
		return nil, errors.New("product service is nil")
	}
	if log == nil {
		log = slog.Default()
	}
	return &ProductHandler{service: service, log: log}, nil
}

func (h *ProductHandler) RegisterRoutes(mux RouteRegistrar, authenticate func(http.Handler) http.Handler, audit func(http.Handler) http.Handler) {
	if authenticate == nil {
		return
	}
	mux.Handle("POST /api/v1/products", authenticatedBusinessHandler(http.HandlerFunc(h.create), authenticate, audit))
	mux.Handle("GET /api/v1/products", authenticate(http.HandlerFunc(h.list)))
	mux.Handle("GET /api/v1/products/{product_id}", authenticate(http.HandlerFunc(h.get)))
	mux.Handle("PUT /api/v1/products/{product_id}", authenticatedBusinessHandler(http.HandlerFunc(h.update), authenticate, audit))
	mux.Handle("DELETE /api/v1/products/{product_id}", authenticatedBusinessHandler(http.HandlerFunc(h.delete), authenticate, audit))
}

func (h *ProductHandler) create(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		writeHTTPError(w, httperror.ErrUnauthorized)
		return
	}

	var req productCreateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeHTTPError(w, err)
		return
	}
	auditRecorder, _ := middleware.AuditFromContext(r.Context())
	auditRecorder.Set("product.create", "product", "", map[string]any{
		"tenant_id":    req.TenantID,
		"product_key":  req.ProductKey,
		"product_name": req.ProductName,
		"node_type":    req.NodeType,
		"auth_type":    req.AuthType,
	})

	product, err := h.service.Create(r.Context(), service.ProductCreateInput{
		UserID:       principal.UserID,
		TenantID:     req.TenantID,
		ProductKey:   req.ProductKey,
		ProductName:  req.ProductName,
		Description:  req.Description,
		NodeType:     req.NodeType,
		AuthType:     req.AuthType,
		ProtocolType: req.ProtocolType,
	})
	if err != nil {
		auditRecorder.SetErrorCode(mapProductError(err).Code)
		h.writeProductError(w, err)
		return
	}
	auditRecorder.SetResourceID(product.ID)

	writeJSON(w, http.StatusCreated, response.SuccessData("product created", http.StatusCreated, product))
}

func (h *ProductHandler) list(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		writeHTTPError(w, httperror.ErrUnauthorized)
		return
	}

	products, err := h.service.List(r.Context(), service.ProductListInput{
		UserID:    principal.UserID,
		TenantID:  r.URL.Query().Get("tenant_id"),
		PageInput: pageInputFromRequest(r),
	})
	if err != nil {
		h.writeProductError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response.SuccessData("ok", http.StatusOK, products))
}

func (h *ProductHandler) get(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		writeHTTPError(w, httperror.ErrUnauthorized)
		return
	}

	product, err := h.service.Get(r.Context(), service.ProductGetInput{
		UserID:    principal.UserID,
		ProductID: r.PathValue("product_id"),
	})
	if err != nil {
		h.writeProductError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response.SuccessData("ok", http.StatusOK, product))
}

func (h *ProductHandler) update(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		writeHTTPError(w, httperror.ErrUnauthorized)
		return
	}

	var req productUpdateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeHTTPError(w, err)
		return
	}
	auditRecorder, _ := middleware.AuditFromContext(r.Context())
	auditRecorder.Set("product.update", "product", r.PathValue("product_id"), map[string]any{
		"product_name":  req.ProductName,
		"node_type":     req.NodeType,
		"auth_type":     req.AuthType,
		"protocol_type": req.ProtocolType,
		"status":        req.Status,
	})

	product, err := h.service.Update(r.Context(), service.ProductUpdateInput{
		UserID:       principal.UserID,
		ProductID:    r.PathValue("product_id"),
		ProductName:  req.ProductName,
		Description:  req.Description,
		NodeType:     req.NodeType,
		AuthType:     req.AuthType,
		ProtocolType: req.ProtocolType,
		Status:       req.Status,
	})
	if err != nil {
		auditRecorder.SetErrorCode(mapProductError(err).Code)
		h.writeProductError(w, err)
		return
	}
	auditRecorder.AddMetadata(map[string]any{
		"tenant_id":   product.TenantID,
		"product_key": product.ProductKey,
	})

	writeJSON(w, http.StatusOK, response.SuccessData("product updated", http.StatusOK, product))
}

func (h *ProductHandler) delete(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		writeHTTPError(w, httperror.ErrUnauthorized)
		return
	}

	auditRecorder, _ := middleware.AuditFromContext(r.Context())
	auditRecorder.Set("product.delete", "product", r.PathValue("product_id"), nil)

	if err := h.service.Delete(r.Context(), service.ProductDeleteInput{
		UserID:    principal.UserID,
		ProductID: r.PathValue("product_id"),
	}); err != nil {
		auditRecorder.SetErrorCode(mapProductError(err).Code)
		h.writeProductError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response.SuccessData("product deleted", http.StatusOK, map[string]bool{"deleted": true}))
}

func (h *ProductHandler) writeProductError(w http.ResponseWriter, err error) {
	httpErr := mapProductError(err)
	if httpErr.HTTPStatus >= http.StatusInternalServerError {
		h.log.Error("product request failed", "err", err)
	}
	writeHTTPError(w, httpErr)
}

func mapProductError(err error) *httperror.Error {
	switch {
	case errors.Is(err, service.ErrInvalidProductInput):
		return httperror.ErrInvalidRequest
	case errors.Is(err, service.ErrProductAlreadyExists):
		return httperror.ErrProductAlreadyExists
	case errors.Is(err, service.ErrProductNotFound):
		return httperror.ErrNotFound
	default:
		return httperror.From(err)
	}
}
