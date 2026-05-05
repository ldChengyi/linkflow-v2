package processor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
	"github.com/ldchengyi/linkflow-v2/service/device-event-processor/internal/schema"
)

type Message struct {
	Key     []byte
	Value   []byte
	Headers map[string]string
}

type Action string

const (
	ActionAck   Action = "ack"
	ActionRetry Action = "retry"
	ActionDrop  Action = "drop"
)

type Result struct {
	Action   Action
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

type EventHandlerFunc func(ctx context.Context, env *event.Envelope) Result

func (f EventHandlerFunc) Handle(ctx context.Context, env *event.Envelope) Result {
	return f(ctx, env)
}

type EventProcessor struct {
	registry *schema.Registry
	handlers map[string]EventHandler
}

func NewEventProcessor(registry *schema.Registry) (*EventProcessor, error) {
	if registry == nil {
		return nil, errors.New("schema registry is nil")
	}

	ep := &EventProcessor{
		registry: registry,
		handlers: make(map[string]EventHandler),
	}
	ep.Register(event.DeviceTelemetryReceived.Key(), EventHandlerFunc(ep.processTelemetryReceived))
	ep.Register(event.DevicePropertySetAcknowledged.Key(), EventHandlerFunc(ep.processPropertySetAcknowledged))
	return ep, nil
}

func (ep *EventProcessor) Register(key string, h EventHandler) {
	ep.handlers[key] = h
}

func (ep *EventProcessor) Process(ctx context.Context, msg Message) Result {
	if err := ctx.Err(); err != nil {
		return Result{Action: ActionRetry, Err: err}
	}

	env, err := ep.registry.ValidateEvent(msg.Value)
	if err != nil {
		return Result{Action: ActionDrop, Err: err}
	}

	key := event.PayloadSchemaKey(env.EventType, env.EventVersion)
	handler, ok := ep.handlers[key]
	if !ok {
		return Result{
			Action:   ActionDrop,
			EventID:  env.EventID,
			TenantID: env.TenantID,
			Err:      fmt.Errorf("unsupported event %s", key),
		}
	}
	return handler.Handle(ctx, env)
}

func (p *EventProcessor) processTelemetryReceived(ctx context.Context, env *event.Envelope) Result {
	result := Result{
		EventID:  env.EventID,
		TenantID: env.TenantID,
	}

	if err := ctx.Err(); err != nil {
		result.Action = ActionRetry
		result.Err = err
		return result
	}

	var payload event.TelemetryReceivedPayload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		result.Action = ActionDrop
		result.Err = fmt.Errorf("decode telemetry payload: %w", err)
		return result
	}

	result.Action = ActionAck
	result.DeviceID = payload.DeviceID
	return result
}

func (p *EventProcessor) processPropertySetAcknowledged(ctx context.Context, env *event.Envelope) Result {
	result := Result{
		EventID:  env.EventID,
		TenantID: env.TenantID,
	}

	if err := ctx.Err(); err != nil {
		result.Action = ActionRetry
		result.Err = err
		return result
	}

	var payload event.PropertySetAcknowledgedPayload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		result.Action = ActionDrop
		result.Err = fmt.Errorf("decode property set acknowledged payload: %w", err)
		return result
	}

	result.Action = ActionAck
	result.DeviceID = payload.DeviceID
	return result
}
