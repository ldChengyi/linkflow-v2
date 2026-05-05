package registry

import (
	"context"
	"log/slog"
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
	r := router.New(slog.Default())
	pub := &fakePublisher{}
	factory := event.EnvelopeFactory{
		Producer: "mqtt-gateway",
		TenantID: "default",
	}

	if err := RegisterAll(r, factory, pub); err != nil {
		t.Fatalf("RegisterAll() error = %v", err)
	}

	r.Dispatch(
		context.Background(),
		"lf/v1/esp32/dev-001/property/up/post",
		[]byte(`{"temperature":23.5}`),
	)
}
