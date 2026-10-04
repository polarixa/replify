package replify

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/polarixa/replify/pkg/strutil"
	"github.com/polarixa/replify/pkg/workergroup"
)

// ErrNilTask is the error recorded in a [WorkerResult] when a nil [WorkerFunc]
// or nil [workergroup.Job] is submitted to a [WorkerGroup] or [Pool]. A nil
// task is treated as an ordinary task failure rather than a panic, so it is
// reported through the same [WorkerResult] / wrapper channel as any other
// error instead of crashing the submitting goroutine.
var ErrNilTask = errors.New("replify: task is nil")

// ErrWorkerPanicked is recorded in a [WorkerResult] when its task panicked
// and the panic unwound past the point where the result would normally be
// recorded. It only appears when the underlying [workergroup.Group] or
// [workergroup.WorkerPool] is configured with panic recovery
// ([workergroup.WithPanicRecovery] / [workergroup.WithPoolPanicRecovery]); the
// recovered value and stack trace itself are delivered to that option's
// handler, not here. Without panic recovery enabled, a panicking task
// crashes the process exactly as it would using [pkg/workergroup] directly —
// [WorkerGroup] and [Pool] do not change that default.
var ErrWorkerPanicked = errors.New("replify: task panicked before completing")

// WorkerFunc is a unit of work submitted to a [WorkerGroup] through [WorkerGroup.Go]
// or [WorkerGroup.TryGo]. It receives the group's Context — the one returned
// alongside the [WorkerGroup] by [NewWorkerGroupWithContext], or
// [context.Background] when the group was created with [NewWorkerGroup] — and
// returns an optional result value alongside an error.
//
// A nil error marks the task as successful and value (if any) is preserved in
// the corresponding [WorkerResult.Value]. A non-nil error marks the task as
// failed; it is preserved verbatim in [WorkerResult.Err] and also reported to
// the underlying [workergroup.Group], wrapped with the task's name via
// fmt.Errorf("worker %q failed: %w", name, err) so that [errors.Is] and
// [errors.As] continue to see the original error through the wrapping.
type WorkerFunc func(ctx context.Context) (any, error)

// WorkerTask pairs a descriptive name with the [WorkerFunc] to execute. It is
// the input accepted by [RunWorkerGroup] for the common case of submitting a
// fixed, known-upfront batch of named tasks. Build one with [NewWorkerTask].
type WorkerTask struct {
	name string // identifies the task in the resulting [WorkerResult] and in any wrapped error
	fn   WorkerFunc
}

// NewWorkerTask creates a new [WorkerTask] pairing name with fn.
//
// Parameters:
//   - name: identifies the task in the resulting [WorkerResult] and in any
//     wrapped error. Defaults to "task-<index>" (its submission index) when
//     empty.
//   - fn: the unit of work to execute. A nil fn is reported as a failed task
//     carrying [ErrNilTask] rather than panicking.
//
// Returns:
//   - A pointer to a newly created [WorkerTask] instance.
func NewWorkerTask(name string, fn WorkerFunc) *WorkerTask {
	return &WorkerTask{name: name, fn: fn}
}

// Available checks whether the [WorkerTask] instance is non-nil.
//
// Returns:
//   - A boolean value indicating whether the [WorkerTask] instance is non-nil.
func (t *WorkerTask) Available() bool {
	return t != nil
}

// Name retrieves the descriptive name of the [WorkerTask] instance.
//
// Returns:
//   - A string representing the task's name, or an empty string if the
//     [WorkerTask] instance is nil.
func (t *WorkerTask) Name() string {
	if !t.Available() {
		return ""
	}
	return t.name
}

// Fn retrieves the unit of work of the [WorkerTask] instance.
//
// Returns:
//   - The [WorkerFunc] to execute, or nil if the [WorkerTask] instance is nil.
func (t *WorkerTask) Fn() WorkerFunc {
	if !t.Available() {
		return nil
	}
	return t.fn
}

// WithName sets the descriptive name of the [WorkerTask] instance.
//
// Parameters:
//   - name: The string to set as the task's name.
//
// Returns:
//   - A pointer to the updated [WorkerTask] instance.
func (t *WorkerTask) WithName(name string) *WorkerTask {
	t.name = name
	return t
}

