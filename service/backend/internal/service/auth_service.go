package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ldchengyi/linkflow-v2/service/backend/internal/domain"
)

var (
	ErrInvalidAuthInput   = errors.New("invalid auth input")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserDisabled       = errors.New("user disabled")
)

const (
	defaultAuthUserRole = "user"
	activeUserStatus    = "active"
	disabledUserStatus  = "disabled"
)

type AuthUser struct {
	ID           string
	Email        string
	PasswordHash string
	Role         string
	Status       string
}

type AuthRegisterInput struct {
	Email    string
	Password string
}

type AuthLoginInput struct {
	Email    string
	Password string
}

type AuthLogoutInput struct {
	TokenID string
	UserID  string
}

type AuthResult struct {
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type AuthUserStore interface {
	CreateUser(ctx context.Context, email string, passwordHash string, role string) (AuthUser, error)
	FindUserByEmail(ctx context.Context, email string) (AuthUser, error)
	MarkLogin(ctx context.Context, userID string) error
}

type AuthSessionStore interface {
	CreateAccessSession(ctx context.Context, session domain.AuthSession) error
	DeleteAccessSession(ctx context.Context, tokenID string) error
}

type AuthPasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash string, password string) error
}

type AuthTokenIssuer interface {
	IssueAccessToken(userID string, role string) (domain.IssuedAccessToken, error)
}

type AuthService struct {
	users    AuthUserStore
	sessions AuthSessionStore
	password AuthPasswordHasher
	tokens   AuthTokenIssuer
}

func NewAuthService(users AuthUserStore, sessions AuthSessionStore, password AuthPasswordHasher, tokens AuthTokenIssuer) (*AuthService, error) {
	if users == nil {
		return nil, fmt.Errorf("auth user store is nil")
	}
	if sessions == nil {
		return nil, fmt.Errorf("auth session store is nil")
	}
	if password == nil {
		return nil, fmt.Errorf("auth password hasher is nil")
	}
	if tokens == nil {
		return nil, fmt.Errorf("auth token issuer is nil")
	}

	return &AuthService{
		users:    users,
		sessions: sessions,
		password: password,
		tokens:   tokens,
	}, nil
}

func (s *AuthService) Register(ctx context.Context, in AuthRegisterInput) (AuthResult, error) {
	if err := ctx.Err(); err != nil {
		return AuthResult{}, err
	}

	email := normalizeAuthEmail(in.Email)
	if email == "" || in.Password == "" {
		return AuthResult{}, ErrInvalidAuthInput
	}

	hash, err := s.password.Hash(in.Password)
	if err != nil {
		return AuthResult{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.users.CreateUser(ctx, email, hash, defaultAuthUserRole)
	if err != nil {
		return AuthResult{}, err
	}

	return s.issue(ctx, user)
}

func (s *AuthService) Login(ctx context.Context, in AuthLoginInput) (AuthResult, error) {
	if err := ctx.Err(); err != nil {
		return AuthResult{}, err
	}

	email := normalizeAuthEmail(in.Email)
	if email == "" || in.Password == "" {
		return AuthResult{}, ErrInvalidAuthInput
	}

	user, err := s.users.FindUserByEmail(ctx, email)
	if err != nil {
		return AuthResult{}, err
	}
	if user.Status == disabledUserStatus {
		return AuthResult{}, ErrUserDisabled
	}
	if user.Status != activeUserStatus {
		return AuthResult{}, ErrInvalidCredentials
	}
	if err := s.password.Compare(user.PasswordHash, in.Password); err != nil {
		return AuthResult{}, ErrInvalidCredentials
	}
	if err := s.users.MarkLogin(ctx, user.ID); err != nil {
		return AuthResult{}, err
	}

	return s.issue(ctx, user)
}

func (s *AuthService) Logout(ctx context.Context, in AuthLogoutInput) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if strings.TrimSpace(in.TokenID) == "" || strings.TrimSpace(in.UserID) == "" {
		return ErrInvalidAuthInput
	}
	if err := s.sessions.DeleteAccessSession(ctx, in.TokenID); err != nil {
		return fmt.Errorf("delete access session: %w", err)
	}
	return nil
}

func (s *AuthService) issue(ctx context.Context, user AuthUser) (AuthResult, error) {
	issued, err := s.tokens.IssueAccessToken(user.ID, user.Role)
	if err != nil {
		return AuthResult{}, fmt.Errorf("issue access token: %w", err)
	}
	if err := s.sessions.CreateAccessSession(ctx, domain.AuthSession{
		TokenID:   issued.TokenID,
		UserID:    issued.UserID,
		Role:      issued.Role,
		ExpiresAt: issued.ExpiresAt,
	}); err != nil {
		return AuthResult{}, fmt.Errorf("create access session: %w", err)
	}

	return AuthResult{
		UserID:      user.ID,
		Email:       user.Email,
		Role:        user.Role,
		AccessToken: issued.Raw,
		TokenType:   "Bearer",
	}, nil
}

func normalizeAuthEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
