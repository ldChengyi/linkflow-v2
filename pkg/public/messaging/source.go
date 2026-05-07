package messaging

import "context"

// Delivery is one received message plus the broker-specific completion actions.
type Delivery interface {
	Message() Message
	Ack(ctx context.Context) error
	Drop(ctx context.Context) error
	Retry(ctx context.Context) error
}

// Source receives broker messages and returns them as broker-neutral deliveries.
type Source interface {
	Receive(ctx context.Context) (Delivery, error)
	Close() error
}
