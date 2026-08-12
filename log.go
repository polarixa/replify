package replify

import (
	"strings"

	"github.com/polarixa/replify/pkg/slogger"
	"github.com/polarixa/replify/pkg/strutil"
)

const (
	// keyCheckpoint is the structured log field key used by [wrapper.Checkpoint].
	// The value is the checkpoint name, e.g. "user.loaded", "auth.completed".
	// This field is always present in Checkpoint entries, making them easy to grep
	// in text output (grep "checkpoint=auth.completed") and filter in JSON pipelines.
	keyCheckpoint = "[checkpoint]"

	// keyStep is the structured log field key used by [wrapper.Step].
	// The value is the progress formatted as "current/total", e.g. "2/5".
	// This field is always present in Step entries, enabling log-level filtering
	// and easy grepping.
	keyStep = "[step]"

	// keyFlow is the structured log field key used by [wrapper.Flow].
	// The value is the joined execution path, e.g. "HTTP >> Handler >> Service >> Repository".
	// This field is always present in Flow entries, making them easy to grep
	// in text output (grep 'flow=') and filter in JSON pipelines.
	keyFlow = "[flow]"
)

// log is the shared low-level dispatch used by Checkpoint, Step, and Flow.
//
// It derives a goroutine-local child from the global logger via
// [slogger.Logger.With] so that per-call caller settings (WithCaller,
// WithCallerSkip) never race against concurrent users of the same logger.
//
// callerSkip is fixed at 3 to account for the three internal frames that sit
// between an exported wrapper method and the slogger dispatch internals:
//
//	application call site   (frame reported)
//	  ↑ +3
//	exported wrapper method   e.g. Checkpoint()
//	  ↑ +2
//	log                  this helper
//	  ↑ +1
//	child.<Level>             slogger level trampoline
//	  ↑
//	dispatch → dispatchContext → getCaller → runtime.Callers(4+3)
//
// This matches the callerSkip used by [wrapper.Logging] and [wrapper.Slogging]
// All five exported logging methods (Checkpoint, Step, Flow, Logging, Slogging)
// route through this helper.
func (w *wrapper) log(l *slogger.Logger, lvl slogger.Level, msg string, fields ...slogger.Field) {
	child := l.With()
	child.WithCaller(true).WithCallerSkip(3)
	switch lvl {
	case slogger.ErrorLevel:
		child.Error(msg, fields...)
	case slogger.WarnLevel:
		child.Warn(msg, fields...)
	case slogger.InfoLevel:
		child.Info(msg, fields...)
	case slogger.DebugLevel:
		child.Debug(msg, fields...)
	default:
		child.Trace(msg, fields...)
	}
}

// Checkpoint emits a slogger level log entry that marks a named execution point.
//
// Use Checkpoint to answer: "Did execution reach this point?"
//
// The name is recorded as both the log message and the value of a structured
// "checkpoint" field, making entries easy to grep in text output
// (grep "checkpoint=auth.completed") and filter in JSON pipelines.
//
// Optional [slogger.Field] values are appended after the checkpoint field,
// allowing callers to attach contextual data without losing the checkpoint
// identity in the structured output.
//
// # Log level
//
// Derived from the wrapper's HTTP status code via [httpStatusLevel]:
// 1xx→Debug, 2xx→Info, 3xx→Warn, 4xx/5xx→Error, no status→Trace.
// Checkpoint is a no-op when the resolved level is below the active
// logger's minimum — zero allocations in that path.
//
// # Thread safety
//
// Safe for concurrent use. A goroutine-local child logger is derived via
// [slogger.Logger.With] on every call; the global logger is never mutated.
//
// # Caller reporting
//
// The reported source location resolves to the application call site of
// Checkpoint, not to an internal helper frame.
//
// Parameters:
//   - name:   the checkpoint identifier, e.g. "user.loaded", "auth.completed".
//   - fields: optional structured [slogger.Field] values appended to the entry.
//
// Returns:
//
// the receiver *wrapper unchanged, enabling method chaining.
//
// Example:
//
//	w.Checkpoint("request.received")
//	w.Checkpoint("user.loaded", slogger.JSON("DATA", map[string]any{"user_id": id}))
//
// Expected output (text formatter, 200 OK wrapper):
//
//	INFO  checkpoint checkpoint=request.received caller=handler.go:42
//	INFO  checkpoint checkpoint=user.loaded DATA={"user_id":"u1"} caller=handler.go:55
func (w *wrapper) Checkpoint(name string, fields ...slogger.Field) *wrapper {
	if !w.Available() {
		return w
	}
	l := slogger.S()
	lvl := httpStatusLevel(w.StatusCode())
	// Fast path: skip all allocations when the resolved level is below the active minimum.
	if !l.IsLevelEnabled(lvl) {
		return w
	}
	fs := make([]slogger.Field, 0, 1+len(fields))
	fs = append(fs, slogger.String("checkpoint", name))
	fs = append(fs, fields...)
	w.log(l, lvl, keyCheckpoint, fs...)
	return w
}

