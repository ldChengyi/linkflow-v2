package schema

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
)

func TestValidateEventAcceptsTelemetryReceived(t *testing.T) {
	registry := newTestRegistry(t)

	env := validTelemetryEvent()
	raw := mustMarshalJSON(t, env)

	got, err := registry.ValidateEvent(raw)
	if err != nil {
		t.Fatalf("ValidateEvent() error = %v", err)
	}

	if got.EventType != event.TypeDeviceTelemetryReceived {
		t.Fatalf("EventType = %q", got.EventType)
	}
	if got.EventVersion != event.VersionDeviceTelemetryReceived {
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
			env := validTelemetryEvent()
			tt.mutate(env)

			if _, err := registry.ValidateEvent(mustMarshalJSON(t, env)); err == nil {
				t.Fatal("ValidateEvent() error = nil, want validation error")
			}
		})
	}
}

func TestValidateEventRejectsUnknownPayloadSchema(t *testing.T) {
	registry := newTestRegistry(t)

	env := validTelemetryEvent()
	env["event_version"] = 2

	if _, err := registry.ValidateEvent(mustMarshalJSON(t, env)); err == nil {
		t.Fatal("ValidateEvent() error = nil, want unknown payload schema error")
	}
}

func TestValidateEventRejectsInvalidTelemetryPayload(t *testing.T) {
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
				"metrics": map[string]any{
					"temperature": 23.5,
				},
			},
		},
		{
			name: "missing metrics",
			payload: map[string]any{
				"device_id":   "dev-001",
				"product_key": "esp32",
				"protocol":    "mqtt",
			},
		},
		{
			name: "empty metrics",
			payload: map[string]any{
				"device_id":   "dev-001",
				"product_key": "esp32",
				"protocol":    "mqtt",
				"metrics":     map[string]any{},
			},
		},
		{
			name: "invalid metric name",
			payload: map[string]any{
				"device_id":   "dev-001",
				"product_key": "esp32",
				"protocol":    "mqtt",
				"metrics": map[string]any{
					"Temperature": 23.5,
				},
			},
		},
		{
			name: "unsupported metric value type",
			payload: map[string]any{
				"device_id":   "dev-001",
				"product_key": "esp32",
				"protocol":    "mqtt",
				"metrics": map[string]any{
					"temperature": map[string]any{"value": 23.5},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := validTelemetryEvent()
			env["payload"] = tt.payload

			if _, err := registry.ValidateEvent(mustMarshalJSON(t, env)); err == nil {
				t.Fatal("ValidateEvent() error = nil, want validation error")
			}
		})
	}
}

func newTestRegistry(t *testing.T) *Registry {
	t.Helper()

	registry, err := NewRegistry(os.DirFS("../../../../contracts"))
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}

	return registry
}

func validTelemetryEvent() map[string]any {
	return map[string]any{
		"event_id":      "018f56d3-7cb7-7f1a-9b41-3f3a63fd3db5",
		"event_type":    event.TypeDeviceTelemetryReceived,
		"event_version": event.VersionDeviceTelemetryReceived,
		"occurred_at":   "2026-05-05T10:00:00Z",
		"producer":      "mqtt-gateway",
		"tenant_id":     "default",
		"payload": map[string]any{
			"device_id":   "dev-001",
			"product_key": "esp32",
			"protocol":    "mqtt",
			"metrics": map[string]any{
				"temperature": 23.5,
				"online":      true,
				"status":      "ok",
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
