package workergroup

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
)

// ErrorMode selects how a Group combines errors returned by its tasks.
// See the constants FirstError and CollectErrors for the two supported
// strategies.
type ErrorMode int

const (
	// FirstError keeps only the first non-nil error reported by any task
	// and discards the rest, exactly like golang.org/x/sync/errgroup. If
	// the Group was created with NewWithContext, its derived Context is
	// canceled with that first error. This is the default mode used by
	// New and NewWithContext when no WithErrorMode option is given.
	FirstError ErrorMode = iota

	// CollectErrors accumulates every non-nil error reported by every
	// task. Wait and Err return the accumulated errors joined together
	// with errors.Join, and Errors returns them as a slice in the order
	// they were reported. Unlike FirstError, an error reported in this
	// mode does NOT cancel the Group's derived Context, so other tasks
	// are left to run to completion. This mode suits long-running
	// supervisors, such as WorkerPool, where a single failed unit of
	// work should not stop the others.
	CollectErrors
)

// String returns a human-readable name for the mode, implementing
// fmt.Stringer.
func (m ErrorMode) String() string {
	switch m {
	case FirstError:
		return "FirstError"
	case CollectErrors:
		return "CollectErrors"
	default:
		return fmt.Sprintf("ErrorMode(%d)", int(m))
	}
}

// Option configures a Group at construction time. Options are applied in
// the order they are given to New or NewWithContext.
type Option func(*Group)

// WithLimit returns an Option that bounds the number of goroutines the
// Group will run concurrently to n. A negative n means unlimited, which is
// also the default when WithLimit is not supplied. The limit can still be
// changed later, at any time, with (*Group).SetLimit.
func WithLimit(n int) Option {
	return func(g *Group) { g.limiter().setLimit(n) }
}

// WithErrorMode returns an Option that selects how the Group combines
// errors from its tasks. See ErrorMode for the available strategies. The
// default, if this option is not supplied, is FirstError.
func WithErrorMode(mode ErrorMode) Option {
	return func(g *Group) { g.mode = mode }
}

// WithPanicRecovery returns an Option that makes the Group recover panics
// raised by a task's function instead of letting them crash the process.
// When a task panics, handler is called with the recovered value and a
// formatted stack trace captured at the point of the panic, and the
// panic is converted into an error (satisfying errors.Is against
// ErrPanic) that is reported to the Group exactly as if the task had
// returned it.
//
// By default (when this option is not used) panics are NOT recovered and
// propagate normally, matching golang.org/x/sync/errgroup's documented
// position that silently turning panics into errors hides crashes from
// monitoring tools and can delay them arbitrarily. Enable this option only
// when you have a specific need, such as a long-lived WorkerPool worker
// where a single misbehaving job should not take down the whole pool.
func WithPanicRecovery(handler func(recovered any, stack []byte)) Option {
	return func(g *Group) { g.onPanic = handler }
}

// ErrPanic is wrapped into the error reported to a Group when
// WithPanicRecovery is enabled and a task panics. Use errors.Is(err,
// ErrPanic) to detect this case.
var ErrPanic = errors.New("workergroup: recovered panic in task")

// Group is a collection of goroutines working on subtasks that are part
// of the same overall task. A Group should not be reused once Wait has
// returned.
//
// Unlike golang.org/x/sync/errgroup.Group, a Group here:
//
//   - may have its concurrency limit changed at any time via SetLimit or
//     Scale, including while goroutines are active, without panicking;
//   - supports an explicit ErrorMode so callers can choose between
//     fail-fast (FirstError) and accumulate-everything (CollectErrors)
//     semantics;
//   - exposes Fail, letting a long-running task report an error without
//     having to return from its function, which WorkerPool relies on to
//     let a worker keep processing jobs after one of them fails.
//
// The zero Group is valid, has no limit on the number of active
// goroutines, uses FirstError semantics, and does not cancel on error —
// the same defaults as a zero golang.org/x/sync/errgroup.Group. Prefer New
// or NewWithContext to configure a Group explicitly in the OOP style this
// package is designed around.
type Group struct {
	cancel context.CancelCauseFunc
	wg     sync.WaitGroup

	lim     *limiter
	limOnce sync.Once

	mode ErrorMode

	mu      sync.Mutex
	errOnce sync.Once
	err     error
	errs    []error

	onPanic func(recovered any, stack []byte)
}

