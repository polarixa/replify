// Package main demonstrates serving a file as a download with FileAttachment.
// A temporary report file is created, streamed through an httptest server, and removed.
//
// Run: go run ./examples/08_file_attachment
// Expected: status 200, Content-Disposition naming report.json, and the file content.
package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
	dir, err := os.MkdirTemp("", "replify-file-*")
	if err != nil {
		return fmt.Errorf("mkdir temp: %w", err)
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "data.json")
	if err := os.WriteFile(path, []byte(`{"report":"ok"}`), 0o600); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	handler := func(rw http.ResponseWriter, _ *http.Request) {
		w := replify.New().OK().WithMessage("File attachment retrieved successfully").WithTotal(1)
		w.FileAttachment(path, "report.json")
		w.Write(rw)
	}
	srv := httptest.NewServer(http.HandlerFunc(handler))
	defer srv.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(srv.URL)
	if err != nil {
		return fmt.Errorf("get: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}
	slogger.Info("attachment received",
		slogger.Int("status", resp.StatusCode),
		slogger.String("disposition", resp.Header.Get("Content-Disposition")),
		slogger.String("content", string(data)))
	return nil
}
