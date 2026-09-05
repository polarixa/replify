package replify_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/polarixa/replify"
	"github.com/polarixa/replify/pkg/slogger"
)

// --- test helpers ------------------------------------------------------------

// panicHandler returns an http.Handler that unconditionally panics with v.
func panicHandler(v any) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(v)
	})
}

// okHandler returns an http.Handler that writes a 200 response with body.
func okHandler(body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	})
}

// captureLogger builds a JSON-formatted logger writing into buf at the given
// level and installs it as the package-level slogger used by Recovery() and
// Logger(), which both capture slogger.S() once at middleware-construction
// time. Callers must build the middleware AFTER calling this, and should
// defer slogger.ResetGlobalLogger() to avoid leaking state into other tests.
func captureLogger(buf *bytes.Buffer, level slogger.Level) {
	l := slogger.New(
		slogger.WithOutput(buf),
		slogger.WithFormatter(slogger.NewJSONFormatter()),
		slogger.WithLevel(level),
	)
	slogger.SetGlobalLogger(l)
}

// logLines decodes each newline-delimited JSON log entry in buf.
func logLines(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()
	var entries []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if line == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("log line is not valid JSON: %v\nline: %s", err, line)
		}
		entries = append(entries, m)
	}
	return entries
}

// --- Recovery: happy path ----------------------------------------------------

func TestRecovery_NoPanic(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ok", nil)

	replify.Recovery()(okHandler("hello")).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if rr.Body.String() != "hello" {
		t.Fatalf("expected body %q, got %q", "hello", rr.Body.String())
	}
}

// --- Recovery: panic value variants ------------------------------------------

func TestRecovery_PanicString(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/crash", nil)

	replify.Recovery()(panicHandler("something went wrong")).ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v\nbody: %s", err, rr.Body.String())
	}
	if body["message"] != "an unexpected error occurred" {
		t.Errorf("expected generic message, got %v", body["message"])
	}
}

func TestRecovery_PanicError(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/crash", nil)

	replify.Recovery()(panicHandler(errors.New("db connection lost"))).ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}

func TestRecovery_PanicNonStringValue(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/crash", nil)

	replify.Recovery()(panicHandler(42)).ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}

func TestRecovery_PanicNilMapWrite(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/crash", nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var m map[string]int
		m["boom"] = 1
	})

	replify.Recovery()(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}

// --- Recovery: response format / information disclosure ---------------------

// TestRecovery_ResponseFormat verifies the recovery response matches the
// Replify JSON contract and does not expose the panic value or stack trace to
// the client — only the minimal, public-safe Issue.
func TestRecovery_ResponseFormat(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/crash", nil)

	replify.Recovery()(panicHandler("internal-secret")).ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v\nbody: %s", err, rr.Body.String())
	}

	if sc, ok := body["status_code"]; !ok {
		t.Error("response missing status_code field")
	} else if int(sc.(float64)) != http.StatusInternalServerError {
		t.Errorf("expected status_code 500, got %v", sc)
	}
	if _, ok := body["message"]; !ok {
		t.Error("response missing message field")
	}

	issue, ok := body["issue"].(map[string]any)
	if !ok {
		t.Fatalf("response missing issue field, got %#v", body)
	}
	if _, ok := issue["id"]; !ok {
		t.Error("issue missing id field")
	}
	if _, ok := issue["fingerprint"]; !ok {
		t.Error("issue missing fingerprint field")
	}
	if msg, ok := issue["message"].(string); !ok || msg != "panic: internal-secret" {
		t.Errorf("expected issue message %q, got %v", "panic: internal-secret", issue["message"])
	}

	// The Issue message intentionally carries the panic's string form (per
	// design), but the raw stack trace must never reach the client.
	dbg, _ := body["debug"].(map[string]any)
	if _, ok := dbg["panic"]; ok {
		t.Error("response debug must not expose the raw panic value")
	}
	if _, ok := dbg["stack"]; ok {
		t.Error("response debug must not expose the raw stack trace")
	}
	if strings.Contains(fmt.Sprint(body), "goroutine ") {
		t.Error("response must not leak a raw goroutine stack trace")
	}
}

func TestRecovery_ContentType(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/crash", nil)

	replify.Recovery()(panicHandler("boom")).ServeHTTP(rr, req)

	if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("expected JSON content type, got %q", ct)
	}
}

