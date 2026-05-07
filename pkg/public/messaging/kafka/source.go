package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/ldchengyi/linkflow-v2/pkg/public/messaging"
	kafkago "github.com/segmentio/kafka-go"
)

const (
	defaultMinBytes = 1
	defaultMaxBytes = 10e6
	defaultMaxWait  = 10 * time.Second
)

// Options configures a Kafka-backed messaging source.
type Options struct {
	Brokers []string
	Topic   string
	GroupID string

	MinBytes int
	MaxBytes int
	MaxWait  time.Duration
}

// Source receives Kafka messages and exposes them as messaging deliveries.
type Source struct {
	reader *kafkago.Reader
}

// NewSource creates a Kafka source backed by kafka-go Reader.
func NewSource(opt Options) (*Source, error) {
	if len(opt.Brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers are required")
	}
	if opt.Topic == "" {
		return nil, fmt.Errorf("kafka topic is required")
	}
	if opt.GroupID == "" {
		return nil, fmt.Errorf("kafka group id is required")
	}
	if opt.MinBytes <= 0 {
		opt.MinBytes = defaultMinBytes
	}
	if opt.MaxBytes <= 0 {
		opt.MaxBytes = defaultMaxBytes
	}
	if opt.MaxWait <= 0 {
		opt.MaxWait = defaultMaxWait
	}

	return &Source{
		reader: kafkago.NewReader(kafkago.ReaderConfig{
			Brokers:  opt.Brokers,
			Topic:    opt.Topic,
			GroupID:  opt.GroupID,
			MinBytes: opt.MinBytes,
			MaxBytes: opt.MaxBytes,
			MaxWait:  opt.MaxWait,
		}),
	}, nil
}

// Receive fetches one Kafka message.
func (s *Source) Receive(ctx context.Context) (messaging.Delivery, error) {
	msg, err := s.reader.FetchMessage(ctx)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("fetch kafka message: %w", err)
	}
	return &Delivery{reader: s.reader, msg: msg}, nil
}

// Close closes the underlying Kafka reader.
func (s *Source) Close() error {
	if s == nil || s.reader == nil {
		return nil
	}
	return s.reader.Close()
}

// Delivery wraps one fetched Kafka message.
type Delivery struct {
	reader *kafkago.Reader
	msg    kafkago.Message
}

// Message converts the Kafka message into the broker-neutral shape.
func (d *Delivery) Message() messaging.Message {
	return messaging.Message{
		Key:     append([]byte(nil), d.msg.Key...),
		Value:   append([]byte(nil), d.msg.Value...),
		Headers: headersToMap(d.msg.Headers),
	}
}

// Ack commits the Kafka message.
func (d *Delivery) Ack(ctx context.Context) error {
	return d.commit(ctx)
}

// Drop commits the Kafka message so poison messages do not block the group.
func (d *Delivery) Drop(ctx context.Context) error {
	return d.commit(ctx)
}

// Retry leaves the Kafka message uncommitted for later redelivery.
func (d *Delivery) Retry(ctx context.Context) error {
	return ctx.Err()
}

func (d *Delivery) commit(ctx context.Context) error {
	if err := d.reader.CommitMessages(ctx, d.msg); err != nil {
		return fmt.Errorf("commit kafka message: %w", err)
	}
	return nil
}

func headersToMap(headers []kafkago.Header) map[string]string {
	if len(headers) == 0 {
		return nil
	}

	out := make(map[string]string, len(headers))
	for _, h := range headers {
		out[h.Key] = string(h.Value)
	}
	return out
}
