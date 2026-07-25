package crontask

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/polarixa/replify/pkg/slogger"
)

// CustomHooks is a struct that implements the Hooks interface with custom
// implementations of the methods you care about. It embeds NoopHooks to
// provide default no-op implementations for the remaining methods.
//
// Example:
//
//	type MyHooks struct {
//		crontask.NoopHooks
//	}
//
//	func (h *MyHooks) OnFailure(_ context.Context, id string, _ time.Duration, err error) {
//		log.Printf("ALERT: job %s failed: %v", id, err)
//	}

// OnStart implements Hooks. It does nothing.
func (NoopHooks) OnStart(_ context.Context, _ string) {}

// OnSuccess implements Hooks. It does nothing.
func (NoopHooks) OnSuccess(_ context.Context, _ string, _ time.Duration) {}

// OnFailure implements Hooks. It does nothing.
func (NoopHooks) OnFailure(_ context.Context, _ string, _ time.Duration, _ error) {}

// OnComplete implements Hooks. It does nothing.
func (NoopHooks) OnComplete(_ context.Context, _ string, _ time.Duration) {}

// ─────────────────────────────────────────────────────────────────────────────
// Builtin Hooks
// ─────────────────────────────────────────────────────────────────────────────

// ChainHooks composes multiple Hooks implementations into a single Hooks that
// dispatches each method call to all members in order. Nil members are silently
// skipped.
//
// The returned chain also implements RetryHook and PanicHook by delegating to
// any member that implements those optional interfaces; members that do not
// implement them are skipped.
//
// When zero non-nil hooks are supplied ChainHooks returns NoopHooks. When
// exactly one non-nil hook is supplied it is returned as-is (no wrapping).
//
// Example:
//
//	hooks := crontask.ChainHooks(
//	    crontask.LoggingHook(),
//	    crontask.MetricsHook(),
//	    &myCustomHook{},
//	)
//	s.Register("@daily", fn, crontask.WithHooks(hooks))
func ChainHooks(hooks ...Hooks) Hooks {
	active := make([]Hooks, 0, len(hooks))
	for _, h := range hooks {
		if h != nil {
			active = append(active, h)
		}
	}
	switch len(active) {
	case 0:
		return NoopHooks{}
	case 1:
		return active[0]
	}
	return &chainedHooks{hooks: active}
}

// OnStart dispatches the call to all hooks in the chain.
//
// chainedHooks dispatches hook method calls to a slice of Hooks implementations.
// Nil members are skipped. It implements RetryHook and PanicHook by forwarding
// to any member that implements those optional interfaces.
func (c *chainedHooks) OnStart(ctx context.Context, jobID string) {
	for _, h := range c.hooks {
		h.OnStart(ctx, jobID)
	}
}

// OnSuccess dispatches the call to all hooks in the chain.
//
// chainedHooks dispatches hook method calls to a slice of Hooks implementations.
// Nil members are skipped. It implements RetryHook and PanicHook by forwarding
// to any member that implements those optional interfaces.
func (c *chainedHooks) OnSuccess(ctx context.Context, jobID string, d time.Duration) {
	for _, h := range c.hooks {
		h.OnSuccess(ctx, jobID, d)
	}
}

// OnFailure dispatches the call to all hooks in the chain.
//
// chainedHooks dispatches hook method calls to a slice of Hooks implementations.
// Nil members are skipped. It implements RetryHook and PanicHook by forwarding
// to any member that implements those optional interfaces.
func (c *chainedHooks) OnFailure(ctx context.Context, jobID string, d time.Duration, err error) {
	for _, h := range c.hooks {
		h.OnFailure(ctx, jobID, d, err)
	}
}

// OnComplete dispatches the call to all hooks in the chain.
//
// chainedHooks dispatches hook method calls to a slice of Hooks implementations.
// Nil members are skipped. It implements RetryHook and PanicHook by forwarding
// to any member that implements those optional interfaces.
func (c *chainedHooks) OnComplete(ctx context.Context, jobID string, d time.Duration) {
	for _, h := range c.hooks {
		h.OnComplete(ctx, jobID, d)
	}
}

// OnRetry implements RetryHook by forwarding to all members that support it.
//
// chainedHooks dispatches hook method calls to a slice of Hooks implementations.
// Nil members are skipped. It implements RetryHook and PanicHook by forwarding
// to any member that implements those optional interfaces.
func (c *chainedHooks) OnRetry(ctx context.Context, jobID string, attempt int, err error) {
	for _, h := range c.hooks {
		if rh, ok := h.(RetryHook); ok {
			rh.OnRetry(ctx, jobID, attempt, err)
		}
	}
}

