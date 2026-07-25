package fj

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/polarixa/replify/pkg/conv"
	"github.com/polarixa/replify/pkg/encoding"
	"github.com/polarixa/replify/pkg/strutil"
)

//////////////////////////////////
/// Type enum / representation //
////////////////////////////////

// String provides a string representation of the `Type` enumeration.
//
// This method converts the `Type` value into a human-readable string.
// It is particularly useful for debugging or logging purposes.
//
// Mapping of `Type` values to strings:
//   - Null: "Null"
//   - False: "False"
//   - Number: "Number"
//   - String: "String"
//   - True: "True"
//   - JSON: "JSON"
//   - Default (unknown type): An empty string is returned.
//
// Returns:
//   - string: A string representation of the `Type` value.
//
// Example Usage:
//
//	var t Type = True
//	fmt.Println(t.String())  // Output: "True"
func (t Type) String() string {
	switch t {
	default:
		return ""
	case Null:
		return "Null"
	case False:
		return "False"
	case Number:
		return "Number"
	case String:
		return "String"
	case True:
		return "True"
	case JSON:
		return "JSON"
	}
}

//////////////////////////////////
/// Core identity & raw access //
//////////////////////////////////

// Kind returns the JSON type of the Context.
// It provides the specific type of the JSON value, such as String, Number, Object, etc.
//
// Returns:
//   - Type: The type of the JSON value represented by the Context.
func (ctx Context) Kind() Type {
	return ctx.kind
}

// Raw returns the raw, unprocessed JSON string for the Context.
// This can be useful for inspecting the original data without any parsing or transformations.
//
// Returns:
//   - string: The unprocessed JSON string.
func (ctx Context) Raw() string {
	return ctx.raw
}

// Number returns the numeric value of the Context, if applicable.
// This is relevant when the Context represents a JSON number.
//
// Returns:
//   - float64: The numeric value of the Context.
//     If the Context does not represent a number, the value may be undefined.
func (ctx Context) Number() float64 {
	return ctx.num
}

// Index returns the index of the unprocessed JSON value in the original JSON string.
// This can be used to track the position of the value in the source data.
// If the index is unknown, it defaults to 0.
//
// Returns:
//   - int: The position of the value in the original JSON string.
func (ctx Context) Index() int {
	return ctx.idx
}

// Indexes returns a slice of indices for elements matching a path containing the '#' character.
// This is useful for handling path queries that involve multiple matches.
//
// Returns:
//   - []int: A slice of indices for matching elements.
func (ctx Context) Indexes() []int {
	return ctx.idxs
}

// Exists returns true if the value exists (i.e., it is not Null and contains data).
//
// Example Usage:
//
//	if fj.Get(json, "user.name").Exists() {
//	  println("value exists")
//	}
//
// Returns:
//   - bool: Returns true if the value is not null and contains non-empty data, otherwise returns false.
func (ctx Context) Exists() bool {
	return ctx.kind != Null || len(ctx.raw) != 0
}

// IsError checks if there is an error associated with the Context.
//
// This function checks whether the `err` field of the Context is set. If there is
// an error (i.e., `err` is not `nil`), the function returns `true`; otherwise, it returns `false`.
//
// Example Usage:
//
//	ctx := Context{err: fmt.Errorf("invalid JSON")}
//	fmt.Println(ctx.IsError()) // Output: true
//
//	ctx = Context{}
//	fmt.Println(ctx.IsError()) // Output: false
//
// Returns:
//   - bool: `true` if the Context has an error, `false` otherwise.
func (ctx Context) IsError() bool {
	return ctx.err != nil
}

// Cause returns the error message if there is an error in the Context.
//
// If the Context has an error (i.e., `err` is not `nil`), this function returns
// the error message as a string. If there is no error, it returns an empty string.
//
// Example Usage:
//
//	ctx := Context{err: fmt.Errorf("parsing error")}
//	fmt.Println(ctx.Cause()) // Output: "parsing error"
//
//	ctx = Context{}
//	fmt.Println(ctx.Cause()) // Output: ""
//
// Returns:
//   - string: The error message if there is an error, or an empty string if there is no error.
func (ctx Context) Cause() string {
	if ctx.IsError() {
		return ctx.err.Error()
	}
	return ""
}

//////////////////////////////////
/// Type predicates            ///
//////////////////////////////////

// IsArray checks if the current `Context` represents a JSON array.
//
// A value is considered a JSON array if:
//   - The `kind` is `JSON`.
//   - The `raw` string starts with the `[` character.
//
// Returns:
//   - bool: Returns `true` if the `Context` is a JSON array; otherwise, `false`.
//
// Example Usage:
//
//	ctx := Context{kind: JSON, raw: "[1, 2, 3]"}
//	isArr := ctx.IsArray()
//	// isArr: true
//
//	ctx = Context{kind: JSON, raw: "{"key": "value"}"}
//	isArr = ctx.IsArray()
//	// isArr: false
func (ctx Context) IsArray() bool {
	return ctx.kind == JSON && len(ctx.raw) > 0 && ctx.raw[0] == '['
}

// IsObject checks if the current `Context` represents a JSON object.
//
// A value is considered a JSON object if:
//   - The `kind` is `JSON`.
//   - The `raw` string starts with the `{` character.
//
// Returns:
//   - bool: Returns `true` if the `Context` is a JSON object; otherwise, `false`.
//
// Example Usage:
//
//	ctx := Context{kind: JSON, raw: "{"key": "value"}"}
//	isObj := ctx.IsObject()
//	// isObj: true
//
//	ctx = Context{kind: JSON, raw: "[1, 2, 3]"}
//	isObj = ctx.IsObject()
//	// isObj: false
func (ctx Context) IsObject() bool {
	return ctx.kind == JSON && len(ctx.raw) > 0 && ctx.raw[0] == '{'
}

// IsBool checks if the current `Context` represents a JSON boolean value.
//
// A value is considered a JSON boolean if:
//   - The `kind` is `True` or `False`.
//
// Returns:
//   - bool: Returns `true` if the `Context` is a JSON boolean; otherwise, `false`.
//
// Example Usage:
//
//	ctx := Context{kind: True}
//	isBool := ctx.IsBool()
//	// isBool: true
//
//	ctx = Context{kind: String, strings: "true"}
//	isBool = ctx.IsBool()
//	// isBool: false
func (ctx Context) IsBool() bool {
	return ctx.kind == True || ctx.kind == False
}

//////////////////////////////////
/// Conversions to Go types    ///
//////////////////////////////////

