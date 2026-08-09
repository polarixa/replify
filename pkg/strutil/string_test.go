package strutil

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestOneline(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		// empty / no-op
		{"empty string", "", ""},
		{"single space", " ", " "},
		{"spaces only", "   ", "   "},
		{"no line breaks", "no breaks here", "no breaks here"},

		// LF (\n)
		{"LF only", "\n", " "},
		{"LF at start", "\nhello", " hello"},
		{"LF at end", "hello\n", "hello "},
		{"LF in middle", "hello\nworld", "hello world"},
		{"two consecutive LFs", "a\n\nb", "a  b"},
		{"three consecutive LFs", "a\n\n\nb", "a   b"},

		// CR (\r)
		{"CR only", "\r", " "},
		{"CR at start", "\rhello", " hello"},
		{"CR at end", "hello\r", "hello "},
		{"CR in middle", "foo\rbar", "foo bar"},
		{"two consecutive CRs", "a\r\rb", "a  b"},

		// CRLF (\r\n) — must count as exactly one space
		{"CRLF only", "\r\n", " "},
		{"CRLF at start", "\r\nhello", " hello"},
		{"CRLF at end", "hello\r\n", "hello "},
		{"CRLF in middle", "foo\r\nbar", "foo bar"},
		{"two consecutive CRLFs", "a\r\n\r\nb", "a  b"},
		{"CRLF then LF", "x\r\n\ny", "x  y"},
		{"LF then CRLF", "x\n\r\ny", "x  y"},

		// tab (\t)
		{"tab only", "\t", " "},
		{"tab at start", "\thello", " hello"},
		{"tab at end", "hello\t", "hello "},
		{"tab in middle", "a\tb", "a b"},
		{"multiple tabs", "a\t\tb", "a  b"},

		// vertical tab (\v) and form feed (\f)
		{"vertical tab", "a\vb", "a b"},
		{"form feed", "a\fb", "a b"},
		{"VT and FF together", "a\v\fb", "a  b"},

		// Unicode line breakers
		{"NEL U+0085", "a\u0085b", "a b"},
		{"NEL at start", "\u0085hello", " hello"},
		{"NEL at end", "hello\u0085", "hello "},
		{"Line Separator U+2028", "a\u2028b", "a b"},
		{"LS at start", "\u2028hello", " hello"},
		{"LS at end", "hello\u2028", "hello "},
		{"Paragraph Separator U+2029", "a\u2029b", "a b"},
		{"PS only", "\u2029", " "},
		{"PS at start", "\u2029hello", " hello"},
		{"NEL+LS+PS sequence", "\u0085\u2028\u2029", "   "},

		// all nine break sequences together in one string
		{"all break types",
			"a\nb\rc\r\nd\te\vf\fg\u0085h\u2028i\u2029j",
			"a b c d e f g h i j"},

		// mixed positions
		{"mixed multi-line", "line1\nline2\r\nline3\rline4", "line1 line2 line3 line4"},
		{"leading and trailing LF", "\nhello\n", " hello "},
		{"multiple leading breaks", "\n\nhello", "  hello"},
		{"multiple trailing breaks", "hello\n\n", "hello  "},
		{"break-only sequence", "\n\r\n\r", "   "},
		{"only tabs", "\t\t\t", "   "},

		// Unicode content preserved
		{"latin accents", "héllo\nwörld", "héllo wörld"},
		{"CJK", "你好\n世界", "你好 世界"},
		{"Arabic RTL", "مرحبا\nبالعالم", "مرحبا بالعالم"},
		{"emoji 2-codepoint", "hello 😀\nworld 🌍", "hello 😀 world 🌍"},
		{"emoji only", "🎉\n🎊", "🎉 🎊"},
		// e + combining acute (U+0301), newline, f + combining circumflex (U+0302)
		{"combining diacritics", "e\u0301\nf\u0302", "e\u0301 f\u0302"},
		{"Devanagari", "नमस्ते\nदुनिया", "नमस्ते दुनिया"},
		{"mixed scripts", "Latin\nCyrillicКирилл\nGreekΕλλάδα", "Latin CyrillicКирилл GreekΕλλάδα"},

		// characters that must NOT be altered
		{"null byte preserved", "a\x00b", "a\x00b"},
		{"non-breaking space U+00A0", "a\u00a0b", "a\u00a0b"},
		{"zero-width space U+200B", "a\u200bb", "a\u200bb"},
		{"soft hyphen U+00AD", "a\u00adb", "a\u00adb"},
		{"bell preserved", "a\x07b", "a\x07b"},
		{"backspace preserved", "a\x08b", "a\x08b"},
		{"escape preserved", "a\x1bb", "a\x1bb"},
		{"delete preserved", "a\x7fb", "a\x7fb"},
		{"en-space U+2002", "a\u2002b", "a\u2002b"},
		{"em-space U+2003", "a\u2003b", "a\u2003b"},
		{"BOM U+FEFF", "a\ufeffb", "a\ufeffb"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Oneline(tt.input)
			if got != tt.want {
				t.Errorf("Oneline(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestOnelineNoLineBreaksRemain asserts the output never contains any line-breaking rune.
func TestOnelineNoLineBreaksRemain(t *testing.T) {
	breakers := []rune{'\n', '\r', '\t', '\v', '\f', '\u0085', '\u2028', '\u2029'}
	inputs := []string{
		"multi\nline\r\nstring",
		"\t\v\f\r\n",
		"a\u0085b\u2028c\u2029d",
		"",
		"no breaks",
		"hello\n\n\nworld",
		"a\nb\rc\r\nd\te\vf\fg\u0085h\u2028i\u2029j",
	}
	for _, input := range inputs {
		got := Oneline(input)
		for _, r := range breakers {
			if strings.ContainsRune(got, r) {
				t.Errorf("Oneline(%q) = %q still contains rune %U", input, got, r)
			}
		}
	}
}

// TestOnelineValidUTF8 asserts the output is always valid UTF-8.
func TestOnelineValidUTF8(t *testing.T) {
	inputs := []string{
		"héllo\nwörld",
		"你好\n世界",
		"مرحبا\nبالعالم",
		"hello 😀\nworld",
		"e\u0301\nf\u0302",
		"",
		"\n",
		"plain ascii",
	}
	for _, input := range inputs {
		got := Oneline(input)
		if !utf8.ValidString(got) {
			t.Errorf("Oneline(%q) produced invalid UTF-8: %q", input, got)
		}
	}
}

// TestOnelineIdempotent asserts that calling Oneline twice gives the same result as once.
func TestOnelineIdempotent(t *testing.T) {
	inputs := []string{
		"hello\nworld",
		"foo\r\nbar",
		"\t\v\f",
		"a\u0085b\u2028c",
		"",
		"no breaks",
		"a\nb\rc\r\nd\te\vf\fg\u0085h\u2028i\u2029j",
	}
	for _, input := range inputs {
		once := Oneline(input)
		twice := Oneline(once)
		if once != twice {
			t.Errorf("not idempotent for %q: first=%q second=%q", input, once, twice)
		}
	}
}
