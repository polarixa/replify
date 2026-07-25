package crontask_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/polarixa/replify/pkg/crontask"
)

// ─────────────────────────────────────────────────────────────────────────────
// ChainHooks
// ─────────────────────────────────────────────────────────────────────────────

// TestChainHooksZero verifies that ChainHooks with no args returns NoopHooks.
func TestChainHooksZero(t *testing.T) {
	t.Parallel()
	h := crontask.ChainHooks()
	// Must satisfy the Hooks interface without panicking.
	ctx := context.Background()
	h.OnStart(ctx, "id")
	h.OnSuccess(ctx, "id", time.Millisecond)
	h.OnFailure(ctx, "id", time.Millisecond, errors.New("e"))
	h.OnComplete(ctx, "id", time.Millisecond)
}

// TestChainHooksSingle verifies that ChainHooks returns the single hook as-is.
func TestChainHooksSingle(t *testing.T) {
	t.Parallel()
	var called int32
	h := &cronTestHooks{
		onStart: func(_, _ string) { atomic.AddInt32(&called, 1) },
	}
	chain := crontask.ChainHooks(h)
	chain.OnStart(context.Background(), "id")
	if atomic.LoadInt32(&called) != 1 {
		t.Error("expected single hook to be called")
	}
}

// TestChainHooksMultiple verifies that ChainHooks dispatches to all members.
func TestChainHooksMultiple(t *testing.T) {
	t.Parallel()
	var count int32
	inc := func(_, _ string) { atomic.AddInt32(&count, 1) }
	h1 := &cronTestHooks{onStart: inc}
	h2 := &cronTestHooks{onStart: inc}
	chain := crontask.ChainHooks(h1, h2)
	chain.OnStart(context.Background(), "id")
	if atomic.LoadInt32(&count) != 2 {
		t.Errorf("expected count=2, got %d", atomic.LoadInt32(&count))
	}
}

// TestChainHooksNilSkipped verifies that nil entries in ChainHooks are ignored.
func TestChainHooksNilSkipped(t *testing.T) {
	t.Parallel()
	var called int32
	h := &cronTestHooks{
		onStart: func(_, _ string) { atomic.AddInt32(&called, 1) },
	}
	chain := crontask.ChainHooks(nil, h, nil)
	chain.OnStart(context.Background(), "id")
	if atomic.LoadInt32(&called) != 1 {
		t.Error("expected exactly one call from non-nil hook")
	}
}