// Bool converts the Context value into a boolean representation.
// The conversion depends on the JSON type of the Context:
//   - For `True` type: Returns `true`.
//   - For `String` type: Attempts to parse the string as a boolean (case-insensitive).
//     If parsing fails, defaults to `false`.
//   - For `Number` type: Returns `true` if the numeric value is non-zero, otherwise `false`.
//   - For all other types: Returns `false`.
//
// Returns:
//   - bool: A boolean representation of the Context value.
func (ctx Context) Bool() bool {
	switch ctx.kind {
	default:
		return false
	case True:
		return true
	case String:
		return conv.BoolOrDefault(ctx.str, false)
	case Number:
		return ctx.num != 0
	}
}

//////////////////////////////////
/// Signed integers            ///
//////////////////////////////////

// Int converts the Context value into an integer representation (int).
// The conversion depends on the JSON type of the Context:
//   - For `True` type: Returns 1.
//   - For `String` type: Attempts to parse the string into an integer. Defaults to 0 on failure.
//   - For `Number` type:
//   - Directly converts the numeric value to an integer if it's within the safe range.
//   - Parses the unprocessed string for integer values as a fallback.
//   - Defaults to converting the float64 numeric value to an int.
//
// Returns:
//   - int: An integer representation of the Context value.
func (ctx Context) Int() int {
	switch ctx.kind {
	default:
		return 0
	case True:
		return 1
	case String:
		return conv.IntOrDefault(ctx.str, 0)
	case Number:
		if ctx.num < math.MinInt32 || ctx.num > math.MaxInt32 {
			return 0
		}
		val, err := conv.Int(ctx.raw)
		if err == nil {
			return val
		}
		return conv.IntOrDefault(ctx.num, 0)
	}
}

// Int8 converts the Context value into an 8-bit integer representation (int8).
// The conversion depends on the JSON type of the Context:
//   - For `True` type: Returns 1.
//   - For `String` type: Attempts to parse the string into an 8-bit integer. Defaults to 0 on failure.
//   - For `Number` type:
//   - Directly converts the numeric value to an 8-bit integer if it's within the safe range.
//   - Parses the unprocessed string for 8-bit integer values as a fallback.
//   - Defaults to converting the float64 numeric value to an int8.
//
// Returns:
//   - int8: An 8-bit integer representation of the Context value.
func (ctx Context) Int8() int8 {
	switch ctx.kind {
	default:
		return 0
	case True:
		return 1
	case String:
		return conv.Int8OrDefault(ctx.str, 0)
	case Number:
		if ctx.num < math.MinInt8 || ctx.num > math.MaxInt8 {
			return 0
		}
		val, err := conv.Int8(ctx.raw)
		if err == nil {
			return val
		}
		return conv.Int8OrDefault(ctx.num, 0)
	}
}

// Int16 converts the Context value into a 16-bit integer representation (int16).
// The conversion depends on the JSON type of the Context:
//   - For `True` type: Returns 1.
//   - For `String` type: Attempts to parse the string into a 16-bit integer. Defaults to 0 on failure.
//   - For `Number` type:
//   - Directly converts the numeric value to a 16-bit integer if it's within the safe range.
//   - Parses the unprocessed string for 16-bit integer values as a fallback.
//   - Defaults to converting the float64 numeric value to an int16.
//
// Returns:
//   - int16: A 16-bit integer representation of the Context value.
func (ctx Context) Int16() int16 {
	switch ctx.kind {
	default:
		return 0
	case True:
		return 1
	case String:
		return conv.Int16OrDefault(ctx.str, 0)
	case Number:
		if ctx.num < math.MinInt16 || ctx.num > math.MaxInt16 {
			return 0
		}
		val, err := conv.Int16(ctx.raw)
		if err == nil {
			return val
		}
		return conv.Int16OrDefault(ctx.num, 0)
	}
}

// Int32 converts the Context value into a 32-bit integer representation (int32).
// The conversion depends on the JSON type of the Context:
//   - For `True` type: Returns 1.
//   - For `String` type: Attempts to parse the string into a 32-bit integer. Defaults to 0 on failure.
//   - For `Number` type:
//   - Directly converts the numeric value to a 32-bit integer if it's within the safe range.
//   - Parses the unprocessed string for 32-bit integer values as a fallback.
//   - Defaults to converting the float64 numeric value to an int32.
//
// Returns:
//   - int32: A 32-bit integer representation of the Context value.
func (ctx Context) Int32() int32 {
	switch ctx.kind {
	default:
		return 0
	case True:
		return 1
	case String:
		return conv.Int32OrDefault(ctx.str, 0)
	case Number:
		if ctx.num < math.MinInt32 || ctx.num > math.MaxInt32 {
			return 0
		}
		val, err := conv.Int32(ctx.raw)
		if err == nil {
			return val
		}
		return conv.Int32OrDefault(ctx.num, 0)
	}
}

// Int64 converts the Context value into an integer representation (int64).
// The conversion depends on the JSON type of the Context:
//   - For `True` type: Returns 1.
//   - For `String` type: Attempts to parse the string into an integer. Defaults to 0 on failure.
//   - For `Number` type:
//   - Directly converts the numeric value to an integer if it's safe.
//   - Parses the unprocessed string for integer values as a fallback.
//   - Defaults to converting the float64 numeric value to an int64.
//
// Returns:
//   - int64: An integer representation of the Context value.
func (ctx Context) Int64() int64 {
	switch ctx.kind {
	default:
		return 0
	case True:
		return 1
	case String:
		return conv.Int64OrDefault(ctx.str, 0)
	case Number:
		if ctx.num < math.MinInt64 || ctx.num > math.MaxInt64 {
			return 0
		}
		val, err := conv.Int64(ctx.raw)
		if err == nil {
			return val
		}
		return conv.Int64OrDefault(ctx.num, 0)
	}
}

//////////////////////////////////
/// Unsigned integers          ///
//////////////////////////////////

// Uint converts the Context value into an unsigned integer representation (uint).
// The conversion depends on the JSON type of the Context:
//   - For `True` type: Returns 1.
//   - For `String` type: Attempts to parse the string into an unsigned integer. Defaults to 0 on failure.
//   - For `Number` type:
//   - Directly converts the numeric value to a uint if it's safe and non-negative.
//   - Parses the unprocessed string for unsigned integer values as a fallback.
//   - Defaults to converting the float64 numeric value to a uint.
//
// Returns:
//   - uint: An unsigned integer representation of the Context value.
func (ctx Context) Uint() uint {
	switch ctx.kind {
	default:
		return 0
	case True:
		return 1
	case String:
		return conv.UintOrDefault(ctx.str, 0)
	case Number:
		if ctx.num < 0 || ctx.num > math.MaxUint {
			return 0
		}
		val, err := conv.Uint(ctx.raw)
		if err == nil {
			return val
		}
		return conv.UintOrDefault(ctx.num, 0)
	}
}

