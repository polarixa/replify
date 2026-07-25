package crontask_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/polarixa/replify/pkg/crontask"
)

// TestNewScheduler verifies that New returns a valid, non-running Scheduler.
func TestNewScheduler(t *testing.T) {
	t.Parallel()
	s, err := crontask.New()
	if err != nil {
		t.Fatalf("New(): unexpected error: %v", err)
	}
	if s.IsRunning() {
		t.Error("scheduler should not be running after New()")
	}
}

// TestStartStop verifies that the scheduler can be started and stopped.
func TestStartStop(t *testing.T) {
	t.Parallel()
	s, _ := crontask.New()
	if err := s.Start(); err != nil {
		t.Fatalf("Start(): %v", err)
	}
	if !s.IsRunning() {
		t.Error("IsRunning() should be true after Start()")
	}
	s.Stop()
	// Allow the goroutine to exit.
	time.Sleep(50 * time.Millisecond)
	if s.IsRunning() {
		t.Error("IsRunning() should be false after Stop()")
	}
}

// TestDoubleStart verifies that starting an already-running scheduler returns
// ErrSchedulerRunning.
func TestDoubleStart(t *testing.T) {
	t.Parallel()
	s, _ := crontask.New()
	_ = s.Start()
	defer s.Stop()
	err := s.Start()
	if !errors.Is(err, crontask.ErrSchedulerRunning) {
		t.Errorf("expected ErrSchedulerRunning, got %v", err)
	}
}

// TestShutdown verifies graceful shutdown completes within a reasonable
// deadline.
func TestShutdown(t *testing.T) {
	t.Parallel()
	s, _ := crontask.New()
	_ = s.Start()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown(): unexpected error: %v", err)
	}
	if s.IsRunning() {
		t.Error("scheduler should not be running after Shutdown()")
	}
}

// TestRegisterInvalidExpression verifies that Register rejects bad expressions.
func TestRegisterInvalidExpression(t *testing.T) {
	t.Parallel()
	s, _ := crontask.New()
	_, err := s.Register("not a cron", func(_ context.Context) error { return nil })
	if err == nil {
		t.Error("expected error for invalid expression, got nil")
	}
}

// TestRegisterNilFunc verifies that Register rejects nil job functions.
func TestRegisterNilFunc(t *testing.T) {
	t.Parallel()
	s, _ := crontask.New()
	_, err := s.Register("* * * * *", nil)
	if err == nil {
		t.Error("expected error for nil function, got nil")
	}
}

// TestRegisterAndRemove verifies the full registration and removal lifecycle.
func TestRegisterAndRemove(t *testing.T) {
	t.Parallel()
	s, _ := crontask.New()
	id, err := s.Register("* * * * *", func(_ context.Context) error { return nil })
	if err != nil {
		t.Fatalf("Register(): %v", err)
	}
	if id == "" {
		t.Error("expected non-empty job ID")
	}
	if err := s.Remove(id); err != nil {
		t.Errorf("Remove(%q): %v", id, err)
	}
	// Removing a second time should return ErrJobNotFound.
	if err := s.Remove(id); !errors.Is(err, crontask.ErrJobNotFound) {
		t.Errorf("expected ErrJobNotFound, got %v", err)
	}
}

// TestJobs verifies that Jobs returns a snapshot of registered jobs.
func TestJobs(t *testing.T) {
	t.Parallel()
	s, _ := crontask.New()
	_, _ = s.Register("* * * * *", func(_ context.Context) error { return nil },
		crontask.WithJobName("test-job"))
	jobs := s.Jobs()
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if jobs[0].Name != "test-job" {
		t.Errorf("expected name %q, got %q", "test-job", jobs[0].Name)
	}
}

// TestSchedulerNextRuns verifies deterministic next-run calculation.
func TestSchedulerNextRuns(t *testing.T) {
	t.Parallel()
	s, _ := crontask.New()
	id, _ := s.Register("0 * * * *", func(_ context.Context) error { return nil })
	ref := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	runs, err := s.NextRuns(id, ref, 3)
	if err != nil {
		t.Fatalf("NextRuns(): %v", err)
	}
	if len(runs) != 3 {
		t.Fatalf("expected 3 runs, got %d", len(runs))
	}
	// "0 * * * *" fires at 01:00, 02:00, 03:00 after midnight.
	want := []time.Time{
		time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 2, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 3, 0, 0, 0, time.UTC),
	}
	for i, r := range runs {
		if !r.Equal(want[i]) {
			t.Errorf("runs[%d] = %v, want %v", i, r, want[i])
		}
	}
}

// TestSchedulerNextRunsNotFound verifies that NextRuns returns ErrJobNotFound
// for an unknown ID.
func TestSchedulerNextRunsNotFound(t *testing.T) {
	t.Parallel()
	s, _ := crontask.New()
	_, err := s.NextRuns("nonexistent", time.Now(), 1)
	if !errors.Is(err, crontask.ErrJobNotFound) {
		t.Errorf("expected ErrJobNotFound, got %v", err)
	}
}

