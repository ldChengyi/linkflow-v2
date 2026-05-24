package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

var (
	ErrInvalidDeviceInput                = errors.New("invalid device input")
	ErrDeviceAlreadyExists               = errors.New("device already exists")
	ErrDeviceNotFound                    = errors.New("device not found")
	ErrDeviceOffline                     = errors.New("device is offline")
	ErrDeviceServiceNotFound             = errors.New("device service not found")
	ErrDeviceCommandPublisherUnavailable = errors.New("device command publisher unavailable")
)

const (
	activeDeviceStatus      = "active"
	disabledDeviceStatus    = "disabled"
	offlineDeviceConnection = "offline"
	defaultDeviceStatus     = activeDeviceStatus
	defaultDeviceConnection = offlineDeviceConnection

	defaultTrendBucketSeconds = 60
	minTrendBucketSeconds     = 10
	maxTrendBucketSeconds     = 86400
	maxTrendRange             = 31 * 24 * time.Hour
	maxTrendProperties        = 8
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

type DevicePropertyTrendInput struct {
	UserID        string
	DeviceID      string
	Properties    []string
	From          time.Time
	To            time.Time
	BucketSeconds int
	Aggregate     string
}

type DevicePropertyTrendPoint struct {
	BucketAt time.Time `json:"bucket_at"`
	Value    float64   `json:"value"`
	Min      float64   `json:"min"`
	Max      float64   `json:"max"`
	Count    int       `json:"count"`
}

type DevicePropertyTrendSeries struct {
	Property string                     `json:"property"`
	Points   []DevicePropertyTrendPoint `json:"points"`
}

type DevicePropertyTrend struct {
	DeviceID      string                      `json:"device_id"`
	TenantID      string                      `json:"tenant_id"`
	ProductID     string                      `json:"product_id"`
	ProductKey    string                      `json:"product_key"`
	DeviceSlug    string                      `json:"device_slug"`
	Properties    []string                    `json:"properties"`
	From          time.Time                   `json:"from"`
	To            time.Time                   `json:"to"`
	BucketSeconds int                         `json:"bucket_seconds"`
	Aggregate     string                      `json:"agg"`
	Series        []DevicePropertyTrendSeries `json:"series"`
}

type DeviceEventHistoryInput struct {
	UserID    string
	DeviceID  string
	EventName string
	PageInput
}

type DeviceServiceCallHistoryInput struct {
	UserID             string
	DeviceID           string
	ServiceName        string
	AckDeadlineSeconds int
	PageInput
}

type DevicePropertySetHistoryInput struct {
	UserID             string
	DeviceID           string
	PropertyName       string
	AckDeadlineSeconds int
	PageInput
}

type DevicePropertySetInput struct {
	UserID     string
	DeviceID   string
	Properties map[string]any
}

type DevicePropertySetResult struct {
	CommandID  string         `json:"command_id"`
	Topic      string         `json:"topic"`
	Properties map[string]any `json:"properties"`
	SentAt     time.Time      `json:"sent_at"`
}

type DeviceEventEntry struct {
	EventID    string         `json:"event_id"`
	TenantID   string         `json:"tenant_id"`
	ProductID  string         `json:"product_id"`
	ProductKey string         `json:"product_key"`
	DeviceSlug string         `json:"device_slug"`
	EventName  string         `json:"event_name"`
	Params     map[string]any `json:"params"`
	OccurredAt time.Time      `json:"occurred_at"`
	ReceivedAt time.Time      `json:"received_at"`
}

type DeviceServiceCallHistoryEntry struct {
	CommandID          string         `json:"command_id"`
	TenantID           string         `json:"tenant_id"`
	ProductID          string         `json:"product_id"`
	ProductKey         string         `json:"product_key"`
	DeviceSlug         string         `json:"device_slug"`
	ServiceName        string         `json:"service_name"`
	Topic              string         `json:"topic"`
	Input              map[string]any `json:"input"`
	OccurredAt         time.Time      `json:"occurred_at"`
	AckDeadlineAt      time.Time      `json:"ack_deadline_at"`
	AckStatus          string         `json:"ack_status"`
	AckEventID         string         `json:"ack_event_id,omitempty"`
	AckSuccess         *bool          `json:"ack_success,omitempty"`
	AckCode            string         `json:"ack_code,omitempty"`
	AckMessage         string         `json:"ack_message,omitempty"`
	AckOutput          map[string]any `json:"ack_output,omitempty"`
	AckOccurredAt      *time.Time     `json:"ack_occurred_at,omitempty"`
	AckReceivedAt      *time.Time     `json:"ack_received_at,omitempty"`
	AckDeadlineSeconds int            `json:"ack_deadline_seconds"`
}

type DevicePropertySetHistoryEntry struct {
	CommandID          string         `json:"command_id"`
	TenantID           string         `json:"tenant_id"`
	ProductID          string         `json:"product_id"`
	ProductKey         string         `json:"product_key"`
	DeviceSlug         string         `json:"device_slug"`
	Topic              string         `json:"topic"`
	Properties         map[string]any `json:"properties"`
	OccurredAt         time.Time      `json:"occurred_at"`
	AckDeadlineAt      time.Time      `json:"ack_deadline_at"`
	AckStatus          string         `json:"ack_status"`
	AckEventID         string         `json:"ack_event_id,omitempty"`
	AckSuccess         *bool          `json:"ack_success,omitempty"`
	AckCode            string         `json:"ack_code,omitempty"`
	AckMessage         string         `json:"ack_message,omitempty"`
	AckProperties      map[string]any `json:"ack_properties,omitempty"`
	AckOccurredAt      *time.Time     `json:"ack_occurred_at,omitempty"`
	AckReceivedAt      *time.Time     `json:"ack_received_at,omitempty"`
	AckDeadlineSeconds int            `json:"ack_deadline_seconds"`
}

type DeviceServiceCallInput struct {
	UserID      string
	DeviceID    string
	ServiceName string
	Input       map[string]any
}

type DeviceServiceCallResult struct {
	CommandID   string         `json:"command_id"`
	Topic       string         `json:"topic"`
	ServiceName string         `json:"service_name"`
	Input       map[string]any `json:"input"`
	SentAt      time.Time      `json:"sent_at"`
}

type DeviceStore interface {
	FindDeviceProductAuthType(ctx context.Context, in DeviceProductAuthInput) (string, error)
	CreateDevice(ctx context.Context, in DeviceCreateInput) (Device, error)
	ListDevices(ctx context.Context, in DeviceListInput) (PageResult[Device], error)
	FindDeviceByID(ctx context.Context, in DeviceGetInput) (Device, error)
	FindDeviceLatestProperties(ctx context.Context, in DeviceLatestPropertiesInput) (DeviceLatestProperties, error)
	FindDevicePropertyTrend(ctx context.Context, in DevicePropertyTrendInput) (DevicePropertyTrend, error)
	ListDeviceEventHistory(ctx context.Context, in DeviceEventHistoryInput) (PageResult[DeviceEventEntry], error)
	ListDeviceServiceCallHistory(ctx context.Context, in DeviceServiceCallHistoryInput) (PageResult[DeviceServiceCallHistoryEntry], error)
	ListDevicePropertySetHistory(ctx context.Context, in DevicePropertySetHistoryInput) (PageResult[DevicePropertySetHistoryEntry], error)
	FindDevicePropertySetTarget(ctx context.Context, in DevicePropertySetTargetInput) (DevicePropertySetTarget, error)
	UpdateDevice(ctx context.Context, in DeviceUpdateInput) (Device, error)
	DeleteDevice(ctx context.Context, in DeviceDeleteInput) error
}

type DevicePropertySetTargetInput struct {
	UserID   string
	DeviceID string
}

type DevicePropertySetTarget struct {
	TenantID         string
	ProductID        string
	DeviceID         string
	TenantSlug       string
	ProductKey       string
	DeviceSlug       string
	DeviceStatus     string
	ConnectionStatus string
	Properties       ThingsModelObject
}

type DeviceServiceCallTargetInput struct {
	UserID   string
	DeviceID string
}

type DeviceServiceCallTarget struct {
	TenantID         string
	ProductID        string
	DeviceID         string
	TenantSlug       string
	ProductKey       string
	DeviceSlug       string
	DeviceStatus     string
	ConnectionStatus string
	Services         ThingsModelObject
}

type DeviceServiceCallTargetStore interface {
	FindDeviceServiceCallTarget(ctx context.Context, in DeviceServiceCallTargetInput) (DeviceServiceCallTarget, error)
}

type DeviceServiceCallMessage struct {
	TenantID    string
	ProductID   string
	DeviceID    string
	TenantSlug  string
	ProductKey  string
	DeviceSlug  string
	Protocol    string
	Topic       string
	CommandID   string
	ServiceName string
	RequestedBy string
	Input       map[string]any
	OccurredAt  time.Time
}

type DevicePropertySetMessage struct {
	TenantID    string
	ProductID   string
	DeviceID    string
	TenantSlug  string
	ProductKey  string
	DeviceSlug  string
	Protocol    string
	Topic       string
	CommandID   string
	RequestedBy string
	Properties  map[string]any
	OccurredAt  time.Time
}

type DeviceServiceCallPublisher interface {
	PublishServiceCall(ctx context.Context, in DeviceServiceCallMessage) error
}

type DevicePropertySetPublisher interface {
	PublishPropertySet(ctx context.Context, in DevicePropertySetMessage) error
}

type DeviceProductAuthInput struct {
	UserID    string
	TenantID  string
	ProductID string
}

type DeviceSecretManager interface {
	Generate() (string, error)
	Hash(secret string) (string, error)
	Compare(hash string, secret string) error
}

type DeviceService struct {
	devices              DeviceStore
	secrets              DeviceSecretManager
	serviceCallTargets   DeviceServiceCallTargetStore
	serviceCallPublisher DeviceServiceCallPublisher
	propertySetPublisher DevicePropertySetPublisher
}

type DeviceServiceOption func(*DeviceService)

func WithDeviceServiceCallTargets(targets DeviceServiceCallTargetStore) DeviceServiceOption {
	return func(s *DeviceService) {
		s.serviceCallTargets = targets
	}
}

func WithDeviceServiceCallPublisher(publisher DeviceServiceCallPublisher) DeviceServiceOption {
	return func(s *DeviceService) {
		s.serviceCallPublisher = publisher
	}
}

func WithDevicePropertySetPublisher(publisher DevicePropertySetPublisher) DeviceServiceOption {
	return func(s *DeviceService) {
		s.propertySetPublisher = publisher
	}
}

func NewDeviceService(devices DeviceStore, secrets DeviceSecretManager, options ...DeviceServiceOption) (*DeviceService, error) {
	if devices == nil {
		return nil, fmt.Errorf("device store is nil")
	}
	if secrets == nil {
		return nil, fmt.Errorf("device secret manager is nil")
	}
	svc := &DeviceService{devices: devices, secrets: secrets}
	for _, option := range options {
		if option != nil {
			option(svc)
		}
	}
	return svc, nil
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

func (s *DeviceService) PropertyTrend(ctx context.Context, in DevicePropertyTrendInput) (DevicePropertyTrend, error) {
	if err := ctx.Err(); err != nil {
		return DevicePropertyTrend{}, err
	}
	in.UserID = strings.TrimSpace(in.UserID)
	in.DeviceID = strings.TrimSpace(in.DeviceID)
	in.Aggregate = strings.ToLower(strings.TrimSpace(in.Aggregate))
	if in.Aggregate == "" {
		in.Aggregate = "avg"
	}
	if in.BucketSeconds <= 0 {
		in.BucketSeconds = defaultTrendBucketSeconds
	}
	properties, ok := normalizeTrendProperties(in.Properties)
	if in.UserID == "" || in.DeviceID == "" || !ok || in.From.IsZero() || in.To.IsZero() || !in.From.Before(in.To) {
		return DevicePropertyTrend{}, ErrInvalidDeviceInput
	}
	if in.To.Sub(in.From) > maxTrendRange {
		return DevicePropertyTrend{}, ErrInvalidDeviceInput
	}
	if in.BucketSeconds < minTrendBucketSeconds || in.BucketSeconds > maxTrendBucketSeconds {
		return DevicePropertyTrend{}, ErrInvalidDeviceInput
	}
	if !validTrendAggregate(in.Aggregate) {
		return DevicePropertyTrend{}, ErrInvalidDeviceInput
	}
	in.Properties = properties
	return s.devices.FindDevicePropertyTrend(ctx, in)
}

func (s *DeviceService) EventHistory(ctx context.Context, in DeviceEventHistoryInput) (PageResult[DeviceEventEntry], error) {
	if err := ctx.Err(); err != nil {
		return PageResult[DeviceEventEntry]{}, err
	}
	in.UserID = strings.TrimSpace(in.UserID)
	in.DeviceID = strings.TrimSpace(in.DeviceID)
	in.EventName = strings.TrimSpace(in.EventName)
	if in.UserID == "" || in.DeviceID == "" {
		return PageResult[DeviceEventEntry]{}, ErrInvalidDeviceInput
	}
	in.PageInput = NormalizePageInput(in.PageInput)
	return s.devices.ListDeviceEventHistory(ctx, in)
}

func (s *DeviceService) ServiceCallHistory(ctx context.Context, in DeviceServiceCallHistoryInput) (PageResult[DeviceServiceCallHistoryEntry], error) {
	if err := ctx.Err(); err != nil {
		return PageResult[DeviceServiceCallHistoryEntry]{}, err
	}
	in.UserID = strings.TrimSpace(in.UserID)
	in.DeviceID = strings.TrimSpace(in.DeviceID)
	in.ServiceName = strings.TrimSpace(in.ServiceName)
	if in.UserID == "" || in.DeviceID == "" {
		return PageResult[DeviceServiceCallHistoryEntry]{}, ErrInvalidDeviceInput
	}
	if in.ServiceName != "" && !validThingsModelIdentifier(in.ServiceName) {
		return PageResult[DeviceServiceCallHistoryEntry]{}, ErrInvalidDeviceInput
	}
	if in.AckDeadlineSeconds <= 0 {
		in.AckDeadlineSeconds = 90
	}
	if in.AckDeadlineSeconds > 3600 {
		in.AckDeadlineSeconds = 3600
	}
	in.PageInput = NormalizePageInput(in.PageInput)
	return s.devices.ListDeviceServiceCallHistory(ctx, in)
}

func (s *DeviceService) PropertySetHistory(ctx context.Context, in DevicePropertySetHistoryInput) (PageResult[DevicePropertySetHistoryEntry], error) {
	if err := ctx.Err(); err != nil {
		return PageResult[DevicePropertySetHistoryEntry]{}, err
	}
	in.UserID = strings.TrimSpace(in.UserID)
	in.DeviceID = strings.TrimSpace(in.DeviceID)
	in.PropertyName = strings.TrimSpace(in.PropertyName)
	if in.UserID == "" || in.DeviceID == "" {
		return PageResult[DevicePropertySetHistoryEntry]{}, ErrInvalidDeviceInput
	}
	if in.PropertyName != "" && !validThingsModelIdentifier(in.PropertyName) {
		return PageResult[DevicePropertySetHistoryEntry]{}, ErrInvalidDeviceInput
	}
	if in.AckDeadlineSeconds <= 0 {
		in.AckDeadlineSeconds = 90
	}
	if in.AckDeadlineSeconds > 3600 {
		in.AckDeadlineSeconds = 3600
	}
	in.PageInput = NormalizePageInput(in.PageInput)
	return s.devices.ListDevicePropertySetHistory(ctx, in)
}

func (s *DeviceService) SetProperties(ctx context.Context, in DevicePropertySetInput) (DevicePropertySetResult, error) {
	if err := ctx.Err(); err != nil {
		return DevicePropertySetResult{}, err
	}
	if s.propertySetPublisher == nil {
		return DevicePropertySetResult{}, ErrDeviceCommandPublisherUnavailable
	}
	in.UserID = strings.TrimSpace(in.UserID)
	in.DeviceID = strings.TrimSpace(in.DeviceID)
	if in.UserID == "" || in.DeviceID == "" || len(in.Properties) == 0 {
		return DevicePropertySetResult{}, ErrInvalidDeviceInput
	}

	target, err := s.devices.FindDevicePropertySetTarget(ctx, DevicePropertySetTargetInput{
		UserID:   in.UserID,
		DeviceID: in.DeviceID,
	})
	if err != nil {
		return DevicePropertySetResult{}, err
	}
	if target.DeviceStatus != activeDeviceStatus {
		return DevicePropertySetResult{}, ErrInvalidDeviceInput
	}
	if target.ConnectionStatus != "online" {
		return DevicePropertySetResult{}, ErrDeviceOffline
	}

	acceptedProperties, err := validateDevicePropertySetInput(target.Properties, in.Properties)
	if err != nil {
		return DevicePropertySetResult{}, err
	}

	commandID, err := newDeviceCommandID()
	if err != nil {
		return DevicePropertySetResult{}, err
	}
	requestedAt := time.Now().UTC()
	topic := devicePropertySetDownTopic(target.TenantSlug, target.ProductKey, target.DeviceSlug)
	message := DevicePropertySetMessage{
		TenantID:    target.TenantID,
		ProductID:   target.ProductID,
		DeviceID:    target.DeviceID,
		TenantSlug:  target.TenantSlug,
		ProductKey:  target.ProductKey,
		DeviceSlug:  target.DeviceSlug,
		Protocol:    "mqtt",
		Topic:       topic,
		CommandID:   commandID,
		RequestedBy: in.UserID,
		Properties:  acceptedProperties,
		OccurredAt:  requestedAt,
	}
	if err := s.propertySetPublisher.PublishPropertySet(ctx, message); err != nil {
		return DevicePropertySetResult{}, fmt.Errorf("publish device property set requested event: %w", err)
	}

	return DevicePropertySetResult{
		CommandID:  commandID,
		Topic:      topic,
		Properties: acceptedProperties,
		SentAt:     requestedAt,
	}, nil
}

func (s *DeviceService) CallService(ctx context.Context, in DeviceServiceCallInput) (DeviceServiceCallResult, error) {
	if err := ctx.Err(); err != nil {
		return DeviceServiceCallResult{}, err
	}
	if s.serviceCallTargets == nil {
		return DeviceServiceCallResult{}, ErrDeviceCommandPublisherUnavailable
	}
	if s.serviceCallPublisher == nil {
		return DeviceServiceCallResult{}, ErrDeviceCommandPublisherUnavailable
	}
	in.UserID = strings.TrimSpace(in.UserID)
	in.DeviceID = strings.TrimSpace(in.DeviceID)
	in.ServiceName = strings.TrimSpace(in.ServiceName)
	if in.UserID == "" || in.DeviceID == "" || in.ServiceName == "" || !validThingsModelIdentifier(in.ServiceName) {
		return DeviceServiceCallResult{}, ErrInvalidDeviceInput
	}
	if in.Input == nil {
		in.Input = map[string]any{}
	}

	target, err := s.serviceCallTargets.FindDeviceServiceCallTarget(ctx, DeviceServiceCallTargetInput{
		UserID:   in.UserID,
		DeviceID: in.DeviceID,
	})
	if err != nil {
		return DeviceServiceCallResult{}, err
	}
	if target.DeviceStatus != activeDeviceStatus {
		return DeviceServiceCallResult{}, ErrInvalidDeviceInput
	}
	if target.ConnectionStatus != "online" {
		return DeviceServiceCallResult{}, ErrDeviceOffline
	}

	acceptedInput, err := validateDeviceServiceCallInput(in.ServiceName, target.Services, in.Input)
	if err != nil {
		return DeviceServiceCallResult{}, err
	}

	commandID, err := newDeviceCommandID()
	if err != nil {
		return DeviceServiceCallResult{}, err
	}
	requestedAt := time.Now().UTC()
	topic := deviceServiceDownTopic(target.TenantSlug, target.ProductKey, target.DeviceSlug, in.ServiceName)
	message := DeviceServiceCallMessage{
		TenantID:    target.TenantID,
		ProductID:   target.ProductID,
		DeviceID:    target.DeviceID,
		TenantSlug:  target.TenantSlug,
		ProductKey:  target.ProductKey,
		DeviceSlug:  target.DeviceSlug,
		Protocol:    "mqtt",
		Topic:       topic,
		CommandID:   commandID,
		ServiceName: in.ServiceName,
		RequestedBy: in.UserID,
		Input:       acceptedInput,
		OccurredAt:  requestedAt,
	}
	if err := s.serviceCallPublisher.PublishServiceCall(ctx, message); err != nil {
		return DeviceServiceCallResult{}, fmt.Errorf("publish device service call requested event: %w", err)
	}

	return DeviceServiceCallResult{
		CommandID:   commandID,
		Topic:       topic,
		ServiceName: in.ServiceName,
		Input:       acceptedInput,
		SentAt:      requestedAt,
	}, nil
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

func normalizeTrendProperties(properties []string) ([]string, bool) {
	seen := make(map[string]struct{}, len(properties))
	out := make([]string, 0, len(properties))
	for _, property := range properties {
		property = strings.TrimSpace(property)
		if property == "" {
			continue
		}
		if !validThingsModelIdentifier(property) {
			return nil, false
		}
		if _, exists := seen[property]; exists {
			continue
		}
		seen[property] = struct{}{}
		out = append(out, property)
	}
	if len(out) == 0 || len(out) > maxTrendProperties {
		return nil, false
	}
	return out, true
}

func validTrendAggregate(aggregate string) bool {
	switch aggregate {
	case "avg", "min", "max", "last":
		return true
	default:
		return false
	}
}

func deviceServiceDownTopic(tenantSlug string, productKey string, deviceSlug string, serviceName string) string {
	return "lf/v1/" + tenantSlug + "/" + productKey + "/" + deviceSlug + "/service/down/" + serviceName
}

func devicePropertySetDownTopic(tenantSlug string, productKey string, deviceSlug string) string {
	return "lf/v1/" + tenantSlug + "/" + productKey + "/" + deviceSlug + "/property/down/set"
}

func newDeviceCommandID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate command id: %w", err)
	}
	out := make([]byte, 36)
	hex.Encode(out[0:8], b[0:4])
	out[8] = '-'
	hex.Encode(out[9:13], b[4:6])
	out[13] = '-'
	hex.Encode(out[14:18], b[6:8])
	out[18] = '-'
	hex.Encode(out[19:23], b[8:10])
	out[23] = '-'
	hex.Encode(out[24:36], b[10:16])
	return string(out), nil
}

func validateDevicePropertySetInput(propertiesDef ThingsModelObject, properties map[string]any) (map[string]any, error) {
	if len(properties) == 0 {
		return nil, ErrInvalidDeviceInput
	}

	accepted := make(map[string]any, len(properties))
	for name, value := range properties {
		if !validThingsModelIdentifier(name) {
			return nil, ErrInvalidDeviceInput
		}
		defRaw, ok := propertiesDef[name]
		if !ok {
			return nil, ErrInvalidDeviceInput
		}
		def, ok := thingsModelDefinitionObject(defRaw)
		if !ok {
			return nil, ErrInvalidDeviceInput
		}
		accessMode := optionalString(def, "access_mode")
		if accessMode != propertyAccessWrite && accessMode != propertyAccessReadWrite {
			return nil, ErrInvalidDeviceInput
		}
		if err := validateDeviceServiceCallValue(name, def, value); err != nil {
			return nil, err
		}
		accepted[name] = value
	}
	return accepted, nil
}

func validateDeviceServiceCallInput(serviceName string, services ThingsModelObject, input map[string]any) (map[string]any, error) {
	rawService, ok := services[serviceName]
	if !ok {
		return nil, ErrDeviceServiceNotFound
	}
	serviceDef, ok := thingsModelDefinitionObject(rawService)
	if !ok {
		return nil, ErrInvalidDeviceInput
	}
	inputDef, ok := thingsModelDefinitionObject(serviceDef["input"])
	if !ok {
		return nil, ErrInvalidDeviceInput
	}

	accepted := make(map[string]any, len(input))
	for name, value := range input {
		defRaw, ok := inputDef[name]
		if !ok || !validThingsModelIdentifier(name) {
			return nil, ErrInvalidDeviceInput
		}
		def, ok := thingsModelDefinitionObject(defRaw)
		if !ok {
			return nil, ErrInvalidDeviceInput
		}
		if err := validateDeviceServiceCallValue(name, def, value); err != nil {
			return nil, err
		}
		accepted[name] = value
	}

	for name, defRaw := range inputDef {
		def, ok := thingsModelDefinitionObject(defRaw)
		if !ok {
			return nil, ErrInvalidDeviceInput
		}
		required, _ := def["required"].(bool)
		if required {
			if _, exists := accepted[name]; !exists {
				return nil, ErrInvalidDeviceInput
			}
		}
	}
	return accepted, nil
}

func validateDeviceServiceCallValue(name string, def map[string]any, value any) error {
	dataType := optionalString(def, "data_type")
	spec, _ := thingsModelDefinitionObject(def["spec"])
	switch dataType {
	case propertyDataTypeInt:
		n, ok := numberFromValue(value)
		if !ok || math.Trunc(n) != n {
			return ErrInvalidDeviceInput
		}
		return validateDeviceServiceCallNumber(name, dataType, spec, n)
	case propertyDataTypeFloat, propertyDataTypeDouble:
		n, ok := numberFromValue(value)
		if !ok {
			return ErrInvalidDeviceInput
		}
		return validateDeviceServiceCallNumber(name, dataType, spec, n)
	case propertyDataTypeBool:
		if _, ok := value.(bool); !ok {
			return ErrInvalidDeviceInput
		}
		return nil
	case propertyDataTypeString:
		if _, ok := value.(string); !ok {
			return ErrInvalidDeviceInput
		}
		return nil
	default:
		return ErrInvalidDeviceInput
	}
}

func validateDeviceServiceCallNumber(name string, dataType string, spec map[string]any, value float64) error {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return ErrInvalidDeviceInput
	}
	min, hasMin, err := optionalNumber(spec, "min")
	if err != nil {
		return ErrInvalidDeviceInput
	}
	max, hasMax, err := optionalNumber(spec, "max")
	if err != nil {
		return ErrInvalidDeviceInput
	}
	step, hasStep, err := optionalNumber(spec, "step")
	if err != nil {
		return ErrInvalidDeviceInput
	}
	precision, hasPrecision, err := optionalNumber(spec, "precision")
	if err != nil {
		return ErrInvalidDeviceInput
	}
	if hasMin && value < min {
		return ErrInvalidDeviceInput
	}
	if hasMax && value > max {
		return ErrInvalidDeviceInput
	}
	if hasPrecision {
		if precision < 0 || math.Trunc(precision) != precision {
			return ErrInvalidDeviceInput
		}
		scale := math.Pow10(int(precision))
		if math.Abs(value*scale-math.Round(value*scale)) > 1e-9*math.Max(1, math.Abs(value*scale)) {
			return ErrInvalidDeviceInput
		}
	}
	if hasStep {
		if step <= 0 {
			return ErrInvalidDeviceInput
		}
		anchor := 0.0
		if hasMin {
			anchor = min
		}
		diff := value - anchor
		quotient := diff / step
		if math.Abs(quotient-math.Round(quotient)) > 1e-9*math.Max(1, math.Abs(quotient)) {
			return ErrInvalidDeviceInput
		}
	}
	if dataType == propertyDataTypeInt && math.Trunc(value) != value {
		return ErrInvalidDeviceInput
	}
	_ = name
	return nil
}

func numberFromValue(value any) (float64, bool) {
	switch n := value.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	default:
		return 0, false
	}
}
