package publisher

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/envelope"
)

type Stdout struct {
	Log *slog.Logger
}

func (s Stdout) Publish(_ context.Context, e envelope.Envelope) error {
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	s.Log.Info("event", "envelope", json.RawMessage(b))
	return nil
}
