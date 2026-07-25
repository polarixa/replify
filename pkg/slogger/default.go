package slogger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/polarixa/replify/pkg/strutil"
	"github.com/polarixa/replify/pkg/sysx"
)

// ///////////////////////////////////////////////////////////////////////////
// Logger constructors
// ///////////////////////////////////////////////////////////////////////////

// New creates a Logger using the provided functional options.
// Unset options fall back to production-safe defaults (InfoLevel, TextFormatter,
// os.Stderr output).
//
// Parameters:
//   - `opts`: zero or more functions that mutate an *Options before the logger
//     is built
//
// Returns:
//
// a ready-to-use *Logger.
func New(opts ...func(*Options)) *Logger {
	o := defaultOptions()
	for _, fn := range opts {
		fn(o)
	}
	l := &Logger{
		wmu:        &sync.Mutex{}, // shared mutex for write synchronization
		formatter:  o.formatter,
		output:     o.output,
		hooks:      NewHooks(),
		fields:     append([]Field(nil), o.fields...),
		name:       o.name,
		caller:     o.caller,
		callerSkip: o.callerSkip,
	}
	l.level.Store(int32(o.level))
	if o.samplingOpts != nil {
		l.sampling = newSampler(*o.samplingOpts)
	}
	if o.rotationOpts != nil {
		lfw, err := newLevelFileWriter(*o.rotationOpts)
		if err != nil {
			// Rotation setup failed; write diagnostic to stderr and continue
			// without rotation so the logger remains usable.
			_, _ = fmt.Fprintf(os.Stderr, "slogger: rotation setup failed: %v", err)
		} else {
			// Use a color-disabled formatter for file output to prevent
			// ANSI escape sequences from appearing in log files.
			fileFormatter := cloneFormatterForFile(o.formatter)
			l.hooks.Add(NewLevelWriterHook(lfw, fileFormatter))
		}
	}
	return l
}

// NewLogger creates a Logger with all default settings (InfoLevel, TextFormatter,
// os.Stderr). It is equivalent to New() and is provided to enable fluent
// builder-style configuration via the With** methods.
//
// Example:
//
//	log := slogger.NewLogger().
//	    WithLevel(slogger.DebugLevel).
//	    WithFormatter(slogger.NewJSONFormatter()).
//	    WithOutput(os.Stdout)
//
// Returns:
//
// a ready-to-use *Logger configured with production-safe defaults.
func NewLogger() *Logger {
	return New()
}

// ///////////////////////////////////////////////////////////////////////////
// Formatter constructors
// ///////////////////////////////////////////////////////////////////////////

// NewJSONFormatter returns a JSONFormatter with sensible production defaults.
//
// Returns:
//
// a *JSONFormatter with keys ts, level, msg, caller, and name.
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{
		timeFormat:  defaultTimeFormat,
		timeKey:     defaultJSONTimeKey,
		levelKey:    defaultJSONLevelKey,
		messageKey:  defaultJSONMessageKey,
		callerKey:   defaultJSONCallerKey,
		nameKey:     defaultJSONNameKey,
		enableColor: false,
	}
}

// NewTextFormatter returns a TextFormatter that writes to output.
//
// Parameters:
//   - `output`: the destination writer (used only for TTY detection)
//
// Returns:
//
// a *TextFormatter with RFC3339 timestamps and colors enabled when output is
// a terminal.
func NewTextFormatter(output io.Writer) *TextFormatter {
	return &TextFormatter{
		timeFormat: defaultTimeFormat,
		output:     output,
	}
}

// ///////////////////////////////////////////////////////////////////////////
// Hook constructors
// ///////////////////////////////////////////////////////////////////////////

// NewHooks returns an empty Hooks registry.
//
// Returns:
//
// a ready-to-use *Hooks.
func NewHooks() *Hooks {
	return &Hooks{
		hooks: make(map[Level][]Hook, 7),
	}
}

// NewLevelWriterHook creates a LevelWriterHook that writes to lfw using formatter.
// If levels is empty, all levels are enabled.
//
// Parameters:
//   - `lfw`: the LevelFileWriter that handles per-level file output
//   - `formatter`: the Formatter used to serialize each Entry before writing
//   - `levels`: the log levels this hook responds to; if empty all levels fire
//
// Returns:
//
// a *LevelWriterHook registered for the given levels.
func NewLevelWriterHook(lfw *LevelFileWriter, formatter Formatter, levels ...Level) *LevelWriterHook {
	if len(levels) == 0 {
		levels = []Level{TraceLevel, DebugLevel, InfoLevel, WarnLevel, ErrorLevel, FatalLevel, PanicLevel}
	}
	return &LevelWriterHook{
		writer:    lfw,
		formatter: formatter,
		levels:    levels,
	}
}

