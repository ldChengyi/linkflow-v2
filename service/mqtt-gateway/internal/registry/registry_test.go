package registry

import (
	"context"
	"testing"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/router"
)

type fakePublisher struct {
	events []event.Envelope
}

func (p *fakePublisher) Publish(ctx context.Context, e event.Envelope) error {
	p.events = append(p.events, e)
	return nil
}

func TestRegisterAllRegistersPropertyPostRoute(t *testing.T) {
	r := router.New()
	pub := &fakePublisher{}
	factory := event.EnvelopeFactory{
		Producer: "mqtt-gateway",
		TenantID: "default",
	}

	subscriptions, err := RegisterAll(r, factory, pub)
	if err != nil {
		t.Fatalf("RegisterAll() error = %v", err)
	}
	if len(subscriptions) != 2 {
		t.Fatalf("subscriptions = %d", len(subscriptions))
	}
	if subscriptions[0].Topic != "lf/v1/+/+/property/up/post" {
		t.Fatalf("subscription topic = %q", subscriptions[0].Topic)
	}
	if subscriptions[1].Topic != "lf/v1/+/+/property/up/set_reply" {
		t.Fatalf("subscription topic = %q", subscriptions[1].Topic)
	}

	if err := r.Dispatch(
		context.Background(),
		"lf/v1/esp32/dev-001/property/up/post",
		[]byte(`{"temperature":23.5}`),
	); err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}
}

func TestRoutesUsePublicEventSpec(t *testing.T) {
	pub := &fakePublisher{}
	factory := event.EnvelopeFactory{
		Producer: "mqtt-gateway",
		TenantID: "default",
	}

	routes := Routes(factory, pub)
	if len(routes) != 2 {
		t.Fatalf("routes = %d", len(routes))
	}
	if routes[0].Event != event.DeviceTelemetryReceived {
		t.Fatalf("route event = %#v", routes[0].Event)
	}
	if routes[1].Event != event.DevicePropertySetAcknowledged {
		t.Fatalf("route event = %#v", routes[1].Event)
	}
}
