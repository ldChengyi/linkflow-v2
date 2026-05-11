package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/ldchengyi/linkflow-v2/service/backend/internal/auth/token"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/domain"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/response"
)

type principalContextKey struct{}

type AccessSessionValidator interface {
	ValidateAccessSession(ctx context.Context, tokenID string) (domain.AuthSession, error)
}

type AccessTokenVerifier interface {
	VerifyAccessToken(raw string) (*token.AccessClaims, error)
}

func Authenticate(tokens AccessTokenVerifier, sessions AccessSessionValidator) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if tokens == nil || sessions == nil {
				writeUnauthorized(w)
				return
			}

			raw, ok := bearerToken(r.Header.Get("Authorization"))
			if !ok {
				writeUnauthorized(w)
				return
			}

			claims, err := tokens.VerifyAccessToken(raw)
			if err != nil {
				writeUnauthorized(w)
				return
			}

			session, err := sessions.ValidateAccessSession(r.Context(), claims.ID)
			if err != nil {
				if errors.Is(err, domain.ErrAuthSessionNotFound) {
					writeUnauthorized(w)
					return
				}
				writeUnauthorized(w)
				return
			}
			if session.UserID != claims.UserID || session.Role != claims.Role {
				writeUnauthorized(w)
				return
			}

			principal := domain.Principal{
				TokenID: session.TokenID,
				UserID:  session.UserID,
				Role:    session.Role,
			}
			next.ServeHTTP(w, r.WithContext(ContextWithPrincipal(r.Context(), principal)))
		})
	}
}

func ContextWithPrincipal(ctx context.Context, principal domain.Principal) context.Context {
	return context.WithValue(ctx, principalContextKey{}, principal)
}

func PrincipalFromContext(ctx context.Context) (domain.Principal, bool) {
	principal, ok := ctx.Value(principalContextKey{}).(domain.Principal)
	return principal, ok
}

func bearerToken(header string) (string, bool) {
	scheme, raw, ok := strings.Cut(strings.TrimSpace(header), " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return "", false
	}
	raw = strings.TrimSpace(raw)
	return raw, raw != ""
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(response.Fail("unauthorized", http.StatusUnauthorized))
}
