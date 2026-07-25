package fj

import (
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/polarixa/replify/pkg/common"
	"github.com/polarixa/replify/pkg/encoding"
	"github.com/polarixa/replify/pkg/strutil"
)

// Transformer is the interface implemented by all transformer types.
// A transformer receives the current JSON string and an optional argument string,
// and returns a transformed JSON string. Transformers are applied via the @ syntax
// in fj path expressions (e.g. "name.@uppercase").
//
// Implement this interface to provide custom, stateful, or composable transformer
// logic that goes beyond plain function closures. For simple, stateless cases use
// the TransformerFunc adapter, which satisfies this interface automatically.
//
// Example (struct-based):
//
//	type prefixTransformer struct{ prefix string }
//
//	func (t *prefixTransformer) Apply(json, arg string) string {
//	    return t.prefix + json
//	}
//
//	fj.AddTransformer("prefix", &prefixTransformer{prefix: "data:"})
type Transformer interface {
	Apply(json, arg string) string
}

// TransformerFunc is a function adapter that implements the Transformer interface.
// It allows any plain function with the signature func(json, arg string) string to
// satisfy Transformer without defining a named struct.
//
// This mirrors the http.HandlerFunc pattern from the standard library.
//
// Example:
//
//	fj.AddTransformer("upper", fj.TransformerFunc(func(json, arg string) string {
//	    return strings.ToUpper(json)
//	}))
type TransformerFunc func(json, arg string) string

// Apply calls f(json, arg), satisfying the Transformer interface.
func (f TransformerFunc) Apply(json, arg string) string {
	return f(json, arg)
}

// transformerRegistry holds the mapping from transformer name to Transformer.
// It guards concurrent access with a sync.RWMutex so that AddTransformer and path
// evaluation can be called safely from multiple goroutines.
type transformerRegistry struct {
	mu           sync.RWMutex
	transformers map[string]Transformer
}

// globalRegistry is the package-level singleton transformer registry.
// All built-in transformers are registered via init(). Custom transformers can be
// added at any time with AddTransformer.
var globalRegistry = &transformerRegistry{
	transformers: make(map[string]Transformer),
}

// Register adds or replaces a Transformer in the registry under the given name.
// It is safe for concurrent use by multiple goroutines.
func (r *transformerRegistry) Register(name string, t Transformer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.transformers[name] = t
}

// Get retrieves a Transformer by name.
// It returns (t, true) if found or (nil, false) if not registered.
// It is safe for concurrent use by multiple goroutines.
func (r *transformerRegistry) Get(name string) (Transformer, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.transformers[name]
	return t, ok
}

// IsRegistered reports whether a transformer with the given name has been registered.
// It is safe for concurrent use by multiple goroutines.
func (r *transformerRegistry) IsRegistered(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.transformers[name]
	return ok
}

// resolveTransformer looks up and returns the Transformer registered under the given
// name in the global registry. It returns nil when no transformer with that name has
// been registered, acting as the central dispatch point for all named transformer
// invocations.
func resolveTransformer(name string) Transformer {
	t, ok := globalRegistry.Get(name)
	if !ok {
		return nil
	}
	return t
}

// applyIdentity is a fallback transformation that simply returns the input JSON string
// without applying any modifications. This function is typically used as a default case
// when no specific transformation is requested or supported.
//
// Parameters:
//   - `json`: The JSON string to be returned as-is.
//   - `arg`: This parameter is unused for this transformation but is included for consistency
//     with other transform functions.
//
// Returns:
//   - The original input JSON string, unchanged.
//
// Example Usage:
//
//	// Input JSON
//	json := `{"name":"Alice","age":25}`
//
//	// No transformation applied, returns original JSON
//	unchangedJSON := applyIdentity(json, "")
//	fmt.Println(unchangedJSON)
//	// Output: {"name":"Alice","age":25}
//
// Notes:
//   - This function is used when no transformation is specified or when the transformation
//     request is unsupported. It ensures that the input JSON is returned unmodified.
func applyIdentity(json, arg string) string {
	return json
}

// applyPrettyFormat formats the input JSON string into a human-readable, indented format.
//
// This function applies "pretty printing" to the provided JSON data, making it easier to read
// and interpret. If additional formatting options are specified in the `arg` parameter, these
// options are parsed and applied to customize the output. Formatting options include sorting
// keys, setting indentation styles, specifying prefixes, and defining maximum line widths.
//
// Parameters:
//   - `json`: The JSON string to be formatted.
//   - `arg`: An optional string containing formatting configuration in JSON format. The configuration
//     can specify the following keys:
//   - `sortKeys`: A boolean value (`true` or `false`) that determines whether keys in JSON objects
//     should be sorted alphabetically.
//   - `indent`: A string containing whitespace characters (e.g., `"  "` or `"\t"`) used for indentation.
//   - `prefix`: A string prepended to each line of the formatted JSON.
//   - `width`: An integer specifying the maximum line width for the formatted output.
//
// Returns:
//   - A string representing the formatted JSON, transformed based on the specified or default options.
//
// Example Usage:
//
//	// Input JSON
//	json := `{"name":"Alice","age":25,"address":{"city":"New York","zip":"10001"}}`
//
//	// Format without additional options
//	prettyJSON := applyPrettyFormat(json, "")
//	fmt.Println(prettyJSON)
//	// Output:
//	// {
//	//   "name": "Alice",
//	//   "age": 25,
//	//   "address": {
//	//     "city": "New York",
//	//     "zip": "10001"
//	//   }
//	// }
//
//	// Format with additional options
//	arg := `{"indent": "\t", "sort_keys": true}`
//	prettyJSONWithOpts := applyPrettyFormat(json, arg)
//	fmt.Println(prettyJSONWithOpts)
//	// Output:
//	// {
//	// 	"address": {
//	// 		"city": "New York",
//	// 		"zip": "10001"
//	// 	},
//	// 	"age": 25,
//	// 	"name": "Alice"
//	// }
//
// Notes:
//   - If `arg` is empty, default formatting is applied with standard indentation.
//   - The function uses `encoding.Pretty` or `encoding.PrettyOptions` for the actual formatting.
//   - Invalid or unrecognized keys in the `arg` parameter are ignored.
//   - The function internally uses `fromStr2Bytes` and `fromBytes2Str` for efficient data conversion.
//
// Implementation Details:
//   - The `arg` string is parsed using the `Parse` function, and each key-value pair is applied
//     to configure the formatting options (`opts`).
//   - The `strutil.StripNonWhitespace` function ensures only whitespace characters are used for `indent`
//     and `prefix` settings to prevent formatting errors.
func applyPrettyFormat(json, arg string) string {
	if len(arg) > 0 {
		opts := *encoding.DefaultOptionsConfig
		Parse(arg).Foreach(func(key, value Context) bool {
			switch key.String() {
			case "sort_keys":
				opts.SortKeys = value.Bool()
			case "indent":
				opts.Indent = strutil.StripNonWhitespace(value.String())
			case "prefix":
				opts.Prefix = strutil.StripNonWhitespace(value.String())
			case "width":
				opts.Width = int(value.Int64())
			}
			return true
		})
		return strutil.SafeStr(encoding.PrettyOptions(UnsafeBytes(json), &opts))
	}
	return strutil.SafeStr(encoding.Pretty(UnsafeBytes(json)))
}

// applyMinify removes all whitespace characters from the input JSON string,
// transforming it into a compact, single-line format.
//
// This function applies a "minified" transformation to the provided JSON data,
// removing all spaces, newlines, and other whitespace characters. The result is
// a more compact representation of the JSON, which is useful for minimizing
// data size, especially when transmitting JSON data over the network or storing
// it in a compact format.
//
// Parameters:
//   - `json`: The JSON string to be transformed into a compact form.
//   - `arg`: This parameter is unused for this transformation but is included
//     for consistency with other transform functions.
//
// Returns:
//   - A string representing the "ugly" JSON, with all whitespace removed.
//
// Example Usage:
//
//	// Input JSON
//	json := `{
//	  "name": "Alice",
//	  "age": 25,
//	  "address": {
//	    "city": "New York",
//	    "zip": "10001"
//	  }
//	}`
//
//	// Transform to minify (compact) JSON
//	uglyJSON := applyMinify(json, "")
//	fmt.Println(uglyJSON)
//	// Output: {"name":"Alice","age":25,"address":{"city":"New York","zip":"10001"}}
//
// Notes:
//   - The `arg` parameter is not used in this transformation, and its value is ignored.
//   - The function uses `encoding.Ugly` for the actual transformation, which removes all
//     whitespace from the JSON data.
//   - This function is often used to reduce the size of JSON data for storage or transmission.
func applyMinify(json, arg string) string {
	return strutil.SafeStr(encoding.CompactJSON(UnsafeBytes(json)))
}

