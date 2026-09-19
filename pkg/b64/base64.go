package b64

import "encoding/base64"

// Validator is the behavior every Base64 validator must expose.
// Depending on the concrete implementation, "valid" can mean "valid for
// one specific encoding" (see Base64Validator) or "valid for at least one
// of several encodings" (see CompositeValidator).
type Validator interface {
	// IsValid reports whether s is a syntactically valid, canonically
	// encoded Base64 string.
	IsValid(s string) bool
}

// Base64Validator validates a string against a single Base64 encoding
// (e.g. base64.StdEncoding, base64.URLEncoding, base64.RawStdEncoding, ...).
//
// It is the direct OOP replacement for the original isBase64 function,
// generalized so the encoding is a field instead of a hard-coded package
// call.
type Base64Validator struct {
	encoding *base64.Encoding
}

// Compile-time check that Base64Validator satisfies Validator.
var _ Validator = (*Base64Validator)(nil)

// NewBase64Validator creates a Validator for the given Base64 encoding.
// If encoding is nil, base64.StdEncoding is used, matching the behavior
// of the original function.
func NewBase64Validator(encoding *base64.Encoding) *Base64Validator {
	if encoding == nil {
		encoding = base64.StdEncoding
	}
	return &Base64Validator{encoding: encoding}
}

// IsValid reports whether s is valid, canonical Base64 for this
// validator's encoding.
//
// The check works in two steps, identical in spirit to the original
// function:
//  1. Reject the empty string outright (an empty string is not
//     considered "a Base64 string" for this validator's purposes).
//  2. Decode s and re-encode the result. If decoding fails, or the
//     round trip doesn't reproduce s exactly, s is rejected. The
//     round-trip comparison is what catches things a bare decode-only
//     check would miss: non-canonical padding, mixed alphabets,
//     embedded whitespace/newlines, etc.
func (v *Base64Validator) IsValid(s string) bool {
	if s == "" {
		return false
	}

	decoded, err := v.encoding.DecodeString(s)
	if err != nil {
		return false
	}

	return v.encoding.EncodeToString(decoded) == s
}

// CompositeValidator validates a string against several encodings and
// reports it valid if ANY of them accept it. This is useful when the
// source of a string is unknown or mixed (e.g. some payloads arrive as
// standard Base64, others as URL-safe Base64).
type CompositeValidator struct {
	validators []Validator
}

// Compile-time check that CompositeValidator satisfies Validator.
var _ Validator = (*CompositeValidator)(nil)

// NewCompositeValidator builds a CompositeValidator out of one or more
// existing Validators.
func NewCompositeValidator(validators ...Validator) *CompositeValidator {
	return &CompositeValidator{validators: validators}
}

// NewMultiEncodingValidator is a convenience constructor that builds a
// CompositeValidator directly from *base64.Encoding values, so callers
// don't have to wrap each one in a Base64Validator by hand.
func NewMultiEncodingValidator(encodings ...*base64.Encoding) *CompositeValidator {
	validators := make([]Validator, 0, len(encodings))
	for _, enc := range encodings {
		validators = append(validators, NewBase64Validator(enc))
	}
	return &CompositeValidator{validators: validators}
}

// IsValid reports whether s is valid for at least one of the underlying
// validators. An empty set of validators always returns false.
func (c *CompositeValidator) IsValid(s string) bool {
	for _, v := range c.validators {
		if v.IsValid(s) {
			return true
		}
	}
	return false
}

// Package-level convenience helpers, kept so existing call sites can
// migrate from the original free function with minimal changes.

// defaultValidator is the package-wide validator for standard Base64,
// matching the original function's behavior exactly.
var defaultValidator = NewBase64Validator(base64.StdEncoding)

// IsBase64 reports whether s is valid, canonical standard Base64
// (RFC 4648 with padding). It is a drop-in replacement for the original
// isBase64 function.
func IsBase64(s string) bool {
	return defaultValidator.IsValid(s)
}

// IsBase64Any reports whether s is valid Base64 under any of the common
// dialects: standard, standard-without-padding, URL-safe, and
// URL-safe-without-padding.
func IsBase64Any(s string) bool {
	return NewMultiEncodingValidator(
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	).IsValid(s)
}
