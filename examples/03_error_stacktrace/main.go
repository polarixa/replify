// Package main demonstrates building an error response across a call chain,
// capturing the stack trace and generating the ordered resolution guide.
//
// Run: go run ./examples/03_error_stacktrace
// Expected: error flag, status code, error text, and a Markdown guide listing the stack frames.
package main

import (
	"os"

	"github.com/polarixa/replify"
	"github.com/polarixa/replify/pkg/slogger"
)

type user struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func level1() replify.R { return level2() }
func level2() replify.R { return level3() }

// level3 is where the simulated failure originates (the root-cause frame).
func level3() replify.R {
	return replify.New().
		InternalServerError().
		WithError("simulated database connection failure").
		WithBody(user{Name: "John Doe", Email: "john.doe@example.com", Age: 30}).
		InjectStackTrace().
		Reply()
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
	w := level1()
	slogger.Error("request failed",
		slogger.Bool("is_error", w.IsError()),
		slogger.Int("status", w.StatusCode()),
		slogger.String("error", w.Error()))
	slogger.Info("resolution guide", slogger.String("doc", w.ResolveESTOrderedDoc().String()))
	return nil
}