// TestRegisterAfterShutdown verifies that Register returns ErrSchedulerStopped
// when called on a stopped scheduler.
func TestRegisterAfterShutdown(t *testing.T) {
	t.Parallel()
	s, _ := crontask.New()
	_ = s.Start()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = s.Shutdown(ctx)
	_, err := s.Register("* * * * *", func(_ context.Context) error { return nil })
	if !errors.Is(err, crontask.ErrSchedulerStopped) {
		t.Errorf("expected ErrSchedulerStopped, got %v", err)
	}
}

// TestJobExecution verifies that a registered job is actually executed by the
// running scheduler within a generous deadline.
func TestJobExecution(t *testing.T) {
	t.Parallel()
	s, _ := crontask.New(crontask.WithSeconds())

	var counter int64
	// "@every 100ms" fires approximately every 100 ms.
	_, err := s.Register("@every 100ms", func(_ context.Context) error {
		atomic.AddInt64(&counter, 1)
		return nil
	})
	if err != nil {
		t.Fatalf("Register(): %v", err)
	}

	_ = s.Start()
	defer s.Stop()

	// Wait up to 2 seconds for at least one execution.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if atomic.LoadInt64(&counter) >= 1 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if atomic.LoadInt64(&counter) < 1 {
		t.Error("expected at least one job execution within 2 seconds")
	}
}

// TestJobWithID verifies that a job registered with WithJobID uses the
// supplied identifier.
func TestJobWithID(t *testing.T) {
	t.Parallel()
	s, _ := crontask.New()
	const want = "my-custom-id"
	id, err := s.Register("* * * * *", func(_ context.Context) error { return nil },
		crontask.WithJobID(want))
	if err != nil {
		t.Fatalf("Register(): %v", err)
	}
	if id != want {
		t.Errorf("got ID %q, want %q", id, want)
	}
}

// TestHooksCalledOnSuccess verifies that OnStart, OnSuccess, and OnComplete
// are all invoked for a successful job.
func TestHooksCalledOnSuccess(t *testing.T) {
	t.Parallel()

	var started, succeeded, completed int32
	h := &cronTestHooks{
		onStart:    func(_, _ string) { atomic.AddInt32(&started, 1) },
		onSuccess:  func(_, _ string, _ time.Duration) { atomic.AddInt32(&succeeded, 1) },
		onComplete: func(_, _ string, _ time.Duration) { atomic.AddInt32(&completed, 1) },
	}

	s, _ := crontask.New(crontask.WithSeconds())
	_, _ = s.Register("@every 100ms", func(_ context.Context) error {
		return nil
	}, crontask.WithHooks(h))

	_ = s.Start()
	defer s.Stop()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(&succeeded) >= 1 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	if atomic.LoadInt32(&started) < 1 {
		t.Error("OnStart was not called")
	}
	if atomic.LoadInt32(&succeeded) < 1 {
		t.Error("OnSuccess was not called")
	}
	if atomic.LoadInt32(&completed) < 1 {
		t.Error("OnComplete was not called")
	}
}

// TestHooksCalledOnFailure verifies OnFailure is invoked for a failing job.
func TestHooksCalledOnFailure(t *testing.T) {
	t.Parallel()

	var failed int32
	h := &cronTestHooks{
		onFailure: func(_, _ string, _ time.Duration, _ error) { atomic.AddInt32(&failed, 1) },
	}

	s, _ := crontask.New(crontask.WithSeconds())
	_, _ = s.Register("@every 100ms", func(_ context.Context) error {
		return errors.New("simulated failure")
	}, crontask.WithHooks(h))

	_ = s.Start()
	defer s.Stop()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(&failed) >= 1 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	if atomic.LoadInt32(&failed) < 1 {
		t.Error("OnFailure was not called")
	}
}

// TestWithErrorHandler verifies the scheduler-level error callback.
func TestWithErrorHandler(t *testing.T) {
	t.Parallel()
	var errCount int32
	s, _ := crontask.New(
		crontask.WithSeconds(),
		crontask.WithErrorHandler(func(_ string, _ error) {
			atomic.AddInt32(&errCount, 1)
		}),
	)
	_, _ = s.Register("@every 100ms", func(_ context.Context) error {
		return errors.New("boom")
	})

	_ = s.Start()
	defer s.Stop()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(&errCount) >= 1 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	if atomic.LoadInt32(&errCount) < 1 {
		t.Error("error handler was not called")
	}
}

// TestRetryPolicy verifies that the executor retries a failing job the
// configured number of times.
func TestRetryPolicy(t *testing.T) {
	t.Parallel()

	var attempts int32
	s, _ := crontask.New(crontask.WithSeconds())
	_, _ = s.Register("@every 200ms", func(_ context.Context) error {
		atomic.AddInt32(&attempts, 1)
		return errors.New("always fails")
	}, crontask.WithMaxRetries(2))

	_ = s.Start()
	defer s.Stop()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(&attempts) >= 3 { // 1 original + 2 retries
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	if atomic.LoadInt32(&attempts) < 3 {
		t.Errorf("expected at least 3 attempts (1 + 2 retries), got %d", atomic.LoadInt32(&attempts))
	}
}

// TestJobTimeout verifies that a job's context is cancelled when the timeout
// expires.
func TestJobTimeout(t *testing.T) {
	t.Parallel()

	done := make(chan struct{})
	s, _ := crontask.New(crontask.WithSeconds())
	_, _ = s.Register("@every 200ms", func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			close(done)
			return ctx.Err()
		case <-time.After(10 * time.Second):
			return nil
		}
	}, crontask.WithTimeout(100*time.Millisecond))

	_ = s.Start()
	defer s.Stop()

	select {
	case <-done:
		// success: context was cancelled due to timeout
	case <-time.After(3 * time.Second):
		t.Error("job context was not cancelled by timeout")
	}
}

