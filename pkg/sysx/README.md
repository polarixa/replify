# sysx

**sysx** is a lightweight, production-grade system utilities toolkit for Go, providing a clean and consistent API for OS detection, runtime introspection, environment variable management, process control, command execution, and file system queries — all built exclusively on the Go standard library.

## Overview

The `sysx` package eliminates the boilerplate of writing low-level system queries from scratch. It addresses common pain points like:

- **OS / Architecture detection** – know at runtime whether you are on Linux, macOS, Windows, 64-bit, or ARM
- **Runtime introspection** – read hostname, PID, UID, GID, number of CPUs, goroutine count, Go version, and memory stats in a single call
- **Environment management** – read, write, and parse environment variables with typed helpers (int, bool, slice) and sensible fallbacks
- **Process utilities** – check whether a PID is alive, send signals, look up processes by PID
- **Command execution** – a builder API plus convenience functions supporting timeout, working-directory override, environment injection, real-time streaming, async launch, and shell-style pipelines
- **File system helpers** – existence checks, type checks, permission checks, size queries, directory lookups, read/write helpers, atomic writes, and concurrency-safe writers

**Problem Solved:** Querying the operating system involves a patchwork of `os`, `os/exec`, `runtime`, `syscall`, and `os/user` calls scattered across many packages. `sysx` unifies these into a single, coherent API with uniform error handling and well-documented behavior.

## Design Philosophy

- **Zero external dependencies** — only the Go standard library is used
- **Safe for concurrent use** — all exported functions are stateless or read-only with respect to shared state
- **Explicit errors, not silent failures** — functions return errors rather than hiding them; `Must*` variants are provided where panicking on failure is a deliberate choice
- **Platform-aware** — differences between Linux, macOS, and Windows are documented and handled gracefully
- **No shell interpolation** — command execution helpers use `os/exec` directly, never `sh -c`, to avoid injection risks
- **DRY internals** — the `Command` builder is the single implementation; all convenience functions delegate to it

## Package Architecture

| File           | Responsibility                                                                                                                           |
| -------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| `type.go`      | Struct definitions (`Command`, `CommandResult`, `SafeFileWriter`) with unexported fields; getter/accessor methods; package-level globals |
| `doc.go`       | Package-level godoc documentation                                                                                                        |
| `os.go`        | OS and architecture detection (`IsLinux`, `IsDarwin`, `IsWindows`, `OSVersion`, …)                                                       |
| `runtime.go`   | Runtime information (`Hostname`, `PID`, `UID`, `GoVersion`, `MemStats`, …)                                                               |
| `env.go`       | Environment variable helpers (`Getenv`, `MustGetenv`, `Hasenv`, typed getters, `EnvMap`)                                                 |
| `process.go`   | Process utilities (`ProcessExists`, `KillProcess`, `FindProcessByPID`, …)                                                                |
| `command.go`   | Command builder (`NewCommand`, `With*` methods, `Execute`, `Run`, `Output`) and all convenience functions                                |
| `file.go`      | File system helpers — existence/type/permission checks, read/write functions, atomic and concurrency-safe writers                        |
| `net.go`       | Network utilities — IP classification, port probing, address resolution, connectivity helpers                                            |
| `utilities.go` | Internal helpers (`parseBoolString`, `splitLines`) and exported `UserInfo()`                                                             |
| `entry.go`     | Top-level convenience (`SystemInfo`, `IsPrivileged`)                                                                                     |
| `resource.go`  | `Resource` builder with chainable `With*` setters and `From*` loaders, plus lifecycle methods (`Close`, `Rewind`, `CopyTo`, `Drain`)     |
| `memblob.go`   | `MemBlob`: in-memory `ReadSeekCloser` backing for small payloads                                                                         |
| `tempfile.go`  | `TempFile`: on-disk temporary file backing with idempotent `Close` and configurable auto-removal                                         |
| `spillbuf.go`  | Private hybrid memory → spill-file backing used by `FromReader`                                                                          |
| `mime.go`      | `MimeFromName` helper, MIME constants, and `ErrNilResource`                                                                              |

## Installation

```bash
go get github.com/polarixa/replify
```

Import the package:

```go
import "github.com/polarixa/replify/pkg/sysx"
```

**Requirements:** Go 1.24.0 or higher

## API Reference

### OS Detection (`os.go`)

```go
sysx.IsLinux()   bool   // true on Linux
sysx.IsDarwin()  bool   // true on macOS
sysx.IsWindows() bool   // true on Windows

sysx.OSName()    string // runtime.GOOS  ("linux", "darwin", "windows", …)
sysx.Arch()      string // runtime.GOARCH ("amd64", "arm64", "386", …)
sysx.Is64Bit()   bool   // true for amd64, arm64, ppc64, …
sysx.IsArm()     bool   // true for arm and arm64

sysx.OSVersion() string // best-effort OS version string
```

**OSVersion resolution:**

| Platform | Source                                     |
| -------- | ------------------------------------------ |
| Linux    | `PRETTY_NAME` field from `/etc/os-release` |
| macOS    | Output of `sw_vers -productVersion`        |
| Windows  | `runtime.GOOS + "/" + runtime.GOARCH`      |
| Other    | `runtime.GOOS`                             |

### Runtime Information (`runtime.go`)

```go
sysx.Hostname()          (string, error) // os.Hostname()
sysx.MustHostname()      string          // panics on error
sysx.PID()               int             // os.Getpid()
sysx.PPID()              int             // os.Getppid()
sysx.UID()               int             // os.Getuid()  (-1 on Windows)
sysx.GID()               int             // os.Getgid()  (-1 on Windows)
sysx.ExecutablePath()    (string, error) // os.Executable()
sysx.MustExecutablePath() string         // panics on error
sysx.NumCPU()            int             // runtime.NumCPU()
sysx.NumGoroutine()      int             // runtime.NumGoroutine()
sysx.GoVersion()         string          // runtime.Version() — e.g. "go1.24.0"
sysx.MemStats()          runtime.MemStats
```

### Environment Utilities (`env.go`)

```go
sysx.Getenv(key, fallback string) string          // env var or fallback
sysx.MustGetenv(key string) string                // panics if absent/empty
sysx.Hasenv(key string) bool                      // true when set and non-empty
sysx.Setenv(key, value any) error              // os.Setenv
sysx.Unsetenv(key string) error                   // os.Unsetenv
sysx.GetenvInt(key string, fallback int) int      // parsed int or fallback
sysx.GetenvBool(key string, fallback bool) bool   // parsed bool or fallback
sysx.GetenvSlice(key, sep string) []string        // split by sep, nil if unset
sysx.Environ() []string                           // os.Environ()
sysx.EnvMap() map[string]string                   // all env vars as map
```

