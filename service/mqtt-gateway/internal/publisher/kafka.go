package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/envelope"
	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/kafka"
)


type KafkaEvent struct {
	client *kafka.Client
}

func NewKafkaEvent(client *kafka.Client) *KafkaEvent {
	return &KafkaEvent{client: client}
}

func topicForEvent(e envelope.Envelope) (string, error) {
	switch {
     	case strings.HasPrefix(e.EventType, "device."):
			return "lf.v1.device.events", nil
		default:
			return "", fmt.Errorf("unknown event domain for event_type %q", e.EventType)
	}
}

func (p *KafkaEvent) Publish(ctx context.Context, e envelope.Envelope) error {
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
			"event_type" : e.EventType,
			"event_version" : strconv.Itoa(e.EventVersion),
			"producer": e.Producer,
			"tenant_id": e.TenantID,
		},
	)
}
