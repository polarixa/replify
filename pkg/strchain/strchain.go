package strchain

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// Compile-time interface satisfaction check.
var _ Weaver = (*StringWeaver)(nil)

// StringWeaver wraps strings.Builder with a fluent API for chainable operations.
// This implementation is NOT thread-safe and should only be used from a single goroutine.
// For concurrent access, use SafeStringWeaver instead.
//
// Example:
//
//	sw := strchain.New().Append("Hello").Space().Append("World")
//	fmt.Println(sw.String()) // Output: Hello World
type StringWeaver struct {
	builder strings.Builder
}

// From creates a new StringWeaver initialized with the given string.
//
// Example:
//
//	sw := strchain.From("Initial")
func From(s string) *StringWeaver {
	sw := New()
	sw.builder.WriteString(s)
	return sw
}

// FromPtr creates a new StringWeaver initialized with the given strings.Builder pointer.
//
// Example:
//
//	sw := strchain.FromPtr(&myBuilder)
func FromPtr(s *strings.Builder) *StringWeaver {
	if s == nil {
		return New()
	}
	sw := New()
	sw.builder.WriteString(s.String())
	return sw
}

// New creates a new StringWeaver instance.
//
// Example:
//
//	sw := strchain.New()
func New() *StringWeaver {
	return &StringWeaver{}
}

// NewWithCapacity creates a new StringWeaver with an initial capacity hint.
//
// Example:
//
//	sw := strchain.NewWithCapacity(1024)
func NewWithCapacity(capacity int) *StringWeaver {
	sw := &StringWeaver{}
	sw.builder.Grow(capacity)
	return sw
}

// Append adds a string and returns the builder for chaining.
//
// Example:
//
//	sw.Append("hello").Append(" world")
func (sw *StringWeaver) Append(s string) Weaver {
	sw.builder.WriteString(s)
	return sw
}

// AppendF adds a formatted string (printf-style) and returns the builder.
//
// Example:
//
//	sw.AppendF("value: %d", 42)
func (sw *StringWeaver) AppendF(format string, args ...any) Weaver {
	fmt.Fprintf(&sw.builder, format, args...)
	return sw
}

// AppendByte adds a single byte and returns the builder.
//
// Example:
//
//	sw.AppendByte('a')
func (sw *StringWeaver) AppendByte(b byte) Weaver {
	sw.builder.WriteByte(b)
	return sw
}

// AppendRune adds a single rune and returns the builder.
//
// Example:
//
//	sw.AppendRune('⌘')
func (sw *StringWeaver) AppendRune(r rune) Weaver {
	sw.builder.WriteRune(r)
	return sw
}

// AppendBytes adds a byte slice and returns the builder.
//
// Example:
//
//	sw.AppendBytes([]byte("data"))
func (sw *StringWeaver) AppendBytes(b []byte) Weaver {
	sw.builder.Write(b)
	return sw
}

// AppendInt adds an integer and returns the builder.
//
// Example:
//
//	sw.AppendInt(123)
func (sw *StringWeaver) AppendInt(i int) Weaver {
	sw.builder.WriteString(strconv.Itoa(i))
	return sw
}

// AppendInt8 adds an int8 and returns the builder.
//
// Example:
//
//	sw.AppendInt8(8)
func (sw *StringWeaver) AppendInt8(i int8) Weaver {
	sw.builder.WriteString(strconv.FormatInt(int64(i), 10))
	return sw
}

// AppendInt16 adds an int16 and returns the builder.
//
// Example:
//
//	sw.AppendInt16(16)
func (sw *StringWeaver) AppendInt16(i int16) Weaver {
	sw.builder.WriteString(strconv.FormatInt(int64(i), 10))
	return sw
}

// AppendInt32 adds an int32 and returns the builder.
//
// Example:
//
//	sw.AppendInt32(32)
func (sw *StringWeaver) AppendInt32(i int32) Weaver {
	sw.builder.WriteString(strconv.FormatInt(int64(i), 10))
	return sw
}

// AppendInt64 adds an int64 and returns the builder.
//
// Example:
//
//	sw.AppendInt64(64)
func (sw *StringWeaver) AppendInt64(i int64) Weaver {
	sw.builder.WriteString(strconv.FormatInt(i, 10))
	return sw
}

// AppendUint adds a uint and returns the builder.
//
// Example:
//
//	sw.AppendUint(123)
func (sw *StringWeaver) AppendUint(i uint) Weaver {
	sw.builder.WriteString(strconv.FormatUint(uint64(i), 10))
	return sw
}

// AppendUint8 adds a uint8 and returns the builder.
//
// Example:
//
//	sw.AppendUint8(8)
func (sw *StringWeaver) AppendUint8(i uint8) Weaver {
	sw.builder.WriteString(strconv.FormatUint(uint64(i), 10))
	return sw
}

// AppendUint16 adds a uint16 and returns the builder.
//
// Example:
//
//	sw.AppendUint16(16)
func (sw *StringWeaver) AppendUint16(i uint16) Weaver {
	sw.builder.WriteString(strconv.FormatUint(uint64(i), 10))
	return sw
}

// AppendUint32 adds a uint32 and returns the builder.
//
// Example:
//
//	sw.AppendUint32(32)
func (sw *StringWeaver) AppendUint32(i uint32) Weaver {
	sw.builder.WriteString(strconv.FormatUint(uint64(i), 10))
	return sw
}

