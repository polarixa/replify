// Package main is a realistic scenario: an HTTP handler streams a large generated body to the
// client through a StreamingWrapper, and the client verifies the download by checksum. The request
// context is passed to Start, so a client disconnect stops the stream.
//
// Run:      go run ./examples/29_stream_http_download
// Expected: status 200, bytes=2097152, checksum_matches=true.
package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
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

// report is the "file" served by the handler.
var report = func() []byte {
	b := make([]byte, 2<<20)
	rand.New(rand.NewSource(42)).Read(b)
	return b
}()

func download(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/octet-stream")
	rw.Header().Set("Content-Length", strconv.Itoa(len(report)))

	cfg := replify.NewStreamConfig()
	cfg.Strategy = replify.StrategyDirect
	cfg.ChunkSize = 128 << 10

	sw := replify.New().WithStreaming(bytes.NewReader(report), cfg)
	sw.WithWriter(rw)

	// The response is committed by the first write, so the outcome can only be logged here.
	if w := sw.Start(r.Context()); w.StatusCode() != http.StatusOK {
		slogger.Error("download stream failed", slogger.Int("status", w.StatusCode()), slogger.String("message", w.Message()))
	}
}

func run() error {
	srv := httptest.NewServer(http.HandlerFunc(download))
	defer srv.Close()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(srv.URL)
	if err != nil {
		return fmt.Errorf("get: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	slogger.Info("download complete",
		slogger.Int("status", resp.StatusCode),
		slogger.Int("bytes", len(body)),
		slogger.Bool("checksum_matches", sha256.Sum256(body) == sha256.Sum256(report)))
	return nil
}
