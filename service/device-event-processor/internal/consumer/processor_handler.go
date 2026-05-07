package consumer

import (
	"context"
	"fmt"

	"github.com/ldchengyi/linkflow-v2/pkg/public/messaging"
	"github.com/ldchengyi/linkflow-v2/service/device-event-processor/internal/processor"
)

// ProcessorHandler adapts the service event processor to the public messaging
// Handler interface.
type ProcessorHandler struct {
	processor processor.Processor
}

// NewProcessorHandler creates a messaging handler backed by the service
// processor.
func NewProcessorHandler(processor processor.Processor) (*ProcessorHandler, error) {
	if processor == nil {
		return nil, fmt.Errorf("processor is nil")
	}
	return &ProcessorHandler{processor: processor}, nil
}

// Handle maps public messaging messages and decisions to the service processor.
func (h *ProcessorHandler) Handle(ctx context.Context, msg messaging.Message) messaging.Result {
	result := h.processor.Process(ctx, processor.Message{
		Key:     msg.Key,
		Value:   msg.Value,
		Headers: msg.Headers,
	})

	return messaging.Result{
		Decision: result.Decision,
		Err:      result.Err,
	}
}