// Uint8 converts the Context value into an 8-bit unsigned integer representation (uint8).
// The conversion depends on the JSON type of the Context:
//   - For `True` type: Returns 1.
//   - For `String` type: Attempts to parse the string into an 8-bit unsigned integer. Defaults to 0 on failure.
//   - For `Number` type:
//   - Directly converts the numeric value to a uint8 if it's safe and non-negative.
//   - Parses the unprocessed string for 8-bit unsigned integer values as a fallback.
//   - Defaults to converting the float64 numeric value to a uint8.
//
// Returns:
//   - uint8: An 8-bit unsigned integer representation of the Context value.
func (ctx Context) Uint8() uint8 {
	switch ctx.kind {
	default:
		return 0
	case True:
		return 1
	case String:
		return conv.Uint8OrDefault(ctx.str, 0)
	case Number:
		if ctx.num < 0 || ctx.num > math.MaxUint8 {
			return 0
		}
		val, err := conv.Uint8(ctx.raw)
		if err == nil {
			return val
		}
		return conv.Uint8OrDefault(ctx.num, 0)
	}
}

// Uint16 converts the Context value into a 16-bit unsigned integer representation (uint16).
// The conversion depends on the JSON type of the Context:
//   - For `True` type: Returns 1.
//   - For `String` type: Attempts to parse the string into a 16-bit unsigned integer. Defaults to 0 on failure.
//   - For `Number` type:
//   - Directly converts the numeric value to a uint16 if it's safe and non-negative.
//   - Parses the unprocessed string for 16-bit unsigned integer values as a fallback.
//   - Defaults to converting the float64 numeric value to a uint16.
//
// Returns:
//   - uint16: A 16-bit unsigned integer representation of the Context value.
func (ctx Context) Uint16() uint16 {
	switch ctx.kind {
	default:
		return 0
	case True:
		return 1
	case String:
		return conv.Uint16OrDefault(ctx.str, 0)
	case Number:
		if ctx.num < 0 || ctx.num > math.MaxUint16 {
			return 0
		}
		val, err := conv.Uint16(ctx.raw)
		if err == nil {
			return val
		}
		return conv.Uint16OrDefault(ctx.num, 0)
	}
}

// Uint32 converts the Context value into a 32-bit unsigned integer representation (uint32).
// The conversion depends on the JSON type of the Context:
//   - For `True` type: Returns 1.
//   - For `String` type: Attempts to parse the string into a 32-bit unsigned integer. Defaults to 0 on failure.
//   - For `Number` type:
//   - Directly converts the numeric value to a uint32 if it's safe and non-negative.
//   - Parses the unprocessed string for 32-bit unsigned integer values as a fallback.
//   - Defaults to converting the float64 numeric value to a uint32.
//
// Returns:
//   - uint32: A 32-bit unsigned integer representation of the Context value.
func (ctx Context) Uint32() uint32 {
	switch ctx.kind {
	default:
		return 0
	case True:
		return 1
	case String:
		return conv.Uint32OrDefault(ctx.str, 0)
	case Number:
		if ctx.num < 0 || ctx.num > math.MaxUint32 {
			return 0
		}
		val, err := conv.Uint32(ctx.raw)
		if err == nil {
			return val
		}
		return conv.Uint32OrDefault(ctx.num, 0)
	}
}

// Uint64 converts the Context value into an unsigned integer representation (uint64).
// The conversion depends on the JSON type of the Context:
//   - For `True` type: Returns 1.
//   - For `String` type: Attempts to parse the string into an unsigned integer. Defaults to 0 on failure.
//   - For `Number` type:
//   - Directly converts the numeric value to a uint64 if it's safe and non-negative.
//   - Parses the unprocessed string for unsigned integer values as a fallback.
//   - Defaults to converting the float64 numeric value to a uint64.
//
// Returns:
//   - uint64: An unsigned integer representation of the Context value.
func (ctx Context) Uint64() uint64 {
	switch ctx.kind {
	default:
		return 0
	case True:
		return 1
	case String:
		return conv.Uint64OrDefault(ctx.str, 0)
	case Number:
		if ctx.num < 0 || ctx.num > math.MaxUint64 {
			return 0
		}
		val, err := conv.Uint64(ctx.raw)
		if err == nil {
			return val
		}
		return conv.Uint64OrDefault(ctx.num, 0)
	}
}

//////////////////////////////////
///           Floats           ///
//////////////////////////////////

// Float32 converts the Context value into a floating-point representation (Float32).
// This function provides a similar conversion mechanism as Float64 but with Float32 precision.
// The conversion depends on the JSON type of the Context:
//   - For `True` type: Returns 1 as a Float32 value.
//   - For `String` type: Attempts to parse the string as a floating-point number (Float32 precision).
//     If the parsing fails, it defaults to 0.
//   - For `Number` type: Returns the numeric value as a Float32, assuming the Context contains
//     a numeric value in its `num` field.
//
// Returns:
//   - Float32: A floating-point representation of the Context value.
//
// Example Usage:
//
//	ctx := Context{kind: String, strings: "123.45"}
//	result := ctx.Float32()
//	// result: 123.45 (parsed as Float32)
//
//	ctx = Context{kind: True}
//	result = ctx.Float32()
//	// result: 1 (True is represented as 1.0)
//
//	ctx = Context{kind: Number, numeric: 678.9}
//	result = ctx.Float32()
//	// result: 678.9 (as Float32)
//
// Details:
//
//   - For the `True` type, the function always returns 1.0, representing the boolean `true` value.
//
//   - For the `String` type, it uses `strconv.ParseFloat` with 32-bit precision to convert the string
//     into a Float32. If parsing fails (e.g., if the string is not a valid numeric representation),
//     the function returns 0 as a fallback.
//
//   - For the `Number` type, the `num` field, assumed to hold a float64 value, is converted
//     to a Float32 for the return value.
//
// Notes:
//
//   - The function gracefully handles invalid string inputs for the `String` type by returning 0,
//     ensuring no runtime panic occurs due to a parsing error.
//
//   - Precision may be lost when converting from float64 (`num` field) to Float32.
func (ctx Context) Float32() float32 {
	switch ctx.kind {
	default:
		return 0
	case True:
		return 1
	case String:
		return conv.Float32OrDefault(ctx.str, 0)
	case Number:
		return float32(ctx.num)
	}
}

