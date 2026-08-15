package replify

import (
	"github.com/polarixa/replify/pkg/slogger"
	"github.com/polarixa/replify/pkg/strchain"
)

// WithNext sets the next cursor value for pagination.
//
// Parameters:
//   - next: A string representing the cursor for the next page of results.
//
// Returns:
//   - A pointer to the [cursor] instance with the updated next cursor value.
func (c *cursor) WithNext(next string) *cursor {
	c.next = next
	return c
}

// WithPrevious sets the previous cursor value for pagination.
//
// Parameters:
//   - previous: A string representing the cursor for the previous page of results.
//
// Returns:
//   - A pointer to the [cursor] instance with the updated previous cursor value.
func (c *cursor) WithPrevious(previous string) *cursor {
	c.previous = previous
	return c
}

// WithHasNext sets the hasNext flag for pagination.
//
// Parameters:
//   - hasNext: A boolean indicating whether there is a next page of results.
//
// Returns:
//   - A pointer to the [cursor] instance with the updated hasNext flag.
func (c *cursor) WithHasNext(hasNext bool) *cursor {
	c.hasNext = hasNext
	return c
}

// WithHasPrevious sets the hasPrevious flag for pagination.
//
// Parameters:
//   - hasPrevious: A boolean indicating whether there is a previous page of results.
//
// Returns:
//   - A pointer to the [cursor] instance with the updated hasPrevious flag.
func (c *cursor) WithHasPrevious(hasPrevious bool) *cursor {
	c.hasPrevious = hasPrevious
	return c
}

// WithLimit sets the limit for pagination.
//
// Parameters:
//   - limit: An integer representing the maximum number of results per page.
//
// Returns:
//   - A pointer to the [cursor] instance with the updated limit.
func (c *cursor) WithLimit(limit int) *cursor {
	c.limit = limit
	return c
}

// Available checks if the cursor instance is available (not nil).
//
// Returns:
//   - A boolean indicating whether the cursor instance is available.
func (c *cursor) Available() bool {
	return c != nil
}

// Accepted checks if the cursor has a next page available.
//
// Returns:
//   - A boolean indicating whether there is a next page of results.
func (c *cursor) Accepted() bool {
	return c != nil && c.hasNext
}

// Rejected checks if the cursor has a previous page available.
//
// Returns:
//   - A boolean indicating whether there is a previous page of results.
func (c *cursor) Rejected() bool {
	return c != nil && c.hasPrevious
}

// RejectedOrAccepted checks if the cursor has either a next or previous page available.
//
// Returns:
//   - A boolean indicating whether there is a next or previous page of results.
func (c *cursor) RejectedOrAccepted() bool {
	return c != nil && (c.hasNext || c.hasPrevious)
}

// Next returns the next cursor value for pagination.
//
// Returns:
//   - A string representing the cursor for the next page of results.
//     Returns an empty string if the cursor instance is nil.
func (c *cursor) Next() string {
	if c == nil {
		return ""
	}
	return c.next
}

// Previous returns the previous cursor value for pagination.
//
// Returns:
//   - A string representing the cursor for the previous page of results.
//     Returns an empty string if the cursor instance is nil.
func (c *cursor) Previous() string {
	if c == nil {
		return ""
	}
	return c.previous
}

// Limit returns the limit for pagination.
//
// Returns:
//   - An integer representing the maximum number of results per page.
//     Returns 0 if the cursor instance is nil.
func (c *cursor) Limit() int {
	if c == nil {
		return 0
	}
	return c.limit
}

// HasNext checks if there is a next page of results available.
//
// Returns:
//   - A boolean indicating whether there is a next page of results.
//     Returns false if the cursor instance is nil.
func (c *cursor) HasNext() bool {
	if c == nil {
		return false
	}
	return c.hasNext
}

// HasPrevious checks if there is a previous page of results available.
//
// Returns:
//   - A boolean indicating whether there is a previous page of results.
//     Returns false if the cursor instance is nil.
func (c *cursor) HasPrevious() bool {
	if c == nil {
		return false
	}
	return c.hasPrevious
}

