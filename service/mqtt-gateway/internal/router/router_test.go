package router

import (
	"context"
	"errors"
	"log/slog"
	"testing"
)

func TestRouterDispatchMatchesRouteAndExtractsVars(t *testing.T) {
	r := New(slog.Default())

	called := false

	err := r.Handle(
		"device.property.post",
		`^lf/v1/(?P<product_key>[^/]+)/(?P<device_id>[^/]+)/property/up/post$`,
		func(ctx context.Context, msg ParsedMessage) error {
			called = true

			if msg.Topic != "lf/v1/esp32/dev-001/property/up/post" {
				t.Fatalf("Topic = %q", msg.Topic)
			}
			if string(msg.Payload) != `{"temperature":23.5}` {
				t.Fatalf("Payload = %s", string(msg.Payload))
			}
			if msg.Vars["product_key"] != "esp32" {
				t.Fatalf("product_key = %q", msg.Vars["product_key"])
			}
			if msg.Vars["device_id"] != "dev-001" {
				t.Fatalf("device_id = %q", msg.Vars["device_id"])
			}

			return nil
		},
	)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	r.Dispatch(
		context.Background(),
		"lf/v1/esp32/dev-001/property/up/post",
		[]byte(`{"temperature":23.5}`),
	)

	if !called {
		t.Fatal("handler was not called")
	}
}

func TestRouterDispatchDoesNotCallHandlerWhenNoRouteMatches(t *testing.T) {
	r := New(slog.Default())

	err := r.Handle(
		"device.property.post",
		`^lf/v1/(?P<product_key>[^/]+)/(?P<device_id>[^/]+)/property/up/post$`,
		func(ctx context.Context, msg ParsedMessage) error {
			t.Fatal("handler should not be called")
			return nil
		},
	)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	r.Dispatch(context.Background(), "lf/v1/esp32/dev-001/event/up/post", []byte(`{}
`))
}

func TestRouterDispatchDoesNotPanicWhenHandlerReturnsError(t *testing.T) {
	r := New(slog.Default())

	err := r.Handle(
		"device.property.post",
		`^lf/v1/(?P<product_key>[^/]+)/(?P<device_id>[^/]+)/property/up/post$`,
		func(ctx context.Context, msg ParsedMessage) error {
			return errors.New("handler failed")
		},
	)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	r.Dispatch(
		context.Background(),
		"lf/v1/esp32/dev-001/property/up/post",
		[]byte(`{"temperature":23.5}`),
	)
}

func TestRouterHandleReturnsCompileErrorForInvalidPattern(t *testing.T) {
	r := New(slog.Default())

	err := r.Handle("bad", `(`, func(ctx context.Context, msg ParsedMessage) error {
		return nil
	})
	if err == nil {
		t.Fatal("Handle() error is nil, want compile error")
	}
}