// AppendUint64 adds a uint64 and returns the builder.
//
// Example:
//
//	sw.AppendUint64(64)
func (sw *StringWeaver) AppendUint64(i uint64) Weaver {
	sw.builder.WriteString(strconv.FormatUint(i, 10))
	return sw
}

// AppendUintptr adds a uintptr and returns the builder.
//
// Example:
//
//	sw.AppendUintptr(0xdeadbeef)
func (sw *StringWeaver) AppendUintptr(i uintptr) Weaver {
	sw.builder.WriteString(strconv.FormatUint(uint64(i), 10))
	return sw
}

// AppendFloat32 adds a float32 and returns the builder.
//
// Example:
//
//	sw.AppendFloat32(3.14)
func (sw *StringWeaver) AppendFloat32(f float32) Weaver {
	sw.builder.WriteString(strconv.FormatFloat(float64(f), 'f', -1, 32))
	return sw
}

// AppendFloat64 adds a float64 and returns the builder.
//
// Example:
//
//	sw.AppendFloat64(3.14159)
func (sw *StringWeaver) AppendFloat64(f float64) Weaver {
	sw.builder.WriteString(strconv.FormatFloat(f, 'f', -1, 64))
	return sw
}

// AppendBool adds a boolean value and returns the builder.
//
// Example:
//
//	sw.AppendBool(true)
func (sw *StringWeaver) AppendBool(b bool) Weaver {
	sw.builder.WriteString(strconv.FormatBool(b))
	return sw
}

// Space adds a single space character and returns the builder.
//
// Example:
//
//	sw.Append("Hello").Space().Append("World")
func (sw *StringWeaver) Space() Weaver {
	sw.builder.WriteByte(' ')
	return sw
}

// Spaces adds n space characters and returns the builder.
//
// Example:
//
//	sw.Append("Key:").Spaces(4).Append("Value")
func (sw *StringWeaver) Spaces(n int) Weaver {
	if n > 0 {
		sw.builder.WriteString(strings.Repeat(" ", n))
	}
	return sw
}

// Tab adds a single tab character and returns the builder.
//
// Example:
//
//	sw.Append("Column1").Tab().Append("Column2")
func (sw *StringWeaver) Tab() Weaver {
	sw.builder.WriteByte('\t')
	return sw
}

// Tabs adds n tab characters and returns the builder.
//
// Example:
//
//	sw.Tabs(2).Append("Indented text")
func (sw *StringWeaver) Tabs(n int) Weaver {
	if n > 0 {
		sw.builder.WriteString(strings.Repeat("\t", n))
	}
	return sw
}

// NewLine adds a single newline character and returns the builder.
//
// Example:
//
//	sw.Append("Line 1").NewLine().Append("Line 2")
func (sw *StringWeaver) NewLine() Weaver {
	sw.builder.WriteByte('\n')
	return sw
}

// NewLines adds n newline characters and returns the builder.
//
// Example:
//
//	sw.Append("Paragraph 1").NewLines(2).Append("Paragraph 2")
func (sw *StringWeaver) NewLines(n int) Weaver {
	if n > 0 {
		sw.builder.WriteString(strings.Repeat("\n", n))
	}
	return sw
}

// Line adds a string followed by a newline and returns the builder.
//
// Example:
//
//	sw.Line("First Line").Line("Second Line")
func (sw *StringWeaver) Line(s string) Weaver {
	sw.builder.WriteString(s)
	sw.builder.WriteByte('\n')
	return sw
}

// LineF adds a formatted string followed by a newline and returns the builder.
//
// Example:
//
//	sw.LineF("User ID: %d", 123)
func (sw *StringWeaver) LineF(format string, args ...any) Weaver {
	fmt.Fprintf(&sw.builder, format, args...)
	sw.builder.WriteByte('\n')
	return sw
}

// Repeat adds a string n times and returns the builder.
//
// Example:
//
//	sw.Repeat("-", 10) // adds "----------"
func (sw *StringWeaver) Repeat(s string, n int) Weaver {
	if n > 0 {
		sw.builder.WriteString(strings.Repeat(s, n))
	}
	return sw
}

// Join adds strings with a separator and returns the builder.
//
// Example:
//
//	sw.Join(", ", "Apple", "Banana", "Cherry")
func (sw *StringWeaver) Join(sep string, elements ...string) Weaver {
	for i, elem := range elements {
		if i > 0 {
			sw.builder.WriteString(sep)
		}
		sw.builder.WriteString(elem)
	}
	return sw
}

// AppendIf conditionally adds a string if the condition is true.
//
// Example:
//
//	sw.AppendIf(isValid, " Validated")
func (sw *StringWeaver) AppendIf(condition bool, s string) Weaver {
	if condition {
		sw.builder.WriteString(s)
	}
	return sw
}

// AppendIfF conditionally adds a formatted string if the condition is true.
//
// Example:
//
//	sw.AppendIfF(debugMode, "Debug: %s", msg)
func (sw *StringWeaver) AppendIfF(condition bool, format string, args ...any) Weaver {
	if condition {
		fmt.Fprintf(&sw.builder, format, args...)
	}
	return sw
}

// When executes a function on the builder if the condition is true.
//
// Example:
//
//	sw.When(isHeader, func(s *StringWeaver) {
//	    s.Append("# ").Line(title)
//	})
func (sw *StringWeaver) When(condition bool, fn func(*StringWeaver)) *StringWeaver {
	if condition {
		fn(sw)
	}
	return sw
}

// Unless executes a function on the builder if the condition is false.
//
// Example:
//
//	sw.Unless(isMinified, func(s *StringWeaver) {
//	    s.NewLine().Indent(1, "")
//	})
func (sw *StringWeaver) Unless(condition bool, fn func(*StringWeaver)) *StringWeaver {
	if !condition {
		fn(sw)
	}
	return sw
}