// applyReverse reverses the order of elements in an array or the order of key-value
// pairs in an object. This function processes the JSON input and applies the reversal
// based on the type of JSON structure: array or object.
//
// If the JSON is an array, it reverses the array elements. If it's an object, it reverses
// the key-value pairs. If the input is neither an array nor an object, the original JSON
// string is returned unchanged.
//
// Parameters:
//   - `json`: The JSON string to be transformed, which may be an array or an object.
//   - `arg`: This parameter is unused for this transformation but is included for consistency
//     with other transform functions.
//
// Returns:
//   - A string representing the transformed JSON with reversed elements or key-value pairs.
//
// Example Usage:
//
//	// Input JSON (array)
//	jsonArray := `[1, 2, 3]`
//
//	// Reverse array elements
//	reversedJSON := applyReverse(jsonArray, "")
//	fmt.Println(reversedJSON)
//	// Output: [3,2,1]
//
//	// Input JSON (object)
//	jsonObject := `{"name":"Alice","age":25}`
//
//	// Reverse key-value pairs
//	reversedObject := applyReverse(jsonObject, "")
//	fmt.Println(reversedObject)
//	// Output: {"age":25,"name":"Alice"}
//
// Notes:
//   - If the input JSON is an array, the array elements are reversed.
//   - If the input JSON is an object, the key-value pairs are reversed.
//   - If the input JSON is neither an array nor an object, the original string is returned unchanged.
func applyReverse(json, arg string) string {
	ctx := Parse(json)
	if ctx.IsArray() {
		var values []Context
		ctx.Foreach(func(_, value Context) bool {
			values = append(values, value)
			return true
		})
		out := make([]byte, 0, len(json))
		out = append(out, '[')
		for i, j := len(values)-1, 0; i >= 0; i, j = i-1, j+1 {
			if j > 0 {
				out = append(out, ',')
			}
			out = append(out, values[i].raw...)
		}
		out = append(out, ']')
		return strutil.SafeStr(out)
	}
	if ctx.IsObject() {
		var keyValues []Context
		ctx.Foreach(func(key, value Context) bool {
			keyValues = append(keyValues, key, value)
			return true
		})
		out := make([]byte, 0, len(json))
		out = append(out, '{')
		for i, j := len(keyValues)-2, 0; i >= 0; i, j = i-2, j+1 {
			if j > 0 {
				out = append(out, ',')
			}
			out = append(out, keyValues[i+0].raw...)
			out = append(out, ':')
			out = append(out, keyValues[i+1].raw...)
		}
		out = append(out, '}')
		return strutil.SafeStr(out)
	}
	return json
}

// applyFlatten flattens a JSON array by removing any nested arrays within it.
//
// This function takes a JSON array (which may contain nested arrays) and flattens it
// into a single array by extracting the elements of any child arrays. The function
// supports both shallow and deep flattening based on the provided argument.
//
// Parameters:
//   - `json`: A string representing the JSON array to be flattened. The array may contain
//     nested arrays that will be flattened into the outer array.
//   - `arg`: An optional string containing configuration options in JSON format. The configuration
//     can specify the following key:
//   - `deep`: A boolean value (`true` or `false`) that determines whether nested arrays should
//     be recursively flattened (deep flattening). If `deep` is `true`, all nested arrays are
//     flattened into the main array, while if `false` (or absent), only the immediate nested arrays
//     are flattened.
//
// Returns:
//   - A string representing the flattened JSON array. The returned array may contain elements
//     from nested arrays, depending on whether deep flattening was requested.
//
// Example Usage:
//
//	// Input JSON (shallow flatten)
//	json := "[1,[2],[3,4],[5,[6,7]]]"
//	shallowFlattened := applyFlatten(json, "")
//	fmt.Println(shallowFlattened)
//	// Output: [1,2,3,4,5,[6,7]]
//
//	// Input JSON (deep flatten)
//	json := "[1,[2],[3,4],[5,[6,7]]]"
//	deepFlattened := applyFlatten(json, "{\"deep\": true}")
//	fmt.Println(deepFlattened)
//	// Output: [1,2,3,4,5,6,7]
//
// Notes:
//
//   - If the input JSON is not an array, the original JSON string is returned unchanged.
//
//   - The function first checks if the provided JSON is an array. If it is not an array, it returns
//     the original input string without any changes.
//
//   - The `deep` option controls whether nested arrays are flattened recursively. If the `deep`
//     option is set to `false` (or omitted), only the immediate nested arrays are flattened.
//
//   - Nested arrays can be flattened either shallowly or deeply depending on the configuration provided
//     in the `arg` parameter.
//
//   - The function uses `removeOuterBraces` to remove the surrounding brackets of nested arrays to
//     achieve the flattening effect.
//
//     [1,[2],[3,4],[5,[6,7]]] -> [1,2,3,4,5,[6,7]]
//
// The {"deep":true} arg can be provide for deep flattening.
//
//	[1,[2],[3,4],[5,[6,7]]] -> [1,2,3,4,5,6,7]
//
// The original json is returned when the json is not an array.
func applyFlatten(json, arg string) string {
	ctx := Parse(json)
	if !ctx.IsArray() {
		return json
	}
	var deep bool
	if strutil.IsNotEmpty(arg) {
		Parse(arg).Foreach(func(key, value Context) bool {
			if key.String() == "deep" {
				deep = value.Bool()
			}
			return true
		})
	}
	var out []byte
	out = append(out, '[')
	var idx int
	ctx.Foreach(func(_, value Context) bool {
		var raw string
		if value.IsArray() {
			if deep {
				raw = trimOuterBrackets(applyFlatten(value.raw, arg))
			} else {
				raw = trimOuterBrackets(value.raw)
			}
		} else {
			raw = value.raw
		}
		raw = strings.TrimSpace(raw)
		if len(raw) > 0 {
			if idx > 0 {
				out = append(out, ',')
			}
			out = append(out, raw...)
			idx++
		}
		return true
	})
	out = append(out, ']')
	return strutil.SafeStr(out)
}