// New returns a new Group configured by opts. The returned Group is not
// associated with a Context; use NewWithContext if tasks should observe
// cancellation.
func New(opts ...Option) *Group {
	g := &Group{}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

// NewWithContext returns a new Group configured by opts, and a Context
// derived from ctx.
//
// If the Group's ErrorMode is FirstError (the default), the derived
// Context is canceled the first time a task passed to Go or TryGo returns
// a non-nil error, the first time Fail is called with a non-nil error, or
// the first time Wait returns, whichever occurs first. If the ErrorMode is
// CollectErrors, task errors do not cancel the derived Context; it is
// still canceled when Wait returns.
func NewWithContext(ctx context.Context, opts ...Option) (*Group, context.Context) {
	ctx, cancel := context.WithCancelCause(ctx)
	g := &Group{cancel: cancel}
	for _, opt := range opts {
		opt(g)
	}
	return g, ctx
}

// limiter lazily initializes and returns the Group's semaphore, so that a
// Group constructed as a bare Group{} (rather than via New) behaves
// correctly as an unlimited group, matching the "zero Group is valid"
// guarantee.
func (g *Group) limiter() *limiter {
	g.limOnce.Do(func() {
		if g.lim == nil {
			g.lim = newLimiter(-1)
		}
	})
	return g.lim
}

// done releases this goroutine's concurrency permit and marks it finished
// in the Group's WaitGroup.
func (g *Group) done() {
	g.limiter().release()
	g.wg.Done()
}

// runTask invokes f, optionally recovering a panic if WithPanicRecovery
// was configured, and returns f's error (or an error describing the
// recovered panic).
func (g *Group) runTask(f func() error) (err error) {
	if g.onPanic != nil {
		defer func() {
			if r := recover(); r != nil {
				buf := make([]byte, 64<<10)
				n := runtime.Stack(buf, false)
				stack := buf[:n]
				g.onPanic(r, stack)
				err = fmt.Errorf("%w: %v", ErrPanic, r)
			}
		}()
	}
	return f()
}

// Go calls f in a new goroutine.
//
// The first call to Go must happen before a call to Wait. It blocks until
// the new goroutine can be added without the number of goroutines in the
// Group exceeding the configured limit; if the limit is raised or lowered
// by SetLimit or Scale while Go is blocked, the change takes effect
// immediately.
//
// If f returns a non-nil error, it is reported to the Group exactly as if
// Fail had been called with that error from inside f.
func (g *Group) Go(f func() error) {
	g.limiter().acquire()
	g.wg.Add(1)
	go func() {
		defer g.done()
		if err := g.runTask(f); err != nil {
			g.Fail(err)
		}
	}()
}

// TryGo calls f in a new goroutine only if the number of active goroutines
// in the Group is currently below the configured limit.
//
// The return value reports whether the goroutine was started. If it was,
// errors from f are reported exactly as described for Go.
func (g *Group) TryGo(f func() error) bool {
	if !g.limiter().tryAcquire() {
		return false
	}
	g.wg.Add(1)
	go func() {
		defer g.done()
		if err := g.runTask(f); err != nil {
			g.Fail(err)
		}
	}()
	return true
}

// Fail reports err to the Group as if it had been returned by a task
// passed to Go or TryGo. It is a no-op if err is nil.
//
// Fail exists so that a long-running task — in particular, a WorkerPool
// worker loop that processes many jobs over its lifetime — can report a
// failure without having to return from its function and thereby stop
// running. Combined with CollectErrors mode, this lets a single failed
// unit of work be recorded without ending the goroutine that reported it.
//
// In FirstError mode, only the first error ever passed to Fail (whether
// from a task's return value or a direct call) is kept, and it cancels the
// Group's derived Context, if any. In CollectErrors mode, every error is
// kept and the Context is not canceled.
func (g *Group) Fail(err error) {
	if err == nil {
		return
	}
	switch g.mode {
	case CollectErrors:
		g.mu.Lock()
		g.errs = append(g.errs, err)
		g.mu.Unlock()
		g.errOnce.Do(func() {
			g.mu.Lock()
			g.err = err
			g.mu.Unlock()
		})
	default: // FirstError
		g.errOnce.Do(func() {
			g.mu.Lock()
			g.err = err
			g.mu.Unlock()
			if g.cancel != nil {
				g.cancel(err)
			}
		})
	}
}

// combinedErr computes the Group's current error under a read of its
// internal state, applying the ErrorMode's combination rule.
func (g *Group) combinedErr() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.mode == CollectErrors {
		if len(g.errs) == 0 {
			return nil
		}
		return errors.Join(g.errs...)
	}
	return g.err
}

