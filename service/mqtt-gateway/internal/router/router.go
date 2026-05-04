package router

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
)

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
type Handler func(ctx context.Context, msg ParsedMessage) error

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
	log    *slog.Logger
}

// New creates an empty Router that uses log for route and handler diagnostics.
func New(log *slog.Logger) *Router {
	return &Router{log: log}
}

// Handle registers a new MQTT topic route.
//
// The pattern argument must be a valid Go regular expression. Named capture
// groups, such as (?P<deviceID>[^/]+), are exposed to the handler through
// ParsedMessage.Vars.
//
// warning: must be finished before Dispatch function call. not goroutine-safe
func (r *Router) Handle(name, pattern string, h Handler) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("router : compile %q: %w", name, err)
	}
	r.routes = append(r.routes, route{name: name, pattern: re, handler: h})
	return nil
}

// Dispatch routes one MQTT message to the first handler whose pattern matches
// topic.
//
// If the matched pattern contains named capture groups, Dispatch copies those
// captured values into ParsedMessage.Vars before invoking the handler. Handler
// errors are logged and not returned because MQTT message delivery is normally
// handled asynchronously by the broker client.
func (r *Router) Dispatch(ctx context.Context, topic string, payload []byte) {
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
		if err := rt.handler(ctx, msg); err != nil {
			r.log.Error("handler failed", "router", rt.name, "topic", topic, "err", err)
		}
		return
	}
	r.log.Warn("no route matched", "topic", topic)
}
