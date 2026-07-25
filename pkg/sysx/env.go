package sysx

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/polarixa/replify/pkg/conv"
	"github.com/polarixa/replify/pkg/encoding"
	"github.com/polarixa/replify/pkg/strutil"
)

// Setenv sets the environment variable named by key to value.
//
// It delegates directly to os.Setenv and propagates any error.
//
// Parameters:
//   - `key`:   the name of the environment variable to set.
//   - `value`: the value to assign.
//
// Returns:
//
//	An error if the variable could not be set, or nil on success.
//
// Example:
//
//	if err := sysx.Setenv("LOG_LEVEL", "debug"); err != nil {
//	    log.Fatal(err)
//	}
func Setenv(key string, value any) error {
	str := conv.StringOrDefault(value, "")
	return setenv(key, str)
}

// Unsetenv removes the environment variable named by key from the process
// environment.
//
// It delegates directly to os.Unsetenv and propagates any error.
//
// Parameters:
//   - `key`: the name of the environment variable to remove.
//
// Returns:
//
//	An error if the variable could not be removed, or nil on success.
//
// Example:
//
//	if err := sysx.Unsetenv("TEMP_TOKEN"); err != nil {
//	    log.Fatal(err)
//	}
func Unsetenv(key string) error {
	if strutil.IsEmpty(key) {
		return errors.New("sysx: key must not be empty")
	}
	return os.Unsetenv(key)
}

// Getenv returns the value of the environment variable named by key.
//
// If the variable is not set, or is set to an empty string, Getenv returns
// the provided fallback value instead.
//
// Parameters:
//   - `key`:      the name of the environment variable.
//   - `fallback`: the value to return when the variable is absent or empty.
//
// Returns:
//
//	A string: either the current value of the variable or fallback.
//
// Example:
//
//	port := sysx.Getenv("PORT", "8080")
func Getenv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && strutil.IsNotEmpty(v) {
		return v
	}
	return fallback
}

// Hasenv reports whether the environment variable named by key exists and
// is non-empty.
//
// A variable that is explicitly set to an empty string is treated as absent
// for the purposes of this function.
//
// Parameters:
//   - `key`: the name of the environment variable.
//
// Returns:
//
//	A boolean value:
//	 - true  when the variable is set and its value is non-empty;
//	 - false otherwise.
//
// Example:
//
//	if sysx.Hasenv("DEBUG") {
//	    fmt.Println("debug mode enabled")
//	}
func Hasenv(key string) bool {
	v, ok := os.LookupEnv(key)
	return ok && strutil.IsNotEmpty(v)
}

// MustGetenv returns the value of the environment variable named by key and
// panics if the variable is not set or is empty.
//
// Use this function only when the absence of the variable is considered an
// unrecoverable configuration error (e.g. during application bootstrap).
//
// Parameters:
//   - `key`: the name of the environment variable.
//
// Returns:
//
//	A non-empty string containing the value of the variable.
//
// Example:
//
//	dsn := sysx.MustGetenv("DATABASE_URL")
func MustGetenv(key string) string {
	v, ok := os.LookupEnv(key)
	if !ok || strutil.IsEmpty(v) {
		panic("sysx: required environment variable not set or empty: " + key)
	}
	return v
}

// GetenvInt returns the value of the environment variable named by key parsed
// as a base-10 integer.
//
// If the variable is not set, is empty, or cannot be parsed as an integer,
// the provided fallback value is returned.
//
// Parameters:
//   - `key`:      the name of the environment variable.
//   - `fallback`: the integer to return when the variable is absent, empty, or non-numeric.
//
// Returns:
//
//	An int: either the parsed variable value or fallback.
//
// Example:
//
//	workers := sysx.GetenvInt("WORKERS", 4)
func GetenvInt(key string, defaultValue int) int {
	v, ok := os.LookupEnv(key)
	if !ok || strutil.IsEmpty(v) {
		return defaultValue
	}
	n, err := conv.Int(v)
	if err != nil {
		return defaultValue
	}
	return n
}

