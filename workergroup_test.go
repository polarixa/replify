package replify_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/polarixa/replify"
	"github.com/polarixa/replify/pkg/workergroup"
)

// resultsOf extracts the []*replify.WorkerResult body from w, failing the
// test if the body is absent or of an unexpected type.
func resultsOf(t *testing.T, w interface {
	Body() any
}) []*replify.WorkerResult {
	t.Helper()
	results, ok := w.Body().([]*replify.WorkerResult)
	if !ok {
		t.Fatalf("expected body of type []*replify.WorkerResult, got %T", w.Body())
	}
	return results
}

// --- RunWorkerGroup: happy path ---------------------------------------------

func TestRunWorkerGroup_SingleTaskSuccess(t *testing.T) {
	t.Parallel()

	w := replify.RunWorkerGroup(context.Background(), 0, []*replify.WorkerTask{
		replify.NewWorkerTask("only", func(ctx context.Context) (any, error) { return 42, nil }),
	})

	if w.StatusCode() != replify.StatusOK.Value() {
		t.Fatalf("expected 200, got %d", w.StatusCode())
	}
	if w.IsError() {
		t.Fatalf("expected no error, got %v", w.Error())
	}
	results := resultsOf(t, w)
	if len(results) != 1 || !results[0].Success() || results[0].Value() != 42 {
		t.Fatalf("unexpected results: %+v", results)
	}
}

func TestRunWorkerGroup_MultipleTasksSuccess(t *testing.T) {
	t.Parallel()

	var tasks []*replify.WorkerTask
	for i := 0; i < 5; i++ {
		i := i
		tasks = append(tasks, replify.NewWorkerTask(
			fmt.Sprintf("task-%d", i),
			func(ctx context.Context) (any, error) { return i * i, nil },
		))
	}

	w := replify.RunWorkerGroup(context.Background(), 3, tasks)

	if w.StatusCode() != replify.StatusOK.Value() {
		t.Fatalf("expected 200, got %d", w.StatusCode())
	}
	if w.Total() != 5 {
		t.Fatalf("expected total 5, got %d", w.Total())
	}
	results := resultsOf(t, w)
	sum := 0
	for _, r := range results {
		if !r.Success() {
			t.Fatalf("unexpected failure: %+v", r)
		}
		sum += r.Value().(int)
	}
	if sum != 0+1+4+9+16 {
		t.Fatalf("expected sum 30, got %d", sum)
	}
}

func TestRunWorkerGroup_Empty(t *testing.T) {
	t.Parallel()

	w := replify.RunWorkerGroup(context.Background(), 0, nil)

	if w.StatusCode() != replify.StatusNoContent.Value() {
		t.Fatalf("expected 204, got %d", w.StatusCode())
	}
	if w.Total() != 0 {
		t.Fatalf("expected total 0, got %d", w.Total())
	}
	if w.IsError() {
		t.Fatalf("expected no error for empty group, got %v", w.Error())
	}
}

// --- RunWorkerGroup: error handling ------------------------------------------

var errBoom = errors.New("boom")

func TestRunWorkerGroup_SingleTaskFailure(t *testing.T) {
	t.Parallel()

	w := replify.RunWorkerGroup(context.Background(), 0, []*replify.WorkerTask{
		replify.NewWorkerTask("failing", func(ctx context.Context) (any, error) { return nil, errBoom }),
	})

	if w.StatusCode() != replify.StatusInternalServerError.Value() {
		t.Fatalf("expected 500, got %d", w.StatusCode())
	}
	if !w.IsError() {
		t.Fatalf("expected an error to be present")
	}
	if !errors.Is(w.Cause(), errBoom) {
		t.Fatalf("expected errors.Is to find errBoom in %v", w.Cause())
	}
	if w.Reason().Code() != replify.ReasonCodeJobFailed {
		t.Fatalf("expected reason code %v, got %v", replify.ReasonCodeJobFailed, w.Reason().Code())
	}
}

func TestRunWorkerGroup_AllTasksFail(t *testing.T) {
	t.Parallel()

	errA := errors.New("task a failed")
	errB := errors.New("task b failed")

	w := replify.RunWorkerGroup(context.Background(), 0, []*replify.WorkerTask{
		replify.NewWorkerTask("a", func(ctx context.Context) (any, error) { return nil, errA }),
		replify.NewWorkerTask("b", func(ctx context.Context) (any, error) { return nil, errB }),
	})

	if w.StatusCode() != replify.StatusInternalServerError.Value() {
		t.Fatalf("expected 500, got %d", w.StatusCode())
	}
	if !errors.Is(w.Cause(), errA) || !errors.Is(w.Cause(), errB) {
		t.Fatalf("expected errors.Is to find both errA and errB in %v", w.Cause())
	}

	results := resultsOf(t, w)
	if len(results) != 2 || results[0].Success() || results[1].Success() {
		t.Fatalf("expected both results to be failures: %+v", results)
	}
}

