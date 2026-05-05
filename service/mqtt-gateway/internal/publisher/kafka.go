package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/kafka"
)

type KafkaEvent struct {
	client *kafka.Client
}

func NewKafkaEvent(client *kafka.Client) *KafkaEvent {
	return &KafkaEvent{client: client}
}

func topicForEvent(e event.Envelope) (string, error) {
	spec, ok := event.Lookup(e.EventType, e.EventVersion)
	if !ok {
		return "", fmt.Errorf("unknown event %s v%d", e.EventType, e.EventVersion)
	}
	return spec.Topic, nil
}

func (p *KafkaEvent) Publish(ctx context.Context, e event.Envelope) error {
	b, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("marshal envelope: %w", err)
	}

	topic, err := topicForEvent(e)
	if err != nil {
		return err
	}

	return p.client.Send(
		ctx,
		topic,
		[]byte(e.EventID),
		b,
		map[string]string{
			"event_type":    e.EventType,
			"event_version": strconv.Itoa(e.EventVersion),
			"producer":      e.Producer,
			"tenant_id":     e.TenantID,
		},
	)
}