// Each iterates over a slice and applies a function for each element.
//
// Example:
//
//	sw.Each([]string{"a", "b", "c"}, func(s *StringWeaver, item string) {
//	    s.Append(item).Comma()
//	})
func (sw *StringWeaver) Each(items []string, fn func(*StringWeaver, string)) *StringWeaver {
	for _, item := range items {
		fn(sw, item)
	}
	return sw
}

// Indent adds indentation (2 spaces per level) before appending text.
//
// Example:
//
//	sw.Indent(2, "Sub-item")
func (sw *StringWeaver) Indent(level int, s string) Weaver {
	if level > 0 {
		sw.builder.WriteString(strings.Repeat(" ", level*2))
	}
	sw.builder.WriteString(s)
	return sw
}

// IndentLine adds indentation (2 spaces per level) before text and ends with a newline.
//
// Example:
//
//	sw.IndentLine(1, "Point 1")
func (sw *StringWeaver) IndentLine(level int, s string) Weaver {
	if level > 0 {
		sw.builder.WriteString(strings.Repeat(" ", level*2))
	}
	sw.builder.WriteString(s)
	sw.builder.WriteByte('\n')
	return sw
}

// Wrap wraps text with a prefix and suffix.
//
// Example:
//
//	sw.Wrap("**", "Bold text", "**")
func (sw *StringWeaver) Wrap(prefix, text, suffix string) Weaver {
	sw.builder.WriteString(prefix)
	sw.builder.WriteString(text)
	sw.builder.WriteString(suffix)
	return sw
}

// Quote wraps text in double quotes.
//
// Example:
//
//	sw.Quote("quoted string") // adds "quoted string"
func (sw *StringWeaver) Quote(s string) Weaver {
	sw.builder.WriteByte('"')
	sw.builder.WriteString(s)
	sw.builder.WriteByte('"')
	return sw
}

// SingleQuote wraps text in single quotes.
//
// Example:
//
//	sw.SingleQuote("char") // adds 'char'
func (sw *StringWeaver) SingleQuote(s string) Weaver {
	sw.builder.WriteByte('\'')
	sw.builder.WriteString(s)
	sw.builder.WriteByte('\'')
	return sw
}

// Parenthesize wraps text in parentheses.
//
// Example:
//
//	sw.Parenthesize("expression") // adds (expression)
func (sw *StringWeaver) Parenthesize(s string) Weaver {
	sw.builder.WriteByte('(')
	sw.builder.WriteString(s)
	sw.builder.WriteByte(')')
	return sw
}

// Bracket wraps text in square brackets.
//
// Example:
//
//	sw.Bracket("index") // adds [index]
func (sw *StringWeaver) Bracket(s string) Weaver {
	sw.builder.WriteByte('[')
	sw.builder.WriteString(s)
	sw.builder.WriteByte(']')
	return sw
}

// Brace wraps text in curly braces.
//
// Example:
//
//	sw.Brace("struct definition") // adds {struct definition}
func (sw *StringWeaver) Brace(s string) Weaver {
	sw.builder.WriteByte('{')
	sw.builder.WriteString(s)
	sw.builder.WriteByte('}')
	return sw
}

// Comma adds a comma to the builder.
//
// Example:
//
//	sw.Append("item1").Comma().Append("item2")
func (sw *StringWeaver) Comma() Weaver {
	sw.builder.WriteByte(',')
	return sw
}

// Dot adds a period to the builder.
//
// Example:
//
//	sw.Append("End of sentence").Dot()
func (sw *StringWeaver) Dot() Weaver {
	sw.builder.WriteByte('.')
	return sw
}

// Colon adds a colon to the builder.
//
// Example:
//
//	sw.Append("Label").Colon().Space().Append("Value")
func (sw *StringWeaver) Colon() Weaver {
	sw.builder.WriteByte(':')
	return sw
}

// Semicolon adds a semicolon to the builder.
//
// Example:
//
//	sw.Append("x = 1").Semicolon()
func (sw *StringWeaver) Semicolon() Weaver {
	sw.builder.WriteByte(';')
	return sw
}

// Equals adds an equals sign to the builder.
//
// Example:
//
//	sw.Append("key").Equals().Append("value")
func (sw *StringWeaver) Equals() Weaver {
	sw.builder.WriteByte('=')
	return sw
}

// Arrow adds an arrow (->) to the builder.
//
// Example:
//
//	sw.Append("source").Arrow().Append("target")
func (sw *StringWeaver) Arrow() Weaver {
	sw.builder.WriteString("->")
	return sw
}

// FatArrow adds a fat arrow (=>) to the builder.
//
// Example:
//
//	sw.Append("map").FatArrow().Append("result")
func (sw *StringWeaver) FatArrow() Weaver {
	sw.builder.WriteString("=>")
	return sw
}

// Grow grows the builder's capacity by n bytes.
//
// Example:
//
//	sw.Grow(512)
func (sw *StringWeaver) Grow(n int) Weaver {
	sw.builder.Grow(n)
	return sw
}

// Reset clears the builder and returns it for reuse.
//
// Example:
//
//	sw.Reset().Append("Fresh start")
func (sw *StringWeaver) Reset() Weaver {
	sw.builder.Reset()
	return sw
}

// Len returns the current length of the built string.
//
// Example:
//
//	length := sw.Len()
func (sw *StringWeaver) Len() int {
	return sw.builder.Len()
}

