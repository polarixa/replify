# match

**match** is a lightweight, high-performance Go library for pattern matching with wildcard support. It provides fast string matching using `*` (any sequence) and `?` (any single character) wildcards, with optional complexity limits to prevent algorithmic complexity attacks.

## Overview

The `match` package implements efficient wildcard pattern matching for strings, similar to shell glob patterns but optimized for Go applications. It's designed to be simple, fast, and safe against patterns that could cause exponential time complexity.

**Key Features:**

- 🚀 **Fast wildcard matching** - optimized algorithm with minimal allocations
- 🎯 **Simple pattern syntax** - uses familiar `*` and `?` wildcards
- 🛡️ **Complexity protection** - optional limits to prevent DoS via complex patterns
- ✅ **UTF-8 support** - correctly handles Unicode characters
- 🔧 **Escape sequences** - support for literal `*`, `?`, and `\` characters
- 📊 **Pattern analysis** - calculate min/max bounds for patterns
- 🎨 **Zero dependencies** - only relies on Go standard library (and `strutil`)

**Use Case:** Perfect for filtering, routing, access control, file matching, and any scenario requiring flexible string pattern matching without regular expressions.

## Use Cases

### When to Use

- ✅ **File path matching** - match file names or paths with wildcards
- ✅ **API routing** - simple wildcard-based route matching
- ✅ **Access control** - match resource names against permission patterns
- ✅ **Configuration filters** - allow users to specify flexible filters
- ✅ **Log filtering** - match log entries against patterns
- ✅ **Data validation** - check if strings conform to expected patterns
- ✅ **User input matching** - search/filter with wildcards (e.g., `*smith`, `john?`)
- ✅ **Simple glob patterns** - replace complex regex with readable wildcards

### When Not to Use

- ❌ **Complex pattern matching** - use `regexp` for advanced patterns (groups, lookahead, etc.)
- ❌ **Regular expressions needed** - when you need character classes like `[a-z]` or `\d+`
- ❌ **Case-insensitive matching** - this library is case-sensitive (convert to lowercase first)
- ❌ **Performance-critical exact matching** - use `==` or `strings.Contains()` instead
- ❌ **Path/file-specific logic** - use `filepath.Match()` for file path patterns with OS-specific behavior

## Installation

```bash
go get github.com/polarixa/replify
```

Import the package in your Go code:

```go
import "github.com/polarixa/replify/pkg/match"
```

## Usage

### Basic Pattern Matching

The simplest way to match strings against patterns:

```go
package main

import (
    "fmt"
    "github.com/polarixa/replify/pkg/match"
)

func main() {
    // Match with wildcards
    fmt.Println(match.Match("hello", "h*o"))        // true
    fmt.Println(match.Match("hello", "h?llo"))      // true
    fmt.Println(match.Match("hello", "*"))          // true (matches everything)
    fmt.Println(match.Match("hello", "world"))      // false

    // Exact match (no wildcards)
    fmt.Println(match.Match("hello", "hello"))      // true
    fmt.Println(match.Match("hello", "Hello"))      // false (case-sensitive)
}
```

### Wildcard Syntax

**Supported wildcards:**

| Pattern | Description                    | Example | Matches          | Doesn't Match |
| ------- | ------------------------------ | ------- | ---------------- | ------------- |
| `*`     | Any sequence (including empty) | `a*`    | `a`, `ab`, `abc` | `ba`          |
| `?`     | Exactly one character          | `a?c`   | `abc`, `axc`     | `ac`, `abcd`  |
| `\*`    | Literal asterisk               | `a\*b`  | `a*b`            | `ab`, `aXb`   |
| `\?`    | Literal question mark          | `a\?b`  | `a?b`            | `ab`, `axb`   |
| `\\`    | Literal backslash              | `a\\b`  | `a\b`            | `ab`          |

### Pattern Examples

```go
// Match email patterns
match.Match("user@example.com", "*@*.com")     // true
match.Match("admin@test.org", "*@*.com")       // false

