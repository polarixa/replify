package workergroup

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// Job is a unit of work submitted to a WorkerPool. It receives the pool's
// Context, which is canceled when the pool's parent Context is canceled.
// A non-nil return value is reported to the pool exactly as described by
// (*WorkerPool).Errors.
type Job func(ctx context.Context) error

// ErrPoolClosed is returned by Submit and TrySubmit, and by ScaleTo and
// ScaleBy, once Close has been called on the WorkerPool.
var ErrPoolClosed = errors.New("workergroup: worker pool is closed")

// PoolOption configures a WorkerPool at construction time.
type PoolOption func(*poolConfig)

// poolConfig collects the settings applied by PoolOption values before a
// WorkerPool is built, so that NewWorkerPool can assemble the pool in one
// pass.
type poolConfig struct {
	queueSize int
	groupOpts []Option
}

// WithQueueSize returns a PoolOption that sets the capacity of the pool's
// internal job queue to n. With the default of 0, Submit blocks until a
// worker is ready to receive the job directly (an unbuffered handoff).
// A larger n lets callers get ahead of the workers by up to n pending
// jobs before Submit starts blocking.
func WithQueueSize(n int) PoolOption {
	return func(c *poolConfig) {
		c.queueSize = n
	}
}

// WithPoolErrorMode returns a PoolOption that selects how the pool
// combines errors returned by jobs; see ErrorMode. The default is
// CollectErrors, which is almost always what a persistent pool wants: one
// failing job should not stop the worker that ran it, let alone the rest
// of the pool. Pass FirstError to instead cancel the whole pool's Context
// the first time any job fails.
func WithPoolErrorMode(mode ErrorMode) PoolOption {
	return func(c *poolConfig) {
		c.groupOpts = append(c.groupOpts, WithErrorMode(mode))
	}
}

// WithPoolPanicRecovery returns a PoolOption that recovers panics raised
// by individual jobs so that one misbehaving job cannot crash the process
// or silently kill off a worker goroutine. See WithPanicRecovery for the
// semantics of handler. Recommended for most production pools, since an
// unrecovered panic in one job would otherwise terminate the worker
// goroutine that ran it, permanently shrinking the pool by one.
func WithPoolPanicRecovery(handler func(recovered any, stack []byte)) PoolOption {
	return func(c *poolConfig) {
		c.groupOpts = append(c.groupOpts, WithPanicRecovery(handler))
	}
}

// WorkerPool is a dynamically scalable pool of persistent worker
// goroutines that consume Jobs from an internal queue.
//
// Unlike Group, whose Go method starts exactly one goroutine per call,
// WorkerPool maintains a standing set of worker goroutines, each running a
// loop that repeatedly pulls a Job from the queue and executes it. The
// number of workers can be grown or shrunk at any time with ScaleTo or
// ScaleBy, which is the intended way to react to changing load — for
// example, scaling up when the queue is deep and scaling back down once
// it drains.
//
// WorkerPool is built directly on top of Group: each worker's loop is one
// long-lived task passed to Group.Go, and job errors are reported to the
// Group through Group.Fail so that a failing job never has to end the
// worker's loop. This is why the pool defaults to CollectErrors mode (see
// WithPoolErrorMode).
//
// A WorkerPool must be created with NewWorkerPool; its zero value is not
// usable. All methods are safe for concurrent use by multiple goroutines.
type WorkerPool struct {
	ctx   context.Context
	group *Group
	jobs  chan Job

	mu      sync.Mutex
	workers map[uint64]chan struct{} // worker id -> its stop signal
	nextID  uint64
	closed  bool

	closeOnce sync.Once
}

// NewWorkerPool creates a WorkerPool whose jobs observe ctx, starts it
// with initialWorkers worker goroutines (which may be 0), and returns it.
// A negative initialWorkers is treated as 0.
//
// If ctx is later canceled, every worker goroutine exits once it finishes
// (or is between) jobs; the pool itself is not otherwise tied to ctx's
// lifetime.
func NewWorkerPool(ctx context.Context, initialWorkers int, opts ...PoolOption) *WorkerPool {
	cfg := &poolConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	if len(cfg.groupOpts) == 0 {
		cfg.groupOpts = []Option{WithErrorMode(CollectErrors)}
	}

	p := &WorkerPool{
		ctx:     ctx,
		group:   New(cfg.groupOpts...),
		jobs:    make(chan Job, cfg.queueSize),
		workers: make(map[uint64]chan struct{}),
	}

	if initialWorkers < 0 {
		initialWorkers = 0
	}
	p.mu.Lock()
	for i := 0; i < initialWorkers; i++ {
		p.spawnWorkerLocked()
	}
	p.mu.Unlock()

	return p
}

