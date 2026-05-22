package validator

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeReader struct {
	def   ThingsModelDefinition
	err   error
	calls int
}

func (f *fakeReader) FindCurrentByProductID(ctx context.Context, tenantID string, productID string) (ThingsModelDefinition, error) {
	f.calls++
	if f.err != nil {
		return ThingsModelDefinition{}, f.err
	}
	def := f.def
	def.TenantID = tenantID
	def.ProductID = productID
	return def, nil
}

func newTestValidator(t *testing.T, reader ThingsModelReader) *PropertyReportValidator {
	t.Helper()
	v, err := NewPropertyReportValidator(reader, DefaultCacheTTL)
	if err != nil {
		t.Fatalf("NewPropertyReportValidator() error = %v", err)
	}
	return v
}

func newTestEventValidator(t *testing.T, reader ThingsModelReader) *EventReportValidator {
	t.Helper()
	v, err := NewEventReportValidator(reader, DefaultCacheTTL)
	if err != nil {
		t.Fatalf("NewEventReportValidator() error = %v", err)
	}
	return v
}

func newTestServiceCallValidator(t *testing.T, reader ThingsModelReader) *ServiceCallValidator {
	t.Helper()
	v, err := NewServiceCallValidator(reader, DefaultCacheTTL)
	if err != nil {
		t.Fatalf("NewServiceCallValidator() error = %v", err)
	}
	return v
}

func sampleDefinition() ThingsModelDefinition {
	return ThingsModelDefinition{
		Properties: map[string]PropertyDefinition{
			"temperature": {DataType: DataTypeFloat, HasMin: true, Min: -40, HasMax: true, Max: 125, HasStep: true, Step: 0.1, HasPrecision: true, Precision: 1},
			"battery":     {DataType: DataTypeInt, HasMin: true, Min: 0, HasMax: true, Max: 100},
			"online":      {DataType: DataTypeBool},
			"status":      {DataType: DataTypeString},
		},
		Events: map[string]EventDefinition{
			"temperature_alarm": {
				Output: map[string]PropertyDefinition{
					"temperature": {DataType: DataTypeFloat, HasMin: true, Min: -40, HasMax: true, Max: 125, HasPrecision: true, Precision: 1},
					"level":       {DataType: DataTypeString},
				},
			},
			"button_pressed": {
				Output: map[string]PropertyDefinition{},
			},
		},
		Services: map[string]ServiceDefinition{
			"reboot": {
				Output: map[string]PropertyDefinition{
					"accepted": {DataType: DataTypeBool},
				},
			},
			"sync_time": {
				Output: map[string]PropertyDefinition{},
			},
		},
	}
}

func TestValidateAcceptsKnownAndDropsUnknown(t *testing.T) {
	reader := &fakeReader{def: sampleDefinition()}
	v := newTestValidator(t, reader)

	result, err := v.Validate(context.Background(), Input{
		TenantID:  "tenant-1",
		ProductID: "product-1",
		Properties: map[string]any{
			"temperature": 23.5,
			"battery":     float64(88),
			"online":      true,
			"status":      "ok",
			"debug_raw":   "ignored",
		},
	})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if len(result.Accepted) != 4 {
		t.Fatalf("accepted count = %d, want 4", len(result.Accepted))
	}
	if _, exists := result.Accepted["debug_raw"]; exists {
		t.Fatal("debug_raw should not be in accepted")
	}
	if len(result.Dropped) != 1 || result.Dropped[0] != "debug_raw" {
		t.Fatalf("dropped = %v, want [debug_raw]", result.Dropped)
	}
}

func TestValidateRejectsValueBelowMin(t *testing.T) {
	v := newTestValidator(t, &fakeReader{def: sampleDefinition()})

	_, err := v.Validate(context.Background(), Input{
		TenantID:  "tenant-1",
		ProductID: "product-1",
		Properties: map[string]any{
			"battery": float64(-1),
		},
	})
	if !errors.Is(err, ErrInvalidPropertyValue) {
		t.Fatalf("err = %v, want ErrInvalidPropertyValue", err)
	}
}

func TestValidateRejectsIntWithFractionalPart(t *testing.T) {
	v := newTestValidator(t, &fakeReader{def: sampleDefinition()})

	_, err := v.Validate(context.Background(), Input{
		TenantID:  "tenant-1",
		ProductID: "product-1",
		Properties: map[string]any{
			"battery": 88.5,
		},
	})
	if !errors.Is(err, ErrInvalidPropertyValue) {
		t.Fatalf("err = %v, want ErrInvalidPropertyValue", err)
	}
}

func TestValidateAcceptsStepAlignedFloat(t *testing.T) {
	v := newTestValidator(t, &fakeReader{def: sampleDefinition()})

	for _, value := range []float64{0.3, -39.9, 23.5, 125.0, -40.0} {
		_, err := v.Validate(context.Background(), Input{
			TenantID:  "tenant-1",
			ProductID: "product-1",
			Properties: map[string]any{
				"temperature": value,
			},
		})
		if err != nil {
			t.Fatalf("Validate(%v) error = %v, want nil", value, err)
		}
	}
}

