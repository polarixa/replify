# truncate

**truncate** is a flexible, OOP-style Go string truncation library that supports multiple truncation strategies, customisable omission markers, and positional control (start, middle, end). It provides both a fluent builder API and ready-to-use strategy types for common truncation patterns.

## Overview

The `truncate` package solves the problem of shortening strings to a target length while preserving readability. Unlike a simple slice operation it handles:

- **Omission markers** — inserting `…` or `...` to indicate removed content
- **Positional control** — truncate from the end, start, or middle of a string
- **Unicode correctness** — operates on runes, not bytes, so multi-byte characters are never split
- **Strategy pattern** — swap truncation behaviour through a common `Strategy` interface

**Key Features:**

- ✂️ **Multiple built-in strategies** — Cut, CutEllipsis, CutEllipsisLeading, EllipsisMiddle
- 🔧 **Fluent builder API** — configure omission, position, and max length with chaining
- 🌐 **Unicode-safe** — correctly handles CJK, emoji, and accented characters
- ⚡ **Zero dependencies** — uses only the Go standard library
- 🧩 **Strategy interface** — implement your own truncation logic and plug it in
- 🔒 **Thread-safe** — all operations are stateless or use immutable configuration

## Use Cases

### When to Use

- ✅ **UI text truncation** — shorten titles, descriptions, or labels for display
- ✅ **Log messages** — cap long values in structured logging
- ✅ **CLI output** — fit strings into fixed-width terminal columns
- ✅ **Notification previews** — generate truncated message previews
- ✅ **Database fields** — ensure strings fit column width constraints
- ✅ **API responses** — limit string field lengths in payloads
- ✅ **File path display** — shorten long paths with middle truncation

### When Not to Use

- ❌ **Word-boundary truncation** — this package operates on runes, not words
- ❌ **HTML-aware truncation** — use a dedicated HTML-safe truncator instead
- ❌ **Binary data** — use `bytes` package for byte-level operations

## Installation

```bash
go get github.com/polarixa/replify
```

Import the package in your Go code:

```go
import "github.com/polarixa/replify/pkg/truncate"
```

## Usage

### Quick Start with Strategies

The simplest way to truncate strings is using the built-in strategy types:

```go
package main

import (
    "fmt"
    "github.com/polarixa/replify/pkg/truncate"
)

func main() {
    text := "Hello, World! This is a long string."

    // Cut without any marker
    fmt.Println(truncate.Apply(text, 10, truncate.NewCutStrategy()))
    // Output: "Hello, Wor"

    // Cut with ellipsis at the end
    fmt.Println(truncate.Apply(text, 10, truncate.NewCutEllipsisStrategy()))
    // Output: "Hello, Wo…"

    // Cut with ellipsis at the start
    fmt.Println(truncate.Apply(text, 10, truncate.NewCutEllipsisLeadingStrategy()))
    // Output: "…g string."

    // Ellipsis in the middle
    fmt.Println(truncate.Apply(text, 10, truncate.NewEllipsisMiddleStrategy()))
    // Output: "Hell…ing."
}
```

### Using the Builder API

For full control, build a `Truncator` with the fluent builder:

```go
package main

import (
    "fmt"
    "github.com/polarixa/replify/pkg/truncate"
)

func main() {
    // Build a reusable truncator
    t := truncate.NewTruncator().
        WithOmission("...").
        WithPosition(truncate.PositionEnd).
        WithMaxLength(15).
        Build()

    fmt.Println(t.Truncate("The quick brown fox jumps over the lazy dog"))
    // Output: "The quick br..."

    // Override length on-the-fly
    fmt.Println(t.TruncateWithLength("The quick brown fox", 10))
    // Output: "The qui..."
}
```

## Examples

### 1. UI Label Truncation

```go
func formatLabel(label string, maxWidth int) string {
    t := truncate.NewTruncator().
        WithOmission("…").
        WithPosition(truncate.PositionEnd).
        WithMaxLength(maxWidth).
        Build()
    return t.Truncate(label)
}

func main() {
    fmt.Println(formatLabel("Dashboard Settings", 12)) // "Dashboard S…"
    fmt.Println(formatLabel("Home", 12))                // "Home"
}
```

### 2. File Path Display (Middle Truncation)

```go
func shortenPath(path string, maxLen int) string {
    t := truncate.NewTruncator().
        WithOmission("…").
        WithPosition(truncate.PositionMiddle).
        WithMaxLength(maxLen).
        Build()
    return t.Truncate(path)
}

func main() {
    path := "/Users/alice/Documents/projects/myapp/src/main.go"
    fmt.Println(shortenPath(path, 30))
    // Output: "/Users/alice/D…rc/main.go"
}
```

