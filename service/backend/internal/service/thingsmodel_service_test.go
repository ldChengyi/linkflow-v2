package service

import (
	"context"
	"errors"
	"testing"
)

type fakeThingsModelStore struct {
	created ThingsModelCreateInput
	listed  ThingsModelListInput
	got     ThingsModelGetInput
	updated ThingsModelUpdateInput
	deleted ThingsModelDeleteInput
	model   ThingsModel
	err     error
}

func (f *fakeThingsModelStore) CreateThingsModel(ctx context.Context, in ThingsModelCreateInput) (ThingsModel, error) {
	f.created = in
	if f.err != nil {
		return ThingsModel{}, f.err
	}
	f.model = ThingsModel{
		ID:           "thingsmodel-1",
		TenantID:     in.TenantID,
		ProductID:    in.ProductID,
		ModelVersion: in.ModelVersion,
		ModelName:    in.ModelName,
		Description:  in.Description,
		Status:       in.Status,
		IsCurrent:    in.IsCurrent,
		Properties:   in.Properties,
		Events:       in.Events,
		Services:     in.Services,
	}
	return f.model, nil
}

func (f *fakeThingsModelStore) ListThingsModels(ctx context.Context, in ThingsModelListInput) (PageResult[ThingsModel], error) {
	f.listed = in
	if f.err != nil {
		return PageResult[ThingsModel]{}, f.err
	}
	return NewPageResult([]ThingsModel{f.model}, 1, in.PageInput), nil
}

func (f *fakeThingsModelStore) FindThingsModelByID(ctx context.Context, in ThingsModelGetInput) (ThingsModel, error) {
	f.got = in
	if f.err != nil {
		return ThingsModel{}, f.err
	}
	return f.model, nil
}

func (f *fakeThingsModelStore) UpdateThingsModel(ctx context.Context, in ThingsModelUpdateInput) (ThingsModel, error) {
	f.updated = in
	if f.err != nil {
		return ThingsModel{}, f.err
	}
	f.model.ModelName = in.ModelName
	f.model.Description = in.Description
	f.model.Status = in.Status
	f.model.IsCurrent = in.IsCurrent
	f.model.Properties = in.Properties
	f.model.Events = in.Events
	f.model.Services = in.Services
	return f.model, nil
}

func (f *fakeThingsModelStore) DeleteThingsModel(ctx context.Context, in ThingsModelDeleteInput) error {
	f.deleted = in
	return f.err
}

