package strchain

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/polarixa/replify/pkg/strutil"
)

// Chain is a lightweight fluent string wrapper that enables method chaining
// directly from a string value without needing to create a StringWeaver constructor.
//
// Usage:
//
//	result := strchain.Chain("Hello").Space().Append("World").Upper().String()
//	// Output: "HELLO WORLD"
//
//	result := strchain.Chain("  trimmed  ").Trim().Quote().String()
//	// Output: "\"trimmed\""
type Chain string

// C is a shorthand alias for Chain.
//
// Example:
//
//	result := strchain.C("Hello").Space().Append("World").String()
func C(s ...string) Chain {
	return Chain(strings.Join(s, ""))
}

// Empty creates an empty Chain for building strings from scratch.
//
// Example:
//
//	result := strchain.Empty().Append("Built").Space().Append("from scratch").String()
func Empty() Chain {
	return Chain("")
}

// String returns the underlying string value.
func (c Chain) String() string {
	return string(c)
}

// Len returns the length of the string.
func (c Chain) Len() int {
	return strutil.Len(string(c))
}

// IsEmpty returns true if the string is empty.
func (c Chain) IsEmpty() bool {
	return strutil.IsEmpty(string(c))
}

// IsNotEmpty returns true if the string is not empty.
func (c Chain) IsNotEmpty() bool {
	return strutil.Len(string(c)) > 0
}

// ──────────────────────────────────────────────────────────────────────────────
// Append Operations
// ──────────────────────────────────────────────────────────────────────────────

// Append adds a string to the end.
func (c Chain) Append(s string) Chain {
	return Chain(string(c) + s)
}

// AppendF adds a formatted string (printf-style).
func (c Chain) AppendF(format string, args ...any) Chain {
	return Chain(string(c) + fmt.Sprintf(format, args...))
}

// Prepend adds a string to the beginning.
func (c Chain) Prepend(s string) Chain {
	return Chain(s + string(c))
}

// PrependF adds a formatted string to the beginning.
func (c Chain) PrependF(format string, args ...any) Chain {
	return Chain(fmt.Sprintf(format, args...) + string(c))
}

// AppendInt adds an integer.
func (c Chain) AppendInt(i int) Chain {
	return Chain(string(c) + strconv.Itoa(i))
}

// AppendInt64 adds an int64.
func (c Chain) AppendInt64(i int64) Chain {
	return Chain(string(c) + strconv.FormatInt(i, 10))
}

// AppendFloat64 adds a float64.
func (c Chain) AppendFloat64(f float64) Chain {
	return Chain(string(c) + strconv.FormatFloat(f, 'f', -1, 64))
}

// AppendBool adds a boolean.
func (c Chain) AppendBool(b bool) Chain {
	return Chain(string(c) + strconv.FormatBool(b))
}

// AppendByte adds a single byte.
func (c Chain) AppendByte(b byte) Chain {
	return Chain(string(c) + string(b))
}

// AppendRune adds a single rune.
func (c Chain) AppendRune(r rune) Chain {
	return Chain(string(c) + string(r))
}

// AppendBytes adds a byte slice.
func (c Chain) AppendBytes(b []byte) Chain {
	return Chain(string(c) + string(b))
}

// ──────────────────────────────────────────────────────────────────────────────
// Whitespace Operations
// ──────────────────────────────────────────────────────────────────────────────

// Space adds a single space.
func (c Chain) Space() Chain {
	return Chain(string(c) + " ")
}

// Spaces adds n spaces.
func (c Chain) Spaces(n int) Chain {
	if n <= 0 {
		return c
	}
	return Chain(string(c) + strings.Repeat(" ", n))
}

// Tab adds a single tab.
func (c Chain) Tab() Chain {
	return Chain(string(c) + "\t")
}

// Tabs adds n tabs.
func (c Chain) Tabs(n int) Chain {
	if n <= 0 {
		return c
	}
	return Chain(string(c) + strings.Repeat("\t", n))
}

