package conv

import (
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/polarixa/replify/pkg/strutil"
)

// ///////////////////////////
// Section: Int conversion interface
// ///////////////////////////

// intConverter is an interface for types that can convert themselves to int64.
type intConverter interface {
	Int64() (int64, error)
}

// ///////////////////////////
// Section: Platform-specific int bounds
// ///////////////////////////

var (
	mathMaxInt  int64             // Maximum value for int on the platform
	mathMinInt  int64             // Minimum value for int on the platform
	mathMaxUint uint64            // Maximum value for uint on the platform
	mathIntSize = strconv.IntSize // Integer size in bits (32 or 64)
)

// init initializes platform-specific integer size constants.
func init() {
	initIntSizes(mathIntSize)
}

// initIntSizes sets the mathMaxInt, mathMinInt, and mathMaxUint
// variables based on the provided integer size (32 or 64 bits).
//
// Parameters:
//   - size: The integer size in bits (32 or 64).
func initIntSizes(size int) {
	switch size {
	case 64:
		mathMaxInt = math.MaxInt64
		mathMinInt = math.MinInt64
		mathMaxUint = math.MaxUint64
	case 32:
		mathMaxInt = math.MaxInt32
		mathMinInt = math.MinInt32
		mathMaxUint = math.MaxUint32
	}
}

// ///////////////////////////
// Section: Int64 conversion
// ///////////////////////////

// Int64 attempts to convert the given value to int64, returns the zero value
// and an error on failure.
//
// Parameters:
//   - from: The value to convert.
//
// Returns:
//   - The converted int64 value.
//   - An error if conversion fails.
func (c *Converter) Int64(from any) (int64, error) {
	// Handle nil
	if from == nil {
		if c.nilAsZero {
			return 0, nil
		}
		return 0, newConvError(from, "int64")
	}

	// Fast path for common types
	switch v := from.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case string:
		return c.stringToInt64(v)
	case float64:
		if c.strictMode && v != math.Trunc(v) {
			return 0, newConvErrorf("lossy conversion: %v has fractional part, cannot convert to int64 in strict mode", v)
		}
		return int64(v), nil
	case float32:
		if c.strictMode && float64(v) != math.Trunc(float64(v)) {
			return 0, newConvErrorf("lossy conversion: %v has fractional part, cannot convert to int64 in strict mode", v)
		}
		return int64(v), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case *int64:
		if v == nil {
			return 0, nil
		}
		return *v, nil
	case *int:
		if v == nil {
			return 0, nil
		}
		return int64(*v), nil
	case *string:
		if v == nil {
			return 0, nil
		}
		return c.stringToInt64(*v)
	}

	// Check for custom converter interface
	if conv, ok := from.(intConverter); ok {
		return conv.Int64()
	}

	// Use reflection for other types
	return c.int64FromReflect(from)
}

// Int attempts to convert the given value to int, returns the zero value
// and an error on failure.
//
// Parameters:
//   - from: The value to convert.
//
// Returns:
//   - The converted int value.
//   - An error if conversion fails.
func (c *Converter) Int(from any) (int, error) {
	if v, ok := from.(int); ok {
		return v, nil
	}

	to64, err := c.Int64(from)
	if err != nil {
		return 0, newConvError(from, "int")
	}

	// Handle overflow on 32-bit systems
	if to64 > mathMaxInt || to64 < mathMinInt {
		if c.strictMode {
			return 0, newConvErrorf("value %d overflows int (size=%d bits)", to64, mathIntSize)
		}
		if to64 > mathMaxInt {
			to64 = mathMaxInt
		} else {
			to64 = mathMinInt
		}
	}

	return int(to64), nil
}

// Int8 attempts to convert the given value to int8, returns the zero value
// and an error on failure.
//
// Parameters:
//   - from: The value to convert.
//
// Returns:
//   - The converted int8 value.
//   - An error if conversion fails.
func (c *Converter) Int8(from any) (int8, error) {
	if v, ok := from.(int8); ok {
		return v, nil
	}

	to64, err := c.Int64(from)
	if err != nil {
		return 0, newConvError(from, "int8")
	}

	if to64 > math.MaxInt8 || to64 < math.MinInt8 {
		if c.strictMode {
			return 0, newConvErrorf("value %d overflows int8", to64)
		}
		if to64 > math.MaxInt8 {
			to64 = math.MaxInt8
		} else {
			to64 = math.MinInt8
		}
	}

	return int8(to64), nil
}

