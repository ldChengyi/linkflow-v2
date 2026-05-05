package router

import (
	"context"
	"errors"
	"fmt"
	"regexp"
)

var ErrNoRoute = errors.New("no route matched")

// ParsedMessage is the normalized MQTT message passed to a route handler.
//
// Topic keeps the original MQTT topic that was received, Vars contains named
// values extracted from the matched route pattern, and Payload contains the raw
// MQTT message body.
type ParsedMessage struct {
	Topic   string
	Vars    map[string]string
	Payload []byte
}

// Handler processes a ParsedMessage after a route pattern matches its topic.
//
// The context allows callers to propagate cancellation, deadlines, and request
// scoped values through the MQTT dispatch pipeline.
type Handler interface {
	Handle(ctx context.Context, msg ParsedMessage) error
}

// HandlerFunc adapts a function to Handler.
type HandlerFunc func(ctx context.Context, msg ParsedMessage) error

func (f HandlerFunc) Handle(ctx context.Context, msg ParsedMessage) error {
	return f(ctx, msg)
}

// Middleware wraps a Handler with cross-cutting behavior such as logging,
// metrics, recover, timeout, or retry.
type Middleware func(Handler) Handler

// Chain applies middleware around h. The first middleware is the outermost one.
func Chain(h Handler, mws ...Middleware) Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// route stores one registered topic matcher and the handler that should process
// messages matching that pattern.
type route struct {
	name    string
	pattern *regexp.Regexp
	handler Handler
}

// Router dispatches MQTT messages to the first registered route whose regular
// expression matches the message topic.
//
// Route order is significant: Dispatch stops after the first match, so register
// more specific patterns before broader fallback patterns.
type Router struct {
	routes []route
}

// New creates an empty Router.
func New() *Router {
	return &Router{}
}

// Handle registers a new MQTT topic route.
//
// The pattern argument must be a valid Go regular expression. Named capture
// groups, such as (?P<deviceID>[^/]+), are exposed to the handler through
// ParsedMessage.Vars.
//
// warning: must be finished before Dispatch function call. not goroutine-safe
func (r *Router) Handle(name, pattern string, h Handler, mws ...Middleware) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("router : compile %q: %w", name, err)
	}
	r.routes = append(r.routes, route{name: name, pattern: re, handler: Chain(h, mws...)})
	return nil
}

// Dispatch routes one MQTT message to the first handler whose pattern matches
// topic.
//
// If the matched pattern contains named capture groups, Dispatch copies those
// captured values into ParsedMessage.Vars before invoking the handler. Handler
// errors are returned so the caller can decide whether to retry, drop, or
// dead-letter the message.
func (r *Router) Dispatch(ctx context.Context, topic string, payload []byte) error {
	for _, rt := range r.routes {
		// FindStringSubmatch returns the full match plus capture-group values.
		// A nil result means this route does not match the MQTT topic.
		m := rt.pattern.FindStringSubmatch(topic)
		if m == nil {
			continue
		}

		// Convert named regexp capture groups into a stable key/value map for
		// handler code. Unnamed capture groups are intentionally ignored.
		vars := make(map[string]string, len(rt.pattern.SubexpNames()))
		for i, n := range rt.pattern.SubexpNames() {
			if i == 0 || n == "" {
				continue
			}
			vars[n] = m[i]
		}

		msg := ParsedMessage{Topic: topic, Vars: vars, Payload: payload}
		if err := rt.handler.Handle(ctx, msg); err != nil {
			return fmt.Errorf("handle route %q: %w", rt.name, err)
		}
		return nil
	}
	return fmt.Errorf("%w: %s", ErrNoRoute, topic)
}
