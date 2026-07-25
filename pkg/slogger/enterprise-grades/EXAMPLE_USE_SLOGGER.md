# Real-Life slogger Usage Examples

This document provides seven production-oriented examples of slogger usage.
Each example includes a problem description, the logging strategy applied, a
complete code snippet, and an analysis of trade-offs.

---

## Example 1 — Web Server Request Logging

### Problem

An HTTP API needs to log every request with enough context to diagnose
production incidents: which endpoint was called, how long it took, what status
was returned, and which user made the request.

### Strategy

Use a Gin middleware (or `net/http` middleware) to capture request metadata into
a child `Logger` and inject the request ID into `context.Context`. All handlers
and services downstream use `log.WithContext(ctx)` to inherit these fields
automatically.

### Code

```go
package main

import (
    "net/http"
    "time"

    "github.com/polarixa/replify/pkg/slogger"
)

func loggingMiddleware(log *slogger.Logger, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        rid   := r.Header.Get("X-Request-ID")

        reqLog := log.With(
            slogger.String("request_id", rid),
            slogger.String("method",     r.Method),
            slogger.String("path",       r.URL.Path),
        )
        ctx := slogger.WithContextFields(r.Context(),
            slogger.String("request_id", rid),
        )

        rw := &statusResponseWriter{ResponseWriter: w, status: http.StatusOK}
        next.ServeHTTP(rw, r.WithContext(ctx))

        reqLog.Info("request",
            slogger.Int("status",       rw.status),
            slogger.Duration("latency", time.Since(start)),
        )
    })
}
```

### Pros and cons

| Pros                                                             | Cons                                                                         |
| ---------------------------------------------------------------- | ---------------------------------------------------------------------------- |
| Every log line in a request carries `request_id` for correlation | Request ID must be propagated via context; forgetting `WithContext` loses it |
| Level-aware: 5xx → Error, 4xx → Warn, 2xx → Info                 | Async handlers must explicitly carry ctx                                     |
| Zero allocation for `String` and `Duration` fields               | —                                                                            |

---

## Example 2 — Database Query Logging

### Problem

A data access layer needs to log slow queries (above a configurable threshold)
and all query errors, with enough context (query name, parameters, duration) to
identify bottlenecks and diagnose failures.

### Strategy

Wrap each database call with timing and log the result at the appropriate level.
The query name becomes a structured field so it can be indexed and aggregated
in a log dashboard.

### Code

```go
package repository

import (
    "context"
    "time"

    "github.com/polarixa/replify/pkg/slogger"
)

const slowQueryThreshold = 100 * time.Millisecond

type UserRepository struct {
    db  DB
    log *slogger.Logger
}

func NewUserRepository(db DB) *UserRepository {
    return &UserRepository{
        db:  db,
        log: slogger.GlobalLogger().Named("db.users"),
    }
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
    start := time.Now()
    user, err := r.db.QueryRow(ctx, `SELECT * FROM users WHERE email = $1`, email)
    elapsed := time.Since(start)

    fields := []slogger.Field{
        slogger.String("query",   "FindByEmail"),
        slogger.String("email",   email),
        slogger.Duration("latency", elapsed),
    }

    switch {
    case err != nil:
        r.log.WithContext(ctx).Error("query failed", append(fields, slogger.Err(err))...)
        return nil, err
    case elapsed > slowQueryThreshold:
        r.log.WithContext(ctx).Warn("slow query", fields...)
    default:
        r.log.WithContext(ctx).Debug("query ok", fields...)
    }

    return user, nil
}
```

### Pros and cons

| Pros                                                           | Cons                                                                |
| -------------------------------------------------------------- | ------------------------------------------------------------------- |
| `latency` is queryable — easy to build a slow-query dashboard  | Adds a timing call to every query; negligible but not zero overhead |
| Consistent field names across all repositories                 | Threshold is hardcoded; should be configurable via Options          |
| `Named("db.users")` distinguishes queries from different repos | —                                                                   |

---

## Example 3 — Background Worker Logging

### Problem

A job queue processes thousands of tasks per hour. Each task execution must be
traceable, failures must be visible, and panics must be recovered and logged
without crashing the worker pool.

### Strategy

Each worker goroutine receives a scoped child logger with its worker ID. Task
executions are bracketed with start/end log lines. A deferred recover logs
panics at `PanicLevel` so they appear in monitoring but do not kill the worker.

### Code

