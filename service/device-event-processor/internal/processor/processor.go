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

type EventProcessor struct {
	registry *schema.Registry
}

func NewEventProcessor(registry *schema.Registry) (*EventProcessor, error) {
	if registry == nil {
		return nil, errors.New("schema registry is nil")
	}

	return &EventProcessor{
		registry: registry,
	}, nil
}

func (ep *EventProcessor) Process(ctx context.Context, msg Message) Result {
	if err := ctx.Err(); err != nil {
		return Result{Action: ActionRetry, Err: err}
	}

	env, err := ep.registry.ValidateEvent(msg.Value)
	if err != nil {
		return Result{Action: ActionDrop, Err: err}
	}

	base := Result{
		EventID:  env.EventID,
		TenantID: env.TenantID,
	}

	switch event.PayloadSchemaKey(env.EventType, env.EventVersion) {
	case event.DeviceTelemetryReceived.Key():
		return ep.processTelemetryReceived(ctx, env)
	default:
		base.Action = ActionDrop
		base.Err = fmt.Errorf("unsupported event %s v%d", env.EventType, env.EventVersion)
		return base
	}

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
