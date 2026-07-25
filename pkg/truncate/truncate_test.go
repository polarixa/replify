package truncate_test

import (
	"testing"

	"github.com/polarixa/replify/pkg/truncate"
)

// ============================================================================
// CUT STRATEGY TESTS
// ============================================================================

func TestCutStrategy_Truncate(t *testing.T) {
	strategy := truncate.NewCutStrategy()

	tests := []struct {
		name     string
		input    string
		length   int
		expected string
	}{
		{"truncate_short", "Hello, World!", 5, "Hello"},
		{"no_truncation", "Hello", 10, "Hello"},
		{"exact_length", "Hello", 5, "Hello"},
		{"length_one", "Hello", 1, "H"},
		{"length_zero", "Hello", 0, ""},
		{"negative_length", "Hello", -1, ""},
		{"empty_string", "", 5, ""},
		{"unicode", "Héllo, Wörld!", 7, "Héllo, "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strategy.Truncate(tt.input, tt.length)
			if result != tt.expected {
				t.Errorf("CutStrategy.Truncate(%q, %d) = %q, want %q", tt.input, tt.length, result, tt.expected)
			}
		})
	}
}

// ============================================================================
// CUT ELLIPSIS STRATEGY TESTS
// ============================================================================

func TestCutEllipsisStrategy_Truncate(t *testing.T) {
	strategy := truncate.NewCutEllipsisStrategy()

	tests := []struct {
		name     string
		input    string
		length   int
		expected string
	}{
		{"truncate_with_ellipsis", "Hello, World!", 8, "Hello, …"},
		{"no_truncation", "Hello", 10, "Hello"},
		{"exact_length", "Hello", 5, "Hello"},
		{"length_two", "Hello", 2, "H…"},
		{"length_one_fallback", "Hello", 1, "H"},
		{"length_zero", "Hello", 0, ""},
		{"empty_string", "", 5, ""},
		{"unicode", "日本語テスト文字列", 5, "日本語テ…"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strategy.Truncate(tt.input, tt.length)
			if result != tt.expected {
				t.Errorf("CutEllipsisStrategy.Truncate(%q, %d) = %q, want %q", tt.input, tt.length, result, tt.expected)
			}
		})
	}
}

// ============================================================================
// CUT ELLIPSIS LEADING STRATEGY TESTS
// ============================================================================

func TestCutEllipsisLeadingStrategy_Truncate(t *testing.T) {
	strategy := truncate.NewCutEllipsisLeadingStrategy()

	tests := []struct {
		name     string
		input    string
		length   int
		expected string
	}{
		{"truncate_leading", "Hello, World!", 8, "… World!"},
		{"no_truncation", "Hello", 10, "Hello"},
		{"exact_length", "Hello", 5, "Hello"},
		{"length_two", "Hello", 2, "…o"},
		{"length_one_fallback", "Hello", 1, "H"},
		{"length_zero", "Hello", 0, ""},
		{"empty_string", "", 5, ""},
		{"unicode", "日本語テスト文字列", 5, "…ト文字列"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strategy.Truncate(tt.input, tt.length)
			if result != tt.expected {
				t.Errorf("CutEllipsisLeadingStrategy.Truncate(%q, %d) = %q, want %q", tt.input, tt.length, result, tt.expected)
			}
		})
	}
}

// ============================================================================
// ELLIPSIS MIDDLE STRATEGY TESTS
// ============================================================================

func TestEllipsisMiddleStrategy_Truncate(t *testing.T) {
	strategy := truncate.NewEllipsisMiddleStrategy()

	tests := []struct {
		name     string
		input    string
		length   int
		expected string
	}{
		{"truncate_middle_even", "Hello, World!", 8, "Hel…rld!"},
		{"no_truncation", "Hello", 10, "Hello"},
		{"exact_length", "Hello", 5, "Hello"},
		{"length_zero", "Hello", 0, ""},
		{"empty_string", "", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strategy.Truncate(tt.input, tt.length)
			if result != tt.expected {
				t.Errorf("EllipsisMiddleStrategy.Truncate(%q, %d) = %q, want %q", tt.input, tt.length, result, tt.expected)
			}
		})
	}
}

// ============================================================================
// TRUNCATOR BUILDER TESTS
// ============================================================================

func TestTruncatorBuilder_DefaultValues(t *testing.T) {
	tr := truncate.NewTruncator().
		WithMaxLength(8).
		Build()

	// Default should be PositionEnd with DefaultOmission
	result := tr.Truncate("Hello, World!")
	expected := "Hello, …"
	if result != expected {
		t.Errorf("Truncator default = %q, want %q", result, expected)
	}
}

