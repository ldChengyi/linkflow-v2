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

type fakeEventWriter struct {
	saved   []event.EventReportedPayload
	saveErr error
}

func (w *fakeEventWriter) SaveEventReport(ctx context.Context, env *event.Envelope, payload event.EventReportedPayload) error {
	w.saved = append(w.saved, payload)
	return w.saveErr
}

type fakeEventValidator struct {
	result validator.EventResult
	err    error
	calls  int
}

func (v *fakeEventValidator) Validate(ctx context.Context, in validator.EventInput) (validator.EventResult, error) {
	v.calls++
	return v.result, v.err
}

type fakeEventPublisher struct {
	published []publisher.PublishInput
	err       error
}

func (p *fakeEventPublisher) Publish(ctx context.Context, in publisher.PublishInput) error {
	p.published = append(p.published, in)
	return p.err
}

func newEventEnvelope(t *testing.T, payload event.EventReportedPayload) *event.Envelope {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return &event.Envelope{
		EventID:      "evt-1",
		EventType:    event.DeviceEventReported.Type,
		EventVersion: event.DeviceEventReported.Version,
		OccurredAt:   "2026-05-05T10:00:00Z",
		TenantID:     "tenant-1",
		Payload:      json.RawMessage(raw),
	}
}

func sampleEventPayload() event.EventReportedPayload {
	return event.EventReportedPayload{
		TenantID:   "tenant-1",
		ProductID:  "product-1",
		DeviceID:   "device-1",
		TenantSlug: "tenant",
		ProductKey: "product",
		DeviceSlug: "device",
		Protocol:   "mqtt",
		EventName:  "temperature_alarm",
		Params: map[string]any{
			"temperature": 85.2,
			"debug_raw":   "ignored",
		},
	}
}

func TestEventReportHandleSavesAcceptedParamsOnSuccess(t *testing.T) {
	writer := &fakeEventWriter{}
	pub := &fakeEventPublisher{}
	v := &fakeEventValidator{
		result: validator.EventResult{
			Accepted: map[string]any{"temperature": 85.2},
			Dropped:  []string{"debug_raw"},
		},
	}
	h, err := NewEventReportHandler(writer, v, pub, nil)
	if err != nil {
		t.Fatalf("NewEventReportHandler() error = %v", err)
	}

	result := h.Handle(context.Background(), newEventEnvelope(t, sampleEventPayload()))

	if result.Decision != messaging.DecisionAck {
		t.Fatalf("decision = %q, want ack (err=%v)", result.Decision, result.Err)
	}
	if len(writer.saved) != 1 {
		t.Fatalf("writer save count = %d, want 1", len(writer.saved))
	}
	saved := writer.saved[0]
	if len(saved.Params) != 1 {
		t.Fatalf("saved params = %v, want only temperature", saved.Params)
	}
	if _, exists := saved.Params["debug_raw"]; exists {
		t.Fatal("debug_raw should not be persisted")
	}
	if len(pub.published) != 1 {
		t.Fatalf("published count = %d, want 1", len(pub.published))
	}
	publishedIn := pub.published[0]
	if publishedIn.EventType != event.TypeDeviceEventReceived {
		t.Fatalf("published event_type = %q, want %q", publishedIn.EventType, event.TypeDeviceEventReceived)
	}
	received, ok := publishedIn.Payload.(event.EventReceivedPayload)
	if !ok {
		t.Fatalf("published payload = %T, want EventReceivedPayload", publishedIn.Payload)
	}
	if received.EventName != "temperature_alarm" || len(received.Params) != 1 {
		t.Fatalf("published payload = %+v, want accepted event params", received)
	}
}

func TestEventReportHandleDropsOnEventNotFound(t *testing.T) {
	writer := &fakeEventWriter{}
	pub := &fakeEventPublisher{}
	v := &fakeEventValidator{err: validator.ErrEventNotFound}
	h, _ := NewEventReportHandler(writer, v, pub, nil)

	result := h.Handle(context.Background(), newEventEnvelope(t, sampleEventPayload()))

	if result.Decision != messaging.DecisionDrop {
		t.Fatalf("decision = %q, want drop", result.Decision)
	}
	if !errors.Is(result.Err, validator.ErrEventNotFound) {
		t.Fatalf("err = %v, want ErrEventNotFound", result.Err)
	}
	if len(writer.saved) != 0 {
		t.Fatal("writer should not be called when event is missing")
	}
	if len(pub.published) != 0 {
		t.Fatal("publisher should not be called when event is missing")
	}
}

func TestEventReportHandleDropsOnInvalidParam(t *testing.T) {
	writer := &fakeEventWriter{}
	pub := &fakeEventPublisher{}
	v := &fakeEventValidator{err: validator.ErrInvalidEventValue}
	h, _ := NewEventReportHandler(writer, v, pub, nil)

	result := h.Handle(context.Background(), newEventEnvelope(t, sampleEventPayload()))

	if result.Decision != messaging.DecisionDrop {
		t.Fatalf("decision = %q, want drop", result.Decision)
	}
	if len(writer.saved) != 0 {
		t.Fatal("writer should not be called when validation fails")
	}
	if len(pub.published) != 0 {
		t.Fatal("publisher should not be called when validation fails")
	}
}

func TestEventReportHandleRetriesOnUnexpectedValidatorError(t *testing.T) {
	writer := &fakeEventWriter{}
	pub := &fakeEventPublisher{}
	v := &fakeEventValidator{err: errors.New("db down")}
	h, _ := NewEventReportHandler(writer, v, pub, nil)

	result := h.Handle(context.Background(), newEventEnvelope(t, sampleEventPayload()))

	if result.Decision != messaging.DecisionRetry {
		t.Fatalf("decision = %q, want retry", result.Decision)
	}
	if len(writer.saved) != 0 {
		t.Fatal("writer should not be called on validator failure")
	}
}

func TestEventReportHandleRetriesOnWriterError(t *testing.T) {
	writer := &fakeEventWriter{saveErr: errors.New("db down")}
	pub := &fakeEventPublisher{}
	v := &fakeEventValidator{
		result: validator.EventResult{Accepted: map[string]any{"temperature": 85.2}},
	}
	h, _ := NewEventReportHandler(writer, v, pub, nil)

	result := h.Handle(context.Background(), newEventEnvelope(t, sampleEventPayload()))

	if result.Decision != messaging.DecisionRetry {
		t.Fatalf("decision = %q, want retry", result.Decision)
	}
	if len(writer.saved) != 1 {
		t.Fatal("writer should have been called before retry decision")
	}
	if len(pub.published) != 0 {
		t.Fatal("publisher should not be called when save fails")
	}
}

func TestEventReportHandleRetriesOnPublisherError(t *testing.T) {
	writer := &fakeEventWriter{}
	pub := &fakeEventPublisher{err: errors.New("kafka down")}
	v := &fakeEventValidator{
		result: validator.EventResult{Accepted: map[string]any{"temperature": 85.2}},
	}
	h, _ := NewEventReportHandler(writer, v, pub, nil)

	result := h.Handle(context.Background(), newEventEnvelope(t, sampleEventPayload()))

	if result.Decision != messaging.DecisionRetry {
		t.Fatalf("decision = %q, want retry", result.Decision)
	}
	if len(writer.saved) != 1 {
		t.Fatal("writer should have been called before publisher retry decision")
	}
	if len(pub.published) != 1 {
		t.Fatal("publisher should have been called")
	}
}
