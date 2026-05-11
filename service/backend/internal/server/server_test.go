package server_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ldchengyi/linkflow-v2/service/backend/internal/config"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/server"
)

func TestHealthRoutes(t *testing.T) {
	server, err := server.New(testConfig(), slog.New(slog.NewTextHandler(io.Discard, nil)), server.Options{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	tests := []string{
		"/healthz",
		"/api/v1/healthz",
		"/internal/emqx/healthz",
	}

	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()

			server.Handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
			}

			var body struct {
				Msg  string            `json:"msg"`
				Code int               `json:"code"`
				Data map[string]string `json:"data"`
			}
			if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body.Msg != "ok" {
				t.Fatalf("msg = %q, want ok", body.Msg)
			}
			if body.Code != http.StatusOK {
				t.Fatalf("code = %d, want %d", body.Code, http.StatusOK)
			}
			if body.Data["status"] != "ok" {
				t.Fatalf("status body = %q, want ok", body.Data["status"])
			}
		})
	}
}

func TestNewRequiresHTTPAddr(t *testing.T) {
	cfg := testConfig()
	cfg.HTTPAddr = ""

	if _, err := server.New(cfg, nil, server.Options{}); err == nil {
		t.Fatal("New() error = nil, want error")
	}
}

func testConfig() config.Config {
	return config.Config{
		HTTPAddr:           "127.0.0.1:0",
		HTTPReadTimeout:    time.Second,
		HTTPWriteTimeout:   time.Second,
		HTTPIdleTimeout:    time.Second,
		HTTPRequestTimeout: time.Second,
	}
}