func TestThingsModelServiceCreateNormalizesInputAndDefaults(t *testing.T) {
	store := &fakeThingsModelStore{}
	svc := newTestThingsModelService(t, store)

	model, err := svc.Create(context.Background(), ThingsModelCreateInput{
		UserID:       " user-1 ",
		TenantID:     " tenant-1 ",
		ProductID:    " product-1 ",
		ModelVersion: 1,
		ModelName:    "  ESP32 Thermostat  ",
		Description:  "  temperature model  ",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if store.created.UserID != "user-1" || store.created.TenantID != "tenant-1" || store.created.ProductID != "product-1" {
		t.Fatalf("scope = %q/%q/%q, want normalized ids", store.created.UserID, store.created.TenantID, store.created.ProductID)
	}
	if store.created.ModelName != "ESP32 Thermostat" || store.created.Description != "temperature model" {
		t.Fatalf("text = %q/%q, want trimmed values", store.created.ModelName, store.created.Description)
	}
	if store.created.Status != thingsModelStatusDraft {
		t.Fatalf("Status = %q, want draft", store.created.Status)
	}
	if store.created.Properties == nil || store.created.Events == nil || store.created.Services == nil {
		t.Fatalf("json objects must default to empty maps")
	}
	if model.ID != "thingsmodel-1" {
		t.Fatalf("model.ID = %q, want thingsmodel-1", model.ID)
	}
}

func TestThingsModelServiceCreateValidatesCurrentRequiresPublished(t *testing.T) {
	svc := newTestThingsModelService(t, &fakeThingsModelStore{})

	if _, err := svc.Create(context.Background(), ThingsModelCreateInput{
		UserID:       "user-1",
		TenantID:     "tenant-1",
		ProductID:    "product-1",
		ModelVersion: 1,
		ModelName:    "ESP32",
		Status:       "draft",
		IsCurrent:    true,
	}); !errors.Is(err, ErrInvalidThingsModelInput) {
		t.Fatalf("Create() error = %v, want ErrInvalidThingsModelInput", err)
	}
}

func TestThingsModelServiceCreateValidatesRequiredInput(t *testing.T) {
	svc := newTestThingsModelService(t, &fakeThingsModelStore{})

	if _, err := svc.Create(context.Background(), ThingsModelCreateInput{
		TenantID:     "tenant-1",
		ProductID:    "product-1",
		ModelVersion: 1,
		ModelName:    "ESP32",
	}); !errors.Is(err, ErrInvalidThingsModelInput) {
		t.Fatalf("Create() error = %v, want ErrInvalidThingsModelInput", err)
	}
	if _, err := svc.Create(context.Background(), ThingsModelCreateInput{
		UserID:    "user-1",
		TenantID:  "tenant-1",
		ProductID: "product-1",
		ModelName: "ESP32",
	}); !errors.Is(err, ErrInvalidThingsModelInput) {
		t.Fatalf("Create() error = %v, want ErrInvalidThingsModelInput", err)
	}
}

func TestThingsModelServiceCreateValidatesThingModelDefinition(t *testing.T) {
	svc := newTestThingsModelService(t, &fakeThingsModelStore{})

	validProperties := ThingsModelObject{
		"temperature": map[string]any{
			"name":        "温度",
			"data_type":   "float",
			"access_mode": "read",
			"required":    true,
			"spec": map[string]any{
				"min":       -40,
				"max":       125,
				"step":      0.1,
				"unit":      "celsius",
				"precision": 1,
			},
		},
	}
	validEvents := ThingsModelObject{
		"temperature_alarm": map[string]any{
			"name":  "温度报警",
			"level": "warning",
			"desc":  "设备检测到温度超过安全阈值",
			"output": map[string]any{
				"temperature": map[string]any{
					"name":      "当前温度",
					"data_type": "float",
					"required":  true,
				},
			},
		},
	}
	validServices := ThingsModelObject{
		"reboot": map[string]any{
			"name":      "重启设备",
			"call_type": "async",
			"desc":      "平台下发重启命令",
			"input":     map[string]any{},
			"output": map[string]any{
				"accepted": map[string]any{
					"name":      "是否接受",
					"data_type": "bool",
					"required":  true,
				},
			},
		},
	}

	if _, err := svc.Create(context.Background(), ThingsModelCreateInput{
		UserID:       "user-1",
		TenantID:     "tenant-1",
		ProductID:    "product-1",
		ModelVersion: 1,
		ModelName:    "ESP32",
		Properties:   validProperties,
		Events:       validEvents,
		Services:     validServices,
	}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	cases := []struct {
		name       string
		properties ThingsModelObject
		events     ThingsModelObject
		services   ThingsModelObject
	}{
		{
			name: "invalid property data type",
			properties: ThingsModelObject{
				"temperature": map[string]any{
					"name":        "温度",
					"data_type":   "number",
					"access_mode": "read",
					"required":    true,
				},
			},
		},
		{
			name: "invalid property spec range",
			properties: ThingsModelObject{
				"temperature": map[string]any{
					"name":        "温度",
					"data_type":   "float",
					"access_mode": "read",
					"required":    true,
					"spec":        map[string]any{"min": 100, "max": 1},
				},
			},
		},
		{
			name: "event output must be object",
			events: ThingsModelObject{
				"temperature_alarm": map[string]any{
					"name":   "温度报警",
					"level":  "warning",
					"output": "temperature",
				},
			},
		},
		{
			name: "event param missing required",
			events: ThingsModelObject{
				"temperature_alarm": map[string]any{
					"name":  "温度报警",
					"level": "warning",
					"output": map[string]any{
						"temperature": map[string]any{
							"name":      "当前温度",
							"data_type": "float",
						},
					},
				},
			},
		},
		{
			name: "service call type invalid",
			services: ThingsModelObject{
				"reboot": map[string]any{
					"name":      "重启设备",
					"call_type": "later",
					"input":     map[string]any{},
					"output":    map[string]any{},
				},
			},
		},
		{
			name: "service input param invalid identifier",
			services: ThingsModelObject{
				"set_report_interval": map[string]any{
					"name":      "设置上报周期",
					"call_type": "sync",
					"input": map[string]any{
						"interval-seconds": map[string]any{
							"name":      "上报间隔",
							"data_type": "int",
							"required":  true,
						},
					},
					"output": map[string]any{},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := svc.Create(context.Background(), ThingsModelCreateInput{
				UserID:       "user-1",
				TenantID:     "tenant-1",
				ProductID:    "product-1",
				ModelVersion: 1,
				ModelName:    "ESP32",
				Properties:   tc.properties,
				Events:       tc.events,
				Services:     tc.services,
			}); !errors.Is(err, ErrInvalidThingsModelInput) {
				t.Fatalf("Create() error = %v, want ErrInvalidThingsModelInput", err)
			}
		})
	}
}

func TestThingsModelServiceListNormalizesPagination(t *testing.T) {
	store := &fakeThingsModelStore{}
	svc := newTestThingsModelService(t, store)

	if _, err := svc.List(context.Background(), ThingsModelListInput{
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

func TestThingsModelServiceUpdateDefaultsAndValidatesStatus(t *testing.T) {
	store := &fakeThingsModelStore{}
	svc := newTestThingsModelService(t, store)

	if _, err := svc.Update(context.Background(), ThingsModelUpdateInput{
		UserID:        "user-1",
		ThingsModelID: "thingsmodel-1",
		ModelName:     "ESP32",
	}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if store.updated.Status != thingsModelStatusDraft {
		t.Fatalf("Status = %q, want draft", store.updated.Status)
	}
	if store.updated.Properties == nil || store.updated.Events == nil || store.updated.Services == nil {
		t.Fatalf("json objects must default to empty maps")
	}

	if _, err := svc.Update(context.Background(), ThingsModelUpdateInput{
		UserID:        "user-1",
		ThingsModelID: "thingsmodel-1",
		ModelName:     "ESP32",
		Status:        "unknown",
	}); !errors.Is(err, ErrInvalidThingsModelInput) {
		t.Fatalf("Update() error = %v, want ErrInvalidThingsModelInput", err)
	}
}

func TestThingsModelServiceGetAndDeletePassInput(t *testing.T) {
	store := &fakeThingsModelStore{model: ThingsModel{ID: "thingsmodel-1"}}
	svc := newTestThingsModelService(t, store)

	if _, err := svc.Get(context.Background(), ThingsModelGetInput{
		UserID:        "user-1",
		ThingsModelID: "thingsmodel-1",
	}); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if store.got.UserID != "user-1" || store.got.ThingsModelID != "thingsmodel-1" {
		t.Fatalf("Get input = %+v, want user-1/thingsmodel-1", store.got)
	}

	if err := svc.Delete(context.Background(), ThingsModelDeleteInput{
		UserID:        "user-1",
		ThingsModelID: "thingsmodel-1",
	}); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if store.deleted.UserID != "user-1" || store.deleted.ThingsModelID != "thingsmodel-1" {
		t.Fatalf("Delete input = %+v, want user-1/thingsmodel-1", store.deleted)
	}
}

func newTestThingsModelService(t *testing.T, store ThingsModelStore) *ThingsModelService {
	t.Helper()

	svc, err := NewThingsModelService(store)
	if err != nil {
		t.Fatalf("NewThingsModelService() error = %v", err)
	}
	return svc
}
