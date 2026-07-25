package conv

import (
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/polarixa/replify/pkg/strutil"
)

// ///////////////////////////
// Section: Float conversion interface
// ///////////////////////////

// floatConverter is an interface for types that can convert themselves to float64.
type floatConverter interface {
	Float64() (float64, error)
}

// ///////////////////////////
// Section: Float64 conversion
// ///////////////////////////

// Float64 attempts to convert the given value to float64, returns the zero value
// and an error on failure.
//
// Parameters:
//   - from: The value to convert.
//
// Returns:
//   - The converted float64 value.
//   - An error if conversion fails.
func (c *Converter) Float64(from any) (float64, error) {
	if from == nil {
		if c.nilAsZero {
			return 0, nil
		}
		return 0, newConvError(from, "float64")
	}

	// Fast path for common types
	switch v := from.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case string:
		return c.stringToFloat64(v)
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case *float64:
		if v == nil {
			return 0, nil
		}
		return *v, nil
	case *float32:
		if v == nil {
			return 0, nil
		}
		return float64(*v), nil
	case *string:
		if v == nil {
			return 0, nil
		}
		return c.stringToFloat64(*v)
	}

	// Check for custom converter interface
	if conv, ok := from.(floatConverter); ok {
		return conv.Float64()
	}

	// Use reflection for other types
	return c.float64FromReflect(from)
}

// ///////////////////////////
// Section:  Float32 conversion
// ///////////////////////////

// Float32 attempts to convert the given value to float32, returns the zero value
// and an error on failure.
//
// Note: Values exceeding float32 range are clamped to max/min float32.
//
// Parameters:
//   - from: The value to convert.
//
// Returns:
//   - The converted float32 value.
//   - An error if conversion fails.
func (c *Converter) Float32(from any) (float32, error) {
	if v, ok := from.(float32); ok {
		return v, nil
	}

	f64, err := c.Float64(from)
	if err != nil {
		return 0, newConvError(from, "float32")
	}

	// Overflow protection
	if f64 > math.MaxFloat32 || f64 < -math.MaxFloat32 {
		if c.strictMode {
			return 0, newConvErrorf("value %v overflows float32", f64)
		}
		if f64 > math.MaxFloat32 {
			f64 = math.MaxFloat32
		} else {
			f64 = -math.MaxFloat32
		}
	}

	return float32(f64), nil
}

// ///////////////////////////
// Section: Special float values
// ///////////////////////////

// IsNaN checks if the given value converts to NaN which indicates "Not a Number".
//
// Parameters:
//   - from: The value to check.
//
// Returns:
//   - A boolean indicating whether the converted value is NaN.
func IsNaN(from any) bool {
	if v, err := defaultConverter.Float64(from); err == nil {
		return math.IsNaN(v)
	}
	return false
}

// IsInf checks if the given value converts to infinity.
// sign > 0 checks for positive infinity, sign < 0 for negative infinity,
// sign == 0 checks for either.
//
// Parameters:
//   - from: The value to check.
//   - sign: The sign of infinity to check for.
//
// Returns:
//   - A boolean indicating whether the converted value is infinite with the specified sign.
//
// Example:
//   - IsInf(value, 1) checks for positive infinity.
//   - IsInf(value, -1) checks for negative infinity.
//   - IsInf(value, 0) checks for either positive or negative infinity.
func IsInf(from any, sign int) bool {
	if v, err := defaultConverter.Float64(from); err == nil {
		return math.IsInf(v, sign)
	}
	return false
}

// IsFinite checks if the given value converts to a finite number. That is, it is neither NaN nor infinite.
//
// Parameters:
//   - from: The value to check.
//
// Returns:
//   - A boolean indicating whether the converted value is finite (not NaN or infinite).
//
// Example:
//   - IsFinite(value) returns true if value is a finite number.
func IsFinite(from any) bool {
	if v, err := defaultConverter.Float64(from); err == nil {
		return !math.IsNaN(v) && !math.IsInf(v, 0)
	}
	return false
}

// stringToFloat64 attempts to convert a string to float64.
//
// Parameters:
//   - v: The string to convert.
//
// Returns:
//   - The converted float64 value.
//   - An error if conversion fails.
func (c *Converter) stringToFloat64(v string) (float64, error) {
	if strutil.IsEmpty(v) {
		if c.emptyAsZero {
			return 0, nil
		}
		return 0, newConvErrorMsg("cannot convert empty string to float64")
	}

	if c.trimStrings {
		v = strings.TrimSpace(v)
	}

	// Try parsing as float
	if parsed, err := strconv.ParseFloat(v, 64); err == nil {
		return parsed, nil
	}

	// Try parsing as bool
	if parsed, err := c.stringToBool(v); err == nil {
		if parsed {
			return 1, nil
		}
		return 0, nil
	}

	return 0, newConvErrorf("cannot convert %q to float64", v)
}

// float64FromReflect converts a reflect.Value to float64.
//
// Parameters:
//   - from: The input value to convert.
//
// Returns:
//   - The converted float64 value.
//   - An error if the conversion fails.
func (c *Converter) float64FromReflect(from any) (float64, error) {
	value := indirectValue(reflect.ValueOf(from))
	if !value.IsValid() {
		if c.nilAsZero {
			return 0, nil
		}
		return 0, newConvError(from, "float64")
	}

	kind := value.Kind()
	switch {
	case kind == reflect.String:
		return c.stringToFloat64(value.String())
	case isKindFloat(kind):
		return value.Float(), nil
	case isKindInt(kind):
		return float64(value.Int()), nil
	case isKindUint(kind):
		return float64(value.Uint()), nil
	case isKindComplex(kind):
		return real(value.Complex()), nil
	case kind == reflect.Bool:
		if value.Bool() {
			return 1, nil
		}
		return 0, nil
	case isKindLength(kind):
		return float64(value.Len()), nil
	}

	return 0, newConvError(from, "float64")
}
