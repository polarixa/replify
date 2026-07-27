package replify

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/polarixa/replify/pkg/slogger"
	"github.com/polarixa/replify/pkg/strchain"
	"github.com/polarixa/replify/pkg/strutil"
)

// WithCode sets the `code` field of the [header] instance.
//
// This function assigns the provided integer value to the `code` field of the [header]
// and returns the updated [header] instance, allowing for method chaining.
//
// Parameters:
//   - `v`: The integer value to set as the HTTP status code.
//
// Returns:
//   - The updated [header] instance with the `code` field set to the provided value.
func (h *header) WithCode(v int) *header {
	if v < 100 || v > 599 {
		v = http.StatusInternalServerError
	}
	h.code = v
	return h
}

// WithText sets the `text` field of the [header] instance.
//
// This function assigns the provided string value to the `text` field of the [header]
// and returns the updated [header] instance, allowing for method chaining.
//
// Parameters:
//   - `v`: The string value to set as the text message.
//
// Returns:
//   - The updated [header] instance with the `text` field set to the provided value.
func (h *header) WithText(v string) *header {
	h.text = v
	return h
}

// WithType sets the `Type` field of the [header] instance.
//
// This function assigns the provided string value to the `Type` field of the [header]
// and returns the updated [header] instance, allowing for method chaining.
//
// Parameters:
//   - `v`: The string value to set as the type of the header.
//
// Returns:
//   - The updated [header] instance with the `Type` field set to the provided value.
func (h *header) WithType(v string) *header {
	h.typez = v
	return h
}

// WithDescription sets the `description` field of the [header] instance.
//
// This function assigns the provided string value to the `description` field of the [header]
// and returns the updated [header] instance, allowing for method chaining.
//
// Parameters:
//   - `v`: The string value to set as the description of the header.
//
// Returns:
//   - The updated [header] instance with the `description` field set to the provided value.
func (h *header) WithDescription(v string) *header {
	h.description = v
	return h
}

// Available checks if the [header] instance is non-nil.
//
// This function ensures that the [header] instance is not nil before performing any operations.
// It returns `true` if the [header] is non-nil, and `false` if the [header] is nil.
//
// Returns:
//   - `true` if the [header] instance is not nil.
//   - `false` if the [header] instance is nil.
func (h *header) Available() bool {
	return h != nil
}

// IsCodePresent checks if the `code` field in the [header] instance is present and greater than zero.
//
// This function first checks if the [header] is available (non-nil), and then checks if the `code`
// field is greater than zero, indicating that it is present and valid.
//
// Returns:
//   - `true` if the `code` field is greater than zero.
//   - `false` if the `code` field is either not present (nil) or zero.
func (h *header) IsCodePresent() bool {
	return h.Available() && h.code > 0
}

// IsTextPresent checks if the `text` field in the [header] instance is present and not empty.
//
// This function verifies if the [header] is available and if the `text` field is not empty, using
// the `strutil.IsNotEmpty` utility to ensure the presence of the `text` field.
//
// Returns:
//   - `true` if the `text` field is non-empty.
//   - `false` if the `text` field is either not present (nil) or empty.
func (h *header) IsTextPresent() bool {
	return h.Available() && strutil.IsNotEmpty(h.text)
}

// IsTypePresent checks if the `Type` field in the [header] instance is present and not empty.
//
// This function checks if the [header] instance is available and if the `Type` field is not empty,
// utilizing the `strutil.IsNotEmpty` utility to determine whether the `Type` field contains a value.
//
// Returns:
//   - `true` if the `Type` field is non-empty.
//   - `false` if the `Type` field is either not present (nil) or empty.
func (h *header) IsTypePresent() bool {
	return h.Available() && strutil.IsNotEmpty(h.typez)
}

// IsDescriptionPresent checks if the `description` field in the [header] instance is present and not empty.
//
// This function ensures that the [header] is available and that the `description` field is not empty,
// using `strutil.IsNotEmpty` to check for non-emptiness.
//
// Returns:
//   - `true` if the `description` field is non-empty.
//   - `false` if the `description` field is either not present (nil) or empty.
func (h *header) IsDescriptionPresent() bool {
	return h.Available() && strutil.IsNotEmpty(h.description)
}

