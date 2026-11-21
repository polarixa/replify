// Package main demonstrates parsing a JSON envelope received from an API back
// into a replify wrapper and reading its parts.
//
// Run: go run ./examples/05_unwrap_json
// Expected: status, message, body and a debug value read from the parsed envelope.
package main

import (
	"fmt"
	"os"

	"github.com/polarixa/replify"
	"github.com/polarixa/replify/pkg/slogger"
)

const payload = `{
  "data": {"id": 123, "name": "John Doe"},
  "debug": {"trace": "session 4919e84f", "refer": 1234},
  "message": "User retrieved successfully",
  "pagination": {"is_last": true, "page": 1, "per_page": 2, "total_items": 2, "total_pages": 1},
  "header": {"code": 200, "text": "OK", "type": "Successful"},
  "path": "/api/v1/users",
  "status_code": 200,
  "total": 1
}`

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
	w, err := replify.UnwrapJSON(payload)
	if err != nil {
		return fmt.Errorf("unwrap: %w", err)
	}
	slogger.Info("envelope parsed",
		slogger.Int("status", w.StatusCode()),
		slogger.String("message", w.Message()),
		slogger.String("body", w.BodyString()),
		slogger.Stringf("trace", "%v", w.OnDebugging("trace")))

	if _, err := replify.UnwrapJSON("not json"); err != nil {
		slogger.Warn("invalid input rejected", slogger.Err(err))
	}
	return nil
}