// Cap returns the current capacity of the builder.
//
// Example:
//
//	capacity := sw.Cap()
func (sw *StringWeaver) Cap() int {
	return sw.builder.Cap()
}

// String returns the final built string.
//
// Example:
//
//	result := sw.String()
func (sw *StringWeaver) String() string {
	return sw.builder.String()
}

// Build is an alias for String() for fluent API consistency.
//
// Example:
//
//	result := sw.Append("data").Build()
func (sw *StringWeaver) Build() string {
	return sw.builder.String()
}

// Clone creates an independent copy of the current builder state.
//
// Example:
//
//	newSw := sw.Clone()
func (sw *StringWeaver) Clone() Weaver {
	clone := New()
	clone.builder.WriteString(sw.builder.String())
	return clone
}

// Inspect executes a function with access to the current state without modification.
//
// Example:
//
//	sw.Inspect(func(current string) {
//	    log.Printf("Current state: %s", current)
//	})
func (sw *StringWeaver) Inspect(fn func(current string)) Weaver {
	fn(sw.builder.String())
	return sw
}

// Builder returns the underlying strings.Builder for advanced operations.
//
// Example:
//
//	builder := sw.Builder()
func (sw *StringWeaver) Builder() *strings.Builder {
	return &sw.builder
}

// IndentF adds indentation (2 spaces per level) before appending a formatted string.
//
// Example:
//
//	sw.IndentF(1, `"id": %q,`, "abc123")
func (sw *StringWeaver) IndentF(level int, format string, args ...any) Weaver {
	if level > 0 {
		sw.builder.WriteString(strings.Repeat(" ", level*2))
	}
	fmt.Fprintf(&sw.builder, format, args...)
	return sw
}

// IndentLineF adds indentation (2 spaces per level) before a formatted string and ends with a newline.
//
// Example:
//
//	sw.IndentLineF(1, `"id": %q,`, "abc123")
func (sw *StringWeaver) IndentLineF(level int, format string, args ...any) Weaver {
	if level > 0 {
		sw.builder.WriteString(strings.Repeat(" ", level*2))
	}
	fmt.Fprintf(&sw.builder, format, args...)
	sw.builder.WriteByte('\n')
	return sw
}

// JSONObjectStart adds an opening curly brace for a JSON object.
//
// Example:
//
//	sw.JSONObjectStart() // adds "{"
func (sw *StringWeaver) JSONObjectStart() Weaver {
	sw.builder.WriteByte('{')
	return sw
}

// JSONObjectEnd adds a closing curly brace for a JSON object.
//
// Example:
//
//	sw.JSONObjectEnd() // adds "}"
func (sw *StringWeaver) JSONObjectEnd() Weaver {
	sw.builder.WriteByte('}')
	return sw
}

// JSONArrayStart adds an opening square bracket for a JSON array.
//
// Example:
//
//	sw.JSONArrayStart() // adds "["
func (sw *StringWeaver) JSONArrayStart() Weaver {
	sw.builder.WriteByte('[')
	return sw
}

// JSONArrayEnd adds a closing square bracket for a JSON array.
//
// Example:
//
//	sw.JSONArrayEnd() // adds "]"
func (sw *StringWeaver) JSONArrayEnd() Weaver {
	sw.builder.WriteByte(']')
	return sw
}

// JSONString adds a quoted and escaped string value.
// The output is proper JSON (RFC 8259) — control characters are encoded as
// \uXXXX sequences, not Go-specific escape sequences.
//
// Example:
//
//	sw.JSONString("hello") // adds "hello"
func (sw *StringWeaver) JSONString(s string) Weaver {
	sw.builder.WriteString(jsonEncodeString(s))
	return sw
}

// JSONKey adds a quoted key followed by a colon and space.
//
// Example:
//
//	sw.JSONKey("name") // adds "name":
func (sw *StringWeaver) JSONKey(key string) Weaver {
	sw.builder.WriteByte('"')
	sw.builder.WriteString(key)
	sw.builder.WriteByte('"')
	sw.builder.WriteByte(':')
	sw.builder.WriteByte(' ')
	return sw
}

// JSONKeyString adds a key-value pair where the value is a string.
// The value is properly JSON-encoded (RFC 8259).
//
// Example:
//
//	sw.JSONKeyString("name", "John") // adds "name": "John"
func (sw *StringWeaver) JSONKeyString(key, value string) Weaver {
	sw.builder.WriteByte('"')
	sw.builder.WriteString(key)
	sw.builder.WriteByte('"')
	sw.builder.WriteByte(':')
	sw.builder.WriteByte(' ')
	sw.builder.WriteString(jsonEncodeString(value))
	return sw
}

// JSONKeyInt adds a key-value pair where the value is an integer.
//
// Example:
//
//	sw.JSONKeyInt("age", 30) // adds "age": 30
func (sw *StringWeaver) JSONKeyInt(key string, value int) Weaver {
	sw.builder.WriteByte('"')
	sw.builder.WriteString(key)
	sw.builder.WriteByte('"')
	sw.builder.WriteByte(':')
	sw.builder.WriteByte(' ')
	sw.builder.WriteString(strconv.Itoa(value))
	return sw
}

// JSONKeyBool adds a key-value pair where the value is a boolean.
//
// Example:
//
//	sw.JSONKeyBool("active", true) // adds "active": true
func (sw *StringWeaver) JSONKeyBool(key string, value bool) Weaver {
	sw.builder.WriteByte('"')
	sw.builder.WriteString(key)
	sw.builder.WriteByte('"')
	sw.builder.WriteByte(':')
	sw.builder.WriteByte(' ')
	sw.builder.WriteString(strconv.FormatBool(value))
	return sw
}

