package handler

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
	"github.com/ldchengyi/linkflow-v2/pkg/public/messaging"
)

type fakePropertySetRequestWriter struct {
	saved   []event.PropertySetRequestedPayload
	saveErr error
}

func (w *fakePropertySetRequestWriter) SavePropertySetRequest(ctx context.Context, env *event.Envelope, payload event.PropertySetRequestedPayload) error {
	w.saved = append(w.saved, payload)
	return w.saveErr
}

type fakePropertySetRequestPublisher struct {
	published []event.PropertySetRequestedPayload
	err       error
}

func (p *fakePropertySetRequestPublisher) PublishPropertySet(ctx context.Context, payload event.PropertySetRequestedPayload) error {
	p.published = append(p.published, payload)
	return p.err
}

func newPropertySetRequestedEnvelope(t *testing.T, payload event.PropertySetRequestedPayload) *event.Envelope {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return &event.Envelope{
		EventID:      "evt-1",
		EventType:    event.DevicePropertySetRequested.Type,
		EventVersion: event.DevicePropertySetRequested.Version,
		OccurredAt:   "2026-05-05T10:00:00Z",
		Producer:     "backend",
		TenantID:     "tenant-1",
		Payload:      json.RawMessage(raw),
	}
}

func samplePropertySetRequestedPayload() event.PropertySetRequestedPayload {
	return event.PropertySetRequestedPayload{
		CommandID:   "018f56d3-7cb7-7f1a-9b41-3f3a63fd3db9",
		TenantID:    "tenant-1",
		ProductID:   "product-1",
		DeviceID:    "device-1",
		TenantSlug:  "tenant",
		ProductKey:  "product",
		DeviceSlug:  "device",
		Protocol:    "mqtt",
		Topic:       "lf/v1/tenant/product/device/property/down/set",
		Properties:  map[string]any{"led1": true},
		RequestedBy: "user-1",
	}
}

func TestPropertySetRequestedHandleRecordsThenPublishes(t *testing.T) {
	writer := &fakePropertySetRequestWriter{}
	publisher := &fakePropertySetRequestPublisher{}
	h, err := NewPropertySetRequestedHandler(writer, publisher, nil)
	if err != nil {
		t.Fatalf("NewPropertySetRequestedHandler() error = %v", err)
	}

	result := h.Handle(context.Background(), newPropertySetRequestedEnvelope(t, samplePropertySetRequestedPayload()))

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

func TestPropertySetRequestedHandleRetriesOnSaveErrorBeforePublish(t *testing.T) {
	writer := &fakePropertySetRequestWriter{saveErr: errors.New("db down")}
	publisher := &fakePropertySetRequestPublisher{}
	h, _ := NewPropertySetRequestedHandler(writer, publisher, nil)

	result := h.Handle(context.Background(), newPropertySetRequestedEnvelope(t, samplePropertySetRequestedPayload()))

	if result.Decision != messaging.DecisionRetry {
		t.Fatalf("decision = %q, want retry", result.Decision)
	}
	if len(publisher.published) != 0 {
		t.Fatal("publisher should not be called when dispatch record cannot be saved")
	}
}

func TestPropertySetRequestedHandleRetriesOnPublishError(t *testing.T) {
	writer := &fakePropertySetRequestWriter{}
	publisher := &fakePropertySetRequestPublisher{err: errors.New("emqx down")}
	h, _ := NewPropertySetRequestedHandler(writer, publisher, nil)

	result := h.Handle(context.Background(), newPropertySetRequestedEnvelope(t, samplePropertySetRequestedPayload()))

	if result.Decision != messaging.DecisionRetry {
		t.Fatalf("decision = %q, want retry", result.Decision)
	}
	if len(writer.saved) != 1 {
		t.Fatalf("writer save count = %d, want 1 before publish retry", len(writer.saved))
	}
}
