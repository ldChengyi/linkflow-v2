package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/service"
)

type PostgresThingsModelStore struct {
	actor actorRLSStore
}

func NewPostgresThingsModelStore(pool *pgxpool.Pool) (*PostgresThingsModelStore, error) {
	if pool == nil {
		return nil, fmt.Errorf("postgres pool is nil")
	}
	return &PostgresThingsModelStore{actor: newActorRLSStore(pool, "thingsmodel")}, nil
}

func (s *PostgresThingsModelStore) CreateThingsModel(ctx context.Context, in service.ThingsModelCreateInput) (service.ThingsModel, error) {
	if err := ctx.Err(); err != nil {
		return service.ThingsModel{}, err
	}

	properties, events, services, err := marshalThingsModelObjects(in.Properties, in.Events, in.Services)
	if err != nil {
		return service.ThingsModel{}, service.ErrInvalidThingsModelInput
	}

	const insertQuery = `
INSERT INTO thingsmodel (
    tenant_id,
    product_id,
    model_version,
    model_name,
    description,
    status,
    is_current,
    properties,
    events,
    services,
    published_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8::jsonb, $9::jsonb, $10::jsonb,
    CASE WHEN $6 = 'published' THEN now() ELSE NULL END
)
RETURNING id::text, tenant_id::text, product_id::text, model_version, model_name, description, status, is_current, properties, events, services, published_at, created_at, updated_at`

	var model service.ThingsModel
	err = s.withThingsModelUser(ctx, in.UserID, func(tx pgx.Tx) error {
		ok, err := productBelongsToTenant(ctx, tx, in.ProductID, in.TenantID)
		if err != nil {
			return err
		}
		if !ok {
			return service.ErrInvalidThingsModelInput
		}
		if in.IsCurrent {
			if err := clearCurrentThingsModel(ctx, tx, in.TenantID, in.ProductID, ""); err != nil {
				return err
			}
		}
		return scanThingsModel(tx.QueryRow(
			ctx,
			insertQuery,
			in.TenantID,
			in.ProductID,
			in.ModelVersion,
			in.ModelName,
			in.Description,
			in.Status,
			in.IsCurrent,
			properties,
			events,
			services,
		), &model)
	})
	if err != nil {
		if isUniqueViolation(err) {
			return service.ThingsModel{}, service.ErrThingsModelAlreadyExists
		}
		if errors.Is(err, service.ErrInvalidThingsModelInput) {
			return service.ThingsModel{}, err
		}
		return service.ThingsModel{}, fmt.Errorf("insert thingsmodel: %w", err)
	}
	return model, nil
}

func (s *PostgresThingsModelStore) ListThingsModels(ctx context.Context, in service.ThingsModelListInput) (service.PageResult[service.ThingsModel], error) {
	if err := ctx.Err(); err != nil {
		return service.PageResult[service.ThingsModel]{}, err
	}

	const listByTenantQuery = `
SELECT id::text, tenant_id::text, product_id::text, model_version, model_name, description, status, is_current, properties, events, services, published_at, created_at, updated_at
FROM thingsmodel
WHERE tenant_id = $1
ORDER BY updated_at DESC, id DESC
LIMIT $2 OFFSET $3`
	const countByTenantQuery = `SELECT count(*) FROM thingsmodel WHERE tenant_id = $1`
	const listByProductQuery = `
SELECT id::text, tenant_id::text, product_id::text, model_version, model_name, description, status, is_current, properties, events, services, published_at, created_at, updated_at
FROM thingsmodel
WHERE tenant_id = $1 AND product_id = $2
ORDER BY updated_at DESC, id DESC
LIMIT $3 OFFSET $4`
	const countByProductQuery = `SELECT count(*) FROM thingsmodel WHERE tenant_id = $1 AND product_id = $2`

	var models []service.ThingsModel
	var total int
	err := s.withThingsModelUser(ctx, in.UserID, func(tx pgx.Tx) error {
		if in.ProductID == "" {
			if err := tx.QueryRow(ctx, countByTenantQuery, in.TenantID).Scan(&total); err != nil {
				return err
			}
			rows, err := tx.Query(ctx, listByTenantQuery, in.TenantID, in.PageInput.Limit(), in.PageInput.Offset())
			if err != nil {
				return err
			}
			defer rows.Close()
			return scanThingsModelRows(rows, &models)
		}

		ok, err := productBelongsToTenant(ctx, tx, in.ProductID, in.TenantID)
		if err != nil {
			return err
		}
		if !ok {
			return service.ErrInvalidThingsModelInput
		}
		if err := tx.QueryRow(ctx, countByProductQuery, in.TenantID, in.ProductID).Scan(&total); err != nil {
			return err
		}
		rows, err := tx.Query(ctx, listByProductQuery, in.TenantID, in.ProductID, in.PageInput.Limit(), in.PageInput.Offset())
		if err != nil {
			return err
		}
		defer rows.Close()
		return scanThingsModelRows(rows, &models)
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidThingsModelInput) {
			return service.PageResult[service.ThingsModel]{}, err
		}
		return service.PageResult[service.ThingsModel]{}, fmt.Errorf("list thingsmodels: %w", err)
	}
	return service.NewPageResult(models, total, in.PageInput), nil
}

