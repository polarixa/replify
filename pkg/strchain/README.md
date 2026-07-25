# StrChain — Fluent String Builder for Go

A high-performance string builder for Go that wraps `strings.Builder` with a fluent, chainable API. Choose between a **pure** implementation for maximum throughput or a **thread-safe** variant for concurrent workloads.

## ✨ Features

- ⛓️ **Fluent API** — All methods return `Weaver` for expressive chainable calls
- 🚀 **Zero Overhead** — `StringWeaver` wraps `strings.Builder` directly, no mutex
- 🔒 **Thread-Safe Option** — `SafeStringWeaver` adds `sync.Mutex` protection
- 🎯 **Type-Safe** — Full numeric type support (int8–64, uint8–64, float32/64, bool)
- 📦 **Zero Dependencies** — Built entirely on Go's standard library
- 🧩 **Interface-Driven** — `Weaver` interface enables polymorphism between implementations
- 💡 **Intuitive** — Natural, readable syntax that mirrors how you think about string construction

## 📥 Installation

```bash
go get github.com/polarixa/replify
```

Import:

```go
import "github.com/polarixa/replify/pkg/strchain"
```

## 🏗️ Architecture

```
                    ┌────────────┐
                    │   Weaver   │ (interface)
                    │  48 methods│
                    └─────┬──────┘
                          │
              ┌───────────┴───────────┐
              │                       │
     ┌────────┴────────┐    ┌─────────┴─────────┐
     │  StringWeaver   │    │ SafeStringWeaver   │
     │  (no mutex)     │    │ (sync.Mutex)       │
     │  max performance│    │ thread-safe        │
     └─────────────────┘    └────────────────────┘
```

**Choose the right type:**

| When to use                            | Type               | Constructor                                        |
| -------------------------------------- | ------------------ | -------------------------------------------------- |
| Single goroutine, max performance      | `StringWeaver`     | `New()`, `From()`, `NewWithCapacity()`             |
| Multiple goroutines, concurrent access | `SafeStringWeaver` | `NewSafe()`, `SafeFrom()`, `NewSafeWithCapacity()` |

## 🚀 Quick Start

### Basic Usage (Single-Threaded)

```go
result := strchain.New().
    Append("Hello").
    Space().
    Append("World").
    Build()

fmt.Println(result) // Output: Hello World
```

### Thread-Safe Usage

```go
logger := strchain.NewSafe()
var wg sync.WaitGroup

for i := 1; i <= 5; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        logger.LineF("[Thread %d] Processing completed", id)
    }(i)
}

wg.Wait()
fmt.Println(logger.Build())
```

### Polymorphism via Weaver Interface

```go
func buildGreeting(w strchain.Weaver, name string) string {
    return w.Append("Hello, ").Append(name).Append("!").Build()
}

// Works with either implementation
buildGreeting(strchain.New(), "Alice")         // fast, single-threaded
buildGreeting(strchain.NewSafe(), "Bob")       // safe, multi-threaded
```

## 📚 Core API

Both `StringWeaver` and `SafeStringWeaver` share the same method set, unified by the `Weaver` interface.

### Constructors

```go
// StringWeaver (non-thread-safe, maximum performance)
strchain.New()                    // Create new instance
strchain.NewWithCapacity(1000)    // Pre-allocate capacity
strchain.From("initial text")     // Start from existing string

// SafeStringWeaver (thread-safe with mutex)
strchain.NewSafe()                // Create new instance
strchain.NewSafeWithCapacity(1000)// Pre-allocate capacity
strchain.SafeFrom("initial text") // Start from existing string
```

### String Operations

```go
.Append("text")                   // Append string
.AppendF("Hello %s", name)        // Formatted append (printf-style)
.AppendByte('A')                  // Single byte
.AppendRune('🔒')                 // Single rune
.AppendBytes([]byte{...})         // Byte slice
```

### Type-Specific Appends

```go
.AppendInt(42)                    // int
.AppendInt8(8)                    // int8
.AppendInt16(16)                  // int16
.AppendInt32(32)                  // int32
.AppendInt64(64)                  // int64
.AppendUint(42)                   // uint
.AppendUint8(8)                   // uint8
.AppendUint16(16)                 // uint16
.AppendUint32(32)                 // uint32
.AppendUint64(64)                 // uint64
.AppendUintptr(100)               // uintptr
.AppendFloat32(3.14)              // float32
.AppendFloat64(2.718)             // float64
.AppendBool(true)                 // bool
```

