package replify

import (
	"fmt"
	"slices"
	"time"

	"github.com/polarixa/replify/pkg/conv"
	"github.com/polarixa/replify/pkg/slogger"
	"github.com/polarixa/replify/pkg/strchain"
	"github.com/polarixa/replify/pkg/strutil"
)

// newSignature creates and returns a new instance of the [signature] struct.
// This function initializes the [signature] object with default values.
//
// Returns:
//   - A pointer to a newly created [signature] instance.
func newSignature() *signature {
	s := &signature{}
	return s
}

// String returns the string representation of the [SignatureAlgorithm] instance.
//
// Returns:
//   - A string representing the [SignatureAlgorithm] instance.
func (s SignatureAlgorithm) String() string {
	return string(s)
}

// Equals checks whether the [SignatureAlgorithm] instance matches any of the provided algorithms.
//
// Parameters:
//   - other: A variadic list of [SignatureAlgorithm] instances to compare against.
//
// Returns:
//   - A boolean value indicating whether the [SignatureAlgorithm] instance matches any of the provided algorithms:
//   - `true` if a match is found.
//   - `false` if no match is found or if no algorithms are provided.
func (s SignatureAlgorithm) Equals(other ...SignatureAlgorithm) bool {
	if len(other) == 0 {
		return false
	}
	return slices.Contains(other, s)
}

// Available checks whether the [signature] instance is non-nil.
//
// This function ensures that the [signature] object exists and is not nil.
// It serves as a safety check to avoid null pointer dereferences when accessing the instance's fields or methods.
//
// Returns:
//   - A boolean value indicating whether the [signature] instance is non-nil:
//   - `true` if the [signature] instance is non-nil.
//   - `false` if the [signature] instance is nil.
func (s *signature) Available() bool {
	return s != nil
}

// Algorithm retrieves the algorithm associated with the [signature] instance.
//
// This function returns the `algorithm` field of the [signature], which typically
// specifies the cryptographic algorithm used for the signature.
//
// Returns:
//   - A [SignatureAlgorithm] representing the algorithm.
func (s *signature) Algorithm() SignatureAlgorithm {
	if !s.Available() {
		return ""
	}
	return s.algorithm
}

// Value retrieves the value associated with the [signature] instance.
//
// This function returns the `value` field of the [signature], which typically
// contains the actual cryptographic signature.
//
// Returns:
//   - A string representing the value of the signature.
func (s *signature) Value() string {
	if !s.Available() {
		return ""
	}
	return s.value
}

// Timestamp retrieves the timestamp associated with the [signature] instance.
//
// This function returns the `timestamp` field of the [signature], which typically
// represents the time at which the signature was created.
//
// Returns:
//   - An int64 value representing the timestamp.
func (s *signature) Timestamp() int64 {
	if !s.Available() {
		return 0
	}
	return s.timestamp
}

// IsTimestampPresent checks whether the timestamp is present in the [signature] instance.
//
// This function returns `true` if the `timestamp` field is non-zero, indicating that
// a valid timestamp is associated with the signature.
//
// Returns:
//   - A boolean value indicating whether the timestamp is present:
//   - `true` if the timestamp is non-zero.
//   - `false` if the timestamp is zero or the [signature] instance is nil.
func (s *signature) IsTimestampPresent() bool {
	if !s.Available() {
		return false
	}
	return s.timestamp != 0
}

// IsValuePresent checks whether the value is present in the [signature] instance.
//
// This function returns `true` if the `value` field is non-empty, indicating that
// a valid cryptographic signature is associated with the instance.
//
// Returns:
//   - A boolean value indicating whether the value is present:
//   - `true` if the value is non-empty.
//   - `false` if the value is empty or the [signature] instance is nil.
func (s *signature) IsValuePresent() bool {
	if !s.Available() {
		return false
	}
	return strutil.IsNotEmpty(s.value)
}

// IsAlgorithmPresent checks whether the algorithm is present in the [signature] instance.
//
// This function returns `true` if the `algorithm` field is non-empty, indicating that
// a valid cryptographic algorithm is associated with the instance.
//
// Returns:
//   - A boolean value indicating whether the algorithm is present:
//   - `true` if the algorithm is non-empty.
//   - `false` if the algorithm is empty or the [signature] instance is nil.
func (s *signature) IsAlgorithmPresent() bool {
	if !s.Available() {
		return false
	}
	return strutil.IsNotEmpty(s.algorithm.String())
}

