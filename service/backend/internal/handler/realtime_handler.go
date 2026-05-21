package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/coder/websocket"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/domain"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/middleware"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/realtime"
)

// TenantLookup returns the tenant ids visible to the given owner user. The
// realtime handler uses this at handshake time to bind a WS connection to the
// user's tenants.
type TenantLookup interface {
	ListTenantIDsByOwner(ctx context.Context, userID string) ([]string, error)
}

type RealtimeHandler struct {
	registry *realtime.Registry
	tokens   middleware.AccessTokenVerifier
	sessions middleware.AccessSessionValidator
	tenants  TenantLookup
	log      *slog.Logger
}

func NewRealtimeHandler(
	registry *realtime.Registry,
	tokens middleware.AccessTokenVerifier,
	sessions middleware.AccessSessionValidator,
	tenants TenantLookup,
	log *slog.Logger,
) (*RealtimeHandler, error) {
	if registry == nil {
		return nil, fmt.Errorf("registry is nil")
	}
	if tokens == nil {
		return nil, fmt.Errorf("access token verifier is nil")
	}
	if sessions == nil {
		return nil, fmt.Errorf("access session validator is nil")
	}
	if tenants == nil {
		return nil, fmt.Errorf("tenant lookup is nil")
	}
	if log == nil {
		log = slog.Default()
	}
	return &RealtimeHandler{
		registry: registry,
		tokens:   tokens,
		sessions: sessions,
		tenants:  tenants,
		log:      log,
	}, nil
}

func (h *RealtimeHandler) RegisterRoutes(mux RouteRegistrar) {
	mux.HandleFunc("GET /api/v1/ws/devices", h.serve)
}

func (h *RealtimeHandler) serve(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.authenticate(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	tenantIDs, err := h.tenants.ListTenantIDsByOwner(r.Context(), principal.UserID)
	if err != nil {
		h.log.Error("realtime tenant lookup failed", "user_id", principal.UserID, "err", err)
		http.Error(w, "tenant lookup failed", http.StatusInternalServerError)
		return
	}
	if len(tenantIDs) == 0 {
		http.Error(w, "no tenants", http.StatusForbidden)
		return
	}

	ws, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // same-origin behind nginx; CSRF is mitigated by cookie SameSite=Lax
	})
	if err != nil {
		h.log.Info("realtime accept failed", "user_id", principal.UserID, "err", err)
		return
	}

	conn := realtime.NewConn(ws, tenantIDs, principal.UserID, h.log)
	h.registry.Register(conn)
	defer h.registry.Unregister(conn)

	if err := conn.Run(r.Context()); err != nil && !errors.Is(err, context.Canceled) {
		h.log.Info("realtime conn ended", "user_id", principal.UserID, "err", err)
	}
}

func (h *RealtimeHandler) authenticate(r *http.Request) (domain.Principal, bool) {
	raw, ok := h.extractToken(r)
	if !ok {
		return domain.Principal{}, false
	}

	claims, err := h.tokens.VerifyAccessToken(raw)
	if err != nil {
		return domain.Principal{}, false
	}

	session, err := h.sessions.ValidateAccessSession(r.Context(), claims.ID)
	if err != nil {
		return domain.Principal{}, false
	}
	if session.UserID != claims.UserID || session.Role != claims.Role {
		return domain.Principal{}, false
	}

	return domain.Principal{
		TokenID: session.TokenID,
		UserID:  session.UserID,
		Role:    session.Role,
	}, true
}

func (h *RealtimeHandler) extractToken(r *http.Request) (string, bool) {
	if cookie, err := r.Cookie(realtimeAccessCookieName); err == nil {
		if value := strings.TrimSpace(cookie.Value); value != "" {
			return value, true
		}
	}

	if header := r.Header.Get("Authorization"); header != "" {
		scheme, raw, ok := strings.Cut(strings.TrimSpace(header), " ")
		if ok && strings.EqualFold(scheme, "Bearer") {
			raw = strings.TrimSpace(raw)
			if raw != "" {
				return raw, true
			}
		}
	}

	return "", false
}
