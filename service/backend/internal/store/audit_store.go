package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/service"
)

type PostgresAuditStore struct {
	pool *pgxpool.Pool
}

func NewPostgresAuditStore(pool *pgxpool.Pool) (*PostgresAuditStore, error) {
	if pool == nil {
		return nil, fmt.Errorf("postgres pool is nil")
	}
	return &PostgresAuditStore{pool: pool}, nil
}

func (s *PostgresAuditStore) WriteAuditLog(ctx context.Context, entry service.AuditEntry) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	metadata, err := service.AuditMetadataJSON(entry.Metadata)
	if err != nil {
		return err
	}

	const query = `
INSERT INTO audit_logs (
    actor_user_id,
    actor_role,
    action,
    resource_type,
    resource_id,
    result,
    error_code,
    method,
    path,
    status_code,
    duration_ms,
    ip,
    user_agent,
    request_id,
    trace_id,
    operation_id,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9, $10, $11, $12,
    $13, $14, $15, $16, $17
)`

	err = s.withAuditUser(ctx, entry.ActorUserID, func(tx pgx.Tx) error {
		_, err := tx.Exec(
			ctx,
			query,
			nullIfEmpty(entry.ActorUserID),
			nullIfEmpty(entry.ActorRole),
			entry.Action,
			entry.ResourceType,
			nullIfEmpty(entry.ResourceID),
			entry.Result,
			nullIfEmpty(entry.ErrorCode),
			entry.Method,
			entry.Path,
			entry.StatusCode,
			entry.DurationMS,
			nullIfEmpty(entry.IP),
			nullIfEmpty(entry.UserAgent),
			nullIfEmpty(entry.RequestID),
			nullIfEmpty(entry.TraceID),
			nullIfEmpty(entry.OperationID),
			metadata,
		)
		return err
	})
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}
	return nil
}

func (s *PostgresAuditStore) ListAuditLogs(ctx context.Context, in service.AuditListInput) (service.PageResult[service.AuditEntry], error) {
	if err := ctx.Err(); err != nil {
		return service.PageResult[service.AuditEntry]{}, err
	}

	const query = `
SELECT
    id::text,
    COALESCE(actor_user_id::text, ''),
    COALESCE(actor_role, ''),
    action,
    resource_type,
    COALESCE(resource_id, ''),
    result,
    COALESCE(error_code, ''),
    method,
    path,
    status_code,
    duration_ms,
    COALESCE(ip, ''),
    COALESCE(user_agent, ''),
    COALESCE(request_id, ''),
    COALESCE(trace_id, ''),
    COALESCE(operation_id, ''),
    metadata,
    created_at
FROM audit_logs
ORDER BY created_at DESC, id DESC
LIMIT $1 OFFSET $2`

	const countQuery = `SELECT count(*) FROM audit_logs`

	var entries []service.AuditEntry
	var total int
	err := s.withAuditUser(ctx, in.UserID, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, countQuery).Scan(&total); err != nil {
			return err
		}

		rows, err := tx.Query(ctx, query, in.PageInput.Limit(), in.PageInput.Offset())
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var entry service.AuditEntry
			var metadata []byte
			if err := rows.Scan(
				&entry.ID,
				&entry.ActorUserID,
				&entry.ActorRole,
				&entry.Action,
				&entry.ResourceType,
				&entry.ResourceID,
				&entry.Result,
				&entry.ErrorCode,
				&entry.Method,
				&entry.Path,
				&entry.StatusCode,
				&entry.DurationMS,
				&entry.IP,
				&entry.UserAgent,
				&entry.RequestID,
				&entry.TraceID,
				&entry.OperationID,
				&metadata,
				&entry.CreatedAt,
			); err != nil {
				return err
			}
			if len(metadata) > 0 {
				if err := json.Unmarshal(metadata, &entry.Metadata); err != nil {
					return fmt.Errorf("unmarshal audit metadata: %w", err)
				}
			}
			if entry.Metadata == nil {
				entry.Metadata = map[string]any{}
			}
			entries = append(entries, entry)
		}
		return rows.Err()
	})
	if err != nil {
		return service.PageResult[service.AuditEntry]{}, fmt.Errorf("list audit logs: %w", err)
	}
	return service.NewPageResult(entries, total, in.PageInput), nil
}

func (s *PostgresAuditStore) withAuditUser(ctx context.Context, userID string, fn func(pgx.Tx) error) error {
	return withRLSUser(ctx, s.pool, userID, "audit", fn)
}