// TestWithLocation verifies that the scheduler respects a non-UTC location.
// Uses the exported Location() accessor added in the refactoring phase.
func TestWithLocation(t *testing.T) {
	t.Parallel()
	tz, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip("America/New_York not available:", err)
	}
	s, err := crontask.New(crontask.WithLocation(tz))
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	if s.Location() != tz {
		t.Errorf("expected location %v, got %v", tz, s.Location())
	}
}

// TestJobInfoRunCount verifies that the run counter increments on each
// execution.
func TestJobInfoRunCount(t *testing.T) {
	t.Parallel()
	s, _ := crontask.New(crontask.WithSeconds())

	id, _ := s.Register("@every 100ms", func(_ context.Context) error { return nil })
	_ = s.Start()
	defer s.Stop()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		jobs := s.Jobs()
		for _, j := range jobs {
			if j.ID == id && j.RunCount >= 2 {
				return
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Error("run count did not reach 2 within the deadline")
}

// TestBackoffPolicies exercises ConstantBackoff and ExponentialBackoff.
func TestBackoffPolicies(t *testing.T) {
	t.Parallel()

	cb := crontask.ConstantBackoff(5 * time.Second)
	for i := 1; i <= 5; i++ {
		if d := cb(i); d != 5*time.Second {
			t.Errorf("ConstantBackoff attempt %d = %v, want 5s", i, d)
		}
	}

	eb := crontask.ExponentialBackoff(time.Second)
	wantEB := []time.Duration{
		time.Second,      // attempt 1
		2 * time.Second,  // attempt 2
		4 * time.Second,  // attempt 3
		8 * time.Second,  // attempt 4
		16 * time.Second, // attempt 5
	}
	for i, w := range wantEB {
		if d := eb(i + 1); d != w {
			t.Errorf("ExponentialBackoff attempt %d = %v, want %v", i+1, d, w)
		}
	}
}

// TestNoopHooks ensures NoopHooks satisfies the Hooks interface and that all
// methods are safe to call without any side effects.
func TestNoopHooks(t *testing.T) {
	t.Parallel()
	var h crontask.Hooks = crontask.NoopHooks{}
	ctx := context.Background()
	h.OnStart(ctx, "id")
	h.OnSuccess(ctx, "id", time.Millisecond)
	h.OnFailure(ctx, "id", time.Millisecond, errors.New("e"))
	h.OnComplete(ctx, "id", time.Millisecond)
}

// TestJobErrorUnwrap verifies the JobError sentinel chain.
func TestJobErrorUnwrap(t *testing.T) {
	t.Parallel()
	inner := errors.New("inner")
	je := &crontask.JobError{JobID: "x", Attempt: 1, Err: inner}
	if !errors.Is(je, inner) {
		t.Error("errors.Is(JobError, inner) should be true")
	}
}

// TestSchedulerWithSecondsOption verifies the six-field mode option is stored.
// Uses the exported WithSecondsEnabled() accessor added in the refactoring phase.
func TestSchedulerWithSecondsOption(t *testing.T) {
	t.Parallel()
	s, _ := crontask.New(crontask.WithSeconds())
	if !s.WithSecondsEnabled() {
		t.Error("WithSecondsEnabled() should be true after WithSeconds()")
	}
}

// cronTestHooks is a lightweight Hooks implementation used in scheduler tests.
// The callback fields use distinct parameter names to avoid confusion with
// context.Context: the first string argument is an unused placeholder, and
// the second is the job ID.
type cronTestHooks struct {
	crontask.NoopHooks
	onStart    func(placeholder, id string)
	onSuccess  func(placeholder, id string, d time.Duration)
	onFailure  func(placeholder, id string, d time.Duration, err error)
	onComplete func(placeholder, id string, d time.Duration)
}

func (h *cronTestHooks) OnStart(_ context.Context, id string) {
	if h.onStart != nil {
		h.onStart("", id)
	}
}

func (h *cronTestHooks) OnSuccess(_ context.Context, id string, d time.Duration) {
	if h.onSuccess != nil {
		h.onSuccess("", id, d)
	}
}

func (h *cronTestHooks) OnFailure(_ context.Context, id string, d time.Duration, err error) {
	if h.onFailure != nil {
		h.onFailure("", id, d, err)
	}
}

func (h *cronTestHooks) OnComplete(_ context.Context, id string, d time.Duration) {
	if h.onComplete != nil {
		h.onComplete("", id, d)
	}
}