// NewLine adds a newline character.
func (c Chain) NewLine() Chain {
	return Chain(string(c) + "\n")
}

// NewLines adds n newline characters.
func (c Chain) NewLines(n int) Chain {
	if n <= 0 {
		return c
	}
	return Chain(string(c) + strings.Repeat("\n", n))
}

// Line adds a string followed by a newline.
func (c Chain) Line(s string) Chain {
	return Chain(string(c) + s + "\n")
}

// LineF adds a formatted string followed by a newline.
func (c Chain) LineF(format string, args ...any) Chain {
	return Chain(string(c) + fmt.Sprintf(format, args...) + "\n")
}

// ──────────────────────────────────────────────────────────────────────────────
// Punctuation Operations
// ──────────────────────────────────────────────────────────────────────────────

// Comma adds a comma.
func (c Chain) Comma() Chain {
	return Chain(string(c) + ",")
}

// Dot adds a period.
func (c Chain) Dot() Chain {
	return Chain(string(c) + ".")
}

// Colon adds a colon.
func (c Chain) Colon() Chain {
	return Chain(string(c) + ":")
}

// Semicolon adds a semicolon.
func (c Chain) Semicolon() Chain {
	return Chain(string(c) + ";")
}

// Equals adds an equals sign.
func (c Chain) Equals() Chain {
	return Chain(string(c) + "=")
}

// Arrow adds an arrow (->).
func (c Chain) Arrow() Chain {
	return Chain(string(c) + "->")
}

// FatArrow adds a fat arrow (=>).
func (c Chain) FatArrow() Chain {
	return Chain(string(c) + "=>")
}

// ──────────────────────────────────────────────────────────────────────────────
// Wrapping Operations
// ──────────────────────────────────────────────────────────────────────────────

// Wrap wraps the current string with prefix and suffix.
func (c Chain) Wrap(prefix, suffix string) Chain {
	return Chain(prefix + string(c) + suffix)
}

// Quote wraps the string in double quotes.
func (c Chain) Quote() Chain {
	return Chain("\"" + string(c) + "\"")
}

// SingleQuote wraps the string in single quotes.
func (c Chain) SingleQuote() Chain {
	return Chain("'" + string(c) + "'")
}

// Parenthesize wraps the string in parentheses.
func (c Chain) Parenthesize() Chain {
	return Chain("(" + string(c) + ")")
}

// Bracket wraps the string in square brackets.
func (c Chain) Bracket() Chain {
	return Chain("[" + string(c) + "]")
}

// Brace wraps the string in curly braces.
func (c Chain) Brace() Chain {
	return Chain("{" + string(c) + "}")
}

// ──────────────────────────────────────────────────────────────────────────────
// Transformation Operations
// ──────────────────────────────────────────────────────────────────────────────

// Upper converts the string to uppercase.
func (c Chain) Upper() Chain {
	return Chain(strings.ToUpper(string(c)))
}

// Lower converts the string to lowercase.
func (c Chain) Lower() Chain {
	return Chain(strings.ToLower(string(c)))
}

// Title converts the string to title case.
func (c Chain) Title() Chain {
	return Chain(strutil.Title(string(c)))
}

// Trim removes leading and trailing whitespace.
func (c Chain) Trim() Chain {
	return Chain(strings.TrimSpace(string(c)))
}

// TrimPrefix removes the specified prefix.
func (c Chain) TrimPrefix(prefix string) Chain {
	return Chain(strings.TrimPrefix(string(c), prefix))
}

// TrimSuffix removes the specified suffix.
func (c Chain) TrimSuffix(suffix string) Chain {
	return Chain(strings.TrimSuffix(string(c), suffix))
}

// TrimChars removes leading and trailing characters in cutset.
func (c Chain) TrimChars(cutset string) Chain {
	return Chain(strings.Trim(string(c), cutset))
}

// TrimLeft removes leading characters in cutset.
func (c Chain) TrimLeft(cutset string) Chain {
	return Chain(strings.TrimLeft(string(c), cutset))
}

