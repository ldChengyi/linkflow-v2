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

func TestValidateEventAcceptsDeviceEventReported(t *testing.T) {
	registry := newTestRegistry(t)

	env := validDeviceEventReportedEvent()
	raw := mustMarshalJSON(t, env)

	got, err := registry.ValidateEvent(raw)
	if err != nil {
		t.Fatalf("ValidateEvent() error = %v", err)
	}

	if got.EventType != event.DeviceEventReported.Type {
		t.Fatalf("EventType = %q", got.EventType)
	}
	if got.EventVersion != event.DeviceEventReported.Version {
		t.Fatalf("EventVersion = %d", got.EventVersion)
	}
}

func TestValidateEventAcceptsDeviceServiceCallRequested(t *testing.T) {
	registry := newTestRegistry(t)

	env := validDeviceServiceCallRequestedEvent()
	raw := mustMarshalJSON(t, env)

	got, err := registry.ValidateEvent(raw)
	if err != nil {
		t.Fatalf("ValidateEvent() error = %v", err)
	}

	if got.EventType != event.DeviceServiceCallRequested.Type {
		t.Fatalf("EventType = %q", got.EventType)
	}
	if got.EventVersion != event.DeviceServiceCallRequested.Version {
		t.Fatalf("EventVersion = %d", got.EventVersion)
	}
}

