package replify

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"

	"github.com/polarixa/replify/pkg/randn"
	"github.com/polarixa/replify/pkg/slogger"
	"github.com/polarixa/replify/pkg/strchain"
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

// newIssue creates a new instance of [issue] with default values.
func newIssue() *issue {
	return &issue{}
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
	i := newIssue()
	i.WithID(NewIssueID()).
		WithFingerprint(NewFingerprint(kind, function, file, line)).
		WithMessage(rootCause(err).Error())
	return i
}

// NewPanicIssue builds an API-facing [issue] from a recovered panic value
// (the return of `recover()`). It must be called from within the deferred
// function that performed the recover; see [panicOrigin] for how the
// original panic site is located regardless of intervening helper calls.
func NewPanicIssue(recovered any) *issue {
	function, file, line := panicOrigin()
	return newIssue().
		WithID(NewIssueID()).
		WithFingerprint(NewFingerprint("panic", function, file, line)).
		WithMessagef("panic: %v", recovered)
}

// Available checks if the issue instance is available (not nil).
//
// Returns:
//   - A boolean indicating whether the issue instance is available.
func (i *issue) Available() bool {
	return i != nil
}

// ID returns the unique identifier for this specific issue occurrence.
//
// Returns:
//   - A string representing the unique identifier for this issue.
func (i *issue) ID() string {
	return i.id
}

// Fingerprint returns the fingerprint identifying the category of failure for this issue.
//
// Returns:
//   - A string representing the fingerprint for this issue.
func (i *issue) Fingerprint() string {
	return i.fingerprint
}

// Message returns the root-cause message for this issue, safe to display to a caller.
//
// Returns:
//   - A string representing the root-cause message for this issue.
func (i *issue) Message() string {
	return i.message
}

// IsIDPresent checks if the unique identifier for this specific issue occurrence is present and non-empty.
//
// Returns:
//   - A boolean indicating whether the unique identifier is present and non-empty.
func (i *issue) IsIDPresent() bool {
	return i.Available() && strutil.IsNotEmpty(i.id)
}

// IsFingerprintPresent checks if the fingerprint identifying the category of failure for this issue is present and non-empty.
//
// Returns:
//   - A boolean indicating whether the fingerprint is present and non-empty.
func (i *issue) IsFingerprintPresent() bool {
	return i.Available() && strutil.IsNotEmpty(i.fingerprint)
}

// IsMessagePresent checks if the root-cause message for this issue is present and non-empty.
//
// Returns:
//   - A boolean indicating whether the root-cause message is present and non-empty.
func (i *issue) IsMessagePresent() bool {
	return i.Available() && strutil.IsNotEmpty(i.message)
}

// WithID sets the unique identifier for this specific issue occurrence.
//
// Parameters:
//   - id: A string representing the unique identifier for this issue.
//
// Returns:
//   - A pointer to the [issue] instance with the updated unique identifier.
func (i *issue) WithID(id string) *issue {
	i.id = id
	return i
}

// WithIDf sets the unique identifier for this specific issue occurrence using a formatted string.
//
// Parameters:
//   - format: A format string.
//   - args: A variadic list of arguments to be formatted according to the format string.
//
// Returns:
//   - A pointer to the [issue] instance with the updated unique identifier.
func (i *issue) WithIDf(format string, args ...any) *issue {
	return i.WithID(fmt.Sprintf(format, args...))
}

// WithFingerprint sets the fingerprint identifying the category of failure for this issue.
//
// Parameters:
//   - fingerprint: A string representing the fingerprint for this issue.
//
// Returns:
//   - A pointer to the [issue] instance with the updated fingerprint.
func (i *issue) WithFingerprint(fingerprint string) *issue {
	i.fingerprint = fingerprint
	return i
}

// WithFingerprintf sets the fingerprint identifying the category of failure for this issue using a formatted string.
//
// Parameters:
//   - format: A format string.
//   - args: A variadic list of arguments to be formatted according to the format string.
//
// Returns:
//   - A pointer to the [issue] instance with the updated fingerprint.
func (i *issue) WithFingerprintf(format string, args ...any) *issue {
	return i.WithFingerprint(fmt.Sprintf(format, args...))
}

// WithMessage sets the root-cause message for this issue, safe to display to a caller.
//
// Parameters:
//   - message: A string representing the root-cause message for this issue.
//
// Returns:
//   - A pointer to the [issue] instance with the updated message.
func (i *issue) WithMessage(message string) *issue {
	i.message = message
	return i
}

// WithMessagef sets the root-cause message for this issue using a formatted string, safe to display to a caller.
//
// Parameters:
//   - format: A format string.
//   - args: A variadic list of arguments to be formatted according to the format string.
//
// Returns:
//   - A pointer to the [issue] instance with the updated message.
func (i *issue) WithMessagef(format string, args ...any) *issue {
	return i.WithMessage(fmt.Sprintf(format, args...))
}

