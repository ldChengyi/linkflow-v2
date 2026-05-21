package kafka

import (
	"context"
	"fmt"

	kafkago "github.com/segmentio/kafka-go"
)

// SinkOptions configures a Kafka-backed message sink.
type SinkOptions struct {
	Brokers []string
}

// Sink is a producer-side counterpart to Source. It accepts broker-neutral
// (topic, key, value, headers) tuples and forwards them to Kafka.
type Sink struct {
	writer *kafkago.Writer
}

// NewSink creates a Kafka sink backed by kafka-go Writer.
func NewSink(opt SinkOptions) (*Sink, error) {
	if len(opt.Brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers are required")
	}
	return &Sink{
		writer: &kafkago.Writer{
			Addr:     kafkago.TCP(opt.Brokers...),
			Balancer: &kafkago.Hash{},
		},
	}, nil
}

// Send publishes one message to the given topic.
func (s *Sink) Send(ctx context.Context, topic string, key []byte, value []byte, headers map[string]string) error {
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

	if err := s.writer.WriteMessages(ctx, kafkago.Message{
		Topic:   topic,
		Key:     key,
		Value:   value,
		Headers: kafkaHeaders,
	}); err != nil {
		return fmt.Errorf("write kafka message to topic %q: %w", topic, err)
	}
	return nil
}

// Close closes the underlying Kafka writer.
func (s *Sink) Close() error {
	if s == nil || s.writer == nil {
		return nil
	}
	return s.writer.Close()
}
