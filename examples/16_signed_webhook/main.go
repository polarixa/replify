// Package main simulates a signed webhook: the sender signs a replify payload and
// ships the signature in HTTP headers; the receiver verifies it before trusting the body.
//
// Run: go run ./examples/16_signed_webhook
// Expected: "valid delivery -> 200", "tampered delivery -> 401".
package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/polarixa/replify"
	"github.com/polarixa/replify/pkg/slogger"
)

const secret = "example-webhook-secret"

func config() *replify.SignatureConfig {
	return replify.NewSignatureConfig(secret).
		WithAlgorithm(replify.HMACSHA256).
		WithMaxAge(time.Minute).
		EnableCanonicalize()
}

// receiver verifies the signature headers against the payload before accepting it.
func receiver(rw http.ResponseWriter, r *http.Request) {
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		replify.New().BadRequest().WithErrorf("read body: %v", err).WriteJSON(rw)
		return
	}
	payload, err := replify.UnwrapJSON(string(raw))
	if err != nil {
		replify.New().BadRequest().WithErrorf("invalid payload: %v", err).WriteJSON(rw)
		return
	}
	if !payload.VerifySignatureHeader(r, config()) {
		replify.New().Unauthorized().WithError("signature verification failed").WriteJSON(rw)
		return
	}
	replify.New().OK().WithMessage("webhook accepted").WriteJSON(rw)
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

func deliver(url string, event *replify.R, tamper bool) (int, error) {
	// A recorder collects the signature headers that ApplySignatureHeader writes.
	rec := httptest.NewRecorder()
	event.ApplySignatureHeader(rec, config())

	body := event.JSON()
	if tamper {
		body = string(bytes.Replace([]byte(body), []byte("paid"), []byte("free"), 1))
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBufferString(body))
	if err != nil {
		return 0, err
	}
	for k, v := range rec.Header() {
		req.Header[k] = v
	}
	resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

func run() error {
	srv := httptest.NewServer(http.HandlerFunc(receiver))
	defer srv.Close()

	event := replify.New().OK().
		WithMessage("invoice.updated").
		WithBody(map[string]any{"invoice": "inv_1", "status": "paid"}).
		Reply()

	for _, c := range []struct {
		name   string
		tamper bool
	}{{"valid delivery", false}, {"tampered delivery", true}} {
		code, err := deliver(srv.URL, &event, c.tamper)
		if err != nil {
			return fmt.Errorf("%s: %w", c.name, err)
		}
		slogger.Info("delivery", slogger.String("case", c.name), slogger.Int("status", code))
	}
	return nil
}
