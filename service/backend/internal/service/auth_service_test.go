package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ldchengyi/linkflow-v2/service/backend/internal/domain"
)

type fakeAuthUsers struct {
	user        AuthUser
	createErr   error
	findErr     error
	markLoginID string
}

func (f *fakeAuthUsers) CreateUser(ctx context.Context, email string, passwordHash string, role string) (AuthUser, error) {
	if f.createErr != nil {
		return AuthUser{}, f.createErr
	}
	f.user = AuthUser{
		ID:           "user-1",
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
		Status:       activeUserStatus,
	}
	return f.user, nil
}

func (f *fakeAuthUsers) FindUserByEmail(ctx context.Context, email string) (AuthUser, error) {
	if f.findErr != nil {
		return AuthUser{}, f.findErr
	}
	return f.user, nil
}

func (f *fakeAuthUsers) MarkLogin(ctx context.Context, userID string) error {
	f.markLoginID = userID
	return nil
}

type fakeAuthPassword struct {
	compareErr error
}

func (f fakeAuthPassword) Hash(password string) (string, error) {
	return "hashed:" + password, nil
}

func (f fakeAuthPassword) Compare(hash string, password string) error {
	return f.compareErr
}

type fakeAuthTokens struct{}

func (fakeAuthTokens) IssueAccessToken(userID string, role string) (domain.IssuedAccessToken, error) {
	return domain.IssuedAccessToken{
		Raw:       "token:" + userID + ":" + role,
		TokenID:   "token-id-1",
		UserID:    userID,
		Role:      role,
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}, nil
}

type fakeAuthSessions struct {
	session        domain.AuthSession
	deletedTokenID string
	err            error
}

func (f *fakeAuthSessions) CreateAccessSession(ctx context.Context, session domain.AuthSession) error {
	if f.err != nil {
		return f.err
	}
	f.session = session
	return nil
}

func (f *fakeAuthSessions) DeleteAccessSession(ctx context.Context, tokenID string) error {
	if f.err != nil {
		return f.err
	}
	f.deletedTokenID = tokenID
	return nil
}

func TestAuthServiceRegister(t *testing.T) {
	users := &fakeAuthUsers{}
	svc := newTestAuthService(t, users, fakeAuthPassword{})

	result, err := svc.Register(context.Background(), AuthRegisterInput{
		Email:    " USER@example.COM ",
		Password: "secret",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if result.Email != "user@example.com" {
		t.Fatalf("Email = %q, want user@example.com", result.Email)
	}
	if result.Role != defaultAuthUserRole {
		t.Fatalf("Role = %q, want %q", result.Role, defaultAuthUserRole)
	}
	if result.AccessToken == "" {
		t.Fatal("AccessToken is empty")
	}
}

func TestAuthServiceLogin(t *testing.T) {
	sessions := &fakeAuthSessions{}
	users := &fakeAuthUsers{
		user: AuthUser{
			ID:           "user-1",
			Email:        "user@example.com",
			PasswordHash: "hash",
			Role:         defaultAuthUserRole,
			Status:       activeUserStatus,
		},
	}
	svc := newTestAuthServiceWithSessions(t, users, sessions, fakeAuthPassword{})

	result, err := svc.Login(context.Background(), AuthLoginInput{
		Email:    "user@example.com",
		Password: "secret",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if result.UserID != "user-1" {
		t.Fatalf("UserID = %q, want user-1", result.UserID)
	}
	if users.markLoginID != "user-1" {
		t.Fatalf("markLoginID = %q, want user-1", users.markLoginID)
	}
	if sessions.session.TokenID != "token-id-1" {
		t.Fatalf("session TokenID = %q, want token-id-1", sessions.session.TokenID)
	}
	if sessions.session.UserID != "user-1" {
		t.Fatalf("session UserID = %q, want user-1", sessions.session.UserID)
	}
}

func TestAuthServiceLoginRejectsBadPassword(t *testing.T) {
	users := &fakeAuthUsers{
		user: AuthUser{
			ID:           "user-1",
			Email:        "user@example.com",
			PasswordHash: "hash",
			Role:         defaultAuthUserRole,
			Status:       activeUserStatus,
		},
	}
	svc := newTestAuthService(t, users, fakeAuthPassword{compareErr: errors.New("mismatch")})

	if _, err := svc.Login(context.Background(), AuthLoginInput{
		Email:    "user@example.com",
		Password: "bad",
	}); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want ErrInvalidCredentials", err)
	}
}

func TestAuthServiceLoginRejectsDisabledUser(t *testing.T) {
	users := &fakeAuthUsers{
		user: AuthUser{
			ID:           "user-1",
			Email:        "user@example.com",
			PasswordHash: "hash",
			Role:         defaultAuthUserRole,
			Status:       disabledUserStatus,
		},
	}
	svc := newTestAuthService(t, users, fakeAuthPassword{})

	if _, err := svc.Login(context.Background(), AuthLoginInput{
		Email:    "user@example.com",
		Password: "secret",
	}); !errors.Is(err, ErrUserDisabled) {
		t.Fatalf("Login() error = %v, want ErrUserDisabled", err)
	}
}

func TestAuthServiceLogoutDeletesAccessSession(t *testing.T) {
	sessions := &fakeAuthSessions{}
	svc := newTestAuthServiceWithSessions(t, &fakeAuthUsers{}, sessions, fakeAuthPassword{})

	if err := svc.Logout(context.Background(), AuthLogoutInput{
		TokenID: "token-id-1",
		UserID:  "user-1",
	}); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}

	if sessions.deletedTokenID != "token-id-1" {
		t.Fatalf("deletedTokenID = %q, want token-id-1", sessions.deletedTokenID)
	}
}

func TestAuthServiceLogoutValidatesInput(t *testing.T) {
	svc := newTestAuthService(t, &fakeAuthUsers{}, fakeAuthPassword{})

	if err := svc.Logout(context.Background(), AuthLogoutInput{
		UserID: "user-1",
	}); !errors.Is(err, ErrInvalidAuthInput) {
		t.Fatalf("Logout() error = %v, want ErrInvalidAuthInput", err)
	}
	if err := svc.Logout(context.Background(), AuthLogoutInput{
		TokenID: "token-id-1",
	}); !errors.Is(err, ErrInvalidAuthInput) {
		t.Fatalf("Logout() error = %v, want ErrInvalidAuthInput", err)
	}
}

func newTestAuthService(t *testing.T, users AuthUserStore, password AuthPasswordHasher) *AuthService {
	t.Helper()

	return newTestAuthServiceWithSessions(t, users, &fakeAuthSessions{}, password)
}

func newTestAuthServiceWithSessions(t *testing.T, users AuthUserStore, sessions AuthSessionStore, password AuthPasswordHasher) *AuthService {
	t.Helper()

	svc, err := NewAuthService(users, sessions, password, fakeAuthTokens{})
	if err != nil {
		t.Fatalf("NewAuthService() error = %v", err)
	}
	return svc
}
