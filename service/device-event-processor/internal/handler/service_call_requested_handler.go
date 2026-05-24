package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
	"github.com/ldchengyi/linkflow-v2/pkg/public/messaging"
	"github.com/ldchengyi/linkflow-v2/service/device-event-processor/internal/processor"
)

type ServiceCallRequestWriter interface {
	SaveServiceCallRequest(ctx context.Context, env *event.Envelope, payload event.ServiceCallRequestedPayload) error
}

type ServiceCallRequestPublisher interface {
	PublishServiceCall(ctx context.Context, payload event.ServiceCallRequestedPayload) error
}

type ServiceCallRequestedHandler struct {
	writer    ServiceCallRequestWriter
	publisher ServiceCallRequestPublisher
	log       *slog.Logger
}

func NewServiceCallRequestedHandler(
	writer ServiceCallRequestWriter,
	publisher ServiceCallRequestPublisher,
	log *slog.Logger,
) (*ServiceCallRequestedHandler, error) {
	if writer == nil {
		return nil, fmt.Errorf("service call request writer is nil")
	}
	if publisher == nil {
		return nil, fmt.Errorf("service call request publisher is nil")
	}
	if log == nil {
		log = slog.Default()
	}
	return &ServiceCallRequestedHandler{
		writer:    writer,
		publisher: publisher,
		log:       log,
	}, nil
}

func (h *ServiceCallRequestedHandler) Handle(ctx context.Context, env *event.Envelope) processor.Result {
	if err := ctx.Err(); err != nil {
		return processor.Result{
			Decision: messaging.DecisionRetry,
			Err:      err,
		}
	}
	if env == nil {
		return processor.Result{
			Decision: messaging.DecisionDrop,
			Err:      fmt.Errorf("event envelope is nil"),
		}
	}

	result := processor.Result{
		EventID:      env.EventID,
		EventType:    env.EventType,
		EventVersion: env.EventVersion,
		TenantID:     env.TenantID,
	}

	var payload event.ServiceCallRequestedPayload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		result.Decision = messaging.DecisionDrop
		result.Err = fmt.Errorf("decode service call requested payload: %w", err)
		return result
	}

	result.ProductKey = payload.ProductKey
	result.DeviceSlug = payload.DeviceSlug

	if err := h.writer.SaveServiceCallRequest(ctx, env, payload); err != nil {
		result.Decision = messaging.DecisionRetry
		result.Err = fmt.Errorf("save service call request: %w", err)
		return result
	}

	if err := h.publisher.PublishServiceCall(ctx, payload); err != nil {
		result.Decision = messaging.DecisionRetry
		result.Err = fmt.Errorf("publish service call to mqtt: %w", err)
		return result
	}

	result.Decision = messaging.DecisionAck
	return result
}
