package fj

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/polarixa/replify/pkg/common"
)

// Parse parses a JSON string and returns a Context representing the parsed value.
//
// This function processes the input JSON string and attempts to determine the type of the value it represents.
// It handles objects, arrays, numbers, strings, booleans, and null values. The function does not validate whether
// the JSON is well-formed, and instead returns a Context object that represents the first valid JSON element found
// in the string. Invalid JSON may result in unexpected behavior, so for input from unpredictable sources, consider
// using the `Valid` function first.
//
// Parameters:
//   - `json`: A string containing the JSON data to be parsed. This function expects well-formed JSON and does not
//     perform comprehensive validation.
//
// Returns:
//   - A `Context` that represents the parsed JSON element. The `Context` contains details about the type, value,
//     and position of the JSON element, including raw and unprocessed string data.
//
// Notes:
//   - The function attempts to determine the type of the JSON element by inspecting the first character in the
//     string. It supports the following types: Object (`{`), Array (`[`), Number, String (`"`), Boolean (`true` / `false`),
//     and Null (`null`).
//   - The function sets the `raw` field of the `Context` to the raw JSON string for further processing, and
//     sets the `kind` field to represent the type of the value (e.g., `String`, `Number`, `True`, `False`, `JSON`, `Null`).
//
// Example Usage:
//
//	json := "{\"name\": \"John\", \"age\": 30}"
//	ctx := Parse(json)
//	fmt.Println(ctx.kind) // Output: JSON (if the input starts with '{')
//
//	json := "12345"
//	ctx := Parse(json)
//	fmt.Println(ctx.kind) // Output: Number (if the input is a numeric value)
//
//	json := "\"Hello, World!\""
//	ctx := Parse(json)
//	fmt.Println(ctx.kind) // Output: String (if the input is a string)
//
// Returns:
//   - `Context`: The parsed result, which may represent an object, array, string, number, boolean, or null.
func Parse(json string) Context {
	var value Context
	i := 0
	for ; i < len(json); i++ {
		if json[i] == '{' || json[i] == '[' {
			value.kind = JSON
			value.raw = json[i:]
			break
		}
		if json[i] <= ' ' {
			continue
		}
		switch json[i] {
		case '+', '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
			'i', 'I', 'N':
			value.kind = Number
			value.raw, value.num = tokenizeNumber(json[i:])
		case 'n':
			if i+1 < len(json) && json[i+1] != 'u' {
				// nan
				value.kind = Number
				value.raw, value.num = tokenizeNumber(json[i:])
			} else {
				// null
				value.kind = Null
				value.raw = leadingLowercase(json[i:])
			}
		case 't':
			value.kind = True
			value.raw = leadingLowercase(json[i:])
		case 'f':
			value.kind = False
			value.raw = leadingLowercase(json[i:])
		case '"':
			value.kind = String
			value.raw, value.str = extractAndUnescapeJSONString(json[i:])
		default:
			return Context{}
		}
		break
	}
	if value.Exists() {
		value.idx = i
	}
	return value
}

// ParseBytes parses a JSON byte slice and returns a Context representing the parsed value.
//
// This function is a wrapper around the `Parse` function, designed specifically for handling JSON data
// in the form of a byte slice. It converts the byte slice into a string and then calls `Parse` to process
// the JSON data. If you're working with raw JSON data as bytes, using this method is preferred over
// manually converting the bytes to a string and passing it to `Parse`.
//
// Parameters:
//   - `json`: A byte slice containing the JSON data to be parsed.
//
// Returns:
//   - A `Context` representing the parsed JSON element, similar to the behavior of `Parse`. The `Context`
//     contains information about the type, value, and position of the JSON element, including the raw and
//     unprocessed string data.
//
// Example Usage:
//
//	json := []byte("{\"name\": \"Alice\", \"age\": 25}")
//	ctx := ParseBytes(json)
//	fmt.Println(ctx.kind) // Output: JSON (if the input is an object)
//
// Returns:
//   - `Context`: The parsed result, representing the parsed JSON element, such as an object, array, string,
//     number, boolean, or null.
func ParseBytes(json []byte) Context {
	return Parse(string(json))
}