// WithFn sets the unit of work of the [WorkerTask] instance.
//
// Parameters:
//   - fn: The [WorkerFunc] to set as the task's unit of work.
//
// Returns:
//   - A pointer to the updated [WorkerTask] instance.
func (t *WorkerTask) WithFn(fn WorkerFunc) *WorkerTask {
	t.fn = fn
	return t
}

// PoolJob pairs a descriptive name with the [workergroup.Job] to submit to a
// [Pool]. It is the input accepted by [RunWorkerPool] for the common case of
// submitting a fixed, known-upfront batch of named jobs. Build one with
// [NewPoolJob].
type PoolJob struct {
	name string // identifies the job in the resulting [WorkerResult] and in any wrapped error
	job  workergroup.Job
}

// NewPoolJob creates a new [PoolJob] pairing name with job.
//
// Parameters:
//   - name: identifies the job in the resulting [WorkerResult] and in any
//     wrapped error. Defaults to "job-<index>" (its submission index) when
//     empty.
//   - job: the unit of work to execute. A nil job is reported as a failed job
//     carrying [ErrNilTask] rather than panicking.
//
// Returns:
//   - A pointer to a newly created [PoolJob] instance.
func NewPoolJob(name string, job workergroup.Job) *PoolJob {
	return &PoolJob{name: name, job: job}
}

// Available checks whether the [PoolJob] instance is non-nil.
//
// Returns:
//   - A boolean value indicating whether the [PoolJob] instance is non-nil.
func (j *PoolJob) Available() bool {
	return j != nil
}

// Name retrieves the descriptive name of the [PoolJob] instance.
//
// Returns:
//   - A string representing the job's name, or an empty string if the
//     [PoolJob] instance is nil.
func (j *PoolJob) Name() string {
	if !j.Available() {
		return ""
	}
	return j.name
}

// Job retrieves the unit of work of the [PoolJob] instance.
//
// Returns:
//   - The [workergroup.Job] to execute, or nil if the [PoolJob] instance is nil.
func (j *PoolJob) Job() workergroup.Job {
	if !j.Available() {
		return nil
	}
	return j.job
}

// WithName sets the descriptive name of the [PoolJob] instance.
//
// Parameters:
//   - name: The string to set as the job's name.
//
// Returns:
//   - A pointer to the updated [PoolJob] instance.
func (j *PoolJob) WithName(name string) *PoolJob {
	j.name = name
	return j
}

// WithJob sets the unit of work of the [PoolJob] instance.
//
// Parameters:
//   - job: The [workergroup.Job] to set as the job's unit of work.
//
// Returns:
//   - A pointer to the updated [PoolJob] instance.
func (j *PoolJob) WithJob(job workergroup.Job) *PoolJob {
	j.job = job
	return j
}

// WorkerResult captures the outcome of a single task executed through a
// [WorkerGroup] or a [Pool]. Results are returned sorted by Index, where
// Index reflects the order in which tasks acquired a goroutine and began
// running — not necessarily the order Go, TryGo, or Submit were called in,
// since [pkg/workergroup] makes no scheduling-order guarantee for concurrent
// tasks. When a concurrency limit of 1 is used (fully sequential execution),
// Index does match call order.
type WorkerResult struct {
	name     string        // identifies the task or job, defaulted to "task-<index>"/"job-<index>" when empty
	index    int           // zero-based position of the result among all tasks or jobs
	value    any           // the value returned by the task, or nil if not applicable
	err      error         // the error returned by the task, or nil if the task succeeded or if not applicable
	duration time.Duration // the time taken to execute the task, or zero if not applicable
}

// newWorkerResult creates a new [WorkerResult] reserved for the task or job
// identified by name at the given index; value, err, and duration are filled
// in once the task completes.
func newWorkerResult(name string, index int) *WorkerResult {
	return &WorkerResult{name: name, index: index}
}

// Available checks whether the [WorkerResult] instance is non-nil.
//
// Returns:
//   - A boolean value indicating whether the [WorkerResult] instance is non-nil.
func (r *WorkerResult) Available() bool {
	return r != nil
}

