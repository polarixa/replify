# conv

**conv** is a powerful, flexible, and type-safe Go library for converting between different data types. It simplifies type conversions with support for primitives, slices, maps, structs, time values, and more.

## Overview

The `conv` package provides a comprehensive set of functions to convert between Go types in a safe and predictable way. It handles edge cases like `nil` values, empty strings, overflow protection, and supports both strict and lenient conversion modes.

**Key Features:**

- 🔄 **Type-safe conversions** using Go generics (`To[T]`, `Slice[T]`)
- ⚡ **Fast-path optimizations** for common types
- 🛡️ **Overflow protection** for numeric conversions
- 🔧 **Configurable behavior** (strict mode, custom date formats, nil/empty handling)
- 📦 **Rich API** with `Must*`, `*OrDefault`, and error-returning variants
- 🧩 **Extensible** via custom converter interfaces
- ⏱️ **Time parsing** with multiple date format support
- 🗂️ **Collection conversions** (slices, maps, structs)
- 📝 **JSON parsing** helpers with generics

## Use Cases

### When to Use

- ✅ Converting user input (strings) to typed values (int, bool, float)
- ✅ Parsing configuration files or environment variables
- ✅ Working with heterogeneous data from JSON/YAML
- ✅ Building APIs that accept flexible input types
- ✅ Type conversion in database query results
- ✅ Normalizing data from external sources

### When Not to Use

- ❌ When you need complex validation logic (use a validation library instead)
- ❌ When source types are already known at compile time (use type assertions)
- ❌ For performance-critical hot paths with known types (use native conversions)
- ❌ When you need bidirectional serialization (use encoding packages)

## Installation

```bash
go get github.com/polarixa/replify
```

Import the package in your Go code:

```go
import "github.com/polarixa/replify/pkg/conv"
```

## Usage

### Basic Conversions

The package provides simple functions for converting between common types:

```go
package main

import (
    "fmt"
    "github.com/polarixa/replify/pkg/conv"
)

func main() {
    // String to integer
    num, err := conv.Int("42")
    if err != nil {
        panic(err)
    }
    fmt.Println(num) // 42

    // String to boolean
    flag, _ := conv.Bool("true")
    fmt.Println(flag) // true

    // Integer to string
    str, _ := conv.String(12345)
    fmt.Println(str) // "12345"

    // Float to integer (with truncation)
    intVal, _ := conv.Int(3.14)
    fmt.Println(intVal) // 3
}
```

### Generic Type Conversion

Use the generic `To[T]` function for type-safe conversions:

```go
// Convert to specific types using generics
age, err := conv.To[int]("25")
price, err := conv.To[float64]("19.99")
name, err := conv.To[string](12345)
active, err := conv.To[bool]("yes")

// All return the correct type without casting
fmt.Printf("%T: %v\n", age, age)       // int: 25
fmt.Printf("%T: %v\n", price, price)   // float64: 19.99
```

### Must and OrDefault Variants

For cleaner code when you're confident about conversion or want fallbacks:

```go
// Must* functions panic on error (use in initialization)
count := conv.MustInt("100")
timeout := conv.MustDuration("30s")

// *OrDefault functions return a default value on error
port := conv.IntOrDefault(os.Getenv("PORT"), 8080)
debug := conv.BoolOrDefault(os.Getenv("DEBUG"), false)
maxRetry := conv.Int64OrDefault(config["max_retry"], 3)
```

### Type Inference

The `Infer` function automatically converts values to the target type:

```go
var age int
err := conv.Infer(&age, "42")
fmt.Println(age) // 42

var duration time.Duration
err = conv.Infer(&duration, "1h30m")
fmt.Println(duration) // 1h30m0s

var timestamp time.Time
err = conv.Infer(&timestamp, "2024-01-15T10:00:00Z")
fmt.Println(timestamp) // 2024-01-15 10:00:00 +0000 UTC
```

## Examples

### 1. Converting Strings to Numbers

```go
// String to various numeric types
i, _ := conv.Int("123")           // int: 123
i8, _ := conv.Int8("127")         // int8: 127
i16, _ := conv.Int16("32000")     // int16: 32000
i32, _ := conv.Int32("2000000")   // int32: 2000000
i64, _ := conv.Int64("9000000")   // int64: 9000000

// Unsigned integers
u, _ := conv.Uint("12345")        // uint: 12345
u8, _ := conv.Uint8("255")        // uint8: 255
u64, _ := conv.Uint64("9876543210") // uint64: 9876543210

// Floating-point
f32, _ := conv.Float32("3.14")    // float32: 3.14
f64, _ := conv.Float64("2.71828") // float64: 2.71828
```