// GetenvInt8 returns the value of the environment variable named by key parsed
// as a base-10 integer.
//
// If the variable is not set, is empty, or cannot be parsed as an integer,
// the provided fallback value is returned.
//
// Parameters:
//   - `key`:      the name of the environment variable.
//   - `fallback`: the integer to return when the variable is absent, empty, or non-numeric.
//
// Returns:
//
//	An int8: either the parsed variable value or fallback.
//
// Example:
//
//	workers := sysx.GetenvInt8("WORKERS", 4)
func GetenvInt8(key string, defaultValue int8) int8 {
	v, ok := os.LookupEnv(key)
	if !ok || strutil.IsEmpty(v) {
		return defaultValue
	}
	n, err := conv.Int8(v)
	if err != nil {
		return defaultValue
	}
	return n
}

// GetenvInt16 returns the value of the environment variable named by key parsed
// as a base-10 integer.
//
// If the variable is not set, is empty, or cannot be parsed as an integer,
// the provided fallback value is returned.
//
// Parameters:
//   - `key`:      the name of the environment variable.
//   - `fallback`: the integer to return when the variable is absent, empty, or non-numeric.
//
// Returns:
//
//	An int16: either the parsed variable value or fallback.
//
// Example:
//
//	workers := sysx.GetenvInt16("WORKERS", 4)
func GetenvInt16(key string, defaultValue int16) int16 {
	v, ok := os.LookupEnv(key)
	if !ok || strutil.IsEmpty(v) {
		return defaultValue
	}
	n, err := conv.Int16(v)
	if err != nil {
		return defaultValue
	}
	return n
}

// GetenvInt32 returns the value of the environment variable named by key parsed
// as a base-10 integer.
//
// If the variable is not set, is empty, or cannot be parsed as an integer,
// the provided fallback value is returned.
//
// Parameters:
//   - `key`:      the name of the environment variable.
//   - `fallback`: the integer to return when the variable is absent, empty, or non-numeric.
//
// Returns:
//
//	An int32: either the parsed variable value or fallback.
//
// Example:
//
//	workers := sysx.GetenvInt32("WORKERS", 4)
func GetenvInt32(key string, defaultValue int32) int32 {
	v, ok := os.LookupEnv(key)
	if !ok || strutil.IsEmpty(v) {
		return defaultValue
	}
	n, err := conv.Int32(v)
	if err != nil {
		return defaultValue
	}
	return n
}

// GetenvInt64 returns the value of the environment variable named by key parsed
// as a base-10 integer.
//
// If the variable is not set, is empty, or cannot be parsed as an integer,
// the provided fallback value is returned.
//
// Parameters:
//   - `key`:      the name of the environment variable.
//   - `fallback`: the integer to return when the variable is absent, empty, or non-numeric.
//
// Returns:
//
//	An int64: either the parsed variable value or fallback.
//
// Example:
//
//	workers := sysx.GetenvInt64("WORKERS", 4)
func GetenvInt64(key string, defaultValue int64) int64 {
	v, ok := os.LookupEnv(key)
	if !ok || strutil.IsEmpty(v) {
		return defaultValue
	}
	n, err := conv.Int64(v)
	if err != nil {
		return defaultValue
	}
	return n
}

// GetenvUint returns the value of the environment variable named by key parsed
// as a base-10 unsigned integer.
//
// If the variable is not set, is empty, or cannot be parsed as an integer,
// the provided fallback value is returned.
//
// Parameters:
//   - `key`:      the name of the environment variable.
//   - `fallback`: the integer to return when the variable is absent, empty, or non-numeric.
//
// Returns:
//
//	An uint: either the parsed variable value or fallback.
//
// Example:
//
//	workers := sysx.GetenvUint("WORKERS", 4)
func GetenvUint(key string, defaultValue uint) uint {
	v, ok := os.LookupEnv(key)
	if !ok || strutil.IsEmpty(v) {
		return defaultValue
	}
	n, err := conv.Uint(v)
	if err != nil {
		return defaultValue
	}
	return n
}

// GetenvUint8 returns the value of the environment variable named by key parsed
// as a base-10 unsigned integer.
//
// If the variable is not set, is empty, or cannot be parsed as an integer,
// the provided fallback value is returned.
//
// Parameters:
//   - `key`:      the name of the environment variable.
//   - `fallback`: the integer to return when the variable is absent, empty, or non-numeric.
//
// Returns:
//
//	An uint8: either the parsed variable value or fallback.
//
// Example:
//
//	workers := sysx.GetenvUint8("WORKERS", 4)
func GetenvUint8(key string, defaultValue uint8) uint8 {
	v, ok := os.LookupEnv(key)
	if !ok || strutil.IsEmpty(v) {
		return defaultValue
	}
	n, err := conv.Uint8(v)
	if err != nil {
		return defaultValue
	}
	return n
}

