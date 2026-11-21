// Package main demonstrates the structured tracing helpers: Checkpoint, Step,
// Flow, Await and Span. Output goes to the replify logger.
//
// Run: go run ./examples/12_tracing
// Expected: log lines for checkpoints, steps 1/3..3/3, the flow, an await and span start/end.
package main

import (
	"os"
	"time"

	"github.com/polarixa/replify"
	"github.com/polarixa/replify/pkg/slogger"
)

func main() {
	// Development logger: text formatter on stdout. Use slogger.NewJSONFormatter() in production.
	slogger.SetGlobalLogger(slogger.New(
		slogger.WithOutput(os.Stdout),
		slogger.WithLevel(slogger.DebugLevel),
	))
	if err := run(); err != nil {
		slogger.Fatal("example failed", slogger.Err(err))
	}
}

func run() error {
	w := replify.New().Processing()

	// Span returns a func that records the elapsed time when deferred.
	done := w.Span("load-user", slogger.String("function", "run"))
	defer done()

	w.Checkpoint("request.received", slogger.String("status", "ok"))
	w.Step(1, 3, "validate request")
	w.Step(2, 3, "authorize user", slogger.String("status", "ok"))
	w.Step(3, 3, "finalize")
	w.Flow("HTTP", "Handler", "UserService", "UserRepository", "Database")
	w.Await(2*time.Millisecond, slogger.String("reason", "backoff"), slogger.Int("attempt", 1))
	w.OK().Checkpoint("request.completed")
	return nil
}
