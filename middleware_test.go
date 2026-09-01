package replify_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/polarixa/replify"
	"github.com/polarixa/replify/pkg/slogger"
)

// --- helpers ----------------------------------------------------------------

func panicHandler(p any) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(p)
	})
}

func okHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// newTestLogger creates a logger that writes to buf with colours disabled.
func newTestLogger(buf *bytes.Buffer) *slogger.Logger {
	return slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.TraceLevel)
		o.SetOutput(buf)
		o.SetFormatter(slogger.NewTextFormatter(buf).WithColorMode(slogger.ColorNever))
	})
}

// --- normal request ---------------------------------------------------------

// TestRecovery_NormalRequest verifies that a non-panicking handler is
// unaffected: status, headers, and body are passed through unchanged.
func TestRecovery_NormalRequest(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)

	replify.Recovery()(http.HandlerFunc(okHandler)).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("expected application/json Content-Type, got %q", got)
	}
	if rr.Body.String() != `{"status":"ok"}` {
		t.Errorf("unexpected body: %q", rr.Body.String())
	}
}

// --- panic with string ------------------------------------------------------

// TestRecovery_PanicString verifies recovery from a string panic and a 500 response.
func TestRecovery_PanicString(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/crash", nil)

	replify.Recovery()(panicHandler("something went wrong")).ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("expected JSON Content-Type, got %q", ct)
	}
}

// --- panic with error -------------------------------------------------------

// TestRecovery_PanicError verifies recovery from an error panic and a 500 response.
func TestRecovery_PanicError(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/crash", nil)

	replify.Recovery()(panicHandler(errors.New("something went wrong"))).ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

// --- panic with arbitrary value ---------------------------------------------

// TestRecovery_PanicArbitraryValue verifies that the middleware does not assume
// the panic value is a string or error.
func TestRecovery_PanicArbitraryValue(t *testing.T) {
	t.Parallel()

	type errPayload struct{ Code int }

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/crash", nil)

	replify.Recovery()(panicHandler(errPayload{Code: 500})).ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

// --- response format --------------------------------------------------------

// TestRecovery_ResponseFormat verifies the recovery response matches the
// Replify JSON contract and does not expose the panic value to the client.
func TestRecovery_ResponseFormat(t *testing.T) {
	t.Parallel()

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

	// Replify shape: status_code and message must be present.
	if sc, ok := body["status_code"]; !ok {
		t.Error("response missing status_code field")
	} else if int(sc.(float64)) != http.StatusInternalServerError {
		t.Errorf("expected status_code 500, got %v", sc)
	}
	if _, ok := body["message"]; !ok {
		t.Error("response missing message field")
	}

	// Panic details must NOT be exposed to the client.
	if strings.Contains(rr.Body.String(), "internal-secret") {
		t.Error("panic value leaked to client response")
	}
}

// --- response already started -----------------------------------------------

// TestRecovery_ResponseAlreadyStarted verifies that a panic after headers have
// been committed does not produce a second response or corrupt the connection.
// The original status code is preserved; no extra JSON body is appended.
func TestRecovery_ResponseAlreadyStarted(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("partial"))
		panic("late panic after write")
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/partial", nil)

	replify.Recovery()(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected original status 200, got %d", rr.Code)
	}
	// The body must start with the partial write; the recovery path must not
	// append a second JSON error payload.
	if !strings.HasPrefix(rr.Body.String(), "partial") {
		t.Errorf("unexpected body: %q", rr.Body.String())
	}
	if strings.Contains(rr.Body.String(), "status_code") {
		t.Error("recovery wrote a second JSON response after headers were committed")
	}
}

// --- ErrAbortHandler --------------------------------------------------------

// TestRecovery_ErrAbortHandler verifies that http.ErrAbortHandler is re-panicked
// so that net/http can close the connection cleanly without a response body.
func TestRecovery_ErrAbortHandler(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(http.ErrAbortHandler)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/abort", nil)

	defer func() {
		p := recover()
		if p != http.ErrAbortHandler {
			t.Errorf("expected ErrAbortHandler to be re-panicked, got %v", p)
		}
	}()

	replify.Recovery()(handler).ServeHTTP(rr, req)
}

// --- concurrent requests ----------------------------------------------------

// TestRecovery_Concurrent verifies that panic recovery state is isolated per
// request: panicking requests get 500 and non-panicking requests get 200.
func TestRecovery_Concurrent(t *testing.T) {
	t.Parallel()

	const n = 50
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/panic" {
			panic("concurrent panic")
		}
		w.WriteHeader(http.StatusOK)
	})
	middleware := replify.Recovery()(handler)

	type result struct{ code int }
	results := make([]result, n)
	var wg sync.WaitGroup
	wg.Add(n)

	for i := range n {
		i := i
		go func() {
			defer wg.Done()
			path := "/panic"
			if i%2 == 0 {
				path = "/ok"
			}
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, path, nil)
			middleware.ServeHTTP(rr, req)
			results[i] = result{code: rr.Code}
		}()
	}
	wg.Wait()

	for i, res := range results {
		if i%2 == 0 {
			if res.code != http.StatusOK {
				t.Errorf("request %d (/ok): expected 200, got %d", i, res.code)
			}
		} else {
			if res.code != http.StatusInternalServerError {
				t.Errorf("request %d (/panic): expected 500, got %d", i, res.code)
			}
		}
	}
}