// Float64 converts the Context value into a floating-point representation (float64).
// The conversion depends on the JSON type of the Context:
//   - For `True` type: Returns 1.
//   - For `String` type: Attempts to parse the string as a floating-point number. Defaults to 0 on failure.
//   - For `Number` type: Returns the numeric value as a float64.
//
// Returns:
//   - float64: A floating-point representation of the Context value.
func (ctx Context) Float64() float64 {
	switch ctx.kind {
	default:
		return 0
	case True:
		return 1
	case String:
		return conv.Float64OrDefault(ctx.str, 0)
	case Number:
		return ctx.num
	}
}

//////////////////////////////////
///        Time / duration     ///
//////////////////////////////////

// Time converts the Context value into a time.Time representation.
// The conversion interprets the Context value as a string in RFC3339 format.
// If parsing fails, the zero time (0001-01-01 00:00:00 UTC) is returned.
//
// Returns:
//   - time.Time: A time.Time representation of the Context value.
//     Defaults to the zero time if parsing fails.
func (ctx Context) Time() time.Time {
	return conv.TimeOrDefault(ctx.String(), time.Time{})
}

// WithTime parses the Context value into a time.Time representation using a custom format.
// This function allows for greater flexibility by enabling parsing with user-defined
// date and time formats, rather than relying on the fixed RFC3339 format used in `Time()`.
//
// Parameters:
//   - format: A string representing the desired format to parse the Context value.
//     This format must conform to the layouts supported by the `time.Parse` function.
//
// Returns:
//   - time.Time: The parsed time.Time representation of the Context value if parsing succeeds.
//   - error: An error value if the parsing fails (e.g., due to an invalid format or mismatched value).
//
// Example Usage:
//
//	ctx := Context{kind: String, strings: "12-25-2023"}
//	t, err := ctx.WithTime("01-02-2006")
//	if err != nil {
//	    fmt.Println("Error parsing time:", err)
//	} else {
//	    fmt.Println("Parsed time:", t)
//	}
//	// Output: Parsed time: 2023-12-25 00:00:00 +0000 UTC
//
// Details:
//
//   - The function relies on the `time.Parse` function to convert the string value from the Context
//     into a time.Time representation.
//
//   - The format parameter determines how the Context value should be interpreted. It must match
//     the layout of the Context's string value. If the value cannot be parsed according to the given
//     format, the function returns an error.
//
//   - Unlike `Time()`, which defaults to the zero time on failure, this function explicitly returns
//     an error to indicate parsing issues.
//
// Notes:
//
//   - The function assumes that `ctx.String()` returns a string representation of the Context value.
//     If the Context does not contain a valid string, the parsing will fail.
//
//   - This function is ideal for cases where date and time formats vary or when the RFC3339 standard
//     is not suitable.
//
//   - To handle parsing failures gracefully, always check the returned error before using the
//     resulting `time.Time` value.
func (ctx Context) WithTime(format string) (time.Time, error) {
	return time.Parse(format, ctx.String())
}

// Duration converts the Context value into a time.Duration representation.
// The conversion depends on the JSON type of the Context:
//   - For `Number` type: Returns the numeric value as a time.Duration.
//   - For `String` type: Attempts to parse the string as a duration using `time.ParseDuration`.
//     If parsing fails, defaults to `time.Duration(0)`.
//   - For all other types: Returns `time.Duration(0)`.
//
// Returns:
//   - time.Duration: A time.Duration representation of the Context value.
//     Defaults to `time.Duration(0)` if parsing fails.
func (ctx Context) Duration() time.Duration {
	return conv.DurationOrDefault(ctx.String(), time.Duration(0))
}

//////////////////////////////////
///          Generic           ///
//////////////////////////////////

// Value returns the corresponding Go type for the JSON value represented by the Context.
//
// The function returns one of the following types based on the JSON value:
//   - bool for JSON booleans (True or False)
//   - float64 for JSON numbers
//   - string for JSON string literals
//   - nil for JSON null
//   - map[string]interface{} for JSON objects
//   - []interface{} for JSON arrays
//
// Example Usage:
//
//	value := ctx.Value()
//	switch v := value.(type) {
//	  case bool:
//	    fmt.Println("Boolean:", v)
//	  case float64:
//	    fmt.Println("Number:", v)
//	  case string:
//	    fmt.Println("String:", v)
//	  case nil:
//	    fmt.Println("Null value")
//	  case map[string]interface{}:
//	    fmt.Println("Object:", v)
//	  case []interface{}:
//	    fmt.Println("Array:", v)
//	}
//
// Returns:
//
//   - interface{}: Returns the corresponding Go type for the JSON value, or nil if the type is not recognized.
func (ctx Context) Value() any {
	if ctx.kind == String {
		return ctx.str
	}
	switch ctx.kind {
	default:
		return nil
	case False:
		return false
	case Number:
		return ctx.num
	case JSON:
		r := ctx.parseJSONElements(0, true)

		switch r.vn {
		case '{':
			return r.opResults
		case '[':
			return r.elems
		}
		return nil
	case True:
		return true
	}
}

/////////////////////////////////////////////////
///    Stringification / display   //////////////
/////////////////////////////////////////////////

// String returns a string representation of the Context value.
// The output depends on the JSON type of the Context:
//   - For `False` type: Returns "false".
//   - For `True` type: Returns "true".
//   - For `Number` type: Returns the numeric value as a string.
//     If the numeric value was calculated, it formats the float value.
//     Otherwise, it preserves the original unprocessed string if valid.
//   - For `String` type: Returns the string value.
//   - For `JSON` type: Returns the raw unprocessed JSON string.
//   - For other types: Returns an empty string.
//
// Returns:
//   - string: A string representation of the Context value.
func (ctx Context) String() string {
	switch ctx.kind {
	default:
		return ""
	case True:
		return "true"
	case False:
		return "false"
	case String:
		return ctx.str
	case JSON:
		return ctx.raw
	case Number:
		if len(ctx.raw) == 0 {
			return strconv.FormatFloat(ctx.num, 'f', -1, 64)
		}
		var i int
		if ctx.raw[0] == '-' {
			i++
		}
		for ; i < len(ctx.raw); i++ {
			if ctx.raw[i] < '0' || ctx.raw[i] > '9' {
				return strconv.FormatFloat(ctx.num, 'f', -1, 64)
			}
		}
		return ctx.raw
	}
}