// JSONKeyFloat adds a key-value pair where the value is a float.
//
// Example:
//
//	sw.JSONKeyFloat("price", 19.99) // adds "price": 19.99
func (sw *StringWeaver) JSONKeyFloat(key string, value float64) Weaver {
	sw.builder.WriteByte('"')
	sw.builder.WriteString(key)
	sw.builder.WriteByte('"')
	sw.builder.WriteByte(':')
	sw.builder.WriteByte(' ')
	sw.builder.WriteString(strconv.FormatFloat(value, 'f', -1, 64))
	return sw
}

// JSONFieldString adds an indented JSON field (key: value) with optional comma and newline.
// The value is properly JSON-encoded (RFC 8259).
//
// Example:
//
//	sw.JSONFieldString(1, "name", `"John"`, true) // adds '  "name": "John",\n'
func (sw *StringWeaver) JSONFieldString(level int, key, value string, addComma bool) Weaver {
	if level > 0 {
		sw.builder.WriteString(strings.Repeat(" ", level*2))
	}
	sw.builder.WriteByte('"')
	sw.builder.WriteString(key)
	sw.builder.WriteByte('"')
	sw.builder.WriteByte(':')
	sw.builder.WriteByte(' ')
	sw.builder.WriteString(jsonEncodeString(value))
	if addComma {
		sw.builder.WriteByte(',')
	}
	sw.builder.WriteByte('\n')
	return sw
}

// JSONFieldInt adds an indented JSON field with an integer value.
//
// Example:
//
//	sw.JSONFieldInt(1, "age", 30, true) // adds '  "age": 30,\n'
func (sw *StringWeaver) JSONFieldInt(level int, key string, value int, addComma bool) Weaver {
	if level > 0 {
		sw.builder.WriteString(strings.Repeat(" ", level*2))
	}
	sw.builder.WriteByte('"')
	sw.builder.WriteString(key)
	sw.builder.WriteByte('"')
	sw.builder.WriteByte(':')
	sw.builder.WriteByte(' ')
	sw.builder.WriteString(strconv.FormatInt(int64(value), 10))
	if addComma {
		sw.builder.WriteByte(',')
	}
	sw.builder.WriteByte('\n')
	return sw
}

// JSONFieldInt8 adds an indented JSON field with an int8 value.
//
// Example:
//
//	sw.JSONFieldInt8(1, "age", 30, true) // adds '  "age": 30,\n'
func (sw *StringWeaver) JSONFieldInt8(level int, key string, value int8, addComma bool) Weaver {
	if level > 0 {
		sw.builder.WriteString(strings.Repeat(" ", level*2))
	}
	sw.builder.WriteByte('"')
	sw.builder.WriteString(key)
	sw.builder.WriteByte('"')
	sw.builder.WriteByte(':')
	sw.builder.WriteByte(' ')
	sw.builder.WriteString(strconv.FormatInt(int64(value), 10))
	if addComma {
		sw.builder.WriteByte(',')
	}
	sw.builder.WriteByte('\n')
	return sw
}

// JSONFieldInt16 adds an indented JSON field with an int16 value.
//
// Example:
//
//	sw.JSONFieldInt16(1, "age", 30, true) // adds '  "age": 30,\n'
func (sw *StringWeaver) JSONFieldInt16(level int, key string, value int16, addComma bool) Weaver {
	if level > 0 {
		sw.builder.WriteString(strings.Repeat(" ", level*2))
	}
	sw.builder.WriteByte('"')
	sw.builder.WriteString(key)
	sw.builder.WriteByte('"')
	sw.builder.WriteByte(':')
	sw.builder.WriteByte(' ')
	sw.builder.WriteString(strconv.FormatInt(int64(value), 10))
	if addComma {
		sw.builder.WriteByte(',')
	}
	sw.builder.WriteByte('\n')
	return sw
}

// JSONFieldInt32 adds an indented JSON field with an int32 value.
//
// Example:
//
//	sw.JSONFieldInt32(1, "age", 30, true) // adds '  "age": 30,\n'
func (sw *StringWeaver) JSONFieldInt32(level int, key string, value int32, addComma bool) Weaver {
	if level > 0 {
		sw.builder.WriteString(strings.Repeat(" ", level*2))
	}
	sw.builder.WriteByte('"')
	sw.builder.WriteString(key)
	sw.builder.WriteByte('"')
	sw.builder.WriteByte(':')
	sw.builder.WriteByte(' ')
	sw.builder.WriteString(strconv.FormatInt(int64(value), 10))
	if addComma {
		sw.builder.WriteByte(',')
	}
	sw.builder.WriteByte('\n')
	return sw
}

// JSONFieldInt64 adds an indented JSON field with an int64 value.
//
// Example:
//
//	sw.JSONFieldInt64(1, "age", 30, true) // adds '  "age": 30,\n'
func (sw *StringWeaver) JSONFieldInt64(level int, key string, value int64, addComma bool) Weaver {
	if level > 0 {
		sw.builder.WriteString(strings.Repeat(" ", level*2))
	}
	sw.builder.WriteByte('"')
	sw.builder.WriteString(key)
	sw.builder.WriteByte('"')
	sw.builder.WriteByte(':')
	sw.builder.WriteByte(' ')
	sw.builder.WriteString(strconv.FormatInt(value, 10))
	if addComma {
		sw.builder.WriteByte(',')
	}
	sw.builder.WriteByte('\n')
	return sw
}