func TestRunWorkerGroup_PartialFailure(t *testing.T) {
	t.Parallel()

	w := replify.RunWorkerGroup(context.Background(), 0, []*replify.WorkerTask{
		replify.NewWorkerTask("ok", func(ctx context.Context) (any, error) { return "done", nil }),
		replify.NewWorkerTask("bad", func(ctx context.Context) (any, error) { return nil, errBoom }),
	})

	if w.StatusCode() != replify.StatusMultiStatus.Value() {
		t.Fatalf("expected 207, got %d", w.StatusCode())
	}
	// A 207 Multi-Status outcome is not itself a top-level error.
	if w.IsErrorPresent() {
		t.Fatalf("expected no top-level error for a partial failure, got %v", w.Error())
	}

	results := resultsOf(t, w)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	var sawSuccess, sawFailure bool
	for _, r := range results {
		switch r.Name() {
		case "ok":
			sawSuccess = r.Success()
		case "bad":
			sawFailure = !r.Success() && errors.Is(r.Err(), errBoom)
		}
	}
	if !sawSuccess || !sawFailure {
		t.Fatalf("expected one success and one errBoom failure, got %+v", results)
	}

	if debug := w.Debugging(); debug["succeeded"] != 1 || debug["failed"] != 1 {
		t.Fatalf("unexpected debug counters: %+v", debug)
	}
}

func TestRunWorkerGroup_ErrorsAs(t *testing.T) {
	t.Parallel()

	var myErr *customError
	w := replify.RunWorkerGroup(context.Background(), 0, []*replify.WorkerTask{
		replify.NewWorkerTask("typed", func(ctx context.Context) (any, error) { return nil, &customError{msg: "typed failure"} }),
	})

	if !errors.As(w.Cause(), &myErr) {
		t.Fatalf("expected errors.As to find *customError in %v", w.Cause())
	}
	if myErr.msg != "typed failure" {
		t.Fatalf("unexpected message: %s", myErr.msg)
	}
}

type customError struct{ msg string }

func (e *customError) Error() string { return e.msg }

// --- Context: cancellation and deadlines ------------------------------------

func TestRunWorkerGroup_ContextCanceledBeforeExecution(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before any task runs

	w := replify.RunWorkerGroup(ctx, 0, []*replify.WorkerTask{
		replify.NewWorkerTask("irrelevant", func(ctx context.Context) (any, error) { return nil, nil }),
	})

	if w.StatusCode() != replify.StatusClientClosedRequest.Value() {
		t.Fatalf("expected 499, got %d", w.StatusCode())
	}
	if !errors.Is(w.Cause(), context.Canceled) {
		t.Fatalf("expected errors.Is to find context.Canceled in %v", w.Cause())
	}
	if w.Reason().Code() != replify.ReasonCodeJobCancelled {
		t.Fatalf("expected reason code %v, got %v", replify.ReasonCodeJobCancelled, w.Reason().Code())
	}
}

func TestRunWorkerGroup_ContextCanceledDuringExecution(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	g, gctx := replify.NewWorkerGroupWithContext(ctx)

	started := make(chan struct{})
	g.Go("waits-for-cancel", func(ctx context.Context) (any, error) {
		close(started)
		<-gctx.Done()
		return nil, ctx.Err()
	})

	<-started
	cancel()

	w := g.Wait()
	if w.StatusCode() != replify.StatusClientClosedRequest.Value() {
		t.Fatalf("expected 499, got %d", w.StatusCode())
	}
}

func TestRunWorkerGroup_DeadlineExceeded(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	g, gctx := replify.NewWorkerGroupWithContext(ctx)
	g.Go("slow", func(ctx context.Context) (any, error) {
		<-gctx.Done()
		return nil, ctx.Err()
	})

	w := g.Wait()
	if w.StatusCode() != replify.StatusGatewayTimeout.Value() {
		t.Fatalf("expected 504, got %d", w.StatusCode())
	}
	if !errors.Is(w.Cause(), context.DeadlineExceeded) {
		t.Fatalf("expected errors.Is to find context.DeadlineExceeded in %v", w.Cause())
	}
	if w.Reason().Code() != replify.ReasonCodeJobTimeout {
		t.Fatalf("expected reason code %v, got %v", replify.ReasonCodeJobTimeout, w.Reason().Code())
	}
}

// --- Concurrency -------------------------------------------------------------

