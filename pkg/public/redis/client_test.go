package redis

import (
	"testing"
	"time"
)

func TestApplyDefaults(t *testing.T) {
	cfg := Config{}

	applyDefaults(&cfg)

	if cfg.DB != defaultDB {
		t.Fatalf("DB = %d, want %d", cfg.DB, defaultDB)
	}
	if cfg.PoolSize != defaultPoolSize {
		t.Fatalf("PoolSize = %d, want %d", cfg.PoolSize, defaultPoolSize)
	}
	if cfg.MinIdleConns != defaultMinIdleConns {
		t.Fatalf("MinIdleConns = %d, want %d", cfg.MinIdleConns, defaultMinIdleConns)
	}
	if cfg.DialTimeout != defaultDialTimeout {
		t.Fatalf("DialTimeout = %s, want %s", cfg.DialTimeout, defaultDialTimeout)
	}
	if cfg.ReadTimeout != defaultReadTimeout {
		t.Fatalf("ReadTimeout = %s, want %s", cfg.ReadTimeout, defaultReadTimeout)
	}
	if cfg.WriteTimeout != defaultWriteTimeout {
		t.Fatalf("WriteTimeout = %s, want %s", cfg.WriteTimeout, defaultWriteTimeout)
	}
	if cfg.PingTimeout != defaultPingTimeout {
		t.Fatalf("PingTimeout = %s, want %s", cfg.PingTimeout, defaultPingTimeout)
	}
}

func TestApplyDefaultsCapsMinIdleConns(t *testing.T) {
	cfg := Config{
		PoolSize:     2,
		MinIdleConns: 8,
		DialTimeout:  time.Second,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		PingTimeout:  time.Second,
	}

	applyDefaults(&cfg)

	if cfg.MinIdleConns != cfg.PoolSize {
		t.Fatalf("MinIdleConns = %d, want %d", cfg.MinIdleConns, cfg.PoolSize)
	}
}
