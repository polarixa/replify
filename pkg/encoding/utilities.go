package encoding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// _jsonMarshalerType is the reflect.Type of the json.Marshaler interface.
// Used to detect types that implement their own JSON serialisation so that
// toJSONCompatible can leave them untouched.
var _jsonMarshalerType = reflect.TypeFor[json.Marshaler]()

// _complexTypeCache caches typeContainsComplex results keyed by reflect.Type.
// After the first reflection walk for a given type every subsequent call is O(1).
var _complexTypeCache sync.Map

// containsComplex reports whether t or any type structurally reachable from t
// contains a complex64 or complex128 field. Results are cached per type, so
// repeated marshaling of the same struct pays only a single sync.Map lookup.
func containsComplex(t reflect.Type) bool {
	if v, ok := _complexTypeCache.Load(t); ok {
		return v.(bool)
	}
	result := typeContainsComplex(t, make(map[reflect.Type]bool))
	_complexTypeCache.Store(t, result)
	return result
}

// typeContainsComplex is the recursive worker for containsComplex.
// seen prevents infinite loops for self-referential types.
func typeContainsComplex(t reflect.Type, seen map[reflect.Type]bool) bool {
	switch t.Kind() {
	case reflect.Complex64, reflect.Complex128:
		return true
	case reflect.Ptr, reflect.Slice, reflect.Array, reflect.Chan:
		return typeContainsComplex(t.Elem(), seen)
	case reflect.Map:
		return typeContainsComplex(t.Key(), seen) || typeContainsComplex(t.Elem(), seen)
	case reflect.Struct:
		if seen[t] {
			return false // recursive type — avoid infinite loop
		}
		seen[t] = true
		for i := 0; i < t.NumField(); i++ {
			if typeContainsComplex(t.Field(i).Type, seen) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

// parseJSONTag extracts the effective JSON key name, the omitempty flag, and
// whether the field should be omitted entirely ("-") from a struct field's tag.
func parseJSONTag(sf reflect.StructField) (name string, omitempty, skip bool) {
	tag := sf.Tag.Get("json")
	if tag == "" {
		return sf.Name, false, false
	}
	if tag == "-" {
		return "", false, true
	}
	before, after, ok := strings.Cut(tag, ",")
	if !ok {
		return tag, false, false
	}
	name = before
	if name == "" {
		name = sf.Name
	}
	omitempty = strings.Contains(after, "omitempty")
	return name, omitempty, false
}

// encodeValue encodes a reflect.Value to compact JSON bytes, handling
// complex64/complex128 at any depth and preserving struct field declaration
// order. Types that implement json.Marshaler are handed off directly so that
// their custom serialisation (e.g. time.Time → RFC3339) is preserved.
func encodeValue(v reflect.Value) ([]byte, error) {
	// Dereference pointer chain.
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return []byte("null"), nil
		}
		v = v.Elem()
	}
	// Unwrap interface layers.
	for v.Kind() == reflect.Interface {
		if v.IsNil() {
			return []byte("null"), nil
		}
		v = v.Elem()
	}
	if !v.IsValid() {
		return []byte("null"), nil
	}

	t := v.Type()
	// Honour custom JSON marshaling (e.g. time.Time, json.RawMessage…).
	if t.Implements(_jsonMarshalerType) || reflect.PointerTo(t).Implements(_jsonMarshalerType) {
		return json.Marshal(v.Interface())
	}

	switch v.Kind() {
	case reflect.Complex64, reflect.Complex128:
		r, i := realFrom(v), imagFrom(v)
		return json.Marshal(map[string]float64{"real": r, "imag": i})

	case reflect.Struct:
		return encodeStruct(v)

	case reflect.Slice:
		if v.IsNil() {
			return []byte("null"), nil
		}
		return encodeSlice(v)

	case reflect.Array:
		return encodeSlice(v)

	case reflect.Map:
		if v.IsNil() {
			return []byte("null"), nil
		}
		return encodeMap(v)

	default:
		return json.Marshal(v.Interface())
	}
}

// isAnonymousPromoted reports whether sf is an anonymous (embedded) struct field
// whose sub-fields should be promoted into the parent JSON object — i.e. it
// carries no explicit JSON key name. This mirrors encoding/json promotion rules.
func isAnonymousPromoted(sf reflect.StructField) bool {
	if !sf.Anonymous {
		return false
	}
	tag := sf.Tag.Get("json")
	if tag == "" {
		return true
	}
	if tag == "-" {
		return false
	}
	// `json:",omitempty"` — comma is the first character, no explicit name → promote.
	// `json:"name"` or `json:"name,omitempty"` — has explicit name → do not promote.
	return strings.IndexByte(tag, ',') == 0
}

// writeStructFields writes the exported fields of struct v to buf in declaration
// order, honouring json struct tags. Anonymous (embedded) struct fields whose
// tag carries no explicit name are promoted — their sub-fields are written
// inline, matching encoding/json embedding semantics.
func writeStructFields(v reflect.Value, buf *bytes.Buffer, first *bool) error {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if !sf.IsExported() {
			continue
		}
		fv := v.Field(i)

		if isAnonymousPromoted(sf) {
			// Dereference an embedded pointer; skip the whole embed if nil.
			fw := fv
			nilPtr := false
			for fw.Kind() == reflect.Ptr {
				if fw.IsNil() {
					nilPtr = true
					break
				}
				fw = fw.Elem()
			}
			if !nilPtr && fw.Kind() == reflect.Struct {
				if err := writeStructFields(fw, buf, first); err != nil {
					return err
				}
			}
			continue
		}

		name, omitempty, skip := parseJSONTag(sf)
		if skip {
			continue
		}
		if omitempty && fv.IsZero() {
			continue
		}
		if !*first {
			buf.WriteByte(',')
		}
		*first = false
		keyBytes, err := json.Marshal(name)
		if err != nil {
			return err
		}
		buf.Write(keyBytes)
		buf.WriteByte(':')
		valBytes, err := encodeValue(fv)
		if err != nil {
			return err
		}
		buf.Write(valBytes)
	}
	return nil
}

