# encoding

**encoding** is a Go library providing utilities for JSON encoding, decoding, and formatting. It offers convenient wrappers around Go's standard `encoding/json` package plus advanced features like pretty-printing, minification, and colorized output for JSON data.

## Overview

The `encoding` package simplifies JSON operations in Go by providing:

- **JSON Marshaling/Unmarshaling**: Simplified wrappers for encoding and decoding
- **Pretty Printing**: Format JSON with customizable indentation and styling
- **Minification**: Remove whitespace from JSON (uglify)
- **Color Styling**: Apply syntax highlighting for terminal output
- **Flexible Options**: Customize width, sorting, indentation, and prefixes

**Problem Solved:** Working with JSON in Go requires verbose error handling and lacks built-in pretty-printing with customization. This package provides clean, reusable functions with sensible defaults while offering advanced formatting options when needed.

## Use Cases

### When to Use

- ✅ **API development** - marshal/unmarshal JSON request/response bodies
- ✅ **Configuration files** - read/write JSON config with pretty formatting
- ✅ **Logging** - format JSON logs with indentation for readability
- ✅ **CLI tools** - output formatted JSON to terminals with colors
- ✅ **Debugging** - pretty-print JSON for easier inspection
- ✅ **JSON minification** - reduce JSON size for transmission
- ✅ **Testing** - create readable JSON test fixtures
- ✅ **Documentation** - generate formatted JSON examples

### When Not to Use

- ❌ **High-performance streaming** - use `encoding/json` Encoder/Decoder directly
- ❌ **Custom JSON parsers** - when you need non-standard JSON handling
- ❌ **Binary protocols** - use Protocol Buffers, MessagePack, etc.
- ❌ **Complex transformations** - consider dedicated JSON processing libraries
- ❌ **Schema validation** - use JSON Schema validators

## Installation

```bash
go get github.com/polarixa/replify
```

Import the package in your Go code:

```go
import "github.com/polarixa/replify/pkg/encoding"
```

**Requirements:** Go 1.13 or higher

## Usage

### Quick Start

```go
package main

import (
    "fmt"
    "github.com/polarixa/replify/pkg/encoding"
)

func main() {
    // Create a struct
    type User struct {
        Name  string `json:"name"`
        Age   int    `json:"age"`
        Email string `json:"email"`
    }

    user := User{Name: "Alice", Age: 30, Email: "alice@example.com"}

    // Marshal to JSON
    jsonBytes, err := encoding.MarshalJSONb(user)
    if err != nil {
        panic(err)
    }
    fmt.Println(string(jsonBytes))
    // Output: {"name":"Alice","age":30,"email":"alice@example.com"}

    // Pretty print
    pretty := encoding.Pretty(jsonBytes)
    fmt.Println(string(pretty))
    // Output:
    // {
    //   "name": "Alice",
    //   "age": 30,
    //   "email": "alice@example.com"
    // }
}
```

## Examples

### 1. Basic JSON Marshaling

```go
package main

import (
	"fmt"
	"log"

	"github.com/polarixa/replify/pkg/encoding"
)

func main() {
	type Product struct {
		ID    int     `json:"id"`
		Name  string  `json:"name"`
		Price float64 `json:"price"`
	}

	product := Product{ID: 1, Name: "Laptop", Price: 999.99}

	// Marshal to bytes
	jsonBytes, err := encoding.MarshalJSONb(product)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(jsonBytes))
	// Output: {"id":1,"name":"Laptop","price":999.99}

	// Marshal to string directly
	jsonString, err := encoding.MarshalJSONs(product)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(jsonString)
	// Output: {"id":1,"name":"Laptop","price":999.99}

	// Marshal with indentation
	indented, err := encoding.MarshalJSONIndent(product, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(indented))
	// Output:
	//
	//	{
	//	  "id": 1,
	//	  "name": "Laptop",
	//	  "price": 999.99
	//	}
}
```

### 2. JSON Unmarshaling

```go
package main

import (
	"fmt"
	"log"

	"github.com/polarixa/replify/pkg/encoding"
)

func main() {
	type User struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email"`
	}

	// From bytes
	jsonBytes := []byte(`{"name":"Bob","age":25,"email":"bob@example.com"}`)
	var user User
	err := encoding.UnmarshalBytes(jsonBytes, &user)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", user)
	// Output: {Name:Bob Age:25 Email:bob@example.com}

	// From string
	jsonString := `{"name":"Charlie","age":35,"email":"charlie@example.com"}`
	var anotherUser User
	err = encoding.UnmarshalJSON(jsonString, &anotherUser)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", anotherUser)
	// Output: {Name:Charlie Age:35 Email:charlie@example.com}
}
```

### 3. Pretty Printing JSON

```go
// Compact JSON
compact := []byte(`{"name":"Alice","age":30,"hobbies":["reading","coding","gaming"],"address":{"city":"New York","zip":"10001"}}`)

