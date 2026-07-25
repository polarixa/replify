package crontask

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/polarixa/replify/pkg/randn"
)

// Start begins the scheduler loop in a background goroutine. It returns
// ErrSchedulerRunning if the scheduler is already active or
// ErrSchedulerStopped if Shutdown has already been called.
func (s *Scheduler) Start() error {
	if atomic.LoadInt32(&s.stopped) == 1 {
		return ErrSchedulerStopped
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !atomic.CompareAndSwapInt32(&s.running, 0, 1) {
		return ErrSchedulerRunning
	}
	s.stopCh = make(chan struct{})
	s.doneCh = make(chan struct{})
	go s.loop(s.stopCh, s.doneCh)
	return nil
}

// Stop halts the scheduler loop without waiting for in-flight jobs to
// complete. To wait for all running jobs to finish, use Shutdown instead.
func (s *Scheduler) Stop() {
	s.mu.Lock()
	if atomic.LoadInt32(&s.running) == 1 {
		close(s.stopCh)
	}
	s.mu.Unlock()
}

// IsRunning reports whether the scheduler loop is currently active.
func (s *Scheduler) IsRunning() bool {
	return atomic.LoadInt32(&s.running) == 1
}

// Shutdown stops the scheduler loop and blocks until the loop goroutine has
// exited or ctx expires. After Shutdown returns the Scheduler cannot be
// restarted.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//	if err := s.Shutdown(ctx); err != nil {
//	    log.Printf("shutdown timed out: %v", err)
//	}
func (s *Scheduler) Shutdown(ctx context.Context) error {
	atomic.StoreInt32(&s.stopped, 1)
	s.mu.Lock()
	stopCh := s.stopCh
	doneCh := s.doneCh
	isRunning := atomic.LoadInt32(&s.running) == 1
	s.mu.Unlock()

	if isRunning {
		// Signal the loop to exit.
		select {
		case <-stopCh:
			// already closed — nothing to do
		default:
			close(stopCh)
		}
		// Wait for the loop to acknowledge.
		select {
		case <-doneCh:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

// Register adds a new job to the scheduler. The job fires according to the
// supplied cron expression and calls fn each time it is due.
//
// Register is safe to call both before and after Start. Jobs registered after
// Start will begin firing at their next scheduled time.
//
// It returns the job's ID (which may be supplied via WithJobID) or an error
// if the expression is invalid or the scheduler has been shut down.
//
// Example:
//
//	id, err := s.Register("0 * * * *", func(ctx context.Context) error {
//	    fmt.Println("every hour")
//	    return nil
//	}, crontask.WithJobName("Hourly ping"))
func (s *Scheduler) Register(expr string, fn JobFunc, opts ...JobOption) (string, error) {
	if atomic.LoadInt32(&s.stopped) == 1 {
		return "", ErrSchedulerStopped
	}
	if fn == nil {
		return "", fmt.Errorf("crontask: job function must not be nil")
	}

	// Apply job options.
	cfg := jobConfig{
		ctx: context.Background(),
	}
	for _, o := range opts {
		o(&cfg)
	}

	// Fall back to the scheduler-level default hooks when the job does not
	// supply its own. Per-job hooks always take precedence.
	if cfg.hooks == nil && s.cfg.defaultHooks != nil {
		cfg.hooks = s.cfg.defaultHooks
	}

	// Assign an ID if not provided.
	if cfg.id == "" {
		id, err := randn.UUID()
		if err != nil {
			// Fall back to a time-based ID on UUID generation failure.
			id = randn.TimeID()
		}
		cfg.id = id
	}

	// Parse the expression.
	sched, err := Parse(expr)
	if err != nil {
		return "", err
	}

	e := &entry{
		id:         cfg.id,
		name:       cfg.name,
		expression: expr,
		schedule:   sched,
		fn:         fn,
		cfg:        cfg,
	}
	e.nextRun = sched.Next(s.now())

	s.registry.add(e)
	return e.id, nil
}

// Remove unregisters the job with the given id. It returns ErrJobNotFound
// when the id is not present in the registry.
func (s *Scheduler) Remove(id string) error {
	return s.registry.remove(id)
}

// Jobs returns a snapshot of all registered jobs in an unspecified order.
// Each element is an immutable JobInfo that reflects the state at the time
// of the call.
func (s *Scheduler) Jobs() []JobInfo {
	entries := s.registry.list()
	out := make([]JobInfo, len(entries))
	for i, e := range entries {
		out[i] = e.snapshot()
	}
	return out
}

// NextRuns returns the next n activation times for the job identified by id,
// starting from the given reference time t. It returns ErrJobNotFound when
// the id is not registered.
//
// Example:
//
//	runs, err := s.NextRuns("my-job-id", time.Now(), 5)
func (s *Scheduler) NextRuns(id string, t time.Time, n int) ([]time.Time, error) {
	e, ok := s.registry.get(id)
	if !ok {
		return nil, ErrJobNotFound
	}
	if n <= 0 {
		return nil, nil
	}
	out := make([]time.Time, 0, n)
	cur := t
	for len(out) < n {
		next := e.schedule.Next(cur)
		if next.IsZero() {
			break
		}
		out = append(out, next)
		cur = next
	}
	return out, nil
}

// Location returns the timezone that this scheduler uses to compute job
// next-run times. The value is the location supplied via WithLocation; it
// defaults to time.UTC when no location option is provided.
func (s *Scheduler) Location() *time.Location {
	return s.cfg.loc
}

// WithSecondsEnabled reports whether the Scheduler was constructed with the
// WithSeconds option, meaning it accepts six-field cron expressions and ticks
// at one-second granularity.
func (s *Scheduler) WithSecondsEnabled() bool {
	return s.cfg.withSecs
}

// now returns the current time in the scheduler's configured timezone.
func (s *Scheduler) now() time.Time {
	return time.Now().In(s.cfg.loc)
}

// loop is the main scheduler goroutine. It sleeps until the next due job(s),
// dispatches them via the executor, then updates their next-run times.
func (s *Scheduler) loop(stop <-chan struct{}, done chan<- struct{}) {
	defer func() {
		atomic.StoreInt32(&s.running, 0)
		close(done)
	}()

	// Compute initial next-run times for all already-registered entries.
	for _, e := range s.registry.list() {
		updateNextRun(e, s.now())
	}

	for {
		now := s.now()
		due, _ := s.registry.nextDue(now)

		for _, e := range due {
			s.exec.dispatch(e, now)
			updateNextRun(e, now)
		}

		// Recompute sleep duration after updating next-run times for dispatched
		// entries. This ensures the loop does not sleep for maxSleep when the
		// only registered entries were just dispatched and have a fresh nextRun.
		_, sleep := s.registry.nextDue(s.now())

		select {
		case <-stop:
			return
		case <-time.After(sleep):
		}
	}
}
