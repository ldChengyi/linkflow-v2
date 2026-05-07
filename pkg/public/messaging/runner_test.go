package messaging_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ldchengyi/linkflow-v2/pkg/public/messaging"
)

func TestRunnerCompletesDeliveryByDecision(t *testing.T) {
	tests := []struct {
		name      string
		decision  messaging.Decision
		wantAck   int
		wantDrop  int
		wantRetry int
	}{
		{name: "ack", decision: messaging.DecisionAck, wantAck: 1},
		{name: "drop", decision: messaging.DecisionDrop, wantDrop: 1},
		{name: "retry", decision: messaging.DecisionRetry, wantRetry: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := &fakeDelivery{
				msg: messaging.Message{
					Key:   []byte("k1"),
					Value: []byte(`{"ok":true}`),
				},
				done: make(chan struct{}),
			}
			source := newFakeSource(delivery)

			runner, err := messaging.NewRunner(
				source,
				messaging.HandlerFunc(func(_ context.Context, msg messaging.Message) messaging.Result {
					if string(msg.Key) != "k1" {
						t.Fatalf("message key = %q", msg.Key)
					}
					return messaging.Result{Decision: tt.decision}
				}),
				messaging.Options{Workers: 1, Buffer: 1, MaxRetries: 1, RetryBackoff: time.Millisecond},
			)
			if err != nil {
				t.Fatalf("NewRunner() error = %v", err)
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			errs := make(chan error, 1)
			go func() {
				errs <- runner.Run(ctx)
			}()

			select {
			case <-delivery.done:
				cancel()
			case <-time.After(time.Second):
				t.Fatal("delivery was not completed")
			}

			err = <-errs
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("Run() error = %v, want context canceled", err)
			}

			if delivery.acked != tt.wantAck {
				t.Fatalf("acked = %d, want %d", delivery.acked, tt.wantAck)
			}
			if delivery.dropped != tt.wantDrop {
				t.Fatalf("dropped = %d, want %d", delivery.dropped, tt.wantDrop)
			}
			if delivery.retried != tt.wantRetry {
				t.Fatalf("retried = %d, want %d", delivery.retried, tt.wantRetry)
			}
		})
	}
}

func TestRunnerRetriesBeforeCompletingDelivery(t *testing.T) {
	delivery := &fakeDelivery{
		msg:  messaging.Message{Key: []byte("k1")},
		done: make(chan struct{}),
	}
	source := newFakeSource(delivery)

	var calls int
	runner, err := messaging.NewRunner(
		source,
		messaging.HandlerFunc(func(context.Context, messaging.Message) messaging.Result {
			calls++
			return messaging.Result{Decision: messaging.DecisionRetry}
		}),
		messaging.Options{Workers: 1, Buffer: 1, MaxRetries: 2, RetryBackoff: time.Millisecond},
	)
	if err != nil {
		t.Fatalf("NewRunner() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errs := make(chan error, 1)
	go func() {
		errs <- runner.Run(ctx)
	}()

	select {
	case <-delivery.done:
		cancel()
	case <-time.After(time.Second):
		t.Fatal("delivery was not completed")
	}

	if calls != 3 {
		t.Fatalf("handler calls = %d, want 3", calls)
	}
	if delivery.retried != 1 {
		t.Fatalf("retried = %d, want 1", delivery.retried)
	}
}

func TestRunnerRecoversHandlerPanicAsRetry(t *testing.T) {
	delivery := &fakeDelivery{
		msg:  messaging.Message{Key: []byte("k1")},
		done: make(chan struct{}),
	}
	source := newFakeSource(delivery)

	runner, err := messaging.NewRunner(
		source,
		messaging.HandlerFunc(func(context.Context, messaging.Message) messaging.Result {
			panic("boom")
		}),
		messaging.Options{Workers: 1, Buffer: 1, MaxRetries: 1, RetryBackoff: time.Millisecond},
	)
	if err != nil {
		t.Fatalf("NewRunner() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errs := make(chan error, 1)
	go func() {
		errs <- runner.Run(ctx)
	}()

	select {
	case <-delivery.done:
		cancel()
	case <-time.After(time.Second):
		t.Fatal("delivery was not completed")
	}

	if delivery.retried != 1 {
		t.Fatalf("retried = %d, want 1", delivery.retried)
	}
}

func TestRunnerReturnsErrorForUnknownDecision(t *testing.T) {
	delivery := &fakeDelivery{
		msg:  messaging.Message{Key: []byte("k1")},
		done: make(chan struct{}),
	}
	source := newFakeSource(delivery)

	runner, err := messaging.NewRunner(
		source,
		messaging.HandlerFunc(func(context.Context, messaging.Message) messaging.Result {
			return messaging.Result{}
		}),
		messaging.Options{Workers: 1, Buffer: 1},
	)
	if err != nil {
		t.Fatalf("NewRunner() error = %v", err)
	}

	err = runner.Run(context.Background())
	if err == nil {
		t.Fatal("Run() error = nil, want unknown decision error")
	}
	if !strings.Contains(err.Error(), "unknown decision") {
		t.Fatalf("Run() error = %v, want unknown decision", err)
	}
}

type fakeSource struct {
	deliveries chan messaging.Delivery
}

func newFakeSource(deliveries ...messaging.Delivery) *fakeSource {
	ch := make(chan messaging.Delivery, len(deliveries))
	for _, delivery := range deliveries {
		ch <- delivery
	}
	return &fakeSource{deliveries: ch}
}

func (s *fakeSource) Receive(ctx context.Context) (messaging.Delivery, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case delivery := <-s.deliveries:
		return delivery, nil
	}
}

func (s *fakeSource) Close() error {
	return nil
}

type fakeDelivery struct {
	msg     messaging.Message
	done    chan struct{}
	acked   int
	dropped int
	retried int
}

func (d *fakeDelivery) Message() messaging.Message {
	return d.msg
}

func (d *fakeDelivery) Ack(context.Context) error {
	d.acked++
	close(d.done)
	return nil
}

func (d *fakeDelivery) Drop(context.Context) error {
	d.dropped++
	close(d.done)
	return nil
}

func (d *fakeDelivery) Retry(context.Context) error {
	d.retried++
	close(d.done)
	return nil
}
