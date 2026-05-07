package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	KafkaBrokers []string

	PostgresDSN             string
	PostgresMaxOpenConns    int
	PostgresMaxIdleConns    int
	PostgresConnMaxLifetime time.Duration
	PostgresConnMaxIdleTime time.Duration

	ContractsDir string
}

func Load() (Config, error) {
	kafkaBrokers := splitCSV(getEnv("KAFKA_BROKERS", "127.0.0.1:19092"))
	if len(kafkaBrokers) == 0 {
		return Config{}, fmt.Errorf("KAFKA_BROKERS is required")
	}

	maxOpenConns, err := getEnvInt("POSTGRES_MAX_OPEN_CONNS", 10)
	if err != nil {
		return Config{}, err
	}

	maxIdleConns, err := getEnvInt("POSTGRES_MAX_IDLE_CONNS", 5)
	if err != nil {
		return Config{}, err
	}

	connMaxLifetime, err := getEnvDuration("POSTGRES_CONN_MAX_LIFETIME", 30*time.Minute)
	if err != nil {
		return Config{}, err
	}

	connMaxIdleTime, err := getEnvDuration("POSTGRES_CONN_MAX_IDLE_TIME", 5*time.Minute)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		KafkaBrokers: kafkaBrokers,

		PostgresDSN: getEnv("POSTGRES_DSN",
			"postgres://linkflow:linkflow123@127.0.0.1:5432/linkflow?sslmode=disable"),
		PostgresMaxOpenConns:    maxOpenConns,
		PostgresMaxIdleConns:    maxIdleConns,
		PostgresConnMaxLifetime: connMaxLifetime,
		PostgresConnMaxIdleTime: connMaxIdleTime,

		ContractsDir: getEnv("CONTRACTS_DIR", "../../contracts"),
	}

	if cfg.PostgresDSN == "" {
		return Config{}, fmt.Errorf("POSTGRES_DSN is required")
	}
	if cfg.ContractsDir == "" {
		return Config{}, fmt.Errorf("CONTRACTS_DIR is required")
	}

	return cfg, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}

	return out
}

func getEnvInt(key string, def int) (int, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return def, nil
	}

	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	if v <= 0 {
		return 0, fmt.Errorf("%s must be positive", key)
	}

	return v, nil
}

func getEnvDuration(key string, def time.Duration) (time.Duration, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return def, nil
	}

	v, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration: %w", key, err)
	}
	if v <= 0 {
		return 0, fmt.Errorf("%s must be positive", key)
	}

	return v, nil
}
