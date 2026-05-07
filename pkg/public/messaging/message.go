package messaging

// Message is the broker-neutral message shape consumed by application handlers.
type Message struct {
	Key     []byte
	Value   []byte
	Headers map[string]string
}

// Decision tells the Runner how to finish a Delivery after the handler returns.
type Decision string

const (
	DecisionAck   Decision = "ack"
	DecisionDrop  Decision = "drop"
	DecisionRetry Decision = "retry"
)

// Result is returned by a Handler after it processes a Message.
type Result struct {
	Decision Decision
	Err      error
}
