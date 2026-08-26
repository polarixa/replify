package conv

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

// ///////////////////////////
// Section: String conversion interface
// ///////////////////////////

// stringConverter is an interface for types that can convert themselves to string.
type stringConverter interface {
	String() string
}

// stringErrorConverter is an interface for types that can convert themselves to string with error.
type stringErrorConverter interface {
	String() (string, error)
}

// ///////////////////////////
// Section:  String conversion
// ///////////////////////////

// String returns the string representation from the given interface{} value.
// This function cannot fail for most types - it will use fmt.Sprintf as fallback.
//
// Parameters:
//   - from:  The value to convert.
//
// Returns:
//   - The converted string value.
//   - An error (currently always nil for compatibility, but check it for future-proofing).
func (c *Converter) String(from any) (string, error) {
	if from == nil {
		if c.nilAsZero {
			return "", nil
		}
		return "", newConvError(from, "string")
	}

	// Fast path for common types
	switch v := from.(type) {
	case string:
		return castString(&v), nil
	case *string:
		return castString(v), nil
	case []byte:
		return castBytes(&v), nil
	case *[]byte:
		return castBytes(v), nil
	case []rune:
		return castRunes(&v), nil
	case *[]rune:
		return castRunes(v), nil
	case bool:
		return castBool(&v), nil
	case *bool:
		return castBool(v), nil
	case int:
		return castInt(&v), nil
	case *int:
		return castInt(v), nil
	case int8:
		return castInt8(&v), nil
	case *int8:
		return castInt8(v), nil
	case int16:
		return castInt16(&v), nil
	case *int16:
		return castInt16(v), nil
	case int32:
		return castInt32(&v), nil
	case *int32:
		return castInt32(v), nil
	case int64:
		return castInt64(&v), nil
	case *int64:
		return castInt64(v), nil
	case uint:
		return castUint(&v), nil
	case *uint:
		return castUint(v), nil
	case uint8:
		return castUint8(&v), nil
	case *uint8:
		return castUint8(v), nil
	case uint16:
		return castUint16(&v), nil
	case *uint16:
		return castUint16(v), nil
	case uint32:
		return castUint32(&v), nil
	case *uint32:
		return castUint32(v), nil
	case uint64:
		return castUint64(&v), nil
	case *uint64:
		return castUint64(v), nil
	case float32:
		return castFloat32(&v), nil
	case *float32:
		return castFloat32(v), nil
	case float64:
		return castFloat64(&v), nil
	case *float64:
		return castFloat64(v), nil
	case complex64:
		return castComplex64(&v)
	case *complex64:
		return castComplex64(v)
	case complex128:
		return castComplex128(&v)
	case *complex128:
		return castComplex128(v)
	case time.Time:
		return castTime(&v), nil
	case *time.Time:
		return castTime(v), nil
	case time.Duration:
		return castDuration(&v), nil
	case *time.Duration:
		return castDuration(v), nil
	case fmt.Stringer:
		return castFmtStringer(&v), nil
	case *fmt.Stringer:
		return castFmtStringer(v), nil
	case json.RawMessage:
		return castJSONRawMessage(&v)
	case *json.RawMessage:
		return castJSONRawMessage(v)
	case json.Marshaler:
		return castJSONMarshaler(&v)
	case *json.Marshaler:
		return castJSONMarshaler(v)
	case error:
		return castError(&v), nil
	case *error:
		return castError(v), nil
	}

	// Check for custom converter interfaces
	if conv, ok := from.(stringConverter); ok {
		return conv.String(), nil
	}
	if conv, ok := from.(stringErrorConverter); ok {
		return conv.String()
	}

	// Use reflection for other types
	return c.stringFromReflect(from)
}