func (s *PostgresThingsModelStore) FindThingsModelByID(ctx context.Context, in service.ThingsModelGetInput) (service.ThingsModel, error) {
	if err := ctx.Err(); err != nil {
		return service.ThingsModel{}, err
	}

	const query = `
SELECT id::text, tenant_id::text, product_id::text, model_version, model_name, description, status, is_current, properties, events, services, published_at, created_at, updated_at
FROM thingsmodel
WHERE id = $1`

	var model service.ThingsModel
	err := s.withThingsModelUser(ctx, in.UserID, func(tx pgx.Tx) error {
		return scanThingsModel(tx.QueryRow(ctx, query, in.ThingsModelID), &model)
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return service.ThingsModel{}, service.ErrThingsModelNotFound
		}
		return service.ThingsModel{}, fmt.Errorf("find thingsmodel by id: %w", err)
	}
	return model, nil
}

func (s *PostgresThingsModelStore) UpdateThingsModel(ctx context.Context, in service.ThingsModelUpdateInput) (service.ThingsModel, error) {
	if err := ctx.Err(); err != nil {
		return service.ThingsModel{}, err
	}

	properties, events, services, err := marshalThingsModelObjects(in.Properties, in.Events, in.Services)
	if err != nil {
		return service.ThingsModel{}, service.ErrInvalidThingsModelInput
	}

	const selectScopeQuery = `SELECT tenant_id::text, product_id::text FROM thingsmodel WHERE id = $1`
	const updateQuery = `
UPDATE thingsmodel
SET model_name = $2,
    description = $3,
    status = $4,
    is_current = $5,
    properties = $6::jsonb,
    events = $7::jsonb,
    services = $8::jsonb,
    published_at = CASE WHEN $4 = 'published' THEN COALESCE(published_at, now()) ELSE NULL END,
    updated_at = now()
WHERE id = $1
RETURNING id::text, tenant_id::text, product_id::text, model_version, model_name, description, status, is_current, properties, events, services, published_at, created_at, updated_at`

	var model service.ThingsModel
	err = s.withThingsModelUser(ctx, in.UserID, func(tx pgx.Tx) error {
		var tenantID string
		var productID string
		if err := tx.QueryRow(ctx, selectScopeQuery, in.ThingsModelID).Scan(&tenantID, &productID); err != nil {
			return err
		}
		if in.IsCurrent {
			if err := clearCurrentThingsModel(ctx, tx, tenantID, productID, in.ThingsModelID); err != nil {
				return err
			}
		}
		return scanThingsModel(tx.QueryRow(
			ctx,
			updateQuery,
			in.ThingsModelID,
			in.ModelName,
			in.Description,
			in.Status,
			in.IsCurrent,
			properties,
			events,
			services,
		), &model)
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return service.ThingsModel{}, service.ErrThingsModelNotFound
		}
		if isUniqueViolation(err) {
			return service.ThingsModel{}, service.ErrThingsModelAlreadyExists
		}
		return service.ThingsModel{}, fmt.Errorf("update thingsmodel: %w", err)
	}
	return model, nil
}