**Bool string recognition (case-insensitive):**

| Truthy                   | Falsy                     |
| ------------------------ | ------------------------- |
| `1`, `true`, `yes`, `on` | `0`, `false`, `no`, `off` |

### Process Utilities (`process.go`)

```go
sysx.ProcessExists(pid int) bool                     // true when process is running
sysx.KillProcess(pid int) error                      // SIGTERM
sysx.KillProcessForcefully(pid int) error            // SIGKILL
sysx.CurrentProcessName() string                     // filepath.Base of executable
sysx.FindProcessByPID(pid int) (*os.Process, error)  // os.FindProcess
```

### Command Execution (`command.go`)

The command subsystem has two API layers that share a single implementation.
All `CommandResult` fields are **unexported**; always use the accessor methods.

#### Builder API — `Command` and `CommandResult`

```go
// Create and configure a Command:
cmd := sysx.NewCommand("bash").
    WithArgs("-c", "echo hello").
    WithTimeout(5 * time.Second).
    WithEnv("APP_ENV=prod").
    WithDir("/tmp").
    WithStdin(os.Stdin).
    WithStdout(os.Stdout).
    WithStderr(os.Stderr)

// Execute and get a structured result:
res := cmd.Execute()

// Read results via accessor methods:
res.Stdout()   string         // captured stdout (empty when WithStdout was used)
res.Stderr()   string         // captured stderr (empty when WithStderr was used)
res.ExitCode() int            // 0 on success, -1 when undetermined
res.Duration() time.Duration  // wall-clock execution time
res.Err()      error          // nil on success

// Convenience methods on *CommandResult:
res.IsSuccess()   bool   // true when Err() == nil
res.Combined()  string // Stdout() + Stderr()

// Shorter variants on *Command:
cmd.Run()           error           // Execute().Err()
cmd.Output()        (string, error) // Execute().Combined(), Execute().Err()
```

#### Command Getters

```go
cmd.Name()     string         // program name or path
cmd.Args()     []string       // positional arguments
cmd.Dir()      string         // working directory
cmd.Env()      []string       // extra KEY=VALUE bindings
cmd.Timeout()  time.Duration  // configured timeout (0 = none)
```

#### Convenience Functions

```go
// Structured result shortcut
sysx.RunCommand(name string, args ...string) *CommandResult

// Discard output
sysx.ExecCommand(name string, args ...string) error
sysx.ExecCommandContext(ctx context.Context, name string, args ...string) error
sysx.ExecCommandWithTimeout(timeout time.Duration, name string, args ...string) error
sysx.ExecCommandInDir(dir, name string, args ...string) error

// Capture combined stdout+stderr
sysx.ExecOutput(name string, args ...string) (string, error)
sysx.ExecOutputWithTimeout(timeout time.Duration, name string, args ...string) (string, error)
sysx.ExecOutputInDir(dir, name string, args ...string) (string, error)

// Capture stdout as lines
sysx.ExecOutputLines(name string, args ...string) ([]string, error)

// Stream output in real time to provided writers
sysx.ExecStreaming(stdout, stderr io.Writer, name string, args ...string) error

// Start asynchronously — caller calls cmd.Wait()
sysx.ExecAsync(name string, args ...string) (*exec.Cmd, error)

// Chain commands as a shell pipeline
sysx.ExecPipeline(commands ...[]string) (string, error)
```

### File System Utilities (`file.go`)

#### Existence and Type Checks

```go
sysx.FileExists(path string) bool   // any file system entry
sysx.DirExists(path string) bool    // directory only
sysx.IsFile(path string) bool       // regular file (follows symlinks)
sysx.IsDir(path string) bool        // directory (follows symlinks)
sysx.IsSymlink(path string) bool    // symbolic link (does NOT follow)
```

#### Permission Checks

```go
sysx.IsExecutable(path string) bool  // owner execute bit (0100)
sysx.IsReadable(path string) bool    // owner read bit    (0400)
sysx.IsWritable(path string) bool    // attempts O_WRONLY open
```

#### File Metadata and Directories

```go
sysx.FileSize(path string) (int64, error)
sysx.TempDir() string
sysx.HomeDir() (string, error)
sysx.MustHomeDir() string
sysx.WorkingDir() (string, error)
sysx.MustWorkingDir() string
```

#### File Reading

```go
sysx.ReadFile(path string) ([]byte, error)
sysx.ReadFileString(path string) (string, error)
sysx.ReadLines(path string) ([]string, error)
sysx.StreamLines(path string, handler func(string) error) error
```

#### File Writing

```go
sysx.WriteFile(path string, data []byte) error
sysx.WriteFileString(path string, content string) error
sysx.AppendFile(path string, data []byte) error
sysx.AppendString(path string, content string) error
sysx.WriteLines(path string, lines []string) error
```

#### Atomic and Concurrency-Safe Writes

```go
// Atomic replace via temp-file + rename (POSIX atomic)
sysx.AtomicWriteFile(path string, data []byte) error

// Per-path in-process mutex (multiple goroutines, same path)
sysx.WriteFileLocked(path string, data []byte) error

// Reusable mutex-protected writer for one path
w := sysx.NewSafeFileWriter(path).WithPerm(0o600)
w.Write(data []byte) error
w.WriteString(s string) error
w.Overwrite(data []byte) error  // atomic replace
// Accessors:
w.Path() string        // the file path
w.Perm() os.FileMode   // file permission
```

### Network Utilities (`net.go`)

#### IP Classification

```go
sysx.IsIPv4("192.168.1.1")  bool  // true for valid dotted-decimal IPv4
sysx.IsIPv6("::1")          bool  // true for valid IPv6 (non-IPv4)
sysx.IsLocalIP("10.0.0.1")  bool  // true for loopback, link-local, RFC 1918
```

Recognized local ranges: `127.0.0.0/8`, `169.254.0.0/16`, `10.0.0.0/8`,
`172.16.0.0/12`, `192.168.0.0/16`, `::1`, `fe80::/10`, `fc00::/7`.

#### Port Probing

```go
sysx.IsPortOpen("db.internal", 5432) bool    // TCP connect attempt, 3s timeout
sysx.IsPortAvailable(8080)           bool    // true when port can be bound
```

#### Network Address Helpers

```go
sysx.GetLocalIP()         (string, error)    // first non-loopback IPv4
sysx.GetPublicIP()        (string, error)    // via https://api.ipify.org (needs internet)
sysx.GetInterfaceIPs()    ([]string, error)  // all IPs across all interfaces
```

#### URL and Host Helpers