### Whitespace

```go
.Space()                          // Single space
.Spaces(5)                        // Multiple spaces
.Tab()                            // Single tab
.Tabs(3)                          // Multiple tabs
.NewLine()                        // Single newline
.NewLines(2)                      // Multiple newlines
.Line("text")                     // Text + newline
.LineF("Hello %s", name)          // Formatted + newline
```

### Collections

```go
.Join(", ", "A", "B", "C")        // Join with separator
.Repeat("*", 10)                  // Repeat string n times
.Each(items, func(sw, item) {     // Iterate over slice
    sw.Append(item)
})
```

### Conditional Building

```go
.AppendIf(condition, "text")      // Append if true
.AppendIfF(condition, "%s", val)  // Formatted append if true
.When(condition, func(sw) {       // Execute block if true
    sw.Append("text")
})
.Unless(condition, func(sw) {     // Execute block if false
    sw.Append("text")
})
```

> **Note:** `When`, `Unless`, and `Each` accept callbacks with concrete receiver types (`*StringWeaver` or `*SafeStringWeaver`). These methods are not part of the `Weaver` interface.

### Formatting

```go
.Indent(2, "text")                // Indent (2 spaces per level)
.IndentLine(1, "text")            // Indent + newline
.Quote("text")                    // "text"
.SingleQuote("text")              // 'text'
.Parenthesize("text")             // (text)
.Bracket("text")                  // [text]
.Brace("text")                    // {text}
.Wrap("<", "text", ">")           // <text>
```

### Punctuation

```go
.Comma()                          // ,
.Dot()                            // .
.Colon()                          // :
.Semicolon()                      // ;
.Equals()                         // =
.Arrow()                          // ->
.FatArrow()                       // =>
```

### Utilities

```go
.Grow(1000)                       // Pre-allocate capacity
.Reset()                          // Clear and reuse
.Len()                            // Current length
.Cap()                            // Current capacity
.Clone()                          // Create independent copy
.Inspect(func(current) {          // Debug current state
    fmt.Println(current)
})
```

### Output

```go
.Build()                          // Get final string
.String()                         // Alias for Build()
```

## 🎯 Real-World Examples

### SQL Query Building

```go
query := strchain.New().
    Append("SELECT ").
    Join(", ", "id", "name", "email").NewLine().
    Append("FROM users").NewLine().
    Append("WHERE active = true").NewLine().
    Append("ORDER BY created_at DESC").
    Build()
```

### JSON Construction

```go
json := strchain.New().
    Line("{").
    IndentLine(1, `"name": "John",`).
    IndentLine(1, `"age": 30,`).
    IndentLine(1, `"active": true`).
    Append("}").
    Build()
```

### HTML Generation

```go
html := strchain.New().
    Append("<div class=").Quote("container").Append(">").NewLine().
    Indent(1, "<h1>").Append("Welcome").Append("</h1>").NewLine().
    Indent(1, "<p>").Append("Content").Append("</p>").NewLine().
    Append("</div>").
    Build()
```

### Concurrent Log Aggregation

```go
logger := strchain.NewSafe()

go logger.LineF("[%s] Service started", time.Now())
go logger.LineF("[%s] Database connected", time.Now())
go logger.LineF("[%s] Cache initialized", time.Now())

// All logs are safely aggregated — no external locking needed
```

### Concurrent CSV Generation

```go
csv := strchain.NewSafe()
csv.Line("ID,Name,Value")
var wg sync.WaitGroup

for i := 1; i <= 100; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        csv.Join(",",
            fmt.Sprintf("%d", id),
            fmt.Sprintf("Item%d", id),
            fmt.Sprintf("%d", id*100),
        ).NewLine()
    }(i)
}

wg.Wait()
```

### Safe Cloning for Branching

```go
template := strchain.NewSafe()
template.Append("Dear ")

for _, name := range []string{"Alice", "Bob", "Charlie"} {
    go func(n string) {
        personalized := template.Clone() // Thread-safe clone
        personalized.Append(n).Append(",").NewLine()
        fmt.Println(personalized.Build())
    }(name)
}
```