func TestRunWorkerGroup_ConcurrentAggregation(t *testing.T) {
	t.Parallel()

	const n = 50
	var tasks []*replify.WorkerTask
	for i := 0; i < n; i++ {
		i := i
		tasks = append(tasks, replify.NewWorkerTask("", func(ctx context.Context) (any, error) { return i, nil }))
	}

	w := replify.RunWorkerGroup(context.Background(), 8, tasks)
	results := resultsOf(t, w)
	if len(results) != n {
		t.Fatalf("expected %d results, got %d", n, len(results))
	}
	seen := make(map[int]bool, n)
	for _, r := range results {
		seen[r.Value().(int)] = true
	}
	if len(seen) != n {
		t.Fatalf("expected all %d distinct values, got %d", n, len(seen))
	}
}

func TestWorkerGroup_SequentialOrderMatchesCallOrder(t *testing.T) {
	t.Parallel()

	g, _ := replify.NewWorkerGroupWithContext(context.Background(), workergroup.WithLimit(1))
	for i := 0; i < 5; i++ {
		i := i
		g.Go(fmt.Sprintf("seq-%d", i), func(ctx context.Context) (any, error) { return i, nil })
	}
	w := g.Wait()
	results := resultsOf(t, w)
	for i, r := range results {
		if r.Index() != i || r.Value() != i {
			t.Fatalf("expected sequential index/value %d, got index=%d value=%v", i, r.Index(), r.Value())
		}
	}
}

// --- Edge cases ----------------------------------------------------------------

func TestRunWorkerGroup_NilWorkerFunc(t *testing.T) {
	t.Parallel()

	w := replify.RunWorkerGroup(context.Background(), 0, []*replify.WorkerTask{
		replify.NewWorkerTask("nil-fn", nil),
	})

	if w.StatusCode() != replify.StatusInternalServerError.Value() {
		t.Fatalf("expected 500, got %d", w.StatusCode())
	}
	if !errors.Is(w.Cause(), replify.ErrNilTask) {
		t.Fatalf("expected errors.Is to find ErrNilTask in %v", w.Cause())
	}
}

func TestWorkerGroup_DefaultNames(t *testing.T) {
	t.Parallel()

	g, _ := replify.NewWorkerGroupWithContext(context.Background())
	g.Go("", func(ctx context.Context) (any, error) { return nil, nil })
	w := g.Wait()
	results := resultsOf(t, w)
	if results[0].Name() != "task-0" {
		t.Fatalf("expected default name task-0, got %q", results[0].Name())
	}
}

func TestWorkerGroup_TryGoRespectsLimit(t *testing.T) {
	t.Parallel()

	g := replify.NewWorkerGroup(workergroup.WithLimit(1))
	block := make(chan struct{})
	started := make(chan struct{})
	g.Go("blocker", func(ctx context.Context) (any, error) {
		close(started)
		<-block
		return nil, nil
	})
	<-started

	if g.TryGo("rejected", func(ctx context.Context) (any, error) { return nil, nil }) {
		t.Fatalf("expected TryGo to be rejected while at the concurrency limit")
	}
	close(block)
	w := g.Wait()
	if w.Total() != 1 {
		t.Fatalf("expected only the started task to be recorded, got total=%d", w.Total())
	}
}

// --- WorkerTask / PoolJob / WorkerResult accessors --------------------------

func TestWorkerTask_GettersAndSetters(t *testing.T) {
	t.Parallel()

	fn := func(ctx context.Context) (any, error) { return nil, nil }
	task := replify.NewWorkerTask("initial", fn)
	if task.Name() != "initial" {
		t.Fatalf("expected name %q, got %q", "initial", task.Name())
	}
	task.WithName("renamed")
	if task.Name() != "renamed" {
		t.Fatalf("expected name %q, got %q", "renamed", task.Name())
	}
	if task.Fn() == nil {
		t.Fatalf("expected non-nil Fn")
	}

	var nilTask *replify.WorkerTask
	if nilTask.Available() {
		t.Fatalf("expected nil *WorkerTask to be unavailable")
	}
	if nilTask.Name() != "" || nilTask.Fn() != nil {
		t.Fatalf("expected zero values from a nil *WorkerTask")
	}
}

func TestPoolJob_GettersAndSetters(t *testing.T) {
	t.Parallel()

	job := replify.NewPoolJob("initial", func(ctx context.Context) error { return nil })
	if job.Name() != "initial" {
		t.Fatalf("expected name %q, got %q", "initial", job.Name())
	}
	job.WithName("renamed").WithJob(func(ctx context.Context) error { return errBoom })
	if job.Name() != "renamed" {
		t.Fatalf("expected name %q, got %q", "renamed", job.Name())
	}
	if err := job.Job()(context.Background()); !errors.Is(err, errBoom) {
		t.Fatalf("expected errBoom from replaced job, got %v", err)
	}

	var nilJob *replify.PoolJob
	if nilJob.Available() {
		t.Fatalf("expected nil *PoolJob to be unavailable")
	}
}

