package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/envelope"
	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/publisher"
	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/router"
)

type telemetryPayload struct {
	DeviceID   string         `json:"device_id"`
	ProductKey string         `json:"product_key"`
	Protocol   string         `json:"protocol"`
	Metrics    map[string]any `json:"metrics"`
}

func Property(b envelope.Builder, pub publisher.Publisher) router.Handler {
	return func(ctx context.Context, msg router.ParsedMessage) error {
		var metrics map[string]any
		if err := json.Unmarshal(msg.Payload, &metrics); err != nil {
			return fmt.Errorf("decode metrics: %w", err)
		}
		p := telemetryPayload{
			DeviceID:   msg.Vars["device_id"],
			ProductKey: msg.Vars["product_key"],
			Protocol:   "mqtt",
			Metrics:    metrics,
		}
		env, err := b.New("device.telemetry.received", 1, p, time.Time{})
		if err != nil {
			return err
		}
		return pub.Publish(ctx, env)
	}
}
