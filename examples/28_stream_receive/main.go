// Package main demonstrates receive mode: data streamed out by one wrapper is read back by a
// second wrapper switched to receive mode with WithReceiveMode(true).
//
// Run:      go run ./examples/28_stream_receive
// Expected: sent_bytes=100000 received_bytes=100000 identical=true.
//
// Uncompressed streams are shown. Receiving a compressed stream is not covered (see
// examples/README.md, "Not covered").
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

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	data := make([]byte, 100_000)
	rand.New(rand.NewSource(42)).Read(data)

	newConfig := func() *replify.StreamConfig {
		cfg := replify.NewStreamConfig()
		cfg.Strategy = replify.StrategyChunked
		cfg.ChunkSize = 8 << 10
		return cfg
	}

	// Sender: reader -> wire.
	var wire bytes.Buffer
	send := replify.New().WithStreaming(bytes.NewReader(data), newConfig())
	send.WithWriter(&wire)
	if w := send.Start(ctx); w.StatusCode() != http.StatusOK {
		return fmt.Errorf("send failed: %d %s", w.StatusCode(), w.Message())
	}

	// Receiver: wire -> final, in receive mode.
	var final bytes.Buffer
	recv := replify.New().WithStreaming(bytes.NewReader(wire.Bytes()), newConfig())
	recv.WithWriter(&final)
	recv.WithReceiveMode(true)
	if w := recv.Start(ctx); w.StatusCode() != http.StatusOK {
		return fmt.Errorf("receive failed: %d %s", w.StatusCode(), w.Message())
	}

	slogger.Info("round trip",
		slogger.Int("sent_bytes", wire.Len()),
		slogger.Int("received_bytes", final.Len()),
		slogger.Bool("identical", bytes.Equal(final.Bytes(), data)))
	return nil
}
