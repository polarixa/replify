package strchain

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// Compile-time interface satisfaction check.
var _ Weaver = (*SafeStringWeaver)(nil)

// SafeStringWeaver wraps strings.Builder with a fluent API for chainable operations.
// This implementation is THREAD-SAFE and can be used concurrently by multiple goroutines.
// Each method acquires a mutex lock to ensure safe concurrent access.
//
// For single-threaded usage where maximum performance is desired, use StringWeaver instead.
//
// Example:
//
//	sw := strchain.NewSafe().Append("Hello").Space().Append("World")
//	fmt.Println(sw.Build()) // Output: Hello World
type SafeStringWeaver struct {
	builder strings.Builder
	mu      sync.Mutex
}

// SafeFrom creates a new SafeStringWeaver initialized with the given string.
//
// Example:
//
//	sw := strchain.SafeFrom("Initial")
func SafeFrom(s string) *SafeStringWeaver {
	sw := NewSafe()
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteString(s)
	return sw
}

// SafeFromPtr creates a new SafeStringWeaver initialized with the given strings.Builder pointer.
//
// Example:
//
//	sw := strchain.SafeFromPtr(&myBuilder)
func SafeFromPtr(s *strings.Builder) *SafeStringWeaver {
	if s == nil {
		return NewSafe()
	}
	sw := NewSafe()
	sw.builder.WriteString(s.String())
	return sw
}

// NewSafe creates a new SafeStringWeaver instance.
//
// Example:
//
//	sw := strchain.NewSafe()
func NewSafe() *SafeStringWeaver {
	return &SafeStringWeaver{}
}

// NewSafeWithCapacity creates a new SafeStringWeaver with an initial capacity hint.
//
// Example:
//
//	sw := strchain.NewSafeWithCapacity(1024)
func NewSafeWithCapacity(capacity int) *SafeStringWeaver {
	sw := &SafeStringWeaver{}
	sw.builder.Grow(capacity)
	return sw
}

// Append adds a string and returns the builder for chaining.
//
// Example:
//
//	sw.Append("hello").Append(" world")
func (sw *SafeStringWeaver) Append(s string) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteString(s)
	return sw
}

// AppendF adds a formatted string (printf-style) and returns the builder.
//
// Example:
//
//	sw.AppendF("value: %d", 42)
func (sw *SafeStringWeaver) AppendF(format string, args ...any) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	fmt.Fprintf(&sw.builder, format, args...)
	return sw
}

// AppendByte adds a single byte and returns the builder.
//
// Example:
//
//	sw.AppendByte('a')
func (sw *SafeStringWeaver) AppendByte(b byte) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteByte(b)
	return sw
}

// AppendRune adds a single rune and returns the builder.
//
// Example:
//
//	sw.AppendRune('⌘')
func (sw *SafeStringWeaver) AppendRune(r rune) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteRune(r)
	return sw
}

// AppendBytes adds a byte slice and returns the builder.
//
// Example:
//
//	sw.AppendBytes([]byte("data"))
func (sw *SafeStringWeaver) AppendBytes(b []byte) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.Write(b)
	return sw
}

// AppendInt adds an integer and returns the builder.
//
// Example:
//
//	sw.AppendInt(123)
func (sw *SafeStringWeaver) AppendInt(i int) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteString(strconv.Itoa(i))
	return sw
}

// AppendInt8 adds an int8 and returns the builder.
//
// Example:
//
//	sw.AppendInt8(8)
func (sw *SafeStringWeaver) AppendInt8(i int8) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteString(strconv.FormatInt(int64(i), 10))
	return sw
}

// AppendInt16 adds an int16 and returns the builder.
//
// Example:
//
//	sw.AppendInt16(16)
func (sw *SafeStringWeaver) AppendInt16(i int16) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteString(strconv.FormatInt(int64(i), 10))
	return sw
}

// AppendInt32 adds an int32 and returns the builder.
//
// Example:
//
//	sw.AppendInt32(32)
func (sw *SafeStringWeaver) AppendInt32(i int32) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteString(strconv.FormatInt(int64(i), 10))
	return sw
}

// AppendInt64 adds an int64 and returns the builder.
//
// Example:
//
//	sw.AppendInt64(64)
func (sw *SafeStringWeaver) AppendInt64(i int64) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteString(strconv.FormatInt(i, 10))
	return sw
}