// Match file extensions
match.Match("document.pdf", "*.pdf")           // true
match.Match("image.jpg", "*.pdf")              // false

// Match prefixes
match.Match("production-server-01", "prod*")   // true (starts with "prod")
match.Match("staging-server-01", "prod*")      // false

// Match suffixes
match.Match("error.log", "*.log")              // true (ends with ".log")
match.Match("data.csv", "*.log")               // false

// Match with multiple wildcards
match.Match("user-123-admin", "user-*-*")      // true
match.Match("2024-01-15", "????-??-??")        // true (10 characters)

// Escape special characters
match.Match("price*tax", `price\*tax`)         // true (literal *)
match.Match("what?", `what\?`)                 // true (literal ?)
```

## Examples

### 1. File Name Matching

```go
func matchFiles(filename string, patterns []string) bool {
    for _, pattern := range patterns {
        if match.Match(filename, pattern) {
            return true
        }
    }
    return false
}

func main() {
    patterns := []string{"*.go", "*.txt", "README*"}

    fmt.Println(matchFiles("main.go", patterns))        // true
    fmt.Println(matchFiles("README.md", patterns))      // true
    fmt.Println(matchFiles("config.json", patterns))    // false
}
```

### 2. Permission Checking

```go
type Permission struct {
    Resource string
    Patterns []string
}

func hasAccess(resource string, permissions []Permission) bool {
    for _, perm := range permissions {
        for _, pattern := range perm.Patterns {
            if match.Match(resource, pattern) {
                return true
            }
        }
    }
    return false
}

func main() {
    permissions := []Permission{
        {Resource: "admin", Patterns: []string{"/admin/*", "/users/*/edit"}},
        {Resource: "public", Patterns: []string{"/public/*", "/images/*"}},
    }

    fmt.Println(hasAccess("/admin/dashboard", permissions))  // true
    fmt.Println(hasAccess("/users/123/edit", permissions))   // true
    fmt.Println(hasAccess("/secret/data", permissions))      // false
}
```

### 3. Complexity-Limited Matching (DoS Protection)

```go
// Protect against patterns that could cause exponential time complexity
func safeMatch(text, pattern string) (bool, error) {
    // Limit complexity to prevent ReDoS-style attacks
    matched, stopped := match.MatchLimit(text, pattern, 10000)

    if stopped {
        return false, fmt.Errorf("pattern too complex: %s", pattern)
    }

    return matched, nil
}

func main() {
    // Normal pattern
    matched, err := safeMatch("hello world", "h*o w*d")
    fmt.Println(matched, err) // true, nil

    // Complex pattern that would take too long
    text := strings.Repeat("a", 1000)
    pattern := strings.Repeat("*a", 500) + "*x"
    matched, err = safeMatch(text, pattern)
    // Returns: false, "pattern too complex: ..."
}
```

### 4. Route Matching for HTTP

```go
type Route struct {
    Pattern string
    Handler string
}

func matchRoute(path string, routes []Route) (string, bool) {
    for _, route := range routes {
        if match.Match(path, route.Pattern) {
            return route.Handler, true
        }
    }
    return "", false
}

func main() {
    routes := []Route{
        {Pattern: "/api/users/*", Handler: "UserHandler"},
        {Pattern: "/api/posts/*/comments", Handler: "CommentHandler"},
        {Pattern: "/static/*", Handler: "StaticHandler"},
        {Pattern: "*", Handler: "NotFoundHandler"}, // Catch-all
    }

    handler, ok := matchRoute("/api/users/123", routes)
    fmt.Println(handler, ok) // "UserHandler", true

    handler, ok = matchRoute("/unknown/path", routes)
    fmt.Println(handler, ok) // "NotFoundHandler", true
}
```

### 5. Log Entry Filtering

```go
type LogFilter struct {
    Include []string
    Exclude []string
}