### 3. Log Message Capping

```go
func capLogValue(value string) string {
    return truncate.Apply(value, 80, truncate.NewCutEllipsisStrategy())
}

func main() {
    longJSON := `{"user":"alice","data":"` + strings.Repeat("x", 200) + `"}`
    fmt.Println(capLogValue(longJSON))
    // Truncates to 80 runes with "…" at the end
}
```

### 4. Notification Preview

```go
func notificationPreview(body string) string {
    t := truncate.NewTruncator().
        WithOmission("…").
        WithPosition(truncate.PositionEnd).
        WithMaxLength(50).
        Build()
    return t.Truncate(body)
}

func main() {
    msg := "Your order #12345 has been shipped and is expected to arrive by Friday, February 14th."
    fmt.Println(notificationPreview(msg))
    // Output: "Your order #12345 has been shipped and is expec…"
}
```

### 5. Custom Strategy Implementation

Implement the `Strategy` interface for custom truncation logic:

```go
// WordBoundaryStrategy truncates at the last space before the limit.
type WordBoundaryStrategy struct {
    Omission string
}

func (w WordBoundaryStrategy) Truncate(str string, length int) string {
    if len([]rune(str)) <= length {
        return str
    }
    r := []rune(str)
    truncated := string(r[:length])
    // Find last space
    if idx := strings.LastIndex(truncated, " "); idx > 0 {
        truncated = truncated[:idx]
    }
    return truncated + w.Omission
}

func main() {
    strategy := WordBoundaryStrategy{Omission: "…"}
    result := truncate.Apply("The quick brown fox jumps over the lazy dog", 20, strategy)
    fmt.Println(result) // "The quick brown fox…"
}
```

### 6. Unicode and Emoji Support

```go
func main() {
    // CJK characters
    cjk := "日本語テスト文字列"
    fmt.Println(truncate.Apply(cjk, 5, truncate.NewCutEllipsisStrategy()))
    // Output: "日本語テ…"

    // Emoji
    emoji := "Hello 👋 World 🌍!"
    fmt.Println(truncate.Apply(emoji, 10, truncate.NewCutEllipsisStrategy()))
    // Output: "Hello 👋 W…"

    // Accented characters
    accented := "Héllo, café résumé"
    fmt.Println(truncate.Apply(accented, 10, truncate.NewCutEllipsisStrategy()))
    // Output: "Héllo, ca…"
}
```

### 7. Strategy Comparison

```go
func main() {
    text := "Hello, World!"
    length := 8

    strategies := map[string]truncate.Strategy{
        "Cut":            truncate.NewCutStrategy(),
        "CutEllipsis":    truncate.NewCutEllipsisStrategy(),
        "LeadingEllipsis": truncate.NewCutEllipsisLeadingStrategy(),
        "MiddleEllipsis": truncate.NewEllipsisMiddleStrategy(),
    }

    for name, s := range strategies {
        fmt.Printf("%-18s → %q\n", name, truncate.Apply(text, length, s))
    }
    // Output:
    // Cut                → "Hello, W"
    // CutEllipsis        → "Hello, …"
    // LeadingEllipsis    → "… World!"
    // MiddleEllipsis     → "Hel…rld!"
}
```

## API Reference

### Package-Level Functions

| Function                       | Description                      | Return Type |
| ------------------------------ | -------------------------------- | ----------- |
| `Apply(str, length, strategy)` | Truncate string using a strategy | `string`    |

### Strategy Factories

| Function                          | Description          | Omission | Position |
| --------------------------------- | -------------------- | -------- | -------- |
| `NewCutStrategy()`                | Plain cut, no marker | `""`     | End      |
| `NewCutEllipsisStrategy()`        | Ellipsis at end      | `"…"`    | End      |
| `NewCutEllipsisLeadingStrategy()` | Ellipsis at start    | `"…"`    | Start    |
| `NewEllipsisMiddleStrategy()`     | Ellipsis in middle   | `"…"`    | Middle   |

### Truncator Builder

```go
t := truncate.NewTruncator().
    WithOmission("...").                 // Custom omission marker
    WithPosition(truncate.PositionEnd).  // PositionEnd | PositionStart | PositionMiddle
    WithMaxLength(20).                   // Maximum rune count
    Build()
```

| Method                   | Description                 | Default       |
| ------------------------ | --------------------------- | ------------- |
| `NewTruncator()`         | Create a new builder        | —             |
| `WithOmission(string)`   | Set omission marker         | `"…"`         |
| `WithPosition(Position)` | Set truncation position     | `PositionEnd` |
| `WithMaxLength(int)`     | Set max rune count          | `0`           |
| `Build()`                | Build immutable `Truncator` | —             |