```go
package worker

import (
    "context"
    "fmt"
    "time"

    "github.com/polarixa/replify/pkg/slogger"
)

type Worker struct {
    id   int
    log  *slogger.Logger
    jobs <-chan Job
}

func NewWorker(id int, jobs <-chan Job) *Worker {
    return &Worker{
        id:   id,
        log:  slogger.GlobalLogger().Named(fmt.Sprintf("worker-%d", id)),
        jobs: jobs,
    }
}

func (w *Worker) Run(ctx context.Context) {
    w.log.Info("worker started")
    defer w.log.Info("worker stopped")

    for {
        select {
        case <-ctx.Done():
            return
        case job, ok := <-w.jobs:
            if !ok {
                return
            }
            w.processJob(ctx, job)
        }
    }
}

func (w *Worker) processJob(ctx context.Context, job Job) {
    start := time.Now()
    jobLog := w.log.With(
        slogger.String("job_id",   job.ID),
        slogger.String("job_type", job.Type),
    )

    defer func() {
        if r := recover(); r != nil {
            jobLog.Error("job panicked",
                slogger.Any("panic", r),
                slogger.Duration("elapsed", time.Since(start)),
            )
        }
    }()

    jobLog.Debug("job started")

    if err := job.Execute(ctx); err != nil {
        jobLog.Error("job failed",
            slogger.Err(err),
            slogger.Duration("elapsed", time.Since(start)),
        )
        return
    }

    jobLog.Info("job completed",
        slogger.Duration("elapsed", time.Since(start)),
    )
}
```

### Pros and cons

| Pros                                                                              | Cons                                                                         |
| --------------------------------------------------------------------------------- | ---------------------------------------------------------------------------- |
| Worker ID is in every log line — easy to correlate spikes with individual workers | Deferred recover adds a goroutine frame; use `runtime.Stack` for full traces |
| Panics are logged instead of silently crashing the pool                           | —                                                                            |
| `elapsed` on every job enables latency percentile analysis                        | —                                                                            |

---

## Example 4 — Microservice Structured Logging

### Problem

A microservice participates in a distributed system with OpenTelemetry tracing.
Every log line must carry `trace_id`, `span_id`, `service`, and `version` so
that log entries can be joined to trace spans in observability tools.

### Strategy

Inject the OpenTelemetry trace context into `slogger`'s context at every
service boundary. Bind `service` and `version` at logger construction time
so they appear on every entry without repetition.

### Code

```go
package main

import (
    "context"
    "os"

    "go.opentelemetry.io/otel/trace"
    "github.com/polarixa/replify/pkg/slogger"
)

func main() {
    log := slogger.New(func(o *slogger.Options) {
        o.Level     = slogger.InfoLevel
        o.Formatter = slogger.NewJSONFormatter()
        o.Output    = os.Stdout
        o.Fields    = []slogger.Field{
            slogger.String("service", "payment-service"),
            slogger.String("version", "2.4.1"),
        }
    })
    slogger.SetGlobalLogger(log)
    // ...
}

// withTraceFields enriches a context with the active OpenTelemetry span IDs.
// Call this at each service boundary (gRPC handler, HTTP handler, consumer).
func withTraceFields(ctx context.Context) context.Context {
    span := trace.SpanFromContext(ctx)
    sc   := span.SpanContext()
    if !sc.IsValid() {
        return ctx
    }
    return slogger.WithContextFields(ctx,
        slogger.String("trace_id", sc.TraceID().String()),
        slogger.String("span_id",  sc.SpanID().String()),
    )
}

func processPayment(ctx context.Context, payment Payment) error {
    ctx = withTraceFields(ctx)
    log := slogger.GlobalLogger()

    log.WithContext(ctx).Info("processing payment",
        slogger.String("payment_id",  payment.ID),
        slogger.String("currency",    payment.Currency),
        slogger.Int64("amount_cents", payment.AmountCents),
    )
    // ...
    return nil
}
```

**Sample JSON output:**

```json
{
  "ts": "2026-01-15T10:00:00Z",
  "level": "INFO",
  "name": "payment-service",
  "msg": "processing payment",
  "service": "payment-service",
  "version": "2.4.1",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "span_id": "00f067aa0ba902b7",
  "payment_id": "pay-789",
  "currency": "USD",
  "amount_cents": 4999
}
```

### Pros and cons

| Pros                                                          | Cons                                                                           |
| ------------------------------------------------------------- | ------------------------------------------------------------------------------ |
| Every log line is directly joinable to a distributed trace    | Requires OpenTelemetry instrumentation in the application                      |
| `service` and `version` are bound once, not repeated per call | Trace IDs are opaque strings; correlation requires a frontend (Jaeger, Zipkin) |
| JSON output integrates with any modern observability stack    | —                                                                              |

---

## Example 5 — CLI Application Logging

### Problem

A command-line tool needs human-readable, optionally coloured output for
interactive use, and machine-parseable output when piped to scripts or
structured via `--json`.