// Code retrieves the code value from the [header] instance.
//
// This function checks if the [header] instance is available (non-nil) before retrieving
// the `code` field. If the [header] instance is unavailable, it returns 0.
//
// Returns:
//   - The `code` as an integer if available.
//   - 0 if the [header] instance is unavailable.
func (h *header) Code() int {
	if !h.Available() {
		return 0
	}
	return h.code
}

// Text retrieves the text value from the [header] instance.
//
// This function checks if the [header] instance is available (non-nil) before retrieving
// the `text` field. If the [header] instance is unavailable, it returns an empty string.
//
// Returns:
//   - The `text` as a string if available.
//   - An empty string if the [header] instance is unavailable.
func (h *header) Text() string {
	if !h.Available() {
		return ""
	}
	return h.text
}

// Type retrieves the type value from the [header] instance.
//
// This function checks if the [header] instance is available (non-nil) before retrieving
// the `Type` field. If the [header] instance is unavailable, it returns an empty string.
//
// Returns:
//   - The `Type` as a string if available.
//   - An empty string if the [header] instance is unavailable.
func (h *header) Type() string {
	if !h.Available() {
		return ""
	}
	return h.typez
}

// Description retrieves the description value from the [header] instance.
//
// This function checks if the [header] instance is available (non-nil) before retrieving
// the `description` field. If the [header] instance is unavailable, it returns an empty string.
//
// Returns:
//   - The `description` as a string if available.
//   - An empty string if the [header] instance is unavailable.
func (h *header) Description() string {
	if !h.Available() {
		return ""
	}
	return h.description
}

// Respond generates a map representation of the [header] instance.
//
// This function checks if the [header] instance is available (non-nil) and includes the
// values of its fields in the returned map. Only the fields that are present (i.e., non-empty)
// are added to the map, ensuring a clean and concise response.
//
// Fields included in the response:
//   - `code`: The HTTP status code, if present and greater than 0.
//   - `text`: The associated text message, if present and not empty.
//   - `type`: The type of the header, if present and not empty.
//   - `description`: A description related to the header, if present and not empty.
//
// Returns:
//   - A `map[string]interface{}` containing the fields of the [header] instance that are present.
func (h *header) Respond() map[string]any {
	m := make(map[string]any)
	if !h.Available() {
		return m
	}
	if h.IsCodePresent() {
		m["code"] = h.code
	}
	if h.IsTextPresent() {
		m["text"] = h.text
	}
	if h.IsTypePresent() {
		m["type"] = h.typez
	}
	if h.IsDescriptionPresent() {
		m["description"] = h.description
	}
	return m
}

// JSON serializes the [header] instance into a compact JSON string.
//
// This function uses the `encoding.JSON` utility to create a compact JSON representation
// of the [header] instance. The resulting string contains only the key information, formatted
// with minimal whitespace, making it suitable for compact storage or transmission of header data.
//
// Returns:
//   - A compact JSON string representation of the [header] instance.
func (h *header) JSON() string {
	return jsonpass(h.Respond())
}

// JSONPretty serializes the [header] instance into a prettified JSON string.
//
// This function uses the `encoding.JSONPretty` utility to produce a formatted, human-readable
// JSON string representation of the [header] instance. The output is structured with indentation
// and newlines, making it ideal for inspecting header data in a clear, easy-to-read format, especially
// during debugging or development.
//
// Returns:
//   - A prettified JSON string representation of the [header] instance, formatted for improved readability.
func (h *header) JSONPretty() string {
	return jsonpretty(h.Respond())
}

// StatusText returns the standard HTTP status text corresponding to the `code` field of the [header] instance.
//
// This function checks if the [header] instance is non-nil and has a valid `code` field (greater than 0).
// If these conditions are met, it uses the `http.StatusText` function to retrieve the standard status text
// associated with the HTTP status code. If the [header] instance is nil or does not have a valid code, it returns an empty string.
//
// Returns:
//   - A string containing the standard HTTP status text corresponding to the `code` field of the [header] instance, or an empty string if the instance is nil or does not have a valid code.
func (h *header) StatusText() string {
	if h == nil || !h.IsCodePresent() {
		return ""
	}
	return fmt.Sprintf("%d (%s)", h.code, http.StatusText(h.code))
}

