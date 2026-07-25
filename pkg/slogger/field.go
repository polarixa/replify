package slogger

import (
	"fmt"
	"strconv"
	"time"

	"github.com/polarixa/replify/pkg/encoding"
	"github.com/polarixa/replify/pkg/strutil"
)

// ///////////////////////////////////////////////////////////////////////////
// Field accessors
// ///////////////////////////////////////////////////////////////////////////

// Key returns the field's name as it will appear in formatted output.
//
// Returns:
//
// the string key that identifies this field in a log entry.
func (f Field) Key() string {
	return f.key
}

// Type returns the concrete value type stored in this field.
// Callers rarely need this; it is primarily used by Formatter implementations.
//
// Returns:
//
// the FieldType constant that describes the stored value variant.
func (f Field) Type() FieldType {
	return f.typ
}

// FieldType returns the concrete value type stored in this field.
// Callers rarely need this; it is primarily used by Formatter implementations.
//
// Deprecated: Use Type() instead. This method will be removed in a future version.
//
// Returns:
//
// the FieldType constant that describes the stored value variant.
func (f Field) FieldType() FieldType {
	return f.Type()
}

// StringValue returns the raw string value stored in this field.
// This is only meaningful when Type() returns StringType, JSONType, or TimefType.
//
// Returns:
//
// the string value stored in this field.
func (f Field) StringValue() string {
	return f.strVal
}

// IntValue returns the raw int64 value stored in this field.
// This is only meaningful when Type() returns Int64Type, Int8Type, Int16Type, or Int32Type.
//
// Returns:
//
// the int64 value stored in this field.
func (f Field) IntValue() int64 {
	return f.intVal
}

// Uint64Value returns the raw uint64 value stored in this field.
// This is only meaningful when Type() returns UintType, Uint8Type, Uint16Type, Uint32Type, or Uint64Type.
//
// Returns:
//
// the uint64 value stored in this field.
func (f Field) Uint64Value() uint64 {
	return f.uint64Val
}

// FloatValue returns the raw float64 value stored in this field.
// This is only meaningful when Type() returns Float64Type or Float32Type.
//
// Returns:
//
// the float64 value stored in this field.
func (f Field) FloatValue() float64 {
	return f.floatVal
}

// BoolValue returns the raw bool value stored in this field.
// This is only meaningful when Type() returns BoolType.
//
// Returns:
//
// the bool value stored in this field.
func (f Field) BoolValue() bool {
	return f.boolVal
}

// ErrValue returns the raw error value stored in this field.
// This is only meaningful when Type() returns ErrorType.
//
// Returns:
//
// the error value stored in this field.
func (f Field) ErrValue() error {
	return f.errVal
}

// TimeValue returns the raw time.Time value stored in this field.
// This is only meaningful when Type() returns TimeType or TimefType.
//
// Returns:
//
// the time.Time value stored in this field.
func (f Field) TimeValue() time.Time {
	return f.timeVal
}

// DurationValue returns the raw time.Duration value stored in this field.
// This is only meaningful when Type() returns DurationType.
//
// Returns:
//
// the time.Duration value stored in this field.
func (f Field) DurationValue() time.Duration {
	return f.durVal
}

// AnyValue returns the raw interface{} value stored in this field.
// This is only meaningful when Type() returns AnyType.
//
// Returns:
//
// the interface{} value stored in this field.
func (f Field) AnyValue() any {
	return f.anyVal
}

// ///////////////////////////////////////////////////////////////////////////
// Field constructors — primitive types
// ///////////////////////////////////////////////////////////////////////////

// String constructs a Field carrying a string value.
//
// Parameters:
//   - `key`: the field name
//   - `val`: the string value
//
// Returns:
//
// a Field of type StringType.
func String(key, val string) Field {
	return Field{key: key, typ: StringType, strVal: val}
}

// Stringf constructs a Field carrying a formatted string value.
//
// Parameters:
//   - `key`: the field name
//   - `format`: the format string
//   - `args`: the arguments to format
//
// Returns:
//
// a Field of type StringType.
func Stringf(key, format string, args ...any) Field {
	return Field{key: key, typ: StringType, strVal: fmt.Sprintf(format, args...)}
}

// Bool constructs a Field carrying a bool value.
//
// Parameters:
//   - `key`: the field name
//   - `val`: the boolean value
//
// Returns:
//
// a Field of type BoolType.
func Bool(key string, val bool) Field {
	return Field{key: key, typ: BoolType, boolVal: val}
}

// ///////////////////////////////////////////////////////////////////////////
// Field constructors — signed integer types
// ///////////////////////////////////////////////////////////////////////////