// Default pretty printing
pretty := encoding.Pretty(compact)
fmt.Println(string(pretty))
// Output:
// {
//   "name": "Alice",
//   "age": 30,
//   "hobbies": [
//     "reading",
//     "coding",
//     "gaming"
//   ],
//   "address": {
//     "city": "New York",
//     "zip": "10001"
//   }
// }

// Custom options
options := &encoding.OptionsConfig{
    Width:    120,
    Prefix:   "  ",
    Indent:   "    ",
    SortKeys: true,
}
customPretty := encoding.PrettyOptions(compact, options)
fmt.Println(string(customPretty))
// Output: (with custom indentation and sorted keys)
```

### 4. JSON Minification (Uglify)

```go
// Pretty JSON with whitespace
prettyJSON := []byte(`{
  "name": "Alice",
  "age": 30,
  "email": "alice@example.com"
}`)

// Remove all whitespace
minified := encoding.CompactJSON(prettyJSON)
fmt.Println(string(minified))
// Output: {"name":"Alice","age":30,"email":"alice@example.com"}

// In-place uglification (reuses buffer)
encoding.CompactJSONUnsafe(prettyJSON)
fmt.Println(string(prettyJSON))
// Output: {"name":"Alice","age":30,"email":"alice@example.com"}
```

### 5. Pretty Printing with Custom Options

```go
type Config struct {
    Server   string            `json:"server"`
    Port     int               `json:"port"`
    Features map[string]bool   `json:"features"`
    Users    []string          `json:"users"`
}

config := Config{
    Server: "localhost",
    Port:   8080,
    Features: map[string]bool{
        "logging":     true,
        "compression": false,
        "caching":     true,
    },
    Users: []string{"admin", "user1", "user2"},
}

jsonBytes, _ := encoding.Marshal(config)

// Option 1: Wide format (arrays on single line if they fit)
wideOptions := &encoding.OptionsConfig{
    Width:    120,
    Indent:   "  ",
    SortKeys: true,
}
wide := encoding.PrettyOptions(jsonBytes, wideOptions)
fmt.Println(string(wide))

// Option 2: Narrow format (more line breaks)
narrowOptions := &encoding.OptionsConfig{
    Width:    40,
    Indent:   "  ",
    SortKeys: false,
}
narrow := encoding.PrettyOptions(jsonBytes, narrowOptions)
fmt.Println(string(narrow))

// Option 3: With prefix (useful for logs)
prefixOptions := &encoding.OptionsConfig{
    Width:    80,
    Prefix:   "[JSON] ",
    Indent:   "  ",
    SortKeys: true,
}
prefixed := encoding.PrettyOptions(jsonBytes, prefixOptions)
fmt.Println(string(prefixed))
```

### 6. Sorting JSON Keys

```go
unsortedJSON := []byte(`{"zebra":1,"apple":2,"banana":3}`)

// Pretty print with sorted keys
options := &encoding.OptionsConfig{
    Width:    80,
    Indent:   "  ",
    SortKeys: true,
}
sorted := encoding.PrettyOptions(unsortedJSON, options)
fmt.Println(string(sorted))
// Output:
// {
//   "apple": 2,
//   "banana": 3,
//   "zebra": 1
// }
```

### 7. Practical Use Cases

#### API Response Formatting

```go
type APIResponse struct {
    Status  string      `json:"status"`
    Data    interface{} `json:"data"`
    Message string      `json:"message"`
}

func formatResponse(data interface{}) (string, error) {
    response := APIResponse{
        Status:  "success",
        Data:    data,
        Message: "Request processed successfully",
    }

    return encoding.MarshalJSONs(response)
}

// Usage
jsonString, err := formatResponse(map[string]int{"count": 42})
if err != nil {
    log.Fatal(err)
}
fmt.Println(jsonString)
```

#### Configuration File Handling

```go
type AppConfig struct {
    Database DatabaseConfig `json:"database"`
    Server   ServerConfig   `json:"server"`
    Logging  LoggingConfig  `json:"logging"`
}

func SaveConfig(config AppConfig, filename string) error {
    // Marshal with indentation for readability
    data, err := encoding.MarshalJSONIndent(config, "", "  ")
    if err != nil {
        return err
    }

    return os.WriteFile(filename, data, 0644)
}

