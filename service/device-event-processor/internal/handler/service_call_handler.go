package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
	"github.com/ldchengyi/linkflow-v2/pkg/public/messaging"
	"github.com/ldchengyi/linkflow-v2/service/device-event-processor/internal/processor"
	"github.com/ldchengyi/linkflow-v2/service/device-event-processor/internal/validator"
)

type ServiceCallWriter interface {
	SaveServiceCallAcknowledgement(ctx context.Context, env *event.Envelope, payload event.ServiceCallAcknowledgedPayload) error
}

type ServiceCallValidator interface {
	Validate(ctx context.Context, in validator.ServiceCallInput) (validator.ServiceCallResult, error)
}

type ServiceCallHandler struct {
	writer    ServiceCallWriter
	validator ServiceCallValidator
	log       *slog.Logger
}

func NewServiceCallHandler(writer ServiceCallWriter, v ServiceCallValidator, log *slog.Logger) (*ServiceCallHandler, error) {
	if writer == nil {
		return nil, fmt.Errorf("service call writer is nil")
	}
	if v == nil {
		return nil, fmt.Errorf("service call validator is nil")
	}
	if log == nil {
		log = slog.Default()
	}
	return &ServiceCallHandler{
		writer:    writer,
		validator: v,
		log:       log,
	}, nil
}

func (h *ServiceCallHandler) Handle(ctx context.Context, env *event.Envelope) processor.Result {
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

	var payload event.ServiceCallAcknowledgedPayload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		result.Decision = messaging.DecisionDrop
		result.Err = fmt.Errorf("decode service call acknowledged payload: %w", err)
		return result
	}

	result.ProductKey = payload.ProductKey
	result.DeviceSlug = payload.DeviceSlug

	validationResult, err := h.validator.Validate(ctx, validator.ServiceCallInput{
		TenantID:    payload.TenantID,
		ProductID:   payload.ProductID,
		ServiceName: payload.ServiceName,
		Output:      payload.Output,
	})
	if err != nil {
		switch {
		case errors.Is(err, validator.ErrThingsModelNotFound), errors.Is(err, validator.ErrServiceNotFound):
			h.log.Warn(
				"thingsmodel service not found, dropping service call acknowledgement",
				"event_id", env.EventID,
				"tenant_id", payload.TenantID,
				"product_id", payload.ProductID,
				"device_id", payload.DeviceID,
				"service_name", payload.ServiceName,
			)
			result.Decision = messaging.DecisionDrop
			result.Err = err
			return result
		case errors.Is(err, validator.ErrInvalidServiceOutput):
			h.log.Info(
				"service call acknowledgement rejected by thingsmodel validation",
				"event_id", env.EventID,
				"product_id", payload.ProductID,
				"device_id", payload.DeviceID,
				"service_name", payload.ServiceName,
				"err", err,
			)
			result.Decision = messaging.DecisionDrop
			result.Err = err
			return result
		default:
			result.Decision = messaging.DecisionRetry
			result.Err = fmt.Errorf("validate service call acknowledgement: %w", err)
			return result
		}
	}

	if len(validationResult.Dropped) > 0 {
		h.log.Info(
			"service call acknowledgement dropped unknown output fields",
			"event_id", env.EventID,
			"product_id", payload.ProductID,
			"device_id", payload.DeviceID,
			"service_name", payload.ServiceName,
			"dropped", validationResult.Dropped,
		)
	}

	payload.Output = validationResult.Accepted
	if payload.Output == nil {
		payload.Output = map[string]any{}
	}

	if err := h.writer.SaveServiceCallAcknowledgement(ctx, env, payload); err != nil {
		result.Decision = messaging.DecisionRetry
		result.Err = fmt.Errorf("save service call acknowledgement: %w", err)
		return result
	}

	result.Decision = messaging.DecisionAck
	return result
}