// encodeStruct serialises a struct to compact JSON bytes in field declaration
// order, honouring json struct tags (name, omitempty, "-") and promoting
// anonymous (embedded) struct fields that carry no explicit JSON name.
func encodeStruct(v reflect.Value) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	first := true
	if err := writeStructFields(v, &buf, &first); err != nil {
		return nil, err
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// encodeSlice serialises a slice or array to compact JSON bytes.
func encodeSlice(v reflect.Value) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('[')
	for i := 0; i < v.Len(); i++ {
		if i > 0 {
			buf.WriteByte(',')
		}
		b, err := encodeValue(v.Index(i))
		if err != nil {
			return nil, err
		}
		buf.Write(b)
	}
	buf.WriteByte(']')
	return buf.Bytes(), nil
}

// encodeMap serialises a map to compact JSON bytes.
// It first attempts a direct json.Marshal; on UnsupportedTypeError (complex
// values) it falls back to encoding each value individually.
func encodeMap(v reflect.Value) ([]byte, error) {
	// Fast path: no complex values — let json.Marshal handle it.
	b, err := json.Marshal(v.Interface())
	if err == nil {
		return b, nil
	}
	if _, ok := err.(*json.UnsupportedTypeError); !ok {
		return nil, err
	}
	// Slow path: at least one value is a complex type.
	var buf bytes.Buffer
	buf.WriteByte('{')
	first := true
	for _, key := range v.MapKeys() {
		if !first {
			buf.WriteByte(',')
		}
		first = false
		keyBytes, err := json.Marshal(fmt.Sprintf("%v", key.Interface()))
		if err != nil {
			return nil, err
		}
		buf.Write(keyBytes)
		buf.WriteByte(':')
		valBytes, err := encodeValue(v.MapIndex(key))
		if err != nil {
			return nil, err
		}
		buf.Write(valBytes)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// safeMarshalJSONString marshals a Go value to its JSON string representation or returns an error if the marshalling fails.
// It uses a deferred function to recover from any panics that may occur during marshalling.
//
// Parameters:
//   - `v`: The Go value to be marshalled into JSON.
//   - `pretty`: A boolean indicating whether the JSON should be pretty-printed.
//
// Returns:
//   - A string containing the JSON representation of the input value.
//   - An error if the marshalling fails.
//
// Example:
//
//	jsonString, err := safeMarshalJSONString(myStruct, false)
func safeMarshalJSONString(v any, pretty bool) (as string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%w: %v", ErrMarshalPanicRecovered, r)
			as = ""
		}
	}()

	rv := reflect.ValueOf(v)

	// Static fast-path: if the concrete type is known to contain complex numbers
	// skip the doomed json.Marshal attempt entirely. containsComplex is O(1)
	// after the first call per type thanks to the sync.Map cache.
	if rv.IsValid() && containsComplex(rv.Type()) {
		b, encErr := encodeValue(rv)
		if encErr != nil {
			return "", encErr
		}
		if pretty {
			var indented bytes.Buffer
			if indentErr := json.Indent(&indented, b, "", "    "); indentErr != nil {
				return "", indentErr
			}
			b = indented.Bytes()
		}
		return string(b), nil
	}

	// No statically known complex fields — attempt direct marshal.
	var b []byte
	if pretty {
		b, err = MarshalJSONIndent(v, "", "    ")
	} else {
		b, err = MarshalJSON(v)
	}
	if err == nil {
		return string(b), nil
	}

	// Dynamic fallback: an any-typed field holding a complex value can only be
	// detected at runtime. Retry once with the order-preserving encoder.
	if _, ok := err.(*json.UnsupportedTypeError); ok {
		b, err = encodeValue(rv)
		if err != nil {
			return "", err
		}
		if pretty {
			var indented bytes.Buffer
			if indentErr := json.Indent(&indented, b, "", "    "); indentErr != nil {
				return "", indentErr
			}
			b = indented.Bytes()
		}
		return string(b), nil
	}

	return "", err
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

// formatJSONFloat formats a float64 to its JSON string representation.
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
//	jsonFloat := formatJSONFloat(1.2345, false)
func formatJSONFloat(f float64, is32 bool) string {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		if floatsUseNullForNonFinite {
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

// encodeComplexJSON encodes a complex number to its JSON string representation.
// It uses 'g' formatting like encoding/json and converts non-finite numbers to null.
//
// Parameters:
//   - `realPart`: The real part of the complex number.
//   - `imagPart`: The imaginary part of the complex number.
//   - `is32`: A boolean indicating whether the complex number is a complex64 (true) or complex128 (false).
//
// Returns:
//   - A string containing the JSON representation of the complex number.
//
// Example:
//
//	r := encodeComplexJSON(1.2345, 6.789, false)
func encodeComplexJSON(realPart, imagPart float64, is32 bool) string {
	// Use 'g' formatting like encoding/json; convert non-finite to null.
	r := formatJSONFloat(realPart, is32)
	i := formatJSONFloat(imagPart, is32)
	if r == "" || i == "" {
		return ""
	}
	// Format complex number as JSON object with "real" and "imag" fields.
	// This is done to avoid the default JSON representation of complex numbers,
	// which is not valid JSON.
	return `{"real":` + r + `,"imag":` + i + `}`
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
	r := formatJSONFloat(realPart, is32)
	i := formatJSONFloat(imagPart, is32)
	// formatFloatJSON returns "" only when floatsUseNullForNonFinite is false and the
	// component is non-finite (NaN/Inf). In that case the token policy is to error.
	if r == "" || i == "" {
		return "", ErrNonFiniteFloat
	}
	return `{"real":` + r + `,"imag":` + i + `}`, nil
}

// marshalJSON converts a Go value to its JSON string representation or returns an error if the marshalling fails.
// It uses a deferred function to recover from any panics that may occur during marshalling.
//
// Parameters:
//   - `data`: The Go value to be converted to JSON.
//   - `pretty`: A boolean indicating whether the JSON should be pretty-printed.
//
// Returns:
//   - A string containing the JSON representation of the input value.
//   - An error if the marshalling fails.
//
// Example:
//
//	jsonString, err := marshalJSON(myStruct, false)
func marshalJSON(data any, pretty bool) string {
	if data == nil {
		return ""
	}

	// if data is string, return it
	s, ok := data.(string)
	if ok {
		return s
	}

	// 1) Pass-through raw JSON if explicitly provided.
	if rm, ok := data.(json.RawMessage); ok {
		if rm == nil {
			return "null"
		}
		if !json.Valid(rm) {
			// Invalid raw JSON => treat as error sentinel.
			return ""
		}
		if !pretty {
			return string(rm)
		}
		var buf bytes.Buffer
		// json.Indent is a no-op for primitives; safe to call for any valid JSON value.
		if err := json.Indent(&buf, rm, "", "    "); err != nil {
			return ""
		}
		return buf.String()
	}

	v := reflect.ValueOf(data)

	// 2) Unwrap *all* interface layers; nil interface => "".
	for v.Kind() == reflect.Interface {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}
	data = v.Interface()

	// 3) Nil-able kinds (Ptr, Map, Slice, Func, Chan) when nil => "null".
	v = reflect.ValueOf(data)
	switch v.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Func, reflect.Chan:
		if v.IsNil() {
			return "null"
		}
	}

	// 4) Scalar fast-path as valid JSON tokens.
	switch v.Kind() {
	case reflect.String:
		// MUST quote/escape to be a valid JSON string token.
		b, err := json.Marshal(v.String())
		if err != nil {
			return ""
		}
		return string(b)

	case reflect.Bool:
		return strconv.FormatBool(v.Bool())

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)

	case reflect.Uintptr:
		// Avoid precision loss & misleading semantics; emit as quoted hex address.
		return fmt.Sprintf("%q", fmt.Sprintf("0x%x", v.Uint()))

	case reflect.Float32, reflect.Float64:
		f := v.Float()
		if math.IsNaN(f) || math.IsInf(f, 0) {
			if floatsUseNullForNonFinite {
				return "null"
			}
			// Optional legacy behavior: treat as marshal error sentinel.
			return ""
		}
		bitSize := 64
		if v.Kind() == reflect.Float32 {
			bitSize = 32
		}
		// Match encoding/json's number formatting preference.
		return strconv.FormatFloat(f, 'g', -1, bitSize)

	case reflect.Complex64, reflect.Complex128:
		// JSON has no complex type; emit as an object {"real":..., "imag":...}.
		r, i := realFrom(v), imagFrom(v)
		return encodeComplexJSON(r, i, v.Kind() == reflect.Complex64)
	}

	// 5) Everything else: marshal with panic protection and optional pretty indent.
	s, err := safeMarshalJSONString(data, pretty)
	if err != nil {
		return ""
	}
	return s
}

// castJSONString converts a pointer to a string into a JSON string representation.
// If the pointer is nil, it returns the JSON literal "null". If the pointer points to an empty string, it returns an empty string.
//
// Parameters:
//   - `value`: A pointer to a string that may be nil or point to an empty string.
//
// Returns:
//   - A string containing the JSON representation of the input string pointer.
//
// Example:
//
//	str := "hello"
//	jsonStr := castJSONString(&str) // jsonStr will be "\"hello\""
func castJSONString(value *string) (string, error) {
	if value == nil {
		return "null", nil
	}
	as := *value
	if as == "" {
		return "", nil
	}
	return as, nil
}

// castJSONStringPtr converts a pointer to a string into a JSON string representation.
// If the pointer is nil, it returns the JSON literal "null". If the pointer points to an empty string, it returns an empty string.
//
// Parameters:
//   - `value`: A pointer to a string that may be nil or point to an empty string.
//
// Returns:
//   - A string containing the JSON representation of the input string pointer.
//
// Example:
//
//	str := "hello"
//	jsonStr := castJSONStringPtr(&str) // jsonStr will be "\"hello\""
func castJSONRawMessage(value *json.RawMessage, pretty bool) (string, error) {
	if value == nil {
		return "null", nil
	}
	as := *value
	if !json.Valid(as) {
		return "", ErrInvalidRawMessage
	}
	if !pretty {
		return string(as), nil
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, as, "", "    "); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// marshalJSONE converts a Go value to its JSON string representation or returns an error if the marshalling fails.
// It uses a deferred function to recover from any panics that may occur during marshalling.
//
// Parameters:
//   - `data`: The Go value to be converted to JSON.
//   - `pretty`: A boolean indicating whether the JSON should be pretty-printed.
//
// Returns:
//   - A string containing the JSON representation of the input value.
//   - An error if the marshalling fails.
//
// Example:
//
//	jsonString, err := marshalJSONE(myStruct, false)
func marshalJSONE(data any, pretty bool) (string, error) {
	if data == nil {
		return "", ErrNilInterface
	}

	switch v := data.(type) {
	case string:
		return castJSONString(&v)
	case *string:
		return castJSONString(v)
	case json.RawMessage:
		return castJSONRawMessage(&v, pretty)
	case *json.RawMessage:
		return castJSONRawMessage(v, pretty)
	}

	v := reflect.ValueOf(data)

	// 2) Unwrap *all* interface layers; nil interface => error.
	for v.Kind() == reflect.Interface {
		if v.IsNil() {
			return "", ErrNilInterface
		}
		v = v.Elem()
	}
	data = v.Interface()

	// 3) Nil-able kinds (Ptr, Map, Slice, Func, Chan) when nil => "null".
	v = reflect.ValueOf(data)
	switch v.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Func, reflect.Chan:
		if v.IsNil() {
			return "null", nil
		}
	}

	// 4) Scalar fast-path as valid JSON tokens.
	switch v.Kind() {
	case reflect.String:
		// MUST quote/escape to be a valid JSON string token.
		b, err := json.Marshal(v.String())
		if err != nil {
			return "", err
		}
		return string(b), nil

	case reflect.Bool:
		return strconv.FormatBool(v.Bool()), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10), nil

	case reflect.Uintptr:
		// Avoid precision loss & misleading semantics; emit as quoted hex address.
		return fmt.Sprintf("%q", fmt.Sprintf("0x%x", v.Uint())), nil

	case reflect.Float32, reflect.Float64:
		f := v.Float()
		if math.IsNaN(f) || math.IsInf(f, 0) {
			if floatsUseNullForNonFinite {
				return "null", nil
			}
			return "", ErrNonFiniteFloat
		}
		bitSize := 64
		if v.Kind() == reflect.Float32 {
			bitSize = 32
		}
		return strconv.FormatFloat(f, 'g', -1, bitSize), nil

	case reflect.Complex64, reflect.Complex128:
		// Encode as {"real":...,"imag":...}
		r, i := realFrom(v), imagFrom(v)
		s, err := encodeComplexJSONToken(r, i, v.Kind() == reflect.Complex64)
		return s, err
	}

	// 5) Everything else: marshal with panic protection and optional pretty indent.
	return safeMarshalJSONString(data, pretty)
}