// AppendUint adds a uint and returns the builder.
//
// Example:
//
//	sw.AppendUint(123)
func (sw *SafeStringWeaver) AppendUint(i uint) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteString(strconv.FormatUint(uint64(i), 10))
	return sw
}

// AppendUint8 adds a uint8 and returns the builder.
//
// Example:
//
//	sw.AppendUint8(8)
func (sw *SafeStringWeaver) AppendUint8(i uint8) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteString(strconv.FormatUint(uint64(i), 10))
	return sw
}

// AppendUint16 adds a uint16 and returns the builder.
//
// Example:
//
//	sw.AppendUint16(16)
func (sw *SafeStringWeaver) AppendUint16(i uint16) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteString(strconv.FormatUint(uint64(i), 10))
	return sw
}

// AppendUint32 adds a uint32 and returns the builder.
//
// Example:
//
//	sw.AppendUint32(32)
func (sw *SafeStringWeaver) AppendUint32(i uint32) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteString(strconv.FormatUint(uint64(i), 10))
	return sw
}

// AppendUint64 adds a uint64 and returns the builder.
//
// Example:
//
//	sw.AppendUint64(64)
func (sw *SafeStringWeaver) AppendUint64(i uint64) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteString(strconv.FormatUint(i, 10))
	return sw
}

// AppendUintptr adds a uintptr and returns the builder.
//
// Example:
//
//	sw.AppendUintptr(0xdeadbeef)
func (sw *SafeStringWeaver) AppendUintptr(i uintptr) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteString(strconv.FormatUint(uint64(i), 10))
	return sw
}

// AppendFloat32 adds a float32 and returns the builder.
//
// Example:
//
//	sw.AppendFloat32(3.14)
func (sw *SafeStringWeaver) AppendFloat32(f float32) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteString(strconv.FormatFloat(float64(f), 'f', -1, 32))
	return sw
}

// AppendFloat64 adds a float64 and returns the builder.
//
// Example:
//
//	sw.AppendFloat64(3.14159)
func (sw *SafeStringWeaver) AppendFloat64(f float64) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteString(strconv.FormatFloat(f, 'f', -1, 64))
	return sw
}

// AppendBool adds a boolean value and returns the builder.
//
// Example:
//
//	sw.AppendBool(true)
func (sw *SafeStringWeaver) AppendBool(b bool) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteString(strconv.FormatBool(b))
	return sw
}

// Space adds a single space character and returns the builder.
//
// Example:
//
//	sw.Append("Hello").Space().Append("World")
func (sw *SafeStringWeaver) Space() Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteByte(' ')
	return sw
}

// Spaces adds n space characters and returns the builder.
//
// Example:
//
//	sw.Append("Key:").Spaces(4).Append("Value")
func (sw *SafeStringWeaver) Spaces(n int) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) Tab() Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteByte('\t')
	return sw
}

// Tabs adds n tab characters and returns the builder.
//
// Example:
//
//	sw.Tabs(2).Append("Indented text")
func (sw *SafeStringWeaver) Tabs(n int) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) NewLine() Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteByte('\n')
	return sw
}

// NewLines adds n newline characters and returns the builder.
//
// Example:
//
//	sw.Append("Paragraph 1").NewLines(2).Append("Paragraph 2")
func (sw *SafeStringWeaver) NewLines(n int) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) Line(s string) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteString(s)
	sw.builder.WriteByte('\n')
	return sw
}

// LineF adds a formatted string followed by a newline and returns the builder.
//
// Example:
//
//	sw.LineF("User ID: %d", 123)
func (sw *SafeStringWeaver) LineF(format string, args ...any) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	fmt.Fprintf(&sw.builder, format, args...)
	sw.builder.WriteByte('\n')
	return sw
}

// Repeat adds a string n times and returns the builder.
//
// Example:
//
//	sw.Repeat("-", 10) // adds "----------"
func (sw *SafeStringWeaver) Repeat(s string, n int) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) Join(sep string, elements ...string) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) AppendIf(condition bool, s string) Weaver {
	if !condition {
		return sw
	}
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteString(s)
	return sw
}

