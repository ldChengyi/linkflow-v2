package token

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	testIssuer = "linkflow-backend"
	testSecret = "01234567890123456789012345678901"
)

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

type fixedTokenIDGenerator struct {
	tokenID string
}

func (g fixedTokenIDGenerator) NewTokenID() (string, error) {
	return g.tokenID, nil
}

func TestIssueAndVerifyAccessToken(t *testing.T) {
	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	manager := newTestManager(t, now, time.Hour)

	issued, err := manager.IssueAccessToken("user-1", "user")
	if err != nil {
		t.Fatalf("IssueAccessToken() error = %v", err)
	}

	claims, err := manager.VerifyAccessToken(issued.Raw)
	if err != nil {
		t.Fatalf("VerifyAccessToken() error = %v", err)
	}

	if issued.TokenID != "token-id-1" {
		t.Fatalf("issued TokenID = %q, want token-id-1", issued.TokenID)
	}
	if !issued.ExpiresAt.Equal(now.Add(time.Hour)) {
		t.Fatalf("issued ExpiresAt = %s, want %s", issued.ExpiresAt, now.Add(time.Hour))
	}
	if claims.ID != "token-id-1" {
		t.Fatalf("claims ID = %q, want token-id-1", claims.ID)
	}
	if claims.UserID != "user-1" {
		t.Fatalf("UserID = %q, want user-1", claims.UserID)
	}
	if claims.Subject != "user-1" {
		t.Fatalf("Subject = %q, want user-1", claims.Subject)
	}
	if claims.Role != "user" {
		t.Fatalf("Role = %q, want user", claims.Role)
	}
	if claims.Purpose != accessTokenPurpose {
		t.Fatalf("Purpose = %q, want %q", claims.Purpose, accessTokenPurpose)
	}
	if !claims.ExpiresAt.Time.Equal(now.Add(time.Hour)) {
		t.Fatalf("ExpiresAt = %s, want %s", claims.ExpiresAt.Time, now.Add(time.Hour))
	}
}

func TestVerifyAccessTokenRejectsWrongSecret(t *testing.T) {
	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	manager := newTestManager(t, now, time.Hour)
	issued, err := manager.IssueAccessToken("user-1", "user")
	if err != nil {
		t.Fatalf("IssueAccessToken() error = %v", err)
	}

	other, err := NewJWTManager(Config{
		Secret: "abcdefghijklmnopqrstuvwxyz123456",
		TTL:    time.Hour,
		Issuer: testIssuer,
		Clock:  fixedClock{now: now},
	})
	if err != nil {
		t.Fatalf("NewJWTManager() error = %v", err)
	}

	if _, err := other.VerifyAccessToken(issued.Raw); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("VerifyAccessToken() error = %v, want ErrInvalidToken", err)
	}
}

func TestVerifyAccessTokenRejectsExpiredToken(t *testing.T) {
	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	manager := newTestManager(t, now, time.Hour)
	issued, err := manager.IssueAccessToken("user-1", "user")
	if err != nil {
		t.Fatalf("IssueAccessToken() error = %v", err)
	}

	manager.clock = fixedClock{now: now.Add(2 * time.Hour)}

	if _, err := manager.VerifyAccessToken(issued.Raw); !errors.Is(err, ErrExpiredToken) {
		t.Fatalf("VerifyAccessToken() error = %v, want ErrExpiredToken", err)
	}
}

func TestVerifyAccessTokenRejectsUnexpectedSigningMethod(t *testing.T) {
	manager := newTestManager(t, time.Now().UTC(), time.Hour)

	raw, err := jwt.NewWithClaims(jwt.SigningMethodNone, AccessClaims{
		UserID:  "user-1",
		Role:    "user",
		Purpose: accessTokenPurpose,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-1",
			ID:        "token-id-1",
			Issuer:    testIssuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	if _, err := manager.VerifyAccessToken(raw); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("VerifyAccessToken() error = %v, want ErrInvalidToken", err)
	}
}

func TestNewJWTManagerValidatesConfig(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
	}{
		{
			name: "short secret",
			cfg: Config{
				Secret: "short",
				TTL:    time.Hour,
				Issuer: testIssuer,
			},
		},
		{
			name: "zero ttl",
			cfg: Config{
				Secret: testSecret,
				Issuer: testIssuer,
			},
		},
		{
			name: "empty issuer",
			cfg: Config{
				Secret: testSecret,
				TTL:    time.Hour,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewJWTManager(tt.cfg); err == nil {
				t.Fatal("NewJWTManager() error = nil, want error")
			}
		})
	}
}

func TestIssueAccessTokenValidatesInput(t *testing.T) {
	manager := newTestManager(t, time.Now().UTC(), time.Hour)

	if _, err := manager.IssueAccessToken("", "user"); err == nil {
		t.Fatal("IssueAccessToken() user id error = nil, want error")
	}
	if _, err := manager.IssueAccessToken("user-1", ""); err == nil {
		t.Fatal("IssueAccessToken() role error = nil, want error")
	}
}

func TestVerifyAccessTokenRejectsEmptyToken(t *testing.T) {
	manager := newTestManager(t, time.Now().UTC(), time.Hour)

	if _, err := manager.VerifyAccessToken(strings.TrimSpace("")); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("VerifyAccessToken() error = %v, want ErrInvalidToken", err)
	}
}

func newTestManager(t *testing.T, now time.Time, ttl time.Duration) *JWTManager {
	t.Helper()

	manager, err := NewJWTManager(Config{
		Secret:           testSecret,
		TTL:              ttl,
		Issuer:           testIssuer,
		Clock:            fixedClock{now: now},
		TokenIDGenerator: fixedTokenIDGenerator{tokenID: "token-id-1"},
	})
	if err != nil {
		t.Fatalf("NewJWTManager() error = %v", err)
	}
	return manager
}
