package conv

import (
	"math"
	"math/cmplx"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/polarixa/replify/pkg/strutil"
)

// ///////////////////////////
// Section:  Time constants
// ///////////////////////////

var (
	emptyTime      = time.Time{}
	typeOfTime     = reflect.TypeFor[time.Time]()
	typeOfDuration = reflect.TypeFor[time.Duration]()
)

// ///////////////////////////
// Section: Time conversion interfaces
// ///////////////////////////

// timeConverter is an interface for types that can convert themselves to time.Time.
type timeConverter interface {
	Time() (time.Time, error)
}

// durationConverter is an interface for types that can convert themselves to time.Duration.
type durationConverter interface {
	Duration() (time.Duration, error)
}

// ///////////////////////////
// Section: Duration conversion
// ///////////////////////////

// Duration attempts to convert the given value to time.Duration, returns the
// zero value and an error on failure.
//
// String parsing supports Go duration format (e.g., "1h30m") and numeric strings.
// Numeric values are treated as nanoseconds.
// Float values are treated as seconds (e.g., 1.5 = 1.5 seconds).
//
// Parameters:
//   - from: The value to convert.
//
// Returns:
//   - The converted time.Duration value.
//   - An error if conversion fails.
func (c *Converter) Duration(from any) (time.Duration, error) {
	if from == nil {
		if c.nilAsZero {
			return 0, nil
		}
		return 0, newConvError(from, "time.Duration")
	}

	// Fast path for common types
	switch v := from.(type) {
	case time.Duration:
		return v, nil
	case *time.Duration:
		if v == nil {
			return 0, nil
		}
		return *v, nil
	case string:
		return c.stringToDuration(v)
	case *string:
		if v == nil {
			return 0, nil
		}
		return c.stringToDuration(*v)
	case int:
		return time.Duration(v), nil
	case int64:
		return time.Duration(v), nil
	case int32:
		return time.Duration(v), nil
	case uint:
		return time.Duration(v), nil
	case uint64:
		if v > math.MaxInt64 {
			v = math.MaxInt64
		}
		return time.Duration(v), nil
	case float64:
		return c.float64ToDuration(v), nil
	case float32:
		return c.float64ToDuration(float64(v)), nil
	}

	// Check for custom converter interface
	if conv, ok := from.(durationConverter); ok {
		return conv.Duration()
	}

	// Use reflection for other types
	return c.durationFromReflect(from)
}

// ///////////////////////////
// Section: Time conversion
// ///////////////////////////

// Time attempts to convert the given value to time.Time, returns the zero value
// of time.Time and an error on failure.
//
// String parsing supports multiple date formats.  Use WithDateFormats() to customize.
//
// Parameters:
//   - from: The value to convert.
//
// Returns:
//   - The converted time.Time value.
//   - An error if conversion fails.
func (c *Converter) Time(from any) (time.Time, error) {
	if from == nil {
		if c.nilAsZero {
			return emptyTime, nil
		}
		return emptyTime, newConvError(from, "time.Time")
	}

	// Fast path for common types
	switch v := from.(type) {
	case time.Time:
		return v, nil
	case *time.Time:
		if v == nil {
			return emptyTime, nil
		}
		return *v, nil
	case string:
		return c.stringToTime(v)
	case *string:
		if v == nil {
			return emptyTime, nil
		}
		return c.stringToTime(*v)
	case int64:
		// Treat as Unix timestamp (seconds)
		return time.Unix(v, 0), nil
	case int:
		return time.Unix(int64(v), 0), nil
	case uint64:
		// Guard against values that exceed the int64 range (time.Unix takes int64).
		// Clamping to math.MaxInt64 avoids silent wrap-around to negative timestamps.
		if v > math.MaxInt64 {
			v = math.MaxInt64
		}
		return time.Unix(int64(v), 0), nil
	case float64:
		// Treat as Unix timestamp with fractional seconds
		sec := int64(v)
		nanoSec := int64((v - float64(sec)) * 1e9)
		return time.Unix(sec, nanoSec), nil
	}

	// Check for custom converter interface
	if conv, ok := from.(timeConverter); ok {
		return conv.Time()
	}

	// Use reflection for other types
	return c.timeFromReflect(from)
}

