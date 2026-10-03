package workergroup

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestGroupBasic(t *testing.T) {
	g := New(WithLimit(2))
	var n int32
	for i := 0; i < 10; i++ {
		g.Go(func() error {
			atomic.AddInt32(&n, 1)
			time.Sleep(time.Millisecond)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 10 {
		t.Fatalf("want 10 ran, got %d", n)
	}
}

func TestGroupFirstErrorCancelsContext(t *testing.T) {
	g, ctx := NewWithContext(context.Background())
	boom := errors.New("boom")
	g.Go(func() error { return boom })
	g.Go(func() error {
		<-ctx.Done()
		return ctx.Err()
	})
	err := g.Wait()
	if !errors.Is(err, boom) {
		t.Fatalf("want boom, got %v", err)
	}
}

func TestGroupCollectErrors(t *testing.T) {
	g := New(WithErrorMode(CollectErrors))
	e1 := errors.New("e1")
	e2 := errors.New("e2")
	g.Go(func() error { return e1 })
	g.Go(func() error { return e2 })
	g.Go(func() error { return nil })
	err := g.Wait()
	if !errors.Is(err, e1) || !errors.Is(err, e2) {
		t.Fatalf("want both errors joined, got %v", err)
	}
	if len(g.Errors()) != 2 {
		t.Fatalf("want 2 collected errors, got %d", len(g.Errors()))
	}
}

func TestGroupDynamicSetLimitWhileActive(t *testing.T) {
	g := New(WithLimit(1))
	started := make(chan struct{}, 10)
	release := make(chan struct{})

	// Start one long-running task that holds the only permit.
	g.Go(func() error {
		started <- struct{}{}
		<-release
		return nil
	})
	<-started

	// Raise the limit while the first goroutine is still active; this
	// must not panic (upstream errgroup panics here).
	g.SetLimit(4)
	if got := g.Limit(); got != 4 {
		t.Fatalf("want limit 4, got %d", got)
	}

	done := make(chan struct{})
	for i := 0; i < 3; i++ {
		g.Go(func() error {
			started <- struct{}{}
			return nil
		})
	}
	go func() {
		for i := 0; i < 3; i++ {
			<-started
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for newly-permitted goroutines to start")
	}

	close(release)
	if err := g.Wait(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGroupScale(t *testing.T) {
	g := New(WithLimit(2))
	if n := g.Scale(3); n != 5 {
		t.Fatalf("want 5, got %d", n)
	}
	if n := g.Scale(-10); n != 0 {
		t.Fatalf("want floor at 0, got %d", n)
	}
}

func TestGroupTryGoRespectsLimit(t *testing.T) {
	g := New(WithLimit(1))
	block := make(chan struct{})
	ok := g.TryGo(func() error { <-block; return nil })
	if !ok {
		t.Fatal("first TryGo should have succeeded")
	}
	if g.TryGo(func() error { return nil }) {
		t.Fatal("second TryGo should have been rejected at limit 1")
	}
	close(block)
	g.Wait()
}

func TestGroupPanicRecovery(t *testing.T) {
	var recovered any
	g := New(WithErrorMode(CollectErrors), WithPanicRecovery(func(r any, stack []byte) {
		recovered = r
	}))
	g.Go(func() error { panic("kaboom") })
	err := g.Wait()
	if !errors.Is(err, ErrPanic) {
		t.Fatalf("want ErrPanic, got %v", err)
	}
	if recovered != "kaboom" {
		t.Fatalf("want recovered value kaboom, got %v", recovered)
	}
}

func TestZeroGroupIsUsable(t *testing.T) {
	var g Group
	g.Go(func() error { return nil })
	if err := g.Wait(); err != nil {
		t.Fatalf("unexpected error from zero Group: %v", err)
	}
}

func TestWorkerPoolBasic(t *testing.T) {
	ctx := context.Background()
	pool := NewWorkerPool(ctx, 3)
	var n int32
	for i := 0; i < 50; i++ {
		if err := pool.Submit(ctx, func(ctx context.Context) error {
			atomic.AddInt32(&n, 1)
			return nil
		}); err != nil {
			t.Fatalf("submit failed: %v", err)
		}
	}
	pool.Close()
	if err := pool.Wait(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 50 {
		t.Fatalf("want 50 jobs run, got %d", n)
	}
}

func TestWorkerPoolScaleUpAndDown(t *testing.T) {
	ctx := context.Background()
	pool := NewWorkerPool(ctx, 2)
	if w := pool.Workers(); w != 2 {
		t.Fatalf("want 2 workers, got %d", w)
	}
	if err := pool.ScaleTo(8); err != nil {
		t.Fatalf("scale up failed: %v", err)
	}
	if w := pool.Workers(); w != 8 {
		t.Fatalf("want 8 workers, got %d", w)
	}
	if n, err := pool.ScaleBy(-5); err != nil || n != 3 {
		t.Fatalf("want 3 workers after ScaleBy(-5), got %d err=%v", n, err)
	}
	if w := pool.Workers(); w != 3 {
		t.Fatalf("want 3 workers, got %d", w)
	}

	pool.Close()
	if err := pool.Wait(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWorkerPoolCollectsJobErrorsWithoutStoppingWorkers(t *testing.T) {
	ctx := context.Background()
	pool := NewWorkerPool(ctx, 2) // default mode: CollectErrors
	boom := errors.New("boom")

	for i := 0; i < 5; i++ {
		i := i
		_ = pool.Submit(ctx, func(ctx context.Context) error {
			if i == 2 {
				return boom
			}
			return nil
		})
	}
	pool.Close()
	err := pool.Wait()
	if !errors.Is(err, boom) {
		t.Fatalf("want boom recorded, got %v", err)
	}
	if len(pool.Errors()) != 1 {
		t.Fatalf("want exactly 1 recorded error, got %d: %v", len(pool.Errors()), pool.Errors())
	}
}

func TestWorkerPoolClosedRejectsSubmit(t *testing.T) {
	ctx := context.Background()
	pool := NewWorkerPool(ctx, 1)
	pool.Close()
	if err := pool.Submit(ctx, func(context.Context) error { return nil }); !errors.Is(err, ErrPoolClosed) {
		t.Fatalf("want ErrPoolClosed, got %v", err)
	}
	if pool.TrySubmit(func(context.Context) error { return nil }) {
		t.Fatal("TrySubmit should fail on a closed pool")
	}
	if err := pool.ScaleTo(5); !errors.Is(err, ErrPoolClosed) {
		t.Fatalf("want ErrPoolClosed from ScaleTo, got %v", err)
	}
	pool.Wait()
}

func TestWorkerPoolContextCancellationStopsWorkers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	pool := NewWorkerPool(ctx, 4)
	cancel()
	if err := pool.Wait(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func ExampleGroup() {
	g, ctx := NewWithContext(context.Background(), WithLimit(2))
	results := make(chan int, 3)
	for i := 1; i <= 3; i++ {
		i := i
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				results <- i * i
				return nil
			}
		})
	}
	if err := g.Wait(); err != nil {
		fmt.Println("error:", err)
	}
	close(results)
	sum := 0
	for r := range results {
		sum += r
	}
	fmt.Println(sum)
	// Output: 14
}

func ExampleWorkerPool() {
	pool := NewWorkerPool(context.Background(), 2)
	total := make(chan int, 4)
	for i := 1; i <= 4; i++ {
		i := i
		_ = pool.Submit(context.Background(), func(ctx context.Context) error {
			total <- i
			return nil
		})
	}
	pool.Close()
	_ = pool.Wait()
	close(total)
	sum := 0
	for v := range total {
		sum += v
	}
	fmt.Println(sum)
	// Output: 10
}