// stringFromReflect converts a reflect.Value to a string based on its kind.
//
// Parameters:
//   - `from`: The input value to be converted to string.
//
// Returns:
//   - A string representation of the input value.
//   - An error if the conversion fails.
func (c *Converter) stringFromReflect(from any) (string, error) {
	value := indirectValue(reflect.ValueOf(from))
	if !value.IsValid() {
		if c.nilAsZero {
			return "", nil
		}
		return "", newConvError(from, "string")
	}

	kind := value.Kind()
	switch kind {
	case reflect.String:
		s := value.String()
		return castString(&s), nil
	case reflect.Bool:
		s := value.Bool()
		return castBool(&s), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		s := value.Int()
		return castInt64(&s), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		s := value.Uint()
		return castUint64(&s), nil
	case reflect.Float32:
		s := value.Float()
		f := float32(s)
		return castFloat32(&f), nil
	case reflect.Float64:
		s := value.Float()
		return castFloat64(&s), nil
	case reflect.Complex64:
		return castComplex64Ref(value)
	case reflect.Complex128:
		return castComplex128Ref(value)
	case reflect.Slice:
		if value.Type().Elem().Kind() == reflect.Uint8 {
			// s := value.Bytes()
			// return castBytes(&s), nil
			return castBytes((*[]byte)(unsafe.Pointer(value.Pointer()))), nil
		}
	case reflect.Struct:
		if value.Type() == reflect.TypeFor[time.Time]() {
			t := value.Interface().(time.Time)
			return castTime(&t), nil
		}
		if value.Type() == reflect.TypeFor[time.Duration]() {
			d := value.Interface().(time.Duration)
			return castDuration(&d), nil
		}
		if conv, ok := value.Interface().(stringConverter); ok {
			return conv.String(), nil
		}
		if conv, ok := value.Interface().(stringErrorConverter); ok {
			return conv.String()
		}
	case reflect.Ptr:
		if value.IsNil() {
			if c.nilAsZero {
				return "", nil
			}
			return "", newConvError(from, "string")
		}
		elem := value.Elem()
		return c.stringFromReflect(elem.Interface())
	case reflect.Interface:
		if value.IsNil() {
			if c.nilAsZero {
				return "", nil
			}
			return "", newConvError(from, "string")
		}
		elem := value.Elem()
		return c.stringFromReflect(elem.Interface())
	}

	// Fallback to fmt.Sprintf
	return fmt.Sprintf("%v", from), nil
}

// ///////////////////////////
// Section: String formatting
// ///////////////////////////

// StringSlice converts a slice of any type to a slice of strings.
//
// Parameters:
//   - from: The input value to be converted to a slice of strings.
//
// Returns:
//   - A slice of strings representing the converted values.
//   - An error if any conversion fails.
func StringSlice(from any) ([]string, error) {
	if from == nil {
		return nil, nil
	}

	// Fast path for string slice
	if v, ok := from.([]string); ok {
		return v, nil
	}

	value := reflect.ValueOf(from)
	if value.Kind() != reflect.Slice && value.Kind() != reflect.Array {
		// Single value - wrap in slice
		s, err := defaultConverter.String(from)
		if err != nil {
			return nil, err
		}
		return []string{s}, nil
	}

	result := make([]string, value.Len())
	for i := 0; i < value.Len(); i++ {
		s, err := defaultConverter.String(value.Index(i).Interface())
		if err != nil {
			return nil, err
		}
		result[i] = s
	}

	return result, nil
}

// ///////////////////////////
// Section: String utility functions
// ///////////////////////////

// StringOrEmpty returns the converted string or empty string if conversion fails.
//
// Parameters:
//   - from:  The value to convert.
//
// Returns:
//   - The converted string value, or empty string if conversion fails.
func StringOrEmpty(from any) string {
	v, _ := defaultConverter.String(from)
	return v
}

// Quote returns a double-quoted string safely escaped with Go syntax.
//
// Parameters:
//   - from:  The value to convert.
//
// Returns:
//   - The quoted string value.
func Quote(from any) string {
	s, _ := defaultConverter.String(from)
	return strconv.Quote(s)
}

// TrimSpace returns the string with all leading and trailing white space removed.
//
// Parameters:
//   - from:  The value to convert.
//
// Returns:
//   - The trimmed string value.
func TrimSpace(from any) string {
	s, _ := defaultConverter.String(from)
	return strings.TrimSpace(s)
}

// Join converts a slice of any type to a single string joined by the specified separator.
//
// Parameters:
//   - from: The input value to be converted to a joined string.
//   - sep:  The separator string to use between elements.
//
// Returns:
//   - A single string with all elements joined by the separator.
//   - An error if any conversion fails.
func Join(from any, sep string) (string, error) {
	slice, err := StringSlice(from)
	if err != nil {
		return "", err
	}
	return strings.Join(slice, sep), nil
}

