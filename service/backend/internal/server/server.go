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
	Authenticate middleware.Middleware
}

func New(cfg config.Config, log *slog.Logger, opt Options) (*http.Server, error) {
	if cfg.HTTPAddr == "" {
		return nil, fmt.Errorf("http addr is required")
	}
	if log == nil {
		log = slog.Default()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", health)
	mux.HandleFunc("GET /api/v1/healthz", health)
	mux.HandleFunc("GET /internal/emqx/healthz", health)
	if opt.Auth != nil {
		opt.Auth.RegisterRoutes(mux, opt.Authenticate)
	}

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