func shouldLogEntry(message string, filter LogFilter) bool {
    // Check exclusions first
    for _, pattern := range filter.Exclude {
        if match.Match(message, pattern) {
            return false
        }
    }

    // Check inclusions
    if len(filter.Include) == 0 {
        return true // No filter means include all
    }

    for _, pattern := range filter.Include {
        if match.Match(message, pattern) {
            return true
        }
    }

    return false
}

func main() {
    filter := LogFilter{
        Include: []string{"ERROR*", "WARN*"},
        Exclude: []string{"*test*", "*debug*"},
    }

    fmt.Println(shouldLogEntry("ERROR: database connection failed", filter))  // true
    fmt.Println(shouldLogEntry("WARN: slow query detected", filter))          // true
    fmt.Println(shouldLogEntry("INFO: request processed", filter))            // false
    fmt.Println(shouldLogEntry("ERROR: test failed", filter))                 // false (excluded)
}
```

### 6. Pattern Bounds Analysis

```go
// Calculate min/max possible values for a pattern
func analyzePattern(pattern string) {
    min, max := match.WildcardPatternLimits(pattern)

    fmt.Printf("Pattern: %s\n", pattern)
    fmt.Printf("Min: %q\n", min)
    fmt.Printf("Max: %q\n", max)
    fmt.Println()
}

func main() {
    analyzePattern("user-*")
    // Pattern: user-*
    // Min: "user-"
    // Max: "user-\xf4\x8f\xbf\xc0" (next possible value)

    analyzePattern("a?c")
    // Pattern: a?c
    // Min: "a\x00c"
    // Max: "a\xf4\x8f\xbf\xbfc"

    analyzePattern("*")
    // Pattern: *
    // Min: ""
    // Max: ""

    analyzePattern("test")
    // Pattern: test
    // Min: "test"
    // Max: "test"
}
```

### 7. Case-Insensitive Matching

```go
// Match ignoring case by converting to lowercase
func matchIgnoreCase(text, pattern string) bool {
    return match.Match(
        strings.ToLower(text),
        strings.ToLower(pattern),
    )
}

func main() {
    fmt.Println(matchIgnoreCase("Hello World", "hello*"))    // true
    fmt.Println(matchIgnoreCase("HELLO", "h?llo"))           // true
    fmt.Println(matchIgnoreCase("TeSt", "test"))             // true
}
```

### 8. Batch Matching

```go
// Check if any pattern matches
func matchAny(text string, patterns []string) bool {
    for _, pattern := range patterns {
        if match.Match(text, pattern) {
            return true
        }
    }
    return false
}

// Check if all patterns match
func matchAll(text string, patterns []string) bool {
    for _, pattern := range patterns {
        if !match.Match(text, pattern) {
            return false
        }
    }
    return true
}