## ⚡ Performance

### Design Philosophy

The `strchain` package provides two implementations to avoid forcing a one-size-fits-all trade-off:

| Implementation     | Overhead                                     | Use Case                   |
| ------------------ | -------------------------------------------- | -------------------------- |
| `StringWeaver`     | **Zero** — direct `strings.Builder` wrapper  | Single-threaded, hot paths |
| `SafeStringWeaver` | **~10–20ns/op** — mutex lock/unlock per call | Concurrent access          |

Both implementations produce **zero allocations** beyond the initial builder allocation — identical to `strings.Builder`.

### Performance Tips

```go
// ✅ Use StringWeaver for single-threaded code
sw := strchain.New()

// ✅ Pre-allocate capacity when the approximate size is known
sw := strchain.NewWithCapacity(4096)

// ✅ Reuse with Reset to avoid re-allocation
sw.Reset()

// ✅ Use SafeStringWeaver only when concurrent access is required
sw := strchain.NewSafe()

// ⚠️ Avoid calling Len() in tight loops with SafeStringWeaver — each call acquires a lock
```

## 🔒 Thread-Safety Details

### StringWeaver

`StringWeaver` is **NOT thread-safe**. It is designed for single-goroutine use where maximum performance is required. Using it concurrently without external synchronization will cause data races.

### SafeStringWeaver

`SafeStringWeaver` is **thread-safe**. All methods use internal mutex synchronization:

- ✅ All Append operations
- ✅ All formatting and punctuation methods
- ✅ `Len()`, `Cap()`, `String()`, `Build()`
- ✅ `Reset()`, `Grow()`, `Clone()`
- ✅ `Inspect()` — locks during state read, unlocks before callback execution

**Callback methods** (`When`, `Unless`, `Each`) do not hold the lock during callback execution. The callback receives the builder instance and can safely call any method (each of which acquires its own lock individually).

### Concurrent Patterns

```go
// ✅ Pattern 1: Shared SafeStringWeaver
logger := strchain.NewSafe()
go logger.Line("A")
go logger.Line("B")

// ✅ Pattern 2: Clone for branching
template := strchain.NewSafe()
template.Append("Base: ")
go func() {
    branch := template.Clone()
    branch.Append("Branch 1")
}()

// ✅ Pattern 3: Separate instances per goroutine
go func() {
    local := strchain.New() // No mutex needed
    local.Append("Local work")
}()
```

## 🆚 Comparison

### StringWeaver vs SafeStringWeaver

| Feature     | StringWeaver               | SafeStringWeaver       |
| ----------- | -------------------------- | ---------------------- |
| Thread-Safe | ❌ No                      | ✅ Yes                 |
| Performance | Baseline (zero overhead)   | ~10–20ns/op mutex cost |
| Allocations | 0                          | 0                      |
| Use Case    | Single goroutine, max perf | Concurrent access      |

### vs. strings.Builder

| Feature      | StringWeaver | SafeStringWeaver  | strings.Builder |
| ------------ | ------------ | ----------------- | --------------- |
| Thread-Safe  | ❌ No        | ✅ Yes            | ❌ No           |
| Fluent API   | ✅ Yes       | ✅ Yes            | ❌ No           |
| Type Methods | ✅ Full      | ✅ Full           | ⚠️ Limited      |
| Performance  | ≈ Baseline   | ~10–20ns overhead | Baseline        |
| Allocations  | 0            | 0                 | 0               |

### vs. strings.Join / fmt.Sprintf

| Feature     | Weaver  | strings.Join | fmt.Sprintf |
| ----------- | ------- | ------------ | ----------- |
| Zero Allocs | ✅ Yes  | ❌ No        | ❌ No       |
| Fluent API  | ✅ Yes  | ❌ No        | ❌ No       |
| Flexibility | ✅ High | ⚠️ Limited   | ⚠️ Medium   |

## 🤝 Contributing

Contributions welcome! Please feel free to submit a Pull Request.

## 📄 License

MIT License — see LICENSE file for details.

## 🙏 Credits

Built with ❤️ on top of Go's excellent `strings.Builder`. Thread-safety powered by `sync.Mutex`.