// JSONFieldUint adds an indented JSON field with a uint value.
//
// Example:
//
//	sw.JSONFieldUint(1, "age", 30, true) // adds '  "age": 30,\n'
func (sw *StringWeaver) JSONFieldUint(level int, key string, value uint, addComma bool) Weaver {
	if level > 0 {
		sw.builder.WriteString(strings.Repeat(" ", level*2))
	}
	sw.builder.WriteByte('"')
	sw.builder.WriteString(key)
	sw.builder.WriteByte('"')
	sw.builder.WriteByte(':')
	sw.builder.WriteByte(' ')
	sw.builder.WriteString(strconv.FormatUint(uint64(value), 10))
	if addComma {
		sw.builder.WriteByte(',')
	}
	sw.builder.WriteByte('\n')
	return sw
}

// JSONFieldUint8 adds an indented JSON field with a uint8 value.
//
// Example:
//
//	sw.JSONFieldUint8(1, "age", 30, true) // adds '  "age": 30,\n'
func (sw *StringWeaver) JSONFieldUint8(level int, key string, value uint8, addComma bool) Weaver {
	if level > 0 {
		sw.builder.WriteString(strings.Repeat(" ", level*2))
	}
	sw.builder.WriteByte('"')
	sw.builder.WriteString(key)
	sw.builder.WriteByte('"')
	sw.builder.WriteByte(':')
	sw.builder.WriteByte(' ')
	sw.builder.WriteString(strconv.FormatUint(uint64(value), 10))
	if addComma {
		sw.builder.WriteByte(',')
	}
	sw.builder.WriteByte('\n')
	return sw
}

// JSONFieldUint16 adds an indented JSON field with a uint16 value.
//
// Example:
//
//	sw.JSONFieldUint16(1, "age", 30, true) // adds '  "age": 30,\n'
func (sw *StringWeaver) JSONFieldUint16(level int, key string, value uint16, addComma bool) Weaver {
	if level > 0 {
		sw.builder.WriteString(strings.Repeat(" ", level*2))
	}
	sw.builder.WriteByte('"')
	sw.builder.WriteString(key)
	sw.builder.WriteByte('"')
	sw.builder.WriteByte(':')
	sw.builder.WriteByte(' ')
	sw.builder.WriteString(strconv.FormatUint(uint64(value), 10))
	if addComma {
		sw.builder.WriteByte(',')
	}
	sw.builder.WriteByte('\n')
	return sw
}

// JSONFieldUint32 adds an indented JSON field with a uint32 value.
//
// Example:
//
//	sw.JSONFieldUint32(1, "age", 30, true) // adds '  "age": 30,\n'
func (sw *StringWeaver) JSONFieldUint32(level int, key string, value uint32, addComma bool) Weaver {
	if level > 0 {
		sw.builder.WriteString(strings.Repeat(" ", level*2))
	}
	sw.builder.WriteByte('"')
	sw.builder.WriteString(key)
	sw.builder.WriteByte('"')
	sw.builder.WriteByte(':')
	sw.builder.WriteByte(' ')
	sw.builder.WriteString(strconv.FormatUint(uint64(value), 10))
	if addComma {
		sw.builder.WriteByte(',')
	}
	sw.builder.WriteByte('\n')
	return sw
}

// JSONFieldUint64 adds an indented JSON field with a uint64 value.
//
// Example:
//
//	sw.JSONFieldUint64(1, "age", 30, true) // adds '  "age": 30,\n'
func (sw *StringWeaver) JSONFieldUint64(level int, key string, value uint64, addComma bool) Weaver {
	if level > 0 {
		sw.builder.WriteString(strings.Repeat(" ", level*2))
	}
	sw.builder.WriteByte('"')
	sw.builder.WriteString(key)
	sw.builder.WriteByte('"')
	sw.builder.WriteByte(':')
	sw.builder.WriteByte(' ')
	sw.builder.WriteString(strconv.FormatUint(value, 10))
	if addComma {
		sw.builder.WriteByte(',')
	}
	sw.builder.WriteByte('\n')
	return sw
}

// JSONFieldFloat32 adds an indented JSON field with a float32 value.
//
// Example:
//
//	sw.JSONFieldFloat32(1, "age", 30, true) // adds '  "age": 30,\n'
func (sw *StringWeaver) JSONFieldFloat32(level int, key string, value float32, addComma bool) Weaver {
	if level > 0 {
		sw.builder.WriteString(strings.Repeat(" ", level*2))
	}
	sw.builder.WriteByte('"')
	sw.builder.WriteString(key)
	sw.builder.WriteByte('"')
	sw.builder.WriteByte(':')
	sw.builder.WriteByte(' ')
	sw.builder.WriteString(strconv.FormatFloat(float64(value), 'f', -1, 32))
	if addComma {
		sw.builder.WriteByte(',')
	}
	sw.builder.WriteByte('\n')
	return sw
}

// JSONFieldFloat64 adds an indented JSON field with a float64 value.
//
// Example:
//
//	sw.JSONFieldFloat64(1, "age", 30, true) // adds '  "age": 30,\n'
func (sw *StringWeaver) JSONFieldFloat64(level int, key string, value float64, addComma bool) Weaver {
	if level > 0 {
		sw.builder.WriteString(strings.Repeat(" ", level*2))
	}
	sw.builder.WriteByte('"')
	sw.builder.WriteString(key)
	sw.builder.WriteByte('"')
	sw.builder.WriteByte(':')
	sw.builder.WriteByte(' ')
	sw.builder.WriteString(strconv.FormatFloat(value, 'f', -1, 64))
	if addComma {
		sw.builder.WriteByte(',')
	}
	sw.builder.WriteByte('\n')
	return sw
}