// Name retrieves the name of the task or job that produced this [WorkerResult].
//
// Returns:
//   - A string representing the task or job's name, or an empty string if the
//     [WorkerResult] instance is nil.
func (r *WorkerResult) Name() string {
	if !r.Available() {
		return ""
	}
	return r.name
}

// Index retrieves the zero-based position of this [WorkerResult] among all
// results collected by the same [WorkerGroup] or [Pool].
//
// Returns:
//   - An integer representing the result's index, or 0 if the [WorkerResult]
//     instance is nil.
func (r *WorkerResult) Index() int {
	if !r.Available() {
		return 0
	}
	return r.index
}

// Value retrieves the task's return value, when produced by a [WorkerFunc].
//
// Returns:
//   - The value returned by the task, always nil for [Pool] jobs (since
//     [workergroup.Job] only reports an error) or if the [WorkerResult]
//     instance is nil.
func (r *WorkerResult) Value() any {
	if !r.Available() {
		return nil
	}
	return r.value
}

// Err retrieves the error returned by the task, or nil on success. It is the
// original, unwrapped error — the task-name wrapping applied for the
// underlying Group/Pool's own error aggregation is not applied here, so
// [errors.Is] and [errors.As] against sentinel errors work directly.
//
// Returns:
//   - The error returned by the task, or nil if it succeeded or if the
//     [WorkerResult] instance is nil.
func (r *WorkerResult) Err() error {
	if !r.Available() {
		return nil
	}
	return r.err
}

// Duration retrieves the wall-clock time the task spent executing, measured
// from immediately before the task function is invoked to immediately after
// it returns (or, for a recovered panic, to the point the panic unwound past
// the measurement point).
//
// Returns:
//   - A [time.Duration] representing the task's execution time, or 0 if the
//     [WorkerResult] instance is nil.
func (r *WorkerResult) Duration() time.Duration {
	if !r.Available() {
		return 0
	}
	return r.duration
}

// WithName sets the name of the [WorkerResult] instance.
//
// Returns:
//   - A pointer to the updated [WorkerResult] instance.
func (r *WorkerResult) WithName(name string) *WorkerResult {
	r.name = name
	return r
}

// WithIndex sets the zero-based position of the [WorkerResult] instance.
//
// Returns:
//   - A pointer to the updated [WorkerResult] instance.
func (r *WorkerResult) WithIndex(index int) *WorkerResult {
	r.index = index
	return r
}

// WithValue sets the return value of the [WorkerResult] instance.
//
// Returns:
//   - A pointer to the updated [WorkerResult] instance.
func (r *WorkerResult) WithValue(value any) *WorkerResult {
	r.value = value
	return r
}

// WithErr sets the error of the [WorkerResult] instance.
//
// Returns:
//   - A pointer to the updated [WorkerResult] instance.
func (r *WorkerResult) WithErr(err error) *WorkerResult {
	r.err = err
	return r
}

// WithDuration sets the execution duration of the [WorkerResult] instance.
//
// Returns:
//   - A pointer to the updated [WorkerResult] instance.
func (r *WorkerResult) WithDuration(d time.Duration) *WorkerResult {
	r.duration = d
	return r
}

// Success reports whether the task completed without error.
//
// Returns:
//   - A boolean value indicating whether the task succeeded: `true` if Err is
//     nil, `false` otherwise (including when the [WorkerResult] instance is nil).
func (r *WorkerResult) Success() bool {
	return r.Available() && r.err == nil
}

