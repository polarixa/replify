# explorer

A **zero-dependency** Go library for detecting whether a program was launched by double-clicking its executable from Windows Explorer — the classic "mousetrap" problem for console applications.

## Features

- **Zero dependency** — uses only the Go standard library
- **Cross-platform safe** — a single API with build-tagged implementations for Windows and all other platforms
- **No false positives on non-Windows** — the non-Windows build always reports `false` rather than attempting unsupported syscalls

## Installation

```bash
go get github.com/polarixa/replify/pkg/explorer
```

## Quick Start

```go
package main

import (
    "fmt"

    "github.com/polarixa/replify/pkg/explorer"
)

func main() {
    if explorer.StartedByExplorer() {
        fmt.Println("Launched from Explorer — press Enter to exit...")
        fmt.Scanln()
    }
}
```

## API Reference

### `StartedByExplorer` — detect an Explorer-launched console app

```go
// StartedByExplorer returns true if the program was invoked by the user
// double-clicking on the executable from explorer.exe.
func StartedByExplorer() bool
```

On Windows, it walks a toolhelp snapshot of running processes to find the parent process and checks whether its image name is `explorer.exe`. On every other platform it always returns `false`.

## Why This Matters

Console applications launched by double-clicking in Windows Explorer run in a new console window that closes immediately when the program exits, often before the user can read any output. Checking `StartedByExplorer` lets a program detect this case and pause (e.g. prompt "press Enter to exit") so error messages and output remain visible.

## Real-world Example

### Pausing before exit only when double-clicked

```go
package main

import (
    "bufio"
    "fmt"
    "os"

    "github.com/polarixa/replify/pkg/explorer"
)

func main() {
    defer func() {
        if explorer.StartedByExplorer() {
            fmt.Print("Press Enter to exit...")
            bufio.NewReader(os.Stdin).ReadString('\n')
        }
    }()

    // ... run the program.
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