// AppendIfF conditionally adds a formatted string if the condition is true.
//
// Example:
//
//	sw.AppendIfF(debugMode, "Debug: %s", msg)
func (sw *SafeStringWeaver) AppendIfF(condition bool, format string, args ...any) Weaver {
	if !condition {
		return sw
	}
	sw.mu.Lock()
	defer sw.mu.Unlock()
	fmt.Fprintf(&sw.builder, format, args...)
	return sw
}

// When executes a function on the builder if the condition is true.
//
// Example:
//
//	sw.When(isHeader, func(s *SafeStringWeaver) {
//	    s.Append("# ").Line(title)
//	})
func (sw *SafeStringWeaver) When(condition bool, fn func(*SafeStringWeaver)) *SafeStringWeaver {
	if condition {
		fn(sw)
	}
	return sw
}

// Unless executes a function on the builder if the condition is false.
//
// Example:
//
//	sw.Unless(isMinified, func(s *SafeStringWeaver) {
//	    s.NewLine().Indent(1, "")
//	})
func (sw *SafeStringWeaver) Unless(condition bool, fn func(*SafeStringWeaver)) *SafeStringWeaver {
	if !condition {
		fn(sw)
	}
	return sw
}

// Each iterates over a slice and applies a function for each element.
//
// Example:
//
//	sw.Each([]string{"a", "b", "c"}, func(s *SafeStringWeaver, item string) {
//	    s.Append(item).Comma()
//	})
func (sw *SafeStringWeaver) Each(items []string, fn func(*SafeStringWeaver, string)) *SafeStringWeaver {
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
func (sw *SafeStringWeaver) Indent(level int, s string) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) IndentLine(level int, s string) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) Wrap(prefix, text, suffix string) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) Quote(s string) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) SingleQuote(s string) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) Parenthesize(s string) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) Bracket(s string) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) Brace(s string) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) Comma() Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteByte(',')
	return sw
}

// Dot adds a period to the builder.
//
// Example:
//
//	sw.Append("End of sentence").Dot()
func (sw *SafeStringWeaver) Dot() Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteByte('.')
	return sw
}

// Colon adds a colon to the builder.
//
// Example:
//
//	sw.Append("Label").Colon().Space().Append("Value")
func (sw *SafeStringWeaver) Colon() Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteByte(':')
	return sw
}

// Semicolon adds a semicolon to the builder.
//
// Example:
//
//	sw.Append("x = 1").Semicolon()
func (sw *SafeStringWeaver) Semicolon() Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteByte(';')
	return sw
}

// Equals adds an equals sign to the builder.
//
// Example:
//
//	sw.Append("key").Equals().Append("value")
func (sw *SafeStringWeaver) Equals() Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteByte('=')
	return sw
}

// Arrow adds an arrow (->) to the builder.
//
// Example:
//
//	sw.Append("source").Arrow().Append("target")
func (sw *SafeStringWeaver) Arrow() Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteString("->")
	return sw
}

// FatArrow adds a fat arrow (=>) to the builder.
//
// Example:
//
//	sw.Append("map").FatArrow().Append("result")
func (sw *SafeStringWeaver) FatArrow() Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteString("=>")
	return sw
}

// Grow grows the builder's capacity by n bytes.
//
// Example:
//
//	sw.Grow(512)
func (sw *SafeStringWeaver) Grow(n int) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.Grow(n)
	return sw
}

// Reset clears the builder and returns it for reuse.
//
// Example:
//
//	sw.Reset().Append("Fresh start")
func (sw *SafeStringWeaver) Reset() Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.Reset()
	return sw
}

// Len returns the current length of the built string.
//
// Example:
//
//	length := sw.Len()
func (sw *SafeStringWeaver) Len() int {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	return sw.builder.Len()
}

// Cap returns the current capacity of the builder.
//
// Example:
//
//	capacity := sw.Cap()
func (sw *SafeStringWeaver) Cap() int {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	return sw.builder.Cap()
}

// String returns the final built string.
//
// Example:
//
//	result := sw.String()
func (sw *SafeStringWeaver) String() string {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	return sw.builder.String()
}

// Build is an alias for String() for fluent API consistency.
//
// Example:
//
//	result := sw.Append("data").Build()
func (sw *SafeStringWeaver) Build() string {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	return sw.builder.String()
}

