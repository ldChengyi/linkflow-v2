package validator

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

type ServiceCallInput struct {
	TenantID    string
	ProductID   string
	ServiceName string
	Output      map[string]any
}

type ServiceCallResult struct {
	Accepted map[string]any
	Dropped  []string
}

type ServiceCallValidator struct {
	reader ThingsModelReader
	ttl    time.Duration
	now    func() time.Time

	mu    sync.Mutex
	cache map[string]cachedDefinition
}

func NewServiceCallValidator(reader ThingsModelReader, ttl time.Duration) (*ServiceCallValidator, error) {
	if reader == nil {
		return nil, errors.New("thingsmodel reader is nil")
	}
	if ttl <= 0 {
		ttl = DefaultCacheTTL
	}
	return &ServiceCallValidator{
		reader: reader,
		ttl:    ttl,
		now:    time.Now,
		cache:  make(map[string]cachedDefinition),
	}, nil
}

func (v *ServiceCallValidator) Validate(ctx context.Context, in ServiceCallInput) (ServiceCallResult, error) {
	if err := ctx.Err(); err != nil {
		return ServiceCallResult{}, err
	}

	in.TenantID = strings.TrimSpace(in.TenantID)
	in.ProductID = strings.TrimSpace(in.ProductID)
	in.ServiceName = strings.TrimSpace(in.ServiceName)
	if in.TenantID == "" || in.ProductID == "" || in.ServiceName == "" {
		return ServiceCallResult{}, fmt.Errorf("validate service call: tenant_id, product_id and service_name are required")
	}

	def, err := v.lookup(ctx, in.TenantID, in.ProductID)
	if err != nil {
		return ServiceCallResult{}, err
	}

	serviceDef, ok := def.Services[in.ServiceName]
	if !ok {
		return ServiceCallResult{}, ErrServiceNotFound
	}
	if len(serviceDef.Output) == 0 {
		return ServiceCallResult{Accepted: map[string]any{}, Dropped: keysOf(in.Output)}, nil
	}
	if len(in.Output) == 0 {
		return ServiceCallResult{Accepted: map[string]any{}}, nil
	}

	accepted := make(map[string]any, len(in.Output))
	dropped := make([]string, 0)
	for name, value := range in.Output {
		outputDef, known := serviceDef.Output[name]
		if !known {
			dropped = append(dropped, name)
			continue
		}
		if err := validateServiceOutputValue(name, outputDef, value); err != nil {
			return ServiceCallResult{}, err
		}
		accepted[name] = value
	}

	return ServiceCallResult{Accepted: accepted, Dropped: dropped}, nil
}

func (v *ServiceCallValidator) lookup(ctx context.Context, tenantID string, productID string) (ThingsModelDefinition, error) {
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