func (s *PostgresThingsModelStore) DeleteThingsModel(ctx context.Context, in service.ThingsModelDeleteInput) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	const query = `DELETE FROM thingsmodel WHERE id = $1`

	err := s.withThingsModelUser(ctx, in.UserID, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, query, in.ThingsModelID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return service.ErrThingsModelNotFound
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, service.ErrThingsModelNotFound) {
			return service.ErrThingsModelNotFound
		}
		return fmt.Errorf("delete thingsmodel: %w", err)
	}
	return nil
}

func (s *PostgresThingsModelStore) withThingsModelUser(ctx context.Context, userID string, fn func(pgx.Tx) error) error {
	return s.actor.withActor(ctx, userID, fn)
}

func clearCurrentThingsModel(ctx context.Context, tx pgx.Tx, tenantID string, productID string, exceptID string) error {
	const query = `
UPDATE thingsmodel
SET is_current = false, updated_at = now()
WHERE tenant_id = $1 AND product_id = $2 AND is_current = true`
	const exceptQuery = `
UPDATE thingsmodel
SET is_current = false, updated_at = now()
WHERE tenant_id = $1 AND product_id = $2 AND is_current = true AND id <> $3::uuid`

	if exceptID == "" {
		if _, err := tx.Exec(ctx, query, tenantID, productID); err != nil {
			return fmt.Errorf("clear current thingsmodel: %w", err)
		}
		return nil
	}
	if _, err := tx.Exec(ctx, exceptQuery, tenantID, productID, exceptID); err != nil {
		return fmt.Errorf("clear current thingsmodel: %w", err)
	}
	return nil
}

func marshalThingsModelObjects(properties, events, services service.ThingsModelObject) ([]byte, []byte, []byte, error) {
	propertiesJSON, err := json.Marshal(properties)
	if err != nil {
		return nil, nil, nil, err
	}
	eventsJSON, err := json.Marshal(events)
	if err != nil {
		return nil, nil, nil, err
	}
	servicesJSON, err := json.Marshal(services)
	if err != nil {
		return nil, nil, nil, err
	}
	return propertiesJSON, eventsJSON, servicesJSON, nil
}

type thingsModelScanner interface {
	Scan(dest ...any) error
}

func scanThingsModel(row thingsModelScanner, model *service.ThingsModel) error {
	var properties []byte
	var events []byte
	var services []byte
	if err := row.Scan(
		&model.ID,
		&model.TenantID,
		&model.ProductID,
		&model.ModelVersion,
		&model.ModelName,
		&model.Description,
		&model.Status,
		&model.IsCurrent,
		&properties,
		&events,
		&services,
		&model.PublishedAt,
		&model.CreatedAt,
		&model.UpdatedAt,
	); err != nil {
		return err
	}
	if err := json.Unmarshal(properties, &model.Properties); err != nil {
		return fmt.Errorf("decode thingsmodel properties: %w", err)
	}
	if err := json.Unmarshal(events, &model.Events); err != nil {
		return fmt.Errorf("decode thingsmodel events: %w", err)
	}
	if err := json.Unmarshal(services, &model.Services); err != nil {
		return fmt.Errorf("decode thingsmodel services: %w", err)
	}
	return nil
}

func scanThingsModelRows(rows pgx.Rows, models *[]service.ThingsModel) error {
	for rows.Next() {
		var model service.ThingsModel
		if err := scanThingsModel(rows, &model); err != nil {
			return err
		}
		*models = append(*models, model)
	}
	return rows.Err()
}
