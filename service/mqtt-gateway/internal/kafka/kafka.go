package kafka

import (
	"context"
	"fmt"
	"log/slog"

	kafkago "github.com/segmentio/kafka-go"
)

type Options struct {
	Brokers []string
}

type Client struct {
	writer *kafkago.Writer
	log    *slog.Logger
}

func New(opt Options, log *slog.Logger) (*Client, error) {
	if len(opt.Brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers are required")
	}

	w := &kafkago.Writer{
		Addr:     kafkago.TCP(opt.Brokers...),
		Balancer: &kafkago.Hash{},
	}

	return &Client{
		writer: w,
		log:    log,
	}, nil

}

func (c *Client) Send(
	ctx context.Context,
	topic string,
	key []byte,
	value []byte,
	headers map[string]string,
) error {
	if topic == "" {
		return fmt.Errorf("kafka topic is required")
	}

	kafkaHeaders := make([]kafkago.Header, 0, len(headers))
	for k, v := range headers {
		kafkaHeaders = append(kafkaHeaders, kafkago.Header{
			Key:   k,
			Value: []byte(v),
		})
	}

	if err := c.writer.WriteMessages(ctx, kafkago.Message{
		Topic:   topic,
		Key:     key,
		Value:   value,
		Headers: kafkaHeaders,
	}); err != nil {
		return fmt.Errorf("write kafka message to topic %q: %w", topic, err)
	}
	c.log.Info("kafka message sent", "topic", topic)
	return nil
}

func (c *Client) Close() error {
	if c.writer == nil {
		return nil
	}
	return c.writer.Close()
}
