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

// DevicePropertyPost handles device-to-cloud property post messages.
//
// MQTT topic shape:
//
//	lf/v1/{tenant_slug}/{product_key}/{device_slug}/property/up/post
func DevicePropertyPost(spec event.Spec, ef event.EnvelopeFactory, pub publisher.Publisher) router.Handler {
	return router.HandlerFunc(func(ctx context.Context, msg router.ParsedMessage) error {
		var properties map[string]any
		if err := json.Unmarshal(msg.Payload, &properties); err != nil {
			return fmt.Errorf("decode properties: %w", err)
		}
		p := event.PropertyReportedPayload{
			TenantSlug: msg.Vars["tenant_slug"],
			DeviceSlug: msg.Vars["device_slug"],
			ProductKey: msg.Vars["product_key"],
			Protocol:   "mqtt",
			Properties: properties,
		}
		env, err := ef.New(
			spec.Type,
			spec.Version,
			p,
			time.Time{},
		)
		if err != nil {
			return err
		}
		return pub.Publish(ctx, env)
	})
}
