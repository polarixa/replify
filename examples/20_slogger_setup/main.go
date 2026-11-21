// Package main demonstrates slogger setup for a service handling replify responses:
// JSON formatter, bound static fields, named sub-loggers, and request-scoped
// context fields (trace ID) that appear on every entry logged with that context.
//
// Run: go run ./examples/20_slogger_setup
// Expected: JSON log lines carrying service/env, then request lines that all share trace_id.
package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/polarixa/replify"
	"github.com/polarixa/replify/pkg/slogger"
)

func main() {
	// Production setup: JSON formatter plus static fields bound once.
	logger := slogger.New(
		slogger.WithOutput(os.Stdout),
		slogger.WithLevel(slogger.InfoLevel),
		slogger.WithFormatter(slogger.NewJSONFormatter()),
		slogger.WithFields(slogger.String("service", "api"), slogger.String("env", "example")),
	)
	slogger.SetGlobalLogger(logger)

	if err := run(logger); err != nil {
		slogger.Fatal("example failed", slogger.Err(err))
	}
}

func run(logger *slogger.Logger) error {
	httpLog := logger.Named("http")

	handler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// Attach a trace ID once; every entry logged with ctx carries it.
		start := time.Now()
		ctx := slogger.WithContextFields(r.Context(), slogger.String("trace_id", "trace-1234"))

		httpLog.WithContext(ctx).Info("request started",
			slogger.String("method", r.Method), slogger.String("path", r.URL.Path))

		w := replify.New().OK().WithMessage("pong").WithRequest(r)
		w.WriteJSON(rw)

		httpLog.WithContext(ctx).Info("request finished",
			slogger.Int("status", w.StatusCode()), slogger.Duration("elapsed", time.Since(start)))
	})

	srv := httptest.NewServer(handler)
	defer srv.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/ping", nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}