// applyMerge merges multiple JSON objects into a single object.
// If the input is an array of JSON objects, it combines their key-value pairs
// into one object. Duplicate keys can be preserved or discarded based on the
// configuration provided in the `arg` parameter.
//
// Parameters:
//   - `json`: A string representing a JSON array, where each element is a JSON object.
//     The objects will be merged into a single object.
//   - `arg`: A string containing a JSON configuration that can specify whether
//     duplicate keys should be preserved. If `arg` is provided and contains
//     the key `preserve` set to `true`, duplicate keys will be kept in the output object.
//
// Returns:
//   - A string representing the merged JSON object. If the input is not an array
//     of JSON objects, the function returns the original `json` string unchanged.
//
// Example Usage:
//
//	// Input JSON (merge objects with duplicate keys discarded)
//	json := `[{"first":"Tom","age":37},{"age":41}]`
//	mergedJSON := applyMerge(json, "")
//	fmt.Println(mergedJSON)
//	// Output: {"first":"Tom","age":41}
//
//	// Input JSON (merge objects with duplicate keys preserved)
//	json := `[{"first":"Tom","age":37},{"age":41}]`
//	mergedJSONWithDupes := applyMerge(json, "{\"preserve\": true}")
//	fmt.Println(mergedJSONWithDupes)
//	// Output: {"first":"Tom","age":37,"age":41}
//
// Notes:
//   - If the input `json` is not a valid array of JSON objects, the function returns
//     the original input string unchanged.
//   - The `preserve` option controls whether duplicate keys should be kept in the merged object.
//     If `preserve` is `false` (or absent), only the last occurrence of each key is kept.
//   - The function uses `removeOuterBraces` to remove any extraneous brackets around JSON objects
//     before merging their contents.
//
// Implementation Details:
//   - If the `preserve` option is set to `true`, all key-value pairs from the objects are
//     appended to the resulting object, even if keys are repeated.
//   - If `preserve` is `false`, the function will deduplicate keys by selecting the last
//     value for each key across all objects in the array. The keys are also added in stable
//     order based on their appearance in the input objects.
func applyMerge(json, arg string) string {
	ctx := Parse(json)
	if !ctx.IsArray() {
		return json
	}
	var preserve bool
	if strutil.IsNotEmpty(arg) {
		Parse(arg).Foreach(func(key, value Context) bool {
			if key.String() == "preserve" {
				preserve = value.Bool()
			}
			return true
		})
	}
	var target []byte
	target = append(target, '{')
	if preserve { // preserve duplicate keys.
		var idx int
		ctx.Foreach(func(_, value Context) bool {
			if !value.IsObject() {
				return true
			}
			if idx > 0 {
				target = append(target, ',')
			}
			target = append(target, trimOuterBrackets(value.raw)...)
			idx++
			return true
		})
	} else { // deduplicate keys and generate an object with stable ordering.
		var keys []Context
		keyVal := make(map[string]Context)
		ctx.Foreach(func(_, value Context) bool {
			if !value.IsObject() {
				return true
			}
			value.Foreach(func(key, value Context) bool {
				k := key.String()
				if _, ok := keyVal[k]; !ok {
					keys = append(keys, key)
				}
				keyVal[k] = value
				return true
			})
			return true
		})
		for i := 0; i < len(keys); i++ {
			if i > 0 {
				target = append(target, ',')
			}
			target = append(target, keys[i].raw...)
			target = append(target, ':')
			target = append(target, keyVal[keys[i].String()].raw...)
		}
	}
	target = append(target, '}')
	return strutil.SafeStr(target)
}

// applyValidityCheck reports whether the input is valid JSON.
// It returns the string "true" if valid, or "false" otherwise. Use this transformer
// in a path expression such as `input.@valid` to gate subsequent processing.
//
// Example:
//
//	fj.Get(`{"a":1}`, `@valid`) → "true"
//	fj.Get(`{bad}`,   `@valid`) → "false"
func applyValidityCheck(json, arg string) string {
	if !IsValidJSONString(json) {
		return "false"
	}
	return "true"
}

// applyKeys extracts the keys from a JSON object and returns them as a JSON array of strings.
// The function processes the input JSON, identifies whether it is an object, and then generates
// an array containing the keys of the object. If the input is not a valid JSON object, it returns
// an empty array.
//
// Parameters:
//   - `json`: A string representing the JSON data, which should be an object from which keys will be extracted.
//   - `arg`: This parameter is not used in this function but is included for consistency with other transformation functions.
//
// Returns:
//   - A string representing a JSON array of keys, or an empty array (`[]`) if the input is not a valid object.
//
// Example Usage:
//
//	// Input JSON (object)
//	json := `{"first":"Tom","last":"Smith"}`
//	keys := applyKeys(json, "")
//	fmt.Println(keys)
//	// Output: ["first","last"]
//
//	// Input JSON (non-object)
//	json := `"Tom"`
//	keys := applyKeys(json, "")
//	fmt.Println(keys)
//	// Output: []
//
// Notes:
//   - If the input JSON is an object, the function will iterate through the keys of the object and return them in
//     a JSON array format.
//   - If the input JSON is not an object (e.g., an array, string, or invalid), the function will return an empty array (`[]`).
//   - The function relies on the `Parse` function to parse the input JSON and the `Foreach` method to iterate over
//     the object keys.
//   - The `raw` method is used to extract the raw key value as a string without further processing.
//
// Implementation Details:
//   - The function first checks if the parsed JSON object exists. If it does, it iterates through the object and extracts
//     the keys. Each key is added to a string builder, and the keys are wrapped in square brackets to form a valid JSON array.
//   - If the JSON is not an object, the function immediately returns an empty array (`[]`).
func applyKeys(json, arg string) string {
	ctx := Parse(json)
	if !ctx.Exists() {
		return "[]"
	}
	var i int
	var builder strings.Builder
	o := ctx.IsObject()
	builder.WriteByte('[')
	ctx.Foreach(func(key, _ Context) bool {
		if i > 0 {
			builder.WriteByte(',')
		}
		if o {
			builder.WriteString(key.raw)
		} else {
			builder.WriteString("null")
		}
		i++
		return true
	})
	builder.WriteByte(']')
	return builder.String()
}

// applyValues extracts the values from a JSON object and returns them as a JSON array of values.
//
// This function parses the input JSON string, and if the JSON is an object, it extracts all the values
// from the key-value pairs and returns them as a JSON array of values. If the input JSON is already an array,
// it simply returns the original JSON string. If the input does not contain a valid JSON object or array,
// it returns an empty array ("[]").
//
// Parameters:
//   - `json`: The JSON string to extract values from. It can be a JSON object or array.
//   - `arg`: An optional argument that is not used in this function, but can be included for consistency
//     with other transformation functions.
//
// Returns:
//   - A string representing a JSON array containing the values extracted from the input JSON object.
//     If the input JSON is already an array, it is returned as-is. If the input is invalid or empty,
//     an empty array ("[]") is returned.
//
// Example Usage:
//
//	// Input JSON representing an object
//	json := `{"first":"Aris","last":"Nguyen"}`
//
//	// Extract the values from the object
//	values := applyValues(json, "")
//	fmt.Println(values) // Output: ["Aris","Nguyen"]
//
//	// Input JSON representing an array
//	jsonArray := `["apple", "banana", "cherry"]`
//
//	// Return the array as-is
//	values := applyValues(jsonArray, "")
//	fmt.Println(values) // Output: ["apple", "banana", "cherry"]
//
//	// Input JSON representing an invalid object
//	invalidJson := `{"key":}` // Invalid JSON
//
//	// Return empty array for invalid JSON
//	values := applyValues(invalidJson, "")
//	fmt.Println(values) // Output: []
//
// Details:
//   - The function first parses the input JSON string using `Parse`.
//   - If the input is an array, the function directly returns the original string as it is.
//   - If the input is an object, the function iterates over its key-value pairs, extracting only the values,
//     and then constructs a JSON array of these values.
//   - If the input JSON does not exist or is invalid, the function returns an empty JSON array ("[]").
func applyValues(json, arg string) string {
	ctx := Parse(json)
	if !ctx.Exists() {
		return "[]"
	}
	if ctx.IsArray() {
		return json
	}
	var i int
	var builder strings.Builder
	builder.WriteByte('[')
	ctx.Foreach(func(_, value Context) bool {
		if i > 0 {
			builder.WriteByte(',')
		}
		builder.WriteString(value.raw)
		i++
		return true
	})
	builder.WriteByte(']')
	return builder.String()
}

// applyToJSON converts a string into a valid JSON representation.
//
// This function ensures that the input string is a valid JSON before attempting to
// parse and convert it into its corresponding JSON format. If the input string is
// a valid JSON, the function returns the formatted JSON as a string. Otherwise,
// it returns an empty string to indicate that the input was not valid JSON.
//
// Parameters:
//   - `json`: A string representing the data that needs to be converted to a valid JSON format.
//   - `arg`: An additional argument which is unused in this function. It may be a placeholder for future extensions.
//
// Returns:
//   - A string representing the input data in valid JSON format. If the input string is not valid JSON,
//     an empty string is returned.
//
// Example Usage:
//
//	// Input string
//	json := "{\"id\":1023,\"name\":\"alert\"}"
//
//	// Convert to valid JSON representation
//	result := applyToJSON(json, "")
//	fmt.Println(result)  // Output: {"id":1023,"name":"alert"}
//
//	// Invalid input string
//	invalidJson := "\"id\":1023,\"name\":\"alert\""
//	result = applyToJSON(invalidJson, "")
//	fmt.Println(result)  // Output: ""
//
// Notes:
//   - This function uses the `IsValidJSON` helper to check if the input string is a valid JSON format.
//   - If the input string is valid JSON, the `Parse` function is used to parse and format it, ensuring it is returned in the proper JSON format.
//   - If the input is invalid, an empty string is returned, indicating that the transformation failed.
func applyToJSON(json, arg string) string {
	if !IsValidJSONString(json) {
		return ""
	}
	return Parse(json).String()
}

