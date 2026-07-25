# strutil

**strutil** is a comprehensive Go string utility library providing 100+ functions for common string manipulation, validation, transformation, and formatting operations. It eliminates boilerplate code and provides a consistent, well-tested API for string handling in Go applications.

## Overview

The `strutil` package offers a rich collection of string utilities that Go's standard library doesn't provide out of the box. It addresses common pain points like:

- **Validation**: Check if strings are empty, blank, numeric, alpha, etc.
- **Transformation**: Convert case, trim, pad, truncate, slugify, and more
- **Analysis**: Count occurrences, check prefixes/suffixes, pattern matching
- **Formatting**: Wrap, repeat, abbreviate, capitalize, title case
- **Manipulation**: Remove, replace, reverse, chop, chomp
- **Comparison**: Case-insensitive checks, contains operations
- **Hashing**: SHA-256 hashing utilities

**Problem Solved:** Writing repetitive string manipulation code is tedious and error-prone. `strutil` provides battle-tested, optimized functions that make string operations readable, maintainable, and consistent across your codebase.

## Use Cases

### When to Use

- ✅ **Input validation** - check for empty, blank, or specific patterns
- ✅ **Data sanitization** - clean user input, normalize whitespace
- ✅ **Text formatting** - title case, sentence case, slugs
- ✅ **String analysis** - check character types, count occurrences
- ✅ **URL/SEO-friendly strings** - slugify names, titles
- ✅ **API responses** - format and validate string data
- ✅ **Configuration parsing** - validate and normalize config values
- ✅ **Template processing** - manipulate strings for templates
- ✅ **Testing** - generate test data, validate outputs

### When Not to Use

- ❌ **Complex parsing** - use `encoding/*` packages for JSON, XML, etc.
- ❌ **Regular expressions** - use `regexp` package directly for complex patterns
- ❌ **Performance-critical hot paths** - use standard library or manual optimization
- ❌ **Internationalization** - use `golang.org/x/text` for i18n/l10n
- ❌ **Binary data** - use `bytes` package for byte operations

## Installation

```bash
go get github.com/polarixa/replify
```

Import the package in your Go code:

```go
import "github.com/polarixa/replify/pkg/strutil"
```

**Requirements:** Go 1.13 or higher

## Usage

### Basic Operations

```go
package main

import (
    "fmt"
    "github.com/polarixa/replify/pkg/strutil"
)

func main() {
    // Check if empty
    isEmpty := strutil.IsEmpty("   ")
    fmt.Println(isEmpty) // true

    // Check if not empty
    isNotEmpty := strutil.IsNotEmpty("hello")
    fmt.Println(isNotEmpty) // true

    // Title case
    title := strutil.Title("hello world")
    fmt.Println(title) // "Hello World"

    // Slugify (URL-friendly)
    slug := strutil.Slugify("Hello World! This is a Test.")
    fmt.Println(slug) // "hello-world-this-is-a-test"

    // Truncate
    short := strutil.Truncate("This is a long sentence", 10)
    fmt.Println(short) // "This is a..."
}
```

## Examples

### 1. Validation and Checking

```go
// Empty and blank checks
strutil.IsEmpty("")           // true
strutil.IsEmpty("   ")        // true
strutil.IsNotEmpty("hello")   // true
strutil.IsBlank("   ")        // true
strutil.IsNotBlank("hello")   // true

// Check any/all empty
strutil.IsAnyEmpty("hello", "", "world")   // true
strutil.IsNoneEmpty("hello", "world")      // true
strutil.IsAllEmpty("", "   ", "")          // true

// Character type checks
strutil.IsAlpha("abc")                 // true
strutil.IsAlpha("abc123")              // false
strutil.IsNumeric("12345")             // true
strutil.IsAlphanumeric("abc123")       // true
strutil.IsWhitespace("   ")            // true

// Case checks
strutil.IsAllLowerCase("hello")        // true
strutil.IsAllUpperCase("HELLO")        // true
```

### 2. String Transformation

