package registry

import (
	"context"
	"log/slog"
	"testing"

	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/envelope"
	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/router"
)

type fakePublisher struct {
	events []envelope.Envelope
}

func (p *fakePublisher) Publish(ctx context.Context, e envelope.Envelope) error {
	p.events = append(p.events, e)
	return nil
}

func TestRegisterAllRegistersPropertyPostRoute(t *testing.T) {
	r := router.New(slog.Default())
	pub := &fakePublisher{}
	builder := envelope.Builder{
		Producer: "mqtt-gateway",
		TenantID: "default",
	}

	if err := RegisterAll(r, builder, pub); err != nil {
		t.Fatalf("RegisterAll() error = %v", err)
	}

	r.Dispatch(
		context.Background(),
		"lf/v1/esp32/dev-001/property/up/post",
		[]byte(`{"temperature":23.5}`),
	)
}
