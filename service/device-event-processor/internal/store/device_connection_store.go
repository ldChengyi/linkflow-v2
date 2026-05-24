package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"
)

const (
	connectionStatusOnline  = "online"
	connectionStatusOffline = "offline"
)

type DeviceConnectionStore struct {
	pool      *pgxpool.Pool
	redis     goredis.Cmdable
	onlineTTL time.Duration
}

func NewDeviceConnectionStore(pool *pgxpool.Pool, redis goredis.Cmdable, onlineTTL time.Duration) (*DeviceConnectionStore, error) {
	if pool == nil {
		return nil, fmt.Errorf("postgres pool is nil")
	}
	if redis == nil {
		return nil, fmt.Errorf("redis client is nil")
	}
	if onlineTTL <= 0 {
		return nil, fmt.Errorf("device online ttl must be positive")
	}
	return &DeviceConnectionStore{pool: pool, redis: redis, onlineTTL: onlineTTL}, nil
}

type DeviceConnectionEvent struct {
	TenantID   string
	ProductID  string
	DeviceID   string
	OccurredAt time.Time
}

func (s *DeviceConnectionStore) MarkOnline(ctx context.Context, in DeviceConnectionEvent) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	applied, err := s.updateConnectionStatus(ctx, in, connectionStatusOnline)
	if err != nil {
		return false, err
	}
	if !applied {
		return false, nil
	}
	if err := s.redis.Set(ctx, deviceOnlineKey(in.TenantID, in.DeviceID), in.OccurredAt.UTC().Format(time.RFC3339Nano), s.onlineTTL).Err(); err != nil {
		return false, fmt.Errorf("set redis online key: %w", err)
	}
	return true, nil
}

func (s *DeviceConnectionStore) MarkOffline(ctx context.Context, in DeviceConnectionEvent) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	applied, err := s.updateConnectionStatus(ctx, in, connectionStatusOffline)
	if err != nil {
		return false, err
	}
	if !applied {
		return false, nil
	}
	if err := s.redis.Del(ctx, deviceOnlineKey(in.TenantID, in.DeviceID)).Err(); err != nil {
		return false, fmt.Errorf("delete redis online key: %w", err)
	}
	return true, nil
}

func (s *DeviceConnectionStore) updateConnectionStatus(ctx context.Context, in DeviceConnectionEvent, status string) (bool, error) {
	const query = `
UPDATE devices
SET connection_status = $1,
    last_seen_at = $2,
    updated_at = now()
WHERE id = $3::uuid
  AND tenant_id = $4::uuid
  AND (last_seen_at IS NULL OR last_seen_at <= $2)`

	var applied bool
	if err := s.withAdmin(ctx, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, query, status, in.OccurredAt, in.DeviceID, in.TenantID)
		if err != nil {
			return fmt.Errorf("update device connection_status: %w", err)
		}
		applied = tag.RowsAffected() > 0
		return nil
	}); err != nil {
		return false, err
	}
	return applied, nil
}

func (s *DeviceConnectionStore) withAdmin(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin device connection transaction: %w", err)
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
		return fmt.Errorf("commit device connection transaction: %w", err)
	}
	return nil
}

func deviceOnlineKey(tenantID string, deviceID string) string {
	return "device:online:" + tenantID + ":" + deviceID
}
