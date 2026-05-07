package processor

import (
	"context"
	"errors"
	"fmt"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event/validation"
	"github.com/ldchengyi/linkflow-v2/pkg/public/messaging"
)

type Message struct {
	Key     []byte
	Value   []byte
	Headers map[string]string
}

type Result struct {
	Decision messaging.Decision
	EventID  string
	TenantID string
	DeviceID string
	Err      error
}

type Processor interface {
	Process(ctx context.Context, msg Message) Result
}

type EventHandler interface {
	Handle(ctx context.Context, env *event.Envelope) Result
}

type EventProcessor struct {
	validator *validation.Validator
	handlers  map[string]EventHandler
}

func NewEventProcessor(validator *validation.Validator) (*EventProcessor, error) {
	if validator == nil {
		return nil, errors.New("event validator is nil")
	}

	return &EventProcessor{
		validator: validator,
		handlers:  make(map[string]EventHandler),
	}, nil
}

func (ep *EventProcessor) Register(spec event.Spec, h EventHandler) error {
	if h == nil {
		return fmt.Errorf("handler for event %q is nil", spec.Key())
	}

	if _, exists := ep.handlers[spec.Key()]; exists {
		return fmt.Errorf("handler for event %q already registered", spec.Key())
	}

	ep.handlers[spec.Key()] = h
	return nil
}

func (ep *EventProcessor) Process(ctx context.Context, msg Message) Result {
	if err := ctx.Err(); err != nil {
		return Result{Decision: messaging.DecisionRetry, Err: err}
	}

	env, err := ep.validator.ValidateEvent(msg.Value)
	if err != nil {
		return Result{Decision: messaging.DecisionDrop, Err: err}
	}

	key := event.PayloadSchemaKey(env.EventType, env.EventVersion)
	handler, ok := ep.handlers[key]
	if !ok {
		return Result{
			Decision: messaging.DecisionDrop,
			EventID:  env.EventID,
			TenantID: env.TenantID,
			Err:      fmt.Errorf("unsupported event %s", key),
		}
	}

	return handler.Handle(ctx, env)
}
