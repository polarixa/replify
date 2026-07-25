package slogger

import (
	"context"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/polarixa/replify/pkg/encoding"
)

// ///////////////////////////////////////////////////////////////////////////
// Level and FieldType — typed enumerations
// ///////////////////////////////////////////////////////////////////////////

// Level represents the severity of a log message.
// Severity increases with numeric value; TraceLevel is the least severe.
type Level int32

// FieldType enumerates the concrete value types that a Field may carry.
// Callers should use the typed constructor functions (String, Int64, Bool, …)
// rather than setting FieldType directly.
type FieldType int

// ColorMode controls when ANSI colour codes are emitted by TextFormatter.
//
// Use ColorAuto (the default) to let the formatter detect whether the output
// is a TTY and enable colours only when appropriate. Use ColorAlways to force
// colours regardless of output type, or ColorNever to guarantee plain text
// output (suitable for files, pipes, CI environments).
type ColorMode int

// ///////////////////////////////////////////////////////////////////////////
// Logger — core type
// ///////////////////////////////////////////////////////////////////////////

// Logger is the core logging type.
// All exported methods are safe for concurrent use.
//
// A Logger is created with New and customised through functional options.
// Child loggers that inherit configuration can be derived with With and Named.
// The zero value is not usable; always create a Logger through New.
type Logger struct {
	mu         sync.RWMutex // local mutex for formatter/output changes
	wmu        *sync.Mutex  // shared mutex for write synchronization across child loggers
	level      atomic.Int32
	formatter  Formatter
	output     io.Writer
	hooks      *Hooks
	fields     []Field
	name       string
	caller     bool
	callerSkip int
	sampling   *sampler
}

// ///////////////////////////////////////////////////////////////////////////
// Entry — in-flight log event
// ///////////////////////////////////////////////////////////////////////////

// Entry represents a single in-flight log event.
// Entries are pooled internally; callers must not retain an Entry after the
// log call that produced it returns.
//
// To obtain an Entry that carries a context, use Logger.WithContext.
// All log-level methods on Entry are safe for concurrent use.
type Entry struct {
	logger  *Logger
	time    time.Time
	level   Level
	message string
	fields  []Field
	caller  *CallerInfo
	ctx     context.Context
}

// CallerInfo holds the runtime source-location data captured when caller
// reporting is enabled on a Logger.
//
// All fields are read-only after construction. Access them through the
// exported accessor methods File, Line, and Function.
type CallerInfo struct {
	file     string
	line     int
	function string
}

// ///////////////////////////////////////////////////////////////////////////
// Field — structured key-value data
// ///////////////////////////////////////////////////////////////////////////

// Field is a typed key-value pair attached to a log entry.
// All value variants are stored inline to minimise heap allocations.
//
// Use the typed constructors — String, Int, Int64, Float64, Bool, Err,
// Time, Duration, Any, Int8, Int16, Int32, Uint, Uint64, Float32, Timef, JSON —
// to build Field values; do not construct them directly.
// Access the key and type via the Key() and FieldType() accessor methods.
type Field struct {
	key       string
	typ       FieldType
	strVal    string
	intVal    int64
	uint64Val uint64
	floatVal  float64
	boolVal   bool
	errVal    error
	timeVal   time.Time
	durVal    time.Duration
	anyVal    any
}

// ///////////////////////////////////////////////////////////////////////////
// Formatter — serialisation interface
// ///////////////////////////////////////////////////////////////////////////

// Formatter serialises a log Entry to a byte slice.
// Implementations must be safe for concurrent use from multiple goroutines.
//
// Two built-in implementations are provided:
//   - TextFormatter: human-readable key=value lines
//   - JSONFormatter: single-line JSON objects
type Formatter interface {
	Format(*Entry) ([]byte, error)
}

// ///////////////////////////////////////////////////////////////////////////
// Hook — side-effect interface
// ///////////////////////////////////////////////////////////////////////////

// Hook fires side-effects (e.g. alerting, metrics, remote shipping) for
// matching log levels. Implementations must be safe for concurrent use.
//
// Register hooks on a Logger with Logger.AddHook.
// Hooks are called after the entry has been written to the primary output.
type Hook interface {
	// Levels returns the set of log levels this hook handles.
	Levels() []Level

	// Fire is called for each matching entry; it must not retain the entry.
	Fire(*Entry) error
}

