package mqtt

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
)

type Dispatcher interface {
	Dispatch(ctx context.Context, topic string, payload []byte)
}

type Client struct {
	c          paho.Client
	dispatcher Dispatcher
	log        *slog.Logger
}

type Options struct {
	BrokerURL string
	ClientID  string
	Username  string
	Password  string
}

func New(opt Options, d Dispatcher, log *slog.Logger) *Client {
	c := &Client{dispatcher: d, log: log}

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

func (c *Client) Connect(timeout time.Duration) error {
	t := c.c.Connect()
	if !t.WaitTimeout(timeout) {
		return fmt.Errorf("mqtt connect timeout after %s", timeout)
	}
	return t.Error()
}

func (c *Client) Subscribe(topic string, qos byte) error {
	t := c.c.Subscribe(topic, qos, func(_ paho.Client, m paho.Message) {
		c.dispatcher.Dispatch(context.Background(), m.Topic(), m.Payload())
	})
	t.Wait()
	if err := t.Error(); err != nil {
		return fmt.Errorf("subscribe %q: %w", topic, err)
	}
	c.log.Info("subscribed", "topic", topic, "qos", qos)
	return nil
}

func (c *Client) Disconnect() {
	c.c.Disconnect(500)
}

func (c *Client) onConnect(_ paho.Client) {
	c.log.Info("mqtt connected")
}