// WithAlgorithm sets the algorithm for the [signature] instance.
//
// This function assigns the provided `algorithm` string to the `algorithm` field
// of the [signature] instance and returns the updated instance.
//
// Parameters:
//   - algorithm: A [SignatureAlgorithm] representing the cryptographic algorithm to be set.
//
// Returns:
//   - The updated [signature] instance with the new algorithm.
func (s *signature) WithAlgorithm(algorithm SignatureAlgorithm) *signature {
	s.algorithm = algorithm
	return s
}

// WithValue sets the value for the [signature] instance.
//
// This function assigns the provided `value` string to the `value` field
// of the [signature] instance and returns the updated instance.
//
// Parameters:
//   - value: A string representing the cryptographic signature value.
//
// Returns:
//   - The updated [signature] instance with the new value.
func (s *signature) WithValue(value string) *signature {
	s.value = value
	return s
}

// WithValuef sets the value for the [signature] instance using a formatted string.
//
// This function assigns the formatted string, created using the provided `format` and `args`,
// to the `value` field of the [signature] instance and returns the updated instance.
//
// Parameters:
//   - format: A format string compatible with [fmt.Sprintf].
//   - args: A variadic list of arguments to be formatted according to the `format` string.
//
// Returns:
//   - The updated [signature] instance with the new formatted value.
func (s *signature) WithValuef(format string, args ...any) *signature {
	s.value = fmt.Sprintf(format, args...)
	return s
}

// WithTextValue sets the value for the [signature] instance using a generic type.
//
// This function assigns the provided `value` to the `value` field of the [signature] instance.
// If the value is nil, the instance remains unchanged. The value is converted to a string
// using the [conv.StringOrDefault] function, with a default of "unsupported".
//
// Parameters:
//   - value: A generic value representing the cryptographic signature value.
//
// Returns:
//   - The updated [signature] instance with the new value.
func (s *signature) WithTextValue(value any) *signature {
	if value == nil {
		return s
	}
	s.value = conv.StringOrDefault(value, "unsupported")
	return s
}

// WithJSONValue sets the value for the [signature] instance using a JSON-compatible value.
//
// This function assigns the provided `value` to the `value` field of the [signature] instance.
// If the value is nil, the instance remains unchanged. The value is converted to a JSON string
// using the [jsonpass] function.
//
// Parameters:
//   - value: A generic value representing the cryptographic signature value in JSON-compatible format.
//
// Returns:
//   - The updated [signature] instance with the new JSON-formatted value.
func (s *signature) WithJSONValue(value any) *signature {
	if value == nil {
		return s
	}
	s.value = jsonpass(value)
	return s
}

// WithTimestamp sets the timestamp for the [signature] instance.
//
// This function assigns the provided `timestamp` value to the `timestamp` field
// of the [signature] instance and returns the updated instance.
//
// Parameters:
//   - timestamp: A [time.Time] representing the timestamp to be set.
//
// Returns:
//   - The updated [signature] instance with the new timestamp.
func (s *signature) WithTimestamp(t time.Time) *signature {
	if t.IsZero() {
		return s
	}
	s.timestamp = t.Unix()
	return s
}

// WithNow sets the timestamp for the [signature] instance to the current time.
//
// This function is a convenience method that assigns the current time to the `timestamp` field
// of the [signature] instance and returns the updated instance.
//
// Returns:
//   - The updated [signature] instance with the current timestamp.
func (s *signature) WithNow() *signature {
	return s.WithTimestamp(time.Now())
}

// Respond generates a map representation of the [signature] instance.
//
// This function creates a map containing the fields of the [signature] instance that are present.
// Only the fields that have been set will be included in the resulting map.
//
// Returns:
//   - A map with the present fields of the [signature] instance.
func (s *signature) Respond() map[string]any {
	m := make(map[string]any)
	if s.IsAlgorithmPresent() {
		m["algorithm"] = s.algorithm.String()
	}
	if s.IsValuePresent() {
		m["value"] = s.value
	}
	if s.IsTimestampPresent() {
		m["timestamp"] = s.timestamp
	}
	return m
}

// JSON generates a JSON string representation of the [signature] instance.
//
// This function creates a JSON-formatted string containing the fields of the [signature] instance that are present.
// Only the fields that have been set will be included in the resulting JSON string.
//
// Returns:
//   - A JSON string with the present fields of the [signature] instance.
func (s *signature) JSON() string {
	return jsonpass(s.Respond())
}

