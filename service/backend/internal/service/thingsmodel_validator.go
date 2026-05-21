package service

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
)

var (
	errInvalidThingsModelDefinition = errors.New("invalid thingsmodel definition")
	thingsModelIdentifierPattern    = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)
)

const (
	propertyDataTypeInt    = "int"
	propertyDataTypeFloat  = "float"
	propertyDataTypeDouble = "double"
	propertyDataTypeBool   = "bool"
	propertyDataTypeString = "string"

	propertyAccessRead      = "read"
	propertyAccessWrite     = "write"
	propertyAccessReadWrite = "readwrite"

	eventLevelInfo    = "info"
	eventLevelWarning = "warning"
	eventLevelError   = "error"

	serviceCallTypeSync  = "sync"
	serviceCallTypeAsync = "async"
)

func validateThingsModelDefinition(properties, events, services ThingsModelObject) error {
	if err := validateThingsModelProperties(properties); err != nil {
		return err
	}
	if err := validateThingsModelEvents(events); err != nil {
		return err
	}
	if err := validateThingsModelServices(services); err != nil {
		return err
	}
	return nil
}

func validateThingsModelProperties(properties ThingsModelObject) error {
	for identifier, raw := range properties {
		if !validThingsModelIdentifier(identifier) {
			return fmt.Errorf("%w: invalid property identifier", errInvalidThingsModelDefinition)
		}
		def, ok := thingsModelDefinitionObject(raw)
		if !ok {
			return fmt.Errorf("%w: property must be object", errInvalidThingsModelDefinition)
		}
		if err := validateThingModelParamDefinition(def, true); err != nil {
			return fmt.Errorf("%w: property %s: %v", errInvalidThingsModelDefinition, identifier, err)
		}
	}
	return nil
}

func validateThingsModelEvents(events ThingsModelObject) error {
	for identifier, raw := range events {
		if !validThingsModelIdentifier(identifier) {
			return fmt.Errorf("%w: invalid event identifier", errInvalidThingsModelDefinition)
		}
		def, ok := thingsModelDefinitionObject(raw)
		if !ok {
			return fmt.Errorf("%w: event must be object", errInvalidThingsModelDefinition)
		}
		if !requiredString(def, "name") {
			return fmt.Errorf("%w: event name is required", errInvalidThingsModelDefinition)
		}
		if !validEventLevel(optionalString(def, "level")) {
			return fmt.Errorf("%w: invalid event level", errInvalidThingsModelDefinition)
		}
		if desc, exists := def["desc"]; exists {
			if _, ok := desc.(string); !ok {
				return fmt.Errorf("%w: event desc must be string", errInvalidThingsModelDefinition)
			}
		}
		output, ok := thingsModelDefinitionObject(def["output"])
		if !ok {
			return fmt.Errorf("%w: event output must be object", errInvalidThingsModelDefinition)
		}
		if err := validateThingModelParamMap(output); err != nil {
			return fmt.Errorf("%w: event %s output: %v", errInvalidThingsModelDefinition, identifier, err)
		}
	}
	return nil
}

func validateThingsModelServices(services ThingsModelObject) error {
	for identifier, raw := range services {
		if !validThingsModelIdentifier(identifier) {
			return fmt.Errorf("%w: invalid service identifier", errInvalidThingsModelDefinition)
		}
		def, ok := thingsModelDefinitionObject(raw)
		if !ok {
			return fmt.Errorf("%w: service must be object", errInvalidThingsModelDefinition)
		}
		if !requiredString(def, "name") {
			return fmt.Errorf("%w: service name is required", errInvalidThingsModelDefinition)
		}
		if !validServiceCallType(optionalString(def, "call_type")) {
			return fmt.Errorf("%w: invalid service call_type", errInvalidThingsModelDefinition)
		}
		if desc, exists := def["desc"]; exists {
			if _, ok := desc.(string); !ok {
				return fmt.Errorf("%w: service desc must be string", errInvalidThingsModelDefinition)
			}
		}
		input, ok := thingsModelDefinitionObject(def["input"])
		if !ok {
			return fmt.Errorf("%w: service input must be object", errInvalidThingsModelDefinition)
		}
		output, ok := thingsModelDefinitionObject(def["output"])
		if !ok {
			return fmt.Errorf("%w: service output must be object", errInvalidThingsModelDefinition)
		}
		if err := validateThingModelParamMap(input); err != nil {
			return fmt.Errorf("%w: service %s input: %v", errInvalidThingsModelDefinition, identifier, err)
		}
		if err := validateThingModelParamMap(output); err != nil {
			return fmt.Errorf("%w: service %s output: %v", errInvalidThingsModelDefinition, identifier, err)
		}
	}
	return nil
}

func validateThingModelParamMap(params map[string]any) error {
	for identifier, raw := range params {
		if !validThingsModelIdentifier(identifier) {
			return fmt.Errorf("%w: invalid param identifier", errInvalidThingsModelDefinition)
		}
		def, ok := thingsModelDefinitionObject(raw)
		if !ok {
			return fmt.Errorf("%w: param must be object", errInvalidThingsModelDefinition)
		}
		if err := validateThingModelParamDefinition(def, false); err != nil {
			return fmt.Errorf("%w: param %s: %v", errInvalidThingsModelDefinition, identifier, err)
		}
	}
	return nil
}

