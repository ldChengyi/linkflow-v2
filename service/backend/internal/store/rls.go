package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const rlsActorUserSetting = "app.current_user_id"
const rlsInternalServiceSetting = "app.internal_service"

type actorRLSStore struct {
	pool  *pgxpool.Pool
	scope string
}

func newActorRLSStore(pool *pgxpool.Pool, scope string) actorRLSStore {
	return actorRLSStore{
		pool:  pool,
		scope: scope,
	}
}

func (s actorRLSStore) withActor(ctx context.Context, actorUserID string, fn func(pgx.Tx) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin %s transaction: %w", s.scope, err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if _, err := tx.Exec(ctx, `SELECT set_config($1, $2, true)`, rlsActorUserSetting, actorUserID); err != nil {
		return fmt.Errorf("set %s rls actor: %w", s.scope, err)
	}
	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit %s transaction: %w", s.scope, err)
	}
	return nil
}

func (s actorRLSStore) withInternalService(ctx context.Context, serviceName string, fn func(pgx.Tx) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin %s internal transaction: %w", s.scope, err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if _, err := tx.Exec(ctx, `SELECT set_config($1, $2, true)`, rlsInternalServiceSetting, serviceName); err != nil {
		return fmt.Errorf("set %s internal service: %w", s.scope, err)
	}
	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit %s internal transaction: %w", s.scope, err)
	}
	return nil
}
