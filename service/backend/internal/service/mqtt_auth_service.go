package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidMQTTAuthInput = errors.New("invalid mqtt auth input")
	ErrInvalidMQTTAuth      = errors.New("invalid mqtt credentials")
)

type MQTTAuthInput struct {
	TenantSlug string
	ProductKey string
	DeviceSlug string
	Password   string
}

type MQTTAuthResult struct {
	TenantID   string `json:"tenant_id"`
	ProductID  string `json:"product_id"`
	DeviceID   string `json:"device_id"`
	TenantSlug string `json:"tenant_slug"`
	ProductKey string `json:"product_key"`
	DeviceSlug string `json:"device_slug"`
}

type MQTTDeviceCredential struct {
	TenantID         string
	ProductID        string
	DeviceID         string
	TenantSlug       string
	ProductKey       string
	DeviceSlug       string
	ProductAuthType  string
	TenantStatus     string
	ProductStatus    string
	DeviceStatus     string
	CredentialStatus string
	SecretHash       string
}

type MQTTAuthStore interface {
	FindMQTTDeviceCredential(ctx context.Context, in MQTTAuthInput) (MQTTDeviceCredential, error)
}

type DeviceSecretVerifier interface {
	Compare(hash string, secret string) error
}

type MQTTAuthService struct {
	devices MQTTAuthStore
	secrets DeviceSecretVerifier
}

func NewMQTTAuthService(devices MQTTAuthStore, secrets DeviceSecretVerifier) (*MQTTAuthService, error) {
	if devices == nil {
		return nil, fmt.Errorf("mqtt auth store is nil")
	}
	if secrets == nil {
		return nil, fmt.Errorf("device secret verifier is nil")
	}
	return &MQTTAuthService{devices: devices, secrets: secrets}, nil
}

func (s *MQTTAuthService) Authenticate(ctx context.Context, in MQTTAuthInput) (MQTTAuthResult, error) {
	if err := ctx.Err(); err != nil {
		return MQTTAuthResult{}, err
	}
	in.TenantSlug = normalizeTenantSlug(in.TenantSlug)
	in.ProductKey = normalizeProductKey(in.ProductKey)
	in.DeviceSlug = normalizeDeviceSlug(in.DeviceSlug)
	in.Password = strings.TrimSpace(in.Password)
	if in.TenantSlug == "" || in.DeviceSlug == "" {
		return MQTTAuthResult{}, ErrInvalidMQTTAuthInput
	}

	cred, err := s.devices.FindMQTTDeviceCredential(ctx, in)
	if err != nil {
		if errors.Is(err, ErrDeviceNotFound) {
			return MQTTAuthResult{}, ErrInvalidMQTTAuth
		}
		return MQTTAuthResult{}, err
	}
	if cred.TenantStatus != activeProductStatus || cred.ProductStatus != activeProductStatus || cred.DeviceStatus != activeDeviceStatus {
		return MQTTAuthResult{}, ErrInvalidMQTTAuth
	}

	switch cred.ProductAuthType {
	case defaultProductAuthType:
		if in.Password == "" || cred.CredentialStatus != "active" {
			return MQTTAuthResult{}, ErrInvalidMQTTAuth
		}
		if err := s.secrets.Compare(cred.SecretHash, in.Password); err != nil {
			return MQTTAuthResult{}, ErrInvalidMQTTAuth
		}
	case anonymousProductAuthType:
	case certificateProductAuthType:
		return MQTTAuthResult{}, ErrInvalidMQTTAuth
	default:
		return MQTTAuthResult{}, ErrInvalidMQTTAuth
	}

	return MQTTAuthResult{
		TenantID:   cred.TenantID,
		ProductID:  cred.ProductID,
		DeviceID:   cred.DeviceID,
		TenantSlug: cred.TenantSlug,
		ProductKey: cred.ProductKey,
		DeviceSlug: cred.DeviceSlug,
	}, nil
}
