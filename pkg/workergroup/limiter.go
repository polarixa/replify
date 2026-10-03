package workergroup

import "sync"

// limiter is an internal counting semaphore whose capacity can be changed
// at any time, including while permits are held. It is the mechanism that
// lets Group and WorkerPool scale their concurrency limit up or down live,
// without the "must not be modified while goroutines are active" panic
// found in golang.org/x/sync/errgroup's channel-based semaphore.
//
// A negative limit means "unlimited": acquire and tryAcquire always
// succeed immediately. The zero value of limiter is unlimited.
type limiter struct {
	mu    sync.Mutex
	cond  *sync.Cond
	limit int // < 0 means unlimited
	inUse int
}

// newLimiter returns a limiter with the given initial capacity. A negative
// n means unlimited.
func newLimiter(n int) *limiter {
	l := &limiter{limit: n}
	l.cond = sync.NewCond(&l.mu)
	return l
}

// cond lazily initializes the sync.Cond for limiters created as a zero
// value (e.g. embedded in a zero-value Group) rather than via newLimiter.
func (l *limiter) condVar() *sync.Cond {
	// l.mu must already be held by the caller.
	if l.cond == nil {
		l.cond = sync.NewCond(&l.mu)
	}
	return l.cond
}

// acquire blocks until a permit is available and then takes it. If the
// limiter is unlimited (capacity < 0), acquire returns immediately.
func (l *limiter) acquire() {
	l.mu.Lock()
	defer l.mu.Unlock()
	cond := l.condVar()
	for l.limit >= 0 && l.inUse >= l.limit {
		cond.Wait()
	}
	l.inUse++
}

// tryAcquire takes a permit only if one is immediately available, without
// blocking. It reports whether a permit was taken.
func (l *limiter) tryAcquire() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.limit >= 0 && l.inUse >= l.limit {
		return false
	}
	l.inUse++
	return true
}

// release returns a permit to the limiter and wakes any goroutines
// blocked in acquire so they can re-check the (possibly just-changed)
// capacity.
func (l *limiter) release() {
	l.mu.Lock()
	l.inUse--
	cond := l.condVar()
	l.mu.Unlock()
	cond.Broadcast()
}

// setLimit changes the capacity to n. A negative n makes the limiter
// unlimited. This is safe to call at any time, including while permits
// are held; goroutines blocked in acquire are woken so they can observe
// the new capacity immediately.
func (l *limiter) setLimit(n int) {
	l.mu.Lock()
	l.limit = n
	cond := l.condVar()
	l.mu.Unlock()
	cond.Broadcast()
}

// adjust changes the capacity by delta relative to its current value and
// returns the resulting capacity. A limiter that is currently unlimited is
// treated as having a baseline capacity of 0 before delta is applied. The
// resulting capacity is never allowed to go below 0.
func (l *limiter) adjust(delta int) int {
	l.mu.Lock()
	if l.limit < 0 {
		l.limit = 0
	}
	l.limit += delta
	if l.limit < 0 {
		l.limit = 0
	}
	n := l.limit
	cond := l.condVar()
	l.mu.Unlock()
	cond.Broadcast()
	return n
}

// cap reports the current capacity. A negative result means unlimited.
func (l *limiter) cap() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.limit
}

// used reports the number of permits currently held.
func (l *limiter) used() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.inUse
}