// compactJSON processes a source byte slice and removes unwanted characters, returning a new cleaned-up byte slice.
//
// This function processes the input `src` byte slice and appends characters to the `dst` byte slice based on certain criteria.
// It specifically filters out characters that are not printable (i.e., characters with ASCII values greater than `' '`).
// Additionally, it handles quoted substrings, ensuring that characters inside properly escaped double quotes are preserved.
// If an unescaped double quote is encountered, it will stop processing further characters.
//
// Parameters:
//   - `dst`: The destination slice of bytes where the cleaned characters will be appended.
//   - `src`: The source slice of bytes to process, which may contain unwanted characters and quoted substrings.
//
// Returns:
//   - A new byte slice (`dst`) with unwanted characters removed. The cleaned-up version of `src`.
//
// Example:
//
//	src := []byte(`hello "world" 1234`)
//	dst := compactJSON([]byte{}, src)
//	// dst will be []byte{'h', 'e', 'l', 'l', 'o', ' ', '"', 'w', 'o', 'r', 'l', 'd', '"', ' ', '1', '2', '3', '4'},
//	// as the function preserves only printable characters and properly handles quoted substrings.
//
// Notes:
//   - This function skips characters that are not printable (less than or equal to ASCII ' ').
//   - When encountering a double quote (`"`), the function ensures that it correctly handles escaped quotes, skipping characters
//     until a valid closing quote is found. If an odd number of backslashes precede the closing quote, it breaks the loop to avoid
//     incorrect parsing of the quotes.
func compactJSON(dst, src []byte) []byte {
	dst = dst[:0] // Reset destination slice to an empty state
	for i := 0; i < len(src); i++ {
		if src[i] > ' ' { // Only include characters that are printable
			dst = append(dst, src[i])
			if src[i] == '"' { // Handle quoted substring (double quotes)
				for i = i + 1; i < len(src); i++ {
					dst = append(dst, src[i])
					if src[i] == '"' {
						// Search backwards for the last non-escaped backslash
						j := i - 1
						for ; ; j-- {
							if src[j] != '\\' {
								break
							}
						}
						// If the number of consecutive backslashes is odd, break the loop
						if (j-i)%2 != 0 {
							break
						}
					}
				}
			}
		}
	}
	return dst
}

