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
	"github.com/ldchengyi/linkflow-v2/pkg/public/postgres"
	publicredis "github.com/ldchengyi/linkflow-v2/pkg/public/redis"
	"github.com/ldchengyi/linkflow-v2/service/device-event-processor/internal/config"
	"github.com/ldchengyi/linkflow-v2/service/device-event-processor/internal/consumer"
	"github.com/ldchengyi/linkflow-v2/service/device-event-processor/internal/handler"
	"github.com/ldchengyi/linkflow-v2/service/device-event-processor/internal/mqtt"
	"github.com/ldchengyi/linkflow-v2/service/device-event-processor/internal/processor"
	"github.com/ldchengyi/linkflow-v2/service/device-event-processor/internal/publisher"
	"github.com/ldchengyi/linkflow-v2/service/device-event-processor/internal/store"
	"github.com/ldchengyi/linkflow-v2/service/device-event-processor/internal/validator"
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

	redisClient, err := publicredis.Open(ctx, publicredis.Config{
		Addr:         cfg.RedisAddr,
		Password:     cfg.RedisPassword,
		DB:           cfg.RedisDB,
		DialTimeout:  cfg.RedisDialTimeout,
		ReadTimeout:  cfg.RedisReadTimeout,
		WriteTimeout: cfg.RedisWriteTimeout,
	})
	if err != nil {
		return fmt.Errorf("open redis: %w", err)
	}
	defer func() {
		_ = redisClient.Close()
	}()

	eventValidator, err := validation.NewValidator(os.DirFS(cfg.ContractsDir))
	if err != nil {
		return fmt.Errorf("create event validator: %w", err)
	}

	eventProcessor, err := processor.NewEventProcessor(eventValidator)
	if err != nil {
		return fmt.Errorf("create event processor: %w", err)
	}

	eventSink, err := messagingkafka.NewSink(messagingkafka.SinkOptions{
		Brokers: cfg.KafkaBrokers,
	})
	if err != nil {
		return fmt.Errorf("create kafka sink: %w", err)
	}
	defer func() {
		_ = eventSink.Close()
	}()

	eventPublisher, err := publisher.NewEventPublisher(eventSink, "device-event-processor")
	if err != nil {
		return fmt.Errorf("create event publisher: %w", err)
	}

	propertyReportStore, err := store.NewPropertyReportStore(pool)
	if err != nil {
		return fmt.Errorf("create property report store: %w", err)
	}
	propertySetStore, err := store.NewPropertySetStore(pool)
	if err != nil {
		return fmt.Errorf("create property set store: %w", err)
	}
	eventReportStore, err := store.NewEventReportStore(pool)
	if err != nil {
		return fmt.Errorf("create event report store: %w", err)
	}
	serviceCallStore, err := store.NewServiceCallStore(pool)
	if err != nil {
		return fmt.Errorf("create service call store: %w", err)
	}
	mqttPublisher, err := mqtt.NewPublisher(mqtt.PublisherOptions{
		BaseURL:   cfg.EMQXAPIURL,
		APIKey:    cfg.EMQXAPIKey,
		APISecret: cfg.EMQXAPISecret,
		Timeout:   cfg.EMQXPublishTimeout,
	})
	if err != nil {
		return fmt.Errorf("create mqtt publisher: %w", err)
	}

	thingsModelReader, err := store.NewThingsModelReader(pool)
	if err != nil {
		return fmt.Errorf("create thingsmodel reader: %w", err)
	}

	propertyValidator, err := validator.NewPropertyReportValidator(thingsModelReader, validator.DefaultCacheTTL)
	if err != nil {
		return fmt.Errorf("create property report validator: %w", err)
	}
	eventReportValidator, err := validator.NewEventReportValidator(thingsModelReader, validator.DefaultCacheTTL)
	if err != nil {
		return fmt.Errorf("create event report validator: %w", err)
	}
	serviceCallValidator, err := validator.NewServiceCallValidator(thingsModelReader, validator.DefaultCacheTTL)
	if err != nil {
		return fmt.Errorf("create service call validator: %w", err)
	}

	propertyReportHandler, err := handler.NewPropertyReportHandler(propertyReportStore, propertyValidator, eventPublisher, log)
	if err != nil {
		return fmt.Errorf("create property report handler: %w", err)
	}
	propertySetHandler, err := handler.NewPropertySetHandler(propertySetStore, log)
	if err != nil {
		return fmt.Errorf("create property set handler: %w", err)
	}
	propertySetRequestedHandler, err := handler.NewPropertySetRequestedHandler(propertySetStore, mqttPublisher, log)
	if err != nil {
		return fmt.Errorf("create property set requested handler: %w", err)
	}
	eventReportHandler, err := handler.NewEventReportHandler(eventReportStore, eventReportValidator, eventPublisher, log)
	if err != nil {
		return fmt.Errorf("create event report handler: %w", err)
	}
	serviceCallHandler, err := handler.NewServiceCallHandler(serviceCallStore, serviceCallValidator, log)
	if err != nil {
		return fmt.Errorf("create service call handler: %w", err)
	}
	serviceCallRequestedHandler, err := handler.NewServiceCallRequestedHandler(serviceCallStore, mqttPublisher, log)
	if err != nil {
		return fmt.Errorf("create service call requested handler: %w", err)
	}

	if err := eventProcessor.Register(event.DevicePropertyReported, propertyReportHandler); err != nil {
		return fmt.Errorf("register property reported handler: %w", err)
	}
	if err := eventProcessor.Register(event.DevicePropertySetRequested, propertySetRequestedHandler); err != nil {
		return fmt.Errorf("register property set requested handler: %w", err)
	}
	if err := eventProcessor.Register(event.DevicePropertySetAcknowledged, propertySetHandler); err != nil {
		return fmt.Errorf("register property set acknowledged handler: %w", err)
	}
	if err := eventProcessor.Register(event.DeviceEventReported, eventReportHandler); err != nil {
		return fmt.Errorf("register event reported handler: %w", err)
	}
	if err := eventProcessor.Register(event.DeviceServiceCallRequested, serviceCallRequestedHandler); err != nil {
		return fmt.Errorf("register service call requested handler: %w", err)
	}
	if err := eventProcessor.Register(event.DeviceServiceCallAcknowledged, serviceCallHandler); err != nil {
		return fmt.Errorf("register service call acknowledged handler: %w", err)
	}

	deviceConnectionStore, err := store.NewDeviceConnectionStore(pool, redisClient, cfg.DeviceOnlineTTL)
	if err != nil {
		return fmt.Errorf("create device connection store: %w", err)
	}

	deviceConnectedHandler, err := handler.NewDeviceConnectedHandler(deviceConnectionStore, eventPublisher, log)
	if err != nil {
		return fmt.Errorf("create device connected handler: %w", err)
	}

	deviceDisconnectedHandler, err := handler.NewDeviceDisconnectedHandler(deviceConnectionStore, eventPublisher, log)
	if err != nil {
		return fmt.Errorf("create device disconnected handler: %w", err)
	}

	if err := eventProcessor.Register(event.DeviceConnected, deviceConnectedHandler); err != nil {
		return fmt.Errorf("register device connected handler: %w", err)
	}
	if err := eventProcessor.Register(event.DeviceDisconnected, deviceDisconnectedHandler); err != nil {
		return fmt.Errorf("register device disconnected handler: %w", err)
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
