package publisher

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
)

type fakeSink struct {
	topic   string
	key     []byte
	value   []byte
	headers map[string]string
	err     error
	calls   int
}

func (s *fakeSink) Send(ctx context.Context, topic string, key []byte, value []byte, headers map[string]string) error {
	s.calls++
	s.topic = topic
	s.key = key
	s.value = value
	s.headers = headers
	return s.err
}

func TestPublishWrapsPayloadInEnvelopeAndRoutesByTopic(t *testing.T) {
	sink := &fakeSink{}
	pub, err := NewEventPublisher(sink, "device-event-processor")
	if err != nil {
		t.Fatalf("NewEventPublisher() error = %v", err)
	}

	err = pub.Publish(context.Background(), PublishInput{
		EventType:    event.DeviceConnectionChanged.Type,
		EventVersion: event.DeviceConnectionChanged.Version,
		TenantID:     "tenant-1",
		Payload: event.ConnectionChangedPayload{
			TenantID: "tenant-1", ProductID: "product-1", DeviceID: "device-1",
			TenantSlug: "tenant", ProductKey: "product", DeviceSlug: "device",
			Status: event.ConnectionStatusOnline,
		},
		CausationID: "upstream-event-id",
	})
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	if sink.topic != event.TopicDeviceStateV1 {
		t.Fatalf("topic = %q, want %q", sink.topic, event.TopicDeviceStateV1)
	}
	if sink.headers["event_type"] != event.TypeDeviceConnectionChanged {
		t.Fatalf("event_type header = %q", sink.headers["event_type"])
	}
	if sink.headers["producer"] != "device-event-processor" {
		t.Fatalf("producer header = %q", sink.headers["producer"])
	}
	if sink.headers["tenant_id"] != "tenant-1" {
		t.Fatalf("tenant_id header = %q", sink.headers["tenant_id"])
	}

	var envelope event.Envelope
	if err := json.Unmarshal(sink.value, &envelope); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	if envelope.EventID == "" {
		t.Fatal("envelope event_id is empty")
	}
	if envelope.CausationID != "upstream-event-id" {
		t.Fatalf("causation_id = %q, want upstream-event-id", envelope.CausationID)
	}
	if string(sink.key) != envelope.EventID {
		t.Fatalf("kafka key = %q, want envelope event_id %q", sink.key, envelope.EventID)
	}
}

func TestPublishRejectsUnknownEvent(t *testing.T) {
	pub, _ := NewEventPublisher(&fakeSink{}, "device-event-processor")

	err := pub.Publish(context.Background(), PublishInput{
		EventType:    "no.such.event",
		EventVersion: 1,
		TenantID:     "tenant-1",
		Payload:      map[string]any{},
	})
	if err == nil {
		t.Fatal("Publish() error = nil, want unknown event error")
	}
}

func TestPublishPropagatesSinkError(t *testing.T) {
	sink := &fakeSink{err: errors.New("kafka down")}
	pub, _ := NewEventPublisher(sink, "device-event-processor")

	err := pub.Publish(context.Background(), PublishInput{
		EventType:    event.DevicePropertyChanged.Type,
		EventVersion: event.DevicePropertyChanged.Version,
		TenantID:     "tenant-1",
		Payload: event.PropertyChangedPayload{
			TenantID: "tenant-1", ProductID: "product-1", DeviceID: "device-1",
			TenantSlug: "tenant", ProductKey: "product", DeviceSlug: "device",
			Properties: map[string]any{"temperature": 23.5},
		},
	})
	if err == nil {
		t.Fatal("Publish() error = nil, want sink error")
	}
}
