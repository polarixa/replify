// Package main shows the common layered pattern: a service returns domain errors,
// and a single mapper converts them to replify responses via errors.Is.
//
// Run: go run ./examples/19_service_error_mapping
// Expected: one line per outcome: 200 ok, 404 not found, 409 conflict, 500 for unexpected errors.
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/polarixa/replify"
	"github.com/polarixa/replify/pkg/slogger"
)

var (
	errNotFound = errors.New("not found")
	errConflict = errors.New("conflict")
)

// findOrder stands in for a service/repository call.
func findOrder(id string) (map[string]any, error) {
	switch id {
	case "o_1":
		return map[string]any{"id": id, "total": 42}, nil
	case "o_2":
		return nil, fmt.Errorf("order %s: %w", id, errConflict)
	case "o_3":
		return nil, errors.New("database connection failure")
	}
	return nil, fmt.Errorf("order %s: %w", id, errNotFound)
}

// toResponse is the single place where domain errors become HTTP semantics.
func toResponse(order map[string]any, err error) *replify.R {
	var w = replify.New()
	switch {
	case err == nil:
		w.OK().WithBody(order)
	case errors.Is(err, errNotFound):
		w.NotFound().WithError(err.Error())
	case errors.Is(err, errConflict):
		w.Conflict().WithError(err.Error())
	default:
		// Unexpected: 500 with a captured stack trace for diagnostics.
		w.InternalServerError().WithError(err.Error()).InjectStackTrace()
	}
	r := w.Reply()
	return &r
}

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
	for _, id := range []string{"o_1", "o_2", "o_3", "o_4"} {
		order, err := findOrder(id)
		r := toResponse(order, err)
		slogger.Info("mapped",
			slogger.String("order", id),
			slogger.Int("status", r.StatusCode()),
			slogger.Bool("client_error", r.IsClientError()),
			slogger.Bool("server_error", r.IsServerError()))
	}
	return nil
}
