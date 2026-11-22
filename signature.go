package replify

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"maps"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/polarixa/replify/pkg/conv"
	"github.com/polarixa/replify/pkg/slogger"
	"github.com/polarixa/replify/pkg/strchain"
	"github.com/polarixa/replify/pkg/strutil"
)

// NewSignature creates and returns a new instance of the [signature] struct.
// This function initializes the [signature] object with default values.
// The default algorithm is set to [HMACSHA256].
//
// Returns:
//   - A pointer to a newly created [signature] instance.
func NewSignature() *signature {
	s := &signature{
		algorithm: HMACSHA256,
	}
	return s
}

// String returns the string representation of the [SignatureAlgorithm] instance.
//
// Returns:
//   - A string representing the [SignatureAlgorithm] instance.
func (s SignatureAlgorithm) String() string {
	return string(s)
}

// IsValid checks whether the [SignatureAlgorithm] instance represents a valid, non-empty algorithm.
//
// Returns:
//   - A boolean value indicating whether the [SignatureAlgorithm] instance is valid:
//   - `true` if the algorithm is non-empty.
//   - `false` if the algorithm is empty.
func (s SignatureAlgorithm) IsValid() bool {
	return strutil.IsNotEmpty(string(s))
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

// Headers retrieves the headers associated with the [signature] instance.
//
// This function returns the `headers` field of the [signature], which typically
// contains optional headers included in the signature.
//
// Returns:
//   - A map of header names to their corresponding values. Returns `nil` if the [signature] instance is nil.
func (s *signature) Headers() map[string]string {
	if !s.Available() {
		return nil
	}
	return s.headers
}

// LenHeaders retrieves the number of headers associated with the [signature] instance.
//
// This function returns the length of the `headers` map of the [signature].
// If the instance is nil or the headers map is nil, it returns 0.
//
// Returns:
//   - An integer representing the number of headers. Returns 0 if the [signature] instance is nil or has no headers.
func (s *signature) LenHeaders() int {
	if !s.Available() {
		return 0
	}
	if s.headers == nil {
		return 0
	}
	return len(s.headers)
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

// IsHeadersPresent checks whether the headers are present in the [signature] instance.
//
// This function returns `true` if the `headers` field is non-nil and contains at least one header,
// indicating that optional headers are associated with the signature.
//
// Returns:
//   - A boolean value indicating whether the headers are present:
//   - `true` if the headers are non-nil and contain at least one entry.
//   - `false` if the headers are nil, empty, or the [signature] instance is nil.
func (s *signature) IsHeadersPresent() bool {
	if !s.Available() {
		return false
	}
	if s.headers == nil {
		return false
	}
	return len(s.headers) > 0
}

// HasHeaderKey checks whether a specific header key is present in the [signature] instance.
//
// This function returns `true` if the `headers` field contains the specified key,
// indicating that the corresponding header is associated with the signature.
//
// Parameters:
//   - key: The header key to check for presence.
//
// Returns:
//   - A boolean value indicating whether the specified header key is present:
//   - `true` if the key exists in the headers.
//   - `false` if the key does not exist, the headers are nil, or the [signature] instance is nil.
func (s *signature) HasHeaderKey(key string) bool {
	if !s.IsHeadersPresent() {
		return false
	}
	if strutil.IsEmpty(key) {
		return false
	}
	_, exists := s.headers[key]
	return exists
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

// WithTime sets the timestamp for the [signature] instance.
//
// This function assigns the provided `timestamp` value to the `timestamp` field
// of the [signature] instance and returns the updated instance.
//
// Parameters:
//   - timestamp: A [time.Time] representing the timestamp to be set.
//
// Returns:
//   - The updated [signature] instance with the new timestamp.
func (s *signature) WithTime(t time.Time) *signature {
	if t.IsZero() {
		return s
	}
	s.timestamp = t.Unix()
	return s
}

// WithTimeUnix sets the timestamp for the [signature] instance using a Unix timestamp.
//
// This function assigns the provided `timestamp` value to the `timestamp` field
// of the [signature] instance and returns the updated instance.
//
// Parameters:
//   - timestamp: An int64 representing the Unix timestamp to be set.
//
// Returns:
//   - The updated [signature] instance with the new Unix timestamp.
func (s *signature) WithTimeUnix(timestamp int64) *signature {
	if timestamp <= 0 {
		return s
	}
	s.timestamp = timestamp
	return s
}

// WithTimeDuration sets the timestamp for the [signature] instance using a duration from the current time.
//
// This function assigns the current time plus the provided `duration` value to the `timestamp` field
// of the [signature] instance and returns the updated instance.
//
// Parameters:
//   - duration: A [time.Duration] representing the duration to be added to the current time.
//
// Returns:
//   - The updated [signature] instance with the new timestamp.
func (s *signature) WithTimeDuration(d time.Duration) *signature {
	if d <= 0 {
		return s
	}
	s.timestamp = time.Now().Add(d).Unix()
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
	return s.WithTime(time.Now())
}

// WithHeaders sets the headers for the [signature] instance.
//
// This function assigns the provided `headers` map to the `headers` field of the [signature] instance.
// If the map is empty, the instance remains unchanged.
//
// Parameters:
//   - headers: A map containing the headers to be set.
//
// Returns:
//   - The updated [signature] instance with the new headers.
func (s *signature) WithHeaders(headers map[string]string) *signature {
	if len(headers) == 0 {
		return s
	}
	s.headers = headers
	return s
}

// WithHeader sets a single header for the [signature] instance.
//
// This function assigns the provided `key` and `value` to the `headers` field of the [signature] instance.
// If the key or value is empty, the instance remains unchanged.
//
// Parameters:
//   - key: The header key to be set.
//   - value: The header value to be set.
//
// Returns:
//   - The updated [signature] instance with the new header.
func (s *signature) WithHeader(key, value string) *signature {
	if strutil.IsEmpty(key) || strutil.IsEmpty(value) {
		return s
	}
	if s.headers == nil {
		s.headers = make(map[string]string)
	}
	s.headers[key] = value
	return s
}

// Nonce returns the single-use nonce signed with the payload, or an empty string when there is none.
//
// Returns:
//   - The nonce, or an empty string if the instance is not available or no nonce is set.
func (s *signature) Nonce() string {
	if !s.Available() {
		return ""
	}
	return s.nonce
}

// IsNoncePresent checks whether a non-empty nonce is set on the [signature] instance.
//
// Returns:
//   - true if a nonce is present, false otherwise.
func (s *signature) IsNoncePresent() bool {
	return s.Available() && strutil.IsNotEmpty(s.nonce)
}

// WithNonce sets the single-use nonce of the [signature] instance. The nonce is only meaningful when
// replay protection is enabled (see [SignatureConfig.EnableReplayProtection]), where it is signed with the payload.
//
// Parameters:
//   - nonce: The nonce to set; an empty string clears it.
//
// Returns:
//   - A pointer to the modified [signature] instance (enabling method chaining).
func (s *signature) WithNonce(nonce string) *signature {
	s.nonce = nonce
	return s
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
	if s.IsHeadersPresent() {
		m["headers"] = s.headers
	}
	if s.IsNoncePresent() {
		m["nonce"] = s.nonce
	}
	return m
}

// RespondIgnoring generates a map representation of the [signature] instance, ignoring the specified top-level fields.
//
// This function creates a map containing the fields of the [signature] instance that are present,
// excluding the fields specified in the `level1fields` parameter.
//
// Parameters:
//   - level1fields: top-level fields in the response body to ignore.
//
// Returns:
//   - A map with the present fields of the [signature] instance, excluding the ignored fields.
func (s *signature) RespondIgnoring(level1fields ...string) map[string]any {
	if len(level1fields) == 0 {
		return s.Respond()
	}
	// Respond() returns the shared cached map; clone it so deleting fields
	// below doesn't mutate the cache backing future Respond()/JSON() calls.
	src := s.Respond()
	m := make(map[string]any, len(src))
	maps.Copy(m, src)
	for _, field := range level1fields {
		if strutil.IsEmpty(field) {
			continue
		}
		delete(m, field)
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

// JSONIgnoring generates a JSON string representation of the [signature] instance, ignoring the specified top-level fields.
//
// This function creates a JSON-formatted string containing the fields of the [signature] instance that are present,
// excluding the fields specified in the `level1fields` parameter.
//
// Parameters:
//   - level1fields: top-level fields in the response body to ignore.
//
// Returns:
//   - A JSON string with the present fields of the [signature] instance, excluding the ignored fields.
func (s *signature) JSONIgnoring(level1fields ...string) string {
	return jsonpass(s.RespondIgnoring(level1fields...))
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

// JSONPrettyIgnoring generates a pretty-printed JSON string representation of the [signature] instance, ignoring the specified top-level fields.
//
// This function creates a JSON-formatted string containing the fields of the [signature] instance that are present,
// excluding the fields specified in the `level1fields` parameter.
//
// Parameters:
//   - level1fields: top-level fields in the response body to ignore.
//
// Returns:
//   - A pretty-printed JSON string with the present fields of the [signature] instance, excluding the ignored fields.
func (s *signature) JSONPrettyIgnoring(level1fields ...string) string {
	return jsonpretty(s.RespondIgnoring(level1fields...))
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
	if s.IsHeadersPresent() != other.IsHeadersPresent() {
		return false
	}
	if s.IsHeadersPresent() && !reflect.DeepEqual(s.headers, other.headers) {
		return false
	}
	if s.IsNoncePresent() != other.IsNoncePresent() {
		return false
	}
	if s.IsNoncePresent() && s.nonce != other.nonce {
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
	if s.IsHeadersPresent() {
		sw.Space()
		sw.AppendF("headers=%v", conv.StringOrEmpty(s.headers))
	}
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

// LoggingIgnoring logs the current [signature] instance using the provided logger(s) or the default logger, ignoring the specified top-level fields in the response body.
//
// This function creates a log entry containing the fields of the [signature] instance that are present,
// excluding the fields specified in the `level1fields` parameter.
//
// Parameters:
//   - level1fields: top-level fields in the response body to ignore.
//
// Returns:
//   - The current [signature] instance.
func (s *signature) LoggingIgnoring(level1fields ...string) *signature {
	if s == nil {
		return s
	}
	l := slogger.S()
	child := l.With()
	child.WithCaller(true).WithCallerSkip(3)

	msg := "replify::signature::logging_ignoring"

	logAtLevel(child, slogger.InfoLevel, msg, slogger.JSON("SIGNATURE", s.RespondIgnoring(level1fields...)))
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
// The headers map is copied, so mutating the clone never affects the original.
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
		headers:   maps.Clone(s.headers),
		nonce:     s.nonce,
	}
	return clone
}

// NewSignatureConfig creates a new instance of [SignatureConfig] with default values.
// It sets the default algorithm to HMAC-SHA256 ([HMACSHA256]), includes the timestamp by default,
// initializes an empty list of headers to sign, sets the maximum age to 5 minutes, and tolerates
// [DefaultSignatureMaxClockSkew] (30 seconds) of clock skew for timestamps in the future. The legacy
// HMAC-SHA1 and HMAC-MD5 algorithms are disabled; see [SignatureConfig.EnableLegacyAlgorithms].
// Note that a zero-value &SignatureConfig{} has a clock skew of 0 and is therefore strict.
//
// Parameters:
//   - secretKey: The secret key to be used for signature generation.
func NewSignatureConfig(secretKey string) *SignatureConfig {
	return &SignatureConfig{
		algorithm:        HMACSHA256,      // Default algorithm for signature generation
		secretKey:        secretKey,       // Set the provided secret key
		includeTimestamp: true,            // Include timestamp in the signature by default
		headersToSign:    []string{},      // Headers to include in the signature (optional)
		maxAge:           5 * time.Minute, // Maximum age for the signature to be considered valid (optional)
		canonicalize:     false,           // Disable canonicalization of the JSON payload by default
		maxClockSkew:     DefaultSignatureMaxClockSkew,
	}
}

// Available checks if the current [SignatureConfig] instance is non-nil.
//
// Returns:
//   - true if the [SignatureConfig] instance is non-nil, false otherwise.
func (s *SignatureConfig) Available() bool {
	return s != nil
}

// IsValid checks if the current [SignatureConfig] instance has valid configuration.
//
// Returns:
//   - true if the [SignatureConfig] instance is non-nil and has both algorithm and secret key set, false otherwise.
func (s *SignatureConfig) IsValid() bool {
	if s == nil {
		return false
	}
	if !s.algorithm.IsValid() || strutil.IsEmpty(s.secretKey) {
		return false
	}
	return true
}

// Algorithm returns the signature algorithm configured in the current [SignatureConfig] instance.
//
// Returns:
//   - The [SignatureAlgorithm] of the current [SignatureConfig] instance, or an empty string if the instance is not available.
func (s *SignatureConfig) Algorithm() SignatureAlgorithm {
	if !s.Available() {
		return ""
	}
	return s.algorithm
}

// SecretKey returns the secret key configured in the current [SignatureConfig] instance.
//
// Returns:
//   - The secret key of the current [SignatureConfig] instance, or an empty string if the instance is not available.
func (s *SignatureConfig) SecretKey() string {
	if !s.Available() {
		return ""
	}
	return s.secretKey
}

// IsIncludeTimestamp returns whether the timestamp is included in the current [SignatureConfig] instance.
//
// Returns:
//   - true if the timestamp is included, false otherwise, or false if the instance is not available.
func (s *SignatureConfig) IsIncludeTimestamp() bool {
	if !s.Available() {
		return false
	}
	return s.includeTimestamp
}

// MaxAge returns the maximum age configured in the current [SignatureConfig] instance.
//
// Returns:
//   - The maximum age of the current [SignatureConfig] instance, or 0 if the instance is not available.
func (s *SignatureConfig) MaxAge() time.Duration {
	if !s.Available() {
		return 0
	}
	return s.maxAge
}

// HeadersToSign returns the list of headers configured to be included in the signature for the current [SignatureConfig] instance.
//
// Returns:
//   - The list of headers to sign of the current [SignatureConfig] instance, or an empty list if the instance is not available.
func (s *SignatureConfig) HeadersToSign() []string {
	if !s.Available() {
		return []string{}
	}
	return s.headersToSign
}

// IgnoringSignatureFields returns the JSON field names excluded when computing the signature for the current [SignatureConfig] instance.
//
// Returns:
//   - The list of ignored signature fields of the current [SignatureConfig] instance, or an empty list if the instance is not available.
func (s *SignatureConfig) IgnoringSignatureFields() []string {
	if !s.Available() {
		return []string{}
	}
	return s.ignoringSignatureFields
}

// IgnoringResponseFields returns the JSON field names excluded from the response written to the client for the current [SignatureConfig] instance.
//
// Returns:
//   - The list of ignored response fields of the current [SignatureConfig] instance, or an empty list if the instance is not available.
func (s *SignatureConfig) IgnoringResponseFields() []string {
	if !s.Available() {
		return []string{}
	}
	return s.ignoringResponseFields
}

// IsAlgorithmPresent checks if a valid signature algorithm is configured in the current [SignatureConfig] instance.
//
// Returns:
//   - true if a valid signature algorithm is configured, false otherwise, or false if the instance is not available.
func (s *SignatureConfig) IsAlgorithmPresent() bool {
	if !s.Available() {
		return false
	}
	return s.algorithm.IsValid()
}

// IsSecretKeyPresent checks if a secret key is configured in the current [SignatureConfig] instance.
//
// Returns:
//   - true if a secret key is configured, false otherwise, or false if the instance is not available.
func (s *SignatureConfig) IsSecretKeyPresent() bool {
	if !s.Available() {
		return false
	}
	return strutil.IsNotEmpty(s.secretKey)
}

// IsHeadersToSignPresent checks if there are headers configured to be included in the signature for the current [SignatureConfig] instance.
//
// Returns:
//   - true if there are headers to sign, false otherwise, or false if the instance is not available.
func (s *SignatureConfig) IsHeadersToSignPresent() bool {
	if !s.Available() {
		return false
	}
	return len(s.headersToSign) > 0
}

// IsHeaderToSignPresent checks if a specific header is configured to be included in the signature for the current [SignatureConfig] instance.
//
// Parameters:
//   - header: The header name to check.
//
// Returns:
//   - true if the specified header is configured to be included in the signature, false otherwise, or false if the instance is not available.
func (s *SignatureConfig) IsHeaderToSignPresent(header string) bool {
	if !s.Available() {
		return false
	}
	if strutil.IsEmpty(header) {
		return false
	}
	return slices.Contains(s.headersToSign, header)
}

// IsMaxAgePresent checks if a maximum age is configured in the current [SignatureConfig] instance.
//
// Returns:
//   - true if a maximum age is configured, false otherwise, or false if the instance is not available.
func (s *SignatureConfig) IsMaxAgePresent() bool {
	if !s.Available() {
		return false
	}
	return s.maxAge > 0
}

// RemainingMaxAge calculates the remaining duration before the current [SignatureConfig] instance is considered expired based on a given start time.
//
// A start time in the future is only accepted within the configured clock skew (see
// [SignatureConfig.MaxClockSkew]): up to that far ahead the full maximum age is returned, and any time
// further ahead returns 0 (expired). This stops a pre-dated signature from staying valid indefinitely.
// A zero-value &SignatureConfig{} has a skew of 0, so any future time is expired.
//
// Parameters:
//   - t: The start time from which to calculate the remaining duration.
//
// Returns:
//   - the remaining duration if the instance is available and the maximum age is configured, or 0 otherwise
//     (including when t is further in the future than the clock skew allows).
func (s *SignatureConfig) RemainingMaxAge(t time.Time) time.Duration {
	if !s.Available() || s.maxAge <= 0 || t.IsZero() {
		return 0
	}

	elapsed := time.Since(t)
	if elapsed < 0 {
		// t is in the future: tolerate only the configured clock skew.
		if -elapsed > max(s.maxClockSkew, 0) {
			return 0
		}
		return s.maxAge
	}

	remaining := s.maxAge - elapsed
	if remaining <= 0 {
		return 0
	}
	return remaining
}

// IsExpiredTime checks if a given [time.Time] value is considered expired based on
// the maximum age configured in the current [SignatureConfig] instance.
//
// Parameters:
//   - t: The [time.Time] value to check.
//
// Returns:
//   - true if the time is expired, false otherwise, or false if the instance is not available or the maximum age is not configured.
func (s *SignatureConfig) IsExpiredTime(t time.Time) bool {
	return s.RemainingMaxAge(t) <= 0
}

// IsExpiredUnix checks if a given Unix timestamp is considered expired based on
// the maximum age configured in the current [SignatureConfig] instance.
//
// Parameters:
//   - timestamp: The Unix timestamp to check.
//
// Returns:
//   - true if the timestamp is expired, false otherwise, or false if the instance is not available or the maximum age is not configured.
func (s *SignatureConfig) IsExpiredUnix(timestamp int64) bool {
	return s.IsExpiredTime(time.Unix(timestamp, 0))
}

// IsExpiredDuration checks if a given [time.Duration] value is considered expired based on
// the maximum age configured in the current [SignatureConfig] instance.
//
// Parameters:
//   - d: The [time.Duration] value to check.
//
// Returns:
//   - true if the duration is expired, false otherwise, or false if the instance is not available or the maximum age is not configured.
func (s *SignatureConfig) IsExpiredDuration(d time.Duration) bool {
	if !s.Available() {
		return false
	}
	if s.maxAge <= 0 {
		return false
	}
	return d > s.maxAge
}

// MaxClockSkew returns the clock skew tolerated for signature timestamps that lie in the future.
// [NewSignatureConfig] sets it to [DefaultSignatureMaxClockSkew]; a zero-value &SignatureConfig{}
// has a skew of 0, which is strict.
//
// Returns:
//   - The tolerated clock skew, or 0 if the instance is not available.
func (s *SignatureConfig) MaxClockSkew() time.Duration {
	if !s.Available() {
		return 0
	}
	return s.maxClockSkew
}

// WithMaxClockSkew sets the clock skew tolerated for signature timestamps that lie in the future.
// A timestamp further ahead than d is treated as expired (see [SignatureConfig.RemainingMaxAge]).
// A zero or negative d is strict: no future timestamp is accepted.
//
// Parameters:
//   - d: The tolerated clock skew.
//
// Returns:
//   - A pointer to the modified [SignatureConfig] instance (enabling method chaining).
func (s *SignatureConfig) WithMaxClockSkew(d time.Duration) *SignatureConfig {
	s.maxClockSkew = d
	return s
}

// IsLegacyAlgorithmsEnabled reports whether the legacy HMAC-SHA1 and HMAC-MD5 algorithms are allowed.
// They are disabled by default; see [SignatureConfig.EnableLegacyAlgorithms].
//
// Returns:
//   - true if legacy algorithms are enabled, false otherwise or if the instance is not available.
func (s *SignatureConfig) IsLegacyAlgorithmsEnabled() bool {
	return s.Available() && s.allowLegacyAlgorithms
}

// WithLegacyAlgorithms allows or forbids the legacy HMAC-SHA1 and HMAC-MD5 algorithms.
// When forbidden (the default), [ValidateSignatureConfig] rejects a config using either of them, which
// in turn makes [GenerateSignature], [GenerateSignatureFromRequest], [VerifySignature] and
// [VerifySignatureFromRequest] fail with a Bad Request.
//
// Parameters:
//   - allow: true to permit HMAC-SHA1 and HMAC-MD5.
//
// Returns:
//   - A pointer to the modified [SignatureConfig] instance (enabling method chaining).
func (s *SignatureConfig) WithLegacyAlgorithms(allow bool) *SignatureConfig {
	s.allowLegacyAlgorithms = allow
	return s
}

// EnableLegacyAlgorithms allows the legacy HMAC-SHA1 and HMAC-MD5 algorithms, for interoperability with
// older systems only. SHA-1 and MD5 are cryptographically weak; prefer [HMACSHA256] or stronger.
//
// Returns:
//   - A pointer to the modified [SignatureConfig] instance (enabling method chaining).
func (s *SignatureConfig) EnableLegacyAlgorithms() *SignatureConfig {
	return s.WithLegacyAlgorithms(true)
}

// NonceStore records which nonces have been used, so that a signed request can be accepted only once
// when replay protection is enabled (see [SignatureConfig.EnableReplayProtection]).
//
// SeenOrMark must be an atomic test-and-set: it reports whether nonce was already recorded and, when it
// was not, records it so that it is reported as seen until ttl has elapsed. Two concurrent calls with the
// same fresh nonce must never both return alreadySeen == false.
//
// A non-nil error means the store could not decide; the verifier then FAILS CLOSED and rejects the request.
//
// The default store is in memory and per-process. A deployment with several instances behind a load
// balancer MUST supply a shared implementation (for example backed by Redis) with
// [SignatureConfig.WithNonceStore], otherwise a request captured on one instance can be replayed against another.
type NonceStore interface {
	SeenOrMark(nonce string, ttl time.Duration) (alreadySeen bool, err error)
}

// memoryNonceStore is the default per-process [NonceStore]: a mutex-protected map from nonce to expiry,
// bounded by capacity. When a new nonce arrives and the map is at capacity, expired entries are swept
// first; if it is still full the call returns an error so the verifier fails closed instead of growing
// without bound or evicting live entries (which would reopen the replay window).
type memoryNonceStore struct {
	mu        sync.Mutex
	entries   map[string]time.Time // nonce -> expiry
	capacity  int
	now       func() time.Time // injectable clock; defaults to time.Now
	nextSweep time.Time        // opportunistic sweep of expired entries is due at or after this time
	minExpiry time.Time        // lower bound of all expiries; no entry can be expired before it
}

// newMemoryNonceStore creates an in-memory nonce store holding at most capacity unexpired nonces.
// A non-positive capacity falls back to [DefaultNonceStoreCapacity].
func newMemoryNonceStore(capacity int) *memoryNonceStore {
	if capacity <= 0 {
		capacity = DefaultNonceStoreCapacity
	}
	return &memoryNonceStore{
		entries:  make(map[string]time.Time),
		capacity: capacity,
		now:      time.Now,
	}
}

// SeenOrMark implements [NonceStore].
func (m *memoryNonceStore) SeenOrMark(nonce string, ttl time.Duration) (bool, error) {
	if ttl <= 0 {
		return false, errors.New("nonce store: ttl must be positive")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	now := m.now()
	if expiry, ok := m.entries[nonce]; ok {
		if expiry.After(now) {
			return true, nil
		}
		delete(m.entries, nonce) // expired: the nonce may be used again
	}

	// Opportunistic sweep, at most once per ttl, so expired entries never accumulate.
	if !now.Before(m.nextSweep) {
		m.sweepLocked(now)
		m.nextSweep = now.Add(ttl)
	}
	if len(m.entries) >= m.capacity {
		// Only worth an O(n) sweep if something can have expired by now.
		if !now.Before(m.minExpiry) {
			m.sweepLocked(now)
		}
		if len(m.entries) >= m.capacity {
			return false, fmt.Errorf("nonce store: capacity of %d unexpired nonces reached", m.capacity)
		}
	}

	expiry := now.Add(ttl)
	m.entries[nonce] = expiry
	if len(m.entries) == 1 || expiry.Before(m.minExpiry) {
		m.minExpiry = expiry
	}
	return false, nil
}

// sweepLocked removes expired entries and recomputes the lower expiry bound. The caller holds m.mu.
func (m *memoryNonceStore) sweepLocked(now time.Time) {
	var min time.Time
	for nonce, expiry := range m.entries {
		if !expiry.After(now) {
			delete(m.entries, nonce)
			continue
		}
		if min.IsZero() || expiry.Before(min) {
			min = expiry
		}
	}
	m.minExpiry = min
}

// len returns the number of stored entries (expired entries included until swept).
func (m *memoryNonceStore) len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.entries)
}

// newNonce returns 16 bytes from crypto/rand, hex-encoded (32 characters).
func newNonce() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

// IsReplayProtectionEnabled reports whether replay protection is enabled. It is off by default.
//
// Returns:
//   - true if replay protection is enabled, false otherwise or if the instance is not available.
func (s *SignatureConfig) IsReplayProtectionEnabled() bool {
	return s.Available() && s.replayProtection
}

// EnableReplayProtection turns on opt-in replay protection and, when no nonce store has been set, installs
// the default in-memory store ([DefaultNonceStoreCapacity] entries).
//
// Threat model: without it, a captured, validly signed request can be replayed any number of times until
// it passes max age. With it, the signer adds a random single-use nonce that is signed as part of the HMAC
// payload and sent in [HeaderXSignatureNonce], and the verifier accepts each nonce at most once for
// max age plus the clock skew. Verification order is: configuration, algorithm, expiry, HMAC (which covers
// the nonce), and only then the nonce store, so unauthenticated requests never touch it. A missing nonce,
// a nonce seen before, or a store error all reject the request (fail closed).
//
// It requires a positive max age and timestamps ([SignatureConfig.WithMaxAge], the defaults of
// [NewSignatureConfig]); [ValidateSignatureConfig] rejects an unbounded window. Both sides must enable it.
// The default store is per-process: multi-instance deployments MUST supply a shared store with
// [SignatureConfig.WithNonceStore].
//
// Returns:
//   - A pointer to the modified [SignatureConfig] instance (enabling method chaining).
func (s *SignatureConfig) EnableReplayProtection() *SignatureConfig {
	return s.WithReplayProtection(true)
}

// WithReplayProtection enables or disables replay protection; see [SignatureConfig.EnableReplayProtection].
// Enabling installs the default in-memory store when none is set; disabling keeps any store that was set.
//
// Parameters:
//   - enable: true to enable replay protection.
//
// Returns:
//   - A pointer to the modified [SignatureConfig] instance (enabling method chaining).
func (s *SignatureConfig) WithReplayProtection(enable bool) *SignatureConfig {
	s.replayProtection = enable
	if enable && s.nonceStore == nil {
		s.nonceStore = newMemoryNonceStore(DefaultNonceStoreCapacity)
	}
	return s
}

// WithNonceStore sets the [NonceStore] used when replay protection is enabled, replacing the default
// in-memory store. Use it to share seen-nonce state across instances. Passing nil removes the store; with
// replay protection enabled, [ValidateSignatureConfig] then rejects the configuration.
//
// Parameters:
//   - store: The nonce store to use.
//
// Returns:
//   - A pointer to the modified [SignatureConfig] instance (enabling method chaining).
func (s *SignatureConfig) WithNonceStore(store NonceStore) *SignatureConfig {
	s.nonceStore = store
	return s
}

// IsExpired always returns false.
//
// A [SignatureConfig] carries no timestamp, so there is nothing to measure the maximum age against.
// The previous implementation compared the age of "now" to itself (about zero) and therefore never
// reported expiry either; it is now an explicit, deterministic false.
//
// Deprecated: IsExpired never reports expiry. Use [SignatureConfig.IsExpiredTime] or
// [SignatureConfig.IsExpiredUnix] with the timestamp of the signature being checked.
//
// Returns:
//   - always false.
func (s *SignatureConfig) IsExpired() bool {
	// No timestamp is available on a config, so expiry cannot be determined here.
	return false
}

// IsCanonicalize checks if the canonicalization of the JSON payload before signing is enabled for the current [SignatureConfig] instance.
//
// Returns:
//   - true if canonicalization is enabled, false otherwise, or false if the instance is not available.
func (s *SignatureConfig) IsCanonicalize() bool {
	if !s.Available() {
		return false
	}
	return s.canonicalize
}

// WithSecretKey sets the secret key for the current [SignatureConfig] instance and returns the updated instance.
//
// Parameters:
//   - secret: The secret key to set.
//
// Returns:
//   - The updated [SignatureConfig] instance.
func (s *SignatureConfig) WithSecretKey(secret string) *SignatureConfig {
	s.secretKey = secret
	return s
}

// WithAlgorithm sets the signature algorithm for the current [SignatureConfig] instance and returns the updated instance.
//
// Parameters:
//   - algorithm: The signature algorithm to set.
//
// Returns:
//   - The updated [SignatureConfig] instance.
func (s *SignatureConfig) WithAlgorithm(algorithm SignatureAlgorithm) *SignatureConfig {
	s.algorithm = algorithm
	return s
}

// EnableIncludeTimestamp enables the inclusion of the timestamp in the current [SignatureConfig] instance and returns the updated instance.
//
// Returns:
//   - The updated [SignatureConfig] instance.
func (s *SignatureConfig) EnableIncludeTimestamp() *SignatureConfig {
	s.includeTimestamp = true
	return s
}

// DisableIncludeTimestamp disables the inclusion of the timestamp in the current [SignatureConfig] instance and returns the updated instance.
//
// Returns:
//   - The updated [SignatureConfig] instance.
func (s *SignatureConfig) DisableIncludeTimestamp() *SignatureConfig {
	s.includeTimestamp = false
	return s
}

// WithIncludeTimestamp sets the inclusion of the timestamp in the current [SignatureConfig] instance and returns the updated instance.
//
// Parameters:
//   - include: A boolean value indicating whether to include the timestamp.
//
// Returns:
//   - The updated [SignatureConfig] instance.
func (s *SignatureConfig) WithIncludeTimestamp(include bool) *SignatureConfig {
	s.includeTimestamp = include
	return s
}

// WithCanonicalize sets whether to canonicalize the JSON payload before signing for the current [SignatureConfig] instance and returns the updated instance.
//
// Parameters:
//   - enable: A boolean value indicating whether to canonicalize the JSON payload before signing.
//
// Returns:
//   - The updated [SignatureConfig] instance.
func (s *SignatureConfig) WithCanonicalize(enable bool) *SignatureConfig {
	s.canonicalize = enable
	return s
}

// EnableCanonicalize enables the canonicalization of the JSON payload before signing for the current [SignatureConfig] instance and returns the updated instance.
//
// Returns:
//   - The updated [SignatureConfig] instance.
func (s *SignatureConfig) EnableCanonicalize() *SignatureConfig {
	return s.WithCanonicalize(true)
}

// DisableCanonicalize disables the canonicalization of the JSON payload before signing for the current [SignatureConfig] instance and returns the updated instance.
//
// Returns:
//   - The updated [SignatureConfig] instance.
func (s *SignatureConfig) DisableCanonicalize() *SignatureConfig {
	return s.WithCanonicalize(false)
}

// WithMaxAge sets the maximum age for the current [SignatureConfig] instance and returns the updated instance.
//
// Parameters:
//   - age: The maximum age to set.
//
// Returns:
//   - The updated [SignatureConfig] instance.
func (s *SignatureConfig) WithMaxAge(age time.Duration) *SignatureConfig {
	s.maxAge = age
	return s
}

// WithHeader adds a single header to be included in the signature for the current [SignatureConfig] instance and returns the updated instance.
//
// Parameters:
//   - key: The name of the header to include in the signature.
//
// Returns:
//   - The updated [SignatureConfig] instance.
func (s *SignatureConfig) WithHeader(header string) *SignatureConfig {
	if strutil.IsEmpty(header) {
		return s
	}
	s.headersToSign = append(s.headersToSign, header)
	return s
}

// WithHeaders adds multiple headers to be included in the signature for the current [SignatureConfig] instance and returns the updated instance.
//
// Parameters:
//   - headers: A variadic list of header names to include in the signature. Each header will be added using the [WithHeader] method.
//
// Returns:
//   - The updated [SignatureConfig] instance.
func (s *SignatureConfig) WithHeaders(headers ...string) *SignatureConfig {
	for _, header := range headers {
		s.WithHeader(header)
	}
	return s
}

// ReleaseHeaders clears all headers to be included in the signature for the current [SignatureConfig] instance and returns the updated instance.
//
// Returns:
//   - The updated [SignatureConfig] instance.
func (s *SignatureConfig) ReleaseHeaders() *SignatureConfig {
	s.headersToSign = []string{}
	return s
}

// WithIgnoringSignatureFields sets the JSON field names excluded when computing the signature
// for the current [SignatureConfig] instance and returns the updated instance.
//
// This is distinct from [WithIgnoringResponseFields]: it only affects the body used to compute
// the signature (e.g. via [ApplySignature]) and has no effect on the JSON ultimately written to
// the client.
//
// Parameters:
//   - fields: A variadic list of JSON field names to exclude when computing the signature.
//
// Returns:
//   - The updated [SignatureConfig] instance.
func (s *SignatureConfig) WithIgnoringSignatureFields(fields ...string) *SignatureConfig {
	if len(fields) == 0 {
		return s
	}
	for _, field := range fields {
		if strutil.IsEmpty(field) {
			continue
		}
		s.ignoringSignatureFields = append(s.ignoringSignatureFields, field)
	}
	return s
}

// WithIgnoringResponseFields sets the JSON field names excluded from the response written to the
// client for the current [SignatureConfig] instance and returns the updated instance.
//
// This is distinct from [WithIgnoringSignatureFields]: it only affects the JSON serialized to the
// [http.ResponseWriter] (e.g. via [WriteJSON]) and has no effect on the signature computation.
//
// Parameters:
//   - fields: A variadic list of JSON field names to exclude from the response.
//
// Returns:
//   - The updated [SignatureConfig] instance.
func (s *SignatureConfig) WithIgnoringResponseFields(fields ...string) *SignatureConfig {
	if len(fields) == 0 {
		return s
	}
	for _, field := range fields {
		if strutil.IsEmpty(field) {
			continue
		}
		s.ignoringResponseFields = append(s.ignoringResponseFields, field)
	}
	return s
}

// HasHeader checks if a specific header is included in the signature for the current [SignatureConfig] instance and returns a boolean result.
//
// Parameters:
//   - header: The name of the header to check.
//
// Returns:
//   - A boolean value indicating whether the specified header is included in the signature.
func (s *SignatureConfig) HasHeader(header string) bool {
	return slices.Contains(s.headersToSign, header)
}

// RemoveHeaderIgnorecase removes a specific header from the list of headers to be included in the signature for the current [SignatureConfig] instance and returns the updated instance.
//
// Parameters:
//   - header: The name of the header to remove (case-insensitive).
//
// Returns:
//   - The updated [SignatureConfig] instance.
func (s *SignatureConfig) RemoveHeaderIgnorecase(header string) *SignatureConfig {
	if strutil.IsEmpty(header) {
		return s
	}
	for i, h := range s.headersToSign {
		if strings.EqualFold(h, header) {
			s.headersToSign = append(s.headersToSign[:i], s.headersToSign[i+1:]...)
			break
		}
	}
	return s
}

// Respond generates a map representation of the current [SignatureConfig] instance, including only the fields that are present.
//
// Returns:
//   - A map containing the present fields and their corresponding values.
func (s *SignatureConfig) Respond() map[string]any {
	m := make(map[string]any)
	m["include_timestamp"] = s.includeTimestamp
	m["canonicalize"] = s.canonicalize
	if s.IsSecretKeyPresent() {
		m["secret_key"] = "******"
	}
	if s.IsAlgorithmPresent() {
		m["algorithm"] = s.algorithm.String()
	}
	if s.IsHeadersToSignPresent() {
		m["headers_to_sign"] = s.headersToSign
	}
	if s.IsMaxAgePresent() {
		m["max_age"] = s.maxAge.String()
	}
	if s.maxClockSkew != DefaultSignatureMaxClockSkew {
		m["max_clock_skew"] = s.maxClockSkew.String()
	}
	if s.allowLegacyAlgorithms {
		m["allow_legacy_algorithms"] = true
	}
	if s.replayProtection {
		m["replay_protection"] = true
	}
	return m
}

// RespondIgnoring generates a map representation of the current [SignatureConfig] instance, excluding the specified level 1 fields.
//
// Parameters:
//   - level1fields: A variadic list of field names to exclude from the resulting map.
//
// Returns:
//   - A map containing the present fields and their corresponding values, excluding the specified fields.
func (s *SignatureConfig) RespondIgnoring(level1fields ...string) map[string]any {
	if len(level1fields) == 0 {
		return s.Respond()
	}
	// Respond() returns the shared cached map; clone it so deleting fields
	// below doesn't mutate the cache backing future Respond()/JSON() calls.
	src := s.Respond()
	m := make(map[string]any, len(src))
	maps.Copy(m, src)
	for _, field := range level1fields {
		if strutil.IsEmpty(field) {
			continue
		}
		delete(m, field)
	}
	return m
}

// RespondOnly generates a map representation of the current [SignatureConfig] instance, including only the specified level 1 fields.
//
// Parameters:
//   - level1fields: A variadic list of field names to include in the resulting map.
//
// Returns:
//   - A map containing only the specified fields and their corresponding values.
func (s *SignatureConfig) RespondOnly(level1fields ...string) map[string]any {
	if len(level1fields) == 0 {
		return s.Respond()
	}
	// Respond() returns the shared cached map; clone it so deleting fields
	// below doesn't mutate the cache backing future Respond()/JSON() calls.
	src := s.Respond()
	m := make(map[string]any, len(src))
	maps.Copy(m, src)
	for key := range m {
		found := false
		for _, field := range level1fields {
			if strutil.IsEmpty(field) {
				continue
			}
			if key == field {
				found = true
				break
			}
		}
		if !found {
			delete(m, key)
		}
	}
	return m
}

// JSON generates a JSON string representation of the current [SignatureConfig] instance, including only the fields that are present.
//
// Returns:
//   - A JSON string containing the present fields and their corresponding values.
func (s *SignatureConfig) JSON() string {
	return jsonpass(s.Respond())
}

// JSONOnly generates a JSON string representation of the current [SignatureConfig] instance, including only the specified level 1 fields.
//
// Parameters:
//   - level1fields: A variadic list of field names to include in the resulting JSON string.
//
// Returns:
//   - A JSON string containing only the specified fields and their corresponding values.
func (s *SignatureConfig) JSONOnly(level1fields ...string) string {
	return jsonpass(s.RespondOnly(level1fields...))
}

// JSONIgnoring generates a JSON string representation of the current [SignatureConfig] instance, excluding the specified level 1 fields.
//
// Parameters:
//   - level1fields: A variadic list of field names to exclude from the resulting JSON string.
//
// Returns:
//   - A JSON string containing the present fields and their corresponding values, excluding the specified fields.
func (s *SignatureConfig) JSONIgnoring(level1fields ...string) string {
	return jsonpass(s.RespondIgnoring(level1fields...))
}

// JSONPretty generates a pretty-printed JSON string representation of the current [SignatureConfig] instance, including only the fields that are present.
//
// Returns:
//   - A pretty-printed JSON string containing the present fields and their corresponding values.
func (s *SignatureConfig) JSONPretty() string {
	return jsonpretty(s.Respond())
}

// JSONPrettyOnly generates a pretty-printed JSON string representation of the current [SignatureConfig] instance, including only the specified level 1 fields.
//
// Parameters:
//   - level1fields: A variadic list of field names to include in the resulting pretty-printed JSON string.
//
// Returns:
//   - A pretty-printed JSON string containing only the specified fields and their corresponding values.
func (s *SignatureConfig) JSONPrettyOnly(level1fields ...string) string {
	return jsonpretty(s.RespondOnly(level1fields...))
}

// JSONPrettyIgnoring generates a pretty-printed JSON string representation of the current [SignatureConfig] instance, excluding the specified level 1 fields.
//
// Parameters:
//   - level1fields: A variadic list of field names to exclude from the resulting pretty-printed JSON string.
//
// Returns:
//   - A pretty-printed JSON string containing the present fields and their corresponding values, excluding the specified fields.
func (s *SignatureConfig) JSONPrettyIgnoring(level1fields ...string) string {
	return jsonpretty(s.RespondIgnoring(level1fields...))
}

// Equal compares the current [SignatureConfig] instance with another [SignatureConfig] instance for equality.
//
// Every field is compared unconditionally: algorithm, secret key, maximum age, timestamp inclusion,
// canonicalization, clock skew, legacy-algorithm opt-in, and the three string lists (headers to sign,
// ignored signature fields, ignored response fields). The lists are compared in order; a nil list and
// an empty list are equal. Two nil configs are equal; a nil and a non-nil config are not.
//
// The replay-protection flag is compared, but the nonce store is deliberately excluded: it is a
// (possibly shared, stateful) service object, and two configs that differ only in which store instance
// they use describe the same signing and verification policy.
//
// Parameters:
//   - other: The other [SignatureConfig] instance to compare against.
//
// Returns:
//   - A boolean value indicating whether the two [SignatureConfig] instances are equal.
func (s *SignatureConfig) Equal(other *SignatureConfig) bool {
	if s == nil && other == nil {
		return true
	}
	if s == nil || other == nil {
		return false
	}
	return s.algorithm == other.algorithm &&
		s.secretKey == other.secretKey &&
		s.maxAge == other.maxAge &&
		s.includeTimestamp == other.includeTimestamp &&
		s.canonicalize == other.canonicalize &&
		s.maxClockSkew == other.maxClockSkew &&
		s.allowLegacyAlgorithms == other.allowLegacyAlgorithms &&
		s.replayProtection == other.replayProtection &&
		slices.Equal(s.headersToSign, other.headersToSign) &&
		slices.Equal(s.ignoringSignatureFields, other.ignoringSignatureFields) &&
		slices.Equal(s.ignoringResponseFields, other.ignoringResponseFields)
}

// String generates a string representation of the current [SignatureConfig] instance.
//
// Returns:
//   - A string containing the present fields and their corresponding values. Sensitive fields like the secret key are masked.
func (s *SignatureConfig) String() string {
	if s == nil {
		return ""
	}
	sw := strchain.New()
	sw.AppendF("algorithm=%s", s.algorithm.String())
	sw.Space()
	if s.IsSecretKeyPresent() {
		sw.AppendF("secret_key=%s", "******")
		sw.Space()
	}
	if s.IsHeadersToSignPresent() {
		sw.AppendF("headers_to_sign=%v", strings.Join(s.headersToSign, ","))
		sw.Space()
	}
	sw.AppendF("max_age=%s", s.maxAge.String())
	sw.Space()
	sw.AppendF("include_timestamp=%t", s.includeTimestamp)
	return sw.String()
}

// Logging logs the current [SignatureConfig] instance using the provided logger or the default logger if none is provided.
//
// Parameters:
//   - logger: An optional [slogger.Logger] instance to use for logging.
//
// Returns:
//   - The current [SignatureConfig] instance.
func (s *SignatureConfig) Logging(logger ...*slogger.Logger) *SignatureConfig {
	if s == nil {
		return s
	}
	l := slogger.S()
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0]
	}

	msg := "replify::signature_config::logging"

	child := l.With()
	child.WithCaller(true).WithCallerSkip(3)

	logAtLevel(child, slogger.InfoLevel, msg, slogger.JSON("SIGNATURE_CONFIG", s.Respond()))
	return s
}

// LoggingIgnoring logs the current [SignatureConfig] instance while ignoring the specified fields using the default logger.
//
// Parameters:
//   - level1fields: A variadic list of field names to ignore in the log output.
//
// Returns:
//   - The current [SignatureConfig] instance.
func (s *SignatureConfig) LoggingIgnoring(level1fields ...string) *SignatureConfig {
	if s == nil {
		return s
	}
	l := slogger.S()
	child := l.With()
	child.WithCaller(true).WithCallerSkip(3)

	msg := "replify::signature_config::logging_ignoring"

	logAtLevel(child, slogger.InfoLevel, msg, slogger.JSON("SIGNATURE_CONFIG", s.RespondIgnoring(level1fields...)))
	return s
}

// Slogging logs the string representation of the current [SignatureConfig] instance using the provided logger or the default logger if none is provided.
//
// Parameters:
//   - logger: An optional [slogger.Logger] instance to use for logging.
//
// Returns:
//   - The current [SignatureConfig] instance.
func (s *SignatureConfig) Slogging(logger ...*slogger.Logger) *SignatureConfig {
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

// Clone creates a deep copy of the current [SignatureConfig] instance.
//
// Every field is copied, including the ignored signature and response field lists, and all slices are
// duplicated so that mutating the clone never affects the original (and vice versa). The nonce store is
// the one exception: it is copied by reference, so a clone and its original share the same seen-nonce
// state, which is what a replay check requires.
//
// Returns:
//   - A new [SignatureConfig] instance that is a copy of the current instance.
func (s *SignatureConfig) Clone() *SignatureConfig {
	if s == nil {
		return nil
	}
	clone := &SignatureConfig{
		secretKey:               s.secretKey,
		algorithm:               s.algorithm,
		headersToSign:           slices.Clone(s.headersToSign),
		maxAge:                  s.maxAge,
		includeTimestamp:        s.includeTimestamp,
		ignoringSignatureFields: slices.Clone(s.ignoringSignatureFields),
		ignoringResponseFields:  slices.Clone(s.ignoringResponseFields),
		canonicalize:            s.canonicalize,
		maxClockSkew:            s.maxClockSkew,
		allowLegacyAlgorithms:   s.allowLegacyAlgorithms,
		replayProtection:        s.replayProtection,
		nonceStore:              s.nonceStore, // shared by reference on purpose: it is a service object, not config data
	}
	return clone
}

// getHashSignature returns the hash function corresponding to the given signature algorithm.
//
// Parameters:
//   - algorithm: The signature algorithm [SignatureAlgorithm] for which to retrieve the hash function.
//
// Returns:
//   - A function that creates a new [hash.Hash] instance for the specified algorithm.
//   - A [wrapper] instance indicating success or failure.
func getHashSignature(algorithm SignatureAlgorithm) (func() hash.Hash, *wrapper) {
	switch algorithm {
	// Standard SHA-2 (32-bit) family
	case HMACSHA256:
		return sha256.New, New().OK()
	case HMACSHA224:
		return sha256.New224, New().OK()
	// Standard SHA-2 (64-bit) family
	case HMACSHA512:
		return sha512.New, New().OK()
	case HMACSHA384:
		return sha512.New384, New().OK()
	case HMACSHA512_256:
		return sha512.New512_256, New().OK()
	// Legacy algorithms (kept for backward compatibility)
	// HMAC-SHA1 and HMAC-MD5 are considered legacy algorithms and are included for backward compatibility.
	case HMACSHA1:
		return sha1.New, New().OK()
	case HMACMD5:
		return md5.New, New().OK()
	default:
		return nil, New().
			BadRequest().
			WithMessagef("unsupported signature algorithm: %s", algorithm.String())
	}
}

// ValidateSignatureConfig checks the validity of the provided [SignatureConfig] instance and returns a [wrapper] containing the validation result.
//
// Besides requiring an algorithm and a secret key, it rejects the legacy [HMACSHA1] and [HMACMD5]
// algorithms with a Bad Request unless they were enabled through [SignatureConfig.EnableLegacyAlgorithms].
// Because every generate and verify function calls it first, the gate applies to all of them.
//
// When replay protection is enabled it also requires a positive max age, timestamps and a nonce store.
//
// Parameters:
//   - config: The [SignatureConfig] instance to validate.
//
// Returns:
//   - A [wrapper] instance containing the validation result.
func ValidateSignatureConfig(config *SignatureConfig) *wrapper {
	if config == nil {
		return New().BadRequest().WithMessage("signature configuration is missing")
	}
	if !config.IsValid() {
		if !config.IsAlgorithmPresent() {
			return New().BadRequest().WithMessage("signature algorithm is missing")
		}
		if !config.IsSecretKeyPresent() {
			return New().BadRequest().WithMessage("signature secret key is missing")
		}
	}
	if (config.algorithm == HMACSHA1 || config.algorithm == HMACMD5) && !config.allowLegacyAlgorithms {
		return New().BadRequest().
			WithMessagef("signature algorithm %s is a legacy algorithm and is disabled by default; call EnableLegacyAlgorithms() on the SignatureConfig to allow it",
				config.algorithm.String())
	}
	if config.replayProtection {
		if config.maxAge <= 0 {
			return New().BadRequest().
				WithMessage("replay protection requires a positive max age: the nonce window would be unbounded; call WithMaxAge")
		}
		if !config.includeTimestamp {
			return New().BadRequest().
				WithMessage("replay protection requires signature timestamps: without them a signature never expires; call EnableIncludeTimestamp")
		}
		if config.nonceStore == nil {
			return New().BadRequest().
				WithMessage("replay protection requires a nonce store; call WithNonceStore or EnableReplayProtection")
		}
	}
	return New().OK().WithMessage("validated signature configuration successfully")
}

// GenerateSignature generates a signature for the given [SignatureConfig] and request body.
//
// Parameters:
//   - config: The [SignatureConfig] instance containing the signature configuration.
//   - body: The request body to be signed.
//
// Returns:
//   - A [signature] instance containing the generated signature.
//   - A [wrapper] instance indicating success or failure.
func GenerateSignature(config *SignatureConfig, body []byte) (s *signature, w *wrapper) {
	v := ValidateSignatureConfig(config)
	if v.IsError() {
		return nil, v
	}
	var h func() hash.Hash
	h, w = getHashSignature(config.Algorithm())
	if w.IsError() {
		return nil, w
	}

	// Canonicalize the request body if the configuration specifies it.
	// This ensures consistent JSON formatting for signature generation.
	if config.IsCanonicalize() {
		if canonicalBody, wc := canonicalJSON(body); wc.IsError() {
			wc.Slogging()
			return nil, wc
		} else {
			body = canonicalBody
		}
	}

	s = NewSignature().
		WithAlgorithm(config.Algorithm())

	var ts int64
	if config.IsIncludeTimestamp() {
		ts = time.Now().Unix()
		s.WithTimeUnix(ts)
	}

	// With replay protection, sign a fresh random nonce together with the payload.
	var nonce string
	if config.IsReplayProtectionEnabled() {
		n, err := newNonce()
		if err != nil {
			return nil, New().InternalServerError().WithErrorAck(err).WithMessage("failed to generate signature nonce")
		}
		nonce = n
		s.WithNonce(nonce)
	}

	// Build the signature payload by concatenating the body, timestamp, and headers (if any).
	// The payload is constructed in a deterministic manner to ensure that both the sender and receiver
	// can generate the same signature for the same request content.
	payload := buildSignaturePayload(body, ts, nil, nonce)

	mac := hmac.New(h, []byte(config.SecretKey()))
	mac.Write(payload)
	sum := mac.Sum(nil)
	signatureValue := base64.StdEncoding.EncodeToString(sum)
	s.WithTextValue(signatureValue)

	return s, New().
		OK().
		WithBody(s.Respond()).
		WithMessagef("generated signature for algorithm %s successfully", config.Algorithm().String())
}

// GenerateSignatureFromRequest generates a signature for the given [SignatureConfig] and HTTP request.
//
// Parameters:
//   - r: The HTTP request containing the headers to be signed.
//   - config: The [SignatureConfig] instance containing the signature configuration.
//   - body: The request body to be signed.
//
// Returns:
//   - A [signature] instance containing the generated signature.
//   - A [wrapper] instance indicating success or failure.
func GenerateSignatureFromRequest(r *http.Request, config *SignatureConfig, body []byte) (s *signature, w *wrapper) {
	v := ValidateSignatureConfig(config)
	if v.IsError() {
		return nil, v
	}
	var h func() hash.Hash
	h, w = getHashSignature(config.Algorithm())
	if w.IsError() {
		return nil, w
	}

	// Canonicalize the request body if the configuration specifies it.
	// This ensures consistent JSON formatting for signature generation.
	if config.IsCanonicalize() {
		if canonicalBody, wc := canonicalJSON(body); wc.IsError() {
			wc.Slogging()
			return nil, wc
		} else {
			body = canonicalBody
		}
	}

	s = NewSignature().
		WithAlgorithm(config.Algorithm())

	var ts int64
	if config.IsIncludeTimestamp() {
		ts = time.Now().Unix()
		s.WithTimeUnix(ts)
	}

	headers := make(map[string]string)
	if config.IsHeadersToSignPresent() {
		for _, header := range config.HeadersToSign() {
			val := r.Header.Get(header)
			s.WithHeader(header, val)
			headers[header] = val
		}
	}

	// With replay protection, sign a fresh random nonce together with the payload.
	var nonce string
	if config.IsReplayProtectionEnabled() {
		n, err := newNonce()
		if err != nil {
			return nil, New().InternalServerError().WithErrorAck(err).WithMessage("failed to generate signature nonce")
		}
		nonce = n
		s.WithNonce(nonce)
	}

	// Build the signature payload by concatenating the body, timestamp, headers and nonce in a canonical form.
	payload := buildSignaturePayload(body, ts, headers, nonce)

	mac := hmac.New(h, []byte(config.SecretKey()))
	mac.Write(payload)
	sum := mac.Sum(nil)
	signatureValue := base64.StdEncoding.EncodeToString(sum)
	s.WithTextValue(signatureValue)

	return s, New().
		OK().
		WithBody(s.Respond()).
		WithMessagef("generated signature from request for algorithm %s successfully", config.Algorithm().String())
}

// SignatureFromHeader builds a [signature] instance from the X-Signature, X-Signature-Algorithm,
// X-Signature-Timestamp and (when present) X-Signature-Nonce headers, mirroring the headers written by [ApplySignatureHeader].
// It is the natural counterpart used to recover a signature before calling [VerifySignature] or
// [VerifySignatureFromRequest].
//
// Parameters:
//   - header: The HTTP header collection to read the signature from (e.g. r.Header).
//
// Returns:
//   - A [signature] instance populated from the header values present in `header`. Fields absent
//     from the header are left unset.
func SignatureFromHeader(header http.Header) *signature {
	s := NewSignature()
	if header == nil {
		return s
	}
	if v := header.Get(HeaderXSignature.String()); strutil.IsNotEmpty(v) {
		s.WithValue(v)
	}
	if a := header.Get(HeaderXSignatureAlgorithm.String()); strutil.IsNotEmpty(a) {
		s.WithAlgorithm(SignatureAlgorithm(a))
	}
	if ts := header.Get(HeaderXSignatureTimestamp.String()); strutil.IsNotEmpty(ts) {
		s.WithTimeUnix(conv.Int64OrDefault(ts, 0))
	}
	if n := header.Get(HeaderXSignatureNonce.String()); strutil.IsNotEmpty(n) {
		s.WithNonce(n)
	}
	return s
}

// verifyNoncePresent runs the replay check that happens before the HMAC: a missing nonce is rejected
// without touching the store. It returns nil when the nonce is acceptable so far.
func verifyNoncePresent(config *SignatureConfig, provided *signature) *wrapper {
	if config.IsReplayProtectionEnabled() && !provided.IsNoncePresent() {
		return New().Unauthorized().WithMessage("signature nonce is missing: replay protection is enabled")
	}
	return nil
}

// checkReplay marks the nonce of an already authenticated signature as used. It must only be called after
// the HMAC has been verified, so unauthenticated requests never reach the store. It returns nil when the
// nonce was fresh, and an Unauthorized wrapper for a replay or a store failure (fail closed).
func checkReplay(config *SignatureConfig, provided *signature) *wrapper {
	if !config.IsReplayProtectionEnabled() {
		return nil
	}
	ttl := config.MaxAge() + max(config.MaxClockSkew(), 0)
	seen, err := config.nonceStore.SeenOrMark(provided.Nonce(), ttl)
	if err != nil {
		return New().Unauthorized().WithErrorAck(err).
			WithMessage("signature replay check failed: nonce store error, request rejected")
	}
	if seen {
		return New().Unauthorized().WithMessage("signature replay detected: nonce has already been used")
	}
	return nil
}

// VerifySignature verifies a previously generated [signature] against the given body using the
// provided [SignatureConfig]. Verification proceeds in order: configuration validity, presence of
// the signature to verify, algorithm agreement, expiry (when [SignatureConfig.IsIncludeTimestamp]
// is enabled), and finally recomputation of the HMAC via [GenerateSignature] compared against the
// provided value using a constant-time comparison ([hmac.Equal]) to avoid timing attacks.
//
// Parameters:
//   - config: The [SignatureConfig] instance containing the signature configuration.
//   - body: The request/response body that was originally signed.
//   - provided: The [signature] to verify, typically obtained via [SignatureFromHeader].
//
// Returns:
//   - true if `provided` is a valid signature of `body` for the given `config`, false otherwise.
//   - A [wrapper] instance describing the verification outcome.
func VerifySignature(config *SignatureConfig, body []byte, provided *signature) (bool, *wrapper) {
	v := ValidateSignatureConfig(config)
	if v.IsError() {
		return false, v
	}
	if config.IsHeadersToSignPresent() {
		return false, New().
			BadRequest().
			WithMessage("headers are configured to be signed, use VerifySignatureFromRequest instead")
	}
	if !provided.IsValuePresent() {
		return false, New().
			BadRequest().
			WithMessage("signature to verify is missing")
	}
	if provided.IsAlgorithmPresent() && !provided.Algorithm().Equals(config.Algorithm()) {
		return false, New().
			Unauthorized().
			WithMessagef("signature algorithm mismatch: expected %s, got %s",
				config.Algorithm().String(),
				provided.Algorithm().String())
	}
	if config.IsIncludeTimestamp() && provided.IsTimestampPresent() {
		if config.IsExpiredUnix(provided.Timestamp()) {
			return false, New().Unauthorized().WithMessage("signature has expired")
		}
	}
	if wn := verifyNoncePresent(config, provided); wn != nil {
		return false, wn
	}

	// Canonicalize the JSON body if the configuration requires it. This ensures consistent
	// formatting for signature verification, avoiding mismatches due to differences in
	// whitespace, key ordering, or other non-semantic variations in the JSON structure.
	if config.IsCanonicalize() {
		if canonicalBody, wc := canonicalJSON(body); wc.IsError() {
			wc.Slogging()
			return false, wc
		} else {
			body = canonicalBody
		}
	}

	var h func() hash.Hash
	h, w := getHashSignature(config.Algorithm())
	if w.IsError() {
		return false, w
	}

	var ts int64
	if config.IsIncludeTimestamp() && provided.IsTimestampPresent() {
		ts = provided.Timestamp()
	}

	// Build the payload that will be used for HMAC verification.
	// This includes the body, timestamp, and any headers if applicable.
	var nonce string
	if config.IsReplayProtectionEnabled() {
		nonce = provided.Nonce()
	}
	payload := buildSignaturePayload(body, ts, nil, nonce)
	mac := hmac.New(h, []byte(config.SecretKey()))
	mac.Write(payload)
	expectedValue := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expectedValue), []byte(provided.Value())) {
		return false, New().
			Unauthorized().
			WithMessage("signature verification failed: value mismatch or timestamp tampered")
	}

	// Only an authenticated signature may reach the nonce store.
	if wr := checkReplay(config, provided); wr != nil {
		return false, wr
	}

	return true, New().
		OK().
		WithMessagef("verified signature for algorithm %s successfully", config.Algorithm().String())
}

// VerifySignatureFromRequest verifies a [signature] against the given HTTP request and body using
// the provided [SignatureConfig]. In addition to the checks performed by [VerifySignature], it
// ensures that every header configured via [SignatureConfig.WithHeaders] still carries the same
// value recorded on `provided`, detecting header tampering that occurred after the signature was
// generated by [GenerateSignatureFromRequest].
//
// Parameters:
//   - r: The HTTP request carrying the current header values to cross-check.
//   - config: The [SignatureConfig] instance containing the signature configuration.
//   - body: The request/response body that was originally signed.
//   - provided: The [signature] to verify, typically obtained via [SignatureFromHeader].
//
// Returns:
//   - true if `provided` (and its recorded headers) are valid for the given request and `config`, false otherwise.
//   - A [wrapper] instance describing the verification outcome.
func VerifySignatureFromRequest(r *http.Request, config *SignatureConfig, body []byte, provided *signature) (bool, *wrapper) {
	v := ValidateSignatureConfig(config)
	if v.IsError() {
		return false, v
	}
	if !provided.IsValuePresent() {
		return false, New().BadRequest().WithMessage("signature to verify is missing")
	}
	if provided.IsAlgorithmPresent() && !provided.Algorithm().Equals(config.Algorithm()) {
		return false, New().
			Unauthorized().
			WithMessagef("signature algorithm mismatch: expected %s, got %s",
				config.Algorithm().String(),
				provided.Algorithm().String())
	}

	if config.IsIncludeTimestamp() && provided.IsTimestampPresent() {
		if config.IsExpiredUnix(provided.Timestamp()) {
			return false, New().Unauthorized().WithMessage("signature has expired")
		}
	}

	if r == nil {
		return false, New().BadRequest().WithMessage("http request is missing, cannot cross-check signed headers")
	}
	if wn := verifyNoncePresent(config, provided); wn != nil {
		return false, wn
	}

	if config.IsCanonicalize() {
		if canonicalBody, wc := canonicalJSON(body); wc.IsError() {
			wc.Slogging()
			return false, wc
		} else {
			body = canonicalBody
		}
	}

	var h func() hash.Hash
	h, w := getHashSignature(config.Algorithm())
	if w.IsError() {
		return false, w
	}

	var ts int64
	if config.IsIncludeTimestamp() && provided.IsTimestampPresent() {
		ts = provided.Timestamp()
	}

	headers := make(map[string]string)
	if config.IsHeadersToSignPresent() {
		if !provided.IsHeadersPresent() {
			return false, New().Unauthorized().WithMessage("signed headers are missing from the provided signature")
		}
		for _, header := range config.HeadersToSign() {
			originalVal, ok := provided.Headers()[header]
			if !ok {
				return false, New().Unauthorized().WithMessagef("signed header %q is missing from the provided signature", header)
			}
			currentVal := r.Header.Get(header)
			if originalVal != currentVal {
				return false, New().
					Unauthorized().
					WithMessagef("signed header %q has changed since the signature was generated", header)
			}
			headers[header] = originalVal
		}
	}

	var nonce string
	if config.IsReplayProtectionEnabled() {
		nonce = provided.Nonce()
	}
	payload := buildSignaturePayload(body, ts, headers, nonce)
	mac := hmac.New(h, []byte(config.SecretKey()))
	mac.Write(payload)
	expectedValue := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expectedValue), []byte(provided.Value())) {
		return false, New().
			Unauthorized().
			WithMessage("signature verification failed: value mismatch or payload tampered")
	}

	// Only an authenticated signature may reach the nonce store.
	if wr := checkReplay(config, provided); wr != nil {
		return false, wr
	}

	return true, New().
		OK().
		WithMessagef("verified signature from request for algorithm %s successfully", config.Algorithm().String())
}