// spawnWorkerLocked starts one new worker goroutine and registers its
// stop channel. The caller must hold p.mu.
func (p *WorkerPool) spawnWorkerLocked() {
	id := p.nextID
	p.nextID++
	stop := make(chan struct{})
	p.workers[id] = stop

	p.group.Go(func() error {
		for {
			select {
			case <-stop:
				return nil
			case <-p.ctx.Done():
				return nil
			case job, ok := <-p.jobs:
				if !ok {
					return nil
				}
				if err := job(p.ctx); err != nil {
					p.group.Fail(fmt.Errorf("workergroup: job failed: %w", err))
				}
			}
		}
	})
}

// Submit adds job to the queue, blocking until a worker is ready to
// accept it (or, if WithQueueSize gave the queue spare capacity, until
// there is room in the queue), ctx is canceled, the pool's own Context is
// canceled, or the pool is closed.
//
// Submit returns ctx.Err(), the pool Context's error, or ErrPoolClosed in
// those respective cases, and nil once the job has been accepted.
// Accepting a job only means it has been queued or handed to a worker,
// not that it has finished or even started running.
func (p *WorkerPool) Submit(ctx context.Context, job Job) error {
	p.mu.Lock()
	closed := p.closed
	p.mu.Unlock()
	if closed {
		return ErrPoolClosed
	}
	select {
	case p.jobs <- job:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-p.ctx.Done():
		return p.ctx.Err()
	}
}

// TrySubmit adds job to the queue only if doing so would not block, i.e.
// only if a worker or free queue slot is immediately available. It
// reports whether job was accepted; it returns false without blocking if
// the pool is closed.
func (p *WorkerPool) TrySubmit(job Job) bool {
	p.mu.Lock()
	closed := p.closed
	p.mu.Unlock()
	if closed {
		return false
	}
	select {
	case p.jobs <- job:
		return true
	default:
		return false
	}
}

// ScaleTo changes the number of active worker goroutines to exactly n,
// starting new workers or stopping existing ones as needed, and returns
// once the change has been requested. A negative n is rejected with an
// error; ErrPoolClosed is returned if Close has already been called.
//
// Workers that are stopped because n is lower than the current worker
// count finish any job they are currently running before exiting; they do
// not abandon in-progress work. Scaling up or down does not affect jobs
// already queued — it only changes how many workers are available to
// pull from that queue.
func (p *WorkerPool) ScaleTo(n int) error {
	if n < 0 {
		return fmt.Errorf("workergroup: ScaleTo: negative worker count %d", n)
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return ErrPoolClosed
	}

	current := len(p.workers)
	switch {
	case n > current:
		for i := 0; i < n-current; i++ {
			p.spawnWorkerLocked()
		}
	case n < current:
		toStop := current - n
		for id, stop := range p.workers {
			if toStop == 0 {
				break
			}
			close(stop)
			delete(p.workers, id)
			toStop--
		}
	}
	return nil
}

// ScaleBy adjusts the number of active worker goroutines by delta
// (positive to scale up, negative to scale down) and returns the
// resulting target worker count along with any error from the underlying
// ScaleTo call. The resulting worker count is never allowed to go below 0.
func (p *WorkerPool) ScaleBy(delta int) (int, error) {
	p.mu.Lock()
	target := len(p.workers) + delta
	p.mu.Unlock()
	if target < 0 {
		target = 0
	}
	return target, p.ScaleTo(target)
}

// Workers reports the current number of active worker goroutines.
func (p *WorkerPool) Workers() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.workers)
}

// Close stops the pool from accepting any further jobs via Submit or
// TrySubmit (both return ErrPoolClosed afterwards) and closes the
// internal queue so that each worker exits once it has drained any jobs
// still queued ahead of it. Close does not wait for workers to finish;
// call Wait afterwards to do that. Close is idempotent and safe to call
// more than once or concurrently with other methods.
func (p *WorkerPool) Close() {
	p.closeOnce.Do(func() {
		p.mu.Lock()
		p.closed = true
		p.mu.Unlock()
		close(p.jobs)
	})
}

// Wait blocks until every worker goroutine has exited, then returns the
// pool's combined error according to its ErrorMode (CollectErrors by
// default; see WithPoolErrorMode). Workers exit when the pool's Context is
// canceled, when they are individually stopped by ScaleTo/ScaleBy, or
// (for workers still running) once Close has been called and the queue
// has drained.
//
// Calling Wait on a pool that is never closed, scaled to zero, or whose
// Context is never canceled will block forever, since its workers loop
// indefinitely waiting for jobs. The typical shutdown sequence is to call
// Close and then Wait.
func (p *WorkerPool) Wait() error {
	return p.group.Wait()
}

// Err returns the pool's current combined error without blocking,
// reflecting whatever jobs have failed so far. See (*Group).Err.
func (p *WorkerPool) Err() error {
	return p.group.Err()
}

// Errors returns every error reported by failed jobs so far, without
// blocking. See (*Group).Errors for the exact semantics, which depend on
// the pool's ErrorMode.
func (p *WorkerPool) Errors() []error {
	return p.group.Errors()
}
