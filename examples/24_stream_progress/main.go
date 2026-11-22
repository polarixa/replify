// Package main demonstrates progress reporting: WithTotalBytes (for percentages), WithCallback
// and WithHook (called once per chunk), and GetProgress for a final, race-free snapshot.
//
// Run:      go run ./examples/24_stream_progress
// Expected: callbacks=8 hooks=8 for 8 chunks, then percentage=100 transferred_bytes=65536.
//
// Callbacks and hooks run in their own goroutines, may still be running when Start returns, and
// receive the live progress object. Do not read its fields from the callback; count calls (as
// here) or read a snapshot with GetProgress after Start returns.
package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sync/atomic"
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
	const size, chunk = 64 << 10, 8 << 10

	cfg := replify.NewStreamConfig()
	cfg.Strategy = replify.StrategyDirect
	cfg.ChunkSize = chunk

	var callbacks, hooks atomic.Int64
	sw := replify.New().WithStreaming(bytes.NewReader(make([]byte, size)), cfg)
	sw.WithWriter(&bytes.Buffer{})
	sw.WithTotalBytes(size) // enables Percentage and TotalChunks
	sw.WithCallback(func(_ *replify.StreamProgress, _ error) { callbacks.Add(1) })
	sw.WithHook(func(_ *replify.StreamProgress, _ *replify.R) { hooks.Add(1) })

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if w := sw.Start(ctx); w.StatusCode() != http.StatusOK {
		return fmt.Errorf("stream failed: %d %s", w.StatusCode(), w.Message())
	}

	// The notifications are asynchronous: wait (bounded) for all of them to arrive.
	deadline := time.Now().Add(2 * time.Second)
	for (callbacks.Load() < size/chunk || hooks.Load() < size/chunk) && time.Now().Before(deadline) {
		runtime.Gosched()
	}

	p := sw.GetProgress()
	slogger.Info("notifications", slogger.Int64("callbacks", callbacks.Load()), slogger.Int64("hooks", hooks.Load()))
	slogger.Info("final progress",
		slogger.Int("percentage", p.Percentage),
		slogger.Int64("transferred_bytes", p.TransferredBytes),
		slogger.Int64("total_bytes", p.TotalBytes),
		slogger.Int64("total_chunks", p.TotalChunks))
	return nil
}