// OnPanic implements PanicHook by forwarding to all members that support it.
//
// chainedHooks dispatches hook method calls to a slice of Hooks implementations.
// Nil members are skipped. It implements RetryHook and PanicHook by forwarding
// to any member that implements those optional interfaces.
func (c *chainedHooks) OnPanic(ctx context.Context, jobID string, recovered any) {
	for _, h := range c.hooks {
		if ph, ok := h.(PanicHook); ok {
			ph.OnPanic(ctx, jobID, recovered)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// LoggingHook
// ─────────────────────────────────────────────────────────────────────────────

// LoggingHook returns a Hooks implementation that logs each lifecycle event
// (start, success, failure, complete) using the standard log package. All
// messages are prefixed with "[crontask]" for easy filtering.
//
// Logging is synchronous and lightweight; it delegates directly to log.Printf
// which is safe for concurrent use.
//
// Example:
//
//	s, _ := crontask.New(
//	    crontask.WithSchedulerHooks(crontask.LoggingHook()),
//	)
func LoggingHook() Hooks {
	return &loggingHook{}
}

// loggingHook is a Hooks implementation that logs each lifecycle event
// (start, success, failure, complete) using the standard log package.
// All messages are prefixed with "[crontask]" for easy filtering.
//
// Logging is synchronous and lightweight; it delegates directly to log.Printf
// which is safe for concurrent use.
//
// Example:
//
//	s, _ := crontask.New(
//	    crontask.WithSchedulerHooks(crontask.LoggingHook()),
//	)
func (h *loggingHook) OnStart(_ context.Context, jobID string) {
	slogger.Infof("[crontask] job %s: starting", jobID)
}

func (h *loggingHook) OnSuccess(_ context.Context, jobID string, d time.Duration) {
	slogger.Infof("[crontask] job %s: succeeded in %s", jobID, d.Round(time.Millisecond))
}

func (h *loggingHook) OnFailure(_ context.Context, jobID string, d time.Duration, err error) {
	slogger.Errorf("[crontask] job %s: failed in %s: %v", jobID, d.Round(time.Millisecond), err)
}

func (h *loggingHook) OnComplete(_ context.Context, jobID string, d time.Duration) {
	slogger.Infof("[crontask] job %s: completed in %s", jobID, d.Round(time.Millisecond))
}

// OnRetry implements RetryHook for loggingHook.
func (h *loggingHook) OnRetry(_ context.Context, jobID string, attempt int, err error) {
	slogger.Warnf("[crontask] job %s: attempt %d failed, will retry: %v", jobID, attempt, err)
}

// OnPanic implements PanicHook for loggingHook.
func (h *loggingHook) OnPanic(_ context.Context, jobID string, recovered any) {
	slogger.Errorf("[crontask] job %s: PANIC recovered: %v", jobID, recovered)
}

// ─────────────────────────────────────────────────────────────────────────────
// MetricsHook
// ─────────────────────────────────────────────────────────────────────────────

// MetricsHook returns a *MetricsHookInstance that implements Hooks, RetryHook,
// and PanicHook. Counters and durations are updated atomically and can be read
// at any time from any goroutine without external synchronisation.
//
// Example:
//
//	m := crontask.MetricsHook()
//	s, _ := crontask.New(crontask.WithSchedulerHooks(m))
//	// ... later ...
//	fmt.Printf("successes=%d failures=%d", m.SuccessCount(), m.FailureCount())
func MetricsHook() *MetricsHookInstance {
	return &MetricsHookInstance{}
}

// SuccessCount returns the total number of successful job invocations.
func (m *MetricsHookInstance) SuccessCount() int64 {
	return atomic.LoadInt64(&m.successCount)
}

// FailureCount returns the total number of failed job invocations (after all
// retries are exhausted).
func (m *MetricsHookInstance) FailureCount() int64 {
	return atomic.LoadInt64(&m.failureCount)
}

// PanicCount returns the total number of job invocations that panicked.
func (m *MetricsHookInstance) PanicCount() int64 {
	return atomic.LoadInt64(&m.panicCount)
}

// TotalDuration returns the cumulative execution time across all invocations.
func (m *MetricsHookInstance) TotalDuration() time.Duration {
	return time.Duration(atomic.LoadInt64(&m.totalDurNs))
}

func (m *MetricsHookInstance) OnSuccess(_ context.Context, _ string, d time.Duration) {
	atomic.AddInt64(&m.successCount, 1)
	atomic.AddInt64(&m.totalDurNs, int64(d))
}

func (m *MetricsHookInstance) OnFailure(_ context.Context, _ string, d time.Duration, _ error) {
	atomic.AddInt64(&m.failureCount, 1)
	atomic.AddInt64(&m.totalDurNs, int64(d))
}

// OnPanic implements PanicHook for MetricsHookInstance.
func (m *MetricsHookInstance) OnPanic(_ context.Context, _ string, _ any) {
	atomic.AddInt64(&m.panicCount, 1)
}

// ─────────────────────────────────────────────────────────────────────────────
// RecoverPanicHook
// ─────────────────────────────────────────────────────────────────────────────

// RecoverPanicHook returns a Hooks implementation that silently recovers
// panics in job functions and logs them with the standard log package. The
// scheduler loop is never interrupted by a panicking job.
//
// To supply a custom panic handler (e.g. to send to Sentry or PagerDuty),
// use RecoverPanicHookWithHandler.
//
// Example:
//
//	s.Register("@daily", riskyFn, crontask.WithHooks(crontask.RecoverPanicHook()))
func RecoverPanicHook() Hooks {
	return RecoverPanicHookWithHandler(func(_ context.Context, jobID string, recovered any) {
		slogger.Errorf("[crontask] job %s: PANIC recovered: %v", jobID, recovered)
	})
}

// RecoverPanicHookWithHandler returns a Hooks implementation that calls
// handler whenever the job function panics. handler is called synchronously
// inside the executor goroutine; it must not panic itself.
//
// Example:
//
//	hook := crontask.RecoverPanicHookWithHandler(func(_ context.Context, id string, r any) {
//	    alerting.Send(fmt.Sprintf("job %s panicked: %v", id, r))
//	})
//	s.Register("@daily", fn, crontask.WithHooks(hook))
func RecoverPanicHookWithHandler(handler func(ctx context.Context, jobID string, recovered any)) Hooks {
	if handler == nil {
		handler = func(_ context.Context, _ string, _ any) {}
	}
	return &recoverPanicHook{handler: handler}
}

// OnPanic implements PanicHook.
func (h *recoverPanicHook) OnPanic(ctx context.Context, jobID string, recovered any) {
	h.handler(ctx, jobID, recovered)
}

// ─────────────────────────────────────────────────────────────────────────────
// RetryLoggerHook
// ─────────────────────────────────────────────────────────────────────────────

// RetryLoggerHook returns a Hooks implementation that logs every retry attempt
// via the standard log package. It implements the optional RetryHook interface
// so it receives per-attempt callbacks rather than only the final failure.
//
// Example:
//
//	s.Register("@every 5m", fn,
//	    crontask.WithMaxRetries(3),
//	    crontask.WithHooks(crontask.RetryLoggerHook()),
//	)
func RetryLoggerHook() Hooks {
	return &retryLoggerHook{}
}

// OnRetry implements RetryHook.
func (h *retryLoggerHook) OnRetry(_ context.Context, jobID string, attempt int, err error) {
	slogger.Warnf("[crontask] job %s: attempt %d failed, retrying: %v", jobID, attempt, err)
}

// ─────────────────────────────────────────────────────────────────────────────
// TimeoutLoggerHook
// ─────────────────────────────────────────────────────────────────────────────

// TimeoutLoggerHook returns a Hooks implementation that logs a warning
// whenever a job invocation is terminated due to a timeout (context.DeadlineExceeded).
// Non-timeout failures are passed through without logging.
//
// Example:
//
//	s.Register("@minutely", fn,
//	    crontask.WithTimeout(10*time.Second),
//	    crontask.WithHooks(crontask.TimeoutLoggerHook()),
//	)
func TimeoutLoggerHook() Hooks {
	return &timeoutLoggerHook{}
}

func (h *timeoutLoggerHook) OnFailure(_ context.Context, jobID string, d time.Duration, err error) {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, ErrJobTimeout) {
		slogger.Warnf("[crontask] job %s: timed out after %s: %v", jobID, d.Round(time.Millisecond), err)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ConcurrencyLimiterHook
// ─────────────────────────────────────────────────────────────────────────────

// ConcurrencyLimiterHook returns a Hooks implementation that limits the number
// of concurrent executions to maxConcurrent. Additional executions block in
// OnStart until a slot becomes available or the job's context is cancelled.
//
// This is most useful when the same scheduler runs many instances of a heavy
// job and you want to avoid overwhelming downstream resources:
//
//	limiter := crontask.ConcurrencyLimiterHook(3)
//	for i := 0; i < 10; i++ {
//	    s.Register("@every 1m", heavyFn, crontask.WithHooks(limiter))
//	}
//
// If maxConcurrent is ≤ 0 it is treated as 1 (serial execution).
//
// The same ConcurrencyLimiterHook instance must be shared across all jobs
// that should count against the same limit; passing different instances to
// different jobs creates independent limits.
func ConcurrencyLimiterHook(maxConcurrent int) *ConcurrencyLimiterHookInstance {
	if maxConcurrent <= 0 {
		maxConcurrent = 1
	}
	return &ConcurrencyLimiterHookInstance{
		sem: make(chan struct{}, maxConcurrent),
	}
}

// OnStart acquires one concurrency slot. It blocks until a slot is available
// or the context is done. When the context is cancelled before a slot is
// acquired the method returns without acquiring, and OnComplete becomes a
// no-op for this invocation (the semaphore remains unmodified).
func (h *ConcurrencyLimiterHookInstance) OnStart(ctx context.Context, _ string) {
	select {
	case h.sem <- struct{}{}:
		// slot acquired
	case <-ctx.Done():
		// context cancelled; skip acquisition so OnComplete does not release
	}
}

// OnComplete releases the previously acquired concurrency slot. If no slot was
// acquired (context was cancelled in OnStart), the non-blocking select ensures
// the semaphore is not under-released.
func (h *ConcurrencyLimiterHookInstance) OnComplete(_ context.Context, _ string, _ time.Duration) {
	select {
	case <-h.sem:
		// slot released
	default:
		// nothing to release (slot was never acquired)
	}
}

// Active returns the number of job invocations currently holding a slot.
func (h *ConcurrencyLimiterHookInstance) Active() int {
	return len(h.sem)
}
