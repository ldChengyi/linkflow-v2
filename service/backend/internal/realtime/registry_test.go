package realtime

import (
	"testing"
)

func newTestConn(tenantIDs []string) *Conn {
	return &Conn{
		tenantIDs: tenantIDs,
		write:     make(chan []byte, 4),
		closed:    make(chan struct{}),
	}
}

func TestRegistryBroadcastDeliversToMatchingTenant(t *testing.T) {
	r := NewRegistry()
	a := newTestConn([]string{"tenant-1"})
	b := newTestConn([]string{"tenant-1", "tenant-2"})
	other := newTestConn([]string{"tenant-3"})

	r.Register(a)
	r.Register(b)
	r.Register(other)

	delivered := r.Broadcast("tenant-1", []byte("evt"))
	if delivered != 2 {
		t.Fatalf("delivered = %d, want 2", delivered)
	}

	for _, c := range []*Conn{a, b} {
		select {
		case msg := <-c.write:
			if string(msg) != "evt" {
				t.Fatalf("delivered = %q", msg)
			}
		default:
			t.Fatal("expected message in conn write buffer")
		}
	}

	select {
	case <-other.write:
		t.Fatal("other tenant should not receive event")
	default:
	}
}

func TestRegistryUnregisterRemovesFromAllTenants(t *testing.T) {
	r := NewRegistry()
	c := newTestConn([]string{"tenant-1", "tenant-2"})

	r.Register(c)
	r.Unregister(c)

	if delivered := r.Broadcast("tenant-1", []byte("x")); delivered != 0 {
		t.Fatalf("delivered = %d, want 0 after unregister", delivered)
	}
	if delivered := r.Broadcast("tenant-2", []byte("x")); delivered != 0 {
		t.Fatalf("delivered = %d, want 0 after unregister", delivered)
	}
}

func TestRegistryBroadcastClosesSlowConn(t *testing.T) {
	r := NewRegistry()
	slow := &Conn{
		tenantIDs: []string{"tenant-1"},
		write:     make(chan []byte), // unbuffered → Send default branch fires immediately
		closed:    make(chan struct{}),
	}
	r.Register(slow)

	r.Broadcast("tenant-1", []byte("evt"))

	select {
	case <-slow.closed:
	default:
		t.Fatal("slow conn should be closed when write buffer is full")
	}
}