// Wait blocks until all function calls from the Go and TryGo methods have
// returned, then returns the combined error according to the Group's
// ErrorMode (see FirstError and CollectErrors). If the Group has a derived
// Context (created via NewWithContext), that Context is canceled before
// Wait returns.
func (g *Group) Wait() error {
	g.wg.Wait()
	err := g.combinedErr()
	if g.cancel != nil {
		g.cancel(err)
	}
	return err
}

// Err returns the Group's current combined error without blocking. It
// reflects whatever tasks have reported so far and may change until Wait
// has returned. It is safe to call concurrently with Go, TryGo, Fail, and
// Wait.
func (g *Group) Err() error {
	return g.combinedErr()
}

// Errors returns every error reported to the Group so far, without
// blocking. In CollectErrors mode this is every error reported by every
// task, in the order Fail observed them. In FirstError mode it is either
// nil or a single-element slice containing the one error that was kept.
// It is safe to call concurrently with Go, TryGo, Fail, and Wait.
func (g *Group) Errors() []error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.mode == CollectErrors {
		out := make([]error, len(g.errs))
		copy(out, g.errs)
		return out
	}
	if g.err == nil {
		return nil
	}
	return []error{g.err}
}

// SetLimit changes the number of goroutines the Group will run
// concurrently to n. A negative n removes the limit entirely. A limit of
// zero prevents any new goroutine from starting until the limit is raised
// again.
//
// Unlike golang.org/x/sync/errgroup.Group.SetLimit, this method never
// panics and may be called at any time, including while goroutines started
// by Go or TryGo are still active: any goroutine currently blocked in Go
// waiting for a free slot observes the new limit immediately, and this is
// precisely the mechanism that lets a Group (and the WorkerPool built on
// top of it) be scaled up or down live.
func (g *Group) SetLimit(n int) {
	g.limiter().setLimit(n)
}

// Scale adjusts the current concurrency limit by delta and returns the
// resulting limit. For example, g.Scale(4) raises the limit by 4 and
// g.Scale(-2) lowers it by 2. If the Group is currently unlimited, it is
// first treated as having a limit of 0 before delta is applied. The
// resulting limit is never allowed to go below 0. Like SetLimit, Scale is
// safe to call at any time, including while goroutines are active.
func (g *Group) Scale(delta int) int {
	return g.limiter().adjust(delta)
}

// Limit reports the Group's current concurrency limit. A negative result
// means the Group is unlimited.
func (g *Group) Limit() int {
	return g.limiter().cap()
}

// Running reports the number of goroutines started by Go or TryGo that
// have not yet returned.
func (g *Group) Running() int {
	return g.limiter().used()
}
