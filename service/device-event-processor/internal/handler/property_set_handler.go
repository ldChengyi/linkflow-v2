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

type PropertySetWriter interface {
	SavePropertySetAcknowledgement(ctx context.Context, env *event.Envelope, payload event.PropertySetAcknowledgedPayload) error
}

type PropertySetHandler struct {
	writer PropertySetWriter
	log    *slog.Logger
}

func NewPropertySetHandler(writer PropertySetWriter, log *slog.Logger) (*PropertySetHandler, error) {
	if writer == nil {
		return nil, fmt.Errorf("property set writer is nil")
	}
	if log == nil {
		log = slog.Default()
	}
	return &PropertySetHandler{
		writer: writer,
		log:    log,
	}, nil
}

func (h *PropertySetHandler) Handle(ctx context.Context, env *event.Envelope) processor.Result {
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

	var payload event.PropertySetAcknowledgedPayload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		result.Decision = messaging.DecisionDrop
		result.Err = fmt.Errorf("decode property set acknowledged payload: %w", err)
		return result
	}

	result.ProductKey = payload.ProductKey
	result.DeviceSlug = payload.DeviceSlug

	if payload.Properties == nil {
		payload.Properties = map[string]any{}
	}
	if payload.Raw == nil {
		payload.Raw = map[string]any{}
	}

	if err := h.writer.SavePropertySetAcknowledgement(ctx, env, payload); err != nil {
		result.Decision = messaging.DecisionRetry
		result.Err = fmt.Errorf("save property set acknowledgement: %w", err)
		return result
	}

	result.Decision = messaging.DecisionAck
	return result
}
