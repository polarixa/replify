package replify

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"sync"
	"time"

	"github.com/polarixa/replify/pkg/slogger"
	"github.com/polarixa/replify/pkg/strutil"
)

// Recovery returns a net/http middleware that catches any panic in a downstream
// handler, logs diagnostic context using the Replify slogger infrastructure,
// and writes a 500 Internal Server Error response in the Replify JSON format
// when the response has not already started.
//
// Panics with [http.ErrAbortHandler] are re-panicked so that net/http can close
// the connection cleanly; that sentinel value signals an intentional abort, not
// an application error.
//
// The [http.ResponseWriter] is wrapped with a minimal tracker that records
// whether WriteHeader or Write has been called. This prevents a double-write
// when the downstream handler has already committed part of the response before
// the panic. The wrapper implements [http.Flusher] forwarding and exposes
// Unwrap() for [http.ResponseController] compatibility (Go 1.20+).
//
// Sensitive request headers (Authorization, Cookie) are never logged or
// included in the client-facing response.
//
// Usage:
//
//	http.Handle("/api/users", Recovery()(myHandler))
//
// Composable with other middleware:
//
//	handler := Recovery()(authMiddleware(myHandler))
func Recovery() func(http.Handler) http.Handler {
	l := slogger.S().Named("replify::recovery") // cached per-middleware-instance
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw := &recoveryWriter{ResponseWriter: w}
			defer func() {
				p := recover()
				if p == nil {
					return
				}
				// http.ErrAbortHandler is a net/http sentinel. Re-panic so
				// the server closes the connection without a response body.
				if p == http.ErrAbortHandler {
					panic(p)
				}
				stack := debug.Stack()
				rw.logRecoveredPanic(l, r, p, stack)
				if !rw.written {
					// Only the minimal, public-safe Issue reaches the client;
					// the raw panic value and full stack stay server-side in
					// the structured log emitted above.
					New().
						InternalServerError().
						WithMessage("an unexpected error occurred").
						WithIssue(NewPanicIssue(p)).
						WriteJSON(w)
				}
			}()
			next.ServeHTTP(rw, r)
		})
	}
}

// recoveryWriter wraps [http.ResponseWriter] to track whether WriteHeader or
// Write has been called before a downstream panic. The written flag prevents
// the recovery path from attempting a second response after the connection has
// already been committed.
//
// Unwrap() enables [http.ResponseController] (Go 1.20+) and middleware chains
// that use the Unwrap convention to reach the underlying writer.
type recoveryWriter struct {
	http.ResponseWriter
	written bool
}

// WriteHeader records that a response has been written and forwards the call to the underlying [http.ResponseWriter].
func (rw *recoveryWriter) WriteHeader(code int) {
	rw.written = true
	rw.ResponseWriter.WriteHeader(code)
}

// Write records that a response has been written and forwards the call to the underlying [http.ResponseWriter].
func (rw *recoveryWriter) Write(b []byte) (int, error) {
	rw.written = true
	return rw.ResponseWriter.Write(b)
}

// Flush implements [http.Flusher] by forwarding to the underlying writer when it
// supports flushing. This preserves streaming/SSE behaviour for handlers that
// type-assert [http.Flusher] directly rather than going through
// [http.ResponseController].
func (rw *recoveryWriter) Flush() {
	if fl, ok := rw.ResponseWriter.(http.Flusher); ok {
		fl.Flush()
	}
}