// ///////////////////////////////////////////////////////////////////////////
// Hooks — level-indexed hook registry
// ///////////////////////////////////////////////////////////////////////////

// Hooks is a level-indexed registry of Hook instances.
// All methods are safe for concurrent use.
type Hooks struct {
	mu    sync.RWMutex
	hooks map[Level][]Hook
}

// ///////////////////////////////////////////////////////////////////////////
// Options — logger construction configuration
// ///////////////////////////////////////////////////////////////////////////

// Options configures a Logger at construction time.
// Pass functional options to New to override the defaults.
//
// All fields are optional; unset fields fall back to production-safe defaults
// (InfoLevel, TextFormatter, os.Stderr).
type Options struct {
	// level is the minimum level that will be logged.
	level Level

	// formatter serialises entries; defaults to TextFormatter on os.Stderr.
	formatter Formatter

	// output is the destination writer; defaults to os.Stderr.
	output io.Writer

	// caller enables automatic source-location capture.
	caller bool

	// callerSkip adds extra skip frames for library wrappers.
	callerSkip int

	// fields are attached to every entry emitted by the logger.
	fields []Field

	// name is a dot-separated logger identifier prepended to output.
	name string

	// samplingOpts, when non-nil, enables per-message rate limiting.
	samplingOpts *SamplingOptions

	// rotationOpts, when non-nil, enables per-level file rotation.
	rotationOpts *RotationOptions
}

// SamplingOptions configures per-message rate limiting for a Logger.
// The first First messages per Period are always logged.
// After that, every Thereafter-th message is logged (0 means drop all remaining).
type SamplingOptions struct {
	// first is the number of identical messages always logged within Period.
	first int

	// period is the window after which the counter resets.
	period time.Duration

	// thereafter logs every Nth message after the First are exhausted.
	// Zero means drop all subsequent messages.
	thereafter int
}

// ///////////////////////////////////////////////////////////////////////////
// Formatter implementations
// ///////////////////////////////////////////////////////////////////////////

// TextFormatter formats log entries as human-readable key=value lines.
//
// Output format:
//
// 2006-01-02 15:04:05.999999 INFO  message  key=value key2=value2\n
//
// Color Behaviour:
//
// By default (ColorAuto), TextFormatter emits ANSI colour codes only when the
// output writer is a terminal (TTY). Use WithColorMode to override this:
//   - ColorAuto:   colours when output is a TTY
//   - ColorAlways: colours unconditionally (useful for piped output to tools like less -R)
//   - ColorNever:  never emit colours (suitable for files, CI, log aggregators)
//
// The legacy WithDisableColor and WithEnableColor methods set ColorNever and
// ColorAuto respectively for backward compatibility.
type TextFormatter struct {
	timeFormat       string
	disableColors    bool
	disableTimestamp bool
	enableCaller     bool
	output           io.Writer
	colorMode        ColorMode
}

// JSONFormatter formats log entries as single-line JSON objects.
//
// Output example:
//
// {"ts":"2006-01-02 15:04:05.999999","level":"INFO","msg":"message","key":"value"}\n
type JSONFormatter struct {
	timeFormat   string
	enableCaller bool
	enableColor  bool
	timeKey      string
	levelKey     string
	messageKey   string
	callerKey    string
	nameKey      string
	color        *encoding.Style // Color style for log levels
}

// ///////////////////////////////////////////////////////////////////////////
// MultiWriter — fan-out writer
// ///////////////////////////////////////////////////////////////////////////

// MultiWriter fans log output to multiple io.Writer targets simultaneously.
// All Write calls are delivered to every registered writer in registration order.
// All methods are safe for concurrent use.
type MultiWriter struct {
	mu      sync.Mutex
	writers []io.Writer
}

// ///////////////////////////////////////////////////////////////////////////
// Rotation types
// ///////////////////////////////////////////////////////////////////////////

// RotationOptions configures the rotating file writer.
// Pass RotationOptions to WithRotation to enable per-level log file rotation.
type RotationOptions struct {
	// dir is the base log directory. Defaults to defaultLogDir ("logs").
	dir string

	// maxBytes is the maximum file size before rotation. Defaults to 10 MiB.
	maxBytes int64

	// maxAge is the maximum age of a log file before rotation. Zero means no age-based rotation.
	maxAge time.Duration

	// compress controls whether rotated files are zipped. Defaults to false.
	compress bool
}