```go
// Case conversion
strutil.ToLower("HELLO")               // "hello"
strutil.ToUpper("hello")               // "HELLO"
strutil.Capitalize("hello world")      // "Hello world"
strutil.Title("hello world")           // "Hello World"

// Whitespace handling
strutil.Strip("  hello  ")             // "hello"
strutil.StripStart("  hello  ")        // "hello  "
strutil.StripEnd("  hello  ")          // "  hello"

// Trimming
text := "Hello, World!"
strutil.Trim(text)                     // "Hello, World!"
strutil.TrimLeft("###hello", "#")      // "hello"
strutil.TrimRight("hello###", "#")     // "hello"

// Newline handling
strutil.Chomp("hello\n")               // "hello"
strutil.Chomp("hello\r\n")             // "hello"
strutil.Chop("hello\n")                // "hello"
strutil.Chop("world")                  // "worl"
```

### 3. Prefix and Suffix Operations

```go
// Check prefix/suffix
strutil.StartsWith("Hello", "He")              // true
strutil.EndsWith("Hello", "lo")                // true
strutil.StartsWithIgnoreCase("Hello", "he")    // true
strutil.EndsWithIgnoreCase("Hello", "LO")      // true

// Check multiple
strutil.StartsWithAny("Hello", "Hi", "He")     // true
strutil.EndsWithAny("Hello", "lo", "llo")      // true

// Remove prefix/suffix
strutil.RemoveStart("Hello", "He")             // "llo"
strutil.RemoveEnd("Hello", "lo")               // "Hel"
strutil.RemoveStartIgnoreCase("Hello", "he")   // "llo"
strutil.RemoveEndIgnoreCase("Hello", "LO")     // "Hel"

// Add if missing
strutil.PrependIfMissing("world", "hello ")    // "hello world"
strutil.AppendIfMissing("hello", " world")     // "hello world"
```

### 4. Substring Operations

```go
text := "Hello World"

// Contains checks
strutil.Contains(text, "World")                // true
strutil.ContainsIgnoreCase(text, "world")      // true
strutil.ContainsAny(text, "Foo", "World")      // true
strutil.ContainsNone(text, "xyz", "123")       // true

```

### 5. String Manipulation

```go
// Remove operations
strutil.Remove("hello world", "world")         // "hello "
strutil.RemoveAll("a-b-c", "-")               // "abc"
strutil.RemovePattern("abc123", "[0-9]+")     // "abc"

// Replace operations
strutil.Replace("hello", "l", "L", 1)          // "heLlo"
strutil.ReplaceAll("hello", "l", "L")          // "heLLo"
strutil.ReplaceIgnoreCase("Hello", "hello", "Hi")  // "Hi"

// Reverse operations
strutil.Reverse("hello")                       // "olleh"
strutil.ReverseDelimited("a-b-c", "-")        // "c-b-a"

// Repeat
strutil.Repeat("abc", 3)                       // "abcabcabc"
```

### 6. Formatting and Padding

```go
// Padding
strutil.PadLeft("5", 3, "0")                   // "005"
strutil.PadRight("5", 3, "0")                  // "500"
strutil.Center("Hi", 6, " ")                   // "  Hi  "

// Wrapping
strutil.Wrap("hello", "***")                   // "***hello***"
strutil.Unwrap("***hello***", "***")           // "hello"

// Truncation
strutil.Truncate("Long text here", 8)          // "Long tex..."
strutil.Abbreviate("Long text", 10)            // "Long te..."

// Joining
parts := []string{"hello", "world"}
strutil.JoinUnary(parts, "-")                  // "hello-world"
```

### 7. Slugification and URL-Safe Strings

```go
// Basic slugify
strutil.Slugify("Hello World!")                    // "hello-world"
strutil.SlugifySpec("Hello_World!", "_")       // "hello_world"

// URL-safe strings
title := "10 Tips for Better Code"
slug := strutil.Slugify(title)                     // "10-tips-for-better-code"

// Custom separators
strutil.SlugifySpec("Hello World", "_")         // "hello_world"
```

### 8. Hashing

```go
// SHA-256 hash
hash := strutil.Hash256("password123")
// Returns: "ef92b778bafe771e89245b89ecbc08a44a4e166c06659911881f383d4473e94f"

// Use for checksums, cache keys, etc.
data := "user-123-profile"
cacheKey := strutil.Hash256(data)
```

### 9. Default Values