// Int16 attempts to convert the given value to int16, returns the zero value
// and an error on failure.
//
// Parameters:
//   - from: The value to convert.
//
// Returns:
//   - The converted int16 value.
//   - An error if conversion fails.
func (c *Converter) Int16(from any) (int16, error) {
	if v, ok := from.(int16); ok {
		return v, nil
	}

	to64, err := c.Int64(from)
	if err != nil {
		return 0, newConvError(from, "int16")
	}

	if to64 > math.MaxInt16 || to64 < math.MinInt16 {
		if c.strictMode {
			return 0, newConvErrorf("value %d overflows int16", to64)
		}
		if to64 > math.MaxInt16 {
			to64 = math.MaxInt16
		} else {
			to64 = math.MinInt16
		}
	}

	return int16(to64), nil
}

// Int32 attempts to convert the given value to int32, returns the zero value
// and an error on failure.
//
// Parameters:
//   - from: The value to convert.
//
// Returns:
//   - The converted int32 value.
//   - An error if conversion fails.
func (c *Converter) Int32(from any) (int32, error) {
	if v, ok := from.(int32); ok {
		return v, nil
	}

	to64, err := c.Int64(from)
	if err != nil {
		return 0, newConvError(from, "int32")
	}

	if to64 > math.MaxInt32 || to64 < math.MinInt32 {
		if c.strictMode {
			return 0, newConvErrorf("value %d overflows int32", to64)
		}
		if to64 > math.MaxInt32 {
			to64 = math.MaxInt32
		} else {
			to64 = math.MinInt32
		}
	}

	return int32(to64), nil
}

// stringToInt64 converts a string to an int64 value.
//
// Parameters:
//   - `v`: The input string to be converted.
//
// Returns:
//   - An int64 value representing the converted string.
//   - An error if the conversion fails.
func (c *Converter) stringToInt64(v string) (int64, error) {
	if strutil.IsEmpty(v) {
		if c.emptyAsZero {
			return 0, nil
		}
		return 0, newConvErrorMsg("cannot convert empty string to int64")
	}

	if c.trimStrings {
		v = strings.TrimSpace(v)
	}

	// Try parsing as integer first
	if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
		return parsed, nil
	}

	// Try parsing as float and truncate
	if parsed, err := strconv.ParseFloat(v, 64); err == nil {
		if c.strictMode && parsed != math.Trunc(parsed) {
			return 0, newConvErrorf("lossy conversion: %q has fractional part, cannot convert to int64 in strict mode", v)
		}
		return int64(parsed), nil
	}

	// Try parsing as bool
	if parsed, err := c.stringToBool(v); err == nil {
		if parsed {
			return 1, nil
		}
		return 0, nil
	}

	return 0, newConvErrorf("cannot convert %q to int64", v)
}

// int64FromReflect converts a reflect.Value to an int64 value based on its kind.
//
// Parameters:
//   - `from`: The input value to be converted.
//
// Returns:
//   - An int64 value representing the converted input.
//   - An error if the conversion fails.
func (c *Converter) int64FromReflect(from any) (int64, error) {
	value := indirectValue(reflect.ValueOf(from))
	if !value.IsValid() {
		if c.nilAsZero {
			return 0, nil
		}
		return 0, newConvError(from, "int64")
	}

	kind := value.Kind()
	switch {
	case kind == reflect.String:
		return c.stringToInt64(value.String())
	case isKindInt(kind):
		return value.Int(), nil
	case isKindUint(kind):
		val := value.Uint()
		if val > math.MaxInt64 {
			val = math.MaxInt64
		}
		return int64(val), nil
	case isKindFloat(kind):
		f := value.Float()
		if c.strictMode && f != math.Trunc(f) {
			return 0, newConvErrorf("lossy conversion: %v has fractional part, cannot convert to int64 in strict mode", f)
		}
		return int64(f), nil
	case isKindComplex(kind):
		return int64(real(value.Complex())), nil
	case kind == reflect.Bool:
		if value.Bool() {
			return 1, nil
		}
		return 0, nil
	case isKindLength(kind):
		return int64(value.Len()), nil
	}

	return 0, newConvError(from, "int64")
}