// applyToString converts a regular string into a valid JSON string format.
//
// This function takes an input string and converts it into a JSON-encoded string
// by wrapping it in double quotes and escaping any necessary characters (such as
// quotes or backslashes) to ensure the resulting string adheres to JSON encoding rules.
//
// Parameters:
//   - `str`: The input string to be converted into JSON string format.
//   - `arg`: This parameter is not used in this function but is present for consistency
//     with the function signature, which might be useful for future extensions or
//     compatibility with other transformations.
//
// Returns:
//   - A string that is a valid JSON representation of the input `str`, with special
//     characters properly escaped (e.g., double quotes, backslashes, control characters).
//
// Example Usage:
//
//	str := "Hello \"world\"\nLine break!"
//	result := applyToString(str, "")
//	fmt.Println(result)
//	// Output: "\"Hello \\\"world\\\"\nLine break!\""
//
// Notes:
//   - This function calls `appendJSON` to handle the conversion of the string into a
//     valid JSON string format, ensuring proper escaping of special characters (like
//     double quotes and newlines) to maintain valid JSON syntax.
//   - The `arg` parameter is included for consistency with other transformation functions
//     that may require it, though it does not affect the behavior of this specific function.
func applyToString(str, arg string) string {
	return string(appendJSONString(nil, str))
}

// applyGroup processes a JSON string containing objects and arrays, and groups the
// elements of arrays within objects by their keys. It converts each array into a group of
// key-value pairs, resulting in a new JSON structure where each object is grouped by its array values.
//
// This function is primarily used to reformat JSON data by grouping values in arrays under
// the same key into new JSON objects, preserving the structure of the original JSON while
// transforming the array elements into key-value pairs associated with the respective key.
//
// Parameters:
//   - `json`: A string representing the JSON data that needs to be transformed. It is assumed
//     to be in the form of a JSON object that contains arrays.
//   - `arg`: A string argument that provides additional options for the transformation. In this
//     case, the argument is not used but is still passed for consistency with the other transformation
//     functions.
//
// Returns:
//   - A new string representing the transformed JSON data. The arrays in the input JSON are grouped
//     under their respective keys, and the resulting structure is a list of objects.
//
// Example Usage:
//
//	json := `{"a": [1, 2, 3], "b": [4, 5]}`
//	result := applyGroup(json, "")
//	fmt.Println(result)
//	// Output: `[
//	//  {"a":1},
//	//  {"a":2},
//	//  {"a":3},
//	//  {"b":4},
//	//  {"b":5}
//	// ]`
//
// Notes:
//   - The function works by iterating over each key-value pair in the input JSON object. If the value
//     associated with a key is an array, it processes the array's elements by creating new objects where
//     the key is used for each element, effectively turning each element into a separate object with its
//     corresponding key.
//   - If the input JSON does not contain arrays, it does not affect the final result.
//
// Implementation Details:
//   - The function first parses the input JSON string into a context object using `Parse(json)`.
//   - It checks if the context is a valid object (`ctx.IsObject()`). If not, it returns an empty string.
//   - It iterates over the object's key-value pairs using `ctx.Foreach()`. For each array value, it creates
//     new objects where each array element is associated with the respective key.
//   - The transformed groups of key-value pairs are collected into a new array format, where each array
//     element corresponds to a new object created from the array values.
func applyGroup(json, arg string) string {
	ctx := Parse(json)
	if !ctx.IsObject() {
		return ""
	}
	var all [][]byte
	ctx.Foreach(func(key, value Context) bool {
		if !value.IsArray() {
			return true
		}
		var idx int
		value.Foreach(func(_, value Context) bool {
			if idx == len(all) {
				all = append(all, []byte{})
			}
			all[idx] = append(all[idx], ("," + key.raw + ":" + value.raw)...)
			idx++
			return true
		})
		return true
	})
	var data []byte
	data = append(data, '[')
	for i, el := range all {
		if i > 0 {
			data = append(data, ',')
		}
		data = append(data, '{')
		data = append(data, el[1:]...)
		data = append(data, '}')
	}
	data = append(data, ']')
	return string(data)
}

// applySearch performs a value lookup on a JSON structure based on the specified path
// and returns a JSON-encoded string containing all matching values found at that path.
//
// This function searches recursively through the JSON structure to find all occurrences
// of the specified path and then aggregates the results into a new JSON array. The results
// are returned as a string, which represents the matched values in their original, unprocessed form.
//
// Parameters:
//   - `json`: A string representing the input JSON data. The function will parse this JSON
//     and search for the specified path within the structure.
//   - `arg`: A string representing the JSON path to search for. The path is used to navigate
//     the nested JSON structure and retrieve values that match the specified key(s).
//
// Returns:
//   - A string representing a JSON array containing all matching values found in the input JSON
//     data. The values are presented in the same order they appear in the original JSON structure,
//     and are enclosed within square brackets ([]). If no matches are found, an empty array is returned.
//
// Example Usage:
//
//	json := `{
//	  "store": {
//	    "book": [
//	      { "category": "fiction", "author": "J.K. Rowling", "title": "Harry Potter" },
//	      { "category": "science", "author": "Stephen Hawking", "title": "A Brief History of Time" }
//	    ],
//	    "music": [
//	      { "artist": "The Beatles", "album": "Abbey Road" },
//	      { "artist": "Pink Floyd", "album": "The Wall" }
//	    ]
//	  }
//	}`
//
//	arg := "book.author"
//	result := applySearch(json, arg)
//
//	// Output: `["J.K. Rowling", "Stephen Hawking"]`
//	// The function will search for the "book.author" path and return all matching author names
//	// found in the nested `book` array.
//
// Notes:
//   - The `deepSearchRecursively` function is used to traverse the JSON structure and find all
//     matches for the specified path, ensuring that nested objects and arrays are searched as well.
//   - The results are accumulated in a slice, which is then converted into a JSON array and returned
//     as a string representation.
//   - If the path doesn't match any elements in the JSON structure, an empty array is returned.
//
// Implementation Details:
//   - The function utilizes `deepSearchRecursively` to perform a depth-first traversal of the JSON
//     structure and collect all matching values along the specified path.
//   - The results are then appended to a byte slice (`seg`), which is later converted into a string.
//   - The final output is a JSON array, even if no results are found.
func applySearch(json, arg string) string {
	all := recurseCollectMatches(nil, Parse(json), arg)
	var seg []byte
	seg = append(seg, '[')
	for i, res := range all {
		if i > 0 {
			seg = append(seg, ',')
		}
		seg = append(seg, res.raw...)
	}
	seg = append(seg, ']')
	return string(seg)
}

// applyUppercase converts the input JSON string to uppercase.
//
// This function takes a JSON string as input and converts all of its characters
// to uppercase. If the input string is empty, it returns the string unchanged.
//
// Parameters:
//   - `json`: The JSON string to be converted to uppercase.
//   - `arg`: An optional string parameter that is currently unused in this function,
//     but it could be extended for future use to modify the behavior of the transformation.
//
// Returns:
//   - A string with all characters converted to uppercase. If the input `json`
//     is empty, it returns the input string unchanged.
//
// Example Usage:
//
//	json := "{\"name\":\"Alice\",\"age\":25}"
//	result := applyUppercase(json, "")
//	fmt.Println(result) // Output: "{\"NAME\":\"ALICE\",\"AGE\":25}"
//
// Notes:
//   - This function uses the standard Go `strings.ToUpper` method to convert the string
//     to uppercase, which applies the transformation to every character in the string.
func applyUppercase(json, arg string) string {
	if strutil.IsEmpty(json) {
		return json
	}
	return strings.ToUpper(json)
}

