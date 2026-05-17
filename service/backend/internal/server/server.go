package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ldchengyi/linkflow-v2/service/backend/internal/config"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/handler"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/middleware"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/response"
)

type Options struct {
	Auth         *handler.AuthHandler
	Tenant       *handler.TenantHandler
	Product      *handler.ProductHandler
	ThingsModel  *handler.ThingsModelHandler
	Device       *handler.DeviceHandler
	EMQXAuth     *handler.EMQXAuthHandler
	AuditLog     *handler.AuditHandler
	Authenticate middleware.Middleware
	Audit        middleware.Middleware
}

func New(cfg config.Config, log *slog.Logger, opt Options) (*http.Server, error) {
	if cfg.HTTPAddr == "" {
		return nil, fmt.Errorf("http addr is required")
	}
	if log == nil {
		log = slog.Default()
	}

	mux := http.NewServeMux()
	routes := newRouteRecorder(mux)
	routes.HandleFunc("GET /healthz", health)
	routes.HandleFunc("GET /api/v1/healthz", health)
	routes.HandleFunc("GET /internal/emqx/healthz", health)
	if opt.Auth != nil {
		opt.Auth.RegisterRoutes(routes, opt.Authenticate)
	}
	if opt.Tenant != nil {
		opt.Tenant.RegisterRoutes(routes, opt.Authenticate, opt.Audit)
	}
	if opt.Product != nil {
		opt.Product.RegisterRoutes(routes, opt.Authenticate, opt.Audit)
	}
	if opt.ThingsModel != nil {
		opt.ThingsModel.RegisterRoutes(routes, opt.Authenticate, opt.Audit)
	}
	if opt.Device != nil {
		opt.Device.RegisterRoutes(routes, opt.Authenticate, opt.Audit)
	}
	if opt.EMQXAuth != nil {
		opt.EMQXAuth.RegisterRoutes(routes)
	}
	if opt.AuditLog != nil {
		opt.AuditLog.RegisterRoutes(routes, opt.Authenticate)
	}
	routes.Log(log)

	handler := middleware.Chain(
		mux,
		middleware.Recover(log),
		middleware.LogRequests(log),
		middleware.Timeout(cfg.HTTPRequestTimeout),
	)

	return &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      handler,
		ReadTimeout:  cfg.HTTPReadTimeout,
		WriteTimeout: cfg.HTTPWriteTimeout,
		IdleTimeout:  cfg.HTTPIdleTimeout,
	}, nil
}

func health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, response.SuccessData("ok", http.StatusOK, map[string]string{"status": "ok"}))
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