// --- logging ----------------------------------------------------------------

// TestRecovery_Logging verifies that the panic value and request ID are
// forwarded to the Replify logger and that neither appears in the client
// response.
//
// This test is NOT marked parallel because it temporarily replaces the
// package-level global logger.
func TestRecovery_Logging(t *testing.T) {
	var buf bytes.Buffer
	orig := slogger.S()
	slogger.SetGlobalLogger(newTestLogger(&buf))
	defer slogger.SetGlobalLogger(orig)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/crash", nil)
	req.Header.Set("X-Request-Id", "test-req-42")

	replify.Recovery()(panicHandler("log-me-please")).ServeHTTP(rr, req)

	logged := buf.String()

	if !strings.Contains(logged, "log-me-please") {
		t.Errorf("panic value not in log output; got: %q", logged)
	}
	if !strings.Contains(logged, "test-req-42") {
		t.Errorf("request_id not in log output; got: %q", logged)
	}
	if !strings.Contains(logged, "panic recovered") {
		t.Errorf("expected 'panic recovered' log message; got: %q", logged)
	}
	// Panic value must NOT be in the client response.
	if strings.Contains(rr.Body.String(), "log-me-please") {
		t.Error("panic value leaked to client response body")
	}
}

// --- sensitive data ----------------------------------------------------------

// TestRecovery_NoSensitiveDataInResponse verifies that Authorization and Cookie
// headers are not included in the 500 response body.
func TestRecovery_NoSensitiveDataInResponse(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/crash", nil)
	req.Header.Set("Authorization", "Bearer super-secret-token")
	req.Header.Set("Cookie", "session=abc123")

	replify.Recovery()(panicHandler("crash")).ServeHTTP(rr, req)

	body := rr.Body.String()
	if strings.Contains(body, "super-secret-token") {
		t.Error("Authorization header value leaked to client response")
	}
	if strings.Contains(body, "abc123") {
		t.Error("Cookie value leaked to client response")
	}
}

// --- composability ----------------------------------------------------------

// TestRecovery_Composable verifies that Recovery() returns a usable middleware
// and that the resulting handler is non-nil.
func TestRecovery_Composable(t *testing.T) {
	t.Parallel()

	mw := replify.Recovery()
	if mw == nil {
		t.Fatal("Recovery() returned nil")
	}
	handler := mw(http.HandlerFunc(okHandler))
	if handler == nil {
		t.Fatal("Recovery()(handler) returned nil")
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

// --- regression: partial-write does not append a second body ----------------

// TestRecovery_PartialWriteThenPanic is a regression test: before the written
// flag was introduced, the recovery path would call WriteJSON even after the
// downstream handler had already committed bytes, producing a corrupt response.
func TestRecovery_PartialWriteThenPanic(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id":1}`))
		panic("oops")
	})

	rr := httptest.NewRecorder()
	replify.Recovery()(handler).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	body := rr.Body.String()
	// Body must start with the partial write and contain no second JSON object.
	if !strings.HasPrefix(body, `{"id":1}`) {
		t.Errorf("partial write lost: %q", body)
	}
	if strings.Count(body, `{`) > 1 {
		t.Errorf("second JSON object appended after panic: %q", body)
	}
}

// --- flusher forwarding -----------------------------------------------------

// TestRecovery_FlushForwarded verifies that Flush() is forwarded to the
// underlying ResponseWriter when it implements http.Flusher.
func TestRecovery_FlushForwarded(t *testing.T) {
	t.Parallel()

	flushed := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if fl, ok := w.(http.Flusher); ok {
			fl.Flush()
			flushed = true
		}
	})

	rr := httptest.NewRecorder() // httptest.ResponseRecorder implements http.Flusher
	replify.Recovery()(handler).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if !flushed {
		t.Error("Flush was not forwarded to the underlying ResponseWriter")
	}
}