func LoadConfig(filename string) (*AppConfig, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }

    var config AppConfig
    err = encoding.UnmarshalBytes(data, &config)
    return &config, err
}
```

#### Debug Logging

```go
func LogJSON(label string, data interface{}) {
    jsonBytes, err := encoding.MarshalJSONb(data)
    if err != nil {
        log.Printf("[ERROR] Failed to marshal %s: %v", label, err)
        return
    }

    // Pretty print for logs
    pretty := encoding.Pretty(jsonBytes)
    log.Printf("[DEBUG] %s:\n%s", label, string(pretty))
}

// Usage
user := User{Name: "Alice", Age: 30}
LogJSON("User Data", user)
```

#### HTTP Request/Response

```go
func HandleRequest(w http.ResponseWriter, r *http.Request) {
    var requestData map[string]interface{}

    // Unmarshal request body
    body, _ := io.ReadAll(r.Body)
    if err := encoding.UnmarshalBytes(body, &requestData); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    // Process data...
    responseData := map[string]interface{}{
        "status": "success",
        "data":   requestData,
    }

    // Marshal response
    jsonBytes, _ := encoding.MarshalJSONb(responseData)
    w.Header().Set("Content-Type", "application/json")
    w.Write(jsonBytes)
}
```

#### CLI Output Formatting

```go
func DisplayResults(results []Result) {
    jsonBytes, err := encoding.MarshalJSONb(results)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        return
    }

    // Pretty print for terminal
    pretty := encoding.Pretty(jsonBytes)
    fmt.Println(string(pretty))
}
```

### 8. Working with Complex Nested Structures

```go
type Company struct {
    Name       string       `json:"name"`
    Employees  []Employee   `json:"employees"`
    Departments []Department `json:"departments"`
}

type Employee struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

type Department struct {
    Name    string   `json:"name"`
    Manager string   `json:"manager"`
    Members []int    `json:"member_ids"`
}

company := Company{
    Name: "TechCorp",
    Employees: []Employee{
        {ID: 1, Name: "Alice", Email: "alice@techcorp.com"},
        {ID: 2, Name: "Bob", Email: "bob@techcorp.com"},
    },
    Departments: []Department{
        {Name: "Engineering", Manager: "Alice", Members: []int{1, 2}},
    },
}