// applyLowercase converts the input JSON string to lowercase.
//
// This function takes a JSON string as input and converts all of its characters
// to lowercase. If the input string is empty, it returns the string unchanged.
//
// Parameters:
//   - `json`: The JSON string to be converted to lowercase.
//   - `arg`: An optional string parameter that is currently unused in this function,
//     but it could be extended for future use to modify the behavior of the transformation.
//
// Returns:
//   - A string with all characters converted to lowercase. If the input `json`
//     is empty, it returns the input string unchanged.
//
// Example Usage:
//
//	json := "{\"name\":\"Alice\",\"age\":25}"
//	result := applyLowercase(json, "")
//	fmt.Println(result) // Output: "{\"name\":\"alice\",\"age\":25}"
//
// Notes:
//   - This function uses the standard Go `strings.ToLower` method to convert the string
//     to lowercase, which applies the transformation to every character in the string.
func applyLowercase(json, arg string) string {
	if strutil.IsEmpty(json) {
		return json
	}
	return strings.ToLower(json)
}

// applyFlip reverses the input JSON string.
//
// This function takes the input string and reverses the order of its characters.
// It returns the reversed string. If the input string is empty, it returns the
// string unchanged.
//
// Parameters:
//   - `json`: The JSON string to be reversed.
//   - `arg`: An optional argument that is currently unused, but could be extended
//     for future transformations.
//
// Returns:
//   - A string with its characters reversed. If the input is empty, the original
//     string is returned unchanged.
//
// Example Usage:
//
//	json := "{\"name\":\"Alice\",\"age\":25}"
//	result := applyFlip(json, "")
//	fmt.Println(result) // Output: "}52ega,\"ecilA\":\"emam{"
func applyFlip(json, arg string) string {
	if strutil.IsEmpty(json) {
		return json
	}
	runes := []rune(json)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// applyTrim removes leading and trailing whitespace from the input JSON string.
//
// This function removes any whitespace characters at the beginning and end of the
// input string. If there is no whitespace, the string remains unchanged.
//
// Parameters:
//   - `json`: The JSON string to be trimmed of whitespace.
//   - `arg`: An optional string argument that is currently unused, but could be
//     extended for future use.
//
// Returns:
//   - A string with leading and trailing whitespace removed. If the input string
//     does not contain whitespace at the edges, it returns the original string.
//
// Example Usage:
//
//	json := "   {\"name\":\"Alice\"}   "
//	result := applyTrim(json, "")
//	fmt.Println(result) // Output: "{\"name\":\"Alice\"}"
//
// Notes:
//   - This function uses Go's `strings.TrimSpace` method to remove whitespace characters.
func applyTrim(json, arg string) string {
	return strutil.TrimWhitespace(strutil.Trim(json))
}

// applySnakeCase converts the input string to snake_case format, which is typically used
// for variable names in many programming languages. The string is transformed to lowercase,
// and spaces or other delimiters are replaced with underscores ('_').
//
// Parameters:
//   - `json`: The input string that will be converted to snake_case.
//   - `arg`: An optional string argument that is currently unused.
//
// Returns:
//   - A string formatted in snake_case. If the input is empty, it returns unchanged.
//
// Example Usage:
//
//	json := "{\"First Name\":\"Alice\",\"Last Name\":\"Smith\"}"
//	result := applySnakeCase(json, "")
//	fmt.Println(result) // Output: "{\"first_name\":\"alice\",\"last_name\":\"smith\"}"
func applySnakeCase(json, arg string) string {
	if strutil.IsEmpty(json) {
		return json
	}
	// Replace spaces with underscores, convert to lowercase.
	json = strings.ReplaceAll(json, " ", "_")
	return strings.ToLower(json)
}

// applyCamelCase converts the input string into camelCase, which is often used for
// variable and function names in JavaScript and other programming languages. The string
// is converted to lowercase with spaces removed, and the first letter of each word after
// the first is capitalized.
//
// Parameters:
//   - `json`: The input string that will be converted to camelCase.
//   - `arg`: An optional argument that is currently unused.
//
// Returns:
//   - A string formatted in camelCase. If the input is empty, it returns unchanged.
//
// Example Usage:
//
//	json := "{\"first name\":\"alice\",\"last name\":\"smith\"}"
//	result := applyCamelCase(json, "")
//	fmt.Println(result) // Output: "{\"firstName\":\"alice\",\"lastName\":\"smith\"}"
func applyCamelCase(json, arg string) string {
	if strutil.IsEmpty(json) {
		return json
	}
	words := strings.Fields(json)
	for i := 1; i < len(words); i++ {
		words[i] = strings.ToUpper(words[i][:1]) + words[i][1:]
	}
	return strings.Join(words, "")
}

// applyKebabCase converts the input string into kebab-case, often used for URL slugs.
// The string is transformed to lowercase, and spaces are replaced with hyphens ('-').
//
// Parameters:
//   - `json`: The input string that will be converted to kebab-case.
//   - `arg`: An optional string argument that is currently unused.
//
// Returns:
//   - A string formatted in kebab-case. If the input is empty, it returns unchanged.
//
// Example Usage:
//
//	json := "{\"First Name\":\"Alice\",\"Last Name\":\"Smith\"}"
//	result := applyKebabCase(json, "")
//	fmt.Println(result) // Output: "{\"first-name\":\"alice\",\"last-name\":\"smith\"}"
func applyKebabCase(json, arg string) string {
	if strutil.IsEmpty(json) {
		return json
	}
	// Replace spaces with hyphens, convert to lowercase.
	json = strings.ReplaceAll(json, " ", "-")
	return strings.ToLower(json)
}

// applyReplace replaces a specific substring within the input string with another string.
//
// This function can be used to replace any substring in the input string with a given replacement.
//
// Parameters:
//   - `json`: The input string in which the replacement will occur.
//   - `arg`: A JSON string containing the target and replacement substrings.
//     Example: `{"target": "apple", "replacement": "orange"}`.
//
// Returns:
//   - The string with the target substring replaced. If the input is empty, it returns unchanged.
//
// Example Usage:
//
//	json := "I love apple pie"
//	arg := "{\"target\":\"apple\",\"replacement\":\"orange\"}"
//	result := applyReplace(json, arg)
//	fmt.Println(result) // Output: "I love orange pie"
func applyReplace(json, arg string) string {
	if strutil.IsEmpty(json) {
		return json
	}
	var target, replacement string
	Parse(arg).Foreach(func(key, value Context) bool {
		if key.String() == "target" {
			target = value.String()
		}
		if key.String() == "replacement" {
			replacement = value.String()
		}
		return true
	})
	return strings.Replace(json, target, replacement, 1)
}

// applyReplaceAll replaces all occurrences of a target substring with a replacement string.
//
// This function performs a global replacement in the input string, replacing every occurrence
// of the target substring with the specified replacement.
//
// Parameters:
//   - `json`: The input string in which the replacement will be performed.
//   - `arg`: A JSON string containing the target and replacement substrings.
//     Example: `{"target": "foo", "replacement": "bar"}`.
//
// Returns:
//   - The string with all target substrings replaced by the replacement. If the input is empty,
//     it returns unchanged.
//
// Example Usage:
//
//	json := "foo bar foo"
//	arg := "{\"target\":\"foo\",\"replacement\":\"baz\"}"
//	result := applyReplaceAll(json, arg)
//	fmt.Println(result) // Output: "baz bar baz"
func applyReplaceAll(json, arg string) string {
	if strutil.IsEmpty(json) {
		return json
	}
	var target, replacement string
	Parse(arg).Foreach(func(key, value Context) bool {
		if key.String() == "target" {
			target = value.String()
		}
		if key.String() == "replacement" {
			replacement = value.String()
		}
		return true
	})
	return strings.ReplaceAll(json, target, replacement)
}

// applyToHex converts the string to its hexadecimal representation.
//
// This function converts each character of the input string to its hexadecimal ASCII value.
//
// Parameters:
//   - `json`: The input string to be converted to hexadecimal.
//   - `arg`: An optional argument that is currently unused.
//
// Returns:
//   - A string containing the hexadecimal representation of the input string.
//
// Example Usage:
//
//	json := "hello"
//	result := applyToHex(json, "")
//	fmt.Println(result) // Output: "68656c6c6f"
func applyToHex(json, arg string) string {
	if strutil.IsEmpty(json) {
		return json
	}
	ctx := Parse(json)
	if !ctx.IsArray() && !ctx.IsObject() && common.IsScalarType(ctx.String()) {
		return fmt.Sprintf("%x", strutil.RemoveDoubleQuotes(json))
	}
	return fmt.Sprintf("%x", json)
}

// applyToBinary converts the string to its binary representation.
//
// This function converts each character of the input string to its binary ASCII value.
//
// Parameters:
//   - `json`: The input string to be converted to binary.
//   - `arg`: An optional argument that is currently unused.
//
// Returns:
//   - A string containing the binary representation of the input string.
//
// Example Usage:
//
//	json := "hello"
//	result := applyToBinary(json, "")
//	fmt.Println(result) // Output: "11010001101101110110011011001101111"
func applyToBinary(json, arg string) string {
	ctx := Parse(json)
	if !ctx.IsArray() && !ctx.IsObject() && common.IsScalarType(ctx.String()) {
		var bin string
		for _, r := range ctx.String() {
			bin += fmt.Sprintf("%08b", r)
		}
		return bin
	}
	var bin string
	for _, r := range json {
		bin += fmt.Sprintf("%08b", r)
	}
	return bin
}

// applyInsertAt inserts a specified string at a given index in the input string.
//
// This function allows you to insert a string at a particular position in the original string.
//
// Parameters:
//   - `json`: The original string where the insertion will happen.
//   - `arg`: A JSON string containing the position (`index`) and the string to insert (`insert`).
//     Example: `{"index": 5, "insert": "XYZ"}`.
//
// Returns:
//   - A new string with the specified insertion. If the input is empty, it returns unchanged.
//
// Example Usage:
//
//	json := "HelloWorld"
//	arg := "{\"index\":5,\"insert\":\"XYZ\"}"
//	result := applyInsertAt(json, arg)
//	fmt.Println(result) // Output: "HelloXYZWorld"
func applyInsertAt(json, arg string) string {
	if strutil.IsEmpty(json) {
		return json
	}
	var index int
	var insert string
	Parse(arg).Foreach(func(key, value Context) bool {
		if key.String() == "index" {
			index = int(value.Number())
		}
		if key.String() == "insert" {
			insert = value.String()
		}
		return true
	})
	if index < 0 || index > len(json) {
		return json // If index is out of bounds, return the original string.
	}
	return json[:index] + insert + json[index:]
}

// applyCountWords counts the number of words in the input string.
//
// This function splits the string into words (by spaces) and returns the count of the words.
//
// Parameters:
//   - `json`: The input string to count the words in.
//   - `arg`: An optional string argument that is currently unused.
//
// Returns:
//   - An integer representing the number of words in the input string. If the string is empty, it returns 0.
//
// Example Usage:
//
//	json := "Hello world"
//	result := applyCountWords(json, "")
//	fmt.Println(result) // Output: 2
func applyCountWords(json, arg string) string {
	ctx := Parse(json)
	json = strutil.TrimWhitespace(ctx.String())
	if strutil.IsEmpty(json) || strutil.IsBlank(json) || strutil.IsWhitespace(json) {
		return json
	}
	words := strings.Fields(json)
	return fmt.Sprintf("%v", len(words))
}

// applyPadLeft pads the input string with a specified character on the left to a given length.
//
// This function adds padding to the left of the string until the string reaches the specified length.
//
// Parameters:
//   - `json`: The input string to be padded.
//   - `arg`: A JSON string containing the padding character and the desired length. Example: `{"padding": "*", "length": 10}`.
//
// Returns:
//   - A new string with the specified padding added on the left. If the string is already the desired length, it returns unchanged.
//
// Example Usage:
//
//	json := "Hello"
//	arg := "{\"padding\": \"*\", \"length\": 10}"
//	result := applyPadLeft(json, arg)
//	fmt.Println(result) // Output: "*****Hello"
func applyPadLeft(json, arg string) string {
	ctx := Parse(json)
	value := ctx.String()
	var padding string
	var length int
	Parse(arg).Foreach(func(key, value Context) bool {
		if key.String() == "padding" {
			padding = value.String()
		}
		if key.String() == "length" {
			length = int(value.Int64())
		}
		return true
	})
	value = strutil.TrimWhitespace(strutil.Trim(value))
	t := encoding.JSON(value)
	if length <= len(t) {
		return t
	}
	return strings.Repeat(padding, length-len(t)) + t
}

// applyPadRight pads the input string with a specified character on the right to a given length.
//
// This function adds padding to the right of the string until the string reaches the specified length.
//
// Parameters:
//   - `json`: The input string to be padded.
//   - `arg`: A JSON string containing the padding character and the desired length. Example: `{"padding": "*", "length": 10}`.
//
// Returns:
//   - A new string with the specified padding added on the right. If the string is already the desired length, it returns unchanged.
//
// Example Usage:
//
//	json := "Hello"
//	arg := "{\"padding\": \"*\", \"length\": 10}"
//	result := applyPadRight(json, arg)
//	fmt.Println(result) // Output: "Hello*****"
func applyPadRight(json, arg string) string {
	ctx := Parse(json)
	value := ctx.String()
	var padding string
	var length int
	Parse(arg).Foreach(func(key, value Context) bool {
		if key.String() == "padding" {
			padding = value.String()
		}
		if key.String() == "length" {
			length = int(value.Int64())
		}
		return true
	})
	value = strutil.TrimWhitespace(strutil.Trim(value))
	t := encoding.JSON(value)
	if length <= len(t) {
		return t
	}
	return t + strings.Repeat(padding, length-len(t))
}

// applyProject returns a JSON object containing only the fields named in the `pick`
// list, optionally renaming them using the `rename` map.
//
// The `arg` must be a JSON object that may contain:
//   - `"pick"`: a JSON array of field names to include (all fields kept when absent).
//   - `"rename"`: a JSON object whose keys are original field names and whose values
//     are the desired output names.
//
// If the input is not a JSON object the input is returned unchanged.
//
// Example:
//
//	// projection only
//	fj.Get(`{"name":"Alice","age":30,"city":"NY"}`, `@project:{"pick":["name","age"]}`)
//	// → {"name":"Alice","age":30}
//
//	// rename only
//	fj.Get(`{"name":"Alice","age":30}`, `@project:{"rename":{"name":"fullName"}}`)
//	// → {"fullName":"Alice","age":30}
//
//	// projection + rename combined
//	fj.Get(`{"name":"Alice","age":30,"city":"NY"}`, `@project:{"pick":["name","age"],"rename":{"name":"fullName","age":"years"}}`)
//	// → {"fullName":"Alice","years":30}
//
// Performance: O(fields) with a single Foreach pass; builds output into a pre-grown
// byte slice, so only one allocation per call.
func applyProject(json, arg string) string {
	ctx := Parse(json)
	if !ctx.IsObject() {
		return json
	}
	var pick []string
	rename := make(map[string]string)
	if strutil.IsNotEmpty(arg) {
		Parse(arg).Foreach(func(key, value Context) bool {
			switch key.String() {
			case "pick":
				value.Foreach(func(_, v Context) bool {
					pick = append(pick, v.String())
					return true
				})
			case "rename":
				value.Foreach(func(k, v Context) bool {
					rename[k.String()] = v.String()
					return true
				})
			}
			return true
		})
	}
	pickSet := make(map[string]bool, len(pick))
	for _, f := range pick {
		pickSet[f] = true
	}
	out := make([]byte, 0, len(json))
	out = append(out, '{')
	var i int
	ctx.Foreach(func(key, value Context) bool {
		fieldName := key.String()
		if len(pickSet) > 0 && !pickSet[fieldName] {
			return true
		}
		outputName := fieldName
		if newName, ok := rename[fieldName]; ok {
			outputName = newName
		}
		if i > 0 {
			out = append(out, ',')
		}
		out = appendJSONString(out, outputName)
		out = append(out, ':')
		out = append(out, value.raw...)
		i++
		return true
	})
	out = append(out, '}')
	return strutil.SafeStr(out)
}

// applyFilter removes elements from a JSON array that do not satisfy a condition.
//
// The `arg` must be a JSON object with:
//   - `"key"` (required): the field name to test on each array element.
//   - `"value"` (required): the value to compare against.
//   - `"op"` (optional): comparison operator, one of:
//     `"eq"` (default), `"ne"`, `"gt"`, `"gte"`, `"lt"`, `"lte"`, `"contains"`.
//
// Elements that are not JSON objects are kept unchanged when the op cannot be
// evaluated (fail-open).  If the input is not a JSON array, it is returned
// unchanged.
//
// Example:
//
//	fj.Get(`[{"name":"Alice","age":30},{"name":"Bob","age":25}]`,
//	        `@filter:{"key":"age","op":"gt","value":28}`)
//	// → [{"name":"Alice","age":30}]
//
//	fj.Get(`[{"status":"active"},{"status":"inactive"}]`,
//	        `@filter:{"key":"status","value":"active"}`)
//	// → [{"status":"active"}]
//
// Performance: single Foreach pass; uses a pre-grown byte slice.
func applyFilter(json, arg string) string {
	ctx := Parse(json)
	if !ctx.IsArray() {
		return json
	}
	var key, op, rawVal string
	op = "eq"
	Parse(arg).Foreach(func(k, v Context) bool {
		switch k.String() {
		case "key":
			key = v.String()
		case "op":
			op = v.String()
		case "value":
			rawVal = v.raw
		}
		return true
	})
	if strutil.IsEmpty(key) {
		return json
	}
	cmpVal := Parse(rawVal)
	out := make([]byte, 0, len(json))
	out = append(out, '[')
	var i int
	ctx.Foreach(func(_, elem Context) bool {
		fieldVal := elem.Get(key)
		if !fieldVal.Exists() {
			return true
		}
		if matchesCondition(fieldVal, cmpVal, op) {
			if i > 0 {
				out = append(out, ',')
			}
			out = append(out, elem.raw...)
			i++
		}
		return true
	})
	out = append(out, ']')
	return strutil.SafeStr(out)
}

// applyDefault injects fallback values for fields that are absent or explicitly null
// in a JSON object.
//
// The `arg` must be a JSON object mapping field names to their default values.
// Fields that already exist with a non-null value are left untouched.
// Fields listed in `arg` that are missing from the input object are appended.
// If the input is not a JSON object, it is returned unchanged.
//
// Example:
//
//	fj.Get(`{"name":"Alice","role":null}`,
//	        `@default:{"role":"user","active":true}`)
//	// → {"name":"Alice","role":"user","active":true}
//
// Performance: two Foreach passes (one to collect existing, one to append defaults);
// single allocation for the output byte slice.
func applyDefault(json, arg string) string {
	ctx := Parse(json)
	if !ctx.IsObject() {
		return json
	}
	if strutil.IsEmpty(arg) {
		return json
	}
	defaults := Parse(arg)
	if !defaults.IsObject() {
		return json
	}
	// Collect the set of all fields that appear in the object (including null ones).
	// The second pass only appends keys that are entirely absent from the input.
	present := make(map[string]bool)
	ctx.Foreach(func(key, _ Context) bool {
		present[key.String()] = true
		return true
	})
	out := make([]byte, 0, len(json)+len(arg))
	out = append(out, '{')
	var i int
	// Emit original fields, substituting null with the default when available.
	ctx.Foreach(func(key, value Context) bool {
		if i > 0 {
			out = append(out, ',')
		}
		out = append(out, key.raw...)
		out = append(out, ':')
		if value.kind == Null {
			if dv := defaults.Get(key.String()); dv.Exists() {
				out = append(out, dv.raw...)
			} else {
				out = append(out, value.raw...)
			}
		} else {
			out = append(out, value.raw...)
		}
		i++
		return true
	})
	// Append default fields that were missing entirely.
	defaults.Foreach(func(key, value Context) bool {
		if !present[key.String()] {
			if i > 0 {
				out = append(out, ',')
			}
			out = append(out, key.raw...)
			out = append(out, ':')
			out = append(out, value.raw...)
			i++
		}
		return true
	})
	out = append(out, '}')
	return strutil.SafeStr(out)
}

// applyCoerce converts a scalar JSON value to the type specified by the `arg`.
//
// Supported target types (case-insensitive):
//   - `"string"`: converts the value to a JSON string.
//   - `"number"`: parses the value as a float64 and re-emits it; returns `null`
//     when conversion is not possible.
//   - `"bool"` / `"boolean"`: interprets truthy values (non-zero numbers, "true",
//     "1", "yes") as `true`, everything else as `false`.
//
// Objects and arrays are returned unchanged for any target type.
//
// Example:
//
//	fj.Get(`42`,    `@coerce:{"to":"string"}`)  // → "42"
//	fj.Get(`"99"`,  `@coerce:{"to":"number"}`)  // → 99
//	fj.Get(`1`,     `@coerce:{"to":"bool"}`)    // → true
//
// Performance: no heap allocations for number→string and bool conversions.
func applyCoerce(json, arg string) string {
	ctx := Parse(json)
	// Objects and arrays pass through unchanged.
	if ctx.IsObject() || ctx.IsArray() {
		return json
	}
	var to string
	Parse(arg).Foreach(func(key, value Context) bool {
		if key.String() == "to" {
			to = strings.ToLower(value.String())
		}
		return true
	})
	switch to {
	case "string":
		s := ctx.String()
		return string(appendJSONString(nil, s))
	case "number":
		f := ctx.Float64()
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return "null"
		}
		return fmt.Sprintf("%g", f)
	case "bool", "boolean":
		if ctx.Bool() {
			return "true"
		}
		return "false"
	}
	return json
}

