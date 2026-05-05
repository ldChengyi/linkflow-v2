package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/publisher"
	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/router"
)

func Property(ef event.EnvelopeFactory, pub publisher.Publisher) router.Handler {
	return func(ctx context.Context, msg router.ParsedMessage) error {
		var metrics map[string]any
		if err := json.Unmarshal(msg.Payload, &metrics); err != nil {
			return fmt.Errorf("decode metrics: %w", err)
		}
		p := event.TelemetryReceivedPayload{
			DeviceID:   msg.Vars["device_id"],
			ProductKey: msg.Vars["product_key"],
			Protocol:   "mqtt",
			Metrics:    metrics,
		}
		env, err := ef.New(
			event.DeviceTelemetryReceived.Type,
			event.DeviceTelemetryReceived.Version,
			p,
			time.Time{},
		)
		if err != nil {
			return err
		}
		return pub.Publish(ctx, env)
	}
}
