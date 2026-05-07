package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
)

type PropertyReportStore struct {
	pool *pgxpool.Pool
}

func NewPropertyReportStore(pool *pgxpool.Pool) (*PropertyReportStore, error) {
	if pool == nil {
		return nil, fmt.Errorf("postgres pool is nil")
	}

	return &PropertyReportStore{pool: pool}, nil
}

func (s *PropertyReportStore) SavePropertyReport(
	ctx context.Context,
	env *event.Envelope,
	payload event.PropertyReportedPayload,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if env == nil {
		return fmt.Errorf("event envelope is nil")
	}

	if len(payload.Properties) == 0 {
		return fmt.Errorf("property report properties are empty")
	}

	occurredAt, err := time.Parse(time.RFC3339Nano, env.OccurredAt)
	if err != nil {
		return fmt.Errorf("parse occurred_at %q: %w", env.OccurredAt, err)
	}

	const query = `
  INSERT INTO device_property_report_events (
        event_id,
        tenant_id,
        product_key,
        device_id,
        protocol,
        occurred_at,
        producer,
        trace_id,
        correlation_id,
        causation_id,
        properties,
        raw
  ) VALUES (
        $1, $2, $3, $4, $5, $6,
        $7, $8, $9, $10, $11, $12
  )
  ON CONFLICT (event_id, occurred_at) DO NOTHING`

	if _, err := s.pool.Exec(
		ctx,
		query,
		env.EventID,
		env.TenantID,
		payload.ProductKey,
		payload.DeviceID,
		payload.Protocol,
		occurredAt,
		env.Producer,
		nullIfEmpty(env.TraceID),
		nullIfEmpty(env.CorrelationID),
		nullIfEmpty(env.CausationID),
		payload.Properties,
		emptyMapAsNil(payload.Raw),
	); err != nil {
		return fmt.Errorf("insert device property report event: %w", err)
	}

	return nil

}
