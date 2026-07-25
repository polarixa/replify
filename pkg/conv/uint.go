package conv

import (
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/polarixa/replify/pkg/strutil"
)

// ///////////////////////////
// Section: Uint conversion interface
// ///////////////////////////

// uintConverter is an interface for types that can convert themselves to uint64.
type uintConverter interface {
	Uint64() (uint64, error)
}

// ///////////////////////////
// Section: Uint64 conversion
// ///////////////////////////

// Uint64 attempts to convert the given value to uint64, returns the zero value
// and an error on failure.
//
// Note: Negative values are converted to 0 (underflow protection).
//
// Parameters:
//   - from: The value to convert.
//
// Returns:
//   - The converted uint64 value.
//   - An error if conversion fails.
func (c *Converter) Uint64(from any) (uint64, error) {
	if from == nil {
		if c.nilAsZero {
			return 0, nil
		}
		return 0, newConvError(from, "uint64")
	}

	// Fast path for common types
	switch v := from.(type) {
	case uint64:
		return v, nil
	case uint:
		return uint64(v), nil
	case uint32:
		return uint64(v), nil
	case uint16:
		return uint64(v), nil
	case uint8:
		return uint64(v), nil
	case int64:
		if v < 0 {
			return 0, nil // Underflow protection, negative to zero
		}
		return uint64(v), nil
	case int:
		if v < 0 {
			return 0, nil
		}
		return uint64(v), nil
	case int32:
		if v < 0 {
			return 0, nil
		}
		return uint64(v), nil
	case string:
		return c.stringToUint64(v)
	case float64:
		if v < 0 {
			return 0, nil
		}
		return uint64(v), nil
	case float32:
		if v < 0 {
			return 0, nil
		}
		return uint64(v), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case *uint64:
		if v == nil {
			return 0, nil
		}
		return *v, nil
	case *uint:
		if v == nil {
			return 0, nil
		}
		return uint64(*v), nil
	case *string:
		if v == nil {
			return 0, nil
		}
		return c.stringToUint64(*v)
	}

	// Check for custom converter interface
	if conv, ok := from.(uintConverter); ok {
		return conv.Uint64()
	}

	// Use reflection for other types
	return c.uint64FromReflect(from)
}

// Uint attempts to convert the given value to uint, returns the zero value
// and an error on failure.
//
// Parameters:
//   - from: The value to convert.
//
// Returns:
//   - The converted uint value.
//   - An error if conversion fails.
func (c *Converter) Uint(from any) (uint, error) {
	if v, ok := from.(uint); ok {
		return v, nil
	}

	to64, err := c.Uint64(from)
	if err != nil {
		return 0, newConvError(from, "uint")
	}

	// Handle overflow on 32-bit systems
	if to64 > mathMaxUint {
		if c.strictMode {
			return 0, newConvErrorf("value %d overflows uint (size=%d bits)", to64, mathIntSize)
		}
		to64 = mathMaxUint
	}

	return uint(to64), nil
}

// Uint8 attempts to convert the given value to uint8, returns the zero value
// and an error on failure.
//
// Parameters:
//   - from: The value to convert.
//
// Returns:
//   - The converted uint8 value.
//   - An error if conversion fails.
func (c *Converter) Uint8(from any) (uint8, error) {
	if v, ok := from.(uint8); ok {
		return v, nil
	}

	to64, err := c.Uint64(from)
	if err != nil {
		return 0, newConvError(from, "uint8")
	}

	if to64 > math.MaxUint8 {
		if c.strictMode {
			return 0, newConvErrorf("value %d overflows uint8", to64)
		}
		to64 = math.MaxUint8
	}

	return uint8(to64), nil
}

// Uint16 attempts to convert the given value to uint16, returns the zero value
// and an error on failure.
//
// Parameters:
//   - from: The value to convert.
//
// Returns:
//   - The converted uint16 value.
//   - An error if conversion fails.
func (c *Converter) Uint16(from any) (uint16, error) {
	if v, ok := from.(uint16); ok {
		return v, nil
	}

	to64, err := c.Uint64(from)
	if err != nil {
		return 0, newConvError(from, "uint16")
	}

	if to64 > math.MaxUint16 {
		if c.strictMode {
			return 0, newConvErrorf("value %d overflows uint16", to64)
		}
		to64 = math.MaxUint16
	}

	return uint16(to64), nil
}

// Uint32 attempts to convert the given value to uint32, returns the zero value
// and an error on failure.
//
// Parameters:
//   - from: The value to convert.
//
// Returns:
//   - The converted uint32 value.
//   - An error if conversion fails.
func (c *Converter) Uint32(from any) (uint32, error) {
	if v, ok := from.(uint32); ok {
		return v, nil
	}

	to64, err := c.Uint64(from)
	if err != nil {
		return 0, newConvError(from, "uint32")
	}

	if to64 > math.MaxUint32 {
		if c.strictMode {
			return 0, newConvErrorf("value %d overflows uint32", to64)
		}
		to64 = math.MaxUint32
	}

	return uint32(to64), nil
}

// stringToUint64 attempts to convert a string to uint64.
//
// Parameters:
//   - v: The string to convert.
//
// Returns:
//   - The converted uint64 value.
//   - An error if the conversion fails.
func (c *Converter) stringToUint64(v string) (uint64, error) {
	if strutil.IsEmpty(v) {
		if c.emptyAsZero {
			return 0, nil
		}
		return 0, newConvErrorMsg("cannot convert empty string to uint64")
	}

	if c.trimStrings {
		v = strings.TrimSpace(v)
	}

	// Try parsing as unsigned integer first
	if parsed, err := strconv.ParseUint(v, 10, 64); err == nil {
		return parsed, nil
	}

	// Try parsing as float and truncate (with underflow protection)
	if parsed, err := strconv.ParseFloat(v, 64); err == nil {
		return uint64(math.Max(0, parsed)), nil
	}

	// Try parsing as bool
	if parsed, err := c.stringToBool(v); err == nil {
		if parsed {
			return 1, nil
		}
		return 0, nil
	}

	return 0, newConvErrorf("cannot convert %q to uint64", v)
}

// uint64FromReflect converts a reflect.Value to uint64.
//
// Parameters:
//   - from: The input value to convert.
//
// Returns:
//   - The converted uint64 value.
//   - An error if the conversion fails.
func (c *Converter) uint64FromReflect(from any) (uint64, error) {
	value := indirectValue(reflect.ValueOf(from))
	if !value.IsValid() {
		if c.nilAsZero {
			return 0, nil
		}
		return 0, newConvError(from, "uint64")
	}

	kind := value.Kind()
	switch {
	case kind == reflect.String:
		return c.stringToUint64(value.String())
	case isKindUint(kind):
		return value.Uint(), nil
	case isKindInt(kind):
		val := value.Int()
		if val < 0 {
			return 0, nil // Underflow protection, negative to zero
		}
		return uint64(val), nil
	case isKindFloat(kind):
		return uint64(math.Max(0, value.Float())), nil
	case isKindComplex(kind):
		return uint64(math.Max(0, real(value.Complex()))), nil
	case kind == reflect.Bool:
		if value.Bool() {
			return 1, nil
		}
		return 0, nil
	case isKindLength(kind):
		return uint64(value.Len()), nil
	}

	return 0, newConvError(from, "uint64")
}
