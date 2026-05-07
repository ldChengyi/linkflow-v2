package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event/validation"
	"github.com/ldchengyi/linkflow-v2/pkg/public/messaging"
	messagingkafka "github.com/ldchengyi/linkflow-v2/pkg/public/messaging/kafka"
	"github.com/ldchengyi/linkflow-v2/service/device-event-processor/internal/config"
	"github.com/ldchengyi/linkflow-v2/service/device-event-processor/internal/consumer"
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

	processorHandler, err := consumer.NewProcessorHandler(eventProcessor)
	if err != nil {
		return fmt.Errorf("create processor messaging handler: %w", err)
	}

	source, err := messagingkafka.NewSource(messagingkafka.Options{
		Brokers: cfg.KafkaBrokers,
		Topic:   event.TopicDeviceEventsV1,
		GroupID: cfg.KafkaGroupID,
	})
	if err != nil {
		return fmt.Errorf("create kafka source: %w", err)
	}

	runner, err := messaging.NewRunner(source, processorHandler, messaging.Options{
		Workers:      cfg.ConsumerWorkers,
		Buffer:       cfg.ConsumerBuffer,
		MaxRetries:   cfg.ConsumerMaxRetries,
		RetryBackoff: cfg.ConsumerRetryBackoff,
		Logger:       log,
	})
	if err != nil {
		return fmt.Errorf("create messaging runner: %w", err)
	}

	log.Info(
		"device event processor started",
		"kafka_topic", event.TopicDeviceEventsV1,
		"kafka_group_id", cfg.KafkaGroupID,
		"workers", cfg.ConsumerWorkers,
		"max_retries", cfg.ConsumerMaxRetries,
	)

	if err := runner.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("run messaging consumer: %w", err)
	}
	return nil
}