// realFrom extracts the real part of a complex number from a reflect.Value.
//
// Parameters:
//   - `v`: The reflect.Value to extract the real part from.
//
// Returns:
//   - The real part of the complex number as a float64.
//
// Example:
//
//	realPart := realFrom(reflect.ValueOf(complex(1.23, 4.56)))
func realFrom(v reflect.Value) float64 {
	switch v.Kind() {
	case reflect.Complex64:
		return float64(real(v.Complex()))
	case reflect.Complex128:
		return real(v.Complex())
	default:
		return 0
	}
}

// imagFrom extracts the imaginary part of a complex number from a reflect.Value.
//
// Parameters:
//   - `v`: The reflect.Value to extract the imaginary part from.
//
// Returns:
//   - The imaginary part of the complex number as a float64.
//
// Example:
//
//	imagPart := imagFrom(reflect.ValueOf(complex(1.23, 4.56)))
func imagFrom(v reflect.Value) float64 {
	switch v.Kind() {
	case reflect.Complex64:
		return float64(imag(v.Complex()))
	case reflect.Complex128:
		return imag(v.Complex())
	default:
		return 0
	}
}

// formatFloatJSON formats a float64 to its JSON string representation.
// It uses 'g' formatting like encoding/json and converts non-finite numbers to null.
//
// Parameters:
//   - `f`: The float64 value to format.
//   - `is32`: A boolean indicating whether the float is a float32 (true) or float64 (false).
//
// Returns:
//   - A string containing the JSON representation of the float.
//
// Example:
//
//	jsonFloat := formatFloatJSON(1.2345, false)
func formatFloatJSON(f float64, is32 bool) string {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		if true { // floatsUseNullForNonFinite, always true in this implementation
			return "null"
		}
		return ""
	}
	bitSize := 64
	if is32 {
		bitSize = 32
	}
	// 'g' format is used for general-purpose formatting. It uses the shortest
	// representation of the float, either in decimal or scientific notation.
	// The -1 precision means that the smallest number of digits necessary to
	// represent the float will be used.
	return strconv.FormatFloat(f, 'g', -1, bitSize)
}

// encodeComplexJSONToken encodes a complex number to its JSON string representation.
// It uses 'g' formatting like encoding/json and converts non-finite numbers to null.
//
// Parameters:
//   - `realPart`: The real part of the complex number.
//   - `imagPart`: The imaginary part of the complex number.
//   - `is32`: A boolean indicating whether the complex number is a complex64 (true) or complex128 (false).
//
// Returns:
//   - A string containing the JSON representation of the complex number.
//   - An error if the marshalling fails.
//
// Example:
//
//	r := encodeComplexJSONToken(1.2345, 6.789, false)
func encodeComplexJSONToken(realPart, imagPart float64, is32 bool) (string, error) {
	r := formatFloatJSON(realPart, is32)
	i := formatFloatJSON(imagPart, is32)
	// formatFloatJSON returns "" only when floatsUseNullForNonFinite is false and the
	// component is non-finite (NaN/Inf). In that case the token policy is to error.
	if r == "" || i == "" {
		return "", errors.New("non-finite float (NaN/Inf)")
	}
	return `{"real":` + r + `,"imag":` + i + `}`, nil
}

// castString safely dereferences a string pointer, returning an empty string if the pointer is nil.
//
// Parameters:
//   - value: A pointer to a string.
//
// Returns:
//   - The dereferenced string value if the pointer is not nil; otherwise, an empty string.
//
// Example:
//
//	var strPtr *string
//	result := castString(strPtr) // result will be ""
func castString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

// castBytes safely dereferences a byte slice pointer, returning an empty string if the pointer is nil or the slice is empty.
//
// Parameters:
//   - value: A pointer to a byte slice.
//
// Returns:
//   - The string representation of the byte slice if the pointer is not nil and the slice is not empty; otherwise, an empty string.
//
// Example:
//
//	var bytesPtr *[]byte
//	result := castBytes(bytesPtr) // result will be ""
func castBytes(value *[]byte) string {
	if value == nil || len(*value) == 0 {
		return ""
	}
	return string(*value)
}

// castRunes safely dereferences a rune slice pointer, returning an empty string if the pointer is nil or the slice is empty.
//
// Parameters:
//   - value: A pointer to a rune slice.
//
// Returns:
//   - The string representation of the rune slice if the pointer is not nil and the slice is not empty; otherwise, an empty string.
//
// Example:
//
//	var runesPtr *[]rune
//	result := castRunes(runesPtr) // result will be ""
func castRunes(value *[]rune) string {
	if value == nil || len(*value) == 0 {
		return ""
	}
	return string(*value)
}

