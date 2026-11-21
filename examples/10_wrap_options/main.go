// Package main demonstrates the functional-options constructor Wrap(...) as an
// alternative to the fluent replify.New() builder.
//
// Run: go run ./examples/10_wrap_options
// Expected: two equivalent envelopes (status 200, same message and pagination).
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
	viaOptions := replify.Wrap(
		replify.WithStatusCode(replify.OK),
		replify.WithMessage("User retrieved successfully"),
		replify.WithBody(map[string]any{"id": 123}),
		replify.WithPagination(replify.FromPages(100, 5)),
	)

	viaFluent := replify.New().
		OK().
		WithMessage("User retrieved successfully").
		WithBody(map[string]any{"id": 123}).
		WithPagination(replify.FromPages(100, 5))

	slogger.Info("built with options", slogger.String("json", viaOptions.JSON()))
	slogger.Info("built fluently", slogger.String("json", viaFluent.JSON()))
	return nil
}