// --- Pool: happy path ---------------------------------------------------------

func TestRunWorkerPool_Success(t *testing.T) {
	t.Parallel()

	sum := make(chan int, 4)
	var jobs []*replify.PoolJob
	for i := 1; i <= 4; i++ {
		i := i
		jobs = append(jobs, replify.NewPoolJob(
			fmt.Sprintf("job-%d", i),
			func(ctx context.Context) error { sum <- i; return nil },
		))
	}

	w := replify.RunWorkerPool(context.Background(), 2, jobs)
	close(sum)

	if w.StatusCode() != replify.StatusOK.Value() {
		t.Fatalf("expected 200, got %d", w.StatusCode())
	}
	total := 0
	for v := range sum {
		total += v
	}
	if total != 10 {
		t.Fatalf("expected sum 10, got %d", total)
	}
}

func TestRunWorkerPool_Empty(t *testing.T) {
	t.Parallel()

	w := replify.RunWorkerPool(context.Background(), 2, nil)
	if w.StatusCode() != replify.StatusNoContent.Value() {
		t.Fatalf("expected 204, got %d", w.StatusCode())
	}
}

func TestRunWorkerPool_PartialFailure(t *testing.T) {
	t.Parallel()

	w := replify.RunWorkerPool(context.Background(), 2, []*replify.PoolJob{
		replify.NewPoolJob("ok", func(ctx context.Context) error { return nil }),
		replify.NewPoolJob("bad", func(ctx context.Context) error { return errBoom }),
	})

	if w.StatusCode() != replify.StatusMultiStatus.Value() {
		t.Fatalf("expected 207, got %d", w.StatusCode())
	}
	results := resultsOf(t, w)
	var failures int
	for _, r := range results {
		if !r.Success() {
			failures++
			if !errors.Is(r.Err(), errBoom) {
				t.Fatalf("expected errBoom, got %v", r.Err())
			}
		}
	}
	if failures != 1 {
		t.Fatalf("expected exactly 1 failure, got %d", failures)
	}
}

func TestRunWorkerPool_AllFail(t *testing.T) {
	t.Parallel()

	w := replify.RunWorkerPool(context.Background(), 2, []*replify.PoolJob{
		replify.NewPoolJob("a", func(ctx context.Context) error { return errBoom }),
		replify.NewPoolJob("b", func(ctx context.Context) error { return errBoom }),
	})

	if w.StatusCode() != replify.StatusInternalServerError.Value() {
		t.Fatalf("expected 500, got %d", w.StatusCode())
	}
	if !errors.Is(w.Cause(), errBoom) {
		t.Fatalf("expected errors.Is to find errBoom in %v", w.Cause())
	}
}

func TestRunWorkerPool_NilJob(t *testing.T) {
	t.Parallel()

	w := replify.RunWorkerPool(context.Background(), 1, []*replify.PoolJob{
		replify.NewPoolJob("nil-job", nil),
	})
	if !errors.Is(w.Cause(), replify.ErrNilTask) {
		t.Fatalf("expected errors.Is to find ErrNilTask in %v", w.Cause())
	}
}

func TestPool_ContextCanceled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	p := replify.NewPool(ctx, 1)
	cancel()
	p.Close()

	w := p.Wait()
	if w.StatusCode() != replify.StatusClientClosedRequest.Value() {
		t.Fatalf("expected 499, got %d", w.StatusCode())
	}
}

func TestPool_SubmitFailureIsRecorded(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // the pool's own ctx is already canceled

	w := replify.RunWorkerPool(ctx, 1, []*replify.PoolJob{
		replify.NewPoolJob("never-runs", func(ctx context.Context) error { return nil }),
	})

	// Submission itself fails because ctx is canceled; it must still show
	// up as a tracked (failed) result rather than being silently dropped.
	if w.Total() == 0 {
		t.Fatalf("expected the submission failure to be recorded, got total=0")
	}
}

// --- Example -------------------------------------------------------------------

// ExampleRunWorkerGroup demonstrates running a batch of named tasks
// concurrently through replify and inspecting the aggregated result.
func ExampleRunWorkerGroup() {
	tasks := []*replify.WorkerTask{
		replify.NewWorkerTask("square-1", func(ctx context.Context) (any, error) { return 1 * 1, nil }),
		replify.NewWorkerTask("square-2", func(ctx context.Context) (any, error) { return 2 * 2, nil }),
		replify.NewWorkerTask("square-3", func(ctx context.Context) (any, error) { return 3 * 3, nil }),
	}

	w := replify.RunWorkerGroup(context.Background(), 2, tasks)

	sum := 0
	for _, r := range w.Body().([]*replify.WorkerResult) {
		sum += r.Value().(int)
	}
	fmt.Println(w.StatusCode(), sum)
	// Output: 200 14
}