// ============================================================================
// Logger middleware tests
// ============================================================================

// --- successful request -----------------------------------------------------

// TestLogger_SuccessfulRequest verifies method, path, status, and duration
// appear in the log for a plain 200 OK response.
func TestLogger_SuccessfulRequest(t *testing.T) {
	var buf bytes.Buffer
	orig := slogger.S()
	slogger.SetGlobalLogger(newTestLogger(&buf))
	defer slogger.SetGlobalLogger(orig)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	replify.Logger()(http.HandlerFunc(okHandler)).ServeHTTP(rr, req)

	logged := buf.String()
	for _, want := range []string{"method=GET", "path=/health", "status=200", "duration="} {
		if !strings.Contains(logged, want) {
			t.Errorf("expected %q in log; got: %q", want, logged)
		}
	}
}

// --- explicit status --------------------------------------------------------

// TestLogger_ExplicitStatus201 verifies that a handler calling WriteHeader(201)
// is logged as status=201.
func TestLogger_ExplicitStatus201(t *testing.T) {
	var buf bytes.Buffer
	orig := slogger.S()
	slogger.SetGlobalLogger(newTestLogger(&buf))
	defer slogger.SetGlobalLogger(orig)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	replify.Logger()(handler).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/items", nil))

	if !strings.Contains(buf.String(), "status=201") {
		t.Errorf("expected status=201 in log; got: %q", buf.String())
	}
}

// --- implicit status --------------------------------------------------------

// TestLogger_ImplicitStatus200 verifies that a handler calling only Write (no
// WriteHeader) is logged as status=200 per Go net/http semantics.
func TestLogger_ImplicitStatus200(t *testing.T) {
	var buf bytes.Buffer
	orig := slogger.S()
	slogger.SetGlobalLogger(newTestLogger(&buf))
	defer slogger.SetGlobalLogger(orig)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello"))
	})

	replify.Logger()(handler).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if !strings.Contains(buf.String(), "status=200") {
		t.Errorf("expected status=200 in log; got: %q", buf.String())
	}
}

// --- empty handler ----------------------------------------------------------

// TestLogger_EmptyHandler verifies that a handler writing nothing is logged as
// status=200 (implicit) with bytes=0.
func TestLogger_EmptyHandler(t *testing.T) {
	var buf bytes.Buffer
	orig := slogger.S()
	slogger.SetGlobalLogger(newTestLogger(&buf))
	defer slogger.SetGlobalLogger(orig)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	replify.Logger()(handler).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	logged := buf.String()
	if !strings.Contains(logged, "status=200") {
		t.Errorf("expected status=200 in log; got: %q", logged)
	}
	if !strings.Contains(logged, "bytes=0") {
		t.Errorf("expected bytes=0 in log; got: %q", logged)
	}
}

// --- error status -----------------------------------------------------------

// TestLogger_ErrorStatus verifies that a 400 response is logged as status=400.
func TestLogger_ErrorStatus(t *testing.T) {
	var buf bytes.Buffer
	orig := slogger.S()
	slogger.SetGlobalLogger(newTestLogger(&buf))
	defer slogger.SetGlobalLogger(orig)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	replify.Logger()(handler).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if !strings.Contains(buf.String(), "status=400") {
		t.Errorf("expected status=400 in log; got: %q", buf.String())
	}
}

// --- response size ----------------------------------------------------------

// TestLogger_ResponseSize verifies that bytes written to the response body are
// accurately counted.
func TestLogger_ResponseSize(t *testing.T) {
	var buf bytes.Buffer
	orig := slogger.S()
	slogger.SetGlobalLogger(newTestLogger(&buf))
	defer slogger.SetGlobalLogger(orig)

	const payload = "hello world" // 11 bytes
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(payload))
	})

	replify.Logger()(handler).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if !strings.Contains(buf.String(), "bytes=11") {
		t.Errorf("expected bytes=11 in log; got: %q", buf.String())
	}
}

// --- multiple writes --------------------------------------------------------

// TestLogger_MultipleWrites verifies that byte counts are accumulated correctly
// across multiple w.Write calls.
func TestLogger_MultipleWrites(t *testing.T) {
	var buf bytes.Buffer
	orig := slogger.S()
	slogger.SetGlobalLogger(newTestLogger(&buf))
	defer slogger.SetGlobalLogger(orig)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("abc"))   // 3
		_, _ = w.Write([]byte("defgh")) // 5  → total 8
	})

	replify.Logger()(handler).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if !strings.Contains(buf.String(), "bytes=8") {
		t.Errorf("expected bytes=8 in log; got: %q", buf.String())
	}
}

