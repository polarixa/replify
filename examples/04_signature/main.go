// Package main demonstrates signing a response with HMAC and verifying the
// signature after the JSON is unwrapped on the receiving side.
//
// Run: go run ./examples/04_signature
// Expected: "valid: true" for the untouched payload, "valid: false" after tampering.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/polarixa/replify"
	"github.com/polarixa/replify/pkg/slogger"
)

type user struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func newConfig() *replify.SignatureConfig {
	return replify.NewSignatureConfig("example-secret").
		WithAlgorithm(replify.HMACSHA256).
		WithMaxAge(time.Minute).
		EnableCanonicalize()
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
	sender := replify.New().
		Created().
		WithBody(user{Name: "John Doe", Email: "john.doe@example.com"})
	sender.ApplySignature(newConfig())

	wire := sender.JSON()

	received, err := replify.UnwrapJSON(wire)
	if err != nil {
		return fmt.Errorf("unwrap: %w", err)
	}
	slogger.Info("verified untouched payload", slogger.Bool("valid", received.VerifySignature(newConfig())))

	// Tamper with the body and verify again.
	tampered, err := replify.UnwrapJSON(wire)
	if err != nil {
		return fmt.Errorf("unwrap: %w", err)
	}
	tampered.WithBody(user{Name: "Mallory", Email: "evil@example.com"})
	slogger.Info("verified tampered payload", slogger.Bool("valid", tampered.VerifySignature(newConfig())))
	return nil
}
