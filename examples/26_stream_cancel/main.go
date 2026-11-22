// Package main demonstrates cancelling a running stream with Cancel(). A reader is parked
// between two chunks, the stream is cancelled, and the reader is released: the stream stops
// with a 500 and records the cancellation error.
//
// Run:      go run ./examples/26_stream_cancel
// Expected: status 500, has_errors=true, cancelled=true (errors.Is context.Canceled).
//
// Start must be called with a nil context for Cancel to take effect: Cancel cancels the wrapper's
// own context, which Start only uses when no context is passed (see examples/README.md, Findings).
package main

import (
	"context"
	"errors"
	"io"
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

// parkedReader serves one chunk, then blocks until release is signalled, so the example does
// not depend on timing.
type parkedReader struct {
	reads   int
	parked  chan struct{}
	release chan struct{}
}

func (r *parkedReader) Read(p []byte) (int, error) {
	r.reads++
	if r.reads == 1 {
		return copy(p, "first chunk"), nil
	}
	close(r.parked) // tell the caller the stream is waiting for more data
	<-r.release
	return copy(p, "second chunk"), nil
}

func run() error {
	cfg := replify.NewStreamConfig()
	cfg.Strategy = replify.StrategyDirect
	cfg.ChunkSize = 64

	pr := &parkedReader{parked: make(chan struct{}), release: make(chan struct{})}
	sw := replify.New().WithStreaming(pr, cfg)
	sw.WithWriter(io.Discard)

	status := make(chan int, 1)
	go func() { status <- sw.Start(nil).StatusCode() }()

	<-pr.parked       // the stream processed the first chunk and waits in Read
	sw.Cancel()       // cancel while it waits
	close(pr.release) // let the pending Read return; the stream notices the cancellation next

	select {
	case code := <-status:
		cancelled := false
		for _, err := range sw.Errors() {
			cancelled = cancelled || errors.Is(err, context.Canceled)
		}
		slogger.Info("stream cancelled",
			slogger.Int("status", code),
			slogger.Bool("has_errors", sw.HasErrors()),
			slogger.Bool("cancelled", cancelled))
	case <-time.After(5 * time.Second):
		return errors.New("stream did not stop after Cancel")
	}
	return nil
}
