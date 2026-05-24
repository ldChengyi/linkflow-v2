package handler

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
	"github.com/ldchengyi/linkflow-v2/pkg/public/messaging"
)

type fakeServiceCallRequestWriter struct {
	saved   []event.ServiceCallRequestedPayload
	saveErr error
}

func (w *fakeServiceCallRequestWriter) SaveServiceCallRequest(ctx context.Context, env *event.Envelope, payload event.ServiceCallRequestedPayload) error {
	w.saved = append(w.saved, payload)
	return w.saveErr
}

type fakeServiceCallRequestPublisher struct {
	published []event.ServiceCallRequestedPayload
	err       error
}

func (p *fakeServiceCallRequestPublisher) PublishServiceCall(ctx context.Context, payload event.ServiceCallRequestedPayload) error {
	p.published = append(p.published, payload)
	return p.err
}

func newServiceCallRequestedEnvelope(t *testing.T, payload event.ServiceCallRequestedPayload) *event.Envelope {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return &event.Envelope{
		EventID:      "evt-1",
		EventType:    event.DeviceServiceCallRequested.Type,
		EventVersion: event.DeviceServiceCallRequested.Version,
		OccurredAt:   "2026-05-05T10:00:00Z",
		Producer:     "backend",
		TenantID:     "tenant-1",
		Payload:      json.RawMessage(raw),
	}
}

func sampleServiceCallRequestedPayload() event.ServiceCallRequestedPayload {
	return event.ServiceCallRequestedPayload{
		CommandID:   "018f56d3-7cb7-7f1a-9b41-3f3a63fd3db9",
		TenantID:    "tenant-1",
		ProductID:   "product-1",
		DeviceID:    "device-1",
		TenantSlug:  "tenant",
		ProductKey:  "product",
		DeviceSlug:  "device",
		Protocol:    "mqtt",
		Topic:       "lf/v1/tenant/product/device/service/down/reboot",
		ServiceName: "reboot",
		Input:       map[string]any{"delay": float64(5)},
		RequestedBy: "user-1",
	}
}

func TestServiceCallRequestedHandleRecordsThenPublishes(t *testing.T) {
	writer := &fakeServiceCallRequestWriter{}
	publisher := &fakeServiceCallRequestPublisher{}
	h, err := NewServiceCallRequestedHandler(writer, publisher, nil)
	if err != nil {
		t.Fatalf("NewServiceCallRequestedHandler() error = %v", err)
	}

	result := h.Handle(context.Background(), newServiceCallRequestedEnvelope(t, sampleServiceCallRequestedPayload()))

	if result.Decision != messaging.DecisionAck {
		t.Fatalf("decision = %q, want ack (err=%v)", result.Decision, result.Err)
	}
	if len(writer.saved) != 1 {
		t.Fatalf("writer save count = %d, want 1", len(writer.saved))
	}
	if len(publisher.published) != 1 {
		t.Fatalf("publisher call count = %d, want 1", len(publisher.published))
	}
	if writer.saved[0].CommandID != publisher.published[0].CommandID {
		t.Fatalf("command ids differ: saved=%q published=%q", writer.saved[0].CommandID, publisher.published[0].CommandID)
	}
}

func TestServiceCallRequestedHandleRetriesOnSaveErrorBeforePublish(t *testing.T) {
	writer := &fakeServiceCallRequestWriter{saveErr: errors.New("db down")}
	publisher := &fakeServiceCallRequestPublisher{}
	h, _ := NewServiceCallRequestedHandler(writer, publisher, nil)

	result := h.Handle(context.Background(), newServiceCallRequestedEnvelope(t, sampleServiceCallRequestedPayload()))

	if result.Decision != messaging.DecisionRetry {
		t.Fatalf("decision = %q, want retry", result.Decision)
	}
	if len(publisher.published) != 0 {
		t.Fatal("publisher should not be called when dispatch record cannot be saved")
	}
}

func TestServiceCallRequestedHandleRetriesOnPublishError(t *testing.T) {
	writer := &fakeServiceCallRequestWriter{}
	publisher := &fakeServiceCallRequestPublisher{err: errors.New("emqx down")}
	h, _ := NewServiceCallRequestedHandler(writer, publisher, nil)

	result := h.Handle(context.Background(), newServiceCallRequestedEnvelope(t, sampleServiceCallRequestedPayload()))

	if result.Decision != messaging.DecisionRetry {
		t.Fatalf("decision = %q, want retry", result.Decision)
	}
	if len(writer.saved) != 1 {
		t.Fatalf("writer save count = %d, want 1 before publish retry", len(writer.saved))
	}
}
