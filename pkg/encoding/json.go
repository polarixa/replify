package encoding

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/polarixa/replify/pkg/strutil"
)

// MarshalJSON converts a Go value into its JSON byte representation.
//
// This function marshals the input value `v` using the standard json library.
// The resulting JSON data is returned as a byte slice. If there is an error
// during marshalling, it returns the error.
//
// Parameters:
//   - `v`: The Go value to be marshalled into JSON.
//
// Returns:
//   - A byte slice containing the JSON representation of the input value.
//   - An error if the marshalling fails.
//
// Example:
//
//	jsonData, err := MarshalJSON(myStruct)
func MarshalJSON(v any) ([]byte, error) {
	return json.Marshal(v)
}

// MarshalJSONString converts a Go value to its JSON string representation.
//
// This function utilizes the standard json library to marshal the input value `v`
// into a JSON string. If the marshalling is successful, it returns the resulting
// JSON string. If an error occurs during the process, it returns an error.
//
// Parameters:
//   - `v`: The Go value to be marshalled into JSON.
//
// Returns:
//   - A string containing the JSON representation of the input value.
//   - An error if the marshalling fails.
//
// Example:
//
//	jsonString, err := MarshalJSONString(myStruct)
func MarshalJSONString(v any) (string, error) {
	data, err := MarshalJSON(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// MarshalJSONIndent converts a Go value to its JSON string representation with indentation.
//
// This function marshals the input value `v` into a formatted JSON string,
// allowing for easy readability by including a specified prefix and indentation.
// It returns the resulting JSON byte slice or an error if marshalling fails.
//
// Parameters:
//   - `v`: The Go value to be marshalled into JSON.
//   - `prefix`: A string that will be prefixed to each line of the output JSON.
//   - `indent`: A string used for indentation (typically a series of spaces or a tab).
//
// Returns:
//   - A byte slice containing the formatted JSON representation of the input value.
//   - An error if the marshalling fails.
//
// Example:
//
//	jsonIndented, err := MarshalJSONIndent(myStruct, "", "    ")
func MarshalJSONIndent(v any, prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(v, prefix, indent)
}

// UnmarshalJSON parses JSON-encoded data and stores the result in the value pointed to by `v`.
//
// This function uses the standard json library to unmarshal JSON data
// (given as a byte slice) into the specified Go value `v`. If the unmarshalling
// is successful, it populates the value `v`. If an error occurs, it returns the error.
//
// Parameters:
//   - `data`: A byte slice containing JSON data to be unmarshalled.
//   - `v`: A pointer to the Go value where the unmarshalled data will be stored.
//
// Returns:
//   - An error if the unmarshalling fails.
//
// Example:
//
//	err := UnmarshalJSON(jsonData, &myStruct)
func UnmarshalJSON(jsonValue []byte, v any) error {
	return json.Unmarshal(jsonValue, v)
}

// UnmarshalJSONString parses JSON-encoded string and stores the result in the value pointed to by `v`.
//
// This function utilizes the standard json library to unmarshal JSON data
// from a string into the specified Go value `v`. If the unmarshalling is
// successful, it populates the value `v`. If an error occurs, it returns the error.
//
// Parameters:
//   - `jsonStr`: A string containing JSON data to be unmarshalled.
//   - `v`: A pointer to the Go value where the unmarshalled data will be stored.
//
// Returns:
//   - An error if the unmarshalling fails.
//
// Example:
//
//	err := UnmarshalJSONString(jsonString, &myStruct)
func UnmarshalJSONString(jsonValue string, v any) error {
	return UnmarshalJSON([]byte(jsonValue), v)
}

// StrictUnmarshalJSON parses JSON-encoded data and stores the result in the value pointed to by `v`.
//
// This function uses the standard json library to unmarshal JSON data
// (given as a byte slice) into the specified Go value `v`. If the unmarshalling
// is successful, it populates the value `v`. If an error occurs, it returns the error.
//
// Parameters:
//   - `data`: A byte slice containing JSON data to be unmarshalled.
//   - `v`: A pointer to the Go value where the unmarshalled data will be stored.
//
// Returns:
//   - An error if the unmarshalling fails.
//
// Example:
//
//	err := StrictUnmarshalJSON(jsonData, &myStruct)
func StrictUnmarshalJSON(jsonValue []byte, v any) error {
	if len(jsonValue) == 0 {
		return fmt.Errorf("%w: JSON data must not be empty", ErrEmptyInput)
	}

	if !IsValidJSON(jsonValue) {
		return fmt.Errorf("%w: input is not valid JSON", ErrInvalidJSON)
	}

	return UnmarshalJSON(jsonValue, v)
}

// StrictUnmarshalJSONString parses JSON-encoded string and stores the result in the value pointed to by `v`.
//
// This function uses the standard json library to unmarshal JSON data
// (given as a string) into the specified Go value `v`. If the unmarshalling
// is successful, it populates the value `v`. If an error occurs, it returns the error.
//
// Parameters:
//   - `jsonStr`: A string containing JSON data to be unmarshalled.
//   - `v`: A pointer to the Go value where the unmarshalled data will be stored.
//
// Returns:
//   - An error if the unmarshalling fails.
//
// Example:
//
//	err := StrictUnmarshalJSONString(jsonString, &myStruct)
func StrictUnmarshalJSONString(json string, v any) error {
	if strutil.IsEmpty(json) {
		return fmt.Errorf("%w: JSON string must not be empty", ErrEmptyInput)
	}

	if !IsValidJSONString(json) {
		return fmt.Errorf("%w: input is not valid JSON", ErrInvalidJSON)
	}

	return UnmarshalJSONString(json, v)
}

// IsValidJSON checks if a given byte slice is a valid JSON format.
//
// This function uses the json.Valid method from the standard json library
// to determine if the input byte slice `data` is a valid JSON representation.
//
// Parameters:
//   - `data`: The byte slice to be validated as JSON.
//
// Returns:
//   - A boolean indicating whether the input byte slice is valid JSON.
func IsValidJSON(jsonValue []byte) bool {
	return json.Valid(jsonValue)
}

// IsValidJSONString checks if a given string is a valid JSON format.
//
// This function uses the json.Valid method from the standard json library
// to determine if the input string `s` is a valid JSON representation.
//
// Parameters:
//   - `s`: The string to be validated as JSON.
//
// Returns:
//   - A boolean indicating whether the input string is valid JSON.
func IsValidJSONString(jsonValue string) bool {
	return IsValidJSON([]byte(jsonValue))
}

// NormalizeJSON attempts to normalize a malformed JSON-like string into valid JSON.
//
// Normalization strategy — passes are applied in sequence; validity is checked
// after each pass that modifies the candidate. The function returns as soon as
// a pass produces a valid JSON string, so no unnecessary work is performed.
//
//  1. Empty / whitespace-only input → return error.
//  2. Already valid JSON → return unchanged (fast path, no allocation).
//  3. Pass 1 - strip a leading UTF-8 BOM (U+FEFF / 0xEF 0xBB 0xBF).
//  4. Pass 2 - remove embedded null bytes (0x00) which are invalid inside JSON text.
//  5. Pass 3 - unescape literal `\"` sequences to `"`.  This is the most common
//     artifact produced when JSON is stored in Go raw string literals or travels
//     through systems that double-escape structural quote characters.
//  6. Pass 4 - remove trailing commas before `}` or `]`.  These are produced by
//     some serializers and are not permitted by the JSON grammar.
//
// Passes are cumulative: each pass operates on the output of the previous one.
// The function does NOT silently corrupt already-valid JSON (step 2 guarantees
// this). Only inputs that fail the initial validation ever enter the pass chain.
//
// Parameters:
//   - s: The input string to normalize.
//
// Returns:
//   - A valid JSON string on success.
//   - An error if the input is empty/whitespace or cannot be normalized to valid JSON.
//
// Example:
//
//	normalized, err := NormalizeJSON(`{\"key\": "value"}`)
func NormalizeJSON(s string) (string, error) {
	if strutil.IsEmpty(s) {
		return "", fmt.Errorf("%w: JSON string must not be empty", ErrEmptyInput)
	}

	// Fast path: already valid JSON — return as-is with no allocation.
	if IsValidJSONString(s) {
		return s, nil
	}

	candidate := s

	// Pass 1: Strip leading UTF-8 BOM (0xEF 0xBB 0xBF).
	if strings.HasPrefix(candidate, "\xEF\xBB\xBF") {
		candidate = candidate[3:]
		if IsValidJSONString(candidate) {
			return candidate, nil
		}
	}

	// Pass 2: Remove embedded null bytes.
	if strings.Contains(candidate, "\x00") {
		candidate = strings.ReplaceAll(candidate, "\x00", "")
		if IsValidJSONString(candidate) {
			return candidate, nil
		}
	}

	// Pass 3: Unescape literal \" → " (structural quote escape artifacts).
	if strings.Contains(candidate, `\"`) {
		candidate = strings.ReplaceAll(candidate, `\"`, `"`)
		if IsValidJSONString(candidate) {
			return candidate, nil
		}
	}

	// Pass 4: Remove trailing commas before } or ] (invalid in JSON grammar).
	if noTrailing := normalizeTrailingCommaRe.ReplaceAllString(candidate, "$1"); noTrailing != candidate {
		candidate = noTrailing
		if IsValidJSONString(candidate) {
			return candidate, nil
		}
	}

	return candidate, fmt.Errorf("%w: input could not be repaired to valid JSON", ErrInvalidJSON)
}

// JSON converts a Go value to its JSON string representation or returns an error if the marshalling fails.
// It uses a deferred function to recover from any panics that may occur during marshalling.
//
// Parameters:
//   - `data`: The Go value to be converted to JSON.
//
// Returns:
//   - A string containing the JSON representation of the input value.
//   - An error if the marshalling fails.
//
// Example:
//
//	 var myStruct = struct {
//	 	Name string
//	 	Age  int
//	 }{
//	 	Name: "John",
//	 	Age:  30,
//	 }
//		jsonString, err := JSON(myStruct)
func JSON(data any) string {
	return marshalJSON(data, false)
}

// JSONE converts a Go value to its JSON string representation or returns an error if the marshalling fails.
// It uses a deferred function to recover from any panics that may occur during marshalling.
//
// Parameters:
//   - `data`: The Go value to be converted to JSON.
//
// Returns:
//   - A string containing the JSON representation of the input value.
//   - An error if the marshalling fails.
//
// Example:
//
//	 var myStruct = struct {
//	 	Name string
//	 	Age  int
//	 }{
//	 	Name: "John",
//	 	Age:  30,
//	 }
//		jsonString, err := JSONE(myStruct)
func JSONE(data any) (string, error) {
	return marshalJSONE(data, false)
}

// JSONPretty converts a Go value to its pretty-printed JSON string representation or returns an error if the marshalling fails.
// It uses a deferred function to recover from any panics that may occur during marshalling.
//
// Parameters:
//   - `data`: The Go value to be converted to JSON.
//
// Returns:
//   - A string containing the JSON representation of the input value.
//   - An error if the marshalling fails.
//
// Example:
//
//	 var myStruct = struct {
//	 	Name string
//	 	Age  int
//	 }{
//	 	Name: "John",
//	 	Age:  30,
//	 }
//		jsonString, err := JSONPretty(myStruct)
func JSONPretty(data any) string {
	return marshalJSON(data, true)
}

// JSONPrettyE converts a Go value to its pretty-printed JSON string representation or returns an error if the marshalling fails.
// It uses a deferred function to recover from any panics that may occur during marshalling.
//
// Parameters:
//   - `data`: The Go value to be converted to JSON.
//
// Returns:
//   - A string containing the JSON representation of the input value.
//   - An error if the marshalling fails.
//
// Example:
//
//	 var myStruct = struct {
//	 	Name string
//	 	Age  int
//	 }{
//	 	Name: "John",
//	 	Age:  30,
//	 }
//		jsonString, err := JSONPrettyE(myStruct)
func JSONPrettyE(data any) (string, error) {
	return marshalJSONE(data, true)
}
