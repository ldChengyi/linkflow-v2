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

type propertySetReplyMessage struct {
	Success    *bool          `json:"success"`
	Code       string         `json:"code"`
	Message    string         `json:"message"`
	Properties map[string]any `json:"properties"`
}

// DevicePropertySetReply handles device-to-cloud property set acknowledgements.
//
// MQTT topic shape:
//
//	lf/v1/{tenant_slug}/{product_key}/{device_slug}/property/up/set_reply
func DevicePropertySetReply(spec event.Spec, ef event.EnvelopeFactory, pub publisher.Publisher) router.Handler {
	return router.HandlerFunc(func(ctx context.Context, msg router.ParsedMessage) error {
		var in propertySetReplyMessage
		if err := json.Unmarshal(msg.Payload, &in); err != nil {
			return fmt.Errorf("decode property set reply: %w", err)
		}
		if in.Success == nil {
			return fmt.Errorf("property set reply missing success")
		}

		raw := make(map[string]any)
		if err := json.Unmarshal(msg.Payload, &raw); err != nil {
			return fmt.Errorf("decode raw property set reply: %w", err)
		}

		p := event.PropertySetAcknowledgedPayload{
			TenantSlug: msg.Vars["tenant_slug"],
			DeviceSlug: msg.Vars["device_slug"],
			ProductKey: msg.Vars["product_key"],
			Protocol:   "mqtt",
			Success:    *in.Success,
			Code:       in.Code,
			Message:    in.Message,
			Properties: in.Properties,
			Raw:        raw,
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