### 2. Boolean Conversions

```go
// Flexible boolean parsing
conv.Bool("true")   // true
conv.Bool("1")      // true
conv.Bool("yes")    // true
conv.Bool("Y")      // true

conv.Bool("false")  // false
conv.Bool("0")      // false
conv.Bool("no")     // false
conv.Bool("N")      // false

// Numeric to boolean
conv.Bool(1)        // true
conv.Bool(0)        // false
conv.Bool(42)       // true (non-zero)
```

### 3. Time and Duration Conversions

```go
// Parse durations
d1, _ := conv.Duration("2h45m")       // 2h45m0s
d2, _ := conv.Duration("1.5")         // 1.5s (float seconds)
d3, _ := conv.Duration(5000000000)    // 5s (nanoseconds)

// Parse dates/times
t1, _ := conv.Time("2024-01-15T10:00:00Z")
t2, _ := conv.Time("2024-01-15")
t3, _ := conv.Time(1705320000)        // Unix timestamp

// Duration helpers
dur := conv.Seconds(90)               // 1m30s
dur = conv.Minutes(2.5)               // 2m30s
dur = conv.Hours(1.5)                 // 1h30m0s
dur = conv.Days(7)                    // 168h0m0s
```

### 4. Slice Conversions

```go
// Convert to typed slices
ints, _ := conv.IntSlice([]any{"1", 2, 3.0})
// []int{1, 2, 3}

floats, _ := conv.Float64Slice([]any{1, "2.5", 3.14})
// []float64{1.0, 2.5, 3.14}

strs, _ := conv.StringSlice([]any{1, 2, 3})
// []string{"1", "2", "3"}

// Generic slice conversion
nums, _ := conv.Slice[int]([]string{"10", "20", "30"})
// []int{10, 20, 30}

// With default fallback
result := conv.SliceOrDefault[int]("invalid", []int{1, 2, 3})
// []int{1, 2, 3}
```

### 5. Map Conversions

```go
// Struct to map
type User struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

user := User{Name: "Alice", Age: 30}
m, _ := conv.MapTo(user)
// map[string]any{"name": "Alice", "age": 30}

// Typed map conversions
strMap, _ := conv.StringMap(m)
// map[string]string{"name": "Alice", "age": "30"}

intMap, _ := conv.IntMap(map[string]any{"a": "1", "b": 2})
// map[string]int{"a": 1, "b": 2}

boolMap, _ := conv.BoolMap(map[string]any{"active": "true", "debug": 1})
// map[string]bool{"active": true, "debug": true}
```

### 6. JSON Parsing

```go
type Config struct {
    Host string `json:"host"`
    Port int    `json:"port"`
}

jsonStr := `{"host": "localhost", "port": 8080}`

// Parse with error handling
var cfg Config
err := conv.FromJSON(jsonStr, &cfg)

// Generic parsing
cfg, err := conv.ParseJSON[Config](jsonStr)

// Must variant (panics on error)
cfg := conv.MustParseJSON[Config](jsonStr)

// From bytes
data := []byte(jsonStr)
cfg, _ := conv.ParseJSONBytes[Config](data)

// Deep copy using JSON
clone, _ := conv.Clone(cfg)
```

### 7. Custom Converter Configuration

```go
// Create a custom converter
c := conv.NewConverter().
    WithStrictMode(true).               // Error on lossy conversions
    WithTrimStrings(true).              // Trim whitespace
    WithNilAsZero(false).               // Nil returns error
    WithEmptyAsZero(false).             // Empty string returns error
    WithDateFormats("2006-01-02", "02/01/2006")

// Use custom converter
value, err := c.Int("  42  ")           // Uses trimStrings
date, err := c.Time("15/01/2024")       // Uses custom date formats

// Clone and modify
c2 := c.Clone().WithStrictMode(false)
```

## API Reference

### Core Conversion Functions

| Function                        | Description              | Example                         |
| ------------------------------- | ------------------------ | ------------------------------- |
| `To[T](from any)`               | Generic type conversion  | `conv.To[int]("42")`            |
| `MustTo[T](from any)`           | Panics on error          | `conv.MustTo[bool]("true")`     |
| `ToOrDefault[T](from, default)` | Returns default on error | `conv.ToOrDefault[int]("x", 0)` |
| `Infer(&into, from)`            | Infers target type       | `conv.Infer(&age, "25")`        |