// --- Recovery: response already started -------------------------------------

// TestRecovery_ResponseAlreadyStarted verifies that a panic after headers have
// been committed does not produce a second response or corrupt the connection.
// The original status code and body are preserved; no JSON error body is
// appended on top of the partial response.
func TestRecovery_ResponseAlreadyStarted(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/crash", nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("partial"))
		panic("late failure")
	})

	replify.Recovery()(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected original status 200 to be preserved, got %d", rr.Code)
	}
	if rr.Body.String() != "partial" {
		t.Fatalf("expected body to remain %q, got %q", "partial", rr.Body.String())
	}
}

func TestRecovery_ResponseAlreadyStarted_WriteOnly(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/crash", nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Write() alone (no explicit WriteHeader) also marks the response as started.
		_, _ = w.Write([]byte("partial"))
		panic("late failure")
	})

	replify.Recovery()(handler).ServeHTTP(rr, req)

	if rr.Body.String() != "partial" {
		t.Fatalf("expected body to remain %q, got %q", "partial", rr.Body.String())
	}
}

// --- Recovery: http.ErrAbortHandler sentinel ---------------------------------

// TestRecovery_ErrAbortHandler verifies that panics with http.ErrAbortHandler
// are re-panicked (so net/http can close the connection) rather than converted
// into a JSON error response.
func TestRecovery_ErrAbortHandler(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/abort", nil)

	var recovered any
	func() {
		defer func() { recovered = recover() }()
		replify.Recovery()(panicHandler(http.ErrAbortHandler)).ServeHTTP(rr, req)
	}()

	if recovered != http.ErrAbortHandler {
		t.Fatalf("expected re-panic with http.ErrAbortHandler, got %v", recovered)
	}
	if rr.Body.Len() != 0 {
		t.Errorf("expected no response body to be written, got %q", rr.Body.String())
	}
}

// --- Recovery: Flush / Unwrap forwarding -------------------------------------

func TestRecovery_FlushForwarding(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ok", nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("streamed"))
		if err := http.NewResponseController(w).Flush(); err != nil {
			t.Errorf("unexpected error flushing through recoveryWriter: %v", err)
		}
	})

	replify.Recovery()(handler).ServeHTTP(rr, req)

	if !rr.Flushed {
		t.Error("expected Flush to be forwarded to the underlying ResponseWriter")
	}
}

func TestRecovery_UnwrapReachesUnderlyingWriter(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ok", nil)

	var unwrapped http.ResponseWriter
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type unwrapper interface{ Unwrap() http.ResponseWriter }
		u, ok := w.(unwrapper)
		if !ok {
			t.Fatal("expected recoveryWriter to implement Unwrap()")
		}
		unwrapped = u.Unwrap()
		w.WriteHeader(http.StatusOK)
	})

	replify.Recovery()(handler).ServeHTTP(rr, req)

	if unwrapped != http.ResponseWriter(rr) {
		t.Error("expected Unwrap() to return the original http.ResponseWriter")
	}
}

// --- Recovery: sensitive header exclusion from logs --------------------------

func TestRecovery_LogsExcludeSensitiveHeaders(t *testing.T) {
	var buf bytes.Buffer
	captureLogger(&buf, slogger.TraceLevel)
	defer slogger.ResetGlobalLogger()

	mw := replify.Recovery()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/crash", nil)
	req.Header.Set("Authorization", "Bearer top-secret-token")
	req.Header.Set("Cookie", "session=super-secret")
	req.Header.Set("X-Request-Id", "req-123")

	mw(panicHandler("boom")).ServeHTTP(rr, req)

	entries := logLines(t, &buf)
	if len(entries) == 0 {
		t.Fatal("expected at least one log entry")
	}
	raw := buf.String()
	if strings.Contains(raw, "top-secret-token") || strings.Contains(raw, "super-secret") {
		t.Error("log output must not contain sensitive header values")
	}

	found := false
	for _, e := range entries {
		if e["msg"] == "panic recovered" {
			found = true
			if e["request_id"] != "req-123" {
				t.Errorf("expected request_id field %q, got %v", "req-123", e["request_id"])
			}
			if e["method"] != http.MethodGet {
				t.Errorf("expected method field %q, got %v", http.MethodGet, e["method"])
			}
			if e["path"] != "/crash" {
				t.Errorf("expected path field %q, got %v", "/crash", e["path"])
			}
			if e["panic"] != "boom" {
				t.Errorf("expected panic field %q, got %v", "boom", e["panic"])
			}
			if _, ok := e["stack"]; !ok {
				t.Error("expected stack field to be present in the server-side log")
			}
			if e["level"] != "ERROR" {
				t.Errorf("expected ERROR level, got %v", e["level"])
			}
		}
	}
	if !found {
		t.Error("expected a \"panic recovered\" log entry")
	}
}