// Int constructs a Field carrying an int value stored as int64.
//
// Parameters:
//   - `key`: the field name
//   - `val`: the integer value
//
// Returns:
//
// a Field of type Int64Type.
func Int(key string, val int) Field {
	return Field{key: key, typ: Int64Type, intVal: int64(val)}
}

// Int8 constructs a Field carrying an int8 value stored as int64.
//
// Parameters:
//   - `key`: the field name
//   - `val`: the int8 value
//
// Returns:
//
// a Field of type Int8Type.
func Int8(key string, val int8) Field {
	return Field{key: key, typ: Int8Type, intVal: int64(val)}
}

// Int16 constructs a Field carrying an int16 value stored as int64.
//
// Parameters:
//   - `key`: the field name
//   - `val`: the int16 value
//
// Returns:
//
// a Field of type Int16Type.
func Int16(key string, val int16) Field {
	return Field{key: key, typ: Int16Type, intVal: int64(val)}
}

// Int32 constructs a Field carrying an int32 value stored as int64.
//
// Parameters:
//   - `key`: the field name
//   - `val`: the int32 value
//
// Returns:
//
// a Field of type Int32Type.
func Int32(key string, val int32) Field {
	return Field{key: key, typ: Int32Type, intVal: int64(val)}
}

// Int64 constructs a Field carrying an int64 value.
//
// Parameters:
//   - `key`: the field name
//   - `val`: the int64 value
//
// Returns:
//
// a Field of type Int64Type.
func Int64(key string, val int64) Field {
	return Field{key: key, typ: Int64Type, intVal: val}
}

// ///////////////////////////////////////////////////////////////////////////
// Field constructors — unsigned integer types
// ///////////////////////////////////////////////////////////////////////////

// Uint constructs a Field carrying a uint value stored as uint64.
//
// Parameters:
//   - `key`: the field name
//   - `val`: the uint value
//
// Returns:
//
// a Field of type UintType.
func Uint(key string, val uint) Field {
	return Field{key: key, typ: UintType, uint64Val: uint64(val)}
}

// Uint8 constructs a Field carrying a uint8 value stored as uint64.
//
// Parameters:
//   - `key`: the field name
//   - `val`: the uint8 value
//
// Returns:
//
// a Field of type Uint8Type.
func Uint8(key string, val uint8) Field {
	return Field{key: key, typ: Uint8Type, uint64Val: uint64(val)}
}

// Uint16 constructs a Field carrying a uint16 value stored as uint64.
//
// Parameters:
//   - `key`: the field name
//   - `val`: the uint16 value
//
// Returns:
//
// a Field of type Uint16Type.
func Uint16(key string, val uint16) Field {
	return Field{key: key, typ: Uint16Type, uint64Val: uint64(val)}
}

// Uint32 constructs a Field carrying a uint32 value stored as uint64.
//
// Parameters:
//   - `key`: the field name
//   - `val`: the uint32 value
//
// Returns:
//
// a Field of type Uint32Type.
func Uint32(key string, val uint32) Field {
	return Field{key: key, typ: Uint32Type, uint64Val: uint64(val)}
}

// Uint64 constructs a Field carrying a uint64 value.
// Use this instead of Int64 when the value may exceed math.MaxInt64.
//
// Parameters:
//   - `key`: the field name
//   - `val`: the uint64 value
//
// Returns:
//
// a Field of type Uint64Type.
func Uint64(key string, val uint64) Field {
	return Field{key: key, typ: Uint64Type, uint64Val: val}
}

// ///////////////////////////////////////////////////////////////////////////
// Field constructors — floating-point types
// ///////////////////////////////////////////////////////////////////////////

// Float32 constructs a Field carrying a float32 value stored as float64.
//
// Parameters:
//   - `key`: the field name
//   - `val`: the float32 value
//
// Returns:
//
// a Field of type Float32Type.
func Float32(key string, val float32) Field {
	return Field{key: key, typ: Float32Type, floatVal: float64(val)}
}

// Float64 constructs a Field carrying a float64 value.
//
// Parameters:
//   - `key`: the field name
//   - `val`: the float64 value
//
// Returns:
//
// a Field of type Float64Type.
func Float64(key string, val float64) Field {
	return Field{key: key, typ: Float64Type, floatVal: val}
}

// ///////////////////////////////////////////////////////////////////////////
// Field constructors — time and duration
// ///////////////////////////////////////////////////////////////////////////

