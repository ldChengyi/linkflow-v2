package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
)

// MessageSink is the broker-neutral byte sink used by EventPublisher. Mirrors
// pkg/public/messaging/kafka.Sink without coupling the publisher to that
// concrete type, which keeps tests easy to wire with fakes.
type MessageSink interface {
	Send(ctx context.Context, topic string, key []byte, value []byte, headers map[string]string) error
}

// PublishInput carries the fields the EventPublisher needs to wrap a payload
// in a v1 envelope and route it to the topic declared by the event spec.
type PublishInput struct {
	EventType    string
	EventVersion int
	TenantID     string
	Payload      any
	OccurredAt   time.Time
	CausationID  string
}

// EventPublisher serializes payloads into envelopes and forwards them to the
// downstream sink. It is the post-validation fan-out point in the processor.
type EventPublisher struct {
	sink     MessageSink
	producer string
}

func NewEventPublisher(sink MessageSink, producer string) (*EventPublisher, error) {
	if sink == nil {
		return nil, fmt.Errorf("message sink is nil")
	}
	if producer == "" {
		return nil, fmt.Errorf("producer is required")
	}
	return &EventPublisher{sink: sink, producer: producer}, nil
}

func (p *EventPublisher) Publish(ctx context.Context, in PublishInput) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	spec, ok := event.Lookup(in.EventType, in.EventVersion)
	if !ok {
		return fmt.Errorf("publish: unknown event %s v%d", in.EventType, in.EventVersion)
	}
	if in.TenantID == "" {
		return fmt.Errorf("publish %s: tenant_id is required", spec.Key())
	}

	payloadJSON, err := json.Marshal(in.Payload)
	if err != nil {
		return fmt.Errorf("publish %s: marshal payload: %w", spec.Key(), err)
	}

	occurredAt := in.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}

	id, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("publish %s: new event id: %w", spec.Key(), err)
	}

	envelope := event.Envelope{
		EventID:      id.String(),
		EventType:    spec.Type,
		EventVersion: spec.Version,
		OccurredAt:   occurredAt.UTC().Format(time.RFC3339Nano),
		Producer:     p.producer,
		TenantID:     in.TenantID,
		CausationID:  in.CausationID,
		Payload:      payloadJSON,
	}

	envelopeJSON, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("publish %s: marshal envelope: %w", spec.Key(), err)
	}

	headers := map[string]string{
		"event_type":    envelope.EventType,
		"event_version": strconv.Itoa(envelope.EventVersion),
		"producer":      envelope.Producer,
		"tenant_id":     envelope.TenantID,
	}

	if err := p.sink.Send(ctx, spec.Topic, []byte(envelope.EventID), envelopeJSON, headers); err != nil {
		return fmt.Errorf("publish %s: send: %w", spec.Key(), err)
	}
	return nil
}
