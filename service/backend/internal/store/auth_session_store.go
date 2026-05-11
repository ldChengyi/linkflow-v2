package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/ldchengyi/linkflow-v2/service/backend/internal/domain"
)

const authAccessSessionKeyPrefix = "auth:access:"

type RedisAuthSessionStore struct {
	client goredis.Cmdable
}

func NewRedisAuthSessionStore(client goredis.Cmdable) (*RedisAuthSessionStore, error) {
	if client == nil {
		return nil, fmt.Errorf("redis client is nil")
	}
	return &RedisAuthSessionStore{client: client}, nil
}

func (s *RedisAuthSessionStore) CreateAccessSession(ctx context.Context, session domain.AuthSession) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if session.TokenID == "" {
		return fmt.Errorf("auth session token id is required")
	}
	if session.UserID == "" {
		return fmt.Errorf("auth session user id is required")
	}
	if session.Role == "" {
		return fmt.Errorf("auth session role is required")
	}

	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		return fmt.Errorf("auth session expires_at must be in the future")
	}

	raw, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("marshal auth session: %w", err)
	}
	if err := s.client.Set(ctx, authAccessSessionKey(session.TokenID), raw, ttl).Err(); err != nil {
		return fmt.Errorf("set auth session: %w", err)
	}
	return nil
}

func (s *RedisAuthSessionStore) ValidateAccessSession(ctx context.Context, tokenID string) (domain.AuthSession, error) {
	if err := ctx.Err(); err != nil {
		return domain.AuthSession{}, err
	}
	if tokenID == "" {
		return domain.AuthSession{}, domain.ErrAuthSessionNotFound
	}

	raw, err := s.client.Get(ctx, authAccessSessionKey(tokenID)).Bytes()
	if err != nil {
		if err == goredis.Nil {
			return domain.AuthSession{}, domain.ErrAuthSessionNotFound
		}
		return domain.AuthSession{}, fmt.Errorf("get auth session: %w", err)
	}

	var session domain.AuthSession
	if err := json.Unmarshal(raw, &session); err != nil {
		return domain.AuthSession{}, fmt.Errorf("unmarshal auth session: %w", err)
	}
	if session.TokenID == "" || session.UserID == "" || session.Role == "" {
		return domain.AuthSession{}, fmt.Errorf("auth session is incomplete")
	}
	if !session.ExpiresAt.IsZero() && time.Now().UTC().After(session.ExpiresAt) {
		return domain.AuthSession{}, domain.ErrAuthSessionNotFound
	}
	return session, nil
}

func (s *RedisAuthSessionStore) DeleteAccessSession(ctx context.Context, tokenID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if tokenID == "" {
		return nil
	}
	if err := s.client.Del(ctx, authAccessSessionKey(tokenID)).Err(); err != nil {
		return fmt.Errorf("delete auth session: %w", err)
	}
	return nil
}

func authAccessSessionKey(tokenID string) string {
	return authAccessSessionKeyPrefix + tokenID
}