// JSONFieldBool adds an indented JSON field with a bool value.
//
// Example:
//
//	sw.JSONFieldBool(1, "age", 30, true) // adds '  "age": 30,\n'
func (sw *StringWeaver) JSONFieldBool(level int, key string, value bool, addComma bool) Weaver {
	if level > 0 {
		sw.builder.WriteString(strings.Repeat(" ", level*2))
	}
	sw.builder.WriteByte('"')
	sw.builder.WriteString(key)
	sw.builder.WriteByte('"')
	sw.builder.WriteByte(':')
	sw.builder.WriteByte(' ')
	sw.builder.WriteString(strconv.FormatBool(value))
	if addComma {
		sw.builder.WriteByte(',')
	}
	sw.builder.WriteByte('\n')
	return sw
}

// WhenCast executes a function with access to the current state without modification.
//
// Example:
//
//	sw.WhenCast(isHeader, func(w Weaver) {
//		w.Append("Header: ")
//	})
func (sw *StringWeaver) WhenCast(condition bool, fn func(w Weaver)) Weaver {
	if condition {
		fn(sw)
	}
	return sw
}

// UnlessCast executes a function with access to the current state without modification.
//
// Example:
//
//	sw.UnlessCast(isHeader, func(w Weaver) {
//		w.Append("Header: ")
//	})
func (sw *StringWeaver) UnlessCast(condition bool, fn func(w Weaver)) Weaver {
	if !condition {
		fn(sw)
	}
	return sw
}

// EachCast executes a function for each item in the slice.
//
// Example:
//
//	sw.EachCast(items, func(w Weaver, item string) {
//		w.Append(item)
//	})
func (sw *StringWeaver) EachCast(items []string, fn func(w Weaver, item string)) Weaver {
	for _, item := range items {
		fn(sw, item)
	}
	return sw
}

// CommaIfNotLast adds a comma if the index is not the last item.
//
// Example:
//
//	sw.CommaIfNotLast(i, len(items)) // adds "," if i < len(items)-1
func (sw *StringWeaver) CommaIfNotLast(index, total int) Weaver {
	if index < total-1 {
		sw.builder.WriteByte(',')
	}
	return sw
}

// JSONFieldArrayStart adds an indented JSON array start.
//
// Example:
//
//	sw.JSONFieldArrayStart(1, "items", true) // adds '  "items": ['
func (sw *StringWeaver) JSONFieldArrayStart(level int, key string) Weaver {
	if level > 0 {
		sw.builder.WriteString(strings.Repeat(" ", level*2))
	}
	sw.builder.WriteByte('"')
	sw.builder.WriteString(key)
	sw.builder.WriteByte('"')
	sw.builder.WriteByte(':')
	sw.builder.WriteByte(' ')
	sw.builder.WriteByte('[')
	return sw
}

// JSONFieldArrayEnd adds an indented JSON array end.
//
// Example:
//
//	sw.JSONFieldArrayEnd(1) // adds '  ]'
func (sw *StringWeaver) JSONFieldArrayEnd(level int) Weaver {
	if level > 0 {
		sw.builder.WriteString(strings.Repeat(" ", level*2))
	}
	sw.builder.WriteByte(']')
	return sw
}

// StartParenthesize adds an opening parenthesis.
//
// Example:
//
//	sw.StartParenthesize() // adds '('
func (sw *StringWeaver) StartParenthesize() Weaver {
	sw.builder.WriteByte('(')
	return sw
}

// EndParenthesize adds a closing parenthesis.
//
// Example:
//
//	sw.EndParenthesize() // adds ')'
func (sw *StringWeaver) EndParenthesize() Weaver {
	sw.builder.WriteByte(')')
	return sw
}

// StartBracket adds an opening bracket.
//
// Example:
//
//	sw.StartBracket() // adds '['
func (sw *StringWeaver) StartBracket() Weaver {
	sw.builder.WriteByte('[')
	return sw
}

// EndBracket adds a closing bracket.
//
// Example:
//
//	sw.EndBracket() // adds ']'
func (sw *StringWeaver) EndBracket() Weaver {
	sw.builder.WriteByte(']')
	return sw
}

// StartBrace adds an opening brace.
//
// Example:
//
//	sw.StartBrace() // adds '{'
func (sw *StringWeaver) StartBrace() Weaver {
	sw.builder.WriteByte('{')
	return sw
}

// EndBrace adds a closing brace.
//
// Example:
//
//	sw.EndBrace() // adds '}'
func (sw *StringWeaver) EndBrace() Weaver {
	sw.builder.WriteByte('}')
	return sw
}

// IndentOnly adds indentation (2 spaces per level) without appending text.
//
// Example:
//
//	sw.IndentOnly(1) // adds '  '
func (sw *StringWeaver) IndentOnly(level int) Weaver {
	if level > 0 {
		sw.builder.WriteString(strings.Repeat(" ", level*2))
	}
	return sw
}

// IndentOnlyLine adds indentation (2 spaces per level) followed by a newline.
//
// Example:
//
//	sw.IndentOnlyLine(1) // adds '  \n'
func (sw *StringWeaver) IndentOnlyLine(level int) Weaver {
	if level > 0 {
		sw.builder.WriteString(strings.Repeat(" ", level*2))
	}
	sw.builder.WriteByte('\n')
	return sw
}