```go
// Return default if empty
strutil.DefaultIfEmpty("", "default")          // "default"
strutil.DefaultIfEmpty("value", "default")     // "value"

// Return default if blank
strutil.DefaultIfBlank("   ", "default")       // "default"
strutil.DefaultIfBlank("value", "default")     // "value"
```

### 10. Advanced String Analysis

```go
text := "Hello World 123"

// Character class checks
hasDigits := false
for _, c := range text {
    if unicode.IsDigit(c) {
        hasDigits = true
        break
    }
}

// Length (Unicode-aware)
length := strutil.Len("Hello 世界")               // 8 (not byte count)

// Comparison
strutil.Equals("hello", "hello")               // true
strutil.EqualsIgnoreCase("Hello", "hello")     // true
```

### 11. Practical Use Cases

#### Input Validation

```go
func validateUsername(username string) error {
    if strutil.IsBlank(username) {
        return errors.New("username cannot be blank")
    }
    if !strutil.IsAlphanumeric(strings.ReplaceAll(username, "_", "")) {
        return errors.New("username must be alphanumeric")
    }
    if strutil.Len(username) < 3 {
        return errors.New("username must be at least 3 characters")
    }
    return nil
}
```

#### Creating URL Slugs

```go
func createPostSlug(title string) string {
    // "My Awesome Post Title!" -> "my-awesome-post-title"
    slug := strutil.Slugify(title)

    // Ensure maximum length
    if len(slug) > 50 {
        slug = strutil.Truncate(slug, 50)
        slug = strutil.TrimRight(slug, "-")
    }

    return slug
}
```

#### Sanitizing User Input

```go
func sanitizeInput(input string) string {
    // Remove extra whitespace
    cleaned := strings.TrimSpace(input)

    // Normalize multiple spaces to single space
    cleaned = strutil.RegexpDupSpaces.ReplaceAllString(cleaned, " ")

    // Remove newlines
    cleaned = strutil.RemoveAll(cleaned, "\n")
    cleaned = strutil.RemoveAll(cleaned, "\r")

    return cleaned
}
```

#### Formatting Names

```go
func formatName(firstName, lastName string) string {
    // Capitalize each name
    first := strutil.Capitalize(strings.ToLower(firstName))
    last := strutil.Capitalize(strings.ToLower(lastName))

    return first + " " + last
}
```

#### Generating Cache Keys

```go
func generateCacheKey(prefix string, params ...string) string {
    // Create a stable cache key
    combined := prefix + "-" + strutil.Join(params, "-")
    return strutil.Hash256(combined)
}
```

## API Reference

### Validation Functions

| Function                        | Description                            | Example                            |
| ------------------------------- | -------------------------------------- | ---------------------------------- |
| `IsEmpty(s string) bool`        | Check if string is empty or whitespace | `IsEmpty("   ") // true`           |
| `IsNotEmpty(s string) bool`     | Check if string is not empty           | `IsNotEmpty("hi") // true`         |
| `IsBlank(s string) bool`        | Check if string is blank               | `IsBlank("   ") // true`           |
| `IsNotBlank(s string) bool`     | Check if string is not blank           | `IsNotBlank("hi") // true`         |
| `IsAnyEmpty(...string) bool`    | Check if any string is empty           | `IsAnyEmpty("a", "", "b") // true` |
| `IsNoneEmpty(...string) bool`   | Check if none are empty                | `IsNoneEmpty("a", "b") // true`    |
| `IsAllEmpty(...string) bool`    | Check if all are empty                 | `IsAllEmpty("", " ") // true`      |
| `IsAlpha(s string) bool`        | Check if only letters                  | `IsAlpha("abc") // true`           |
| `IsNumeric(s string) bool`      | Check if only digits                   | `IsNumeric("123") // true`         |
| `IsAlphanumeric(s string) bool` | Check if letters and digits            | `IsAlphanumeric("abc123") // true` |
| `IsWhitespace(s string) bool`   | Check if only whitespace               | `IsWhitespace("   ") // true`      |
| `IsAllLowerCase(s string) bool` | Check if all lowercase                 | `IsAllLowerCase("abc") // true`    |
| `IsAllUpperCase(s string) bool` | Check if all uppercase                 | `IsAllUpperCase("ABC") // true`    |

### Transformation Functions