// TrimRight removes trailing characters in cutset.
func (c Chain) TrimRight(cutset string) Chain {
	return Chain(strings.TrimRight(string(c), cutset))
}

// Replace replaces all occurrences of old with new.
func (c Chain) Replace(old, new string) Chain {
	return Chain(strings.ReplaceAll(string(c), old, new))
}

// ReplaceN replaces the first n occurrences of old with new.
func (c Chain) ReplaceN(old, new string, n int) Chain {
	return Chain(strings.Replace(string(c), old, new, n))
}

// Repeat repeats the string n times.
func (c Chain) Repeat(n int) Chain {
	if n <= 0 {
		return Chain("")
	}
	return Chain(strings.Repeat(string(c), n))
}

// Reverse reverses the string (rune-aware).
func (c Chain) Reverse() Chain {
	runes := []rune(string(c))
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return Chain(string(runes))
}

// ──────────────────────────────────────────────────────────────────────────────
// Conditional Operations
// ──────────────────────────────────────────────────────────────────────────────

// AppendIf appends a string if the condition is true.
func (c Chain) AppendIf(condition bool, s string) Chain {
	if condition {
		return Chain(string(c) + s)
	}
	return c
}

// AppendIfF appends a formatted string if the condition is true.
func (c Chain) AppendIfF(condition bool, format string, args ...any) Chain {
	if condition {
		return Chain(string(c) + fmt.Sprintf(format, args...))
	}
	return c
}

// PrependIf prepends a string if the condition is true.
func (c Chain) PrependIf(condition bool, s string) Chain {
	if condition {
		return Chain(s + string(c))
	}
	return c
}

// When applies a transformation function if the condition is true.
func (c Chain) When(condition bool, fn func(Chain) Chain) Chain {
	if condition {
		return fn(c)
	}
	return c
}

// Unless applies a transformation function if the condition is false.
func (c Chain) Unless(condition bool, fn func(Chain) Chain) Chain {
	if !condition {
		return fn(c)
	}
	return c
}

// IfEmpty returns the fallback if the string is empty.
func (c Chain) IfEmpty(fallback string) Chain {
	if len(c) == 0 {
		return Chain(fallback)
	}
	return c
}

// IfNotEmpty applies a transformation if the string is not empty.
func (c Chain) IfNotEmpty(fn func(Chain) Chain) Chain {
	if len(c) > 0 {
		return fn(c)
	}
	return c
}

// ──────────────────────────────────────────────────────────────────────────────
// Join & Split Operations
// ──────────────────────────────────────────────────────────────────────────────

// Join joins multiple strings using the current string as separator.
func (c Chain) Join(elements ...string) Chain {
	return Chain(strings.Join(elements, string(c)))
}

// JoinSlice joins a slice of strings using the current string as separator.
func (c Chain) JoinSlice(elements []string) Chain {
	return Chain(strings.Join(elements, string(c)))
}

// Split splits the string by separator and returns the parts.
func (c Chain) Split(sep string) []string {
	return strings.Split(string(c), sep)
}

// SplitN splits the string by separator into at most n parts.
func (c Chain) SplitN(sep string, n int) []string {
	return strings.SplitN(string(c), sep, n)
}

// ──────────────────────────────────────────────────────────────────────────────
// Query Operations
// ────────────────────────────────────────────────────────────────────────────

// Contains returns true if the string contains substr.
func (c Chain) Contains(substr string) bool {
	return strings.Contains(string(c), substr)
}

// HasPrefix returns true if the string starts with prefix.
func (c Chain) HasPrefix(prefix string) bool {
	return strings.HasPrefix(string(c), prefix)
}

// HasSuffix returns true if the string ends with suffix.
func (c Chain) HasSuffix(suffix string) bool {
	return strings.HasSuffix(string(c), suffix)
}

// Index returns the index of the first occurrence of substr, or -1.
func (c Chain) Index(substr string) int {
	return strings.Index(string(c), substr)
}