// Clone creates a thread-safe copy of the current builder state.
//
// Example:
//
//	newSw := sw.Clone()
func (sw *SafeStringWeaver) Clone() Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	clone := NewSafe()
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
func (sw *SafeStringWeaver) Inspect(fn func(current string)) Weaver {
	sw.mu.Lock()
	current := sw.builder.String()
	sw.mu.Unlock()
	fn(current)
	return sw
}

// Builder returns the underlying strings.Builder for advanced operations.
//
// WARNING: This method bypasses ALL thread-safety guarantees of SafeStringWeaver.
// The mutex is NOT held while the caller uses the returned pointer. Any concurrent
// read or write through this pointer — including calling other SafeStringWeaver
// methods simultaneously — will cause a data race. Use this method only when the
// SafeStringWeaver is exclusively owned by a single goroutine at that point in time,
// or when you have external synchronization. For normal concurrent use, prefer the
// fluent methods which each acquire and release the lock atomically.
//
// Example:
//
//	builder := sw.Builder()
func (sw *SafeStringWeaver) Builder() *strings.Builder {
	return &sw.builder
}

// IndentF adds indentation (2 spaces per level) before appending a formatted string.
//
// Example:
//
//	sw.IndentF(1, `"id": %q,`, "abc123")
func (sw *SafeStringWeaver) IndentF(level int, format string, args ...any) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) IndentLineF(level int, format string, args ...any) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
//	sw.JSONObjectStart() // adds "{" (thread-safe)
func (sw *SafeStringWeaver) JSONObjectStart() Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteByte('{')
	return sw
}

// JSONObjectEnd adds a closing curly brace for a JSON object.
//
// Example:
//
//	sw.JSONObjectEnd() // adds "}"
func (sw *SafeStringWeaver) JSONObjectEnd() Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteByte('}')
	return sw
}

// JSONArrayStart adds an opening square bracket for a JSON array.
//
// Example:
//
//	sw.JSONArrayStart() // adds "["
func (sw *SafeStringWeaver) JSONArrayStart() Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteByte('[')
	return sw
}

// JSONArrayEnd adds a closing square bracket for a JSON array.
//
// Example:
//
//	sw.JSONArrayEnd() // adds "]"
func (sw *SafeStringWeaver) JSONArrayEnd() Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) JSONString(s string) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteString(jsonEncodeString(s))
	return sw
}

// JSONKey adds a quoted key followed by a colon and space.
//
// Example:
//
//	sw.JSONKey("name") // adds "name":
func (sw *SafeStringWeaver) JSONKey(key string) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) JSONKeyString(key, value string) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) JSONKeyInt(key string, value int) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) JSONKeyBool(key string, value bool) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) JSONKeyFloat(key string, value float64) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) JSONFieldString(level int, key, value string, addComma bool) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) JSONFieldInt(level int, key string, value int, addComma bool) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) JSONFieldInt8(level int, key string, value int8, addComma bool) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) JSONFieldInt16(level int, key string, value int16, addComma bool) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) JSONFieldInt32(level int, key string, value int32, addComma bool) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) JSONFieldInt64(level int, key string, value int64, addComma bool) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) JSONFieldUint(level int, key string, value uint, addComma bool) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) JSONFieldUint8(level int, key string, value uint8, addComma bool) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) JSONFieldUint16(level int, key string, value uint16, addComma bool) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) JSONFieldUint32(level int, key string, value uint32, addComma bool) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) JSONFieldUint64(level int, key string, value uint64, addComma bool) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) JSONFieldFloat32(level int, key string, value float32, addComma bool) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) JSONFieldFloat64(level int, key string, value float64, addComma bool) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) JSONFieldBool(level int, key string, value bool, addComma bool) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) WhenCast(condition bool, fn func(w Weaver)) Weaver {
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
func (sw *SafeStringWeaver) UnlessCast(condition bool, fn func(w Weaver)) Weaver {
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
//		w.Append(item).Comma()
//	})
func (sw *SafeStringWeaver) EachCast(items []string, fn func(w Weaver, item string)) Weaver {
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
func (sw *SafeStringWeaver) CommaIfNotLast(index, total int) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	if index < total-1 {
		sw.builder.WriteByte(',')
	}
	return sw
}

// JSONFieldArrayStart adds an indented JSON array start.
//
// Example:
//
//	sw.JSONFieldArrayStart(1, "items") // adds '  "items": ['
func (sw *SafeStringWeaver) JSONFieldArrayStart(level int, key string) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) JSONFieldArrayEnd(level int) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) StartParenthesize() Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteByte('(')
	return sw
}

