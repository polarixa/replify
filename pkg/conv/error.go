package conv

import (
	"errors"
	"fmt"

	"github.com/polarixa/replify/pkg/strutil"
)

// ///////////////////////////
// Section:  Conversion errors
// ///////////////////////////

// ConvError represents a type-conversion failure and its diagnostic context.
// It includes the source value, the target type name, an optional custom message,
// and an optional wrapped cause error for error-chain compatibility.
type ConvError struct {
	From    any    // The source value
	To      string // The target type name
	Message string // Additional error message
	Cause   error  // Underlying wrapped error (supports errors.Is/As chaining)
}

// Error implements the error interface for ConvError.
//
// It returns the custom message if provided; otherwise, it constructs a default error message
// indicating the source value and target type.
//
// Returns:
//   - A string representing the error message.
func (e *ConvError) Error() string {
	if strutil.IsNotEmpty(e.Message) {
		return e.Message
	}
	return fmt.Sprintf("cannot convert %#v (type %[1]T) to %v", e.From, e.To)
}

// Unwrap returns the wrapped cause error, enabling errors.Is/errors.As chain traversal.
//
// Returns:
//   - The underlying cause error, or nil if none was set.
func (e *ConvError) Unwrap() error {
	return e.Cause
}

// ///////////////////////////
// Section: Error checking
// ///////////////////////////

// IsConvError reports whether any error in err's chain is a *ConvError.
// It uses errors.As to traverse wrapped error chains.
//
// Parameters:
//   - `err`: The error to be checked.
//
// Returns:
//   - A boolean indicating whether the error chain contains a ConvError.
func IsConvError(err error) bool {
	var ce *ConvError
	return errors.As(err, &ce)
}

// AsConvError attempts to find the first *ConvError in err's chain.
// It uses errors.As to traverse wrapped error chains.
//
// Parameters:
//   - `err`: The error to be cast.
//
// Returns:
//   - A pointer to the ConvError if one is found in the chain.
func AsConvError(err error) (*ConvError, bool) {
	var ce *ConvError
	if errors.As(err, &ce) {
		return ce, true
	}
	return nil, false
}

// ///////////////////////////
// Section: Error constructors
// ///////////////////////////

// newConvError creates a new ConvError instance with the given source value and target type name.
//
// Parameters:
//   - `from`: The source value that failed to convert.
//   - `to`: The target type name to which the conversion was attempted.
//
// Returns:
//   - A pointer to a ConvError instance representing the conversion error.
func newConvError(from any, to string) error {
	return &ConvError{From: from, To: to}
}

// newConvErrorf creates a new ConvError instance with a formatted error message.
//
// Parameters:
//   - `format`: The format string for the error message.
//   - `args`: The arguments to be formatted into the message.
//
// Returns:
//   - A pointer to a ConvError instance representing the conversion error with the formatted message.
func newConvErrorf(format string, args ...any) error {
	return &ConvError{Message: fmt.Sprintf(format, args...)}
}

// newConvErrorMsg creates a new ConvError instance with a custom error message.
//
// Parameters:
//   - `msg`: The custom error message.
//
// Returns:
//   - A pointer to a ConvError instance representing the conversion error with the custom message.
func newConvErrorMsg(msg string) error {
	return &ConvError{Message: msg}
}