```go
sysx.IsValidHost("localhost")          bool             // DNS resolve or valid IP
sysx.ParseHostPort("host:8080")        (host, port, error)
sysx.IsValidURL("https://example.com") bool             // valid scheme + host
```

#### Connectivity

```go
sysx.PingHost("google.com")                       error  // TCP port 80, 5s timeout
sysx.CheckTCPConn("db", 5432, 3*time.Second) error  // any port, custom timeout
```

### Entry-Point Helpers (`entry.go`)

```go
sysx.SystemInfo() map[string]string
// keys: "os", "arch", "hostname", "pid", "go_version", "executable", "num_cpu"

sysx.IsPrivileged() bool  // true when UID() == 0 (root on Unix)
```

### Utility Helpers (`utilities.go`)

```go
sysx.UserInfo() string  // "uid=1000 gid=1000"
```

### Resource — Storage-Agnostic Exports (`resource.go`, `memblob.go`, `tempfile.go`, `spillbuf.go`, `mime.go`)

`Resource` is the storage-agnostic envelope every data-exporting workflow returns instead of a raw `*os.File` — reports, dumps, archives, attachments, backups, media generation, document rendering, etc. Consumers (Telegram, S3, email, HTTP download, Discord, storage, backup workers) depend only on `Resource` and the standard `io` interfaces; they never know whether the bytes live in memory, on disk, or in a streaming pipe.

#### Contract

```go
type ReadSeekCloser interface {
    io.Reader
    io.Seeker
    io.Closer
}

// All fields are unexported; access via getters / chainable setters.
type Resource struct{ /* name, size, contentType, content, ... */ }
```

Accessors:

```go
res.Name()           string
res.Size()           int64
res.ContentType()    string
res.Content()        sysx.ReadSeekCloser
res.SpillThreshold() int64
res.TempPattern()    string
res.TempDir()        string
res.RemoveOnClose()  bool
```

Chainable setters (configure-then-load builder):

```go
res.WithName(name)              *Resource
res.WithContentType(mime)       *Resource
res.WithSize(n)                 *Resource
res.WithContent(rsc)            *Resource
res.WithSpillThreshold(bytes)   *Resource
res.WithTempPattern(pattern)    *Resource
res.WithTempDir(dir)            *Resource
res.WithRemoveOnClose(bool)     *Resource
```

Loaders (each takes a single source):

```go
res.FromBytes([]byte)                          *Resource
res.FromString(string)                         *Resource
res.FromFile(*os.File)                         (*Resource, error)
res.FromTempFile(func(io.Writer) error)        (*Resource, error)
res.FromReader(io.Reader)                      (*Resource, error)
```

Lifecycle / consumption:

```go
res.Close()                  // release backing (idempotent, nil-safe)
res.Rewind()                 // seek to offset 0
res.CopyTo(io.Writer)        // stream payload into a sink
res.Drain()                  // consume and discard
```

Backings:

| Type             | Storage                                   | `Close()`                       |
| ---------------- | ----------------------------------------- | ------------------------------- |
| `*sysx.MemBlob`  | byte slice                                | no-op                           |
| `*sysx.TempFile` | on-disk temporary file                    | closes + unlinks (configurable) |
| _(private)_      | hybrid memory → spill file (`FromReader`) | closes spill file when present  |

MIME helpers (in `mime.go`):

```go
sysx.MimeFromName("dump.tar.gz")  // → "application/gzip"
sysx.MimeFromName("report.csv")   // → "text/csv; charset=utf-8"
// Constants: MimeOctetStream, MimeText, MimeCSV, MimeJSON, MimeXML,
//            MimeHTML, MimePDF, MimeZIP, MimeGZIP, MimeSQL
```

#### Producer examples

```go
// 1) In-memory CSV
res := sysx.NewResource().
    WithName("hello.csv").
    FromString("a,b\n1,2\n")
defer res.Close()

// 2) Large export spilled to a temp file (auto-removed on Close)
res, err := sysx.NewResource().
    WithName("audit.csv").
    WithTempPattern("audit-*.csv").
    FromTempFile(func(w io.Writer) error {
        _, err := w.Write(payload)
        return err
    })

// 3) Streaming pg_dump | gzip — seekable for S3 multipart
cmd := exec.CommandContext(ctx, "pg_dump", "my_db")
stdout, _ := cmd.StdoutPipe()
_ = cmd.Start()
res, err := sysx.NewResource().
    WithName("dump.sql.gz").
    WithSpillThreshold(16 << 20).      // 16 MiB
    FromReader(stdout)

// 4) Adopt an existing on-disk file without copying
f, _ := os.Open("/tmp/snapshot.bin")
res, err := sysx.NewResource().
    WithRemoveOnClose(false).          // do not unlink the file on Close
    FromFile(f)
```

#### Consumer examples

Every consumer looks identical regardless of how the bytes were produced.

```go
// HTTP download
func DownloadHandler(w http.ResponseWriter, r *http.Request, res *sysx.Resource) error {
    defer res.Close()
    w.Header().Set("Content-Disposition", "attachment; filename="+res.Name())
    w.Header().Set("Content-Type", res.ContentType())
    _, err := res.CopyTo(w)
    return err
}

// S3 upload
func UploadToS3(ctx context.Context, c S3Client, res *sysx.Resource) error {
    defer res.Close()
    _, err := c.PutObject(ctx, res.Name(), res.Content(), res.Size())
    return err
}

// Telegram upload
func SendToTelegram(ctx context.Context, bot TelegramBot, res *sysx.Resource) error {
    defer res.Close()
    return bot.SendDocument(ctx, res.Name(), res.Content())
}
```

#### Architectural principles

1. Business code returns `*sysx.Resource`, never `*os.File`.
2. Consumers depend on `Resource` and on the standard `io` interfaces.
3. Implementations are interchangeable: memory, temp file, hybrid stream.
4. Cleanup is explicit and lives behind `Resource.Close` (idempotent, nil-safe).
5. New backings (object storage stream, encrypted pipe, database blob, …) plug in through `WithContent` without changes to existing consumers.

## Resource Cookbook — Complete `main` Programs

This section walks through every backing (`MemBlob`, `TempFile`, the private
hybrid spill buffer, and the high-level `Resource`) with self-contained
`main` programs that compile as-is. Examples progress from the most basic
shape to advanced production patterns.

> All snippets assume the import block:

```go
import (
     "bytes"
     "context"
     "encoding/csv"
     "fmt"
     "io"
     "log"
     "net/http"
     "os"
     "os/exec"
     "strings"
     "time"

     "github.com/polarixa/replify/pkg/sysx"
 )
```

### `MemBlob` — In-Memory Backing