// Equal compares the current [header] instance with another [header] instance for equality.
//
// This function checks if both [header] instances are nil, in which case they are considered equal.
// If one instance is nil and the other is not, they are considered not equal. If both instances
// are non-nil, it compares their `code` and `text` fields for equality. The comparison focuses on
// these two fields as they are the primary identifiers of a header's status and message.
//
// Parameters:
//   - `other`: A pointer to another [header] instance to compare against.
//
// Returns:
//   - A boolean value indicating whether the two [header] instances are considered equal based on their `code` and `text` fields.
func (h *header) Equal(other *header) bool {
	if h == nil && other == nil {
		return true
	}
	if h == nil || other == nil {
		return false
	}
	return h.code == other.code &&
		strutil.EqualsIgnoreCase(other.text, h.text)
}

// String provides a string representation of the [header] instance.
//
// This function constructs a string that includes the values of the [header] instance's fields,
// such as `code`, `text`, `type`, and `description`. The output is formatted in a readable manner,
// making it useful for logging or debugging purposes. If the [header] instance is nil, it returns
// an empty string.
//
// Returns:
//   - A string representation of the [header] instance, including its fields, or an empty string if the instance is nil.
func (h *header) String() string {
	sw := strchain.New()
	if h == nil {
		return sw.String()
	}
	sw.AppendF("code=%d", h.code)
	if strutil.IsNotEmpty(h.text) {
		sw.Space()
		sw.AppendF("text=%s", h.text)
	}
	if strutil.IsNotEmpty(h.typez) {
		sw.Space()
		sw.AppendF("type=%s", h.typez)
	}
	if strutil.IsNotEmpty(h.description) {
		sw.Space()
		sw.AppendF("description=%s", h.description)
	}
	return sw.String()
}

// Logging dispatches a structured log entry for this response using [slogger] with the log message set to a default or custom message.
// The log level is automatically selected based on the HTTP status code range:
//
//   - 1xx → Debug  (informational)
//   - 2xx → Info   (success)
//   - 3xx → Warn   (redirection)
//   - 4xx → Error  (client error)
//   - 5xx → Error  (server error; [slogger.Logger.Fatal] is intentionally avoided because it calls os.Exit(1))
//   - other → Trace (no status code set)
//
// The log field key is "HEADER" and its value is the structured map returned
// by [header.Respond], serialized as JSON by the active formatter.
//
// # Thread-safety
//
// Logging is safe for concurrent use. The supplied logger is never mutated: a
// goroutine-local child is derived via [slogger.Logger.With] on every call so
// that the caller-skip adjustment and caller-enable flag stay local to the
// current goroutine. Concurrent callers sharing the same *[slogger.Logger]
// will not race. The header fields (e.g., contentType, contentLength) are read exactly once per call to give a consistent snapshot; header fields are expected to be immutable after construction via [Header] / [With*] options.
//
// # Caller reporting
//
// Caller information is always enabled for this call. callerSkip is set to 3
// to skip Logging, the slogger trampoline (Trace/Debug/Info/Warn/Error), and the caller of Logging, so the reported file and line resolve to the actual call site of Logging.
//
// Parameters:
//   - `logger`: optional *[slogger.Logger] to use. When omitted or nil, the
//     package-level global logger ([slogger.GlobalLogger]) is used.
//
// Returns:
//
// the receiver *header unchanged, enabling method chaining.
//
// Example:
//
//	replify.Header().
//	    WithContentType("application/json").
//	    WithContentLength(1234).
//	    Logging()
func (h *header) Logging(logger ...*slogger.Logger) *header {
	if h == nil {
		return h
	}
	l := slogger.S()
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0]
	}

	msg := "replify::header::logging"

	child := l.With()
	child.WithCaller(true).WithCallerSkip(3)

	logAtLevel(child, slogger.InfoLevel, msg, slogger.JSON("HEADER", h.Respond()))
	return h
}