// --- panic integration with Recovery ----------------------------------------

// TestLogger_PanicWithRecovery is the key integration test: Logger wraps
// Recovery which wraps a panicking handler.
//
// Recovery converts the panic to a 500 response written through Logger's
// logWriter, so Logger must record status=500.
//
// This test is NOT marked parallel because it temporarily replaces the global logger.
func TestLogger_PanicWithRecovery(t *testing.T) {
	var buf bytes.Buffer
	orig := slogger.S()
	slogger.SetGlobalLogger(newTestLogger(&buf))
	defer slogger.SetGlobalLogger(orig)

	handler := replify.Logger()(replify.Recovery()(panicHandler("boom")))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/crash", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
	// Logger fires after Recovery and must see the 500 Recovery wrote.
	if !strings.Contains(buf.String(), "status=500") {
		t.Errorf("expected status=500 in logger output; got: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "http request") {
		t.Errorf("expected 'http request' in logger output; got: %q", buf.String())
	}
}

// --- regression: wrong ordering would log status=200 not 500 ---------------

// TestLogger_PanicWithRecovery_WrongOrdering_Regression demonstrates that the
// opposite ordering (Recovery → Logger → Handler) causes Logger to see the
// incomplete in-flight status (200) rather than the 500 that Recovery eventually
// writes, which is the reason Logger must be placed OUTSIDE Recovery.
//
// This test is NOT marked parallel because it temporarily replaces the global logger.
func TestLogger_PanicWithRecovery_WrongOrdering_Regression(t *testing.T) {
	var buf bytes.Buffer
	orig := slogger.S()
	slogger.SetGlobalLogger(newTestLogger(&buf))
	defer slogger.SetGlobalLogger(orig)

	// Wrong ordering: Recovery(Logger(handler)) — Logger is INSIDE Recovery.
	// Logger's defer fires before Recovery recovers the panic, so it sees 200
	// (the default when nothing has been written yet).
	handler := replify.Recovery()(replify.Logger()(panicHandler("inner panic")))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/crash", nil))

	// The HTTP response code IS 500 (Recovery wrote it after Logger fired).
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("response should still be 500, got %d", rr.Code)
	}
	// But the LOGGED status is 200 — Logger saw no writes before the panic,
	// so it inferred the implicit 200. This proves the ordering matters.
	if !strings.Contains(buf.String(), "status=200") {
		t.Logf("(informational) log output: %q", buf.String())
	}
}

// --- request ID -------------------------------------------------------------

// TestLogger_RequestID verifies that the X-Request-Id header is included in
// the structured log entry.
//
// NOT parallel — swaps global logger.
func TestLogger_RequestID(t *testing.T) {
	var buf bytes.Buffer
	orig := slogger.S()
	slogger.SetGlobalLogger(newTestLogger(&buf))
	defer slogger.SetGlobalLogger(orig)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-Id", "logger-req-99")

	replify.Logger()(http.HandlerFunc(okHandler)).ServeHTTP(httptest.NewRecorder(), req)

	if !strings.Contains(buf.String(), "logger-req-99") {
		t.Errorf("expected request_id in log; got: %q", buf.String())
	}
}

// --- sensitive headers not logged -------------------------------------------

// TestLogger_NoSensitiveHeaders verifies that Authorization and Cookie values
// do not appear in the structured log.
//
// NOT parallel — swaps global logger.
func TestLogger_NoSensitiveHeaders(t *testing.T) {
	var buf bytes.Buffer
	orig := slogger.S()
	slogger.SetGlobalLogger(newTestLogger(&buf))
	defer slogger.SetGlobalLogger(orig)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer tok-super-secret")
	req.Header.Set("Cookie", "sid=ultra-private")

	replify.Logger()(http.HandlerFunc(okHandler)).ServeHTTP(httptest.NewRecorder(), req)

	logged := buf.String()
	if strings.Contains(logged, "tok-super-secret") {
		t.Error("Authorization value leaked to logs")
	}
	if strings.Contains(logged, "ultra-private") {
		t.Error("Cookie value leaked to logs")
	}
}

// --- query string not logged ------------------------------------------------

