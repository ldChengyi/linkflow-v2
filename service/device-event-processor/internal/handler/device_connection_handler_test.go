package handler

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
	"github.com/ldchengyi/linkflow-v2/pkg/public/messaging"
	"github.com/ldchengyi/linkflow-v2/service/device-event-processor/internal/store"
)

type fakeConnectionStore struct {
	online         []store.DeviceConnectionEvent
	offline        []store.DeviceConnectionEvent
	onlineIgnored  bool
	offlineIgnored bool
	onlineErr      error
	offErr         error
}

func (s *fakeConnectionStore) MarkOnline(ctx context.Context, in store.DeviceConnectionEvent) (bool, error) {
	s.online = append(s.online, in)
	return !s.onlineIgnored, s.onlineErr
}

func (s *fakeConnectionStore) MarkOffline(ctx context.Context, in store.DeviceConnectionEvent) (bool, error) {
	s.offline = append(s.offline, in)
	return !s.offlineIgnored, s.offErr
}

func newConnectedEnvelope(t *testing.T, payload event.ConnectedPayload) *event.Envelope {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return &event.Envelope{
		EventID:      "evt-c1",
		EventType:    event.DeviceConnected.Type,
		EventVersion: event.DeviceConnected.Version,
		TenantID:     payload.TenantID,
		OccurredAt:   "2026-05-05T10:00:00.000Z",
		Payload:      json.RawMessage(raw),
	}
}

func newDisconnectedEnvelope(t *testing.T, payload event.DisconnectedPayload) *event.Envelope {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return &event.Envelope{
		EventID:      "evt-d1",
		EventType:    event.DeviceDisconnected.Type,
		EventVersion: event.DeviceDisconnected.Version,
		TenantID:     payload.TenantID,
		OccurredAt:   "2026-05-05T10:01:00.000Z",
		Payload:      json.RawMessage(raw),
	}
}

func TestDeviceConnectedHandlerMarksOnline(t *testing.T) {
	s := &fakeConnectionStore{}
	pub := &fakePublisher{}
	h, err := NewDeviceConnectedHandler(s, pub, nil)
	if err != nil {
		t.Fatalf("NewDeviceConnectedHandler() error = %v", err)
	}

	result := h.Handle(context.Background(), newConnectedEnvelope(t, event.ConnectedPayload{
		TenantID:   "tenant-1",
		ProductID:  "product-1",
		DeviceID:   "device-1",
		TenantSlug: "tenant",
		ProductKey: "product",
		DeviceSlug: "device",
		Protocol:   "mqtt",
		Keepalive:  60,
	}))

	if result.Decision != messaging.DecisionAck {
		t.Fatalf("decision = %q, want ack (err=%v)", result.Decision, result.Err)
	}
	if len(s.online) != 1 || s.online[0].DeviceID != "device-1" {
		t.Fatalf("online calls = %+v", s.online)
	}
	if len(s.offline) != 0 {
		t.Fatal("MarkOffline should not be called")
	}
	if len(pub.published) != 1 {
		t.Fatalf("publish count = %d, want 1", len(pub.published))
	}
	publishedIn := pub.published[0]
	if publishedIn.EventType != event.TypeDeviceConnectionChanged {
		t.Fatalf("published event_type = %q", publishedIn.EventType)
	}
	changed, ok := publishedIn.Payload.(event.ConnectionChangedPayload)
	if !ok {
		t.Fatalf("payload type = %T", publishedIn.Payload)
	}
	if changed.Status != event.ConnectionStatusOnline {
		t.Fatalf("status = %q, want online", changed.Status)
	}
}

func TestDeviceConnectedHandlerRetriesOnStoreError(t *testing.T) {
	s := &fakeConnectionStore{onlineErr: errors.New("redis down")}
	pub := &fakePublisher{}
	h, _ := NewDeviceConnectedHandler(s, pub, nil)

	result := h.Handle(context.Background(), newConnectedEnvelope(t, event.ConnectedPayload{
		TenantID: "tenant-1", ProductID: "product-1", DeviceID: "device-1",
		TenantSlug: "tenant", ProductKey: "product", DeviceSlug: "device", Protocol: "mqtt",
	}))

	if result.Decision != messaging.DecisionRetry {
		t.Fatalf("decision = %q, want retry", result.Decision)
	}
	if len(pub.published) != 0 {
		t.Fatal("publisher should not be called when store fails")
	}
}