// WorkerGroup wraps a [workergroup.Group], recording the name, return value
// (if any), error, and duration of every task submitted through [WorkerGroup.Go]
// or [WorkerGroup.TryGo]. [WorkerGroup.Wait] turns that bookkeeping into a
// [wrapper] that distinguishes full success, partial failure, group-level
// failure, and context cancellation/timeout — the same distinctions a plain
// [workergroup.Group] leaves to the caller to reconstruct from a single
// combined error.
//
// All other [workergroup.Group] behavior — concurrency limiting, error
// modes, panic recovery, live scaling — is unchanged and reachable through
// the embedded *[workergroup.Group], whose SetLimit, Scale, Limit, Running,
// Fail, Err, and Errors methods are promoted directly onto WorkerGroup. Only
// Wait is overridden, to return a *wrapper instead of a plain error; call
// g.Group.Wait() directly if the plain error is what's needed.
//
// Like [workergroup.Group], a WorkerGroup should not be reused once Wait has
// returned, and is safe for concurrent use by multiple goroutines up to that
// point.
type WorkerGroup struct {
	*workergroup.Group

	// ctx is passed to every WorkerFunc; it is the Context returned by
	// NewWorkerGroupWithContext, or context.Background() for NewWorkerGroup.
	ctx context.Context

	// parent is the caller-supplied Context, kept separately from ctx (which
	// workergroup always cancels internally once Wait returns, even on
	// success) so that Wait can tell a genuine external cancellation/timeout
	// apart from that internal cleanup cancel. Nil for NewWorkerGroup.
	parent context.Context

	started time.Time

	mu      sync.Mutex
	results []*WorkerResult
}

// NewWorkerGroup creates a [WorkerGroup] with no associated Context; tasks
// observe context.Background() and cannot be canceled from outside the
// group. opts configure the underlying [workergroup.Group] exactly as they
// would for [workergroup.New] (concurrency limit, error mode, panic
// recovery).
func NewWorkerGroup(opts ...workergroup.Option) *WorkerGroup {
	return &WorkerGroup{
		Group:   workergroup.New(opts...),
		ctx:     context.Background(),
		started: time.Now(),
	}
}

// NewWorkerGroupWithContext creates a [WorkerGroup] configured by opts and a
// Context derived from ctx, exactly as [workergroup.NewWithContext] does. The
// returned Context is what each [WorkerFunc] receives; it is additionally
// made available via the embedded [workergroup.Group].
//
// If the group's error mode is [workergroup.FirstError] (the default), the
// returned Context is canceled the first time a task fails. In either error
// mode, [WorkerGroup.Wait] inspects ctx itself (not the returned derived
// Context) to detect genuine external cancellation or deadline expiry, since
// the derived Context is always canceled internally once Wait returns.
func NewWorkerGroupWithContext(ctx context.Context, opts ...workergroup.Option) (*WorkerGroup, context.Context) {
	g, derived := workergroup.NewWithContext(ctx, opts...)
	wg := &WorkerGroup{
		Group:   g,
		ctx:     derived,
		parent:  ctx,
		started: time.Now(),
	}
	return wg, derived
}

// Go runs fn in a new goroutine, exactly as [workergroup.Group.Go] does, and
// records its name, return value, error, and duration for the [WorkerResult]
// that [WorkerGroup.Wait] will return. A nil fn is recorded as a failed task
// carrying [ErrNilTask] instead of panicking.
//
// As with [workergroup.Group.Go], the first call to Go must happen before a
// call to Wait, and Go blocks until the task can start without exceeding the
// group's configured concurrency limit.
func (g *WorkerGroup) Go(name string, fn WorkerFunc) {
	g.Group.Go(func() error {
		return g.run(name, fn)
	})
}

// TryGo runs fn in a new goroutine only if doing so would not exceed the
// group's configured concurrency limit, exactly as [workergroup.Group.TryGo]
// does. It reports whether the task was started. When it was, fn's outcome
// is recorded for [WorkerGroup.Wait] exactly as described for [WorkerGroup.Go].
func (g *WorkerGroup) TryGo(name string, fn WorkerFunc) bool {
	return g.Group.TryGo(func() error {
		return g.run(name, fn)
	})
}

