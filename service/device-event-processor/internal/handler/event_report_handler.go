package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
	"github.com/ldchengyi/linkflow-v2/pkg/public/messaging"
	"github.com/ldchengyi/linkflow-v2/service/device-event-processor/internal/processor"
	"github.com/ldchengyi/linkflow-v2/service/device-event-processor/internal/publisher"
	"github.com/ldchengyi/linkflow-v2/service/device-event-processor/internal/validator"
)

type EventReportWriter interface {
	SaveEventReport(ctx context.Context, env *event.Envelope, payload event.EventReportedPayload) error
}

type EventReportValidator interface {
	Validate(ctx context.Context, in validator.EventInput) (validator.EventResult, error)
}

type EventReportHandler struct {
	writer    EventReportWriter
	validator EventReportValidator
	publisher EventPublisher
	log       *slog.Logger
}

func NewEventReportHandler(writer EventReportWriter, v EventReportValidator, pub EventPublisher, log *slog.Logger) (*EventReportHandler, error) {
	if writer == nil {
		return nil, fmt.Errorf("event report writer is nil")
	}
	if v == nil {
		return nil, fmt.Errorf("event report validator is nil")
	}
	if pub == nil {
		return nil, fmt.Errorf("event publisher is nil")
	}
	if log == nil {
		log = slog.Default()
	}
	return &EventReportHandler{
		writer:    writer,
		validator: v,
		publisher: pub,
		log:       log,
	}, nil
}

func (h *EventReportHandler) Handle(ctx context.Context, env *event.Envelope) processor.Result {
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

	var payload event.EventReportedPayload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		result.Decision = messaging.DecisionDrop
		result.Err = fmt.Errorf("decode event reported payload: %w", err)
		return result
	}

	result.ProductKey = payload.ProductKey
	result.DeviceSlug = payload.DeviceSlug

	validationResult, err := h.validator.Validate(ctx, validator.EventInput{
		TenantID:  payload.TenantID,
		ProductID: payload.ProductID,
		EventName: payload.EventName,
		Params:    payload.Params,
	})
	if err != nil {
		switch {
		case errors.Is(err, validator.ErrThingsModelNotFound), errors.Is(err, validator.ErrEventNotFound):
			h.log.Warn(
				"thingsmodel event not found, dropping event report",
				"event_id", env.EventID,
				"tenant_id", payload.TenantID,
				"product_id", payload.ProductID,
				"device_id", payload.DeviceID,
				"event_name", payload.EventName,
			)
			result.Decision = messaging.DecisionDrop
			result.Err = err
			return result
		case errors.Is(err, validator.ErrInvalidEventValue), errors.Is(err, validator.ErrNoAcceptedEventParams):
			h.log.Info(
				"event report rejected by thingsmodel validation",
				"event_id", env.EventID,
				"product_id", payload.ProductID,
				"device_id", payload.DeviceID,
				"event_name", payload.EventName,
				"err", err,
			)
			result.Decision = messaging.DecisionDrop
			result.Err = err
			return result
		default:
			result.Decision = messaging.DecisionRetry
			result.Err = fmt.Errorf("validate event report: %w", err)
			return result
		}
	}

	if len(validationResult.Dropped) > 0 {
		h.log.Info(
			"event report dropped unknown params",
			"event_id", env.EventID,
			"product_id", payload.ProductID,
			"device_id", payload.DeviceID,
			"event_name", payload.EventName,
			"dropped", validationResult.Dropped,
		)
	}

	payload.Params = validationResult.Accepted
	if payload.Params == nil {
		payload.Params = map[string]any{}
	}

	if err := h.writer.SaveEventReport(ctx, env, payload); err != nil {
		result.Decision = messaging.DecisionRetry
		result.Err = fmt.Errorf("save event report: %w", err)
		return result
	}

	occurredAt, _ := time.Parse(time.RFC3339Nano, env.OccurredAt)
	if err := h.publisher.Publish(ctx, publisher.PublishInput{
		EventType:    event.TypeDeviceEventReceived,
		EventVersion: event.VersionDeviceEventReceived,
		TenantID:     env.TenantID,
		CausationID:  env.EventID,
		OccurredAt:   occurredAt,
		Payload: event.EventReceivedPayload{
			TenantID:   payload.TenantID,
			ProductID:  payload.ProductID,
			DeviceID:   payload.DeviceID,
			TenantSlug: payload.TenantSlug,
			ProductKey: payload.ProductKey,
			DeviceSlug: payload.DeviceSlug,
			EventName:  payload.EventName,
			Params:     payload.Params,
		},
	}); err != nil {
		result.Decision = messaging.DecisionRetry
		result.Err = fmt.Errorf("publish event received: %w", err)
		return result
	}

	result.Decision = messaging.DecisionAck
	return result
}