`MemBlob` wraps a `[]byte` in a `ReadSeekCloser`. It is the cheapest backing,
allocates nothing beyond the byte slice itself, and `Close()` is a no-op.

#### MB1. Basic: read a byte slice as a stream

```go
package main

func main() {
    blob := sysx.NewMemBlob([]byte("hello, world\n"))
    defer blob.Close() // no-op, but kept for symmetry

    data, err := io.ReadAll(blob)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("len=%d bytes=%q\n", blob.Len(), data)
}
```

#### MB2. Seek, partial reads, and rewind

```go
package main

func main() {
    blob := sysx.NewMemBlob([]byte("0123456789"))
    defer blob.Close()

    // Skip the first 4 bytes.
    if _, err := blob.Seek(4, io.SeekStart); err != nil {
        log.Fatal(err)
    }
    chunk := make([]byte, 3)
    if _, err := io.ReadFull(blob, chunk); err != nil {
        log.Fatal(err)
    }
    fmt.Printf("middle chunk: %q\n", chunk) // "456"

    // Rewind and read everything.
    if _, err := blob.Seek(0, io.SeekStart); err != nil {
        log.Fatal(err)
    }
    all, _ := io.ReadAll(blob)
    fmt.Printf("full payload: %q\n", all)
}
```

#### MB3. Inspect the underlying slice without copying

```go
package main

func main() {
    payload := []byte("retain-not-copy")
    blob := sysx.NewMemBlob(payload)

    // Bytes() returns the same slice that was passed to NewMemBlob.
    // Callers must not mutate it for the lifetime of the MemBlob.
    same := blob.Bytes()
    fmt.Println("aliased:", &payload[0] == &same[0]) // true
    fmt.Println("len:    ", blob.Len())
}
```

#### MB4. Promote a `MemBlob` into a `Resource`

```go
package main

func main() {
    res := sysx.NewResource().
        WithName("greeting.txt").
        FromBytes([]byte("hello\n")) // installs a *MemBlob internally
    defer res.Close()

    fmt.Println("name:        ", res.Name())
    fmt.Println("size:        ", res.Size())
    fmt.Println("content-type:", res.ContentType()) // text/plain; charset=utf-8

    n, err := res.CopyTo(os.Stdout)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("\ncopied %d bytes\n", n)
}
```

### `TempFile` — On-Disk Backing

`TempFile` is the on-disk counterpart of `MemBlob`. It owns an `*os.File`
created by `os.CreateTemp`, exposes idempotent `Close`, and can auto-remove
the file when closed (default) or hand off ownership when disarmed.

#### TF1. Basic: write, rewind, read, auto-remove

```go
package main

func main() {
    tf, err := sysx.NewTempFile()
    if err != nil {
        log.Fatal(err)
    }
    defer tf.Close() // closes + removes from disk

    if _, err := tf.Write([]byte("temporary payload")); err != nil {
        log.Fatal(err)
    }
    if _, err := tf.Seek(0, io.SeekStart); err != nil {
        log.Fatal(err)
    }

    out, _ := io.ReadAll(tf)
    fmt.Println("path:    ", tf.Path())
    fmt.Printf("payload: %q\n", out)
    fmt.Println("removeOnClose:", tf.RemoveOnClose()) // true
}
```

#### TF2. Custom pattern + custom directory

```go
package main

func main() {
    // System temp dir, "audit-*.csv" pattern.
    tf1, err := sysx.NewTempFilename("audit-*.csv")
    if err != nil {
        log.Fatal(err)
    }
    defer tf1.Close()
    fmt.Println("tf1:", tf1.Path())

    // Explicit directory + pattern.
    tf2, err := sysx.NewTempFileAt(os.TempDir(), "session-*.bin")
    if err != nil {
        log.Fatal(err)
    }
    defer tf2.Close()
    fmt.Println("tf2:", tf2.Path())
}
```

#### TF3. Promote a `TempFile` into a permanent file

`WithRemoveOnClose(false)` disarms auto-removal so a downstream consumer
can rename the file to its final destination without losing the bytes.

```go
package main

func main() {
    tf, err := sysx.NewTempFilename("report-*.csv")
    if err != nil {
        log.Fatal(err)
    }

    if _, err := tf.Write([]byte("col1,col2\n1,2\n")); err != nil {
        _ = tf.Close()
        log.Fatal(err)
    }

    // Hand ownership to the filesystem; do NOT unlink on Close.
    tf.WithRemoveOnClose(false)
    if err := tf.Close(); err != nil {
        log.Fatal(err)
    }

    final := "/tmp/report-final.csv"
    if err := os.Rename(tf.Path(), final); err != nil {
        log.Fatal(err)
    }
    fmt.Println("kept at:", final)
}
```

#### TF4. Idempotent `Close` (safe to defer twice)

```go
package main

func main() {
    tf, err := sysx.NewTempFile()
    if err != nil {
        log.Fatal(err)
    }
    defer tf.Close() // safety net

    _, _ = tf.Write([]byte("once"))

    // First Close performs the work.
    if err := tf.Close(); err != nil {
        log.Fatal(err)
    }
    fmt.Println("closed?", tf.Closed()) // true

    // Subsequent calls (including the deferred one) are no-ops.
}
```

#### TF5. Stat the underlying file before consumption

```go
package main

func main() {
    tf, err := sysx.NewTempFilename("size-*.bin")
    if err != nil {
        log.Fatal(err)
    }
    defer tf.Close()

    _, _ = tf.Write(bytes.Repeat([]byte{0xAB}, 4096))

    info, err := tf.Stat()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("name=%s size=%d mode=%s\n", info.Name(), info.Size(), info.Mode())
}
```

### `spillBuffer` — Hybrid Memory → Disk Backing (via `Resource.FromReader`)

The spill buffer is private; consumers reach it through `Resource.FromReader`.
It keeps the first `SpillThreshold` bytes in memory and transparently
overflows the rest to a temporary file, producing a `ReadSeekCloser` even
when the source reader is one-shot (a process pipe, an HTTP body, …).

#### SB1. Small payload — stays in memory

```go
package main

func main() {
    src := strings.NewReader("compact payload that fits in RAM")

    res, err := sysx.NewResource().
        WithName("note.txt").
        WithSpillThreshold(1 << 20). // 1 MiB ceiling
        FromReader(src)
    if err != nil {
        log.Fatal(err)
    }
    defer res.Close()

    fmt.Println("size:", res.Size())
    _, _ = res.CopyTo(os.Stdout)
}
```

#### SB2. Large payload — spills to disk transparently

