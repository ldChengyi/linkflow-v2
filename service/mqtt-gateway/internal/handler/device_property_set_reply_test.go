package handler

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/router"
)

func TestDevicePropertySetReplyPublishesAcknowledgedEvent(t *testing.T) {
	pub := &fakePublisher{}
	factory := event.EnvelopeFactory{
		Producer: "mqtt-gateway",
		TenantID: "default",
	}

	h := DevicePropertySetReply(event.DevicePropertySetAcknowledged, factory, pub)

	err := h.Handle(context.Background(), router.ParsedMessage{
		Topic: "lf/v1/default/esp32/dev-001/property/up/set_reply",
		Vars: map[string]string{
			"tenant_slug": "default",
			"product_key": "esp32",
			"device_slug": "dev-001",
		},
		Payload: []byte(`{"success":true,"code":"ok","message":"applied","properties":{"led":true}}`),
	})
	if err != nil {
		t.Fatalf("DevicePropertySetReply() error = %v", err)
	}

	if len(pub.events) != 1 {
		t.Fatalf("published events = %d", len(pub.events))
	}

	env := pub.events[0]
	if env.EventType != event.DevicePropertySetAcknowledged.Type {
		t.Fatalf("EventType = %q", env.EventType)
	}
	if env.EventVersion != event.DevicePropertySetAcknowledged.Version {
		t.Fatalf("EventVersion = %d", env.EventVersion)
	}

	var payload event.PropertySetAcknowledgedPayload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		t.Fatalf("json.Unmarshal(env.Payload) error = %v", err)
	}

	if payload.TenantSlug != "default" {
		t.Fatalf("TenantSlug = %q", payload.TenantSlug)
	}
	if payload.DeviceSlug != "dev-001" {
		t.Fatalf("DeviceSlug = %q", payload.DeviceSlug)
	}
	if payload.ProductKey != "esp32" {
		t.Fatalf("ProductKey = %q", payload.ProductKey)
	}
	if payload.Protocol != "mqtt" {
		t.Fatalf("Protocol = %q", payload.Protocol)
	}
	if !payload.Success {
		t.Fatal("Success = false")
	}
	if payload.Code != "ok" {
		t.Fatalf("Code = %q", payload.Code)
	}
	if payload.Message != "applied" {
		t.Fatalf("Message = %q", payload.Message)
	}
	if payload.Properties["led"] != true {
		t.Fatalf("led = %v", payload.Properties["led"])
	}
}

func TestDevicePropertySetReplyReturnsErrorForMissingSuccess(t *testing.T) {
	pub := &fakePublisher{}
	factory := event.EnvelopeFactory{
		Producer: "mqtt-gateway",
		TenantID: "default",
	}

	h := DevicePropertySetReply(event.DevicePropertySetAcknowledged, factory, pub)

	err := h.Handle(context.Background(), router.ParsedMessage{
		Topic: "lf/v1/default/esp32/dev-001/property/up/set_reply",
		Vars: map[string]string{
			"tenant_slug": "default",
			"product_key": "esp32",
			"device_slug": "dev-001",
		},
		Payload: []byte(`{"code":"ok"}`),
	})
	if err == nil {
		t.Fatal("DevicePropertySetReply() error is nil, want missing success error")
	}

	if len(pub.events) != 0 {
		t.Fatalf("published events = %d", len(pub.events))
	}
}