// Respond returns a map representation of the [cursor], suitable for JSON serialization.
//
// Returns:
//   - A map[string]any containing the cursor's next, previous, hasNext, hasPrevious, and limit values.
//     Returns nil if the cursor instance is nil.
func (c *cursor) Respond() map[string]any {
	if c == nil {
		return nil
	}
	return map[string]any{
		"next":         c.next,
		"previous":     c.previous,
		"has_next":     c.hasNext,
		"has_previous": c.hasPrevious,
		"limit":        c.limit,
	}
}

// JSON returns a JSON string representation of the [cursor], suitable for API responses.
//
// Returns:
//   - A JSON string containing the cursor's next, previous, hasNext, hasPrevious, and limit values.
//     Returns an empty string if the cursor instance is nil.
func (c *cursor) JSON() string {
	return jsonpass(c.Respond())
}

// JSONPretty returns a pretty-printed JSON string representation of the [cursor], suitable for API responses.
//
// Returns:
//   - A pretty-printed JSON string containing the cursor's next, previous, hasNext, hasPrevious, and limit values.
//     Returns an empty string if the cursor instance is nil.
func (c *cursor) JSONPretty() string {
	return jsonpretty(c.Respond())
}

// Equal checks if two [cursor] instances are equal by comparing their fields.
//
// Parameters:
//   - other: A pointer to another [cursor] instance to compare with.
//
// Returns:
//   - A boolean indicating whether the two cursor instances are equal.
func (c *cursor) Equal(other *cursor) bool {
	if c == nil && other == nil {
		return true
	}
	if c == nil || other == nil {
		return false
	}
	if c.next != other.next {
		return false
	}
	if c.previous != other.previous {
		return false
	}
	if c.hasNext != other.hasNext {
		return false
	}
	if c.hasPrevious != other.hasPrevious {
		return false
	}
	if c.limit != other.limit {
		return false
	}
	return true
}

// String returns a string representation of the [cursor], suitable for logging and debugging.
//
// Returns:
//   - A string containing the cursor's next, previous, hasNext, hasPrevious, and limit values.
//     Returns an empty string if the cursor instance is nil.
func (c *cursor) String() string {
	if c == nil {
		return ""
	}
	sw := strchain.New()
	sw.AppendF("next=%s", c.next)
	sw.Space()
	sw.AppendF("previous=%s", c.previous)
	sw.Space()
	sw.AppendF("has_next=%t", c.hasNext)
	sw.Space()
	sw.AppendF("has_previous=%t", c.hasPrevious)
	sw.Space()
	sw.AppendF("limit=%d", c.limit)
	return sw.String()
}

// Logging logs the current state of the [cursor] using the provided logger or the default logger if none is provided.
//
// Parameters:
//   - logger: An optional pointer to a [slogger.Logger] instance. If not provided, the default logger is used.
//
// Returns:
//   - A pointer to the [cursor] instance for method chaining.
func (c *cursor) Logging(logger ...*slogger.Logger) *cursor {
	if c == nil {
		return c
	}
	l := slogger.S()
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0]
	}

	msg := "replify::cursor::logging"

	child := l.With()
	child.WithCaller(true).WithCallerSkip(3)

	logAtLevel(child, slogger.InfoLevel, msg, slogger.JSON("CURSOR", c.Respond()))
	return c
}

// Slogging logs the current state of the [cursor] using structured logging with the provided logger or the default logger if none is provided.
//
// Parameters:
//   - logger: An optional pointer to a [slogger.Logger] instance. If not provided, the default logger is used.
//
// Returns:
//   - A pointer to the [cursor] instance for method chaining.
func (c *cursor) Slogging(logger ...*slogger.Logger) *cursor {
	if c == nil {
		return c
	}
	l := slogger.S()
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0]
	}

	child := l.With()
	child.WithCaller(true).WithCallerSkip(3)

	slogAtLevel(child, slogger.InfoLevel, c.String())
	return c
}

// Reply returns a [C] instance that wraps the current [cursor] instance, providing a structured way to access and manipulate pagination cursor details.
//
// Returns:
//   - A [C] instance that encapsulates the current [cursor] instance.
func (c *cursor) Reply() C {
	return C{cursor: c}
}

// ReplyPtr returns a pointer to a [C] instance that wraps the current [cursor] instance, providing a structured way to access and manipulate pagination cursor details.
//
// Returns:
//   - A pointer to a [C] instance that encapsulates the current [cursor] instance.
func (c *cursor) ReplyPtr() *C {
	return &C{cursor: c}
}
