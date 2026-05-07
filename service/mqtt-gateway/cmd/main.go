package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ldchengyi/linkflow-v2/pkg/public/contracts/event"
	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/config"
	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/kafka"
	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/mqtt"
	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/publisher"
	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/registry"
	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/router"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Error("load config", "err", err)
		os.Exit(1)
	}

	kafkaClient, err := kafka.New(kafka.Options{
		Brokers: cfg.KafkaBrokers,
	}, log)
	if err != nil {
		log.Error("create kafka client", "err", err)
		os.Exit(1)
	}
	defer func() {
		if err := kafkaClient.Close(); err != nil {
			log.Error("close kafka client", "err", err)
		}
	}()

	pub := publisher.NewKafkaEvent(kafkaClient)
	ef := event.EnvelopeFactory{Producer: cfg.Producer, TenantID: cfg.TenantID}

	r := router.New()
	subscriptions, err := registry.RegisterAll(r, ef, pub,
		router.Recover(log),
		router.Logging(log),
	)
	if err != nil {
		log.Error("register routes", "err", err)
		os.Exit(1)
	}

	cli := mqtt.New(mqtt.Options{
		BrokerURL:      cfg.BrokerURL,
		ClientID:       cfg.ClientID,
		Username:       cfg.Username,
		Password:       cfg.Password,
		CleanSession:   cfg.CleanSession,
		MessageBuffer:  cfg.MQTTMessageBuffer,
		WorkerCount:    cfg.MQTTWorkerCount,
		HandlerTimeout: cfg.MQTTHandlerTimeout,
	}, r, log)
	cli.Run(ctx)

	if err := cli.Connect(ctx, 10*time.Second); err != nil {
		log.Error("mqtt connect", "err", err)
		os.Exit(1)
	}
	defer cli.Disconnect()

	for _, sub := range subscriptions {
		if err := cli.Subscribe(ctx, sub.Topic, sub.QOS); err != nil {
			log.Error("subscribe", "topic", sub.Topic, "err", err)
			os.Exit(1)
		}
	}

	<-ctx.Done()
	log.Info("shutting down")
}
