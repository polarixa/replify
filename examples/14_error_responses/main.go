// Package main demonstrates typical 4xx error responses served by a handler:
// 400 with a formatted error, 404 with a reason, and a 204 with no body.
//
// Run: go run ./examples/14_error_responses
// Expected: one status line per route plus the JSON error envelope (none for 204).
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
	mux.HandleFunc("/bad", func(rw http.ResponseWriter, r *http.Request) {
		replify.New().BadRequest().
			WithErrorf("invalid query parameter %q", "page").
			WithRequest(r).
			WriteJSON(rw)
	})
	mux.HandleFunc("/missing", func(rw http.ResponseWriter, _ *http.Request) {
		replify.New().NotFound().
			WithMessage("User not found").
			WithReason(replify.NewReason().WithCode(replify.ReasonCodeAccessDenied)).
			WriteJSON(rw)
	})
	mux.HandleFunc("/empty", func(rw http.ResponseWriter, _ *http.Request) {
		replify.New().NoContent().WriteJSON(rw)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	for _, p := range []string{"/bad", "/missing", "/empty"} {
		resp, err := client.Get(srv.URL + p)
		if err != nil {
			return fmt.Errorf("get %s: %w", p, err)
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("read %s: %w", p, err)
		}
		slogger.Info("client received", slogger.String("path", p), slogger.Int("status", resp.StatusCode), slogger.String("body", string(body)))
	}
	return nil
}