func TestValidateEventAcceptsDeviceServiceCallAcknowledged(t *testing.T) {
	registry := newTestRegistry(t)

	env := validDeviceServiceCallAcknowledgedEvent()
	raw := mustMarshalJSON(t, env)

	got, err := registry.ValidateEvent(raw)
	if err != nil {
		t.Fatalf("ValidateEvent() error = %v", err)
	}

	if got.EventType != event.DeviceServiceCallAcknowledged.Type {
		t.Fatalf("EventType = %q", got.EventType)
	}
	if got.EventVersion != event.DeviceServiceCallAcknowledged.Version {
		t.Fatalf("EventVersion = %d", got.EventVersion)
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
			name: "missing tenant_id",
			payload: map[string]any{
				"product_id":  "product-1",
				"device_id":   "device-1",
				"tenant_slug": "default",
				"device_slug": "dev-001",
				"product_key": "esp32",
				"protocol":    "mqtt",
				"properties": map[string]any{
					"temperature": 23.5,
				},
			},
		},
		{
			name: "missing product_id",
			payload: map[string]any{
				"tenant_id":   "default",
				"device_id":   "device-1",
				"tenant_slug": "default",
				"device_slug": "dev-001",
				"product_key": "esp32",
				"protocol":    "mqtt",
				"properties": map[string]any{
					"temperature": 23.5,
				},
			},
		},
		{
			name: "missing device_id",
			payload: map[string]any{
				"tenant_id":   "default",
				"product_id":  "product-1",
				"tenant_slug": "default",
				"device_slug": "dev-001",
				"product_key": "esp32",
				"protocol":    "mqtt",
				"properties": map[string]any{
					"temperature": 23.5,
				},
			},
		},
		{
			name: "missing tenant_slug",
			payload: map[string]any{
				"tenant_id":   "default",
				"product_id":  "product-1",
				"device_id":   "device-1",
				"device_slug": "dev-001",
				"product_key": "esp32",
				"protocol":    "mqtt",
				"properties": map[string]any{
					"temperature": 23.5,
				},
			},
		},
		{
			name: "missing device_slug",
			payload: map[string]any{
				"tenant_id":   "default",
				"product_id":  "product-1",
				"device_id":   "device-1",
				"tenant_slug": "default",
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
				"tenant_id":   "default",
				"product_id":  "product-1",
				"device_id":   "device-1",
				"tenant_slug": "default",
				"device_slug": "dev-001",
				"product_key": "esp32",
				"protocol":    "mqtt",
			},
		},
		{
			name: "empty properties",
			payload: map[string]any{
				"tenant_id":   "default",
				"product_id":  "product-1",
				"device_id":   "device-1",
				"tenant_slug": "default",
				"device_slug": "dev-001",
				"product_key": "esp32",
				"protocol":    "mqtt",
				"properties":  map[string]any{},
			},
		},
		{
			name: "invalid property name",
			payload: map[string]any{
				"tenant_id":   "default",
				"product_id":  "product-1",
				"device_id":   "device-1",
				"tenant_slug": "default",
				"device_slug": "dev-001",
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
				"tenant_id":   "default",
				"product_id":  "product-1",
				"device_id":   "device-1",
				"tenant_slug": "default",
				"device_slug": "dev-001",
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

func TestValidateEventRejectsInvalidPropertySetAcknowledgedPayload(t *testing.T) {
	registry := newTestRegistry(t)

	env := validPropertySetAcknowledgedEvent()
	payload := env["payload"].(map[string]any)
	delete(payload, "tenant_slug")

	if _, err := registry.ValidateEvent(mustMarshalJSON(t, env)); err == nil {
		t.Fatal("ValidateEvent() error = nil, want validation error")
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
			"tenant_id":   "default",
			"product_id":  "product-1",
			"device_id":   "device-1",
			"tenant_slug": "default",
			"device_slug": "dev-001",
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
			"tenant_id":   "default",
			"product_id":  "product-1",
			"device_id":   "device-1",
			"tenant_slug": "default",
			"device_slug": "dev-001",
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

func validDeviceEventReportedEvent() map[string]any {
	return map[string]any{
		"event_id":      "018f56d3-7cb7-7f1a-9b41-3f3a63fd3db7",
		"event_type":    event.DeviceEventReported.Type,
		"event_version": event.DeviceEventReported.Version,
		"occurred_at":   "2026-05-05T10:00:00Z",
		"producer":      "emqx-rule-engine",
		"tenant_id":     "default",
		"payload": map[string]any{
			"tenant_id":   "default",
			"product_id":  "product-1",
			"device_id":   "device-1",
			"tenant_slug": "default",
			"product_key": "esp32",
			"device_slug": "dev-001",
			"protocol":    "mqtt",
			"event_name":  "overheat",
			"params": map[string]any{
				"temperature": 85.2,
			},
		},
	}
}

func validDeviceServiceCallRequestedEvent() map[string]any {
	return map[string]any{
		"event_id":       "018f56d3-7cb7-7f1a-9b41-3f3a63fd3db6",
		"event_type":     event.DeviceServiceCallRequested.Type,
		"event_version":  event.DeviceServiceCallRequested.Version,
		"occurred_at":    "2026-05-05T10:00:00Z",
		"producer":       "backend",
		"tenant_id":      "default",
		"correlation_id": "018f56d3-7cb7-7f1a-9b41-3f3a63fd3db9",
		"payload": map[string]any{
			"command_id":   "018f56d3-7cb7-7f1a-9b41-3f3a63fd3db9",
			"tenant_id":    "default",
			"product_id":   "product-1",
			"device_id":    "device-1",
			"tenant_slug":  "default",
			"product_key":  "esp32",
			"device_slug":  "dev-001",
			"protocol":     "mqtt",
			"topic":        "lf/v1/default/esp32/dev-001/service/down/reboot",
			"service_name": "reboot",
			"input": map[string]any{
				"delay": 5,
			},
			"requested_by": "user-1",
		},
	}
}

func validDeviceServiceCallAcknowledgedEvent() map[string]any {
	return map[string]any{
		"event_id":      "018f56d3-7cb7-7f1a-9b41-3f3a63fd3db8",
		"event_type":    event.DeviceServiceCallAcknowledged.Type,
		"event_version": event.DeviceServiceCallAcknowledged.Version,
		"occurred_at":   "2026-05-05T10:00:00Z",
		"producer":      "emqx-rule-engine",
		"tenant_id":     "default",
		"causation_id":  "018f56d3-7cb7-7f1a-9b41-3f3a63fd3db9",
		"payload": map[string]any{
			"command_id":   "018f56d3-7cb7-7f1a-9b41-3f3a63fd3db9",
			"tenant_id":    "default",
			"product_id":   "product-1",
			"device_id":    "device-1",
			"tenant_slug":  "default",
			"product_key":  "esp32",
			"device_slug":  "dev-001",
			"protocol":     "mqtt",
			"service_name": "reboot",
			"success":      true,
			"code":         "ok",
			"message":      "done",
			"output": map[string]any{
				"accepted": true,
			},
		},
	}
}

func validDeviceConnectedEvent() map[string]any {
	return map[string]any{
		"event_id":      "018f56d3-7cb7-7f1a-9b41-3f3a63fd3dc1",
		"event_type":    event.DeviceConnected.Type,
		"event_version": event.DeviceConnected.Version,
		"occurred_at":   "2026-05-05T10:00:00Z",
		"producer":      "emqx-rule-engine",
		"tenant_id":     "default",
		"payload": map[string]any{
			"tenant_id":   "default",
			"product_id":  "product-1",
			"device_id":   "device-1",
			"tenant_slug": "default",
			"product_key": "esp32",
			"device_slug": "dev-001",
			"protocol":    "mqtt",
			"keepalive":   60,
		},
	}
}

func validDeviceConnectionChangedEvent() map[string]any {
	return map[string]any{
		"event_id":      "018f56d3-7cb7-7f1a-9b41-3f3a63fd3dd1",
		"event_type":    event.DeviceConnectionChanged.Type,
		"event_version": event.DeviceConnectionChanged.Version,
		"occurred_at":   "2026-05-05T10:00:01Z",
		"producer":      "device-event-processor",
		"tenant_id":     "default",
		"causation_id":  "018f56d3-7cb7-7f1a-9b41-3f3a63fd3dc1",
		"payload": map[string]any{
			"tenant_id":   "default",
			"product_id":  "product-1",
			"device_id":   "device-1",
			"tenant_slug": "default",
			"product_key": "esp32",
			"device_slug": "dev-001",
			"status":      "online",
		},
	}
}

func validDevicePropertyChangedEvent() map[string]any {
	return map[string]any{
		"event_id":      "018f56d3-7cb7-7f1a-9b41-3f3a63fd3dd2",
		"event_type":    event.DevicePropertyChanged.Type,
		"event_version": event.DevicePropertyChanged.Version,
		"occurred_at":   "2026-05-05T10:00:02Z",
		"producer":      "device-event-processor",
		"tenant_id":     "default",
		"causation_id":  "018f56d3-7cb7-7f1a-9b41-3f3a63fd3db5",
		"payload": map[string]any{
			"tenant_id":   "default",
			"product_id":  "product-1",
			"device_id":   "device-1",
			"tenant_slug": "default",
			"product_key": "esp32",
			"device_slug": "dev-001",
			"properties": map[string]any{
				"temperature": 23.5,
			},
		},
	}
}

func validDeviceEventReceivedEvent() map[string]any {
	return map[string]any{
		"event_id":      "018f56d3-7cb7-7f1a-9b41-3f3a63fd3dd3",
		"event_type":    event.DeviceEventReceived.Type,
		"event_version": event.DeviceEventReceived.Version,
		"occurred_at":   "2026-05-05T10:00:03Z",
		"producer":      "device-event-processor",
		"tenant_id":     "default",
		"causation_id":  "018f56d3-7cb7-7f1a-9b41-3f3a63fd3db6",
		"payload": map[string]any{
			"tenant_id":   "default",
			"product_id":  "product-1",
			"device_id":   "device-1",
			"tenant_slug": "default",
			"product_key": "esp32",
			"device_slug": "dev-001",
			"event_name":  "temperature_alarm",
			"params": map[string]any{
				"temperature": 85.2,
			},
		},
	}
}

func validDeviceDisconnectedEvent() map[string]any {
	return map[string]any{
		"event_id":      "018f56d3-7cb7-7f1a-9b41-3f3a63fd3dc2",
		"event_type":    event.DeviceDisconnected.Type,
		"event_version": event.DeviceDisconnected.Version,
		"occurred_at":   "2026-05-05T10:01:00Z",
		"producer":      "emqx-rule-engine",
		"tenant_id":     "default",
		"payload": map[string]any{
			"tenant_id":   "default",
			"product_id":  "product-1",
			"device_id":   "device-1",
			"tenant_slug": "default",
			"product_key": "esp32",
			"device_slug": "dev-001",
			"protocol":    "mqtt",
			"reason":      "keepalive_timeout",
		},
	}
}

func TestValidateEventAcceptsDeviceConnectedAndDisconnected(t *testing.T) {
	registry := newTestRegistry(t)

	for _, env := range []map[string]any{
		validDeviceConnectedEvent(),
		validDeviceDisconnectedEvent(),
	} {
		if _, err := registry.ValidateEvent(mustMarshalJSON(t, env)); err != nil {
			t.Fatalf("ValidateEvent(%q) error = %v", env["event_type"], err)
		}
	}
}

func TestValidateEventAcceptsDeviceStateChangedEvents(t *testing.T) {
	registry := newTestRegistry(t)

	for _, env := range []map[string]any{
		validDeviceConnectionChangedEvent(),
		validDevicePropertyChangedEvent(),
		validDeviceEventReceivedEvent(),
	} {
		if _, err := registry.ValidateEvent(mustMarshalJSON(t, env)); err != nil {
			t.Fatalf("ValidateEvent(%q) error = %v", env["event_type"], err)
		}
	}
}

func TestValidateEventRejectsDeviceConnectedMissingDeviceID(t *testing.T) {
	registry := newTestRegistry(t)

	env := validDeviceConnectedEvent()
	payload := env["payload"].(map[string]any)
	delete(payload, "device_id")

	if _, err := registry.ValidateEvent(mustMarshalJSON(t, env)); err == nil {
		t.Fatal("ValidateEvent() error = nil, want missing device_id error")
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
