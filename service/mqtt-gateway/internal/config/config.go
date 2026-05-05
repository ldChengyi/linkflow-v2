package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	BrokerURL string
	ClientID  string
	Username  string
	Password  string
	TenantID  string
	Producer  string

	KafkaBrokers []string

	MQTTMessageBuffer  int
	MQTTWorkerCount    int
	MQTTHandlerTimeout time.Duration
}

func Load() (Config, error) {
	messageBuffer, err := getEnvInt("MQTT_MESSAGE_BUFFER", 128)
	if err != nil {
		return Config{}, err
	}
	workerCount, err := getEnvInt("MQTT_WORKER_COUNT", 4)
	if err != nil {
		return Config{}, err
	}
	handlerTimeout, err := getEnvDuration("MQTT_HANDLER_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}
	kafkaBrokers := splitCSV(getEnv("KAFKA_BROKER_URL", "127.0.0.1:19092"))
	if len(kafkaBrokers) == 0 {
		return Config{}, fmt.Errorf("KAFKA_BROKER_URL is required")
	}

	cfg := Config{
		BrokerURL:          getEnv("MQTT_BROKER_URL", "tcp://127.0.0.1:1883"),
		ClientID:           getEnv("MQTT_CLIENT_ID", "linkflow-mqtt-gateway"),
		Username:           os.Getenv("MQTT_USERNAME"),
		Password:           os.Getenv("MQTT_PASSWORD"),
		TenantID:           getEnv("LF_TENANT_ID", "default"),
		Producer:           getEnv("LF_PRODUCER", "mqtt-gateway"),
		KafkaBrokers:       kafkaBrokers,
		MQTTMessageBuffer:  messageBuffer,
		MQTTWorkerCount:    workerCount,
		MQTTHandlerTimeout: handlerTimeout,
	}

	if cfg.BrokerURL == "" {
		return cfg, fmt.Errorf("MQTT_BROKER URL IS REQUIRED!")
	}
	return cfg, nil
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

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
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