// StringColored returns a colored string representation of the Context value.
// It applies the default style defined in `defaultStyle` to the string
// representation of the Context value.
//
// Details:
//   - The function first retrieves the plain string representation using `ctx.String()`.
//   - If the string is empty (determined by `isEmpty`), it returns an empty string.
//   - Otherwise, it applies the coloring rules from `defaultStyle` using the
//     `encoding.Color` function.
//
// Returns:
//   - string: A colored string representation of the Context value if not empty.
//     Returns an empty string for empty or invalid Context values.
//
// Example Usage:
//
//	ctx := Context{kind: True}
//	fmt.Println(ctx.StringColored()) // Output: "\033[1;35mtrue\033[0m" (colored)
//
// Notes:
//   - Requires the `encoding` library for styling and the `isEmpty` utility function
//     to check for empty strings.
func (ctx Context) StringColored() string {
	s := []byte(ctx.String())
	if strutil.IsEmpty(string(s)) {
		return ""
	}
	return string(encoding.Color(s, defaultStyle))
}

// WithStringColored applies a customizable colored styling to the string representation of the Context value.
//
// This function enhances the default coloring functionality by allowing the caller to specify a custom
// style for highlighting the Context value. If no custom style is provided, the default styling rules
// (`defaultStyle`) are used.
//
// Parameters:
//   - style (*encoding.Style): A pointer to a Style structure that defines custom styling rules
//     for JSON elements. If `style` is nil, the `defaultStyle` is applied.
//
// Details:
//   - Retrieves the plain string representation of the Context value using `ctx.String()`.
//   - Checks if the string is empty using the `isEmpty` utility function. If empty, it returns
//     an empty string immediately.
//   - If a custom style is provided, it applies the given style to the string representation
//     using the `encoding.Color` function. Otherwise, it applies the default style.
//
// Returns:
//   - string: A styled string representation of the Context value based on the provided or default style.
//
// Example Usage:
//
//	customStyle := &encoding.Style{
//	    Key:      [2]string{"\033[1;36m", "\033[0m"},
//	    String:   [2]string{"\033[1;33m", "\033[0m"},
//	    // Additional styling rules...
//	}
//
//	ctx := Context{kind: True}
//	fmt.Println(ctx.WithStringColored(customStyle)) // Output: "\033[1;35mtrue\033[0m" (custom colored)
//
// Notes:
//   - The function uses the `encoding.Color` utility to apply the color rules defined in the style.
//   - Requires the `isEmpty` utility function to check for empty strings.
func (ctx Context) WithStringColored(style *encoding.Style) string {
	s := []byte(ctx.String())
	if strutil.IsEmpty(string(s)) {
		return ""
	}
	if style == nil {
		style = defaultStyle
	}
	return string(encoding.Color(s, style))
}

/////////////////////////////////////////////////
///  Structural extraction & traversal   ////////
/////////////////////////////////////////////////

// Array returns an array of `Context` values derived from the current `Context`.
//
// Behavior:
//   - If the current `Context` represents a `Null` value, it returns an empty array.
//   - If the current `Context` is not a JSON array, it returns an array containing itself as a single element.
//   - If the current `Context` is a JSON array, it parses and returns the array's elements.
//
// Returns:
//   - []Context: A slice of `Context` values representing the array elements.
//
// Example Usage:
//
//	ctx := Context{kind: Null}
//	arr := ctx.Array()
//	// arr: []
//
//	ctx = Context{kind: JSON, raw: "[1, 2, 3]"}
//	arr = ctx.Array()
//	// arr: [Context, Context, Context]
//
// Notes:
//   - This function uses `parseJSONElements` internally to extract array elements.
//   - If the JSON is malformed or does not represent an array, the behavior may vary.
func (ctx Context) Array() []Context {
	if ctx.kind == Null {
		return []Context{}
	}
	if !ctx.IsArray() {
		return []Context{ctx}
	}
	r := ctx.parseJSONElements('[', false)
	return r.arr
}

// Map returns a map of values extracted from a JSON object.
//
// The function assumes that the `Context` represents a JSON object. It parses the JSON object and returns a map
// where the keys are strings, and the values are `Context` elements representing the corresponding JSON values.
//
// If the `Context` does not represent a valid JSON object, the function will return an empty map.
//
// Parameters:
//   - ctx: The `Context` instance that holds the raw JSON string. The function checks if the context represents
//     a JSON object and processes it accordingly.
//
// Returns:
//   - map[string]Context: A map where the keys are strings (representing the keys in the JSON object), and
//     the values are `Context` instances representing the corresponding JSON values. If the context does not represent
//     a valid JSON object, a nil is returned.
//
// Example Usage:
//
//	ctx := Context{kind: JSON, raw: "{\"key1\": \"value1\", \"key2\": 42}"}
//	result := ctx.Map()
//	// result.OpMap contains the parsed key-value pairs: {"key1": "value1", "key2": 42}
//
// Notes:
//   - The function calls `parseJSONElements` with the expected JSON object indicator ('{') to parse the JSON.
//   - If the `Context` is not a valid JSON object, it returns an empty map, which can be used to safely handle errors.
func (ctx Context) Map() map[string]Context {
	if ctx.kind != JSON {
		return nil
	}
	e := ctx.parseJSONElements('{', false)
	return e.ops
}

// Foreach iterates through the values of a JSON object or array, applying the provided iterator function.
//
// If the `Context` represents a non existent value (Null or invalid JSON), no iteration occurs.
// For JSON objects, the iterator receives both the key and value of each item.
// For JSON arrays, the iterator receives only the value of each item.
// If the `Context` is not an array or object, the iterator is called once with the whole value.
//
// Example Usage:
//
//	ctx.Foreach(func(key, value Context) bool {
//	  if key.str != "" {
//	    fmt.Printf("Key: %s, Value: %v\n", key.str, value)
//	  } else {
//	    fmt.Printf("Value: %v\n", value)
//	  }
//	  return true // Continue iteration
//	})
//
// Parameters:
//   - iterator: A function that receives a `key` (for objects) and `value` (for both objects and arrays).
//     The function should return `true` to continue iteration or `false` to stop.
//
// Notes:
//   - If the result is a JSON object, the iterator receives key-value pairs.
//   - If the result is a JSON array, the iterator receives only the values.
//   - If the result is not an object or array, the iterator is invoked once with the value.
//
// Returns:
//   - None. The iteration continues until all items are processed or the iterator returns `false`.
func (ctx Context) Foreach(iterator func(key, value Context) bool) {
	if !ctx.Exists() {
		return
	}
	if ctx.kind != JSON {
		iterator(Context{}, ctx)
		return
	}
	json := ctx.raw
	var obj bool
	var i int
	var key, value Context
	for ; i < len(json); i++ {
		if json[i] == '{' {
			i++
			key.kind = String
			obj = true
			break
		} else if json[i] == '[' {
			i++
			key.kind = Number
			key.num = -1
			break
		}
		if json[i] > ' ' {
			return
		}
	}
	var str string
	var _esc bool
	var ok bool
	var idx int
	for ; i < len(json); i++ {
		if obj {
			if json[i] != '"' {
				continue
			}
			s := i
			i, str, _esc, ok = extractJSONString(json, i+1)
			if !ok {
				return
			}
			if _esc {
				key.str = unescape(str[1 : len(str)-1])
			} else {
				key.str = str[1 : len(str)-1]
			}
			key.raw = str
			key.idx = s + ctx.idx
		} else {
			key.num += 1
		}
		for ; i < len(json); i++ {
			if json[i] <= ' ' || json[i] == ',' || json[i] == ':' {
				continue
			}
			break
		}
		s := i
		i, value, ok = extractJSONValueAt(json, i, true)
		if !ok {
			return
		}
		if ctx.idxs != nil {
			if idx < len(ctx.idxs) {
				value.idx = ctx.idxs[idx]
			}
		} else {
			value.idx = s + ctx.idx
		}
		if !iterator(key, value) {
			return
		}
		idx++
	}
}

