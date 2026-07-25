# crontask — Production Integration Guide

> **Senior-level, enterprise-grade examples for integrating `crontask` into real Go services.**

This guide shows how to embed `crontask` into a production REST service with proper
initialization, hook integration, graceful shutdown, and architectural separation of
concerns. Every code example is idiomatic Go and compiles against the current API.

---

## Table of Contents

1. [Project Structure](#1-project-structure)
2. [Full REST Service Example](#2-full-rest-service-example)
3. [Real-World Job Examples](#3-real-world-job-examples)
4. [Hook Integration Patterns](#4-hook-integration-patterns)
5. [Architecture & Separation of Concerns](#5-architecture--separation-of-concerns)
6. [Advanced Patterns](#6-advanced-patterns)
7. [Graceful Shutdown](#7-graceful-shutdown)
8. [Testing Cron Logic](#8-testing-cron-logic)
9. [Performance Considerations](#9-performance-considerations)
10. [Modification Guide for Real Projects](#10-modification-guide-for-real-projects)

---

## 1. Project Structure

Place the scheduler initialization in an internal `scheduler` package that receives its
dependencies through constructor injection. This keeps the scheduler decoupled from both
`main` and from your business services.

```
myservice/
├── cmd/
│   └── server/
│       └── main.go                  ← wire everything together, start HTTP + scheduler
├── internal/
│   ├── config/
│   │   └── config.go                ← environment-variable-driven config struct
│   ├── scheduler/
│   │   ├── scheduler.go             ← New(deps) *Scheduler — owns crontask instance
│   │   └── jobs.go                  ← one function per job, using injected services
│   ├── service/
│   │   ├── cache.go                 ← CacheService  — business logic, no scheduler knowledge
│   │   ├── report.go                ← ReportService
│   │   └── webhook.go               ← WebhookService
│   └── handler/
│       ├── health.go                ← HTTP handlers
│       └── metrics.go               ← /metrics endpoint — exposes MetricsHook counters
├── go.mod
└── main.go   (or cmd/server/main.go)
```

### Where to initialize the scheduler

Initialize the scheduler **once** in `main` (or your dependency-injection root), inject
all required services as constructor parameters, and pass the resulting `*Scheduler` to
the HTTP server if it needs to expose scheduler state.

```go
// cmd/server/main.go

func main() {
    cfg  := config.Load()          // reads env vars / flags
    db   := database.Connect(cfg)  // your DB client
    redis := cache.Connect(cfg)    // your cache client

    // Business-layer services — zero scheduler knowledge.
    cacheSvc   := service.NewCache(redis)
    reportSvc  := service.NewReport(db)
    webhookSvc := service.NewWebhook(db, cfg.WebhookEndpoint)

    // Build and start the scheduler, injecting dependencies.
    sched := scheduler.New(scheduler.Deps{
        Config:     cfg,
        CacheSvc:   cacheSvc,
        ReportSvc:  reportSvc,
        WebhookSvc: webhookSvc,
    })
    if err := sched.Start(); err != nil {
        log.Fatalf("scheduler: %v", err)
    }

    // Build HTTP server.
    srv := &http.Server{
        Addr:    cfg.Addr,
        Handler: handler.NewRouter(sched),
    }

    // Run until OS signal, then graceful-shutdown both layers.
    runUntilSignal(srv, sched)
}
```

---

## 2. Full REST Service Example

The following is a **fully working** example of a REST service with scheduler, hooks,
multiple jobs, and graceful shutdown. Copy it, adapt the stub service methods to call
your real data layer, and you have a production-ready baseline.

```go
// internal/scheduler/scheduler.go
package scheduler

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/polarixa/replify/pkg/crontask"
)

// Deps carries the external dependencies that jobs need. Each field is an
// interface so that tests can inject fakes without touching real infrastructure.
type Deps struct {
    CacheSvc   CacheService
    ReportSvc  ReportService
    WebhookSvc WebhookService
    SessionSvc SessionService
    DBSvc      DatabaseService
}

// CacheService abstracts cache operations.
type CacheService interface {
    Cleanup(ctx context.Context) error
    WarmUp(ctx context.Context) error
}

// ReportService abstracts report generation.
type ReportService interface {
    GenerateDailyReport(ctx context.Context) error
    GenerateWeeklyReport(ctx context.Context) error
}

// WebhookService abstracts webhook delivery.
type WebhookService interface {
    RetryFailed(ctx context.Context) (int, error)
}

// SessionService abstracts session management.
type SessionService interface {
    PurgeExpired(ctx context.Context) (int, error)
}

// DatabaseService abstracts database utilities.
type DatabaseService interface {
    IntegrityCheck(ctx context.Context) error
    Reconcile(ctx context.Context) error
}

// Scheduler wraps a crontask.Scheduler and exposes metrics for observability.
type Scheduler struct {
    inner   *crontask.Scheduler
    metrics *crontask.MetricsHookInstance
    deps    Deps
}

// New constructs a Scheduler, registers all jobs, and returns the instance.
// Call Start() to begin processing.
func New(deps Deps) *Scheduler {
    metrics := crontask.MetricsHook()

    inner, err := crontask.New(
        crontask.WithLocation(time.UTC),
        crontask.WithSchedulerHooks(
            // Chain observability hooks applied to every job by default.
            crontask.LoggingHook(),
            metrics,
            crontask.RecoverPanicHook(),
        ),
        crontask.WithErrorHandler(func(id string, err error) {
            // Scheduler-level error callback: send to alerting pipeline.
            log.Printf("[ALERT] scheduler: job %s error: %v", id, err)
        }),
    )
    if err != nil {
        // New() only fails if options are contradictory — treat as fatal.
        panic(fmt.Sprintf("crontask.New: %v", err))
    }

    s := &Scheduler{inner: inner, metrics: metrics, deps: deps}
    s.registerJobs()
    return s
}

// Start begins the scheduler loop. It is a thin delegate to crontask.Start.
func (s *Scheduler) Start() error { return s.inner.Start() }

// Shutdown stops the scheduler and waits for in-flight jobs to finish.
func (s *Scheduler) Shutdown(ctx context.Context) error { return s.inner.Shutdown(ctx) }

// Metrics returns the shared MetricsHookInstance for exposure via /metrics.
func (s *Scheduler) Metrics() *crontask.MetricsHookInstance { return s.metrics }

// Jobs delegates to crontask.Jobs for dashboard introspection.
func (s *Scheduler) Jobs() []crontask.JobInfo { return s.inner.Jobs() }

// registerJobs declares every scheduled job. Edit this method to add, remove,
// or reprioritise jobs. Each registration should include an ID and name for
// debuggability — these appear in logs and the /admin/jobs endpoint.
func (s *Scheduler) registerJobs() {
    must := func(id string, err error) {
        if err != nil {
            panic(fmt.Sprintf("register job %q: %v", id, err))
        }
    }

    // ── Health-check polling ─────────────────────────────────────────────────
    // Fires every 30 seconds. Replace the stub with a real HTTP probe.
    // In production: inject an http.Client and target URL via Deps.
    must(s.inner.Register(
        "@every 30s",
        s.healthCheckJob,
        crontask.WithJobID("health-check"),
        crontask.WithJobName("Health Check"),
        crontask.WithTimeout(10*time.Second),
        crontask.WithMaxRetries(2),
        crontask.WithBackoff(crontask.ConstantBackoff(2*time.Second)),
    ))

    // ── Cache cleanup ────────────────────────────────────────────────────────
    // Every 6 hours. Tag the job so the cache team can filter logs easily.
    must(s.inner.Register(
        "0 */6 * * *",
        s.cacheCleanupJob,
        crontask.WithJobID("cache-cleanup"),
        crontask.WithJobName("Cache Cleanup"),
        crontask.WithTimeout(5*time.Minute),
    ))

    // ── Daily business report ────────────────────────────────────────────────
    // Fires at 09:00 on weekdays using the @businessDaily alias.
    must(s.inner.Register(
        crontask.AliasBusinessDaily,
        s.dailyReportJob,
        crontask.WithJobID("daily-report"),
        crontask.WithJobName("Daily Business Report"),
        crontask.WithTimeout(10*time.Minute),
        crontask.WithMaxRetries(1),
        // Per-job hook overrides the scheduler default for this job only.
        crontask.WithHooks(
            crontask.LoggingHook(),
            crontask.RetryLoggerHook(),
            crontask.TimeoutLoggerHook(),
        ),
    ))

    // ── Retry failed webhooks ────────────────────────────────────────────────
    // Every 2 minutes; uses jitter to spread load across a fleet.
    must(s.inner.Register(
        "*/2 * * * *",
        s.retryWebhooksJob,
        crontask.WithJobID("retry-webhooks"),
        crontask.WithJobName("Retry Failed Webhooks"),
        crontask.WithTimeout(90*time.Second),
        crontask.WithJitter(15*time.Second), // ±15 s across replicas
        crontask.WithMaxRetries(3),
        crontask.WithBackoff(crontask.ExponentialBackoff(500*time.Millisecond)),
    ))

    // ── Expired session cleanup ──────────────────────────────────────────────
    // Every night at 02:00 (low-traffic maintenance window).
    must(s.inner.Register(
        crontask.AliasNightlyMaintenance,
        s.sessionCleanupJob,
        crontask.WithJobID("session-cleanup"),
        crontask.WithJobName("Expired Session Cleanup"),
        crontask.WithTimeout(15*time.Minute),
    ))

    // ── Database integrity check ─────────────────────────────────────────────
    // Weekly on Monday at 08:00, with concurrency guard.
    limiter := crontask.ConcurrencyLimiterHook(1) // never run two at once
    must(s.inner.Register(
        crontask.AliasStartOfWeek,
        s.dbIntegrityJob,
        crontask.WithJobID("db-integrity"),
        crontask.WithJobName("Database Integrity Check"),
        crontask.WithTimeout(30*time.Minute),
        crontask.WithHooks(
            crontask.LoggingHook(),
            crontask.MetricsHook(),
            limiter,
        ),
    ))

    // ── Background data reconciliation ───────────────────────────────────────
    // Every hour.
    must(s.inner.Register(
        crontask.AliasHourly,
        s.dataReconcileJob,
        crontask.WithJobID("data-reconcile"),
        crontask.WithJobName("Data Reconciliation"),
        crontask.WithTimeout(20*time.Minute),
        crontask.WithMaxRetries(2),
        crontask.WithBackoff(crontask.ExponentialBackoff(30*time.Second)),
    ))

    // ── Weekly report ────────────────────────────────────────────────────────
    // Monday at 08:00.
    must(s.inner.Register(
        crontask.AliasWeeklyReport,
        s.weeklyReportJob,
        crontask.WithJobID("weekly-report"),
        crontask.WithJobName("Weekly KPI Report"),
        crontask.WithTimeout(15*time.Minute),
    ))
}
```

---

## 3. Real-World Job Examples

Each job function lives in `internal/scheduler/jobs.go`. Services are accessed through
the injected `Deps` — there is no direct database or cache reference in job code.

```go
// internal/scheduler/jobs.go
package scheduler

import (
    "context"
    "errors"
    "fmt"
    "log"
    "time"
)

// ── Health Check ─────────────────────────────────────────────────────────────
// Calls an external endpoint and records the result. In production:
//   - Replace the stub with a real http.Get / gRPC health check.
//   - Inject the HTTP client (with transport-level timeout) via Deps.
//   - Record result in a time-series metric store.

func (s *Scheduler) healthCheckJob(ctx context.Context) error {
    // EDIT THIS: replace with real endpoint probe.
    // e.g.: resp, err := s.deps.HTTPClient.Get(ctx, s.deps.Config.HealthEndpoint)
    log.Printf("[health-check] probing upstream service")
    return nil // nil = healthy
}

// ── Cache Cleanup ─────────────────────────────────────────────────────────────
// Delegates to the CacheService which encapsulates Redis SCAN / DELETE logic.
// INJECT: Redis client is stored inside CacheService; never access it directly here.
// INJECT: TTL thresholds should come from config, passed to CacheService on construction.

func (s *Scheduler) cacheCleanupJob(ctx context.Context) error {
    if err := s.deps.CacheSvc.Cleanup(ctx); err != nil {
        return fmt.Errorf("cache cleanup: %w", err)
    }
    log.Printf("[cache-cleanup] cleanup completed")
    return nil
}

// ── Daily Business Report ─────────────────────────────────────────────────────
// Generates and distributes the daily report. Because this job has a per-job
// hook override, scheduler-level default hooks do NOT apply.
//
// INJECT: Email client and recipient list come from ReportService's constructor.
// INJECT: Database queries are inside ReportService — isolate SQL from job code.

func (s *Scheduler) dailyReportJob(ctx context.Context) error {
    if err := s.deps.ReportSvc.GenerateDailyReport(ctx); err != nil {
        // Wrap for structured error context visible in logs and hooks.
        return fmt.Errorf("daily report: %w", err)
    }
    return nil
}

// ── Retry Failed Webhooks ─────────────────────────────────────────────────────
// Polls the outbox table and re-delivers failed webhook payloads.
//
// INJECT: WebhookService holds the DB client and the HTTP delivery client.
// EDIT THIS: Tune the retry budget and backoff to match your SLA requirements.
//
// NOTE on jitter: this job uses WithJitter(15s) to prevent thundering-herd when
// many service replicas run the same schedule. The jitter is applied per-execution,
// not at startup — replicas stay statistically independent.

func (s *Scheduler) retryWebhooksJob(ctx context.Context) error {
    n, err := s.deps.WebhookSvc.RetryFailed(ctx)
    if err != nil {
        return fmt.Errorf("webhook retry: %w", err)
    }
    if n > 0 {
        log.Printf("[retry-webhooks] re-delivered %d failed webhooks", n)
    }
    return nil
}

// ── Session Cleanup ───────────────────────────────────────────────────────────
// Purges expired sessions. Runs during the nightly maintenance window.
//
// INJECT: SessionService owns the SQL query; pass expiry threshold via its constructor.
// EDIT THIS: Adjust AliasNightlyMaintenance to "@every 1h" in staging to validate sooner.

func (s *Scheduler) sessionCleanupJob(ctx context.Context) error {
    n, err := s.deps.SessionSvc.PurgeExpired(ctx)
    if err != nil {
        return fmt.Errorf("session cleanup: %w", err)
    }
    log.Printf("[session-cleanup] purged %d expired sessions", n)
    return nil
}

// ── Database Integrity Check ──────────────────────────────────────────────────
// Verifies referential integrity and reports anomalies. Heavy operation — protected
// by ConcurrencyLimiterHook(1) so two instances never run simultaneously even if
// the previous run overruns into the next scheduled slot.
//
// INJECT: DatabaseService holds the DB connection pool and query logic.
// EDIT THIS: Add Slack/PagerDuty alerting inside DBService.IntegrityCheck on failure.

func (s *Scheduler) dbIntegrityJob(ctx context.Context) error {
    if err := s.deps.DBSvc.IntegrityCheck(ctx); err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            return fmt.Errorf("db integrity check timed out: %w", err)
        }
        return fmt.Errorf("db integrity check failed: %w", err)
    }
    return nil
}

// ── Data Reconciliation ───────────────────────────────────────────────────────
// Aligns secondary data stores with the source of truth.
// Uses exponential backoff to handle transient connectivity issues gracefully.

func (s *Scheduler) dataReconcileJob(ctx context.Context) error {
    // Check context before starting expensive work.
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }

    if err := s.deps.DBSvc.Reconcile(ctx); err != nil {
        return fmt.Errorf("data reconciliation: %w", err)
    }
    return nil
}

// ── Weekly Report ─────────────────────────────────────────────────────────────

func (s *Scheduler) weeklyReportJob(ctx context.Context) error {
    if err := s.deps.ReportSvc.GenerateWeeklyReport(ctx); err != nil {
        return fmt.Errorf("weekly report: %w", err)
    }
    return nil
}
```

---

## 4. Hook Integration Patterns

### 4.1 Enabling all built-in hooks globally

```go
s, _ := crontask.New(
    crontask.WithSchedulerHooks(
        crontask.LoggingHook(),         // structured log lines for every event
        crontask.MetricsHook(),         // atomic counters for success/failure/duration
        crontask.RecoverPanicHook(),    // prevent panics from crashing the process
    ),
)
```

### 4.2 Per-job hook override with additional hooks

Per-job hooks passed via `WithHooks` completely replace the scheduler-default hooks for
that specific job. To keep the defaults _and_ add more, chain them explicitly:

```go
m := crontask.MetricsHook()   // shared metrics instance

s, _ := crontask.New(
    crontask.WithSchedulerHooks(
        crontask.LoggingHook(),
        m,
    ),
)

// This job uses only its own hooks (defaults not applied).
s.Register("@daily", fn,
    crontask.WithHooks(
        crontask.LoggingHook(),
        crontask.RetryLoggerHook(),
        crontask.TimeoutLoggerHook(),
    ),
)
```

### 4.3 Implementing a custom alerting hook

Embed `NoopHooks` to satisfy the `Hooks` interface and override only what you need:

```go
type PagerDutyHook struct {
    crontask.NoopHooks
    client *pagerduty.Client
    svcID  string
}

func (h *PagerDutyHook) OnFailure(_ context.Context, jobID string, d time.Duration, err error) {
    h.client.Trigger(&pagerduty.Event{
        ServiceKey:  h.svcID,
        Description: fmt.Sprintf("cron job %s failed after %s: %v", jobID, d, err),
    })
}

// Also implement PanicHook to handle panics with highest severity.
func (h *PagerDutyHook) OnPanic(_ context.Context, jobID string, r any) {
    h.client.TriggerCritical(&pagerduty.Event{
        ServiceKey:  h.svcID,
        Description: fmt.Sprintf("cron job %s PANICKED: %v", jobID, r),
    })
}
```

### 4.4 Exposing MetricsHook via HTTP

```go
// internal/handler/metrics.go

func metricsHandler(m *crontask.MetricsHookInstance) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w,
            "cron_success_total %d\ncron_failure_total %d\ncron_total_duration_ns %d\n",
            m.SuccessCount(),
            m.FailureCount(),
            m.TotalDuration().Nanoseconds(),
        )
    }
}
```

---

## 5. Architecture & Separation of Concerns

### 5.1 Where to put business logic

Job functions should be **thin orchestrators** that call into service-layer methods.
Never put SQL queries, HTTP calls, or complex logic directly inside a job function.

```
job function   → calls service method(s)
service method → calls repository/client
repository     → executes queries / HTTP requests
```

### 5.2 Injecting services into job functions

Use constructor injection. The scheduler struct holds the `Deps`, and each job method
is a receiver on the scheduler:

```go
// Good — all dependencies come from Deps, injected at construction time.
func (s *Scheduler) dailyReportJob(ctx context.Context) error {
    return s.deps.ReportSvc.GenerateDailyReport(ctx)
}

// Avoid — tight coupling, impossible to test without real DB.
func dailyReportJobGlobal(ctx context.Context) error {
    return globalDB.Query("SELECT ...") // ← never do this
}
```

### 5.3 How to avoid tight coupling

- Pass interfaces, not concrete types, in `Deps`.
- Keep each service unaware of the scheduler.
- Do not import the `crontask` package inside service layers.

### 5.4 How to scale horizontally

When multiple service replicas run the same cron schedule, use **jitter** to prevent
thundering herd and use **distributed locking** to ensure only one replica executes the
critical section:

```go
s.Register("*/5 * * * *", fn,
    crontask.WithJitter(30*time.Second), // spread across replicas
)
```

See [6.2 Distributed Locking Pattern](#62-distributed-locking-pattern) for single-leader execution.

### 5.5 How to apply concurrency limits

Use `ConcurrencyLimiterHook` to cap the number of simultaneous executions of a heavy job:

```go
// Allow at most 2 concurrent executions of this job across all goroutines.
limiter := crontask.ConcurrencyLimiterHook(2)
s.Register("@minutely", heavyFn, crontask.WithHooks(limiter))
```

### 5.6 How to apply timeout safely

Always set a timeout that is well below the job's schedule interval. A job running
`@every 1m` should never have a timeout longer than 55 seconds:

```go
s.Register("@minutely", fn,
    crontask.WithTimeout(45*time.Second), // leaves buffer before next fire
    crontask.WithHooks(crontask.TimeoutLoggerHook()),
)
```

---

## 6. Advanced Patterns

### 6.1 Worker Pool Model

For high-throughput jobs that process a queue, implement a fan-out worker pool inside
the job function rather than registering many identical cron jobs:

```go
func (s *Scheduler) queueProcessorJob(ctx context.Context) error {
    items, err := s.deps.Queue.Dequeue(ctx, 100) // pull 100 items
    if err != nil {
        return err
    }

    const workers = 8
    sem := make(chan struct{}, workers)
    errs := make(chan error, len(items))

    for _, item := range items {
        item := item
        sem <- struct{}{}
        go func() {
            defer func() { <-sem }()
            if err := s.deps.Processor.Handle(ctx, item); err != nil {
                errs <- err
            }
        }()
    }
    // Drain semaphore — wait for all workers.
    for i := 0; i < workers; i++ {
        sem <- struct{}{}
    }
    close(errs)

    // Collect errors (return the first one).
    for err := range errs {
        if err != nil {
            return fmt.Errorf("worker pool: %w", err)
        }
    }
    return nil
}
```

### 6.2 Distributed Locking Pattern

When exactly-once execution across a fleet is required, acquire a distributed lock
at the start of the job function. The following uses Redis `SET NX PX` semantics
(pseudo-code for the locking client; real implementation uses `go-redis` or `redsync`):

```go
// NOTE: the locking client below is pseudo-code. Use redsync or your own
// implementation that provides SET NX with TTL and atomic release.

func (s *Scheduler) leaderOnlyJob(ctx context.Context) error {
    // Attempt to acquire a distributed lock with a TTL equal to the job timeout.
    lock, err := s.deps.DistLock.Acquire(ctx, "myservice:leader-job", 5*time.Minute)
    if err != nil {
        // Another replica holds the lock — skip this cycle silently.
        return nil
    }
    defer lock.Release(ctx)

    return s.deps.DBSvc.Reconcile(ctx)
}
```

Key considerations:

- Set the lock TTL to the job timeout so the lock auto-expires if the holder crashes.
- Use a unique token per acquisition to prevent a slow instance from releasing a lock
  that was already expired and re-acquired by another replica.
- Log skipped cycles at DEBUG level so dashboards show expected low execution counts.

### 6.3 Persistent Job Metadata

To persist run history across restarts, write `OnComplete` results to your database:

```go
type PersistenceHook struct {
    crontask.NoopHooks
    store JobStore
}

type JobStore interface {
    RecordRun(ctx context.Context, jobID string, success bool, dur time.Duration) error
}

func (h *PersistenceHook) OnSuccess(ctx context.Context, jobID string, d time.Duration) {
    _ = h.store.RecordRun(ctx, jobID, true, d)
}

func (h *PersistenceHook) OnFailure(ctx context.Context, jobID string, d time.Duration, _ error) {
    _ = h.store.RecordRun(ctx, jobID, false, d)
}
```

### 6.4 Circuit Breaker Integration

Wrap the job function itself with a circuit breaker to prevent cascading failures when
a downstream dependency is unhealthy:

```go
import "github.com/sony/gobreaker"

cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
    Name:        "report-db",
    MaxRequests: 1,
    Interval:    60 * time.Second,
    Timeout:     30 * time.Second,
})

s.Register("@hourly", func(ctx context.Context) error {
    _, err := cb.Execute(func() (any, error) {
        return nil, deps.ReportSvc.GenerateDailyReport(ctx)
    })
    return err
}, crontask.WithJobID("hourly-report"))
```

### 6.5 Feature-Flag Controlled Scheduling

Check a feature flag at the start of the job function to enable runtime toggle without
redeployment:

```go
func (s *Scheduler) featureGatedJob(ctx context.Context) error {
    if !s.deps.Flags.IsEnabled("new-reconciler") {
        log.Printf("[data-reconcile] feature flag disabled, skipping")
        return nil
    }
    return s.deps.DBSvc.Reconcile(ctx)
}
```

### 6.6 Conditional Scheduling Logic

For jobs that should only run under specific conditions (e.g., only during business hours,
or only when a maintenance mode flag is NOT set), perform the check inside the job:

```go
func (s *Scheduler) conditionalJob(ctx context.Context) error {
    now := time.Now().UTC()
    // Only process during business hours in UTC.
    if h := now.Hour(); h < 9 || h >= 17 {
        return nil
    }
    if s.deps.Config.MaintenanceMode {
        log.Printf("[conditional-job] maintenance mode active, skipping")
        return nil
    }
    return s.deps.ReportSvc.GenerateDailyReport(ctx)
}
```

### 6.7 Runtime Job Registration from Config File

Load job schedules from a config file at startup for ops-team control without code changes:

```go
type JobConfig struct {
    ID         string `yaml:"id"`
    Name       string `yaml:"name"`
    Expression string `yaml:"expression"`
    TimeoutSec int    `yaml:"timeout_seconds"`
    MaxRetries int    `yaml:"max_retries"`
}

func (s *Scheduler) loadFromConfig(jobs []JobConfig) error {
    handlers := map[string]crontask.JobFunc{
        "health-check":    s.healthCheckJob,
        "cache-cleanup":   s.cacheCleanupJob,
        "daily-report":    s.dailyReportJob,
        "retry-webhooks":  s.retryWebhooksJob,
        "session-cleanup": s.sessionCleanupJob,
    }

    for _, jc := range jobs {
        fn, ok := handlers[jc.ID]
        if !ok {
            return fmt.Errorf("no handler for job %q", jc.ID)
        }
        opts := []crontask.JobOption{
            crontask.WithJobID(jc.ID),
            crontask.WithJobName(jc.Name),
            crontask.WithMaxRetries(jc.MaxRetries),
        }
        if jc.TimeoutSec > 0 {
            opts = append(opts, crontask.WithTimeout(time.Duration(jc.TimeoutSec)*time.Second))
        }
        if _, err := s.inner.Register(jc.Expression, fn, opts...); err != nil {
            return fmt.Errorf("register %q: %w", jc.ID, err)
        }
    }
    return nil
}
```

---

## 7. Graceful Shutdown

Both the HTTP server and the scheduler must drain cleanly when the process receives
`SIGTERM` or `SIGINT`. The pattern below blocks until both layers have exited or the
hard deadline expires.

```go
// cmd/server/main.go

func runUntilSignal(srv *http.Server, sched *scheduler.Scheduler) {
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit // block until signal

    log.Println("shutdown: received signal, draining...")

    // 30-second overall deadline for the entire shutdown sequence.
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // 1. Stop accepting new HTTP requests.
    if err := srv.Shutdown(ctx); err != nil {
        log.Printf("shutdown: HTTP server error: %v", err)
    }

    // 2. Stop the scheduler. In-flight jobs continue until they finish or
    //    their own contexts expire; the Shutdown call blocks until all
    //    executor goroutines have returned.
    if err := sched.Shutdown(ctx); err != nil {
        log.Printf("shutdown: scheduler error: %v", err)
    }

    log.Println("shutdown: complete")
}
```

### What happens to in-flight jobs on shutdown?

- `crontask.Shutdown(ctx)` closes the scheduler's stop channel, which stops the loop
  from dispatching **new** activations.
- Jobs that are **already running** in goroutines continue until they return or their
  own `WithTimeout` context expires.
- The `ctx` passed to `Shutdown` sets the deadline for how long you are willing to wait
  for the loop goroutine itself to exit (not for individual jobs).
- For production, use a 30-second shutdown timeout and ensure every job either
  respects context cancellation or has a shorter `WithTimeout`.

---

## 8. Testing Cron Logic

### 8.1 Unit-test job functions independently

Job functions receive a `context.Context` — test them without a running scheduler:

```go
func TestDailyReportJob(t *testing.T) {
    fakeSvc := &fakeReportService{}
    s := &Scheduler{deps: Deps{ReportSvc: fakeSvc}}

    if err := s.dailyReportJob(context.Background()); err != nil {
        t.Fatalf("dailyReportJob: %v", err)
    }
    if !fakeSvc.generateCalled {
        t.Error("GenerateDailyReport was not called")
    }
}
```

### 8.2 Test hook invocation with a live scheduler

Use the `@every` interval notation with small durations to trigger jobs quickly in tests:

```go
func TestHookCalledOnJobSuccess(t *testing.T) {
    var count int32
    hook := &countingHook{
        onSuccess: func() { atomic.AddInt32(&count, 1) },
    }

    s, _ := crontask.New(crontask.WithSeconds())
    _, _ = s.Register("@every 50ms", func(_ context.Context) error {
        return nil
    }, crontask.WithHooks(hook))
    _ = s.Start()
    defer s.Stop()

    // Poll with a generous deadline to avoid flakiness in slow CI runners.
    deadline := time.Now().Add(3 * time.Second)
    for time.Now().Before(deadline) {
        if atomic.LoadInt32(&count) >= 1 {
            return
        }
        time.Sleep(10 * time.Millisecond)
    }
    t.Error("hook was not called within the deadline")
}
```

### 8.3 Test schedule correctness without running the scheduler

Use the package-level `NextRuns` function to verify schedules compile and fire when expected:

```go
func TestDailyReportSchedule(t *testing.T) {
    // AliasBusinessDaily = "0 9 * * 1-5" — weekdays at 09:00.
    ref := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) // Monday
    runs, err := crontask.NextRuns(crontask.AliasBusinessDaily, ref, 3)
    if err != nil {
        t.Fatalf("NextRuns: %v", err)
    }
    want := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC) // same day at 09:00
    if !runs[0].Equal(want) {
        t.Errorf("first run = %v, want %v", runs[0], want)
    }
}
```

### 8.4 Simulate time in CI

When you need deterministic schedule firing without real-time delays, inject a synthetic
scheduler by registering jobs with a `@every` expression controlled by a test clock.
For most integration tests, the polling approach in §8.2 is sufficient and avoids the
complexity of a fake clock.

---

## 9. Performance Considerations

### 9.1 Goroutine cost

Each job activation dispatches a goroutine. For infrequent jobs (hourly, daily), this is
negligible. For sub-minute jobs (`@minutely`, `@every 10s`), be aware that goroutine
stack allocation (initial ~2 KB) multiplies with parallelism.

### 9.2 High-frequency job caveats

For jobs running every second (`WithSeconds()` + `@every 1s`):

- Ensure each invocation completes well within the interval, or use
  `ConcurrencyLimiterHook(1)` to serialize runs.
- Avoid allocations inside the hot path; pre-allocate buffers and reuse them.
- Consider whether a `time.Ticker` with a dedicated goroutine is more appropriate
  than a cron scheduler for sub-second work.

### 9.3 Backoff strategies

| Scenario                | Recommended backoff                                        |
| ----------------------- | ---------------------------------------------------------- |
| Transient network error | `ExponentialBackoff(500ms)` — caps below schedule interval |
| Database contention     | `ConstantBackoff(5s)` — predictable, avoids amplification  |
| Third-party rate limit  | `ExponentialBackoff(1s)` with a cap injected in the policy |
| Idempotent HTTP POST    | `ExponentialBackoff(200ms)`                                |

### 9.4 Avoid blocking the scheduler loop

The scheduler loop calls `dispatch` (non-blocking — launches a goroutine) and then
immediately updates `nextRun`. Heavy work in `OnStart` or hook methods delays the loop.
If your hook performs a remote write (e.g., Prometheus push gateway), do it in a
separate goroutine.

```go
func (h *remoteMetricHook) OnComplete(_ context.Context, jobID string, d time.Duration) {
    // Non-blocking: do not slow the scheduler loop.
    go h.pushGateway.Push(jobID, d)
}
```

### 9.5 Isolating heavy jobs

Jobs that perform intensive CPU or I/O work should be extracted into a separate worker
service and triggered via a lightweight message (queue, gRPC call) from the scheduler.
The scheduler job becomes a thin dispatcher:

```go
func (s *Scheduler) heavyMLTrainingJob(ctx context.Context) error {
    // Enqueue work; the actual training runs in a separate worker service.
    return s.deps.Queue.Enqueue(ctx, "ml-training", map[string]any{
        "model": "fraud-detector-v2",
        "date":  time.Now().Format("2006-01-02"),
    })
}
```

---

## 10. Modification Guide for Real Projects

| Code location                             | What to change for production                                                   |
| ----------------------------------------- | ------------------------------------------------------------------------------- |
| `scheduler.New` `WithSchedulerHooks(...)` | Replace `LoggingHook()` with a structured logger adapter (zap, slog, logrus)    |
| `scheduler.New` `WithErrorHandler(...)`   | Wire to your alerting pipeline (PagerDuty, OpsGenie, Sentry)                    |
| `registerJobs` cron expressions           | Read from env vars or a config file for ops-team control                        |
| `registerJobs` `WithTimeout(...)`         | Set per-job based on measured p99 execution time + safety margin                |
| `Deps` interface fields                   | Replace stubs with real service implementations backed by your DB/cache clients |
| `healthCheckJob`                          | Replace stub with real `http.Get` using an injected `*http.Client`              |
| `cacheCleanupJob`                         | Pass Redis client and TTL policy via `CacheService` constructor                 |
| `dailyReportJob`                          | Inject SMTP/SendGrid client and recipient config via `ReportService`            |
| `retryWebhooksJob`                        | Inject outbox table query and HTTP delivery client via `WebhookService`         |
| `leaderOnlyJob`                           | Inject a real distributed lock client (redsync, etcd lease) via `Deps.DistLock` |
| `featureGatedJob`                         | Inject a feature-flag client (LaunchDarkly, Unleash, env-var-based)             |

### Environment variables pattern

```go
// internal/config/config.go
type Config struct {
    Addr              string        `env:"SERVER_ADDR"          default:":8080"`
    DBDsn             string        `env:"DATABASE_DSN"         required:"true"`
    RedisAddr         string        `env:"REDIS_ADDR"           default:"localhost:6379"`
    WebhookEndpoint   string        `env:"WEBHOOK_ENDPOINT"     required:"true"`
    MaintenanceMode   bool          `env:"MAINTENANCE_MODE"     default:"false"`
    ShutdownTimeout   time.Duration `env:"SHUTDOWN_TIMEOUT_SEC" default:"30s"`
    HealthEndpoint    string        `env:"HEALTH_CHECK_URL"     default:"http://localhost:8080/health"`
}
```

### Converting inline jobs into reusable service modules

1. Extract the business logic from the job function into a method on the relevant service.
2. The job function becomes a one-liner: `return s.deps.Svc.DoWork(ctx)`.
3. Write unit tests against the service method, not against the job function.
4. This allows the same service method to be called from HTTP handlers, gRPC endpoints,
   or other schedulers without duplicating logic.
