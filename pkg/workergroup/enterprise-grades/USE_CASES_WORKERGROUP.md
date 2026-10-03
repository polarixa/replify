# workergroup — When and How to Use It

> **Senior-level, enterprise-grade case catalog for `workergroup`.**

This guide answers one question exhaustively: **given a concrete concurrency
problem, which `workergroup` primitive should you reach for, and how should you
wire it up?** Every case below is idiomatic Go, compiles against the current
API, and is paired with the symptom that tells you it applies.

---

## Table of Contents

- [workergroup — When and How to Use It](#workergroup--when-and-how-to-use-it)
  - [Table of Contents](#table-of-contents)
  - [1. Decision Matrix](#1-decision-matrix)
  - [2. Case Catalog: `Group`](#2-case-catalog-group)
    - [2.1 Bounded fan-out of independent I/O calls](#21-bounded-fan-out-of-independent-io-calls)
    - [2.2 Fail-fast pipelines with context cancellation](#22-fail-fast-pipelines-with-context-cancellation)
    - [2.3 Best-effort batch processing that must not stop early](#23-best-effort-batch-processing-that-must-not-stop-early)
    - [2.4 Opportunistic work that should never block the caller](#24-opportunistic-work-that-should-never-block-the-caller)
    - [2.5 Long-running supervisors that report errors without exiting](#25-long-running-supervisors-that-report-errors-without-exiting)
    - [2.6 Live concurrency tuning under changing load](#26-live-concurrency-tuning-under-changing-load)
    - [2.7 Isolating a single task's panic from the rest of the batch](#27-isolating-a-single-tasks-panic-from-the-rest-of-the-batch)
  - [3. Case Catalog: `WorkerPool`](#3-case-catalog-workerpool)
    - [3.1 Persistent background job / queue consumer](#31-persistent-background-job--queue-consumer)
    - [3.2 HTTP request offloading with backpressure](#32-http-request-offloading-with-backpressure)
    - [3.3 Autoscaling workers in response to load](#33-autoscaling-workers-in-response-to-load)
    - [3.4 Bounded in-memory queueing ahead of slow consumers](#34-bounded-in-memory-queueing-ahead-of-slow-consumers)
    - [3.5 Graceful shutdown of a pool](#35-graceful-shutdown-of-a-pool)
    - [3.6 Driving a `WorkerPool` from a `crontask` job](#36-driving-a-workerpool-from-a-crontask-job)
  - [4. Combined / Advanced Patterns](#4-combined--advanced-patterns)
    - [4.1 Fan-out/fan-in pipeline](#41-fan-outfan-in-pipeline)
    - [4.2 Rate-limited webhook delivery pool](#42-rate-limited-webhook-delivery-pool)
    - [4.3 Application-wide graceful shutdown ordering](#43-application-wide-graceful-shutdown-ordering)
  - [5. Anti-Patterns — When _Not_ to Use `workergroup`](#5-anti-patterns--when-not-to-use-workergroup)
  - [6. Observability \& Monitoring](#6-observability--monitoring)
  - [7. Summary Checklist](#7-summary-checklist)

---

## 1. Decision Matrix

| Symptom / Requirement                                                            | Use                                                    | Why                                                                                     |
| -------------------------------------------------------------------------------- | ------------------------------------------------------ | --------------------------------------------------------------------------------------- |
| A handful of independent calls, started once, awaited once                       | `Group` (+ `WithLimit`)                                | `Go`/`Wait` matches the call's lifetime exactly; no need for a standing pool            |
| One failure should abort the whole batch immediately                             | `Group` with `FirstError` (default) + `NewWithContext` | The derived `Context` cancels on first error, stopping siblings                         |
| One failure must not stop the other tasks                                        | `Group` with `WithErrorMode(CollectErrors)`            | Errors accumulate via `errors.Join`; the context is never canceled on error             |
| A task should run only if capacity is free, otherwise skip it                    | `Group.TryGo`                                          | Non-blocking variant of `Go`; returns `false` instead of waiting for a slot             |
| A long-lived loop needs to report failures without ending the goroutine          | `Group.Fail`                                           | Reports an error exactly like a returned error, without returning from the task         |
| Concurrency limit must change while goroutines are already running               | `Group.SetLimit` / `Group.Scale`                       | Never panics, unlike `errgroup.SetLimit`; takes effect immediately                      |
| A goroutine processes jobs that keep arriving over the service's lifetime        | `WorkerPool`                                           | Persistent workers pull from a queue instead of spawning one goroutine per job          |
| Callers must block (with backpressure) when all workers are busy                 | `WorkerPool.Submit`                                    | Blocks until a worker or queue slot frees up, `ctx` is canceled, or the pool is closed  |
| Callers must never block; drop or reject when the pool is saturated              | `WorkerPool.TrySubmit`                                 | Non-blocking enqueue; returns `false` immediately if nothing is free                    |
| Worker count must track load (deep queue → scale up, idle → scale down)          | `WorkerPool.ScaleTo` / `ScaleBy`                       | Adjusts the live worker count without restarting the pool or losing queued jobs         |
| A single call site, no need to reuse workers, task count is known up front       | `Group` (not `WorkerPool`)                             | `WorkerPool` is for a _standing_ pool; a one-shot batch doesn't need worker persistence |
| A single sequential operation with no concurrency need                           | Plain function call                                    | `workergroup` adds synchronization overhead with no benefit for strictly serial work    |
| Exactly one unit of background work, fire-and-forget, no error tracking required | Plain `go func() { ... }()`                            | A `Group` only pays for itself when you need to wait, limit, or collect errors          |

---

## 2. Case Catalog: `Group`

### 2.1 Bounded fan-out of independent I/O calls

**When:** you need to call N independent, slow I/O operations (HTTP requests,
database queries, file reads) concurrently, wait for all of them, and bound how
many run at once so you don't exhaust connection pools or rate limits.

```go
package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/polarixa/replify/pkg/workergroup"
)

type Profile struct {
	ID   string
	Name string
}

// fetchProfile simulates a slow network call to a profile service.
func fetchProfile(ctx context.Context, id string) (Profile, error) {
	return Profile{ID: id, Name: "user-" + id}, nil
}

func fetchAllProfiles(ctx context.Context, ids []string) (map[string]Profile, error) {
	g, ctx := workergroup.NewWithContext(ctx, workergroup.WithLimit(8))

	var mu sync.Mutex
	out := make(map[string]Profile, len(ids))

	for _, id := range ids {
		id := id
		g.Go(func() error {
			p, err := fetchProfile(ctx, id)
			if err != nil {
				return fmt.Errorf("fetch profile %s: %w", id, err)
			}
			mu.Lock()
			out[id] = p
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return out, nil
}

func main() {
	ids := []string{"1", "2", "3", "4", "5"}

	profiles, err := fetchAllProfiles(context.Background(), ids)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(profiles)
}
```

`WithLimit(8)` caps concurrent outbound calls at 8 regardless of `len(ids)`,
protecting the downstream service and the local connection pool.

### 2.2 Fail-fast pipelines with context cancellation

**When:** any single failure makes the rest of the work pointless — e.g. a
multi-step validation where the first invalid record should stop all other
validators from wasting work.

```go
package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/polarixa/replify/pkg/workergroup"
)

type Order struct {
	ID string
}

func validateInventory(ctx context.Context, order Order) error { return nil }

func validatePayment(ctx context.Context, order Order) error {
	return errors.New("payment declined")
}

func validateShippingAddress(ctx context.Context, order Order) error { return nil }

func validateOrder(ctx context.Context, order Order) error {
	g, ctx := workergroup.NewWithContext(ctx) // FirstError is the default mode

	g.Go(func() error { return validateInventory(ctx, order) })
	g.Go(func() error { return validatePayment(ctx, order) })
	g.Go(func() error { return validateShippingAddress(ctx, order) })

	return g.Wait()
}

func main() {
	order := Order{ID: "order-123"}

	if err := validateOrder(context.Background(), order); err != nil {
		fmt.Println("order rejected:", err)
		return
	}
	fmt.Println("order accepted")
}
```

As soon as any validator returns an error, `ctx` is canceled, so the other two
validators can observe `ctx.Done()` and return early instead of continuing
pointless work.

### 2.3 Best-effort batch processing that must not stop early

**When:** you are processing a batch (e.g. sending 500 emails, updating 500
rows) where one failure should be recorded but must not prevent the rest of
the batch from running.

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/polarixa/replify/pkg/workergroup"
)

// sendEmail simulates delivering one digest email; every 7th recipient fails.
func sendEmail(ctx context.Context, recipient string, index int) error {
	if index%7 == 0 {
		return fmt.Errorf("smtp rejected %s", recipient)
	}
	return nil
}

func sendDigestEmails(ctx context.Context, recipients []string) error {
	g := workergroup.New(
		workergroup.WithLimit(20),
		workergroup.WithErrorMode(workergroup.CollectErrors),
	)

	for i, r := range recipients {
		i, r := i, r
		g.Go(func() error {
			if err := sendEmail(ctx, r, i); err != nil {
				return fmt.Errorf("send to %s: %w", r, err)
			}
			return nil
		})
	}

	// err joins every failed send; nil only if all recipients succeeded.
	if err := g.Wait(); err != nil {
		log.Printf("digest send: %d failures: %v", len(g.Errors()), err)
	}
	return nil // a partial failure does not fail the whole operation
}

func main() {
	recipients := make([]string, 50)
	for i := range recipients {
		recipients[i] = fmt.Sprintf("user%d@example.com", i)
	}

	_ = sendDigestEmails(context.Background(), recipients)
}
```

Without `CollectErrors`, the default `FirstError` mode would still run every
task (errors never stop already-started goroutines) but would discard all
errors except the first one — unacceptable when you need to know which
recipients failed.

### 2.4 Opportunistic work that should never block the caller

**When:** a task is a "nice to have" — cache warming, speculative
prefetching, best-effort analytics — and the caller must never wait for a
concurrency slot to free up.

```go
package main

import (
	"context"
	"fmt"

	"github.com/polarixa/replify/pkg/workergroup"
)

func warmCache(ctx context.Context, key string) error {
	fmt.Println("warming cache for", key)
	return nil
}

func maybeWarmCache(g *workergroup.Group, key string) bool {
	started := g.TryGo(func() error {
		return warmCache(context.Background(), key)
	})
	if !started {
		// Pool is saturated — skip silently; this was opportunistic anyway.
		fmt.Println("skipped warming", key)
	}
	return started
}

func main() {
	g := workergroup.New(workergroup.WithLimit(2))

	keys := []string{"a", "b", "c", "d", "e"}
	skipped := 0
	for _, key := range keys {
		if !maybeWarmCache(g, key) {
			skipped++
		}
	}

	g.Wait()
	fmt.Println("total skipped:", skipped)
}
```

`TryGo` never blocks: if the `Group`'s limit is already reached, it returns
`false` immediately instead of queueing the request.

### 2.5 Long-running supervisors that report errors without exiting

**When:** a goroutine runs an indefinite loop (e.g. a `WorkerPool` worker, a
stream consumer) and a single failed unit of work must be recorded without
ending the loop.

```go
package main

import (
	"fmt"

	"github.com/polarixa/replify/pkg/workergroup"
)

type Message struct {
	ID   string
	Fail bool
}

func handle(msg Message) error {
	if msg.Fail {
		return fmt.Errorf("handler rejected message %s", msg.ID)
	}
	return nil
}

func main() {
	messages := make(chan Message, 10)
	go func() {
		for i := 0; i < 10; i++ {
			messages <- Message{ID: fmt.Sprintf("msg-%d", i), Fail: i%4 == 0}
		}
		close(messages)
	}()

	g := workergroup.New(workergroup.WithErrorMode(workergroup.CollectErrors))

	g.Go(func() error {
		for msg := range messages {
			if err := handle(msg); err != nil {
				// Record the failure but keep consuming — returning here
				// would end the goroutine and stop processing messages.
				g.Fail(fmt.Errorf("handle message %s: %w", msg.ID, err))
				continue
			}
		}
		return nil
	})

	g.Wait()
	fmt.Println("failures:", g.Errors())
}
```

This is exactly the mechanism `WorkerPool` itself uses internally
(`spawnWorkerLocked` calls `p.group.Fail` on a failed `Job` instead of
returning, so the worker keeps pulling from the queue).

### 2.6 Live concurrency tuning under changing load

**When:** the right concurrency limit isn't known up front, or needs to react
to backpressure signals (e.g. a downstream 429, CPU pressure, or a manual
ops lever) while the `Group` is already running.

```go
package main

import (
	"fmt"
	"time"

	"github.com/polarixa/replify/pkg/workergroup"
)

// onDownstreamThrottled reacts to a monitored backpressure signal.
func onDownstreamThrottled(g *workergroup.Group) {
	if n := g.Limit(); n > 1 {
		g.SetLimit(n / 2) // halve concurrency, floor at 1
	}
}

func onDownstreamRecovered(g *workergroup.Group) {
	g.Scale(+2) // ease back up incrementally rather than snapping to max
}

func main() {
	g := workergroup.New(workergroup.WithLimit(4))

	for i := 0; i < 20; i++ {
		i := i
		g.Go(func() error {
			time.Sleep(10 * time.Millisecond)
			if i == 5 {
				onDownstreamThrottled(g)
			}
			if i == 12 {
				onDownstreamRecovered(g)
			}
			return nil
		})
	}

	g.Wait()
	fmt.Println("final limit:", g.Limit())
}
```

Unlike `golang.org/x/sync/errgroup`, which panics if `SetLimit` is called
while goroutines from `Go` are still active, `Group.SetLimit`/`Scale` are
safe to call at any time — goroutines blocked waiting for a slot wake up and
observe the new limit immediately.

### 2.7 Isolating a single task's panic from the rest of the batch

**When:** tasks call into less-trusted code (e.g. a user-supplied plugin, a
reflection-heavy serializer) where a single panic must not crash the whole
process or silently kill one goroutine out of many.

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/polarixa/replify/pkg/workergroup"
)

type Plugin struct {
	Name string
}

func (p Plugin) Run(ctx context.Context) error {
	if p.Name == "broken" {
		panic("plugin exploded")
	}
	return nil
}

func main() {
	ctx := context.Background()
	plugins := []Plugin{{Name: "a"}, {Name: "broken"}, {Name: "b"}}

	g := workergroup.New(
		workergroup.WithLimit(10),
		workergroup.WithErrorMode(workergroup.CollectErrors),
		workergroup.WithPanicRecovery(func(recovered any, stack []byte) {
			log.Printf("task panic: %v\n%s", recovered, stack)
		}),
	)

	for _, plugin := range plugins {
		plugin := plugin
		g.Go(func() error { return plugin.Run(ctx) })
	}

	err := g.Wait() // panics surface as errors satisfying errors.Is(err, workergroup.ErrPanic)
	fmt.Println("result:", err)
}
```

---

## 3. Case Catalog: `WorkerPool`

### 3.1 Persistent background job / queue consumer

**When:** work arrives continuously over the lifetime of the service (e.g.
consuming from a message broker, processing an outbox table) rather than as
one fixed batch, so you want a standing set of workers instead of spawning a
goroutine per item.

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/polarixa/replify/pkg/workergroup"
)

type Message struct {
	Key string
}

func handleMessage(ctx context.Context, msg Message) error {
	fmt.Println("processed", msg.Key)
	return nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool := workergroup.NewWorkerPool(ctx, 8, workergroup.WithPoolPanicRecovery(
		func(recovered any, stack []byte) {
			log.Printf("job panic: %v\n%s", recovered, stack)
		},
	))

	// simulates messages continuously arriving from a broker.
	messages := make(chan Message, 100)
	go func() {
		for i := 0; i < 20; i++ {
			messages <- Message{Key: fmt.Sprintf("key-%d", i)}
		}
		close(messages)
	}()

	go func() {
		for msg := range messages {
			msg := msg
			if err := pool.Submit(ctx, func(ctx context.Context) error {
				return handleMessage(ctx, msg)
			}); err != nil {
				log.Printf("submit message %s: %v", msg.Key, err)
			}
		}
		pool.Close()
	}()

	if err := pool.Wait(); err != nil {
		log.Printf("pool finished with errors: %v", err)
	}
}
```

Workers are long-lived: 8 goroutines are created once and reused for every
message, instead of creating and tearing down a goroutine per message.

### 3.2 HTTP request offloading with backpressure

**When:** an HTTP handler needs to hand off slow work to a background pool
without the request thread doing the work itself, while still applying
backpressure so the service degrades predictably under load instead of
spawning unbounded goroutines.

```go
package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/polarixa/replify/pkg/workergroup"
)

func processUpload(ctx context.Context, name string) error {
	log.Println("processing upload", name)
	return nil
}

func recordEvent(ctx context.Context, event string) error {
	log.Println("recording event", event)
	return nil
}

type Handler struct {
	pool *workergroup.WorkerPool
}

func (h *Handler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Query().Get("file")
	if file == "" {
		http.Error(w, "missing file", http.StatusBadRequest)
		return
	}

	// Submit blocks briefly if every worker is busy, naturally slowing down
	// the caller instead of letting an unbounded number of goroutines pile up.
	err := h.pool.Submit(r.Context(), func(ctx context.Context) error {
		return processUpload(ctx, file)
	})
	switch {
	case errors.Is(err, workergroup.ErrPoolClosed):
		http.Error(w, "service shutting down", http.StatusServiceUnavailable)
	case err != nil:
		http.Error(w, "request canceled", http.StatusRequestTimeout)
	default:
		w.WriteHeader(http.StatusAccepted)
	}
}

// HandleEvent is best-effort: respond immediately instead of waiting for capacity.
func (h *Handler) HandleEvent(w http.ResponseWriter, r *http.Request) {
	event := r.URL.Query().Get("event")

	if !h.pool.TrySubmit(func(ctx context.Context) error {
		return recordEvent(ctx, event)
	}) {
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func main() {
	h := &Handler{pool: workergroup.NewWorkerPool(context.Background(), 4)}

	mux := http.NewServeMux()
	mux.HandleFunc("/upload", h.HandleUpload)
	mux.HandleFunc("/event", h.HandleEvent)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
```

Use `TrySubmit` instead of `Submit` when the handler must respond
immediately (e.g. `202 Accepted` vs `503`) rather than wait, such as a
best-effort telemetry endpoint — see `HandleEvent` above.

### 3.3 Autoscaling workers in response to load

**When:** the ideal worker count varies with queue depth or external signals
(time of day, downstream capacity) and must change without rebuilding the
pool or losing already-queued jobs.

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/polarixa/replify/pkg/workergroup"
)

// rebalance is run periodically (e.g. from a crontask job, see §3.6, or a ticker).
func rebalance(pool *workergroup.WorkerPool, queueDepth int) {
	switch {
	case queueDepth > 400:
		pool.ScaleTo(32) // queue is nearly full — scale aggressively
	case queueDepth > 100:
		pool.ScaleBy(+4)
	case queueDepth < 10 && pool.Workers() > 4:
		pool.ScaleBy(-4) // scale back down to the floor once load subsides
	}
}

func main() {
	pool := workergroup.NewWorkerPool(context.Background(), 4, workergroup.WithQueueSize(500))
	defer func() {
		pool.Close()
		pool.Wait()
	}()

	depths := []int{50, 150, 420, 300, 5}
	for _, depth := range depths {
		rebalance(pool, depth)
		fmt.Printf("queueDepth=%d -> workers=%d\n", depth, pool.Workers())
		time.Sleep(10 * time.Millisecond)
	}
}
```

Workers that are stopped by `ScaleTo`/`ScaleBy` finish their current job
before exiting — scaling down never abandons in-progress work.

### 3.4 Bounded in-memory queueing ahead of slow consumers

**When:** producers can briefly outpace workers and you want a small buffer
to absorb bursts, rather than having every `Submit` call block the instant
all workers are busy.

```go
package main

import (
	"context"
	"fmt"

	"github.com/polarixa/replify/pkg/workergroup"
)

func main() {
	ctx := context.Background()

	pool := workergroup.NewWorkerPool(
		ctx,
		4,
		workergroup.WithQueueSize(200), // absorb bursts of up to 200 pending jobs
	)

	for i := 0; i < 50; i++ {
		i := i
		if !pool.TrySubmit(func(ctx context.Context) error {
			fmt.Println("processing job", i)
			return nil
		}) {
			fmt.Println("queue full, job", i, "rejected")
		}
	}

	pool.Close()
	pool.Wait()
}
```

With the default `WithQueueSize(0)`, `Submit` hands a job directly to a
ready worker (an unbuffered rendezvous); raising the queue size lets callers
get up to `n` jobs ahead of the workers before `Submit` starts blocking —
useful when producer bursts are short-lived and memory for queued jobs is
cheap relative to blocking the producer.

### 3.5 Graceful shutdown of a pool

**When:** the service is shutting down (SIGTERM, `context.Context`
cancellation) and in-flight jobs must be allowed to finish while new jobs
are rejected.

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/polarixa/replify/pkg/workergroup"
)

func shutdownPool(ctx context.Context, pool *workergroup.WorkerPool) error {
	pool.Close() // reject new Submit/TrySubmit calls; let queued jobs drain

	done := make(chan error, 1)
	go func() { done <- pool.Wait() }()

	select {
	case err := <-done:
		return err // every worker exited; err is the combined job error
	case <-ctx.Done():
		return ctx.Err() // hard shutdown deadline exceeded
	}
}

func main() {
	pool := workergroup.NewWorkerPool(context.Background(), 4)

	for i := 0; i < 10; i++ {
		i := i
		pool.Submit(context.Background(), func(ctx context.Context) error {
			time.Sleep(20 * time.Millisecond)
			fmt.Println("finished job", i)
			return nil
		})
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := shutdownPool(shutdownCtx, pool); err != nil {
		fmt.Println("shutdown error:", err)
	}
}
```

Always call `Close` before `Wait` — a pool that is never closed, scaled to
zero, or whose `Context` is never canceled will block in `Wait` forever,
since idle workers loop indefinitely waiting for jobs.

### 3.6 Driving a `WorkerPool` from a `crontask` job

**When:** a scheduled job (see `pkg/crontask`) needs to fan out a large,
variable-sized batch of sub-tasks on each tick — e.g. "every 5 minutes,
reconcile every stale account" — without spawning an unbounded number of
goroutines per tick.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/polarixa/replify/pkg/crontask"
	"github.com/polarixa/replify/pkg/workergroup"
)

type Account struct {
	ID string
}

func listStaleAccounts(ctx context.Context) ([]Account, error) {
	return []Account{{ID: "a1"}, {ID: "a2"}, {ID: "a3"}}, nil
}

func reconcileAccount(ctx context.Context, id string) error {
	fmt.Println("reconciling", id)
	return nil
}

// reconcileJob fans out a bounded batch of sub-tasks on every scheduler tick.
func reconcileJob(ctx context.Context) error {
	accounts, err := listStaleAccounts(ctx)
	if err != nil {
		return fmt.Errorf("list stale accounts: %w", err)
	}

	g := workergroup.New(
		workergroup.WithLimit(16),
		workergroup.WithErrorMode(workergroup.CollectErrors),
	)
	for _, acct := range accounts {
		acct := acct
		g.Go(func() error { return reconcileAccount(ctx, acct.ID) })
	}
	return g.Wait()
}

func main() {
	sched, err := crontask.New()
	if err != nil {
		log.Fatalf("crontask.New: %v", err)
	}

	if _, err := sched.Register(
		"*/5 * * * *",
		reconcileJob,
		crontask.WithJobID("account-reconcile"),
		crontask.WithTimeout(4*time.Minute),
	); err != nil {
		log.Fatalf("register job: %v", err)
	}

	if err := sched.Start(); err != nil {
		log.Fatalf("start scheduler: %v", err)
	}
	defer sched.Shutdown(context.Background())

	select {} // block forever; replace with real service lifecycle
}
```

A plain `Group` (not a `WorkerPool`) is the right choice here because the
work is a bounded, one-shot batch per tick, not a continuously arriving
stream — see the [Decision Matrix](#1-decision-matrix).

---

## 4. Combined / Advanced Patterns

### 4.1 Fan-out/fan-in pipeline

**When:** each unit of work produces a result that must be collected, in
addition to an error, without a shared map and mutex.

```go
package main

import (
	"context"
	"fmt"

	"github.com/polarixa/replify/pkg/workergroup"
)

type Item struct {
	ID string
}

type Result struct {
	ItemID string
	Value  int
}

func transform(ctx context.Context, item Item) (Result, error) {
	return Result{ItemID: item.ID, Value: len(item.ID)}, nil
}

func transformAll(ctx context.Context, items []Item) ([]Result, error) {
	g, ctx := workergroup.NewWithContext(ctx, workergroup.WithLimit(8))
	results := make([]Result, len(items))

	for i, item := range items {
		i, item := i, item
		g.Go(func() error {
			r, err := transform(ctx, item)
			if err != nil {
				return fmt.Errorf("transform item %d: %w", i, err)
			}
			results[i] = r // each goroutine owns a distinct index — no mutex needed
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return results, nil
}

func main() {
	items := []Item{{ID: "a"}, {ID: "bb"}, {ID: "ccc"}}

	results, err := transformAll(context.Background(), items)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(results)
}
```

Pre-allocating `results` and writing to a distinct index per goroutine
avoids the mutex needed in [§2.1](#21-bounded-fan-out-of-independent-io-calls),
since indices never collide.

### 4.2 Rate-limited webhook delivery pool

**When:** a persistent pool must also respect an external rate limit (e.g.
the receiving webhook endpoint allows only N requests/sec) in addition to
the fixed worker count.

```go
package main

import (
	"context"
	"fmt"

	"github.com/polarixa/replify/pkg/workergroup"
	"golang.org/x/time/rate"
)

type Payload struct {
	ID string
}

func postWebhook(ctx context.Context, payload Payload) error {
	fmt.Println("delivered", payload.ID)
	return nil
}

func main() {
	ctx := context.Background()
	pool := workergroup.NewWorkerPool(ctx, 10, workergroup.WithQueueSize(1000))
	defer func() {
		pool.Close()
		pool.Wait()
	}()

	limiter := rate.NewLimiter(rate.Limit(50), 50) // 50 req/s across all workers

	deliver := func(ctx context.Context, payload Payload) error {
		return pool.Submit(ctx, func(ctx context.Context) error {
			if err := limiter.Wait(ctx); err != nil {
				return err
			}
			return postWebhook(ctx, payload)
		})
	}

	for i := 0; i < 20; i++ {
		if err := deliver(ctx, Payload{ID: fmt.Sprintf("evt-%d", i)}); err != nil {
			fmt.Println("deliver error:", err)
		}
	}
}
```

The worker count bounds _concurrency_; the `rate.Limiter` inside each job
bounds _throughput_ — the two are complementary, not redundant.

### 4.3 Application-wide graceful shutdown ordering

**When:** a service combines an HTTP server, a `crontask.Scheduler`, and one
or more `WorkerPool`s, and shutdown must drain them in a safe order.

```go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/polarixa/replify/pkg/crontask"
	"github.com/polarixa/replify/pkg/workergroup"
)

func runUntilSignal(srv *http.Server, sched *crontask.Scheduler, pool *workergroup.WorkerPool, timeout time.Duration) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 1. Stop accepting new HTTP requests first, so nothing new is submitted.
	_ = srv.Shutdown(ctx)

	// 2. Stop the scheduler so no new ticks enqueue more pool jobs.
	_ = sched.Shutdown(ctx)

	// 3. Stop accepting pool jobs and let in-flight ones drain.
	pool.Close()
	if err := pool.Wait(); err != nil {
		log.Printf("pool drain: %v", err)
	}
}

func main() {
	ctx := context.Background()

	pool := workergroup.NewWorkerPool(ctx, 8)
	sched, err := crontask.New()
	if err != nil {
		log.Fatalf("crontask.New: %v", err)
	}
	if err := sched.Start(); err != nil {
		log.Fatalf("start scheduler: %v", err)
	}

	srv := &http.Server{Addr: ":8080", Handler: http.NewServeMux()}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server: %v", err)
		}
	}()

	runUntilSignal(srv, sched, pool, 30*time.Second)
}
```

Reversing steps 1–3 risks the scheduler or HTTP layer submitting new jobs to
a pool that is already closed, which `Submit`/`TrySubmit` correctly reject
with `ErrPoolClosed`, but it's cleaner to stop producers before consumers.

---

## 5. Anti-Patterns — When _Not_ to Use `workergroup`

| Situation                                                             | Why `workergroup` is the wrong tool                                                                 | Use instead                                                        |
| --------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------ |
| A single, short, sequential operation                                 | Synchronization overhead with zero concurrency benefit                                              | A direct function call                                             |
| One fire-and-forget goroutine with no result, error, or limit needed  | A `Group` only pays for itself when you need to wait, bound concurrency, or collect errors          | A plain `go func() { ... }()`                                      |
| Streaming transformation where each stage must process items in order | `Group`/`WorkerPool` do not guarantee ordering between concurrent tasks                             | A single goroutine, or an ordered pipeline with per-stage channels |
| Strict one-at-a-time mutual exclusion on a shared resource            | `WithLimit(1)` works, but a `sync.Mutex` around the critical section is simpler and clearer         | `sync.Mutex` / `sync.RWMutex`                                      |
| CPU-bound work that should map 1:1 onto `GOMAXPROCS`                  | `workergroup` does not pin or tune for CPU affinity; an unbounded `WithLimit(-1)` can oversubscribe | A pool sized explicitly to `runtime.GOMAXPROCS(0)`                 |
| A `WorkerPool` created and discarded per request                      | Defeats the purpose of a _persistent_ pool; pays worker start-up cost every request                 | A `Group` for the one-shot batch, or a pool shared across requests |

---

## 6. Observability & Monitoring

`Group` and `WorkerPool` expose enough state for basic health checks without
a dedicated metrics hook (unlike `crontask`'s `MetricsHookInstance`, there is
no built-in metrics collector here — wrap the calls below yourself):

| Signal                    | How to read it                                   | What it tells you                                      |
| ------------------------- | ------------------------------------------------ | ------------------------------------------------------ |
| Current concurrency limit | `(*Group).Limit()`                               | The configured ceiling, useful to log after `SetLimit` |
| Active goroutines         | `(*Group).Running()`                             | How saturated the `Group` currently is                 |
| Accumulated errors        | `(*Group).Err()` / `(*Group).Errors()`           | Non-blocking failure inspection mid-run                |
| Active workers            | `(*WorkerPool).Workers()`                        | Current pool size, to verify a `ScaleTo` took effect   |
| Pool failures             | `(*WorkerPool).Err()` / `(*WorkerPool).Errors()` | Job failures accumulated under `CollectErrors` mode    |

Example: exposing pool health on a metrics endpoint.

```go
package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/polarixa/replify/pkg/workergroup"
)

type Handler struct {
	pool *workergroup.WorkerPool
}

func (h *Handler) PoolStats(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]any{
		"workers":     h.pool.Workers(),
		"error_count": len(h.pool.Errors()),
	})
}

func main() {
	h := &Handler{pool: workergroup.NewWorkerPool(context.Background(), 4)}

	mux := http.NewServeMux()
	mux.HandleFunc("/pool/stats", h.PoolStats)

	http.ListenAndServe(":8080", mux)
}
```

For panics, always wire `WithPanicRecovery` / `WithPoolPanicRecovery` to a
handler that logs and increments a counter — otherwise a recovered panic is
only visible as an opaque error wrapping `workergroup.ErrPanic`.

---

## 7. Summary Checklist

Before reaching for `workergroup`, confirm:

1. **Is there real concurrency?** If the work is inherently sequential, skip
   `workergroup` entirely.
2. **One-shot batch or continuous stream?** One-shot → `Group`. Continuous →
   `WorkerPool`.
3. **Should one failure cancel the rest?** Yes → default `FirstError` +
   `NewWithContext`. No → `WithErrorMode(CollectErrors)`.
4. **Must the caller ever skip rather than wait?** Yes → `TryGo` /
   `TrySubmit`. No → `Go` / `Submit`.
5. **Will the concurrency limit need to change at runtime?** Yes →
   `SetLimit`/`Scale` on `Group`, `ScaleTo`/`ScaleBy` on `WorkerPool`. Both
   are always safe to call live.
6. **Can a single task panic?** If the code is not fully trusted, enable
   `WithPanicRecovery` / `WithPoolPanicRecovery` and log every recovery.
7. **Is there a clean shutdown path?** For `WorkerPool`, always
   `Close()` before `Wait()`, and order it after any producer (HTTP server,
   scheduler) that might still submit jobs.
