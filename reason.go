package replify

import (
	"slices"

	"github.com/polarixa/replify/pkg/slogger"
	"github.com/polarixa/replify/pkg/strchain"
	"github.com/polarixa/replify/pkg/strutil"
)

// ReasonCategoryOf retrieves the category associated with the given reason code.
//
// Parameters:
//   - code: The [ReasonCode] for which to retrieve the category.
//
// Returns:
//   - The [ReasonCategory] associated with the provided code, if it exists.
//   - A boolean value indicating whether the category was found:
//   - `true` if the category was found.
//   - `false` if the category was not found.
func ReasonCategoryOf(code ReasonCode) (ReasonCategory, bool) {
	r, ok := ReasonCodes[code]
	if !ok {
		return "", false
	}
	return r.category, true
}

// IsReasonKnown checks whether the given reason code is known.
//
// Parameters:
//   - code: The [ReasonCode] to check.
//
// Returns:
//   - A boolean value indicating whether the reason code is known:
//   - `true` if the reason code exists in the [ReasonCodes] map.
//   - `false` if the reason code does not exist in the [ReasonCodes] map.
func IsReasonKnown(code ReasonCode) bool {
	_, ok := ReasonCodes[code]
	return ok
}

// String returns the string representation of the [ReasonCode] instance.
//
// Returns:
//   - A string representing the [ReasonCode] instance.
func (r ReasonCode) String() string {
	return string(r)
}

// IsValid checks whether the [ReasonCode] instance represents a valid, non-empty code.
//
// Returns:
//   - A boolean value indicating whether the [ReasonCode] instance is valid:
//   - `true` if the code is non-empty.
//   - `false` if the code is empty.
func (r ReasonCode) IsValid() bool {
	return strutil.IsNotEmpty(string(r))
}

// Equals checks whether the [ReasonCode] instance matches any of the provided codes.
//
// Parameters:
//   - other: A variadic list of [ReasonCode] instances to compare against.
//
// Returns:
//   - A boolean value indicating whether the [ReasonCode] instance matches any of the provided codes:
//   - `true` if a match is found.
//   - `false` if no match is found or if no codes are provided.
func (r ReasonCode) Equals(other ...ReasonCode) bool {
	if len(other) == 0 {
		return false
	}
	return slices.Contains(other, r)
}

// String returns the string representation of the [ReasonCategory] instance.
//
// Returns:
//   - A string representing the [ReasonCategory] instance.
func (r ReasonCategory) String() string {
	return string(r)
}

// IsValid checks whether the [ReasonCategory] instance represents a valid, non-empty category.
//
// Returns:
//   - A boolean value indicating whether the [ReasonCategory] instance is valid:
//   - `true` if the category is non-empty.
//   - `false` if the category is empty.
func (r ReasonCategory) IsValid() bool {
	return strutil.IsNotEmpty(string(r))
}

// Equals checks whether the [ReasonCategory] instance matches any of the provided categories.
//
// Parameters:
//   - other: A variadic list of [ReasonCategory] instances to compare against.
//
// Returns:
//   - A boolean value indicating whether the [ReasonCategory] instance matches any of the provided categories:
//   - `true` if a match is found.
//   - `false` if no match is found or if no categories are provided.
func (r ReasonCategory) Equals(other ...ReasonCategory) bool {
	if len(other) == 0 {
		return false
	}
	return slices.Contains(other, r)
}

// CastReasonCode creates a new [ReasonCode] instance from the provided string.
//
// Parameters:
//   - code: A string representing the reason code.
//
// Returns:
//   - A [ReasonCode] instance initialized with the provided code.
func CastReasonCode(code string) ReasonCode {
	return ReasonCode(code)
}

// CastReasonCategory creates a new [ReasonCategory] instance from the provided string.
//
// Parameters:
//   - category: A string representing the reason category.
//
// Returns:
//   - A [ReasonCategory] instance initialized with the provided category.
func CastReasonCategory(category string) ReasonCategory {
	return ReasonCategory(category)
}

// CastReasonWithCategory creates a new [reason] instance with the specified category and an empty code.
//
// Parameters:
//   - category: A string representing the reason category.
//
// Returns:
//   - A pointer to a newly created [reason] instance initialized with the provided category and an empty code.
func CastReasonWithCategory(category string) *reason {
	return &reason{
		category: CastReasonCategory(category),
	}
}

// CastReasonWithCode creates a new [reason] instance with the specified code and an empty category.
//
// Parameters:
//   - code: A string representing the reason code.
//
// Returns:
//   - A pointer to a newly created [reason] instance initialized with the provided code and an empty category.
func CastReasonWithCode(code string) *reason {
	return &reason{
		code: CastReasonCode(code),
	}
}

// NewReason creates a new [reason] instance with empty category and code.
//
// Returns:
//   - A pointer to a newly created [reason] instance with empty category and code.
func NewReason() *reason {
	return &reason{}
}

// Available checks whether the [reason] instance is non-nil.
//
// This function ensures that the [reason] object exists and is not nil.
// It serves as a safety check to avoid null pointer dereferences when accessing the instance's fields or methods.
//
// Returns:
//   - A boolean value indicating whether the [reason] instance is non-nil:
//   - `true` if the [reason] instance is non-nil.
//   - `false` if the [reason] instance is nil.
func (r *reason) Available() bool {
	return r != nil
}