// Marshal and pretty print
jsonBytes, _ := encoding.MarshalJSONb(company)
pretty := encoding.Pretty(jsonBytes)
fmt.Println(string(pretty))
```

## API Reference

### Core Functions

#### Marshal Functions

| Function            | Signature                                        | Description                       |
| ------------------- | ------------------------------------------------ | --------------------------------- |
| `MarshalJSONb`      | `(v any) ([]byte, error)`                        | Convert Go value to JSON bytes    |
| `MarshalJSONs`      | `(v any) (string, error)`                        | Convert Go value to JSON string   |
| `MarshalJSONIndent` | `(v any, prefix, indent string) ([]byte, error)` | Convert Go value to indented JSON |

**Examples:**

```go
bytes, err := encoding.MarshalJSONb(data)
str, err := encoding.MarshalJSONs(data)
indented, err := encoding.MarshalJSONIndent(data, "", "  ")
```

---

#### Unmarshal Functions

| Function         | Signature                    | Description                     |
| ---------------- | ---------------------------- | ------------------------------- |
| `UnmarshalBytes` | `(data []byte, v any) error` | Parse JSON bytes into Go value  |
| `UnmarshalJSON`  | `(str string, v any) error`  | Parse JSON string into Go value |

**Examples:**

```go
err := encoding.UnmarshalBytes(jsonBytes, &myStruct)
err := encoding.UnmarshalJSON(jsonString, &myStruct)
```

---

#### Formatting Functions

| Function        | Signature                                   | Description                          |
| --------------- | ------------------------------------------- | ------------------------------------ |
| `Pretty`        | `(json []byte) []byte`                      | Format JSON with default options     |
| `PrettyOptions` | `(json []byte, opts *OptionsConfig) []byte` | Format JSON with custom options      |
| `Ugly`          | `(json []byte) []byte`                      | Remove whitespace (minify JSON)      |
| `UglyInPlace`   | `(json []byte) []byte`                      | Minify JSON in-place (reuses buffer) |

**Examples:**

```go
pretty := encoding.Pretty(compactJSON)
custom := encoding.PrettyOptions(json, options)
minified := encoding.CompactJSON(prettyJSON)
```

---

### Options Configuration

#### OptionsConfig Struct

```go
type OptionsConfig struct {
    Width    int    // Max column width for single-line arrays (default: 80)
    Prefix   string // Prefix for all lines (default: "")
    Indent   string // Indentation string (default: "  ")
    SortKeys bool   // Sort object keys alphabetically (default: false)
}
```

**Fields:**

- `Width`: Maximum column width before wrapping arrays/objects (default: 80 characters)
- `Prefix`: String prepended to each line (useful for logging)
- `Indent`: Indentation string (typically `"  "` or `"    "`)
- `SortKeys`: Whether to sort JSON object keys alphabetically

**Example:**

```go
options := &encoding.OptionsConfig{
    Width:    120,
    Prefix:   "[LOG] ",
    Indent:   "    ",
    SortKeys: true,
}
formatted := encoding.PrettyOptions(jsonBytes, options)
```

---

### Default Configuration

```go
var DefaultOptionsConfig = &OptionsConfig{
    Width:    80,
    Prefix:   "",
    Indent:   "  ",
    SortKeys: false,
}
```

Use this when calling `Pretty()` or pass `nil` to `PrettyOptions()`.

## Best Practices & Notes

### ⚠️ Common Pitfalls

1. **Forgetting Error Handling**

   ```go
   // ❌ Bad: ignoring errors
   jsonBytes, _ := encoding.MarshalJSONb(data)

   // ✅ Good: handle errors
   jsonBytes, err := encoding.MarshalJSONb(data)
   if err != nil {
       log.Printf("Failed to marshal: %v", err)
       return err
   }
   ```

2. **Unmarshaling into Wrong Type**

   ```go
   // ❌ Bad: unmarshaling into non-pointer
   var user User
   encoding.UnmarshalBytes(data, user) // Won't work!

   // ✅ Good: pass pointer
   var user User
   encoding.UnmarshalBytes(data, &user)
   ```

3. **Using Pretty in Production APIs**

   ```go
   // ❌ Bad: pretty-printing adds overhead
   jsonBytes, _ := encoding.MarshalJSONb(data)
   pretty := encoding.Pretty(jsonBytes)
   w.Write(pretty) // Extra processing!

   // ✅ Good: use compact format
   jsonBytes, _ := encoding.MarshalJSONb(data)
   w.Write(jsonBytes)
   ```

4. **Modifying Original Buffer with UglyInPlace**

   ```go
   // ⚠️ Warning: modifies original
   original := []byte(`{ "key": "value" }`)
   minified := encoding.CompactJSONUnsafe(original)
   // original is now modified!

   // ✅ Safe: use Ugly for new buffer
   minified := encoding.CompactJSON(original)
   // original unchanged
   ```

### 💡 Recommendations

✅ **Use appropriate functions for the context**

```go
// For APIs: compact format
jsonBytes, _ := encoding.MarshalJSONb(data)

// For configuration files: indented format
indented, _ := encoding.MarshalJSONIndent(config, "", "  ")

// For debugging: pretty print
pretty := encoding.Pretty(jsonBytes)

// For transmission: minified
minified := encoding.CompactJSON(jsonBytes)
```

✅ **Validate input before unmarshaling**

```go
func SafeUnmarshal(data []byte, v interface{}) error {
    if len(data) == 0 {
        return errors.New("empty JSON data")
    }

    if !json.Valid(data) {
        return errors.New("invalid JSON")
    }

    return encoding.UnmarshalBytes(data, v)
}
```

✅ **Use struct tags for control**

```go
type User struct {
    ID        int    `json:"id"`
    Name      string `json:"name"`
    Password  string `json:"-"`              // Never marshaled
    CreatedAt time.Time `json:"created_at,omitempty"` // Omit if zero
}
```

✅ **Cache OptionsConfig for reuse**

```go
var prettyOptions = &encoding.OptionsConfig{
    Width:    120,
    Indent:   "  ",
    SortKeys: true,
}

// Reuse in multiple places
formatted1 := encoding.PrettyOptions(json1, prettyOptions)
formatted2 := encoding.PrettyOptions(json2, prettyOptions)
```

✅ **Handle edge cases**

```go
// Empty JSON
empty := []byte(`{}`)
pretty := encoding.Pretty(empty) // Safe

// Null values
null := []byte(`null`)
pretty = encoding.Pretty(null) // Safe