// applyCount returns the number of elements in a JSON array, or the number of
// key-value pairs in a JSON object, as a plain JSON integer.
//
// For scalar values (strings, numbers, booleans, null) the result is always 0.
//
// Example:
//
//	fj.Get(`[1,2,3]`,          `@count`) // → 3
//	fj.Get(`{"a":1,"b":2}`,    `@count`) // → 2
//	fj.Get(`"hello"`,          `@count`) // → 0
//
// Performance: single Foreach pass with no allocations.
func applyCount(json, arg string) string {
	ctx := Parse(json)
	if !ctx.IsArray() && !ctx.IsObject() {
		return "0"
	}
	var n int
	ctx.Foreach(func(_, _ Context) bool {
		n++
		return true
	})
	return fmt.Sprintf("%d", n)
}

// applyFirst returns the first element of a JSON array as a raw JSON value.
// Returns `null` if the array is empty or the input is not an array.
//
// Example:
//
//	fj.Get(`[10,20,30]`, `@first`) // → 10
//	fj.Get(`[]`,         `@first`) // → null
//
// Performance: early-exit Foreach; at most one element is examined.
func applyFirst(json, arg string) string {
	ctx := Parse(json)
	if !ctx.IsArray() {
		return "null"
	}
	var first string
	ctx.Foreach(func(_, value Context) bool {
		first = value.raw
		return false // stop after the first element
	})
	if strutil.IsEmpty(first) {
		return "null"
	}
	return first
}

