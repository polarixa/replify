// Package main demonstrates dumping a response to files (JSON and body only)
// inside a temporary directory that is removed afterwards.
//
// Run: go run ./examples/07_dump
// Expected: the dump paths and the content read back from each file.
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
	dir, err := os.MkdirTemp("", "replify-dump-*")
	if err != nil {
		return fmt.Errorf("mkdir temp: %w", err)
	}
	defer os.RemoveAll(dir)

	w := replify.New().
		OK().
		WithMessage("Report generated").
		WithBody(map[string]any{"id": 123, "name": "John Doe"})

	d, wd := w.DumpJSONTo(filepath.Join(dir, "response.json"))
	if wd.IsError() {
		return fmt.Errorf("dump json: %s", wd.Error())
	}
	slogger.Info("json dump written", slogger.String("file", filepath.Base(d.Filepath())))
	slogger.Info("json dump content", slogger.String("content", d.ReadAll()))
	if err := d.Close(); err != nil {
		return fmt.Errorf("close json dump: %w", err)
	}

	b, wb := w.DumpBodyTo(filepath.Join(dir, "body.json"))
	if wb.IsError() {
		return fmt.Errorf("dump body: %s", wb.Error())
	}
	slogger.Info("body dump written", slogger.String("file", filepath.Base(b.Filepath())))
	slogger.Info("body dump content", slogger.String("content", b.ReadAll()))
	if err := b.Close(); err != nil {
		return fmt.Errorf("close body dump: %w", err)
	}
	return nil
}
