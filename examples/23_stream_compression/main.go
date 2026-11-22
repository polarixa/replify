// Package main demonstrates sending a stream with gzip and flate compression and reading the
// compression statistics. The gzip output is a standard gzip stream and is decoded here with
// compress/gzip.
//
// Run:      go run ./examples/23_stream_compression
// Expected: gzip and flate lines with compression_ratio below 1, then gzip_roundtrip=true.
//
// Only the send side is shown. Receiving compressed data with WithReceiveMode is not covered
// (see examples/README.md, "Not covered").
package main

import (
	"bytes"
	"compress/gzip"
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

func send(ctx context.Context, data []byte, comp replify.CompressionType) ([]byte, *replify.StreamingStats, error) {
	cfg := replify.NewStreamConfig()
	cfg.Strategy = replify.StrategyChunked
	cfg.Compression = comp
	cfg.ChunkSize = 32 << 10

	var wire bytes.Buffer
	sw := replify.New().WithStreaming(bytes.NewReader(data), cfg)
	sw.WithWriter(&wire)
	if w := sw.Start(ctx); w.StatusCode() != http.StatusOK {
		return nil, nil, fmt.Errorf("%s: %d %s", comp, w.StatusCode(), w.Message())
	}
	return wire.Bytes(), sw.GetStats(), nil
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	data := bytes.Repeat([]byte("highly compressible log line\n"), 20_000)
	var gzipWire []byte
	for _, comp := range []replify.CompressionType{replify.CompressGzip, replify.CompressFlate} {
		wire, stats, err := send(ctx, data, comp)
		if err != nil {
			return err
		}
		if comp == replify.CompressGzip {
			gzipWire = wire
		}
		slogger.Info("compressed stream",
			slogger.String("compression", string(comp)),
			slogger.Int64("total_bytes", stats.TotalBytes),
			slogger.Int64("compressed_bytes", stats.CompressedBytes),
			slogger.Float64("compression_ratio", stats.CompressionRatio),
			slogger.Bool("smaller_on_the_wire", len(wire) < len(data)))
	}

	zr, err := gzip.NewReader(bytes.NewReader(gzipWire))
	if err != nil {
		return fmt.Errorf("open gzip stream: %w", err)
	}
	defer zr.Close()
	back, err := io.ReadAll(zr)
	if err != nil {
		return fmt.Errorf("read gzip stream: %w", err)
	}
	slogger.Info("decoded with compress/gzip", slogger.Bool("gzip_roundtrip", bytes.Equal(back, data)))
	return nil
}