func TestRecovery_LogsWithoutRequestID(t *testing.T) {
	var buf bytes.Buffer
	captureLogger(&buf, slogger.TraceLevel)
	defer slogger.ResetGlobalLogger()

	mw := replify.Recovery()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/crash", nil)
	mw(panicHandler("boom")).ServeHTTP(rr, req)

	entries := logLines(t, &buf)
	if len(entries) != 1 {
		t.Fatalf("expected exactly one log entry, got %d", len(entries))
	}
	if _, ok := entries[0]["request_id"]; ok {
		t.Error("expected no request_id field when X-Request-Id header is absent")
	}
}

// --- Recovery: concurrency ----------------------------------------------------

// TestRecovery_Concurrent exercises Recovery() under concurrent load, mixing
// panicking and non-panicking requests, to catch data races around the
// shared fieldsPool. Run with -race to be effective.
func TestRecovery_Concurrent(t *testing.T) {
	mw := replify.Recovery()

	const n = 100
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/x", nil)
			var h http.Handler
			if i%2 == 0 {
				h = panicHandler(fmt.Sprintf("panic-%d", i))
			} else {
				h = okHandler("ok")
			}
			mw(h).ServeHTTP(rr, req)
			if i%2 == 0 && rr.Code != http.StatusInternalServerError {
				t.Errorf("request %d: expected 500, got %d", i, rr.Code)
			}
			if i%2 != 0 && rr.Code != http.StatusOK {
				t.Errorf("request %d: expected 200, got %d", i, rr.Code)
			}
		}(i)
	}
	wg.Wait()
}

// --- Logger: basic fields -----------------------------------------------------

func TestLogger_BasicFields(t *testing.T) {
	var buf bytes.Buffer
	captureLogger(&buf, slogger.TraceLevel)
	defer slogger.ResetGlobalLogger()

	mw := replify.Logger()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users?token=secret", nil)
	req.RemoteAddr = "192.0.2.1:1234"

	mw(okHandler("hi")).ServeHTTP(rr, req)

	entries := logLines(t, &buf)
	if len(entries) != 1 {
		t.Fatalf("expected exactly one log entry, got %d", len(entries))
	}
	e := entries[0]

	if e["msg"] != "http request" {
		t.Errorf("expected msg %q, got %v", "http request", e["msg"])
	}
	if e["method"] != http.MethodGet {
		t.Errorf("expected method %q, got %v", http.MethodGet, e["method"])
	}
	if e["path"] != "/users" {
		t.Errorf("expected path without query string, got %v", e["path"])
	}
	if strings.Contains(fmt.Sprint(e), "secret") {
		t.Error("query string must not be logged")
	}
	if status, ok := e["status"].(float64); !ok || int(status) != http.StatusOK {
		t.Errorf("expected status 200, got %v", e["status"])
	}
	if bytesField, ok := e["bytes"].(float64); !ok || int(bytesField) != len("hi") {
		t.Errorf("expected bytes %d, got %v", len("hi"), e["bytes"])
	}
	if _, ok := e["duration"]; !ok {
		t.Error("expected duration field")
	}
	if e["remote_addr"] != "192.0.2.1:1234" {
		t.Errorf("expected remote_addr %q, got %v", "192.0.2.1:1234", e["remote_addr"])
	}
	if e["level"] != "INFO" {
		t.Errorf("expected INFO level for 2xx status, got %v", e["level"])
	}
	if _, ok := e["request_id"]; ok {
		t.Error("expected no request_id field when header is absent")
	}
	if _, ok := e["user_agent"]; ok {
		t.Error("expected no user_agent field when header is absent")
	}
}

