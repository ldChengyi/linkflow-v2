package handler

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
	"github.com/ldchengyi/linkflow-v2/pkg/public/messaging"
	"github.com/ldchengyi/linkflow-v2/service/device-event-processor/internal/publisher"
	"github.com/ldchengyi/linkflow-v2/service/device-event-processor/internal/validator"
)

type fakeWriter struct {
	saved   []event.PropertyReportedPayload
	saveErr error
}

func (w *fakeWriter) SavePropertyReport(ctx context.Context, env *event.Envelope, payload event.PropertyReportedPayload) error {
	w.saved = append(w.saved, payload)
	return w.saveErr
}

type fakeValidator struct {
	result validator.Result
	err    error
	calls  int
}

func (v *fakeValidator) Validate(ctx context.Context, in validator.Input) (validator.Result, error) {
	v.calls++
	return v.result, v.err
}

type fakePublisher struct {
	published []publisher.PublishInput
	err       error
}

func (p *fakePublisher) Publish(ctx context.Context, in publisher.PublishInput) error {
	p.published = append(p.published, in)
	return p.err
}

func newEnvelope(t *testing.T, payload event.PropertyReportedPayload) *event.Envelope {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return &event.Envelope{
		EventID:      "evt-1",
		EventType:    event.DevicePropertyReported.Type,
		EventVersion: event.DevicePropertyReported.Version,
		TenantID:     "tenant-1",
		Payload:      json.RawMessage(raw),
	}
}

func samplePayload() event.PropertyReportedPayload {
	return event.PropertyReportedPayload{
		TenantID:   "tenant-1",
		ProductID:  "product-1",
		DeviceID:   "device-1",
		TenantSlug: "tenant",
		ProductKey: "product",
		DeviceSlug: "device",
		Protocol:   "mqtt",
		Properties: map[string]any{
			"temperature": 23.5,
			"debug_raw":   "ignored",
		},
	}
}

func TestHandleSavesAcceptedPropertiesOnSuccess(t *testing.T) {
	writer := &fakeWriter{}
	v := &fakeValidator{
		result: validator.Result{
			Accepted: map[string]any{"temperature": 23.5},
			Dropped:  []string{"debug_raw"},
		},
	}
	pub := &fakePublisher{}
	h, err := NewPropertyReportHandler(writer, v, pub, nil)
	if err != nil {
		t.Fatalf("NewPropertyReportHandler() error = %v", err)
	}

	result := h.Handle(context.Background(), newEnvelope(t, samplePayload()))

	if result.Decision != messaging.DecisionAck {
		t.Fatalf("decision = %q, want ack (err=%v)", result.Decision, result.Err)
	}
	if len(writer.saved) != 1 {
		t.Fatalf("writer save count = %d, want 1", len(writer.saved))
	}
	saved := writer.saved[0]
	if len(saved.Properties) != 1 {
		t.Fatalf("saved properties = %v, want only temperature", saved.Properties)
	}
	if _, exists := saved.Properties["debug_raw"]; exists {
		t.Fatal("debug_raw should not be persisted")
	}
	if len(pub.published) != 1 {
		t.Fatalf("publish count = %d, want 1", len(pub.published))
	}
	publishedIn := pub.published[0]
	if publishedIn.EventType != event.TypeDevicePropertyChanged {
		t.Fatalf("published event_type = %q, want %q", publishedIn.EventType, event.TypeDevicePropertyChanged)
	}
	if publishedIn.CausationID != "evt-1" {
		t.Fatalf("causation_id = %q, want evt-1", publishedIn.CausationID)
	}
	changed, ok := publishedIn.Payload.(event.PropertyChangedPayload)
	if !ok {
		t.Fatalf("payload type = %T", publishedIn.Payload)
	}
	if _, exists := changed.Properties["debug_raw"]; exists {
		t.Fatal("changed event must not carry filtered field")
	}
}

func TestHandleDropsOnThingsModelNotFound(t *testing.T) {
	writer := &fakeWriter{}
	v := &fakeValidator{err: validator.ErrThingsModelNotFound}
	pub := &fakePublisher{}
	h, _ := NewPropertyReportHandler(writer, v, pub, nil)

	result := h.Handle(context.Background(), newEnvelope(t, samplePayload()))

	if result.Decision != messaging.DecisionDrop {
		t.Fatalf("decision = %q, want drop", result.Decision)
	}
	if !errors.Is(result.Err, validator.ErrThingsModelNotFound) {
		t.Fatalf("err = %v, want ErrThingsModelNotFound", result.Err)
	}
	if len(writer.saved) != 0 {
		t.Fatal("writer should not be called when thingsmodel missing")
	}
	if len(pub.published) != 0 {
		t.Fatal("publisher should not be called when validator drops")
	}
}

func TestHandleDropsOnInvalidPropertyValue(t *testing.T) {
	writer := &fakeWriter{}
	v := &fakeValidator{err: validator.ErrInvalidPropertyValue}
	pub := &fakePublisher{}
	h, _ := NewPropertyReportHandler(writer, v, pub, nil)

	result := h.Handle(context.Background(), newEnvelope(t, samplePayload()))

	if result.Decision != messaging.DecisionDrop {
		t.Fatalf("decision = %q, want drop", result.Decision)
	}
	if len(writer.saved) != 0 {
		t.Fatal("writer should not be called when validation fails")
	}
	if len(pub.published) != 0 {
		t.Fatal("publisher should not be called when validator drops")
	}
}

func TestHandleRetriesOnUnexpectedValidatorError(t *testing.T) {
	writer := &fakeWriter{}
	v := &fakeValidator{err: errors.New("db down")}
	pub := &fakePublisher{}
	h, _ := NewPropertyReportHandler(writer, v, pub, nil)

	result := h.Handle(context.Background(), newEnvelope(t, samplePayload()))

	if result.Decision != messaging.DecisionRetry {
		t.Fatalf("decision = %q, want retry", result.Decision)
	}
	if len(writer.saved) != 0 {
		t.Fatal("writer should not be called on validator failure")
	}
}

func TestHandleRetriesOnPublisherError(t *testing.T) {
	writer := &fakeWriter{}
	v := &fakeValidator{
		result: validator.Result{Accepted: map[string]any{"temperature": 23.5}},
	}
	pub := &fakePublisher{err: errors.New("kafka down")}
	h, _ := NewPropertyReportHandler(writer, v, pub, nil)

	result := h.Handle(context.Background(), newEnvelope(t, samplePayload()))

	if result.Decision != messaging.DecisionRetry {
		t.Fatalf("decision = %q, want retry", result.Decision)
	}
	if len(writer.saved) != 1 {
		t.Fatal("writer should still have been called before publish")
	}
}
