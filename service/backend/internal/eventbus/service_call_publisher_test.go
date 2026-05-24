package eventbus

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event/validation"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/service"
)

type fakeMessageSink struct {
	topic   string
	key     []byte
	value   []byte
	headers map[string]string
}

func (s *fakeMessageSink) Send(ctx context.Context, topic string, key []byte, value []byte, headers map[string]string) error {
	s.topic = topic
	s.key = append([]byte(nil), key...)
	s.value = append([]byte(nil), value...)
	s.headers = headers
	return nil
}

func TestServiceCallPublisherPublishesContractEvent(t *testing.T) {
	sink := &fakeMessageSink{}
	pub, err := NewServiceCallPublisher(sink, "backend")
	if err != nil {
		t.Fatalf("NewServiceCallPublisher() error = %v", err)
	}

	err = pub.PublishServiceCall(context.Background(), service.DeviceServiceCallMessage{
		CommandID:   "018f56d3-7cb7-7f1a-9b41-3f3a63fd3db9",
		TenantID:    "tenant-1",
		ProductID:   "product-1",
		DeviceID:    "device-1",
		TenantSlug:  "default",
		ProductKey:  "esp32",
		DeviceSlug:  "dev-1",
		Protocol:    "mqtt",
		Topic:       "lf/v1/default/esp32/dev-1/service/down/reboot",
		ServiceName: "reboot",
		RequestedBy: "user-1",
		Input:       map[string]any{"delay": float64(5)},
		OccurredAt:  time.Date(2026, 5, 5, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("PublishServiceCall() error = %v", err)
	}

	if sink.topic != event.TopicDeviceEventsV1 {
		t.Fatalf("topic = %q, want %q", sink.topic, event.TopicDeviceEventsV1)
	}
	if string(sink.key) != "018f56d3-7cb7-7f1a-9b41-3f3a63fd3db9" {
		t.Fatalf("key = %q, want command id", string(sink.key))
	}
	if sink.headers["event_type"] != event.TypeDeviceServiceCallRequested {
		t.Fatalf("event_type header = %q", sink.headers["event_type"])
	}

	validator, err := validation.NewValidator(os.DirFS("../../../../contracts"))
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}
	env, err := validator.ValidateEvent(sink.value)
	if err != nil {
		t.Fatalf("ValidateEvent() error = %v", err)
	}
	if env.EventType != event.TypeDeviceServiceCallRequested {
		t.Fatalf("EventType = %q", env.EventType)
	}
	if env.CorrelationID != "018f56d3-7cb7-7f1a-9b41-3f3a63fd3db9" {
		t.Fatalf("CorrelationID = %q, want command id", env.CorrelationID)
	}
}

func TestPropertySetPublisherPublishesContractEvent(t *testing.T) {
	sink := &fakeMessageSink{}
	pub, err := NewPropertySetPublisher(sink, "backend")
	if err != nil {
		t.Fatalf("NewPropertySetPublisher() error = %v", err)
	}

	err = pub.PublishPropertySet(context.Background(), service.DevicePropertySetMessage{
		CommandID:   "018f56d3-7cb7-7f1a-9b41-3f3a63fd3db9",
		TenantID:    "tenant-1",
		ProductID:   "product-1",
		DeviceID:    "device-1",
		TenantSlug:  "default",
		ProductKey:  "esp32",
		DeviceSlug:  "dev-1",
		Protocol:    "mqtt",
		Topic:       "lf/v1/default/esp32/dev-1/property/down/set",
		RequestedBy: "user-1",
		Properties:  map[string]any{"led1": true},
		OccurredAt:  time.Date(2026, 5, 5, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("PublishPropertySet() error = %v", err)
	}

	if sink.topic != event.TopicDeviceEventsV1 {
		t.Fatalf("topic = %q, want %q", sink.topic, event.TopicDeviceEventsV1)
	}
	if string(sink.key) != "018f56d3-7cb7-7f1a-9b41-3f3a63fd3db9" {
		t.Fatalf("key = %q, want command id", string(sink.key))
	}
	if sink.headers["event_type"] != event.TypeDevicePropertySetRequested {
		t.Fatalf("event_type header = %q", sink.headers["event_type"])
	}

	validator, err := validation.NewValidator(os.DirFS("../../../../contracts"))
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}
	env, err := validator.ValidateEvent(sink.value)
	if err != nil {
		t.Fatalf("ValidateEvent() error = %v", err)
	}
	if env.EventType != event.TypeDevicePropertySetRequested {
		t.Fatalf("EventType = %q", env.EventType)
	}
	if env.CorrelationID != "018f56d3-7cb7-7f1a-9b41-3f3a63fd3db9" {
		t.Fatalf("CorrelationID = %q, want command id", env.CorrelationID)
	}
}
