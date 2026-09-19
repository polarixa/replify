# b64

A **zero-dependency** Go library for validating Base64-encoded strings. Built on top of the standard `encoding/base64` package with an object-oriented, extensible design that supports multiple encodings and dialects.

## Features

- **Zero dependency** — uses only the Go standard library
- **Canonical validation** — a decode-then-re-encode round trip catches non-canonical padding, mixed alphabets, and embedded whitespace that a bare decode-only check would miss
- **Multiple dialects** — validate against standard, URL-safe, and unpadded ("raw") Base64 encodings
- **Composable** — combine any number of encodings into a single validator via `CompositeValidator`
- **Interface-driven** — the `Validator` interface makes it easy to mock or substitute in tests

## Installation

```bash
go get github.com/polarixa/replify/pkg/b64
```

## Quick Start

```go
package main

import (
    "fmt"

    "github.com/polarixa/replify/pkg/b64"
)

func main() {
    fmt.Println(b64.IsBase64("SGVsbG8sIFdvcmxkIQ==")) // true
    fmt.Println(b64.IsBase64("not_base64!"))           // false

    // Accept any common Base64 dialect (std, raw-std, URL-safe, raw-URL-safe).
    fmt.Println(b64.IsBase64Any("SGVsbG8h")) // true
}
```

## API Reference

### Package-level convenience functions

```go
// IsBase64 checks against standard Base64 (RFC 4648 with padding) only.
b64.IsBase64("SGVsbG8sIFdvcmxkIQ==") // true

// IsBase64Any checks against standard, raw-standard, URL-safe, and raw-URL-safe.
b64.IsBase64Any("PDw_Pz8-Pg") // true
```

### `Validator` interface

Every validator in this package — single-encoding or composite — implements:

```go
type Validator interface {
    IsValid(s string) bool
}
```

### `Base64Validator` — validate against a single encoding

```go
v := b64.NewBase64Validator(base64.URLEncoding)
v.IsValid("PDw_Pz8-Pg==") // true

// A nil encoding defaults to base64.StdEncoding.
v := b64.NewBase64Validator(nil)
v.IsValid("SGVsbG8h") // true
```

### `CompositeValidator` — validate against several encodings at once

A string is considered valid if **any** of the underlying validators accept it.

```go
v := b64.NewCompositeValidator(
    b64.NewBase64Validator(base64.StdEncoding),
    b64.NewBase64Validator(base64.URLEncoding),
)
v.IsValid("PDw_Pz8-Pg==") // true
```

### `NewMultiEncodingValidator` — build a composite directly from encodings

```go
v := b64.NewMultiEncodingValidator(
    base64.StdEncoding,
    base64.RawStdEncoding,
    base64.URLEncoding,
    base64.RawURLEncoding,
)
v.IsValid("SGVsbG8h") // true
```

## Why Round-Trip Validation?

A naive check that only calls `base64.StdEncoding.DecodeString` and looks for an error will accept strings that decode successfully but are not canonical Base64 — e.g. non-canonical trailing bits, or strings mixing padding in unexpected positions. Every validator in this package instead decodes the input and **re-encodes** the result, rejecting the string unless the round trip reproduces the original input exactly. This catches:

- Non-canonical padding (e.g. `"AB=="`)
- Wrong alphabets for the target encoding
- Embedded whitespace or newlines
- Truncated or malformed padding (e.g. `"QQ"`, `"===="`)

## Real-world Example

### Validating an API payload that may arrive in different Base64 dialects

```go
package handler

import (
    "encoding/base64"
    "net/http"

    "github.com/polarixa/replify/pkg/b64"
)

// payloadValidator accepts standard and URL-safe Base64, since clients may
// send either depending on their platform.
var payloadValidator = b64.NewMultiEncodingValidator(
    base64.StdEncoding,
    base64.URLEncoding,
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
    payload := r.FormValue("data")
    if !payloadValidator.IsValid(payload) {
        http.Error(w, "invalid base64 payload", http.StatusBadRequest)
        return
    }
    // ... process the payload.
}
```

## Running Tests

```bash
# Unit tests + race detector
go test -race ./...

# With coverage report
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```