// GetenvUint16 returns the value of the environment variable named by key parsed
// as a base-10 unsigned integer.
//
// If the variable is not set, is empty, or cannot be parsed as an integer,
// the provided fallback value is returned.
//
// Parameters:
//   - `key`:      the name of the environment variable.
//   - `fallback`: the integer to return when the variable is absent, empty, or non-numeric.
//
// Returns:
//
//	An uint16: either the parsed variable value or fallback.
//
// Example:
//
//	workers := sysx.GetenvUint16("WORKERS", 4)
func GetenvUint16(key string, defaultValue uint16) uint16 {
	v, ok := os.LookupEnv(key)
	if !ok || strutil.IsEmpty(v) {
		return defaultValue
	}
	n, err := conv.Uint16(v)
	if err != nil {
		return defaultValue
	}
	return n
}

// GetenvUint32 returns the value of the environment variable named by key parsed
// as a base-10 unsigned integer.
//
// If the variable is not set, is empty, or cannot be parsed as an integer,
// the provided fallback value is returned.
//
// Parameters:
//   - `key`:      the name of the environment variable.
//   - `fallback`: the integer to return when the variable is absent, empty, or non-numeric.
//
// Returns:
//
//	An uint32: either the parsed variable value or fallback.
//
// Example:
//
//	workers := sysx.GetenvUint32("WORKERS", 4)
func GetenvUint32(key string, defaultValue uint32) uint32 {
	v, ok := os.LookupEnv(key)
	if !ok || strutil.IsEmpty(v) {
		return defaultValue
	}
	n, err := conv.Uint32(v)
	if err != nil {
		return defaultValue
	}
	return n
}

// GetenvUint64 returns the value of the environment variable named by key parsed
// as a base-10 unsigned integer.
//
// If the variable is not set, is empty, or cannot be parsed as an integer,
// the provided fallback value is returned.
//
// Parameters:
//   - `key`:      the name of the environment variable.
//   - `fallback`: the integer to return when the variable is absent, empty, or non-numeric.
//
// Returns:
//
//	An uint64: either the parsed variable value or fallback.
//
// Example:
//
//	workers := sysx.GetenvUint64("WORKERS", 4)
func GetenvUint64(key string, defaultValue uint64) uint64 {
	v, ok := os.LookupEnv(key)
	if !ok || strutil.IsEmpty(v) {
		return defaultValue
	}
	n, err := conv.Uint64(v)
	if err != nil {
		return defaultValue
	}
	return n
}

// GetenvFloat32 returns the value of the environment variable named by key parsed
// as a float32.
//
// If the variable is not set, is empty, or cannot be parsed as a float32,
// the provided fallback value is returned.
//
// Parameters:
//   - `key`:      the name of the environment variable.
//   - `fallback`: the float32 to return when the variable is absent, empty, or non-numeric.
//
// Returns:
//
//	A float32: either the parsed variable value or fallback.
//
// Example:
//
//	workers := sysx.GetenvFloat32("WORKERS", 4)
func GetenvFloat32(key string, defaultValue float32) float32 {
	v, ok := os.LookupEnv(key)
	if !ok || strutil.IsEmpty(v) {
		return defaultValue
	}
	n, err := conv.Float32(v)
	if err != nil {
		return defaultValue
	}
	return n
}

// GetenvFloat64 returns the value of the environment variable named by key parsed
// as a float64.
//
// If the variable is not set, is empty, or cannot be parsed as a float64,
// the provided fallback value is returned.
//
// Parameters:
//   - `key`:      the name of the environment variable.
//   - `fallback`: the float64 to return when the variable is absent, empty, or non-numeric.
//
// Returns:
//
//	A float64: either the parsed variable value or fallback.
//
// Example:
//
//	workers := sysx.GetenvFloat64("WORKERS", 4)
func GetenvFloat64(key string, defaultValue float64) float64 {
	v, ok := os.LookupEnv(key)
	if !ok || strutil.IsEmpty(v) {
		return defaultValue
	}
	n, err := conv.Float64(v)
	if err != nil {
		return defaultValue
	}
	return n
}

