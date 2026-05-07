package messaging

import "context"

// Handler processes one broker-neutral Message.
type Handler interface {
	Handle(ctx context.Context, msg Message) Result
}

// HandlerFunc adapts a function to Handler.
type HandlerFunc func(ctx context.Context, msg Message) Result

func (f HandlerFunc) Handle(ctx context.Context, msg Message) Result {
	return f(ctx, msg)
}

// Middleware wraps a Handler with cross-cutting behavior.
type Middleware func(Handler) Handler

// Chain applies middleware around h. The first middleware is the outermost one.
func Chain(h Handler, mws ...Middleware) Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
