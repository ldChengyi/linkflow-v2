package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/service"
)

type PostgresTenantStore struct {
	pool *pgxpool.Pool
}

func NewPostgresTenantStore(pool *pgxpool.Pool) (*PostgresTenantStore, error) {
	if pool == nil {
		return nil, fmt.Errorf("postgres pool is nil")
	}
	return &PostgresTenantStore{pool: pool}, nil
}

func (s *PostgresTenantStore) CreateTenant(ctx context.Context, in service.TenantCreateInput) (service.Tenant, error) {
	if err := ctx.Err(); err != nil {
		return service.Tenant{}, err
	}

	const query = `
INSERT INTO tenants (owner_user_id, tenant_slug, tenant_name)
VALUES ($1, $2, $3)
RETURNING id::text, owner_user_id::text, tenant_slug, tenant_name, status, created_at, updated_at`

	var tenant service.Tenant
	err := s.withTenantUser(ctx, in.UserID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, query, in.UserID, in.TenantSlug, in.TenantName).Scan(
			&tenant.ID,
			&tenant.OwnerUserID,
			&tenant.TenantSlug,
			&tenant.TenantName,
			&tenant.Status,
			&tenant.CreatedAt,
			&tenant.UpdatedAt,
		)
	})
	if err != nil {
		if isUniqueViolation(err) {
			return service.Tenant{}, service.ErrTenantAlreadyExists
		}
		return service.Tenant{}, fmt.Errorf("insert tenant: %w", err)
	}
	return tenant, nil
}

func (s *PostgresTenantStore) ListTenants(ctx context.Context, in service.TenantListInput) (service.PageResult[service.Tenant], error) {
	if err := ctx.Err(); err != nil {
		return service.PageResult[service.Tenant]{}, err
	}

	const query = `
SELECT id::text, owner_user_id::text, tenant_slug, tenant_name, status, created_at, updated_at
FROM tenants
ORDER BY created_at DESC, id DESC
LIMIT $1 OFFSET $2`

	const countQuery = `SELECT count(*) FROM tenants`

	var tenants []service.Tenant
	var total int
	err := s.withTenantUser(ctx, in.UserID, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, countQuery).Scan(&total); err != nil {
			return err
		}

		rows, err := tx.Query(ctx, query, in.PageInput.Limit(), in.PageInput.Offset())
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var tenant service.Tenant
			if err := rows.Scan(
				&tenant.ID,
				&tenant.OwnerUserID,
				&tenant.TenantSlug,
				&tenant.TenantName,
				&tenant.Status,
				&tenant.CreatedAt,
				&tenant.UpdatedAt,
			); err != nil {
				return err
			}
			tenants = append(tenants, tenant)
		}
		return rows.Err()
	})
	if err != nil {
		return service.PageResult[service.Tenant]{}, fmt.Errorf("list tenants: %w", err)
	}
	return service.NewPageResult(tenants, total, in.PageInput), nil
}

func (s *PostgresTenantStore) FindTenantByID(ctx context.Context, in service.TenantGetInput) (service.Tenant, error) {
	if err := ctx.Err(); err != nil {
		return service.Tenant{}, err
	}

	const query = `
SELECT id::text, owner_user_id::text, tenant_slug, tenant_name, status, created_at, updated_at
FROM tenants
WHERE id = $1`

	var tenant service.Tenant
	err := s.withTenantUser(ctx, in.UserID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, query, in.TenantID).Scan(
			&tenant.ID,
			&tenant.OwnerUserID,
			&tenant.TenantSlug,
			&tenant.TenantName,
			&tenant.Status,
			&tenant.CreatedAt,
			&tenant.UpdatedAt,
		)
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return service.Tenant{}, service.ErrTenantNotFound
		}
		return service.Tenant{}, fmt.Errorf("find tenant by id: %w", err)
	}
	return tenant, nil
}

func (s *PostgresTenantStore) UpdateTenant(ctx context.Context, in service.TenantUpdateInput) (service.Tenant, error) {
	if err := ctx.Err(); err != nil {
		return service.Tenant{}, err
	}

	const query = `
UPDATE tenants
SET tenant_name = $2,
    status = $3,
    updated_at = now()
WHERE id = $1
RETURNING id::text, owner_user_id::text, tenant_slug, tenant_name, status, created_at, updated_at`

	var tenant service.Tenant
	err := s.withTenantUser(ctx, in.UserID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, query, in.TenantID, in.TenantName, in.Status).Scan(
			&tenant.ID,
			&tenant.OwnerUserID,
			&tenant.TenantSlug,
			&tenant.TenantName,
			&tenant.Status,
			&tenant.CreatedAt,
			&tenant.UpdatedAt,
		)
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return service.Tenant{}, service.ErrTenantNotFound
		}
		return service.Tenant{}, fmt.Errorf("update tenant: %w", err)
	}
	return tenant, nil
}

func (s *PostgresTenantStore) DeleteTenant(ctx context.Context, in service.TenantDeleteInput) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	const query = `DELETE FROM tenants WHERE id = $1`

	err := s.withTenantUser(ctx, in.UserID, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, query, in.TenantID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return service.ErrTenantNotFound
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, service.ErrTenantNotFound) {
			return service.ErrTenantNotFound
		}
		return fmt.Errorf("delete tenant: %w", err)
	}
	return nil
}

func (s *PostgresTenantStore) withTenantUser(ctx context.Context, userID string, fn func(pgx.Tx) error) error {
	return withRLSUser(ctx, s.pool, userID, "tenant", fn)
}