// isNaNOrInf checks if a byte slice represents a special numeric value: NaN (Not-a-Number) or Infinity.
//
// This function inspects the first character of the input byte slice to determine if it represents
// either NaN or Infinity, including variations such as `Inf`, `+Inf`, `inf`, `NaN`, and `nan`.
// The function returns `true` if the input matches any of these special values.
//
// Parameters:
//   - `src`: A byte slice to inspect, typically a numeric string.
//
// Returns:
//   - `true` if the byte slice represents NaN or Infinity, `false` otherwise.
//
// Example:
//
//	src1 := []byte("Inf")
//	src2 := []byte("NaN")
//	src3 := []byte("+Inf")
//	src4 := []byte("infinity")
//	result1 := isNaNOrInf(src1) // result1 will be true
//	result2 := isNaNOrInf(src2) // result2 will be true
//	result3 := isNaNOrInf(src3) // result3 will be true
//	result4 := isNaNOrInf(src4) // result4 will be false
//
// Notes:
//   - The function only inspects the first character (or first two characters for lowercase `nan`) to make a quick determination.
//   - It supports the variations `Inf`, `+Inf`, `inf`, `NaN`, and `nan` as valid representations.
func isNaNOrInf(src []byte) bool {
	if len(src) == 0 {
		return false
	}
	return src[0] == 'i' || //Inf
		src[0] == 'I' || // inf
		src[0] == '+' || // +Inf
		src[0] == 'N' || // Nan
		(src[0] == 'n' && len(src) > 1 && src[1] != 'u') // nan
}

// getJSONType identifies the JSON type of a given byte slice based on its first character.
//
// This function analyzes the first character in the byte slice `v` to determine which JSON data type it represents.
// Based on the initial character, it categorizes the input as one of the following types:
// `jsonNull`, `jsonFalse`, `jsonTrue`, `jsonString`, `jNumber`, or `jsonJson` (indicating either a JSON object or array).
//
// Parameters:
//   - `v`: A byte slice representing a JSON value.
//
// Returns:
//   - A `jsonType` value that represents the JSON type of `v`, based on its first character.
//
// Example:
//
//	value1 := []byte(`"hello"`)
//	value2 := []byte("false")
//	value3 := []byte("123")
//	value4 := []byte("null")
//	value5 := []byte("[1, 2, 3]")
//	result1 := getJSONType(value1) // result1 will be jsonString
//	result2 := getJSONType(value2) // result2 will be jsonFalse
//	result3 := getJSONType(value3) // result3 will be jNumber
//	result4 := getJSONType(value4) // result4 will be jsonNull
//	result5 := getJSONType(value5) // result5 will be jsonJson
//
// Notes:
//   - If the byte slice is empty, the function returns `jsonNull`.
//   - The function uses the initial character of `v` to distinguish types, assuming `true`, `false`, and `null` are valid JSON values.
func getJSONType(v []byte) jsonType {
	if len(v) == 0 {
		return jsonNull
	}
	switch v[0] {
	case '"':
		return jsonString
	case 'f':
		return jsonFalse
	case 't':
		return jsonTrue
	case 'n':
		return jsonNull
	case '[', '{':
		return jsonJson
	default:
		return jNumber
	}
}

