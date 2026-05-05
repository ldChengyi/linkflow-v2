package mqtt

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
)

type Dispatcher interface {
	Dispatch(ctx context.Context, topic string, payload []byte) error
}

type Client struct {
	c              paho.Client
	dispatcher     Dispatcher
	messages       chan Message
	workerCount    int
	handlerTimeout time.Duration
	log            *slog.Logger
}

type Options struct {
	BrokerURL      string
	ClientID       string
	Username       string
	Password       string
	MessageBuffer  int
	WorkerCount    int
	HandlerTimeout time.Duration
}

type Message struct {
	Topic   string
	Payload []byte
}

func New(opt Options, d Dispatcher, log *slog.Logger) *Client {
	if opt.MessageBuffer <= 0 {
		opt.MessageBuffer = 128
	}
	if opt.WorkerCount <= 0 {
		opt.WorkerCount = 4
	}
	if opt.HandlerTimeout <= 0 {
		opt.HandlerTimeout = 5 * time.Second
	}

	c := &Client{
		dispatcher:     d,
		messages:       make(chan Message, opt.MessageBuffer),
		workerCount:    opt.WorkerCount,
		handlerTimeout: opt.HandlerTimeout,
		log:            log,
	}

	o := paho.NewClientOptions().
		AddBroker(opt.BrokerURL).
		SetClientID(opt.ClientID).
		SetUsername(opt.Username).
		SetPassword(opt.Password).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(2 * time.Second).
		SetCleanSession(false).
		SetOrderMatters(false).
		SetOnConnectHandler(c.onConnect).
		SetConnectionLostHandler(func(_ paho.Client, err error) {
			log.Warn("mqtt connection lost", "err", err)
		})

	c.c = paho.NewClient(o)
	return c
}

func (c *Client) Connect(ctx context.Context, timeout time.Duration) error {
	t := c.c.Connect()
	return waitToken(ctx, t, timeout, "mqtt connect")
}

func (c *Client) Subscribe(ctx context.Context, topic string, qos byte) error {
	t := c.c.Subscribe(topic, qos, func(_ paho.Client, m paho.Message) {
		msg := Message{
			Topic:   m.Topic(),
			Payload: append([]byte(nil), m.Payload()...),
		}
		select {
		case c.messages <- msg:
		default:
			c.log.Error("mqtt message dropped: worker queue full", "topic", msg.Topic)
		}
	})
	if err := waitToken(ctx, t, 10*time.Second, "mqtt subscribe"); err != nil {
		return fmt.Errorf("subscribe %q: %w", topic, err)
	}
	c.log.Info("subscribed", "topic", topic, "qos", qos)
	return nil
}

func (c *Client) Run(ctx context.Context) {
	for i := 0; i < c.workerCount; i++ {
		go c.worker(ctx, i)
	}
}

func (c *Client) Disconnect() {
	c.c.Disconnect(500)
}

func (c *Client) onConnect(_ paho.Client) {
	c.log.Info("mqtt connected")
}

func (c *Client) worker(ctx context.Context, id int) {
	c.log.Info("mqtt worker started", "worker_id", id)
	for {
		select {
		case <-ctx.Done():
			c.log.Info("mqtt worker stopped", "worker_id", id)
			return
		case msg := <-c.messages:
			c.handle(ctx, msg)
		}
	}
}

func (c *Client) handle(parent context.Context, msg Message) {
	ctx, cancel := context.WithTimeout(parent, c.handlerTimeout)
	defer cancel()

	if err := c.dispatcher.Dispatch(ctx, msg.Topic, msg.Payload); err != nil {
		c.log.Error("mqtt dispatch failed", "topic", msg.Topic, "err", err)
	}
}

func waitToken(ctx context.Context, t paho.Token, timeout time.Duration, op string) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return fmt.Errorf("%s timeout after %s", op, timeout)
	case <-t.Done():
		return t.Error()
	}
}
