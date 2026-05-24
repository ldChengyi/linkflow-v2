package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/service"
)

type PropertySetPublisher struct {
	sink     MessageSink
	producer string
}

func NewPropertySetPublisher(sink MessageSink, producer string) (*PropertySetPublisher, error) {
	if sink == nil {
		return nil, fmt.Errorf("kafka sink is nil")
	}
	if producer == "" {
		return nil, fmt.Errorf("producer is required")
	}
	return &PropertySetPublisher{sink: sink, producer: producer}, nil
}

func (p *PropertySetPublisher) PublishPropertySet(ctx context.Context, in service.DevicePropertySetMessage) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if in.CommandID == "" {
		return fmt.Errorf("property set command_id is required")
	}
	if in.TenantID == "" {
		return fmt.Errorf("property set tenant_id is required")
	}
	if len(in.Properties) == 0 {
		return fmt.Errorf("property set properties are required")
	}

	payload := event.PropertySetRequestedPayload{
		CommandID:   in.CommandID,
		TenantID:    in.TenantID,
		ProductID:   in.ProductID,
		DeviceID:    in.DeviceID,
		TenantSlug:  in.TenantSlug,
		ProductKey:  in.ProductKey,
		DeviceSlug:  in.DeviceSlug,
		Protocol:    in.Protocol,
		Topic:       in.Topic,
		Properties:  in.Properties,
		RequestedBy: in.RequestedBy,
	}

	occurredAt := in.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}

	factory := event.EnvelopeFactory{
		Producer: p.producer,
		TenantID: in.TenantID,
	}
	envelope, err := factory.New(event.TypeDevicePropertySetRequested, event.VersionDevicePropertySetRequested, payload, occurredAt)
	if err != nil {
		return fmt.Errorf("create property set requested envelope: %w", err)
	}
	envelope.CorrelationID = in.CommandID

	raw, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("encode property set requested envelope: %w", err)
	}

	return p.sink.Send(ctx, event.TopicDeviceEventsV1, []byte(in.CommandID), raw, map[string]string{
		"event_type":    envelope.EventType,
		"event_version": strconv.Itoa(envelope.EventVersion),
		"producer":      envelope.Producer,
		"tenant_id":     envelope.TenantID,
		"command_id":    in.CommandID,
	})
}
