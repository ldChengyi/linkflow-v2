package realtime

import (
	"sync"
)

// Registry maps tenant_id to the set of live WebSocket connections that should
// receive that tenant's state events. It is the in-memory fan-out point for the
// backend Kafka consumer.
type Registry struct {
	mu       sync.RWMutex
	byTenant map[string]map[*Conn]struct{}
}

func NewRegistry() *Registry {
	return &Registry{
		byTenant: make(map[string]map[*Conn]struct{}),
	}
}

func (r *Registry) Register(c *Conn) {
	if c == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, tenantID := range c.tenantIDs {
		conns, ok := r.byTenant[tenantID]
		if !ok {
			conns = make(map[*Conn]struct{})
			r.byTenant[tenantID] = conns
		}
		conns[c] = struct{}{}
	}
}

func (r *Registry) Unregister(c *Conn) {
	if c == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, tenantID := range c.tenantIDs {
		conns, ok := r.byTenant[tenantID]
		if !ok {
			continue
		}
		delete(conns, c)
		if len(conns) == 0 {
			delete(r.byTenant, tenantID)
		}
	}
}

// Broadcast forwards the raw event bytes to every connection registered under
// tenantID. Slow consumers whose write buffer is full are dropped silently;
// their writeLoop will close the connection.
func (r *Registry) Broadcast(tenantID string, msg []byte) int {
	if tenantID == "" {
		return 0
	}
	r.mu.RLock()
	conns := r.byTenant[tenantID]
	targets := make([]*Conn, 0, len(conns))
	for c := range conns {
		targets = append(targets, c)
	}
	r.mu.RUnlock()

	delivered := 0
	for _, c := range targets {
		if c.Send(msg) {
			delivered++
		}
	}
	return delivered
}
