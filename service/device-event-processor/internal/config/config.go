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
	KafkaGroupID string

	ConsumerWorkers      int
	ConsumerBuffer       int
	ConsumerMaxRetries   int
	ConsumerRetryBackoff time.Duration

	PostgresDSN             string
	PostgresMaxOpenConns    int
	PostgresMaxIdleConns    int
	PostgresConnMaxLifetime time.Duration
	PostgresConnMaxIdleTime time.Duration

	RedisAddr         string
	RedisPassword     string
	RedisDB           int
	RedisDialTimeout  time.Duration
	RedisReadTimeout  time.Duration
	RedisWriteTimeout time.Duration

	DeviceOnlineTTL time.Duration

	EMQXAPIURL         string
	EMQXAPIKey         string
	EMQXAPISecret      string
	EMQXPublishTimeout time.Duration

	ContractsDir string
}

func Load() (Config, error) {
	kafkaBrokers := splitCSV(getEnv("KAFKA_BROKERS", "127.0.0.1:19092"))
	if len(kafkaBrokers) == 0 {
		return Config{}, fmt.Errorf("KAFKA_BROKERS is required")
	}

	consumerWorkers, err := getEnvInt("CONSUMER_WORKERS", 4)
	if err != nil {
		return Config{}, err
	}

	consumerBuffer, err := getEnvInt("CONSUMER_BUFFER", 128)
	if err != nil {
		return Config{}, err
	}

	consumerMaxRetries, err := getEnvInt("CONSUMER_MAX_RETRIES", 3)
	if err != nil {
		return Config{}, err
	}

	consumerRetryBackoff, err := getEnvDuration("CONSUMER_RETRY_BACKOFF", 200*time.Millisecond)
	if err != nil {
		return Config{}, err
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

	redisDB, err := getEnvIntAllowZero("REDIS_DB", 0)
	if err != nil {
		return Config{}, err
	}

	redisDialTimeout, err := getEnvDuration("REDIS_DIAL_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}

	redisReadTimeout, err := getEnvDuration("REDIS_READ_TIMEOUT", 3*time.Second)
	if err != nil {
		return Config{}, err
	}

	redisWriteTimeout, err := getEnvDuration("REDIS_WRITE_TIMEOUT", 3*time.Second)
	if err != nil {
		return Config{}, err
	}

	deviceOnlineTTL, err := getEnvDuration("DEVICE_ONLINE_TTL", 120*time.Second)
	if err != nil {
		return Config{}, err
	}
	emqxPublishTimeout, err := getEnvDuration("EMQX_PUBLISH_TIMEOUT", 3*time.Second)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		KafkaBrokers:         kafkaBrokers,
		KafkaGroupID:         getEnv("KAFKA_GROUP_ID", "device-event-processor"),
		ConsumerWorkers:      consumerWorkers,
		ConsumerBuffer:       consumerBuffer,
		ConsumerMaxRetries:   consumerMaxRetries,
		ConsumerRetryBackoff: consumerRetryBackoff,

		PostgresDSN: getEnv("POSTGRES_DSN",
			"postgres://linkflow:linkflow123@127.0.0.1:5432/linkflow?sslmode=disable"),
		PostgresMaxOpenConns:    maxOpenConns,
		PostgresMaxIdleConns:    maxIdleConns,
		PostgresConnMaxLifetime: connMaxLifetime,
		PostgresConnMaxIdleTime: connMaxIdleTime,

		RedisAddr:         getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPassword:     os.Getenv("REDIS_PASSWORD"),
		RedisDB:           redisDB,
		RedisDialTimeout:  redisDialTimeout,
		RedisReadTimeout:  redisReadTimeout,
		RedisWriteTimeout: redisWriteTimeout,

		DeviceOnlineTTL: deviceOnlineTTL,

		EMQXAPIURL:         getEnv("EMQX_API_URL", "http://127.0.0.1:18083"),
		EMQXAPIKey:         getEnv("EMQX_API_KEY", "linkflow-init"),
		EMQXAPISecret:      getEnv("EMQX_API_SECRET", "linkflow-init-secret"),
		EMQXPublishTimeout: emqxPublishTimeout,

		ContractsDir: getEnv("CONTRACTS_DIR", "../../contracts"),
	}

	if cfg.PostgresDSN == "" {
		return Config{}, fmt.Errorf("POSTGRES_DSN is required")
	}
	if cfg.KafkaGroupID == "" {
		return Config{}, fmt.Errorf("KAFKA_GROUP_ID is required")
	}
	if cfg.RedisAddr == "" {
		return Config{}, fmt.Errorf("REDIS_ADDR is required")
	}
	if cfg.EMQXAPIURL == "" {
		return Config{}, fmt.Errorf("EMQX_API_URL is required")
	}
	if cfg.EMQXAPIKey == "" {
		return Config{}, fmt.Errorf("EMQX_API_KEY is required")
	}
	if cfg.EMQXAPISecret == "" {
		return Config{}, fmt.Errorf("EMQX_API_SECRET is required")
	}
	if cfg.ContractsDir == "" {
		return Config{}, fmt.Errorf("CONTRACTS_DIR is required")
	}

	return cfg, nil
}

func getEnvIntAllowZero(key string, def int) (int, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return def, nil
	}

	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	if v < 0 {
		return 0, fmt.Errorf("%s must be non-negative", key)
	}

	return v, nil
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
