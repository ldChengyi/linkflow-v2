package consumer

import (
	"context"
	"errors"
	"testing"

	"github.com/ldchengyi/linkflow-v2/pkg/public/messaging"
	"github.com/ldchengyi/linkflow-v2/service/device-event-processor/internal/processor"
)

func TestProcessorHandlerReturnsProcessorDecision(t *testing.T) {
	tests := []struct {
		name     string
		decision messaging.Decision
	}{
		{name: "ack", decision: messaging.DecisionAck},
		{name: "drop", decision: messaging.DecisionDrop},
		{name: "retry", decision: messaging.DecisionRetry},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantErr := errors.New("processor error")
			handler, err := NewProcessorHandler(fakeProcessor{
				result: processor.Result{
					Decision: tt.decision,
					Err:      wantErr,
				},
			})
			if err != nil {
				t.Fatalf("NewProcessorHandler() error = %v", err)
			}

			got := handler.Handle(context.Background(), messaging.Message{
				Key:     []byte("k1"),
				Value:   []byte("v1"),
				Headers: map[string]string{"event_type": "device.property.reported"},
			})
			if got.Decision != tt.decision {
				t.Fatalf("decision = %q, want %q", got.Decision, tt.decision)
			}
			if !errors.Is(got.Err, wantErr) {
				t.Fatalf("err = %v, want %v", got.Err, wantErr)
			}
		})
	}
}

func TestNewProcessorHandlerRejectsNilProcessor(t *testing.T) {
	if _, err := NewProcessorHandler(nil); err == nil {
		t.Fatal("NewProcessorHandler() error = nil, want nil processor error")
	}
}

type fakeProcessor struct {
	result processor.Result
}

func (p fakeProcessor) Process(context.Context, processor.Message) processor.Result {
	return p.result
}