// ///////////////////////////
// Section:  Package-level Time functions
// ///////////////////////////

// IsZeroTime checks if the given value converts to a zero time.Time.
//
// Parameters:
//   - from: The value to check.
//
// Returns:
//   - true if the converted time.Time is zero, false otherwise.
func IsZeroTime(from any) bool {
	if t, err := defaultConverter.Time(from); err == nil {
		return t.IsZero()
	}
	return true
}

// ///////////////////////////
// Section: Duration helpers
// ///////////////////////////

// Seconds creates a Duration representing the given number of seconds.
//
// Parameters:
//   - n: The number of seconds.
//
// Returns:
//   - A time.Duration representing the specified number of seconds.
func Seconds(n float64) time.Duration {
	return time.Duration(n * float64(time.Second))
}

// Minutes creates a Duration representing the given number of minutes.
//
// Parameters:
//   - n: The number of minutes.
//
// Returns:
//   - A time.Duration representing the specified number of minutes.
func Minutes(n float64) time.Duration {
	return time.Duration(n * float64(time.Minute))
}

// Hours creates a Duration representing the given number of hours.
//
// Parameters:
//   - n: The number of hours.
//
// Returns:
//   - A time.Duration representing the specified number of hours.
func Hours(n float64) time.Duration {
	return time.Duration(n * float64(time.Hour))
}

// Days creates a Duration representing the given number of days.
//
// Parameters:
//   - n: The number of days.
//
// Returns:
//   - A time.Duration representing the specified number of days.
func Days(n float64) time.Duration {
	return time.Duration(n * 24 * float64(time.Hour))
}

// stringToDuration converts a string to time.Duration.
//
// Parameters:
//   - v: The string value to convert.
//
// Returns:
//   - The converted time.Duration value.
//   - An error if the conversion fails.
func (c *Converter) stringToDuration(v string) (time.Duration, error) {
	if strutil.IsEmpty(v) {
		if c.emptyAsZero {
			return 0, nil
		}
		return 0, newConvErrorMsg("cannot convert empty string to time.Duration")
	}

	if c.trimStrings {
		v = strings.TrimSpace(v)
	}

	// Try Go duration format first (e.g., "1h30m")
	if parsed, err := time.ParseDuration(v); err == nil {
		return parsed, nil
	}

	// Try as integer nanoseconds
	if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
		return time.Duration(parsed), nil
	}

	// Try as float seconds
	if parsed, err := strconv.ParseFloat(v, 64); err == nil {
		return c.float64ToDuration(parsed), nil
	}

	return 0, newConvErrorf("cannot parse %q as time.Duration", v)
}

// float64ToDuration converts a float64 value representing seconds to time.Duration.
//
// Values that would overflow or underflow int64 (the underlying type of time.Duration)
// are clamped to the maximum or minimum representable duration.
//
// Parameters:
//   - v: The float64 value to convert.
//
// Returns:
//   - The converted time.Duration value.
func (c *Converter) float64ToDuration(v float64) time.Duration {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	// Compute the duration in nanoseconds as a float64, then check for overflow
	// before casting to int64. math.MaxInt64 / 1e9 ≈ 9.22e9, so durations beyond
	// roughly ±292 years would overflow.
	const nanosPerSecond = float64(time.Second) // 1e9 nanoseconds per second
	nanos := v * nanosPerSecond
	if nanos >= float64(math.MaxInt64) {
		return time.Duration(math.MaxInt64)
	}
	if nanos <= float64(math.MinInt64) {
		return time.Duration(math.MinInt64)
	}
	return time.Duration(nanos)
}

