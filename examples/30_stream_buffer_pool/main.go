// Package main demonstrates BufferPool (Get/Put reuse, a full pool discarding extras, nil being
// ignored) and the WithBufferPooling switch on a stream.
//
// Run:      go run ./examples/30_stream_buffer_pool
// Expected: reused=true, discarded=true, then both pooled and unpooled streams with status 200.
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
	// A pool of 1 buffer of 4 KiB.
	pool := replify.NewBufferPool(4096, 1)

	buf := pool.Get() // the pool is empty, so this allocates a new 4096-byte buffer
	pool.Put(buf)
	again := pool.Get()
	slogger.Info("get after put", slogger.Int("len", len(again)), slogger.Bool("reused", &again[0] == &buf[0]))

	pool.Put(again)
	extra := make([]byte, 4096)
	pool.Put(extra) // the pool holds 1 buffer, so this one is dropped
	first, second := pool.Get(), pool.Get()
	slogger.Info("full pool",
		slogger.Bool("kept_first", &first[0] == &again[0]),
		slogger.Bool("discarded", &second[0] != &extra[0]))
	pool.Put(nil) // ignored

	// Pooling on a stream is a configuration switch.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	for _, pooled := range []bool{true, false} {
		cfg := replify.NewStreamConfig()
		cfg.Strategy = replify.StrategyDirect
		cfg.UseBufferPool = pooled

		sw := replify.New().WithStreaming(bytes.NewReader(make([]byte, 100_000)), cfg)
		sw.WithWriter(io.Discard)
		sw.WithBufferPooling(pooled)
		if w := sw.Start(ctx); w.StatusCode() != http.StatusOK {
			return fmt.Errorf("pooling=%v: %d %s", pooled, w.StatusCode(), w.Message())
		}
		slogger.Info("stream", slogger.Bool("buffer_pooling", pooled), slogger.Int64("total_bytes", sw.GetStats().TotalBytes))
	}
	return nil
}