// applyLast returns the last element of a JSON array as a raw JSON value.
// Returns `null` if the array is empty or the input is not an array.
//
// Example:
//
//	fj.Get(`[10,20,30]`, `@last`) // → 30
//	fj.Get(`[]`,         `@last`) // → null
//
// Performance: full Foreach pass; the last raw value overwrites on each iteration.
func applyLast(json, arg string) string {
	ctx := Parse(json)
	if !ctx.IsArray() {
		return "null"
	}
	var last string
	ctx.Foreach(func(_, value Context) bool {
		last = value.raw
		return true
	})
	if strutil.IsEmpty(last) {
		return "null"
	}
	return last
}

// applySum returns the arithmetic sum of all numeric values in a JSON array.
// Non-numeric elements (strings, objects, arrays, booleans, null) are skipped.
// Returns `0` when the input is not an array or the array contains no numbers.
//
// Example:
//
//	fj.Get(`[1,2,3,4]`,          `@sum`) // → 10
//	fj.Get(`[1.5,2.5,"x",null]`, `@sum`) // → 4
//
// Performance: single Foreach pass; no allocations beyond the return string.
func applySum(json, arg string) string {
	ctx := Parse(json)
	if !ctx.IsArray() {
		return "0"
	}
	var sum float64
	ctx.Foreach(func(_, value Context) bool {
		if value.kind == Number {
			sum += value.Float64()
		}
		return true
	})
	return formatNumber(sum)
}

