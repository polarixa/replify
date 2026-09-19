package b64

import (
	"encoding/base64"
	"testing"
)

// TestIsBase64 exercises the package-level convenience function, which
// mirrors the exact behavior of the original isBase64 function
// (standard encoding only).
func TestIsBase64(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"empty string", "", false},
		{"valid simple text", "SGVsbG8sIFdvcmxkIQ==", true}, // "Hello, World!"
		{"valid no padding needed", "SGVsbG8h", true},       // "Hello!"
		{"valid single padding char", "TWFu", true},         // "Man"
		{"valid all-padding edge", "YQ==", true},            // "a"
		{"invalid characters", "not_base64!", false},
		{"invalid length (not multiple of 4)", "QQ", false},
		{"whitespace inside payload", "SGVs bG8h", false},
		{"trailing newline", "SGVsbG8h\n", false},
		{"valid single trailing padding", "SGVsbG8=", true},
		{"non-canonical trailing bits", "AB==", false},
		{"URL-safe characters rejected by std encoding", "PDw_Pz8-Pg==", false},
		{"decodes but round-trip mismatch (extra padding)", "YQ===", false},
		{"only padding characters", "====", false},
		{"single character", "A", false},
		{"random garbage", "@@@@", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := IsBase64(tc.in)
			if got != tc.want {
				t.Errorf("IsBase64(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

// TestBase64Validator_StdEncoding re-validates the same cases directly
// against the OOP type to make sure the convenience function and the
// underlying type agree.
func TestBase64Validator_StdEncoding(t *testing.T) {
	v := NewBase64Validator(base64.StdEncoding)

	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"empty string", "", false},
		{"valid padded", "SGVsbG8sIFdvcmxkIQ==", true},
		{"invalid alphabet", "not_base64!", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := v.IsValid(tc.in); got != tc.want {
				t.Errorf("IsValid(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

// TestBase64Validator_NilEncodingDefaultsToStd ensures passing nil falls
// back to StdEncoding, same as the zero-value behavior implied by the
// original function.
func TestBase64Validator_NilEncodingDefaultsToStd(t *testing.T) {
	v := NewBase64Validator(nil)
	if !v.IsValid("SGVsbG8h") {
		t.Errorf("expected nil encoding to default to StdEncoding and accept valid std Base64")
	}
}

// TestBase64Validator_URLEncoding proves the type generalizes cleanly to
// other encodings, which the original function could not do without
// copy-pasting.
func TestBase64Validator_URLEncoding(t *testing.T) {
	v := NewBase64Validator(base64.URLEncoding)

	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"valid URL-safe with padding", "PDw_Pz8-Pg==", true},
		{"std-safe chars rejected by URL encoding", "PDw/Pz8+Pg==", false},
		{"empty string", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := v.IsValid(tc.in); got != tc.want {
				t.Errorf("IsValid(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

// TestBase64Validator_RawEncoding checks unpadded variants, another
// dialect the original function had no way to express.
func TestBase64Validator_RawEncoding(t *testing.T) {
	v := NewBase64Validator(base64.RawStdEncoding)

	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"valid raw (no padding)", "SGVsbG8h", true},
		{"padded input rejected by raw encoding", "SGVsbG8sIFdvcmxkIQ==", false},
		{"empty string", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := v.IsValid(tc.in); got != tc.want {
				t.Errorf("IsValid(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

// TestCompositeValidator verifies the "accept if any encoding accepts"
// composition logic.
func TestCompositeValidator(t *testing.T) {
	c := NewMultiEncodingValidator(
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	)

	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"empty string", "", false},
		{"valid std padded", "SGVsbG8sIFdvcmxkIQ==", true},
		{"valid raw std (no padding)", "SGVsbG8h", true},
		{"valid URL-safe padded", "PDw_Pz8-Pg==", true},
		{"invalid under every encoding", "not_base64!!", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := c.IsValid(tc.in); got != tc.want {
				t.Errorf("IsValid(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

// TestCompositeValidator_NoValidators ensures the degenerate empty-set
// case is well-defined (always false) rather than panicking.
func TestCompositeValidator_NoValidators(t *testing.T) {
	c := NewCompositeValidator()
	if c.IsValid("SGVsbG8h") {
		t.Errorf("expected composite validator with no members to reject everything")
	}
}

// TestIsBase64Any covers the multi-dialect convenience helper.
func TestIsBase64Any(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"empty string", "", false},
		{"std padded", "SGVsbG8sIFdvcmxkIQ==", true},
		{"raw unpadded", "SGVsbG8h", true},
		{"url-safe padded", "PDw_Pz8-Pg==", true},
		{"garbage", "!!!not-base64!!!", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsBase64Any(tc.in); got != tc.want {
				t.Errorf("IsBase64Any(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

// TestValidatorInterfaceSatisfaction is a compile-time-flavored test that
// exercises both implementations purely through the Validator interface,
// confirming the OOP design is actually polymorphic and not just
// structurally similar.
func TestValidatorInterfaceSatisfaction(t *testing.T) {
	var validators []Validator = []Validator{
		NewBase64Validator(base64.StdEncoding),
		NewCompositeValidator(NewBase64Validator(base64.StdEncoding)),
	}

	for i, v := range validators {
		if !v.IsValid("SGVsbG8h") {
			t.Errorf("validators[%d].IsValid returned false for a valid Base64 string", i)
		}
		if v.IsValid("") {
			t.Errorf("validators[%d].IsValid returned true for an empty string", i)
		}
	}
}