```go
package main

func main() {
    // 4 MiB of repeating data — well above the 64 KiB threshold below.
    src := bytes.NewReader(bytes.Repeat([]byte("X"), 4<<20))

    res, err := sysx.NewResource().
        WithName("dump.bin").
        WithSpillThreshold(64 << 10). // 64 KiB → triggers spill
        FromReader(src)
    if err != nil {
        log.Fatal(err)
    }
    defer res.Close() // unlinks the spill file (RemoveOnClose default = true)

    fmt.Println("size:", res.Size()) // 4 MiB
    n, _ := res.Drain()              // read & discard
    fmt.Println("drained:", n)
}
```

#### SB3. Custom temp dir and pattern for the spill file

```go
package main

func main() {
    // Pretend pg_dump produces 50 MiB of SQL.
    src := bytes.NewReader(bytes.Repeat([]byte("INSERT ...;\n"), 4_000_000))

    res, err := sysx.NewResource().
        WithName("dump.sql").
        WithContentType(sysx.MimeSQL).
        WithSpillThreshold(8 << 20).      // 8 MiB
        WithTempDir("/var/tmp").          // dedicated spill volume
        WithTempPattern("pgdump-*.sql").  // human-readable name
        FromReader(src)
    if err != nil {
        log.Fatal(err)
    }
    defer res.Close()

    fmt.Println("name:", res.Name(), "size:", res.Size())
}
```

#### SB4. Streaming a subprocess into a seekable `Resource`

This is the canonical use case: a one-shot pipe becomes seekable so it can
be uploaded by an S3 multipart client (which retries chunks).

```go
package main

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    cmd := exec.CommandContext(ctx, "pg_dump", "my_db")
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        log.Fatal(err)
    }
    if err := cmd.Start(); err != nil {
        log.Fatal(err)
    }

    res, err := sysx.NewResource().
        WithName("my_db.sql").
        WithContentType(sysx.MimeSQL).
        WithSpillThreshold(16 << 20). // 16 MiB
        FromReader(stdout)
    if err != nil {
        _ = cmd.Wait()
        log.Fatal(err)
    }
    defer res.Close()

    if err := cmd.Wait(); err != nil {
        log.Fatal(err)
    }

    // Re-readable: the spill backing is seekable.
    _ = res.Rewind()
    fmt.Println("dump bytes:", res.Size())
}
```

### `Resource` — End-to-End Programs

`Resource` is the public envelope. Producers configure a backing through one
of the `From*` loaders; consumers depend only on `Resource` plus standard
`io` interfaces.

#### R1. Basic: in-memory CSV produced and copied to stdout

```go
package main

func main() {
    res := sysx.NewResource().
        WithName("hello.csv").
        FromString("a,b\n1,2\n")
    defer res.Close()

    if _, err := res.CopyTo(os.Stdout); err != nil {
        log.Fatal(err)
    }
}
```

#### R2. Producer that writes through `FromTempFile`

`FromTempFile` is the right choice when the payload is large or expensive to
re-compute and you want it on disk from the start. The producer never sees
the underlying `*os.File` — only an `io.Writer` view.

```go
package main

func main() {
    res, err := sysx.NewResource().
        WithName("users.csv").
        WithContentType(sysx.MimeCSV).
        WithTempPattern("users-*.csv").
        FromTempFile(func(w io.Writer) error {
            cw := csv.NewWriter(w)
            _ = cw.Write([]string{"id", "email"})
            _ = cw.Write([]string{"1", "alice@example.com"})
            _ = cw.Write([]string{"2", "bob@example.com"})
            cw.Flush()
            return cw.Error()
        })
    if err != nil {
        log.Fatal(err)
    }
    defer res.Close() // closes + unlinks the temp file

    fmt.Println("name:", res.Name(), "size:", res.Size())
    _, _ = res.CopyTo(os.Stdout)
}
```

#### R3. Adopt an existing `*os.File` with `FromFile`

`FromFile` is the only loader that accepts a raw `*os.File`. Use it when a
file already exists on disk (snapshot, cached export, …) and you want to
hand it to a consumer through the `Resource` contract.

```go
package main

func main() {
    f, err := os.Open("/tmp/snapshot.bin")
    if err != nil {
        log.Fatal(err)
    }

    res, err := sysx.NewResource().
        WithRemoveOnClose(false). // keep the file after Close()
        FromFile(f)
    if err != nil {
        _ = f.Close()
        log.Fatal(err)
    }
    defer res.Close()

    fmt.Println("name:", res.Name(), "size:", res.Size(),
        "ct:", res.ContentType())
}
```

#### R4. Multi-consumer dispatch with `Rewind`

A single `Resource` can feed several consumers (S3, local archive,
checksum) by rewinding between consumers.

```go
package main

func main() {
    res := sysx.NewResource().
        WithName("payload.bin").
        FromBytes(bytes.Repeat([]byte{0x42}, 1024))
    defer res.Close()

    // Consumer 1: write to disk.
    out, err := os.Create("/tmp/payload.bin")
    if err != nil {
        log.Fatal(err)
    }
    if _, err := res.CopyTo(out); err != nil {
        _ = out.Close()
        log.Fatal(err)
    }
    _ = out.Close()

    // Consumer 2: compute size by draining (rewinding first).
    if err := res.Rewind(); err != nil {
        log.Fatal(err)
    }
    n, err := res.Drain()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("drained:", n)
}
```

#### R5. HTTP download handler — uniform consumer

The handler depends only on `*sysx.Resource`; it does not care whether the
backing is memory, a temp file, or a spill buffer.

```go
package main

func downloadHandler(produce func() (*sysx.Resource, error)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        res, err := produce()
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        defer res.Close()

        w.Header().Set("Content-Type", res.ContentType())
        w.Header().Set("Content-Disposition",
            fmt.Sprintf("attachment; filename=%q", res.Name()))
        if res.Size() >= 0 {
            w.Header().Set("Content-Length", fmt.Sprint(res.Size()))
        }
        if _, err := res.CopyTo(w); err != nil {
            log.Printf("copy failed: %v", err)
        }
    }
}

func main() {
    produce := func() (*sysx.Resource, error) {
        return sysx.NewResource().
            WithName("hello.txt").
            FromString("hello from sysx\n"), nil
    }
    http.HandleFunc("/download", downloadHandler(produce))
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

#### R6. Pick the right backing at runtime

The same producer chooses memory, on-disk, or hybrid streaming based on the
size hint reported by the data source.

```go
package main