### Strategy

Use `TextFormatter` for interactive use (auto-detects TTY, applies ANSI colour).
When a `--json` flag is passed, switch to `JSONFormatter`. Use `DebugLevel`
during development and `InfoLevel` in normal operation.

### Code

```go
package main

import (
    "flag"
    "os"

    "github.com/polarixa/replify/pkg/slogger"
)

var (
    jsonFlag    = flag.Bool("json",    false, "Output JSON logs")
    verboseFlag = flag.Bool("verbose", false, "Enable debug output")
)

func main() {
    flag.Parse()

    var formatter slogger.Formatter
    if *jsonFlag {
        formatter = slogger.NewJSONFormatter()
    } else {
        formatter = slogger.NewTextFormatter(os.Stderr).
            WithDisableTimestamp() // terminal has a clock; timestamps are noise
    }

    level := slogger.InfoLevel
    if *verboseFlag {
        level = slogger.DebugLevel
    }

    log := slogger.New(func(o *slogger.Options) {
        o.Level     = level
        o.Formatter = formatter
        o.Output    = os.Stderr // stderr for logs; stdout for actual CLI output
    })
    slogger.SetGlobalLogger(log)

    log.Info("starting", slogger.String("command", flag.Arg(0)))

    if err := run(flag.Args()); err != nil {
        log.Error("command failed", slogger.Err(err))
        os.Exit(1)
    }

    log.Info("done")
}
```

**Interactive terminal output (coloured):**

```
 INFO  starting command=deploy
 INFO  deploying service app=my-api env=staging
 INFO  done
```

**JSON output (`--json`):**

```json
{
  "ts": "2026-01-15T10:00:00Z",
  "level": "INFO",
  "msg": "starting",
  "command": "deploy"
}
```

### Pros and cons

| Pros                                                            | Cons                                                                                 |
| --------------------------------------------------------------- | ------------------------------------------------------------------------------------ |
| Single formatter flag switches between human and machine output | Developers must remember to use `os.Stderr` for logs and `os.Stdout` for data output |
| ANSI colors only on TTY — no escape sequences in pipes          | `--verbose` only affects the root level; sub-commands cannot override independently  |
| `WithDisableTimestamp()` keeps output clean in interactive use  | —                                                                                    |

---

## Example 6 — Audit Logging

### Problem

A financial or healthcare application must log security-sensitive events
(logins, access to records, data modifications, privilege escalations) to an
immutable, append-only store for compliance (SOC 2, HIPAA, PCI-DSS).

### Strategy

Create a dedicated **audit logger** separate from the application logger.
Use a structured format with mandatory fields for every audit event. Write to
a file with rotation and optional remote shipping via a hook.

### Code

```go
package audit

import (
    "context"
    "os"
    "time"

    "github.com/polarixa/replify/pkg/slogger"
)

var auditLog *slogger.Logger

// Init initialises the package-level audit logger.
// Must be called once during application bootstrap.
func Init() error {
    lfw, err := slogger.NewLevelFileWriter(slogger.RotationOptions{
        Dir:      "/var/log/myapp/audit",
        MaxBytes: 500 * 1024 * 1024, // 500 MiB — audits are important; keep large files
        MaxAge:   30 * 24 * time.Hour,
        Compress: true,
    })
    if err != nil {
        return err
    }

    hook := slogger.NewLevelWriterHook(
        lfw,
        slogger.NewJSONFormatter(),
        slogger.InfoLevel, // audit events are all INFO
    )

    auditLog = slogger.New(func(o *slogger.Options) {
        o.Level     = slogger.InfoLevel
        o.Formatter = slogger.NewJSONFormatter()
        o.Output    = os.Stdout // also goes to stdout for streaming
        o.Name      = "audit"
        o.Fields    = []slogger.Field{
            slogger.String("log_type", "audit"),
        }
    })
    auditLog.AddHook(hook)
    return nil
}

// Event logs a mandatory-retention audit event.
//
// Every audit event must include actor, action, and resource.
// Additional context (outcome, reason, metadata) is highly recommended.
func Event(ctx context.Context, actor, action, resource string, fields ...slogger.Field) {
    base := []slogger.Field{
        slogger.String("actor",    actor),
        slogger.String("action",   action),
        slogger.String("resource", resource),
        slogger.String("event_id", generateEventID()),
        slogger.Time("occurred_at", time.Now().UTC()),
    }
    auditLog.WithContext(ctx).Info("audit.event",
        append(base, fields...)...,
    )
}

func generateEventID() string {
    return time.Now().Format("20060102150405.000000000")
}
```

**Usage:**

