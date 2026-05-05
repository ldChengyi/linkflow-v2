package router

import (
	"context"
	"errors"
	"testing"
)

func TestRouterDispatchMatchesRouteAndExtractsVars(t *testing.T) {
	r := New()

	called := false

	err := r.Handle(
		"device.property.post",
		`^lf/v1/(?P<product_key>[^/]+)/(?P<device_id>[^/]+)/property/up/post$`,
		HandlerFunc(func(ctx context.Context, msg ParsedMessage) error {
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
		}),
	)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if err := r.Dispatch(
		context.Background(),
		"lf/v1/esp32/dev-001/property/up/post",
		[]byte(`{"temperature":23.5}`),
	); err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}

	if !called {
		t.Fatal("handler was not called")
	}
}

func TestRouterDispatchDoesNotCallHandlerWhenNoRouteMatches(t *testing.T) {
	r := New()

	err := r.Handle(
		"device.property.post",
		`^lf/v1/(?P<product_key>[^/]+)/(?P<device_id>[^/]+)/property/up/post$`,
		HandlerFunc(func(ctx context.Context, msg ParsedMessage) error {
			t.Fatal("handler should not be called")
			return nil
		}),
	)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if err := r.Dispatch(context.Background(), "lf/v1/esp32/dev-001/event/up/post", []byte(`{}
`)); err == nil {
		t.Fatal("Dispatch() error is nil, want no route error")
	}
}

func TestRouterDispatchDoesNotPanicWhenHandlerReturnsError(t *testing.T) {
	r := New()

	err := r.Handle(
		"device.property.post",
		`^lf/v1/(?P<product_key>[^/]+)/(?P<device_id>[^/]+)/property/up/post$`,
		HandlerFunc(func(ctx context.Context, msg ParsedMessage) error {
			return errors.New("handler failed")
		}),
	)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if err := r.Dispatch(
		context.Background(),
		"lf/v1/esp32/dev-001/property/up/post",
		[]byte(`{"temperature":23.5}`),
	); err == nil {
		t.Fatal("Dispatch() error is nil, want handler error")
	}
}

func TestRouterHandleReturnsCompileErrorForInvalidPattern(t *testing.T) {
	r := New()

	err := r.Handle("bad", `(`, HandlerFunc(func(ctx context.Context, msg ParsedMessage) error {
		return nil
	}))
	if err == nil {
		t.Fatal("Handle() error is nil, want compile error")
	}
}
