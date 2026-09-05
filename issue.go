package replify

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/polarixa/replify/pkg/randn"
	"github.com/polarixa/replify/pkg/strutil"
)

// stackTracer is satisfied by any error carrying an embedded call stack, i.e.
// one created via [NewError], [NewErrorf], [NewErrorAck], [NewErrorAckf],
// [AppendErrorAck], or [AppendErrorAckf]. It mirrors the interface used by
// [wrapper.StackTrace].
type stackTracer interface {
	StackTrace() StackTrace
}

// NewIssueID generates a short, collision-resistant, publicly-safe occurrence
// identifier in the form "ISS-XXXXXXXX" (8 uppercase hex characters).
//
// Algorithm: 4 bytes read from crypto/rand ([randn.RandIDHex]), hex-encoded.
// crypto/rand is preferred over math/rand because occurrence IDs are
// frequently pasted into support tickets and logs, sometimes across trust
// boundaries; a predictable generator would let one leaked ID be used to
// guess adjacent ones. No external module is required — crypto/rand and
// encoding/hex are both standard library.
//
// Collision resistance: 4 bytes yield 2^32 possible values. By the birthday
// bound, a 50% collision probability is only reached after roughly 77,000
// generated IDs. That is a deliberate trade-off: occurrence IDs exist for
// human-driven incident correlation (searching logs, quoting in a ticket),
// not for cryptographic uniqueness, so 8 characters keeps them short and
// easy to read/transcribe while still being safe for any realistic
// single-service error volume. Services expecting far higher volumes should
// pair the ID with the request timestamp (already available via [meta]) to
// further reduce ambiguity.
func NewIssueID() string {
	return fmt.Sprintf("ISS-%s", strings.ToUpper(randn.RandIDHex(4)))
}

// NewFingerprint computes a stable "IFP-XXXXXX" fingerprint (6 uppercase hex
// characters) identifying a category of failure. The same (kind, function,
// file, line) tuple always produces the same fingerprint, letting dashboards
// and alerting group occurrences of the same underlying failure together.
//
// Hashing algorithm: SHA-256 over "kind|function|file|line". SHA-256 is used
// instead of a non-cryptographic hash (FNV, CRC32) because it has no known
// clustering patterns for short, structured, near-duplicate inputs (e.g. two
// call sites that differ only in line number) — a fast non-cryptographic
// hash can produce visibly correlated output for such inputs, which would
// undermine the fingerprint's grouping guarantee. SHA-256 is also part of
// the Go standard library, keeping this zero-dependency.
//
// Truncation strategy: only the first 3 bytes (24 bits) of the digest are
// kept, hex-encoded to 6 characters. This is a deliberate space/readability
// trade-off — fingerprints are meant to be scanned by humans in dashboards,
// not compared byte-for-byte for security purposes.
//
// Collision considerations: with a 24-bit output, 50% collision probability
// is reached at roughly 4,800 distinct fingerprints (birthday bound). A
// typical service has, at most, a few hundred distinct panic/error call
// sites, so the practical collision risk is low; a collision would merge two
// unrelated failure categories in a dashboard, which is a graceful, low-harm
// degradation (never a security or correctness issue, since the fingerprint
// is never used to resolve or dedupe request-level data — [issue.ID] carries
// occurrence identity).
//
// Performance: SHA-256 over a short (<100 byte) string is sub-microsecond
// and allocation-light; safe to compute on every error path, including hot
// request paths.
func NewFingerprint(kind, function, file string, line int) string {
	form := fmt.Sprintf("%s|%s|%s|%d", kind, function, file, line)
	sum := sha256.Sum256([]byte(form))
	return fmt.Sprintf("IFP-%s", strings.ToUpper(hex.EncodeToString(sum[:3])))
}

// NewIssue builds an API-facing [issue] from an arbitrary error. The message
// resolves to the root cause via [rootCause] (supporting errors.Unwrap chains
// from both this package and the standard fmt.Errorf("...: %w", err)
// convention); the fingerprint groups by the error's own captured call site
// when available, falling back to its concrete type otherwise.
//
// Returns nil when err is nil.
func NewIssue(err error) *issue {
	if err == nil {
		return nil
	}
	function, file, line := errorOrigin(err)
	kind := reflect.TypeOf(rootCause(err)).String()
	return &issue{
		ID:          NewIssueID(),
		Fingerprint: NewFingerprint(kind, function, file, line),
		Message:     rootCause(err).Error(),
	}
}

// NewPanicIssue builds an API-facing [issue] from a recovered panic value
// (the return of `recover()`). It must be called from within the deferred
// function that performed the recover; see [panicOrigin] for how the
// original panic site is located regardless of intervening helper calls.
func NewPanicIssue(recovered any) *issue {
	function, file, line := panicOrigin()
	return &issue{
		ID:          NewIssueID(),
		Fingerprint: NewFingerprint("panic", function, file, line),
		Message:     fmt.Sprintf("panic: %v", recovered),
	}
}

