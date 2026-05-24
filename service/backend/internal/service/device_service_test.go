package service

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeDeviceStore struct {
	created        DeviceCreateInput
	listed         DeviceListInput
	got            DeviceGetInput
	latest         DeviceLatestPropertiesInput
	trend          DevicePropertyTrendInput
	events         DeviceEventHistoryInput
	calls          DeviceServiceCallHistoryInput
	propertySets   DevicePropertySetHistoryInput
	updated        DeviceUpdateInput
	deleted        DeviceDeleteInput
	device         Device
	target         DeviceServiceCallTarget
	propertyTarget DevicePropertySetTarget
	authType       string
	err            error
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

func (f *fakeDeviceStore) FindDevicePropertyTrend(ctx context.Context, in DevicePropertyTrendInput) (DevicePropertyTrend, error) {
	f.trend = in
	if f.err != nil {
		return DevicePropertyTrend{}, f.err
	}
	return DevicePropertyTrend{
		DeviceID:      in.DeviceID,
		TenantID:      "tenant-1",
		ProductID:     "product-1",
		ProductKey:    "esp32",
		DeviceSlug:    "dev-1",
		Properties:    in.Properties,
		From:          in.From,
		To:            in.To,
		BucketSeconds: in.BucketSeconds,
		Aggregate:     in.Aggregate,
		Series: []DevicePropertyTrendSeries{{
			Property: in.Properties[0],
			Points: []DevicePropertyTrendPoint{{
				BucketAt: in.From,
				Value:    23.5,
				Min:      23,
				Max:      24,
				Count:    2,
			}},
		}},
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

func (f *fakeDeviceStore) ListDeviceServiceCallHistory(ctx context.Context, in DeviceServiceCallHistoryInput) (PageResult[DeviceServiceCallHistoryEntry], error) {
	f.calls = in
	if f.err != nil {
		return PageResult[DeviceServiceCallHistoryEntry]{}, f.err
	}
	return NewPageResult([]DeviceServiceCallHistoryEntry{{
		CommandID:          "018f56d3-7cb7-7f1a-9b41-3f3a63fd3db9",
		ServiceName:        "reboot",
		AckStatus:          "pending",
		AckDeadlineSeconds: in.AckDeadlineSeconds,
	}}, 1, in.PageInput), nil
}

func (f *fakeDeviceStore) ListDevicePropertySetHistory(ctx context.Context, in DevicePropertySetHistoryInput) (PageResult[DevicePropertySetHistoryEntry], error) {
	f.propertySets = in
	if f.err != nil {
		return PageResult[DevicePropertySetHistoryEntry]{}, f.err
	}
	return NewPageResult([]DevicePropertySetHistoryEntry{{
		CommandID:          "018f56d3-7cb7-7f1a-9b41-3f3a63fd3db9",
		Properties:         map[string]any{"led1": true},
		AckStatus:          "pending",
		AckDeadlineSeconds: in.AckDeadlineSeconds,
	}}, 1, in.PageInput), nil
}

func (f *fakeDeviceStore) FindDeviceServiceCallTarget(ctx context.Context, in DeviceServiceCallTargetInput) (DeviceServiceCallTarget, error) {
	if f.err != nil {
		return DeviceServiceCallTarget{}, f.err
	}
	if f.target.DeviceID != "" {
		return f.target, nil
	}
	return DeviceServiceCallTarget{
		TenantID:         "tenant-1",
		ProductID:        "product-1",
		DeviceID:         in.DeviceID,
		TenantSlug:       "default",
		ProductKey:       "esp32",
		DeviceSlug:       "dev-1",
		DeviceStatus:     activeDeviceStatus,
		ConnectionStatus: "online",
		Services: ThingsModelObject{
			"reboot": map[string]any{
				"name":      "Reboot",
				"call_type": "async",
				"input": map[string]any{
					"delay": map[string]any{
						"name":      "Delay",
						"data_type": "int",
						"required":  true,
						"spec":      map[string]any{"min": 0, "max": 60},
					},
				},
				"output": map[string]any{},
			},
		},
	}, nil
}

func (f *fakeDeviceStore) FindDevicePropertySetTarget(ctx context.Context, in DevicePropertySetTargetInput) (DevicePropertySetTarget, error) {
	if f.err != nil {
		return DevicePropertySetTarget{}, f.err
	}
	if f.propertyTarget.DeviceID != "" {
		return f.propertyTarget, nil
	}
	return DevicePropertySetTarget{
		TenantID:         "tenant-1",
		ProductID:        "product-1",
		DeviceID:         in.DeviceID,
		TenantSlug:       "default",
		ProductKey:       "esp32",
		DeviceSlug:       "dev-1",
		DeviceStatus:     activeDeviceStatus,
		ConnectionStatus: "online",
		Properties: ThingsModelObject{
			"led1": map[string]any{
				"name":        "LED 1",
				"data_type":   "bool",
				"required":    false,
				"access_mode": "readwrite",
			},
			"temperature": map[string]any{
				"name":        "Temperature",
				"data_type":   "float",
				"required":    false,
				"access_mode": "read",
			},
		},
	}, nil
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

type fakeDeviceServiceCallPublisher struct {
	message DeviceServiceCallMessage
	err     error
}

func (f *fakeDeviceServiceCallPublisher) PublishServiceCall(ctx context.Context, in DeviceServiceCallMessage) error {
	f.message = in
	return f.err
}

type fakeDevicePropertySetPublisher struct {
	message DevicePropertySetMessage
	err     error
}

func (f *fakeDevicePropertySetPublisher) PublishPropertySet(ctx context.Context, in DevicePropertySetMessage) error {
	f.message = in
	return f.err
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

func TestDeviceServiceCallServicePublishesRequestedEventWithValidatedInput(t *testing.T) {
	store := &fakeDeviceStore{}
	publisher := &fakeDeviceServiceCallPublisher{}
	svc := newTestDeviceService(t, store, WithDeviceServiceCallTargets(store), WithDeviceServiceCallPublisher(publisher))

	result, err := svc.CallService(context.Background(), DeviceServiceCallInput{
		UserID:      "user-1",
		DeviceID:    "device-1",
		ServiceName: "reboot",
		Input: map[string]any{
			"delay": float64(5),
		},
	})
	if err != nil {
		t.Fatalf("CallService() error = %v", err)
	}
	if result.CommandID == "" {
		t.Fatal("CommandID should be generated")
	}
	if result.Topic != "lf/v1/default/esp32/dev-1/service/down/reboot" {
		t.Fatalf("Topic = %q, want service down topic", result.Topic)
	}
	if publisher.message.CommandID != result.CommandID {
		t.Fatalf("published command_id = %q, want %q", publisher.message.CommandID, result.CommandID)
	}
	if publisher.message.TenantID != "tenant-1" || publisher.message.DeviceID != "device-1" {
		t.Fatalf("published target = %+v, want resolved target ids", publisher.message)
	}
	if publisher.message.RequestedBy != "user-1" {
		t.Fatalf("published requested_by = %q, want user-1", publisher.message.RequestedBy)
	}
	if publisher.message.Input["delay"] != float64(5) {
		t.Fatalf("published input = %v, want delay", publisher.message.Input)
	}
}

func TestDeviceServiceSetPropertiesPublishesRequestedEventWithWritableProperties(t *testing.T) {
	store := &fakeDeviceStore{}
	publisher := &fakeDevicePropertySetPublisher{}
	svc := newTestDeviceService(t, store, WithDevicePropertySetPublisher(publisher))

	result, err := svc.SetProperties(context.Background(), DevicePropertySetInput{
		UserID:   "user-1",
		DeviceID: "device-1",
		Properties: map[string]any{
			"led1": true,
		},
	})
	if err != nil {
		t.Fatalf("SetProperties() error = %v", err)
	}
	if result.CommandID == "" {
		t.Fatal("CommandID should be generated")
	}
	if result.Topic != "lf/v1/default/esp32/dev-1/property/down/set" {
		t.Fatalf("Topic = %q, want property down set topic", result.Topic)
	}
	if publisher.message.CommandID != result.CommandID {
		t.Fatalf("published command_id = %q, want %q", publisher.message.CommandID, result.CommandID)
	}
	if publisher.message.Properties["led1"] != true {
		t.Fatalf("published properties = %v, want led1", publisher.message.Properties)
	}
}

func TestDeviceServiceSetPropertiesRejectsReadOnlyProperty(t *testing.T) {
	store := &fakeDeviceStore{}
	publisher := &fakeDevicePropertySetPublisher{}
	svc := newTestDeviceService(t, store, WithDevicePropertySetPublisher(publisher))

	_, err := svc.SetProperties(context.Background(), DevicePropertySetInput{
		UserID:   "user-1",
		DeviceID: "device-1",
		Properties: map[string]any{
			"temperature": float64(23.5),
		},
	})
	if !errors.Is(err, ErrInvalidDeviceInput) {
		t.Fatalf("err = %v, want ErrInvalidDeviceInput", err)
	}
	if publisher.message.Topic != "" {
		t.Fatal("publisher should not be called")
	}
}

func TestDeviceServiceCallServiceRejectsUnknownService(t *testing.T) {
	store := &fakeDeviceStore{}
	publisher := &fakeDeviceServiceCallPublisher{}
	svc := newTestDeviceService(t, store, WithDeviceServiceCallTargets(store), WithDeviceServiceCallPublisher(publisher))

	_, err := svc.CallService(context.Background(), DeviceServiceCallInput{
		UserID:      "user-1",
		DeviceID:    "device-1",
		ServiceName: "unknown",
		Input:       map[string]any{},
	})
	if !errors.Is(err, ErrDeviceServiceNotFound) {
		t.Fatalf("err = %v, want ErrDeviceServiceNotFound", err)
	}
	if publisher.message.Topic != "" {
		t.Fatal("publisher should not be called")
	}
}

func TestDeviceServiceCallServiceRejectsMissingRequiredInput(t *testing.T) {
	store := &fakeDeviceStore{}
	publisher := &fakeDeviceServiceCallPublisher{}
	svc := newTestDeviceService(t, store, WithDeviceServiceCallTargets(store), WithDeviceServiceCallPublisher(publisher))

	_, err := svc.CallService(context.Background(), DeviceServiceCallInput{
		UserID:      "user-1",
		DeviceID:    "device-1",
		ServiceName: "reboot",
		Input:       map[string]any{},
	})
	if !errors.Is(err, ErrInvalidDeviceInput) {
		t.Fatalf("err = %v, want ErrInvalidDeviceInput", err)
	}
	if publisher.message.Topic != "" {
		t.Fatal("publisher should not be called")
	}
}

func TestDeviceServiceCallServiceRejectsOfflineDevice(t *testing.T) {
	store := &fakeDeviceStore{
		target: DeviceServiceCallTarget{
			TenantSlug:       "default",
			ProductKey:       "esp32",
			DeviceSlug:       "dev-1",
			DeviceStatus:     activeDeviceStatus,
			ConnectionStatus: "offline",
			Services:         ThingsModelObject{},
			DeviceID:         "device-1",
		},
	}
	publisher := &fakeDeviceServiceCallPublisher{}
	svc := newTestDeviceService(t, store, WithDeviceServiceCallTargets(store), WithDeviceServiceCallPublisher(publisher))

	_, err := svc.CallService(context.Background(), DeviceServiceCallInput{
		UserID:      "user-1",
		DeviceID:    "device-1",
		ServiceName: "reboot",
		Input:       map[string]any{},
	})
	if !errors.Is(err, ErrDeviceOffline) {
		t.Fatalf("err = %v, want ErrDeviceOffline", err)
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

func TestDeviceServicePropertyTrendNormalizesInput(t *testing.T) {
	store := &fakeDeviceStore{}
	svc := newTestDeviceService(t, store)
	from := time.Date(2026, 5, 22, 0, 0, 0, 0, time.UTC)
	to := from.Add(2 * time.Hour)

	result, err := svc.PropertyTrend(context.Background(), DevicePropertyTrendInput{
		UserID:        " user-1 ",
		DeviceID:      " device-1 ",
		Properties:    []string{" temperature ", "humidity", "temperature"},
		From:          from,
		To:            to,
		BucketSeconds: 0,
		Aggregate:     "",
	})
	if err != nil {
		t.Fatalf("PropertyTrend() error = %v", err)
	}
	if store.trend.UserID != "user-1" || store.trend.DeviceID != "device-1" {
		t.Fatalf("PropertyTrend input = %+v, want normalized user/device", store.trend)
	}
	if got := store.trend.Properties; len(got) != 2 || got[0] != "temperature" || got[1] != "humidity" {
		t.Fatalf("Properties = %v, want deduplicated temperature/humidity", got)
	}
	if store.trend.BucketSeconds != defaultTrendBucketSeconds || store.trend.Aggregate != "avg" {
		t.Fatalf("trend defaults = %d/%q, want %d/avg", store.trend.BucketSeconds, store.trend.Aggregate, defaultTrendBucketSeconds)
	}
	if result.Aggregate != "avg" || len(result.Series) != 1 || result.Series[0].Property != "temperature" {
		t.Fatalf("PropertyTrend result = %+v, want temperature series", result)
	}
}

func TestDeviceServicePropertyTrendRejectsInvalidInput(t *testing.T) {
	svc := newTestDeviceService(t, &fakeDeviceStore{})
	from := time.Date(2026, 5, 22, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		in   DevicePropertyTrendInput
	}{
		{
			name: "missing property",
			in: DevicePropertyTrendInput{
				UserID:   "user-1",
				DeviceID: "device-1",
				From:     from,
				To:       from.Add(time.Hour),
			},
		},
		{
			name: "invalid property",
			in: DevicePropertyTrendInput{
				UserID:     "user-1",
				DeviceID:   "device-1",
				Properties: []string{"bad property"},
				From:       from,
				To:         from.Add(time.Hour),
			},
		},
		{
			name: "invalid time range",
			in: DevicePropertyTrendInput{
				UserID:     "user-1",
				DeviceID:   "device-1",
				Properties: []string{"temperature"},
				From:       from,
				To:         from,
			},
		},
		{
			name: "range too large",
			in: DevicePropertyTrendInput{
				UserID:     "user-1",
				DeviceID:   "device-1",
				Properties: []string{"temperature"},
				From:       from,
				To:         from.Add(maxTrendRange + time.Second),
			},
		},
		{
			name: "invalid bucket",
			in: DevicePropertyTrendInput{
				UserID:        "user-1",
				DeviceID:      "device-1",
				Properties:    []string{"temperature"},
				From:          from,
				To:            from.Add(time.Hour),
				BucketSeconds: 1,
			},
		},
		{
			name: "invalid aggregate",
			in: DevicePropertyTrendInput{
				UserID:     "user-1",
				DeviceID:   "device-1",
				Properties: []string{"temperature"},
				From:       from,
				To:         from.Add(time.Hour),
				Aggregate:  "sum",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := svc.PropertyTrend(context.Background(), tc.in); !errors.Is(err, ErrInvalidDeviceInput) {
				t.Fatalf("PropertyTrend() error = %v, want ErrInvalidDeviceInput", err)
			}
		})
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

func TestDeviceServiceCallHistoryNormalizesInputDeadlineAndPagination(t *testing.T) {
	store := &fakeDeviceStore{}
	svc := newTestDeviceService(t, store)

	result, err := svc.ServiceCallHistory(context.Background(), DeviceServiceCallHistoryInput{
		UserID:             " user-1 ",
		DeviceID:           " device-1 ",
		ServiceName:        " reboot ",
		AckDeadlineSeconds: 0,
		PageInput:          PageInput{Page: -1, PageSize: 1000},
	})
	if err != nil {
		t.Fatalf("ServiceCallHistory() error = %v", err)
	}
	if store.calls.UserID != "user-1" || store.calls.DeviceID != "device-1" || store.calls.ServiceName != "reboot" {
		t.Fatalf("ServiceCallHistory input = %+v, want normalized user/device/service", store.calls)
	}
	if store.calls.AckDeadlineSeconds != 90 {
		t.Fatalf("AckDeadlineSeconds = %d, want default 90", store.calls.AckDeadlineSeconds)
	}
	if store.calls.Page != defaultPage || store.calls.PageSize != maxPageSize {
		t.Fatalf("PageInput = %+v, want page %d page_size %d", store.calls.PageInput, defaultPage, maxPageSize)
	}
	if result.Total != 1 || len(result.Items) != 1 || result.Items[0].ServiceName != "reboot" {
		t.Fatalf("ServiceCallHistory result = %+v, want one call", result)
	}
}

func TestDeviceServicePropertySetHistoryNormalizesInputDeadlineAndPagination(t *testing.T) {
	store := &fakeDeviceStore{}
	svc := newTestDeviceService(t, store)

	result, err := svc.PropertySetHistory(context.Background(), DevicePropertySetHistoryInput{
		UserID:             " user-1 ",
		DeviceID:           " device-1 ",
		PropertyName:       " led1 ",
		AckDeadlineSeconds: 0,
		PageInput:          PageInput{Page: -1, PageSize: 1000},
	})
	if err != nil {
		t.Fatalf("PropertySetHistory() error = %v", err)
	}
	if store.propertySets.UserID != "user-1" || store.propertySets.DeviceID != "device-1" || store.propertySets.PropertyName != "led1" {
		t.Fatalf("PropertySetHistory input = %+v, want normalized user/device/property", store.propertySets)
	}
	if store.propertySets.AckDeadlineSeconds != 90 {
		t.Fatalf("AckDeadlineSeconds = %d, want default 90", store.propertySets.AckDeadlineSeconds)
	}
	if store.propertySets.Page != defaultPage || store.propertySets.PageSize != maxPageSize {
		t.Fatalf("PageInput = %+v, want page %d page_size %d", store.propertySets.PageInput, defaultPage, maxPageSize)
	}
	if result.Total != 1 || len(result.Items) != 1 || result.Items[0].Properties["led1"] != true {
		t.Fatalf("PropertySetHistory result = %+v, want one property set", result)
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

func newTestDeviceService(t *testing.T, store DeviceStore, options ...DeviceServiceOption) *DeviceService {
	t.Helper()

	svc, err := NewDeviceService(store, fakeDeviceSecretManager{}, options...)
	if err != nil {
		t.Fatalf("NewDeviceService() error = %v", err)
	}
	return svc
}