// run executes fn, recording its outcome into a reserved [WorkerResult] slot
// before returning an error suitable for the underlying Group's own error
// aggregation (task errors wrapped with the task name).
func (g *WorkerGroup) run(name string, fn WorkerFunc) (groupErr error) {
	g.mu.Lock()
	idx := len(g.results)
	name = strutil.DefaultIfEmpty(name, fmt.Sprintf("task-%d", idx))
	res := newWorkerResult(name, idx)
	g.results = append(g.results, res)
	g.mu.Unlock()

	if fn == nil {
		res.WithErr(ErrNilTask)
		return fmt.Errorf("worker %q failed: %w", name, ErrNilTask)
	}

	start := time.Now()
	completed := false
	defer func() {
		if !completed {
			// A panic unwound past this task (panic recovery is configured
			// on the underlying Group; otherwise the process has already
			// crashed and this defer never runs).
			res.WithErr(ErrWorkerPanicked)
		}
		res.WithDuration(time.Since(start))
	}()

	value, err := fn(g.ctx)
	completed = true
	res.WithValue(value).WithErr(err)

	if err != nil {
		return fmt.Errorf("worker %q failed: %w", name, err)
	}
	return nil
}

// Wait blocks until every task started by [WorkerGroup.Go] or [WorkerGroup.TryGo]
// has returned, exactly as [workergroup.Group.Wait] does, then builds a
// [wrapper] summarizing the outcome:
//
//   - [ClientClosedRequest] (499) or [GatewayTimeout] (504) when the Context
//     passed to [NewWorkerGroupWithContext] was itself canceled or exceeded
//     its deadline (detected independently of task-level errors).
//   - [NoContent] (204) when no tasks were ever submitted.
//   - [OK] (200) when every task succeeded.
//   - [InternalServerError] (500) when every task failed. The combined error
//     from the underlying Group (see [workergroup.Group.Wait]) is attached via
//     WithErrorAck, preserving [errors.Is] / [errors.As] compatibility.
//   - [MultiStatus] (207) when some tasks succeeded and others failed. No
//     top-level error is attached in this case — a 207 response is, by HTTP
//     convention, not itself an error — but every per-task outcome, including
//     each failure's original error, remains available via [wrapper.Body] as
//     a []WorkerResult.
//
// The full, ordered []WorkerResult is always available via [wrapper.Body] in
// every case above, so no outcome information is ever discarded merely
// because it didn't fit the summary status.
func (g *WorkerGroup) Wait() *wrapper {
	err := g.Group.Wait()
	results := g.snapshot()
	return buildGroupWrapper("worker group", g.parent, results, err, time.Since(g.started))
}

// snapshot returns a defensive copy of the results collected so far,
// ordered by Index.
func (g *WorkerGroup) snapshot() []*WorkerResult {
	g.mu.Lock()
	defer g.mu.Unlock()
	out := make([]*WorkerResult, len(g.results))
	copy(out, g.results)
	return out
}

// RunWorkerGroup is the one-call convenience form of [WorkerGroup]: it
// creates a group bound to ctx, submits every task in tasks via
// [WorkerGroup.Go], waits for all of them to complete, and returns the
// resulting [wrapper]. See [WorkerGroup.Wait] for exactly what the returned
// wrapper represents.
//
// concurrency bounds how many tasks run at once; a non-positive value means
// unlimited (matching [workergroup.WithLimit]'s convention). Passing 1 runs
// every task sequentially, in submission order, which also makes the
// resulting []WorkerResult's Index match call order exactly.
//
// Task errors do not cancel ctx or stop remaining tasks — RunWorkerGroup
// always uses [workergroup.CollectErrors] so that a single failing task
// cannot hide the outcome of the others, which is what makes the
// [MultiStatus] partial-failure case meaningful.
func RunWorkerGroup(ctx context.Context, concurrency int, tasks []*WorkerTask) *wrapper {
	opts := []workergroup.Option{workergroup.WithErrorMode(workergroup.CollectErrors)}
	if concurrency > 0 {
		opts = append(opts, workergroup.WithLimit(concurrency))
	}
	g, _ := NewWorkerGroupWithContext(ctx, opts...)
	for _, t := range tasks {
		g.Go(t.Name(), t.Fn())
	}
	return g.Wait()
}