// Respond constructs a map representation of the [issue] instance, including only the fields that are present and non-empty.
//
// Returns:
//   - A map with keys "id", "fingerprint", and "message" for the corresponding fields that are present and non-empty.
func (i *issue) Respond() map[string]any {
	m := make(map[string]any)
	if i.IsIDPresent() {
		m["id"] = i.id
	}
	if i.IsFingerprintPresent() {
		m["fingerprint"] = i.fingerprint
	}
	if i.IsMessagePresent() {
		m["message"] = i.message
	}
	return m
}

// JSON constructs a JSON string representation of the [issue] instance, including only the fields that are present and non-empty.
//
// Returns:
//   - A JSON string representing the [issue] instance.
func (i *issue) JSON() string {
	return jsonpass(i.Respond())
}

// JSONPretty constructs a pretty-printed JSON string representation of the [issue] instance, including only the fields that are present and non-empty.
//
// Returns:
//   - A pretty-printed JSON string representing the [issue] instance.
func (i *issue) JSONPretty() string {
	return jsonpretty(i.Respond())
}

// Equal compares the current [issue] instance with another [issue] instance for equality.
//
// Parameters:
//   - other: A pointer to the [issue] instance to compare with.
//
// Returns:
//   - A boolean value indicating whether the two [issue] instances are equal.
func (i *issue) Equal(other *issue) bool {
	if i == nil && other == nil {
		return true
	}
	if i == nil || other == nil {
		return false
	}
	if i.id != other.id {
		return false
	}
	if i.fingerprint != other.fingerprint {
		return false
	}
	if i.message != other.message {
		return false
	}
	return true
}

// String constructs a string representation of the [issue] instance, including only the fields that are present and non-empty.
//
// Returns:
//   - A string representing the [issue] instance.
func (i *issue) String() string {
	if i == nil {
		return ""
	}
	sw := strchain.New()
	sw.AppendF("id=%s", i.id)
	sw.Space()
	sw.AppendF("fingerprint=%s", i.fingerprint)
	sw.Space()
	sw.AppendF("message=%s", i.message)
	return sw.String()
}

// Logging logs the current [issue] instance using the provided logger or the default logger if none is provided. It includes only the fields that are present and non-empty.
//
// Parameters:
//   - logger: An optional pointer to a [slogger.Logger] instance to use for logging.
//
// Returns:
//   - A pointer to the [issue] instance (enabling method chaining).
func (i *issue) Logging(logger ...*slogger.Logger) *issue {
	if i == nil {
		return i
	}
	l := slogger.S()
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0]
	}

	msg := "replify::issue::logging"

	child := l.With()
	child.WithCaller(true).WithCallerSkip(3)

	logAtLevel(child, slogger.InfoLevel, msg, slogger.JSON("ISSUE", i.Respond()))
	return i
}

// Slogging logs the current [issue] instance using the provided logger or the default logger if none is provided, in a structured manner. It includes only the fields that are present and non-empty.
//
// Parameters:
//   - logger: An optional pointer to a [slogger.Logger] instance to use for logging.
//
// Returns:
//   - A pointer to the [issue] instance (enabling method chaining).
func (i *issue) Slogging(logger ...*slogger.Logger) *issue {
	if i == nil {
		return i
	}
	l := slogger.S()
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0]
	}

	child := l.With()
	child.WithCaller(true).WithCallerSkip(3)

	slogAtLevel(child, slogger.InfoLevel, i.String())
	return i
}

// Reply constructs an [I] instance wrapping the current [issue] instance, providing a structured way to access and manipulate the issue details for API responses.
//
// Returns:
//   - An [I] instance encapsulating the current [issue] instance.
func (i *issue) Reply() I {
	return I{issue: i}
}

// ReplyPtr constructs a pointer to an [I] instance wrapping the current [issue] instance, providing a structured way to access and manipulate the issue details for API responses.
//
// Returns:
//   - A pointer to an [I] instance encapsulating the current [issue] instance.
func (i *issue) ReplyPtr() *I {
	return &I{issue: i}
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

// WithIssue attaches the provided [issue] instance to the [wrapper], replacing any existing issue.
//
// Parameters:
//   - i: A pointer to the [issue] instance to attach.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) WithIssue(i *issue) *wrapper {
	if i == nil {
		return w
	}
	w.issue = i
	return w
}

// AutoIssue computes the [issue] for this [wrapper]'s current error (if any)
// and attaches it to the [wrapper]. It is a no-op when no error is present.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) AutoIssue() *wrapper {
	issue := w.Issue()
	return w.WithIssue(issue)
}
