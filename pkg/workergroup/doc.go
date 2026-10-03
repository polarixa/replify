// Package workergroup is a standalone, dependency-free reimagining of
// golang.org/x/sync/errgroup, redesigned around three goals:
//
//   - Object-oriented ergonomics: behavior is configured once, through
//     constructors and functional [Option] values, instead of mutating
//     exported struct fields directly.
//   - Dynamic, live scaling: the concurrency limit of a [Group] (and the
//     worker count of a [WorkerPool]) can be changed at any time, including
//     while goroutines are actively running. The upstream errgroup panics
//     if SetLimit is called while goroutines are active; this package
//     never does.
//   - A genuine, persistent worker pool: [WorkerPool] builds on [Group] to
//     provide long-lived worker goroutines that pull [Job] values from an
//     internal queue, rather than the one-goroutine-per-call model of a
//     plain error group. Workers can be added or removed on the fly to
//     scale the pool up or down in response to load.
//
// # Error group
//
// [Group] is API-compatible in spirit with errgroup.Group but adds two
// error-handling strategies via [ErrorMode]:
//
//   - [FirstError] (the default) keeps only the first error returned by any
//     task and cancels the Group's derived [context.Context], exactly like
//     the upstream package.
//   - [CollectErrors] accumulates every error from every task (joined with
//     [errors.Join]) and never cancels the context on error, which suits
//     long-running supervisors that must keep working after a single
//     failure.
//
// Basic usage:
//
//	g, ctx := workergroup.NewWithContext(context.Background(), workergroup.WithLimit(4))
//	for _, url := range urls {
//		url := url
//		g.Go(func() error {
//			return fetch(ctx, url)
//		})
//	}
//	if err := g.Wait(); err != nil {
//		log.Fatal(err)
//	}
//
// Scaling a running group up or down requires no special handling:
//
//	g.SetLimit(16) // safe even while goroutines are in flight
//
// # Worker pool
//
// [WorkerPool] is for workloads that are better modeled as a fixed (but
// adjustable) number of persistent workers consuming a queue, such as a
// request processor or a background job runner, rather than as one
// goroutine per unit of work:
//
//	pool := workergroup.NewWorkerPool(context.Background(), 4)
//	for _, job := range jobs {
//		job := job
//		_ = pool.Submit(context.Background(), func(ctx context.Context) error {
//			return process(ctx, job)
//		})
//	}
//	pool.ScaleTo(16) // scale up under load
//	pool.Close()      // stop accepting new jobs
//	err := pool.Wait() // wait for the queue to drain
//
// All exported types in this package are safe for concurrent use by
// multiple goroutines unless their documentation says otherwise.
package workergroup
