// Package main demonstrates the smallest working use of replify: build a
// standardized response with a message and a map-builder body, then print it.
//
// Run: go run ./examples/01_quickstart
// Expected: a pretty JSON envelope (status_code 200, message, data, meta.request_id).
package main

import (
	"os"

	"github.com/polarixa/replify"
	"github.com/polarixa/replify/pkg/codegen"
	"github.com/polarixa/replify/pkg/mapbuilder"
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
	body := mapbuilder.New().
		Add("id", "u_123").
		Add("username", "jdoe").
		AddPath("profile.avatar", "https://example.com/avatar.jpg").
		AddPath("profile.bio", "Go developer").
		Append("roles", "admin", "editor")

	w := replify.New().
		OK().
		WithMessage("User retrieved successfully").
		WithBodyBuilder(body).
		WithMetaRequestIDGenerated(
			codegen.WithLength(24),
			codegen.WithCharset(codegen.CharsetAlphanumeric),
			codegen.WithPrefix("req_"),
		)

	// JSONPrettyIgnoring drops the named top-level fields from the output.
	slogger.Info("response built", slogger.String("json", w.JSONIgnoring("header")))
	return nil
}
