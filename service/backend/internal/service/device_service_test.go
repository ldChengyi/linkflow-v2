package service

import (
	"context"
	"errors"
	"testing"
)

type fakeDeviceStore struct {
	created  DeviceCreateInput
	listed   DeviceListInput
	got      DeviceGetInput
	latest   DeviceLatestPropertiesInput
	events   DeviceEventHistoryInput
	updated  DeviceUpdateInput
	deleted  DeviceDeleteInput
	device   Device
	authType string
	err      error
}

func (f *fakeDeviceStore) FindDeviceProductAuthType(ctx context.Context, in DeviceProductAuthInput) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	if f.authType == "" {
		return defaultProductAuthType, nil
	}
	return f.authType, nil
}

func (f *fakeDeviceStore) CreateDevice(ctx context.Context, in DeviceCreateInput) (Device, error) {
	f.created = in
	if f.err != nil {
		return Device{}, f.err
	}
	f.device = Device{
		ID:               "device-1",
		TenantID:         in.TenantID,
		ProductID:        in.ProductID,
		DeviceSlug:       in.DeviceSlug,
		DeviceName:       in.DeviceName,
		Description:      in.Description,
		Status:           defaultDeviceStatus,
		ConnectionStatus: defaultDeviceConnection,
		GatewayDeviceID:  in.GatewayDeviceID,
	}
	return f.device, nil
}

func (f *fakeDeviceStore) ListDevices(ctx context.Context, in DeviceListInput) (PageResult[Device], error) {
	f.listed = in
	if f.err != nil {
		return PageResult[Device]{}, f.err
	}
	return NewPageResult([]Device{f.device}, 1, in.PageInput), nil
}

func (f *fakeDeviceStore) FindDeviceByID(ctx context.Context, in DeviceGetInput) (Device, error) {
	f.got = in
	if f.err != nil {
		return Device{}, f.err
	}
	return f.device, nil
}

func (f *fakeDeviceStore) FindDeviceLatestProperties(ctx context.Context, in DeviceLatestPropertiesInput) (DeviceLatestProperties, error) {
	f.latest = in
	if f.err != nil {
		return DeviceLatestProperties{}, f.err
	}
	return DeviceLatestProperties{
		ID:         in.DeviceID,
		TenantID:   "tenant-1",
		ProductID:  "product-1",
		ProductKey: "esp32",
		DeviceSlug: "dev-1",
		Reported:   true,
		Properties: map[string]any{"temperature": 23.5},
	}, nil
}

func (f *fakeDeviceStore) ListDeviceEventHistory(ctx context.Context, in DeviceEventHistoryInput) (PageResult[DeviceEventEntry], error) {
	f.events = in
	if f.err != nil {
		return PageResult[DeviceEventEntry]{}, f.err
	}
	return NewPageResult([]DeviceEventEntry{{
		EventID:    "event-1",
		TenantID:   "tenant-1",
		ProductID:  "product-1",
		ProductKey: "esp32",
		DeviceSlug: "dev-1",
		EventName:  in.EventName,
		Params:     map[string]any{"code": "ok"},
	}}, 1, in.PageInput), nil
}

func (f *fakeDeviceStore) UpdateDevice(ctx context.Context, in DeviceUpdateInput) (Device, error) {
	f.updated = in
	if f.err != nil {
		return Device{}, f.err
	}
	f.device.DeviceName = in.DeviceName
	f.device.Description = in.Description
	f.device.Status = in.Status
	f.device.GatewayDeviceID = in.GatewayDeviceID
	return f.device, nil
}

func (f *fakeDeviceStore) DeleteDevice(ctx context.Context, in DeviceDeleteInput) error {
	f.deleted = in
	return f.err
}

type fakeDeviceSecretManager struct {
	secret string
	hash   string
	err    error
}

