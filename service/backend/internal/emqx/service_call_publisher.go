package emqx

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ldchengyi/linkflow-v2/service/backend/internal/service"
)

type ServiceCallPublisher struct {
	publisher *Publisher
}

func NewServiceCallPublisher(publisher *Publisher) (*ServiceCallPublisher, error) {
	if publisher == nil {
		return nil, fmt.Errorf("emqx publisher is nil")
	}
	return &ServiceCallPublisher{publisher: publisher}, nil
}

func (p *ServiceCallPublisher) PublishServiceCall(ctx context.Context, in service.DeviceServiceCallMessage) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	payload, err := json.Marshal(map[string]any{
		"command_id":   in.CommandID,
		"service_name": in.ServiceName,
		"input":        in.Input,
	})
	if err != nil {
		return fmt.Errorf("encode service call payload: %w", err)
	}
	return p.publisher.Publish(ctx, PublishInput{
		Topic:   in.Topic,
		Payload: payload,
		QOS:     1,
		Retain:  false,
	})
}
