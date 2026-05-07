package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event/validation"
	"github.com/ldchengyi/linkflow-v2/service/device-event-processor/internal/config"
	"github.com/ldchengyi/linkflow-v2/service/device-event-processor/internal/handler"
	"github.com/ldchengyi/linkflow-v2/service/device-event-processor/internal/postgres"
	"github.com/ldchengyi/linkflow-v2/service/device-event-processor/internal/processor"
	"github.com/ldchengyi/linkflow-v2/service/device-event-processor/internal/store"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, log); err != nil {
		log.Error("device event processor stopped", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, log *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	pool, err := postgres.Open(ctx, postgres.Config{
		DSN: cfg.PostgresDSN,
	})
	if err != nil {
		return fmt.Errorf("open postgres: %w", err)
	}
	defer pool.Close()

	validator, err := validation.NewValidator(os.DirFS(cfg.ContractsDir))
	if err != nil {
		return fmt.Errorf("create event validator: %w", err)
	}

	eventProcessor, err := processor.NewEventProcessor(validator)
	if err != nil {
		return fmt.Errorf("create event processor: %w", err)
	}

	propertyReportStore, err := store.NewPropertyReportStore(pool)
	if err != nil {
		return fmt.Errorf("create property report store: %w", err)
	}

	propertyReportHandler, err := handler.NewPropertyReportHandler(propertyReportStore)
	if err != nil {
		return fmt.Errorf("create property report handler: %w", err)
	}

	if err := eventProcessor.Register(event.DevicePropertyReported, propertyReportHandler); err != nil {
		return fmt.Errorf("register property reported handler: %w", err)
	}

	log.Info("device event processor wiring ready")

	<-ctx.Done()
	return ctx.Err()
}