// applyMin returns the minimum numeric value found in a JSON array.
// Non-numeric elements are skipped.  Returns `null` if the array is empty or
// contains no numbers.
//
// Example:
//
//	fj.Get(`[3,1,4,1,5]`, `@min`) // → 1
//
// Performance: single Foreach pass; no heap allocations.
func applyMin(json, arg string) string {
	ctx := Parse(json)
	if !ctx.IsArray() {
		return "null"
	}
	min := math.MaxFloat64
	found := false
	ctx.Foreach(func(_, value Context) bool {
		if value.kind == Number {
			f := value.Float64()
			if !found || f < min {
				min = f
				found = true
			}
		}
		return true
	})
	if !found {
		return "null"
	}
	return formatNumber(min)
}

// applyMax returns the maximum numeric value found in a JSON array.
// Non-numeric elements are skipped.  Returns `null` if the array is empty or
// contains no numbers.
//
// Example:
//
//	fj.Get(`[3,1,4,1,5]`, `@max`) // → 5
//
// Performance: single Foreach pass; no heap allocations.
func applyMax(json, arg string) string {
	ctx := Parse(json)
	if !ctx.IsArray() {
		return "null"
	}
	max := -math.MaxFloat64
	found := false
	ctx.Foreach(func(_, value Context) bool {
		if value.kind == Number {
			f := value.Float64()
			if !found || f > max {
				max = f
				found = true
			}
		}
		return true
	})
	if !found {
		return "null"
	}
	return formatNumber(max)
}

// applyPluck extracts a named field from every element of a JSON array,
// returning a new JSON array of the extracted values.  Elements that do not
// contain the named field are omitted from the result.
//
// The `arg` is the plain field name (not a JSON string).  Nested path
// expressions are supported (e.g. `"address.city"`).
//
// Example:
//
//	fj.Get(`[{"name":"Alice","age":30},{"name":"Bob","age":25}]`, `@pluck:name`)
//	// → ["Alice","Bob"]
//
//	fj.Get(`[{"addr":{"city":"NY"}},{"addr":{"city":"LA"}}]`, `@pluck:addr.city`)
//	// → ["NY","LA"]
//
// Performance: single Foreach pass; pre-grown output slice.
func applyPluck(json, arg string) string {
	ctx := Parse(json)
	if !ctx.IsArray() {
		return "[]"
	}
	fieldPath := strings.TrimSpace(arg)
	if strutil.IsEmpty(fieldPath) {
		return "[]"
	}
	out := make([]byte, 0, len(json))
	out = append(out, '[')
	var i int
	ctx.Foreach(func(_, elem Context) bool {
		val := elem.Get(fieldPath)
		if val.Exists() {
			if i > 0 {
				out = append(out, ',')
			}
			raw := val.raw
			if raw == "" {
				raw = "null"
			}
			out = append(out, raw...)
			i++
		}
		return true
	})
	out = append(out, ']')
	return strutil.SafeStr(out)
}

// init registers all built-in transformers into globalRegistry at package startup.
// Aliases allow shorter names for commonly used transformers.
func init() {
	// Core transformers
	globalRegistry.Register("pretty", TransformerFunc(applyPrettyFormat))
	globalRegistry.Register("minify", TransformerFunc(applyMinify))
	globalRegistry.Register("ugly", TransformerFunc(applyMinify)) // alias for minify
	globalRegistry.Register("reverse", TransformerFunc(applyReverse))
	globalRegistry.Register("flatten", TransformerFunc(applyFlatten))
	globalRegistry.Register("join", TransformerFunc(applyMerge))
	globalRegistry.Register("valid", TransformerFunc(applyValidityCheck))
	globalRegistry.Register("keys", TransformerFunc(applyKeys))
	globalRegistry.Register("values", TransformerFunc(applyValues))
	globalRegistry.Register("json", TransformerFunc(applyToJSON))
	globalRegistry.Register("string", TransformerFunc(applyToString))
	globalRegistry.Register("group", TransformerFunc(applyGroup))
	globalRegistry.Register("search", TransformerFunc(applySearch))
	globalRegistry.Register("this", TransformerFunc(applyIdentity))

	// String case transformers
	globalRegistry.Register("uppercase", TransformerFunc(applyUppercase))
	globalRegistry.Register("upper", TransformerFunc(applyUppercase)) // alias
	globalRegistry.Register("lowercase", TransformerFunc(applyLowercase))
	globalRegistry.Register("lower", TransformerFunc(applyLowercase)) // alias
	globalRegistry.Register("flip", TransformerFunc(applyFlip))
	globalRegistry.Register("trim", TransformerFunc(applyTrim))
	globalRegistry.Register("snakecase", TransformerFunc(applySnakeCase))
	globalRegistry.Register("snakeCase", TransformerFunc(applySnakeCase)) // legacy alias
	globalRegistry.Register("snake", TransformerFunc(applySnakeCase))     // short alias
	globalRegistry.Register("camelcase", TransformerFunc(applyCamelCase))
	globalRegistry.Register("camelCase", TransformerFunc(applyCamelCase)) // legacy alias
	globalRegistry.Register("camel", TransformerFunc(applyCamelCase))     // short alias
	globalRegistry.Register("kebabcase", TransformerFunc(applyKebabCase))
	globalRegistry.Register("kebabCase", TransformerFunc(applyKebabCase)) // legacy alias
	globalRegistry.Register("kebab", TransformerFunc(applyKebabCase))     // short alias

	// Extra string transformers
	globalRegistry.Register("replace", TransformerFunc(applyReplace))
	globalRegistry.Register("replaceAll", TransformerFunc(applyReplaceAll))
	globalRegistry.Register("hex", TransformerFunc(applyToHex))
	globalRegistry.Register("bin", TransformerFunc(applyToBinary))
	globalRegistry.Register("insertAt", TransformerFunc(applyInsertAt))
	globalRegistry.Register("wc", TransformerFunc(applyCountWords))        // short alias for wordCount
	globalRegistry.Register("wordCount", TransformerFunc(applyCountWords)) // legacy alias
	globalRegistry.Register("padLeft", TransformerFunc(applyPadLeft))
	globalRegistry.Register("padRight", TransformerFunc(applyPadRight))

	// Structural / object transformers
	globalRegistry.Register("project", TransformerFunc(applyProject))

	// Array filtering and aggregation
	globalRegistry.Register("filter", TransformerFunc(applyFilter))
	globalRegistry.Register("count", TransformerFunc(applyCount))
	globalRegistry.Register("first", TransformerFunc(applyFirst))
	globalRegistry.Register("last", TransformerFunc(applyLast))
	globalRegistry.Register("sum", TransformerFunc(applySum))
	globalRegistry.Register("min", TransformerFunc(applyMin))
	globalRegistry.Register("max", TransformerFunc(applyMax))
	globalRegistry.Register("pluck", TransformerFunc(applyPluck))

	// Value normalization
	globalRegistry.Register("default", TransformerFunc(applyDefault))
	globalRegistry.Register("coerce", TransformerFunc(applyCoerce))
}