func (f fakeDeviceSecretManager) Generate() (string, error) {
	if f.err != nil {
		return "", f.err
	}
	if f.secret == "" {
		return "device-secret", nil
	}
	return f.secret, nil
}

func (f fakeDeviceSecretManager) Hash(secret string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	if f.hash == "" {
		return "hashed-" + secret, nil
	}
	return f.hash, nil
}

func (f fakeDeviceSecretManager) Compare(hash string, secret string) error {
	if f.err != nil {
		return f.err
	}
	if hash != "hashed-"+secret {
		return errors.New("invalid secret")
	}
	return nil
}

func (f *fakeDeviceStore) FindMQTTDeviceCredential(ctx context.Context, in MQTTAuthInput) (MQTTDeviceCredential, error) {
	if f.err != nil {
		return MQTTDeviceCredential{}, f.err
	}
	return MQTTDeviceCredential{
		TenantID:         "tenant-1",
		ProductID:        "product-1",
		DeviceID:         "device-1",
		TenantSlug:       in.TenantSlug,
		ProductKey:       in.ProductKey,
		DeviceSlug:       in.DeviceSlug,
		ProductAuthType:  defaultProductAuthType,
		TenantStatus:     activeProductStatus,
		ProductStatus:    activeProductStatus,
		DeviceStatus:     activeDeviceStatus,
		CredentialStatus: "active",
		SecretHash:       "hashed-device-secret",
	}, nil
}

