# codegen

A **zero-dependency** Go library for generating configurable-length random codes (numbers and letters). Designed for **fulfillment services** where unique, unpredictable order codes are required.

## Features

- **Zero dependency** — uses only the Go standard library
- **Thread-safe** — a single `Generator` instance can be shared across the entire application
- **Cryptographically secure** — uses `crypto/rand` instead of `math/rand`
- **Functional options** — flexible and extensible API
- **OOP style** — `Generator` struct with a complete set of methods

## Requirements

- Go 1.21 or later

## Installation

```bash
go get github.com/polarixa/replify/pkg/codegen
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/polarixa/replify/pkg/codegen"
)

func main() {
    // Create a generator for order codes.
    g, err := codegen.New(
        codegen.WithLength(10),
        codegen.WithCharset(codegen.CharsetAlphanumericUpper),
        codegen.WithPrefix("ORD-"),
    )
    if err != nil {
        log.Fatal(err)
    }

    code, err := g.Generate()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(code) // "ORD-A3BF9KP2XQ"
}
```

## Creating a Generator

### `New` — initialization with error handling

```go
g, err := codegen.New(
    codegen.WithLength(12),
    codegen.WithCharset(codegen.CharsetAlphanumericUpper),
    codegen.WithPrefix("ORD-"),
    codegen.WithSuffix("-VN"),
)
```

### `MustNew` — initialization at startup (panics on error)

```go
// Declare at package level. Fail fast if the configuration is invalid.
var orderGen = codegen.MustNew(
    codegen.WithLength(12),
    codegen.WithCharset(codegen.CharsetAlphanumericUpper),
    codegen.WithPrefix("ORD-"),
)
```

## Configuration Options

| Option                        | Description                                             | Default               |
| ----------------------------- | ------------------------------------------------------- | --------------------- |
| `WithLength(n int)`           | Number of random characters (excluding prefix/suffix)   | `8`                   |
| `WithCharset(c Charset)`      | Character set used for code generation                  | `CharsetAlphanumeric` |
| `WithCustomCharset(s string)` | Custom character set (duplicates removed automatically) | —                     |
| `WithPrefix(s string)`        | Static string prepended to every code                   | `""`                  |
| `WithSuffix(s string)`        | Static string appended to every code                    | `""`                  |

## Built-in Character Sets

| Constant                   | Contents      | Length |
| -------------------------- | ------------- | ------ |
| `CharsetNumeric`           | `0-9`         | 10     |
| `CharsetAlphaLower`        | `a-z`         | 26     |
| `CharsetAlphaUpper`        | `A-Z`         | 26     |
| `CharsetAlpha`             | `a-z A-Z`     | 52     |
| `CharsetAlphanumeric`      | `0-9 a-z A-Z` | 62     |
| `CharsetAlphanumericUpper` | `0-9 A-Z`     | 36     |
| `CharsetAlphanumericLower` | `0-9 a-z`     | 36     |

## API Reference

### Generate a single code

```go
code, err := g.Generate()
// "ORD-A3BF9KP2XQ"
```

### Generate multiple codes

```go
codes, err := g.GenerateN(100)
// ["ORD-A3BF9KP2XQ", "ORD-B7CD4MN8RT", ...]
```

### Update the configuration

```go
// Atomic update — the previous configuration remains unchanged if an error occurs.
err := g.SetOptions(
    codegen.WithLength(16),
    codegen.WithPrefix("INV-"),
)
```

### Read the current configuration

```go
opts := g.Options()
fmt.Printf("Length=%d, Charset=%s\n", opts.Length, opts.Charset)
```

### Package-level convenience function (one-time generation)

```go
// Convenient when the Generator does not need to be reused.
code, err := codegen.Generate(
    codegen.WithLength(10),
    codegen.WithCharset(codegen.CharsetNumeric),
)
```

## Real-world Examples

### Order Management (Fulfillment)

```go
package order

import "github.com/polarixa/replify/pkg/codegen"

// Initialize once and reuse throughout the service.
var orderCodeGen = codegen.MustNew(
    codegen.WithLength(10),
    codegen.WithCharset(codegen.CharsetAlphanumericUpper),
    codegen.WithPrefix("ORD-"),
)

var invoiceCodeGen = codegen.MustNew(
    codegen.WithLength(8),
    codegen.WithCharset(codegen.CharsetNumeric),
    codegen.WithPrefix("INV-"),
    codegen.WithSuffix("-VN"),
)

// GenerateOrderCode generates an order code (thread-safe).
func GenerateOrderCode() (string, error) {
    return orderCodeGen.Generate()
    // "ORD-A3BF9KP2XQ"
}

// GenerateInvoiceCode generates an invoice code (thread-safe).
func GenerateInvoiceCode() (string, error) {
    return invoiceCodeGen.Generate()
    // "INV-84729163-VN"
}

// CreateOrders generates codes in bulk for batch processing.
func CreateOrders(count int) ([]string, error) {
    return orderCodeGen.GenerateN(count)
}
```

### Using an Unambiguous Character Set

```go
// Exclude visually ambiguous characters: 0/O and 1/I/l.
g, _ := codegen.New(
    codegen.WithLength(8),
    codegen.WithCustomCharset("23456789ABCDEFGHJKLMNPQRSTUVWXYZ"),
)
```

### Integration with an HTTP Handler (Concurrent-Safe)

```go
var gen = codegen.MustNew(
    codegen.WithLength(12),
    codegen.WithCharset(codegen.CharsetAlphanumericUpper),
    codegen.WithPrefix("TXN-"),
)

func CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
    // Generator is safe for concurrent use across multiple requests.
    code, err := gen.Generate()
    if err != nil {
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }
    // ... process the order using the generated code.
}
```

## Thread Safety

`Generator` uses `sync.Mutex` to protect its internal state and relies on `crypto/rand` (which is already thread-safe) for random number generation. A single instance can be safely shared among thousands of concurrent goroutines.

```go
// This is the recommended pattern — share a single Generator.
var gen = codegen.MustNew(codegen.WithLength(12))

// Safe to call concurrently from multiple goroutines.
go func() { gen.Generate() }()
go func() { gen.Generate() }()
go func() { gen.SetOptions(codegen.WithPrefix("NEW-")) }()
```

## Running Tests

```bash
# Unit tests + race detector
go test -race ./...

# With coverage report
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Benchmarks
go test -bench=. -benchmem ./...
```