//////////////////////////////////
///          Querying          ///
//////////////////////////////////

// Get searches for a specified path within a JSON structure and returns the corresponding result.
//
// This function allows you to search for a specific path in the JSON structure and retrieve the corresponding
// value as a `Context`. The path is represented as a string and can be used to navigate nested arrays or objects.
//
// The `path` parameter specifies the JSON path to search for, and the function will attempt to retrieve the value
// associated with that path. The result is returned as a `Context`, which contains information about the matched
// JSON value, including its type, string representation, numeric value, and index in the original JSON.
//
// Parameters:
//   - path: A string representing the path in the JSON structure to search for. The path may include array indices
//     and object keys separated by dots or brackets (e.g., "user.name", "items[0].price").
//
// Returns:
//   - Context: A `Context` instance containing the result of the search. The `Context` may represent various types of
//     JSON values (e.g., string, number, object, array). If no match is found, the `Context` will be empty.
//
// Example Usage:
//
//	ctx := Context{kind: JSON, raw: "{\"user\": {\"name\": \"John\"}, \"items\": [1, 2, 3]}"}
//	result := ctx.Get("user.name")
//	// result.str will contain "John", representing the value found at the "user.name" path.
//
// Notes:
//   - The function uses the `Get` function (presumably another function) to process the `raw` JSON string
//     and search for the specified path.
//   - The function adjusts the indices of the results (if any) to account for the original position of the `Context`
//     in the JSON string.
func (ctx Context) Get(path string) Context {
	q := Get(ctx.raw, path)
	if q.idxs != nil {
		for i := 0; i < len(q.idxs); i++ {
			q.idxs[i] += ctx.idx
		}
	} else {
		q.idx += ctx.idx
	}
	return q
}

// GetMulti searches for multiple paths within a JSON structure and returns a slice of results.
//
// This function allows you to search for multiple paths in the JSON structure, each represented as a string.
// It returns a slice of `Context` instances for each of the specified paths. The paths can be used to navigate
// nested arrays or objects.
//
// The `path` parameters specify the JSON paths to search for, and the function will attempt to retrieve the values
// associated with those paths. Each result is returned as a `Context` containing information about the matched
// JSON value, including its type, string representation, numeric value, and index in the original JSON.
//
// Parameters:
//   - path: One or more strings representing the paths in the JSON structure to search for. The paths may include
//     array indices and object keys separated by dots or brackets (e.g., "user.name", "items[0].price").
//
// Returns:
//   - []Context: A slice of `Context` instances, each containing the result of searching for one of the paths.
//     Each `Context` may represent various types of JSON values (e.g., string, number, object, array). If no match
//     is found for a path, the corresponding `Context` will be empty.
//
// Example Usage:
//
//	ctx := Context{kind: JSON, raw: "{\"user\": {\"name\": \"John\"}, \"items\": [1, 2, 3]}"}
//	results := ctx.GetMulti("user.name", "items[1]")
//	// results[0].str will contain "John" for the "user.name" path,
//	// results[1].num will contain 2 for the "items[1]" path.
//
// Notes:
//   - This function uses the `GetMulti` function (presumably another function) to process the `raw` JSON string
//     and search for each of the specified paths.
//   - Each result is returned as a separate `Context` for each path, allowing for multiple values to be retrieved
//     at once from the JSON structure.
func (ctx Context) GetMulti(path ...string) []Context {
	return GetMulti(ctx.raw, path...)
}

/////////////////////////////////////////////////
///  Path reconstruction / origin  //////////////
/////////////////////////////////////////////////