### Primitive Types

**Integers:** `Int`, `Int8`, `Int16`, `Int32`, `Int64`, `Uint`, `Uint8`, `Uint16`, `Uint32`, `Uint64`

**Floats:** `Float32`, `Float64`

**Boolean:** `Bool`, `BoolOrDefault`, `MustBool`

**String:** `String`, `StringOrDefault`, `MustString`

**Time:** `Time`, `Duration`, `TimeOrDefault`, `DurationOrDefault`

### Collections

**Slices:**

- `Slice[T](from)` - Generic slice conversion
- `IntSlice`, `Int64Slice`, `Float64Slice`, `StringSlice`, `BoolSlice`
- `SliceOrDefault[T](from, default)`

**Maps:**

- `MapTo(from)` - Convert to `map[string]any`
- `StringMap`, `IntMap`, `Float64Map`, `BoolMap`

### JSON Utilities

- `FromJSON(jsonStr, &target)` - Parse JSON string
- `ParseJSON[T](jsonStr)` - Generic JSON parsing
- `MustParseJSON[T](jsonStr)` - Panics on error
- `Clone[T](value)` - Deep copy via JSON

### Special Functions

- `IsNaN(from)` - Check if value is NaN
- `IsInf(from, sign)` - Check if value is infinite
- `IsFinite(from)` - Check if value is finite
- `IsZeroTime(from)` - Check if time is zero

## Best Practices & Notes

### ⚠️ Common Pitfalls

1. **Overflow Behavior**: Numeric conversions clamp to the target type's range instead of erroring:

   ```go
   val, _ := conv.Int8("200")  // Returns 127 (max int8), not error
   ```

2. **String Trimming**: By default, strings are trimmed before conversion:

   ```go
   conv.Int("  42  ") // Works and returns 42
   ```

3. **Nil Handling**: By default, `nil` returns zero value:

   ```go
   conv.Int(nil) // Returns 0, not error
   ```

4. **Float to Int**: Truncates decimal part without rounding:
   ```go
   conv.Int(3.99) // Returns 3, not 4
   ```

### 💡 Recommendations

✅ **Use `Must*` functions only in safe contexts** (initialization, constants)

✅ **Use `*OrDefault` for environment variables and config**

✅ **Check errors for user input conversions**

✅ **Create custom converters for domain-specific behavior**

✅ **Use generics (`To[T]`, `Slice[T]`) for cleaner code**

✅ **Use `Infer` when target type is known at compile time**

### 🔒 Thread Safety

The default converter is safe for concurrent use. Custom converters should not be modified after being shared between goroutines.

```go
// Safe - default converter is read-only
go conv.Int("42")
go conv.String(123)

// Safe - create once, use many times
c := conv.NewConverter().WithStrictMode(true)
go c.Int("42")
go c.Int("100")

// Unsafe - don't modify after sharing
c.WithStrictMode(false) // ❌ Race condition if used concurrently
```

### ⚡ Performance Tips

- Use type assertions when source type is known at compile time
- Reuse converters instead of creating new ones
- Use `Must*` variants to eliminate error checks when safe
- Prefer specific functions (`Int`, `String`) over generic `To[T]` in hot paths

### 🔧 Customization

Disable automatic trimming and nil handling for stricter behavior:

```go
strict := conv.NewConverter().
    DisableTrimStrings().
    DisableNilAsZero().
    DisableEmptyAsZero()

// Now fails on edge cases
_, err := strict.Int("  42  ")  // Error: whitespace
_, err = strict.Int(nil)        // Error: nil not allowed
_, err = strict.Int("")         // Error: empty string
```

## Error Handling

All conversion errors implement the `ConvError` type:

```go
val, err := conv.Int("invalid")
if err != nil {
    if conv.IsConvError(err) {
        // Handle conversion-specific error
        if convErr, ok := conv.AsConvError(err); ok {
            fmt.Printf("Failed to convert %v to %s\n",
                convErr.From, convErr.To)
        }
    }
}
```

## Contributing

Contributions are welcome! Please see the main [replify repository](https://github.com/polarixa/replify) for contribution guidelines.

## License

This library is part of the [replify](https://github.com/polarixa/replify) project.