### Truncator Methods

| Method                                              | Description                           |
| --------------------------------------------------- | ------------------------------------- |
| `Truncate(str string) string`                       | Truncate using configured `maxLength` |
| `TruncateWithLength(str string, length int) string` | Truncate using the given length       |

### Strategy Interface

```go
type Strategy interface {
    Truncate(str string, length int) string
}
```

Implement this interface to create custom truncation strategies.

### Position Enum

| Constant         | Value | Description                   |
| ---------------- | ----- | ----------------------------- |
| `PositionEnd`    | `0`   | Omission at the end (default) |
| `PositionStart`  | `1`   | Omission at the start         |
| `PositionMiddle` | `2`   | Omission in the middle        |

### Constants

| Constant          | Value | Description               |
| ----------------- | ----- | ------------------------- |
| `DefaultOmission` | `"…"` | Unicode ellipsis (U+2026) |

## Best Practices & Notes

### ⚠️ Common Pitfalls

1. **Byte Length vs Rune Count**

   ```go
   // ❌ Wrong: len() counts bytes, not characters
   text := "café"
   len(text) // 5 bytes (é is 2 bytes)

   // ✅ Correct: truncate operates on runes
   truncate.Apply(text, 3, truncate.NewCutEllipsisStrategy())
   // Returns "ca…" (3 runes)
   ```

2. **Omission Counts Toward Length**

   ```go
   // The omission marker is included in the length budget
   truncate.Apply("Hello", 3, truncate.NewCutEllipsisStrategy())
   // Returns "He…" (2 chars + 1 omission = 3 runes total)
   ```

3. **Length Shorter Than Omission**

   ```go
   // When length ≤ omission length, falls back to plain cut
   truncate.Apply("Hello", 1, truncate.NewCutEllipsisStrategy())
   // Returns "H" (not "…")
   ```

4. **No Truncation Needed**
   ```go
   // Returns original string unchanged if it already fits
   truncate.Apply("Hi", 10, truncate.NewCutEllipsisStrategy())
   // Returns "Hi"
   ```

### 💡 Recommendations

✅ **Reuse Truncator instances** — build once, call many times

```go
// ✅ Build once
var titleTruncator = truncate.NewTruncator().
    WithMaxLength(50).
    Build()

func TruncateTitle(title string) string {
    return titleTruncator.Truncate(title)
}
```

✅ **Use middle truncation for file paths** — preserves both directory and filename

✅ **Use leading truncation for log context** — preserves the most recent (rightmost) information

✅ **Choose omission style by context** — `"…"` for UI, `"..."` for plain text / terminals

✅ **Implement `Strategy` for domain-specific rules** — e.g. word-boundary or sentence-boundary truncation

### 🔒 Thread Safety

All strategies and built `Truncator` instances are safe for concurrent use. They hold only immutable configuration and perform no shared-state mutation.

```go
var t = truncate.NewTruncator().WithMaxLength(50).Build()

// Safe — concurrent truncation
var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func(text string) {
        defer wg.Done()
        _ = t.Truncate(text)
    }(fmt.Sprintf("Text #%d with some long content...", i))
}
wg.Wait()
```

### ⚡ Performance Tips

- **All operations are O(n)** where n = rune count of the input string
- **Middle truncation** allocates a result rune slice — slightly more expensive than start/end
- **Reuse `Truncator`** to avoid repeated builder allocations
- **Avoid unnecessary truncation** — check length first if unsure

## Limitations

- **Rune-based, not grapheme-based** — combined emoji (e.g. family emoji 👨‍👩‍👧‍👦) may be split
- **No word-boundary awareness** — truncation may cut in the middle of a word
- **No HTML/markup awareness** — HTML tags will be treated as plain text
- **No locale-aware behaviour** — does not perform locale-specific collation

## Contributing

Contributions are welcome! Please see the main [replify repository](https://github.com/polarixa/replify) for contribution guidelines.

## License

This library is part of the [replify](https://github.com/polarixa/replify) project.

## Related

Part of the **replify** ecosystem:

- [replify](https://github.com/polarixa/replify) — API response wrapping library
- [strutil](https://github.com/polarixa/replify/tree/master/pkg/strutil) — String utility functions
- [hashy](https://github.com/polarixa/replify/tree/master/pkg/hashy) — Deterministic hashing
- [conv](https://github.com/polarixa/replify/tree/master/pkg/conv) — Type conversion utilities
- [match](https://github.com/polarixa/replify/tree/master/pkg/match) — Wildcard pattern matching
- [coll](https://github.com/polarixa/replify/tree/master/pkg/coll) — Collection utilities