// LevelFileWriter writes log entries to separate files per log level.
// It supports automatic rotation and ZIP compression of archived logs.
//
// Create instances with NewLevelFileWriter and register them on a logger using
// NewLevelWriterHook together with Logger.AddHook.
type LevelFileWriter struct {
	mu      sync.Mutex
	opts    RotationOptions
	writers map[Level]*rotatingFile
}

// LevelWriterHook is a Hook that writes log entries to level-specific files.
// Use it with AddHook to enable automatic per-level file logging.
type LevelWriterHook struct {
	writer    *LevelFileWriter
	formatter Formatter
	levels    []Level
}

// SloggerConfig mirrors the slogger section of configs/slogger.yaml.
//
// All fields are optional; unset fields fall back to production-safe defaults
// (InfoLevel, TextFormatter, os.Stderr).
type SloggerConfig struct {
	// Level is the minimum severity that will be emitted.
	// Accepted values (case-insensitive): "trace", "debug", "info", "warn", "error".
	// Invalid or empty values fall back to InfoLevel.
	Level string `yaml:"level" json:"level"`

	// Formatter controls the output structure of each log line.
	// Use "text" for human-readable key=value lines (development / CLI tools)
	// or "json" for single-line JSON objects (log aggregators, Loki, Datadog).
	// Defaults to "text" when unset.
	Formatter string `yaml:"formatter" json:"formatter"`

	// Output configures where log entries are delivered (console, file, or both).
	Output OutputConfig `yaml:"output" json:"output"`

	// File configures the base directory and per-level file names used when
	// file output is enabled.
	File FileConfig `yaml:"file" json:"file"`

	// Rotation configures automatic log-file rotation by size and/or age.
	Rotation RotationConfig `yaml:"rotation" json:"rotation"`

	// Archive configures how rotated files are stored after rotation.
	Archive ArchiveConfig `yaml:"archive" json:"archive"`

	// Caller controls capture of the source file and line number on each entry.
	Caller CallerConfig `yaml:"caller" json:"caller"`

	// Color controls ANSI colour codes in text-format output.
	Color ColorConfig `yaml:"color" json:"color"`
}

// OutputConfig configures where log entries are delivered.
type OutputConfig struct {
	// IsConsole writes every log entry to the standard output stream (os.Stdout).
	// Recommended for containerised workloads where the runtime collects stdout.
	IsConsole bool `yaml:"console" json:"console"`

	// IsFile writes every log entry to level-specific files in addition to any
	// console output. Requires IsFile.Directory to be set.
	// Set to false when running in Kubernetes or other stdout-first environments.
	IsFile bool `yaml:"file" json:"file"`
}

// FileConfig configures the base directory and per-level file names for
// file-based log output. All paths are relative to the working directory
// unless an absolute path is given.
type FileConfig struct {
	// Directory is the base directory for all log files.
	// It is created automatically (with mode 0755) if it does not exist.
	// Recommended: use an absolute path in production (e.g. /var/log/myapp).
	Directory string `yaml:"directory" json:"directory"`

	// InfoFile is the filename for INFO-level entries inside Directory.
	InfoFile string `yaml:"info_file" json:"info_file"`

	// WarnFile is the filename for WARN-level entries inside Directory.
	WarnFile string `yaml:"warn_file" json:"warn_file"`

	// ErrorFile is the filename for ERROR-level (and above) entries inside Directory.
	// Fatal and Panic entries are also routed here.
	ErrorFile string `yaml:"error_file" json:"error_file"`

	// DebugFile is the filename for DEBUG-level entries inside Directory.
	// Trace-level entries are also routed here when no dedicated trace file exists.
	DebugFile string `yaml:"debug_file" json:"debug_file"`
}

