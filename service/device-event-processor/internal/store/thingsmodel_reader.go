package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ldchengyi/linkflow-v2/service/device-event-processor/internal/validator"
)

const adminServiceName = "device-event-processor"

type ThingsModelReader struct {
	pool *pgxpool.Pool
}

func NewThingsModelReader(pool *pgxpool.Pool) (*ThingsModelReader, error) {
	if pool == nil {
		return nil, fmt.Errorf("postgres pool is nil")
	}
	return &ThingsModelReader{pool: pool}, nil
}

func (r *ThingsModelReader) FindCurrentByProductID(ctx context.Context, tenantID string, productID string) (validator.ThingsModelDefinition, error) {
	if err := ctx.Err(); err != nil {
		return validator.ThingsModelDefinition{}, err
	}

	const query = `
SELECT properties::text, events::text
FROM thingsmodel
WHERE tenant_id = $1::uuid
  AND product_id = $2::uuid
  AND is_current = true
LIMIT 1`

	var propertiesRaw string
	var eventsRaw string
	err := r.withAdmin(ctx, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, query, tenantID, productID).Scan(&propertiesRaw, &eventsRaw)
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return validator.ThingsModelDefinition{}, validator.ErrThingsModelNotFound
		}
		return validator.ThingsModelDefinition{}, fmt.Errorf("find current thingsmodel by product_id: %w", err)
	}
	return validator.ParseDefinition(tenantID, productID, []byte(propertiesRaw), []byte(eventsRaw))
}

func (r *ThingsModelReader) withAdmin(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin thingsmodel reader transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if _, err := tx.Exec(ctx, `SELECT set_config($1, $2, true)`, "app.admin_service", adminServiceName); err != nil {
		return fmt.Errorf("set admin service: %w", err)
	}
	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit thingsmodel reader transaction: %w", err)
	}
	return nil
}
