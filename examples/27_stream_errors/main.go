// Package main demonstrates the streaming error paths: a nil reader (400), invalid setter values
// (400), an unknown strategy (500), and a reader that fails mid-stream (500 with Errors() and
// FailedChunks).
//
// Run:      go run ./examples/27_stream_errors
// Expected: one line per case with the HTTP status; the mid-stream case reports wraps_reader_error=true.
package main

import (
	"bytes"
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

var errDisk = errors.New("disk unplugged")

// flakyReader returns one chunk of data and then fails.
type flakyReader struct{ sent bool }

func (f *flakyReader) Read(p []byte) (int, error) {
	if !f.sent {
		f.sent = true
		return copy(p, "some data"), nil
	}
	return 0, errDisk
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. No reader.
	slogger.Info("nil reader", slogger.Int("status", replify.NewStreaming(nil, nil).Start(ctx).StatusCode()))

	// 2. Invalid setter values are rejected with 400 and leave the configuration unchanged.
	sw := replify.NewStreaming(bytes.NewReader(nil), nil)
	slogger.Info("invalid setters",
		slogger.Int("chunk_size_0", sw.WithChunkSize(0).StatusCode()),
		slogger.Int("max_concurrent_0", sw.WithMaxConcurrentChunks(0).StatusCode()),
		slogger.Int("throttle_negative", sw.WithThrottleRate(-1).StatusCode()),
		slogger.Int("empty_strategy", sw.WithStreamingStrategy("").StatusCode()))

	// 3. Unknown strategy.
	cfg := replify.NewStreamConfig()
	cfg.Strategy = replify.StreamingStrategy("teleport")
	bad := replify.NewStreaming(bytes.NewReader([]byte("x")), cfg)
	slogger.Info("unknown strategy", slogger.Int("status", bad.Start(ctx).StatusCode()))

	// 4. The reader fails mid-stream.
	cfg = replify.NewStreamConfig()
	cfg.Strategy = replify.StrategyDirect
	fl := replify.NewStreaming(&flakyReader{}, cfg)
	fl.WithWriter(io.Discard)
	status := fl.Start(ctx).StatusCode()
	wraps := false
	for _, err := range fl.Errors() {
		wraps = wraps || errors.Is(err, errDisk)
	}
	slogger.Info("reader failed mid-stream",
		slogger.Int("status", status),
		slogger.Bool("has_errors", fl.HasErrors()),
		slogger.Bool("wraps_reader_error", wraps),
		slogger.Int64("failed_chunks", fl.GetStats().FailedChunks))
	return nil
}