func TestDeviceConnectedHandlerAcksIgnoredStoreEvent(t *testing.T) {
	s := &fakeConnectionStore{onlineIgnored: true}
	pub := &fakePublisher{}
	h, _ := NewDeviceConnectedHandler(s, pub, nil)

	result := h.Handle(context.Background(), newConnectedEnvelope(t, event.ConnectedPayload{
		TenantID: "tenant-1", ProductID: "product-1", DeviceID: "device-1",
		TenantSlug: "tenant", ProductKey: "product", DeviceSlug: "device", Protocol: "mqtt",
	}))

	if result.Decision != messaging.DecisionAck {
		t.Fatalf("decision = %q, want ack", result.Decision)
	}
	if len(s.online) != 1 {
		t.Fatal("store should still receive the connection event")
	}
	if len(pub.published) != 0 {
		t.Fatal("publisher should not be called when store ignores a stale event")
	}
}

func TestDeviceConnectedHandlerRetriesOnPublisherError(t *testing.T) {
	s := &fakeConnectionStore{}
	pub := &fakePublisher{err: errors.New("kafka down")}
	h, _ := NewDeviceConnectedHandler(s, pub, nil)

	result := h.Handle(context.Background(), newConnectedEnvelope(t, event.ConnectedPayload{
		TenantID: "tenant-1", ProductID: "product-1", DeviceID: "device-1",
		TenantSlug: "tenant", ProductKey: "product", DeviceSlug: "device", Protocol: "mqtt",
	}))

	if result.Decision != messaging.DecisionRetry {
		t.Fatalf("decision = %q, want retry", result.Decision)
	}
	if len(s.online) != 1 {
		t.Fatal("store should be called before publish")
	}
}

func TestDeviceDisconnectedHandlerMarksOffline(t *testing.T) {
	s := &fakeConnectionStore{}
	pub := &fakePublisher{}
	h, err := NewDeviceDisconnectedHandler(s, pub, nil)
	if err != nil {
		t.Fatalf("NewDeviceDisconnectedHandler() error = %v", err)
	}

	result := h.Handle(context.Background(), newDisconnectedEnvelope(t, event.DisconnectedPayload{
		TenantID: "tenant-1", ProductID: "product-1", DeviceID: "device-1",
		TenantSlug: "tenant", ProductKey: "product", DeviceSlug: "device",
		Protocol: "mqtt", Reason: "keepalive_timeout",
	}))

	if result.Decision != messaging.DecisionAck {
		t.Fatalf("decision = %q, want ack (err=%v)", result.Decision, result.Err)
	}
	if len(s.offline) != 1 || s.offline[0].DeviceID != "device-1" {
		t.Fatalf("offline calls = %+v", s.offline)
	}
	if len(s.online) != 0 {
		t.Fatal("MarkOnline should not be called")
	}
	if len(pub.published) != 1 {
		t.Fatalf("publish count = %d, want 1", len(pub.published))
	}
	changed, ok := pub.published[0].Payload.(event.ConnectionChangedPayload)
	if !ok {
		t.Fatalf("payload type = %T", pub.published[0].Payload)
	}
	if changed.Status != event.ConnectionStatusOffline {
		t.Fatalf("status = %q, want offline", changed.Status)
	}
	if changed.Reason != "keepalive_timeout" {
		t.Fatalf("reason = %q, want keepalive_timeout", changed.Reason)
	}
}

func TestDeviceDisconnectedHandlerAcksIgnoredStoreEvent(t *testing.T) {
	s := &fakeConnectionStore{offlineIgnored: true}
	pub := &fakePublisher{}
	h, _ := NewDeviceDisconnectedHandler(s, pub, nil)

	result := h.Handle(context.Background(), newDisconnectedEnvelope(t, event.DisconnectedPayload{
		TenantID: "tenant-1", ProductID: "product-1", DeviceID: "device-1",
		TenantSlug: "tenant", ProductKey: "product", DeviceSlug: "device",
		Protocol: "mqtt", Reason: "discarded",
	}))

	if result.Decision != messaging.DecisionAck {
		t.Fatalf("decision = %q, want ack", result.Decision)
	}
	if len(s.offline) != 1 {
		t.Fatal("store should still receive the disconnection event")
	}
	if len(pub.published) != 0 {
		t.Fatal("publisher should not be called when store ignores a stale event")
	}
}

func TestDeviceConnectedHandlerDropsOnBadOccurredAt(t *testing.T) {
	s := &fakeConnectionStore{}
	pub := &fakePublisher{}
	h, _ := NewDeviceConnectedHandler(s, pub, nil)

	env := newConnectedEnvelope(t, event.ConnectedPayload{
		TenantID: "tenant-1", ProductID: "product-1", DeviceID: "device-1",
		TenantSlug: "tenant", ProductKey: "product", DeviceSlug: "device", Protocol: "mqtt",
	})
	env.OccurredAt = "not-a-date"

	result := h.Handle(context.Background(), env)
	if result.Decision != messaging.DecisionDrop {
		t.Fatalf("decision = %q, want drop", result.Decision)
	}
	if len(s.online) != 0 {
		t.Fatal("store should not be called with bad occurred_at")
	}
	if len(pub.published) != 0 {
		t.Fatal("publisher should not be called with bad occurred_at")
	}
}