// Issue returns the API-facing [issue] for this [wrapper]'s current error, or
// nil when no error is present. Use this — never the internal `errors`
// field — when surfacing failure details to API consumers.
func (w *wrapper) Issue() *issue {
	if !w.Available() || !w.IsError() {
		return nil
	}
	w.autoAdjust()
	if w.errors == nil {
		return nil
	}
	return NewIssue(w.errors)
}

// WithIssue computes the [issue] for this [wrapper]'s current error (if any)
// and attaches it to the debug map under the "issue" key, ready to be
// serialized to API consumers in place of the raw internal error. It is a
// no-op when no error is present.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) WithIssue() *wrapper {
	issue := w.Issue()
	if issue == nil {
		return w
	}
	w.issue = issue
	return w
}

// rootCause unwraps err as far as possible using the standard errors.Unwrap
// chain and returns the innermost error. It is used to resolve [issue.Message]
// to the original failure rather than whatever contextual wrapping was
// layered on top (e.g. `fmt.Errorf("query user: %w", sql.ErrNoRows)` resolves
// to sql.ErrNoRows).
//
// Because Replify's own wrapping types ([underlyingStack], [underlyingMessage])
// implement Unwrap(), this walk transparently supports both Replify-created
// errors and errors created with the standard fmt.Errorf("...: %w", err)
// convention — which also means errors.Is, errors.As, and errors.Unwrap all
// work correctly against any error chain built by this package.
func rootCause(err error) error {
	for {
		next := errors.Unwrap(err)
		if next == nil {
			return err
		}
		err = next
	}
}

// shortenFile trims an absolute source path down to its last two path
// segments (e.g. ".../service/user_service.go" -> "service/user_service.go").
// Keeping one directory of context (rather than the bare filename) avoids
// collapsing distinct call sites that happen to share a filename in
// different packages (e.g. multiple "utilities.go" files) into a single
// fingerprint bucket.
func shortenFile(file string) string {
	if strutil.IsEmpty(file) {
		return "unknown"
	}
	idx := strings.LastIndex(file, "/")
	if idx < 0 {
		return file
	}
	idx2 := strings.LastIndex(file[:idx], "/")
	if idx2 < 0 {
		return file
	}
	return file[idx2+1:]
}

// errorOrigin extracts the (function, file, line) triple used to build a
// fingerprint from err's own embedded stack trace, when err was created via a
// stack-capturing constructor. It only inspects the top-level error (not the
// full unwrap chain) because the outermost stack-aware wrapper is always the
// one closest to the actual failure site in this package's error API.
//
// Returns ("unknown", "unknown", 0) when err carries no embedded stack (e.g.
// a bare third-party error, or one created with [AppendError]/[AppendErrorf]).
// In that case [NewFingerprint] still produces a stable, useful grouping key
// from the error's kind alone.
func errorOrigin(err error) (function, file string, line int) {
	if st, ok := err.(stackTracer); ok {
		if trace := st.StackTrace(); len(trace) > 0 {
			f := trace[0]
			displayName, _, isRuntime := parseStackFrame(f.name())
			if isRuntime {
				displayName = "runtime"
			}
			return displayName, shortenFile(f.file()), f.line()
		}
	}
	return "unknown", "unknown", 0
}

// panicOrigin walks the goroutine's call stack, captured at the point of a
// recovered panic, looking for the frame immediately following
// runtime.gopanic — i.e. the actual function whose panic() call triggered
// the recovery.
//
// This is more robust than skipping a fixed number of frames: Go keeps the
// panicking function's frames on the stack until every deferred call along
// the way returns, so a recover() handler (however many internal helper
// calls deep it lives) can always walk back through its own frames, through
// runtime.gopanic, and land on the original call site. Fixed frame-skip
// counts break the moment an extra helper function is added between recover()
// and this call; scanning for runtime.gopanic does not.
//
// Returns ("unknown", "unknown", 0) if no panic frame is found, which only
// happens when this function is (incorrectly) called outside of an active
// panic/recover.
func panicOrigin() (function, file string, line int) {
	const maxDepth = 64
	pcs := make([]uintptr, maxDepth)
	n := runtime.Callers(0, pcs)
	if n == 0 {
		return "unknown", "unknown", 0
	}
	frames := runtime.CallersFrames(pcs[:n])
	seenPanic := false
	for {
		f, more := frames.Next()
		switch {
		case strings.HasPrefix(f.Function, "runtime.gopanic"):
			seenPanic = true
		case seenPanic && !strings.HasPrefix(f.Function, "runtime."):
			displayName, _, isRuntime := parseStackFrame(f.Function)
			if isRuntime {
				displayName = "runtime"
			}
			return displayName, shortenFile(f.File), f.Line
		}
		if !more {
			break
		}
	}
	return "unknown", "unknown", 0
}