// unescapeJSONString extracts a JSON string from a byte slice, handling escaped characters when present.
//
// This function takes a JSON byte slice representing a string value and extracts the unescaped content.
// It iterates through the byte slice, checking for either escaped characters or the closing double quote (`"`).
// If an escape character (`\`) is detected, `unescapeJSONString` uses JSON unmarshalling to handle any escape sequences.
// If no escape character is encountered, it returns the substring between the opening and closing quotes.
//
// Parameters:
//   - `s`: A byte slice containing a JSON string, including the enclosing double quotes and potentially escaped characters.
//
// Returns:
//   - A byte slice with the unescaped content of the JSON string if valid, or `nil` if an error occurs.
//
// Example:
//
//	s := []byte(`"Hello, world!"`)
//	result := unescapeJSONString(s)
//	// result will be []byte{'H', 'e', 'l', 'l', 'o', ',', ' ', 'w', 'o', 'r', 'l', 'd', '!'}
//	s := []byte(`"Line1\nLine2"`)
//	result := unescapeJSONString(s)
//	// result will be []byte{'L', 'i', 'n', 'e', '1', '\n', 'L', 'i', 'n', 'e', '2'}
//
// Notes:
//   - If an escape sequence is encountered (`\`), JSON unmarshalling is used to correctly interpret it.
//   - If the byte slice does not contain a properly closed string, the function returns `nil`.
func unescapeJSONString(s []byte) []byte {
	for i := 1; i < len(s); i++ {
		if s[i] == '\\' {
			var str string
			if err := json.Unmarshal(s, &str); err != nil {
				return nil
			}
			return []byte(str)
		}
		if s[i] == '"' {
			return s[1:i]
		}
	}
	return nil
}

// sortObjectMembers sorts JSON key-value pairs in a stable order, preserving original formatting,
// and returns the updated buffer containing the sorted pairs.
//
// This function takes the JSON data, a buffer to store formatted values, and a list of key-value pairs (`pairs`).
// If there are no pairs to sort, it directly returns the buffer. Otherwise, it initializes a `byKeyVal` struct
// with the JSON data, buffer, and pairs, and sorts them by key (and by value if keys are identical) using
// `sort.Stable`. After sorting, it constructs a new byte slice with the sorted pairs in order, each followed
// by a comma and newline, and replaces the original content in `buf`.
//
// Parameters:
//   - `json`: The original JSON data as a byte slice.
//   - `buf`: A byte slice that holds formatted key-value pairs and can be modified in-place.
//   - `pairs`: A slice of `pair` structs representing the key-value pairs in the JSON data.
//
// Returns:
//   - A byte slice containing the buffer with sorted key-value pairs, in stable order.
//
// Example:
//
//	json := []byte(`{"b":2, "a":1}`)
//	pairs := []pair{ /* initialized with positions of keys and values */ }
//	buf := make([]byte, len(json))
//	result := sortObjectMembers(json, buf, pairs)
//	// result will contain sorted pairs by key.
//
// Notes:
//   - If `pairs` is empty, `buf` is returned unchanged.
//   - `sort.Stable` is used to ensure that pairs with identical keys maintain their original relative order.
//   - If `byKeyVal` marks pairs as unsorted, it skips replacing the original buffer.
func sortObjectMembers(json, buf []byte, pairs []kvPairs) []byte {
	if len(pairs) == 0 {
		return buf
	}
	_valStart := pairs[0].valueStart
	_valEnd := pairs[len(pairs)-1].valueEnd
	_keyVal := kvSorter{false, json, buf, pairs}
	sort.Stable(&_keyVal)
	if !_keyVal.sorted {
		return buf
	}
	n := make([]byte, 0, _valEnd-_valStart)
	for i, p := range pairs {
		n = append(n, buf[p.valueStart:p.valueEnd]...)
		if i < len(pairs)-1 {
			n = append(n, ',')
			n = append(n, '\n')
		}
	}
	return append(buf[:_valStart], n...)
}

// appendJSONString appends a JSON string value from the input JSON byte slice (`json`) to a buffer (`buf`),
// handling any escaped characters within the string, and returns the updated buffer and indices.
//
// This function begins at a given index `i` within a JSON byte slice `json`, assuming the current character is
// the start of a JSON string (`"`). It appends the entire string (from opening to closing quote) to the buffer `buf`,
// handling escaped quotes within the string. If an escape sequence (`\`) precedes a closing quote, it continues searching
// until it finds an unescaped closing quote, marking the end of the string.
//
// Parameters:
//   - `buf`: The destination byte slice to which the JSON string value is appended.
//   - `json`: The source JSON byte slice containing the entire JSON structure.
//   - `i`: The starting index in `json`, pointing to the beginning of the JSON string (initial quote).
//   - `nl`: The current newline position (used for pretty-printing in larger context).
//
// Returns:
//   - `buf`: The updated buffer, containing the appended JSON string.
//   - `i`: The updated index in `json` after the end of the string (right after the closing quote).
//   - `nl`: The unchanged newline position (for tracking in pretty-printing).
//   - `true`: A boolean flag indicating the function processed a string (for handling in other contexts).
//
// Example:
//
//	json := []byte(`"example \"string\" value"`)
//	buf := []byte{}
//	buf, i, nl, processed := appendJSONString(buf, json, 0, 0)
//	// buf will contain `example \"string\" value`, i will point to the next index after the closing quote,
//	// nl remains the same, and processed is true.
//
// Notes:
//   - The function counts consecutive backslashes before each closing quote to determine if it is escaped.
//   - It appends the entire string (including quotes) to `buf` for easy integration in pretty-printing or formatting routines.
func appendJSONString(buf, json []byte, i, nl int) ([]byte, int, int, bool) {
	s := i
	i++
	for ; i < len(json); i++ {
		if json[i] == '"' {
			var sc int
			for j := i - 1; j > s; j-- {
				if json[j] == '\\' {
					sc++
				} else {
					break
				}
			}
			if sc%2 == 1 {
				continue
			}
			i++
			break
		}
	}
	return append(buf, json[s:i]...), i, nl, true
}