func TestValidateRejectsStepMismatch(t *testing.T) {
	v := newTestValidator(t, &fakeReader{def: sampleDefinition()})

	_, err := v.Validate(context.Background(), Input{
		TenantID:  "tenant-1",
		ProductID: "product-1",
		Properties: map[string]any{
			"temperature": 23.55,
		},
	})
	if !errors.Is(err, ErrInvalidPropertyValue) {
		t.Fatalf("err = %v, want ErrInvalidPropertyValue", err)
	}
}

func TestValidateReturnsNotFoundFromReader(t *testing.T) {
	v := newTestValidator(t, &fakeReader{err: ErrThingsModelNotFound})

	_, err := v.Validate(context.Background(), Input{
		TenantID:   "tenant-1",
		ProductID:  "product-1",
		Properties: map[string]any{"temperature": 23.5},
	})
	if !errors.Is(err, ErrThingsModelNotFound) {
		t.Fatalf("err = %v, want ErrThingsModelNotFound", err)
	}
}

func TestValidateReturnsNoAcceptedWhenAllUnknown(t *testing.T) {
	v := newTestValidator(t, &fakeReader{def: sampleDefinition()})

	_, err := v.Validate(context.Background(), Input{
		TenantID:  "tenant-1",
		ProductID: "product-1",
		Properties: map[string]any{
			"unknown_a": 1.0,
			"unknown_b": "x",
		},
	})
	if !errors.Is(err, ErrNoAcceptedProperties) {
		t.Fatalf("err = %v, want ErrNoAcceptedProperties", err)
	}
}

func TestValidateEventAcceptsKnownAndDropsUnknownParams(t *testing.T) {
	reader := &fakeReader{def: sampleDefinition()}
	v := newTestEventValidator(t, reader)

	result, err := v.Validate(context.Background(), EventInput{
		TenantID:  "tenant-1",
		ProductID: "product-1",
		EventName: "temperature_alarm",
		Params: map[string]any{
			"temperature": 85.2,
			"level":       "warning",
			"debug_raw":   "ignored",
		},
	})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if len(result.Accepted) != 2 {
		t.Fatalf("accepted count = %d, want 2", len(result.Accepted))
	}
	if _, exists := result.Accepted["debug_raw"]; exists {
		t.Fatal("debug_raw should not be in accepted")
	}
	if len(result.Dropped) != 1 || result.Dropped[0] != "debug_raw" {
		t.Fatalf("dropped = %v, want [debug_raw]", result.Dropped)
	}
}

func TestValidateEventReturnsEventNotFound(t *testing.T) {
	v := newTestEventValidator(t, &fakeReader{def: sampleDefinition()})

	_, err := v.Validate(context.Background(), EventInput{
		TenantID:  "tenant-1",
		ProductID: "product-1",
		EventName: "unknown_event",
		Params:    map[string]any{"temperature": 85.2},
	})
	if !errors.Is(err, ErrEventNotFound) {
		t.Fatalf("err = %v, want ErrEventNotFound", err)
	}
}

func TestValidateEventRejectsInvalidParamValue(t *testing.T) {
	v := newTestEventValidator(t, &fakeReader{def: sampleDefinition()})

	_, err := v.Validate(context.Background(), EventInput{
		TenantID:  "tenant-1",
		ProductID: "product-1",
		EventName: "temperature_alarm",
		Params: map[string]any{
			"temperature": "hot",
		},
	})
	if !errors.Is(err, ErrInvalidEventValue) {
		t.Fatalf("err = %v, want ErrInvalidEventValue", err)
	}
}

func TestValidateEventAcceptsNoOutputEvent(t *testing.T) {
	v := newTestEventValidator(t, &fakeReader{def: sampleDefinition()})

	result, err := v.Validate(context.Background(), EventInput{
		TenantID:  "tenant-1",
		ProductID: "product-1",
		EventName: "button_pressed",
		Params: map[string]any{
			"ignored": true,
		},
	})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if len(result.Accepted) != 0 {
		t.Fatalf("accepted = %v, want empty", result.Accepted)
	}
	if len(result.Dropped) != 1 || result.Dropped[0] != "ignored" {
		t.Fatalf("dropped = %v, want [ignored]", result.Dropped)
	}
}

func TestValidateServiceCallAcceptsKnownAndDropsUnknownOutput(t *testing.T) {
	reader := &fakeReader{def: sampleDefinition()}
	v := newTestServiceCallValidator(t, reader)

	result, err := v.Validate(context.Background(), ServiceCallInput{
		TenantID:    "tenant-1",
		ProductID:   "product-1",
		ServiceName: "reboot",
		Output: map[string]any{
			"accepted":  true,
			"debug_raw": "ignored",
		},
	})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if len(result.Accepted) != 1 {
		t.Fatalf("accepted count = %d, want 1", len(result.Accepted))
	}
	if _, exists := result.Accepted["debug_raw"]; exists {
		t.Fatal("debug_raw should not be in accepted")
	}
	if len(result.Dropped) != 1 || result.Dropped[0] != "debug_raw" {
		t.Fatalf("dropped = %v, want [debug_raw]", result.Dropped)
	}
}

