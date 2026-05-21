package validator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const (
	DataTypeInt    = "int"
	DataTypeFloat  = "float"
	DataTypeDouble = "double"
	DataTypeBool   = "bool"
	DataTypeString = "string"
)

var (
	ErrThingsModelNotFound  = errors.New("things model not found")
	ErrInvalidPropertyValue = errors.New("invalid property value")
	ErrNoAcceptedProperties = errors.New("no accepted properties remain after filtering")
)

type ThingsModelReader interface {
	FindCurrentByProductID(ctx context.Context, tenantID string, productID string) (ThingsModelDefinition, error)
}

type ThingsModelDefinition struct {
	TenantID   string
	ProductID  string
	Properties map[string]PropertyDefinition
}

type PropertyDefinition struct {
	DataType     string
	Required     bool
	HasMin       bool
	Min          float64
	HasMax       bool
	Max          float64
	HasStep      bool
	Step         float64
	HasPrecision bool
	Precision    int
}

func ParseDefinition(tenantID string, productID string, propertiesJSON []byte) (ThingsModelDefinition, error) {
	var raw map[string]map[string]any
	if err := json.Unmarshal(propertiesJSON, &raw); err != nil {
		return ThingsModelDefinition{}, fmt.Errorf("decode thingsmodel properties: %w", err)
	}

	properties := make(map[string]PropertyDefinition, len(raw))
	for identifier, body := range raw {
		def, err := parsePropertyDefinition(body)
		if err != nil {
			return ThingsModelDefinition{}, fmt.Errorf("property %q: %w", identifier, err)
		}
		properties[identifier] = def
	}

	return ThingsModelDefinition{
		TenantID:   tenantID,
		ProductID:  productID,
		Properties: properties,
	}, nil
}

func parsePropertyDefinition(body map[string]any) (PropertyDefinition, error) {
	dataType, _ := body["data_type"].(string)
	required, _ := body["required"].(bool)

	def := PropertyDefinition{
		DataType: strings.TrimSpace(dataType),
		Required: required,
	}

	spec, _ := body["spec"].(map[string]any)
	if min, ok := numberFromAny(spec["min"]); ok {
		def.HasMin = true
		def.Min = min
	}
	if max, ok := numberFromAny(spec["max"]); ok {
		def.HasMax = true
		def.Max = max
	}
	if step, ok := numberFromAny(spec["step"]); ok {
		def.HasStep = true
		def.Step = step
	}
	if precision, ok := numberFromAny(spec["precision"]); ok {
		def.HasPrecision = true
		def.Precision = int(precision)
	}

	return def, nil
}

func numberFromAny(value any) (float64, bool) {
	switch n := value.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		if err != nil {
			return 0, false
		}
		return f, true
	}
	return 0, false
}
