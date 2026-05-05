package router

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"time"
)

func Recover(log *slog.Logger) Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, msg ParsedMessage) (err error) {
			defer func() {
				if r := recover(); r != nil {
					log.Error("mqtt handler panic", "topic", msg.Topic, "panic", r, "stack", string(debug.Stack()))
					err = fmt.Errorf("mqtt handler panic: %v", r)
				}
			}()
			return next.Handle(ctx, msg)
		})
	}
}

func Logging(log *slog.Logger) Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, msg ParsedMessage) error {
			start := time.Now()
			err := next.Handle(ctx, msg)
			attrs := []any{
				"topic", msg.Topic,
				"duration", time.Since(start),
			}
			if err != nil {
				attrs = append(attrs, "err", err)
				log.Error("mqtt message handled", attrs...)
				return err
			}
			log.Info("mqtt message handled", attrs...)
			return nil
		})
	}
}
