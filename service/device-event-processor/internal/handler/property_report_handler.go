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

type PropertyReportWriter interface {
	SavePropertyReport(ctx context.Context, env *event.Envelope, payload event.PropertyReportedPayload) error
}

type PropertyReportValidator interface {
	Validate(ctx context.Context, in validator.Input) (validator.Result, error)
}

type EventPublisher interface {
	Publish(ctx context.Context, in publisher.PublishInput) error
}

type PropertyReportHandler struct {
	writer    PropertyReportWriter
	validator PropertyReportValidator
	publisher EventPublisher
	log       *slog.Logger
}

func NewPropertyReportHandler(writer PropertyReportWriter, v PropertyReportValidator, pub EventPublisher, log *slog.Logger) (*PropertyReportHandler, error) {
	if writer == nil {
		return nil, fmt.Errorf("property report writer is nil")
	}
	if v == nil {
		return nil, fmt.Errorf("property report validator is nil")
	}
	if pub == nil {
		return nil, fmt.Errorf("event publisher is nil")
	}
	if log == nil {
		log = slog.Default()
	}

	return &PropertyReportHandler{
		writer:    writer,
		validator: v,
		publisher: pub,
		log:       log,
	}, nil
}

func (h *PropertyReportHandler) Handle(ctx context.Context, env *event.Envelope) processor.Result {
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

	var payload event.PropertyReportedPayload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		result.Decision = messaging.DecisionDrop
		result.Err = fmt.Errorf("decode property reported payload: %w", err)
		return result
	}

	result.ProductKey = payload.ProductKey
	result.DeviceSlug = payload.DeviceSlug

	validationResult, err := h.validator.Validate(ctx, validator.Input{
		TenantID:   payload.TenantID,
		ProductID:  payload.ProductID,
		Properties: payload.Properties,
	})
	if err != nil {
		switch {
		case errors.Is(err, validator.ErrThingsModelNotFound):
			h.log.Warn(
				"thingsmodel not found, dropping property report",
				"event_id", env.EventID,
				"tenant_id", payload.TenantID,
				"product_id", payload.ProductID,
				"device_id", payload.DeviceID,
			)
			result.Decision = messaging.DecisionDrop
			result.Err = err
			return result
		case errors.Is(err, validator.ErrInvalidPropertyValue), errors.Is(err, validator.ErrNoAcceptedProperties):
			h.log.Info(
				"property report rejected by thingsmodel validation",
				"event_id", env.EventID,
				"product_id", payload.ProductID,
				"device_id", payload.DeviceID,
				"err", err,
			)
			result.Decision = messaging.DecisionDrop
			result.Err = err
			return result
		default:
			result.Decision = messaging.DecisionRetry
			result.Err = fmt.Errorf("validate property report: %w", err)
			return result
		}
	}

	if len(validationResult.Dropped) > 0 {
		h.log.Info(
			"property report dropped unknown fields",
			"event_id", env.EventID,
			"product_id", payload.ProductID,
			"device_id", payload.DeviceID,
			"dropped", validationResult.Dropped,
		)
	}

	payload.Properties = validationResult.Accepted

	if err := h.writer.SavePropertyReport(ctx, env, payload); err != nil {
		result.Decision = messaging.DecisionRetry
		result.Err = fmt.Errorf("save property report: %w", err)
		return result
	}

	occurredAt, _ := time.Parse(time.RFC3339Nano, env.OccurredAt)
	if err := h.publisher.Publish(ctx, publisher.PublishInput{
		EventType:    event.TypeDevicePropertyChanged,
		EventVersion: event.VersionDevicePropertyChanged,
		TenantID:     env.TenantID,
		CausationID:  env.EventID,
		OccurredAt:   occurredAt,
		Payload: event.PropertyChangedPayload{
			TenantID:   payload.TenantID,
			ProductID:  payload.ProductID,
			DeviceID:   payload.DeviceID,
			TenantSlug: payload.TenantSlug,
			ProductKey: payload.ProductKey,
			DeviceSlug: payload.DeviceSlug,
			Properties: validationResult.Accepted,
		},
	}); err != nil {
		result.Decision = messaging.DecisionRetry
		result.Err = fmt.Errorf("publish property changed: %w", err)
		return result
	}

	result.Decision = messaging.DecisionAck
	return result
}