// appendJSONNumber appends a JSON number value from the input JSON byte slice (`json`) to a buffer (`buf`),
// and returns the updated buffer and indices.
//
// This function starts at a given index `i` within a JSON byte slice `json` (assuming the current character is the start
// of a JSON number). It appends the entire number to the buffer `buf` and handles all characters up to the next
// non-number character, such as spaces, commas, colons, or closing brackets/braces.
//
// Parameters:
//   - `buf`: The destination byte slice to which the JSON number value will be appended.
//   - `json`: The source JSON byte slice containing the entire JSON structure.
//   - `i`: The starting index in `json`, pointing to the first character of the JSON number.
//   - `nl`: The current newline position (used for pretty-printing in larger context).
//
// Returns:
//   - `buf`: The updated buffer, containing the appended JSON number value.
//   - `i`: The updated index in `json` after the number (right after the last character of the number).
//   - `nl`: The unchanged newline position (for tracking in pretty-printing).
//   - `true`: A boolean flag indicating that a number was processed successfully.
//
// Example:
//
//	json := []byte(`12345`)
//	buf := []byte{}
//	buf, i, nl, processed := appendJSONNumber(buf, json, 0, 0)
//	// buf will contain `12345`, i will point to the next index after the number,
//	// nl remains unchanged, and processed will be true.
//
// Notes:
//   - The function scans for all characters that are part of the number (digits, decimal point, etc.) until it
//     encounters a character that is not part of a valid number, such as a space, comma, colon, or closing bracket/braces.
//   - It assumes that the number is well-formed and does not handle error cases like invalid numbers.
func appendJSONNumber(buf, json []byte, i, nl int) ([]byte, int, int, bool) {
	s := i // Record the start index of the number
	i++    // Move past the initial digit (or minus sign if present)
	for ; i < len(json); i++ {
		// Break the loop if a non-number character is encountered (e.g., space, comma, colon, bracket, or brace)
		if json[i] <= ' ' || json[i] == ',' || json[i] == ':' || json[i] == ']' || json[i] == '}' {
			break
		}
	}
	// Append the number from the start index `s` to the updated index `i` (excluding non-number characters)
	return append(buf, json[s:i]...), i, nl, true
}

// appendJSONContainer processes the next JSON object or array in the input JSON byte slice (`json`)
// and appends it to the buffer (`buf`), while handling pretty-printing, sorting of object keys, and enforcing width constraints.
//
// This function handles the parsing and formatting of JSON objects (`{}`) and arrays (`[]`), adding appropriate indentation,
// newlines, and sorting of keys (if specified). It ensures that the resulting object or array is correctly formatted and
// inserted into the buffer, maintaining the structure and respecting pretty-printing preferences.
//
// It also handles the optional width limit (to control the length of single-line arrays) and can process objects with sorted keys,
// ensuring proper formatting of both simple and complex JSON objects.
//
// Parameters:
//   - `buf`: The destination byte slice to which the processed JSON object or array will be appended.
//   - `json`: The source JSON byte slice containing the entire JSON structure.
//   - `i`: The starting index in `json` from where the next object or array should be processed.
//   - `open`: The opening byte (either '{' for an object or '[' for an array).
//   - `close`: The closing byte (either '}' for an object or ']' for an array).
//   - `pretty`: A boolean flag indicating whether pretty-printing should be applied (i.e., adding newlines and indentation).
//   - `width`: The width used for pretty-printing (influences line breaks for large arrays).
//   - `prefix`: A string prefix used for leading indentation (used in pretty-printing).
//   - `indent`: The string used for indentation in pretty-printing.
//   - `sortKeys`: A boolean flag indicating whether the keys in objects should be sorted.
//   - `tabs`: The number of tabs for indentation (used in pretty-printing).
//   - `nl`: The current newline position, used for managing where to insert newlines during pretty-printing.
//   - `max`: The maximum number of characters to pretty-print before breaking into a new line (relevant for width-based formatting).
//
// Returns:
//   - `buf`: The updated buffer containing the appended JSON object or array (pretty-printed if the `pretty` flag is true).
//   - `i`: The updated index in `json`, pointing to the position after the processed object or array.
//   - `nl`: The updated newline position, adjusted for pretty-printing.
//   - `true` or `false`: A boolean flag indicating whether the processing was successful. The function returns `false` if
//     there is an issue (for example, exceeding the width limit or malformed data).
//
// Example usage:
//
//	json := []byte(`{ "key1": 123, "key2": "value", "key3": [1, 2, 3] }`)
//	buf := []byte{}
//	buf, i, nl, processed := appendJSONContainer(buf, json, 0, '{', '}', true, 80, "", "  ", false, 0, 0, -1)
//	// buf will contain the pretty-printed JSON object, i will point to the next index,
//	// nl will be adjusted for newlines, and processed will be true.
//
// Notes:
//   - This function handles both JSON objects and arrays, depending on the value of `open` and `close` (either '{', '}' for objects
//     or '[' and ']' for arrays).
//   - Pretty-printing is applied based on the `pretty` flag, including indentation and line breaks.
//   - If `sortKeys` is set to true, the keys in the JSON object will be sorted lexicographically before being appended to the buffer.
//   - The `max` value helps control the number of characters in a single line for arrays, ensuring that arrays are properly wrapped into multiple lines if necessary.
//   - The function can also handle arrays and objects with nested structures, ensuring the formatting remains correct throughout.
func appendJSONContainer(buf, json []byte, i int, open, close byte, pretty bool, width int, prefix, indent string, sortKeys bool, tabs, nl, max int) ([]byte, int, int, bool) {
	var ok bool
	if width > 0 {
		if pretty && open == '[' && max == -1 {
			// here we try to create a single line array
			max := width - (len(buf) - nl)
			if max > 3 {
				s1, s2 := len(buf), i
				buf, i, _, ok = appendJSONContainer(buf, json, i, '[', ']', false, width, prefix, "", sortKeys, 0, 0, max)
				if ok && len(buf)-s1 <= max {
					return buf, i, nl, true
				}
				buf = buf[:s1]
				i = s2
			}
		} else if max != -1 && open == '{' {
			return buf, i, nl, false
		}
	}
	buf = append(buf, open)
	i++
	var pairs []kvPairs
	if open == '{' && sortKeys {
		pairs = make([]kvPairs, 0, 8)
	}
	var n int
	for ; i < len(json); i++ {
		if json[i] <= ' ' {
			continue
		}
		if json[i] == close {
			if pretty {
				if open == '{' && sortKeys {
					buf = sortObjectMembers(json, buf, pairs)
				}
				if n > 0 {
					nl = len(buf)
					if buf[nl-1] == ' ' {
						buf[nl-1] = '\n'
					} else {
						buf = append(buf, '\n')
					}
				}
				if buf[len(buf)-1] != open {
					buf = appendIndent(buf, prefix, indent, tabs)
				}
			}
			buf = append(buf, close)
			return buf, i + 1, nl, open != '{'
		}
		if open == '[' || json[i] == '"' {
			if n > 0 {
				buf = append(buf, ',')
				if width != -1 && open == '[' {
					buf = append(buf, ' ')
				}
			}
			var p kvPairs
			if pretty {
				nl = len(buf)
				if buf[nl-1] == ' ' {
					buf[nl-1] = '\n'
				} else {
					buf = append(buf, '\n')
				}
				if open == '{' && sortKeys {
					p.keyStart = i
					p.valueStart = len(buf)
				}
				buf = appendIndent(buf, prefix, indent, tabs+1)
			}
			if open == '{' {
				buf, i, nl, _ = appendJSONString(buf, json, i, nl)
				if sortKeys {
					p.keyEnd = i
				}
				buf = append(buf, ':')
				if pretty {
					buf = append(buf, ' ')
				}
			}
			buf, i, nl, ok = appendJSONValue(buf, json, i, pretty, width, prefix, indent, sortKeys, tabs+1, nl, max)
			if max != -1 && !ok {
				return buf, i, nl, false
			}
			if pretty && open == '{' && sortKeys {
				p.valueEnd = len(buf)
				if p.keyStart > p.keyEnd || p.valueStart > p.valueEnd {
					// bad data. disable sorting
					sortKeys = false
				} else {
					pairs = append(pairs, p)
				}
			}
			i--
			n++
		}
	}
	return buf, i, nl, open != '{'
}

