package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultAuthTokenIssuer       = "linkflow-backend"
	defaultAuthAccessTokenSecret = "dev-only-linkflow-access-token-secret-change-before-production"
)

type Config struct {
	HTTPAddr           string
	HTTPReadTimeout    time.Duration
	HTTPWriteTimeout   time.Duration
	HTTPIdleTimeout    time.Duration
	HTTPRequestTimeout time.Duration

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

	AuthAccessTokenTTL    time.Duration
	AuthAccessTokenSecret string
	AuthTokenIssuer       string
	AuthBCryptCost        int

	MQTTGatewayUsername string
	MQTTGatewayPassword string

	KafkaBrokers       []string
	RealtimeGroupID    string
	RealtimeAccessName string
}

func Load() (Config, error) {
	httpReadTimeout, err := getEnvDuration("HTTP_READ_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}

	httpWriteTimeout, err := getEnvDuration("HTTP_WRITE_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}

	httpIdleTimeout, err := getEnvDuration("HTTP_IDLE_TIMEOUT", 60*time.Second)
	if err != nil {
		return Config{}, err
	}

	httpRequestTimeout, err := getEnvDuration("HTTP_REQUEST_TIMEOUT", 5*time.Second)
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

	accessTokenTTL, err := getEnvDuration("AUTH_ACCESS_TOKEN_TTL", 24*time.Hour)
	if err != nil {
		return Config{}, err
	}

	bcryptCost, err := getEnvInt("AUTH_BCRYPT_COST", 12)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		HTTPAddr:           getEnv("HTTP_ADDR", ":18080"),
		HTTPReadTimeout:    httpReadTimeout,
		HTTPWriteTimeout:   httpWriteTimeout,
		HTTPIdleTimeout:    httpIdleTimeout,
		HTTPRequestTimeout: httpRequestTimeout,

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

		AuthAccessTokenTTL:    accessTokenTTL,
		AuthAccessTokenSecret: getEnv("AUTH_ACCESS_TOKEN_SECRET", defaultAuthAccessTokenSecret),
		AuthTokenIssuer:       getEnv("AUTH_TOKEN_ISSUER", defaultAuthTokenIssuer),
		AuthBCryptCost:        bcryptCost,
		MQTTGatewayUsername:   getEnv("MQTT_GATEWAY_USERNAME", "linkflow-mqtt-gateway"),
		MQTTGatewayPassword:   getEnv("MQTT_GATEWAY_PASSWORD", "linkflow-mqtt-gateway-secret"),

		KafkaBrokers:       splitCSV(getEnv("KAFKA_BROKERS", "127.0.0.1:19092")),
		RealtimeGroupID:    getEnv("REALTIME_GROUP_ID", ""),
		RealtimeAccessName: getEnv("REALTIME_ACCESS_COOKIE", "lf_access"),
	}

	if cfg.HTTPAddr == "" {
		return Config{}, fmt.Errorf("HTTP_ADDR is required")
	}
	if cfg.PostgresDSN == "" {
		return Config{}, fmt.Errorf("POSTGRES_DSN is required")
	}
	if cfg.RedisAddr == "" {
		return Config{}, fmt.Errorf("REDIS_ADDR is required")
	}
	if cfg.AuthBCryptCost < 4 || cfg.AuthBCryptCost > 31 {
		return Config{}, fmt.Errorf("AUTH_BCRYPT_COST must be between 4 and 31")
	}
	if cfg.AuthAccessTokenSecret == "" {
		return Config{}, fmt.Errorf("AUTH_ACCESS_TOKEN_SECRET is required")
	}
	if cfg.AuthTokenIssuer == "" {
		return Config{}, fmt.Errorf("AUTH_TOKEN_ISSUER is required")
	}
	if cfg.MQTTGatewayUsername == "" {
		return Config{}, fmt.Errorf("MQTT_GATEWAY_USERNAME is required")
	}
	if cfg.MQTTGatewayPassword == "" {
		return Config{}, fmt.Errorf("MQTT_GATEWAY_PASSWORD is required")
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
		return 0, fmt.Errorf("%s must be zero or positive", key)
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