// ParseReader reads a JSON string from an `io.Reader` and parses it into a Context.
//
// This function combines the reading and parsing operations. It first reads the JSON
// string using the `BufioRead` function, then processes the string using the `Parse`
// function to extract details about the first valid JSON element. If reading the JSON
// fails, the returned Context contains the encountered error.
//
// Parameters:
//   - in: An `io.Reader` from which the JSON string will be read. This could be a file,
//     network connection, or any other source that implements the `io.Reader` interface.
//
// Returns:
//   - A `Context` representing the parsed JSON element. If an error occurs during the
//     reading process, the `err` field of the `Context` is set with the error, and the
//     other fields remain empty.
//
// Example Usage:
//
//	// Reading JSON from a file
//	file, err := os.Open("example.json")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer file.Close()
//
//	ctx := ParseReader(file)
//	if ctx.err != nil {
//	    log.Fatalf("Failed to parse JSON: %v", ctx.err)
//	}
//	fmt.Println(ctx.kind)
//
//	// Reading JSON from standard input
//	fmt.Println("Enter JSON:")
//	ctx = ParseReader(os.Stdin)
//	if ctx.err != nil {
//	    log.Fatalf("Failed to parse JSON: %v", ctx.err)
//	}
//	fmt.Println(ctx.kind)
//
// Notes:
//   - This function is particularly useful when working with JSON data from streams or large files.
//   - The `Parse` function is responsible for the actual parsing, while this function ensures the
//     JSON string is read correctly before parsing begins.
//   - If the input JSON is malformed or invalid, the returned Context from `Parse` will reflect
//     the issue as an empty Context or an error in the `err` field.
func ParseReader(in io.Reader) Context {
	json, err := common.SlurpLine(in)
	if err != nil {
		return Context{
			err: err,
		}
	}
	return Parse(json)
}

// ParseJSONFile reads a JSON string from a file specified by the filepath and parses it into a Context.
//
// This function opens the specified file, reads its contents using the `ParseReader` function,
// and returns a Context representing the parsed JSON element. If any error occurs during
// file reading or JSON parsing, the returned Context will include the error details.
//
// Parameters:
//   - filepath: The path to the JSON file to be read and parsed.
//
// Returns:
//   - A `Context` representing the parsed JSON element. If an error occurs during the
//     file reading or parsing process, the `err` field of the returned Context will be set
//     with the error, and the other fields will remain empty.
//
// Example Usage:
//
//	ctx := ParseJSONFile("example.json")
//	if ctx.err != nil {
//	    log.Fatalf("Failed to parse JSON: %v", ctx.err)
//	}
//	fmt.Println(ctx.String())
//
// Notes:
//   - This function is useful for reading and parsing JSON directly from files.
//   - The file is opened and closed automatically within this function, and any errors encountered
//     are captured in the returned `Context`.
func ParseJSONFile(filepath string) Context {
	if !strings.HasSuffix(filepath, ".json") {
		return Context{
			err: fmt.Errorf("filepath is not a JSON file: %s", filepath),
		}
	}
	file, err := os.Open(filepath)
	if err != nil {
		return Context{
			err: err,
		}
	}
	defer file.Close()
	return ParseReader(file)
}

