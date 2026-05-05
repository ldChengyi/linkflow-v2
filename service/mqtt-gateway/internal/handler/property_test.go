package handler

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/router"
)

type fakePublisher struct {
	events []event.Envelope
}

func (p *fakePublisher) Publish(ctx context.Context, e event.Envelope) error {
	p.events = append(p.events, e)
	return nil
}

func TestPropertyPublishesTelemetryReceivedEvent(t *testing.T) {
	pub := &fakePublisher{}
	factory := event.EnvelopeFactory{
		Producer: "mqtt-gateway",
		TenantID: "default",
	}

	h := Property(factory, pub)

	err := h(context.Background(), router.ParsedMessage{
		Topic: "lf/v1/esp32/dev-001/property/up/post",
		Vars: map[string]string{
			"product_key": "esp32",
			"device_id":   "dev-001",
		},
		Payload: []byte(`{"temperature":23.5,"online":true,"status":"ok"}`),
	})
	if err != nil {
		t.Fatalf("Property() error = %v", err)
	}

	if len(pub.events) != 1 {
		t.Fatalf("published events = %d", len(pub.events))
	}

	env := pub.events[0]
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

	var payload struct {
		DeviceID   string         `json:"device_id"`
		ProductKey string         `json:"product_key"`
		Protocol   string         `json:"protocol"`
		Metrics    map[string]any `json:"metrics"`
	}

	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		t.Fatalf("json.Unmarshal(env.Payload) error = %v", err)
	}

	if payload.DeviceID != "dev-001" {
		t.Fatalf("DeviceID = %q", payload.DeviceID)
	}
	if payload.ProductKey != "esp32" {
		t.Fatalf("ProductKey = %q", payload.ProductKey)
	}
	if payload.Protocol != "mqtt" {
		t.Fatalf("Protocol = %q", payload.Protocol)
	}
	if payload.Metrics["temperature"] != 23.5 {
		t.Fatalf("temperature = %v", payload.Metrics["temperature"])
	}
	if payload.Metrics["online"] != true {
		t.Fatalf("online = %v", payload.Metrics["online"])
	}
	if payload.Metrics["status"] != "ok" {
		t.Fatalf("status = %v", payload.Metrics["status"])
	}
}

func TestPropertyReturnsErrorForInvalidJSON(t *testing.T) {
	pub := &fakePublisher{}
	factory := event.EnvelopeFactory{
		Producer: "mqtt-gateway",
		TenantID: "default",
	}

	h := Property(factory, pub)

	err := h(context.Background(), router.ParsedMessage{
		Topic: "lf/v1/esp32/dev-001/property/up/post",
		Vars: map[string]string{
			"product_key": "esp32",
			"device_id":   "dev-001",
		},
		Payload: []byte(`{bad json`),
	})
	if err == nil {
		t.Fatal("Property() error is nil, want JSON decode error")
	}

	if len(pub.events) != 0 {
		t.Fatalf("published events = %d", len(pub.events))
	}
}