// GetenvBool returns the value of the environment variable named by key parsed
// as a boolean.
//
// The following string values (case-insensitive) are recognised:
//   - true  strings: "1", "true", "yes", "on"
//   - false strings: "0", "false", "no", "off"
//
// If the variable is not set, is empty, or contains an unrecognised value,
// the provided fallback value is returned.
//
// Parameters:
//   - `key`:      the name of the environment variable.
//   - `fallback`: the boolean to return when the variable cannot be resolved.
//
// Returns:
//
//	A bool: either the parsed variable value or fallback.
//
// Example:
//
//	debug := sysx.GetenvBool("DEBUG", false)
func GetenvBool(key string, defaultValue bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok || strutil.IsEmpty(v) {
		return defaultValue
	}

	result, ok := parseBool(v)
	if !ok {
		return defaultValue
	}
	return result
}

// GetenvDuration returns the value of the environment variable named by key parsed
// as a duration.
//
// If the variable is not set, is empty, or cannot be parsed as a duration,
// the provided fallback value is returned.
//
// Parameters:
//   - `key`:      the name of the environment variable.
//   - `fallback`: the duration to return when the variable cannot be resolved.
//
// Returns:
//
//	A time.Duration: either the parsed variable value or fallback.
//
// Example:
//
//	workers := sysx.GetenvDuration("WORKERS", 4)
func GetenvDuration(key string, defaultValue time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok || strutil.IsEmpty(v) {
		return defaultValue
	}
	n, err := conv.Duration(v)
	if err != nil {
		return defaultValue
	}
	return n
}

// GetenvSlice returns the value of the environment variable named by key
// split by the provided separator.
//
// If the variable is not set or is empty, a nil slice is returned. Empty
// elements produced by consecutive separators are preserved in the result.
//
// Parameters:
//   - `key`: the name of the environment variable.
//   - `sep`: the separator string used to split the variable's value.
//
// Returns:
//
//	A []string containing the split elements, or nil when the variable is absent or empty.
//
// Example:
//
//	hosts := sysx.GetenvSlice("HOSTS", ",") // ["a.example.com", "b.example.com"]
func GetenvSlice(key, sep string) []string {
	if strutil.IsEmpty(key) {
		return nil
	}
	v, ok := os.LookupEnv(key)
	if !ok || strutil.IsEmpty(v) {
		return nil
	}
	return strings.Split(v, sep)
}

// Environ returns a copy of all environment variables in the form
// "key=value".
//
// It delegates directly to os.Environ. The returned slice is a snapshot;
// subsequent changes to the process environment are not reflected.
//
// Returns:
//
//	A []string where each element is a "KEY=value" pair.
//
// Example:
//
//	for _, e := range sysx.Environ() {
//	    fmt.Println(e)
//	}
func Environ() []string {
	return os.Environ()
}

// EnvMap returns all current environment variables as a map from key to
// value.
//
// Each entry in os.Environ is split on the first '=' character. Variables
// without a value portion are stored with an empty string value. The
// returned map is a snapshot; subsequent changes to the process environment
// are not reflected.
//
// Returns:
//
//	A non-nil map[string]string containing all current environment variables.
//
// Example:
//
//	env := sysx.EnvMap()
//	fmt.Println(env["HOME"])
func EnvMap() map[string]string {
	pairs := os.Environ()
	m := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		idx := strings.IndexByte(pair, '=')
		if idx < 0 {
			m[pair] = ""
		} else {
			m[pair[:idx]] = pair[idx+1:]
		}
	}
	return m
}

// JSONEnvMap returns all current environment variables as a JSON string.
//
// It delegates directly to os.Environ. The returned slice is a snapshot;
// subsequent changes to the process environment are not reflected.
//
// Returns:
//
//	A string containing the JSON representation of all current environment variables.
//
// Example:
//
//	env := sysx.JSONEnvMap()
//	fmt.Println(env)
func JSONEnvMap() string {
	return encoding.JSON(EnvMap())
}
