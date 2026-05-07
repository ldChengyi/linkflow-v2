package validation

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
)

func TestValidateEventAcceptsPropertyReported(t *testing.T) {
	registry := newTestRegistry(t)

	env := validPropertyEvent()
	raw := mustMarshalJSON(t, env)

	got, err := registry.ValidateEvent(raw)
	if err != nil {
		t.Fatalf("ValidateEvent() error = %v", err)
	}

	if got.EventType != event.DevicePropertyReported.Type {
		t.Fatalf("EventType = %q", got.EventType)
	}
	if got.EventVersion != event.DevicePropertyReported.Version {
		t.Fatalf("EventVersion = %d", got.EventVersion)
	}
	if len(got.Payload) == 0 {
		t.Fatal("Payload is empty")
	}
}

func TestValidateEventAcceptsPropertySetAcknowledged(t *testing.T) {
	registry := newTestRegistry(t)

	env := validPropertySetAcknowledgedEvent()
	raw := mustMarshalJSON(t, env)

	got, err := registry.ValidateEvent(raw)
	if err != nil {
		t.Fatalf("ValidateEvent() error = %v", err)
	}

	if got.EventType != event.DevicePropertySetAcknowledged.Type {
		t.Fatalf("EventType = %q", got.EventType)
	}
	if got.EventVersion != event.DevicePropertySetAcknowledged.Version {
		t.Fatalf("EventVersion = %d", got.EventVersion)
	}
	if len(got.Payload) == 0 {
		t.Fatal("Payload is empty")
	}
}

func TestValidateEventRejectsInvalidEnvelope(t *testing.T) {
	registry := newTestRegistry(t)

	tests := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{
			name: "missing event_id",
			mutate: func(env map[string]any) {
				delete(env, "event_id")
			},
		},
		{
			name: "missing tenant_id",
			mutate: func(env map[string]any) {
				delete(env, "tenant_id")
			},
		},
		{
			name: "invalid event_id format",
			mutate: func(env map[string]any) {
				env["event_id"] = "not-a-uuid"
			},
		},
		{
			name: "invalid occurred_at format",
			mutate: func(env map[string]any) {
				env["occurred_at"] = "2026/05/05 10:00:00"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := validPropertyEvent()
			tt.mutate(env)

			if _, err := registry.ValidateEvent(mustMarshalJSON(t, env)); err == nil {
				t.Fatal("ValidateEvent() error = nil, want validation error")
			}
		})
	}
}

func TestValidateEventRejectsUnknownPayloadSchema(t *testing.T) {
	registry := newTestRegistry(t)

	env := validPropertyEvent()
	env["event_version"] = 2

	if _, err := registry.ValidateEvent(mustMarshalJSON(t, env)); err == nil {
		t.Fatal("ValidateEvent() error = nil, want unknown payload schema error")
	}
}

func TestValidateEventRejectsInvalidPropertyPayload(t *testing.T) {
	registry := newTestRegistry(t)

	tests := []struct {
		name    string
		payload map[string]any
	}{
		{
			name: "missing device_id",
			payload: map[string]any{
				"product_key": "esp32",
				"protocol":    "mqtt",
				"properties": map[string]any{
					"temperature": 23.5,
				},
			},
		},
		{
			name: "missing properties",
			payload: map[string]any{
				"device_id":   "dev-001",
				"product_key": "esp32",
				"protocol":    "mqtt",
			},
		},
		{
			name: "empty properties",
			payload: map[string]any{
				"device_id":   "dev-001",
				"product_key": "esp32",
				"protocol":    "mqtt",
				"properties":  map[string]any{},
			},
		},
		{
			name: "invalid property name",
			payload: map[string]any{
				"device_id":   "dev-001",
				"product_key": "esp32",
				"protocol":    "mqtt",
				"properties": map[string]any{
					"Temperature": 23.5,
				},
			},
		},
		{
			name: "unsupported property value type",
			payload: map[string]any{
				"device_id":   "dev-001",
				"product_key": "esp32",
				"protocol":    "mqtt",
				"properties": map[string]any{
					"temperature": map[string]any{"value": 23.5},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := validPropertyEvent()
			env["payload"] = tt.payload

			if _, err := registry.ValidateEvent(mustMarshalJSON(t, env)); err == nil {
				t.Fatal("ValidateEvent() error = nil, want validation error")
			}
		})
	}
}

func newTestRegistry(t *testing.T) *Validator {
	t.Helper()

	registry, err := NewValidator(os.DirFS("../../../../../contracts"))
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	return registry
}

func validPropertyEvent() map[string]any {
	return map[string]any{
		"event_id":      "018f56d3-7cb7-7f1a-9b41-3f3a63fd3db5",
		"event_type":    event.DevicePropertyReported.Type,
		"event_version": event.DevicePropertyReported.Version,
		"occurred_at":   "2026-05-05T10:00:00Z",
		"producer":      "mqtt-gateway",
		"tenant_id":     "default",
		"payload": map[string]any{
			"device_id":   "dev-001",
			"product_key": "esp32",
			"protocol":    "mqtt",
			"properties": map[string]any{
				"temperature": 23.5,
				"online":      true,
				"status":      "ok",
			},
		},
	}
}

func validPropertySetAcknowledgedEvent() map[string]any {
	return map[string]any{
		"event_id":      "018f56d3-7cb7-7f1a-9b41-3f3a63fd3db6",
		"event_type":    event.DevicePropertySetAcknowledged.Type,
		"event_version": event.DevicePropertySetAcknowledged.Version,
		"occurred_at":   "2026-05-05T10:00:00Z",
		"producer":      "mqtt-gateway",
		"tenant_id":     "default",
		"payload": map[string]any{
			"device_id":   "dev-001",
			"product_key": "esp32",
			"protocol":    "mqtt",
			"success":     true,
			"code":        "ok",
			"message":     "applied",
			"properties": map[string]any{
				"led": true,
			},
		},
	}
}

func mustMarshalJSON(t *testing.T, v any) []byte {
	t.Helper()

	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	return b
}