// Pool wraps a [workergroup.WorkerPool], recording the name, error, and
// duration of every [workergroup.Job] submitted through [Pool.Submit] or
// [Pool.TrySubmit] so that [Pool.Wait] can report an accurate [wrapper] once
// every worker has exited.
//
// All other [workergroup.WorkerPool] behavior — live worker scaling via
// ScaleTo/ScaleBy, queueing, closing — is unchanged and reachable through the
// embedded *[workergroup.WorkerPool], whose ScaleTo, ScaleBy, Workers, Close,
// Err, and Errors methods are promoted directly onto Pool. Only Submit,
// TrySubmit, and Wait are overridden: Submit and TrySubmit to accept a job
// Name, and Wait to return a *wrapper instead of a plain error.
type Pool struct {
	*workergroup.WorkerPool

	// ctx is the caller-supplied Context given to NewPool, kept separately
	// so Wait can detect its cancellation/deadline independently of
	// individual job errors.
	ctx context.Context

	started time.Time

	mu      sync.Mutex
	results []*WorkerResult
}

// NewPool creates a [Pool] whose jobs observe ctx and starts it with
// initialWorkers worker goroutines (which may be 0), exactly as
// [workergroup.NewWorkerPool] does. opts configure the underlying pool
// exactly as they would for [workergroup.NewWorkerPool] (queue size, error
// mode, panic recovery).
func NewPool(ctx context.Context, initialWorkers int, opts ...workergroup.PoolOption) *Pool {
	return &Pool{
		WorkerPool: workergroup.NewWorkerPool(ctx, initialWorkers, opts...),
		ctx:        ctx,
		started:    time.Now(),
	}
}

// Submit adds a named job to the pool queue, with the blocking and
// cancellation semantics of [workergroup.WorkerPool.Submit]: it blocks until
// a worker accepts the job, ctx is canceled, the pool's own Context is
// canceled, or the pool is closed, and returns the corresponding error in
// each of those cases. A nil job is still accepted and is reported as a
// failed job carrying [ErrNilTask] once a worker reaches it.
//
// On success, the job's outcome (including a nil job's) is recorded for
// [Pool.Wait] exactly as described for [Pool.Submit]'s sibling, [Pool.TrySubmit].
func (p *Pool) Submit(ctx context.Context, name string, job workergroup.Job) error {
	return p.WorkerPool.Submit(ctx, p.wrap(name, job))
}

// TrySubmit adds a named job to the pool queue only if doing so would not
// block, exactly as [workergroup.WorkerPool.TrySubmit] does. It reports
// whether the job was accepted. When it was, the job's outcome is recorded
// for [Pool.Wait]: name identifies it in the resulting [WorkerResult] and in
// any wrapped error, defaulting to "job-<index>" when empty. A nil job is
// still accepted and is reported as a failed job carrying [ErrNilTask] once a
// worker reaches it.
func (p *Pool) TrySubmit(name string, job workergroup.Job) bool {
	return p.WorkerPool.TrySubmit(p.wrap(name, job))
}

// wrap returns a [workergroup.Job] that executes job (recording name, error,
// and duration into a reserved [WorkerResult] slot) and reports job errors to
// the underlying pool wrapped with the job's name, mirroring [WorkerGroup.run].
func (p *Pool) wrap(name string, job workergroup.Job) workergroup.Job {
	p.mu.Lock()
	idx := len(p.results)
	name = strutil.DefaultIfEmpty(name, fmt.Sprintf("job-%d", idx))
	res := newWorkerResult(name, idx)
	p.results = append(p.results, res)
	p.mu.Unlock()

	return func(ctx context.Context) error {
		if job == nil {
			res.WithErr(ErrNilTask)
			return fmt.Errorf("job %q failed: %w", name, ErrNilTask)
		}

		start := time.Now()
		completed := false
		defer func() {
			if !completed {
				res.WithErr(ErrWorkerPanicked)
			}
			res.WithDuration(time.Since(start))
		}()

		err := job(ctx)
		completed = true
		res.WithErr(err)

		if err != nil {
			return fmt.Errorf("job %q failed: %w", name, err)
		}
		return nil
	}
}

