package validator

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

const DefaultCacheTTL = 30 * time.Second

type Input struct {
	TenantID   string
	ProductID  string
	Properties map[string]any
}

type Result struct {
	Accepted map[string]any
	Dropped  []string
}

type PropertyReportValidator struct {
	reader ThingsModelReader
	ttl    time.Duration
	now    func() time.Time

	mu    sync.Mutex
	cache map[string]cachedDefinition
}

type cachedDefinition struct {
	definition ThingsModelDefinition
	expiresAt  time.Time
}

func NewPropertyReportValidator(reader ThingsModelReader, ttl time.Duration) (*PropertyReportValidator, error) {
	if reader == nil {
		return nil, errors.New("thingsmodel reader is nil")
	}
	if ttl <= 0 {
		ttl = DefaultCacheTTL
	}
	return &PropertyReportValidator{
		reader: reader,
		ttl:    ttl,
		now:    time.Now,
		cache:  make(map[string]cachedDefinition),
	}, nil
}

func (v *PropertyReportValidator) Validate(ctx context.Context, in Input) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	in.TenantID = strings.TrimSpace(in.TenantID)
	in.ProductID = strings.TrimSpace(in.ProductID)
	if in.TenantID == "" || in.ProductID == "" {
		return Result{}, fmt.Errorf("validate property report: tenant_id and product_id are required")
	}
	if len(in.Properties) == 0 {
		return Result{}, ErrNoAcceptedProperties
	}

	def, err := v.lookup(ctx, in.TenantID, in.ProductID)
	if err != nil {
		return Result{}, err
	}

	accepted := make(map[string]any, len(in.Properties))
	dropped := make([]string, 0)
	for name, value := range in.Properties {
		propDef, known := def.Properties[name]
		if !known {
			dropped = append(dropped, name)
			continue
		}
		if err := validatePropertyValue(name, propDef, value); err != nil {
			return Result{}, err
		}
		accepted[name] = value
	}

	if len(accepted) == 0 {
		return Result{Dropped: dropped}, ErrNoAcceptedProperties
	}

	return Result{Accepted: accepted, Dropped: dropped}, nil
}

func (v *PropertyReportValidator) lookup(ctx context.Context, tenantID string, productID string) (ThingsModelDefinition, error) {
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
