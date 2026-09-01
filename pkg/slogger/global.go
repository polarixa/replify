package slogger

import (
	"context"
	"os"
	"sync/atomic"
	"time"
	"unsafe"
)

func init() {
	l := New()
	atomic.StorePointer(&global, unsafe.Pointer(l))
}

// SetGlobalLogger replaces the package-level logger.
//
// Parameters:
//   - `l`: the logger to use for all package-level calls
func SetGlobalLogger(l *Logger) {
	if l == nil {
		return
	}
	atomic.StorePointer(&global, unsafe.Pointer(l))
}

// ResetGlobalLogger resets the package-level logger to a new default logger instance.
// This is useful for testing or when you want to clear any custom configuration and return to the default logging behavior.
func ResetGlobalLogger() {
	atomic.StorePointer(&global, unsafe.Pointer(New()))
}

// GlobalLogger returns the current package-level logger.
//
// Returns:
//
// the active *Logger used by all package-level functions.
//
// Deprecated: Use S() instead. This function is retained for backward compatibility.
func GlobalLogger() *Logger {
	return (*Logger)(atomic.LoadPointer(&global))
}

// S returns the package-level logger, equivalent to GlobalLogger().
//
// Returns:
//
// the active *Logger used by all package-level functions.
func S() *Logger {
	return GlobalLogger()
}

// Trace logs a TRACE-level message via the global logger.
func Trace(msg string, fields ...Field) {
	GlobalLogger().dispatch(TraceLevel, msg, fields...)
}

// Debug logs a DEBUG-level message via the global logger.
func Debug(msg string, fields ...Field) {
	GlobalLogger().dispatch(DebugLevel, msg, fields...)
}

// Info logs an INFO-level message via the global logger.
func Info(msg string, fields ...Field) {
	GlobalLogger().dispatch(InfoLevel, msg, fields...)
}

// Warn logs a WARN-level message via the global logger.
func Warn(msg string, fields ...Field) {
	GlobalLogger().dispatch(WarnLevel, msg, fields...)
}

// Error logs an ERROR-level message via the global logger.
func Error(msg string, fields ...Field) {
	GlobalLogger().dispatch(ErrorLevel, msg, fields...)
}

// Fatal logs a FATAL-level message via the global logger.
func Fatal(msg string, fields ...Field) {
	GlobalLogger().dispatch(FatalLevel, msg, fields...)
}

// Tracef logs a TRACE-level formatted message via the global logger.
func Tracef(format string, args ...any) {
	GlobalLogger().Tracef(format, args...)
}

// Debugf logs a DEBUG-level formatted message via the global logger.
func Debugf(format string, args ...any) {
	GlobalLogger().Debugf(format, args...)
}

// Infof logs an INFO-level formatted message via the global logger.
func Infof(format string, args ...any) {
	GlobalLogger().Infof(format, args...)
}

// Warnf logs a WARN-level formatted message via the global logger.
func Warnf(format string, args ...any) {
	GlobalLogger().Warnf(format, args...)
}

// Errorf logs an ERROR-level formatted message via the global logger.
func Errorf(format string, args ...any) {
	GlobalLogger().Errorf(format, args...)
}

// Fatalf logs a FATAL-level formatted message via the global logger.
func Fatalf(format string, args ...any) {
	GlobalLogger().Fatalf(format, args...)
}

// GlobalWithContextFields returns a new context that carries the provided fields for
// use with the global logger.
func GlobalWithContextFields(ctx context.Context, fields ...Field) context.Context {
	return WithContextFields(ctx, fields...)
}

// ApplyGlobalConfig builds a Logger from cfg and installs it as the package-level
// global logger, replacing any previously set logger atomically.
//
// The function performs the following steps in order:
//
//  1. Parse cfg.Level into a Level constant; falls back to InfoLevel on error.
//  2. Choose the output writer: os.Stdout when colour is enabled (to preserve
//     ANSI escape codes), otherwise a colour-stripped Stdout wrapper.
//  3. Instantiate the Formatter selected by cfg.Formatter ("json" or "text").
//     When cfg.Color.IsEnabled is true the formatter is configured with colour
//     support; otherwise colours are explicitly disabled.
//  4. Build a Logger with the resolved level, formatter, output writer, and
//     caller-reporting flag.
//  5. If cfg.Output.File and cfg.Rotation.IsEnabled are both true, attach a
//     rotating LevelFileWriter that writes to cfg.File.Directory, rotating
//     after cfg.Rotation.MaxSizeMB megabytes or cfg.Rotation.MaxAgeDays days,
//     optionally compressing archived files as ZIP.
//  6. Install the resulting Logger as the global logger via SetGlobalLogger.
//
// ApplyGlobalConfig is safe to call multiple times; each call fully replaces the
// previous global logger. It is not safe to call concurrently with itself.
//
// Parameters:
//   - `cfg`: the configuration to apply; see [SloggerConfig] for field details.
//
// Returns:
//
// nil (reserved for future validation errors).
//
// Example:
//
//	func main() {
//		err := slogger.ApplyGlobalConfig(slogger.SloggerConfig{
//			Level:     "debug",
//			Formatter: "text",
//			Output: slogger.OutputConfig{
//				Console: true,
//				File:    true,
//			},
//			File: slogger.FileConfig{
//				Directory: "logs",
//				InfoFile:  "info.log",
//				WarnFile:  "warn.log",
//				ErrorFile: "error.log",
//				DebugFile: "debug.log",
//			},
//			Rotation: slogger.RotationConfig{
//				IsEnabled:  true,
//				MaxSizeMB:  100,
//				MaxAgeDays: 30,
//				Compress:   true,
//			},
//			Archive: slogger.ArchiveConfig{
//				IsEnabled: true,
//				Path:      "logs/archived",
//				Format:    "2006-01-02",
//			},
//			Caller: slogger.CallerConfig{IsEnabled: true},
//			Color:  slogger.ColorConfig{IsEnabled: true},
//		})
//		if err != nil {
//			panic(err)
//		}
//		slogger.Info("logger ready")
//	}
func ApplyGlobalConfig(cfg SloggerConfig) error {
	// 1. Parse level.
	lvl, err := ParseLevel(cfg.Level)
	if err != nil {
		lvl = InfoLevel
	}

	// 2. Choose output writer.
	output := Stdout()
	if cfg.Color.IsEnabled {
		output = os.Stdout
	}

	// 3. Choose formatter.
	var formatter Formatter
	switch cfg.Formatter {
	case "json":
		if cfg.Color.IsEnabled {
			formatter = NewJSONFormatter().WithEnableColor()
		} else {
			formatter = NewJSONFormatter()
		}
	default:
		if cfg.Color.IsEnabled {
			formatter = NewTextFormatter(output).WithEnableColor()
		} else {
			formatter = NewTextFormatter(output).WithDisableColor()
		}
	}

	// 4. Build the logger using the fluent API.
	log := NewLogger().
		WithLevel(lvl).
		WithFormatter(formatter).
		WithOutput(output).
		WithCaller(cfg.Caller.IsEnabled)

	// 5. Enable file rotation if configured.
	if cfg.Output.IsFile && cfg.Rotation.IsEnabled {
		rotOpts := RotationOptions{
			dir:      cfg.File.Directory,
			maxBytes: cfg.Rotation.MaxSizeMB * int64(1024) * 1024,
			maxAge:   time.Duration(cfg.Rotation.MaxAgeDays) * 24 * time.Hour,
			compress: cfg.Rotation.IsCompress,
		}
		log = log.WithRotation(&rotOpts)
	}

	// 6. Set the global logger.
	SetGlobalLogger(log)
	return nil
}
