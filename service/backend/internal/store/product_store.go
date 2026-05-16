package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/service"
)

type PostgresProductStore struct {
	actor actorRLSStore
}

func NewPostgresProductStore(pool *pgxpool.Pool) (*PostgresProductStore, error) {
	if pool == nil {
		return nil, fmt.Errorf("postgres pool is nil")
	}
	return &PostgresProductStore{actor: newActorRLSStore(pool, "product")}, nil
}

func (s *PostgresProductStore) CreateProduct(ctx context.Context, in service.ProductCreateInput) (service.Product, error) {
	if err := ctx.Err(); err != nil {
		return service.Product{}, err
	}

	const query = `
INSERT INTO products (
    tenant_id,
    product_key,
    product_name,
    description,
    node_type,
    auth_type,
    protocol_type
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING id::text, tenant_id::text, product_key, product_name, description, node_type, auth_type, protocol_type, status, created_at, updated_at`

	var product service.Product
	err := s.withProductUser(ctx, in.UserID, func(tx pgx.Tx) error {
		ok, err := tenantExists(ctx, tx, in.TenantID)
		if err != nil {
			return err
		}
		if !ok {
			return service.ErrInvalidProductInput
		}

		return tx.QueryRow(
			ctx,
			query,
			in.TenantID,
			in.ProductKey,
			in.ProductName,
			in.Description,
			in.NodeType,
			in.AuthType,
			in.ProtocolType,
		).Scan(
			&product.ID,
			&product.TenantID,
			&product.ProductKey,
			&product.ProductName,
			&product.Description,
			&product.NodeType,
			&product.AuthType,
			&product.ProtocolType,
			&product.Status,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
	})
	if err != nil {
		if isUniqueViolation(err) {
			return service.Product{}, service.ErrProductAlreadyExists
		}
		return service.Product{}, fmt.Errorf("insert product: %w", err)
	}
	return product, nil
}

func (s *PostgresProductStore) ListProducts(ctx context.Context, in service.ProductListInput) (service.PageResult[service.Product], error) {
	if err := ctx.Err(); err != nil {
		return service.PageResult[service.Product]{}, err
	}

	const query = `
SELECT id::text, tenant_id::text, product_key, product_name, description, node_type, auth_type, protocol_type, status, created_at, updated_at
FROM products
WHERE tenant_id = $1
ORDER BY created_at DESC, id DESC
LIMIT $2 OFFSET $3`

	const countQuery = `SELECT count(*) FROM products WHERE tenant_id = $1`

	var products []service.Product
	var total int
	err := s.withProductUser(ctx, in.UserID, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, countQuery, in.TenantID).Scan(&total); err != nil {
			return err
		}

		rows, err := tx.Query(ctx, query, in.TenantID, in.PageInput.Limit(), in.PageInput.Offset())
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var product service.Product
			if err := rows.Scan(
				&product.ID,
				&product.TenantID,
				&product.ProductKey,
				&product.ProductName,
				&product.Description,
				&product.NodeType,
				&product.AuthType,
				&product.ProtocolType,
				&product.Status,
				&product.CreatedAt,
				&product.UpdatedAt,
			); err != nil {
				return err
			}
			products = append(products, product)
		}
		return rows.Err()
	})
	if err != nil {
		return service.PageResult[service.Product]{}, fmt.Errorf("list products: %w", err)
	}
	return service.NewPageResult(products, total, in.PageInput), nil
}

func (s *PostgresProductStore) FindProductByID(ctx context.Context, in service.ProductGetInput) (service.Product, error) {
	if err := ctx.Err(); err != nil {
		return service.Product{}, err
	}

	const query = `
SELECT id::text, tenant_id::text, product_key, product_name, description, node_type, auth_type, protocol_type, status, created_at, updated_at
FROM products
WHERE id = $1`

	var product service.Product
	err := s.withProductUser(ctx, in.UserID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, query, in.ProductID).Scan(
			&product.ID,
			&product.TenantID,
			&product.ProductKey,
			&product.ProductName,
			&product.Description,
			&product.NodeType,
			&product.AuthType,
			&product.ProtocolType,
			&product.Status,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return service.Product{}, service.ErrProductNotFound
		}
		return service.Product{}, fmt.Errorf("find product by id: %w", err)
	}
	return product, nil
}

func (s *PostgresProductStore) UpdateProduct(ctx context.Context, in service.ProductUpdateInput) (service.Product, error) {
	if err := ctx.Err(); err != nil {
		return service.Product{}, err
	}

	const query = `
UPDATE products
SET product_name = $2,
    description = $3,
    node_type = $4,
    auth_type = $5,
    protocol_type = $6,
    status = $7,
    updated_at = now()
WHERE id = $1
RETURNING id::text, tenant_id::text, product_key, product_name, description, node_type, auth_type, protocol_type, status, created_at, updated_at`

	var product service.Product
	err := s.withProductUser(ctx, in.UserID, func(tx pgx.Tx) error {
		return tx.QueryRow(
			ctx,
			query,
			in.ProductID,
			in.ProductName,
			in.Description,
			in.NodeType,
			in.AuthType,
			in.ProtocolType,
			in.Status,
		).Scan(
			&product.ID,
			&product.TenantID,
			&product.ProductKey,
			&product.ProductName,
			&product.Description,
			&product.NodeType,
			&product.AuthType,
			&product.ProtocolType,
			&product.Status,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return service.Product{}, service.ErrProductNotFound
		}
		return service.Product{}, fmt.Errorf("update product: %w", err)
	}
	return product, nil
}

func (s *PostgresProductStore) DeleteProduct(ctx context.Context, in service.ProductDeleteInput) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	const query = `DELETE FROM products WHERE id = $1`

	err := s.withProductUser(ctx, in.UserID, func(tx pgx.Tx) error {
		inUse, err := productHasDevices(ctx, tx, in.ProductID)
		if err != nil {
			return err
		}
		if inUse {
			return service.ErrInvalidProductInput
		}

		tag, err := tx.Exec(ctx, query, in.ProductID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return service.ErrProductNotFound
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			return service.ErrProductNotFound
		}
		return fmt.Errorf("delete product: %w", err)
	}
	return nil
}

func (s *PostgresProductStore) withProductUser(ctx context.Context, userID string, fn func(pgx.Tx) error) error {
	return s.actor.withActor(ctx, userID, fn)
}

func tenantExists(ctx context.Context, tx pgx.Tx, tenantID string) (bool, error) {
	const query = `SELECT EXISTS (SELECT 1 FROM tenants WHERE id = $1)`

	var ok bool
	if err := tx.QueryRow(ctx, query, tenantID).Scan(&ok); err != nil {
		return false, fmt.Errorf("check tenant: %w", err)
	}
	return ok, nil
}

func productHasDevices(ctx context.Context, tx pgx.Tx, productID string) (bool, error) {
	const query = `SELECT EXISTS (SELECT 1 FROM devices WHERE product_id = $1)`

	var ok bool
	if err := tx.QueryRow(ctx, query, productID).Scan(&ok); err != nil {
		return false, fmt.Errorf("check product devices: %w", err)
	}
	return ok, nil
}
