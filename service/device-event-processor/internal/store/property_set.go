package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
)

type PropertySetStore struct {
	pool *pgxpool.Pool
}

func NewPropertySetStore(pool *pgxpool.Pool) (*PropertySetStore, error) {
	if pool == nil {
		return nil, fmt.Errorf("postgres pool is nil")
	}
	return &PropertySetStore{pool: pool}, nil
}

func (s *PropertySetStore) SavePropertySetRequest(
	ctx context.Context,
	env *event.Envelope,
	payload event.PropertySetRequestedPayload,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if env == nil {
		return fmt.Errorf("event envelope is nil")
	}
	if payload.CommandID == "" {
		return fmt.Errorf("property set command_id is empty")
	}
	if len(payload.Properties) == 0 {
		return fmt.Errorf("property set properties are empty")
	}

	occurredAt, err := time.Parse(time.RFC3339Nano, env.OccurredAt)
	if err != nil {
		return fmt.Errorf("parse occurred_at %q: %w", env.OccurredAt, err)
	}
	properties, err := json.Marshal(payload.Properties)
	if err != nil {
		return fmt.Errorf("encode property set properties: %w", err)
	}

	const query = `
	  INSERT INTO device_property_set_events (
	        command_id,
	        tenant_id,
	        product_key,
	        device_slug,
	        protocol,
	        topic,
	        occurred_at,
	        producer,
	        requested_by,
	        properties
	  ) VALUES (
	        $1, $2, $3, $4, $5, $6, $7,
	        $8,
	        $9,
	        $10::jsonb
	  )
	  ON CONFLICT (command_id, occurred_at) DO NOTHING`

	if _, err := s.pool.Exec(
		ctx,
		query,
		payload.CommandID,
		env.TenantID,
		payload.ProductKey,
		payload.DeviceSlug,
		payload.Protocol,
		payload.Topic,
		occurredAt,
		env.Producer,
		payload.RequestedBy,
		properties,
	); err != nil {
		return fmt.Errorf("insert device property set event: %w", err)
	}

	return nil
}

func (s *PropertySetStore) SavePropertySetAcknowledgement(
	ctx context.Context,
	env *event.Envelope,
	payload event.PropertySetAcknowledgedPayload,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if env == nil {
		return fmt.Errorf("event envelope is nil")
	}
	if payload.Properties == nil {
		payload.Properties = map[string]any{}
	}
	if payload.Raw == nil {
		payload.Raw = map[string]any{}
	}

	occurredAt, err := time.Parse(time.RFC3339Nano, env.OccurredAt)
	if err != nil {
		return fmt.Errorf("parse occurred_at %q: %w", env.OccurredAt, err)
	}

	const query = `
	  INSERT INTO device_property_set_ack_events (
	        event_id,
	        command_id,
	        tenant_id,
	        product_key,
	        device_slug,
	        protocol,
	        success,
	        code,
	        message,
	        occurred_at,
	        producer,
	        trace_id,
	        correlation_id,
	        causation_id,
	        properties,
	        raw
	  ) VALUES (
	        $1, $2, $3, $4, $5, $6, $7,
	        $8, $9, $10, $11, $12, $13, $14, $15, $16
	  )
	  ON CONFLICT (event_id, occurred_at) DO NOTHING`

	if _, err := s.pool.Exec(
		ctx,
		query,
		env.EventID,
		nullIfEmpty(payload.CommandID),
		env.TenantID,
		payload.ProductKey,
		payload.DeviceSlug,
		payload.Protocol,
		payload.Success,
		nullIfEmpty(payload.Code),
		nullIfEmpty(payload.Message),
		occurredAt,
		env.Producer,
		nullIfEmpty(env.TraceID),
		nullIfEmpty(env.CorrelationID),
		nullIfEmpty(env.CausationID),
		payload.Properties,
		emptyMapAsNil(payload.Raw),
	); err != nil {
		return fmt.Errorf("insert device property set ack event: %w", err)
	}

	return nil
}