// Step emits a slogger level log entry for one stage of a multi-step operation.
//
// Use Step to answer: "Which stage of the operation is executing?"
//
// The progress is formatted as "current/total" in the structured "step" field
// (e.g. step=2/5), enabling log-level filtering and easy grepping. The stage
// name is recorded in the "name" field.
//
// Calling Step once per stage of an operation produces a sequential trace that
// shows exactly how far execution progressed — especially useful when a later
// step fails and only the successful steps appear in the log.
//
// # Signature rationale
//
// The signature Step(current, total int, name string, fields ...slogger.Field)
// was chosen over alternatives such as Step(name, current, total) or
// Step(current, name) because:
//   - Numeric progress bounds (current, total) read as a fraction naturally.
//   - The string name comes last to allow the variadic fields to follow
//     without ambiguity.
//   - The total is explicit, avoiding the need to track it out-of-band.
//
// # Log level
//
// Derived from the wrapper's HTTP status code via [httpStatusLevel]:
// 1xx→Debug, 2xx→Info, 3xx→Warn, 4xx/5xx→Error, no status→Trace.
// Step is a no-op when the resolved level is below the active logger's
// minimum — zero allocations in that path.
//
// # Thread safety
//
// Safe for concurrent use. See [wrapper.Checkpoint] for details.
//
// Parameters:
//   - current: 1-based index of the current step (e.g. 1 for the first step).
//   - total:   total number of steps in the operation.
//   - name:    human-readable stage description, e.g. "validate request".
//   - fields:  optional structured [slogger.Field] values appended to the entry.
//
// Returns:
//
// the receiver *wrapper unchanged, enabling method chaining.
//
// Example:
//
//	w.Step(1, 4, "validate request")
//	w.Step(2, 4, "load user", slogger.JSON("DATA", map[string]any{"user_id": id}))
//	w.Step(3, 4, "authorize")
//	w.Step(4, 4, "build response")
//
// Expected output (text formatter, 200 OK wrapper):
//
//	INFO  step step=1/4 name="validate request" caller=handler.go:44
//	INFO  step step=2/4 name="load user" DATA={"user_id":"u1"} caller=handler.go:50
//	INFO  step step=3/4 name=authorize caller=handler.go:55
//	INFO  step step=4/4 name="build response" caller=handler.go:60
func (w *wrapper) Step(current, total int, name string, fields ...slogger.Field) *wrapper {
	if !w.Available() {
		return w
	}
	l := slogger.S()
	lvl := httpStatusLevel(w.StatusCode())
	// Fast path: skip all allocations when the resolved level is below the active minimum.
	if !l.IsLevelEnabled(lvl) {
		return w
	}
	fs := make([]slogger.Field, 0, 2+len(fields))
	fs = append(fs, slogger.Stringf("step", "%d/%d", current, total))
	fs = append(fs, slogger.String("name", name))
	fs = append(fs, fields...)
	w.log(l, lvl, keyStep, fs...)
	return w
}

// Flow emits a slogger level log entry that visualises the execution path as a
// left-to-right sequence of named layers joined by " -> ".
//
// Use Flow to answer: "What path did execution take through the system?"
//
// The joined path is stored in the structured "flow" field, making entries
// easy to grep (grep 'flow=') and meaningful in both text and JSON output.
// Using a single joined string rather than a []string slice ensures that the
// value remains readable in text formatters and fully greppable as one token.
//
// # Signature rationale
//
// Flow(layers ...string) was chosen over Flow(path string) because:
//   - Callers never need to construct or quote the separator manually.
//   - Each layer is an independent, auditable argument at the call site.
//   - The variadic form is consistent with the way other Go tracing libraries
//     (e.g. opentelemetry span names) accept component paths.
//
// Flow takes no optional structured fields because the caller-supplied layers
// already fully describe the path; additional data is better placed on the
// surrounding Checkpoint or Logging calls.
//
// # Log level
//
// Derived from the wrapper's HTTP status code via [httpStatusLevel]:
// 1xx→Debug, 2xx→Info, 3xx→Warn, 4xx/5xx→Error, no status→Trace.
// Flow is a no-op when the resolved level is below the active logger's
// minimum — zero allocations in that path.
//
// # Thread safety
//
// Safe for concurrent use. See [wrapper.Checkpoint] for details.
//
// Parameters:
//   - layers: one or more layer names that describe the execution path,
//     e.g. "HTTP", "Handler", "Service", "Repository".
//     A call with zero layers is silently ignored.
//
// Returns:
//
// the receiver *wrapper unchanged, enabling method chaining.
//
// Example:
//
//	w.Flow("HTTP", "Handler", "UserService", "UserRepository", "Database")
//
// Expected output (text formatter, 200 OK wrapper):
//
//	INFO  flow flow="HTTP -> Handler -> UserService -> UserRepository -> Database" caller=handler.go:65
func (w *wrapper) Flow(layers ...string) *wrapper {
	if !w.Available() {
		return w
	}
	if len(layers) == 0 {
		return w
	}
	l := slogger.S()
	lvl := httpStatusLevel(w.StatusCode())
	// Fast path: skip string joining and allocation when the resolved level is below the active minimum.
	if !l.IsLevelEnabled(lvl) {
		return w
	}
	w.log(l, lvl, keyFlow, slogger.String("flow", strings.Join(layers, " -> ")))
	return w
}

