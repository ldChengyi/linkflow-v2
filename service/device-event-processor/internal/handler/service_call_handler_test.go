package handler

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
	"github.com/ldchengyi/linkflow-v2/pkg/public/messaging"
	"github.com/ldchengyi/linkflow-v2/service/device-event-processor/internal/validator"
)

type fakeServiceCallWriter struct {
	saved   []event.ServiceCallAcknowledgedPayload
	saveErr error
}

func (w *fakeServiceCallWriter) SaveServiceCallAcknowledgement(ctx context.Context, env *event.Envelope, payload event.ServiceCallAcknowledgedPayload) error {
	w.saved = append(w.saved, payload)
	return w.saveErr
}

type fakeServiceCallValidator struct {
	result validator.ServiceCallResult
	err    error
	calls  int
}

func (v *fakeServiceCallValidator) Validate(ctx context.Context, in validator.ServiceCallInput) (validator.ServiceCallResult, error) {
	v.calls++
	return v.result, v.err
}

func newServiceCallEnvelope(t *testing.T, payload event.ServiceCallAcknowledgedPayload) *event.Envelope {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return &event.Envelope{
		EventID:      "evt-1",
		EventType:    event.DeviceServiceCallAcknowledged.Type,
		EventVersion: event.DeviceServiceCallAcknowledged.Version,
		OccurredAt:   "2026-05-05T10:00:00Z",
		TenantID:     "tenant-1",
		Payload:      json.RawMessage(raw),
	}
}

func sampleServiceCallPayload() event.ServiceCallAcknowledgedPayload {
	return event.ServiceCallAcknowledgedPayload{
		CommandID:   "018f56d3-7cb7-7f1a-9b41-3f3a63fd3db9",
		TenantID:    "tenant-1",
		ProductID:   "product-1",
		DeviceID:    "device-1",
		TenantSlug:  "tenant",
		ProductKey:  "product",
		DeviceSlug:  "device",
		Protocol:    "mqtt",
		ServiceName: "reboot",
		Success:     true,
		Code:        "ok",
		Message:     "done",
		Output: map[string]any{
			"accepted":  true,
			"debug_raw": "ignored",
		},
	}
}

func TestServiceCallHandleSavesAcceptedOutputOnSuccess(t *testing.T) {
	writer := &fakeServiceCallWriter{}
	v := &fakeServiceCallValidator{
		result: validator.ServiceCallResult{
			Accepted: map[string]any{"accepted": true},
			Dropped:  []string{"debug_raw"},
		},
	}
	h, err := NewServiceCallHandler(writer, v, nil)
	if err != nil {
		t.Fatalf("NewServiceCallHandler() error = %v", err)
	}

	result := h.Handle(context.Background(), newServiceCallEnvelope(t, sampleServiceCallPayload()))

	if result.Decision != messaging.DecisionAck {
		t.Fatalf("decision = %q, want ack (err=%v)", result.Decision, result.Err)
	}
	if len(writer.saved) != 1 {
		t.Fatalf("writer save count = %d, want 1", len(writer.saved))
	}
	saved := writer.saved[0]
	if len(saved.Output) != 1 {
		t.Fatalf("saved output = %v, want only accepted", saved.Output)
	}
	if saved.CommandID != "018f56d3-7cb7-7f1a-9b41-3f3a63fd3db9" {
		t.Fatalf("saved command_id = %q, want command id from payload", saved.CommandID)
	}
	if _, exists := saved.Output["debug_raw"]; exists {
		t.Fatal("debug_raw should not be persisted")
	}
}

func TestServiceCallHandleDropsOnServiceNotFound(t *testing.T) {
	writer := &fakeServiceCallWriter{}
	v := &fakeServiceCallValidator{err: validator.ErrServiceNotFound}
	h, _ := NewServiceCallHandler(writer, v, nil)

	result := h.Handle(context.Background(), newServiceCallEnvelope(t, sampleServiceCallPayload()))

	if result.Decision != messaging.DecisionDrop {
		t.Fatalf("decision = %q, want drop", result.Decision)
	}
	if !errors.Is(result.Err, validator.ErrServiceNotFound) {
		t.Fatalf("err = %v, want ErrServiceNotFound", result.Err)
	}
	if len(writer.saved) != 0 {
		t.Fatal("writer should not be called when service is missing")
	}
}

func TestServiceCallHandleDropsOnInvalidOutput(t *testing.T) {
	writer := &fakeServiceCallWriter{}
	v := &fakeServiceCallValidator{err: validator.ErrInvalidServiceOutput}
	h, _ := NewServiceCallHandler(writer, v, nil)

	result := h.Handle(context.Background(), newServiceCallEnvelope(t, sampleServiceCallPayload()))

	if result.Decision != messaging.DecisionDrop {
		t.Fatalf("decision = %q, want drop", result.Decision)
	}
	if len(writer.saved) != 0 {
		t.Fatal("writer should not be called when validation fails")
	}
}

func TestServiceCallHandleRetriesOnUnexpectedValidatorError(t *testing.T) {
	writer := &fakeServiceCallWriter{}
	v := &fakeServiceCallValidator{err: errors.New("db down")}
	h, _ := NewServiceCallHandler(writer, v, nil)

	result := h.Handle(context.Background(), newServiceCallEnvelope(t, sampleServiceCallPayload()))

	if result.Decision != messaging.DecisionRetry {
		t.Fatalf("decision = %q, want retry", result.Decision)
	}
	if len(writer.saved) != 0 {
		t.Fatal("writer should not be called on validator failure")
	}
}

func TestServiceCallHandleRetriesOnWriterError(t *testing.T) {
	writer := &fakeServiceCallWriter{saveErr: errors.New("db down")}
	v := &fakeServiceCallValidator{
		result: validator.ServiceCallResult{Accepted: map[string]any{"accepted": true}},
	}
	h, _ := NewServiceCallHandler(writer, v, nil)

	result := h.Handle(context.Background(), newServiceCallEnvelope(t, sampleServiceCallPayload()))

	if result.Decision != messaging.DecisionRetry {
		t.Fatalf("decision = %q, want retry", result.Decision)
	}
	if len(writer.saved) != 1 {
		t.Fatal("writer should have been called before retry decision")
	}
}
