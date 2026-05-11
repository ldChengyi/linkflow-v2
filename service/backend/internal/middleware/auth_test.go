package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/ldchengyi/linkflow-v2/service/backend/internal/auth/token"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/domain"
)

type fakeAccessTokenVerifier struct {
	claims *token.AccessClaims
	err    error
}

func (f fakeAccessTokenVerifier) VerifyAccessToken(raw string) (*token.AccessClaims, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.claims, nil
}

type fakeAccessSessionValidator struct {
	session domain.AuthSession
	err     error
}

func (f fakeAccessSessionValidator) ValidateAccessSession(ctx context.Context, tokenID string) (domain.AuthSession, error) {
	if f.err != nil {
		return domain.AuthSession{}, f.err
	}
	return f.session, nil
}

func TestAuthenticateAddsPrincipal(t *testing.T) {
	mw := Authenticate(
		fakeAccessTokenVerifier{
			claims: &token.AccessClaims{
				UserID: "user-1",
				Role:   "user",
				RegisteredClaims: jwt.RegisteredClaims{
					ID: "token-id-1",
				},
			},
		},
		fakeAccessSessionValidator{
			session: domain.AuthSession{
				TokenID:   "token-id-1",
				UserID:    "user-1",
				Role:      "user",
				ExpiresAt: time.Now().UTC().Add(time.Hour),
			},
		},
	)

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		principal, ok := PrincipalFromContext(r.Context())
		if !ok {
			t.Fatal("principal missing from context")
		}
		if principal.UserID != "user-1" {
			t.Fatalf("principal UserID = %q, want user-1", principal.UserID)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer raw-token")
	rec := httptest.NewRecorder()

	mw(next).ServeHTTP(rec, req)

	if !called {
		t.Fatal("next handler was not called")
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestAuthenticateRejectsMissingBearerToken(t *testing.T) {
	mw := Authenticate(fakeAccessTokenVerifier{}, fakeAccessSessionValidator{})
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()

	mw(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