// ///////////////////////////////////////////////////////////////////////////
// Rotation constructors
// ///////////////////////////////////////////////////////////////////////////

// NewLevelFileWriter creates a LevelFileWriter with the given options.
// The logs directory and per-level files are created automatically if they
// do not already exist.
//
// Parameters:
//   - `opts`: rotation configuration including directory, size limits, and compression
//
// Returns:
//
// a ready-to-use *LevelFileWriter and any initialisation error.
func NewLevelFileWriter(opts RotationOptions) (*LevelFileWriter, error) {
	return newLevelFileWriter(opts)
}

// ///////////////////////////////////////////////////////////////////////////
// Writer constructors
// ///////////////////////////////////////////////////////////////////////////

// NewMultiWriter returns a MultiWriter that writes to all provided writers.
//
// Parameters:
//   - `writers`: one or more destination writers
//
// Returns:
//
// a *MultiWriter targeting every supplied writer.
func NewMultiWriter(writers ...io.Writer) *MultiWriter {
	dst := make([]io.Writer, len(writers))
	copy(dst, writers)
	return &MultiWriter{writers: dst}
}

// NewRotationOptions creates a new RotationOptions with default values.
//
// Returns:
//
// a new *RotationOptions with sensible defaults.
func NewRotationOptions() *RotationOptions {
	return &RotationOptions{}
}

// NewSamplingOptions creates a new SamplingOptions with default values.
//
// Returns:
//
// a new *SamplingOptions.
func NewSamplingOptions() *SamplingOptions {
	return &SamplingOptions{}
}

// defaultOptions returns production-ready defaults:
// InfoLevel, TextFormatter writing to os.Stderr, no caller capture.
//
// Returns:
//
// an *Options pre-populated with sensible production settings.
func defaultOptions() *Options {
	out := os.Stderr
	o := &Options{
		level:     TraceLevel,
		output:    out,
		formatter: NewTextFormatter(out),
	}
	return o
}

// ///////////////////////////////////////////////////////////////////////////
// Sampling helpers
// ///////////////////////////////////////////////////////////////////////////

// newSampler creates a sampler with the given options.
//
// Parameters:
//   - `opts`: the sampling configuration
//
// Returns:
//
// a ready-to-use *sampler.
func newSampler(opts SamplingOptions) *sampler {
	return &sampler{opts: opts}
}

// ///////////////////////////////////////////////////////////////////////////
// Rotation helpers
// ///////////////////////////////////////////////////////////////////////////

// NewLevelFileWriter creates a LevelFileWriter with the given options.
// The logs directory and per-level files are created automatically if they
// do not already exist.
//
// Parameters:
//   - `opts`: rotation configuration including directory, size limits, and compression
//
// Returns:
//
// a ready-to-use *LevelFileWriter and any initialisation error.
func newLevelFileWriter(opts RotationOptions) (*LevelFileWriter, error) {
	if strutil.IsEmpty(opts.dir) {
		opts.dir = defaultLogDir
	}
	if opts.maxBytes <= 0 {
		opts.maxBytes = defaultMaxBytes
	}
	if !sysx.DirExists(opts.dir) {
		if err := os.MkdirAll(opts.dir, 0755); err != nil {
			return nil, fmt.Errorf("slogger: cannot create log directory %q: %w", opts.dir, err)
		}
	}

	w := &LevelFileWriter{
		opts:    opts,
		writers: make(map[Level]*rotatingFile),
	}

	// Four files are created: debug, info, warn, and error.
	// Trace routes to debug; Fatal and Panic route to error (see WriteLevel).
	for _, lvl := range []Level{DebugLevel, InfoLevel, WarnLevel, ErrorLevel} {
		rf, err := newRotatingFile(opts, lvl)
		if err != nil {
			_ = w.Close()
			return nil, err
		}
		w.writers[lvl] = rf
	}
	return w, nil
}

// newRotatingFile creates and opens a rotating log file for the given level.
func newRotatingFile(opts RotationOptions, level Level) (*rotatingFile, error) {
	rf := &rotatingFile{
		path:     filepath.Join(opts.dir, levelFileName(level)),
		maxBytes: opts.maxBytes,
		maxAge:   opts.maxAge,
		compress: opts.compress,
		dir:      opts.dir,
		level:    level,
	}
	if err := rf.open(); err != nil {
		return nil, err
	}
	return rf, nil
}