| Function                         | Description                        | Example                           |
| -------------------------------- | ---------------------------------- | --------------------------------- |
| `ToLower(s string) string`       | Convert to lowercase               | `ToLower("HI") // "hi"`           |
| `ToUpper(s string) string`       | Convert to uppercase               | `ToUpper("hi") // "HI"`           |
| `Capitalize(s string) string`    | Capitalize first letter            | `Capitalize("hi") // "Hi"`        |
| `Title(s string) string`         | Title case (each word)             | `Title("hi world") // "Hi World"` |
| `Reverse(s string) string`       | Reverse string                     | `Reverse("abc") // "cba"`         |
| `Strip(s string) string`         | Remove leading/trailing whitespace | `Strip(" hi ") // "hi"`           |
| `StripStart(s string) string`    | Remove leading whitespace          | `StripStart(" hi") // "hi"`       |
| `StripEnd(s string) string`      | Remove trailing whitespace         | `StripEnd("hi ") // "hi"`         |
| `Repeat(s string, n int) string` | Repeat string n times              | `Repeat("a", 3) // "aaa"`         |

### String Analysis

| Function                                      | Description                    | Example                                    |
| --------------------------------------------- | ------------------------------ | ------------------------------------------ |
| `Contains(s, sub string) bool`                | Check if contains substring    | `Contains("hi", "i") // true`              |
| `ContainsIgnoreCase(s, sub string) bool`      | Contains (case-insensitive)    | `ContainsIgnoreCase("Hi", "hi") // true`   |
| `ContainsAny(s string, ...string) bool`       | Contains any substring         | `ContainsAny("hi", "x", "i") // true`      |
| `ContainsNone(s string, ...string) bool`      | Contains none                  | `ContainsNone("hi", "x", "y") // true`     |
| `StartsWith(s, prefix string) bool`           | Starts with prefix             | `StartsWith("hi", "h") // true`            |
| `EndsWith(s, suffix string) bool`             | Ends with suffix               | `EndsWith("hi", "i") // true`              |
| `StartsWithIgnoreCase(s, prefix string) bool` | Starts with (case-insensitive) | `StartsWithIgnoreCase("Hi", "hi") // true` |
| `EndsWithIgnoreCase(s, suffix string) bool`   | Ends with (case-insensitive)   | `EndsWithIgnoreCase("Hi", "HI") // true`   |
| `CountMatches(s, sub string) int`             | Count substring occurrences    | `CountMatches("aaa", "a") // 3`            |
| `IndexOf(s, sub string) int`                  | Find first index               | `IndexOf("hello", "l") // 2`               |
| `LastIndexOf(s, sub string) int`              | Find last index                | `LastIndexOf("hello", "l") // 3`           |

### String Manipulation

| Function                                    | Description             | Example                                |
| ------------------------------------------- | ----------------------- | -------------------------------------- |
| `Remove(s, remove string) string`           | Remove all occurrences  | `Remove("hello", "l") // "heo"`        |
| `RemoveStart(s, prefix string) string`      | Remove prefix           | `RemoveStart("hello", "he") // "llo"`  |
| `RemoveEnd(s, suffix string) string`        | Remove suffix           | `RemoveEnd("hello", "lo") // "hel"`    |
| `Replace(s, old, new string, n int) string` | Replace n occurrences   | `Replace("aaa", "a", "b", 2) // "bba"` |
| `ReplaceAll(s, old, new string) string`     | Replace all             | `ReplaceAll("aaa", "a", "b") // "bbb"` |
| `Chomp(s string) string`                    | Remove trailing newline | `Chomp("hi\n") // "hi"`                |
| `Chop(s string) string`                     | Remove last char        | `Chop("hello") // "hell"`              |
| `Wrap(s, wrapWith string) string`           | Wrap string             | `Wrap("hi", "*") // "*hi*"`            |
| `Truncate(s string, maxLen int) string`     | Truncate with ellipsis  | `Truncate("hello", 3) // "hel..."`     |

### Formatting

