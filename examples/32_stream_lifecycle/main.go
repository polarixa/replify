// Package main demonstrates the lifecycle accessors: IsStreaming while a stream runs, the
// wrapper's StreamingContext, GetWrapper for the response wrapper, and Close, which cancels the
// context and closes the reader and writer when they implement io.Closer.
//
// Run:      go run ./examples/32_stream_lifecycle
// Expected: streaming_during=true streaming_after=false, then closed_reader=true closed_writer=true context_cancelled=true.
package main

import (
	"bytes"
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

// parkedReader waits for release before returning its data, then reports EOF. It records Close.
type parkedReader struct {
	parked  chan struct{}
	release chan struct{}
	served  bool
	closed  bool
}

func (r *parkedReader) Read(p []byte) (int, error) {
	if r.served {
		return 0, io.EOF
	}
	r.served = true
	close(r.parked)
	<-r.release
	return copy(p, "payload"), nil
}

func (r *parkedReader) Close() error { r.closed = true; return nil }

// closingWriter collects output and records Close.
type closingWriter struct {
	bytes.Buffer
	closed bool
}

func (w *closingWriter) Close() error { w.closed = true; return nil }

func run() error {
	cfg := replify.NewStreamConfig()
	cfg.Strategy = replify.StrategyDirect

	reader := &parkedReader{parked: make(chan struct{}), release: make(chan struct{})}
	writer := &closingWriter{}
	sw := replify.New().WithStreaming(reader, cfg)
	sw.WithWriter(writer)

	done := make(chan int, 1)
	go func() { done <- sw.Start(nil).StatusCode() }()

	<-reader.parked
	during := sw.IsStreaming()
	close(reader.release)

	select {
	case status := <-done:
		slogger.Info("stream finished",
			slogger.Bool("streaming_during", during),
			slogger.Bool("streaming_after", sw.IsStreaming()),
			slogger.Int("status", sw.GetWrapper().StatusCode()),
			slogger.Int("start_status", status),
			slogger.Bool("context_active", sw.StreamingContext().Err() == nil))
	case <-time.After(5 * time.Second):
		return errors.New("stream did not finish")
	}

	sw.Close() // cancels the context and closes the reader and the writer
	slogger.Info("closed",
		slogger.Bool("closed_reader", reader.closed),
		slogger.Bool("closed_writer", writer.closed),
		slogger.Bool("context_cancelled", sw.StreamingContext().Err() != nil))
	return nil
}