// Slogging dispatches a structured log entry for this response using [slogger] with the log message set to the header's string representation.
// The log level is automatically selected based on the HTTP status code range:
//
//   - 1xx → Debug  (informational)
//   - 2xx → Info   (success)
//   - 3xx → Warn   (redirection)
//   - 4xx → Error  (client error)
//   - 5xx → Error  (server error; [slogger.Logger.Fatal] is intentionally avoided because it calls os.Exit(1))
//   - other → Trace (no status code set)
//
// The log field key is "HEADER" and its value is the structured map returned
// by [header.Respond], serialized as Text by the active formatter.
//
// # Thread-safety
//
// Slogging is safe for concurrent use. The supplied logger is never mutated: a
// goroutine-local child is derived via [slogger.Logger.With] on every call so
// that the caller-skip adjustment and caller-enable flag stay local to the
// current goroutine. Concurrent callers sharing the same *[slogger.Logger]
// will not race. The header fields (e.g., contentType, contentLength) are read exactly once per call to give a consistent snapshot; header fields are expected to be immutable after construction via [Header] / [With*] options.
//
// # Caller reporting
//
// Caller information is always enabled for this call. callerSkip is set to 3
// to skip Slogging, the slogger trampoline (Trace/Debug/Info/Warn/Error), and the caller of Slogging, so the reported file and line resolve to the actual call site of Slogging.
//
// Parameters:
//   - `logger`: optional *[slogger.Logger] to use. When omitted or nil, the
//     package-level global logger ([slogger.GlobalLogger]) is used.
//
// Returns:
//
//	the receiver *header unchanged, enabling method chaining.
//
// Example:
//
//	replify.Header().
//	    WithContentType("application/json").
//	    WithContentLength(1234).
//	    Slogging()
func (h *header) Slogging(logger ...*slogger.Logger) *header {
	if h == nil {
		return h
	}
	l := slogger.S()
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0]
	}

	child := l.With()
	child.WithCaller(true).WithCallerSkip(3)

	slogAtLevel(child, slogger.InfoLevel, h.String())
	return h
}

// Reply returns a new [H] instance that wraps the current [header] instance.
//
// This function creates a new [H] instance, which is a wrapper around the [header] struct.
// It allows for easier handling of response headers while maintaining the flexibility of the underlying [header] structure.
//
// Returns:
//   - A new [H] instance that wraps the current [header] instance.
func (h *header) Reply() H {
	return H{header: h}
}

// ReplyPtr returns a pointer to a new [H] instance that wraps the current [header] instance.
//
// This function creates a new [H] instance, which is a wrapper around the [header] struct,
// and returns a pointer to it. It allows for easier handling of response headers while maintaining
// the flexibility of the underlying [header] structure.
//
// Returns:
//   - A pointer to a new [H] instance that wraps the current [header] instance.
func (h *header) ReplyPtr() *H {
	return &H{header: h}
}

// String returns the string representation of the [HeaderType].
func (h HeaderType) String() string {
	return string(h)
}

// String returns the string representation of the [MediaType].
func (m MediaType) String() string {
	return string(m)
}

// Equal compares two [HeaderType] values for equality.
func (h HeaderType) Equal(other HeaderType) bool {
	return h.String() == other.String()
}

// EqualsIgnoreCase compares two [HeaderType] values for equality, ignoring case differences.
// This method is useful when you want to check if two header types are the same without considering the case of the characters.
func (h HeaderType) EqualsIgnoreCase(other HeaderType) bool {
	return strutil.EqualsIgnoreCase(string(h), string(other))
}

// Equals checks if the current [HeaderType] is equal to any of the provided [HeaderType] values.
// It returns true if the current [HeaderType] matches any of the provided values, and false otherwise.
// This method is useful for checking if a header type belongs to a set of predefined types.
func (h HeaderType) Equals(o ...HeaderType) bool {
	return slices.ContainsFunc(o, h.Equal)
}

// Equal compares two [MediaType] values for equality.
func (m MediaType) Equal(other MediaType) bool {
	return m.String() == other.String()
}

// EqualsIgnoreCase compares two [MediaType] values for equality, ignoring case differences.
// This method is useful when you want to check if two media types are the same without considering the case of the characters.
func (m MediaType) EqualsIgnoreCase(other MediaType) bool {
	return strutil.EqualsIgnoreCase(string(m), string(other))
}

// Equals checks if the current [MediaType] is equal to any of the provided [MediaType] values.
// It returns true if the current [MediaType] matches any of the provided values, and false otherwise.
// This method is useful for checking if a media type belongs to a set of predefined types.
func (m MediaType) Equals(o ...MediaType) bool {
	return slices.ContainsFunc(o, m.Equal)
}

// IsEmpty checks if the [HeaderType] is empty (i.e., has no value).
func (h HeaderType) IsEmpty() bool {
	return strutil.IsEmpty(string(h))
}

// IsEmpty checks if the [MediaType] is empty (i.e., has no value).
func (m MediaType) IsEmpty() bool {
	return strutil.IsEmpty(string(m))
}
