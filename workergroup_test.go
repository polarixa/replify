package replify_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
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
	if len(results) != 1 || !results[0].IsSuccess() || results[0].Value() != 42 {
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
		if !r.IsSuccess() {
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
	if len(results) != 2 || results[0].IsSuccess() || results[1].IsSuccess() {
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
			sawSuccess = r.IsSuccess()
		case "bad":
			sawFailure = !r.IsSuccess() && errors.Is(r.Cause(), errBoom)
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

// --- Respond / JSON / Logging / Slogging ------------------------------------

func TestWorkerTask_RespondAndJSON(t *testing.T) {
	t.Parallel()

	task := replify.NewWorkerTask("t1", func(ctx context.Context) (any, error) { return nil, nil })
	m := task.Respond()
	if m["name"] != "t1" || m["has_fn"] != true {
		t.Fatalf("unexpected Respond() map: %+v", m)
	}
	if task.JSON() == "" {
		t.Fatalf("expected non-empty JSON")
	}

	task.Logging()
	task.Slogging()

	var nilTask *replify.WorkerTask
	if len(nilTask.Respond()) != 0 {
		t.Fatalf("expected empty Respond() map for nil *WorkerTask")
	}
	nilTask.Logging() // must not panic
	nilTask.Slogging()
}

func TestPoolJob_RespondAndJSON(t *testing.T) {
	t.Parallel()

	job := replify.NewPoolJob("j1", func(ctx context.Context) error { return nil })
	m := job.Respond()
	if m["name"] != "j1" || m["has_job"] != true {
		t.Fatalf("unexpected Respond() map: %+v", m)
	}
	if job.JSON() == "" {
		t.Fatalf("expected non-empty JSON")
	}

	job.Logging()
	job.Slogging()

	var nilJob *replify.PoolJob
	if len(nilJob.Respond()) != 0 {
		t.Fatalf("expected empty Respond() map for nil *PoolJob")
	}
	nilJob.Logging() // must not panic
	nilJob.Slogging()
}

func TestWorkerResult_RespondAndJSON(t *testing.T) {
	t.Parallel()

	w := replify.RunWorkerGroup(context.Background(), 0, []*replify.WorkerTask{
		replify.NewWorkerTask("failing", func(ctx context.Context) (any, error) { return nil, errBoom }),
	})
	result := resultsOf(t, w)[0]

	m := result.Respond()
	if m["name"] != "failing" || m["success"] != false || m["error"] != errBoom.Error() {
		t.Fatalf("unexpected Respond() map: %+v", m)
	}
	if result.JSON() == "" {
		t.Fatalf("expected non-empty JSON")
	}

	result.Logging()
	result.Slogging()

	var nilResult *replify.WorkerResult
	if len(nilResult.Respond()) != 0 {
		t.Fatalf("expected empty Respond() map for nil *WorkerResult")
	}
	nilResult.Logging() // must not panic
	nilResult.Slogging()
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
		if !r.IsSuccess() {
			failures++
			if !errors.Is(r.Cause(), errBoom) {
				t.Fatalf("expected errBoom, got %v", r.Cause())
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

func TestPool_SubmitReply_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := replify.NewPool(ctx, 2)
	defer p.Close()

	w := p.SubmitReply(ctx, "job-1", func(ctx context.Context) error { return nil })
	if w.StatusCode() != replify.StatusAccepted.Value() {
		t.Fatalf("expected 202, got %d", w.StatusCode())
	}
	if w.IsError() {
		t.Fatalf("expected no error, got %v", w.Error())
	}
}

func TestPool_SubmitReply_ContextCanceled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel both the pool's own ctx and the per-call ctx before any worker can race for the job

	p := replify.NewPool(ctx, 1)
	w := p.SubmitReply(ctx, "job-1", func(ctx context.Context) error { return nil })

	if w.StatusCode() != replify.StatusClientClosedRequest.Value() {
		t.Fatalf("expected 499, got %d", w.StatusCode())
	}
	if !errors.Is(w.Cause(), context.Canceled) {
		t.Fatalf("expected errors.Is to find context.Canceled in %v", w.Cause())
	}
}

func TestPool_SubmitReply_PoolClosed(t *testing.T) {
	t.Parallel()

	p := replify.NewPool(context.Background(), 1)
	p.Close()

	w := p.SubmitReply(context.Background(), "job-1", func(ctx context.Context) error { return nil })
	if w.StatusCode() != replify.StatusServiceUnavailable.Value() {
		t.Fatalf("expected 503, got %d", w.StatusCode())
	}
	if !errors.Is(w.Cause(), workergroup.ErrPoolClosed) {
		t.Fatalf("expected errors.Is to find ErrPoolClosed in %v", w.Cause())
	}
}

// --- Pool: result accounting (reserve-on-accept) -----------------------------

// noopJob is a job that always succeeds.
func noopJob(ctx context.Context) error { return nil }

// notifyErrCtx wraps a Context and signals on every Err() call. Pool.Submit
// calls Err() exactly once per rejected submission (and never on an accepted
// one), which lets a test block a worker until a known number of submissions
// have been rejected, with no sleeping.
type notifyErrCtx struct {
	context.Context
	errCalled chan struct{}
}

func (c *notifyErrCtx) Err() error {
	select {
	case c.errCalled <- struct{}{}:
	default:
	}
	return c.Context.Err()
}

func TestPool_SubmitRejectedLeavesNoResult(t *testing.T) {
	t.Parallel()

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	tests := []struct {
		name    string
		submit  func(p *replify.Pool) error
		wantErr error
	}{
		{
			name: "T1 closed pool",
			submit: func(p *replify.Pool) error {
				p.Close()
				return p.Submit(context.Background(), "never-runs", noopJob)
			},
			wantErr: workergroup.ErrPoolClosed,
		},
		{
			// Zero workers and zero queue: no receiver exists, so the only
			// ready case in Submit's select is the canceled per-call context.
			name: "T3 canceled per-call context",
			submit: func(p *replify.Pool) error {
				return p.Submit(canceledCtx, "never-runs", noopJob)
			},
			wantErr: context.Canceled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := replify.NewPool(context.Background(), 0)
			if err := tt.submit(p); !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected Submit error %v, got %v", tt.wantErr, err)
			}
			p.Close()

			w := p.Wait()
			if w.StatusCode() != replify.StatusNoContent.Value() {
				t.Fatalf("expected 204 for a pool with only rejected submissions, got %d", w.StatusCode())
			}
			if w.Total() != 0 {
				t.Fatalf("expected total 0, got %d", w.Total())
			}
			if n := len(resultsOf(t, w)); n != 0 {
				t.Fatalf("expected zero results (no phantom slots), got %d", n)
			}
		})
	}
}

func TestPool_TrySubmitRejectedLeavesNoResult(t *testing.T) {
	t.Parallel()

	// Zero workers and an unbuffered queue: nothing can ever receive the job,
	// so TrySubmit is guaranteed to report false.
	p := replify.NewPool(context.Background(), 0)
	if p.TrySubmit("never-runs", noopJob) {
		t.Fatalf("expected TrySubmit to be rejected with no workers and no queue")
	}
	p.Close()

	w := p.Wait()
	if w.StatusCode() != replify.StatusNoContent.Value() {
		t.Fatalf("expected 204, got %d", w.StatusCode())
	}
	if n := len(resultsOf(t, w)); n != 0 {
		t.Fatalf("expected zero results, got %d", n)
	}
}

func TestRunWorkerPool_AllSubmissionsRejectedCountedOnce(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Zero workers: no receiver, so every Submit is rejected with ctx.Err().
	jobs := []*replify.PoolJob{
		replify.NewPoolJob("a", noopJob),
		replify.NewPoolJob("b", noopJob),
		replify.NewPoolJob("", noopJob), // unnamed: defaults to job-<index>
	}
	w := replify.RunWorkerPool(ctx, 0, jobs)

	results := resultsOf(t, w)
	if len(results) != 3 || w.Total() != 3 {
		t.Fatalf("expected exactly 3 results and total 3, got %d results, total %d", len(results), w.Total())
	}
	for _, r := range results {
		if r.IsSuccess() {
			t.Fatalf("a rejected submission must never be a success: %+v", r)
		}
		if !errors.Is(r.Cause(), context.Canceled) {
			t.Fatalf("expected submission error context.Canceled, got %v", r.Cause())
		}
	}
	if got := results[2].Name(); got != "job-2" {
		t.Fatalf("expected default name job-2 for the unnamed rejected job, got %q", got)
	}
	// The pool's own context is canceled, so Wait reports 499 regardless of
	// job outcomes (see buildGroupWrapper); only the accounting is under test.
	if w.StatusCode() != replify.StatusClientClosedRequest.Value() {
		t.Fatalf("expected 499, got %d", w.StatusCode())
	}
}

func TestRunWorkerPool_MixedAcceptedAndRejected(t *testing.T) {
	t.Parallel()

	base, cancel := context.WithCancel(context.Background())
	ctx := &notifyErrCtx{Context: base, errCalled: make(chan struct{}, 16)}

	// One worker. Job A cancels the context and then stays busy until both
	// later submissions have been rejected, so B and C can never be accepted.
	jobA := func(context.Context) error {
		cancel()
		<-ctx.errCalled // B rejected
		<-ctx.errCalled // C rejected
		return nil
	}
	w := replify.RunWorkerPool(ctx, 1, []*replify.PoolJob{
		replify.NewPoolJob("A", jobA),
		replify.NewPoolJob("B", noopJob),
		replify.NewPoolJob("C", noopJob),
	})

	results := resultsOf(t, w)
	if len(results) != 3 {
		t.Fatalf("expected exactly 3 results, got %d: %+v", len(results), results)
	}
	var succeeded, failed int
	for _, r := range results {
		if r.IsSuccess() {
			succeeded++
			if r.Name() != "A" {
				t.Fatalf("only A may succeed, got success for %q", r.Name())
			}
			continue
		}
		failed++
		if !errors.Is(r.Cause(), context.Canceled) {
			t.Fatalf("expected %q to fail with context.Canceled, got %v", r.Name(), r.Cause())
		}
	}
	if succeeded != 1 || failed != 2 {
		t.Fatalf("expected 1 succeeded / 2 failed, got %d / %d", succeeded, failed)
	}
}

func TestPool_HappyPathAccounting(t *testing.T) {
	t.Parallel()

	var jobs []*replify.PoolJob
	for i := 0; i < 5; i++ {
		jobs = append(jobs, replify.NewPoolJob(fmt.Sprintf("named-%d", i), noopJob))
	}
	w := replify.RunWorkerPool(context.Background(), 3, jobs)

	if w.StatusCode() != replify.StatusOK.Value() || w.Total() != 5 {
		t.Fatalf("expected 200 and total 5, got %d / %d", w.StatusCode(), w.Total())
	}
	seen := make(map[int]bool)
	for _, r := range resultsOf(t, w) {
		if !r.IsSuccess() {
			t.Fatalf("unexpected failure: %+v", r)
		}
		if r.Index() < 0 || r.Index() > 4 || seen[r.Index()] {
			t.Fatalf("indices must be unique and within 0..4, got %d", r.Index())
		}
		seen[r.Index()] = true
	}
}

func TestPool_UnnamedJobsDefaultToIndex(t *testing.T) {
	t.Parallel()

	p := replify.NewPool(context.Background(), 2)
	for i := 0; i < 4; i++ {
		if err := p.Submit(context.Background(), "", noopJob); err != nil {
			t.Fatalf("Submit: %v", err)
		}
	}
	p.Close()

	results := resultsOf(t, p.Wait())
	if len(results) != 4 {
		t.Fatalf("expected 4 results, got %d", len(results))
	}
	for _, r := range results {
		if want := fmt.Sprintf("job-%d", r.Index()); r.Name() != want {
			t.Fatalf("expected default name %q, got %q", want, r.Name())
		}
	}
}

func TestPool_AcceptedNilJobIsOneFailedResult(t *testing.T) {
	t.Parallel()

	p := replify.NewPool(context.Background(), 1)
	if err := p.Submit(context.Background(), "nil-job", nil); err != nil {
		t.Fatalf("a nil job is still accepted, got %v", err)
	}
	p.Close()

	w := p.Wait()
	results := resultsOf(t, w)
	if len(results) != 1 {
		t.Fatalf("expected exactly one result, got %d", len(results))
	}
	if results[0].IsSuccess() || !errors.Is(results[0].Cause(), replify.ErrNilTask) {
		t.Fatalf("expected a failed result carrying ErrNilTask, got %+v", results[0])
	}
}

func TestWorkerGroup_TryGoRejectionLeavesNoResult(t *testing.T) {
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

	for i := 0; i < 3; i++ {
		if g.TryGo("rejected", func(ctx context.Context) (any, error) { return nil, nil }) {
			t.Fatalf("expected TryGo to be rejected while at the concurrency limit")
		}
	}
	close(block)

	results := resultsOf(t, g.Wait())
	if len(results) != 1 || results[0].Name() != "blocker" || !results[0].IsSuccess() {
		t.Fatalf("expected only the blocker's result, got %+v", results)
	}
}

func TestPool_ConcurrentSubmitCountMatchesAccepted(t *testing.T) {
	t.Parallel()

	const goroutines, perGoroutine = 8, 50

	p := replify.NewPool(context.Background(), 4, workergroup.WithQueueSize(4))
	var accepted atomic.Int64
	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < perGoroutine; i++ {
				if i%2 == 0 {
					if err := p.Submit(context.Background(), "", noopJob); err == nil {
						accepted.Add(1)
					}
				} else if p.TrySubmit("", noopJob) {
					accepted.Add(1)
				}
			}
		}()
	}
	wg.Wait()
	p.Close()

	w := p.Wait()
	results := resultsOf(t, w)
	if int64(len(results)) != accepted.Load() {
		t.Fatalf("expected one result per accepted job: %d results, %d accepted", len(results), accepted.Load())
	}
	for _, r := range results {
		if !r.IsSuccess() {
			t.Fatalf("unexpected failure: %+v", r)
		}
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
