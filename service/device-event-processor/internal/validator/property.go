package validator

import (
	"fmt"
	"math"
)

const precisionEpsilon = 1e-9

func validatePropertyValue(name string, def PropertyDefinition, value any) error {
	return validateDefinedValue(name, def, value, ErrInvalidPropertyValue)
}

func validateEventParamValue(name string, def PropertyDefinition, value any) error {
	return validateDefinedValue(name, def, value, ErrInvalidEventValue)
}

func validateDefinedValue(name string, def PropertyDefinition, value any, sentinel error) error {
	switch def.DataType {
	case DataTypeInt:
		n, ok := numberFromAny(value)
		if !ok {
			return valueError(sentinel, name, "expected integer value")
		}
		if math.Trunc(n) != n {
			return valueError(sentinel, name, fmt.Sprintf("value %v has fractional part for int", value))
		}
		return validateNumericConstraints(name, def, n, sentinel)
	case DataTypeFloat, DataTypeDouble:
		n, ok := numberFromAny(value)
		if !ok {
			return valueError(sentinel, name, "expected numeric value")
		}
		return validateNumericConstraints(name, def, n, sentinel)
	case DataTypeBool:
		if _, ok := value.(bool); !ok {
			return valueError(sentinel, name, "expected boolean value")
		}
		return nil
	case DataTypeString:
		if _, ok := value.(string); !ok {
			return valueError(sentinel, name, "expected string value")
		}
		return nil
	default:
		return valueError(sentinel, name, fmt.Sprintf("unsupported data_type %q", def.DataType))
	}
}

func validateNumericConstraints(name string, def PropertyDefinition, value float64, sentinel error) error {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return valueError(sentinel, name, "value must be finite number")
	}
	if def.HasMin && value < def.Min {
		return valueError(sentinel, name, fmt.Sprintf("value %v below min %v", value, def.Min))
	}
	if def.HasMax && value > def.Max {
		return valueError(sentinel, name, fmt.Sprintf("value %v above max %v", value, def.Max))
	}

	precision := effectivePrecision(def)
	scale := math.Pow10(precision)

	scaledValue := math.Round(value * scale)
	if math.Abs(value*scale-scaledValue) > precisionEpsilon*math.Max(1, math.Abs(value*scale)) {
		return valueError(sentinel, name, fmt.Sprintf("value %v exceeds precision %d", value, precision))
	}

	if !def.HasStep {
		return nil
	}

	scaledStep := math.Round(def.Step * scale)
	if scaledStep <= 0 {
		return valueError(sentinel, name, fmt.Sprintf("invalid step %v", def.Step))
	}

	anchor := 0.0
	if def.HasMin {
		anchor = math.Round(def.Min * scale)
	}

	diff := scaledValue - anchor
	if math.Mod(diff, scaledStep) != 0 {
		return valueError(sentinel, name, fmt.Sprintf("value %v not aligned with step %v", value, def.Step))
	}
	return nil
}

func effectivePrecision(def PropertyDefinition) int {
	if def.HasPrecision {
		return def.Precision
	}
	digits := 0
	if def.HasStep {
		digits = maxInt(digits, decimalDigits(def.Step))
	}
	if def.HasMin {
		digits = maxInt(digits, decimalDigits(def.Min))
	}
	if def.HasMax {
		digits = maxInt(digits, decimalDigits(def.Max))
	}
	return digits
}

func decimalDigits(value float64) int {
	abs := math.Abs(value)
	for i := 0; i <= 8; i++ {
		scale := math.Pow10(i)
		if math.Abs(abs*scale-math.Round(abs*scale)) < 1e-9 {
			return i
		}
	}
	return 8
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func valueError(sentinel error, name string, reason string) error {
	return fmt.Errorf("%w: %s: %s", sentinel, name, reason)
}