// fetchPayload returns a Resource backed by the cheapest viable storage.
func fetchPayload(name string, sizeHint int64, src io.Reader) (*sysx.Resource, error) {
    base := sysx.NewResource().WithName(name)

    switch {
    case sizeHint >= 0 && sizeHint <= 64<<10:
        // Tiny — keep in memory.
        data, err := io.ReadAll(src)
        if err != nil {
            return nil, err
        }
        return base.FromBytes(data), nil

    case sizeHint >= 0 && sizeHint <= 8<<20:
        // Medium — buffer to a temp file in one shot.
        return base.WithTempPattern("payload-*.bin").
            FromTempFile(func(w io.Writer) error {
                _, err := io.Copy(w, src)
                return err
            })

    default:
        // Unknown or large — hybrid spill.
        return base.WithSpillThreshold(8 << 20).
            FromReader(src)
    }
}

func main() {
    res, err := fetchPayload("data.bin", -1, strings.NewReader("streamed payload"))
    if err != nil {
        log.Fatal(err)
    }
    defer res.Close()

    fmt.Println("size:", res.Size())
    _, _ = res.CopyTo(os.Stdout)
}
```

#### R7. Custom `ReadSeekCloser` via `WithContent`

Plug in any backing that satisfies `sysx.ReadSeekCloser` — encrypted pipes,
object-storage streams, database blobs — without touching consumer code.

```go
package main

// readSeekNopCloser wraps a *bytes.Reader and adds a no-op Close.
type readSeekNopCloser struct{ *bytes.Reader }

func (readSeekNopCloser) Close() error { return nil }

func main() {
    payload := []byte("custom-backed payload")
    rsc := readSeekNopCloser{bytes.NewReader(payload)}

    res := sysx.NewResource().
        WithName("custom.bin").
        WithContentType(sysx.MimeOctetStream).
        WithSize(int64(len(payload))).
        WithContent(rsc)
    defer res.Close()

    _, _ = res.CopyTo(os.Stdout)
    fmt.Println()
}
```

#### R8. MIME inference from the filename

`MimeFromName` powers the auto-detection performed by every `From*` loader
when `ContentType` is left empty.

```go
package main

func main() {
    cases := []string{
        "report.csv",
        "dump.tar.gz",
        "page.HTML",
        "schema.sql",
        "no-extension",
    }
    for _, name := range cases {
        fmt.Printf("%-15s → %s\n", name, sysx.MimeFromName(name))
    }
    // report.csv      → text/csv; charset=utf-8
    // dump.tar.gz     → application/gzip
    // page.HTML       → text/html; charset=utf-8
    // schema.sql      → application/sql
    // no-extension    → application/octet-stream
}
```

#### R9. Defensive consumer — nil-safe `Close` and `Rewind`

Every accessor and lifecycle method tolerates a nil receiver, so callers can
defer `Close` even on the error path.

```go
package main

func produce(fail bool) (*sysx.Resource, error) {
    if fail {
        return nil, fmt.Errorf("upstream failure")
    }
    return sysx.NewResource().FromString("ok"), nil
}

func main() {
    res, err := produce(true)
    defer res.Close() // safe even when res == nil

    if err != nil {
        fmt.Println("error path handled cleanly:", err)
        return
    }
    _, _ = res.CopyTo(os.Stdout)
}
```

## Usage Examples

### Basic OS / Architecture Check

```go
fmt.Println("OS:      ", sysx.OSName())
fmt.Println("Arch:    ", sysx.Arch())
fmt.Println("Version: ", sysx.OSVersion())
fmt.Println("64-bit:  ", sysx.Is64Bit())
fmt.Println("ARM:     ", sysx.IsArm())
```

### Environment Configuration

```go
host  := sysx.Getenv("DB_HOST", "localhost")
port  := sysx.GetenvInt("DB_PORT", 5432)
debug := sysx.GetenvBool("DEBUG", false)
tags  := sysx.GetenvSlice("TAGS", ",")

// Panics at startup if DATABASE_URL is not configured
dsn := sysx.MustGetenv("DATABASE_URL")
```

### System Information Snapshot

```go
info := sysx.SystemInfo()
for k, v := range info {
    fmt.Printf("%-12s = %s\n", k, v)
}
// os           = linux
// arch         = amd64
// hostname     = myserver
// pid          = 12345
// go_version   = go1.24.0
// executable   = /usr/local/bin/myapp
// num_cpu      = 8
```

## Command Execution Utilities

### Simple execution

```go
// Fire and forget
if err := sysx.ExecCommand("go", "generate", "./..."); err != nil {
    log.Fatal(err)
}

// Capture combined output
out, err := sysx.ExecOutput("git", "log", "--oneline", "-5")
fmt.Print(out)

// Capture stdout as lines
lines, err := sysx.ExecOutputLines("git", "branch", "-a")
for _, l := range lines {
    fmt.Println(l)
}
```

### Advanced Command Execution Examples

#### Builder with full configuration

```go
res := sysx.NewCommand("bash").
    WithArgs("-c", "echo $APP_ENV && ls /tmp").
    WithTimeout(10 * time.Second).
    WithEnv("APP_ENV=production").
    WithDir("/tmp").
    Execute()

if !res.IsSuccess() {
    log.Printf("command failed (exit %d, %.2fs):\nstdout: %s\nstderr: %s",
        res.ExitCode, res.Duration.Seconds(), res.Stdout, res.Stderr)
}
```

#### Stream output in real time

```go
// Print build output line-by-line as it happens
err := sysx.ExecStreaming(os.Stdout, os.Stderr, "go", "build", "-v", "./...")
```

#### Async execution

```go
cmd, err := sysx.ExecAsync("server", "--port", "8080")
if err != nil {
    log.Fatal(err)
}
// ... do other work while server runs ...
if err := cmd.Wait(); err != nil {
    log.Println("server exited:", err)
}
```

#### Shell-style pipeline

```go
out, err := sysx.ExecPipeline(
    []string{"cat", "/var/log/syslog"},
    []string{"grep", "ERROR"},
    []string{"wc", "-l"},
)
fmt.Println("error lines:", strings.TrimSpace(out))
```

#### Context-controlled execution

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
if err := sysx.ExecCommandContext(ctx, "go", "test", "-race", "./..."); err != nil {
    log.Fatal(err)
}
```

#### Structured result inspection

```go
res := sysx.RunCommand("golangci-lint", "run", "./...")
fmt.Printf("exit=%d duration=%v\n", res.ExitCode, res.Duration.Round(time.Millisecond))
if !res.IsSuccess() {
    fmt.Println("--- stdout ---")
    fmt.Print(res.Stdout)
    fmt.Println("--- stderr ---")
    fmt.Print(res.Stderr)
}
```

## File I/O Utilities

### Reading files

```go
// Read entire file
data, err := sysx.ReadFile("/etc/hosts")

// Read as string
content, err := sysx.ReadFileString("/etc/hostname")
fmt.Println(strings.TrimSpace(content))

// Read line-by-line into slice
lines, err := sysx.ReadLines("/var/log/app.log")
fmt.Printf("%d log entries\n", len(lines))
```