// Path returns the original fj path for a Result where the Result came
// from a simple query path that returns a single value. For example, if the
// `Get` function was called with a query path like:
//
//	fj.Get(json, "employees.#(first=Admin)")
//
// This function will return the original path that corresponds to the single
// value in the result, formatted as a JSON path.
//
// The returned value will be in the form of a JSON string:
//
//	"employees.0"
//
// The param 'json' must be the original JSON used when calling Get.
//
// Returns:
//   - A string representing the original path for the single value in the result.
//   - If the paths cannot be determined (e.g., due to the result being from
//     a multi-path, transformer, or a nested query), an empty string will be returned.
//
// Notes:
//   - The `Path` function operates by tracing the position of the result within
//     the original JSON string and reconstructing the query path based on this position.
//   - The function checks the surrounding JSON context (such as whether the result
//     is within an array or object) and extracts the relevant path information.
//   - The path components are identified by traversing the string from the result's index
//     and extracting the array or object keys that lead to the specific value.
//
// Example Usage:
//
//	json := `{
//	  "employees": [
//	    {"id": 1, "name": {"first": "John", "last": "Doe"}, "department": "HR"},
//	    {"id": 2, "name": {"first": "Jane", "last": "Smith"}, "department": "Engineering"},
//	    {"id": 3, "name": {"first": "Admin", "last": "Land"}, "department": "Marketing"},
//	    {"id": 4, "name": {"first": "Emily", "last": "Jones"}, "department": "Engineering"}
//	  ],
//	  "companies": [
//	    {"name": "TechCorp", "employees": [1, 2]},
//	    {"name": "BizGroup", "employees": [3, 4]}
//	  ]
//	}`
//
//	// Get the employee's last name who works in the Engineering department
//	result := fj.Get(json, "employees.#(department=Engineering).name.last")
//	path := result.Path(json)
//
//	// Output: "employees.1.name.last"
//
//	// Explanation:
//	// The `Path` function returns the path to the "last" name of the second
//	// employee in the "employees" array who works in the "Engineering" department.
//	// The path "employees.1.name.last" corresponds to the "Jane Smith" employee,
//	// and the query specifically looks at the "last" name of that employee.
func (ctx Context) Path(json string) string {
	var path []byte
	var components []string
	i := ctx.idx - 1
	// Ensure the index is within bounds of the original JSON
	if ctx.idx+len(ctx.raw) > len(json) {
		// JSON cannot safely contain the Result.
		goto fail
	}
	// Ensure that the unprocessed part matches the expected JSON structure
	if !strings.HasPrefix(json[ctx.idx:], ctx.raw) {
		// Result is not at the expected index in the JSON.
		goto fail
	}
	// Traverse the JSON from the result's index to extract the path
	for ; i >= 0; i-- {
		if json[i] <= ' ' {
			continue
		}
		if json[i] == ':' {
			for ; i >= 0; i-- {
				if json[i] != '"' {
					continue
				}
				break
			}
			raw := extractOutermostValue(json[:i+1])
			i = i - len(raw)
			components = append(components, raw)
			// Key obtained, now process the next component
			raw = extractOutermostValue(json[:i+1])
			i = i - len(raw)
			i++ // Move index for next loop step
		} else if json[i] == '{' {
			// Encountered an open object, this is likely not a valid result
			goto fail
		} else if json[i] == ',' || json[i] == '[' {
			// Inside an array, count the position of the element
			var arrayIdx int
			if json[i] == ',' {
				arrayIdx++
				i--
			}
			for ; i >= 0; i-- {
				if json[i] == ':' {
					// Unexpected colon indicates an object key
					goto fail
				} else if json[i] == ',' {
					arrayIdx++
				} else if json[i] == '[' {
					components = append(components, strconv.Itoa(arrayIdx))
					break
				} else if json[i] == ']' || json[i] == '}' || json[i] == '"' {
					raw := extractOutermostValue(json[:i+1])
					i = i - len(raw) + 1
				}
			}
		}
	}
	// If no components are found, return a default path for "this"
	if len(components) == 0 {
		if DisableTransformers {
			goto fail
		}
		return "@this"
	}
	// Build the final path by appending each component
	for i := len(components) - 1; i >= 0; i-- {
		rawComplexity := Parse(components[i])
		if !rawComplexity.Exists() {
			goto fail
		}
		comp := escapeUnsafeChars(rawComplexity.String())
		path = append(path, '.')
		path = append(path, comp...)
	}
	// Remove the leading dot and return the final path
	if len(path) > 0 {
		path = path[1:]
	}
	return string(path)
fail:
	// Return an empty string if the path could not be determined
	return ""
}

// Paths returns the original fj paths for a Result where the Result came
// from a simple query path that returns an array. For example, if the
// `Get` function was called with a query path like:
//
//	fj.Get(json, "friends.#.first")
//
// This function will return the paths for each element in the resulting array,
// formatted as a JSON array. The returned paths are the original query paths
// for each item in the array, reflecting the specific positions of the elements
// in the original JSON structure.
//
// The returned value will be in the form of a JSON array, such as:
//
//	["friends.0.first", "friends.1.first", "friends.2.first"]
//
// Parameters:
//   - `json`: A string representing the original JSON used in the query.
//     This is required for resolving the specific paths corresponding to
//     each element in the resulting array.
//
// Returns:
//   - A slice of strings (`[]string`), each containing the original path for
//     an element in the result array.
//   - If the result was a simple query that returns an array, each string
//     will be a path to an individual element in the array.
//   - If the paths cannot be determined (e.g., due to the result being
//     from a multi-path, transformer, or a nested query), an empty slice will
//     be returned.
//
// Notes:
//   - The `Paths` function relies on the `idxs` field in the `Context`
//     object. If the `idxs` field is `nil`, the function will return `nil`.
//   - The function iterates over each element in the result (which is expected
//     to be an array) and appends the corresponding path to the `paths` slice.
//   - If the paths cannot be determined (e.g., due to the result coming from
//     a multi-path or more complex query), an empty slice will be returned.
//   - This function is useful for extracting the specific query paths for
//     elements within a larger result array, providing a way to inspect or
//     manipulate the paths of individual items.
//
// Example Usage:
//
//	json := `{
//	  "friends": [
//	    {"first": "Tom", "last": "Smith"},
//	    {"first": "Sophia", "last": "Davis"},
//	    {"first": "James", "last": "Miller"}
//	  ]
//	}`
//
//	result := fj.Get(json, "friends.#.first")
//	paths := result.Paths(json)
//
//	// Output: ["friends.0.first", "friends.1.first", "friends.2.first"]
func (ctx Context) Paths(json string) []string {
	if ctx.idxs == nil {
		return nil
	}
	paths := make([]string, 0, len(ctx.idxs))
	ctx.Foreach(func(_, value Context) bool {
		paths = append(paths, value.Path(json))
		return true
	})
	if len(paths) != len(ctx.idxs) {
		return nil
	}
	return paths
}

//////////////////////////////////
///        Comparison          ///
//////////////////////////////////

// Less compares two Context values (tokens) and returns true if the first token is considered less than the second one.
// It performs comparisons based on the type of the tokens and their respective values.
// The comparison order follows: Null < False < Number < String < True < JSON.
// This function also supports case-insensitive comparisons for String type tokens based on the caseSensitive parameter.
//
// Parameters:
//   - token: The other Context token to compare with the current one (t).
//   - caseSensitive: A boolean flag that indicates whether the comparison for String type tokens should be case-sensitive.
//   - If true, the comparison is case-sensitive (i.e., "a" < "b" but "A" < "b").
//   - If false, the comparison is case-insensitive (i.e., "a" == "A").
//
// Returns:
//   - true: If the current token (t) is considered less than the provided token.
//   - false: If the current token (t) is not considered less than the provided token.
//
// The function first compares the `kind` of both tokens, which represents their JSON types.
// If both tokens have the same kind, it proceeds to compare based on their specific types:
// - For String types, it compares the strings based on the case-sensitive flag.
// - For Number types, it compares the numeric values directly.
// - For other types, it compares the unprocessed JSON values as raw strings (this could be useful for types like Null, Boolean, etc.).
//
// Example usage:
//
//	context1 := Context{kind: String, strings: "apple"}
//	context2 := Context{kind: String, strings: "banana"}
//	result := context1.Less(context2, true) // This would return true because "apple" < "banana" and case-sensitive comparison is used.
func (ctx Context) Less(token Context, caseSensitive bool) bool {
	if ctx.kind < token.kind {
		return true
	}
	if ctx.kind > token.kind {
		return false
	}
	if ctx.kind == String {
		if caseSensitive {
			return ctx.str < token.str
		}
		return lessFold(ctx.str, token.str)
	}
	if ctx.kind == Number {
		return ctx.num < token.num
	}
	return ctx.raw < token.raw
}