func TestLogger_RequestIDAndUserAgent(t *testing.T) {
	var buf bytes.Buffer
	captureLogger(&buf, slogger.TraceLevel)
	defer slogger.ResetGlobalLogger()

	mw := replify.Logger()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	req.Header.Set("X-Request-Id", "req-abc")
	req.Header.Set("User-Agent", "test-agent/1.0")

	mw(okHandler("ok")).ServeHTTP(rr, req)

	entries := logLines(t, &buf)
	if len(entries) != 1 {
		t.Fatalf("expected exactly one log entry, got %d", len(entries))
	}
	e := entries[0]
	if e["request_id"] != "req-abc" {
		t.Errorf("expected request_id %q, got %v", "req-abc", e["request_id"])
	}
	if e["user_agent"] != "test-agent/1.0" {
		t.Errorf("expected user_agent %q, got %v", "test-agent/1.0", e["user_agent"])
	}
}

func TestLogger_ImplicitOKWhenNoWriteHeaderCalled(t *testing.T) {
	var buf bytes.Buffer
	captureLogger(&buf, slogger.TraceLevel)
	defer slogger.ResetGlobalLogger()

	mw := replify.Logger()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/implicit", nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("no explicit header"))
	})
	mw(handler).ServeHTTP(rr, req)

	entries := logLines(t, &buf)
	if len(entries) != 1 {
		t.Fatalf("expected exactly one log entry, got %d", len(entries))
	}
	if status, ok := entries[0]["status"].(float64); !ok || int(status) != http.StatusOK {
		t.Errorf("expected implicit status 200, got %v", entries[0]["status"])
	}
}

func TestLogger_HandlerWritesNothing(t *testing.T) {
	var buf bytes.Buffer
	captureLogger(&buf, slogger.TraceLevel)
	defer slogger.ResetGlobalLogger()

	mw := replify.Logger()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/empty", nil)

	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).ServeHTTP(rr, req)

	entries := logLines(t, &buf)
	if len(entries) != 1 {
		t.Fatalf("expected exactly one log entry, got %d", len(entries))
	}
	if status, ok := entries[0]["status"].(float64); !ok || int(status) != http.StatusOK {
		t.Errorf("expected implicit status 200 when nothing is written, got %v", entries[0]["status"])
	}
	if bytesField, ok := entries[0]["bytes"].(float64); !ok || int(bytesField) != 0 {
		t.Errorf("expected 0 bytes written, got %v", entries[0]["bytes"])
	}
}

// --- Logger: status-to-level mapping ------------------------------------------

func TestLogger_StatusLevelMapping(t *testing.T) {
	cases := []struct {
		name   string
		status int
		level  string
	}{
		{"client error", http.StatusBadRequest, "ERROR"},
		{"server error", http.StatusInternalServerError, "ERROR"},
		{"redirect", http.StatusMovedPermanently, "WARN"},
		{"success", http.StatusOK, "INFO"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			captureLogger(&buf, slogger.TraceLevel)
			defer slogger.ResetGlobalLogger()

			mw := replify.Logger()

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/status", nil)

			status := tc.status
			mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
			})).ServeHTTP(rr, req)

			entries := logLines(t, &buf)
			if len(entries) != 1 {
				t.Fatalf("expected exactly one log entry, got %d", len(entries))
			}
			if entries[0]["level"] != tc.level {
				t.Errorf("status %d: expected level %s, got %v", tc.status, tc.level, entries[0]["level"])
			}
		})
	}
}

// TestLogger_SuppressedBelowConfiguredLevel verifies that Logger() honors the
// logger's configured level and skips work entirely (no line emitted) for
// levels below threshold — e.g. a 1xx informational status maps to Debug,
// which an Info-level logger should not emit.
func TestLogger_SuppressedBelowConfiguredLevel(t *testing.T) {
	var buf bytes.Buffer
	captureLogger(&buf, slogger.InfoLevel)
	defer slogger.ResetGlobalLogger()

	mw := replify.Logger()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/informational", nil)

	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusProcessing) // 102, maps to Debug level
	})).ServeHTTP(rr, req)

	if buf.Len() != 0 {
		t.Errorf("expected no log output below the configured level, got %q", buf.String())
	}
}

// --- Logger: WriteHeader/Write semantics --------------------------------------