func TestDeviceServiceCreateNormalizesInputAndDefaults(t *testing.T) {
	store := &fakeDeviceStore{}
	svc := newTestDeviceService(t, store)

	result, err := svc.Create(context.Background(), DeviceCreateInput{
		UserID:          "user-1",
		TenantID:        "tenant-1",
		ProductID:       "product-1",
		DeviceSlug:      " ESP32-DEV-001 ",
		DeviceName:      "  Sensor 001  ",
		Description:     "  lab sensor  ",
		GatewayDeviceID: " gateway-1 ",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	device := result.Device

	if store.created.DeviceSlug != "esp32-dev-001" {
		t.Fatalf("DeviceSlug = %q, want esp32-dev-001", store.created.DeviceSlug)
	}
	if store.created.DeviceName != "Sensor 001" {
		t.Fatalf("DeviceName = %q, want Sensor 001", store.created.DeviceName)
	}
	if store.created.Description != "lab sensor" {
		t.Fatalf("Description = %q, want lab sensor", store.created.Description)
	}
	if store.created.GatewayDeviceID != "gateway-1" {
		t.Fatalf("GatewayDeviceID = %q, want gateway-1", store.created.GatewayDeviceID)
	}
	if device.Status != defaultDeviceStatus || device.ConnectionStatus != defaultDeviceConnection {
		t.Fatalf("defaults = %q/%q, want active/offline", device.Status, device.ConnectionStatus)
	}
	if result.DeviceSecret != "device-secret" {
		t.Fatalf("DeviceSecret = %q, want device-secret", result.DeviceSecret)
	}
	if store.created.DeviceSecretHash != "hashed-device-secret" {
		t.Fatalf("DeviceSecretHash = %q, want hashed-device-secret", store.created.DeviceSecretHash)
	}
}

func TestDeviceServiceCreateAnonymousProductDoesNotCreateSecret(t *testing.T) {
	store := &fakeDeviceStore{authType: anonymousProductAuthType}
	svc := newTestDeviceService(t, store)

	result, err := svc.Create(context.Background(), DeviceCreateInput{
		UserID:     "user-1",
		TenantID:   "tenant-1",
		ProductID:  "product-1",
		DeviceSlug: "dev-1",
		DeviceName: "Device",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if result.DeviceSecret != "" {
		t.Fatalf("DeviceSecret = %q, want empty", result.DeviceSecret)
	}
	if store.created.DeviceSecretHash != "" {
		t.Fatalf("DeviceSecretHash = %q, want empty", store.created.DeviceSecretHash)
	}
}

func TestDeviceServiceCreateRejectsUnsupportedCertificateAuth(t *testing.T) {
	store := &fakeDeviceStore{authType: certificateProductAuthType}
	svc := newTestDeviceService(t, store)

	if _, err := svc.Create(context.Background(), DeviceCreateInput{
		UserID:     "user-1",
		TenantID:   "tenant-1",
		ProductID:  "product-1",
		DeviceSlug: "dev-1",
		DeviceName: "Device",
	}); !errors.Is(err, ErrInvalidDeviceInput) {
		t.Fatalf("Create() error = %v, want ErrInvalidDeviceInput", err)
	}
}

func TestDeviceServiceCreateValidatesRequiredInput(t *testing.T) {
	svc := newTestDeviceService(t, &fakeDeviceStore{})

	if _, err := svc.Create(context.Background(), DeviceCreateInput{
		TenantID:   "tenant-1",
		ProductID:  "product-1",
		DeviceSlug: "dev-1",
		DeviceName: "Device",
	}); !errors.Is(err, ErrInvalidDeviceInput) {
		t.Fatalf("Create() error = %v, want ErrInvalidDeviceInput", err)
	}
	if _, err := svc.Create(context.Background(), DeviceCreateInput{
		UserID:     "user-1",
		ProductID:  "product-1",
		DeviceSlug: "dev-1",
		DeviceName: "Device",
	}); !errors.Is(err, ErrInvalidDeviceInput) {
		t.Fatalf("Create() error = %v, want ErrInvalidDeviceInput", err)
	}
}

func TestDeviceServiceListNormalizesPagination(t *testing.T) {
	store := &fakeDeviceStore{}
	svc := newTestDeviceService(t, store)

	if _, err := svc.List(context.Background(), DeviceListInput{
		UserID:    "user-1",
		TenantID:  "tenant-1",
		ProductID: " product-1 ",
		PageInput: PageInput{Page: -1, PageSize: 1000},
	}); err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if store.listed.ProductID != "product-1" {
		t.Fatalf("ProductID = %q, want product-1", store.listed.ProductID)
	}
	if store.listed.Page != defaultPage || store.listed.PageSize != maxPageSize {
		t.Fatalf("PageInput = %+v, want page %d page_size %d", store.listed.PageInput, defaultPage, maxPageSize)
	}
}

func TestDeviceServiceUpdateDefaultsAndValidatesVariants(t *testing.T) {
	store := &fakeDeviceStore{}
	svc := newTestDeviceService(t, store)

	if _, err := svc.Update(context.Background(), DeviceUpdateInput{
		UserID:     "user-1",
		DeviceID:   "device-1",
		DeviceName: "Device",
	}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if store.updated.Status != defaultDeviceStatus {
		t.Fatalf("Status = %q, want active", store.updated.Status)
	}
	if _, err := svc.Update(context.Background(), DeviceUpdateInput{
		UserID:     "user-1",
		DeviceID:   "device-1",
		DeviceName: "Device",
		Status:     "deleted",
	}); !errors.Is(err, ErrInvalidDeviceInput) {
		t.Fatalf("Update() error = %v, want ErrInvalidDeviceInput", err)
	}
}

func TestDeviceServiceGetAndDeletePassInput(t *testing.T) {
	store := &fakeDeviceStore{device: Device{ID: "device-1"}}
	svc := newTestDeviceService(t, store)

	if _, err := svc.Get(context.Background(), DeviceGetInput{
		UserID:   "user-1",
		DeviceID: "device-1",
	}); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if store.got.UserID != "user-1" || store.got.DeviceID != "device-1" {
		t.Fatalf("Get input = %+v, want user-1/device-1", store.got)
	}

	if err := svc.Delete(context.Background(), DeviceDeleteInput{
		UserID:   "user-1",
		DeviceID: "device-1",
	}); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if store.deleted.UserID != "user-1" || store.deleted.DeviceID != "device-1" {
		t.Fatalf("Delete input = %+v, want user-1/device-1", store.deleted)
	}
}

func TestDeviceServiceLatestPropertiesPassesNormalizedInput(t *testing.T) {
	store := &fakeDeviceStore{}
	svc := newTestDeviceService(t, store)

	latest, err := svc.LatestProperties(context.Background(), DeviceLatestPropertiesInput{
		UserID:   " user-1 ",
		DeviceID: " device-1 ",
	})
	if err != nil {
		t.Fatalf("LatestProperties() error = %v", err)
	}
	if store.latest.UserID != "user-1" || store.latest.DeviceID != "device-1" {
		t.Fatalf("LatestProperties input = %+v, want user-1/device-1", store.latest)
	}
	if !latest.Reported || latest.Properties["temperature"] != 23.5 {
		t.Fatalf("LatestProperties result = %+v, want reported temperature", latest)
	}
}

func TestDeviceServiceEventHistoryNormalizesInputAndPagination(t *testing.T) {
	store := &fakeDeviceStore{}
	svc := newTestDeviceService(t, store)

	result, err := svc.EventHistory(context.Background(), DeviceEventHistoryInput{
		UserID:    " user-1 ",
		DeviceID:  " device-1 ",
		EventName: " alarm ",
		PageInput: PageInput{Page: -1, PageSize: 1000},
	})
	if err != nil {
		t.Fatalf("EventHistory() error = %v", err)
	}
	if store.events.UserID != "user-1" || store.events.DeviceID != "device-1" || store.events.EventName != "alarm" {
		t.Fatalf("EventHistory input = %+v, want normalized user/device/event", store.events)
	}
	if store.events.Page != defaultPage || store.events.PageSize != maxPageSize {
		t.Fatalf("PageInput = %+v, want page %d page_size %d", store.events.PageInput, defaultPage, maxPageSize)
	}
	if result.Total != 1 || len(result.Items) != 1 || result.Items[0].Params["code"] != "ok" {
		t.Fatalf("EventHistory result = %+v, want one event", result)
	}
}

func TestMQTTAuthServiceAuthenticatesSecretDevice(t *testing.T) {
	store := &fakeDeviceStore{}
	svc, err := NewMQTTAuthService(store, fakeDeviceSecretManager{})
	if err != nil {
		t.Fatalf("NewMQTTAuthService() error = %v", err)
	}

	result, err := svc.Authenticate(context.Background(), MQTTAuthInput{
		TenantSlug: " Default ",
		ProductKey: " ESP32 ",
		DeviceSlug: " DEV-001 ",
		Password:   "device-secret",
	})
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}

	if result.TenantSlug != "default" || result.ProductKey != "esp32" || result.DeviceSlug != "dev-001" {
		t.Fatalf("Authenticate result = %+v, want normalized slugs", result)
	}
}

func TestMQTTAuthServiceRejectsInvalidSecret(t *testing.T) {
	store := &fakeDeviceStore{}
	svc, err := NewMQTTAuthService(store, fakeDeviceSecretManager{})
	if err != nil {
		t.Fatalf("NewMQTTAuthService() error = %v", err)
	}

	_, err = svc.Authenticate(context.Background(), MQTTAuthInput{
		TenantSlug: "default",
		ProductKey: "esp32",
		DeviceSlug: "dev-001",
		Password:   "wrong",
	})
	if !errors.Is(err, ErrInvalidMQTTAuth) {
		t.Fatalf("Authenticate() error = %v, want ErrInvalidMQTTAuth", err)
	}
}

func newTestDeviceService(t *testing.T, store DeviceStore) *DeviceService {
	t.Helper()

	svc, err := NewDeviceService(store, fakeDeviceSecretManager{})
	if err != nil {
		t.Fatalf("NewDeviceService() error = %v", err)
	}
	return svc
}
