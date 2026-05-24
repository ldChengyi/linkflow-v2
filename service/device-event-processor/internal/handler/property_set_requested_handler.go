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

type PropertySetRequestWriter interface {
	SavePropertySetRequest(ctx context.Context, env *event.Envelope, payload event.PropertySetRequestedPayload) error
}

type PropertySetRequestPublisher interface {
	PublishPropertySet(ctx context.Context, payload event.PropertySetRequestedPayload) error
}

type PropertySetRequestedHandler struct {
	writer    PropertySetRequestWriter
	publisher PropertySetRequestPublisher
	log       *slog.Logger
}

func NewPropertySetRequestedHandler(
	writer PropertySetRequestWriter,
	publisher PropertySetRequestPublisher,
	log *slog.Logger,
) (*PropertySetRequestedHandler, error) {
	if writer == nil {
		return nil, fmt.Errorf("property set request writer is nil")
	}
	if publisher == nil {
		return nil, fmt.Errorf("property set request publisher is nil")
	}
	if log == nil {
		log = slog.Default()
	}
	return &PropertySetRequestedHandler{
		writer:    writer,
		publisher: publisher,
		log:       log,
	}, nil
}

func (h *PropertySetRequestedHandler) Handle(ctx context.Context, env *event.Envelope) processor.Result {
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

	var payload event.PropertySetRequestedPayload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		result.Decision = messaging.DecisionDrop
		result.Err = fmt.Errorf("decode property set requested payload: %w", err)
		return result
	}

	result.ProductKey = payload.ProductKey
	result.DeviceSlug = payload.DeviceSlug

	if err := h.writer.SavePropertySetRequest(ctx, env, payload); err != nil {
		result.Decision = messaging.DecisionRetry
		result.Err = fmt.Errorf("save property set request: %w", err)
		return result
	}

	if err := h.publisher.PublishPropertySet(ctx, payload); err != nil {
		result.Decision = messaging.DecisionRetry
		result.Err = fmt.Errorf("publish property set to mqtt: %w", err)
		return result
	}

	result.Decision = messaging.DecisionAck
	return result
}