| Function                                          | Description                | Example                                      |
| ------------------------------------------------- | -------------------------- | -------------------------------------------- |
| `PadLeft(s string, size int, pad string) string`  | Pad on left                | `PadLeft("5", 3, "0") // "005"`              |
| `PadRight(s string, size int, pad string) string` | Pad on right               | `PadRight("5", 3, "0") // "500"`             |
| `Center(s string, size int, pad string) string`   | Center string              | `Center("hi", 6, " ") // "  hi  "`           |
| `Slugify(s string) string`                        | Create URL slug            | `Slugify("Hi!") // "hi"`                     |
| `SlugifySpec(s, sep string) string`               | Slug with custom separator | `SlugifySpec("Hi", "_") // "hi"`             |
| `Abbreviate(s string, maxLen int) string`         | Abbreviate string          | `Abbreviate("hello world", 8) // "hello..."` |

### Utilities

| Function                                  | Description               | Example                                 |
| ----------------------------------------- | ------------------------- | --------------------------------------- |
| `Len(s string) int`                       | UTF-8 character count     | `Len("hello") // 5`                     |
| `Hash256(s string) string`                | SHA-256 hash              | `Hash256("text")`                       |
| `Join(slice []string, sep string) string` | Join with separator       | `Join([]string{"a","b"}, ",") // "a,b"` |
| `DefaultIfEmpty(s, def string) string`    | Return default if empty   | `DefaultIfEmpty("", "x") // "x"`        |
| `DefaultIfBlank(s, def string) string`    | Return default if blank   | `DefaultIfBlank(" ", "x") // "x"`       |
| `Equals(a, b string) bool`                | String equality           | `Equals("hi", "hi") // true`            |
| `EqualsIgnoreCase(a, b string) bool`      | Case-insensitive equality | `EqualsIgnoreCase("Hi", "hi") // true`  |

### Global Variables

- `Len` - Alias for `utf8.RuneCountInString` (Unicode-aware length)
- `RegexpDupSpaces` - Compiled regex for matching duplicate spaces
- `MaxRuneBytes` - Maximum valid UTF-8 encoding bytes

## Best Practices & Notes

### ⚠️ Common Pitfalls

1. **Unicode vs Byte Length**

   ```go
   // ❌ Wrong: len() counts bytes
   text := "Hello 世界"
   length := len(text) // 13 bytes

   // ✅ Correct: Use Len() for Unicode characters
   length := strutil.Len(text) // 8 characters
   ```

2. **Empty vs Blank**

   ```go
   // IsEmpty: checks for empty or whitespace
   strutil.IsEmpty("")       // true
   strutil.IsEmpty("   ")    // true

   // IsBlank: same as IsEmpty
   strutil.IsBlank("   ")    // true

   // IsNotEmpty: opposite
   strutil.IsNotEmpty("   ") // false
   ```

3. **Case-Sensitive Operations**

   ```go
   // ❌ Case-sensitive by default
   strutil.Contains("Hello", "hello") // false

   // ✅ Use case-insensitive variant
   strutil.ContainsIgnoreCase("Hello", "hello") // true
   ```

4. **Immutability**
   ```go
   // All functions return new strings, originals unchanged
   original := "hello"
   upper := strutil.ToUpper(original)
   // original is still "hello"
   // upper is "HELLO"
   ```

### 💡 Recommendations

✅ **Use semantic names** for clarity

```go
// ✅ Clear intent
isEmpty := strutil.IsEmpty(userInput)
if isEmpty {
    return errors.New("input required")
}
```

✅ **Chain operations** for complex transformations

```go
// Clean and format user input
result := strutil.Strip(input)
result = strutil.ToLower(result)
result = strutil.Slugify(result)
```

✅ **Validate early** in your application flow

```go
func ProcessUser(name, email string) error {
    if strutil.IsAnyEmpty(name, email) {
        return errors.New("name and email required")
    }
    // ... continue processing
}
```

✅ **Use appropriate functions** for the task

```go
// For URLs and slugs
slug := strutil.Slugify(title)

// For user display
formatted := strutil.Title(name)

// For comparison
if strutil.EqualsIgnoreCase(status, "active") {
    // ...
}
```

✅ **Cache hash results** when used frequently

```go
// Cache expensive hash operations
type User struct {
    Email     string
    EmailHash string // Cached hash
}

func NewUser(email string) *User {
    return &User{
        Email:     email,
        EmailHash: strutil.Hash256(email),
    }
}
```

### 🔒 Thread Safety

All functions are **thread-safe** as they don't maintain internal state. They can be safely used in concurrent goroutines.

