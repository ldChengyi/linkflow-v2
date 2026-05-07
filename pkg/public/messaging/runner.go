package messaging

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

const (
	defaultWorkers = 4
	defaultBuffer  = 128
)

// Options controls Runner concurrency and backpressure.
type Options struct {
	Workers int
	Buffer  int
}

// Runner receives deliveries from a Source and processes them with a Handler.
type Runner struct {
	source  Source
	handler Handler
	workers int
	buffer  int
}

// NewRunner creates a Runner with bounded workers and a bounded delivery buffer.
func NewRunner(source Source, handler Handler, opt Options, mws ...Middleware) (*Runner, error) {
	if source == nil {
		return nil, errors.New("messaging source is nil")
	}
	if handler == nil {
		return nil, errors.New("messaging handler is nil")
	}
	if opt.Workers <= 0 {
		opt.Workers = defaultWorkers
	}
	if opt.Buffer <= 0 {
		opt.Buffer = defaultBuffer
	}

	return &Runner{
		source:  source,
		handler: Chain(handler, mws...),
		workers: opt.Workers,
		buffer:  opt.Buffer,
	}, nil
}

// Run starts the receive loop and worker pool until ctx is cancelled or an
// unrecoverable receive/completion error occurs.
func (r *Runner) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	deliveries := make(chan Delivery, r.buffer)
	workerErrs := make(chan error, 1)

	var wg sync.WaitGroup
	wg.Add(r.workers)
	for i := 0; i < r.workers; i++ {
		go func() {
			defer wg.Done()
			r.worker(ctx, deliveries, workerErrs, cancel)
		}()
	}

	receiveErr := r.receive(ctx, deliveries)
	close(deliveries)
	wg.Wait()

	var runErr error
	select {
	case err := <-workerErrs:
		runErr = err
	default:
		runErr = receiveErr
	}

	if err := r.source.Close(); err != nil {
		if runErr != nil {
			return errors.Join(runErr, fmt.Errorf("close source: %w", err))
		}
		return fmt.Errorf("close source: %w", err)
	}

	return runErr
}

func (r *Runner) receive(ctx context.Context, deliveries chan<- Delivery) error {
	for {
		delivery, err := r.source.Receive(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("receive delivery: %w", err)
		}
		if delivery == nil {
			return errors.New("received nil delivery")
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case deliveries <- delivery:
		}
	}
}

func (r *Runner) worker(
	ctx context.Context,
	deliveries <-chan Delivery,
	errs chan<- error,
	cancel context.CancelFunc,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case delivery, ok := <-deliveries:
			if !ok {
				return
			}
			if err := r.handle(ctx, delivery); err != nil {
				select {
				case errs <- err:
					cancel()
				default:
				}
				return
			}
		}
	}
}

func (r *Runner) handle(ctx context.Context, delivery Delivery) error {
	result := r.handler.Handle(ctx, delivery.Message())

	switch result.Decision {
	case DecisionAck:
		if err := delivery.Ack(ctx); err != nil {
			return fmt.Errorf("ack delivery: %w", err)
		}
	case DecisionDrop:
		if err := delivery.Drop(ctx); err != nil {
			return fmt.Errorf("drop delivery: %w", err)
		}
	case DecisionRetry:
		if err := delivery.Retry(ctx); err != nil {
			return fmt.Errorf("retry delivery: %w", err)
		}
	default:
		if result.Err != nil {
			return fmt.Errorf("handler returned no decision: %w", result.Err)
		}
		return fmt.Errorf("handler returned unknown decision %q", result.Decision)
	}

	return nil
}