// Arrays
array := []byte(`[1,2,3]`)
pretty = encoding.Pretty(array) // Safe
```

### 🔒 Thread Safety

All functions are **thread-safe** as they don't maintain internal state. Safe for concurrent use:

```go
var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        data := getData(id)
        jsonBytes, _ := encoding.MarshalJSONb(data)
        pretty := encoding.Pretty(jsonBytes)
        // Safe concurrent use
    }(i)
}
wg.Wait()
```

### ⚡ Performance Tips

**Fast operations:**

- `MarshalJSONb` - Direct encoding/json wrapper
- `UnmarshalBytes` - Direct encoding/json wrapper
- `UglyInPlace` - Modifies in-place (no allocation)

**Moderate operations:**

- `Ugly` - Creates new buffer
- `Pretty` - Formatting overhead
- `MarshalJSONIndent` - Extra formatting

**Slower operations:**

- `PrettyOptions` with `SortKeys: true` - Sorting overhead

**Optimization strategies:**

```go
// ✅ Reuse buffers
var buf bytes.Buffer
encoder := json.NewEncoder(&buf)
encoder.Encode(data)

// ✅ Skip pretty-printing in production
if !debug {
    w.Write(jsonBytes) // Compact
} else {
    w.Write(encoding.Pretty(jsonBytes)) // Readable
}

// ✅ Use streaming for large data
encoder := json.NewEncoder(w)
encoder.Encode(data)
```

### 🐛 Debugging Tips

**Validate JSON before processing:**

```go
import "encoding/json"

if !json.Valid(data) {
    fmt.Println("Invalid JSON")
}
```

**Pretty-print for inspection:**

```go
fmt.Println("Before:", string(data))
pretty := encoding.Pretty(data)
fmt.Println("After:", string(pretty))
```

**Check types:**

```go
var v interface{}
encoding.UnmarshalBytes(data, &v)
fmt.Printf("Type: %T, Value: %v\n", v, v)
```

### 📝 Testing

Example test cases:

```go
func TestMarshalUnmarshal(t *testing.T) {
    type TestStruct struct {
        Name  string `json:"name"`
        Value int    `json:"value"`
    }

    original := TestStruct{Name: "test", Value: 42}

    // Marshal
    jsonBytes, err := encoding.MarshalJSONb(original)
    if err != nil {
        t.Fatalf("Marshal failed: %v", err)
    }

    // Unmarshal
    var result TestStruct
    err = encoding.UnmarshalBytes(jsonBytes, &result)
    if err != nil {
        t.Fatalf("Unmarshal failed: %v", err)
    }

    if result != original {
        t.Errorf("Got %+v, want %+v", result, original)
    }
}

func TestPrettyFormatting(t *testing.T) {
    compact := []byte(`{"name":"test","value":42}`)
    pretty := encoding.Pretty(compact)

    // Check that output is longer (has whitespace)
    if len(pretty) <= len(compact) {
        t.Error("Pretty output should be longer than compact")
    }

    // Check it's still valid JSON
    var v interface{}
    if err := encoding.UnmarshalBytes(pretty, &v); err != nil {
        t.Errorf("Pretty output is not valid JSON: %v", err)
    }
}
```

## Limitations

- **Not a JSON validator**: Use `json.Valid()` from standard library for validation
- **No streaming support**: For large files, use `json.Encoder`/`Decoder`
- **No custom encoding**: Uses standard `encoding/json` under the hood
- **Pretty-printing overhead**: Adds processing time and memory
- **No comments support**: JSON doesn't support comments by spec

## When to Use Standard Library

Use `encoding/json` directly when you need:

- **Streaming**: `json.Encoder`/`json.Decoder` for large files
- **Custom marshaling**: Implement `json.Marshaler`/`json.Unmarshaler`
- **Raw messages**: `json.RawMessage` for deferred parsing
- **Number precision**: `json.Number` for exact numeric handling

## Contributing

Contributions are welcome! Please see the main [replify repository](https://github.com/polarixa/replify) for contribution guidelines.

## License

This library is part of the [replify](https://github.com/polarixa/replify) project.

## Related

Part of the **replify** ecosystem:

- [replify](https://github.com/polarixa/replify) - API response wrapping library
- [conv](https://github.com/polarixa/replify/pkg/conv) - Type conversion utilities
- [hashy](https://github.com/polarixa/replify/pkg/hashy) - Deterministic hashing
- [match](https://github.com/polarixa/replify/pkg/match) - Wildcard pattern matching
- [coll](https://github.com/polarixa/replify/pkg/coll) - Collection utilities
- [strutil](https://github.com/polarixa/replify/pkg/strutil) - String utilities
- [randn](https://github.com/polarixa/replify/pkg/randn) - Random data generation
