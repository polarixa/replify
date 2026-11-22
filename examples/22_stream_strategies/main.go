// Package main compares the Direct and Chunked streaming strategies on the same input: both
// produce identical output, in order. (The Buffered strategy is deliberately not shown; see
// examples/README.md, "Not covered".)
//
// Run:      go run ./examples/22_stream_strategies
// Expected: one line per strategy with identical=true, then identical_to_each_other=true.
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

func payload(n int) []byte {
	b := make([]byte, n)
	rand.New(rand.NewSource(42)).Read(b)
	return b
}

// stream sends data through the given strategy and returns what the writer received.
func stream(ctx context.Context, data []byte, strategy replify.StreamingStrategy) ([]byte, error) {
	cfg := replify.NewStreamConfig()
	cfg.Strategy = strategy
	cfg.ChunkSize = 16 << 10

	var out bytes.Buffer
	sw := replify.New().WithStreaming(bytes.NewReader(data), cfg)
	sw.WithWriter(&out)
	if w := sw.Start(ctx); w.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("%s: %d %s", strategy, w.StatusCode(), w.Message())
	}
	return out.Bytes(), nil
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	data := payload(200_000)
	var outputs [][]byte
	for _, strategy := range []replify.StreamingStrategy{replify.StrategyDirect, replify.StrategyChunked} {
		got, err := stream(ctx, data, strategy)
		if err != nil {
			return err
		}
		slogger.Info("strategy finished",
			slogger.String("strategy", string(strategy)),
			slogger.Int("bytes", len(got)),
			slogger.Bool("identical", bytes.Equal(got, data)))
		outputs = append(outputs, got)
	}
	slogger.Info("compared", slogger.Bool("identical_to_each_other", bytes.Equal(outputs[0], outputs[1])))
	return nil
}
