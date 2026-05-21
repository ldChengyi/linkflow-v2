package realtime

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/coder/websocket"
)

const (
	defaultWriteBuffer = 64
	defaultWriteWait   = 5 * time.Second
	defaultPingPeriod  = 30 * time.Second
)

// Conn wraps one client WebSocket and serializes writes through a buffered
// channel so Broadcast callers never block on a slow client.
type Conn struct {
	ws        *websocket.Conn
	tenantIDs []string
	userID    string
	write     chan []byte
	closeOnce sync.Once
	closed    chan struct{}
	log       *slog.Logger
}

func NewConn(ws *websocket.Conn, tenantIDs []string, userID string, log *slog.Logger) *Conn {
	if log == nil {
		log = slog.Default()
	}
	return &Conn{
		ws:        ws,
		tenantIDs: append([]string(nil), tenantIDs...),
		userID:    userID,
		write:     make(chan []byte, defaultWriteBuffer),
		closed:    make(chan struct{}),
		log:       log,
	}
}

// Send queues msg for delivery on this connection. Returns false if the
// connection's write buffer is full, in which case the caller drops the
// message and the slow connection is closed by the write loop.
func (c *Conn) Send(msg []byte) bool {
	select {
	case <-c.closed:
		return false
	case c.write <- msg:
		return true
	default:
		c.closeWithReason(websocket.StatusPolicyViolation, "write buffer full")
		return false
	}
}

// Run drives the connection until the peer closes, the context is cancelled,
// or a write fails. It blocks the calling goroutine (the HTTP handler).
func (c *Conn) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	readErr := make(chan error, 1)
	go c.readLoop(ctx, readErr)

	writeErr := c.writeLoop(ctx)

	select {
	case err := <-readErr:
		return err
	default:
		return writeErr
	}
}

func (c *Conn) readLoop(ctx context.Context, out chan<- error) {
	for {
		_, _, err := c.ws.Read(ctx)
		if err != nil {
			out <- err
			c.closeWithReason(websocket.StatusNormalClosure, "")
			return
		}
	}
}

func (c *Conn) writeLoop(ctx context.Context) error {
	ping := time.NewTicker(defaultPingPeriod)
	defer ping.Stop()

	for {
		select {
		case <-ctx.Done():
			c.closeWithReason(websocket.StatusGoingAway, "server shutdown")
			return ctx.Err()
		case <-c.closed:
			return errors.New("connection closed")
		case msg := <-c.write:
			writeCtx, cancel := context.WithTimeout(ctx, defaultWriteWait)
			err := c.ws.Write(writeCtx, websocket.MessageText, msg)
			cancel()
			if err != nil {
				c.closeWithReason(websocket.StatusInternalError, "write failed")
				return err
			}
		case <-ping.C:
			pingCtx, cancel := context.WithTimeout(ctx, defaultWriteWait)
			err := c.ws.Ping(pingCtx)
			cancel()
			if err != nil {
				c.closeWithReason(websocket.StatusInternalError, "ping failed")
				return err
			}
		}
	}
}

func (c *Conn) closeWithReason(code websocket.StatusCode, reason string) {
	c.closeOnce.Do(func() {
		close(c.closed)
		if c.ws != nil {
			_ = c.ws.Close(code, reason)
		}
	})
}