func TestTruncatorBuilder_CustomOmission(t *testing.T) {
	tr := truncate.NewTruncator().
		WithOmission("...").
		WithMaxLength(8).
		Build()

	result := tr.Truncate("Hello, World!")
	expected := "Hello..."
	if result != expected {
		t.Errorf("Truncator custom omission = %q, want %q", result, expected)
	}
}

func TestTruncatorBuilder_PositionStart(t *testing.T) {
	tr := truncate.NewTruncator().
		WithPosition(truncate.PositionStart).
		WithMaxLength(8).
		Build()

	result := tr.Truncate("Hello, World!")
	expected := "… World!"
	if result != expected {
		t.Errorf("Truncator PositionStart = %q, want %q", result, expected)
	}
}

func TestTruncatorBuilder_PositionMiddle(t *testing.T) {
	tr := truncate.NewTruncator().
		WithPosition(truncate.PositionMiddle).
		WithMaxLength(8).
		Build()

	result := tr.Truncate("Hello, World!")
	expected := "Hel…rld!"
	if result != expected {
		t.Errorf("Truncator PositionMiddle = %q, want %q", result, expected)
	}
}

func TestTruncatorBuilder_NoOmission(t *testing.T) {
	tr := truncate.NewTruncator().
		WithOmission("").
		WithMaxLength(5).
		Build()

	result := tr.Truncate("Hello, World!")
	expected := "Hello"
	if result != expected {
		t.Errorf("Truncator no omission = %q, want %q", result, expected)
	}
}

func TestTruncatorBuilder_TruncateWithLength(t *testing.T) {
	tr := truncate.NewTruncator().
		WithOmission("...").
		WithPosition(truncate.PositionEnd).
		Build()

	result := tr.TruncateWithLength("Hello, World!", 8)
	expected := "Hello..."
	if result != expected {
		t.Errorf("Truncator.TruncateWithLength = %q, want %q", result, expected)
	}
}

// ============================================================================
// APPLY CONVENIENCE FUNCTION TESTS
// ============================================================================

func TestApply(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		length   int
		strategy truncate.Strategy
		expected string
	}{
		{"apply_cut", "Hello, World!", 5, truncate.NewCutStrategy(), "Hello"},
		{"apply_ellipsis", "Hello, World!", 8, truncate.NewCutEllipsisStrategy(), "Hello, …"},
		{"apply_leading", "Hello, World!", 8, truncate.NewCutEllipsisLeadingStrategy(), "… World!"},
		{"apply_middle", "Hello, World!", 8, truncate.NewEllipsisMiddleStrategy(), "Hel…rld!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate.Apply(tt.input, tt.length, tt.strategy)
			if result != tt.expected {
				t.Errorf("Apply(%q, %d) = %q, want %q", tt.input, tt.length, result, tt.expected)
			}
		})
	}
}

// ============================================================================
// EDGE CASE TESTS
// ============================================================================

func TestTruncate_EdgeCases(t *testing.T) {
	strategy := truncate.NewCutEllipsisStrategy()

	tests := []struct {
		name     string
		input    string
		length   int
		expected string
	}{
		{"single_char_string", "A", 1, "A"},
		{"single_char_string_long_limit", "A", 100, "A"},
		{"two_char_truncated", "AB", 1, "A"},
		{"emoji", "Hello 👋 World!", 8, "Hello 👋…"},
		{"multi_codepoint", "café", 3, "ca…"},
		{"whitespace_only", "      ", 3, "  …"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strategy.Truncate(tt.input, tt.length)
			if result != tt.expected {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tt.input, tt.length, result, tt.expected)
			}
		})
	}
}

// ============================================================================
// BENCHMARK TESTS
// ============================================================================

func BenchmarkCutStrategy(b *testing.B) {
	strategy := truncate.NewCutStrategy()
	input := "The quick brown fox jumps over the lazy dog"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		strategy.Truncate(input, 20)
	}
}

func BenchmarkCutEllipsisStrategy(b *testing.B) {
	strategy := truncate.NewCutEllipsisStrategy()
	input := "The quick brown fox jumps over the lazy dog"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		strategy.Truncate(input, 20)
	}
}

func BenchmarkEllipsisMiddleStrategy(b *testing.B) {
	strategy := truncate.NewEllipsisMiddleStrategy()
	input := "The quick brown fox jumps over the lazy dog"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		strategy.Truncate(input, 20)
	}
}

func BenchmarkTruncatorBuilder(b *testing.B) {
	tr := truncate.NewTruncator().
		WithOmission("...").
		WithPosition(truncate.PositionMiddle).
		WithMaxLength(20).
		Build()
	input := "The quick brown fox jumps over the lazy dog"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Truncate(input)
	}
}