// Get searches for a specified path within the provided JSON string and returns the corresponding value as a Context.
// The path is provided in dot notation, where each segment represents a key or index. The function supports wildcards
// (`*` and `?`), array indexing, and special characters like '#' to access array lengths or child paths. The function
// will return the first matching result it finds along the specified path.
//
// Path Syntax:
// - Dot notation: "name.last" or "age" for direct key lookups.
// - Wildcards: "*" matches any key, "?" matches a single character.
// - Array indexing: "children.0" accesses the first item in the "children" array.
// - The '#' character returns the number of elements in an array (e.g., "children.#" returns the array length).
// - The dot (`.`) and wildcard characters (`*`, `?`) can be escaped with a backslash (`\`).
//
// Example Usage:
//
//	json := `{
//	  "user": {"firstName": "Alice", "lastName": "Johnson"},
//	  "age": 29,
//	  "siblings": ["Ben", "Clara", "David"],
//	  "friends": [
//	    {"firstName": "Tom", "lastName": "Smith"},
//	    {"firstName": "Sophia", "lastName": "Davis"}
//	  ],
//	  "address": {"city": "New York", "zipCode": "10001"}
//	}`
//
//	// Examples of Get function with paths:
//	Get(json, "user.lastName")        // Returns: "Johnson"
//	Get(json, "age")                  // Returns: 29
//	Get(json, "siblings.#")           // Returns: 3 (number of siblings)
//	Get(json, "siblings.1")           // Returns: "Clara" (second sibling)
//	Get(json, "friends.#.firstName")  // Returns: ["Tom", "Sophia"]
//	Get(json, "address.zipCode")      // Returns: "10001"
//
// Details:
//   - The function does not validate JSON format but expects well-formed input.
//     Invalid JSON may result in unexpected behavior.
//   - transformers (e.g., `@` for adjusting paths) and special sub-selectors (e.g., `[` and `{`) are supported and processed
//     in the path before extracting values.
//   - For complex structures, the function analyzes the provided path, handles nested arrays or objects, and returns
//     a Context containing the value found at the specified location.
//
// Parameters:
//   - `json`: A string containing the JSON data to search through.
//   - `path`: A string representing the path to the desired value, using dot notation or other special characters as described.
//
// Returns:
//   - `Context`: A Context object containing the value found at the specified path, including information such as the
//     type (`kind`), the raw JSON string (`raw`), and the parsed value if available (e.g., `str` for strings).
//
// Notes:
//   - If the path is not found, the returned Context will reflect this with an empty or null value.
func Get(json, path string) Context {
	if len(path) > 1 {
		if (path[0] == '@' && !DisableTransformers) || path[0] == '!' {
			var ok bool
			var cPath string
			var cJson string
			if path[0] == '@' && !DisableTransformers {
				cPath, cJson, ok = applyTransformerAt(json, path)
			} else if path[0] == '!' {
				cPath, cJson, ok = splitStaticAndLiteral(path)
			}
			if ok {
				path = cPath
				if len(path) > 0 && (path[0] == '|' || path[0] == '.') {
					res := Get(cJson, path[1:])
					res.idx = 0
					res.idxs = nil
					return res
				}
				return Parse(cJson)
			}
		}
		if path[0] == '[' || path[0] == '{' {
			kind := path[0] // using a sub-selector path
			var ok bool
			var subs []sel
			subs, path, ok = parseSubSelectors(path)
			if ok {
				if len(path) == 0 || (path[0] == '|' || path[0] == '.') {
					var b []byte
					b = append(b, kind)
					var i int
					for _, sub := range subs {
						res := Get(json, sub.path)
						if res.Exists() {
							if i > 0 {
								b = append(b, ',')
							}
							if kind == '{' {
								if len(sub.name) > 0 {
									if sub.name[0] == '"' && IsValidJSONString(sub.name) {
										b = append(b, sub.name...)
									} else {
										b = appendJSONString(b, sub.name)
									}
								} else {
									last := lastSegment(sub.path)
									if isValidName(last) {
										b = appendJSONString(b, last)
									} else {
										b = appendJSONString(b, "_")
									}
								}
								b = append(b, ':')
							}
							var raw string
							if len(res.raw) == 0 {
								raw = res.String()
								if len(raw) == 0 {
									raw = "null"
								}
							} else {
								raw = res.raw
							}
							b = append(b, raw...)
							i++
						}
					}
					b = append(b, kind+2)
					var res Context
					res.raw = string(b)
					res.kind = JSON
					if len(path) > 0 {
						res = res.Get(path[1:])
					}
					res.idx = 0
					return res
				}
			}
		}
	}
	var i int
	var c = &parser{json: json}
	if len(path) >= 2 && path[0] == '.' && path[1] == '.' {
		c.lines = true
		matchJSONArrayAt(c, 0, path[2:])
	} else {
		for ; i < len(c.json); i++ {
			if c.json[i] == '{' {
				i++
				matchJSONObjectAt(c, i, path)
				break
			}
			if c.json[i] == '[' {
				i++
				matchJSONArrayAt(c, i, path)
				break
			}
		}
	}
	if c.piped {
		res := c.val.Get(c.pipe)
		res.idx = 0
		return res
	}
	computeOffset(json, c)
	return c.val
}

// GetMulti searches json for multiple paths.
// The return value is a slice of `Context` objects, where the number of items
// will be equal to the number of input paths. Each `Context` represents the value
// extracted for the corresponding path.
//
// Parameters:
//   - `json`: A string containing the JSON data to search through.
//   - `path`: A variadic list of paths to search for within the JSON data.
//
// Returns:
//   - A slice of `Context` objects, one for each path provided in the `path` parameter.
//
// Notes:
//   - The function will return a `Context` for each path, and the order of the `Context`
//     objects in the result will match the order of the paths provided.
//
// Example:
//
//	json := `{
//	  "user": {"firstName": "Alice", "lastName": "Johnson"},
//	  "age": 29,
//	  "siblings": ["Ben", "Clara", "David"],
//	  "friends": [
//	    {"firstName": "Tom", "lastName": "Smith"},
//	    {"firstName": "Sophia", "lastName": "Davis"}
//	  ]
//	}`
//	paths := []string{"user.lastName", "age", "siblings.#", "friends.#.firstName"}
//	results := GetMulti(json, paths...)
//	// The result will contain Contexts for each path: ["Johnson", 29, 3, ["Tom", "Sophia"]]
func GetMulti(json string, path ...string) []Context {
	ctx := make([]Context, len(path))
	for i, path := range path {
		ctx[i] = Get(json, path)
	}
	return ctx
}

