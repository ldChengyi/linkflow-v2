package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
)

type ServiceCallStore struct {
	pool *pgxpool.Pool
}

func NewServiceCallStore(pool *pgxpool.Pool) (*ServiceCallStore, error) {
	if pool == nil {
		return nil, fmt.Errorf("postgres pool is nil")
	}
	return &ServiceCallStore{pool: pool}, nil
}

func (s *ServiceCallStore) SaveServiceCallAcknowledgement(
	ctx context.Context,
	env *event.Envelope,
	payload event.ServiceCallAcknowledgedPayload,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if env == nil {
		return fmt.Errorf("event envelope is nil")
	}
	if payload.ServiceName == "" {
		return fmt.Errorf("service call service_name is empty")
	}
	if payload.CommandID == "" {
		return fmt.Errorf("service call command_id is empty")
	}
	if payload.Output == nil {
		payload.Output = map[string]any{}
	}

	occurredAt, err := time.Parse(time.RFC3339Nano, env.OccurredAt)
	if err != nil {
		return fmt.Errorf("parse occurred_at %q: %w", env.OccurredAt, err)
	}

	const query = `
	  INSERT INTO device_service_call_ack_events (
	        event_id,
	        command_id,
	        tenant_id,
	        product_key,
	        device_slug,
	        service_name,
	        protocol,
	        success,
	        code,
	        message,
	        occurred_at,
	        producer,
	        trace_id,
	        correlation_id,
	        causation_id,
	        output,
	        raw
	  ) VALUES (
	        $1, $2, $3, $4, $5, $6, $7, $8,
	        $9, $10, $11, $12, $13, $14, $15, $16, $17
	  )
	  ON CONFLICT (event_id, occurred_at) DO NOTHING`

	if _, err := s.pool.Exec(
		ctx,
		query,
		env.EventID,
		payload.CommandID,
		env.TenantID,
		payload.ProductKey,
		payload.DeviceSlug,
		payload.ServiceName,
		payload.Protocol,
		payload.Success,
		nullIfEmpty(payload.Code),
		nullIfEmpty(payload.Message),
		occurredAt,
		env.Producer,
		nullIfEmpty(env.TraceID),
		nullIfEmpty(env.CorrelationID),
		nullIfEmpty(env.CausationID),
		payload.Output,
		emptyMapAsNil(payload.Raw),
	); err != nil {
		return fmt.Errorf("insert device service call ack event: %w", err)
	}

	return nil
}
