package realtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
	"github.com/ldchengyi/linkflow-v2/pkg/public/messaging"
	messagingkafka "github.com/ldchengyi/linkflow-v2/pkg/public/messaging/kafka"
)

// Consumer reads post-validation state events from the device.state topic and
// forwards them through the Registry to local WebSocket connections.
type Consumer struct {
	source   *messagingkafka.Source
	registry *Registry
	log      *slog.Logger
}

func NewConsumer(brokers []string, groupID string, registry *Registry, log *slog.Logger) (*Consumer, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers are required")
	}
	if groupID == "" {
		return nil, fmt.Errorf("kafka group id is required")
	}
	if registry == nil {
		return nil, fmt.Errorf("registry is nil")
	}
	if log == nil {
		log = slog.Default()
	}

	source, err := messagingkafka.NewSource(messagingkafka.Options{
		Brokers: brokers,
		Topic:   event.TopicDeviceStateV1,
		GroupID: groupID,
	})
	if err != nil {
		return nil, fmt.Errorf("create realtime source: %w", err)
	}

	return &Consumer{
		source:   source,
		registry: registry,
		log:      log,
	}, nil
}

// Run blocks until the context is cancelled or the source returns an
// unrecoverable error.
func (c *Consumer) Run(ctx context.Context) error {
	for {
		delivery, err := c.source.Receive(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		}
		c.handle(ctx, delivery)
	}
}

func (c *Consumer) Close() error {
	if c == nil || c.source == nil {
		return nil
	}
	return c.source.Close()
}

func (c *Consumer) handle(ctx context.Context, delivery messaging.Delivery) {
	msg := delivery.Message()

	var envelope event.Envelope
	if err := json.Unmarshal(msg.Value, &envelope); err != nil {
		c.log.Warn("realtime drop malformed envelope", "err", err)
		_ = delivery.Drop(ctx)
		return
	}

	if envelope.TenantID == "" {
		c.log.Warn("realtime drop envelope without tenant_id", "event_id", envelope.EventID)
		_ = delivery.Drop(ctx)
		return
	}

	c.registry.Broadcast(envelope.TenantID, msg.Value)
	_ = delivery.Ack(ctx)
}
