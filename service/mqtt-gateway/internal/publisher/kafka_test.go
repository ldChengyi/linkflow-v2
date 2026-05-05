package publisher

import (
	"testing"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
)

func TestTopicForEventRoutesDeviceEvents(t *testing.T) {
	topic, err := topicForEvent(event.Envelope{
		EventType: "device.telemetry.received",
	})
	if err != nil {
		t.Fatalf("topicForEvent() error = %v", err)
	}

	if topic != "lf.v1.device.events" {
		t.Fatalf("topic = %q", topic)
	}
}

func TestTopicForEventReturnsErrorForUnknownDomain(t *testing.T) {
	_, err := topicForEvent(event.Envelope{
		EventType: "tenant.created",
	})
	if err == nil {
		t.Fatal("topicForEvent() error is nil, want unknown domain error")
	}
}