// appendJSONByte appends a byte `c` to the destination byte slice `dst`,
// handling special control characters by escaping them in Unicode format.
// If `c` is a control character (ASCII value less than 32) and not one of the
// common whitespace characters (`\r`, `\n`, `\t`, `\v`), it appends
// the Unicode escape sequence `\u00XX` to `dst`, where `XX` is the hexadecimal
// representation of `c`. Otherwise, it appends `c` directly to `dst`.
func appendJSONByte(dst []byte, c byte) []byte {
	if c < ' ' && (c != '\r' && c != '\n' && c != '\t' && c != '\v') {
		dst = append(dst, "\\u00"...)
		dst = append(dst, hexDigit((c>>4)&0xF))
		return append(dst, hexDigit(c&0xF))
	}
	return append(dst, c)
}

// appendJSONValue processes the next JSON value in the input JSON byte slice (`json`) and appends it to the buffer (`buf`),
// while handling different types of JSON values (strings, numbers, objects, arrays, and literals).
// It returns the updated buffer and indices, as well as a boolean flag indicating whether a value was processed.
//
// This function is responsible for recognizing the type of the next JSON value and delegating the task of appending that value
// to the appropriate helper function (such as `appendPrettyString` for strings, `appendPrettyNumber` for numbers, and others
// for objects, arrays, and literals). It processes the JSON byte slice one element at a time and handles all value types
// correctly, ensuring that each value is pretty-printed if required.
//
// Parameters:
//   - `buf`: The destination byte slice to which the processed JSON value will be appended.
//   - `json`: The source JSON byte slice containing the entire JSON structure.
//   - `i`: The starting index in `json` from where the next value should be processed.
//   - `pretty`: A boolean flag indicating whether pretty-printing should be applied (i.e., adding newlines and indentation).
//   - `width`: The width used for pretty-printing (not used in this function, but passed for consistency in pretty-printing logic).
//   - `prefix`: A string prefix (used in pretty-printing to add leading indentation, not used here).
//   - `indent`: The string used for indentation (not used here but part of the pretty-printing configuration).
//   - `sortKeys`: A boolean flag indicating whether the keys in objects should be sorted (not used here but passed for consistency).
//   - `tabs`: The number of tabs for indentation (not used here but passed for consistency in pretty-printing logic).
//   - `nl`: The current newline position (used for pretty-printing, ensuring line breaks are maintained correctly).
//   - `max`: The maximum number of characters to pretty-print before breaking into a new line (not used in this function).
//
// Returns:
//   - `buf`: The updated buffer, containing the appended JSON value (pretty-printed if the `pretty` flag is true).
//   - `i`: The updated index in `json`, pointing to the position after the processed value.
//   - `nl`: The unchanged newline position (used for pretty-printing in the larger context).
//   - `true`: A boolean flag indicating that a JSON value was successfully processed.
//
// Example usage:
//
//	json := []byte(`{ "key1": 123, "key2": "value", "key3": [1, 2, 3] }`)
//	buf := []byte{}
//	buf, i, nl, processed := appendJSONValue(buf, json, 0, true, 0, "", "  ", false, 0, 0, 0)
//	// buf will contain the pretty-printed JSON value, i will point to the next index,
//	// nl remains unchanged, and processed will be true.
//
// Notes:
//   - This function processes and appends various JSON data types, including strings, numbers, objects, arrays,
//     and literals (`true`, `false`, `null`).
//   - The function assumes the JSON is valid and well-formed; it does not handle parsing errors for invalid JSON.
func appendJSONValue(buf, json []byte, i int, pretty bool, width int, prefix, indent string, sortKeys bool, tabs, nl, max int) ([]byte, int, int, bool) {
	for ; i < len(json); i++ {
		if json[i] <= ' ' {
			continue
		}
		if json[i] == '"' {
			return appendJSONString(buf, json, i, nl)
		}
		if (json[i] >= '0' && json[i] <= '9') || json[i] == '-' || isNaNOrInf(json[i:]) {
			return appendJSONNumber(buf, json, i, nl)
		}
		if json[i] == '{' {
			return appendJSONContainer(buf, json, i, '{', '}', pretty, width, prefix, indent, sortKeys, tabs, nl, max)
		}
		if json[i] == '[' {
			return appendJSONContainer(buf, json, i, '[', ']', pretty, width, prefix, indent, sortKeys, tabs, nl, max)
		}
		switch json[i] {
		case 't':
			return append(buf, 't', 'r', 'u', 'e'), i + 4, nl, true
		case 'f':
			return append(buf, 'f', 'a', 'l', 's', 'e'), i + 5, nl, true
		case 'n':
			return append(buf, 'n', 'u', 'l', 'l'), i + 4, nl, true
		}
	}
	return buf, i, nl, true
}