// EndParenthesize adds a closing parenthesis.
//
// Example:
//
//	sw.EndParenthesize() // adds ')'
func (sw *SafeStringWeaver) EndParenthesize() Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteByte(')')
	return sw
}

// StartBracket adds an opening bracket.
//
// Example:
//
//	sw.StartBracket() // adds '['
func (sw *SafeStringWeaver) StartBracket() Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteByte('[')
	return sw
}

// EndBracket adds a closing bracket.
//
// Example:
//
//	sw.EndBracket() // adds ']'
func (sw *SafeStringWeaver) EndBracket() Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteByte(']')
	return sw
}

// StartBrace adds an opening brace.
//
// Example:
//
//	sw.StartBrace() // adds '{'
func (sw *SafeStringWeaver) StartBrace() Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteByte('{')
	return sw
}

// EndBrace adds a closing brace.
//
// Example:
//
//	sw.EndBrace() // adds '}'
func (sw *SafeStringWeaver) EndBrace() Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteByte('}')
	return sw
}

// IndentOnly adds indentation (2 spaces per level) without appending text.
//
// Example:
//
//	sw.IndentOnly(1) // adds '  '
func (sw *SafeStringWeaver) IndentOnly(level int) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) IndentOnlyLine(level int) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
func (sw *SafeStringWeaver) Bytes() []byte {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	var buf bytes.Buffer
	buf.WriteString(sw.builder.String())
	return buf.Bytes()
}

// CarriageReturn adds a carriage return character and returns the builder.
//
// Example:
//
//	sw.Append("Line 1").CarriageReturn().Append("Line 2")
func (sw *SafeStringWeaver) CarriageReturn() Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteByte('\r')
	return sw
}

// Pipe adds a pipe character and returns the builder.
//
// Example:
//
//	sw.Append("Column 1").Pipe().Append("Column 2")
func (sw *SafeStringWeaver) Pipe() Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteByte('|')
	return sw
}

// Pipes adds n pipe characters and returns the builder.
//
// Example:
//
//	sw.Append("Column 1").Pipes(3).Append("Column 2")
func (sw *SafeStringWeaver) Pipes(n int) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	if n > 0 {
		sw.builder.WriteString(strings.Repeat("|", n))
	}
	return sw
}

// Dash adds a dash character and returns the builder.
//
// Example:
//
//	sw.Append("Item 1").Dash().Append("Item 2")
func (sw *SafeStringWeaver) Dash() Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteByte('-')
	return sw
}

// Dashes adds n dash characters and returns the builder.
//
// Example:
//
//	sw.Append("Item 1").Dashes(5).Append("Item 2")
func (sw *SafeStringWeaver) Dashes(n int) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	if n > 0 {
		sw.builder.WriteString(strings.Repeat("-", n))
	}
	return sw
}

// Underscore adds an underscore character and returns the builder.
//
// Example:
//
//	sw.Append("Item 1").Underscore().Append("Item 2")
func (sw *SafeStringWeaver) Underscore() Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.builder.WriteByte('_')
	return sw
}

// Underscores adds n underscore characters and returns the builder.
//
// Example:
//
//	sw.Append("Item 1").Underscores(5).Append("Item 2")
func (sw *SafeStringWeaver) Underscores(n int) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	if n > 0 {
		sw.builder.WriteString(strings.Repeat("_", n))
	}
	return sw
}

// CodeBlock adds a code block with the specified language and content.
//
// Example:
//
//	sw.CodeBlock("go", sw.Append("fmt.Println(\"Hello, World!\")"))
func (sw *SafeStringWeaver) CodeBlock(language string, s Weaver) Weaver {
	sw.mu.Lock()
	defer sw.mu.Unlock()
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
//	    // Handle empty case
//	}
func (sw *SafeStringWeaver) IsEmpty() bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	return sw.builder.Len() == 0
}

// IsNotEmpty checks if the builder is not empty.
//
// Example:
//
//	if sw.IsNotEmpty() {
//	    // Handle non-empty case
//	}
func (sw *SafeStringWeaver) IsNotEmpty() bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	return sw.builder.Len() > 0
}
