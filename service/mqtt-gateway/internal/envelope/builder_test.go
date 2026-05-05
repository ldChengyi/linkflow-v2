package envelope

import (
	"encoding/json"
	"testing"
	"time"
)

func TestBuilderNewCreatesEnvelope(t *testing.T) {
	builder := Builder{
		Producer: "mqtt-gateway",
		TenantID: "default",
	}

	payload := map[string]any{
		"device_id":   "dev-001",
		"product_key": "esp32",
		"protocol":    "mqtt",
		"metrics": map[string]any{
			"temperature": 23.5,
		},
	}

	occurredAt := time.Date(2026, 5, 5, 10, 0, 0, 0, time.UTC)

	env, err := builder.New("device.telemetry.received", 1, payload, occurredAt)
	if err != nil {
		t.Fatalf("Builder.New() error = %v", err)
	}

	if env.EventID == "" {
		t.Fatal("EventID is empty")
	}
	if env.EventType != "device.telemetry.received" {
		t.Fatalf("EventType = %q", env.EventType)
	}
	if env.EventVersion != 1 {
		t.Fatalf("EventVersion = %d", env.EventVersion)
	}
	if env.Producer != "mqtt-gateway" {
		t.Fatalf("Producer = %q", env.Producer)
	}
	if env.TenantID != "default" {
		t.Fatalf("TenantID = %q", env.TenantID)
	}
	if env.OccurredAt != occurredAt.Format(time.RFC3339Nano) {
		t.Fatalf("OccurredAt = %q", env.OccurredAt)
	}

	var gotPayload map[string]any
	if err := json.Unmarshal(env.Payload, &gotPayload); err != nil {
		t.Fatalf("json.Unmarshal(env.Payload) error = %v", err)
	}
	if gotPayload["device_id"] != "dev-001" {
		t.Fatalf("payload device_id = %v", gotPayload["device_id"])
	}
}

func TestBuilderNewUsesCurrentTimeWhenOccurredAtIsZero(t *testing.T) {
	builder := Builder{
		Producer: "mqtt-gateway",
		TenantID: "default",
	}

	env, err := builder.New("device.telemetry.received", 1, map[string]any{"ok": true},
		time.Time{})
	if err != nil {
		t.Fatalf("Builder.New() error = %v", err)
	}

	if env.OccurredAt == "" {
		t.Fatal("OccurredAt is empty")
	}

	if _, err := time.Parse(time.RFC3339Nano, env.OccurredAt); err != nil {
		t.Fatalf("OccurredAt is not RFC3339Nano: %v", err)
	}
}

func TestEnvelopeJSONOmitEmptyFields(t *testing.T) {
	env := Envelope{
		EventID:      "018f56d3-7cb7-7f1a-9b41-3f3a63fd3db5",
		EventType:    "device.telemetry.received",
		EventVersion: 1,
		OccurredAt:   "2026-05-05T10:00:00Z",
		Producer:     "mqtt-gateway",
		TenantID:     "default",
		Payload:      json.RawMessage(`{"ok":true}`),
	}

	b, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("json.Marshal(Envelope) error = %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("json.Unmarshal(envelope JSON) error = %v", err)
	}

	for _, field := range []string{"trace_id", "correlation_id", "causation_id"} {
		if _, ok := got[field]; ok {
			t.Fatalf("empty field %q should be omitted, got JSON: %s", field, string(b))
		}
	}
}
