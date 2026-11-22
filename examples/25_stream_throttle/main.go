// Package main demonstrates WithThrottleRate: a 64 KiB stream limited to 256 KiB/s takes about a
// quarter of a second.
//
// Run:      go run ./examples/25_stream_throttle
// Expected: rate_within_20_percent=true (the elapsed time itself varies from run to run).
package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
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
	const size, rate = 64 << 10, 256 << 10 // bytes, bytes per second

	cfg := replify.NewStreamConfig()
	cfg.Strategy = replify.StrategyDirect
	cfg.ChunkSize = 8 << 10

	sw := replify.New().WithStreaming(bytes.NewReader(make([]byte, size)), cfg)
	sw.WithWriter(io.Discard)
	sw.WithThrottleRate(rate)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	start := time.Now()
	if w := sw.Start(ctx); w.StatusCode() != http.StatusOK {
		return fmt.Errorf("stream failed: %d %s", w.StatusCode(), w.Message())
	}
	elapsed := time.Since(start)

	effective := float64(size) / elapsed.Seconds()
	slogger.Info("throttled stream",
		slogger.Int("configured_bytes_per_second", rate),
		slogger.Bool("rate_within_20_percent", math.Abs(effective-rate)/rate < 0.2))
	return nil
}