func TestLogger_WriteHeaderCalledOnceKeepsFirstStatus(t *testing.T) {
	var buf bytes.Buffer
	captureLogger(&buf, slogger.TraceLevel)
	defer slogger.ResetGlobalLogger()

	mw := replify.Logger()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/double", nil)

	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		w.WriteHeader(http.StatusInternalServerError) // net/http semantics: no-op after first call
	})).ServeHTTP(rr, req)

	entries := logLines(t, &buf)
	if len(entries) != 1 {
		t.Fatalf("expected exactly one log entry, got %d", len(entries))
	}
	if status, ok := entries[0]["status"].(float64); !ok || int(status) != http.StatusAccepted {
		t.Errorf("expected first status code %d to be recorded, got %v", http.StatusAccepted, entries[0]["status"])
	}
}

func TestLogger_BytesAccumulateAcrossMultipleWrites(t *testing.T) {
	var buf bytes.Buffer
	captureLogger(&buf, slogger.TraceLevel)
	defer slogger.ResetGlobalLogger()

	mw := replify.Logger()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/chunks", nil)

	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello "))
		_, _ = w.Write([]byte("world"))
	})).ServeHTTP(rr, req)

	entries := logLines(t, &buf)
	if len(entries) != 1 {
		t.Fatalf("expected exactly one log entry, got %d", len(entries))
	}
	want := len("hello world")
	if bytesField, ok := entries[0]["bytes"].(float64); !ok || int(bytesField) != want {
		t.Errorf("expected accumulated bytes %d, got %v", want, entries[0]["bytes"])
	}
}

// --- Logger: duration ----------------------------------------------------------

func TestLogger_DurationReflectsHandlerLatency(t *testing.T) {
	var buf bytes.Buffer
	captureLogger(&buf, slogger.TraceLevel)
	defer slogger.ResetGlobalLogger()

	mw := replify.Logger()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/slow", nil)

	const sleep = 20 * time.Millisecond
	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(sleep)
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	entries := logLines(t, &buf)
	if len(entries) != 1 {
		t.Fatalf("expected exactly one log entry, got %d", len(entries))
	}
	dur, ok := entries[0]["duration"].(string)
	if !ok || dur == "" {
		t.Fatalf("expected non-empty duration string, got %v", entries[0]["duration"])
	}
	parsed, err := time.ParseDuration(dur)
	if err != nil {
		t.Fatalf("duration %q is not parseable: %v", dur, err)
	}
	if parsed < sleep {
		t.Errorf("expected recorded duration >= %s, got %s", sleep, parsed)
	}
}

// --- Composition: Recovery + Logger -------------------------------------------

// TestRecoveryAndLogger_Composition verifies the documented middleware
// ordering (Recovery outside Logger). Because Logger has no recover of its
// own, a panic in the wrapped handler unwinds straight through Logger's
// stack frame to Recovery's deferred handler without ever reaching Logger's
// post-request logging code — so only Recovery's "panic recovered" entry is
// emitted, not Logger's "http request" entry.
func TestRecoveryAndLogger_Composition(t *testing.T) {
	var buf bytes.Buffer
	captureLogger(&buf, slogger.TraceLevel)
	defer slogger.ResetGlobalLogger()

	handler := replify.Recovery()(replify.Logger()(panicHandler("kaboom")))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/composed", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}

	entries := logLines(t, &buf)
	var sawRequestLog, sawRecoveryLog bool
	for _, e := range entries {
		switch e["msg"] {
		case "http request":
			sawRequestLog = true
		case "panic recovered":
			sawRecoveryLog = true
		}
	}
	if sawRequestLog {
		t.Error("did not expect an \"http request\" log entry when the handler panics")
	}
	if !sawRecoveryLog {
		t.Error("expected a \"panic recovered\" log entry from Recovery")
	}
}

func TestRecoveryAndLogger_Composition_NoPanic(t *testing.T) {
	var buf bytes.Buffer
	captureLogger(&buf, slogger.TraceLevel)
	defer slogger.ResetGlobalLogger()

	handler := replify.Recovery()(replify.Logger()(okHandler("fine")))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/composed-ok", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if rr.Body.String() != "fine" {
		t.Fatalf("expected body %q, got %q", "fine", rr.Body.String())
	}

	entries := logLines(t, &buf)
	if len(entries) != 1 {
		t.Fatalf("expected exactly one log entry for a non-panicking request, got %d", len(entries))
	}
	if entries[0]["msg"] != "http request" {
		t.Errorf("expected \"http request\" log entry, got %v", entries[0]["msg"])
	}
}
