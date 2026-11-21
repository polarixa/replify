// Package main demonstrates cursor pagination, a machine-readable reason code,
// custom fields, debugging data and switching the HTTP header.
//
// Run: go run ./examples/09_cursor_reason_headers
// Expected: pretty JSON containing cursor, reason, debug, custom_fields in meta, and a 400 header.
package main

import (
	"os"

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
	w := replify.New().
		OK().
		WithMessage("Page of results").
		WithCursor(replify.Cursors().
			WithLimit(10).
			WithHasNext(true).
			WithHasPrevious(false).
			WithNext("cursor-abc")).
		WithCustomFields(map[string]any{"tenant": "acme"}).
		WithDebugging(map[string]any{"query_ms": 12}).
		WithReason(replify.NewReason().WithCode(replify.ReasonCodeAccountSuspended))

	// WithHeader replaces the status header (code, text, type).
	w.WithHeader(replify.BadRequest)

	slogger.Info("response built",
		slogger.String("json", w.JSON()),
		slogger.Stringf("reason_code", "%v", w.Reason().Code()),
		slogger.Int("query_ms", w.DebuggingInt("query_ms", -1)))
	return nil
}