// TestChainHooksRetryAndPanic verifies that ChainHooks forwards optional
// RetryHook and PanicHook calls to members that implement them.
func TestChainHooksRetryAndPanic(t *testing.T) {
	t.Parallel()

	var retryCalled, panicCalled int32
	h := &fullTestHook{
		onRetry: func(_, _ string, _ int, _ error) { atomic.AddInt32(&retryCalled, 1) },
		onPanic: func(_, _ string, _ any) { atomic.AddInt32(&panicCalled, 1) },
	}
	chain := crontask.ChainHooks(crontask.NoopHooks{}, h) // mix of plain + full

	// The chain implements RetryHook because at least one member does.
	if rh, ok := chain.(crontask.RetryHook); ok {
		rh.OnRetry(context.Background(), "id", 1, errors.New("e"))
	} else {
		t.Error("chain should implement RetryHook")
	}
	if atomic.LoadInt32(&retryCalled) != 1 {
		t.Error("OnRetry was not forwarded")
	}

	if ph, ok := chain.(crontask.PanicHook); ok {
		ph.OnPanic(context.Background(), "id", "boom")
	} else {
		t.Error("chain should implement PanicHook")
	}
	if atomic.LoadInt32(&panicCalled) != 1 {
		t.Error("OnPanic was not forwarded")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// LoggingHook
// ─────────────────────────────────────────────────────────────────────────────

// TestLoggingHookInterface verifies that LoggingHook satisfies Hooks, RetryHook,
// and PanicHook without panicking.
func TestLoggingHookInterface(t *testing.T) {
	t.Parallel()
	h := crontask.LoggingHook()
	ctx := context.Background()
	h.OnStart(ctx, "job-1")
	h.OnSuccess(ctx, "job-1", 50*time.Millisecond)
	h.OnFailure(ctx, "job-1", 10*time.Millisecond, errors.New("oops"))
	h.OnComplete(ctx, "job-1", 50*time.Millisecond)

	if rh, ok := h.(crontask.RetryHook); ok {
		rh.OnRetry(ctx, "job-1", 1, errors.New("retry"))
	}
	if ph, ok := h.(crontask.PanicHook); ok {
		ph.OnPanic(ctx, "job-1", "panic value")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// MetricsHook
// ─────────────────────────────────────────────────────────────────────────────

// TestMetricsHookCounters verifies that success and failure counters increment
// correctly.
func TestMetricsHookCounters(t *testing.T) {
	t.Parallel()

	m := crontask.MetricsHook()
	ctx := context.Background()
	d := 10 * time.Millisecond

	m.OnSuccess(ctx, "j", d)
	m.OnSuccess(ctx, "j", d)
	m.OnFailure(ctx, "j", d, errors.New("e"))

	if m.SuccessCount() != 2 {
		t.Errorf("SuccessCount = %d, want 2", m.SuccessCount())
	}
	if m.FailureCount() != 1 {
		t.Errorf("FailureCount = %d, want 1", m.FailureCount())
	}
	if m.TotalDuration() != 3*d {
		t.Errorf("TotalDuration = %v, want %v", m.TotalDuration(), 3*d)
	}
}

// TestMetricsHookPanic verifies that the panic counter increments via OnPanic.
func TestMetricsHookPanic(t *testing.T) {
	t.Parallel()
	m := crontask.MetricsHook()
	ph, ok := crontask.Hooks(m).(crontask.PanicHook)
	if !ok {
		t.Fatal("MetricsHookInstance should implement PanicHook")
	}
	ph.OnPanic(context.Background(), "j", "boom")
	if m.PanicCount() != 1 {
		t.Errorf("PanicCount = %d, want 1", m.PanicCount())
	}
}

// TestMetricsHookWithScheduler verifies that MetricsHook accumulates counts in
// a running scheduler.
func TestMetricsHookWithScheduler(t *testing.T) {
	t.Parallel()

	m := crontask.MetricsHook()
	s, _ := crontask.New(
		crontask.WithSeconds(),
		crontask.WithSchedulerHooks(m),
	)
	_, _ = s.Register("@every 100ms", func(_ context.Context) error { return nil })
	_ = s.Start()
	defer s.Stop()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if m.SuccessCount() >= 1 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if m.SuccessCount() < 1 {
		t.Error("MetricsHook: expected at least one success")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// RecoverPanicHook
// ─────────────────────────────────────────────────────────────────────────────

// TestRecoverPanicHookInterface verifies that RecoverPanicHook implements
// the optional PanicHook interface.
func TestRecoverPanicHookInterface(t *testing.T) {
	t.Parallel()
	h := crontask.RecoverPanicHook()
	if _, ok := h.(crontask.PanicHook); !ok {
		t.Error("RecoverPanicHook should implement PanicHook")
	}
}

// TestRecoverPanicHookWithHandlerCalled verifies that the custom handler is
// invoked on panic.
func TestRecoverPanicHookWithHandlerCalled(t *testing.T) {
	t.Parallel()
	var captured any
	h := crontask.RecoverPanicHookWithHandler(func(_ context.Context, _ string, r any) {
		captured = r
	})
	ph := h.(crontask.PanicHook)
	ph.OnPanic(context.Background(), "j", "test-panic")
	if captured != "test-panic" {
		t.Errorf("expected recovered=%q, got %v", "test-panic", captured)
	}
}

// TestPanicRecoveryInScheduler verifies that a panicking job does not crash
// the process and that OnPanic is called.
func TestPanicRecoveryInScheduler(t *testing.T) {
	t.Parallel()

	var panicCalled int32
	hook := crontask.RecoverPanicHookWithHandler(func(_ context.Context, _ string, _ any) {
		atomic.AddInt32(&panicCalled, 1)
	})

	s, _ := crontask.New(crontask.WithSeconds())
	_, _ = s.Register("@every 150ms", func(_ context.Context) error {
		panic("deliberate test panic")
	}, crontask.WithHooks(hook))

	_ = s.Start()
	defer s.Stop()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(&panicCalled) >= 1 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if atomic.LoadInt32(&panicCalled) < 1 {
		t.Error("OnPanic was not called for a panicking job")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// RetryLoggerHook
// ─────────────────────────────────────────────────────────────────────────────

// TestRetryLoggerHookInterface verifies that RetryLoggerHook implements the
// optional RetryHook interface.
func TestRetryLoggerHookInterface(t *testing.T) {
	t.Parallel()
	h := crontask.RetryLoggerHook()
	if _, ok := h.(crontask.RetryHook); !ok {
		t.Error("RetryLoggerHook should implement RetryHook")
	}
	// Call OnRetry to ensure it does not panic.
	rh := h.(crontask.RetryHook)
	rh.OnRetry(context.Background(), "j", 1, errors.New("err"))
}

// TestRetryHookCalledDuringRetry verifies that the RetryHook.OnRetry method is
// invoked for each failed attempt in the retry loop.
func TestRetryHookCalledDuringRetry(t *testing.T) {
	t.Parallel()

	var retryCalls int32
	h := &fullTestHook{
		onRetry: func(_, _ string, _ int, _ error) {
			atomic.AddInt32(&retryCalls, 1)
		},
	}

	s, _ := crontask.New(crontask.WithSeconds())
	_, _ = s.Register("@every 200ms", func(_ context.Context) error {
		return errors.New("always fails")
	}, crontask.WithMaxRetries(2), crontask.WithHooks(h))

	_ = s.Start()
	defer s.Stop()

	// 2 retries per activation; wait for at least 2 retry callbacks.
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(&retryCalls) >= 2 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if atomic.LoadInt32(&retryCalls) < 2 {
		t.Errorf("expected at least 2 OnRetry calls, got %d", atomic.LoadInt32(&retryCalls))
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// TimeoutLoggerHook
// ─────────────────────────────────────────────────────────────────────────────

// TestTimeoutLoggerHookLogsOnTimeout verifies that TimeoutLoggerHook logs when
// the error is context.DeadlineExceeded.
func TestTimeoutLoggerHookLogsOnTimeout(t *testing.T) {
	t.Parallel()
	h := crontask.TimeoutLoggerHook()
	// Should not panic; logs are written to the standard logger.
	h.OnFailure(context.Background(), "j", 100*time.Millisecond, context.DeadlineExceeded)
	// Non-timeout errors should also not panic.
	h.OnFailure(context.Background(), "j", 100*time.Millisecond, errors.New("other"))
}

// ─────────────────────────────────────────────────────────────────────────────
// ConcurrencyLimiterHook
// ─────────────────────────────────────────────────────────────────────────────

// TestConcurrencyLimiterHookActive verifies that the Active() counter reflects
// in-flight executions.
func TestConcurrencyLimiterHookActive(t *testing.T) {
	t.Parallel()
	lim := crontask.ConcurrencyLimiterHook(2)
	ctx := context.Background()

	if lim.Active() != 0 {
		t.Errorf("initial Active() = %d, want 0", lim.Active())
	}
	lim.OnStart(ctx, "j")
	if lim.Active() != 1 {
		t.Errorf("Active() after one OnStart = %d, want 1", lim.Active())
	}
	lim.OnStart(ctx, "j")
	if lim.Active() != 2 {
		t.Errorf("Active() after two OnStart = %d, want 2", lim.Active())
	}
	lim.OnComplete(ctx, "j", time.Millisecond)
	if lim.Active() != 1 {
		t.Errorf("Active() after OnComplete = %d, want 1", lim.Active())
	}
	lim.OnComplete(ctx, "j", time.Millisecond)
	if lim.Active() != 0 {
		t.Errorf("Active() after second OnComplete = %d, want 0", lim.Active())
	}
}

// TestConcurrencyLimiterHookContextCancel verifies that a cancelled context
// does not block OnStart and that a subsequent OnComplete is safe.
func TestConcurrencyLimiterHookContextCancel(t *testing.T) {
	t.Parallel()
	lim := crontask.ConcurrencyLimiterHook(1)

	// Fill the one available slot.
	lim.OnStart(context.Background(), "j")

	// Attempt to acquire with a cancelled context — must return immediately.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan struct{})
	go func() {
		lim.OnStart(ctx, "j") // should unblock because ctx is already done
		close(done)
	}()

	select {
	case <-done:
		// pass — OnStart returned without blocking
	case <-time.After(500 * time.Millisecond):
		t.Error("OnStart blocked on a cancelled context")
	}

	// Releasing should be a no-op for the cancelled acquire.
	lim.OnComplete(ctx, "j", 0)
}

// ─────────────────────────────────────────────────────────────────────────────
// WithSchedulerHooks (scheduler-level default hooks)
// ─────────────────────────────────────────────────────────────────────────────

// TestWithSchedulerHooksApplied verifies that jobs without per-job hooks
// receive the scheduler-level default hooks.
func TestWithSchedulerHooksApplied(t *testing.T) {
	t.Parallel()

	var started int32
	h := &cronTestHooks{
		onStart: func(_, _ string) { atomic.AddInt32(&started, 1) },
	}
	s, _ := crontask.New(
		crontask.WithSeconds(),
		crontask.WithSchedulerHooks(h),
	)
	// Register a job with no per-job hooks — default should be used.
	_, _ = s.Register("@every 100ms", func(_ context.Context) error { return nil })
	_ = s.Start()
	defer s.Stop()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(&started) >= 1 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if atomic.LoadInt32(&started) < 1 {
		t.Error("scheduler-level default hook OnStart was not called")
	}
}

// TestWithSchedulerHooksOverriddenByJobHooks verifies that per-job hooks take
// precedence over the scheduler-level default hooks.
func TestWithSchedulerHooksOverriddenByJobHooks(t *testing.T) {
	t.Parallel()

	var defaultCalled, jobCalled int32
	defaultHook := &cronTestHooks{
		onStart: func(_, _ string) { atomic.AddInt32(&defaultCalled, 1) },
	}
	jobHook := &cronTestHooks{
		onStart: func(_, _ string) { atomic.AddInt32(&jobCalled, 1) },
	}

	s, _ := crontask.New(
		crontask.WithSeconds(),
		crontask.WithSchedulerHooks(defaultHook),
	)
	// This job supplies its own hook — default should NOT be called for it.
	_, _ = s.Register("@every 100ms", func(_ context.Context) error { return nil },
		crontask.WithHooks(jobHook))
	_ = s.Start()
	defer s.Stop()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(&jobCalled) >= 1 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if atomic.LoadInt32(&jobCalled) < 1 {
		t.Error("per-job hook was not called")
	}
	if atomic.LoadInt32(&defaultCalled) > 0 {
		t.Error("scheduler-level default hook should not be called when job has its own hooks")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// WithHooks variadic (backward-compat)
// ─────────────────────────────────────────────────────────────────────────────

// TestWithHooksVariadic verifies that multiple hooks can be passed to WithHooks
// and all are invoked.
func TestWithHooksVariadic(t *testing.T) {
	t.Parallel()

	var count int32
	inc := func(_, _ string) { atomic.AddInt32(&count, 1) }
	h1 := &cronTestHooks{onStart: inc}
	h2 := &cronTestHooks{onStart: inc}

	s, _ := crontask.New(crontask.WithSeconds())
	_, _ = s.Register("@every 100ms", func(_ context.Context) error { return nil },
		crontask.WithHooks(h1, h2))
	_ = s.Start()
	defer s.Stop()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(&count) >= 2 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if atomic.LoadInt32(&count) < 2 {
		t.Errorf("expected at least 2 OnStart calls (one per hook), got %d", count)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

// fullTestHook is a test helper that implements Hooks, RetryHook, and PanicHook.
type fullTestHook struct {
	crontask.NoopHooks
	onRetry func(placeholder, id string, attempt int, err error)
	onPanic func(placeholder, id string, recovered any)
}

func (h *fullTestHook) OnRetry(_ context.Context, id string, attempt int, err error) {
	if h.onRetry != nil {
		h.onRetry("", id, attempt, err)
	}
}

func (h *fullTestHook) OnPanic(_ context.Context, id string, recovered any) {
	if h.onPanic != nil {
		h.onPanic("", id, recovered)
	}
}