// JSONPretty generates a pretty-printed JSON string representation of the [signature] instance.
//
// This function creates a JSON-formatted string containing the fields of the [signature] instance that are present.
// The resulting JSON string is formatted with indentation for better readability.
//
// Returns:
//   - A pretty-printed JSON string with the present fields of the [signature] instance.
func (s *signature) JSONPretty() string {
	return jsonpretty(s.Respond())
}

// Equal compares the current [signature] instance with another [signature] instance.
//
// This function checks if all the present fields in both instances are equal.
// It returns true if all present fields match, and false otherwise.
//
// Parameters:
//   - other: The [signature] instance to compare with.
//
// Returns:
//   - A boolean indicating whether the two [signature] instances are equal.
func (s *signature) Equal(other *signature) bool {
	if s == nil && other == nil {
		return true
	}
	if s == nil || other == nil {
		return false
	}
	if s.IsAlgorithmPresent() != other.IsAlgorithmPresent() {
		return false
	}
	if s.IsAlgorithmPresent() && s.algorithm != other.algorithm {
		return false
	}
	if s.IsValuePresent() != other.IsValuePresent() {
		return false
	}
	if s.IsValuePresent() && s.value != other.value {
		return false
	}
	if s.IsTimestampPresent() != other.IsTimestampPresent() {
		return false
	}
	if s.IsTimestampPresent() && s.timestamp != other.timestamp {
		return false
	}
	return true
}

// HasAlgorithm checks if the current [signature] instance has the specified algorithm.
//
// Parameters:
//   - a: The [SignatureAlgorithm] to check for.
//
// Returns:
//   - A boolean indicating whether the current [signature] instance has the specified algorithm.
func (s *signature) HasAlgorithm(a SignatureAlgorithm) bool {
	if !s.IsAlgorithmPresent() {
		return false
	}
	return s.algorithm.Equals(a)
}

// String generates a string representation of the [signature] instance.
//
// This function creates a human-readable string containing the fields of the [signature] instance.
// All fields are included in the resulting string, regardless of whether they are present or not.
//
// Returns:
//   - A string representation of the [signature] instance.
func (s *signature) String() string {
	if s == nil {
		return ""
	}
	sw := strchain.New()
	sw.AppendF("algorithm=%s", s.algorithm.String())
	sw.Space()
	sw.AppendF("value=%s", s.value)
	sw.Space()
	sw.AppendF("timestamp=%d", s.timestamp)
	return sw.String()
}

// Logging logs the current [signature] instance using the provided logger(s) or the default logger.
//
// This function creates a log entry containing the fields of the [signature] instance.
// If no logger is provided, the default logger is used.
//
// Parameters:
//   - logger: Optional variadic parameter specifying the logger(s) to use.
//
// Returns:
//   - The current [signature] instance.
func (s *signature) Logging(logger ...*slogger.Logger) *signature {
	if s == nil {
		return s
	}
	l := slogger.S()
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0]
	}

	msg := "replify::signature::logging"

	child := l.With()
	child.WithCaller(true).WithCallerSkip(3)

	logAtLevel(child, slogger.InfoLevel, msg, slogger.JSON("SIGNATURE", s.Respond()))
	return s
}

// Slogging logs the string representation of the current [signature] instance using the provided logger(s) or the default logger.
//
// This function creates a log entry containing the string representation of the [signature] instance.
// If no logger is provided, the default logger is used.
//
// Parameters:
//   - logger: Optional variadic parameter specifying the logger(s) to use.
//
// Returns:
//   - The current [signature] instance.
func (s *signature) Slogging(logger ...*slogger.Logger) *signature {
	if s == nil {
		return s
	}
	l := slogger.S()
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0]
	}

	child := l.With()
	child.WithCaller(true).WithCallerSkip(3)

	slogAtLevel(child, slogger.InfoLevel, s.String())
	return s
}

// Reply wraps the current [signature] instance in an [S] wrapper and returns it.
//
// Returns:
//   - An [S] wrapper containing the current [signature] instance.
func (s *signature) Reply() S {
	return S{signature: s}
}

// ReplyPtr wraps the current [signature] instance in an [S] wrapper and returns a pointer to it.
//
// Returns:
//   - A pointer to an [S] wrapper containing the current [signature] instance.
func (s *signature) ReplyPtr() *S {
	return &S{signature: s}
}

// Clone creates a deep copy of the current [signature] instance and returns it.
//
// Returns:
//   - A new [signature] instance that is a copy of the current one.
func (s *signature) Clone() *signature {
	if s == nil {
		return nil
	}
	clone := &signature{
		algorithm: s.algorithm,
		value:     s.value,
		timestamp: s.timestamp,
	}
	return clone
}
