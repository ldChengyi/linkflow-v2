package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
	"github.com/ldchengyi/linkflow-v2/pkg/public/messaging"
	"github.com/ldchengyi/linkflow-v2/service/device-event-processor/internal/processor"
	"github.com/ldchengyi/linkflow-v2/service/device-event-processor/internal/publisher"
	"github.com/ldchengyi/linkflow-v2/service/device-event-processor/internal/store"
)

type DeviceConnectionStore interface {
	MarkOnline(ctx context.Context, in store.DeviceConnectionEvent) error
	MarkOffline(ctx context.Context, in store.DeviceConnectionEvent) error
}

type DeviceConnectedHandler struct {
	store     DeviceConnectionStore
	publisher EventPublisher
	log       *slog.Logger
}

func NewDeviceConnectedHandler(s DeviceConnectionStore, pub EventPublisher, log *slog.Logger) (*DeviceConnectedHandler, error) {
	if s == nil {
		return nil, fmt.Errorf("device connection store is nil")
	}
	if pub == nil {
		return nil, fmt.Errorf("event publisher is nil")
	}
	if log == nil {
		log = slog.Default()
	}
	return &DeviceConnectedHandler{store: s, publisher: pub, log: log}, nil
}

func (h *DeviceConnectedHandler) Handle(ctx context.Context, env *event.Envelope) processor.Result {
	if err := ctx.Err(); err != nil {
		return processor.Result{Decision: messaging.DecisionRetry, Err: err}
	}
	if env == nil {
		return processor.Result{Decision: messaging.DecisionDrop, Err: fmt.Errorf("event envelope is nil")}
	}

	result := processor.Result{
		EventID:      env.EventID,
		EventType:    env.EventType,
		EventVersion: env.EventVersion,
		TenantID:     env.TenantID,
	}

	var payload event.ConnectedPayload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		result.Decision = messaging.DecisionDrop
		result.Err = fmt.Errorf("decode connected payload: %w", err)
		return result
	}
	result.ProductKey = payload.ProductKey
	result.DeviceSlug = payload.DeviceSlug

	occurredAt, err := parseEnvelopeOccurredAt(env)
	if err != nil {
		result.Decision = messaging.DecisionDrop
		result.Err = err
		return result
	}

	if err := h.store.MarkOnline(ctx, store.DeviceConnectionEvent{
		TenantID:   payload.TenantID,
		ProductID:  payload.ProductID,
		DeviceID:   payload.DeviceID,
		OccurredAt: occurredAt,
	}); err != nil {
		h.log.Warn(
			"mark device online failed",
			"event_id", env.EventID,
			"tenant_id", payload.TenantID,
			"device_id", payload.DeviceID,
			"err", err,
		)
		result.Decision = messaging.DecisionRetry
		result.Err = fmt.Errorf("mark device online: %w", err)
		return result
	}

	if err := h.publisher.Publish(ctx, publisher.PublishInput{
		EventType:    event.TypeDeviceConnectionChanged,
		EventVersion: event.VersionDeviceConnectionChanged,
		TenantID:     env.TenantID,
		CausationID:  env.EventID,
		OccurredAt:   occurredAt,
		Payload: event.ConnectionChangedPayload{
			TenantID:   payload.TenantID,
			ProductID:  payload.ProductID,
			DeviceID:   payload.DeviceID,
			TenantSlug: payload.TenantSlug,
			ProductKey: payload.ProductKey,
			DeviceSlug: payload.DeviceSlug,
			Status:     event.ConnectionStatusOnline,
		},
	}); err != nil {
		result.Decision = messaging.DecisionRetry
		result.Err = fmt.Errorf("publish connection changed: %w", err)
		return result
	}

	result.Decision = messaging.DecisionAck
	return result
}

type DeviceDisconnectedHandler struct {
	store     DeviceConnectionStore
	publisher EventPublisher
	log       *slog.Logger
}

func NewDeviceDisconnectedHandler(s DeviceConnectionStore, pub EventPublisher, log *slog.Logger) (*DeviceDisconnectedHandler, error) {
	if s == nil {
		return nil, fmt.Errorf("device connection store is nil")
	}
	if pub == nil {
		return nil, fmt.Errorf("event publisher is nil")
	}
	if log == nil {
		log = slog.Default()
	}
	return &DeviceDisconnectedHandler{store: s, publisher: pub, log: log}, nil
}

func (h *DeviceDisconnectedHandler) Handle(ctx context.Context, env *event.Envelope) processor.Result {
	if err := ctx.Err(); err != nil {
		return processor.Result{Decision: messaging.DecisionRetry, Err: err}
	}
	if env == nil {
		return processor.Result{Decision: messaging.DecisionDrop, Err: fmt.Errorf("event envelope is nil")}
	}

	result := processor.Result{
		EventID:      env.EventID,
		EventType:    env.EventType,
		EventVersion: env.EventVersion,
		TenantID:     env.TenantID,
	}

	var payload event.DisconnectedPayload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		result.Decision = messaging.DecisionDrop
		result.Err = fmt.Errorf("decode disconnected payload: %w", err)
		return result
	}
	result.ProductKey = payload.ProductKey
	result.DeviceSlug = payload.DeviceSlug

	occurredAt, err := parseEnvelopeOccurredAt(env)
	if err != nil {
		result.Decision = messaging.DecisionDrop
		result.Err = err
		return result
	}

	if err := h.store.MarkOffline(ctx, store.DeviceConnectionEvent{
		TenantID:   payload.TenantID,
		ProductID:  payload.ProductID,
		DeviceID:   payload.DeviceID,
		OccurredAt: occurredAt,
	}); err != nil {
		h.log.Warn(
			"mark device offline failed",
			"event_id", env.EventID,
			"tenant_id", payload.TenantID,
			"device_id", payload.DeviceID,
			"err", err,
		)
		result.Decision = messaging.DecisionRetry
		result.Err = fmt.Errorf("mark device offline: %w", err)
		return result
	}

	if err := h.publisher.Publish(ctx, publisher.PublishInput{
		EventType:    event.TypeDeviceConnectionChanged,
		EventVersion: event.VersionDeviceConnectionChanged,
		TenantID:     env.TenantID,
		CausationID:  env.EventID,
		OccurredAt:   occurredAt,
		Payload: event.ConnectionChangedPayload{
			TenantID:   payload.TenantID,
			ProductID:  payload.ProductID,
			DeviceID:   payload.DeviceID,
			TenantSlug: payload.TenantSlug,
			ProductKey: payload.ProductKey,
			DeviceSlug: payload.DeviceSlug,
			Status:     event.ConnectionStatusOffline,
			Reason:     payload.Reason,
		},
	}); err != nil {
		result.Decision = messaging.DecisionRetry
		result.Err = fmt.Errorf("publish connection changed: %w", err)
		return result
	}

	result.Decision = messaging.DecisionAck
	return result
}

func parseEnvelopeOccurredAt(env *event.Envelope) (time.Time, error) {
	t, err := time.Parse(time.RFC3339Nano, env.OccurredAt)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse occurred_at %q: %w", env.OccurredAt, err)
	}
	return t, nil
}
