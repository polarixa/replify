package conv

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"

	"github.com/polarixa/replify/pkg/strutil"
)

// FromJSON parses a JSON string and populates the provided variable with the resulting data structure.
//
// Parameters:
//   - `jsonStr`: The JSON string to be parsed.
//   - `into`: A pointer to the variable where the parsed data will be stored.
//
// Returns:
//   - An error if the parsing fails or if the input JSON string is empty (unless `emptyAsZero` is set).
//
// Example:
//
//	var myData MyStruct
//	err := conv.FromJSON(jsonString, &myData)
//	if err != nil {
//	    // handle error
//	}
func (c *Converter) FromJSON(jsonStr string, into any) error {
	if strutil.IsEmpty(jsonStr) {
		if c.emptyAsZero {
			return nil
		}
		return newConvErrorMsg("cannot parse empty JSON string")
	}

	decoder := json.NewDecoder(strings.NewReader(jsonStr))
	decoder.UseNumber() // Preserve number precision

	if err := decoder.Decode(into); err != nil {
		return newConvErrorf("JSON parse error: %v", err)
	}

	return nil
}

// FromJSONBytes parses a JSON byte slice and populates the provided variable with the resulting data structure.
//
// Parameters:
//   - `data`: The JSON byte slice to be parsed.
//   - `into`: A pointer to the variable where the parsed data will be stored.
//
// Returns:
//   - An error if the parsing fails or if the input byte slice is empty (unless `emptyAsZero` is set).
//
// Example:
//
//	var myData MyStruct
//	err := conv.FromJSONBytes(jsonBytes, &myData)
//	if err != nil {
//	    // handle error
//	}
func (c *Converter) FromJSONBytes(data []byte, into any) error {
	if len(data) == 0 {
		if c.emptyAsZero {
			return nil
		}
		return newConvErrorMsg("cannot parse empty JSON bytes")
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	if err := decoder.Decode(into); err != nil {
		return newConvErrorf("JSON parse error: %v", err)
	}

	return nil
}

// FromJSONReader parses JSON data from an io.Reader and populates the provided variable with the resulting data structure.
//
// Parameters:
//   - `r`: An io.Reader from which to read the JSON data.
//   - `into`: A pointer to the variable where the parsed data will be stored.
//
// Returns:
//   - An error if the parsing fails.
//
// Example:
//
//	var myData MyStruct
//	reader := strings.NewReader(jsonString) // or use file reader, network reader, etc.
//	file, err := os.Open("data.json") // use file reader
//	err := conv.FromJSONReader(reader, &myData)
//	if err != nil {
//		   // handle error
//	}
func (c *Converter) FromJSONReader(r io.Reader, into any) error {
	decoder := json.NewDecoder(r)
	decoder.UseNumber()

	if err := decoder.Decode(into); err != nil {
		return newConvErrorf("JSON parse error: %v", err)
	}

	return nil
}

// ParseJSON is a package-level helper that parses a JSON string into a variable of type T.
//
// Parameters:
//   - `jsonStr`: The JSON string to be parsed.
//
// Returns:
//   - A variable of type T populated with the parsed data.
//   - An error if the parsing fails.
//
// Example:
//
//	var myData MyStruct
//	myData, err := conv.ParseJSON[MyStruct](jsonString)
//	if err != nil {
//	    // handle error
//	}
func ParseJSON[T any](jsonStr string) (T, error) {
	var result T
	if err := defaultConverter.FromJSON(jsonStr, &result); err != nil {
		return result, err
	}
	return result, nil
}

// ParseJSONBytes is a package-level helper that parses a JSON byte slice into a variable of type T.
//
// Parameters:
//   - `data`: The JSON byte slice to be parsed.
//
// Returns:
//   - A variable of type T populated with the parsed data.
//   - An error if the parsing fails.
//
// Example:
//
//	var myData MyStruct
//	myData, err := conv.ParseJSONBytes[MyStruct](jsonBytes)
//	if err != nil {
//	    // handle error
//	}
func ParseJSONBytes[T any](data []byte) (T, error) {
	var result T
	if err := defaultConverter.FromJSONBytes(data, &result); err != nil {
		return result, err
	}
	return result, nil
}

// Clone creates a deep copy of the given value using JSON serialization and deserialization.
//
// Parameters:
//   - `v`: The value to be cloned.
//
// Returns:
//   - A deep copy of the input value.
//
// Example:
//
//	original := MyStruct{Field: "value"}
//	clone, err := conv.Clone(original)
//	if err != nil {
//	// handle error
//	}
func Clone[T any](v T) (T, error) {
	var result T

	data, err := json.Marshal(v)
	if err != nil {
		return result, newConvErrorf("cannot marshal for clone: %v", err)
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return result, newConvErrorf("cannot unmarshal for clone: %v", err)
	}

	return result, nil
}