// Count returns the number of non-overlapping occurrences of substr.
func (c Chain) Count(substr string) int {
	return strings.Count(string(c), substr)
}

// ──────────────────────────────────────────────────────────────────────────────
// Indentation Operations
// ──────────────────────────────────────────────────────────────────────────────

// Indent adds indentation (2 spaces per level) before the current text.
func (c Chain) Indent(level int) Chain {
	if level <= 0 {
		return c
	}
	return Chain(strings.Repeat("  ", level) + string(c))
}

// IndentLine adds indentation before the text and appends a newline.
func (c Chain) IndentLine(level int) Chain {
	if level <= 0 {
		return Chain(string(c) + "\n")
	}
	return Chain(strings.Repeat("  ", level) + string(c) + "\n")
}

// ──────────────────────────────────────────────────────────────────────────────
// Conversion Operations
// ──────────────────────────────────────────────────────────────────────────────

// Bytes returns the string as a byte slice.
func (c Chain) Bytes() []byte {
	return []byte(c)
}

// Runes returns the string as a rune slice.
func (c Chain) Runes() []rune {
	return []rune(c)
}

// ToWeaver converts the Chain to a StringWeaver for more complex operations.
func (c Chain) ToWeaver() *StringWeaver {
	return From(string(c))
}

// ToSafeWeaver converts the Chain to a SafeStringWeaver for thread-safe operations.
func (c Chain) ToSafeWeaver() *SafeStringWeaver {
	return SafeFrom(string(c))
}

// ──────────────────────────────────────────────────────────────────────────────
// Map/Transform Operations
// ──────────────────────────────────────────────────────────────────────────────

// Map applies a transformation function to the string.
func (c Chain) Map(fn func(string) string) Chain {
	return Chain(fn(string(c)))
}

// MapRunes applies a function to each rune.
func (c Chain) MapRunes(fn func(rune) rune) Chain {
	return Chain(strings.Map(fn, string(c)))
}

// Filter keeps only runes that satisfy the predicate.
func (c Chain) Filter(fn func(rune) bool) Chain {
	var result strings.Builder
	for _, r := range string(c) {
		if fn(r) {
			result.WriteRune(r)
		}
	}
	return Chain(result.String())
}

// ──────────────────────────────────────────────────────────────────────────────
// Padding Operations
// ──────────────────────────────────────────────────────────────────────────────

// PadLeft pads the string on the left to reach the specified width.
func (c Chain) PadLeft(width int, pad string) Chain {
	s := string(c)
	if len(s) >= width || pad == "" {
		return c
	}
	padLen := width - len(s)
	padding := strings.Repeat(pad, (padLen/len(pad))+1)[:padLen]
	return Chain(padding + s)
}

// PadRight pads the string on the right to reach the specified width.
func (c Chain) PadRight(width int, pad string) Chain {
	s := string(c)
	if len(s) >= width || pad == "" {
		return c
	}
	padLen := width - len(s)
	padding := strings.Repeat(pad, (padLen/len(pad))+1)[:padLen]
	return Chain(s + padding)
}

// Center centers the string within the specified width.
func (c Chain) Center(width int, pad string) Chain {
	s := string(c)
	if len(s) >= width || pad == "" {
		return c
	}
	total := width - len(s)
	left := total / 2
	right := total - left
	leftPad := strings.Repeat(pad, (left/len(pad))+1)[:left]
	rightPad := strings.Repeat(pad, (right/len(pad))+1)[:right]
	return Chain(leftPad + s + rightPad)
}

// Truncate truncates the string to the specified length.
func (c Chain) Truncate(length int) Chain {
	s := string(c)
	if len(s) <= length {
		return c
	}
	return Chain(s[:length])
}

// TruncateWithSuffix truncates and adds a suffix if truncated.
func (c Chain) TruncateWithSuffix(length int, suffix string) Chain {
	s := string(c)
	if len(s) <= length {
		return c
	}
	if length <= len(suffix) {
		return Chain(suffix[:length])
	}
	return Chain(s[:length-len(suffix)] + suffix)
}
