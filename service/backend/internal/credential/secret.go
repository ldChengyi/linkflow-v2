package credential

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const defaultSecretBytes = 32

type SecretManager struct {
	cost int
}

func NewSecretManager(cost int) (*SecretManager, error) {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		return nil, fmt.Errorf("bcrypt cost must be between %d and %d", bcrypt.MinCost, bcrypt.MaxCost)
	}
	return &SecretManager{cost: cost}, nil
}

func (m *SecretManager) Generate() (string, error) {
	raw := make([]byte, defaultSecretBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate device secret: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func (m *SecretManager) Hash(secret string) (string, error) {
	if secret == "" {
		return "", fmt.Errorf("device secret is required")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(secret), m.cost)
	if err != nil {
		return "", fmt.Errorf("hash device secret: %w", err)
	}
	return string(hash), nil
}

func (m *SecretManager) Compare(hash string, secret string) error {
	if hash == "" || secret == "" {
		return fmt.Errorf("device secret hash and secret are required")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(secret)); err != nil {
		return fmt.Errorf("compare device secret: %w", err)
	}
	return nil
}
