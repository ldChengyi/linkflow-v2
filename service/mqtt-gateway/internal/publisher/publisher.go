package publisher

import (
	"context"

	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/envelope"
)

type Publisher interface {
	Publish(ctx context.Context, e envelope.Envelope) error
}

