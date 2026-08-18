package strchain

import "strings"

// Weaver defines the common interface for fluent string builder operations.
// Both StringWeaver (non-thread-safe, maximum performance) and SafeStringWeaver
// (thread-safe with mutex synchronization) implement this interface.
//
// Use Weaver when you need polymorphism — for example, a function that accepts
// either a StringWeaver or a SafeStringWeaver as a parameter.
//
// Methods that accept callback functions with concrete receiver types
// (When, Unless, Each) are not included in this interface because their
// function parameter types differ between StringWeaver and SafeStringWeaver.
// These methods are available only on the concrete types.
//
// Example:
//
//	var w strchain.Weaver = strchain.New()
//	result := w.Append("Hello").Space().Append("World").Build()
//
//	var sw strchain.Weaver = strchain.NewSafe()
//	result = sw.Append("Hello").Space().Append("World").Build()
type Weaver interface {

	// Bytes returns the current content as a byte slice.
	Bytes() []byte

	// Append adds a string and returns the builder for chaining.
	Append(s string) Weaver

	// AppendF adds a formatted string (printf-style) and returns the builder.
	AppendF(format string, args ...any) Weaver

	// AppendByte adds a single byte and returns the builder.
	AppendByte(b byte) Weaver

	// AppendRune adds a single rune and returns the builder.
	AppendRune(r rune) Weaver

	// AppendBytes adds a byte slice and returns the builder.
	AppendBytes(b []byte) Weaver

	// AppendInt adds an integer and returns the builder.
	AppendInt(i int) Weaver

	// AppendInt8 adds an int8 and returns the builder.
	AppendInt8(i int8) Weaver

	// AppendInt16 adds an int16 and returns the builder.
	AppendInt16(i int16) Weaver

	// AppendInt32 adds an int32 and returns the builder.
	AppendInt32(i int32) Weaver

	// AppendInt64 adds an int64 and returns the builder.
	AppendInt64(i int64) Weaver

	// AppendUint adds a uint and returns the builder.
	AppendUint(i uint) Weaver

	// AppendUint8 adds a uint8 and returns the builder.
	AppendUint8(i uint8) Weaver

	// AppendUint16 adds a uint16 and returns the builder.
	AppendUint16(i uint16) Weaver

	// AppendUint32 adds a uint32 and returns the builder.
	AppendUint32(i uint32) Weaver

	// AppendUint64 adds a uint64 and returns the builder.
	AppendUint64(i uint64) Weaver

	// AppendUintptr adds a uintptr and returns the builder.
	AppendUintptr(i uintptr) Weaver

	// AppendFloat32 adds a float32 and returns the builder.
	AppendFloat32(f float32) Weaver

	// AppendFloat64 adds a float64 and returns the builder.
	AppendFloat64(f float64) Weaver

	// AppendBool adds a boolean value and returns the builder.
	AppendBool(b bool) Weaver

	// Space adds a single space character and returns the builder.
	Space() Weaver

	// Spaces adds n space characters and returns the builder.
	Spaces(n int) Weaver

	// Tab adds a single tab character and returns the builder.
	Tab() Weaver

	// Tabs adds n tab characters and returns the builder.
	Tabs(n int) Weaver

	// NewLine adds a single newline character and returns the builder.
	NewLine() Weaver

	// NewLines adds n newline characters and returns the builder.
	NewLines(n int) Weaver

	// Line adds a string followed by a newline and returns the builder.
	Line(s string) Weaver

	// LineF adds a formatted string followed by a newline and returns the builder.
	LineF(format string, args ...any) Weaver

	// Repeat adds a string n times and returns the builder.
	Repeat(s string, n int) Weaver

	// Join adds strings with a separator and returns the builder.
	Join(sep string, elements ...string) Weaver

	// AppendIf conditionally adds a string if the condition is true.
	AppendIf(condition bool, s string) Weaver

	// AppendIfF conditionally adds a formatted string if the condition is true.
	AppendIfF(condition bool, format string, args ...any) Weaver

	// Indent adds indentation (2 spaces per level) before appending text.
	Indent(level int, s string) Weaver

	// IndentLine adds indentation (2 spaces per level) before text and ends with a newline.
	IndentLine(level int, s string) Weaver

	// Wrap wraps text with a prefix and suffix.
	Wrap(prefix, text, suffix string) Weaver

	// Quote wraps text in double quotes.
	Quote(s string) Weaver

	// SingleQuote wraps text in single quotes.
	SingleQuote(s string) Weaver

	// Parenthesize wraps text in parentheses.
	Parenthesize(s string) Weaver

	// Bracket wraps text in square brackets.
	Bracket(s string) Weaver

	// Brace wraps text in curly braces.
	Brace(s string) Weaver

	// Comma adds a comma to the builder.
	Comma() Weaver

	// Dot adds a period to the builder.
	Dot() Weaver

	// Colon adds a colon to the builder.
	Colon() Weaver

	// Semicolon adds a semicolon to the builder.
	Semicolon() Weaver

	// Equals adds an equals sign to the builder.
	Equals() Weaver

	// Arrow adds an arrow (->) to the builder.
	Arrow() Weaver

	// FatArrow adds a fat arrow (=>) to the builder.
	FatArrow() Weaver

	// Grow grows the builder's capacity by n bytes.
	Grow(n int) Weaver

	// Reset clears the builder and returns it for reuse.
	Reset() Weaver

	// Len returns the current length of the built string.
	Len() int

	// Cap returns the current capacity of the builder.
	Cap() int

	// String returns the final built string.
	String() string

	// Build is an alias for String() for fluent API consistency.
	Build() string

	// Clone creates an independent copy of the current builder state.
	Clone() Weaver

	// Inspect executes a function with access to the current state without modification.
	Inspect(fn func(current string)) Weaver

	// Builder returns the underlying strings.Builder for advanced operations.
	Builder() *strings.Builder

	// IndentF adds indentation (2 spaces per level) before appending a formatted string.
	IndentF(level int, format string, args ...any) Weaver

	// IndentLineF adds indentation (2 spaces per level) before a formatted string and ends with a newline.
	IndentLineF(level int, format string, args ...any) Weaver

	// JSONObjectStart adds an opening curly brace for a JSON object.
	JSONObjectStart() Weaver

	// JSONObjectEnd adds a closing curly brace for a JSON object.
	JSONObjectEnd() Weaver

	// JSONArrayStart adds an opening square bracket for a JSON array.
	JSONArrayStart() Weaver

	// JSONArrayEnd adds a closing square bracket for a JSON array.
	JSONArrayEnd() Weaver

	// JSONString adds a quoted and escaped string value.
	JSONString(s string) Weaver

	// JSONKey adds a quoted key followed by a colon and space.
	JSONKey(key string) Weaver

	// JSONKeyString adds a key-value pair where the value is a string.
	JSONKeyString(key, value string) Weaver

	// JSONKeyInt adds a key-value pair where the value is an integer.
	JSONKeyInt(key string, value int) Weaver

	// JSONKeyBool adds a key-value pair where the value is a boolean.
	JSONKeyBool(key string, value bool) Weaver

	// JSONKeyFloat adds a key-value pair where the value is a float.
	JSONKeyFloat(key string, value float64) Weaver

	// JSONFieldString adds an indented JSON field (key: value) with optional comma and newline.
	JSONFieldString(level int, key, value string, addComma bool) Weaver

	// JSONFieldInt adds an indented JSON field with an integer value.
	JSONFieldInt(level int, key string, value int, addComma bool) Weaver

	// JSONFieldInt8 adds an indented JSON field with an int8 value.
	JSONFieldInt8(level int, key string, value int8, addComma bool) Weaver

	// JSONFieldInt16 adds an indented JSON field with an int16 value.
	JSONFieldInt16(level int, key string, value int16, addComma bool) Weaver

	// JSONFieldInt32 adds an indented JSON field with an int32 value.
	JSONFieldInt32(level int, key string, value int32, addComma bool) Weaver

	// JSONFieldInt64 adds an indented JSON field with an int64 value.
	JSONFieldInt64(level int, key string, value int64, addComma bool) Weaver

	// JSONFieldUint adds an indented JSON field with a uint value.
	JSONFieldUint(level int, key string, value uint, addComma bool) Weaver

	// JSONFieldUint8 adds an indented JSON field with a uint8 value.
	JSONFieldUint8(level int, key string, value uint8, addComma bool) Weaver

	// JSONFieldUint16 adds an indented JSON field with a uint16 value.
	JSONFieldUint16(level int, key string, value uint16, addComma bool) Weaver

	// JSONFieldUint32 adds an indented JSON field with a uint32 value.
	JSONFieldUint32(level int, key string, value uint32, addComma bool) Weaver

	// JSONFieldUint64 adds an indented JSON field with a uint64 value.
	JSONFieldUint64(level int, key string, value uint64, addComma bool) Weaver

	// JSONFieldFloat32 adds an indented JSON field with a float32 value.
	JSONFieldFloat32(level int, key string, value float32, addComma bool) Weaver

	// JSONFieldFloat64 adds an indented JSON field with a float64 value.
	JSONFieldFloat64(level int, key string, value float64, addComma bool) Weaver

	// JSONFieldBool adds an indented JSON field with a bool value.
	JSONFieldBool(level int, key string, value bool, addComma bool) Weaver

	// CommaIfNotLast adds a comma if the index is not the last item.
	CommaIfNotLast(index, total int) Weaver

	// WhenCast executes a function with access to the current state without modification.
	WhenCast(condition bool, fn func(w Weaver)) Weaver

	// UnlessCast executes a function with access to the current state without modification.
	UnlessCast(condition bool, fn func(w Weaver)) Weaver

	// EachCast executes a function for each item in the slice.
	EachCast(items []string, fn func(w Weaver, item string)) Weaver

	// JSONFieldArrayStart adds an indented JSON array start.
	//
	// Example:
	//
	//	sw.JSONFieldArrayStart(1, "items") // adds '  "items": ['
	JSONFieldArrayStart(level int, key string) Weaver

	// JSONFieldArrayEnd adds an indented JSON array end.
	//
	// Example:
	//
	//	sw.JSONFieldArrayEnd(1) // adds '  ]'
	JSONFieldArrayEnd(level int) Weaver

	// StartParenthesize adds an opening parenthesis.
	//
	// Example:
	//
	//	sw.StartParenthesize() // adds '('
	StartParenthesize() Weaver

	// EndParenthesize adds a closing parenthesis.
	//
	// Example:
	//
	//	sw.EndParenthesize() // adds ')'
	EndParenthesize() Weaver

	// StartBracket adds an opening bracket.
	//
	// Example:
	//
	//	sw.StartBracket() // adds '['
	StartBracket() Weaver

	// EndBracket adds a closing bracket.
	//
	// Example:
	//
	//	sw.EndBracket() // adds ']'
	EndBracket() Weaver

	// StartBrace adds an opening curly brace.
	//
	// Example:
	//
	//	sw.StartBrace() // adds '{'
	StartBrace() Weaver

	// EndBrace adds a closing curly brace.
	//
	// Example:
	//
	//	sw.EndBrace() // adds '}'
	EndBrace() Weaver

	// IndentOnly adds indentation (2 spaces per level) without appending text.
	//
	// Example:
	//
	//	sw.IndentOnly(1) // adds '  '
	IndentOnly(level int) Weaver

	// IndentOnlyLine adds indentation (2 spaces per level) followed by a newline.
	//
	// Example:
	//
	//	sw.IndentOnlyLine(1) // adds '  \n'
	IndentOnlyLine(level int) Weaver

	// CarriageReturn adds a carriage return character and returns the builder.
	//
	// Example:
	//
	//	sw.CarriageReturn() // adds '\r'
	CarriageReturn() Weaver

	// Pipe adds a pipe character and returns the builder.
	//
	// Example:
	//
	//	sw.Pipe() // adds '|'
	Pipe() Weaver

	// Pipes adds n pipe characters and returns the builder.
	//
	// Example:
	//
	//	sw.Pipes(3) // adds '|||'
	Pipes(n int) Weaver

	// Dash adds a '-' character and returns the builder.
	//
	// Example:
	//
	//	sw.Dash() // adds '-'
	Dash() Weaver

	// Dashes adds n '-' characters and returns the builder.
	//
	// Example:
	//
	//	sw.Dashes(3) // adds '---'
	Dashes(n int) Weaver

	// Underscore adds a '_' character and returns the builder.
	//
	// Example:
	//
	//	sw.Underscore() // adds '_'
	Underscore() Weaver

	// Underscores adds n '_' characters and returns the builder.
	//
	// Example:
	//
	//	sw.Underscores(3) // adds '___'
	Underscores(n int) Weaver

	// CodeBlock adds a code block with optional language specifier.
	//
	// Example:
	//
	// 	sw.CodeBlock("go", "fmt.Println(\"Hello, World!\")") // adds '```go\nfmt.Println("Hello, World!")\n```'
	CodeBlock(language string, s Weaver) Weaver
}