// castBool safely dereferences a bool pointer, returning an empty string if the pointer is nil.
//
// Parameters:
//   - value: A pointer to a bool.
//
// Returns:
//   - The string representation of the bool value if the pointer is not nil; otherwise, an empty string.
//
// Example:
//
//	var boolPtr *bool
//	result := castBool(boolPtr) // result will be ""
func castBool(value *bool) string {
	if value == nil {
		return ""
	}
	return strconv.FormatBool(*value)
}

// castInt safely dereferences an int pointer, returning an empty string if the pointer is nil.
//
// Parameters:
//   - value: A pointer to an int.
//
// Returns:
//   - The string representation of the int value if the pointer is not nil; otherwise, an empty string.
//
// Example:
//
//	var intPtr *int
//	result := castInt(intPtr) // result will be ""
func castInt(value *int) string {
	if value == nil {
		return ""
	}
	return strconv.Itoa(*value)
}

// castInt8 safely dereferences an int8 pointer, returning an empty string if the pointer is nil.
//
// Parameters:
//   - value: A pointer to an int8.
//
// Returns:
//   - The string representation of the int8 value if the pointer is not nil; otherwise, an empty string.
//
// Example:
//
//	var int8Ptr *int8
//	result := castInt8(int8Ptr) // result will be ""
func castInt8(value *int8) string {
	if value == nil {
		return ""
	}
	return strconv.FormatInt(int64(*value), 10)
}

// castInt16 safely dereferences an int16 pointer, returning an empty string if the pointer is nil.
//
// Parameters:
//   - value: A pointer to an int16.
//
// Returns:
//   - The string representation of the int16 value if the pointer is not nil; otherwise, an empty string.
//
// Example:
//
//	var int16Ptr *int16
//	result := castInt16(int16Ptr) // result will be ""
func castInt16(value *int16) string {
	if value == nil {
		return ""
	}
	return strconv.FormatInt(int64(*value), 10)
}

// castInt32 safely dereferences an int32 pointer, returning an empty string if the pointer is nil.
//
// Parameters:
//   - value: A pointer to an int32.
//
// Returns:
//   - The string representation of the int32 value if the pointer is not nil; otherwise, an empty string.
//
// Example:
//
//	var int32Ptr *int32
//	result := castInt32(int32Ptr) // result will be ""
func castInt32(value *int32) string {
	if value == nil {
		return ""
	}
	return strconv.FormatInt(int64(*value), 10)
}

// castInt64 safely dereferences an int64 pointer, returning an empty string if the pointer is nil.
//
// Parameters:
//   - value: A pointer to an int64.
//
// Returns:
//   - The string representation of the int64 value if the pointer is not nil; otherwise, an empty string.
//
// Example:
//
//	var int64Ptr *int64
//	result := castInt64(int64Ptr) // result will be ""
func castInt64(value *int64) string {
	if value == nil {
		return ""
	}
	return strconv.FormatInt(*value, 10)
}

// castUint safely dereferences a uint pointer, returning an empty string if the pointer is nil.
//
// Parameters:
//   - value: A pointer to a uint.
//
// Returns:
//   - The string representation of the uint value if the pointer is not nil; otherwise, an empty string.
//
// Example:
//
//	var uintPtr *uint
//	result := castUint(uintPtr) // result will be ""
func castUint(value *uint) string {
	if value == nil {
		return ""
	}
	return strconv.FormatUint(uint64(*value), 10)
}

// castUint8 safely dereferences a uint8 pointer, returning an empty string if the pointer is nil.
//
// Parameters:
//   - value: A pointer to a uint8.
//
// Returns:
//   - The string representation of the uint8 value if the pointer is not nil; otherwise, an empty string.
//
// Example:
//
//	var uint8Ptr *uint8
//	result := castUint8(uint8Ptr) // result will be ""
func castUint8(value *uint8) string {
	if value == nil {
		return ""
	}
	return strconv.FormatUint(uint64(*value), 10)
}

