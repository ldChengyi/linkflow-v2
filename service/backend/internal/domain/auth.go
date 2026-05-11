package domain

import (
	"errors"
	"time"
)

var ErrAuthSessionNotFound = errors.New("auth session not found")

type Principal struct {
	TokenID string
	UserID  string
	Role    string
}

type AuthSession struct {
	TokenID   string
	UserID    string
	Role      string
	ExpiresAt time.Time
}

type IssuedAccessToken struct {
	Raw       string
	TokenID   string
	UserID    string
	Role      string
	ExpiresAt time.Time
}
