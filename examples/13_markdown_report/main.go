// Package main demonstrates generating Markdown documents from a response:
// the full report (DumpMDDocTo) and the section docs (SummaryDoc, PagingDoc).
//
// Run: go run ./examples/13_markdown_report
// Expected: the generated Markdown report and the paging section.
package main

import (
	"fmt"
	"os"
	"path/filepath"

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
	dir, err := os.MkdirTemp("", "replify-md-*")
	if err != nil {
		return fmt.Errorf("mkdir temp: %w", err)
	}

	defer os.RemoveAll(dir)

	w := replify.New().
		OK().
		WithMessage("User retrieved successfully").
		WithBody(map[string]any{"id": 123, "name": "John Doe"}).
		WithPagination(replify.FromPages(100, 5))

	slogger.Info("paging section", slogger.String("doc", w.PagingDoc().String()))

	d, wd := w.DumpMDDocTo(filepath.Join(dir, "report.md"))
	if wd.IsError() {
		return fmt.Errorf("dump markdown: %s", wd.Error())
	}
	defer d.Close()

	data, err := os.ReadFile(d.Filepath())
	if err != nil {
		return fmt.Errorf("read report: %w", err)
	}
	slogger.Info("markdown report", slogger.String("doc", string(data)))
	return nil
}
