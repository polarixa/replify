// Package main shows the streaming configuration: the defaults from NewStreamConfig, applying
// every setter, and the JSON form of the configuration, statistics and progress.
//
// Run:      go run ./examples/31_stream_config
// Expected: the default config JSON, the configured JSON, the result (total_bytes=50000, percentage=100), then the stats and progress JSON helpers.
//
// WithReadTimeout and WithWriteTimeout store their values (milliseconds) but the stream does not
// enforce them yet (see examples/README.md, Findings).
package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
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
	cfg := replify.NewStreamConfig()
	slogger.Info("defaults", slogger.String("config", cfg.JSON()))

	sw := replify.New().WithStreaming(bytes.NewReader(make([]byte, 50_000)), cfg)
	sw.WithWriter(io.Discard)
	sw.WithReceiveMode(false)
	sw.WithStreamingStrategy(replify.StrategyChunked)
	sw.WithCompressionType(replify.CompressGzip)
	sw.WithChunkSize(10_000)
	sw.WithMaxConcurrentChunks(2)
	sw.WithThrottleRate(0) // 0 means unlimited
	sw.WithBufferPooling(false)
	sw.WithReadTimeout(5000)
	sw.WithWriteTimeout(5000)
	sw.WithTotalBytes(50_000)
	slogger.Info("configured", slogger.String("config", cfg.JSON()))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if w := sw.Start(ctx); w.StatusCode() != http.StatusOK {
		return fmt.Errorf("stream failed: %d %s", w.StatusCode(), w.Message())
	}
	stats, progress := sw.GetStats(), sw.GetProgress()
	slogger.Info("result",
		slogger.Int64("total_bytes", stats.TotalBytes),
		slogger.Int64("compressed_bytes", stats.CompressedBytes),
		slogger.Int("percentage", progress.Percentage))

	// StreamingStats and StreamProgress have the same JSON helper as StreamConfig.
	slogger.Info("stats json", slogger.String("json", (&replify.StreamingStats{TotalBytes: 50_000, CompressedBytes: 240}).JSON()))
	slogger.Info("progress json", slogger.String("json", (&replify.StreamProgress{TotalBytes: 50_000, TransferredBytes: 25_000, Percentage: 50}).JSON()))
	return nil
}
