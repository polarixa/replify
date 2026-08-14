package replify

import (
	"strings"
	"time"

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

	// The structured log field key used by [wrapper.Logging] and [wrapper.Slogging].
	// The value is the structured map returned by [wrapper.Respond], serialized
	// as JSON (Logging) or Text (Slogging) by the active formatter.
	// This field is always present in Logging and Slogging entries, making them
	// easy to grep in text output (grep 'REPLY=') and filter in JSON pipelines.
	keyReply = "REPLY"

	// keySpanStart is the log message used by [wrapper.Span] at operation start.
	// grep 'span.start' to find all operation beginnings in text output.
	keySpanStart = "[span.start]"

	// keySpanEnd is the log message used by [wrapper.Span] at operation completion.
	// The entry also carries an elapsed_ms field with the measured duration.
	keySpanEnd = "[span.end]"

	// keyScopeEnter is the log message used by [wrapper.Scope] at block entry.
	// grep 'scope.enter' to trace execution boundaries in text output.
	keyScopeEnter = "[scope.enter]"

	// keyScopeExit is the log message used by [wrapper.Scope] at block exit.
	// The entry also carries an elapsed_ms field with time spent inside the block.
	keyScopeExit = "[scope.exit]"

	// keyAwait is the log message used by [wrapper.Await].
	// grep '[await]' to find all intentional wait points in text output.
	keyAwait = "[await]"
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
func (w *wrapper) log(l *slogger.Logger, lvl slogger.Level, message string, fields ...slogger.Field) {
	child := l.With()
	child.WithCaller(true).WithCallerSkip(3)
	switch lvl {
	case slogger.ErrorLevel:
		child.Error(message, fields...)
	case slogger.WarnLevel:
		child.Warn(message, fields...)
	case slogger.InfoLevel:
		child.Info(message, fields...)
	case slogger.DebugLevel:
		child.Debug(message, fields...)
	default:
		child.Trace(message, fields...)
	}
}

// S returns the package-level global logger ([slogger.GlobalLogger]) for convenience.
//
// Use S() when you need to access the global logger directly, e.g. to set
// the minimum log level or change the formatter. For per-call logging, use
// [wrapper.Logging] or [wrapper.Slogging] instead of calling S() directly.
func (w *wrapper) S() *slogger.Logger {
	return slogger.S()
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
	w.autoAdjust()
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
	w.autoAdjust()
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
	w.autoAdjust()
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
	w.log(l, lvl, msg, slogger.JSON(keyReply, w.Respond()))
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

// Span measures the execution duration of a named operation and emits two
// structured log entries: one at the start of the operation and one at its
// completion, including the elapsed time in milliseconds.
//
// Use Span to answer: "How long did this operation take?"
//
// Span is designed for discrete, measurable units of work — such as a
// database query, a cache lookup, an external API call, or a validation step.
// The start entry records that the operation has begun; the end entry confirms
// it has finished and reports how long it ran.
//
// # Conceptual distinction from Scope
//
// Span focuses on timing a unit of work. Scope focuses on visualising
// execution structure (entry/exit of a logical block). Use Span for operations
// whose duration is the primary question; use Scope for request lifecycles,
// transactions, or layered boundaries where visibility into nesting matters.
//
// # Log level
//
// Derived from the wrapper's HTTP status code via [httpStatusLevel]:
// 1xx→Debug, 2xx→Info, 3xx→Warn, 4xx/5xx→Error, no status→Trace.
// Both the start and end entries are emitted at the same level. When the
// resolved level is below the active logger's minimum, Span returns a no-op
// function — zero allocations in that path.
//
// # Thread safety
//
// Safe for concurrent use. The logger, level, name, and start time are
// captured once in Span and shared read-only with the returned closure.
// A goroutine-local child logger is derived via [slogger.Logger.With] inside
// [wrapper.log] on every dispatch; the global logger is never mutated.
//
// # Caller reporting
//
// The start entry resolves to the application call site of Span. The end
// entry resolves to the call site of the returned function (i.e. where done()
// or defer done() appears), using the same callerSkip=3 path through
// [wrapper.log] that all other exported logging methods use.
//
// Parameters:
//   - name:   the operation identifier, e.g. "database.query", "cache.lookup".
//   - fields: optional structured [slogger.Field] values appended to both the
//     start and end entries.
//
// Returns:
//
// a completion func() that, when called, emits the end log entry with the
// elapsed duration. Callers should invoke this via defer immediately after
// calling Span to ensure it fires even on early returns or panics.
//
// Example:
//
//	done := w.Span("load-user")
//	defer done()
//
//	user, err := repo.LoadUser(ctx, id)
//
// Expected output (text formatter, 200 OK wrapper):
//
//	INFO  [span.start] span=load-user state=start caller=handler.go:42
//	INFO  [span.end]   span=load-user state=end elapsed_ms=42 caller=handler.go:43
//
// Structured fields (JSON formatter):
//
//	{"level":"INFO","msg":"[span.start]","span":"load-user","state":"start"}
//	{"level":"INFO","msg":"[span.end]",  "span":"load-user","state":"end","elapsed_ms":42}
func (w *wrapper) Span(name string, fields ...slogger.Field) func() {
	if !w.Available() {
		return func() {}
	}
	w.autoAdjust()
	l := slogger.S()
	lvl := httpStatusLevel(w.StatusCode())
	// Fast path: skip all allocations when the resolved level is below the active minimum.
	if !l.IsLevelEnabled(lvl) {
		return func() {}
	}

	start := time.Now()
	fs := make([]slogger.Field, 0, 2+len(fields))
	fs = append(fs, slogger.String("span", name))
	fs = append(fs, slogger.String("state", "start"))
	fs = append(fs, fields...)
	w.log(l, lvl, keySpanStart, fs...)

	return func() {
		elapsed := time.Since(start)
		lvl := httpStatusLevel(w.StatusCode())
		fe := make([]slogger.Field, 0, 3+len(fields))
		fe = append(fe, slogger.String("span", name))
		fe = append(fe, slogger.String("state", "end"))
		fe = append(fe, slogger.Int64("elapsed_ms", elapsed.Milliseconds()))
		fe = append(fe, fields...)
		w.log(l, lvl, keySpanEnd, fe...)
	}
}

// Scope tracks the entry and exit of a logical execution block, emitting two
// structured log entries: one when execution enters the block and one when it
// leaves, including the elapsed duration in milliseconds.
//
// Use Scope to answer: "When did execution enter and leave this block?"
//
// Scope is designed for logical execution boundaries — HTTP request handlers,
// service methods, transactions, authorization gates, or any named region of
// code whose entry and exit are meaningful for debugging. The entry entry
// records that execution has entered the block; the exit entry confirms it has
// left and reports how long the block ran.
//
// # Conceptual distinction from Span
//
// Scope focuses on visualising execution structure. Span focuses on timing a
// unit of work. Use Scope for request lifecycles, transactions, or layered
// boundaries where nested visibility matters; use Span for discrete operations
// like database queries or external calls whose duration is the primary signal.
//
// # Nesting
//
// Scope calls may be freely nested. Each call captures its own start time
// independently, so inner scopes report only their own duration. The log
// output naturally reflects the nesting order: enter entries appear in
// call order and exit entries appear in reverse (LIFO), matching how defer
// unwinds.
//
// # Log level
//
// Derived from the wrapper's HTTP status code via [httpStatusLevel]:
// 1xx→Debug, 2xx→Info, 3xx→Warn, 4xx/5xx→Error, no status→Trace.
// Both the enter and exit entries are emitted at the same level. When the
// resolved level is below the active logger's minimum, Scope returns a no-op
// function — zero allocations in that path.
//
// # Thread safety
//
// Safe for concurrent use. The logger, level, name, and start time are
// captured once in Scope and shared read-only with the returned closure.
// A goroutine-local child logger is derived via [slogger.Logger.With] inside
// [wrapper.log] on every dispatch; the global logger is never mutated.
//
// # Caller reporting
//
// The enter entry resolves to the application call site of Scope. The exit
// entry resolves to the call site of the returned function (i.e. where the
// deferred call fires), using the same callerSkip=3 path through [wrapper.log]
// that all other exported logging methods use.
//
// Parameters:
//   - name:   the scope identifier, e.g. "CreateUser", "HTTP Request", "Transaction".
//   - fields: optional structured [slogger.Field] values appended to both the
//     enter and exit entries.
//
// Returns:
//
// a completion func() that, when called, emits the exit log entry with the
// elapsed duration. The idiomatic usage is defer w.Scope("name")() — the
// outer call emits the enter log immediately; the deferred inner call emits
// the exit log when the enclosing function returns.
//
// Example:
//
//	defer w.Scope("CreateUser")()
//
//	user, err := svc.CreateUser(ctx, req)
//
// Nested example:
//
//	defer w.Scope("HTTP Request")()
//
//	{
//	    defer w.Scope("Authorization")()
//	}
//	{
//	    defer w.Scope("Repository")()
//	}
//
// Expected output (text formatter, 200 OK wrapper):
//
//	INFO  [scope.enter] scope="HTTP Request"   state=enter caller=handler.go:10
//	INFO  [scope.enter] scope=Authorization    state=enter caller=handler.go:13
//	INFO  [scope.exit]  scope=Authorization    state=exit  elapsed_ms=3  caller=handler.go:13
//	INFO  [scope.enter] scope=Repository       state=enter caller=handler.go:17
//	INFO  [scope.exit]  scope=Repository       state=exit  elapsed_ms=11 caller=handler.go:17
//	INFO  [scope.exit]  scope="HTTP Request"   state=exit  elapsed_ms=23 caller=handler.go:10
//
// Structured fields (JSON formatter):
//
//	{"level":"INFO","msg":"[scope.enter]","scope":"CreateUser","state":"enter"}
//	{"level":"INFO","msg":"[scope.exit]", "scope":"CreateUser","state":"exit","elapsed_ms":84}
func (w *wrapper) Scope(name string, fields ...slogger.Field) func() {
	if !w.Available() {
		return func() {}
	}
	w.autoAdjust()
	l := slogger.S()
	lvl := httpStatusLevel(w.StatusCode())
	// Fast path: skip all allocations when the resolved level is below the active minimum.
	if !l.IsLevelEnabled(lvl) {
		return func() {}
	}

	start := time.Now()
	fs := make([]slogger.Field, 0, 2+len(fields))
	fs = append(fs, slogger.String("scope", name))
	fs = append(fs, slogger.String("state", "enter"))
	fs = append(fs, fields...)
	w.log(l, lvl, keyScopeEnter, fs...)

	return func() {
		elapsed := time.Since(start)
		lvl := httpStatusLevel(w.StatusCode())
		fe := make([]slogger.Field, 0, 3+len(fields))
		fe = append(fe, slogger.String("scope", name))
		fe = append(fe, slogger.String("state", "exit"))
		fe = append(fe, slogger.Int64("elapsed_ms", elapsed.Milliseconds()))
		fe = append(fe, fields...)
		w.log(l, lvl, keyScopeExit, fe...)
	}
}

// Await records an intentional waiting period in the current execution flow,
// emits a structured log entry describing the upcoming pause, then sleeps for
// the specified duration before returning.
//
// Use Await to answer: "Why is execution paused, and for how long?"
//
// Await is designed for deliberate idle time: retry backoff, polling loops,
// rate limiting, throttling, synchronisation waits, circuit-breaker recovery,
// workflow scheduling, or eventual-consistency delays. The log entry makes the
// wait visible in both text and JSON output before it begins, so operators
// reading a log stream can distinguish intentional pauses from stalls.
//
// # Conceptual distinction from Span and Scope
//
// Span measures active work ("how long did this operation take?").
// Scope measures the lifetime of a logical block ("when did execution enter
// and leave this region?"). Await records intentional idle time ("execution
// is paused — for how long and why?"). The three primitives are complementary:
//
//	database query          → Span()
//	HTTP request lifecycle  → Scope()
//	retry backoff           → Await()
//
// # Behavior
//
// Await emits the log entry before sleeping, so the entry is always visible
// even if the process is interrupted during the wait. After logging, it calls
// [time.Sleep] with the given duration and returns the receiver unchanged.
//
// # Duration rules
//
// A duration of zero or below zero does not panic. Behavior follows
// [time.Sleep]: the log entry is still emitted and the call returns
// immediately without blocking.
//
// # Log level
//
// Derived from the wrapper's HTTP status code via [httpStatusLevel]:
// 1xx→Debug, 2xx→Info, 3xx→Warn, 4xx/5xx→Error, no status→Trace.
// Await is a no-op when the resolved level is below the active logger's
// minimum — zero allocations in that path, and no sleep is performed.
//
// # Thread safety
//
// Safe for concurrent use. A goroutine-local child logger is derived via
// [slogger.Logger.With] inside [wrapper.log] on every dispatch; the global
// logger is never mutated.
//
// # Caller reporting
//
// The reported source location resolves to the application call site of
// Await, using the same callerSkip=3 path through [wrapper.log] that all
// other exported logging methods use.
//
// Parameters:
//   - d:      the duration to wait, e.g. 500*time.Millisecond, 2*time.Second.
//     Zero or negative values are logged and return immediately.
//   - fields: optional structured [slogger.Field] values appended after the
//     built-in await and duration_ms fields.
//
// Returns:
//
// the receiver *wrapper unchanged, enabling method chaining.
//
// Example:
//
//	w.
//		Checkpoint("retry.started").
//		Await(2*time.Second, slogger.String("reason", "backoff"), slogger.Int("attempt", 3)).
//		Checkpoint("retry.completed")
//
// Expected output (text formatter, 200 OK wrapper):
//
//	INFO  [checkpoint] checkpoint=retry.started  caller=handler.go:42
//	INFO  [await]      await=2s duration_ms=2000 reason=backoff attempt=3 caller=handler.go:44
//	--- 2 second pause ---
//	INFO  [checkpoint] checkpoint=retry.completed caller=handler.go:45
//
// Structured fields (JSON formatter):
//
//	{"level":"INFO","msg":"[await]","await":"2s","duration_ms":2000,"reason":"backoff","attempt":3}
func (w *wrapper) Await(d time.Duration, fields ...slogger.Field) *wrapper {
	if !w.Available() {
		return w
	}
	w.autoAdjust()
	l := slogger.S()
	lvl := httpStatusLevel(w.StatusCode())
	// Fast path: skip all allocations and the sleep when the resolved level is below the active minimum.
	if !l.IsLevelEnabled(lvl) {
		return w
	}
	fs := make([]slogger.Field, 0, 2+len(fields))
	fs = append(fs, slogger.String("await", d.String()))
	fs = append(fs, slogger.Int64("duration_ms", d.Milliseconds()))
	fs = append(fs, fields...)
	w.log(l, lvl, keyAwait, fs...)
	time.Sleep(d)
	return w
}
