package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidDeviceInput  = errors.New("invalid device input")
	ErrDeviceAlreadyExists = errors.New("device already exists")
	ErrDeviceNotFound      = errors.New("device not found")
)

const (
	activeDeviceStatus      = "active"
	disabledDeviceStatus    = "disabled"
	offlineDeviceConnection = "offline"
	defaultDeviceStatus     = activeDeviceStatus
	defaultDeviceConnection = offlineDeviceConnection
)

type Device struct {
	ID               string     `json:"id"`
	TenantID         string     `json:"tenant_id"`
	ProductID        string     `json:"product_id"`
	DeviceSlug       string     `json:"device_slug"`
	DeviceName       string     `json:"device_name"`
	Description      string     `json:"description"`
	Status           string     `json:"status"`
	ConnectionStatus string     `json:"connection_status"`
	GatewayDeviceID  string     `json:"gateway_device_id,omitempty"`
	FirmwareVersion  string     `json:"firmware_version,omitempty"`
	IPAddress        string     `json:"ip_address,omitempty"`
	LastSeenAt       *time.Time `json:"last_seen_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type DeviceCreateInput struct {
	UserID           string
	TenantID         string
	ProductID        string
	DeviceSlug       string
	DeviceName       string
	Description      string
	GatewayDeviceID  string
	DeviceSecretHash string
}

type DeviceCreateResult struct {
	Device       Device `json:"device"`
	DeviceSecret string `json:"device_secret,omitempty"`
}

type DeviceUpdateInput struct {
	UserID          string
	DeviceID        string
	DeviceName      string
	Description     string
	Status          string
	GatewayDeviceID string
}

type DeviceGetInput struct {
	UserID   string
	DeviceID string
}

type DeviceDeleteInput struct {
	UserID   string
	DeviceID string
}

type DeviceListInput struct {
	UserID    string
	TenantID  string
	ProductID string
	PageInput
}

type DeviceLatestPropertiesInput struct {
	UserID   string
	DeviceID string
}

type DeviceLatestProperties struct {
	ID         string         `json:"id"`
	TenantID   string         `json:"tenant_id"`
	ProductID  string         `json:"product_id"`
	ProductKey string         `json:"product_key"`
	DeviceSlug string         `json:"device_slug"`
	Reported   bool           `json:"reported"`
	Properties map[string]any `json:"properties"`
	OccurredAt *time.Time     `json:"occurred_at,omitempty"`
	ReceivedAt *time.Time     `json:"received_at,omitempty"`
}

type DeviceStore interface {
	FindDeviceProductAuthType(ctx context.Context, in DeviceProductAuthInput) (string, error)
	CreateDevice(ctx context.Context, in DeviceCreateInput) (Device, error)
	ListDevices(ctx context.Context, in DeviceListInput) (PageResult[Device], error)
	FindDeviceByID(ctx context.Context, in DeviceGetInput) (Device, error)
	FindDeviceLatestProperties(ctx context.Context, in DeviceLatestPropertiesInput) (DeviceLatestProperties, error)
	UpdateDevice(ctx context.Context, in DeviceUpdateInput) (Device, error)
	DeleteDevice(ctx context.Context, in DeviceDeleteInput) error
}

type DeviceProductAuthInput struct {
	UserID    string
	TenantID  string
	ProductID string
}

type DeviceSecretManager interface {
	Generate() (string, error)
	Hash(secret string) (string, error)
}

type DeviceService struct {
	devices DeviceStore
	secrets DeviceSecretManager
}

func NewDeviceService(devices DeviceStore, secrets DeviceSecretManager) (*DeviceService, error) {
	if devices == nil {
		return nil, fmt.Errorf("device store is nil")
	}
	if secrets == nil {
		return nil, fmt.Errorf("device secret manager is nil")
	}
	return &DeviceService{devices: devices, secrets: secrets}, nil
}

func (s *DeviceService) Create(ctx context.Context, in DeviceCreateInput) (DeviceCreateResult, error) {
	if err := ctx.Err(); err != nil {
		return DeviceCreateResult{}, err
	}
	in.UserID = strings.TrimSpace(in.UserID)
	in.TenantID = strings.TrimSpace(in.TenantID)
	in.ProductID = strings.TrimSpace(in.ProductID)
	in.DeviceSlug = normalizeDeviceSlug(in.DeviceSlug)
	in.DeviceName = strings.TrimSpace(in.DeviceName)
	in.Description = strings.TrimSpace(in.Description)
	in.GatewayDeviceID = strings.TrimSpace(in.GatewayDeviceID)
	if in.UserID == "" || in.TenantID == "" || in.ProductID == "" || in.DeviceSlug == "" || in.DeviceName == "" {
		return DeviceCreateResult{}, ErrInvalidDeviceInput
	}

	authType, err := s.devices.FindDeviceProductAuthType(ctx, DeviceProductAuthInput{
		UserID:    in.UserID,
		TenantID:  in.TenantID,
		ProductID: in.ProductID,
	})
	if err != nil {
		return DeviceCreateResult{}, err
	}

	var plainSecret string
	switch authType {
	case defaultProductAuthType:
		plainSecret, err = s.secrets.Generate()
		if err != nil {
			return DeviceCreateResult{}, err
		}
		in.DeviceSecretHash, err = s.secrets.Hash(plainSecret)
		if err != nil {
			return DeviceCreateResult{}, err
		}
	case anonymousProductAuthType:
	case certificateProductAuthType:
		return DeviceCreateResult{}, ErrInvalidDeviceInput
	default:
		return DeviceCreateResult{}, ErrInvalidDeviceInput
	}

	device, err := s.devices.CreateDevice(ctx, in)
	if err != nil {
		return DeviceCreateResult{}, err
	}
	return DeviceCreateResult{
		Device:       device,
		DeviceSecret: plainSecret,
	}, nil
}

func (s *DeviceService) List(ctx context.Context, in DeviceListInput) (PageResult[Device], error) {
	if err := ctx.Err(); err != nil {
		return PageResult[Device]{}, err
	}
	in.UserID = strings.TrimSpace(in.UserID)
	in.TenantID = strings.TrimSpace(in.TenantID)
	in.ProductID = strings.TrimSpace(in.ProductID)
	if in.UserID == "" || in.TenantID == "" {
		return PageResult[Device]{}, ErrInvalidDeviceInput
	}
	in.PageInput = NormalizePageInput(in.PageInput)
	return s.devices.ListDevices(ctx, in)
}

func (s *DeviceService) Get(ctx context.Context, in DeviceGetInput) (Device, error) {
	if err := ctx.Err(); err != nil {
		return Device{}, err
	}
	in.UserID = strings.TrimSpace(in.UserID)
	in.DeviceID = strings.TrimSpace(in.DeviceID)
	if in.UserID == "" || in.DeviceID == "" {
		return Device{}, ErrInvalidDeviceInput
	}
	return s.devices.FindDeviceByID(ctx, in)
}

func (s *DeviceService) LatestProperties(ctx context.Context, in DeviceLatestPropertiesInput) (DeviceLatestProperties, error) {
	if err := ctx.Err(); err != nil {
		return DeviceLatestProperties{}, err
	}
	in.UserID = strings.TrimSpace(in.UserID)
	in.DeviceID = strings.TrimSpace(in.DeviceID)
	if in.UserID == "" || in.DeviceID == "" {
		return DeviceLatestProperties{}, ErrInvalidDeviceInput
	}
	return s.devices.FindDeviceLatestProperties(ctx, in)
}

func (s *DeviceService) Update(ctx context.Context, in DeviceUpdateInput) (Device, error) {
	if err := ctx.Err(); err != nil {
		return Device{}, err
	}
	in.UserID = strings.TrimSpace(in.UserID)
	in.DeviceID = strings.TrimSpace(in.DeviceID)
	in.DeviceName = strings.TrimSpace(in.DeviceName)
	in.Description = strings.TrimSpace(in.Description)
	in.Status = normalizeDeviceStatus(in.Status)
	in.GatewayDeviceID = strings.TrimSpace(in.GatewayDeviceID)
	if in.UserID == "" || in.DeviceID == "" || in.DeviceName == "" {
		return Device{}, ErrInvalidDeviceInput
	}
	if in.GatewayDeviceID != "" && in.GatewayDeviceID == in.DeviceID {
		return Device{}, ErrInvalidDeviceInput
	}
	if !validDeviceStatus(in.Status) {
		return Device{}, ErrInvalidDeviceInput
	}
	return s.devices.UpdateDevice(ctx, in)
}

func (s *DeviceService) Delete(ctx context.Context, in DeviceDeleteInput) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	in.UserID = strings.TrimSpace(in.UserID)
	in.DeviceID = strings.TrimSpace(in.DeviceID)
	if in.UserID == "" || in.DeviceID == "" {
		return ErrInvalidDeviceInput
	}
	return s.devices.DeleteDevice(ctx, in)
}

func normalizeDeviceSlug(slug string) string {
	return strings.ToLower(strings.TrimSpace(slug))
}

func normalizeDeviceStatus(status string) string {
	status = strings.TrimSpace(status)
	if status == "" {
		return defaultDeviceStatus
	}
	return status
}

func validDeviceStatus(status string) bool {
	switch status {
	case activeDeviceStatus, disabledDeviceStatus:
		return true
	default:
		return false
	}
}
