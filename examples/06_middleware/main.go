// Package main demonstrates the Logger and Recovery middleware. Recovery turns a
// handler panic into a replify 500 response; Logger records the final status.
//
// Run: go run ./examples/06_middleware
// Expected: 200 for /ok, 500 for /panic (server keeps running), plus request log lines.
package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
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
	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(rw http.ResponseWriter, r *http.Request) {
		replify.New().OK().WithMessage("fine").WriteJSON(rw)
	})
	mux.HandleFunc("/panic", func(http.ResponseWriter, *http.Request) {
		panic("intentional panic - testing Recovery middleware")
	})

	// Logger outer, Recovery inner: the logged status reflects the final response.
	srv := httptest.NewServer(replify.Logger()(replify.Recovery()(mux)))
	defer srv.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	for _, path := range []string{"/ok", "/panic"} {
		resp, err := client.Get(srv.URL + path)
		if err != nil {
			return fmt.Errorf("get %s: %w", path, err)
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		slogger.Info("client received", slogger.String("path", path), slogger.Int("status", resp.StatusCode), slogger.Int("bytes", len(body)))
	}
	return nil
}