// GetBytes searches the provided JSON byte slice for the specified path and returns a `Context`
// representing the extracted data. This method is preferred over `Get(string(data), path)` when working
// with JSON data in byte slice format, as it directly operates on the byte slice, minimizing memory
// allocations and unnecessary copies.
//
// Parameters:
//   - `json`: A byte slice containing the JSON data to process.
//   - `path`: A string representing the path in the JSON data to extract.
//
// Returns:
//   - A `Context` struct containing the processed JSON data. The `Context` struct includes both
//     the raw unprocessed JSON string and the specific extracted string based on the given path.
//
// Notes:
//   - This function internally calls the `getBytes` function, which uses unsafe pointer operations
//     to minimize allocations and efficiently handle string slice headers.
//   - The function avoids unnecessary memory allocations by directly processing the byte slice and
//     utilizing memory safety features to manage substring extraction when the `str` part is
//     a substring of the `raw` part of the JSON data.
//
// Example:
//
//	jsonBytes := []byte(`{"key": "value", "nested": {"innerKey": "innerValue"}}`)
//	path := "nested.innerKey"
//	context := GetBytes(jsonBytes, path)
//	fmt.Println("Unprocessed:", context.raw) // Output: `{"key": "value", "nested": {"innerKey": "innerValue"}}`
//	fmt.Println("Strings:", context.str)         // Output: `"innerValue"`
func GetBytes(json []byte, path string) Context {
	return parseShared(json, path)
}

// GetBytesMulti searches json for multiple paths in the provided JSON byte slice.
// The return value is a slice of `Context` objects, where the number of items
// will be equal to the number of input paths. Each `Context` represents the value
// extracted for the corresponding path. This method operates directly on the byte slice,
// which is preferred when working with JSON data in byte format to minimize memory allocations.
//
// Parameters:
//   - `json`: A byte slice containing the JSON data to search through.
//   - `path`: A variadic list of paths to search for within the JSON data.
//
// Returns:
//   - A slice of `Context` objects, one for each path provided in the `path` parameter.
//
// Notes:
//   - The function will return a `Context` for each path, and the order of the `Context`
//     objects in the result will match the order of the paths provided.
//
// Example:
//
//	jsonBytes := []byte(`{"user": {"firstName": "Alice", "lastName": "Johnson"}, "age": 29}`)
//	paths := []string{"user.lastName", "age"}
//	results := GetBytesMulti(jsonBytes, paths...)
//	// The result will contain Contexts for each path: ["Johnson", 29]
func GetBytesMulti(json []byte, path ...string) []Context {
	ctx := make([]Context, len(path))
	for i, path := range path {
		ctx[i] = GetBytes(json, path)
	}
	return ctx
}

// Foreach iterates through each line of JSON data in the JSON Lines format (http://jsonlines.org/),
// and applies a provided iterator function to each line. This is useful for processing large JSON data
// sets where each line is a separate JSON object, allowing for efficient parsing and handling of each object.
//
// Parameters:
//   - `json`: A string containing JSON Lines formatted data, where each line is a separate JSON object.
//   - `iterator`: A callback function that is called for each line. It receives a `Context` representing
//     the parsed JSON object for the current line. The iterator function should return `true` to continue
//     processing the next line, or `false` to stop the iteration.
//
// Example Usage:
//
//	json := `{"name": "Alice"}\n{"name": "Bob"}`
//	iterator := func(line Context) bool {
//	    fmt.Println(line)
//	    return true
//	}
//	Foreach(json, iterator)
//	// Output:
//	// {"name": "Alice"}
//	// {"name": "Bob"}
//
// Notes:
//   - This function assumes the input `json` is formatted as JSON Lines, where each line is a valid JSON object.
//   - The function stops processing as soon as the `iterator` function returns `false` for a line.
//   - The function handles each line independently, meaning it processes one JSON object at a time and provides
//     it to the iterator, which can be used to process or filter lines.
//
// Returns:
//   - This function does not return a value. It processes the JSON data line-by-line and applies the iterator to each.
func Foreach(json string, iterator func(line Context) bool) {
	var ctx Context
	var i int
	for {
		i, ctx, _ = extractJSONValueAt(json, i, true)
		if !ctx.Exists() {
			break
		}
		if !iterator(ctx) {
			return
		}
	}
}

