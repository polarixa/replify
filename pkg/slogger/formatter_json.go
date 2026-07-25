package slogger

import (
	"strings"

	"github.com/polarixa/replify/pkg/encoding"
	"github.com/polarixa/replify/pkg/strutil"
)

// WithTimeFormat overrides the timestamp layout used in JSON output.
//
// Parameters:
//   - `fmt`: the Go time layout string (e.g. time.RFC3339Nano)
//
// Returns:
//
// the receiver, for method chaining.
func (f *JSONFormatter) WithTimeFormat(fmt string) *JSONFormatter {
	f.timeFormat = fmt
	return f
}

// WithEnableCaller includes the caller source location (file:line) in the
// JSON output under the callerKey field.
//
// Returns:
//
// the receiver, for method chaining.
func (f *JSONFormatter) WithEnableCaller() *JSONFormatter {
	f.enableCaller = true
	return f
}

// WithColor controls ANSI colour output for the JSON formatter.
// When enabled is true, colour is applied only when the logger's output is a
// TTY terminal; it is suppressed automatically for files and pipes.
// When enabled is false, output is always plain JSON.
//
// Parameters:
//   - `enabled`: true to enable colour (subject to TTY detection), false to disable
//
// Returns:
//
// the receiver, for method chaining.
func (f *JSONFormatter) WithColor(enabled bool) *JSONFormatter {
	f.enableColor = enabled
	return f
}

// WithEnableColor enables ANSI colour output for the JSON formatter.
//
// Returns:
//
// the receiver, for method chaining.
func (f *JSONFormatter) WithEnableColor() *JSONFormatter {
	f.enableColor = true
	return f
}

// WithColorStyle sets the color style for the JSON formatter.
//
// Parameters:
//   - `style`: the color style to use
//
// Returns:
//
// the receiver, for method chaining.
func (f *JSONFormatter) WithColorStyle(style *encoding.Style) *JSONFormatter {
	f.color = style
	return f
}

// WithTimeKey sets the JSON key for the timestamp field.
//
// Parameters:
//   - `key`: the desired JSON key name
//
// Returns:
//
// the receiver, for method chaining.
func (f *JSONFormatter) WithTimeKey(key string) *JSONFormatter {
	f.timeKey = key
	return f
}

// WithLevelKey sets the JSON key for the level field.
//
// Parameters:
//   - `key`: the desired JSON key name
//
// Returns:
//
// the receiver, for method chaining.
func (f *JSONFormatter) WithLevelKey(key string) *JSONFormatter {
	f.levelKey = key
	return f
}

// WithMessageKey sets the JSON key for the message field.
//
// Parameters:
//   - `key`: the desired JSON key name
//
// Returns:
//
// the receiver, for method chaining.
func (f *JSONFormatter) WithMessageKey(key string) *JSONFormatter {
	f.messageKey = key
	return f
}

// WithCallerKey sets the JSON key for the caller field.
//
// Parameters:
//   - `key`: the desired JSON key name
//
// Returns:
//
// the receiver, for method chaining.
func (f *JSONFormatter) WithCallerKey(key string) *JSONFormatter {
	f.callerKey = key
	return f
}

// WithNameKey sets the JSON key for the logger name field.
//
// Parameters:
//   - `key`: the desired JSON key name
//
// Returns:
//
// the receiver, for method chaining.
func (f *JSONFormatter) WithNameKey(key string) *JSONFormatter {
	f.nameKey = key
	return f
}

// TimeFormat returns the time layout string used when formatting timestamps.
//
// Returns:
//
// the Go time layout string.
func (f *JSONFormatter) TimeFormat() string {
	if f == nil {
		return ""
	}
	return f.timeFormat
}

// IsEnableCaller returns whether caller information is included in the JSON output.
//
// Returns:
//
// true if caller reporting is enabled.
func (f *JSONFormatter) IsEnableCaller() bool {
	if f == nil {
		return false
	}
	return f.enableCaller
}