// Logging dispatches a structured log entry for this response using [slogger].
// The log level is automatically selected based on the HTTP status code range:
//
//   - 1xx → Debug  (informational)
//   - 2xx → Info   (success)
//   - 3xx → Warn   (redirection)
//   - 4xx → Error  (client error)
//   - 5xx → Error  (server error; [slogger.Logger.Fatal] is intentionally avoided because it calls os.Exit(1))
//   - other → Trace (no status code set)
//
// The log field key is "REPLY" and its value is the structured map returned
// by [wrapper.Respond], serialized as JSON by the active formatter.
//
// # Thread-safety
//
// Logging is safe for concurrent use. The supplied logger is never mutated: a
// goroutine-local child is derived via [slogger.Logger.With] on every call so
// that the caller-skip adjustment and caller-enable flag stay local to the
// current goroutine. Concurrent callers sharing the same *[slogger.Logger]
// will not race. The wrapper fields (statusCode, message) are read exactly
// once per call to give a consistent snapshot; wrapper fields are expected to
// be immutable after construction via [Wrap] / [With*] options.
//
// # Caller reporting
//
// Caller information is always enabled for this call. callerSkip is set to 2
// to skip both the Logging frame and the slogger level trampoline
// (Trace/Debug/Info/Warn/Error), so the reported file and line resolve to the
// actual call site of Logging.
//
// Parameters:
//   - `logger`: optional *[slogger.Logger] to use. When omitted or nil, the
//     package-level global logger ([slogger.GlobalLogger]) is used.
//
// Returns:
//
// the receiver *wrapper unchanged, enabling method chaining.
//
// Example:
//
//	replify.Wrap(
//		replify.WithStatusCode(replify.OK),
//		replify.WithMessage("User retrieved successfully"),
//		replify.WithBody(user),
//	).Logging()
func (w *wrapper) Logging(logger ...*slogger.Logger) *wrapper {
	if !w.Available() {
		return w
	}
	w.autoAdjust()
	l := slogger.S()
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0]
	}

	lvl := httpStatusLevel(w.StatusCode())
	msg := strutil.DefaultIfEmpty(w.message, "replify::logging")
	w.log(l, lvl, msg, slogger.JSON("REPLY", w.Respond()))
	return w
}

// Slogging dispatches a structured log entry for this response using [slogger], with the log message set to the wrapper's string representation.
// The log level is automatically selected based on the HTTP status code range:
//
//   - 1xx → Debug  (informational)
//   - 2xx → Info   (success)
//   - 3xx → Warn   (redirection)
//   - 4xx → Error  (client error)
//   - 5xx → Error  (server error; [slogger.Logger.Fatal] is intentionally avoided because it calls os.Exit(1))
//   - other → Trace (no status code set)
//
// The log field key is "REPLY" and its value is the structured map returned
// by [wrapper.Respond], serialized as Text by the active formatter.
//
// # Thread-safety
//
// Slogging is safe for concurrent use. The supplied logger is never mutated: a
// goroutine-local child is derived via [slogger.Logger.With] on every call so
// that the caller-skip adjustment and caller-enable flag stay local to the
// current goroutine. Concurrent callers sharing the same *[slogger.Logger]
// will not race. The wrapper fields (statusCode, message) are read exactly once per call to give a consistent snapshot; wrapper fields are expected to be immutable after construction via [Wrap] / [With*] options.
//
// # Caller reporting
//
// Caller information is always enabled for this call. callerSkip is set to 3
// to skip Slogging, the slogger trampoline (Trace/Debug/Info/Warn/Error), and the caller of Slogging, so the reported file and line resolve to the actual call site of Slogging.
//
// Parameters:
//   - `logger`: optional *[slogger.Logger] to use. When omitted or nil, the
//     package-level global logger ([slogger.GlobalLogger]) is used.
//
// Returns:
//
// the receiver *wrapper unchanged, enabling method chaining.
//
// Example:
//
//	replify.Wrap(
//		replify.WithStatusCode(replify.OK),
//		replify.WithMessage("User retrieved successfully"),
//		replify.WithBody(user),
//	).Slogging()
func (w *wrapper) Slogging(logger ...*slogger.Logger) *wrapper {
	if !w.Available() {
		return w
	}
	w.autoAdjust()
	l := slogger.S()
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0]
	}

	lvl := httpStatusLevel(w.StatusCode())
	w.log(l, lvl, w.String())
	return w
}
