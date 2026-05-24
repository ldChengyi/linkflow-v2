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

type MessageSink interface {
	Send(ctx context.Context, topic string, key []byte, value []byte, headers map[string]string) error
}

type ServiceCallPublisher struct {
	sink     MessageSink
	producer string
}

func NewServiceCallPublisher(sink MessageSink, producer string) (*ServiceCallPublisher, error) {
	if sink == nil {
		return nil, fmt.Errorf("kafka sink is nil")
	}
	if producer == "" {
		return nil, fmt.Errorf("producer is required")
	}
	return &ServiceCallPublisher{sink: sink, producer: producer}, nil
}

func (p *ServiceCallPublisher) PublishServiceCall(ctx context.Context, in service.DeviceServiceCallMessage) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if in.CommandID == "" {
		return fmt.Errorf("service call command_id is required")
	}
	if in.TenantID == "" {
		return fmt.Errorf("service call tenant_id is required")
	}
	if in.Input == nil {
		in.Input = map[string]any{}
	}

	payload := event.ServiceCallRequestedPayload{
		CommandID:   in.CommandID,
		TenantID:    in.TenantID,
		ProductID:   in.ProductID,
		DeviceID:    in.DeviceID,
		TenantSlug:  in.TenantSlug,
		ProductKey:  in.ProductKey,
		DeviceSlug:  in.DeviceSlug,
		Protocol:    in.Protocol,
		Topic:       in.Topic,
		ServiceName: in.ServiceName,
		Input:       in.Input,
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
	envelope, err := factory.New(event.TypeDeviceServiceCallRequested, event.VersionDeviceServiceCallRequested, payload, occurredAt)
	if err != nil {
		return fmt.Errorf("create service call requested envelope: %w", err)
	}
	envelope.CorrelationID = in.CommandID

	raw, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("encode service call requested envelope: %w", err)
	}

	return p.sink.Send(ctx, event.TopicDeviceEventsV1, []byte(in.CommandID), raw, map[string]string{
		"event_type":    envelope.EventType,
		"event_version": strconv.Itoa(envelope.EventVersion),
		"producer":      envelope.Producer,
		"tenant_id":     envelope.TenantID,
		"command_id":    in.CommandID,
	})
}
