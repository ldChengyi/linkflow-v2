package config

import (
	"fmt"
	"os"
)

type Config struct {
	BrokerURL string
	ClientID  string
	Username  string
	Password  string
	TenantID  string
	Producer  string

	KafkaBrokers string
}

func Load() (Config, error) {
	cfg := Config{
		BrokerURL:    getEnv("MQTT_BROKER_URL", "tcp://127.0.0.1:1883"),
		ClientID:     getEnv("MQTT_CLIENT_ID", "linkflow-mqtt-gateway"),
		Username:     os.Getenv("MQTT_USERNAME"),
		Password:     os.Getenv("MQTT_PASSWORD"),
		TenantID:     getEnv("LF_TENANT_ID", "default"),
		Producer:     getEnv("LF_PRODUCER", "mqtt-gateway"),
		KafkaBrokers: getEnv("KAFKA_BROKER_URL", "127.0.0.1:19092"),
	}

	if cfg.BrokerURL == "" {
		return cfg, fmt.Errorf("MQTT_BROKER URL IS REQUIRED!")
	}
	return cfg, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
