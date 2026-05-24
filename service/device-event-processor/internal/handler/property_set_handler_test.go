package handler

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
	"github.com/ldchengyi/linkflow-v2/pkg/public/messaging"
)

type fakePropertySetWriter struct {
	saved   []event.PropertySetAcknowledgedPayload
	saveErr error
}

func (w *fakePropertySetWriter) SavePropertySetAcknowledgement(ctx context.Context, env *event.Envelope, payload event.PropertySetAcknowledgedPayload) error {
	w.saved = append(w.saved, payload)
	return w.saveErr
}

func newPropertySetEnvelope(t *testing.T, payload event.PropertySetAcknowledgedPayload) *event.Envelope {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return &event.Envelope{
		EventID:      "evt-1",
		EventType:    event.DevicePropertySetAcknowledged.Type,
		EventVersion: event.DevicePropertySetAcknowledged.Version,
		OccurredAt:   "2026-05-05T10:00:00Z",
		TenantID:     "tenant-1",
		Payload:      json.RawMessage(raw),
	}
}

func samplePropertySetPayload() event.PropertySetAcknowledgedPayload {
	return event.PropertySetAcknowledgedPayload{
		CommandID:  "018f56d3-7cb7-7f1a-9b41-3f3a63fd3db9",
		TenantID:   "tenant-1",
		ProductID:  "product-1",
		DeviceID:   "device-1",
		TenantSlug: "tenant",
		ProductKey: "product",
		DeviceSlug: "device",
		Protocol:   "mqtt",
		Success:    true,
		Code:       "ok",
		Message:    "applied",
		Properties: map[string]any{"led1": true},
	}
}

func TestPropertySetHandleSavesAcknowledgement(t *testing.T) {
	writer := &fakePropertySetWriter{}
	h, err := NewPropertySetHandler(writer, nil)
	if err != nil {
		t.Fatalf("NewPropertySetHandler() error = %v", err)
	}

	result := h.Handle(context.Background(), newPropertySetEnvelope(t, samplePropertySetPayload()))

	if result.Decision != messaging.DecisionAck {
		t.Fatalf("decision = %q, want ack (err=%v)", result.Decision, result.Err)
	}
	if len(writer.saved) != 1 {
		t.Fatalf("writer save count = %d, want 1", len(writer.saved))
	}
	if writer.saved[0].CommandID != "018f56d3-7cb7-7f1a-9b41-3f3a63fd3db9" {
		t.Fatalf("saved command_id = %q, want command id", writer.saved[0].CommandID)
	}
}

func TestPropertySetHandleRetriesOnWriterError(t *testing.T) {
	writer := &fakePropertySetWriter{saveErr: errors.New("db down")}
	h, _ := NewPropertySetHandler(writer, nil)

	result := h.Handle(context.Background(), newPropertySetEnvelope(t, samplePropertySetPayload()))

	if result.Decision != messaging.DecisionRetry {
		t.Fatalf("decision = %q, want retry", result.Decision)
	}
	if len(writer.saved) != 1 {
		t.Fatal("writer should have been called before retry decision")
	}
}
