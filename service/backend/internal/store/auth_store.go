package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/service"
)

const uniqueViolationCode = "23505"

type PostgresAuthStore struct {
	pool *pgxpool.Pool
}

func NewPostgresAuthStore(pool *pgxpool.Pool) (*PostgresAuthStore, error) {
	if pool == nil {
		return nil, fmt.Errorf("postgres pool is nil")
	}
	return &PostgresAuthStore{pool: pool}, nil
}

func (s *PostgresAuthStore) CreateUser(ctx context.Context, email string, passwordHash string, role string) (service.AuthUser, error) {
	if err := ctx.Err(); err != nil {
		return service.AuthUser{}, err
	}

	const query = `
INSERT INTO users (email, password_hash, role)
VALUES ($1, $2, $3)
RETURNING id::text, email, password_hash, role, status`

	var user service.AuthUser
	if err := s.pool.QueryRow(ctx, query, email, passwordHash, role).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
	); err != nil {
		if isUniqueViolation(err) {
			return service.AuthUser{}, service.ErrUserAlreadyExists
		}
		return service.AuthUser{}, fmt.Errorf("insert user: %w", err)
	}

	return user, nil
}

func (s *PostgresAuthStore) FindUserByEmail(ctx context.Context, email string) (service.AuthUser, error) {
	if err := ctx.Err(); err != nil {
		return service.AuthUser{}, err
	}

	const query = `
SELECT id::text, email, password_hash, role, status
FROM users
WHERE lower(email) = lower($1)`

	var user service.AuthUser
	if err := s.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return service.AuthUser{}, service.ErrInvalidCredentials
		}
		return service.AuthUser{}, fmt.Errorf("find user by email: %w", err)
	}

	return user, nil
}

func (s *PostgresAuthStore) MarkLogin(ctx context.Context, userID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	const query = `UPDATE users SET last_login_at = now(), updated_at = now() WHERE id = $1`
	if _, err := s.pool.Exec(ctx, query, userID); err != nil {
		return fmt.Errorf("mark user login: %w", err)
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode
}