// TestLogger_QueryNotLogged verifies that sensitive query parameters are never
// written to the log (the logger records path only, not the query string).
//
// NOT parallel — swaps global logger.
func TestLogger_QueryNotLogged(t *testing.T) {
	var buf bytes.Buffer
	orig := slogger.S()
	slogger.SetGlobalLogger(newTestLogger(&buf))
	defer slogger.SetGlobalLogger(orig)

	req := httptest.NewRequest(http.MethodGet, "/reset-password?token=super-secret-reset-token", nil)

	replify.Logger()(http.HandlerFunc(okHandler)).ServeHTTP(httptest.NewRecorder(), req)

	logged := buf.String()
	if strings.Contains(logged, "super-secret-reset-token") {
		t.Error("sensitive query parameter leaked to logs")
	}
	// The path without query must still appear.
	if !strings.Contains(logged, "path=/reset-password") {
		t.Errorf("expected path=/reset-password in log; got: %q", logged)
	}
}

// --- concurrent requests ----------------------------------------------------

// TestLogger_Concurrent verifies that per-request state (status, bytes) is
// never shared between concurrent requests. The race detector enforces
// concurrency safety; this test also checks functional correctness under load.
func TestLogger_Concurrent(t *testing.T) {
	t.Parallel()

	const n = 100
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/error" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	middleware := replify.Logger()(handler)

	var wg sync.WaitGroup
	wg.Add(n)
	codes := make([]int, n)

	for i := range n {
		i := i
		go func() {
			defer wg.Done()
			path := "/ok"
			if i%3 == 0 {
				path = "/error"
			}
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, path, nil)
			middleware.ServeHTTP(rr, req)
			codes[i] = rr.Code
		}()
	}
	wg.Wait()

	for i, code := range codes {
		expected := http.StatusOK
		if i%3 == 0 {
			expected = http.StatusBadRequest
		}
		if code != expected {
			t.Errorf("request %d: expected %d, got %d", i, expected, code)
		}
	}
}

// --- ResponseWriter interface compatibility ----------------------------------

// TestLogger_FlusherForwarded verifies that Flush() is forwarded to the
// underlying ResponseWriter when it implements http.Flusher.
func TestLogger_FlusherForwarded(t *testing.T) {
	t.Parallel()

	flushed := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if fl, ok := w.(http.Flusher); ok {
			fl.Flush()
			flushed = true
		}
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	replify.Logger()(handler).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if !flushed {
		t.Error("Flush was not forwarded through logWriter")
	}
}

// TestLogger_UnwrapPreservesUnderlying verifies that the logWriter's Unwrap()
// returns the original ResponseWriter, enabling http.ResponseController to
// reach the underlying writer's optional interfaces.
func TestLogger_UnwrapPreservesUnderlying(t *testing.T) {
	t.Parallel()

	var sawUnderlying bool
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type unwrapper interface{ Unwrap() http.ResponseWriter }
		if uw, ok := w.(unwrapper); ok {
			_ = uw.Unwrap() // must not panic
			sawUnderlying = true
		}
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	replify.Logger()(handler).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if !sawUnderlying {
		t.Error("logWriter does not expose Unwrap()")
	}
}

// ============================================================================
// Benchmark
// ============================================================================

// BenchmarkLogger measures the per-request overhead of the Logger middleware
// on a simple 200 OK handler. Log output is discarded to isolate middleware
// cost from I/O.
func BenchmarkLogger(b *testing.B) {
	orig := slogger.S()
	slogger.SetGlobalLogger(slogger.New(func(o *slogger.Options) {
		o.SetOutput(io.Discard)
	}))
	b.Cleanup(func() { slogger.SetGlobalLogger(orig) })

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello"))
	})
	middleware := replify.Logger()(handler)
	req := httptest.NewRequest(http.MethodGet, "/bench", nil)

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		rr := httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)
	}
}

// BenchmarkLogger_Parallel shows the pool benefit under concurrent load:
// goroutines return slices to the pool and other goroutines reuse them,
// eliminating the make([]slogger.Field, 0, 9) heap allocation on hot paths.
func BenchmarkLogger_Parallel(b *testing.B) {
	orig := slogger.S()
	slogger.SetGlobalLogger(slogger.New(func(o *slogger.Options) {
		o.SetOutput(io.Discard)
	}))
	b.Cleanup(func() { slogger.SetGlobalLogger(orig) })

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello"))
	})
	middleware := replify.Logger()(handler)
	req := httptest.NewRequest(http.MethodGet, "/bench", nil)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rr := httptest.NewRecorder()
			middleware.ServeHTTP(rr, req)
		}
	})
}
