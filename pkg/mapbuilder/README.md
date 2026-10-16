# mapbuilder

A **zero-dependency** Go library for fluently building and manipulating `map[string]any` values. It provides a chainable `MapBuilder` API for setting, appending, merging, and removing values — including nested values addressed by dot-separated paths.

## Features

- **Zero dependency** — uses only the Go standard library and `pkg/strutil`
- **Fluent chaining** — every mutating method returns the builder, so calls compose naturally
- **Dot-separated paths** — read and write deeply nested values with `AddPath`, `GetPath`, `AppendPath`, `MergePath`, and `RemovePath`, creating intermediate maps as needed
- **Deep copy semantics** — `From` and `Build` deep-copy their maps so the builder and its source/output never alias each other
- **Deferred error handling** — once an operation fails, the builder short-circuits and keeps the first error, letting you chain freely and check the error once at the end
- **Recursive merge** — `Merge` and `MergePath` combine nested maps key-by-key instead of overwriting whole subtrees

## Installation

```bash
go get github.com/polarixa/replify/pkg/mapbuilder
```

## Quick Start

```go
package main

import (
    "fmt"

    "github.com/polarixa/replify/pkg/mapbuilder"
)

func main() {
    m, err := mapbuilder.New().
        Add("name", "Alice").
        AddPath("address.city", "Metropolis").
        Append("roles", "admin", "editor").
        Build()
    if err != nil {
        panic(err)
    }
    fmt.Println(m) // map[address:map[city:Metropolis] name:Alice roles:[admin editor]]
}
```

## API Reference

### Constructors

```go
// New creates an empty builder.
b := mapbuilder.New()

// From creates a builder seeded with a deep copy of an existing map.
b := mapbuilder.From(map[string]any{"name": "Alice"})
```

### Setting values

```go
b.Add("name", "Alice")                  // sets a top-level key
b.AddAll(map[string]any{"a": 1, "b": 2}) // merges in several top-level keys
b.AddIfAbsent("name", "Bob")             // only sets if the key is not already present

b.AddPath("address.city", "Metropolis")     // sets a nested value, creating intermediate maps
b.AddPathIfAbsent("address.city", "Gotham") // only sets if the path is not already present
```

### Appending to slices

```go
b.Append("roles", "admin", "editor")           // creates or appends to a []any at a key
b.AppendPath("user.roles", "admin", "editor")  // same, but addressed by a dot-separated path
```

### Merging maps

```go
// Merge recursively combines nested maps instead of overwriting them wholesale.
b.Merge(map[string]any{"address": map[string]any{"zip": "12345"}})

// MergePath merges into the map located at a dot-separated path.
b.MergePath("user.address", map[string]any{"zip": "12345"})
```

### Removing values

```go
b.Remove("name")           // deletes a top-level key
b.RemovePath("address.zip") // deletes a nested value, pruning empty intermediate maps
```

### Reading values

```go
v, ok := b.Get("name")           // reads a top-level key
v, ok := b.GetPath("address.city") // reads a nested value
b.Has("name")                     // reports whether a top-level key exists
b.HasPath("address.city")         // reports whether a nested path exists
```

### Finalizing the result

```go
m, err := b.Build()  // returns a deep copy of the built map, or an error
m := b.MustBuild()    // same, but panics instead of returning an error

b.IsError() // reports whether an error has occurred
b.Cause()   // returns the underlying error (never nil)
b.Error()   // returns the error message, or "" if no error

b.Reset() // clears the builder back to its initial, empty state
```

## Why Deferred Errors?

Every mutating method checks for a prior error before doing any work and returns immediately if one is already set. This lets you chain any number of operations without checking an error after each call:

```go
m, err := mapbuilder.New().
    Add("", "invalid").      // sets an error: empty key
    Add("name", "Alice").    // no-op: builder is already in an error state
    AppendPath("roles", "admin"). // no-op: builder is already in an error state
    Build()
// err: "Add: empty key", m: nil
```

`Build` surfaces that first error, and all read operations (`Get`, `GetPath`, `Has`, `HasPath`) simply report "not found" once the builder is in an error state rather than panicking or masking the failure.

## Real-world Example

### Building a structured API response

```go
package handler

import (
    "github.com/polarixa/replify/pkg/mapbuilder"
)

func BuildUserResponse(id, name, city string, roles []string) (map[string]any, error) {
    b := mapbuilder.New().
        Add("id", id).
        Add("name", name).
        AddPath("address.city", city)

    for _, role := range roles {
        b.AppendPath("roles", role)
    }

    return b.Build()
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