func main() {
    text := "error_log_2024.txt"

    // Any match
    fmt.Println(matchAny(text, []string{"*.log", "*.txt"}))      // true
    fmt.Println(matchAny(text, []string{"*.pdf", "*.doc"}))      // false

    // All match
    fmt.Println(matchAll(text, []string{"error*", "*.txt"}))     // true
    fmt.Println(matchAll(text, []string{"error*", "*.log"}))     // false
}
```

## API Reference

### Core Functions

#### `Match(str, pattern string) bool`

Basic wildcard pattern matching.

**Parameters:**

- `str` - The input string to match
- `pattern` - Pattern with wildcards (`*`, `?`)

**Returns:**

- `true` if the string matches the pattern

**Example:**

```go
match.Match("hello", "h*o")  // true
```

---

#### `MatchLimit(str, pattern string, maxComplexity int) (matched, stopped bool)`

Pattern matching with complexity limit to prevent algorithmic attacks.

**Parameters:**

- `str` - The input string to match
- `pattern` - Pattern with wildcards
- `maxComplexity` - Maximum allowed complexity (recommended: 10000)

**Returns:**

- `matched` - `true` if string matches pattern within complexity limit
- `stopped` - `true` if complexity limit was exceeded

**Example:**

```go
matched, stopped := match.MatchLimit("test", "t*t", 10000)
if stopped {
    // Pattern was too complex
}
```

---

#### `WildcardPatternLimits(pattern string) (min, max string)`

Calculate the minimum and maximum possible string values that could match a pattern.

**Parameters:**

- `pattern` - Pattern with wildcards

**Returns:**

- `min` - Minimum possible matching string
- `max` - Maximum possible matching string (exclusive upper bound)

**Example:**

```go
min, max := match.WildcardPatternLimits("user-*")
// min: "user-"
// max: "user-\xf4\x8f\xbf\xc0"
```

**Use case:** Range queries in databases or sorted data structures.

---

### Pattern Syntax Reference

| Element | Matches             | Example Pattern | Matches          | Doesn't Match    |
| ------- | ------------------- | --------------- | ---------------- | ---------------- |
| Literal | Exact character     | `hello`         | `hello`          | `Hello`, `helo`  |
| `*`     | 0+ characters       | `h*o`           | `ho`, `hello`    | `hey`            |
| `?`     | Exactly 1 character | `h?llo`         | `hello`, `hallo` | `hllo`, `helllo` |
| `\*`    | Literal `*`         | `a\*b`          | `a*b`            | `ab`             |
| `\?`    | Literal `?`         | `a\?b`          | `a?b`            | `ab`             |
| `\\`    | Literal `\`         | `a\\b`          | `a\b`            | `ab`             |

**Pattern Rules:**

- Patterns are **case-sensitive**
- Multiple `*` in a row are treated as one
- `*` at the end always matches
- Empty pattern only matches empty string
- Pattern `*` matches everything

## Best Practices & Notes

### ⚠️ Common Pitfalls

1. **Case Sensitivity**: Matching is case-sensitive by default

   ```go
   match.Match("Hello", "hello")  // false
   // Solution: Convert both to lowercase
   match.Match(strings.ToLower("Hello"), strings.ToLower("hello"))  // true
   ```

2. **Pattern Order Matters**: When matching against multiple patterns, order them from specific to general

   ```go
   // ❌ Bad: catch-all comes first
   patterns := []string{"*", "*.go"}

   // ✅ Good: specific patterns first
   patterns := []string{"*.go", "*.txt", "*"}
   ```

3. **Escaping Backslashes**: Remember to escape backslashes in Go strings

   ```go
   // ❌ Wrong: single backslash
   match.Match("a*b", "a\*b")  // Won't work as expected

   // ✅ Correct: double backslash (or raw string)
   match.Match("a*b", `a\*b`)  // Works
   ```

4. **UTF-8 Handling**: The library correctly handles multi-byte UTF-8 characters

   ```go
   match.Match("café", "caf?")  // true - é is one character
   ```

5. **Empty Patterns**: Be explicit about empty pattern behavior
   ```go
   match.Match("", "")      // true
   match.Match("test", "")  // false
   match.Match("", "*")     // true
   ```

### 💡 Recommendations

✅ **Use `MatchLimit` for user input** to prevent DoS attacks via complex patterns

✅ **Validate patterns** before using them in production (ensure they're not too complex)

✅ **Cache match results** if matching the same pattern repeatedly

✅ **Use specific patterns** when possible (avoid `*` at the start for better performance)

✅ **Document pattern syntax** if exposing to end users

✅ **Test edge cases** like empty strings, Unicode, and special characters

### 🔒 Security Considerations

**Algorithmic Complexity Attack:**

Certain patterns can cause exponential time complexity:

```go
// ❌ Dangerous: nested wildcards with user input
pattern := strings.Repeat("*a", 100) + "*x"
text := strings.Repeat("a", 1000)
// This could take very long!

