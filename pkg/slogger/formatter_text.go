package slogger

import (
	"io"
	"strconv"
	"strings"

	"github.com/polarixa/replify/pkg/encoding"
	"github.com/polarixa/replify/pkg/fj"
	"github.com/polarixa/replify/pkg/strutil"
)

// WithTimeFormat sets the time layout string used when formatting timestamps.
//
// Parameters:
//   - `fmt`: the Go time layout string (e.g. time.RFC3339Nano)
//
// Returns:
//
// the receiver, for method chaining.
func (f *TextFormatter) WithTimeFormat(fmt string) *TextFormatter {
	f.timeFormat = fmt
	return f
}

// WithDisableColor disables ANSI colour codes in the output.
// Useful when writing to files, pipes, or CI environments that do not
// interpret escape sequences.
//
// Deprecated: Use WithColorMode(ColorNever) instead for explicit control.
// This method is retained for backward compatibility and sets ColorNever.
//
// Returns:
//
// the receiver, for method chaining.
func (f *TextFormatter) WithDisableColor() *TextFormatter {
	f.disableColors = true
	f.colorMode = ColorNever
	return f
}

// WithEnableColor re-enables ANSI colour codes in the output.
// This is the counterpart to WithDisableColors and is useful when colour was
// previously suppressed but the destination is known to be a colour-capable
// terminal or when constructing a formatter that starts with colours disabled
// and opts back in conditionally (e.g. based on a config flag).
//
// Deprecated: Use WithColorMode(ColorAuto) or WithColorMode(ColorAlways) instead.
// This method is retained for backward compatibility and sets ColorAuto.
//
// Returns:
//
// the receiver, for method chaining.
func (f *TextFormatter) WithEnableColor() *TextFormatter {
	f.disableColors = false
	f.colorMode = ColorAuto
	return f
}

// WithColorMode sets the colour mode for the formatter.
//
// ColorMode controls when ANSI colour codes are emitted:
//   - ColorAuto (default): colours when output is a TTY
//   - ColorAlways: colours unconditionally
//   - ColorNever: never emit colours
//
// This method supersedes the legacy WithDisableColor and WithEnableColor methods
// and provides more explicit control over colour behaviour.
//
// Parameters:
//   - `mode`: the ColorMode to use
//
// Returns:
//
// the receiver, for method chaining.
//
// Example:
//
//	// File output: no colours
//	f := slogger.NewTextFormatter(file).WithColorMode(slogger.ColorNever)
//
//	// Force colours even when not a TTY
//	f := slogger.NewTextFormatter(os.Stdout).WithColorMode(slogger.ColorAlways)
func (f *TextFormatter) WithColorMode(mode ColorMode) *TextFormatter {
	f.colorMode = mode
	// Sync legacy field for backward compatibility with IsDisableColors()
	f.disableColors = (mode == ColorNever)
	return f
}

// WithDisableTimestamp omits the timestamp from formatted output.
// Useful when the surrounding infrastructure (systemd, Docker) adds its own
// timestamps.
//
// Returns:
//
// the receiver, for method chaining.
func (f *TextFormatter) WithDisableTimestamp() *TextFormatter {
	f.disableTimestamp = true
	return f
}

// WithEnableCaller appends the source file and line number (caller=file:line)
// to formatted output, aiding in debugging.
//
// Returns:
//
// the receiver, for method chaining.
func (f *TextFormatter) WithEnableCaller() *TextFormatter {
	f.enableCaller = true
	return f
}

// TimeFormat returns the time layout string used when formatting timestamps.
//
// Returns:
//
// the Go time layout string.
func (f *TextFormatter) TimeFormat() string {
	if f == nil {
		return ""
	}
	return f.timeFormat
}

// IsDisableColors returns whether ANSI colour codes are disabled.
//
// Returns:
//
// true if colour output is disabled.
func (f *TextFormatter) IsDisableColors() bool {
	if f == nil {
		return false
	}
	return f.disableColors
}

// IsDisableTimestamp returns whether timestamps are omitted from output.
//
// Returns:
//
// true if timestamps are disabled.
func (f *TextFormatter) IsDisableTimestamp() bool {
	if f == nil {
		return false
	}
	return f.disableTimestamp
}

// IsEnableCaller returns whether caller information is appended to output.
//
// Returns:
//
// true if caller reporting is enabled.
func (f *TextFormatter) IsEnableCaller() bool {
	if f == nil {
		return false
	}
	return f.enableCaller
}

// Output returns the output writer used for TTY detection.
//
// Returns:
//
// the io.Writer used for colour detection.
func (f *TextFormatter) Output() io.Writer {
	if f == nil {
		return nil
	}
	return f.output
}

// ColorMode returns the colour mode configured for this formatter.
//
// Returns:
//
// the ColorMode (ColorAuto, ColorAlways, or ColorNever).
func (f *TextFormatter) ColorMode() ColorMode {
	if f == nil {
		return ColorAuto
	}
	return f.colorMode
}

// shouldUseColor determines whether ANSI colour codes should be emitted
// based on the configured ColorMode and output writer.
//
// Returns:
//
// true if colours should be used for the current output.
func (f *TextFormatter) shouldUseColor() bool {
	switch f.colorMode {
	case ColorAlways:
		return true
	case ColorNever:
		return false
	default: // ColorAuto
		// Legacy behavior: check disableColors flag and TTY detection
		return !f.disableColors && istty(f.output)
	}
}

// Format serializes e to a human-readable key=value byte slice.
//
// Parameters:
//   - `e`: the log entry to format
//
// Returns:
//
// the formatted bytes and any encoding error.
func (f *TextFormatter) Format(e *Entry) ([]byte, error) {
	var b strings.Builder

	useColor := f.shouldUseColor()

	// Determine whether to include caller info based on entry's caller or formatter setting
	// Note: We use a local variable instead of modifying f.enableCaller to maintain thread-safety
	includeCaller := f.enableCaller || e.caller != nil

	if !f.disableTimestamp {
		b.WriteString(e.Time().Format(f.timeFormat))
		b.WriteByte(' ')
	}

	levelStr := levelPad(e.Level())
	if useColor {
		b.WriteString(levelColor(e.Level()))
		b.WriteString(colorBold)
		b.WriteString(levelStr)
		b.WriteString(colorReset)
	} else {
		b.WriteString(levelStr)
	}
	b.WriteByte(' ')

	if l := e.Logger(); l != nil && strutil.IsNotEmpty(l.name) {
		b.WriteByte('[')
		b.WriteString(l.name)
		b.WriteString("] ")
	}

	b.WriteString(e.Message())

	for _, fld := range e.Fields() {
		b.WriteByte(' ')
		b.WriteString(fld.key)
		b.WriteByte('=')
		v := fld.Value()
		if encoding.IsValidJSONString(v) && fj.IsValidJSONString(v) {
			b.WriteString(v)
		} else {
			if shouldQuoting(v) {
				b.WriteString(strconv.Quote(v))
			} else {
				b.WriteString(v)
			}
		}
	}

	if includeCaller {
		if c := e.Caller(); c != nil {
			// if the previous has end with whitespace,
			// we can directly append caller info without an extra space
			if !strings.HasSuffix(b.String(), " ") {
				b.WriteByte(' ')
			}
			b.WriteString("caller=")
			b.WriteString(c.File())
			b.WriteByte(':')
			b.WriteString(strconv.Itoa(c.Line()))
		}
	}

	b.WriteByte('\n')
	return []byte(b.String()), nil
}
