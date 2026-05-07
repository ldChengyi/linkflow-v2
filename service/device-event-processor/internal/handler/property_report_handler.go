package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
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
			Action: processor.ActionRetry,
			Err:    err,
		}
	}
	if env == nil {
		return processor.Result{
			Action: processor.ActionDrop,
			Err:    fmt.Errorf("event envelope is nil"),
		}
	}

	result := processor.Result{
		EventID:  env.EventID,
		TenantID: env.TenantID,
	}

	var payload event.PropertyReportedPayload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		result.Action = processor.ActionDrop
		result.Err = fmt.Errorf("decode property reported payload: %w", err)
		return result
	}

	result.DeviceID = payload.DeviceID
	if err := h.writer.SavePropertyReport(ctx, env, payload); err != nil {
		result.Action = processor.ActionRetry
		result.Err = fmt.Errorf("save property report: %w", err)
		return result
	}

	result.Action = processor.ActionAck
	return result
}