// Code retrieves the reason code associated with the [reason] instance.
//
// Returns:
//   - A [ReasonCode] representing the reason code. Returns an empty string if the [reason] instance is nil.
func (r *reason) Code() ReasonCode {
	if !r.Available() {
		return ""
	}
	return r.code
}

// Category retrieves the reason category associated with the [reason] instance.
//
// Returns:
//   - A [ReasonCategory] representing the reason category. Returns an empty string if the [reason] instance is nil.
func (r *reason) Category() ReasonCategory {
	if !r.Available() {
		return ""
	}
	return r.category
}

// WithCode sets the reason code for the [reason] instance.
// This method also updates the reason category based on the provided reason code, if the category can be determined.
//
// Parameters:
//   - code: A [ReasonCode] representing the reason code to set.
//
// Returns:
//   - A pointer to the updated [reason] instance.
func (r *reason) WithCode(code ReasonCode) *reason {
	c, ok := ReasonCategoryOf(code) // Retrieve the category associated with the given reason code.
	if ok {
		r.category = c
	}
	r.code = code
	return r
}

// WithCategory sets the reason category for the [reason] instance.
//
// Parameters:
//   - category: A [ReasonCategory] representing the reason category to set.
//
// Returns:
//   - A pointer to the updated [reason] instance.
func (r *reason) WithCategory(category ReasonCategory) *reason {
	r.category = category
	return r
}

// Respond generates a map representation of the [reason] instance.
//
// Returns:
//   - A map with keys "code" and "category" representing the reason's code and category, respectively.
func (r *reason) Respond() map[string]any {
	m := make(map[string]any)
	m["code"] = r.code.String()
	m["category"] = r.category.String()
	return m
}

// JSON generates a JSON string representation of the [reason] instance.
//
// Returns:
//   - A JSON string representing the reason's code and category.
func (r *reason) JSON() string {
	return jsonpass(r.Respond())
}

// JSONPretty generates a pretty-printed JSON string representation of the [reason] instance.
//
// Returns:
//   - A pretty-printed JSON string representing the reason's code and category.
func (r *reason) JSONPretty() string {
	return jsonpretty(r.Respond())
}

// Equal compares the [reason] instance with another [reason] instance for equality.
//
// Parameters:
//   - other: A pointer to the [reason] instance to compare with.
//
// Returns:
//   - A boolean indicating whether the two [reason] instances are equal.
func (r *reason) Equal(other *reason) bool {
	if r == nil && other == nil {
		return true
	}
	if r == nil || other == nil {
		return false
	}
	if r.Code().IsValid() && other.Code().IsValid() {
		if !r.code.Equals(other.code) {
			return false
		}
	}
	if r.Category().IsValid() && other.Category().IsValid() {
		if !r.category.Equals(other.category) {
			return false
		}
	}
	return true
}

// String generates a string representation of the [reason] instance.
//
// Returns:
//   - A string representing the reason's code and category.
func (r *reason) String() string {
	if r == nil {
		return ""
	}
	sw := strchain.New()
	sw.AppendF("code=%s", r.code.String())
	sw.Space()
	sw.AppendF("category=%s", r.category.String())
	return sw.String()
}

// Logging logs the [reason] instance using the provided logger or the default logger.
//
// Parameters:
//   - logger: An optional pointer to a [slogger.Logger] instance to use for logging.
//
// Returns:
//   - A pointer to the [reason] instance.
func (r *reason) Logging(logger ...*slogger.Logger) *reason {
	if r == nil {
		return r
	}
	l := slogger.S()
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0]
	}

	msg := "replify::reason::logging"

	child := l.With()
	child.WithCaller(true).WithCallerSkip(3)

	logAtLevel(child, slogger.InfoLevel, msg, slogger.JSON("REASON", r.Respond()))
	return r
}

// Slogging logs the [reason] instance using the provided logger or the default logger in a simplified manner.
//
// Parameters:
//   - logger: An optional pointer to a [slogger.Logger] instance to use for logging.
//
// Returns:
//   - A pointer to the [reason] instance.
func (r *reason) Slogging(logger ...*slogger.Logger) *reason {
	if r == nil {
		return r
	}
	l := slogger.S()
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0]
	}

	child := l.With()
	child.WithCaller(true).WithCallerSkip(3)

	slogAtLevel(child, slogger.InfoLevel, r.String())
	return r
}

// Reply creates a [B] instance encapsulating the current [reason] instance.
//
// Returns:
//   - A [B] instance containing the [reason] instance.
func (r *reason) Reply() B {
	return B{
		reason: r,
	}
}

// ReplyPtr creates a pointer to a [B] instance encapsulating the current [reason] instance.
//
// Returns:
//   - A pointer to a [B] instance containing the [reason] instance.
func (r *reason) ReplyPtr() *B {
	return &B{
		reason: r,
	}
}

// Clone creates a deep copy of the current [reason] instance.
//
// Returns:
//   - A pointer to a new [reason] instance that is a copy of the current instance.
func (r *reason) Clone() *reason {
	if r == nil {
		return nil
	}
	clone := &reason{
		code:     r.code,
		category: r.category,
	}
	return clone
}
