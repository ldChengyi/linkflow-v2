package validator

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

type EventInput struct {
	TenantID  string
	ProductID string
	EventName string
	Params    map[string]any
}

type EventResult struct {
	Accepted map[string]any
	Dropped  []string
}

type EventReportValidator struct {
	reader ThingsModelReader
	ttl    time.Duration
	now    func() time.Time

	mu    sync.Mutex
	cache map[string]cachedDefinition
}

func NewEventReportValidator(reader ThingsModelReader, ttl time.Duration) (*EventReportValidator, error) {
	if reader == nil {
		return nil, errors.New("thingsmodel reader is nil")
	}
	if ttl <= 0 {
		ttl = DefaultCacheTTL
	}
	return &EventReportValidator{
		reader: reader,
		ttl:    ttl,
		now:    time.Now,
		cache:  make(map[string]cachedDefinition),
	}, nil
}

func (v *EventReportValidator) Validate(ctx context.Context, in EventInput) (EventResult, error) {
	if err := ctx.Err(); err != nil {
		return EventResult{}, err
	}

	in.TenantID = strings.TrimSpace(in.TenantID)
	in.ProductID = strings.TrimSpace(in.ProductID)
	in.EventName = strings.TrimSpace(in.EventName)
	if in.TenantID == "" || in.ProductID == "" || in.EventName == "" {
		return EventResult{}, fmt.Errorf("validate event report: tenant_id, product_id and event_name are required")
	}

	def, err := v.lookup(ctx, in.TenantID, in.ProductID)
	if err != nil {
		return EventResult{}, err
	}

	eventDef, ok := def.Events[in.EventName]
	if !ok {
		return EventResult{}, ErrEventNotFound
	}
	if len(eventDef.Output) == 0 {
		return EventResult{Accepted: map[string]any{}, Dropped: keysOf(in.Params)}, nil
	}
	if len(in.Params) == 0 {
		return EventResult{}, ErrNoAcceptedEventParams
	}

	accepted := make(map[string]any, len(in.Params))
	dropped := make([]string, 0)
	for name, value := range in.Params {
		paramDef, known := eventDef.Output[name]
		if !known {
			dropped = append(dropped, name)
			continue
		}
		if err := validateEventParamValue(name, paramDef, value); err != nil {
			return EventResult{}, err
		}
		accepted[name] = value
	}

	if len(accepted) == 0 {
		return EventResult{Dropped: dropped}, ErrNoAcceptedEventParams
	}

	return EventResult{Accepted: accepted, Dropped: dropped}, nil
}

func (v *EventReportValidator) lookup(ctx context.Context, tenantID string, productID string) (ThingsModelDefinition, error) {
	key := tenantID + "|" + productID

	v.mu.Lock()
	if entry, ok := v.cache[key]; ok && v.now().Before(entry.expiresAt) {
		v.mu.Unlock()
		return entry.definition, nil
	}
	v.mu.Unlock()

	def, err := v.reader.FindCurrentByProductID(ctx, tenantID, productID)
	if err != nil {
		return ThingsModelDefinition{}, err
	}

	v.mu.Lock()
	v.cache[key] = cachedDefinition{definition: def, expiresAt: v.now().Add(v.ttl)}
	v.mu.Unlock()

	return def, nil
}

func keysOf(m map[string]any) []string {
	if len(m) == 0 {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