// IsEnableColor returns whether ANSI colour codes are enabled.
//
// Returns:
//
// true if colour output is enabled.
func (f *JSONFormatter) IsEnableColor() bool {
	if f == nil {
		return false
	}
	return f.enableColor
}

// TimeKey returns the JSON key for the timestamp field.
//
// Returns:
//
// the JSON key name for timestamps.
func (f *JSONFormatter) TimeKey() string {
	if f == nil {
		return ""
	}
	return f.timeKey
}

// LevelKey returns the JSON key for the level field.
//
// Returns:
//
// the JSON key name for log levels.
func (f *JSONFormatter) LevelKey() string {
	if f == nil {
		return ""
	}
	return f.levelKey
}

// MessageKey returns the JSON key for the message field.
//
// Returns:
//
// the JSON key name for messages.
func (f *JSONFormatter) MessageKey() string {
	if f == nil {
		return ""
	}
	return f.messageKey
}

// CallerKey returns the JSON key for the caller field.
//
// Returns:
//
// the JSON key name for caller information.
func (f *JSONFormatter) CallerKey() string {
	if f == nil {
		return ""
	}
	return f.callerKey
}

// NameKey returns the JSON key for the logger name field.
//
// Returns:
//
// the JSON key name for logger names.
func (f *JSONFormatter) NameKey() string {
	if f == nil {
		return ""
	}
	return f.nameKey
}

// Color returns the color style used for JSON formatting.
//
// Returns:
//
// the *encoding.Style used for colour output.
func (f *JSONFormatter) Color() *encoding.Style {
	if f == nil {
		return nil
	}
	return f.color
}

// Format serializes e to a single-line JSON byte slice.
// When colour is enabled and the logger's output writer is a TTY terminal,
// ANSI colour escape sequences are applied to keys, values, and level fields
// using encoding.Color with encoding.TerminalStyle. When the output is not a
// terminal (e.g. a file or pipe), plain JSON is returned without modification.
//
// Parameters:
//   - `e`: the log entry to format
//
// Returns:
//
// the formatted bytes and any encoding error.
func (f *JSONFormatter) Format(e *Entry) ([]byte, error) {
	var b strings.Builder
	b.WriteByte('{')

	writeJSONKey(&b, f.timeKey)
	b.WriteByte(':')
	writeJSONString(&b, e.Time().Format(f.timeFormat))

	b.WriteByte(',')
	writeJSONKey(&b, f.levelKey)
	b.WriteByte(':')
	writeJSONString(&b, e.Level().String())

	if l := e.Logger(); l != nil && strutil.IsNotEmpty(l.name) {
		b.WriteByte(',')
		writeJSONKey(&b, f.nameKey)
		b.WriteByte(':')
		writeJSONString(&b, l.name)
	}

	b.WriteByte(',')
	writeJSONKey(&b, f.messageKey)
	b.WriteByte(':')
	writeJSONString(&b, e.Message())

	for _, fld := range e.Fields() {
		b.WriteByte(',')
		writeJSONKey(&b, fld.key)
		b.WriteByte(':')
		writeJSONValue(&b, &fld)
	}

	if f.enableCaller {
		if c := e.Caller(); c != nil {
			b.WriteByte(',')
			writeJSONKey(&b, f.callerKey)
			b.WriteByte(':')
			writeJSONString(&b, c.File()+":"+itoa(c.Line()))
		}
	}

	b.WriteByte('}')
	b.WriteByte('\n')

	data := []byte(b.String())

	// Apply ANSI colour only when color is enabled and the output is a terminal.
	if f.enableColor && e.logger != nil && istty(e.logger.output) {
		if f.color == nil {
			f.color = encoding.TerminalStyle // Default color style
		}
		// encoding.Color operates on the JSON body without the trailing newline.
		colored := encoding.Color(data[:len(data)-1], f.color)
		return append(colored, '\n'), nil
	}

	return data, nil
}
