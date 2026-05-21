package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
)

type EventReportStore struct {
	pool *pgxpool.Pool
}

func NewEventReportStore(pool *pgxpool.Pool) (*EventReportStore, error) {
	if pool == nil {
		return nil, fmt.Errorf("postgres pool is nil")
	}
	return &EventReportStore{pool: pool}, nil
}

func (s *EventReportStore) SaveEventReport(
	ctx context.Context,
	env *event.Envelope,
	payload event.EventReportedPayload,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if env == nil {
		return fmt.Errorf("event envelope is nil")
	}
	if payload.EventName == "" {
		return fmt.Errorf("event report event_name is empty")
	}
	if payload.Params == nil {
		payload.Params = map[string]any{}
	}

	occurredAt, err := time.Parse(time.RFC3339Nano, env.OccurredAt)
	if err != nil {
		return fmt.Errorf("parse occurred_at %q: %w", env.OccurredAt, err)
	}

	const query = `
  INSERT INTO device_event_report_events (
        event_id,
        tenant_id,
        product_key,
        device_slug,
        event_name,
        protocol,
        occurred_at,
        producer,
        trace_id,
        correlation_id,
        causation_id,
        params,
        raw
  ) VALUES (
        $1, $2, $3, $4, $5, $6, $7,
        $8, $9, $10, $11, $12, $13
  )
  ON CONFLICT (event_id, occurred_at) DO NOTHING`

	if _, err := s.pool.Exec(
		ctx,
		query,
		env.EventID,
		env.TenantID,
		payload.ProductKey,
		payload.DeviceSlug,
		payload.EventName,
		payload.Protocol,
		occurredAt,
		env.Producer,
		nullIfEmpty(env.TraceID),
		nullIfEmpty(env.CorrelationID),
		nullIfEmpty(env.CausationID),
		payload.Params,
		emptyMapAsNil(payload.Raw),
	); err != nil {
		return fmt.Errorf("insert device event report event: %w", err)
	}

	return nil
}