```go
audit.Event(ctx,
    claims.UserID,
    "READ",
    "patient_record:"+recordID,
    slogger.String("outcome", "success"),
    slogger.String("ip",      clientIP),
)
```

**Sample JSON output:**

```json
{
  "ts": "2026-01-15T10:00:00Z",
  "level": "INFO",
  "name": "audit",
  "msg": "audit.event",
  "log_type": "audit",
  "actor": "u-42",
  "action": "READ",
  "resource": "patient_record:r-99",
  "event_id": "20260115100000.123456789",
  "occurred_at": "2026-01-15T10:00:00Z",
  "outcome": "success",
  "ip": "10.0.1.5",
  "request_id": "req-001",
  "trace_id": "4bf92f35..."
}
```

### Pros and cons

| Pros                                                                      | Cons                                                                                                |
| ------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------- |
| Mandatory fields (`actor`, `action`, `resource`) ensure schema compliance | A schema contract enforced only by convention; no compile-time guarantee                            |
| Separate logger prevents audit noise from mixing with application logs    | Audit logger must never be silenced — `FatalLevel` + `os.Exit(1)` would lose events                 |
| File rotation keeps archives small and manageable                         | Compliance requirements may mandate a tamper-evident log store (use a hook to ship to WORM storage) |

---

## Example 7 — File-Based Production Logging

### Problem

A production service running on bare metal or a VM (not containerised) must
write logs to local files, rotate them on schedule, compress old archives, and
retain them for a configurable period.

### Strategy

Use `LevelFileWriter` for per-level file rotation. Write JSON to both stdout
(for real-time monitoring) and rotating files (for retention). Configure
aggressive rotation thresholds and enable compression to conserve disk space.

### Code

```go
package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/polarixa/replify/pkg/slogger"
)

func main() {
    // File rotation configuration.
    rotOpts := slogger.RotationOptions{
        Dir:      "/var/log/my-service",
        MaxBytes: 100 * 1024 * 1024, // 100 MiB per level
        MaxAge:   24 * time.Hour,    // daily rotation regardless of size
        Compress: true,
    }

    lfw, err := slogger.NewLevelFileWriter(rotOpts)
    if err != nil {
        fmt.Fprintf(os.Stderr, "failed to open log files: %v\n", err)
        os.Exit(1)
    }
    defer lfw.Close()

    jsonFmt   := slogger.NewJSONFormatter()
    fileHook  := slogger.NewLevelWriterHook(lfw, jsonFmt)

    log := slogger.New(func(o *slogger.Options) {
        o.Level     = slogger.InfoLevel
        o.Formatter = jsonFmt
        o.Output    = os.Stdout // stdout for systemd journal / journalctl
        o.Fields    = []slogger.Field{
            slogger.String("service", "my-service"),
            slogger.String("host",    mustHostname()),
        }
    })
    log.AddHook(fileHook)
    slogger.SetGlobalLogger(log)

    log.Info("service started")

    // Handle SIGHUP for log rotation (e.g. triggered by logrotate).
    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGHUP)
    go func() {
        for range sigs {
            log.Info("rotating log files")
            if err := lfw.Rotate(); err != nil {
                log.Error("rotation failed", slogger.Err(err))
            }
        }
    }()

    // Application main loop ...
    select {}
}

func mustHostname() string {
    h, err := os.Hostname()
    if err != nil {
        return "unknown"
    }
    return h
}
```

**Archive structure after several rotations:**

```
/var/log/my-service/
├── debug.log                          (active, current day)
├── info.log                           (active, current day)
├── warn.log                           (active, current day)
├── error.log                          (active, current day)
└── archived/
    ├── 2026-01-13/
    │   ├── 20260113000000_info.zip
    │   └── 20260113000000_error.zip
    ├── 2026-01-14/
    │   ├── 20260114000000_debug.zip
    │   ├── 20260114000000_info.zip
    │   ├── 20260114000000_warn.zip
    │   └── 20260114000000_error.zip
    └── 2026-01-15/
        └── 20260115060000_info.zip   (morning rotation)
```

### Pros and cons

| Pros                                                                             | Cons                                                                        |
| -------------------------------------------------------------------------------- | --------------------------------------------------------------------------- |
| Per-level files allow `grep`/`tail` to focus on severity                         | Four open file handles per service instance; manageable but worth tracking  |
| Compressed archives (`.zip`) reduce disk usage by 80-95% for typical log content | Archived files must be explicitly decompressed before `grep`/analysis       |
| SIGHUP integration allows logrotate(8) compatibility                             | Go's `os.Rename` is atomic on Linux but not on all OSes                     |
| `host` field in every entry simplifies multi-node debugging                      | For Kubernetes, prefer stdout and let the container runtime handle rotation |