// castUint16 safely dereferences a uint16 pointer, returning an empty string if the pointer is nil.
//
// Parameters:
//   - value: A pointer to a uint16.
//
// Returns:
//   - The string representation of the uint16 value if the pointer is not nil; otherwise, an empty string.
//
// Example:
//
//	var uint16Ptr *uint16
//	result := castUint16(uint16Ptr) // result will be ""
func castUint16(value *uint16) string {
	if value == nil {
		return ""
	}
	return strconv.FormatUint(uint64(*value), 10)
}

// castUint32 safely dereferences a uint32 pointer, returning an empty string if the pointer is nil.
//
// Parameters:
//   - value: A pointer to a uint32.
//
// Returns:
//   - The string representation of the uint32 value if the pointer is not nil; otherwise, an empty string.
//
// Example:
//
//	var uint32Ptr *uint32
//	result := castUint32(uint32Ptr) // result will be ""
func castUint32(value *uint32) string {
	if value == nil {
		return ""
	}
	return strconv.FormatUint(uint64(*value), 10)
}

// castUint64 safely dereferences a uint64 pointer, returning an empty string if the pointer is nil.
//
// Parameters:
//   - value: A pointer to a uint64.
//
// Returns:
//   - The string representation of the uint64 value if the pointer is not nil; otherwise, an empty string.
//
// Example:
//
//	var uint64Ptr *uint64
//	result := castUint64(uint64Ptr) // result will be ""
func castUint64(value *uint64) string {
	if value == nil {
		return ""
	}
	return strconv.FormatUint(*value, 10)
}

// castFloat32 safely dereferences a float32 pointer, returning an empty string if the pointer is nil.
//
// Parameters:
//   - value: A pointer to a float32.
//
// Returns:
//   - The string representation of the float32 value if the pointer is not nil; otherwise, an empty string.
//
// Example:
//
//	var float32Ptr *float32
//	result := castFloat32(float32Ptr) // result will be ""
func castFloat32(value *float32) string {
	if value == nil {
		return ""
	}
	return strconv.FormatFloat(float64(*value), 'f', -1, 32)
}

// castFloat64 safely dereferences a float64 pointer, returning an empty string if the pointer is nil.
//
// Parameters:
//   - value: A pointer to a float64.
//
// Returns:
//   - The string representation of the float64 value if the pointer is not nil; otherwise, an empty string.
//
// Example:
//
//	var float64Ptr *float64
//	result := castFloat64(float64Ptr) // result will be ""
func castFloat64(value *float64) string {
	if value == nil {
		return ""
	}
	return strconv.FormatFloat(*value, 'f', -1, 64)
}

// castComplex64 safely dereferences a complex64 pointer, returning an empty string if the pointer is nil.
//
// Parameters:
//   - value: A pointer to a complex64.
//
// Returns:
//   - The string representation of the complex64 value if the pointer is not nil; otherwise, an empty string.
//
// Example:
//
//	var complex64Ptr *complex64
//	result := castComplex64(complex64Ptr) // result will be ""
func castComplex64(value *complex64) (string, error) {
	if value == nil {
		return "", nil
	}
	return encodeComplexJSONToken(float64(real(*value)), float64(imag(*value)), true)
}

// castComplex64Ref safely dereferences a reflect.Value containing a complex64, returning an empty string if the value is nil or invalid.
//
// Parameters:
//   - value: A reflect.Value containing a complex64.
//
// Returns:
//   - The string representation of the complex64 value if the value is valid; otherwise, an empty string or an error.
//
// Example:
//
//	var complex64Val complex64 = 1.23 + 4.56i
//	result, err := castComplex64Ref(reflect.ValueOf(complex64Val)) // result will be the JSON representation of the complex number
func castComplex64Ref(value reflect.Value) (string, error) {
	if value.IsNil() {
		return "", nil
	}
	if !value.IsValid() {
		return "", newConvError(value.Interface(), "invalid complex64 value")
	}
	r, i := realFrom(value), imagFrom(value)
	return encodeComplexJSONToken(r, i, true)
}

// castComplex128 safely dereferences a complex128 pointer, returning an empty string if the pointer is nil.
//
// Parameters:
//   - value: A pointer to a complex128.
//
// Returns:
//   - The string representation of the complex128 value if the pointer is not nil; otherwise, an empty string.
//
// Example:
//
//	var complex128Ptr *complex128
//	result := castComplex128(complex128Ptr) // result will be ""
func castComplex128(value *complex128) (string, error) {
	if value == nil {
		return "", nil
	}
	return encodeComplexJSONToken(real(*value), imag(*value), false)
}

