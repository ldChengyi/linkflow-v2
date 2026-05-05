package publisher

import (
	"testing"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
)

func TestTopicForEventRoutesDeviceEvents(t *testing.T) {
	topic, err := topicForEvent(event.Envelope{
		EventType:    event.DeviceTelemetryReceived.Type,
		EventVersion: event.DeviceTelemetryReceived.Version,
	})
	if err != nil {
		t.Fatalf("topicForEvent() error = %v", err)
	}

	if topic != event.DeviceTelemetryReceived.Topic {
		t.Fatalf("topic = %q", topic)
	}
}

func TestTopicForEventReturnsErrorForUnknownEvent(t *testing.T) {
	_, err := topicForEvent(event.Envelope{
		EventType:    "tenant.created",
		EventVersion: 1,
	})
	if err == nil {
		t.Fatal("topicForEvent() error is nil, want unknown event error")
	}
}
