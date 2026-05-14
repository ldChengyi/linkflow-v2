package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func withRLSUser(ctx context.Context, pool *pgxpool.Pool, userID string, scope string, fn func(pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin %s transaction: %w", scope, err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if _, err := tx.Exec(ctx, `SELECT set_config('app.current_user_id', $1, true)`, userID); err != nil {
		return fmt.Errorf("set %s rls user: %w", scope, err)
	}
	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit %s transaction: %w", scope, err)
	}
	return nil
}