func validateThingModelParamDefinition(def map[string]any, requireAccessMode bool) error {
	if !requiredString(def, "name") {
		return fmt.Errorf("%w: name is required", errInvalidThingsModelDefinition)
	}
	dataType := optionalString(def, "data_type")
	if !validPropertyDataType(dataType) {
		return fmt.Errorf("%w: invalid data_type", errInvalidThingsModelDefinition)
	}
	if _, ok := def["required"].(bool); !ok {
		return fmt.Errorf("%w: required must be bool", errInvalidThingsModelDefinition)
	}
	if requireAccessMode && !validPropertyAccessMode(optionalString(def, "access_mode")) {
		return fmt.Errorf("%w: invalid access_mode", errInvalidThingsModelDefinition)
	}
	if spec, exists := def["spec"]; exists {
		specObject, ok := thingsModelDefinitionObject(spec)
		if !ok {
			return fmt.Errorf("%w: spec must be object", errInvalidThingsModelDefinition)
		}
		if err := validateThingModelSpec(dataType, specObject); err != nil {
			return err
		}
	}
	return nil
}

func validateThingModelSpec(dataType string, spec map[string]any) error {
	min, hasMin, err := optionalNumber(spec, "min")
	if err != nil {
		return err
	}
	max, hasMax, err := optionalNumber(spec, "max")
	if err != nil {
		return err
	}
	step, hasStep, err := optionalNumber(spec, "step")
	if err != nil {
		return err
	}
	precision, hasPrecision, err := optionalNumber(spec, "precision")
	if err != nil {
		return err
	}
	if unit, exists := spec["unit"]; exists {
		value, ok := unit.(string)
		if !ok || strings.TrimSpace(value) == "" {
			return fmt.Errorf("%w: unit must be non-empty string", errInvalidThingsModelDefinition)
		}
	}
	if (hasMin || hasMax || hasStep || hasPrecision) && !numericThingsModelDataType(dataType) {
		return fmt.Errorf("%w: numeric spec requires numeric data_type", errInvalidThingsModelDefinition)
	}
	if hasMin && hasMax && min > max {
		return fmt.Errorf("%w: min cannot exceed max", errInvalidThingsModelDefinition)
	}
	if hasStep && step <= 0 {
		return fmt.Errorf("%w: step must be greater than zero", errInvalidThingsModelDefinition)
	}
	if hasPrecision {
		if precision < 0 || math.Trunc(precision) != precision {
			return fmt.Errorf("%w: precision must be non-negative integer", errInvalidThingsModelDefinition)
		}
		if dataType == propertyDataTypeInt && precision != 0 {
			return fmt.Errorf("%w: int precision must be zero", errInvalidThingsModelDefinition)
		}
	}
	if dataType == propertyDataTypeInt {
		if (hasMin && math.Trunc(min) != min) || (hasMax && math.Trunc(max) != max) || (hasStep && math.Trunc(step) != step) {
			return fmt.Errorf("%w: int spec values must be integers", errInvalidThingsModelDefinition)
		}
	}
	return nil
}

func validThingsModelIdentifier(identifier string) bool {
	return thingsModelIdentifierPattern.MatchString(identifier)
}

func thingsModelDefinitionObject(value any) (map[string]any, bool) {
	switch object := value.(type) {
	case map[string]any:
		return object, true
	case ThingsModelObject:
		return map[string]any(object), true
	default:
		return nil, false
	}
}

func requiredString(value map[string]any, key string) bool {
	raw, ok := value[key].(string)
	return ok && strings.TrimSpace(raw) != ""
}

func optionalString(value map[string]any, key string) string {
	raw, _ := value[key].(string)
	return strings.TrimSpace(raw)
}

func optionalNumber(value map[string]any, key string) (float64, bool, error) {
	raw, exists := value[key]
	if !exists {
		return 0, false, nil
	}
	var number float64
	switch value := raw.(type) {
	case float64:
		number = value
	case float32:
		number = float64(value)
	case int:
		number = float64(value)
	case int8:
		number = float64(value)
	case int16:
		number = float64(value)
	case int32:
		number = float64(value)
	case int64:
		number = float64(value)
	case uint:
		number = float64(value)
	case uint8:
		number = float64(value)
	case uint16:
		number = float64(value)
	case uint32:
		number = float64(value)
	case uint64:
		number = float64(value)
	default:
		return 0, false, fmt.Errorf("%w: %s must be number", errInvalidThingsModelDefinition, key)
	}
	if math.IsNaN(number) || math.IsInf(number, 0) {
		return 0, false, fmt.Errorf("%w: %s must be number", errInvalidThingsModelDefinition, key)
	}
	return number, true, nil
}

func validPropertyDataType(dataType string) bool {
	switch dataType {
	case propertyDataTypeInt, propertyDataTypeFloat, propertyDataTypeDouble, propertyDataTypeBool, propertyDataTypeString:
		return true
	default:
		return false
	}
}

func numericThingsModelDataType(dataType string) bool {
	switch dataType {
	case propertyDataTypeInt, propertyDataTypeFloat, propertyDataTypeDouble:
		return true
	default:
		return false
	}
}

func validPropertyAccessMode(accessMode string) bool {
	switch accessMode {
	case propertyAccessRead, propertyAccessWrite, propertyAccessReadWrite:
		return true
	default:
		return false
	}
}

func validEventLevel(level string) bool {
	switch level {
	case eventLevelInfo, eventLevelWarning, eventLevelError:
		return true
	default:
		return false
	}
}

func validServiceCallType(callType string) bool {
	switch callType {
	case serviceCallTypeSync, serviceCallTypeAsync:
		return true
	default:
		return false
	}
}