func TestValidateServiceCallReturnsServiceNotFound(t *testing.T) {
	v := newTestServiceCallValidator(t, &fakeReader{def: sampleDefinition()})

	_, err := v.Validate(context.Background(), ServiceCallInput{
		TenantID:    "tenant-1",
		ProductID:   "product-1",
		ServiceName: "unknown_service",
		Output:      map[string]any{"accepted": true},
	})
	if !errors.Is(err, ErrServiceNotFound) {
		t.Fatalf("err = %v, want ErrServiceNotFound", err)
	}
}

func TestValidateServiceCallRejectsInvalidOutputValue(t *testing.T) {
	v := newTestServiceCallValidator(t, &fakeReader{def: sampleDefinition()})

	_, err := v.Validate(context.Background(), ServiceCallInput{
		TenantID:    "tenant-1",
		ProductID:   "product-1",
		ServiceName: "reboot",
		Output: map[string]any{
			"accepted": "yes",
		},
	})
	if !errors.Is(err, ErrInvalidServiceOutput) {
		t.Fatalf("err = %v, want ErrInvalidServiceOutput", err)
	}
}

func TestValidateServiceCallAcceptsNoOutputService(t *testing.T) {
	v := newTestServiceCallValidator(t, &fakeReader{def: sampleDefinition()})

	result, err := v.Validate(context.Background(), ServiceCallInput{
		TenantID:    "tenant-1",
		ProductID:   "product-1",
		ServiceName: "sync_time",
		Output: map[string]any{
			"ignored": true,
		},
	})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if len(result.Accepted) != 0 {
		t.Fatalf("accepted = %v, want empty", result.Accepted)
	}
	if len(result.Dropped) != 1 || result.Dropped[0] != "ignored" {
		t.Fatalf("dropped = %v, want [ignored]", result.Dropped)
	}
}

func TestValidateCachesLookup(t *testing.T) {
	reader := &fakeReader{def: sampleDefinition()}
	v := newTestValidator(t, reader)

	for i := 0; i < 3; i++ {
		if _, err := v.Validate(context.Background(), Input{
			TenantID:  "tenant-1",
			ProductID: "product-1",
			Properties: map[string]any{
				"temperature": 23.5,
			},
		}); err != nil {
			t.Fatalf("Validate() error = %v", err)
		}
	}
	if reader.calls != 1 {
		t.Fatalf("reader calls = %d, want 1 (cache hit)", reader.calls)
	}
}

func TestValidateCacheExpiresAndReloads(t *testing.T) {
	reader := &fakeReader{def: sampleDefinition()}
	v := newTestValidator(t, reader)

	current := time.Unix(0, 0)
	v.now = func() time.Time { return current }

	if _, err := v.Validate(context.Background(), Input{
		TenantID:   "tenant-1",
		ProductID:  "product-1",
		Properties: map[string]any{"temperature": 23.5},
	}); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	current = current.Add(v.ttl + time.Second)

	if _, err := v.Validate(context.Background(), Input{
		TenantID:   "tenant-1",
		ProductID:  "product-1",
		Properties: map[string]any{"temperature": 23.5},
	}); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if reader.calls != 2 {
		t.Fatalf("reader calls = %d, want 2 (cache expired)", reader.calls)
	}
}

func TestParseDefinitionExtractsSpec(t *testing.T) {
	raw := []byte(`{
		"temperature": {
			"name": "Temperature",
			"data_type": "float",
			"access_mode": "readwrite",
			"required": true,
			"spec": {"min": -40, "max": 125, "step": 0.1, "precision": 1, "unit": "celsius"}
		},
		"switch": {
			"name": "Switch",
			"data_type": "bool",
			"required": false
		}
	}`)
	def, err := ParseDefinition("tenant-1", "product-1", raw, []byte(`{}`))
	if err != nil {
		t.Fatalf("ParseDefinition() error = %v", err)
	}
	temp, ok := def.Properties["temperature"]
	if !ok {
		t.Fatal("temperature property missing")
	}
	if temp.DataType != DataTypeFloat || !temp.Required || !temp.HasMin || temp.Min != -40 || !temp.HasStep || temp.Step != 0.1 || !temp.HasPrecision || temp.Precision != 1 {
		t.Fatalf("temperature definition = %+v", temp)
	}
	sw, ok := def.Properties["switch"]
	if !ok {
		t.Fatal("switch property missing")
	}
	if sw.DataType != DataTypeBool || sw.Required {
		t.Fatalf("switch definition = %+v", sw)
	}
}
