package replify

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
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
	m := s.Respond()
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
		headers:   s.headers,
	}
	return clone
}

// NewSignatureConfig creates a new instance of [SignatureConfig] with default values.
// It sets the default algorithm to HMAC-SHA256 ([HMACSHA256]), includes the timestamp by default,
// initializes an empty list of headers to sign, and sets the maximum age to 5 minutes.
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
// Parameters:
//   - t: The start time from which to calculate the remaining duration.
//
// Returns:
//   - the remaining duration if the instance is available and the maximum age is configured, or 0 otherwise.
func (s *SignatureConfig) RemainingMaxAge(t time.Time) time.Duration {
	if !s.Available() || s.maxAge <= 0 || t.IsZero() {
		return 0
	}

	elapsed := time.Since(t)
	if elapsed <= 0 {
		return s.maxAge
	}
	if elapsed >= s.maxAge {
		return 0
	}
	return s.maxAge - elapsed
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

// IsExpired checks if the current time is considered expired based on
// the maximum age configured in the current [SignatureConfig] instance.
//
// Returns:
//   - true if the current time is expired, false otherwise, or false if the instance is not available or the maximum age is not configured.
func (s *SignatureConfig) IsExpired() bool {
	if !s.Available() {
		return false
	}
	if s.maxAge <= 0 {
		return false
	}
	return s.IsExpiredDuration(time.Since(time.Now()))
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