// castComplex128Ref safely dereferences a reflect.Value containing a complex128, returning an empty string if the value is nil or invalid.
//
// Parameters:
//   - value: A reflect.Value containing a complex128.
//
// Returns:
//   - The string representation of the complex128 value if the value is valid; otherwise, an empty string or an error.
//
// Example:
//
//	var complex128Val complex128 = 1.23 + 4.56i
//	result, err := castComplex128Ref(reflect.ValueOf(complex128Val)) // result will be the JSON representation of the complex number
func castComplex128Ref(value reflect.Value) (string, error) {
	if value.IsNil() {
		return "", nil
	}
	if !value.IsValid() {
		return "", newConvError(value.Interface(), "invalid complex128 value")
	}
	r, i := realFrom(value), imagFrom(value)
	return encodeComplexJSONToken(r, i, false)
}

// castTime safely dereferences a time.Time pointer, returning an empty string if the pointer is nil.
//
// Parameters:
//   - value: A pointer to a time.Time.
//
// Returns:
//   - The string representation of the time.Time value in RFC3339 format if the pointer is not nil; otherwise, an empty string.
//
// Example:
//
//	var timePtr *time.Time
//	result := castTime(timePtr) // result will be ""
func castTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format(time.RFC3339)
}

// castDuration safely dereferences a time.Duration pointer, returning an empty string if the pointer is nil.
//
// Parameters:
//   - value: A pointer to a time.Duration.
//
// Returns:
//   - The string representation of the time.Duration value if the pointer is not nil; otherwise, an empty string.
//
// Example:
//
//	var durationPtr *time.Duration
//	result := castDuration(durationPtr) // result will be ""
func castDuration(value *time.Duration) string {
	if value == nil {
		return ""
	}
	return value.String()
}

// castError safely dereferences an error pointer, returning an empty string if the pointer is nil.
//
// Parameters:
//   - value: A pointer to an error.
//
// Returns:
//   - The string representation of the error if the pointer is not nil; otherwise, an empty string.
//
// Example:
//
//	var errPtr *error
//	result := castError(errPtr) // result will be ""
func castError(value *error) string {
	if value == nil || *value == nil {
		return ""
	}
	return (*value).Error()
}

// castFmtStringer safely dereferences a fmt.Stringer pointer, returning an empty string if the pointer is nil.
//
// Parameters:
//   - value: A pointer to a fmt.Stringer.
//
// Returns:
//   - The string representation of the fmt.Stringer if the pointer is not nil; otherwise, an empty string.
//
// Example:
//
//	var stringerPtr *fmt.Stringer
//	result := castFmtStringer(stringerPtr) // result will be ""
func castFmtStringer(value *fmt.Stringer) string {
	if value == nil || *value == nil {
		return ""
	}
	return (*value).String()
}

// castJSONRawMessage safely dereferences a json.RawMessage pointer, returning an empty string if the pointer is nil or the JSON is invalid.
//
// Parameters:
//   - value: A pointer to a json.RawMessage.
//
// Returns:
//   - The string representation of the JSON if the pointer is not nil and the JSON is valid; otherwise, an empty string or an error.
//
// Example:
//
//	var rawMsgPtr *json.RawMessage
//	result, err := castJSONRawMessage(rawMsgPtr) // result will be "", err will be nil
func castJSONRawMessage(value *json.RawMessage) (string, error) {
	if value == nil {
		return "", nil
	}
	if !json.Valid(*value) {
		return "", newConvError(value, "invalid JSON")
	}
	return string(*value), nil
}

// castJSONMarshaler safely dereferences a json.Marshaler pointer, returning an empty string if the pointer is nil or if marshalling fails.
//
// Parameters:
//   - value: A pointer to a json.Marshaler.
//
// Returns:
//   - The string representation of the marshalled JSON if the pointer is not nil and marshalling succeeds; otherwise, an empty string or an error.
//
// Example:
//
//	var marshalerPtr *json.Marshaler
//	result, err := castJSONMarshaler(marshalerPtr) // result will be "", err will be nil
func castJSONMarshaler(value *json.Marshaler) (string, error) {
	if value == nil {
		return "", nil
	}
	b, err := (*value).MarshalJSON()
	if err != nil {
		return "", err
	}
	return string(b), nil
}
