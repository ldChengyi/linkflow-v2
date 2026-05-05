package publisher

import (
	"context"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
)

type Publisher interface {
	Publish(ctx context.Context, e event.Envelope) error
}