// Wait blocks until every worker goroutine has exited, exactly as
// [workergroup.WorkerPool.Wait] does (which, per that method, means Close
// must have been called, or the pool's Context canceled, or every worker
// individually stopped — otherwise Wait blocks forever), then builds a
// [wrapper] summarizing every submitted job's outcome using the same rules
// as [WorkerGroup.Wait]: [ClientClosedRequest]/[GatewayTimeout] for the
// pool's own Context being canceled/expired, [NoContent] when no jobs were
// ever submitted, [OK] when every job succeeded, [InternalServerError] when
// every job failed (with the combined error attached via WithErrorAck), and
// [MultiStatus] for a mix of both (with no top-level error attached, but the
// full []WorkerResult — including every failure's original error — always
// available via [wrapper.Body]).
func (p *Pool) Wait() *wrapper {
	err := p.WorkerPool.Wait()
	results := p.snapshot()
	return buildGroupWrapper("worker pool", p.ctx, results, err, time.Since(p.started))
}

// snapshot returns a defensive copy of the results collected so far,
// ordered by Index.
func (p *Pool) snapshot() []*WorkerResult {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]*WorkerResult, len(p.results))
	copy(out, p.results)
	return out
}

// RunWorkerPool is the one-call convenience form of [Pool]: it creates a
// pool bound to ctx with workers worker goroutines, submits every job in
// jobs, closes the pool once submission is complete, waits for all workers
// to drain the queue and exit, and returns the resulting [wrapper]. See
// [Pool.Wait] for exactly what the returned wrapper represents.
//
// If submitting a job fails (ctx canceled, or the pool's own Context
// canceled before every job could be accepted), that job is still recorded
// as a failed [WorkerResult] carrying the submission error, rather than
// being silently dropped from the result set.
func RunWorkerPool(ctx context.Context, workers int, jobs []*PoolJob) *wrapper {
	p := NewPool(ctx, workers)
	for _, j := range jobs {
		if err := p.Submit(ctx, j.Name(), j.Job()); err != nil {
			p.mu.Lock()
			p.results = append(p.results, newWorkerResult(j.Name(), len(p.results)).WithErr(err))
			p.mu.Unlock()
		}
	}
	p.Close()
	return p.Wait()
}

// buildGroupWrapper turns a collected []WorkerResult plus the underlying
// Group/Pool's combined error into the [wrapper] returned by
// [WorkerGroup.Wait] and [Pool.Wait]. kind ("worker group" or "worker pool")
// only affects wording in Message; parent is the caller-supplied Context
// (nil when none was given) used to detect genuine external
// cancellation/deadline expiry, independent of task-level errors.
func buildGroupWrapper(kind string, parent context.Context, results []*WorkerResult, groupErr error, elapsed time.Duration) *wrapper {
	w := New().
		WithTotal(len(results)).
		WithBody(results).
		WithDebuggingKV("elapsed", elapsed)

	if parent != nil {
		if cErr := parent.Err(); cErr != nil {
			if errors.Is(cErr, context.DeadlineExceeded) {
				return w.
					WithHeader(GatewayTimeout).
					WithErrorAck(cErr).
					WithReasonCode(ReasonCodeJobTimeout).
					WithMessagef("%s: context deadline exceeded", kind)
			}
			return w.
				WithHeader(ClientClosedRequest).
				WithErrorAck(cErr).
				WithReasonCode(ReasonCodeJobCancelled).
				WithMessagef("%s: context canceled", kind)
		}
	}

	if len(results) == 0 {
		return w.
			WithHeader(NoContent).
			WithMessagef("%s: no tasks were submitted", kind)
	}

	succeeded, failed := 0, 0
	var firstFailed *WorkerResult
	for _, r := range results {
		if r.Success() {
			succeeded++
			continue
		}
		failed++
		if firstFailed == nil {
			firstFailed = r
		}
	}
	w.WithDebuggingKV("succeeded", succeeded).WithDebuggingKV("failed", failed)

	switch {
	case failed == 0:
		return w.
			WithHeader(OK).
			WithMessagef("%s: %d task(s) completed successfully", kind, succeeded)
	case succeeded == 0:
		return w.
			WithHeader(InternalServerError).
			WithErrorAck(groupErr).
			WithReasonCode(ReasonCodeJobFailed).
			WithMessagef("%s: all %d task(s) failed", kind, failed)
	default:
		w.WithDebuggingKV("first_failed", firstFailed.Name())
		return w.
			WithHeader(MultiStatus).
			WithMessagef("%s: %d of %d task(s) failed", kind, failed, len(results))
	}
}