// Unwrap returns the underlying [http.ResponseWriter] for [http.ResponseController] and
// middleware chains that use the Unwrap convention (Go 1.20+).
func (rw *recoveryWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

// logRecoveredPanic emits a structured error entry using the supplied logger.
// Sensitive headers (Authorization, Cookie) are deliberately excluded.
func (rw *recoveryWriter) logRecoveredPanic(l *slogger.Logger, r *http.Request, p any, stack []byte) {
	fb := fieldsPool.Get().(*fieldsBuf)
	fields := fb.v[:0]
	if id := r.Header.Get("X-Request-Id"); strutil.IsNotEmpty(id) {
		fields = append(fields, slogger.String("request_id", id))
	}
	fields = append(fields,
		slogger.String("method", r.Method),
		slogger.String("path", r.URL.Path),
		slogger.String("remote_addr", r.RemoteAddr),
		slogger.String("panic", fmt.Sprintf("%v", p)),
		slogger.String("stack", string(stack)),
	)
	l.Error("panic recovered", fields...)
	fb.v = fields
	fieldsPool.Put(fb)
}

// Logger returns a net/http middleware that emits one structured log entry per
// request after the downstream handler returns. The entry is logged at a level
// derived from the response status code, using the same [httpStatusLevel]
// convention used across the Replify logging infrastructure:
// 1xx→Debug, 2xx→Info, 3xx→Warn, 4xx/5xx→Error.
//
// Fields emitted per request:
//
//   - request_id  — from the X-Request-Id header, when present
//   - method      — HTTP method
//   - path        — URL path only; the query string is excluded to prevent
//     accidental logging of sensitive query parameters (e.g. ?token=…)
//   - proto       — HTTP protocol version (HTTP/1.1, HTTP/2.0, …)
//   - status      — final HTTP response status code
//   - bytes       — bytes written to the response body
//   - duration    — total request duration from receipt to handler return
//   - remote_addr — network address from r.RemoteAddr (never spoofable)
//   - user_agent  — User-Agent header value, when present
//
// Client IP is taken from r.RemoteAddr. X-Forwarded-For and X-Real-IP are
// intentionally not read because they can be spoofed when the application is
// not behind a trusted proxy. Applications operating behind a known-good proxy
// should strip or validate those headers upstream.
//
// # Middleware ordering
//
// Place Logger outside Recovery so that panics are converted to 500 responses
// before the logger records the final status:
//
//	Recovery → Logger → Application Handler
//
// In code:
//
//	handler := replify.Recovery()(replify.Logger()(myHandler))
//
// This ensures that when Recovery writes a 500 response it does so through
// Logger's [logWriter], giving the logger an accurate view of the final status
// and byte count.
func Logger() func(http.Handler) http.Handler {
	l := slogger.S().Named("replify::logger") // cached per-middleware-instance
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lw := &logWriter{ResponseWriter: w}
			start := time.Now()

			next.ServeHTTP(lw, r)

			code := lw.status
			if code == 0 {
				code = http.StatusOK // implicit 200 when handler wrote nothing
			}
			lvl := httpStatusLevel(code)
			if !l.IsLevelEnabled(lvl) {
				return
			}

			fb := fieldsPool.Get().(*fieldsBuf)
			fields := fb.v[:0]
			if id := r.Header.Get("X-Request-Id"); strutil.IsNotEmpty(id) {
				fields = append(fields, slogger.String("request_id", id))
			}
			if ua := r.Header.Get("User-Agent"); strutil.IsNotEmpty(ua) {
				fields = append(fields, slogger.String("user_agent", ua))
			}
			fields = append(fields,
				slogger.String("method", r.Method),
				slogger.String("path", r.URL.Path),
				slogger.String("proto", r.Proto),
				slogger.Int("status", code),
				slogger.Int("bytes", lw.bytes),
				slogger.Duration("duration", time.Since(start)),
				slogger.String("remote_addr", r.RemoteAddr),
			)
			logFieldsAtLevel(l, lvl, "http request", fields)
			fb.v = fields // write back in case append grew the slice
			fieldsPool.Put(fb)
		})
	}
}

// fieldsBuf holds a reusable []slogger.Field slice. Keeping it in a sync.Pool
// avoids one heap allocation per request in the Logger hot path.
type fieldsBuf struct{ v []slogger.Field }

// fieldsPool stores fieldsBuf instances between requests.
// Capacity 9 covers all Logger fields (8 fixed + optional user_agent).
var fieldsPool = sync.Pool{
	New: func() any { return &fieldsBuf{v: make([]slogger.Field, 0, 9)} },
}

// logWriter wraps [http.ResponseWriter] to capture the HTTP status code and the
// total number of bytes written. Both fields are per-request locals; there is
// no shared state between concurrent requests.
//
// Default status is 0 (not written); the Logger interprets 0 as an implicit
// 200 OK per Go net/http semantics.
type logWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

// WriteHeader captures the status code on the first call only, matching
// net/http semantics where subsequent WriteHeader calls after the first are
// no-ops.
func (lw *logWriter) WriteHeader(code int) {
	if lw.status == 0 {
		lw.status = code
	}
	lw.ResponseWriter.WriteHeader(code)
}

// Write marks the implicit 200 status when WriteHeader has not been called yet,
// then accumulates the byte count.
func (lw *logWriter) Write(b []byte) (int, error) {
	if lw.status == 0 {
		lw.status = http.StatusOK
	}
	n, err := lw.ResponseWriter.Write(b)
	lw.bytes += n
	return n, err
}

// Flush implements [http.Flusher] by forwarding to the underlying writer when it
// supports flushing.
func (lw *logWriter) Flush() {
	if fl, ok := lw.ResponseWriter.(http.Flusher); ok {
		fl.Flush()
	}
}

// Unwrap returns the underlying [http.ResponseWriter] for [http.ResponseController]
// and middleware chains that use the Unwrap convention (Go 1.20+).
func (lw *logWriter) Unwrap() http.ResponseWriter {
	return lw.ResponseWriter
}