### Streaming large files

```go
// Memory-efficient: one line in memory at a time
err := sysx.StreamLines("/var/log/big.log", func(line string) error {
    if strings.Contains(line, "CRITICAL") {
        alertOnCall(line)
    }
    return nil
})
```

**Early termination example:**

```go
var found string
err := sysx.StreamLines("/etc/passwd", func(line string) error {
    if strings.HasPrefix(line, "root:") {
        found = line
        return errors.New("stop") // stop processing immediately
    }
    return nil
})
```

### Writing files

```go
// Create / truncate
sysx.WriteFile("/tmp/output.bin", data)
sysx.WriteFileString("/tmp/config.txt", configContent)

// Append
sysx.AppendFile("/var/log/app.log", []byte("new entry\n"))
sysx.AppendString("/var/log/app.log", "another entry\n")

// Write line slice (each element newline-terminated, buffered writer)
sysx.WriteLines("/tmp/list.txt", []string{"alpha", "beta", "gamma"})
```

## Concurrency-Safe File Operations

### Atomic writes — no partial reads

`AtomicWriteFile` uses the temp-file + rename pattern. On POSIX systems
`os.Rename` is atomic within the same filesystem, so readers will never
observe a partial write.

```go
// Safe for readers running concurrently:
if err := sysx.AtomicWriteFile("/etc/app/config.json", newJSON); err != nil {
    log.Fatal(err)
}
```

### Per-path in-process locking

`WriteFileLocked` serialize concurrent writes to the same path using a
package-level `sync.Map` of mutexes. No goroutine observes a half-written file
within the same process.

```go
// Safe to call from multiple goroutines targeting the same path:
go sysx.WriteFileLocked("/tmp/shared.dat", payload1)
go sysx.WriteFileLocked("/tmp/shared.dat", payload2)
```

### `SafeFileWriter` — reusable concurrent writer

Share a single `SafeFileWriter` instance across goroutines for safe append
operations. The internal mutex serialize every `Write`, `WriteString`, and
`Overwrite` call.

```go
w := sysx.NewSafeFileWriter("/var/log/app.log").WithPerm(0o600)

// Fan-out writes from multiple goroutines — no external lock required:
for i := 0; i < 100; i++ {
    go func(n int) {
        w.WriteString(fmt.Sprintf("goroutine %d\n", n))
    }(i)
}
```

Use `Overwrite` for atomic replacement through the same writer:

```go
w.Overwrite(updatedConfig) // temp-file + rename under the mutex
```

## Performance Considerations

| Operation                   | Implementation detail                                                        |
| --------------------------- | ---------------------------------------------------------------------------- |
| `ReadLines` / `StreamLines` | `bufio.Scanner` — one line in memory, large files handled efficiently        |
| `WriteLines`                | `bufio.Writer` — batches small writes into a single syscall                  |
| `commandBuffer` (internal)  | `strings.Builder` — avoids `bytes.Buffer` allocations for command output     |
| `AtomicWriteFile`           | single `os.Rename` syscall; no extra read pass                               |
| `WriteFileLocked`           | `sync.Map` + per-path `sync.Mutex` — O(1) contention lookup                  |
| `SafeFileWriter`            | single mutex per instance — lowest overhead for repeated appends to one file |

**Guidelines:**

- For large files, always prefer `StreamLines` over `ReadLines` to avoid loading the whole file into memory.
- Re-use a `SafeFileWriter` instance across goroutines rather than calling `AppendFile` concurrently (which uses no lock).
- Use `AtomicWriteFile` when readers and writers run concurrently; it prevents any reader from seeing partial data.

## Best Practices

1. **Prefer `Must*` only at startup** — use error-returning variants in library code; reserve `Must*` for `main()` or `init()` where a missing resource is fatal.
2. **Check `ProcessExists` before signalling** — `os.FindProcess` on Unix always succeeds; pair it with `ProcessExists` to avoid signalling phantom PIDs.
3. **Always set a timeout in long-running services** — use `WithTimeout` or `ExecCommandWithTimeout` when calling external programs.
4. **`ExecOutput*` merges stdout+stderr** — use the `Command` builder with `WithStdout`/`WithStderr` to separate them when needed.
5. **`EnvMap` is a snapshot** — the returned map is not updated when the environment changes; call again if you need fresh values.
6. **Prefer `AtomicWriteFile` over `WriteFile` for critical data** — `WriteFile` uses `os.WriteFile` which is not atomic on all kernels.
7. **`WriteFileLocked` is in-process only** — for cross-process safety combine it with `AtomicWriteFile` or use a platform lock.

## Platform Caveats

| Feature                          | Linux             | macOS            | Windows                                |
| -------------------------------- | ----------------- | ---------------- | -------------------------------------- |
| `UID()` / `GID()`                | numeric           | numeric          | always -1                              |
| `PPID()`                         | supported         | supported        | always 0                               |
| `ProcessExists`                  | signal(0)         | signal(0)        | FindProcess (unreliable for dead PIDs) |
| `KillProcess`                    | SIGTERM           | SIGTERM          | limited                                |
| `KillProcessForcefully`          | SIGKILL           | SIGKILL          | limited                                |
| `IsExecutable` / `IsReadable`    | mode bits         | mode bits        | approximation                          |
| `IsWritable`                     | open test         | open test        | open test                              |
| `OSVersion`                      | `/etc/os-release` | `sw_vers`        | GOOS/GOARCH string                     |
| `IsPrivileged`                   | UID==0            | UID==0           | always false                           |
| `AtomicWriteFile` rename         | atomic (same fs)  | atomic (same fs) | not guaranteed atomic                  |
| `ExecStreaming` / `ExecPipeline` | full support      | full support     | partial support                        |

## When to Use `sysx` vs stdlib

| Task                           | Prefer                                                                           |
| ------------------------------ | -------------------------------------------------------------------------------- |
| Quick OS check                 | `sysx.IsLinux()` vs `runtime.GOOS == "linux"` — both work; sysx adds readability |
| Env var with fallback          | `sysx.Getenv` — stdlib requires a conditional around `os.LookupEnv`              |
| Typed env vars                 | `sysx.GetenvInt` / `sysx.GetenvBool` — no stdlib equivalent                      |
| Run command and capture output | `sysx.ExecOutput` — saves wiring `cmd.Stdout`, `cmd.Stderr`                      |
| Command with timeout           | `sysx.ExecCommandWithTimeout` — saves `context.WithTimeout` boilerplate          |
| Stream command output          | `sysx.ExecStreaming` — pass any `io.Writer`                                      |
| Check if a file exists         | `sysx.FileExists` — stdlib `os.Stat` + error check                               |
| Read file as lines             | `sysx.ReadLines` / `sysx.StreamLines` — stdlib requires `bufio.Scanner` setup    |
| Write file atomically          | `sysx.AtomicWriteFile` — stdlib has no atomic helper                             |
| Concurrent file writes         | `sysx.SafeFileWriter` / `WriteFileLocked` — stdlib requires manual mutex         |
| Check if a process is alive    | `sysx.ProcessExists` — stdlib requires `syscall.Kill(pid, 0)` and type assertion |
| Complex I/O piping             | `os/exec` directly — `sysx.ExecPipeline` covers linear chains only               |
| Watching the file system       | `fsnotify` or stdlib — outside `sysx` scope                                      |

