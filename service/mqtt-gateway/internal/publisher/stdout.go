package publisher

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
)

type Stdout struct {
	Log *slog.Logger
}

func (s Stdout) Publish(_ context.Context, e event.Envelope) error {
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	s.Log.Info("event", "envelope", json.RawMessage(b))
	return nil
}
