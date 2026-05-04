package envelope

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Envelope struct {
	EventID       string `json:"event_id"`
	EventType     string `json:"event_type"`
	EventVersion  int `json:"event_version"`
	OccurredAt    string `json:"occurred_at"`
	Producer      string `json:"producer"`
	TenantID      string `json:"tenant_id"`
	TraceID       string `json:"trace_id,omitempty"`
	CorrelationID string `json:"correlation_id",omitempty`
	CausationID   string `json:"causation_id", omitempty`
	Payload       json.RawMessage `json:"payload"`
}

type Builder struct {
	Producer string
	TenantID string
}

func (b Builder) New(eventType string, version int, payload any, occurredAt time.Time) (Envelope, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return Envelope{}, err
	}
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	
	id, _ := uuid.NewV7()
	return Envelope{ 
		EventID: id.String(),
		EventType: eventType,
		EventVersion: version,
		OccurredAt: occurredAt.UTC().Format(time.RFC3339Nano),
		Producer: b.Producer,
		TenantID: b.TenantID,
		Payload: raw,
	}, nil
}


