// RotationConfig configures automatic log-file rotation.
// Either or both of the size and age policies can be active simultaneously;
// the first threshold that is exceeded triggers rotation.
type RotationConfig struct {
	// IsEnabled activates the rotation subsystem. When false, no rotation
	// occurs regardless of the other fields.
	IsEnabled bool `yaml:"enabled" json:"enabled"`

	// MaxSizeMB is the maximum size in megabytes a log file may reach before
	// it is rotated. Very small values (< 10 MB) cause frequent I/O on
	// high-throughput services. Recommended range: 50–200 MB.
	// Set to 0 to disable size-based rotation.
	MaxSizeMB int64 `yaml:"max_size_mb" json:"max_size_mb"`

	// MaxAgeDays is the maximum number of days a log file may be kept open
	// before it is rotated, regardless of its current size.
	// Set to 0 to disable age-based rotation.
	MaxAgeDays int `yaml:"max_age_days" json:"max_age_days"`

	// IsCompress zips rotated files to reduce disk usage.
	// Compression ratios for plain-text logs are typically 80–95 %.
	// Recommended for long-retention or low-disk environments.
	IsCompress bool `yaml:"compress" json:"compress"`
}

// ArchiveConfig configures how rotated log files are archived.
// Rotated files are stored under Path/<date>/ where the date sub-directory
// name is derived from Format (a Go time layout string, e.g. "2006-01-02").
// When IsEnabled is false the archive step is skipped and rotated files are
// written directly next to the active log file.
type ArchiveConfig struct {
	// IsEnabled activates date-bucketed archival of rotated files.
	// When false, rotated files are placed alongside the active log file
	// without a date sub-directory.
	IsEnabled bool `yaml:"enabled" json:"enabled"`

	// Path is the base directory for archived files.
	// Rotated files are stored at Path/<date>/<timestamp>_<level>.{log,zip}.
	// The directory is created automatically if it does not exist.
	Path string `yaml:"path" json:"path"`

	// Format is the Go time layout string used to name the date sub-directory
	// inside Path (e.g. "2006-01-02" for ISO 8601 dates).
	// Must be a valid Go time layout; defaults to "2006-01-02" when unset.
	Format string `yaml:"format" json:"format"`
}

// CallerConfig controls whether the source file and line number of the log
// call-site are captured and appended to each log entry.
// Enabling this adds a small runtime overhead (runtime.Callers) per log call;
// it is recommended in development but can be disabled in production.
type CallerConfig struct {
	// IsEnabled turns on call-site capture (file:line) for every log entry.
	// Recommended: true in development/debugging, false in production.
	IsEnabled bool `yaml:"enabled" json:"enabled"`
}

// ColorConfig controls whether ANSI color codes are applied to log output.
// Color is applied only when the formatter is TextFormatter and the
// destination writer is a TTY; non-TTY writers (files, pipes, CI) are never
// coloured regardless of this setting.
type ColorConfig struct {
	// IsEnabled opts in to ANSI color codes for text-format output.
	// Has no effect when Formatter is set to "json" or when the output
	// writer is not a TTY (e.g. a file or a CI pipeline).
	IsEnabled bool `yaml:"enabled" json:"enabled"`
}

// ///////////////////////////////////////////////////////////////////////////
// Internal — sampler and bucket
// ///////////////////////////////////////////////////////////////////////////

// samplerBucket tracks message counts for a single message key.
type samplerBucket struct {
	mu      sync.Mutex
	count   uint64
	resetAt time.Time
}

// sampler applies per-message rate limiting.
type sampler struct {
	opts    SamplingOptions
	buckets sync.Map // string (message) -> *samplerBucket
}

// ///////////////////////////////////////////////////////////////////////////
// Internal — rotating file handle
// ///////////////////////////////////////////////////////////////////////////

// rotatingFile manages a single level-specific log file with size/age rotation.
type rotatingFile struct {
	mu       sync.Mutex
	path     string
	file     *os.File
	size     int64
	maxBytes int64
	openedAt time.Time
	maxAge   time.Duration
	compress bool
	dir      string
	level    Level
}

// ///////////////////////////////////////////////////////////////////////////
// Internal — context key
// ///////////////////////////////////////////////////////////////////////////

// contextKey is the unexported key type used to store log fields in a context.
// Using a named type prevents collisions with keys from other packages.
type contextKey struct{}

// ///////////////////////////////////////////////////////////////////////////
// Package-level variables
// ///////////////////////////////////////////////////////////////////////////

// entryPool is the package-level sync.Pool for recycling Entry objects.
// Capacity is pre-allocated to defaultEntryFieldCap to avoid early reallocations.
var entryPool = sync.Pool{New: func() any {
	return &Entry{fields: make([]Field, 0, defaultEntryFieldCap)}
}}

// global holds the package-level logger pointer.
// Access is guarded by atomic operations so callers need no external lock.
var global unsafe.Pointer