// IsValidJSON checks whether the provided byte slice contains valid JSON data.
// It attempts to parse the JSON and returns a boolean indicating if the JSON is well-formed.
//
// Parameters:
//   - `json`: A byte slice (`[]byte`) representing the JSON data that needs to be validated.
//
// Returns:
//   - A boolean value (`true` or `false`):
//   - `true`: The provided JSON byte slice is valid and well-formed.
//   - `false`: The provided JSON byte slice is invalid or malformed.
//
// Notes:
//   - This function works directly with a byte slice (`[]byte`) rather than a string, making it more efficient
//     when dealing with raw byte data that represents JSON. It avoids the need to convert between strings and
//     byte slices, which can improve performance and memory usage when working with large or binary JSON data.
//   - The function utilizes the `verifyJSON` function to check if the byte slice conforms to valid JSON syntax.
//   - If the input byte slice represents invalid JSON, the function will return `false`, indicating that the JSON
//     cannot be parsed or is improperly structured.
//   - The function does not perform deep validation of the content of the JSON, but only checks whether the byte slice
//     adheres to the syntax rules defined for JSON data structures.
//
// Example Usage:
//
//	jsonBytes := []byte(`{"name": {"first": "Alice", "last": "Johnson"}, "age": 30}`)
//	if !IsValidJSON(jsonBytes) {
//	    fmt.Println("Invalid JSON")
//	} else {
//	    fmt.Println("Valid JSON")
//	}
//
//	// Output: "Valid JSON"
func IsValidJSON(json []byte) bool {
	_, ok := expectJSON(json, 0)
	return ok
}

// IsValidJSONString checks whether the provided string contains valid JSON data.
// It attempts to parse the JSON and returns a boolean indicating if the JSON is well-formed.
//
// Parameters:
//   - `json`: A string representing the JSON data that needs to be validated.
//
// Returns:
//   - A boolean value (`true` or `false`):
//   - `true`: The provided JSON string is valid and well-formed.
//   - `false`: The provided JSON string is invalid or malformed.
//
// Notes:
//   - This function utilizes the `fromStr2Bytes` function to efficiently convert the input string
//     into a byte slice without allocating new memory. It then passes the byte slice to the
//     `verifyJSON` function to check if the string conforms to valid JSON syntax.
//   - If the input JSON is invalid, the function will return `false`, indicating that the JSON
//     cannot be parsed or is improperly structured.
//   - The function does not perform deep validation of the content of the JSON, but merely
//     checks if the string is syntactically correct according to JSON rules.
//
// Example Usage:
//
//	json := `{"name": {"first": "Alice", "last": "Johnson"}, "age": 30}`
//	if !IsValidJSONString(json) {
//	    fmt.Println("Invalid JSON")
//	} else {
//	    fmt.Println("IsValidJSONString JSON")
//	}
//
//	// Output: "IsValidJSONString JSON"
func IsValidJSONString(json string) bool {
	_, ok := expectJSON(UnsafeBytes(json), 0)
	return ok
}

// AddTransformer registers a Transformer implementation under the given name in the
// global registry. Registering with an existing name overwrites the previous entry.
// This function is safe for concurrent use.
//
// To register a plain function, wrap it with TransformerFunc:
//
//	fj.AddTransformer("upper", fj.TransformerFunc(func(json, arg string) string {
//	    return strings.ToUpper(json)
//	}))
//
// To register a struct-based implementation, pass a value that satisfies the
// Transformer interface:
//
//	type myTransformer struct{ prefix string }
//
//	func (t *myTransformer) Apply(json, arg string) string {
//	    return t.prefix + json
//	}
//
//	fj.AddTransformer("prefixed", &myTransformer{prefix: "data:"})
func AddTransformer(name string, t Transformer) {
	globalRegistry.Register(name, t)
}

// IsTransformerRegistered reports whether a transformer with the given name has been
// registered in the global registry.
//
// This function is safe for concurrent use by multiple goroutines.
func IsTransformerRegistered(name string) bool {
	return globalRegistry.IsRegistered(name)
}
