# workergroup

A **zero-dependency** Go library for running concurrent tasks, reimagining `golang.org/x/sync/errgroup` with an object-oriented design and **live, dynamic scaling** — the concurrency limit of a group (or the worker count of a pool) can be changed at any time, even while goroutines are actively running.

## Features

- **Zero dependency** — uses only the Go standard library
- **Live rescaling** — unlike `errgroup.Group.SetLimit`, which panics if called while goroutines are active, `Group.SetLimit` and `Group.Scale` never panic and take effect immediately, even mid-flight
- **Two error-handling strategies** — `FirstError` cancels on the first failure like `errgroup`; `CollectErrors` accumulates every error (joined with `errors.Join`) without canceling, ideal for long-running supervisors
- **Persistent worker pool** — `WorkerPool` builds on `Group` to provide long-lived workers pulling `Job` values from a queue, instead of one goroutine per call, and can be scaled up or down on the fly
- **Optional panic recovery** — `WithPanicRecovery` / `WithPoolPanicRecovery` convert a task's panic into an error instead of crashing the process
- **Object-oriented ergonomics** — behavior is configured once via constructors and functional `Option`/`PoolOption` values, never by mutating exported fields

## Installation

```bash
go get github.com/polarixa/replify/pkg/workergroup
```

## Quick Start

### Error group

```go
package main

import (
    "context"
    "log"

    "github.com/polarixa/replify/pkg/workergroup"
)

func main() {
    g, ctx := workergroup.NewWithContext(context.Background(), workergroup.WithLimit(4))

    for _, url := range urls {
        url := url
        g.Go(func() error {
            return fetch(ctx, url)
        })
    }

    if err := g.Wait(); err != nil {
        log.Fatal(err)
    }
}
```

Scaling a running group requires no special handling:

```go
g.SetLimit(16) // safe even while goroutines are in flight
g.Scale(-4)    // lower the limit by 4
```

### Worker pool

```go
pool := workergroup.NewWorkerPool(context.Background(), 4)

for _, job := range jobs {
    job := job
    _ = pool.Submit(context.Background(), func(ctx context.Context) error {
        return process(ctx, job)
    })
}

pool.ScaleTo(16)   // scale up under load
pool.Close()       // stop accepting new jobs
err := pool.Wait() // wait for the queue to drain
```

## API Reference

### `Group` — concurrent tasks with a shared error outcome

```go
g := workergroup.New(workergroup.WithLimit(4))
g.Go(func() error { return doWork() })
err := g.Wait()
```

- `New(opts ...Option) *Group` — a `Group` with no associated `context.Context`
- `NewWithContext(ctx context.Context, opts ...Option) (*Group, context.Context)` — returns a `Group` and a derived `Context`, canceled according to the `ErrorMode`
- `(*Group).Go(f func() error)` — runs `f` in a new goroutine, blocking until a concurrency slot is free
- `(*Group).TryGo(f func() error) bool` — like `Go`, but only starts `f` if a slot is immediately available
- `(*Group).Fail(err error)` — reports `err` without returning from the task, so a long-running worker loop can keep going
- `(*Group).Wait() error` — blocks until all tasks finish and returns the combined error
- `(*Group).Err() error` / `(*Group).Errors() []error` — inspect the combined error, or every reported error, without blocking
- `(*Group).SetLimit(n int)` / `(*Group).Scale(delta int) int` / `(*Group).Limit() int` / `(*Group).Running() int` — read or change concurrency live

The zero `Group` is valid: unlimited concurrency, `FirstError` semantics, no context — the same defaults as a zero `errgroup.Group`.

### `Option` — configure a `Group` at construction time

```go
g := workergroup.New(
    workergroup.WithLimit(8),
    workergroup.WithErrorMode(workergroup.CollectErrors),
    workergroup.WithPanicRecovery(func(recovered any, stack []byte) {
        log.Printf("recovered panic: %v\n%s", recovered, stack)
    }),
)
```

