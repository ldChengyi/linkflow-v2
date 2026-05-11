package token

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/ldchengyi/linkflow-v2/service/backend/internal/domain"
)

const (
	minSecretLength    = 32
	accessTokenPurpose = "access"
)

var (
	ErrInvalidToken = errors.New("invalid access token")
	ErrExpiredToken = errors.New("expired access token")
)

type Clock interface {
	Now() time.Time
}

type TokenIDGenerator interface {
	NewTokenID() (string, error)
}

type realClock struct{}

func (realClock) Now() time.Time {
	return time.Now().UTC()
}

type randomTokenIDGenerator struct{}

func (randomTokenIDGenerator) NewTokenID() (string, error) {
	var buf [32]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("read random token id: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf[:]), nil
}

type Config struct {
	Secret           string
	TTL              time.Duration
	Issuer           string
	Clock            Clock
	TokenIDGenerator TokenIDGenerator
}

type AccessClaims struct {
	UserID  string `json:"user_id"`
	Role    string `json:"role"`
	Purpose string `json:"purpose"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secret           []byte
	ttl              time.Duration
	issuer           string
	clock            Clock
	tokenIDGenerator TokenIDGenerator
}

func NewJWTManager(cfg Config) (*JWTManager, error) {
	if len(cfg.Secret) < minSecretLength {
		return nil, fmt.Errorf("jwt secret must be at least %d characters", minSecretLength)
	}
	if cfg.TTL <= 0 {
		return nil, fmt.Errorf("jwt access token ttl must be positive")
	}
	if cfg.Issuer == "" {
		return nil, fmt.Errorf("jwt issuer is required")
	}
	if cfg.Clock == nil {
		cfg.Clock = realClock{}
	}
	if cfg.TokenIDGenerator == nil {
		cfg.TokenIDGenerator = randomTokenIDGenerator{}
	}

	return &JWTManager{
		secret:           []byte(cfg.Secret),
		ttl:              cfg.TTL,
		issuer:           cfg.Issuer,
		clock:            cfg.Clock,
		tokenIDGenerator: cfg.TokenIDGenerator,
	}, nil
}

func (m *JWTManager) IssueAccessToken(userID, role string) (domain.IssuedAccessToken, error) {
	if userID == "" {
		return domain.IssuedAccessToken{}, fmt.Errorf("user id is required")
	}
	if role == "" {
		return domain.IssuedAccessToken{}, fmt.Errorf("role is required")
	}

	tokenID, err := m.tokenIDGenerator.NewTokenID()
	if err != nil {
		return domain.IssuedAccessToken{}, fmt.Errorf("new token id: %w", err)
	}

	now := m.clock.Now().UTC()
	expiresAt := now.Add(m.ttl)
	claims := AccessClaims{
		UserID:  userID,
		Role:    role,
		Purpose: accessTokenPurpose,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ID:        tokenID,
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return domain.IssuedAccessToken{}, fmt.Errorf("sign access token: %w", err)
	}

	return domain.IssuedAccessToken{
		Raw:       raw,
		TokenID:   tokenID,
		UserID:    userID,
		Role:      role,
		ExpiresAt: expiresAt,
	}, nil
}

func (m *JWTManager) VerifyAccessToken(raw string) (*AccessClaims, error) {
	if raw == "" {
		return nil, ErrInvalidToken
	}

	claims := AccessClaims{}
	parsed, err := jwt.ParseWithClaims(
		raw,
		&claims,
		func(t *jwt.Token) (any, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected signing method %q", t.Method.Alg())
			}
			return m.secret, nil
		},
		jwt.WithIssuer(m.issuer),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithTimeFunc(func() time.Time {
			return m.clock.Now().UTC()
		}),
		jwt.WithLeeway(0),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	if !parsed.Valid {
		return nil, ErrInvalidToken
	}
	if claims.Subject == "" || claims.UserID == "" || claims.Subject != claims.UserID {
		return nil, fmt.Errorf("%w: missing or inconsistent user id", ErrInvalidToken)
	}
	if claims.ID == "" {
		return nil, fmt.Errorf("%w: missing token id", ErrInvalidToken)
	}
	if claims.Role == "" {
		return nil, fmt.Errorf("%w: missing role", ErrInvalidToken)
	}
	if claims.Purpose != accessTokenPurpose {
		return nil, fmt.Errorf("%w: unexpected token purpose", ErrInvalidToken)
	}

	return &claims, nil
}
