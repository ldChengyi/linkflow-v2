package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/config"
	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/envelope"
	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/kafka"
	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/mqtt"
	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/publisher"
	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/registry"
	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/router"
	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/util"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		log.Error("load config", "err", err)
		os.Exit(1)
	}

	kafkaClient, err := kafka.New(kafka.Options{
		Brokers: util.SpiltCSV(cfg.KafkaBrokers),
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
	eb := envelope.Builder{Producer: cfg.Producer, TenantID: cfg.TenantID}

	r := router.New(log)
	if err := registry.RegisterAll(r, eb, pub); err != nil {
		log.Error("register routes", "err", err)
		os.Exit(1)
	}

	cli := mqtt.New(mqtt.Options{
		BrokerURL: cfg.BrokerURL,
		ClientID:  cfg.ClientID,
		Username:  cfg.Username,
		Password:  cfg.Password,
	}, r, log)

	if err := cli.Connect(10 * time.Second); err != nil {
		log.Error("mqtt connect", "err", err)
		os.Exit(1)
	}
	defer cli.Disconnect()

	if err := cli.Subscribe("lf/v1/+/+/+/up/#", 1); err != nil {
		log.Error("subscribe", "err", err)
		os.Exit(1)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Info("shutting down")

}
