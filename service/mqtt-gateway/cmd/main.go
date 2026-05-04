package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/config"
	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/envelope"
	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/handler"
	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/mqtt"
	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/publisher"
	"github.com/ldchengyi/linkflow-v2/service/mqtt-gateway/internal/router"
)

func mustRoute(r *router.Router, name, pattern string, h router.Handler) {
	if err := r.Handle(name, pattern, h); err != nil {
		panic(err)
	}
}

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		log.Error("load config", "err", err)
		os.Exit(1)
	}

	pub := publisher.Stdout{Log: log}
	eb := envelope.Builder{Producer: cfg.Producer, TenantID: cfg.TenantID}

	r := router.New(log)
	mustRoute(r, "device.property", `^lf/v1/(?P<product_key>[^/]+)/(?P<device_name>[^/]+)/property/up/post$`,
		handler.Property(eb, pub),
	)

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