// Time constructs a Field carrying a time.Time value.
// The value is formatted using time.RFC3339 in formatters.
//
// Parameters:
//   - `key`: the field name
//   - `val`: the time value
//
// Returns:
//
// a Field of type TimeType.
func Time(key string, val time.Time) Field {
	return Field{key: key, typ: TimeType, timeVal: val}
}

// Timef constructs a Field carrying a time.Time value formatted with a custom
// layout string. Use this when the default RFC3339 format is not suitable.
//
// Parameters:
//   - `key`: the field name
//   - `val`: the time value
//   - `layout`: a Go time layout string (e.g. "2006/01/02 15:04:05")
//
// Returns:
//
// a Field of type TimefType.
func Timef(key string, val time.Time, layout string) Field {
	return Field{key: key, typ: TimefType, timeVal: val, strVal: layout}
}

// Duration constructs a Field carrying a time.Duration value.
//
// Parameters:
//   - `key`: the field name
//   - `val`: the duration value
//
// Returns:
//
// a Field of type DurationType.
func Duration(key string, val time.Duration) Field {
	return Field{key: key, typ: DurationType, durVal: val}
}

// ///////////////////////////////////////////////////////////////////////////
// Field constructors — error and generic types
// ///////////////////////////////////////////////////////////////////////////

// Err constructs a Field carrying an error under the "error" key.
//
// Parameters:
//   - `err`: the error to log (may be nil)
//
// Returns:
//
// a Field of type ErrorType with key "error".
func Err(err error) Field {
	return Field{key: "error", typ: ErrorType, errVal: err}
}

// Any constructs a Field carrying an arbitrary value.
// The value is rendered via fmt.Sprintf("%v", val) in text formatters and
// marshalled to JSON in JSON formatters.
//
// Parameters:
//   - `key`: the field name
//   - `val`: any value
//
// Returns:
//
// a Field of type AnyType.
func Any(key string, val any) Field {
	return Field{key: key, typ: AnyType, anyVal: val}
}

// JSON constructs a Field carrying a JSON-encoded representation of data.
// The value is encoded at construction time using pkg/encoding.MarshalJSONs.
// If encoding fails, the field stores the fmt.Sprintf("%v") fallback instead.
// This is useful for logging structured objects as compact JSON strings.
//
// Parameters:
//   - `key`: the field name
//   - `data`: any value; will be JSON-encoded at construction time
//
// Returns:
//
// a Field of type JSONType with the JSON string stored inline.
func JSON(key string, data any) Field {
	s := encoding.JSON(data)
	if strutil.IsEmpty(s) {
		s = fmt.Sprintf("%v", data)
	}
	return Field{key: key, typ: JSONType, strVal: s}
}

// ///////////////////////////////////////////////////////////////////////////
// Field value rendering
// ///////////////////////////////////////////////////////////////////////////

// Value returns the field's value formatted as a string for use by formatters.
//
// Returns:
//
// the string representation of the stored value. The format depends on the
// field type: numeric types use decimal notation; ErrorType returns err.Error()
// or "<nil>"; TimeType uses time.RFC3339; TimefType uses the stored layout;
// DurationType uses Duration.String(); JSONType returns the pre-encoded JSON
// string; AnyType uses fmt.Sprintf("%v", val).
func (f *Field) Value() string {
	switch f.typ {
	case StringType, JSONType:
		return f.strVal
	case Int64Type, Int8Type, Int16Type, Int32Type:
		return strconv.FormatInt(f.intVal, 10)
	case UintType, Uint8Type, Uint16Type, Uint32Type, Uint64Type:
		return strconv.FormatUint(f.uint64Val, 10)
	case Float64Type:
		return strconv.FormatFloat(f.floatVal, 'f', -1, 64)
	case Float32Type:
		return strconv.FormatFloat(f.floatVal, 'f', -1, 32)
	case BoolType:
		if f.boolVal {
			return "true"
		}
		return "false"
	case ErrorType:
		if f.errVal == nil {
			return "<nil>"
		}
		return f.errVal.Error()
	case TimeType:
		if f.timeVal.IsZero() {
			return ""
		}
		return f.timeVal.Format(defaultTimeFormat) // time.RFC3339
	case TimefType:
		if f.timeVal.IsZero() {
			return ""
		}
		if strutil.IsEmpty(f.strVal) {
			return f.timeVal.Format(defaultTimeFormat) // time.RFC3339
		}
		return f.timeVal.Format(f.strVal)
	case DurationType:
		return f.durVal.String()
	case AnyType:
		return fmt.Sprintf("%v", f.anyVal)
	default:
		return ""
	}
}