```go
var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        slug := strutil.Slugify(fmt.Sprintf("Title %d", id))
        // Safe concurrent access
    }(i)
}
wg.Wait()
```

### ⚡ Performance Tips

**Fast operations:**

- ✅ `IsEmpty`, `IsNotEmpty` - O(1) after trim
- ✅ `StartsWith`, `EndsWith` - O(n) where n = prefix/suffix length
- ✅ `ToLower`, `ToUpper` - O(n)

**Moderate operations:**

- ⚠️ `Slugify` - Multiple regex operations
- ⚠️ `Title` - Iterates through words
- ⚠️ `Remove`, `Replace` - O(n) per occurrence

**Expensive operations:**

- 🐌 `Hash256` - Cryptographic hash (use for security, not performance)
- 🐌 `Reverse` - O(n) with UTF-8 handling

**Optimization strategies:**

1. **Reuse compiled regexes** - Use `RegexpDupSpaces` global
2. **Avoid unnecessary conversions** - Don't `ToLower` multiple times
3. **Batch operations** - Combine multiple string ops
4. **Cache results** - Store computed values (hashes, slugs)

### 🐛 Debugging Tips

**Print intermediate results:**

```go
input := " Hello World! "
fmt.Printf("Original: %q\n", input)
stripped := strutil.Strip(input)
fmt.Printf("Stripped: %q\n", stripped)
slug := strutil.Slugify(stripped)
fmt.Printf("Slug: %q\n", slug)
```

**Check for empty strings:**

```go
if strutil.IsEmpty(result) {
    log.Printf("Warning: got empty result from %q", input)
}
```

**Verify transformations:**

```go
original := "Test String"
transformed := strutil.Slugify(original)
fmt.Printf("%s -> %s\n", original, transformed)
```

### 📝 Testing

Example test cases:

```go
func TestSlugify(t *testing.T) {
    tests := []struct {
        input    string
        expected string
    }{
        {"Hello World", "hello-world"},
        {"Hello  World", "hello-world"},
        {"Hello-World!", "hello-world"},
        {"", ""},
        {"   ", ""},
        {"CamelCase", "camelcase"},
        {"with_underscore", "with-underscore"},
    }

    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            got := strutil.Slugify(tt.input)
            if got != tt.expected {
                t.Errorf("Slugify(%q) = %q, want %q",
                    tt.input, got, tt.expected)
            }
        })
    }
}
```

### 🔍 Common Patterns

**Validation middleware:**

```go
func ValidateRequest(r *http.Request) error {
    name := r.FormValue("name")
    email := r.FormValue("email")

    if strutil.IsAnyEmpty(name, email) {
        return errors.New("name and email required")
    }

    if !strutil.ContainsIgnoreCase(email, "@") {
        return errors.New("invalid email format")
    }

    return nil
}
```

**Text normalization:**

```go
func NormalizeText(text string) string {
    // Remove extra whitespace
    text = strutil.Strip(text)
    text = strutil.RegexpDupSpaces.ReplaceAllString(text, " ")

    // Remove newlines
    text = strutil.RemoveAll(text, "\n")
    text = strutil.RemoveAll(text, "\r")

    return text
}
```

## Limitations

- **No internationalization** - Use `golang.org/x/text` for i18n/l10n
- **Basic pattern matching** - Use `regexp` for complex patterns
- **ASCII-focused slugify** - May not handle all Unicode correctly
- **No streaming operations** - All operations load full strings into memory
- **No locale-aware operations** - Case conversion uses Go's defaults

## Performance Characteristics

| Operation Type   | Time Complexity | Memory |
| ---------------- | --------------- | ------ |
| Validation       | O(n)            | O(1)   |
| Case conversion  | O(n)            | O(n)   |
| Substring search | O(n×m)          | O(1)   |
| Trim/Strip       | O(n)            | O(n)   |
| Replace          | O(n)            | O(n)   |
| Slugify          | O(n)            | O(n)   |
| Hash256          | O(n)            | O(1)   |

Where n = string length, m = substring length

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
- Other sivaosorg utilities

---

**Note:** This package contains 100+ utility functions. For a complete API reference, please visit the [strutil package on GitHub](https://github.com/polarixa/replify/tree/master/pkg/strutil).