// Bytes returns the built string as a byte slice.
//
// Example:
//
//	data := sw.Bytes()
func (sw *StringWeaver) Bytes() []byte {
	var buf bytes.Buffer
	buf.WriteString(sw.builder.String())
	return buf.Bytes()
}

// CarriageReturn adds a carriage return character and returns the builder.
//
// Example:
//
//	sw.Append("Line 1").CarriageReturn().Append("Line 2")
func (sw *StringWeaver) CarriageReturn() Weaver {
	sw.builder.WriteByte('\r')
	return sw
}

// Pipe adds a pipe character and returns the builder.
//
// Example:
//
//	sw.Append("Column1").Pipe().Append("Column2")
func (sw *StringWeaver) Pipe() Weaver {
	sw.builder.WriteByte('|')
	return sw
}

// Pipes adds n pipe characters and returns the builder.
//
// Example:
//
//	sw.Pipes(3) // adds "|||"
func (sw *StringWeaver) Pipes(n int) Weaver {
	if n > 0 {
		sw.builder.WriteString(strings.Repeat("|", n))
	}
	return sw
}

// Dash adds a dash character and returns the builder.
//
// Example:
//
//	sw.Append("Item1").Dash().Append("Item2")
func (sw *StringWeaver) Dash() Weaver {
	sw.builder.WriteByte('-')
	return sw
}

// Dashes adds n dash characters and returns the builder.
//
// Example:
//
//	sw.Dashes(5) // adds "-----"
func (sw *StringWeaver) Dashes(n int) Weaver {
	if n > 0 {
		sw.builder.WriteString(strings.Repeat("-", n))
	}
	return sw
}

// Underscore adds a single underscore character and returns the builder.
//
// Example:
//
//	sw.Append("variable").Underscore().Append("name")
func (sw *StringWeaver) Underscore() Weaver {
	sw.builder.WriteByte('_')
	return sw
}

// Underscores adds n underscore characters and returns the builder.
//
// Example:
//
//	sw.Underscores(4) // adds "____"
func (sw *StringWeaver) Underscores(n int) Weaver {
	if n > 0 {
		sw.builder.WriteString(strings.Repeat("_", n))
	}
	return sw
}

// CodeBlock adds a code block with the specified language and content.
//
// Example:
//
//	sw.CodeBlock("go", New().Append("fmt.Println(\"Hello, World!\")"))
func (sw *StringWeaver) CodeBlock(language string, s Weaver) Weaver {
	sw.builder.WriteString("```")
	sw.builder.WriteString(language)
	sw.builder.WriteByte('\n')
	sw.builder.WriteString(s.String())
	sw.builder.WriteByte('\n')
	sw.builder.WriteString("```")
	return sw
}

// IsEmpty checks if the builder is empty.
//
// Example:
//
//	if sw.IsEmpty() {
//	    fmt.Println("Builder is empty")
//	}
func (sw *StringWeaver) IsEmpty() bool {
	return sw.builder.Len() == 0
}

// IsNotEmpty checks if the builder is not empty.
//
// Example:
//
//	if sw.IsNotEmpty() {
//	    fmt.Println("Builder has content")
//	}
func (sw *StringWeaver) IsNotEmpty() bool {
	return sw.builder.Len() > 0
}

// InlineCode adds a string wrapped in backticks for code representation.
//
// Example:
//
//	sw.InlineCode("fmt.Println(\"Hello, World!\")") // adds `fmt.Println("Hello, World!")`
func (sw *StringWeaver) InlineCode(s string) Weaver {
	sw.builder.WriteString("`")
	sw.builder.WriteString(s)
	sw.builder.WriteString("`")
	return sw
}

// InlineCodeF adds a formatted inline code (printf-style) to the builder.
//
// Example:
//
//	sw.InlineCodeF("inline %s", "code") // adds '`inline code`'
func (sw *StringWeaver) InlineCodeF(format string, args ...any) Weaver {
	sw.builder.WriteString("`")
	fmt.Fprintf(&sw.builder, format, args...)
	sw.builder.WriteString("`")
	return sw
}

// Bold adds a string wrapped in double asterisks for bold representation.
//
// Example:
//
//	sw.Bold("Important") // adds "**Important**"
func (sw *StringWeaver) Bold(s string) Weaver {
	sw.builder.WriteString("**")
	sw.builder.WriteString(s)
	sw.builder.WriteString("**")
	return sw
}

// Italic adds a string wrapped in single asterisks for italic representation.
//
// Example:
//
//	sw.Italic("Emphasized") // adds "*Emphasized*"
func (sw *StringWeaver) Italic(s string) Weaver {
	sw.builder.WriteString("*")
	sw.builder.WriteString(s)
	sw.builder.WriteString("*")
	return sw
}

// Strikethrough adds a string wrapped in double tildes for strikethrough representation.
//
// Example:
//
//	sw.Strikethrough("Obsolete") // adds "~~Obsolete~~"
func (sw *StringWeaver) Strikethrough(s string) Weaver {
	sw.builder.WriteString("~~")
	sw.builder.WriteString(s)
	sw.builder.WriteString("~~")
	return sw
}

// Underline adds a string wrapped in <u> tags for underline representation.
//
// Example:
//
//	sw.Underline("Highlighted") // adds "<u>Highlighted</u>"
func (sw *StringWeaver) Underline(s string) Weaver {
	sw.builder.WriteString("<u>")
	sw.builder.WriteString(s)
	sw.builder.WriteString("</u>")
	return sw
}
