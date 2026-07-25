# slogger

Zero-dependency, production-grade structured logging library for Go applications.

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-blue)](https://go.dev/)
[![Go Reference](https://pkg.go.dev/badge/github.com/polarixa/replify/pkg/slogger.svg)](https://pkg.go.dev/github.com/polarixa/replify/pkg/slogger)
[![License](https://img.shields.io/badge/license-GPL--3.0-green)](../../LICENSE)

---

## Overview

Modern Go applications require structured, levelled logging that integrates seamlessly with log aggregators, distributed tracing systems, and observability platforms. Traditional logging approaches using `fmt.Printf` or the standard `log` package lack structure, making it difficult to filter, query, and alert on specific fields at scale.

**slogger** solves this by providing a lightweight, allocation-efficient structured logger built entirely on the Go standard library. It enables developers to attach typed key-value pairs to every log entry, supporting both human-readable text output for development and machine-parseable JSON for production environments.

### Key features

- **Zero allocations on hot paths** — entry pooling and inline field storage eliminate heap pressure
- **Structured fields** — typed constructors for strings, integers, floats, errors, durations, and more
- **Dual formatters** — human-readable `TextFormatter` and JSON `JSONFormatter` for any environment
- **Context propagation** — fields stored in `context.Context` are merged automatically at log time
- **File rotation** — size-based and age-based rotation with optional ZIP compression
- **Rate limiting** — per-message sampling prevents log storms during error spikes
- **Side-effect hooks** — level-indexed callbacks for alerting, metrics, and remote shipping

---

## Installation

```bash
go get github.com/polarixa/replify/pkg/slogger@latest
```

### Requirements

- Go 1.21 or higher
- No external dependencies (standard library only)

---

## Quick start

```go
package main

import (
    "errors"
    "os"

    "github.com/polarixa/replify/pkg/slogger"
)

func main() {
    // Create a JSON logger writing to stdout at DEBUG level.
    log := slogger.NewLogger().
        WithLevel(slogger.DebugLevel).
        WithFormatter(slogger.NewJSONFormatter()).
        WithOutput(os.Stdout)

    // Log structured fields with typed constructors.
    log.Info("server started",
        slogger.String("addr", ":8080"),
        slogger.String("env", "production"),
    )

    // Handle errors with structured context.
    if err := processRequest(); err != nil {
        log.Error("request failed",
            slogger.Err(err),
            slogger.String("endpoint", "/api/users"),
        )
    }
}

func processRequest() error {
    return errors.New("connection timeout")
}
// Output: {"ts":"2026-01-15 10:00:00.000000","level":"INFO","msg":"server started","addr":":8080","env":"production"}
// Output: {"ts":"2026-01-15 10:00:00.000001","level":"ERROR","msg":"request failed","error":"connection timeout","endpoint":"/api/users"}
```

---

## Usage examples

### Child loggers with bound fields

Use `With` to create child loggers that prepend fields to every entry, ideal for request-scoped logging.

```go
package main

import (
    "os"

    "github.com/polarixa/replify/pkg/slogger"
)

func main() {
    log := slogger.NewLogger().
        WithLevel(slogger.DebugLevel).
        WithFormatter(slogger.NewTextFormatter(os.Stdout)).
        WithOutput(os.Stdout)

    // Create a child logger with request-scoped fields.
    requestID := "req-abc-123"
    userID := "user-42"

    reqLog := log.With(
        slogger.String("request_id", requestID),
        slogger.String("user_id", userID),
    )

    reqLog.Info("processing request")
    reqLog.Debug("validating input")
    reqLog.Info("request completed")
    // Output: 2026-01-15 10:00:00.000000 INFO processing request request_id=req-abc-123 user_id=user-42
}
```

### Context-aware logging

Store fields in `context.Context` for automatic propagation through call stacks without threading loggers.

```go
package main

import (
    "context"
    "os"
    "time"

    "github.com/polarixa/replify/pkg/slogger"
)

func main() {
    log := slogger.NewLogger().
        WithLevel(slogger.DebugLevel).
        WithFormatter(slogger.NewJSONFormatter()).
        WithOutput(os.Stdout)

    // Inject trace context at the entry point.
    ctx := context.Background()
    ctx = slogger.WithContextFields(ctx,
        slogger.String("trace_id", "trace-xyz-789"),
        slogger.String("span_id", "span-001"),
    )

    // Deep in the call stack, log with context fields automatically merged.
    processOrder(ctx, log, "order-123")
}

func processOrder(ctx context.Context, log *slogger.Logger, orderID string) {
    start := time.Now()
    // Context fields (trace_id, span_id) are merged automatically.
    log.WithContext(ctx).Info("processing order",
        slogger.String("order_id", orderID),
        slogger.Duration("latency", time.Since(start)),
    )
    // Output: {"ts":"...","level":"INFO","msg":"processing order","trace_id":"trace-xyz-789","span_id":"span-001","order_id":"order-123","latency":"42µs"}
}
```

### File rotation with compression

Enable automatic per-level log file rotation with size limits and optional ZIP compression.

```go
package main

import (
	"os"
	"time"

	"github.com/polarixa/replify/pkg/slogger"
)

func main() {
	// Create a logger with file rotation enabled.

	log := slogger.NewLogger().
		WithLevel(slogger.InfoLevel).
		WithFormatter(slogger.NewTextFormatter(os.Stdout)).
		WithOutput(os.Stdout).
		WithRotation(slogger.NewRotationOptions().
			WithDirectory("logs"). // base directory; created if absent
			WithMaxBytes(50 * 1024 * 1024). // rotate at 50 MiB
			WithMaxAge(24 * time.Hour). // or after 24 hours, whichever comes first
			WithCompress(true)) // zip rotated files

	log.Info("application started with rotation enabled")
	log.Warn("disk space low", slogger.Int("percent_used", 85))
	// Creates: logs/info.log, logs/warn.log, logs/error.log, logs/debug.log
	// Rotated files: logs/archived/2026-01-15/20260115100000_info.zip
}
```

### Custom hooks for alerting

Implement the `Hook` interface to trigger side-effects like alerting or metrics on specific log levels.

```go
package main

import (
    "fmt"
    "os"

    "github.com/polarixa/replify/pkg/slogger"
)

// AlertHook sends alerts for ERROR and FATAL level logs.
type AlertHook struct{}

func (h *AlertHook) Levels() []slogger.Level {
    return []slogger.Level{slogger.ErrorLevel, slogger.FatalLevel}
}

func (h *AlertHook) Fire(e *slogger.Entry) error {
    // Send to alerting system (Slack, PagerDuty, etc.).
    fmt.Printf("[ALERT] %s: %s\n", e.Level(), e.Message())
    return nil
}

func main() {
    log := slogger.NewLogger().
        WithLevel(slogger.InfoLevel).
        WithFormatter(slogger.NewTextFormatter(os.Stderr)).
        WithOutput(os.Stderr)

    // Register the alert hook.
    log.AddHook(&AlertHook{})

    log.Info("normal operation")                       // No alert
    log.Error("database connection failed")            // Triggers alert
    // Output: [ALERT] ERROR: database connection failed
}
```

### Global logger configuration

Use the package-level global logger for simple applications or replace it during bootstrap.

```go
package main

import (
    "os"

    "github.com/polarixa/replify/pkg/slogger"
)

func main() {
    // Configure and install a custom global logger.
    log := slogger.NewLogger().
        WithLevel(slogger.DebugLevel).
        WithFormatter(slogger.NewJSONFormatter()).
        WithOutput(os.Stdout)

    slogger.SetGlobalLogger(log)

    // Use package-level functions anywhere in the application.
    slogger.Info("application ready",
        slogger.String("version", "1.0.0"),
    )
    slogger.Debugf("loaded %d plugins", 5)
    // Output: {"ts":"...","level":"INFO","msg":"application ready","version":"1.0.0"}
    // Output: {"ts":"...","level":"DEBUG","msg":"loaded 5 plugins"}
}
```

---

## API overview

| Type/Function                                             | Description                                                          |
| --------------------------------------------------------- | -------------------------------------------------------------------- |
| `Logger`                                                  | Core logging type; all methods are goroutine-safe                    |
| `Entry`                                                   | In-flight log event; do not retain after log call returns            |
| `Field`                                                   | Typed key-value pair for structured logging                          |
| `Formatter`                                               | Interface for serialising entries (`TextFormatter`, `JSONFormatter`) |
| `Hook`                                                    | Interface for side-effect callbacks on log events                    |
| `New(opts...)`                                            | Creates a logger with functional options                             |
| `NewLogger()`                                             | Creates a logger with defaults for fluent configuration              |
| `With(fields...)`                                         | Returns a child logger with bound fields                             |
| `Named(name)`                                             | Returns a child logger with a hierarchical name                      |
| `WithContext(ctx)`                                        | Returns an entry bound to the given context                          |
| `ParseLevel(s)`                                           | Parses a level string (e.g., "info", "DEBUG")                        |
| `String`, `Int`, `Bool`, `Err`, `Duration`, `Time`, `Any` | Typed field constructors                                             |

For complete API documentation, see [pkg.go.dev](https://pkg.go.dev/github.com/polarixa/replify/pkg/slogger).

---

## Configuration

### Log levels

Seven severity levels are defined, in increasing order:

| Constant     | Numeric | Meaning                                       |
| ------------ | ------- | --------------------------------------------- |
| `TraceLevel` | 0       | Fine-grained diagnostic output; very verbose  |
| `DebugLevel` | 1       | Developer debugging; enabled in test/staging  |
| `InfoLevel`  | 2       | General operational messages; default minimum |
| `WarnLevel`  | 3       | Potentially harmful situations                |
| `ErrorLevel` | 4       | Errors that do not stop the application       |
| `FatalLevel` | 5       | Logs the message then calls `os.Exit(1)`      |
| `PanicLevel` | 6       | Logs the message then panics                  |

### Environment-based configuration

```go
lvl, err := slogger.ParseLevel(os.Getenv("LOG_LEVEL"))
if err != nil {
    lvl = slogger.InfoLevel
}
log.SetLevel(lvl)
```

### Logger options

| Option     | Type        | Default         | Description                     |
| ---------- | ----------- | --------------- | ------------------------------- |
| Level      | `Level`     | `InfoLevel`     | Minimum log level               |
| Formatter  | `Formatter` | `TextFormatter` | Entry serialiser                |
| Output     | `io.Writer` | `os.Stderr`     | Primary write destination       |
| Caller     | `bool`      | `false`         | Capture source file and line    |
| CallerSkip | `int`       | `0`             | Additional stack frames to skip |

### Rotation options

| Option   | Type            | Default        | Description                      |
| -------- | --------------- | -------------- | -------------------------------- |
| Dir      | `string`        | `"logs"`       | Base log directory               |
| MaxBytes | `int64`         | `10 MiB`       | File size threshold for rotation |
| MaxAge   | `time.Duration` | `0` (disabled) | Age threshold for rotation       |
| Compress | `bool`          | `false`        | ZIP-compress rotated files       |

### Sampling options

| Option     | Type            | Description                                           |
| ---------- | --------------- | ----------------------------------------------------- |
| First      | `int`           | Number of identical messages always logged per Period |
| Period     | `time.Duration` | Sliding window for the counter                        |
| Thereafter | `int`           | Log every Nth message after First; 0 = drop all       |

---

## Platform support

slogger is built entirely on the Go standard library and supports all platforms where Go runs.

| OS      | Architecture      | Status             |
| ------- | ----------------- | ------------------ |
| Linux   | amd64, arm64, arm | ✅ Fully supported |
| macOS   | amd64, arm64      | ✅ Fully supported |
| Windows | amd64, arm64      | ✅ Fully supported |
| FreeBSD | amd64             | ✅ Fully supported |

### Platform-specific notes

- **TTY detection**: Color output is automatically disabled when stdout/stderr is not a terminal (e.g., in CI pipelines or when piped to files)
- **File rotation**: Archive directories use forward slashes internally; path handling is OS-aware via `filepath`
- **Atomic operations**: Logger level changes use `sync/atomic` for lock-free reads on all platforms

---

## Contributing

Contributions are welcome! Please see the [contributing guidelines](../../CONTRIBUTING.md) for details.

### Local development

```bash
# Clone the repository
git clone https://github.com/polarixa/replify.git
cd replify

# Run tests for the slogger package
go test -v ./pkg/slogger/...

# Run tests with race detection
go test -race ./pkg/slogger/...

# Run benchmarks
go test -bench=. ./pkg/slogger/...
```

---

## License

This project is licensed under the GNU General Public License v3.0 — see the [LICENSE](../../LICENSE) file for details.

---

## Architecture

```
pkg/slogger/
│
├── constant.go        — All package constants (levels, field types, ANSI codes,
│                        default dirs, file names, timestamp formats, JSON keys)
│
├── type.go            — All type definitions, ordered highest→lowest abstraction
│                        Logger, Entry, CallerInfo, Field, Formatter (interface),
│                        Hook (interface), Hooks, Options, SamplingOptions,
│                        TextFormatter, JSONFormatter, MultiWriter,
│                        RotationOptions, LevelFileWriter, LevelWriterHook,
│                        sampler, rotatingFile, contextKey, entryPool, global
│
├── utilities.go       — All unexported helper functions:
│                        entry pool, trimFile, defaultOptions, newSampler,
│                        rotation helpers, color helpers, JSON/text helpers
│
├── level.go           — Level methods: String, ParseLevel, IsEnabled
├── field.go           — Field constructors: String, Int, Int64, Float64, …
│
├── entry.go           — Entry/CallerInfo methods and accessors
│                        reset, WithContext, Logger, Time, GetLevel, Message,
│                        Fields, Caller, Context, File, Line, Function
│                        Entry-level log methods: Trace … Panic
│
├── formatter.go       — (placeholder; Formatter interface in type.go)
├── formatter_text.go  — TextFormatter: human-readable key=value output
├── formatter_json.go  — JSONFormatter: single-line JSON output
│
├── hook.go            — Hooks registry: NewHooks, Add, Fire, Len
├── writer.go          — MultiWriter, Stdout, Stderr
├── color.go           — IsTTY terminal detection
├── options.go         — WithRotation functional option
├── context.go         — WithContextFields, FieldsFromContext
├── sampling.go        — sampler.allow (rate-limiting logic)
│
├── logger.go          — Logger: New, With, Named, SetLevel, GetLevel,
│                        SetOutput, SetFormatter, AddHook, IsLevelEnabled,
│                        Trace … Panic, Tracef … Panicf, log, logCtx, getCaller
│
├── global.go          — Package-level delegates to a global Logger
│                        SetGlobalLogger, GetGlobalLogger, Trace … Errorf,
│                        GlobalWithContextFields
│
└── rotation.go        — LevelFileWriter, LevelWriterHook:
                         NewLevelFileWriter, WriteLevel, Write, Close, Rotate,
                         NewLevelWriterHook, Levels, Fire
```

---

## Additional documentation

### Component deep-dive

#### Logger

`Logger` is the root type. It holds:

- **`level`** (`atomic.Int32`) — the minimum severity threshold; reads/writes
  are lock-free via `sync/atomic`.
- **`formatter`** (`Formatter`) — the serialiser that converts an `Entry` to
  `[]byte`. Protected by `mu` for safe runtime replacement.
- **`output`** (`io.Writer`) — the primary write destination. Protected by `mu`.
- **`hooks`** (`*Hooks`) — the level-indexed registry of side-effect handlers.
- **`fields`** (`[]Field`) — logger-bound fields prepended to every entry.
- **`name`** (`string`) — dot-separated logger name shown in output.
- **`caller`** / **`callerSkip`** — source-location capture settings.
- **`sampling`** (`*sampler`) — optional per-message rate limiter.

All exported `Logger` methods are goroutine-safe. The formatter and output are
double-protected: reading under a read-lock, writing under an exclusive lock.
The level itself uses atomic operations so that `SetLevel` does not block
concurrent log calls.

#### Entry

`Entry` is an in-flight log event. It is obtained from `sync.Pool` via
`acquireEntry` and returned via `releaseEntry` after the entry has been
formatted and written. The pool pre-allocates a `fields` slice of capacity
`defaultEntryFieldCap` to avoid reallocations for the common case of a small
number of fields per entry.

**Important:** callers must not retain an `*Entry` after the logging call that
produced it returns; the entry will be recycled and its fields overwritten.

#### Field

`Field` is a discriminated union: a `FieldType` tag plus a set of value slots
(string, int64, float64, bool, error, time.Time, time.Duration, interface{}).
Primitive types are stored inline, avoiding heap allocation. The `Value()`
method renders any variant as a `string` for formatters that need text output.

#### Formatter

`Formatter` is a single-method interface:

```go
type Formatter interface {
    Format(*Entry) ([]byte, error)
}
```

Two built-in implementations are provided:

| Formatter       | Output style                                   | Best for                    |
| --------------- | ---------------------------------------------- | --------------------------- |
| `TextFormatter` | `timestamp LEVEL [name] message key=value`     | Development, CLI tools      |
| `JSONFormatter` | `{"ts":"…","level":"…","msg":"…","key":value}` | Production, log aggregators |

Both formatters are stateless once constructed and are safe for concurrent use.

#### Hook

A `Hook` fires side-effects on matching log levels after the entry has been
written to the primary output. Typical uses include:

- Sending high-severity alerts (Slack, PagerDuty)
- Emitting metrics (Prometheus counters)
- Shipping entries to a remote log aggregator

Hooks are registered per-logger with `Logger.AddHook` and stored in a
level-indexed registry (`Hooks`) protected by a read-write mutex.

#### Rotation

`LevelFileWriter` maintains four open file handles (debug, info, warn, error).
On each write it checks whether the current file exceeds `MaxBytes` or `MaxAge`
and triggers rotation if needed. Rotation:

1. Closes the active file.
2. Creates a date-based archive directory (`logs/archived/2006-01-02/`).
3. Moves or compresses (ZIP) the old file with a timestamp prefix.
4. Opens a fresh file at the original path.

The `LevelWriterHook` bridges the `Hook` interface and `LevelFileWriter`, so
rotation integrates seamlessly with the hook system.

---

## Logging modes

### Structured logging

Structured logging attaches discrete, typed key-value pairs to every log
entry rather than embedding values inside a free-form string.

```go
log.Info("user login",
    slogger.String("user_id", uid),
    slogger.String("ip",      ip),
    slogger.Duration("latency", latency),
)
```

**JSON output:**

```json
{
  "ts": "2026-01-15T10:00:00Z",
  "level": "INFO",
  "msg": "user login",
  "user_id": "u42",
  "ip": "10.0.0.1",
  "latency": "3.2ms"
}
```

**Advantages:**

- Fields are indexable and queryable in log aggregators (Elasticsearch, Loki, Datadog).
- No string parsing needed downstream; every field has a predictable name and type.
- New fields can be added without changing parsers or dashboards.

**Disadvantages:**

- Slightly more verbose at the call site compared to `fmt.Sprintf`.
- Requires a log aggregator that understands structured logs to realise the full benefit.

### Unstructured (formatted) logging

`slogger` also provides `fmt.Sprintf`-style methods for situations where
structured fields would be cumbersome:

```go
log.Infof("user %q logged in from %s after %v", uid, ip, latency)
```

**Advantages:**

- Familiar pattern for Go developers.
- Concise for one-off messages.

**Disadvantages:**

- Values are embedded in the message string; downstream tooling cannot parse
  individual fields reliably.
- Harder to filter, alert on, or aggregate by specific values.

**Recommendation:** Use structured fields for any value you might want to query
or alert on. Reserve formatted messages for truly free-form diagnostic output
(e.g., startup banners, developer debugging).

### Logger construction

`New` accepts zero or more functional options:

```go
log := slogger.New(
		func(o *slogger.Options) {
			o.SetLevel(slogger.DebugLevel)
			o.SetFormatter(slogger.NewTextFormatter(os.Stdout))
			o.SetOutput(os.Stdout)
			o.SetCaller(true)
			o.SetCallerSkip(0)
			o.SetName("my-service")
			o.AddFields(slogger.String("version", "1.2.3"))
		},
	)
```

If `New` is called with no options, the logger uses:

- `InfoLevel`
- `TextFormatter` writing to `os.Stderr`
- No caller capture
- No sampling, no rotation

### Formatters

#### TextFormatter

Human-readable output, ideal for development and CLI tools:

```
2026-01-15T10:00:00Z INFO  [api] server started addr=:8080 env=production
```

```go
f := slogger.NewTextFormatter(os.Stderr).
    WithTimeFormat(time.RFC3339Nano).  // custom timestamp layout
    WithEnableCaller().                // append caller=pkg/foo/bar.go:42
    WithColorMode(slogger.ColorNever). // explicit color control (recommended)
    WithDisableTimestamp()             // omit timestamp (when infra adds its own)
```

**Color Mode:**

`TextFormatter` supports three color modes via `WithColorMode()`:

| Mode                  | Behaviour                                                     |
| --------------------- | ------------------------------------------------------------- |
| `ColorAuto` (default) | Emit colours only when output is a TTY                        |
| `ColorAlways`         | Always emit ANSI colour codes                                 |
| `ColorNever`          | Never emit ANSI colour codes (for files, CI, log aggregators) |

```go
// Terminal output with auto-detection (default)
termFormatter := slogger.NewTextFormatter(os.Stdout).WithColorMode(slogger.ColorAuto)

// File output without colours
fileFormatter := slogger.NewTextFormatter(file).WithColorMode(slogger.ColorNever)

// Force colours even when not a TTY (e.g., piping to less -R)
colorFormatter := slogger.NewTextFormatter(os.Stdout).WithColorMode(slogger.ColorAlways)
```

> **Note:** When using `WithRotation()`, file output automatically uses `ColorNever`
> to prevent ANSI escape sequences from appearing in log files, regardless of the
> formatter's configured color mode.

**Legacy Methods:**

The deprecated `WithDisableColor()` and `WithEnableColor()` methods are retained for
backward compatibility but `WithColorMode()` is recommended for new code.

#### JSONFormatter

Machine-parseable single-line JSON, ideal for production and log aggregators:

```json
{
  "ts": "2026-01-15T10:00:00Z",
  "level": "INFO",
  "name": "api",
  "msg": "server started",
  "addr": ":8080"
}
```

```go
f := slogger.NewLogger().
		WithFormatter(slogger.NewJSONFormatter().
			WithTimeFormat(time.RFC3339Nano).
			WithTimeKey("timestamp").  // override default "ts"
			WithLevelKey("severity").  // override default "level"
			WithMessageKey("message"). // override default "msg"
			WithEnableCaller())        // add "caller":"pkg/foo/bar.go:42"
```

### Output writers

Any `io.Writer` can be used as the log destination:

```go
// Write to a file.
f, _ := os.OpenFile("app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
log.SetOutput(f)

// Write to multiple destinations simultaneously.
w := slogger.NewMultiWriter(os.Stdout, f)
log.SetOutput(w)

// Replace output at runtime (safe; protected by an internal mutex).
log.SetOutput(os.Stdout)
```

Convenience helpers:

```go
slogger.Stdout() // returns os.Stdout as io.Writer
slogger.Stderr() // returns os.Stderr as io.Writer
```

### Hooks

Hooks are called after every log write for their registered levels.

```go
// Define a hook.
type errorAlertHook struct {
    // ... alerting client, metrics client, etc.
}

func (h *errorAlertHook) Levels() []slogger.Level {
    return []slogger.Level{slogger.ErrorLevel, slogger.FatalLevel}
}

func (h *errorAlertHook) Fire(e *slogger.Entry) error {
    // Send to alerting system, emit a Prometheus counter, etc.
    // Do NOT retain e after this function returns.
    return nil
}

// Register the hook.
log.AddHook(&errorAlertHook{})
```

**Performance note:** hooks are called synchronously on the logging goroutine.
If a hook performs I/O (HTTP requests, database writes), consider buffering or
dispatching to a background goroutine inside the hook's `Fire` implementation.

### Log rotation

Enable automatic per-level log file rotation:

```go
log := slogger.NewLogger().
		WithLevel(slogger.InfoLevel).
		WithFormatter(slogger.NewJSONFormatter()).
		WithOutput(os.Stdout).
		WithRotation(slogger.NewRotationOptions().
			WithDirectory("logs"). // base directory; created if absent
			WithMaxBytes(50 * 1024 * 1024). // rotate at 50 MiB
			WithMaxAge(24 * time.Hour). // or after 24 hours, whichever comes first
			WithCompress(true)) // zip rotated files
```

This creates four files: `logs/debug.log`, `logs/info.log`, `logs/warn.log`,
`logs/error.log`. When a file exceeds the size or age threshold, it is moved to
`logs/archived/2006-01-02/20060102150405_<level>.zip` (or `.log` without
compression).

### Color support

Color is applied automatically when the output writer is a TTY. Colors are
disabled automatically for file writers, pipes, and CI environments.

Force-disable colors:

```go
f := slogger.NewTextFormatter(os.Stderr).WithDisableColor()
```

Color mapping:

| Level                 | Color  |
| --------------------- | ------ |
| TRACE                 | Cyan   |
| DEBUG                 | Blue   |
| INFO                  | Green  |
| WARN                  | Yellow |
| ERROR / FATAL / PANIC | Red    |

---

## Structured fields

All field constructors return a zero-allocation `Field` value:

```go
slogger.String("key", "value")
slogger.Int("count", 42)
slogger.Int64("id", int64(9007199254740993))
slogger.Float64("ratio", 3.14159)
slogger.Bool("ok", true)
slogger.Err(err)                          // key = "error"
slogger.Time("at", time.Now())
slogger.Duration("elapsed", 500*time.Millisecond)
slogger.Any("meta", anyValue)             // uses fmt.Sprintf("%v", val)
```

Fields are merged in this order when an entry is dispatched:

1. Logger-bound fields (from `With` or `Options.Fields`)
2. Context fields (from `WithContextFields`)
3. Call-site fields (passed directly to `Info`, `Error`, etc.)

---

## Child loggers

### With — add persistent fields

```go
// All entries from reqLog include request_id and user_id.
reqLog := log.With(
    slogger.String("request_id", rid),
    slogger.String("user_id",    uid),
)
reqLog.Info("handler called")
reqLog.Error("validation failed", slogger.Err(err))
```

`With` is safe to call concurrently and returns a new `*Logger` that shares the
parent's hooks, formatter, and output writer but has its own independent field
list.

### Named — hierarchical logger names

```go
dbLog  := log.Named("db")             // name = "db"
rdLog  := dbLog.Named("reader")       // name = "db.reader"
wrLog  := dbLog.Named("writer")       // name = "db.writer"

rdLog.Info("query executed")
// → 2026-01-15T10:00:00Z INFO  [db.reader] query executed
```

Names are rendered in square brackets in `TextFormatter` output and as the
`name` field in `JSONFormatter` output.

---

## Sampling

Rate-limit identical log messages to prevent log storms during error spikes:

```go
package main

import (
	"time"

	"github.com/polarixa/replify/pkg/slogger"
)

func main() {
	log := slogger.NewLogger().
		WithLevel(slogger.InfoLevel).
		WithSampling(slogger.NewSamplingOptions().
			WithFirst(10). // log the first 10 identical messages per second
			WithPeriod(time.Second). // sliding window duration
			WithThereafter(100)) // then log every 100th message

	log.Info("this message will appear 10 times in the first second, then once every 100 messages")
}
```

Sampling is keyed on the exact message string. Each unique message maintains an
independent bucket so that one chatty message does not suppress others.

Setting `Thereafter` to `0` drops all messages after the first `First` within
the window — useful for suppressing completely repetitive events.

---

## Best practices

### Microservices

Always include a `service` or `app` field so entries can be distinguished in
a shared aggregator:

```go
log := slogger.NewLogger().
    WithLevel(slogger.InfoLevel).
    WithFormatter(slogger.NewJSONFormatter()).
    WithOutput(os.Stdout).
    With(
        slogger.String("service", "order-service"),
        slogger.String("version", version),
        slogger.String("env",     os.Getenv("ENV")),
    )
```

### Background workers

Use `Named` to scope each worker's logs:

```go
for i := 0; i < numWorkers; i++ {
    workerLog := log.Named(fmt.Sprintf("worker-%d", i))
    go runWorker(ctx, workerLog)
}
```

### High-volume logging

For services that log tens of thousands of entries per second:

1. Use `JSONFormatter` — single-pass builder with no reflection.
2. Enable sampling for chatty message categories.
3. Route to a `MultiWriter` that includes a buffered writer (`bufio.NewWriter`)
   to reduce syscall frequency.
4. Use `With` to bind high-cardinality fields at the service level so they are
   not reallocated per entry.

---

## Performance considerations

### Memory allocations

- **Entry pooling** — `Entry` objects are recycled through `sync.Pool`.
  At `InfoLevel`, a typical structured log call with three `String` fields
  performs **zero heap allocations**.
- **Inline field storage** — `Field` is a value type; primitive variants
  (string, int64, float64, bool, Duration) are stored directly in the struct
  without boxing.
- **Builder-based formatting** — both `TextFormatter` and `JSONFormatter`
  use `strings.Builder` with a single `[]byte` conversion at the end, avoiding
  intermediate allocations.

### Concurrency model

- `Logger.level` is an `atomic.Int32` — level reads/writes never block.
- `Logger.formatter` and `Logger.output` are protected by a `sync.RWMutex`:
  concurrent reads (format + write) hold a read lock; `SetFormatter`/`SetOutput`
  acquire an exclusive lock.
- `Hooks` uses a separate `sync.RWMutex` — hook dispatch does not block
  concurrent log calls on the same logger.
- `LevelFileWriter` uses a per-instance `sync.Mutex` — file writes are
  serialised per writer, but multiple loggers writing to different files are
  fully concurrent.