// ✅ Safe: Use MatchLimit
matched, stopped := match.MatchLimit(text, pattern, 10000)
if stopped {
    return fmt.Errorf("pattern rejected: too complex")
}
```

**Recommended `maxComplexity` values:**

- `10000` - Good default for general use
- `100000` - For trusted patterns
- `1000` - For highly untrusted input

**Formula:** Complexity ≈ `maxComplexity * len(str)`

### ⚡ Performance Tips

**Fast patterns:**

- ✅ No wildcards: `O(n)` - fastest
- ✅ `*` at end: `O(n)` - very fast
- ✅ Single `*`: `O(n)` - fast

**Slow patterns:**

- ⚠️ `*` at start: `O(n²)` - slower
- ⚠️ Multiple `*`: `O(n²)` - can be slow
- 🐌 Nested wildcards: `O(2ⁿ)` - potentially exponential

**Optimization tips:**

1. **Avoid leading wildcards** when possible
2. **Use exact matches** when you don't need wildcards
3. **Limit pattern complexity** for user-provided patterns
4. **Pre-compile patterns** (validate once, use many times)

### 🐛 Debugging

If matches aren't working as expected:

1. **Print the pattern and string:**

   ```go
   fmt.Printf("Pattern: %q\nString: %q\n", pattern, str)
   ```

2. **Check for hidden characters:**

   ```go
   fmt.Printf("Pattern bytes: %v\n", []byte(pattern))
   ```

3. **Test incrementally:**

   ```go
   match.Match("test", "t*")    // true
   match.Match("test", "te*")   // true
   match.Match("test", "tes*")  // true
   match.Match("test", "test")  // true
   ```

4. **Verify escaping:**
   ```go
   // These are different:
   match.Match("a*b", `a*b`)    // Matches a, ab, aXXXb (wildcard)
   match.Match("a*b", `a\*b`)   // Only matches literal "a*b"
   ```

### 📊 Pattern Analysis

Use `WildcardPatternLimits` for:

- **Range queries**: Find database records matching pattern
- **Sorting**: Determine where pattern matches would appear
- **Optimization**: Skip impossible matches early

```go
min, max := match.WildcardPatternLimits("user-2024-*")
// Query: WHERE username >= min AND username < max
```

### 🧪 Testing

Recommended test cases for your patterns:

```go
func TestPatternMatching(t *testing.T) {
    tests := []struct {
        str     string
        pattern string
        want    bool
    }{
        {"", "", true},
        {"", "*", true},
        {"test", "", false},
        {"test", "test", true},
        {"test", "t*", true},
        {"test", "*t", true},
        {"test", "t?st", true},
        {"test", "t??t", true},
        {"café", "caf?", true},
        {"a*b", `a\*b`, true},
    }

    for _, tt := range tests {
        got := match.Match(tt.str, tt.pattern)
        if got != tt.want {
            t.Errorf("Match(%q, %q) = %v, want %v",
                tt.str, tt.pattern, got, tt.want)
        }
    }
}
```

## Limitations

- **No character classes**: Use `regexp` for patterns like `[a-z]` or `\d+`
- **No alternation**: Cannot express "A or B" (use multiple patterns)
- **No negation**: Cannot match "not this" (filter results instead)
- **Case-sensitive**: Convert to lowercase for case-insensitive matching
- **No anchors**: Patterns always match the entire string (not substrings)

## Performance

Typical performance characteristics:

| Pattern Type | Time Complexity   | Example  |
| ------------ | ----------------- | -------- |
| Exact match  | O(n)              | `hello`  |
| Trailing `*` | O(n)              | `hello*` |
| Leading `*`  | O(n×m)            | `*world` |
| Multiple `*` | O(n×m) worst-case | `h*o*d`  |

Where `n` = string length, `m` = pattern length.

## Contributing

Contributions are welcome! Please see the main [replify repository](https://github.com/polarixa/replify) for contribution guidelines.

## License

This library is part of the [replify](https://github.com/polarixa/replify) project.
