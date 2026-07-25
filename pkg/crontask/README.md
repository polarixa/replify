# crontask

> **A production-grade cron and task scheduling engine for Go.**

`crontask` is the canonical scheduling sub-package of the
[replify](https://github.com/polarixa/replify) ecosystem. It provides
reliable, expressive, and observable periodic job execution built on a
clean four-layer architecture.

> 📖 **Looking for real-world integration examples?**
> See [EXAMPLE.md](./EXAMPLE.md) for a senior-level, production-oriented guide covering
> REST service integration, hook patterns, graceful shutdown, distributed locking, and more.

---

## Table of Contents

1. [Introduction to Scheduling in Go](#1-introduction-to-scheduling-in-go)
2. [Cron Expression Fundamentals](#2-cron-expression-fundamentals)
3. [Understanding Time, Timezones, and DST](#3-understanding-time-timezones-and-dst)
4. [Architecture of crontask](#4-architecture-of-crontask)
5. [Core Features Overview](#5-core-features-overview)
6. [Utility Functions](#6-utility-functions)
7. [Business Aliases](#7-business-aliases)
8. [Concurrency Model](#8-concurrency-model)
9. [Hooks & Extensibility](#9-hooks--extensibility)
10. [Performance Considerations](#10-performance-considerations)
11. [Limitations & Practical Mitigations](#11-limitations--practical-mitigations)
12. [Edge Cases in Production Environments](#12-edge-cases-in-production-environments)
13. [Real-World Usage Examples](#13-real-world-usage-examples)
14. [Comparison with robfig/cron and gronx](#14-comparison-with-robfigcron-and-gronx)

---

## 1. Introduction to Scheduling in Go

Periodic task execution is a fundamental pattern in backend systems:
database vacuums, report generation, cache warming, health checks, and
financial settlements all depend on a reliable scheduler.

Go's goroutine model makes building a scheduler straightforward—the
standard library supplies `time.Sleep`, `time.Ticker`, and `time.After`.
For _structured, observable, retryable_ scheduling however, raw
primitives are insufficient. A production scheduler must handle:

- **Expression parsing** — translating a human-authored schedule string
  ("every weekday at 9 AM") into a precise next-activation function.
- **Timezone correctness** — fire at the right wall-clock time even
  across DST transitions and negative-UTC offsets.
- **Concurrency safety** — the job registry, the scheduler loop, and
  external callers (adding/removing jobs at runtime) all run
  concurrently.
- **Observability** — each job must expose metadata (next run, last
  error, run count) and emit lifecycle callbacks for metrics.
- **Resilience** — transient failures should be retried with backoff;
  slow jobs should not block the scheduler tick.

`crontask` addresses every one of these requirements as a cohesive,
standalone package.

---

## 2. Cron Expression Fundamentals

### 2.1 The Five-Field Standard

A standard cron expression contains five space-separated fields:

```
┌─────────── minute       (0–59)
│  ┌────────── hour         (0–23)
│  │  ┌───────── day-of-month (1–31)
│  │  │  ┌──────── month       (1–12 or jan–dec)
│  │  │  │  ┌───── day-of-week (0–7, 0=Sun, 7=Sun)
│  │  │  │  │
*  *  *  *  *
```

### 2.2 Field Operators

| Operator       | Meaning         | Example                       |
| -------------- | --------------- | ----------------------------- |
| `*`            | Every value     | `*` in hour = every hour      |
| `,`            | List of values  | `1,15` in DOM = 1st and 15th  |
| `-`            | Inclusive range | `1-5` in DOW = Mon–Fri        |
| `/`            | Step            | `*/5` in minute = every 5 min |
| `/` with range | Step over range | `0-30/5` = 0,5,10,15,20,25,30 |

### 2.3 The Optional Seconds Field

`crontask` supports an optional **leading seconds** field, making it a
six-field expression when `WithSeconds()` is passed to `New()`:

```
S  M  H  DOM  MON  DOW
```

The seconds field follows the same grammar as the minutes field (0–59).

### 2.4 Named Aliases

Frequently used schedules are available as `@` aliases:

| Alias                   | Equivalent  |
| ----------------------- | ----------- |
| `@yearly` / `@annually` | `0 0 1 1 *` |
| `@monthly`              | `0 0 1 * *` |
| `@weekly`               | `0 0 * * 0` |
| `@daily` / `@midnight`  | `0 0 * * *` |
| `@hourly`               | `0 * * * *` |
| `@minutely`             | `* * * * *` |

### 2.5 Interval Expressions

The `@every <duration>` syntax fires at fixed intervals from the
scheduler start time, using Go's `time.Duration` notation:

```
@every 30s       # every 30 seconds
@every 5m        # every 5 minutes
@every 2h30m     # every 2.5 hours
```

Duration must be positive and non-zero.

---

## 3. Understanding Time, Timezones, and DST

### 3.1 Timezone-Prefixed Expressions

Individual expressions can carry an optional `TZ=` prefix:

```
TZ=America/New_York  0 9 * * 1-5
```

This causes the schedule to fire at 09:00 in New York time regardless
of the server's system timezone.

### 3.2 Scheduler-Level Timezone

Alternatively, configure all jobs in a scheduler with a single
timezone:

```go
loc, _ := time.LoadLocation("Europe/London")
s, err := crontask.New(crontask.WithLocation(loc))
```

### 3.3 DST Behaviour

Daylight Saving Time creates two anomalies:

| Event                           | Effect                                                 |
| ------------------------------- | ------------------------------------------------------ |
| Spring-forward (clock skips 1h) | Jobs scheduled in the skipped hour are silently missed |
| Fall-back (clock repeats 1h)    | Jobs in the repeated hour fire **twice**               |

`crontask` uses Go's `time.Time` representation, which stores wall
clock and monotonic readings. When the system timezone transitions,
`Next()` simply advances the wall-clock calendar, producing the correct
behaviour for all zones defined in the IANA timezone database.

**Example — checking DST behaviour:**

```go
loc, _ := time.LoadLocation("America/New_York")
sched, _ := crontask.Parse("TZ=America/New_York 30 2 * * *")

// On the spring-forward night (2024-03-10), 02:30 does not exist.
// Next() skips to 02:30 on the following day.
t := time.Date(2024, 3, 10, 2, 0, 0, 0, loc)
fmt.Println(sched.Next(t)) // 2024-03-11 02:30:00 EDT
```

---

## 4. Architecture of crontask

```
┌─────────────────────────────────────────────────┐
│  entry.go  (public API surface)                 │
│  Expression · MustParse · IsValidCronExpr       │
│  IsDue · NextRun · NextRuns · Explain           │
│  RegisterAlias · DeleteAlias                    │
└───────────────────┬─────────────────────────────┘
                    │ delegates to
        ┌───────────┴────────────────────────┐
        │                                    │
┌───────▼──────────┐          ┌─────────────▼────────────┐
│  Expression      │          │  Scheduler (scheduler.go)│
│  Layer           │          │  owns main goroutine     │
│  expression.go   │          │  tick · dispatch · stop  │
│                  │          └─────────────┬────────────┘
│                  │                        │
└───────┬──────────┘          ┌─────────────▼────────────┐
        │                     │  Job Layer (job.go)      │
        │                     │  registry · JobInfo      │
        │                     │  RWMutex-guarded         │
        │                     └─────────────┬────────────┘
        │                                   │
        │                     ┌─────────────▼─────────────┐
        │                     │  Execution Layer          │
        └─────────────────────│  executor.go              │
                              │  timeout · retry · hooks  │
                              └───────────────────────────┘
```

**Layer responsibilities:**

| Layer      | Files           | Responsibility                     |
| ---------- | --------------- | ---------------------------------- |
| Public API | `entry.go`      | All exported top-level functions   |
| Expression | `expression.go` | Parse, validate, explain           |
| Job        | `job.go`        | Registry, metadata, statistics     |
| Execution  | `executor.go`   | Goroutine dispatch, retry, backoff |
| Scheduler  | `scheduler.go`  | Main loop, tick, graceful shutdown |

---

## 5. Core Features Overview

### 5.1 Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/polarixa/replify/pkg/crontask"
)

func main() {
    s, err := crontask.New()
    if err != nil {
        log.Fatal(err)
    }

    _, err = s.Register("0 * * * *", func(ctx context.Context) error {
        fmt.Println("top of the hour:", time.Now())
        return nil
    }, crontask.WithJobName("hourly-ping"))
    if err != nil {
        log.Fatal(err)
    }

    if err := s.Start(); err != nil {
        log.Fatal(err)
    }

    // Graceful shutdown on signal.
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    _ = s.Shutdown(ctx)
}
```

### 5.2 Registering Jobs

```go
// Minimal registration.
id, err := s.Register("*/5 * * * *", myFunc)

// With options.
id, err := s.Register(
    "0 9 * * 1-5",
    myFunc,
    crontask.WithJobName("daily-report"),
    crontask.WithJobID("report-prod"),
    crontask.WithTimeout(2*time.Minute),
    crontask.WithMaxRetries(3),
    crontask.WithBackoff(crontask.ExponentialBackoff(time.Second)),
    crontask.WithJitter(5*time.Second),
    crontask.WithHooks(myHooks),
)
```

### 5.3 Scheduler Options

```go
s, err := crontask.New(
    crontask.WithSeconds(),                        // six-field mode
    crontask.WithLocation(loc),                    // timezone
    crontask.WithErrorHandler(func(id string, err error) {
        slog.Error("job failed", "id", id, "err", err)
    }),
)
```

### 5.4 Job Lifecycle Management

```go
// List all registered jobs.
for _, j := range s.Jobs() {
    fmt.Printf("%s next=%v runs=%d\n", j.Name, j.NextRun, j.RunCount)
}

// Get next N runs for a specific job.
runs, err := s.NextRuns(jobID, time.Now(), 5)

// Remove a job while the scheduler is running.
err = s.Remove(jobID)
```

---

## 6. Utility Functions

These standalone helpers work without a running `Scheduler` instance
and are safe for concurrent use.

### 6.1 Validation

```go
// Boolean validity check.
ok := crontask.IsValidCronExpr("0 9 * * 1-5") // true

// Descriptive error on failure.
err := crontask.ValidateCronExpr("99 * * * *")
// err: invalid minute value 99 (valid range 0–59)
```

### 6.2 Due-Time Check

```go
now := time.Now().Truncate(time.Second)
if crontask.IsDue("0 9 * * 1-5", now) {
    sendMorningReport()
}
```

### 6.3 Next-Run Computation

```go
// Single next activation.
next, err := crontask.NextRun("0 9 * * 1-5", time.Now())

// Five future activations.
runs, err := crontask.NextRuns("0 9 * * 1-5", time.Now(), 5)
```

### 6.4 Parsed Expression Object

```go
// MustParse panics for invalid expressions — safe for package-level vars.
var morning = crontask.MustParse("0 9 * * 1-5")

next  := morning.Next(time.Now())
due   := morning.IsDue(time.Now().Truncate(time.Second))
runs  := morning.NextN(time.Now(), 3)
raw   := morning.Raw() // "0 9 * * 1-5"
```

### 6.5 Human-Readable Explanation

```go
desc, err := crontask.Explain("0 9 * * 1-5")
// desc == "At 09:00, Monday through Friday"

desc, _ = crontask.Explain("@every 5m")
// desc == "Every 5 minutes"

desc, _ = crontask.Explain("*/30 * * * * *")
// desc == "Every 30 seconds"

desc, _ = crontask.Explain("TZ=America/New_York 0 0 * * 1-5")
// desc == "At 00:00, Monday through Friday"
```

---

## 7. Business Aliases

### 7.1 Built-In Business Aliases

`crontask` ships with aliases tailored for common business scheduling
patterns:

| Alias             | Expression         | Description                        |
| ----------------- | ------------------ | ---------------------------------- |
| `@weekdays`       | `0 0 * * 1-5`      | Midnight, Monday–Friday            |
| `@weekends`       | `0 0 * * 0,6`      | Midnight, Saturday and Sunday      |
| `@businessDaily`  | `0 9 * * 1-5`      | 09:00 every weekday                |
| `@businessHourly` | `0 9-17 * * 1-5`   | Every hour 09:00–17:00, Mon–Fri    |
| `@quarterly`      | `0 0 1 1,4,7,10 *` | 1st of each quarter                |
| `@semiMonthly`    | `0 0 1,15 * *`     | 1st and 15th of each month         |
| `@workhours`      | `* 9-17 * * 1-5`   | Every minute during business hours |
| `@marketOpen`     | `30 9 * * 1-5`     | 09:30, Mon–Fri (US market open)    |
| `@marketClose`    | `0 16 * * 1-5`     | 16:00, Mon–Fri (US market close)   |

### 7.2 Custom Alias Registration

```go
// Register a custom alias (must begin with "@").
err := crontask.RegisterAlias("@nightly", "0 2 * * *")

// Use it anywhere a cron expression is accepted.
_, err = s.Register("@nightly", myFunc)

// Explain it.
desc, _ := crontask.Explain("@nightly") // "At 02:00"

// Remove it when no longer needed.
err = crontask.DeleteAlias("@nightly")
```

**Rules:**

- Names must begin with `@`.
- The right-hand expression must be a valid five-field or six-field
  standard cron expression (no nested aliases, no `@every`).
- Names are matched case-insensitively.
- Registering an existing name overwrites it; the alias map is
  protected by a `sync.RWMutex` for concurrent safety.

---

## 8. Concurrency Model

### 8.1 Goroutine-per-Invocation

Each job invocation runs in its own goroutine, dispatched by
`executor.dispatch`. This design ensures that a slow or hung job never
delays the scheduler tick or blocks other jobs.

```
Scheduler loop goroutine
        │
        ├── tick fires
        │       │
        │       └── executor.dispatch(job1) ──▶ goroutine
        │       └── executor.dispatch(job2) ──▶ goroutine
        │
        └── tick fires (next second)
```

### 8.2 Limiting Concurrency

If unbounded parallelism is undesirable, use a semaphore in your job
function:

```go
var sem = make(chan struct{}, 4) // max 4 concurrent invocations

s.Register("* * * * *", func(ctx context.Context) error {
    select {
    case sem <- struct{}{}:
        defer func() { <-sem }()
    case <-ctx.Done():
        return ctx.Err()
    }
    return doWork(ctx)
})
```

### 8.3 WithTimeout Best Practices

Always set a timeout for jobs that perform I/O. A job without a
deadline can accumulate goroutines and exhaust file descriptors:

```go
s.Register("@every 30s", healthCheck,
    crontask.WithTimeout(10*time.Second))
```

---

## 9. Hooks & Extensibility

### 9.1 The Hooks Interface

```go
type Hooks interface {
    OnStart(ctx context.Context, jobID string)
    OnSuccess(ctx context.Context, jobID string, elapsed time.Duration)
    OnFailure(ctx context.Context, jobID string, elapsed time.Duration, err error)
    OnComplete(ctx context.Context, jobID string, elapsed time.Duration)
}
```

Embed `crontask.NoopHooks` to implement only the callbacks you need:

```go
type metricsHooks struct {
    crontask.NoopHooks
}

func (h *metricsHooks) OnSuccess(_ context.Context, id string, d time.Duration) {
    metrics.JobDuration.WithLabelValues(id).Observe(d.Seconds())
}

func (h *metricsHooks) OnFailure(_ context.Context, id string, _ time.Duration, err error) {
    metrics.JobErrors.WithLabelValues(id).Inc()
    slog.Error("job failed", "id", id, "err", err)
}
```

### 9.2 Optional Hook Interfaces

Two optional extensions are automatically dispatched by the executor when
the attached `Hooks` implementation also satisfies them:

**`RetryHook`** — called after each failed attempt that will be retried:

```go
type RetryHook interface {
    OnRetry(ctx context.Context, jobID string, attempt int, err error)
}
```

**`PanicHook`** — called when the job function panics (the panic is recovered
automatically; the job is marked as failed):

```go
type PanicHook interface {
    OnPanic(ctx context.Context, jobID string, recovered any)
}
```

### 9.3 Built-in Base Hooks

`crontask` ships a set of ready-to-use hooks that cover the most common
production needs. Enable them individually or in any combination.

| Hook                | Constructor                 | Description                                                       |
| ------------------- | --------------------------- | ----------------------------------------------------------------- |
| Logging             | `LoggingHook()`             | Logs lifecycle events via `log.Printf`                            |
| Metrics             | `MetricsHook()`             | Atomic success/failure/duration counters                          |
| Panic recovery      | `RecoverPanicHook()`        | Recovers panics; custom handler via `RecoverPanicHookWithHandler` |
| Retry logger        | `RetryLoggerHook()`         | Logs each retry attempt (`RetryHook`)                             |
| Timeout logger      | `TimeoutLoggerHook()`       | Warns on `context.DeadlineExceeded` failures                      |
| Concurrency limiter | `ConcurrencyLimiterHook(n)` | Caps simultaneous executions to `n`                               |
| Chaining            | `ChainHooks(hooks...)`      | Composes any number of hooks into one                             |

**Scheduler-level defaults** — applied to every job that does not supply its own hooks:

```go
s, _ := crontask.New(
    crontask.WithSchedulerHooks(
        crontask.LoggingHook(),
        crontask.MetricsHook(),
        crontask.RecoverPanicHook(),
    ),
)
```

**Per-job hooks** — override the scheduler defaults for a specific job:

```go
s.Register("@daily", fn,
    crontask.WithHooks(
        crontask.LoggingHook(),
        crontask.RetryLoggerHook(),
        crontask.TimeoutLoggerHook(),
    ),
)
```

**Composing multiple hooks** with `ChainHooks`:

```go
hooks := crontask.ChainHooks(
    crontask.LoggingHook(),
    crontask.MetricsHook(),
    &myCustomHook{},
)
s.Register("@hourly", fn, crontask.WithHooks(hooks))
```

See [EXAMPLE.md](./EXAMPLE.md) for production-grade integration examples.

### 9.4 Persistence Hook Pattern

`crontask` is in-memory by default. To persist job state across
restarts, define a persistence adapter around the `Hooks` interface:

```go
// PersistenceHook defines the contract for storing job execution records.
type PersistenceHook interface {
    RecordStart(jobID string, at time.Time)
    RecordResult(jobID string, at time.Time, err error, elapsed time.Duration)
    LoadLastRun(jobID string) (time.Time, error)
}

// DatabaseHook is an example adapter backed by a SQL database.
type DatabaseHook struct {
    crontask.NoopHooks
    db PersistenceHook
}

func (h *DatabaseHook) OnStart(_ context.Context, id string) {
    h.db.RecordStart(id, time.Now())
}

func (h *DatabaseHook) OnComplete(_ context.Context, id string, elapsed time.Duration) {
    h.db.RecordResult(id, time.Now(), nil, elapsed)
}
```

### 9.5 Distributed Lock Middleware

In a multi-instance deployment every scheduler fires the same jobs.
Wrap each job with a distributed lock to ensure single-execution:

```go
// redisLock acquires a Redis SET NX lock and returns true when acquired.
func redisLock(client *redis.Client, key string, ttl time.Duration) bool {
    ok, _ := client.SetNX(context.Background(), key, "1", ttl).Result()
    return ok
}

func withDistributedLock(rdb *redis.Client, jobID string, fn crontask.JobFunc) crontask.JobFunc {
    return func(ctx context.Context) error {
        if !redisLock(rdb, "crontask:lock:"+jobID, 55*time.Second) {
            return nil // another instance owns this tick
        }
        return fn(ctx)
    }
}

// Usage:
s.Register("0 * * * *", withDistributedLock(rdb, "hourly-report", myFunc))
```

---

## 10. Performance Considerations

| Concern                       | Recommendation                                                                                                |
| ----------------------------- | ------------------------------------------------------------------------------------------------------------- |
| Many jobs (1 000+)            | `nextDue` scans the registry linearly; consider partitioning into multiple schedulers by domain               |
| High-frequency jobs (< 100ms) | Use `WithSeconds()` and `@every` intervals; avoid five-field expressions where sub-minute precision is needed |
| Large job payloads            | Pass data via the job closure or a queue; never store large objects in `JobInfo`                              |
| Memory leaks                  | Always call `s.Remove(id)` for dynamically registered one-shot jobs                                           |
| Goroutine accumulation        | Set `WithTimeout` on every I/O-bound job                                                                      |

---

## 11. Limitations & Practical Mitigations

### 11.1 No Persistent Storage

**Current behaviour:** Job state (run count, last run, last error) is
held in memory. A restart resets all state.

**Mitigation:** Implement the `PersistenceHook` adapter shown in
[§9.2](#92-persistence-hook-pattern) to persist state to a database,
Redis, or a flat file. On startup, call `LoadLastRun` to reconstruct
state and conditionally skip the first missed activation.

### 11.2 No Distributed Coordination

**Current behaviour:** Running multiple instances of the same service
causes every instance to execute every job.

**Mitigation:** Use the distributed lock wrapper from [§9.3](#93-distributed-lock-middleware).
For a leader-election approach, combine with a service like etcd or
Consul to elect a single scheduler per cluster.

### 11.3 Goroutine-per-Invocation

**Current behaviour:** Each invocation spawns a goroutine. A surge of
simultaneously due jobs (or a job that never finishes) creates
unbounded goroutines.

**Mitigation:**

- Always set `WithTimeout` for any job that performs I/O.
- Use a semaphore (see [§8.2](#82-limiting-concurrency)) to cap
  parallelism.
- For CPU-bound jobs, consider a worker pool and submit work items
  instead of running logic directly in the job function.

### 11.4 No Missed-Job Detection

**Current behaviour:** Jobs missed while the scheduler is down are
never recovered.

**Mitigation (CatchUp pattern):** In the job itself, check whether the
last recorded run is older than expected and perform catch-up work:

```go
s.Register("0 9 * * 1-5", func(ctx context.Context) error {
    last, err := persistence.LoadLastRun("daily-report")
    if err == nil && time.Since(last) > 25*time.Hour {
        // Scheduler was down; run catch-up.
        return generateReportFor(ctx, last)
    }
    return generateReport(ctx)
})
```

### 11.5 DST Behaviour

See [§3.3](#33-dst-behaviour) for full details. Key points:

- Jobs in the skipped spring-forward hour are **missed** — no
  automatic recovery.
- Jobs in the repeated fall-back hour **fire twice** — implement
  idempotent job logic or use distributed locks to prevent double
  execution.

---

## 12. Edge Cases in Production Environments

| Scenario                        | Behaviour                                          | Recommendation                                               |
| ------------------------------- | -------------------------------------------------- | ------------------------------------------------------------ |
| **DST spring-forward**          | Job in skipped hour is missed                      | Accept the miss or implement CatchUp                         |
| **DST fall-back**               | Job in repeated hour fires twice                   | Use idempotent jobs + distributed locks                      |
| **Scheduler restart**           | All state reset; missed jobs not replayed          | Implement PersistenceHook + CatchUp                          |
| **High-frequency jobs**         | Goroutine-per-invocation overhead at < 10ms        | Use `@every` intervals; set `WithTimeout`                    |
| **Long-running tasks**          | Job may still run when next tick fires             | Always set `WithTimeout`; use a semaphore to cap concurrency |
| **Concurrent job modification** | `Register`/`Remove` during `Start()` is safe       | Allowed; registry uses `sync.RWMutex`                        |
| **Alias conflicts**             | `RegisterAlias` silently overwrites existing names | Use unique, namespaced alias prefixes                        |
| **Invalid expression**          | `Parse`/`Register` return `ErrInvalidExpression`   | Validate at startup with `ValidateCronExpr`                  |
| **Timezone drift**              | Long-running processes may accumulate drift        | Use `time.Now().In(loc)` and keep IANA tz data up to date    |
| **Leap year**                   | `0 0 29 2 *` fires once every four years           | Safe; `Next()` searches up to four years ahead               |

---

## 13. Real-World Usage Examples

### Email Report Every Weekday at 9 AM

```go
s.Register("@businessDaily", func(ctx context.Context) error {
    return email.SendDailySummary(ctx)
}, crontask.WithJobName("daily-email-report"),
   crontask.WithTimeout(5*time.Minute),
   crontask.WithMaxRetries(2),
   crontask.WithBackoff(crontask.ExponentialBackoff(30*time.Second)))
```

### Cleanup Job Every 6 Hours

```go
s.Register("0 */6 * * *", func(ctx context.Context) error {
    return storage.PurgeExpiredFiles(ctx)
}, crontask.WithJobName("storage-cleanup"),
   crontask.WithTimeout(10*time.Minute))
```

### Financial Settlement at Market Close

```go
s.Register("@marketClose", func(ctx context.Context) error {
    return settlement.RunEndOfDay(ctx)
}, crontask.WithJobName("eod-settlement"),
   crontask.WithTimeout(30*time.Minute),
   crontask.WithMaxRetries(1))
```

### Retry Webhook with Exponential Backoff

```go
s.Register("@every 2m", func(ctx context.Context) error {
    return webhook.Deliver(ctx, payload)
}, crontask.WithJobName("webhook-retry"),
   crontask.WithMaxRetries(5),
   crontask.WithBackoff(crontask.ExponentialBackoff(time.Second)),
   crontask.WithTimeout(30*time.Second))
```

### Health Check Every 30 Seconds

```go
s, _ := crontask.New(crontask.WithSeconds())
s.Register("@every 30s", func(ctx context.Context) error {
    return health.PingDatabase(ctx)
}, crontask.WithJobName("db-health"),
   crontask.WithTimeout(5*time.Second))
```

### Business-Only Scheduling

```go
// Only schedule on weekdays using the built-in alias.
s.Register("@weekdays", func(ctx context.Context) error {
    return cache.WarmProductCatalogue(ctx)
}, crontask.WithJobName("catalogue-warm"))
```

### Using Explain

```go
exprs := []string{
    "@businessDaily",
    "*/30 * * * * *",
    "0 0 1,15 * *",
    "@every 5m",
}
for _, expr := range exprs {
    desc, err := crontask.Explain(expr)
    if err != nil {
        log.Printf("invalid: %s — %v", expr, err)
        continue
    }
    log.Printf("%-30s → %s", expr, desc)
}
// Output:
// @businessDaily                 → At 09:00, Monday through Friday
// */30 * * * * *                 → Every 30 seconds
// 0 0 1,15 * *                   → At 00:00, on the 1st and 15th of each month
// @every 5m                      → Every 5 minutes
```

---

## 14. Comparison with robfig/cron and gronx

| Feature                      | robfig/cron | gronx | crontask             |
| ---------------------------- | ----------- | ----- | -------------------- |
| Five-field cron              | ✓           | ✓     | ✓                    |
| Six-field (seconds)          | ✓ (opt-in)  | ✓     | ✓ (opt-in)           |
| @aliases                     | ✓           | ✓     | ✓ + business aliases |
| @every intervals             | ✓           | —     | ✓                    |
| Custom alias registration    | —           | —     | ✓ (`RegisterAlias`)  |
| Human-readable Explain       | —           | —     | ✓                    |
| Per-job timeout              | —           | —     | ✓                    |
| Retry with backoff           | —           | —     | ✓                    |
| Jitter                       | —           | —     | ✓                    |
| Lifecycle hooks              | —           | —     | ✓                    |
| Job metadata & introspection | —           | —     | ✓                    |
| Graceful shutdown            | ✓           | —     | ✓                    |
| Distributed lock example     | —           | —     | documented           |
| Persistence hook pattern     | —           | —     | documented           |

**When to choose `crontask`:** You need a complete scheduling engine —
not just a parser — with built-in retry, observability hooks, a rich
expression vocabulary, and production-ready documentation.

**When to choose `robfig/cron`:** You are in an ecosystem already
built on it and only need the basic scheduler functionality.

**When to choose `gronx`:** You only need expression validation and
next-run computation without any scheduler runtime.

---

## License

Part of the [replify](https://github.com/polarixa/replify) project.
See the root `LICENSE` file for terms.