// durationFromReflect converts a reflect.Value to time.Duration.
//
// Parameters:
//   - from: The input value to convert.
//
// Returns:
//   - The converted time.Duration value.
//   - An error if the conversion fails.
func (c *Converter) durationFromReflect(from any) (time.Duration, error) {
	value := indirectValue(reflect.ValueOf(from))
	if !value.IsValid() {
		if c.nilAsZero {
			return 0, nil
		}
		return 0, newConvError(from, "time.Duration")
	}

	kind := value.Kind()
	switch {
	case kind == reflect.String:
		return c.stringToDuration(value.String())
	case isKindInt(kind):
		return time.Duration(value.Int()), nil
	case isKindUint(kind):
		v := value.Uint()
		if v > math.MaxInt64 {
			v = math.MaxInt64
		}
		return time.Duration(v), nil
	case isKindFloat(kind):
		return c.float64ToDuration(value.Float()), nil
	case isKindComplex(kind):
		v := value.Complex()
		if cmplx.IsNaN(v) || cmplx.IsInf(v) {
			return 0, nil
		}
		return c.float64ToDuration(real(v)), nil
	}

	return 0, newConvError(from, "time.Duration")
}

// stringToTime converts a string to time.Time.
//
// Parameters:
//   - v: The string value to convert.
//
// Returns:
//   - The converted time.Time value.
//   - An error if the conversion fails.
func (c *Converter) stringToTime(v string) (time.Time, error) {
	if strutil.IsEmpty(v) {
		if c.emptyAsZero {
			return emptyTime, nil
		}
		return emptyTime, newConvErrorMsg("cannot convert empty string to time.Time")
	}

	if c.trimStrings {
		v = strings.TrimSpace(v)
	}

	// Take a defensive copy of dateFormats under the read lock so that concurrent
	// calls to WithDateFormats (which replaces the slice) cannot affect iteration,
	// and individual elements cannot be mutated underneath the caller.
	c.mu.RLock()
	formats := append([]string(nil), c.dateFormats...)
	c.mu.RUnlock()

	// Try each configured format
	for _, format := range formats {
		if t, err := time.Parse(format, v); err == nil {
			return t, nil
		}
	}

	// Try parsing as Unix timestamp
	if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
		return time.Unix(ts, 0), nil
	}

	// Try parsing as float timestamp
	if ts, err := strconv.ParseFloat(v, 64); err == nil {
		sec := int64(ts)
		nanoSec := int64((ts - float64(sec)) * 1e9)
		return time.Unix(sec, nanoSec), nil
	}

	return emptyTime, newConvErrorf("cannot parse %q as time.Time", v)
}

// timeFromReflect converts a reflect.Value to time.Time.
//
// Parameters:
//   - from: The input value to convert.
//
// Returns:
//   - The converted time.Time value.
//   - An error if the conversion fails.
func (c *Converter) timeFromReflect(from any) (time.Time, error) {
	value := indirectValue(reflect.ValueOf(from))
	if !value.IsValid() {
		if c.nilAsZero {
			return emptyTime, nil
		}
		return emptyTime, newConvError(from, "time.Time")
	}

	kind := value.Kind()
	switch {
	case kind == reflect.String:
		return c.stringToTime(value.String())
	case kind == reflect.Struct:
		if value.Type() == typeOfTime && value.CanInterface() {
			if t, ok := value.Interface().(time.Time); ok {
				return t, nil
			}
		}
	case isKindInt(kind):
		return time.Unix(value.Int(), 0), nil
	case isKindUint(kind):
		return time.Unix(int64(value.Uint()), 0), nil
	case isKindFloat(kind):
		v := value.Float()
		sec := int64(v)
		nanoSec := int64((v - float64(sec)) * 1e9)
		return time.Unix(sec, nanoSec), nil
	}

	return emptyTime, newConvError(from, "time.Time")
}
