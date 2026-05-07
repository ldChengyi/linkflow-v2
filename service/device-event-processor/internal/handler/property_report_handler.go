package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
	"github.com/ldchengyi/linkflow-v2/pkg/public/messaging"
	"github.com/ldchengyi/linkflow-v2/service/device-event-processor/internal/processor"
)

type PropertyReportWriter interface {
	SavePropertyReport(ctx context.Context, env *event.Envelope, payload event.PropertyReportedPayload) error
}

type PropertyReportHandler struct {
	writer PropertyReportWriter
}

func NewPropertyReportHandler(writer PropertyReportWriter) (*PropertyReportHandler, error) {
	if writer == nil {
		return nil, fmt.Errorf("property report writer is nil")
	}

	return &PropertyReportHandler{
		writer: writer,
	}, nil
}

func (h *PropertyReportHandler) Handle(ctx context.Context, env *event.Envelope) processor.Result {
	if err := ctx.Err(); err != nil {
		return processor.Result{
			Decision: messaging.DecisionRetry,
			Err:      err,
		}
	}
	if env == nil {
		return processor.Result{
			Decision: messaging.DecisionDrop,
			Err:      fmt.Errorf("event envelope is nil"),
		}
	}

	result := processor.Result{
		EventID:  env.EventID,
		TenantID: env.TenantID,
	}

	var payload event.PropertyReportedPayload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		result.Decision = messaging.DecisionDrop
		result.Err = fmt.Errorf("decode property reported payload: %w", err)
		return result
	}

	result.DeviceID = payload.DeviceID
	if err := h.writer.SavePropertyReport(ctx, env, payload); err != nil {
		result.Decision = messaging.DecisionRetry
		result.Err = fmt.Errorf("save property report: %w", err)
		return result
	}

	result.Decision = messaging.DecisionAck
	return result
}
