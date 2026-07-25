# Integrating `crontask` into a Production Gin Service

> **Audience:** Principal engineers, staff engineers, and senior engineers designing or
> reviewing a production microservice that embeds scheduled background work.  
> **Scope:** End-to-end architectural blueprint — from folder structure through
> PostgreSQL persistence, Redis distributed locking, YAML-driven configuration,
> observability, scaling, and graceful shutdown.

---

## Table of Contents

1. [Enterprise Folder Structure](#1-enterprise-folder-structure)
2. [Gin Integration](#2-gin-integration)
3. [PostgreSQL Job State Persistence](#3-postgresql-job-state-persistence)
4. [Redis Distributed Lock Integration](#4-redis-distributed-lock-integration)
5. [YAML Configuration for Tasks](#5-yaml-configuration-for-tasks)
6. [Full Job Lifecycle Explanation](#6-full-job-lifecycle-explanation)
7. [Graceful Shutdown Strategy](#7-graceful-shutdown-strategy)
8. [Observability & Monitoring](#8-observability--monitoring)
9. [Scaling Strategy](#9-scaling-strategy)
10. [Production Pitfalls & Edge Cases](#10-production-pitfalls--edge-cases)

---

## 1. Enterprise Folder Structure

```
myservice/
├── cmd/
│   └── api/
│       └── main.go                    ← binary entry-point; wires all layers
├── internal/
│   ├── config/
│   │   └── config.go                  ← Config struct; loaded from YAML + env overrides
│   ├── database/
│   │   └── postgres.go                ← *sql.DB / pgxpool bootstrap
│   ├── scheduler/
│   │   ├── scheduler.go               ← New(Deps) wraps crontask.Scheduler
│   │   └── jobs.go                    ← one exported function per cron job
│   ├── service/
│   │   ├── health.go                  ← HealthService
│   │   ├── cleanup.go                 ← CleanupService
│   │   └── report.go                  ← ReportService
│   ├── handler/
│   │   ├── health.go                  ← GET /health, GET /ready
│   │   └── metrics.go                 ← GET /metrics (Prometheus scrape endpoint)
│   └── repository/
│       └── job_state.go               ← JobStateRepository — reads/writes cron_job_states
├── pkg/
│   └── crontask/                      ← (this library; vendored or module-dependency)
├── config/
│   └── config.yaml                    ← default runtime configuration
├── migrations/
│   └── 001_create_cron_job_states.sql ← DDL managed by a migration tool
├── go.mod
└── go.sum
```

### Architectural rationale

| Layer                 | Responsibility                        | Why separate?                                                    |
| --------------------- | ------------------------------------- | ---------------------------------------------------------------- |
| `cmd/api`             | Wire dependencies; own `main`         | Single binary entry-point, zero business logic                   |
| `internal/config`     | Load + validate YAML/env config       | Centralises all runtime knobs; swappable at test time            |
| `internal/database`   | Open + health-check DB pool           | Pool lifecycle is independent of scheduler lifecycle             |
| `internal/scheduler`  | Own the `crontask.Scheduler` instance | Encapsulates job registration + hook wiring away from HTTP layer |
| `internal/service`    | Business logic                        | Zero scheduler knowledge; injectable; independently testable     |
| `internal/handler`    | HTTP handlers                         | Depend on services and, optionally, scheduler for introspection  |
| `internal/repository` | DB queries for job state              | Isolates SQL behind an interface; mocked in unit tests           |

**Rule of thumb:** the scheduler is infrastructure, not business logic. It lives in
`internal/scheduler` exactly as the HTTP server lives in `internal/handler`. Both are
wired together in `cmd/api/main.go` and share nothing except injected services.

---

## 2. Gin Integration

### 2.1 `internal/config/config.go`

```go
package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config is the top-level runtime configuration struct.
// Fields tagged with `env` can be overridden by environment variables at startup.
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Crontask CrontaskConfig `yaml:"crontask"`
}

type ServerConfig struct {
	Port            int           `yaml:"port"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

type DatabaseConfig struct {
	DSN             string        `yaml:"dsn"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

type RedisConfig struct {
	Address  string        `yaml:"address"`
	Password string        `yaml:"password"`
	DB       int           `yaml:"db"`
	LockTTL  time.Duration `yaml:"lock_ttl"`
}

type CrontaskConfig struct {
	Timezone string       `yaml:"timezone"`
	Tasks    []TaskConfig `yaml:"tasks"`
}

type TaskConfig struct {
	Name              string        `yaml:"name"`
	Expression        string        `yaml:"expression"`
	Timeout           time.Duration `yaml:"timeout"`
	Retry             int           `yaml:"retry"`
	Jitter            time.Duration `yaml:"jitter"`
	DistributedLock   bool          `yaml:"distributed_lock"`
}

// Load reads the YAML file at path and applies environment-variable overrides.
// It returns a validated Config or a descriptive error.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("config: open %s: %w", path, err)
	}
	defer f.Close()

	var cfg Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("config: decode %s: %w", path, err)
	}

	// Environment-variable overrides for secrets that must not live in YAML.
	if v := os.Getenv("DATABASE_DSN"); v != "" {
		cfg.Database.DSN = v
	}
	if v := os.Getenv("REDIS_ADDRESS"); v != "" {
		cfg.Redis.Address = v
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		cfg.Redis.Password = v
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config: validation: %w", err)
	}
	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Database.DSN == "" {
		return fmt.Errorf("database.dsn is required")
	}
	if c.Redis.Address == "" {
		return fmt.Errorf("redis.address is required")
	}
	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}
	if c.Server.ShutdownTimeout == 0 {
		c.Server.ShutdownTimeout = 30 * time.Second
	}
	if c.Redis.LockTTL == 0 {
		c.Redis.LockTTL = 30 * time.Second
	}
	if c.Crontask.Timezone == "" {
		c.Crontask.Timezone = "UTC"
	}
	return nil
}
```

### 2.2 `internal/scheduler/scheduler.go`

```go
package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/polarixa/replify/pkg/crontask"

	"myservice/internal/config"
	"myservice/internal/repository"
	"myservice/internal/service"
)

// Deps carries all injectable dependencies for the scheduler layer.
// All fields are required; zero-value checks are performed in New.
type Deps struct {
	Config       *config.Config
	HealthSvc    *service.HealthService
	CleanupSvc   *service.CleanupService
	ReportSvc    *service.ReportService
	JobStateRepo repository.JobStateRepository
	LockAcquirer LockAcquirer // see §4
}

// Scheduler wraps crontask.Scheduler and owns all registered jobs.
// Expose it only as the interface your handlers actually need.
type Scheduler struct {
	inner *crontask.Scheduler
	deps  Deps
}

// New constructs the Scheduler, registers all jobs from config, and returns
// the ready-to-start instance. Call Start() separately so the caller controls
// the exact moment background work begins.
func New(deps Deps) (*Scheduler, error) {
	loc, err := time.LoadLocation(deps.Config.Crontask.Timezone)
	if err != nil {
		return nil, fmt.Errorf("scheduler: invalid timezone %q: %w", deps.Config.Crontask.Timezone, err)
	}

	metrics := crontask.MetricsHook()

	inner, err := crontask.New(
		crontask.WithLocation(loc),
		crontask.WithSchedulerHooks(
			crontask.LoggingHook(),
			metrics,
			crontask.RecoverPanicHook(),
		),
		crontask.WithErrorHandler(func(id string, err error) {
			log.Printf("[scheduler] job %s terminal failure: %v", id, err)
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("scheduler: init: %w", err)
	}

	s := &Scheduler{inner: inner, deps: deps}
	if err := s.registerAll(); err != nil {
		return nil, err
	}
	return s, nil
}

// registerAll reads deps.Config.Crontask.Tasks and registers each job.
// This is the single place where task configuration drives job registration;
// new tasks are added to config.yaml, not to Go source.
func (s *Scheduler) registerAll() error {
	for _, t := range s.deps.Config.Crontask.Tasks {
		t := t // capture range variable
		fn, err := s.resolveJobFunc(t.Name)
		if err != nil {
			return fmt.Errorf("scheduler: unknown task %q: %w", t.Name, err)
		}

		// Wrap with distributed lock if configured.
		if t.DistributedLock {
			fn = withDistributedLock(s.deps.LockAcquirer, t.Name, s.deps.Config.Redis.LockTTL, fn)
		}

		// Wrap with job state persistence.
		fn = withStatePersistence(s.deps.JobStateRepo, t.Name, fn)

		opts := buildJobOptions(t)
		if _, err := s.inner.Register(t.Expression, fn, opts...); err != nil {
			return fmt.Errorf("scheduler: register %q: %w", t.Name, err)
		}
		log.Printf("[scheduler] registered job %q (%s)", t.Name, t.Expression)
	}
	return nil
}

// resolveJobFunc maps a task name to its implementation.
// Adding a new task = add an entry here + add it to config.yaml.
func (s *Scheduler) resolveJobFunc(name string) (crontask.JobFunc, error) {
	switch name {
	case "health-check":
		return healthCheckJob(s.deps.HealthSvc), nil
	case "cleanup":
		return cleanupJob(s.deps.CleanupSvc), nil
	case "weekly-report":
		return weeklyReportJob(s.deps.ReportSvc), nil
	default:
		return nil, fmt.Errorf("no implementation registered")
	}
}

// buildJobOptions converts a TaskConfig into the corresponding crontask options.
func buildJobOptions(t config.TaskConfig) []crontask.JobOption {
	opts := []crontask.JobOption{
		crontask.WithJobName(t.Name),
		crontask.WithJobID(t.Name),
	}
	if t.Timeout > 0 {
		opts = append(opts, crontask.WithTimeout(t.Timeout))
	}
	if t.Retry > 0 {
		opts = append(opts, crontask.WithMaxRetries(t.Retry))
		opts = append(opts, crontask.WithBackoff(crontask.ExponentialBackoff(500*time.Millisecond)))
	}
	if t.Jitter > 0 {
		opts = append(opts, crontask.WithJitter(t.Jitter))
	}
	return opts
}

// Start begins the scheduler loop in a background goroutine.
func (s *Scheduler) Start() error {
	return s.inner.Start()
}

// Shutdown stops the scheduler and blocks until the loop exits or ctx expires.
func (s *Scheduler) Shutdown(ctx context.Context) error {
	return s.inner.Shutdown(ctx)
}

// Jobs returns a snapshot of all registered jobs for introspection endpoints.
func (s *Scheduler) Jobs() []crontask.JobInfo {
	return s.inner.Jobs()
}
```

### 2.3 `internal/scheduler/jobs.go`

```go
package scheduler

import (
	"context"

	"github.com/polarixa/replify/pkg/crontask"

	"myservice/internal/service"
)

// healthCheckJob returns a JobFunc that verifies downstream dependency health.
// It is deliberately thin: all business logic lives in HealthService.
func healthCheckJob(svc *service.HealthService) crontask.JobFunc {
	return func(ctx context.Context) error {
		return svc.CheckAll(ctx)
	}
}

// cleanupJob returns a JobFunc that purges expired records from the database.
func cleanupJob(svc *service.CleanupService) crontask.JobFunc {
	return func(ctx context.Context) error {
		return svc.PurgeExpiredRecords(ctx)
	}
}

// weeklyReportJob returns a JobFunc that generates and dispatches the weekly report.
func weeklyReportJob(svc *service.ReportService) crontask.JobFunc {
	return func(ctx context.Context) error {
		return svc.GenerateWeeklyReport(ctx)
	}
}
```

### 2.4 `cmd/api/main.go` — full wiring

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"myservice/internal/config"
	"myservice/internal/database"
	"myservice/internal/handler"
	"myservice/internal/repository"
	"myservice/internal/scheduler"
	"myservice/internal/service"
)

func main() {
	cfgPath := envOrDefault("CONFIG_PATH", "config/config.yaml")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// ── Infrastructure ─────────────────────────────────────────────────────
	db, err := database.Open(cfg.Database)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer rdb.Close()

	// ── Repositories ───────────────────────────────────────────────────────
	jobStateRepo := repository.NewJobStateRepository(db)

	// ── Services ───────────────────────────────────────────────────────────
	healthSvc := service.NewHealthService(db, rdb)
	cleanupSvc := service.NewCleanupService(db)
	reportSvc := service.NewReportService(db)

	// ── Scheduler ──────────────────────────────────────────────────────────
	sched, err := scheduler.New(scheduler.Deps{
		Config:       cfg,
		HealthSvc:    healthSvc,
		CleanupSvc:   cleanupSvc,
		ReportSvc:    reportSvc,
		JobStateRepo: jobStateRepo,
		LockAcquirer: scheduler.NewRedisLockAcquirer(rdb, cfg.Redis.LockTTL),
	})
	if err != nil {
		log.Fatalf("scheduler: %v", err)
	}
	if err := sched.Start(); err != nil {
		log.Fatalf("scheduler start: %v", err)
	}

	// ── HTTP server ────────────────────────────────────────────────────────
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(requestLogger())

	handler.RegisterRoutes(router, healthSvc, sched)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// ── Run until signal ───────────────────────────────────────────────────
	runUntilSignal(srv, sched, cfg.Server.ShutdownTimeout)
}

// runUntilSignal blocks until SIGINT/SIGTERM, then shuts down in order:
//  1. Stop accepting new HTTP requests.
//  2. Wait for in-flight HTTP requests to drain.
//  3. Stop the scheduler loop.
//  4. Wait for in-flight jobs to drain (or timeout).
func runUntilSignal(srv *http.Server, sched *scheduler.Scheduler, timeout time.Duration) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("[main] shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Shut down HTTP first so no new requests arrive while jobs are draining.
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[main] HTTP shutdown error: %v", err)
	}

	// Then stop the scheduler.
	if err := sched.Shutdown(ctx); err != nil {
		log.Printf("[main] scheduler shutdown error: %v", err)
	}

	log.Println("[main] shutdown complete")
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Printf("[http] %s %s %d %s",
			c.Request.Method, c.Request.URL.Path,
			c.Writer.Status(), time.Since(start))
	}
}
```

### 2.5 `internal/handler/routes.go`

```go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/polarixa/replify/pkg/crontask"

	"myservice/internal/service"
)

type schedulerInspector interface {
	Jobs() []crontask.JobInfo
}

// RegisterRoutes wires all HTTP routes onto r.
func RegisterRoutes(r *gin.Engine, healthSvc *service.HealthService, sched schedulerInspector) {
	r.GET("/health", healthHandler(healthSvc))
	r.GET("/ready", readyHandler(healthSvc))
	r.GET("/scheduler/jobs", jobsHandler(sched))
}

func healthHandler(svc *service.HealthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := svc.CheckAll(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func readyHandler(svc *service.HealthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	}
}

// jobsHandler exposes the current job registry for operational dashboards.
func jobsHandler(sched schedulerInspector) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, sched.Jobs())
	}
}
```

---

## 3. PostgreSQL Job State Persistence

### 3.1 DDL

```sql
-- migrations/001_create_cron_job_states.sql

CREATE TYPE cron_job_status AS ENUM (
    'pending',
    'running',
    'succeeded',
    'failed',
    'retrying',
    'skipped'
);

CREATE TABLE cron_job_states (
    id                BIGSERIAL         PRIMARY KEY,

    -- Stable identifier that matches crontask.WithJobID / TaskConfig.Name.
    job_id            TEXT              NOT NULL,

    -- Human-readable label for dashboards.
    job_name          TEXT              NOT NULL DEFAULT '',

    -- Raw cron expression registered for this job.
    expression        TEXT              NOT NULL DEFAULT '',

    -- Lifecycle state machine value.
    status            cron_job_status   NOT NULL DEFAULT 'pending',

    -- Wall-clock time this invocation was dispatched by the scheduler.
    scheduled_at      TIMESTAMPTZ,

    -- Wall-clock time the job function was actually invoked (after jitter).
    started_at        TIMESTAMPTZ,

    -- Wall-clock time the invocation completed (succeeded, failed, or panicked).
    finished_at       TIMESTAMPTZ,

    -- Elapsed execution time in milliseconds; NULL while still running.
    duration_ms       BIGINT,

    -- One-based attempt number; incremented on each retry.
    attempt           INT               NOT NULL DEFAULT 1,

    -- Total number of retries attempted across the lifetime of this invocation.
    retry_count       INT               NOT NULL DEFAULT 0,

    -- Maximum retries configured for the job (denormalised for audit clarity).
    max_retries       INT               NOT NULL DEFAULT 0,

    -- Serialised error string from the last failed attempt; NULL on success.
    last_error        TEXT,

    -- Predicted next activation time; updated after each invocation.
    next_run_at       TIMESTAMPTZ,

    -- Audit columns.
    created_at        TIMESTAMPTZ       NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ       NOT NULL DEFAULT NOW()
);

-- ── Indexes ────────────────────────────────────────────────────────────────

-- Primary operational lookup: "give me the latest state for job X."
CREATE INDEX idx_cron_job_states_job_id
    ON cron_job_states (job_id, created_at DESC);

-- Scheduler introspection: "which jobs are due next?"
CREATE INDEX idx_cron_job_states_next_run
    ON cron_job_states (next_run_at)
    WHERE next_run_at IS NOT NULL;

-- Partial index: fast retrieval of all currently running jobs (for health checks).
CREATE INDEX idx_cron_job_states_running
    ON cron_job_states (job_id, started_at)
    WHERE status = 'running';

-- Failed-job alerting queue: find all unacknowledged failures quickly.
CREATE INDEX idx_cron_job_states_failed
    ON cron_job_states (job_id, finished_at DESC)
    WHERE status = 'failed';

-- ── Unique constraint ──────────────────────────────────────────────────────

-- Prevent duplicate in-flight rows for the same job. The application layer
-- must INSERT with ON CONFLICT DO NOTHING and fall back to a state update.
CREATE UNIQUE INDEX uq_cron_job_states_running_per_job
    ON cron_job_states (job_id)
    WHERE status IN ('running', 'retrying');

-- ── Trigger: keep updated_at current ───────────────────────────────────────

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    IF NEW IS DISTINCT FROM OLD THEN
        NEW.updated_at = NOW();
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_cron_job_states_updated_at
BEFORE UPDATE ON cron_job_states
FOR EACH ROW EXECUTE FUNCTION set_updated_at();
```

### 3.2 Column rationale

| Column                       | Purpose                                                                                      |
| ---------------------------- | -------------------------------------------------------------------------------------------- |
| `job_id`                     | Stable identity key matching `crontask.WithJobID`; used to correlate rows across invocations |
| `status`                     | State machine; drives alerting rules and the partial index on running jobs                   |
| `scheduled_at`               | When the scheduler decided to fire the job; enables drift analysis                           |
| `started_at` / `finished_at` | Compute actual execution window; detect long-running jobs                                    |
| `duration_ms`                | Pre-computed for dashboard queries; avoids repeated timestamp arithmetic                     |
| `attempt` / `retry_count`    | Distinguish first attempt from retries; feeds SLA calculations                               |
| `last_error`                 | Persisted for post-mortem without requiring log correlation                                  |
| `next_run_at`                | Enables "jobs overdue" alerting even when the scheduler itself is down                       |
| `created_at` / `updated_at`  | Immutable audit trail; `updated_at` maintained by DB trigger                                 |

### 3.3 Repository implementation

```go
package repository

import (
	"context"
	"database/sql"
	"time"
)

// JobState is the read model returned by the repository.
type JobState struct {
	ID          int64
	JobID       string
	JobName     string
	Status      string
	ScheduledAt *time.Time
	StartedAt   *time.Time
	FinishedAt  *time.Time
	DurationMs  *int64
	Attempt     int
	RetryCount  int
	LastError   *string
	NextRunAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UpsertParams carries the fields written on each scheduler tick.
type UpsertParams struct {
	JobID      string
	JobName    string
	Expression string
	Status     string
	ScheduledAt *time.Time
	StartedAt   *time.Time
	FinishedAt  *time.Time
	DurationMs  *int64
	Attempt     int
	RetryCount  int
	MaxRetries  int
	LastError   *string
	NextRunAt   *time.Time
}

// JobStateRepository defines the persistence contract.
// The interface boundary allows mock injection in unit tests.
type JobStateRepository interface {
	Upsert(ctx context.Context, p UpsertParams) error
	LatestByJobID(ctx context.Context, jobID string) (*JobState, error)
	ListRunning(ctx context.Context) ([]JobState, error)
}

type pgJobStateRepo struct {
	db *sql.DB
}

// NewJobStateRepository returns a Postgres-backed JobStateRepository.
func NewJobStateRepository(db *sql.DB) JobStateRepository {
	return &pgJobStateRepo{db: db}
}

// Upsert inserts a new row or updates the existing running row for the job.
// The unique partial index on (job_id) WHERE status IN ('running','retrying')
// enforces single-in-flight semantics at the DB level.
func (r *pgJobStateRepo) Upsert(ctx context.Context, p UpsertParams) error {
	const q = `
INSERT INTO cron_job_states
    (job_id, job_name, expression, status,
     scheduled_at, started_at, finished_at, duration_ms,
     attempt, retry_count, max_retries, last_error, next_run_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
ON CONFLICT (job_id) WHERE status IN ('running','retrying')
DO UPDATE SET
    status       = EXCLUDED.status,
    started_at   = EXCLUDED.started_at,
    finished_at  = EXCLUDED.finished_at,
    duration_ms  = EXCLUDED.duration_ms,
    attempt      = EXCLUDED.attempt,
    retry_count  = EXCLUDED.retry_count,
    last_error   = EXCLUDED.last_error,
    next_run_at  = EXCLUDED.next_run_at,
    updated_at   = NOW()`

	_, err := r.db.ExecContext(ctx, q,
		p.JobID, p.JobName, p.Expression, p.Status,
		p.ScheduledAt, p.StartedAt, p.FinishedAt, p.DurationMs,
		p.Attempt, p.RetryCount, p.MaxRetries, p.LastError, p.NextRunAt,
	)
	return err
}

func (r *pgJobStateRepo) LatestByJobID(ctx context.Context, jobID string) (*JobState, error) {
	const q = `
SELECT id, job_id, job_name, status,
       scheduled_at, started_at, finished_at, duration_ms,
       attempt, retry_count, last_error, next_run_at,
       created_at, updated_at
FROM   cron_job_states
WHERE  job_id = $1
ORDER BY created_at DESC
LIMIT  1`

	row := r.db.QueryRowContext(ctx, q, jobID)
	var s JobState
	err := row.Scan(
		&s.ID, &s.JobID, &s.JobName, &s.Status,
		&s.ScheduledAt, &s.StartedAt, &s.FinishedAt, &s.DurationMs,
		&s.Attempt, &s.RetryCount, &s.LastError, &s.NextRunAt,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func (r *pgJobStateRepo) ListRunning(ctx context.Context) ([]JobState, error) {
	const q = `
SELECT id, job_id, job_name, status, started_at, created_at, updated_at
FROM   cron_job_states
WHERE  status IN ('running','retrying')
ORDER BY started_at`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []JobState
	for rows.Next() {
		var s JobState
		if err := rows.Scan(&s.ID, &s.JobID, &s.JobName, &s.Status,
			&s.StartedAt, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
```

### 3.4 Persistence wrapper

The `withStatePersistence` middleware wraps any `JobFunc` and records state transitions
before and after execution without modifying the job's business logic.

```go
package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/polarixa/replify/pkg/crontask"

	"myservice/internal/repository"
)

// withStatePersistence wraps fn, recording state transitions in cron_job_states.
func withStatePersistence(repo repository.JobStateRepository, jobID string, fn crontask.JobFunc) crontask.JobFunc {
	return func(ctx context.Context) error {
		now := time.Now()

		if err := repo.Upsert(ctx, repository.UpsertParams{
			JobID:       jobID,
			Status:      "running",
			ScheduledAt: &now,
			StartedAt:   &now,
		}); err != nil {
			log.Printf("[scheduler] persist start state for %s: %v", jobID, err)
		}

		err := fn(ctx)

		finished := time.Now()
		durationMs := finished.Sub(now).Milliseconds()
		status := "succeeded"
		var lastErr *string
		if err != nil {
			status = "failed"
			s := err.Error()
			lastErr = &s
		}

		if uerr := repo.Upsert(ctx, repository.UpsertParams{
			JobID:      jobID,
			Status:     status,
			StartedAt:  &now,
			FinishedAt: &finished,
			DurationMs: &durationMs,
			LastError:  lastErr,
		}); uerr != nil {
			log.Printf("[scheduler] persist final state for %s: %v", jobID, uerr)
		}

		return err
	}
}
```

---

## 4. Redis Distributed Lock Integration

In a horizontally scaled deployment, every running instance of the service has its own
`crontask.Scheduler`. Without coordination, each instance fires the job simultaneously.
Redis SET NX (set-if-not-exists) with a per-key TTL is the canonical solution for
coarse-grained distributed mutual exclusion.

### 4.1 Lock abstraction

```go
package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// LockAcquirer is the interface the scheduler uses to request a distributed lock.
// The concrete implementation is injected; tests can substitute a no-op.
type LockAcquirer interface {
	// Acquire attempts to obtain an exclusive lock for key with the given TTL.
	// It returns (true, release, nil) on success; (false, nil, nil) when
	// another holder owns the lock; (false, nil, err) on infrastructure failure.
	// release must be called when the critical section completes — even on error.
	Acquire(ctx context.Context, key string, ttl time.Duration) (acquired bool, release func(context.Context) error, err error)
}

type redisLockAcquirer struct {
	rdb     *redis.Client
	lockTTL time.Duration
}

// NewRedisLockAcquirer returns an Acquirer backed by Redis SET NX.
func NewRedisLockAcquirer(rdb *redis.Client, lockTTL time.Duration) LockAcquirer {
	return &redisLockAcquirer{rdb: rdb, lockTTL: lockTTL}
}

// Acquire performs a Redis SET NX PX with a unique token as value.
// The release function issues a Lua-script DEL that checks the token,
// preventing a slow holder from deleting a lock that was re-acquired by
// a different instance after expiry.
func (r *redisLockAcquirer) Acquire(ctx context.Context, key string, ttl time.Duration) (bool, func(context.Context) error, error) {
	token := fmt.Sprintf("%d", time.Now().UnixNano())
	ok, err := r.rdb.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return false, nil, fmt.Errorf("redis lock acquire %q: %w", key, err)
	}
	if !ok {
		return false, nil, nil
	}

	release := func(ctx context.Context) error {
		// Atomic check-and-delete: only delete if we still own the key.
		const script = `
if redis.call("get", KEYS[1]) == ARGV[1] then
    return redis.call("del", KEYS[1])
else
    return 0
end`
		return r.rdb.Eval(ctx, script, []string{key}, token).Err()
	}
	return true, release, nil
}
```

### 4.2 Lock wrapper for job functions

```go
package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/polarixa/replify/pkg/crontask"
)

// withDistributedLock wraps fn so it only executes when the caller acquires
// the distributed lock for jobID. If the lock is held by another instance,
// the invocation is skipped (returns nil — not an error).
func withDistributedLock(
	acquirer LockAcquirer,
	jobID string,
	ttl time.Duration,
	fn crontask.JobFunc,
) crontask.JobFunc {
	return func(ctx context.Context) error {
		lockKey := fmt.Sprintf("crontask:lock:%s", jobID)
		acquired, release, err := acquirer.Acquire(ctx, lockKey, ttl)
		if err != nil {
			// Infrastructure failure — log and skip rather than error, to avoid
			// triggering retry logic for a transient Redis hiccup.
			log.Printf("[scheduler] lock acquire error for %s: %v — skipping invocation", jobID, err)
			return nil
		}
		if !acquired {
			log.Printf("[scheduler] job %s skipped — lock held by another instance", jobID)
			return nil
		}
		defer func() {
			if rerr := release(ctx); rerr != nil {
				log.Printf("[scheduler] lock release error for %s: %v", jobID, rerr)
			}
		}()
		return fn(ctx)
	}
}
```

### 4.3 Lock expiration strategy

**TTL must exceed the maximum expected job execution time** — including any retry delays
configured via `WithBackoff`.

```
lock_ttl = max_job_duration + (max_retries × max_backoff_delay) + safety_margin
```

For a job with `timeout=30s`, `retry=3`, `ExponentialBackoff(500ms)`, and a safety
margin of 10 s:

```
lock_ttl = 30s + (3 × 4s) + 10s = 52s  →  round up to 60s
```

Expose `lock_ttl` in `config.yaml` (`redis.lock_ttl`) so it can be tuned per environment
without a code change.

### 4.4 Failure scenarios

| Scenario                                | Outcome                                                                         | Mitigation                                                        |
| --------------------------------------- | ------------------------------------------------------------------------------- | ----------------------------------------------------------------- |
| Instance crashes while holding lock     | Lock expires after TTL; another instance picks up the next scheduled tick       | Set `lock_ttl` conservatively; ensure jobs are idempotent         |
| Redis connection lost during job        | Lock key orphaned; `release` returns error (logged, not propagated)             | TTL natural expiry; circuit-breaker on Redis client               |
| Clock drift between instances           | SET NX TTL is server-side on Redis; client clock skew is irrelevant             | No additional action needed                                       |
| Job takes longer than TTL               | Another instance may re-acquire the lock mid-execution; duplicate work possible | Enforce `crontask.WithTimeout` ≤ `lock_ttl`; make jobs idempotent |
| Lua `release` races with natural expiry | DEL is a no-op; no phantom deletion                                             | The Lua check-and-delete pattern handles this atomically          |

### 4.5 Idempotency requirements

A distributed lock is a **best-effort** guard, not a correctness guarantee. Every job
that uses `distributed_lock: true` **must** be idempotent:

- **Database writes:** use `INSERT … ON CONFLICT DO NOTHING` or upsert patterns.
- **External API calls:** use idempotency keys in HTTP headers.
- **File operations:** write to a temp path, then rename (atomic on POSIX).
- **Message publishing:** use at-least-once delivery with consumer-side deduplication.

---

## 5. YAML Configuration for Tasks

### 5.1 `config/config.yaml`

```yaml
server:
  port: 8080
  read_timeout: 10s
  write_timeout: 30s
  shutdown_timeout: 30s

database:
  # Override in production via DATABASE_DSN environment variable.
  dsn: "postgres://app:secret@localhost:5432/myservice?sslmode=disable"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m

redis:
  # Override in production via REDIS_ADDRESS / REDIS_PASSWORD environment variables.
  address: "localhost:6379"
  password: ""
  db: 0
  lock_ttl: 60s

crontask:
  timezone: "UTC"
  tasks:
    - name: "health-check"
      expression: "@every 30s"
      timeout: 5s
      retry: 0
      jitter: 2s
      distributed_lock: false

    - name: "cleanup"
      expression: "0 */6 * * *"
      timeout: 30s
      retry: 1
      jitter: 10s
      distributed_lock: true

    - name: "weekly-report"
      expression: "0 8 * * 1"
      timeout: 120s
      retry: 2
      jitter: 30s
      distributed_lock: true
```

### 5.2 Dynamic registration pattern

`scheduler.New` calls `registerAll()`, which iterates `cfg.Crontask.Tasks` and calls
`resolveJobFunc(name)` to look up the implementation. This pattern:

- Keeps job registration **data-driven** — operators tune schedules without Go changes.
- Fails **fast at startup** if a task name in YAML has no implementation.
- Allows per-environment `config.yaml` files (dev, staging, prod) with different
  expressions and retry budgets without branching in Go code.

### 5.3 Environment-variable override pattern

Production secrets (DSN, Redis password) must never appear in YAML committed to source
control. The `config.Load` function applies a simple precedence rule:

```
env var > config.yaml > struct default
```

For Kubernetes deployments, inject secrets via `envFrom: secretRef` in the pod spec.
For plain Docker, pass `--env DATABASE_DSN=...`.

### 5.4 Validating the config at startup

`config.validate()` enforces required fields and applies sensible defaults. Fail-fast
at startup is preferred over silent misconfiguration that only manifests at the first
scheduled tick. All validation errors are surfaced to `log.Fatalf` in `main.go` before
the process binds to any port.

---

## 6. Full Job Lifecycle Explanation

```
Scheduler tick (every second or next-due interval)
│
├── Is job due? (nextRun ≤ now)
│   └── No  → sleep until next due time; goto tick
│
├── Yes → dispatch(entry, now)  [runs in a new goroutine]
│
│   ├── 1. Jitter sleep         (random [0, jitter))
│   │
│   ├── 2. Lock acquisition     (if distributed_lock=true)
│   │   ├── Acquired → continue
│   │   └── Not acquired → log skip; return nil (no retry)
│   │
│   ├── 3. Derive execution context
│   │   └── context.WithTimeout(base, timeout)  [if timeout > 0]
│   │
│   ├── 4. hooks.OnStart(ctx, jobID)
│   │
│   ├── 5. Persist state → status=running
│   │
│   ├── 6. fn(ctx)              ← actual job function
│   │   ├── Success → goto 8
│   │   └── Error →
│   │       ├── attempt < maxRetries → hooks.OnRetry; backoff sleep; goto 6
│   │       └── attempt == maxRetries → goto 7
│   │
│   ├── 7. (Failure path)
│   │   ├── hooks.OnFailure(ctx, jobID, duration, err)
│   │   ├── Persist state → status=failed, last_error=err
│   │   └── onError callback (if registered on Scheduler)
│   │
│   ├── 8. (Success path)
│   │   ├── hooks.OnSuccess(ctx, jobID, duration)
│   │   └── Persist state → status=succeeded
│   │
│   ├── 9. hooks.OnComplete(ctx, jobID, duration)
│   │
│   ├── 10. Release distributed lock
│   │
│   └── 11. Emit Prometheus metrics (via MetricsHookInstance counters)
│
└── Update entry.nextRun = schedule.Next(now)
    Emit next_run_at to cron_job_states
```

### State transition diagram

```
pending → running → succeeded
                 ↘ retrying → succeeded
                            ↘ failed
                 ↘ failed   (first attempt, no retries)
                 ↘ skipped  (lock not acquired)
```

---

## 7. Graceful Shutdown Strategy

### 7.1 Signal handling and ordering

```go
// From cmd/api/main.go — runUntilSignal (shown in §2.4)
//
// Shutdown order:
//   1. srv.Shutdown(ctx)  — stop accepting HTTP connections; drain in-flight requests.
//   2. sched.Shutdown(ctx) — close stopCh; wait for loop goroutine to exit via doneCh.
//
// The shared ctx enforces a hard deadline across both layers.
```

**Why this order matters:**

1. Stopping HTTP first ensures no new requests arrive that might trigger a job via an
   admin endpoint while the scheduler is tearing down.
2. The scheduler's `Shutdown` closes `stopCh`, which causes `loop()` to return at the
   next iteration boundary. In-flight job goroutines are **not** killed; they continue
   until their own context or timeout expires.
3. The outer `context.WithTimeout(background, shutdownTimeout)` enforces a hard wall;
   if jobs are still running after `shutdownTimeout`, `Shutdown` returns `context.DeadlineExceeded`.

### 7.2 Job drain behaviour

`crontask.Scheduler.Shutdown` does **not** wait for in-flight job goroutines; it only
waits for the scheduler loop itself to exit. To drain jobs, pass the jobs a cancellable
context derived from the application context:

```go
// In scheduler.New, propagate an application-level context to each job.
appCtx, appCancel := context.WithCancel(context.Background())

// When shutting down:
appCancel()                    // signals all job contexts to cancel
sched.inner.Shutdown(ctx)      // waits for the loop to stop
```

Use `WithContext(appCtx)` when registering long-running jobs so they honour the
application shutdown signal:

```go
inner.Register(expr, fn,
    crontask.WithContext(appCtx),
    crontask.WithTimeout(30*time.Second),
)
```

### 7.3 Failure modes

| Failure                                  | Behaviour                                                                                             | Mitigation                                                                          |
| ---------------------------------------- | ----------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------- |
| Job ignores context cancellation         | Job runs past shutdown deadline; process killed by orchestrator after `terminationGracePeriodSeconds` | Ensure all blocking I/O uses the ctx parameter                                      |
| DB connection closed before job finishes | Job returns error; persisted as `failed` on next startup                                              | Use connection-pool `max_lifetime` and graceful pool close after scheduler shutdown |
| Redis lock released after process exit   | TTL expiry handles recovery; next pod picks up next tick                                              | Lock TTL must be less than Kubernetes `terminationGracePeriodSeconds`               |

### 7.4 Kubernetes configuration recommendation

```yaml
# Align with the shutdown_timeout in config.yaml.
terminationGracePeriodSeconds: 60

lifecycle:
  preStop:
    exec:
      command: ["/bin/sleep", "5"] # drain load-balancer connections
```

---

## 8. Observability & Monitoring

### 8.1 Structured logging

Replace `log.Printf` with a structured logger (`slog`, `zap`, or `logrus`) in
production. The pattern below uses `slog` (stdlib since Go 1.21):

```go
// In internal/scheduler/scheduler.go
import "log/slog"

crontask.WithErrorHandler(func(id string, err error) {
    slog.Error("crontask job terminal failure",
        slog.String("job_id", id),
        slog.String("error", err.Error()),
    )
}),
```

Key log events and their recommended fields:

| Event              | Fields                                      |
| ------------------ | ------------------------------------------- |
| Job started        | `job_id`, `job_name`, `scheduled_at`        |
| Job succeeded      | `job_id`, `duration_ms`                     |
| Job failed         | `job_id`, `duration_ms`, `attempt`, `error` |
| Job retrying       | `job_id`, `attempt`, `error`, `backoff_ms`  |
| Lock skipped       | `job_id`, `lock_key`                        |
| Scheduler shutdown | `timeout_ms`, `jobs_in_flight`              |

### 8.2 Prometheus metrics

Expose `MetricsHookInstance` counters through a custom Prometheus collector:

```go
package handler

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/polarixa/replify/pkg/crontask"
)

// metricsCollector bridges MetricsHookInstance to prometheus.Collector.
type metricsCollector struct {
	hook       *crontask.MetricsHookInstance
	successes  *prometheus.Desc
	failures   *prometheus.Desc
	panics     *prometheus.Desc
	totalDurNs *prometheus.Desc
}

func NewMetricsCollector(hook *crontask.MetricsHookInstance) prometheus.Collector {
	ns := "crontask"
	return &metricsCollector{
		hook: hook,
		successes: prometheus.NewDesc(
			ns+"_job_successes_total",
			"Total number of successful job invocations.", nil, nil),
		failures: prometheus.NewDesc(
			ns+"_job_failures_total",
			"Total number of failed job invocations (all retries exhausted).", nil, nil),
		panics: prometheus.NewDesc(
			ns+"_job_panics_total",
			"Total number of job invocations that panicked.", nil, nil),
		totalDurNs: prometheus.NewDesc(
			ns+"_job_duration_nanoseconds_total",
			"Cumulative nanoseconds spent in job execution.", nil, nil),
	}
}

func (c *metricsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.successes
	ch <- c.failures
	ch <- c.panics
	ch <- c.totalDurNs
}

func (c *metricsCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(c.successes, prometheus.CounterValue,
		float64(c.hook.SuccessCount()))
	ch <- prometheus.MustNewConstMetric(c.failures, prometheus.CounterValue,
		float64(c.hook.FailureCount()))
	ch <- prometheus.MustNewConstMetric(c.panics, prometheus.CounterValue,
		float64(c.hook.PanicCount()))
	ch <- prometheus.MustNewConstMetric(c.totalDurNs, prometheus.CounterValue,
		float64(c.hook.TotalDuration()))
}

// MetricsHandler wires the collector into a Gin route.
func MetricsHandler(hook *crontask.MetricsHookInstance) gin.HandlerFunc {
	reg := prometheus.NewRegistry()
	reg.MustRegister(NewMetricsCollector(hook))
	h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
```

Register on the router:

```go
router.GET("/metrics", handler.MetricsHandler(metricsHook))
```

### 8.3 Alerting strategy

| Alert             | Condition                                                   | Severity |
| ----------------- | ----------------------------------------------------------- | -------- |
| Job not running   | `next_run_at < NOW() - 2×interval`                          | Warning  |
| High failure rate | `failures / (successes + failures) > 0.05` over 15 min      | Critical |
| Long-running job  | `status = 'running' AND started_at < NOW() - timeout × 1.5` | Warning  |
| Scheduler down    | No `scheduler loop` log line in last 60 s                   | Critical |
| Lock starvation   | Same instance always wins lock (others always skip)         | Warning  |

### 8.4 SLA considerations

- Define per-job SLA windows in config (e.g. `cleanup` must finish within 6 h of
  scheduled time) and enforce them via `WithTimeout`.
- Store `scheduled_at` and `finished_at` in `cron_job_states` to compute scheduling
  latency (time from `scheduled_at` to `started_at`) and execution latency separately.
- Query the DB for jobs that have not completed within their SLA window and page on-call.

---

## 9. Scaling Strategy

### 9.1 Horizontal scaling with distributed locking

Every replica runs its own `crontask.Scheduler`. The `withDistributedLock` wrapper
(§4.2) ensures that only one replica executes a given job at any tick. Replicas that
fail to acquire the lock log a skip and return nil — no retry, no error.

```
Pod A  →  lock acquired  →  fn() executes
Pod B  →  lock NOT acquired → skip (no-op)
Pod C  →  lock NOT acquired → skip (no-op)
```

This design is **AP** (available, partition-tolerant) under CAP: if Redis is
unreachable, every pod skips the job rather than risking a duplicate or deadlock.

### 9.2 Leader election alternative

For jobs where exactly-once semantics are critical (e.g. billing runs), consider
replacing the Redis SET NX pattern with a proper leader-election library
(`etcd` lease, `Consul` session, or `k8s` Lease object). The `withDistributedLock`
interface accepts any `LockAcquirer` implementation, so swapping the backend requires
only a new struct that satisfies the interface.

### 9.3 Worker service isolation

When the scheduler carries significant CPU or I/O load, isolate it in a dedicated
`worker` binary separate from the HTTP `api` binary:

```
cmd/
├── api/main.go     ← HTTP only; no scheduler
└── worker/main.go  ← scheduler only; no HTTP (or minimal health endpoint)
```

Deploy the `worker` as a single replica (or use a `Deployment` with `replicas: 1`
guarded by a `PodDisruptionBudget`) and scale the `api` independently. The distributed
lock is still needed if you ever scale the worker to >1 replica for availability.

### 9.4 High-frequency job partitioning

For jobs running every second or sub-second, a single goroutine-per-tick model breaks
down. Partition the work:

1. Register one job that assigns N work units per tick.
2. Distribute units to a worker pool (`golang.org/x/sync/errgroup` or channel-based).
3. Limit concurrency via `ConcurrencyLimiterHook`.

```go
limiter := crontask.ConcurrencyLimiterHook(8) // max 8 parallel sub-workers

inner.Register("@every 1s", func(ctx context.Context) error {
    units := partitioner.Assign(ctx) // returns work units for this tick
    g, gctx := errgroup.WithContext(ctx)
    for _, u := range units {
        u := u
        g.Go(func() error { return processUnit(gctx, u) })
    }
    return g.Wait()
}, crontask.WithHooks(limiter))
```

---

## 10. Production Pitfalls & Edge Cases

### 10.1 DST (Daylight Saving Time) transitions

When the scheduler's timezone transitions into DST, one hour is skipped; when
transitioning out, one hour repeats.

- **Skipped hour:** jobs with expressions that fall in the skipped window (e.g.
  `0 2 * * *` in `America/New_York` on the spring-forward Sunday) will not fire
  until the next matching time.
- **Repeated hour:** jobs may fire twice.

**Mitigation:** schedule high-value jobs in UTC (`crontask.WithLocation(time.UTC)`)
and use offset expressions for apparent local times. Alternatively, accept DST
skips/doubles and make jobs idempotent.

### 10.2 Long-running jobs

A job that runs longer than its scheduled interval causes the next tick to start before
the previous one finishes. `crontask` dispatches each due job in its own goroutine;
concurrent executions are possible by design.

**Mitigation:**

- Set `WithTimeout` to a value strictly less than the job interval.
- Use `ConcurrencyLimiterHook(1)` for jobs that must never overlap.
- Monitor `status = 'running'` rows whose `started_at` is older than expected.

### 10.3 Database transaction timeouts

PostgreSQL's `statement_timeout` and `lock_timeout` session parameters are separate
from the application-level context timeout. A job holding a long transaction may be
killed by the DB but the Go context is still active.

**Mitigation:**

- Always pass the job's `ctx` to every database call (`db.ExecContext`, `db.QueryContext`).
- Set `statement_timeout` in the DSN: `postgres://...?options=statement_timeout%3D25000`.
- Keep the application timeout (`WithTimeout`) slightly shorter than the DB timeout.

### 10.4 Redis lock expiry race

If a job takes longer than the lock TTL, the key expires and another instance may
acquire the lock while the first is still executing.

**Mitigation:**

- Enforce `crontask.WithTimeout` ≤ `lock_ttl` as a startup validation:

```go
for _, t := range cfg.Crontask.Tasks {
    if t.DistributedLock && t.Timeout > 0 && t.Timeout > cfg.Redis.LockTTL {
        return fmt.Errorf("task %q: timeout (%s) exceeds redis.lock_ttl (%s)",
            t.Name, t.Timeout, cfg.Redis.LockTTL)
    }
}
```

- Implement lock extension (refresh TTL mid-job) for jobs with unpredictable duration.

### 10.5 Idempotency requirements

Any job that may execute more than once (due to retries, lock race, or duplicate
scheduling) must be idempotent. Verify idempotency at the design stage:

| Operation       | Idempotent approach                                             |
| --------------- | --------------------------------------------------------------- |
| Insert record   | `ON CONFLICT DO NOTHING`                                        |
| Send email      | Track sent emails in DB with unique `(job_id, recipient, date)` |
| Charge payment  | Use payment gateway's idempotency key                           |
| Publish message | Use `message_id` + consumer-side deduplication                  |

### 10.6 Duplicate scheduling risk

Registering the same job ID twice will create two entries that fire independently:

```go
// WRONG — registers two jobs both named "cleanup"
inner.Register("0 */6 * * *", cleanupFn, crontask.WithJobID("cleanup"))
inner.Register("0 */6 * * *", cleanupFn, crontask.WithJobID("cleanup"))
```

`crontask` does not deduplicate on job ID at registration time. The `registerAll`
function in §2.2 is the canonical call site; ensure it is called exactly once and
does not loop over duplicated task names.

### 10.7 YAML misconfiguration

A typo in a cron expression causes `Register` to return an error, which `registerAll`
surfaces as a startup failure (fast-fail). Validate expressions in CI:

```go
// In a test file: TestConfigExpressions
for _, t := range cfg.Crontask.Tasks {
    _, err := crontask.Parse(t.Expression) // or crontask.Validate(t.Expression)
    if err != nil {
        t.Errorf("task %q: invalid expression %q: %v", t.Name, t.Expression, err)
    }
}
```

### 10.8 Schema migration risks

The `cron_job_states` table is written by every job invocation. Schema changes require
care:

- **Adding a nullable column:** safe; existing inserts are unaffected.
- **Adding a NOT NULL column without a default:** will break existing inserts;
  always provide a `DEFAULT`.
- **Renaming a column:** requires application-side coordination (dual-write during
  transition or a maintenance window).
- **Dropping a column:** remove application references first; deploy; then `ALTER TABLE
DROP COLUMN`.
- **Index changes:** use `CREATE INDEX CONCURRENTLY` in Postgres to avoid table locks.

Use a migration tool (`golang-migrate`, `goose`, or `Flyway`) and enforce that
migrations are applied before the application starts (`migrate up` as an init-container
or `main.go` pre-flight).

---

## Summary

This blueprint elevates `crontask` from a standalone scheduler library into an
enterprise-ready scheduling framework by:

1. **Encapsulating** scheduler initialization behind `internal/scheduler`, keeping
   `main.go` as a pure wiring layer.
2. **Externalising** all task configuration to `config.yaml`, with environment-variable
   overrides for secrets.
3. **Persisting** job state to PostgreSQL with a production DDL that supports
   dashboards, alerting, and post-mortem analysis.
4. **Coordinating** multi-instance deployments with Redis SET NX distributed locks that
   degrade gracefully on Redis outage.
5. **Observing** the scheduler through `MetricsHookInstance`, Prometheus, structured
   logging, and targeted alerting rules.
6. **Shutting down** cleanly by draining HTTP before stopping the scheduler, enforcing
   a hard timeout, and surfacing errors to the operator.
7. **Documenting** the edge cases — DST, long-running jobs, lock drift, idempotency,
   schema migrations — so the team can design defensively from day one.
