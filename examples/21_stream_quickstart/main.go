// Package main demonstrates the smallest working use of the streaming API: wrap a reader in a
// StreamingWrapper, point it at a writer, Start it, and read the statistics.
//
// Run:      go run ./examples/21_stream_quickstart
// Expected: status 200, total_bytes=1048576, percentage=100, identical=true.
//
// Setters on a StreamingWrapper return the underlying response wrapper, so they cannot be
// chained; configure through a StreamConfig first and call setters as separate statements.
package main

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
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

// payload returns n deterministic pseudo-random bytes.
func payload(n int) []byte {
	b := make([]byte, n)
	rand.New(rand.NewSource(42)).Read(b)
	return b
}

func run() error {
	data := payload(1 << 20)

	cfg := replify.NewStreamConfig()
	cfg.Strategy = replify.StrategyDirect
	cfg.ChunkSize = 64 << 10

	var out bytes.Buffer
	sw := replify.New().WithStreaming(bytes.NewReader(data), cfg)
	sw.WithWriter(&out)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	w := sw.Start(ctx)
	if w.StatusCode() != http.StatusOK {
		return fmt.Errorf("stream failed: %d %s", w.StatusCode(), w.Message())
	}

	stats := sw.GetStats()
	slogger.Info("stream finished",
		slogger.Int("status", w.StatusCode()),
		slogger.Int64("total_bytes", stats.TotalBytes),
		slogger.Int("percentage", sw.GetProgress().Percentage),
		slogger.Bool("identical", bytes.Equal(out.Bytes(), data)))
	return nil
}
