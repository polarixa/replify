package conv

import (
	"reflect"
	"strings"
	"time"

	"github.com/polarixa/replify/pkg/strutil"
)

// ///////////////////////////
// Section: Bool conversion interface
// ///////////////////////////

// boolConverter is an interface for types that can convert themselves to bool.
type boolConverter interface {
	Bool() (bool, error)
}

// ///////////////////////////
// Section:  Bool conversion methods
// ///////////////////////////

// Bool attempts to convert the given value to bool, returns the zero value
// and an error on failure.
//
// Supported conversions:
//   - string: "1", "t", "T", "true", "True", "TRUE", "y", "Y", "yes", "Yes", "YES" → true
//   - string: "0", "f", "F", "false", "False", "FALSE", "n", "N", "no", "No", "NO" → false
//   - numeric: non-zero → true, zero → false
//   - bool: returned as-is
//   - slice/map/array: len > 0 → true, empty → false
//   - time.Time: non-zero → true, zero → false
//
// Parameters:
//   - from: The value to convert.
//
// Returns:
//   - The converted bool value.
//   - An error if conversion fails.
func (c *Converter) Bool(from any) (bool, error) {
	// Handle nil
	if from == nil {
		if c.nilAsZero {
			return false, nil
		}
		return false, newConvError(from, "bool")
	}

	// Fast path for common types
	switch v := from.(type) {
	case bool:
		return v, nil
	case string:
		return c.stringToBool(v)
	case *bool:
		if v == nil {
			if c.nilAsZero {
				return false, nil
			}
			return false, newConvError(from, "bool")
		}
		return *v, nil
	case *string:
		if v == nil {
			if c.nilAsZero {
				return false, nil
			}
			return false, newConvError(from, "bool")
		}
		return c.stringToBool(*v)
	}

	// Check for custom converter interface
	if conv, ok := from.(boolConverter); ok {
		return conv.Bool()
	}

	// Use reflection for other types
	return c.boolFromReflect(from)
}

// ///////////////////////////
// Section: Helper methods
// ///////////////////////////

// stringToBool converts a string to a boolean value based on common representations.
//
// Parameters:
//   - `v`: The input string to be converted.
//
// Returns:
//   - A boolean value representing the converted string.
func (c *Converter) stringToBool(v string) (bool, error) {
	if strutil.IsEmpty(v) {
		if c.emptyAsZero {
			return false, nil
		}
		return false, newConvErrorMsg("cannot convert empty string to bool")
	}

	if c.trimStrings {
		v = strings.TrimSpace(v)
	}

	// Length check for performance
	if len(v) > 5 {
		return false, newConvErrorf("cannot parse %q as bool", v) // "cannot parse string with len > 5 as bool"
	}

	switch v {
	case "1", "t", "T", "true", "True", "TRUE", "y", "Y", "yes", "Yes", "YES", "on", "On", "ON":
		return true, nil
	case "0", "f", "F", "false", "False", "FALSE", "n", "N", "no", "No", "NO", "off", "Off", "OFF":
		return false, nil
	}

	return false, newConvErrorf("cannot parse %q as bool", v)
}

// boolFromReflect converts a reflect.Value to a boolean value based on its kind.
//
// Parameters:
//   - `from`: The input value to be converted.
//
// Returns:
//   - A boolean value representing the converted input.
//   - An error if the conversion fails.
func (c *Converter) boolFromReflect(from any) (bool, error) {
	value := indirectValue(reflect.ValueOf(from))
	if !value.IsValid() {
		if c.nilAsZero {
			return false, nil
		}
		return false, newConvError(from, "bool")
	}

	kind := value.Kind()
	switch {
	case kind == reflect.String:
		return c.stringToBool(value.String())
	case kind == reflect.Bool:
		return value.Bool(), nil
	case isKindInt(kind):
		return value.Int() != 0, nil
	case isKindUint(kind):
		return value.Uint() != 0, nil
	case isKindFloat(kind):
		return value.Float() != 0, nil
	case isKindComplex(kind):
		return real(value.Complex()) != 0, nil
	case isKindLength(kind):
		return value.Len() > 0, nil
	case kind == reflect.Struct && value.CanInterface():
		if t, ok := value.Interface().(time.Time); ok {
			return !t.IsZero(), nil
		}
	}

	return false, newConvError(from, "bool")
}
