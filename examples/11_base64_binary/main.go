// Package main demonstrates body type detection and Base64 encoding/decoding:
// JSON bodies, binary bodies and EncodeBodyBase64 / DecodeBodyBase64.
//
// Run: go run ./examples/11_base64_binary
// Expected: IsJSONBody/IsBinaryBody flags and the body before and after Base64 round trip.
package main

import (
	"fmt"
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
	w := replify.New().OK().WithBody(map[string]any{"name": "John Doe"})
	slogger.Info("body type", slogger.Bool("json", w.IsJSONBody()), slogger.Bool("binary", w.IsBinaryBody()))

	w.EncodeBodyBase64()
	slogger.Info("body encoded", slogger.String("body", w.BodyString()))
	w.DecodeBodyBase64()
	slogger.Info("body decoded", slogger.String("body", w.BodyString()))

	// A []byte body is treated as binary data.
	bin := replify.New().OK()
	if _, err := bin.WithBodySuppressed([]byte{0x48, 0x00, 0x65, 0x00, 0x6C, 0x00}); err != nil {
		return fmt.Errorf("set binary body: %w", err)
	}
	slogger.Info("binary body set", slogger.Bool("binary", bin.IsBinaryBody()))
	return nil
}