## Real-World Integration Examples

### Health-Check Endpoint

```go
func healthHandler(w http.ResponseWriter, r *http.Request) {
    info := sysx.SystemInfo()
    json.NewEncoder(w).Encode(map[string]any{
        "status":     "ok",
        "hostname":   info["hostname"],
        "pid":        info["pid"],
        "go_version": info["go_version"],
        "num_cpu":    info["num_cpu"],
    })
}
```

### Startup Configuration Validation

```go
func mustLoadConfig() Config {
    return Config{
        DSN:     sysx.MustGetenv("DATABASE_URL"),
        Port:    sysx.GetenvInt("PORT", 8080),
        Debug:   sysx.GetenvBool("DEBUG", false),
        Origins: sysx.GetenvSlice("CORS_ORIGINS", ","),
    }
}
```

### Conditional Platform Logic

```go
func configureLogging() {
    switch {
    case sysx.IsLinux():
        // Use journald integration
    case sysx.IsDarwin():
        // Use os_log
    default:
        // Fallback to stderr
    }
}
```

### CI Pipeline Step Runner

```go
steps := [][]string{
    {"go", "vet", "./..."},
    {"go", "build", "./..."},
    {"go", "test", "-race", "-cover", "./..."},
}
for _, step := range steps {
    res := sysx.RunCommand(step[0], step[1:]...)
    fmt.Printf("$ %s => exit=%d (%.2fs)\n",
        strings.Join(step, " "), res.ExitCode, res.Duration.Seconds())
    if !res.IsSuccess() {
        fmt.Fprintln(os.Stderr, res.Combined())
        os.Exit(1)
    }
}
```

### Concurrent Log Aggregator

```go
w := sysx.NewSafeFileWriter("/var/log/aggregated.log")
var wg sync.WaitGroup
for _, source := range logSources {
    wg.Add(1)
    go func(src string) {
        defer wg.Done()
        sysx.StreamLines(src, func(line string) error {
            return w.WriteString(line + "\n")
        })
    }(source)
}
wg.Wait()
```

### Atomic Config Reload

```go
func reloadConfig(newCfg []byte) error {
    // Validate before replacing
    if err := validateConfig(newCfg); err != nil {
        return err
    }
    // Replace atomically so readers always see a complete file
    return sysx.AtomicWriteFile("/etc/app/config.json", newCfg)
}
```

## Network Utilities

The `net.go` file provides helpers for common network operations used in
production backend services. All functions use only the Go standard library
and are safe for concurrent use.

### IP and Address Utilities

```go
// Classify IP addresses
sysx.IsIPv4("192.168.1.1")   // true
sysx.IsIPv6("::1")           // true
sysx.IsLocalIP("10.0.0.1")   // true (RFC 1918 private range)
sysx.IsLocalIP("8.8.8.8")    // false (public)

// Get machine addresses
localIP, _ := sysx.GetLocalIP()           // first non-loopback IPv4
publicIP, _ := sysx.GetPublicIP()         // via api.ipify.org
allIPs, _ := sysx.GetInterfaceIPs()       // all IPs from all interfaces
```

### Port Probing

```go
// Check whether a remote port is reachable (production service health check)
if !sysx.IsPortOpen("postgres.internal", 5432) {
    log.Fatal("database unreachable")
}

// Check whether a local port is free before binding
if !sysx.IsPortAvailable(8080) {
    log.Fatal("port 8080 already in use")
}
```

### URL and Host Validation

```go
// Validate URLs before making requests
if !sysx.IsValidURL(webhookURL) {
    return fmt.Errorf("invalid webhook URL: %q", webhookURL)
}

// Parse host:port strings from configuration
host, port, err := sysx.ParseHostPort(config.DBAddr)
// host == "db.internal", port == 5432

// Verify a host is DNS-resolvable before connecting
if !sysx.IsValidHost(host) {
    return fmt.Errorf("cannot resolve host: %q", host)
}
```

### Connectivity Checks

```go
// Quick reachability probe (TCP port 80)
if err := sysx.PingHost("api.internal"); err != nil {
    log.Printf("API host unreachable: %v", err)
}

// Explicit TCP connection check with custom timeout
err := sysx.CheckTCPConn("redis.cache", 6379, 2*time.Second)
if err != nil {
    log.Printf("Redis unavailable: %v", err)
}
```

### Real-World: Startup Dependency Check

```go
func checkDependencies() error {
    deps := []struct{ name, host string; port int }{
        {"PostgreSQL", "db.internal", 5432},
        {"Redis",      "cache.internal", 6379},
        {"Kafka",      "broker.internal", 9092},
    }
    for _, dep := range deps {
        if err := sysx.CheckTCPConn(dep.host, dep.port, 5*time.Second); err != nil {
            return fmt.Errorf("%s is not reachable: %w", dep.name, err)
        }
        log.Printf("✓ %s (%s:%d) reachable", dep.name, dep.host, dep.port)
    }
    return nil
}
```

### Real-World: Dynamic Port Selection

```go
func findFreePort(start, end int) (int, error) {
    for port := start; port <= end; port++ {
        if sysx.IsPortAvailable(port) {
            return port, nil
        }
    }
    return 0, fmt.Errorf("no free port in range %d-%d", start, end)
}
```

### Platform Notes for Network Utilities

| Feature           | Linux          | macOS          | Windows        |
| ----------------- | -------------- | -------------- | -------------- |
| `GetLocalIP`      | supported      | supported      | supported      |
| `GetPublicIP`     | needs internet | needs internet | needs internet |
| `IsPortAvailable` | TCP listen     | TCP listen     | TCP listen     |
| `PingHost`        | TCP port 80    | TCP port 80    | TCP port 80    |
| `CheckTCPConn`    | full support   | full support   | full support   |

> **Note:** `GetPublicIP` requires outbound internet access. It will fail in
> air-gapped environments. Always handle its error gracefully in production code.

---