- `WithLimit(n int)` — bounds concurrent goroutines; negative means unlimited (the default)
- `WithErrorMode(mode ErrorMode)` — `FirstError` (default) or `CollectErrors`
- `WithPanicRecovery(handler func(recovered any, stack []byte))` — recovers task panics into an error satisfying `errors.Is(err, workergroup.ErrPanic)` instead of crashing the process

### `ErrorMode` — how a `Group` combines task errors

```go
type ErrorMode int

const (
    FirstError ErrorMode = iota
    CollectErrors
)
```

- `FirstError` — keeps only the first error and cancels the derived `Context`, exactly like `errgroup`
- `CollectErrors` — accumulates every error (joined with `errors.Join`) and never cancels the context on error; suited to long-running supervisors such as `WorkerPool`

### `WorkerPool` — a persistent, rescalable pool of workers

```go
pool := workergroup.NewWorkerPool(
    context.Background(),
    4, // initial workers
    workergroup.WithQueueSize(100),
    workergroup.WithPoolErrorMode(workergroup.CollectErrors),
    workergroup.WithPoolPanicRecovery(func(recovered any, stack []byte) {
        log.Printf("job panic: %v\n%s", recovered, stack)
    }),
)
```

- `NewWorkerPool(ctx context.Context, initialWorkers int, opts ...PoolOption) *WorkerPool` — starts the pool with `initialWorkers` persistent workers
- `(*WorkerPool).Submit(ctx context.Context, job Job) error` — queues `job`, blocking until accepted, `ctx` is canceled, the pool's context is canceled, or the pool is closed
- `(*WorkerPool).TrySubmit(job Job) bool` — queues `job` only if doing so would not block
- `(*WorkerPool).ScaleTo(n int) error` / `(*WorkerPool).ScaleBy(delta int) (int, error)` — grow or shrink the worker count live; workers being stopped finish their current job first
- `(*WorkerPool).Workers() int` — current number of active workers
- `(*WorkerPool).Close()` — stops accepting new jobs and lets workers drain the queue
- `(*WorkerPool).Wait() error` — blocks until every worker has exited and returns the combined error
- `(*WorkerPool).Err() error` / `(*WorkerPool).Errors() []error` — inspect failures without blocking

`Job` is `func(ctx context.Context) error`, invoked with the pool's context.

### `PoolOption` — configure a `WorkerPool` at construction time

- `WithQueueSize(n int)` — capacity of the internal job queue; `0` (the default) makes `Submit` hand jobs directly to a ready worker
- `WithPoolErrorMode(mode ErrorMode)` — defaults to `CollectErrors`, so one failing job never stops the worker that ran it
- `WithPoolPanicRecovery(handler func(recovered any, stack []byte))` — recovers panics raised by individual jobs so one bad job can't permanently shrink the pool

## Why Live Rescaling?

`golang.org/x/sync/errgroup.Group.SetLimit` panics if called after `Go` has been invoked and goroutines are still active, which forces callers to pick a concurrency limit up front and live with it. Every `Group` and `WorkerPool` in this package is built on an internal `limiter` whose capacity can be changed at any time — goroutines blocked waiting for a free slot are woken immediately to observe the new limit — so a long-running group or pool can be scaled up under load and back down once it subsides, without restarting it.

## Real-world Example

### Processing a queue of jobs with a pool that scales with load

```go
package service

import (
    "context"
    "log"

    "github.com/polarixa/replify/pkg/workergroup"
)

// pool runs background jobs with a panic-safe CollectErrors strategy, so a
// single bad job is recorded but never brings down the other workers.
var pool = workergroup.NewWorkerPool(
    context.Background(),
    4,
    workergroup.WithPoolPanicRecovery(func(recovered any, stack []byte) {
        log.Printf("job panic: %v\n%s", recovered, stack)
    }),
)

func EnqueueJob(ctx context.Context, job workergroup.Job) error {
    return pool.Submit(ctx, job)
}

// ScaleForLoad grows or shrinks the pool in response to queue depth.
func ScaleForLoad(depth int) error {
    switch {
    case depth > 100:
        return pool.ScaleTo(16)
    case depth < 10:
        return pool.ScaleTo(4)
    default:
        return nil
    }
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