// appendIndent appends indentation to the provided buffer (`buf`) based on the specified `prefix`, `indent`,
// and the number of `tabs` to insert.
//
// This function adds a specific number of tab or space-based indents to the `buf`, depending on the `indent` value.
// If the `indent` string contains exactly two spaces, it will append spaces for each tab; otherwise, it uses
// the provided `indent` string (which can represent any indenting character, such as tabs or custom strings).
// Additionally, if a `prefix` is provided, it will be prepended to the buffer before any indentation is added.
//
// Parameters:
//   - `buf`: The byte slice to which the indentations are appended.
//   - `prefix`: A string (byte slice) that will be prepended to `buf` before any indentation (optional).
//   - `indent`: The string to use for indentation, typically consisting of spaces or tabs (e.g., `"\t"` or `"  "`).
//   - `tabs`: The number of times the `indent` should be repeated to represent the desired level of indentation.
//
// Returns:
//   - The updated `buf` with the appropriate amount of indentation based on `tabs` and `indent`.
//
// Example:
//
//	buf := []byte{}
//	prefix := "  "
//	indent := "\t"
//	tabs := 3
//	buf = appendIndent(buf, prefix, indent, tabs)
//	// buf will be `{"  "\t\t\t` (prefix followed by 3 tab characters).
//
// Notes:
//   - If the `indent` string is exactly two spaces (`"  "`), the function will append two spaces for each `tab`.
//   - If the `indent` string is anything else, it will be appended `tabs` times.
func appendIndent(buf []byte, prefix, indent string, tabs int) []byte {
	if len(prefix) != 0 { // Append prefix if it's not an empty string
		buf = append(buf, prefix...)
	}
	// Check if the indent is exactly two spaces and append spaces for each tab
	if len(indent) == 2 && indent[0] == ' ' && indent[1] == ' ' {
		for i := 0; i < tabs; i++ {
			buf = append(buf, ' ', ' ')
		}
	} else {
		// Otherwise, append the custom indent string for each tab
		for i := 0; i < tabs; i++ {
			buf = append(buf, indent...)
		}
	}
	return buf
}

// hexDigit converts a numeric value to its corresponding hexadecimal character.
//
// This function takes a single byte `p`, which represents a numeric value, and converts it to its
// hexadecimal equivalent as a byte. The function assumes that the input `p` is in the range of 0 to 15
// (i.e., it represents a single hexadecimal digit).
//
// Parameters:
//   - `p`: A byte representing a numeric value between 0 and 15 (inclusive).
//
// Returns:
//   - A byte representing the corresponding hexadecimal character.
//   - If `p` is less than 10, it returns the ASCII character for the corresponding digit (0-9).
//   - If `p` is 10 or greater, it returns the lowercase ASCII character for the corresponding letter (a-f).
//
// Example:
//
//	hexDigit(0)  // returns '0'
//	hexDigit(9)  // returns '9'
//	hexDigit(10) // returns 'a'
//	hexDigit(15) // returns 'f'
//
// Notes:
//   - The function works only for values between 0 and 15 (inclusive).
//   - For input values greater than 15, the behavior is not defined and may lead to unexpected results.
func hexDigit(p byte) byte {
	// If p is less than 10, return the corresponding digit character ('0' to '9')
	switch {
	case p < 10:
		return p + '0' // Add ASCII value of '0' to convert to character
	default:
		// If p is 10 or greater, return the corresponding letter character ('a' to 'f')
		// Add ASCII value of 'a' to get 'a' to 'f'
		return (p - 10) + 'a'
	}
}

// sanitizeJSON processes a source byte slice (source) and removes or replaces comment sections,
// formatting them into a cleaned-up destination byte slice (destination).
// It handles both single-line (`//`) and multi-line (`/* */`) comments and strips them out,
// replacing them with spaces or newlines as appropriate.
//
// Parameters:
//   - `source`: The input byte slice containing the source code to process.
//   - `destination`: The output byte slice that will contain the cleaned code (without comments).
//     It is assumed to be initialized as an empty slice.
//
// Returns:
//   - A byte slice representing the cleaned source code with comments removed or replaced.
//   - Single-line comments (`//`) are replaced with spaces.
//   - Multi-line comments (`/* */`) are replaced with spaces and preserved newlines for line breaks.
//
// Example:
//
//	source := []byte("int x = 10; // initialize x\n/* multi-line\n comment */")
//	destination := sanitizeJSON(source, []byte{})
//	// destination will contain: "int x = 10;    \n   "
//
// Notes:
//   - The function handles the following types of comments:
//   - Single-line comments starting with `//` and ending with the newline (`\n`).
//   - Multi-line comments enclosed in `/* */`, even if they span multiple lines.
//   - Strings inside double quotes (`"`) and special characters like `}` or `]` are preserved as is,
//     with careful handling of quotes and escape sequences within the string.
func sanitizeJSON(source, destination []byte) []byte {
	destination = destination[:0]
	for i := 0; i < len(source); i++ {
		if source[i] == '/' {
			if i < len(source)-1 {
				if source[i+1] == '/' {
					destination = append(destination, ' ', ' ')
					i += 2
					for ; i < len(source); i++ {
						if source[i] == '\n' {
							destination = append(destination, '\n')
							break
						} else if source[i] == '\t' || source[i] == '\r' {
							destination = append(destination, source[i])
						} else {
							destination = append(destination, ' ')
						}
					}
					continue
				}
				if source[i+1] == '*' {
					destination = append(destination, ' ', ' ')
					i += 2
					for ; i < len(source)-1; i++ {
						if source[i] == '*' && source[i+1] == '/' {
							destination = append(destination, ' ', ' ')
							i++
							break
						} else if source[i] == '\n' || source[i] == '\t' ||
							source[i] == '\r' {
							destination = append(destination, source[i])
						} else {
							destination = append(destination, ' ')
						}
					}
					continue
				}
			}
		}
		destination = append(destination, source[i])
		if source[i] == '"' {
			for i = i + 1; i < len(source); i++ {
				destination = append(destination, source[i])
				if source[i] == '"' {
					j := i - 1
					for ; ; j-- {
						if source[j] != '\\' {
							break
						}
					}
					if (j-i)%2 != 0 {
						break
					}
				}
			}
		} else if source[i] == '}' || source[i] == ']' {
			for j := len(destination) - 2; j >= 0; j-- {
				if destination[j] <= ' ' {
					continue
				}
				if destination[j] == ',' {
					destination[j] = ' '
				}
				break
			}
		}
	}
	return destination
}