//////////////////////////////////
///      Internal helpers      ///
//////////////////////////////////

// parseJSONElements processes a JSON string (from the `Context`) and attempts to parse it as either a JSON array or a JSON object.
//
// The function examines the raw JSON string and determines whether it represents an array or an object by looking at
// the first character ('[' for arrays, '{' for objects). It then processes the content accordingly and returns the
// parsed results as a `qCtx`, which contains either an array or an object, depending on the type of the JSON structure.
//
// Parameters:
//   - vc: A byte representing the expected JSON structure type to parse ('[' for arrays, '{' for objects).
//   - valueSize: A boolean flag that indicates whether intermediary values should be stored as raw types (`true`)
//     or parsed into `Context` objects (`false`).
//
// Returns:
//   - qCtx: A `qCtx` struct containing the parsed elements. This can include:
//   - ArrayResult: A slice of `Context` elements for arrays.
//   - ArrayIns: A slice of `interface{}` elements for arrays when `valueSize` is true.
//   - OpMap: A map of string keys to `Context` values for objects when `valueSize` is false.
//   - OpIns: A map of string keys to `interface{}` values for objects when `valueSize` is true.
//   - valueN: The byte value indicating the start of the JSON array or object ('[' or '{').
//
// Function Process:
//
//  1. **Identifying JSON Structure**:
//     The function starts by checking the first non-whitespace character in the JSON string to determine if it's an object (`{`)
//     or an array (`[`). If the expected structure is detected, the function proceeds accordingly.
//
//  2. **Creating Appropriate Containers**:
//     Based on the type of JSON being parsed (array or object), the function initializes an empty slice or map
//     to store the parsed elements. The `OpMap` or `OpIns` is used for objects, while the `ArrayResult` or `ArrayIns`
//     is used for arrays. If `valueSize` is `true`, the values will be stored as raw types (`interface{}`), otherwise,
//     they will be stored as `Context` objects.
//
//  3. **Parsing JSON Elements**:
//     The function then loops through the JSON string, identifying and parsing individual elements. Each element could
//     be a string, number, boolean, `null`, array, or object. For each identified element, it is added to the appropriate
//     container (array or map) as determined by the type of JSON being processed.
//
//  4. **Handling Key-Value Pairs (for Objects)**:
//     If parsing an object (denoted by `{`), the function identifies key-value pairs and alternates between storing the
//     key (as a string) and its corresponding value (as a `Context` object or raw type) in the `OpMap` or `OpIns` container.
//
//  5. **Assigning Indices**:
//     After parsing the elements, the function assigns the correct index to each element in the `ArrayResult` based on
//     the `idxs` from the parent `Context`. If the number of elements in the array does not match the expected
//     number of indexes, the indices are reset to 0 for each element.
//
// Example Usage:
//
//	ctx := Context{kind: JSON, raw: "[1, 2, 3]"}
//	result := ctx.parseJSONElements('[', false)
//	// result.ArrayResult contains the parsed `Context` elements for the array.
//
//	ctx = Context{kind: JSON, raw: "{\"key\": \"value\"}"}
//	result = ctx.parseJSONElements('{', false)
//	// result.OpMap contains the parsed key-value pair for the object.
//
// Notes:
//   - The function handles various JSON value types, including numbers, strings, booleans, null, and nested arrays/objects.
//   - The function uses internal helper functions like `getNumeric`, `squash`, `lowerPrefix`, and `unescapeJSONEncoded`
//     to parse the raw JSON string into appropriate `Context` elements.
//   - The `valueSize` flag controls whether the elements are stored as raw types (`interface{}`) or as `Context` objects.
//   - If `valueSize` is `false`, the result will contain structured `Context` elements, which can be used for further processing or queries.
func (ctx Context) parseJSONElements(vc byte, valueSize bool) (result qCtx) {
	var json = ctx.raw
	var i int
	var value Context
	var count int
	var key Context
	if vc == 0 {
		for ; i < len(json); i++ {
			if json[i] == '{' || json[i] == '[' {
				result.vn = json[i]
				i++
				break
			}
			if json[i] > ' ' {
				goto end
			}
		}
	} else {
		for ; i < len(json); i++ {
			if json[i] == vc {
				i++
				break
			}
			if json[i] > ' ' {
				goto end
			}
		}
		result.vn = vc
	}
	if result.vn == '{' {
		if valueSize {
			result.opResults = make(map[string]any)
		} else {
			result.ops = make(map[string]Context)
		}
	} else {
		if valueSize {
			result.elems = make([]any, 0)
		} else {
			result.arr = make([]Context, 0)
		}
	}
	for ; i < len(json); i++ {
		if json[i] <= ' ' {
			continue
		}
		if json[i] == ']' || json[i] == '}' {
			break
		}
		switch json[i] {
		default:
			if (json[i] >= '0' && json[i] <= '9') || json[i] == '-' {
				value.kind = Number
				value.raw, value.num = tokenizeNumber(json[i:])
				value.str = ""
			} else {
				continue
			}
		case '{', '[':
			value.kind = JSON
			value.raw = compactJSON(json[i:])
			value.str, value.num = "", 0
		case 'n':
			value.kind = Null
			value.raw = leadingLowercase(json[i:])
			value.str, value.num = "", 0
		case 't':
			value.kind = True
			value.raw = leadingLowercase(json[i:])
			value.str, value.num = "", 0
		case 'f':
			value.kind = False
			value.raw = leadingLowercase(json[i:])
			value.str, value.num = "", 0
		case '"':
			value.kind = String
			value.raw, value.str = extractAndUnescapeJSONString(json[i:])
			value.num = 0
		}
		value.idx = i + ctx.idx

		i += len(value.raw) - 1

		if result.vn == '{' {
			if count%2 == 0 {
				key = value
			} else {
				if valueSize {
					if _, ok := result.opResults[key.str]; !ok {
						result.opResults[key.str] = value.Value()
					}
				} else {
					if _, ok := result.ops[key.str]; !ok {
						result.ops[key.str] = value
					}
				}
			}
			count++
		} else {
			if valueSize {
				result.elems = append(result.elems, value.Value())
			} else {
				result.arr = append(result.arr, value)
			}
		}
	}
end:
	if ctx.idxs != nil {
		if len(ctx.idxs) != len(result.arr) {
			for i := 0; i < len(result.arr); i++ {
				result.arr[i].idx = 0
			}
		} else {
			for i := 0; i < len(result.arr); i++ {
				result.arr[i].idx = ctx.idxs[i]
			}
		}
	}
	return
}
