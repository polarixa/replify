package fj

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
	"unsafe"

	"github.com/polarixa/replify/pkg/conv"
	"github.com/polarixa/replify/pkg/match"
	"github.com/polarixa/replify/pkg/strutil"
)

// UnsafeBytes converts a string into a byte slice without allocating new memory for the data.
// This function uses unsafe operations to directly reinterpret the string's underlying data
// structure as a byte slice. This allows efficient access to the string's content as a mutable
// byte slice, but it also comes with risks.
//
// Parameters:
//   - `s`: The input string that needs to be converted to a byte slice.
//
// Returns:
//   - A byte slice (`[]byte`) that shares the same underlying data as the input string.
//
// Notes:
//   - This function leverages Go's `unsafe` package to bypass the usual safety mechanisms
//     of the Go runtime. It does this by manipulating memory layouts using `unsafe.Pointer`.
//   - The resulting byte slice must be treated with care. Modifying the byte slice can lead
//     to undefined behavior since strings in Go are immutable by design.
//   - Any operation that depends on the immutability of the original string should avoid using this function.
//
// Safety Considerations:
//   - Since this function operates on unsafe pointers, it is not portable across different
//     Go versions or architectures.
//   - Direct modifications to the returned byte slice will violate Go's immutability guarantees
//     for strings and may corrupt program state.
//
// Example Usage:
//
//	s := "immutable string"
//	b := UnsafeBytes(s) // Efficiently converts the string to []byte
//	// WARNING: Modifying 'b' here can lead to undefined behavior.
func UnsafeBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&sliceHeader{
		data: (*stringHeader)(unsafe.Pointer(&s)).data,
		n:    len(s),
		cap:  len(s),
	}))
}

// tokenizeNumber scans json starting at position 0—assumed to be the beginning
// of a numeric literal—and returns the raw numeric token and its float64 value.
//
// The scan proceeds forward until a delimiter that ends a JSON number is
// encountered: any ASCII whitespace, a comma ',', or a closing ']' or '}'.
// Characters commonly found inside JSON numbers (digits, '.', exponent markers
// 'e'/'E', and explicit signs '+'/'-' after an exponent) are treated as part of
// the token. The function is a lightweight tokenizer; it does not perform full
// JSON validation.
//
// Parameters:
//   - json: a string whose first byte must be part of a number (e.g., '-', '0'–'9').
//     The caller retains ownership and must not mutate it concurrently.
//
// Returns:
//   - raw: the zero-copy substring of json that forms the numeric token
//   - num: the float64 parsed from raw; if parsing fails, num is 0 and raw still
//     contains the token that was attempted
//
// Limitations & Nuances:
//   - Heuristic tokenization only: content is not validated beyond delimiter scanning.
//   - Parsing uses strconv.ParseFloat on the token and ignores the returned error,
//     yielding 0 on failure—callers should validate raw if 0 is ambiguous.
//   - Leading/trailing whitespace before the number is not skipped; json[0] must
//     start at the number for correct results.
//   - No allocations aside from those performed internally by ParseFloat; raw
//     shares backing storage with json.
//
// Examples:
//
//	// Typical number before a comma
//	raw, n := tokenizeNumber("123.45, more")
//	// raw == "123.45", n == 123.45
//
//	// Scientific notation ending at closing bracket
//	raw, n = tokenizeNumber("-1.2e+3]")
//	// raw == "-1.2e+3", n == -1200
//
//	// Invalid numeric token: parsing yields 0, but raw reflects the token
//	raw, n = tokenizeNumber("-]")
//	// raw == "-", n == 0
func tokenizeNumber(json string) (raw string, num float64) {
	for i := 1; i < len(json); i++ {
		// check for characters that signify the end of a numeric value.
		if json[i] <= '-' {
			// break if the character is a whitespace or comma.
			if json[i] <= ' ' || json[i] == ',' {
				raw = json[:i]
				num, _ = strconv.ParseFloat(raw, 64) // convert the numeric substring to a float.
				return
			}
			// if the character is '+' or '-', assume it could be part of the number.
		} else if json[i] == ']' || json[i] == '}' {
			// break on closing brackets or braces (']' or '}')
			raw = json[:i]
			num, _ = strconv.ParseFloat(raw, 64)
			return
		}
	}
	// if no delimiters are encountered, process the entire string.
	raw = json
	num, _ = strconv.ParseFloat(raw, 64) // convert the entire string to a float.
	return
}

// parseShared efficiently processes a JSON byte slice and a path string to produce a `Context`.
// This function minimizes memory allocations and copies, leveraging unsafe operations
// to handle large JSON strings and slice conversions.
//
// Parameters:
//   - `json`: A byte slice containing the JSON data to process.
//   - `path`: A string representing the path to extract data from the JSON.
//
// Returns:
//   - A `Context` struct containing processed and unprocessed strings representing
//     the result of applying the path to the JSON data.
//
// Notes:
//   - The function uses unsafe pointer operations to avoid unnecessary allocations and copies.
//   - It extracts string and byte slice headers and ensures memory safety by copying headers
//     to strings when needed.
//   - The function checks whether the substring (`str`) is part of the raw string (`raw`)
//     and handles memory overlap efficiently.
//
// Example:
//
//	jsonBytes := []byte(`{"key": "value", "nested": {"innerKey": "innerValue"}}`)
//	path := "nested.innerKey"
//	context := parseShared(jsonBytes, path)
//	fmt.Println("Unprocessed:", context.raw) // Output: `{"key": "value", "nested": {"innerKey": "innerValue"}}`
//	fmt.Println("Strings:", context.str)         // Output: `{"innerKey": "innerValue"}`
func parseShared(json []byte, path string) Context {
	var result Context
	if json != nil {
		// unsafe cast json bytes to a string and process it using the Get function.
		result = Get(*(*string)(unsafe.Pointer(&json)), path)
		// extract the string headers for unprocessed and strings.
		rawSafe := *(*stringHeader)(unsafe.Pointer(&result.raw))
		stringSafe := *(*stringHeader)(unsafe.Pointer(&result.str))
		// create byte slice headers for the raw and str fields.
		rawSliceHeader := sliceHeader{data: rawSafe.data, n: rawSafe.n, cap: rawSafe.n}
		strSliceHeader := sliceHeader{data: stringSafe.data, n: stringSafe.n, cap: rawSafe.n}
		// check for nil data and safely copy headers to strings if necessary.
		if strSliceHeader.data == nil {
			if rawSliceHeader.data == nil {
				result.raw = ""
			} else {
				// raw has data, safely copy the slice header to a string
				result.raw = string(*(*[]byte)(unsafe.Pointer(&rawSliceHeader)))
			}
			result.str = ""
		} else if rawSliceHeader.data == nil {
			result.raw = ""
			result.str = string(*(*[]byte)(unsafe.Pointer(&strSliceHeader)))
		} else if uintptr(strSliceHeader.data) >= uintptr(rawSliceHeader.data) &&
			uintptr(strSliceHeader.data)+uintptr(strSliceHeader.n) <=
				uintptr(rawSliceHeader.data)+uintptr(rawSliceHeader.n) {
			// str is a substring of raw.
			start := uintptr(strSliceHeader.data) - uintptr(rawSliceHeader.data)
			// safely copy the raw slice header
			result.raw = string(*(*[]byte)(unsafe.Pointer(&rawSliceHeader)))
			result.str = result.raw[start : start+uintptr(strSliceHeader.n)]
		} else {
			// safely copy both headers to strings.
			result.raw = string(*(*[]byte)(unsafe.Pointer(&rawSliceHeader)))
			result.str = string(*(*[]byte)(unsafe.Pointer(&strSliceHeader)))
		}
	}
	return result
}

// isSafePathKeyByte reports whether c is considered a "safe" byte for use in a
// path‑key segment.
//
// A byte is treated as safe if it falls within a restricted, ASCII‑only set
// suitable for lightweight path parsing and key matching. The accepted
// characters include:
//   - letters (A–Z, a–z)
//   - digits (0–9)
//   - '_', '-', ':'
//   - any control or whitespace byte (<= ' ')
//   - any non‑ASCII byte (> '~')
//
// This helper performs a simple classification and is intended for
// performance‑sensitive code that validates or tokenizes path segments.
// It assumes ASCII semantics and does not attempt normalization or Unicode
// category checks.
//
// Examples:
//
//	// Typical ASCII key
//	ok := isSafePathKeyByte('a')
//	// ok == true
//
//	// Disallowed punctuation
//	ok = isSafePathKeyByte('%')
//	// ok == false
//
//	// Non‑ASCII bytes are treated as safe
//	ok = isSafePathKeyByte(0xC3) // 'Ã' in UTF‑8
//	// ok == true
func isSafePathKeyByte(c byte) bool {
	return c <= ' ' || c > '~' || c == '_' || c == '-' || c == ':' ||
		(c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9')
}

// extractOutermostValue returns the syntactic outermost JSON value beginning at
// its start position within the provided string.
//
// The function walks the input backward to locate the boundary of the top‑level
// JSON value. It supports strings, arrays, objects, and parenthesized values by
// tracking bracket depth and correctly handling escaped quotes. This makes it
// suitable for lightweight extraction tasks where the input may contain
// surrounding text or multiple concatenated JSON-like fragments.
//
// This function performs a heuristic scan rather than full JSON parsing. It does
// not validate that the input is well‑formed JSON and may yield unexpected
// results on malformed or deeply irregular input. Escaped quotes inside strings
// are handled, but control characters and Unicode escapes are not interpreted.
// No allocations are introduced beyond the returned substring.
//
// Parameters:
//   - json: a raw JSON or JSON‑like string; must contain at least one syntactic
//     value at the end of the string. The caller owns the memory and must not
//     mutate it concurrently.
//
// Returns:
//   - string: a substring of the input representing the outermost JSON value,
//     sharing memory with the original string. If no structural boundaries can
//     be determined, the entire input is returned.
//
// Examples:
//
//	// Extracting a simple object
//	v := extractOutermostValue(`prefix {"a":1,"b":2}`)
//	// v == {"a":1,"b":2}
//
//	// Extracting an array from trailing content
//	v := extractOutermostValue(`header [1,2,3] footer`)
//	// v == [1,2,3]
//
//	// Strings with escaped quotes
//	v := extractOutermostValue(`data "he said \"hello\"" extra`)
//	// v == "he said \"hello\""
func extractOutermostValue(json string) string {
	i := len(json) - 1
	var depth int
	// if the last character is not a quote, assume it's part of a value and increase depth
	if json[i] != '"' {
		depth++
	}
	// if the last character is a closing bracket, brace, or parenthesis, adjust the index to skip it
	if json[i] == '}' || json[i] == ']' || json[i] == ')' {
		i--
	}
	// loop backwards through the string
	for ; i >= 0; i-- {
		switch json[i] {
		case '"':
			// handle strings enclosed in double quotes
			i-- // skip the opening quote
			for ; i >= 0; i-- {
				if json[i] == '"' {
					// check for escape sequences (e.g., \")
					esc := 0
					for i > 0 && json[i-1] == '\\' {
						i-- // move back over escape characters
						esc++
					}
					// if the quote is escaped, continue
					if esc%2 == 1 {
						continue
					}
					// if the quote is not escaped, break out of the loop
					i += esc
					break
				}
			}
			// if the depth is 0, we've found the outermost value
			if depth == 0 {
				if i < 0 {
					i = 0
				}
				return json[i:] // return the substring starting from the outermost value
			}
		case '}', ']', ')': // increase depth when encountering closing brackets, braces, or parentheses
			depth++
		case '{', '[', '(': // decrease depth when encountering opening brackets, braces, or parentheses
			depth--
			// if depth reaches 0, we've found the outermost value
			if depth == 0 {
				return json[i:] // return the substring starting from the outermost value
			}
		}
	}
	return json
}

// computeOffset sets c.val.idx to the byte offset of c.val.raw within the
// provided json string.
//
// The function assumes that c.val.raw is a substring (sharing backing storage)
// of json. When c.calc is false and c.val.raw is non‑empty, computeOffset derives
// the offset by comparing the underlying data pointers of both strings and
// writes the result into c.val.idx. If the computed offset falls outside the
// bounds of json, the index is clamped to 0.
//
// This routine is a low‑level, performance‑oriented helper that uses unsafe
// pointer arithmetic to avoid allocations and additional scans. It does not
// validate substring relationships; correctness relies on the invariant that
// c.val.raw was obtained from json without copying. If c.val.raw does not share
// the same backing array as json (e.g., constructed independently, decoded, or
// allocated separately), the computed index is undefined and will be coerced to
// 0 by the bounds check.
//
// Parameters:
//   - json: the original source string that backs c.val.raw
//   - c: a parser whose current value (c.val) contains raw and idx fields;
//     ownership remains with the caller and must not be mutated concurrently
//
// Side Effects & Performance:
//   - Writes to c.val.idx; no other state is modified.
//   - Zero additional allocations; uses unsafe pointer math for O(1) index
//     computation.
//   - No errors are returned; invalid relationships are handled by setting
//     c.val.idx to 0.
//
// Limitations & Nuances:
//   - Requires that c.val.raw and json share the same backing storage.
//   - Behavior is undefined if the strings originate from different allocations,
//     even if they contain identical content.
//   - Skips computation when c.calc is true or c.val.raw is empty.
//
// Example:
//
//	json := `{"key": "value"}`
//	value := Context{raw: `"value"`}
//	c := &parser{json: json, value: value}
//	computeOffset(json, c)
//	fmt.Println(c.val.idx) // Outputs the starting position of `"value"` in the JSON string.
func computeOffset(json string, c *parser) {
	if len(c.val.raw) > 0 && !c.calc {
		jsonHeader := *(*stringHeader)(unsafe.Pointer(&json))
		unprocessedHeader := *(*stringHeader)(unsafe.Pointer(&(c.val.raw)))
		c.val.idx = int(uintptr(unprocessedHeader.data) - uintptr(jsonHeader.data))
		if c.val.idx < 0 || c.val.idx >= len(json) {
			c.val.idx = 0
		}
	}
}

// leadingLowercase extracts the initial contiguous sequence of lowercase alphabetic characters
// ('a' to 'z') from the input string `json`. It stops when it encounters a character
// outside this range and returns the substring up to that point.
//
// Parameters:
//   - `json`: The input string from which the initial sequence of lowercase alphabetic
//     characters is extracted.
//
// Returns:
//   - `raw`: A substring of `json` containing only the leading lowercase alphabetic characters.
//     If the input string starts with no lowercase alphabetic characters, the return
//     value will be an empty string.
//
// Notes:
//   - The function starts iterating from the second character (index 1) since it assumes the
//     first character does not need validation or extraction.
//   - The comparison checks (`json[i] < 'a' || json[i] > 'z'`) ensure that the character
//     falls outside the range of lowercase alphabetic ASCII values.
//
// Examples:
//
//	leadingLowercase(`{"name": "Alice"}`)     // → "name"
//	leadingLowercase(`{"id": 42, "active": true}`)  // → "id"
//	leadingLowercase(`{"UserID": 123}`)       // → "" (first letter is uppercase)
//	leadingLowercase(`{`)                     // → ""
//	leadingLowercase(`{""}`)                  // → ""
//	leadingLowercase(`not json`)              // → ""
func leadingLowercase(json string) (raw string) {
	for i := 1; i < len(json); i++ {
		if json[i] < 'a' || json[i] > 'z' {
			return json[:i]
		}
	}
	return json
}

// compactJSON returns the shortest prefix of json that forms a complete top‑level
// JSON token and discards the rest. If json starts with a double quote, the
// token is treated as a JSON string literal; otherwise it is treated as a
// delimited structure and scanned by nesting depth until the matching closing
// brace/bracket/paren is found.
//
// The scanner is tolerant and lightweight: it does not perform full JSON
// validation and only tracks enough structure to identify the end of the first
// token. String literals are handled with proper recognition of escaped quotes
// by counting backslashes. For structural tokens, nesting depth is incremented
// on '{', '[', '(' and decremented on '}', ']', ')'.
//
// Parameters:
//   - json: non-empty input whose first byte must be either '"' (for a string
//     literal) or the beginning of a structured value such as '{' or '['.
//     Parentheses are also recognized for depth tracking even though they are
//     not part of standard JSON.
//
// Returns:
//   - string: a substring of the input containing exactly the first complete
//     top-level token. If the input does not contain a matching closing
//     delimiter (unterminated token), the original input is returned.
//
// Assumptions & Invariants:
//   - The input must be non-empty; otherwise this function will panic due to
//     indexing json[0].
//   - Escaped quotes inside string literals are identified by counting
//     consecutive backslashes immediately preceding a quote.
//   - Only byte-level scanning is performed; it does not interpret Unicode
//     escapes or validate numeric/boolean/null tokens.
//
// Limitations & Nuances:
//   - This is a heuristic tokenizer, not a complete JSON parser. It may accept
//     some non-JSON delimiters (parentheses) purely for depth tracking.
//   - It does not verify overall JSON correctness; it only finds the end of
//     the first top-level token if present.
//   - If no closing delimiter is found for structured input, the function
//     returns the original string unchanged.
//   - Only ASCII-relevant characters are examined for structure; other bytes
//     are scanned transparently.
//
// Performance & Memory:
//   - Runs in O(n) time over the input.
//   - Zero allocations: the returned value is a substring that shares the
//     original string’s backing storage; callers should copy it if long-term
//     retention without holding the full input is required.
//
// Examples:
//
//	// Object token
//	in := `{"a":1,"b":[2,3]} trailing`
//	out := compactJSON(in)
//	// out == `{"a":1,"b":[2,3]}`
//
//	// String with escaped quote
//	in = `"he said \"hi\"" extra`
//	out = compactJSON(in)
//	// out == `"he said \"hi\""`
//
//	// Unterminated structure returns original
//	in = `{"a":1`
//	out = compactJSON(in)
//	// out == `{"a":1`
func compactJSON(json string) string {
	var i, depth int
	// If the first character is not a quote, initialize i and depth for the JSON object/array parsing.
	if json[0] != '"' {
		i, depth = 1, 1
	}
	// Iterate through the string starting from index 1 to process the content.
	for ; i < len(json); i++ {
		// Process characters that are within the range of valid JSON characters (from '"' to '}').
		if json[i] >= '"' && json[i] <= '}' {
			switch json[i] {
			// Handle string literals, ensuring to escape any escaped quotes inside.
			case '"':
				i++
				s2 := i
				for ; i < len(json); i++ {
					if json[i] > '\\' {
						continue
					}
					// If an unescaped quote is found, break out of the loop.
					if json[i] == '"' {
						// look for an escaped slash
						if json[i-1] == '\\' {
							n := 0
							// Count the number of preceding backslashes.
							for j := i - 2; j > s2-1; j-- {
								if json[j] != '\\' {
									break
								}
								n++
							}
							// If there is an even number of backslashes, continue, as this quote is escaped.
							if n%2 == 0 {
								continue
							}
						}
						// If quote is found and it's not escaped, break the loop.
						break
					}
				}
				// If depth is 0, we've finished processing the top-level string, return it.
				if depth == 0 {
					if i >= len(json) {
						return json
					}
					return json[:i+1]
				}
			// Process nested objects/arrays (opening braces or brackets).
			case '{', '[', '(':
				depth++
			// Process closing of nested objects/arrays (closing braces, brackets, or parentheses).
			case '}', ']', ')':
				depth--
				// If depth becomes 0, we've reached the end of the top-level object/array.
				if depth == 0 {
					return json[:i+1]
				}
			}
		}
	}
	return json
}

// unescape takes a JSON-encoded string as input and processes any escape sequences (e.g., \n, \t, \u) within it,
// returning a new string with the escape sequences replaced by their corresponding characters.
//
// Parameters:
//   - `json`: A string representing a JSON-encoded string which may contain escape sequences (e.g., `\n`, `\"`).
//
// Returns:
//   - A new string where all escape sequences in the input JSON string are replaced by their corresponding
//     character representations. If an escape sequence is invalid or incomplete, it returns the string up to
//     that point without applying the escape.
//
// Notes:
//   - The function processes escape sequences commonly found in JSON, such as `\\`, `\/`, `\b`, `\f`, `\n`, `\r`, `\t`, `\"`, and `\u` (Unicode).
//   - If an invalid or incomplete escape sequence is encountered (for example, an incomplete Unicode sequence), it returns the string up to that point.
//   - If a non-printable character (less than ASCII value 32) is encountered, the function terminates early and returns the string up to that point.
//   - The function handles Unicode escape sequences (e.g., `\uXXXX`) by decoding them into their respective Unicode characters and converting them into UTF-8.
//
// Example Usage:
//
//	input := "\"Hello\\nWorld\""
//	result := unescape(input)
//	// result: "Hello\nWorld"
//
//	input := "\"Unicode \\u0048\\u0065\\u006C\\u006C\\u006F\""
//	result := unescape(input)
//	// result: "Unicode Hello"
func unescape(json string) string {
	var str = make([]byte, 0, len(json))
	for i := 0; i < len(json); i++ {
		switch {
		default:
			str = append(str, json[i])
		case json[i] < ' ': // If the character is a non-printable character (ASCII value less than 32), terminate early.
			return string(str)
		case json[i] == '\\': // If the current character is a backslash, process the escape sequence.
			i++
			if i >= len(json) {
				return string(str)
			}
			switch json[i] {
			default:
				return string(str)
			case '\\':
				str = append(str, '\\')
			case '/':
				str = append(str, '/')
			case 'b':
				str = append(str, '\b')
			case 'f':
				str = append(str, '\f')
			case 'n':
				str = append(str, '\n')
			case 'r':
				str = append(str, '\r')
			case 't':
				str = append(str, '\t')
			case '"':
				str = append(str, '"')
			case 'u': // Handle Unicode escape sequences (\uXXXX).
				if i+5 > len(json) {
					return string(str)
				}
				r := hexToRune(json[i+1:]) // Decode the Unicode code point (assuming `goRune` is a helper function).
				i += 5
				if utf16.IsSurrogate(r) { // Check for surrogate pairs (used for characters outside the Basic Multilingual Plane).
					// If a second surrogate is found, decode it into the correct rune.
					if len(json[i:]) >= 6 && json[i] == '\\' &&
						json[i+1] == 'u' {
						// Decode the second part of the surrogate pair.
						r = utf16.DecodeRune(r, hexToRune(json[i+2:]))
						i += 6
					}
				}
				// Allocate enough space to encode the decoded rune as UTF-8.
				str = append(str, 0, 0, 0, 0, 0, 0, 0, 0)
				// Encode the rune as UTF-8 and append it to the result slice.
				n := utf8.EncodeRune(str[len(str)-8:], r)
				str = str[:len(str)-8+n]
				i-- // Backtrack the index to account for the additional character read.
			}
		}
	}
	return string(str)
}

// hexToRune converts the first four characters of s from a hexadecimal code
// point into a rune. It is a small helper used when interpreting "\uXXXX"
// escape sequences.
//
// If s is empty, shorter than four bytes, or the substring cannot be parsed as
// valid hexadecimal, the function returns unicode.ReplacementChar (U+FFFD).
// The returned rune is produced without allocation.
//
// Parameters:
//   - s: a string expected to contain at least four hexadecimal digits
//     beginning at index 0
//
// Returns:
//   - rune: the Unicode code point represented by the first four hex digits of s,
//     or unicode.ReplacementChar on invalid or insufficient input.
//
// Limitations:
//   - Only the first four ASCII bytes are parsed; it does not handle surrogate
//     pairs or extended escape formats.
//   - No validation is performed on remaining characters in s.
//
// Examples:
//
//	// Valid hexadecimal escape
//	r := hexToRune("0041rest")
//	// r == 'A'
//
//	// Invalid hex sequence returns replacement character
//	r = hexToRune("ZZZZ")
//	// r == unicode.ReplacementChar
//
//	// Input too short
//	r = hexToRune("12")
//	// r == unicode.ReplacementChar
func hexToRune(s string) rune {
	if strutil.IsEmpty(s) || len(s) < 4 {
		return unicode.ReplacementChar
	}
	n, err := strconv.ParseUint(s[:4], 16, 32) // strconv.ParseUint(json[:4], 16, 64): Parse the first 4 characters of the input string as a 16-bit hexadecimal number.
	if err != nil {
		return unicode.ReplacementChar
	}
	return rune(n)
}

// lessFold reports whether a lexicographically precedes b under simple Unicode
// case-folding (a case-insensitive comparison).
//
// It compares byte-by-byte, folding uppercase ASCII letters to lowercase on the fly
// without allocation or full string conversion. Only ASCII letters are case-folded;
// non-ASCII characters are compared as-is.
//
// This is a lightweight alternative to strings.Compare with case-insensitivity for
// ASCII-heavy keys (e.g. HTTP headers, identifiers, simple sorting).
//
// It returns true if a < b (case-insensitive), false if a >= b.
//
// For full Unicode correctness prefer strings.EqualFold + custom less logic or
// golang.org/x/text/collate when performance is not critical.
//
// Parameters:
//   - `a`: The first string to compare.
//   - `b`: The second string to compare.
//
// Returns:
//   - `true`: If string a is lexicographically smaller than string b in a case-insensitive comparison.
//   - `false`: Otherwise.
//
// Notes:
//   - The function compares the strings character by character. If both characters are uppercase, they are compared directly.
//   - If one character is uppercase and the other is lowercase, the uppercase character is treated as the corresponding lowercase character.
//   - If neither character is uppercase, they are compared directly without any transformation.
//   - The function handles cases where the strings have different lengths and returns true if the strings are equal up to the point where the shorter string ends.
//
// Example Usage:
//
//	result := lessFold("apple", "Apple")
//	// result: false, because "apple" and "Apple" are considered equal when case is ignored
//
//	result := lessFold("apple", "banana")
//	// result: true, because "apple" is lexicographically smaller than "banana"
func lessFold(a, b string) bool {
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] >= 'A' && a[i] <= 'Z' {
			if b[i] >= 'A' && b[i] <= 'Z' {
				// both are uppercase, do nothing
				if a[i] < b[i] {
					return true
				} else if a[i] > b[i] {
					return false
				}
			} else {
				// a is uppercase, convert a to lowercase
				if a[i]+32 < b[i] {
					return true
				} else if a[i]+32 > b[i] {
					return false
				}
			}
		} else if b[i] >= 'A' && b[i] <= 'Z' {
			// b is uppercase, convert b to lowercase
			if a[i] < b[i]+32 {
				return true
			} else if a[i] > b[i]+32 {
				return false
			}
		} else {
			// neither are uppercase
			if a[i] < b[i] {
				return true
			} else if a[i] > b[i] {
				return false
			}
		}
	}
	return len(a) < len(b)
}

// expectTrue checks if the given byte slice starting at index i represents the string "true".
// It returns the next index after "true" and true if the sequence matches, otherwise it returns the current index and false.
//
// Parameters:
//   - data: A byte slice containing the data to validate.
//   - i: The index in the byte slice to start checking from.
//
// Returns:
//   - value: The index immediately after the "true" string if it matches, or the current index if it doesn't.
//   - ok: A boolean indicating whether the "true" string was found starting at index i.
//
// Notes:
//   - The function checks if the characters at positions i, i+1, and i+2 correspond to the letters 't', 'r', and 'u'.
//     If the substring matches the word "true", it returns the index after the "true" string and true.
//   - If the substring does not match "true", it returns the current index i and false.
//
// Example Usage:
//
//	data := []byte("this is true")
//	i := 10
//	value, ok := expectTrue(data, i)
//	// value: 13 (the index after the word "true")
//	// ok: true (because "true" was found starting at index 10)
func expectTrue(data []byte, i int) (newPos int, ok bool) {
	if i+3 <= len(data) && data[i] == 'r' && data[i+1] == 'u' &&
		data[i+2] == 'e' {
		return i + 3, true
	}
	return i, false
}

// expectFalse checks if the given byte slice starting at index i represents the string "false".
// It returns the next index after "false" and true if the sequence matches, otherwise it returns the current index and false.
//
// Parameters:
//   - data: A byte slice containing the data to validate.
//   - i: The index in the byte slice to start checking from.
//
// Returns:
//   - val: The index immediately after the "false" string if it matches, or the current index if it doesn't.
//   - ok: A boolean indicating whether the "false" string was found starting at index i.
//
// Notes:
//   - The function checks if the characters at positions i, i+1, i+2, and i+3 correspond to the letters 'f', 'a', 'l', and 's', respectively.
//     If the substring matches the word "false", it returns the index after the "false" string and true.
//   - If the substring does not match "false", it returns the current index i and false.
//
// Example Usage:
//
//	data := []byte("this is false")
//	i := 8
//	val, ok := expectFalse(data, i)
//	// val: 13 (the index after the word "false")
//	// ok: true (because "false" was found starting at index 8)
func expectFalse(data []byte, i int) (newPos int, ok bool) {
	if i+4 <= len(data) && data[i] == 'a' && data[i+1] == 'l' &&
		data[i+2] == 's' && data[i+3] == 'e' {
		return i + 4, true
	}
	return i, false
}

// expectNull checks if the given byte slice starting at index i represents the string "null".
// It returns the next index after "null" and true if the sequence matches, otherwise it returns the current index and false.
//
// Parameters:
//   - data: A byte slice containing the data to validate.
//   - i: The index in the byte slice to start checking from.
//
// Returns:
//   - val: The index immediately after the "null" string if it matches, or the current index if it doesn't.
//   - ok: A boolean indicating whether the "null" string was found starting at index i.
//
// Notes:
//   - The function checks if the characters at positions i, i+1, and i+2 correspond to the letters 'n', 'u', and 'l', respectively.
//     If the substring matches the word "null", it returns the index after the "null" string and true.
//   - If the substring does not match "null", it returns the current index i and false.
//
// Example Usage:
//
//	data := []byte("value is null")
//	i := 9
//	val, ok := expectNull(data, i)
//	// val: 13 (the index after the word "null")
//	// ok: true (because "null" was found starting at index 9)
func expectNull(data []byte, i int) (newPos int, ok bool) {
	if i+3 <= len(data) && data[i] == 'u' && data[i+1] == 'l' &&
		data[i+2] == 'l' {
		return i + 3, true
	}
	return i, false
}

// expectNumber validates whether the byte slice starting at index i represents a valid numeric value.
// It supports integer, floating-point, and exponential number formats as per JSON specifications.
//
// Parameters:
//   - data: A byte slice containing the input to validate.
//   - i: The starting index to check in the byte slice.
//
// Returns:
//   - val: The index immediately after the numeric value if it is valid, or the current index if it isn't.
//   - ok: A boolean indicating whether the input from index i represents a valid numeric value.
//
// Notes:
//   - This function validates numbers following the JSON number format, which includes:
//   - Optional sign ('-' for negative numbers).
//   - Integer component (digits, starting with '0' or other digits).
//   - Optional fractional part (a dot '.' followed by one or more digits).
//   - Optional exponent part ('e' or 'E', optionally signed, followed by one or more digits).
//   - The function iterates over the byte slice, checking each part of the number sequentially.
//   - If the numeric value is valid, the function returns the index after the number and true.
//     Otherwise, it returns the starting index and false.
//
// Example Usage:
//
//	data := []byte("-123.45e+6")
//	i := 1 // Start after the '-' sign
//	val, ok := expectNumber(data, i)
//	// val: 10 (the index after the number)
//	// ok: true (because "-123.45e+6" is a valid numeric value)
//
// Details:
//
//   - The function handles three major components of a number: sign, integer part, and optional components
//     (fractional and exponential parts).
//   - Each component is validated, and the function exits early with a false result if a part is invalid.
func expectNumber(data []byte, i int) (newPos int, ok bool) {
	// Check if i is within valid range
	if i <= 0 || i >= len(data) {
		return i, false
	}
	i--
	// Check for a sign ('-') at the start of the number.
	if data[i] == '-' {
		i++
		// A sign without any digits is invalid.
		// The character after the sign must be a digit.
		if i == len(data) || data[i] < '0' || data[i] > '9' {
			return i, false
		}
	}
	// Validate the integer part of the number.
	if i == len(data) {
		return i, false
	}
	if data[i] == '0' {
		// A leading '0' is valid but must not be followed by other digits.
		i++
	} else {
		// Consume digits in the integer part.
		for ; i < len(data); i++ {
			if data[i] >= '0' && data[i] <= '9' {
				continue
			}
			break
		}
	}
	// Validate the fractional part, if present.
	if i == len(data) {
		return i, true
	}
	if data[i] == '.' {
		i++
		// A dot without digits after it is invalid.
		// The character after the dot must be a digit.
		if i == len(data) || data[i] < '0' || data[i] > '9' {
			return i, false
		}
		i++
		// Consume digits in the fractional part.
		for ; i < len(data); i++ {
			if data[i] >= '0' && data[i] <= '9' {
				continue
			}
			break
		}
	}
	// Validate the exponential part, if present.
	if i == len(data) {
		return i, true
	}
	if data[i] == 'e' || data[i] == 'E' {
		i++
		if i == len(data) {
			// An 'e' or 'E' without any exponent value is invalid.
			return i, false
		}
		// Check for an optional sign in the exponent.
		if data[i] == '+' || data[i] == '-' {
			i++
		}
		// A sign without any digits in the exponent is invalid.
		// The character after the exponent must be a digit.
		if i == len(data) || data[i] < '0' || data[i] > '9' {
			return i, false
		}
		i++
		// Consume digits in the exponent part.
		for ; i < len(data); i++ {
			if data[i] >= '0' && data[i] <= '9' {
				continue
			}
			break
		}
	}
	return i, true
}

// expectString validates whether the byte slice starting at index i represents a valid JSON string.
// The function ensures the string adheres to the JSON string format, including proper escaping of special characters.
//
// Parameters:
//   - data: A byte slice containing the input to validate.
//   - i: The starting index to check in the byte slice.
//
// Returns:
//   - val: The index immediately after the string if it is valid, or the current index if it isn't.
//   - ok: A boolean indicating whether the input from index i represents a valid JSON string.
//
// Notes:
//   - JSON strings must start and end with double quotes ('"').
//   - The function handles escaped characters such as '\\', '\"', and unicode escapes (e.g., '\\u1234').
//   - The function iterates over the byte slice, validating each character and ensuring proper escape sequences.
//   - If the string is valid, the function returns the index after the closing double quote and true.
//     Otherwise, it returns the current index and false.
//
// Example Usage:
//
//	data := []byte("\"Hello, \\\"world!\\\"\"")
//	i := 0 // Start at the first character
//	val, ok := expectString(data, i)
//	// val: 20 (the index after the string)
//	// ok: true (because "\"Hello, \\\"world!\\\"\"" is a valid JSON string)
//
// Details:
//   - The function iterates over the characters, checking for valid JSON string content.
//   - It handles special escape sequences, ensuring their correctness.
//   - Early exit occurs if any invalid sequence or character is detected.
func expectString(data []byte, i int) (newPos int, ok bool) {
	for ; i < len(data); i++ {
		if data[i] < ' ' {
			return i, false
		} else if data[i] == '\\' {
			i++
			if i == len(data) {
				return i, false
			}
			switch data[i] {
			default:
				return i, false
			case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
			case 'u':
				for j := 0; j < 4; j++ {
					i++
					if i >= len(data) {
						return i, false
					}
					if !((data[i] >= '0' && data[i] <= '9') ||
						(data[i] >= 'a' && data[i] <= 'f') ||
						(data[i] >= 'A' && data[i] <= 'F')) {
						return i, false
					}
				}
			}
		} else if data[i] == '"' {
			return i + 1, true
		}
	}
	return i, false
}

// expectCommaOrEnd checks for the presence of a comma (',') or the specified end character in the given byte slice
// starting at index i. It skips over any whitespace characters and ensures valid structure.
//
// Parameters:
//   - data: A byte slice containing the input to validate.
//   - i: The starting index to check in the byte slice.
//   - end: The specific byte (character) to treat as a valid stopping point, in addition to a comma.
//
// Returns:
//   - val: The index of the comma or end character if found, or the current index if invalid.
//   - ok: A boolean indicating whether a valid comma or end character was found.
//
// Notes:
//   - Whitespace characters (' ', '\t', '\n', '\r') are skipped during validation.
//   - The function exits early with false if an invalid character is encountered before finding a comma or end character.
//   - If the comma or end character is found, the function returns its index and true.
//
// Example Usage:
//
//	data := []byte(" , next")
//	i := 0
//	end := byte('n')
//	val, ok := expectCommaOrEnd(data, i, end)
//	// val: 1 (the index of the comma)
//	// ok: true (because a comma was found)
//
// Details:
//   - Iterates over characters, skipping valid whitespace.
//   - Checks for either a comma or the specified end character.
//   - Returns false if an invalid character is encountered.
func expectCommaOrEnd(data []byte, i int, end byte) (newPos int, ok bool) {
	for ; i < len(data); i++ {
		switch data[i] {
		default:
			return i, false
		case ' ', '\t', '\n', '\r':
			continue
		case ',':
			return i, true
		case end:
			return i, true
		}
	}
	return i, false
}

// expectColon checks for the presence of a colon (':') in the given byte slice starting at index i.
// It skips over any whitespace characters and ensures valid JSON structure.
//
// Parameters:
//   - data: A byte slice containing the input to validate.
//   - i: The starting index to check in the byte slice.
//
// Returns:
//   - val: The index immediately after the colon if found, or the current index if invalid.
//   - ok: A boolean indicating whether a valid colon was found.
//
// Notes:
//   - Whitespace characters (' ', '\t', '\n', '\r') are skipped during validation.
//   - The function exits early with false if an invalid character is encountered before finding the colon.
//   - If the colon is found, the function returns the index after it and true.
//
// Example Usage:
//
//	data := []byte(" : value")
//	i := 0
//	val, ok := expectColon(data, i)
//	// val: 2 (the index after the colon)
//	// ok: true (because a colon was found)
//
// Details:
//   - Iterates over characters, skipping valid whitespace.
//   - Checks for a colon and returns the next index upon finding it.
//   - Returns false if an invalid character is encountered.
func expectColon(data []byte, i int) (newPos int, ok bool) {
	for ; i < len(data); i++ {
		switch data[i] {
		default:
			return i, false
		case ' ', '\t', '\n', '\r':
			continue
		case ':':
			return i + 1, true
		}
	}
	return i, false
}

// expectArray validates whether the byte slice starting at index i represents a valid JSON array.
// It ensures that the array starts and ends with square brackets ('[' and ']') and contains valid JSON values.
//
// Parameters:
//   - data: A byte slice containing the input to validate.
//   - i: The starting index to check in the byte slice.
//
// Returns:
//   - val: The index immediately after the valid array if found, or the current index if it isn't.
//   - ok: A boolean indicating whether the input from index i represents a valid JSON array.
//
// Notes:
//   - The function handles arrays that may contain:
//   - Whitespace (skipped).
//   - Comma-separated JSON values, validated using the `validateAny` function.
//   - Empty arrays ([]).
//   - The function ensures that the array ends with a closing square bracket (']').
//
// Example Usage:
//
//	data := []byte("[123, \"string\", false]")
//	i := 0
//	val, ok := expectArray(data, i)
//	// val: 21 (the index after the array)
//	// ok: true (because the input is a valid JSON array)
//
// Details:
//   - Skips leading whitespace.
//   - Checks for an initial ']' to handle empty arrays.
//   - Iteratively validates JSON values and ensures proper use of commas.
//   - Returns false if an invalid character or structure is encountered.
func expectArray(data []byte, i int) (newPos int, ok bool) {
	for ; i < len(data); i++ {
		switch data[i] {
		default:
			for ; i < len(data); i++ {
				if i, ok = expectValue(data, i); !ok {
					return i, false
				}
				if i, ok = expectCommaOrEnd(data, i, ']'); !ok {
					return i, false
				}
				if data[i] == ']' {
					return i + 1, true
				}
			}
		case ' ', '\t', '\n', '\r':
			continue
		case ']':
			return i + 1, true
		}
	}
	return i, false
}

// expectObject validates whether the byte slice starting at index i represents a valid JSON object.
// It ensures that the object starts and ends with curly braces ('{' and '}') and contains valid key-value pairs.
//
// Parameters:
//   - data: A byte slice containing the input to validate.
//   - i: The starting index to check in the byte slice.
//
// Returns:
//   - val: The index immediately after the valid object if found, or the current index if it isn't.
//   - ok: A boolean indicating whether the input from index i represents a valid JSON object.
//
// Notes:
//   - The function handles objects that may contain:
//   - Whitespace (skipped).
//   - Key-value pairs, where keys are JSON strings and values are validated using `validateAny`.
//   - Empty objects ({}).
//   - Ensures proper use of colons (:) and commas (,) in separating keys and values.
//   - The function iteratively validates keys and values until the closing curly brace ('}') is found.
//
// Example Usage:
//
//	data := []byte(`{"key1": 123, "key2": "value"}`)
//	i := 0
//	val, ok := expectObject(data, i)
//	// val: 28 (the index after the object)
//	// ok: true (because the input is a valid JSON object)
//
// Details:
//   - Skips leading whitespace.
//   - Validates the presence of keys (JSON strings) and their corresponding values.
//   - Ensures that the structure adheres to JSON object syntax, returning false for any invalid structure.
func expectObject(data []byte, i int) (newPos int, ok bool) {
	for ; i < len(data); i++ {
		switch data[i] {
		default:
			return i, false
		case ' ', '\t', '\n', '\r':
			continue
		case '}':
			return i + 1, true
		case '"':
		key:
			if i, ok = expectString(data, i+1); !ok {
				return i, false
			}
			if i, ok = expectColon(data, i); !ok {
				return i, false
			}
			if i, ok = expectValue(data, i); !ok {
				return i, false
			}
			if i, ok = expectCommaOrEnd(data, i, '}'); !ok {
				return i, false
			}
			if data[i] == '}' {
				return i + 1, true
			}
			i++
			for ; i < len(data); i++ {
				switch data[i] {
				default:
					return i, false
				case ' ', '\t', '\n', '\r':
					continue
				case '"':
					goto key
				}
			}
			return i, false
		}
	}
	return i, false
}

// expectValue attempts to validate the data starting at index i as one of the possible JSON value types.
// It recognizes and validates the following JSON types:
//   - Object: represented by curly braces `{}`
//   - Array: represented by square brackets `[]`
//   - String: represented by double quotes `""`
//   - Numeric values: including integers and floating-point numbers
//   - Boolean values: `true` or `false`
//   - Null: represented by `null`
//
// Parameters:
//   - data: A byte slice containing the JSON input to validate.
//   - i: The starting index in the byte slice where the validation should begin.
//
// Returns:
//   - val: The index immediately after the valid value if found, or the current index if it isn't.
//   - ok: A boolean indicating whether the input starting at index i is a valid JSON value of one of the recognized types.
//
// Notes:
//   - The function handles a variety of JSON data types, attempting to validate the input by matching
//     the character at the current index and calling the appropriate validation function for the recognized type.
//   - It will skip any whitespace characters (spaces, tabs, newlines, carriage returns) before checking the data.
//   - The function calls other helper functions to validate specific types of JSON values, such as `verifyObject`, `verifyArray`,
//     `verifyString`, `verifyNumeric`, `verifyBoolTrue`, `verifyBoolFalse`, and `verifyNullable`.
//
// Example Usage:
//
//	data := []byte(`{"key1": 123, "key2": "value"}`)
//	i := 0
//	val, ok := expectValue(data, i)
//	// val: 28 (the index after the object)
//	// ok: true (because the input is a valid JSON object)
//
// Details:
//   - If the input data at index i is a valid JSON value (object, array, string, numeric, boolean, or null),
//     the function will return the index immediately after the valid value and true.
//   - If the data does not match a valid JSON value, it returns the current index and false.
func expectValue(data []byte, i int) (newPos int, ok bool) {
	for ; i < len(data); i++ {
		switch data[i] {
		default:
			return i, false
		case ' ', '\t', '\n', '\r':
			continue
		case '{':
			return expectObject(data, i+1)
		case '[':
			return expectArray(data, i+1)
		case '"':
			return expectString(data, i+1)
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			return expectNumber(data, i+1)
		case 't':
			return expectTrue(data, i+1)
		case 'f':
			return expectFalse(data, i+1)
		case 'n':
			return expectNull(data, i+1)
		}
	}
	return i, false
}

// expectJSON attempts to validate the data starting at index i as a valid JSON payload. It checks for
// the presence of valid JSON values after skipping any whitespace characters (spaces, tabs, newlines, etc.).
// It calls the `verifyAny` function to validate the first JSON value and ensures that there are no unexpected
// characters after the valid value.
//
// Parameters:
//   - data: A byte slice containing the JSON input to validate.
//   - i: The starting index in the byte slice where the validation should begin.
//
// Returns:
//   - val: The index immediately after the validated JSON payload, or the current index if validation fails.
//   - ok: A boolean indicating whether the payload is valid. If true, the input is a valid JSON payload;
//     if false, the payload is not valid.
//
// Notes:
//   - The function starts by checking the first non-whitespace character in the data. It skips over any
//     whitespace (spaces, tabs, newlines, carriage returns) before trying to validate the payload.
//   - It then calls the `verifyAny` function to check for a valid JSON value (object, array, string, numeric,
//     boolean, or null) at the current index.
//   - After the first valid value is found, the function ensures that the rest of the input contains only
//     whitespace (ignoring spaces, tabs, newlines) before validating that the payload ends correctly.
//   - If the first value is valid and no unexpected characters are found, it returns true along with the
//     index after the valid value. If any issue is encountered, it returns false.
//
// Example Usage:
//
//	data := []byte(`{"key1": 123, "key2": "value"}`)
//	i := 0
//	val, ok := expectJSON(data, i)
//	// val: 28 (index after the valid payload)
//	// ok: true (because the input is a valid JSON payload)
//
// Details:
//   - The function ensures that the input data starts with a valid JSON value, as recognized by the `verifyAny`
//     function, and that there are no unexpected characters after that value.
//   - The function returns false if the JSON payload is incomplete or invalid.
func expectJSON(data []byte, i int) (newPos int, ok bool) {
	for ; i < len(data); i++ { // Iterate through the data starting from index i.
		// Handle unexpected characters in the payload.
		switch data[i] {
		default:
			// If the character is unexpected, call verifyAny to validate the first valid JSON value.
			i, ok = expectValue(data, i)
			if !ok {
				return i, false // Return false if the value is not valid.
			}
			// After a valid JSON value, continue to check if the rest of the data only contains whitespace.
			for ; i < len(data); i++ {
				switch data[i] {
				default:
					return i, false // Return false if there is any invalid character after the value.
				case ' ', '\t', '\n', '\r': // Skip over whitespace characters.
					continue
				}
			}
			// If all subsequent characters are whitespace, return true along with the index after the valid value.
			return i, true
		case ' ', '\t', '\n', '\r': // Skip over whitespace characters (spaces, tabs, newlines, carriage returns).
			continue
		}
	}
	return i, false // Return false if the end of data is reached without a valid payload.
}

// extractJSONValue processes a JSON string starting from a given index `i`, squashing (flattening) any nested JSON structures
// (such as arrays, objects, or even parentheses) into a single value. The function handles strings, nested objects,
// arrays, and parentheses while ignoring the nested structures themselves, only returning the top-level JSON structure
// from the starting point.
//
// Parameters:
//   - `json`: A string representing the JSON data to be parsed. This string can include various JSON constructs like
//     strings, objects, arrays, and nested structures within parentheses.
//   - `i`: The index in the `json` string from which parsing should begin. The function assumes that the character
//     at this index is the opening character of a JSON array ('['), object ('{'), or parentheses ('(').
//
// Returns:
//   - `int`: The new index after the parsing of the JSON structure, which is after the closing bracket/parenthesis/brace.
//   - `string`: A string containing the flattened JSON structure starting from the opening character and squashing all
//     nested structures until the corresponding closing character is reached.
//
// Example Usage:
//
//	json := "[{ \"key\": \"value\" }, { \"nested\": [1, 2, 3] }]"
//	i, result := extractJSONValue(json, 0)
//	// result: "{ \"key\": \"value\" }, { \"nested\": [1, 2, 3] }" (flattened to top-level content)
//	// i: the index after the closing ']' of the outer array
//
// Details:
//   - The function expects that the character at index `i` is an opening character for an array, object, or parentheses,
//     and it will proceed to skip over any nested structures of the same type (i.e., arrays, objects, or parentheses).
//   - The depth of nesting is tracked, and whenever the function encounters a closing bracket (']'), brace ('}'), or parenthesis
//     (')'), it checks if the depth has returned to 0 (indicating the end of the top-level structure).
//   - If a string is encountered (enclosed in double quotes), it processes the string contents carefully, respecting escape sequences.
//   - The function ensures that nested structures (arrays, objects, or parentheses) are ignored, effectively "squashing" the
//     content into the outermost structure, while the depth ensures that only the highest-level structure is returned.
func extractJSONValue(json string, i int) (newPos int, value string) {
	if strutil.IsEmpty(json) || i < 0 {
		return i, ""
	}
	s := i
	i++
	depth := 1
	for ; i < len(json); i++ {
		if json[i] >= '"' && json[i] <= '}' {
			switch json[i] {
			case '"':
				i++
				s2 := i
				for ; i < len(json); i++ {
					if json[i] > '\\' {
						continue
					}
					if json[i] == '"' {
						// look for an escaped slash
						if json[i-1] == '\\' {
							n := 0
							for j := i - 2; j > s2-1; j-- {
								if json[j] != '\\' {
									break
								}
								n++
							}
							if n%2 == 0 {
								continue
							}
						}
						break
					}
				}
			case '{', '[', '(':
				depth++
			case '}', ']', ')':
				depth--
				if depth == 0 {
					i++
					return i, json[s:i]
				}
			}
		}
	}
	return i, json[s:]
}

// extractJSONString scans json starting at index i (immediately after an opening
// double quote at json[i-1]) and returns the position just past the matching
// closing quote, the quoted substring including both quotes, whether any escape
// sequences were encountered, and whether a complete closing quote was found.
//
// This helper performs lightweight, byte-wise scanning suitable for JSON-style
// string literals. It recognizes escaped quotes by counting preceding backslashes
// so that \" inside the string does not terminate it. The function does not
// perform full JSON validation or unescaping; it merely locates the terminating
// quote boundary.
//
// Parameters:
//   - json: the source text containing a JSON string literal.
//   - i: starting index, which must point to the first byte *after* the opening
//     quote (i.e., json[i-1] == '"'). If json is empty or i < 0, the function
//     returns i, json, false, false.
//
// Returns:
//   - newPos: index of the first byte after the closing quote when complete is true;
//     otherwise the index where scanning stopped (typically len(json)).
//   - quoted: the slice of json that spans the string literal including its
//     surrounding quotes. On incomplete input, this is json[s-1:] (the remainder
//     starting at the opening quote).
//   - hadEscape: true if an escape sequence was observed while scanning.
//   - complete: true if a matching, unescaped closing quote was found.
//
// Assumptions & Invariants:
//   - The caller is responsible for providing i such that i > 0 and json[i-1] == '"';
//     otherwise results may be meaningless.
//   - Escaped quotes are detected by counting consecutive backslashes immediately
//     preceding a quote; an even count means the quote is escaped and scanning continues.
//   - Scanning is byte-oriented (ASCII for structure). It does not interpret Unicode
//     escapes, surrogate pairs, or other JSON token types.
//
// Limitations & Nuances:
//   - This is a structural scanner, not a full JSON parser or unescaper.
//   - On incomplete input (no closing quote), complete is false and quoted contains
//     the remainder of json from the opening quote to the end.
//   - No errors are returned; callers should inspect complete to determine success.
//   - Substrings share the backing array with json (zero-copy). Copy if you need
//     to retain quoted independently of json.
//
// Performance:
//   - Runs in O(n) time over the scanned region and performs no allocations.
//
// Examples:
//
//	// Typical well-formed JSON string
//	pos, quoted, hadEsc, ok := extractJSONString(`"hello" trailing`, 1)
//	// pos == 7, quoted == `"hello"`, hadEsc == false, ok == true
//
//	// String containing an escaped quote
//	pos, quoted, hadEsc, ok = extractJSONString(`"he said \"hi\"" rest`, 1)
//	// pos == 16, quoted == `"he said \"hi\""`, hadEsc == true, ok == true
//
//	// Unterminated string: no closing quote found
//	pos, quoted, hadEsc, ok = extractJSONString(`"unterminated`, 1)
//	// ok == false, quoted == `"unterminated`, pos == len(`"unterminated`)
func extractJSONString(json string, i int) (newPos int, quoted string, hadEscape bool, complete bool) {
	if strutil.IsEmpty(json) || i < 0 {
		return i, json, false, false
	}
	var s = i
	for ; i < len(json); i++ {
		if json[i] > '\\' {
			continue
		}
		if json[i] == '"' {
			return i + 1, json[s-1 : i+1], false, true
		}
		if json[i] == '\\' {
			i++
			for ; i < len(json); i++ {
				if json[i] > '\\' {
					continue
				}
				if json[i] == '"' {
					// look for an escaped slash
					if json[i-1] == '\\' {
						n := 0
						for j := i - 2; j > 0; j-- {
							if json[j] != '\\' {
								break
							}
							n++
						}
						if n%2 == 0 {
							continue
						}
					}
					return i + 1, json[s-1 : i+1], true, true
				}
			}
			break
		}
	}
	return i, json[s-1:], false, false
}

// extractJSONNumber scans json starting at index i (which should point to the
// first byte of a JSON number) and returns the index just past the number along
// with the substring that forms that number.
//
// The scan proceeds byte-by-byte and stops at the first structural or
// terminating delimiter: whitespace, ',', ']', or '}'. It does not perform full
// JSON validation (e.g., it does not enforce RFC 7159 number grammar, handle
// leading '+' signs, or normalize exponents); it simply slices out the maximal
// run that appears to be a number until a delimiter is reached.
//
// Parameters:
//   - json: the source text containing a JSON value.
//   - i: the starting index; callers are expected to position i at the first
//     byte of the numeric token. If json is empty or i < 0, the function
//     returns i and json unchanged.
//
// Returns:
//   - newPos: the index of the first byte after the parsed number (or the point
//     where scanning stopped).
//   - number: a substring of json that shares the original backing array
//     (zero-copy). Copy it if you need to retain it independently.
//
// Assumptions & Invariants:
//   - Scanning is ASCII/byte-oriented and heuristic; it does not validate that
//     the substring is a well-formed JSON number.
//   - The caller is responsible for ensuring i is within [0, len(json)) and
//     points to the beginning of a number. If not, results may be unintended.
//
// Limitations & Nuances:
//   - No error is reported; callers should validate the returned substring if
//     strict number conformance is required.
//   - The function does not skip leading whitespace; callers should advance i to
//     the first non-space if necessary.
//   - If the end of json is reached without encountering a delimiter, number
//     extends to the end.
//
// Performance:
//   - O(n) over the scanned suffix; no allocations in the common case.
//
// Examples:
//
//	// Basic integer
//	pos, num := extractJSONNumber("123, rest", 0)
//	// pos == 3, num == "123"
//
//	// Decimal with exponent, followed by array delimiter
//	pos, num = extractJSONNumber("-1.23e+10]tail", 0)
//	// pos == len("-1.23e+10"), num == "-1.23e+10"
//
//	// Stops at whitespace
//	pos, num = extractJSONNumber("42  ,next", 0)
//	// pos == 2, num == "42"
func extractJSONNumber(json string, i int) (newPos int, number string) {
	if strutil.IsEmpty(json) || i < 0 {
		return i, json
	}
	var s = i
	i++
	for ; i < len(json); i++ {
		if json[i] <= ' ' || json[i] == ',' || json[i] == ']' ||
			json[i] == '}' {
			return i, json[s:i]
		}
	}
	return i, json[s:]
}

// extractJSONLowerLiteral scans json beginning at index i—expected to point to
// the first byte of a lowercase JSON literal (such as true, false, or null in
// lowercase form)—and returns the position just after the contiguous run of
// lowercase ASCII letters along with the substring that forms that literal.
//
// The scan is purely lexical. It does not validate that the substring
// corresponds to a valid JSON literal; it only consumes consecutive bytes in
// the range 'a'–'z'. Callers must validate the returned literal if correctness
// is required.
//
// Parameters:
//   - json: the source text containing a JSON value or literal.
//   - i: index of the first lowercase rune to scan. If json is empty or i < 0,
//     the function returns i and json unchanged.
//
// Returns:
//   - newPos: index immediately following the last consumed lowercase letter.
//   - literal: the zero‑copy substring of json containing the scanned run.
//     Copy it if you need to retain it independently.
//
// Assumptions & Invariants:
//   - The caller positions i at the first character of the literal. If i does
//     not point to a lowercase ASCII letter, the returned slice may be empty or
//     unintended.
//   - Only ASCII lowercase a–z are consumed; no attempt is made to interpret
//     Unicode letters or validate against JSON’s allowed literal set.
//
// Limitations & Nuances:
//   - This function does not confirm whether the extracted text is "true",
//     "false", "null", or any other valid JSON literal.
//   - No error value is returned; callers must rely on semantic validation
//     after extraction.
//   - If the end of json is reached with no delimiter encountered, literal
//     extends to the end.
//
// Performance:
//   - Runs in O(n) over the scanned suffix and performs no allocations.
//
// Examples:
//
//	// Basic literal
//	pos, lit := extractJSONLowerLiteral("true,", 0)
//	// pos == 4, lit == "true"
//
//	// Mixed characters: stops at non-lowercase
//	pos, lit = extractJSONLowerLiteral("falseXrest", 0)
//	// pos == 5, lit == "false"
//
//	// Input too short or invalid start index
//	pos, lit = extractJSONLowerLiteral("", 0)
//	// pos == 0, lit == ""
func extractJSONLowerLiteral(json string, i int) (newPos int, literal string) {
	if strutil.IsEmpty(json) || i < 0 {
		return i, json
	}
	var s = i
	i++
	for ; i < len(json); i++ {
		if json[i] < 'a' || json[i] > 'z' {
			return i, json[s:i]
		}
	}
	return i, json[s:]
}

// extractJSONValueAt parses the next JSON value from a given JSON string starting at the specified index `i`.
// The function identifies and processes a variety of JSON value types including objects, arrays, strings, literals (true, false, null),
// and numeric values. The result of parsing is returned as a `Context` containing relevant information about the parsed value.
//
// Parameters:
//   - `json`: A string representing the JSON data to be parsed. This string can include objects, arrays, strings, literals, and numbers.
//   - `i`: The starting index in the `json` string where parsing should begin. The function will parse the value starting at this index.
//   - `hit`: A boolean flag indicating whether to capture the parsed result into the `Context` object. If true, the context will be populated with the parsed value.
//
// Returns:
//   - `i`: The updated index after parsing the JSON value. This is the index immediately after the parsed value.
//   - `ctx`: A `Context` object containing information about the parsed value, including the type (`kind`), the raw unprocessed string (`raw`),
//     and for strings or numbers, the parsed value (e.g., `str` for strings, `num` for numbers).
//   - `ok`: A boolean indicating whether the parsing was successful. If the function successfully identifies a JSON value, it returns true; otherwise, false.
//
// Example Usage:
//
//	json := `{"key": "value", "age": 25}`
//	i := 0
//	hit := true
//	i, ctx, ok := extractJSONValueAt(json, i, hit)
//	// i: the index after parsing the first JSON value (e.g., after the closing quote of "value").
//	// ctx: contains the parsed context information (e.g., for strings, the kind would be String, unprocessed would be the raw value, etc.)
//	// ok: true if the value was successfully parsed.
//
// Details:
//   - The function processes various JSON types, including objects, arrays, strings, literals (true, false, null), and numeric values.
//   - It recognizes objects (`{}`), arrays (`[]`), and string literals (`""`), and calls the appropriate helper functions for each type.
//   - When parsing a string, it handles escape sequences, and when parsing numeric values, it checks for valid numbers (including integers, floats, and special numeric literals like `NaN`).
//   - For literals like `true`, `false`, and `null`, the function parses the exact keywords and stores them in the `Context` object as `True`, `False`, or `Null` respectively.
//
// The function ensures flexibility by checking each character in the JSON string and delegating to specialized functions for handling different value types.
// If no valid JSON value is found at the given position, it returns false.
func extractJSONValueAt(json string, i int, hit bool) (newPos int, _ctx Context, _ok bool) {
	var ctx Context
	var val string
	for ; i < len(json); i++ {
		if json[i] == '{' || json[i] == '[' {
			i, val = extractJSONValue(json, i)
			if hit {
				ctx.raw = val
				ctx.kind = JSON
			}
			var tmp parser
			tmp.val = ctx
			computeOffset(json, &tmp)
			return i, tmp.val, true
		}
		if json[i] <= ' ' {
			continue
		}
		var num bool
		switch json[i] {
		case '"':
			i++
			var escVal bool
			var ok bool
			i, val, escVal, ok = extractJSONString(json, i)
			if !ok {
				return i, ctx, false
			}
			if hit {
				ctx.kind = String
				ctx.raw = val
				if escVal {
					ctx.str = unescape(val[1 : len(val)-1])
				} else {
					ctx.str = val[1 : len(val)-1]
				}
			}
			return i, ctx, true
		case 'n':
			if i+1 < len(json) && json[i+1] != 'u' {
				num = true
				break
			}
			fallthrough
		case 't', 'f':
			vc := json[i]
			i, val = extractJSONLowerLiteral(json, i)
			if hit {
				ctx.raw = val
				switch vc {
				case 't':
					ctx.kind = True
				case 'f':
					ctx.kind = False
				}
				return i, ctx, true
			}
		case '+', '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
			'i', 'I', 'N':
			num = true
		}
		if num {
			i, val = extractJSONNumber(json, i)
			if hit {
				ctx.raw = val
				ctx.kind = Number
				ctx.num, _ = strconv.ParseFloat(val, 64)
			}
			return i, ctx, true
		}

	}
	return i, ctx, false
}

// extractAndUnescapeJSONString extracts a JSON-encoded string and returns both the full JSON string (with quotes) and the unescaped string content.
// The function processes the input string to handle escaped characters and returns a clean, unescaped version of the string
// as well as the portion of the JSON string that includes the enclosing quotes.
//
// Parameters:
//   - `json`: A JSON-encoded string, which is expected to start and end with a double quote (").
//     This string may contain escape sequences (e.g., `\"`, `\\`, `\n`, etc.) within the string value.
//
// Returns:
//   - `raw`: The full JSON string including the enclosing quotes (e.g., `"[Hello]"`).
//   - `unescaped`: The unescaped content of the string inside the quotes (e.g., `"Hello"` becomes `Hello` after unescaping).
//
// Example Usage:
//
//	input := "\"Hello\\nWorld\""
//	raw, str := extractAndUnescapeJSONString(input)
//	// raw: "\"Hello\\nWorld\"" (the full JSON string with quotes)
//	// str: "Hello\nWorld" (the unescaped string content)
//
//	input := "\"This is a \\\"quoted\\\" word\""
//	raw, str := extractAndUnescapeJSONString(input)
//	// raw: "\"This is a \\\"quoted\\\" word\""
//	// str: "This is a \"quoted\" word" (the unescaped string content)
//
// Details:
//
//   - The function processes the input string starting from the second character (ignoring the initial quote).
//
//   - It handles escape sequences inside the string, skipping over escaped quotes (`\"`) and other escape sequences.
//
//   - When a closing quote (`"`) is encountered, the function checks for the escape sequences to ensure the string is correctly unescaped.
//
//   - The function also checks if there are any escaped slashes (`\\`) and validates if they are part of an even or odd sequence.
//     If an escaped slash is found, it is taken into account to avoid terminating the string early.
//
//   - If the string is well-formed, the function returns the entire JSON string with quotes (`raw`) and the unescaped string (`str`).
//
//   - If an unescaped string is not found or the JSON string doesn't match expected formats, the function returns the string as is.
func extractAndUnescapeJSONString(json string) (quoted string, unescaped string) {
	for i := 1; i < len(json); i++ {
		if json[i] > '\\' {
			continue
		}
		if json[i] == '"' {
			return json[:i+1], json[1:i]
		}
		if json[i] == '\\' {
			i++
			for ; i < len(json); i++ {
				if json[i] > '\\' {
					continue
				}
				if json[i] == '"' {
					// look for an escaped slash
					if json[i-1] == '\\' {
						n := 0
						for j := i - 2; j > 0; j-- {
							if json[j] != '\\' {
								break
							}
							n++
						}
						if n%2 == 0 {
							continue
						}
					}
					return json[:i+1], unescape(json[1:i])
				}
			}
			var ret string
			if i+1 < len(json) {
				ret = json[:i+1]
			} else {
				ret = json[:i]
			}
			return ret, unescape(json[1:i])
		}
	}
	return json, json[1:]
}

// lastSegment extracts the last part of a given path string, where the path segments are separated by
// either a pipe ('|') or a dot ('.'). The function returns the substring after the last separator,
// taking escape sequences (backslashes) into account. It ensures that any escaped separator is ignored.
//
// Parameters:
//   - path: A string representing the full path, which may contain segments separated by '|' or '.'.
//
// Returns:
//   - A string representing the last segment in the path after the last occurrence of either '|' or '.'.
//     If no separator is found, it returns the entire input string.
//
// Notes:
//   - The function handles escape sequences where separators are preceded by a backslash ('\').
//   - If there is no valid separator in the string, the entire path is returned as-is.
//   - The returned substring is the part after the last separator, which could be the last portion of the path.
//
// Example Usage:
//
//	path := "foo|bar.baz.qux"
//	segment := lastSegment(path)
//	// segment: "qux" (the last segment after the last dot or pipe)
//
// Details:
//   - The function iterates from the end of the string towards the beginning, looking for the last
//     occurrence of '|' or '.' that is not preceded by a backslash.
//   - It handles edge cases where the separator is escaped or there are no separators at all.
func lastSegment(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '|' || path[i] == '.' {
			if i > 0 {
				if path[i-1] == '\\' {
					continue
				}
			}
			return path[i+1:]
		}
	}
	return path
}

// isValidName checks if a given string component is a "simple name" according to specific rules.
// A "simple name" is a string that does not contain any control characters or any of the following special characters:
// '[' , ']' , '{' , '}' , '(' , ')' , '#' , '|' , '!'. The function returns true if the string meets these criteria.
//
// Parameters:
//   - component: A string to be checked for validity as a simple name.
//
// Returns:
//   - A boolean indicating whether the input string is a valid simple name.
//   - Returns true if the string contains only printable characters and does not include any of the restricted special characters.
//   - Returns false if the string contains any control characters or restricted special characters.
//
// Notes:
//   - The function checks each character of the string to ensure it is printable and does not contain any of the restricted characters.
//   - Control characters are defined as any character with a Unicode value less than a space (' ').
//   - The function assumes that the string is not empty and contains at least one character.
//
// Example Usage:
//
//	component := "validName"
//	isValid := isValidName(component)
//	// isValid: true (the string contains only valid characters)
//
//	component = "invalid|name"
//	isValid = isValidName(component)
//	// isValid: false (the string contains an invalid character '|')
//
// Details:
//   - The function iterates through each character of the string and checks whether it is a printable character and whether it
//     is not one of the restricted special characters. If any invalid character is found, the function returns false immediately.
func isValidName(component string) bool {
	if strutil.IsEmpty(component) {
		return false
	}
	if strutil.ContainsAny(component, " ") {
		return false
	}
	for i := 0; i < len(component); i++ {
		if component[i] < ' ' {
			return false
		}
		switch component[i] {
		case '[', ']', '{', '}', '(', ')', '#', '|', '!':
			return false
		}
	}
	return true
}

// appendHex appends the hexadecimal representation of a 16-bit unsigned integer (uint16)
// to a byte slice. The integer is converted to a 4-character hexadecimal string, and each character
// is appended to the input byte slice in sequence. The function uses a pre-defined set of hexadecimal
// digits ('0'–'9' and 'a'–'f') for the conversion.
//
// Parameters:
//   - bytes: A byte slice to which the hexadecimal characters will be appended.
//   - x: A 16-bit unsigned integer to be converted to hexadecimal and appended to the byte slice.
//
// Returns:
//   - A new byte slice containing the original bytes with the appended hexadecimal digits
//     representing the 16-bit integer.
//
// Example Usage:
//
//	var result []byte
//	x := uint16(3055) // Decimal 3055 is 0x0BEF in hexadecimal
//	result = appendHex(result, x)
//	// result: []byte{'0', 'b', 'e', 'f'} (hexadecimal representation of 3055)
//
// Details:
//   - The function shifts and masks the 16-bit integer to extract each of the four hexadecimal digits.
//   - It uses the pre-defined `hexDigits` array to convert the integer's nibbles (4 bits) into their
//     corresponding hexadecimal characters.
func appendHex(bytes []byte, x uint16) []byte {
	return append(bytes,
		hexDigits[x>>12&0xF], hexDigits[x>>8&0xF],
		hexDigits[x>>4&0xF], hexDigits[x>>0&0xF],
	)
}

// parseUint64 parses a string as an unsigned integer (uint64).
// It attempts to convert the given string to a numeric value, where each character in the string
// must be a digit between '0' and '9'. If any non-digit character is encountered, the function
// returns false, indicating the string does not represent a valid unsigned integer.
//
// Parameters:
//   - s: A string representing the unsigned integer to be parsed.
//
// Returns:
//   - n: The parsed unsigned integer value (of type uint64) if the string represents a valid number.
//   - ok: A boolean indicating whether the parsing was successful. If true, the string was successfully
//     parsed into an unsigned integer; if false, the string was invalid.
//
// Example Usage:
//
//	str := "12345"
//	n, ok := parseUint64(str)
//	// n: 12345 (the parsed unsigned integer)
//	// ok: true (the string is a valid unsigned integer)
//
//	str = "12a45"
//	n, ok = parseUint64(str)
//	// n: 0 (parsing failed)
//	// ok: false (the string contains invalid characters)
//
// Details:
//   - The function iterates through each character of the string. If it encounters a digit ('0'–'9'),
//     it accumulates the corresponding integer value into the result `n`. The result is multiplied by 10
//     with each new digit to shift the previous digits left.
//   - If any non-digit character is encountered, the function returns `0` and `false`.
//   - The function assumes that the input string is non-empty and only contains valid ASCII digits if valid.
func parseUint64(s string) (n uint64, ok bool) {
	num, err := conv.Uint64(s)
	if err != nil {
		return 0, false
	}
	return num, true
}

// matchesGlob checks if a string matches a pattern with a complexity limit to
// avoid excessive computational cost, such as those from ReDos (Regular Expression Denial of Service) attacks.
//
// This function utilizes the `MatchLimit` function from `unify4g` to perform the matching, enforcing a maximum
// complexity limit of 10,000. The function aims to prevent situations where matching could lead to long or
// excessive computation, particularly when dealing with user-controlled input.
//
// Parameters:
//   - `str`: The string to match against the pattern.
//   - `pattern`: The pattern string to match, which may include wildcards or other special characters.
//
// Returns:
//   - `bool`: `true` if the `str` matches the `pattern` within the set complexity limit; otherwise `false`.
//
// Example:
//
//	result := matchesGlob("hello", "h*o") // Returns `true` if the pattern matches the string within the complexity limit.
func matchesGlob(str, pattern string) bool {
	matched, _ := match.MatchLimit(str, pattern, 10000)
	return matched
}

// splitStaticAndLiteral parses a string path to find a static value, such as a boolean, null, or number.
// The function expects that the input path starts with a '!', indicating a static value. It identifies the static
// value by looking for valid characters and structures that represent literal values in a path. If a valid static value
// is found, it returns the remaining path, the static value, and a success flag. Otherwise, it returns false.
//
// Parameters:
//   - path: A string representing the path to parse. The path should start with a '!' to indicate a static value.
//     The function processes the string following the '!' to find the static value.
//
// Returns:
//   - pathOut: The remaining part of the path after the static value has been identified. This is the portion of the
//     string that follows the literal value, such as any further path segments or operators.
//   - result: The static value found in the path, which can be a boolean ("true" or "false"), null, NaN, or Inf, or
//     a numeric value, or an empty string if no valid static value is found.
//   - ok: A boolean indicating whether the function successfully identified a static value. Returns true if a valid
//     static value is found, and false otherwise.
//
// Example Usage:
//
//	path := "!true.some.other.path"
//	pathOut, result, ok := splitStaticAndLiteral(path)
//	// pathOut: ".some.other.path" (remaining path)
//	// result: "true" (the static value found)
//	// ok: true (successful identification of static value)
//
//	path = "!123.abc"
//	pathOut, result, ok = splitStaticAndLiteral(path)
//	// pathOut: ".abc" (remaining path)
//	// result: "123" (the static value found)
//	// ok: true (successful identification of static value)
//
// Details:
//
//   - The function looks for the first character after the '!' to determine if the value starts with a valid static
//     value, such as a number or a boolean literal ("true", "false"), null, NaN, or Inf.
//
//   - It processes the string to extract the static value, then identifies the rest of the path (if any) after the static
//     value, which is returned as the remaining portion of the path.
//
//   - If the function encounters a delimiter like '.' or '|', it stops further parsing of the static value and returns
//     the remaining path.
//
//   - If no static value is identified, the function returns false.
//
//     Notes:
//
//   - The function assumes that the input path is well-formed and follows the expected format (starting with '!').
//
//   - The value can be a boolean, null, NaN, Inf, or a number in the path.
func splitStaticAndLiteral(path string) (staticPrefix, literalValue string, ok bool) {
	name := path[1:]
	if len(name) > 0 {
		switch name[0] {
		case '{', '[', '"', '+', '-', '0', '1', '2', '3', '4', '5', '6', '7',
			'8', '9':
			_, literalValue = extractJSONValue(name, 0)
			staticPrefix = name[len(literalValue):]
			return staticPrefix, literalValue, true
		}
	}
	for i := 1; i < len(path); i++ {
		if path[i] == '|' {
			staticPrefix = path[i:]
			name = path[1:i]
			break
		}
		if path[i] == '.' {
			staticPrefix = path[i:]
			name = path[1:i]
			break
		}
	}
	switch strings.ToLower(name) {
	case "true", "false", "null", "nan", "inf":
		return staticPrefix, name, true
	}
	return staticPrefix, literalValue, false
}

// splitAtUnescapedPipe scans path and splits it at the first unescaped '|'
// that is not inside a JSON object literal or a balanced selector. It returns
// the left and right substrings around the pipe and a boolean indicating
// whether such a split point was found.
//
// Behavior overview:
//   - If path begins with '{', the function treats the leading bytes as a JSON
//     object literal and skips it (using compact semantics). If the byte
//     immediately following that object is '|', the split occurs there.
//   - Otherwise, the function walks path byte-by-byte, honoring backslash
//     escapes. It also recognizes selector segments starting with ".#[" or ".#("
//     and balances the corresponding brackets/parentheses, including quoted
//     strings within selectors, so a '|' inside a selector or inside quotes does
//     not trigger a split.
//   - An unescaped '|' found outside of those protected regions triggers a split.
//
// Parameters:
//   - path: the input to scan. Backslashes escape the following byte. A leading
//     '{' denotes a JSON object literal that will be skipped as a whole for the
//     purpose of finding the split.
//
// Returns:
//   - left, right: the substrings on each side of the first valid split point,
//     excluding the '|' itself.
//   - ok: true if a split point was found; otherwise left, right are empty and
//     ok is false.
//
// Assumptions & Invariants:
//   - The function performs byte-wise scanning; it assumes ASCII for structural
//     characters ('\\', '|', '.', '#', '[', ']', '(', ')', '"', '{', '}').
//   - A backslash escapes only the next byte; it does not implement full JSON
//     or string unescaping semantics beyond skipping escaped delimiters.
//   - When path starts with '{', only the leading JSON object is considered;
//     malformed or incomplete objects result in no split.
//   - No state is kept between calls; there are no side effects.
//
// Limitations & Nuances:
//   - This is a structural/heuristic scanner, not a full parser. It balances
//     only brackets/parentheses introduced by a ".#[" or ".#(" selector and
//     double-quoted strings within that selector.
//   - Pipes inside balanced selectors or inside the initial JSON object are
//     ignored for splitting.
//   - If no valid '|' is found, ok is false.
//   - Substrings share the backing array with path (zero-copy). Copy if you need
//     to retain them independently of the original string.
//
// Performance:
//   - Runs in O(n) time over path with no allocations in the common case.
//
// Examples:
//
//	// Simple split on first unescaped pipe
//	left, right, ok := splitAtUnescapedPipe(`foo|bar.baz`)
//	// left == "foo", right == "bar.baz", ok == true
//
//	// Escaped pipe is ignored
//	left, right, ok = splitAtUnescapedPipe(`foo\|bar|rest`)
//	// left == `foo\|bar`, right == "rest", ok == true
//
//	// Pipe after a leading JSON object
//	left, right, ok = splitAtUnescapedPipe(`{"a":[1,2,{"b":"x|y"}]}|tail`)
//	// left == `{"a":[1,2,{"b":"x|y"}]}`, right == "tail", ok == true
//
//	// Pipe inside a balanced selector is ignored
//	left, right, ok = splitAtUnescapedPipe(`root.#["a|b"](nested)|after`)
//	// left == `root.#["a|b"](nested)`, right == "after", ok == true
func splitAtUnescapedPipe(path string) (left, right string, ok bool) {
	var possible bool
	for i := 0; i < len(path); i++ {
		if path[i] == '|' {
			possible = true
			break
		}
	}
	if !possible {
		return
	}
	if len(path) > 0 && path[0] == '{' {
		squashed := compactJSON(path[1:])
		if len(squashed) < len(path)-1 {
			squashed = path[:len(squashed)+1]
			remain := path[len(squashed):]
			if remain[0] == '|' {
				return squashed, remain[1:], true
			}
		}
		return
	}
	for i := 0; i < len(path); i++ {
		if path[i] == '\\' {
			i++
		} else if path[i] == '.' {
			if i == len(path)-1 {
				return
			}
			if path[i+1] == '#' {
				i += 2
				if i == len(path) {
					return
				}
				if path[i] == '[' || path[i] == '(' {
					var start, end byte
					if path[i] == '[' {
						start, end = '[', ']'
					} else {
						start, end = '(', ')'
					}
					// inside selector, balance brackets
					i++
					depth := 1
					for ; i < len(path); i++ {
						if path[i] == '\\' {
							i++
						} else if path[i] == start {
							depth++
						} else if path[i] == end {
							depth--
							if depth == 0 {
								break
							}
						} else if path[i] == '"' {
							// inside selector string, balance quotes
							i++
							for ; i < len(path); i++ {
								if path[i] == '\\' {
									i++
								} else if path[i] == '"' {
									break
								}
							}
						}
					}
				}
			}
		} else if path[i] == '|' {
			return path[:i], path[i+1:], true
		}
	}
	return
}

// looksLikeJSONOrTransformer checks whether the first character of the input string `s` is a special character
// (such as '@', '[', or '{') that might indicate a transformer or a JSON structure in the context of processing.
//
// The function performs the following checks:
//   - If the first character is '@', it further inspects if the following characters indicate a transformer.
//   - If the first character is '[' or '{', it returns `true`, indicating a potential JSON array or object.
//   - The function will return `false` for any other characters or if transformers are disabled.
//
// Parameters:
//   - `s`: A string to be checked, which can be a part of a JSON structure or an identifier with a transformer.
//
// Returns:
//   - `bool`: `true` if the first character is '@' followed by a transformer, or if the first character is '[' or '{'.
//     `false` otherwise.
//
// Example Usage:
//
//	s1 := "@transformer|value"
//	looksLikeJSONOrTransformer(s1)
//	// Returns: true (because it starts with '@' and is followed by a transformer)
//
//	s2 := "[1, 2, 3]"
//	looksLikeJSONOrTransformer(s2)
//	// Returns: true (because it starts with '[')
//
//	s3 := "{ \"key\": \"value\" }"
//	looksLikeJSONOrTransformer(s3)
//	// Returns: true (because it starts with '{')
//
//	s4 := "normalString"
//	looksLikeJSONOrTransformer(s4)
//	// Returns: false (no '@', '[', or '{')
//
// Details:
//   - The function first checks if transformers are disabled (by `DisableTransformers` flag). If they are, it returns `false` immediately.
//   - If the string starts with '@', it scans for a potential transformer by checking if there is a '.' or '|' after it,
//     and verifies whether the transformer exists in the `transformers` map.
//   - If the string starts with '[' or '{', it immediately returns `true`, as those characters typically indicate the start of a JSON array or object.
func looksLikeJSONOrTransformer(s string) bool {
	if DisableTransformers {
		return false
	}
	c := s[0]
	if c == '@' {
		i := 1
		for ; i < len(s); i++ {
			if s[i] == '.' || s[i] == '|' || s[i] == ':' {
				break
			}
		}
		ok := globalRegistry.IsRegistered(s[1:i])
		return ok
	}
	return c == '[' || c == '{'
}

// splitPathSegment parses a given path string, extracting different components such as parts, pipes, paths, and wildcards.
// It identifies special characters ('.', '|', '*', '?', '\\') in the path and processes them accordingly. The function
// breaks the string into a part and further splits it into pipes or paths, marking certain flags when necessary.
// It also handles escaped characters by stripping escape sequences and processing them correctly.
//
// Parameters:
//   - `path`: A string representing the path to be parsed. It can contain various special characters like
//     dots ('.'), pipes ('|'), wildcards ('*', '?'), and escape sequences ('\\').
//
// Returns:
//   - `r`: A `wildcard` struct containing the parsed components of the path. It will include:
//   - `Part`: The part of the path before any special character or wildcard.
//   - `Path`: The portion of the path after a dot ('.') if present.
//   - `Pipe`: The portion of the path after a pipe ('|') if present.
//   - `Piped`: A boolean flag indicating if a pipe ('|') was encountered.
//   - `Wild`: A boolean flag indicating if a wildcard ('*' or '?') was encountered.
//   - `More`: A boolean flag indicating if the path is further segmented by a dot ('.').
//
// Example Usage:
//
//	path1 := "field.subfield|anotherField"
//	result := splitPathSegment(path1)
//	// result.part: "field"
//	// result.path: "subfield"
//	// result.pipe: "anotherField"
//	// result.piped: true
//
//	path2 := "object.field"
//	result = splitPathSegment(path2)
//	// result.part: "object"
//	// result.path: "field"
//	// result.more: true
//
//	path3 := "path\\.*.field"
//	result = splitPathSegment(path3)
//	// result.part: "path.*"
//	// result.wild: true
//
// Details:
//   - The function scans through the path string character by character, processing the first encountered special
//     character (either '.', '|', '*', '?', '\\') and extracting the relevant components.
//   - If a '.' is encountered, the part before it is extracted as the `Part`, and the string following it is assigned
//     to `Path`. If there are transformers or JSON structure indicators (like '[' or '{'), the path is marked accordingly.
//   - If a pipe ('|') is found, the `Part` is separated from the string after the pipe, and the `Piped` flag is set to true.
//   - Wildcard characters ('*' or '?') are detected, and the `Wild` flag is set.
//   - Escape sequences (indicated by '\\') are processed by appending the escaped character(s) and stripping the escape character.
//   - If no special characters are found, the entire path is assigned to `Part`, and the function returns the parsed result.
func splitPathSegment(path string) (r wildcard) {
	for i := 0; i < len(path); i++ {
		if path[i] == '|' {
			r.part = path[:i]
			r.pipe = path[i+1:]
			r.piped = true
			return
		}
		if path[i] == '.' {
			r.part = path[:i]
			if i < len(path)-1 && looksLikeJSONOrTransformer(path[i+1:]) {
				r.pipe = path[i+1:]
				r.piped = true
			} else {
				r.path = path[i+1:]
				r.more = true
			}
			return
		}
		if path[i] == '*' || path[i] == '?' {
			r.wild = true
			continue
		}
		if path[i] == '\\' {
			// go into escape mode
			// a slower path that strips off the escape character from the part.
			escapePart := []byte(path[:i])
			i++
			if i < len(path) {
				escapePart = append(escapePart, path[i])
				i++
				for ; i < len(path); i++ {
					if path[i] == '\\' {
						i++
						if i < len(path) {
							escapePart = append(escapePart, path[i])
						}
						continue
					} else if path[i] == '.' {
						r.part = string(escapePart)
						if i < len(path)-1 && looksLikeJSONOrTransformer(path[i+1:]) {
							r.pipe = path[i+1:]
							r.piped = true
						} else {
							r.path = path[i+1:]
							r.more = true
						}
						return
					} else if path[i] == '|' {
						r.part = string(escapePart)
						r.pipe = path[i+1:]
						r.piped = true
						return
					} else if path[i] == '*' || path[i] == '?' {
						r.wild = true
					}
					escapePart = append(escapePart, path[i])
				}
			}
			r.part = string(escapePart)
			return
		}
	}
	r.part = path
	return
}

// matchJSONObjectAt parses a JSON object structure from a given JSON string, extracting key-value pairs based on a specified path.
//
// The function processes a JSON object (denoted by curly braces '{' and '}') and looks for matching keys. It handles both
// simple key-value pairs and nested structures (objects or arrays) within the object. If the path to a key contains wildcards
// or transformers, the function matches the keys accordingly. It also processes escape sequences for both keys and values,
// ensuring proper handling of special characters within JSON strings.
//
// Parameters:
//   - `c`: A pointer to a `parser` object that holds the JSON string (`json`), and context information (`value`).
//   - `i`: The current index in the JSON string from where the parsing should begin. This index should point to the
//     opening curly brace '{' of the JSON object.
//   - `path`: The string representing the path to be parsed. It may include transformers or wildcards, guiding the matching
//     of specific keys in the object.
//
// Returns:
//   - `i` (int): The index in the JSON string after the parsing is completed. This index points to the character
//     immediately after the parsed object.
//   - `bool`: `true` if a match for the specified path was found, `false` if no match was found.
//
// Example Usage:
//
//	json := `{"name": "John", "age": 30, "address": {"city": "New York"}}`
//	i, found := matchJSONObjectAt(c, 0, "name")
//	// found: true (if the "name" key was found in the JSON object)
//
// Details:
//   - The function first searches for a key enclosed in double quotes ('"'). It handles both normal keys and escaped keys.
//   - It then checks if the key matches the specified path, which may contain wildcards or exact matches.
//   - If the key matches and there are no more transformers in the path, the corresponding value is extracted and stored in the `parser` object.
//   - If the key points to a nested object or array, the function recursively parses those structures to extract the required data.
//   - The function handles various types of JSON values including strings, numbers, booleans, objects, and arrays.
//   - The function also handles escape sequences within JSON strings and ensures that they are processed correctly.
//
// Notes:
//   - The function makes use of the `parsePathWithtransformers` function to parse and process the path for matching keys.
//   - If the path contains wildcards ('*' or '?'), the function uses `matchSafely` to ensure safe matching within a complexity limit.
//   - If the key is matched, the function will return the parsed value. If no match is found, the parsing continues.
//
// Key functions used:
//   - `parsePathWithtransformers`: Extracts and processes the path to identify the key and transformers.
//   - `matchSafely`: Performs the safe matching of the key using a wildcard pattern, avoiding excessive complexity.
func matchJSONObjectAt(c *parser, i int, path string) (newPos int, found bool) {
	var _match, keyEsc, escVal, ok, hit bool
	var key, val string
	spseg := splitPathSegment(path)
	if !spseg.more && spseg.piped {
		c.pipe = spseg.pipe
		c.piped = true
	}
	for i < len(c.json) {
		for ; i < len(c.json); i++ {
			if c.json[i] == '"' {
				i++
				var s = i
				for ; i < len(c.json); i++ {
					if c.json[i] > '\\' {
						continue
					}
					if c.json[i] == '"' {
						i, key, keyEsc, ok = i+1, c.json[s:i], false, true
						goto parse_key_completed
					}
					if c.json[i] == '\\' {
						i++
						for ; i < len(c.json); i++ {
							if c.json[i] > '\\' {
								continue
							}
							if c.json[i] == '"' {
								// look for an escaped slash
								if c.json[i-1] == '\\' {
									n := 0
									for j := i - 2; j > 0; j-- {
										if c.json[j] != '\\' {
											break
										}
										n++
									}
									if n%2 == 0 {
										continue
									}
								}
								i, key, keyEsc, ok = i+1, c.json[s:i], true, true
								goto parse_key_completed
							}
						}
						break
					}
				}
				key, keyEsc, ok = c.json[s:], false, false
			parse_key_completed:
				break
			}
			if c.json[i] == '}' {
				return i + 1, false
			}
		}
		if !ok {
			return i, false
		}
		if spseg.wild {
			if keyEsc {
				_match = matchesGlob(unescape(key), spseg.part)
			} else {
				_match = matchesGlob(key, spseg.part)
			}
		} else {
			if keyEsc {
				_match = spseg.part == unescape(key)
			} else {
				_match = spseg.part == key
			}
		}
		hit = _match && !spseg.more
		for ; i < len(c.json); i++ {
			var num bool
			switch c.json[i] {
			default:
				continue
			case '"':
				i++
				i, val, escVal, ok = extractJSONString(c.json, i)
				if !ok {
					return i, false
				}
				if hit {
					if escVal {
						c.val.str = unescape(val[1 : len(val)-1])
					} else {
						c.val.str = val[1 : len(val)-1]
					}
					c.val.raw = val
					c.val.kind = String
					return i, true
				}
			case '{':
				if _match && !hit {
					i, hit = matchJSONObjectAt(c, i+1, spseg.path)
					if hit {
						return i, true
					}
				} else {
					i, val = extractJSONValue(c.json, i)
					if hit {
						c.val.raw = val
						c.val.kind = JSON
						return i, true
					}
				}
			case '[':
				if _match && !hit {
					i, hit = matchJSONArrayAt(c, i+1, spseg.path)
					if hit {
						return i, true
					}
				} else {
					i, val = extractJSONValue(c.json, i)
					if hit {
						c.val.raw = val
						c.val.kind = JSON
						return i, true
					}
				}
			case 'n':
				if i+1 < len(c.json) && c.json[i+1] != 'u' {
					num = true
					break
				}
				fallthrough
			case 't', 'f':
				vc := c.json[i]
				i, val = extractJSONLowerLiteral(c.json, i)
				if hit {
					c.val.raw = val
					switch vc {
					case 't':
						c.val.kind = True
					case 'f':
						c.val.kind = False
					}
					return i, true
				}
			case '+', '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
				'i', 'I', 'N':
				num = true
			}
			if num {
				i, val = extractJSONNumber(c.json, i)
				if hit {
					c.val.raw = val
					c.val.kind = Number
					c.val.num, _ = strconv.ParseFloat(val, 64)
					return i, true
				}
			}
			break
		}
	}
	return i, false
}

// parseFilterQuery parses a query string into its constituent parts and identifies its structure.
// It is designed to handle queries that involve filtering or accessing nested structures,
// particularly in JSON-like or similar data representations.
//
// Parameters:
//   - query (string): The query string to parse. It must start with `#(` or `#[` to be valid.
//     The query string may include paths, operators, and values, and can contain nested structures.
//
// Returns:
//   - path (string): The portion of the query string representing the path to the field or property.
//   - op (string): The operator used in the query (e.g., `==`, `!=`, `<`, `>`, etc.).
//   - value (string): The value to compare against or use in the query.
//   - remain (string): The remaining portion of the query after processing.
//   - i (int): The index in the query string where parsing ended.
//   - _vEsc (bool): Indicates whether the value part contains any escaped characters.
//   - ok (bool): Indicates whether the parsing was successful or not.
//
// Example Usage:
//
//	For a query `#(first_name=="Aris").last`:
//	  - path: "first_name"
//	  - op: "=="
//	  - value: "Aris"
//	  - remain: ".last"
//
//	For a query `#(user_roles.#(=="admin")).privilege`:
//	  - path: "user_roles.#(=="admin")"
//	  - op: ""
//	  - value: ""
//	  - remain: ".privilege"
//
// Details:
//   - The function starts by verifying the query's validity, ensuring it begins with `#(` or `#[`.
//   - It processes the query character by character, accounting for nested structures, operators, and escaped characters.
//   - The `path` is extracted from the portion of the query before the operator or value.
//   - The `op` and `value` are identified and split if an operator is present.
//   - Remaining characters in the query, such as `.last` in the example, are captured in `remain`.
//
// Notes:
//   - The function supports a variety of operators (`==`, `!=`, `<`, `>`, etc.).
//   - It handles nested brackets or parentheses and ensures balanced nesting.
//   - Escaped characters (e.g., `\"`) within the query are processed correctly, with `_vEsc` indicating their presence.
//   - If the query is invalid or incomplete, the function will return `ok` as `false`.
//
// Edge Cases:
//   - Handles nested queries with multiple levels of depth.
//   - Ensures proper handling of invalid or malformed queries by returning appropriate values.
func parseFilterQuery(query string) (path, op, value, remain string, endPos int, valueEsc bool, ok bool) {
	if len(query) < 2 || query[0] != '#' ||
		(query[1] != '(' && query[1] != '[') {
		return "", "", "", "", endPos, false, false
	}
	endPos = 2
	j := 0
	depth := 1
	for ; endPos < len(query); endPos++ {
		if depth == 1 && j == 0 {
			switch query[endPos] {
			case '!', '=', '<', '>', '%':
				j = endPos
				continue
			}
		}
		if query[endPos] == '\\' {
			endPos++
		} else if query[endPos] == '[' || query[endPos] == '(' {
			depth++
		} else if query[endPos] == ']' || query[endPos] == ')' {
			depth--
			if depth == 0 {
				break
			}
		} else if query[endPos] == '"' {
			endPos++
			for ; endPos < len(query); endPos++ {
				if query[endPos] == '\\' {
					valueEsc = true
					endPos++
				} else if query[endPos] == '"' {
					break
				}
			}
		}
	}
	if depth > 0 {
		return "", "", "", "", endPos, false, false
	}
	if j > 0 {
		path = strutil.Trim(query[2:j])
		value = strutil.Trim(query[j:endPos])
		remain = query[endPos+1:]
		var trail int
		switch {
		case len(value) == 1:
			trail = 1
		case value[0] == '!' && value[1] == '=':
			trail = 2
		case value[0] == '!' && value[1] == '%':
			trail = 2
		case value[0] == '<' && value[1] == '=':
			trail = 2
		case value[0] == '>' && value[1] == '=':
			trail = 2
		case value[0] == '=' && value[1] == '=':
			value = value[1:]
			trail = 1
		case value[0] == '<':
			trail = 1
		case value[0] == '>':
			trail = 1
		case value[0] == '=':
			trail = 1
		case value[0] == '%':
			trail = 1
		}
		op = value[:trail]
		value = strutil.Trim(value[trail:])
	} else {
		path = strutil.Trim(query[2:endPos])
		remain = query[endPos+1:]
	}
	return path, op, value, remain, endPos + 1, valueEsc, true
}

// classifyPathSegment parses a string path into its structural components, breaking it into meaningful parts
// such as the main path, pipe (if present), query parameters, and nested paths. This function is particularly
// useful for processing JSON-like paths or other hierarchical data representations.
//
// Parameters:
//   - path (string): The input string path to be analyzed. It may contain various symbols such as '|', '.',
//     or '#' that represent different parts or behaviors.
//
// Returns:
//   - r (deeper): A struct containing the parsed components of the path, such as `Part`, `Pipe`, `Path`,
//     and additional meta (e.g., `piped`, `more`, `arch`, and query-related fields).
//
// Fields in `deeper`:
//   - `part`: The main part of the path before special characters like '.', '|', or '#'.
//   - `path`: The remaining part of the path after '.' or other separators.
//   - `pipe`: A piped portion of the path, if separated by '|', indicating a subsequent operation.
//   - `piped`: A boolean indicating whether the path contains a pipe ('|').
//   - `more`: A boolean indicating whether there is more of the path to process after the first separator.
//   - `arch`: A boolean indicating the presence of a '#' in the path, signifying an archive or query operation.
//   - `logOk`: A boolean indicating a valid archive log if the path starts with `#.`.
//   - `logKey`: The key following `#.` for archive logging, if applicable.
//   - `query`: A nested struct providing details about a query if the path contains query operations:
//   - `on`: Indicates whether the path contains a query (e.g., starting with `#(`).
//   - `all`: Indicates whether the query applies to all elements.
//   - `path`: The path portion of the query.
//   - `opt`: The operator used in the query (e.g., `==`, `!=`, etc.).
//   - `val`: The value used in the query.
//   - `opt`: The operator for comparison.
//   - `val`: The query value.
//
// Details:
//   - The function iterates through the `path` string character by character, identifying and processing special symbols
//     such as '|', '.', and '#'.
//   - If the path contains a '|', the portion before it is stored in `Part`, and the portion after it is stored in `Pipe`.
//     The `Piped` flag is set to `true`.
//   - If the path contains a '.', the portion before it is stored in `Part`, and the remaining part in `Path`.
//     If the path after the '.' starts with a transformer or JSON, it is stored in `Pipe` instead, with `Piped` set to `true`.
//   - If the path contains a '#', the `Arch` flag is set to `true`. It may also indicate an archive log (`#.key`) or a query (`#(...)`).
//     Queries are parsed using the `analyzeQuery` function, and relevant fields in the `query` struct are populated.
//   - For archive logs starting with `#.` (e.g., `#.key`), the `ALogOk` flag is set, and `ALogKey` contains the key.
//   - If the path contains a query, the function extracts and processes the query's path, operator, and value.
//     Queries are denoted by a '#' followed by '[' or '(' (e.g., `#[...]` or `#(...)`).
//
// Example Usage:
//
//	For Input: "data|filter.name"
//	   Part: "data"
//	   Pipe: "filter.name"
//	   Piped: true
//	   Path: ""
//	   More: false
//	   Arch: false
//
//	For Input: "items.#(value=='42').details"
//	   Part: "items"
//	   Path: "#(value=='42').details"
//	   Arch: true
//	   query.On: true
//	   query.path: "value"
//	   query.Option: "=="
//	   query.Value: "42"
//	   query.All: false
//
//	For Input: "#.log"
//	   Part: "#"
//	   Path: ""
//	   ALogOk: true
//	   ALogKey: "log"
//
// Notes:
//   - The function is robust against malformed paths but assumes valid inputs for proper operation.
//   - It ensures nested paths and queries are correctly identified and processed.
//
// Edge Cases:
//   - If no special characters are found, the entire input is stored in `Part`.
//   - If the path contains an incomplete or invalid query, the function skips the query parsing gracefully.
func classifyPathSegment(path string) (r meta) {
	for i := 0; i < len(path); i++ {
		if path[i] == '|' {
			r.part = path[:i]
			r.pipe = path[i+1:]
			r.piped = true
			return
		}
		if path[i] == '.' {
			r.part = path[:i]
			if !r.arch && i < len(path)-1 && looksLikeJSONOrTransformer(path[i+1:]) {
				r.pipe = path[i+1:]
				r.piped = true
			} else {
				r.path = path[i+1:]
				r.more = true
			}
			return
		}
		if path[i] == '#' {
			r.arch = true
			if i == 0 && len(path) > 1 {
				if path[1] == '.' {
					r.logOk = true
					r.logKey = path[2:]
					r.path = path[:1]
				} else if path[1] == '[' || path[1] == '(' {
					// query
					r.query.on = true
					queryPath, op, value, _, fi, escVal, ok :=
						parseFilterQuery(path[i:])
					if !ok {
						break
					}
					if len(value) >= 2 && value[0] == '"' &&
						value[len(value)-1] == '"' {
						value = value[1 : len(value)-1]
						if escVal {
							value = unescape(value)
						}
					}
					r.query.path = queryPath
					r.query.opt = op
					r.query.val = value

					i = fi - 1
					if i+1 < len(path) && path[i+1] == '#' {
						r.query.all = true
					}
				}
			}
			continue
		}
	}
	r.part = path
	r.path = ""
	return
}

// matchJSONArrayAt processes and evaluates the path in the context of an array, checking
// for matches and executing queries on elements within the array. It is responsible
// for handling nested structures and queries, as well as determining if the analysis
// matches the current path and value in the context.
//
// Parameters:
//   - c (parser*): A pointer to the `parser` object that holds the current context,
//     including the JSON data and the parsed path.
//   - i (int): The current index in the JSON string being processed.
//   - path (string): The path to be analyzed for array processing.
//
// Returns:
//   - (int): The updated index after processing the array.
//   - (bool): A boolean indicating whether the analysis was successful or not.
//
// Details:
//   - The function analyzes a path related to arrays and performs various checks to
//     determine if the array elements match the specified path and conditions.
//   - It checks for array literals, objects, and nested structures, invoking appropriate
//     parsing functions for each type.
//   - If a query is present, it will evaluate the query on the current element and decide
//     whether to continue the search or return a match.
//   - The function supports queries on array elements (e.g., matching specific values),
//     and it can return results in JSON format or execute specific actions (like calculating a value).
//
// Flow:
//   - The function first processes the path and ensures that it is valid for array analysis.
//   - It checks if the path includes an archive log, and if so, handles logging operations.
//   - The core loop processes each element of the array, checking for string, numeric, object,
//     or array elements and evaluating whether they match the query conditions, if any.
//   - If the query is satisfied, the function performs further processing on the matching element,
//     such as storing the result or calculating a value. If no query is provided, it directly
//     sets the `c.val` with the matched result.
//   - It also handles special cases like archive logs and nested array structures.
//   - If no valid match is found, the function returns `false`, and the search continues.
//
// Example:
//
//	Input: `["apple", "banana", "cherry"]`
//	If the query was for "banana", the function would find a match and return the result.
//
// Edge Cases:
//   - Handles situations where no array is found or the query fails to match any element.
//   - Properly handles nested arrays or objects within the JSON data, maintaining structure.
//   - Takes into account escaped characters and special syntax (e.g., queries, JSON objects).
func matchJSONArrayAt(c *parser, i int, path string) (int, bool) {
	var _match, escVal, ok, hit bool
	var val string
	var h int
	var aLog []int
	var partIdx int
	var multics []byte
	var queryIndexes []int
	analysis := classifyPathSegment(path)
	if !analysis.arch {
		n, ok := parseUint64(analysis.part)
		if !ok {
			partIdx = -1
		} else {
			partIdx = int(n)
		}
	}
	if !analysis.more && analysis.piped {
		c.pipe = analysis.pipe
		c.piped = true
	}

	executeQuery := func(eVal Context) bool {
		if analysis.query.all {
			if len(multics) == 0 {
				multics = append(multics, '[')
			}
		}
		var tmp parser
		tmp.val = eVal
		computeOffset(c.json, &tmp)
		parentIndex := tmp.val.idx
		var res Context
		if eVal.kind == JSON {
			res = eVal.Get(analysis.query.path)
		} else {
			if analysis.query.path != "" {
				return false
			}
			res = eVal
		}
		if matchesQueryCondition(&analysis, res) {
			if analysis.more {
				left, right, ok := splitAtUnescapedPipe(analysis.path)
				if ok {
					analysis.path = left
					c.pipe = right
					c.piped = true
				}
				res = eVal.Get(analysis.path)
			} else {
				res = eVal
			}
			if analysis.query.all {
				raw := res.raw
				if len(raw) == 0 {
					raw = res.String()
				}
				if raw != "" {
					if len(multics) > 1 {
						multics = append(multics, ',')
					}
					multics = append(multics, raw...)
					queryIndexes = append(queryIndexes, res.idx+parentIndex)
				}
			} else {
				c.val = res
				return true
			}
		}
		return false
	}
	for i < len(c.json)+1 {
		if !analysis.arch {
			_match = partIdx == h
			hit = _match && !analysis.more
		}
		h++
		if analysis.logOk {
			aLog = append(aLog, i)
		}
		for ; ; i++ {
			var ch byte
			if i > len(c.json) {
				break
			} else if i == len(c.json) {
				ch = ']'
			} else {
				ch = c.json[i]
			}
			var num bool
			switch ch {
			default:
				continue
			case '"':
				i++
				i, val, escVal, ok = extractJSONString(c.json, i)
				if !ok {
					return i, false
				}
				if analysis.query.on {
					var cVal Context
					if escVal {
						cVal.str = unescape(val[1 : len(val)-1])
					} else {
						cVal.str = val[1 : len(val)-1]
					}
					cVal.raw = val
					cVal.kind = String
					if executeQuery(cVal) {
						return i, true
					}
				} else if hit {
					if analysis.logOk {
						break
					}
					if escVal {
						c.val.str = unescape(val[1 : len(val)-1])
					} else {
						c.val.str = val[1 : len(val)-1]
					}
					c.val.raw = val
					c.val.kind = String
					return i, true
				}
			case '{':
				if _match && !hit {
					i, hit = matchJSONObjectAt(c, i+1, analysis.path)
					if hit {
						if analysis.logOk {
							break
						}
						return i, true
					}
				} else {
					i, val = extractJSONValue(c.json, i)
					if analysis.query.on {
						if executeQuery(Context{raw: val, kind: JSON}) {
							return i, true
						}
					} else if hit {
						if analysis.logOk {
							break
						}
						c.val.raw = val
						c.val.kind = JSON
						return i, true
					}
				}
			case '[':
				if _match && !hit {
					i, hit = matchJSONArrayAt(c, i+1, analysis.path)
					if hit {
						if analysis.logOk {
							break
						}
						return i, true
					}
				} else {
					i, val = extractJSONValue(c.json, i)
					if analysis.query.on {
						if executeQuery(Context{raw: val, kind: JSON}) {
							return i, true
						}
					} else if hit {
						if analysis.logOk {
							break
						}
						c.val.raw = val
						c.val.kind = JSON
						return i, true
					}
				}
			case 'n':
				if i+1 < len(c.json) && c.json[i+1] != 'u' {
					num = true
					break
				}
				fallthrough
			case 't', 'f':
				vc := c.json[i]
				i, val = extractJSONLowerLiteral(c.json, i)
				if analysis.query.on {
					var cVal Context
					cVal.raw = val
					switch vc {
					case 't':
						cVal.kind = True
					case 'f':
						cVal.kind = False
					}
					if executeQuery(cVal) {
						return i, true
					}
				} else if hit {
					if analysis.logOk {
						break
					}
					c.val.raw = val
					switch vc {
					case 't':
						c.val.kind = True
					case 'f':
						c.val.kind = False
					}
					return i, true
				}
			case '+', '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
				'i', 'I', 'N':
				num = true
			case ']':
				if analysis.arch && analysis.part == "#" {
					if analysis.logOk {
						left, right, ok := splitAtUnescapedPipe(analysis.logKey)
						if ok {
							analysis.logKey = left
							c.pipe = right
							c.piped = true
						}
						var indexes = make([]int, 0, 64)
						var jsonVal = make([]byte, 0, 64)
						jsonVal = append(jsonVal, '[')
						for j, k := 0, 0; j < len(aLog); j++ {
							idx := aLog[j]
							for idx < len(c.json) {
								switch c.json[idx] {
								case ' ', '\t', '\r', '\n':
									idx++
									continue
								}
								break
							}
							if idx < len(c.json) && c.json[idx] != ']' {
								_, res, ok := extractJSONValueAt(c.json, idx, true)
								if ok {
									res := res.Get(analysis.logKey)
									if res.Exists() {
										if k > 0 {
											jsonVal = append(jsonVal, ',')
										}
										raw := res.raw
										if len(raw) == 0 {
											raw = res.String()
										}
										jsonVal = append(jsonVal, []byte(raw)...)
										indexes = append(indexes, res.idx)
										k++
									}
								}
							}
						}
						jsonVal = append(jsonVal, ']')
						c.val.kind = JSON
						c.val.raw = string(jsonVal)
						c.val.idxs = indexes
						return i + 1, true
					}
					if analysis.logOk {
						break
					}

					c.val.kind = Number
					c.val.num = float64(h - 1)
					c.val.raw = strconv.Itoa(h - 1)
					c.calc = true
					return i + 1, true
				}
				if !c.val.Exists() {
					if len(multics) > 0 {
						c.val = Context{
							raw:  string(append(multics, ']')),
							kind: JSON,
							idxs: queryIndexes,
						}
					} else if analysis.query.all {
						c.val = Context{
							raw:  "[]",
							kind: JSON,
						}
					}
				}
				return i + 1, false
			}
			if num {
				i, val = extractJSONNumber(c.json, i)
				if analysis.query.on {
					var cVal Context
					cVal.raw = val
					cVal.kind = Number
					cVal.num, _ = strconv.ParseFloat(val, 64)
					if executeQuery(cVal) {
						return i, true
					}
				} else if hit {
					if analysis.logOk {
						break
					}
					c.val.raw = val
					c.val.kind = Number
					c.val.num, _ = strconv.ParseFloat(val, 64)
					return i, true
				}
			}
			break
		}
	}
	return i, false
}

// parseSubSelectors parses a sub-selection string, which can either be in the form of
// '[path1,path2]' or '{"field1":path1,"field2":path2}' type structure. It returns the parsed
// selectors from the given path, which includes the name and path of each selector within
// the structure. The function assumes that the first character in the path is either '[' or '{',
// and this check is expected to be performed before calling the function.
//
// Parameters:
//   - path: A string representing the sub-selection in either array or object format. The string
//     must begin with either '[' (array) or '{' (object), and the structure should contain
//     valid selectors or field-path pairs.
//
// Returns:
//   - selectors: A slice of `sel` structs containing the parsed selectors and their associated paths.
//   - remaining: The remaining part of the path after parsing the selectors. This will be the part following the
//     closing bracket (']') or brace ('}') if applicable.
//   - ok: A boolean indicating whether the parsing was successful. It returns true if the parsing was
//     successful and the structure was valid, or false if there was an error during parsing.
//
// Example Usage:
//
//	path := "[field1:subpath1,field2:subpath2]"
//	selectors, remaining, ok := parseSubSelectors(path)
//	// selectors: [{name: "field1", path: "subpath1"}, {name: "field2", path: "subpath2"}]
//	// remaining: "" (no remaining part of the path)
//	// ok: true (parsing was successful)
//
// Details:
//   - The function iterates through each character of the input path and identifies different types
//     of characters (e.g., commas, colons, brackets, braces, and quotes).
//   - It tracks the depth of nested structures (array or object) using the `depth` variable. This ensures
//     proper handling of nested elements within the sub-selection.
//   - The function supports escaping characters with backslashes ('\') and handles this case while parsing.
//   - If a colon (':') is encountered, it indicates a potential name-path pair. The function captures
//     the name and path accordingly, and if no colon is found, it assumes the value is just a path.
//   - The function handles both array-style sub-selections (e.g., [path1,path2]) and object-style
//     sub-selections (e.g., {"field1":path1,"field2":path2}).
//   - If an error is encountered during parsing (e.g., mismatched brackets or braces), the function
//     returns an empty slice and `false` to indicate a failure.
//
// Flow:
//   - The function first initializes tracking variables like `transformer`, `depth`, `colon`, and `start`.
//   - It iterates through the path, checking for different characters, such as backslashes (escape),
//     colons (for name-path pair separation), commas (for separating selectors), and brackets/braces (for
//     nested structures).
//   - If a valid selector is found, it is stored in the `selectors` slice.
//   - The function returns the parsed selectors, the remaining path, and a success flag.
func parseSubSelectors(path string) (selectors []sel, remaining string, ok bool) {
	transformer := 0
	depth := 1
	colon := 0
	start := 1
	i := 1
	pushSelectors := func() {
		var selector sel
		if colon == 0 {
			selector.path = path[start:i]
		} else {
			selector.name = path[start:colon]
			selector.path = path[colon+1 : i]
		}
		selectors = append(selectors, selector)
		colon = 0
		transformer = 0
		start = i + 1
	}
	for ; i < len(path); i++ {
		switch path[i] {
		case '\\':
			i++
		case '@':
			if transformer == 0 && i > 0 && (path[i-1] == '.' || path[i-1] == '|') {
				transformer = i
			}
		case ':':
			if transformer == 0 && colon == 0 && depth == 1 {
				colon = i
			}
		case ',':
			if depth == 1 {
				pushSelectors()
			}
		case '"':
			i++
		loop:
			for ; i < len(path); i++ {
				switch path[i] {
				case '\\':
					i++
				case '"':
					break loop
				}
			}
		case '[', '(', '{':
			depth++
		case ']', ')', '}':
			depth--
			if depth == 0 {
				pushSelectors()
				path = path[i+1:]
				return selectors, path, true
			}
		}
	}
	return
}

// applyTransformerAt parses a given path to identify a transformer function and its associated arguments,
// then applies the transformer to the provided JSON string based on the parsed path. This function expects
// that the path starts with a '@', indicating the presence of a transformer. It identifies the transformer's
// name, extracts any potential arguments, and returns the modified result along with the remaining path
// after processing the transformer.
//
// Parameters:
//   - json: A string containing the JSON data that the transformer will operate on.
//   - path: A string representing the path, which includes a transformer prefixed by '@'. The path may
//     contain an optional argument to be processed by the transformer.
//
// Returns:
//   - remainingPath: The remaining portion of the path after parsing the transformer and its arguments.
//   - result: The result obtained by applying the transformer to the JSON string, or an empty string
//     if no valid transformer is found.
//   - ok: A boolean indicating whether the transformer was successfully identified and applied. If true,
//     the transformer was found and applied; if false, the transformer was not found.
//
// Example Usage:
//
//	json := `{"key": "value"}`
//	path := "@transformerName:argument"
//	remainingPath, result, ok := applyTransformerAt(json, path)
//	// remainingPath: remaining path after the transformer
//	// result: the modified JSON result based on the transformer applied
//	// ok: true if the transformer was found and applied successfully
//
// Details:
//   - The function first removes the '@' character from the beginning of the path and processes the
//     remaining portion of the path to extract the transformer's name and its optional arguments.
//   - The function handles various formats of arguments, including JSON-like objects, arrays, strings,
//     and other specific cases based on the character delimiters such as '{', '[', '"', or '('.
//   - If a valid transformer function is found in the `transformers` map, it applies the function to the JSON
//     string and returns the result along with the remaining path. If no valid transformer is found, it
//     returns the original path and an empty result.
func applyTransformerAt(json, path string) (remainingPath, result string, ok bool) {
	name := path[1:] // remove the '@' character and initialize the name to the remaining path.
	var hasArgs bool
	// iterate over the path to find the transformer name and any arguments.
	for i := 1; i < len(path); i++ {
		// check for argument delimiter (':'), process if found.
		if path[i] == ':' {
			remainingPath = path[i+1:]
			name = path[1:i]
			hasArgs = len(remainingPath) > 0
			break
		}
		// check for pipe ('|'), dot ('.'), or other delimiters to separate the transformer name and arguments.
		if path[i] == '|' {
			remainingPath = path[i:]
			name = path[1:i]
			break
		}
		if path[i] == '.' {
			remainingPath = path[i:]
			name = path[1:i]
			break
		}
	}
	// check if the transformer exists in the global registry and apply it if found.
	if t := resolveTransformer(name); t != nil {
		var args string
		if hasArgs { // if arguments are found, parse and handle them.
			var parsedArgs bool
			// process the arguments based on their type (e.g., JSON, string, etc.).
			switch remainingPath[0] {
			case '{', '[', '"': // handle JSON-like arguments.
				ctx := Parse(remainingPath)
				if ctx.Exists() {
					args = compactJSON(remainingPath) // squash the JSON to remove nested structures.
					remainingPath = remainingPath[len(args):]
					parsedArgs = true
				}
			}
			if !parsedArgs { // process arguments if not already parsed as JSON.
				i := 0
				// iterate through the arguments and process any nested structures or strings.
				for ; i < len(remainingPath); i++ {
					if remainingPath[i] == '|' {
						break
					}
					switch remainingPath[i] {
					case '{', '[', '"', '(': // handle nested structures like arrays or objects.
						s := compactJSON(remainingPath[i:])
						i += len(s) - 1
					}
				}
				args = remainingPath[:i]          // extract the argument portion.
				remainingPath = remainingPath[i:] // update the remaining path.
			}
		}
		// dispatch to the transformer and return the result.
		return remainingPath, t.Apply(json, args), true
	}
	// if no transformer is found, return the path and an empty result.
	return remainingPath, result, false
}

// isNull checks whether a given `Context` represents a JSON null value.
//
// Parameters:
//   - t: A `Context` struct that contains information about a specific JSON value.
//
// Returns:
//   - bool: Returns `true` if the `kind` field of the provided `Context` is `Null`,
//     indicating that the JSON value is null. Otherwise, returns `false`.
//
// Example Usage:
//
//	ctx := Context{kind: Null}
//	isNull := isNull(ctx)
//	// isNull: true
//
//	ctx = Context{kind: String, strings: "example"}
//	isNull = isNull(ctx)
//	// isNull: false
//
// Notes:
//   - This function provides a convenient way to check if a JSON value is null,
//     allowing for easier handling of such cases in JSON processing.
func isNull(t Context) bool {
	return t.kind == Null
}

// isFalsy determines if the given `Context` represents a "falsy" value.
//
// A value is considered "falsy" if:
//   - It is a JSON null (`Null`).
//   - It is a JSON false (`False`).
//   - It is a string that can be parsed as a boolean and evaluates to `false` (e.g., "false", "0").
//   - It is a number and equals zero.
//
// Parameters:
//   - t: A `Context` struct that contains information about a specific JSON value.
//
// Returns:
//   - bool: Returns `true` if the `Context` represents a falsy value; otherwise, returns `false`.
//
// Example Usage:
//
//	ctx := Context{kind: False}
//	isFalse := isFalsy(ctx)
//	// isFalse: true
//
//	ctx = Context{kind: String, strings: "false"}
//	isFalse = isFalsy(ctx)
//	// isFalse: true
//
//	ctx = Context{kind: Number, numeric: 1.0}
//	isFalse = isFalsy(ctx)
//	// isFalse: false
//
// Notes:
//   - For string values, the function attempts to parse the string as a boolean.
//     If parsing fails, the value is not considered falsy.
//   - Numeric values are considered falsy only if they equal zero.
func isFalsy(t Context) bool {
	switch t.kind {
	case Null:
		return true
	case False:
		return true
	case String:
		b := conv.BoolOrDefault(t.str, false)
		return !b
	case Number:
		return t.num == 0
	default:
		return false
	}
}

// isTruthy determines if the given `Context` represents a "truthy" value.
//
// A value is considered "truthy" if:
//   - It is a JSON true (`True`).
//   - It is a string that can be parsed as a boolean and evaluates to `true` (e.g., "true", "1").
//   - It is a number and does not equal zero.
//
// Parameters:
//   - t: A `Context` struct that contains information about a specific JSON value.
//
// Returns:
//   - bool: Returns `true` if the `Context` represents a truthy value; otherwise, returns `false`.
//
// Example Usage:
//
//	ctx := Context{kind: True}
//	isTrue := isTruthy(ctx)
//	// isTrue: true
//
//	ctx = Context{kind: String, strings: "true"}
//	isTrue = isTruthy(ctx)
//	// isTrue: true
//
//	ctx = Context{kind: Number, numeric: 0.0}
//	isTrue = isTruthy(ctx)
//	// isTrue: false
//
// Notes:
//   - For string values, the function attempts to parse the string as a boolean.
//     If parsing fails, the value is not considered truthy.
//   - Numeric values are considered truthy if they do not equal zero.
func isTruthy(t Context) bool {
	switch t.kind {
	case True:
		return true
	case String:
		b := conv.BoolOrDefault(t.str, false)
		return b
	case Number:
		return t.num != 0
	default:
		return false
	}
}

// matchesQueryCondition determines whether a given `Context` value matches the conditions specified in the [meta] query.
//
// This function evaluates a JSON path query against a specific `Context` value, checking for matching conditions such as
// existence, equality, inequality, and other relational operations. It supports operations on strings, numbers, and booleans.
//
// Parameters:
//   - dp: A pointer to the [meta] structure containing query details, such as the value to match (`Value`) and
//     the comparison option (`Option`).
//   - value: A `Context` structure representing the JSON value to be evaluated against the query.
//
// Returns:
//   - bool: Returns `true` if the `Context` matches the query conditions; otherwise, returns `false`.
//
// Query Matching Process:
//  1. If the query value (`Value`) starts with a `~`, it is treated as a special type such as a wildcard
//     (`*`), `null`, `true`, or `false`.
//  2. The function evaluates whether the `value` exists in the JSON structure. If it doesn't exist, the result is `false`.
//  3. If no `Option` is provided in the query, the function checks for the existence of the value (`Exists`).
//  4. Based on the type of the `value` (e.g., `String`, `Number`, `True`, `False`), the function applies the query's
//     `Option` to perform comparisons or match patterns.
//
// Supported Query Options:
//   - `=`: Checks for equality.
//   - `!=`: Checks for inequality.
//   - `<`, `<=`: Checks if the value is less than or equal to the query value.
//   - `>`, `>=`: Checks if the value is greater than or equal to the query value.
//   - `%`: Checks if the value matches a regular expression (string only).
//   - `!%`: Checks if the value does not match a regular expression (string only).
//
// Example Usage:
//
//	dp := &meta{query: {opt: "=", val: "example"}}
//	value := Context{kind: String, str: "example"}
//	matches := matchesQueryCondition(dp, value)
//	// matches: true
//
// Notes:
//   - For wildcard queries (`~`), special handling applies to determine if the `Context` satisfies the query type.
//   - Boolean values (`True` or `False`) are evaluated based on string representations of "true" and "false".
//   - Numeric comparisons rely on parsing the query value into a float64.
//
// Limitations:
//   - String pattern matching (`%`, `!%`) relies on the `matchSafely` function, which is not defined here.
//   - Unsupported types or operations return `false`.
func matchesQueryCondition(dp *meta, value Context) bool {
	mt := dp.query.val
	if len(mt) > 0 {
		if mt[0] == '~' {
			mt = mt[1:]
			var ish, ok bool
			switch mt {
			case "*":
				ish, ok = value.Exists(), true
			case "null":
				ish, ok = isNull(value), true
			case "true":
				ish, ok = isTruthy(value), true
			case "false":
				ish, ok = isFalsy(value), true
			}
			if ok {
				mt = "true"
				if ish {
					value = Context{kind: True}
				} else {
					value = Context{kind: False}
				}
			} else {
				mt = ""
				value = Context{}
			}
		}
	}
	if !value.Exists() {
		return false
	}
	if dp.query.opt == "" {
		return true
	}
	switch value.kind {
	case String:
		switch dp.query.opt {
		case "=":
			return value.str == mt
		case "!=":
			return value.str != mt
		case "<":
			return value.str < mt
		case "<=":
			return value.str <= mt
		case ">":
			return value.str > mt
		case ">=":
			return value.str >= mt
		case "%":
			return matchesGlob(value.str, mt)
		case "!%":
			return !matchesGlob(value.str, mt)
		}
	case Number:
		_rightVal, _ := strconv.ParseFloat(mt, 64)
		switch dp.query.opt {
		case "=":
			return value.num == _rightVal
		case "!=":
			return value.num != _rightVal
		case "<":
			return value.num < _rightVal
		case "<=":
			return value.num <= _rightVal
		case ">":
			return value.num > _rightVal
		case ">=":
			return value.num >= _rightVal
		}
	case True:
		switch dp.query.opt {
		case "=":
			return mt == "true"
		case "!=":
			return mt != "true"
		case ">":
			return mt == "false"
		case ">=":
			return true
		}
	case False:
		switch dp.query.opt {
		case "=":
			return mt == "false"
		case "!=":
			return mt != "false"
		case "<":
			return mt == "true"
		case "<=":
			return true
		}
	}
	return false
}

// appendJSONString converts a given string into a valid JSON string format
// and appends it to the provided byte slice `dst`.
//
// This function escapes special characters in the input string `s` to ensure
// that it adheres to the JSON string encoding rules, such as escaping double
// quotes, backslashes, and control characters. Additionally, it handles UTF-8
// characters and appends them in their proper encoded format.
//
// Parameters:
//   - dst: A byte slice to which the encoded JSON string will be appended.
//   - s: The input string to be converted into JSON string format.
//
// Returns:
//   - []byte: The resulting byte slice containing the original content of `dst`
//     with the JSON-encoded string appended.
//
// Details:
//   - The function begins by appending space for the string `s` and wrapping
//     it in double quotes.
//   - It iterates through the input string `s` character by character and checks
//     for specific cases where escaping or additional encoding is required:
//   - Control characters (`\n`, `\r`, `\t`) are replaced with their escape
//     sequences (`\\n`, `\\r`, `\\t`).
//   - Characters like `<`, `>`, and `&` are escaped using Unicode notation
//     to ensure the resulting JSON string is safe for embedding in HTML or XML.
//   - Backslashes (`\`) and double quotes (`"`) are escaped with a preceding
//     backslash (`\\`).
//   - UTF-8 characters are properly encoded, and unsupported characters or
//     decoding errors are replaced with the Unicode replacement character
//     (`\ufffd`).
//
// Example Usage:
//
//	dst := []byte("Current JSON: ")
//	s := "Hello \"world\"\nLine break!"
//	result := appendJSONString(dst, s)
//	// result: []byte(`Current JSON: "Hello \"world\"\nLine break!"`)
//
// Notes:
//   - This function is useful for building JSON-encoded strings dynamically
//     without allocating new memory for each operation.
//   - It ensures that the resulting JSON string is safe and adheres to
//     encoding rules for use in various contexts such as web APIs or
//     configuration files.
func appendJSONString(target []byte, s string) []byte {
	target = append(target, make([]byte, len(s)+2)...)
	target = append(target[:len(target)-len(s)-2], '"')
	for i := 0; i < len(s); i++ {
		if s[i] < ' ' {
			target = append(target, '\\')
			switch s[i] {
			case '\n':
				target = append(target, 'n')
			case '\r':
				target = append(target, 'r')
			case '\t':
				target = append(target, 't')
			default:
				target = append(target, 'u')
				target = appendHex(target, uint16(s[i]))
			}
		} else if s[i] == '>' || s[i] == '<' || s[i] == '&' {
			target = append(target, '\\', 'u')
			target = appendHex(target, uint16(s[i]))
		} else if s[i] == '\\' {
			target = append(target, '\\', '\\')
		} else if s[i] == '"' {
			target = append(target, '\\', '"')
		} else if s[i] > 127 {
			r, n := utf8.DecodeRuneInString(s[i:]) // read utf8 character
			if n == 0 {
				break
			}
			if r == utf8.RuneError && n == 1 {
				target = append(target, `\ufffd`...)
			} else if r == '\u2028' || r == '\u2029' {
				target = append(target, `\u202`...)
				target = append(target, hexDigits[r&0xF])
			} else {
				target = append(target, s[i:i+n]...)
			}
			i = i + n - 1
		} else {
			target = append(target, s[i])
		}
	}
	return append(target, '"')
}

// recurseCollectMatches recursively traverses a JSON structure to find all matches
// for a specified path within nested objects or arrays.
//
// This function performs a depth-first traversal of the JSON structure starting from
// a given parent `Context`, and it collects all the `Context` results that match
// the specified path. It works by first attempting to find a match for the path at
// the current level and then recursively explores any nested objects or arrays to
// find additional matches.
//
// Parameters:
//   - `all`: A slice of `Context` that accumulates the results. It is initially empty
//     and is populated with matching `Context` objects found during the traversal.
//   - `parent`: The `Context` representing the current JSON element being processed.
//     It acts as the starting point for the search in this recursive descent.
//   - `path`: A string representing the JSON path to search for. This path is used
//     to query the current level and to guide the search deeper into nested structures.
//
// Returns:
//   - A slice of `Context` containing all the results that match the specified path.
//     The slice is accumulated during the recursive descent, and all matches, including
//     those found in nested objects and arrays, are added to the result.
//
// Example Usage:
//
//	json := `{
//	  "store": {
//	    "book": [
//	      { "category": "fiction", "author": "J.K. Rowling", "title": "Harry Potter" },
//	      { "category": "science", "author": "Stephen Hawking", "title": "A Brief History of Time" }
//	    ],
//	    "music": [
//	      { "artist": "The Beatles", "album": "Abbey Road" },
//	      { "artist": "Pink Floyd", "album": "The Wall" }
//	    ]
//	  }
//	}`
//
//	parent := fj.Get(json, "store")
//	results := recurseCollectMatches(nil, parent, "book.title")
//
//	// `results` will contain:
//	// ["Harry Potter", "A Brief History of Time"]
//	// The function searches for the "book.title" path in the store and collects all matches
//	// found within the nested book array in the store object.
//
// Notes:
//   - The function leverages recursive descent to explore nested JSON objects and arrays,
//     ensuring that all levels of the structure are searched for matches.
//   - If the `parent` element is an object or array, it will iterate over its elements and
//     perform recursive descent for each of them.
func recurseCollectMatches(all []Context, parent Context, path string) []Context {
	if matched := parent.Get(path); matched.Exists() {
		all = append(all, matched)
	}
	if parent.IsArray() || parent.IsObject() {
		parent.Foreach(func(_, ctx Context) bool {
			all = recurseCollectMatches(all, ctx, path)
			return true
		})
	}
	return all
}

// escapeUnsafeChars processes a string `component` to escape characters that are not considered safe
// according to the `isSafePathKeyByte` function. It inserts a backslash (`\`) before each unsafe
// character, ensuring that the resulting string contains only safe characters.
//
// Parameters:
//   - `component`: A string that may contain unsafe characters that need to be escaped.
//
// Returns:
//   - A new string with unsafe characters escaped by prefixing them with a backslash (`\`).
//
// Notes:
//   - The function iterates through the input string and checks each character using the
//     `isSafePathKeyByte` function. When it encounters an unsafe character, it escapes it with a backslash.
//   - Once an unsafe character is found, the function adds a backslash before each subsequent unsafe character
//     and continues until the end of the string.
//
// Example:
//
//	component := "key-with$pecial*chars"
//	escaped := escapeUnsafeChars(component) // escaped: "key-with\$pecial\*chars"
func escapeUnsafeChars(component string) string {
	for i := 0; i < len(component); i++ {
		if !isSafePathKeyByte(component[i]) {
			noneComponent := []byte(component[:i])
			for ; i < len(component); i++ {
				if !isSafePathKeyByte(component[i]) {
					noneComponent = append(noneComponent, '\\')
				}
				noneComponent = append(noneComponent, component[i])
			}
			return string(noneComponent)
		}
	}
	return component
}

// trimOuterBrackets removes the surrounding '[]' or '{}' characters from a JSON string.
// This function is useful when you want to extract the content inside a JSON array or object,
// effectively unwrapping the outermost brackets or braces.
//
// Parameters:
//   - json: A string representing a JSON object or array. The string may include square brackets ('[]') or
//     curly braces ('{}') at the beginning and end, which will be removed if they exist.
//
// Returns:
//   - A new string with the outermost '[]' or '{}' characters removed. If the string does not start
//     and end with matching brackets or braces, the string remains unchanged.
//
// Example Usage:
//
//	json := "[1, 2, 3]"
//	unwrapped := trimOuterBrackets(json)
//	// unwrapped: "1, 2, 3" (the array removed)
//
//	json = "{ \"name\": \"John\" }"
//	unwrapped = trimOuterBrackets(json)
//	// unwrapped: " \"name\": \"John\" " (the object removed)
//
//	str := "hello world"
//	unwrapped = trimOuterBrackets(str)
//	// unwrapped: "hello world" (no change since no surrounding brackets or braces)
//
// Details:
//
//   - The function first trims any leading or trailing whitespace from the input string using the `trim` function.
//
//   - It then checks if the string has at least two characters and if the first character is either '[' or '{'.
//
//   - If the first character is an opening bracket or brace, and the last character matches its pair (']' or '}'),
//     the function removes both the first and last characters.
//
//   - If the string does not start and end with matching brackets or braces, the original string is returned unchanged.
//
//   - The function handles cases where the string may contain additional whitespace at the beginning or end by trimming it first.
func trimOuterBrackets(json string) string {
	json = strutil.Trim(json)
	if len(json) >= 2 && (json[0] == '[' || json[0] == '{') {
		json = json[1 : len(json)-1]
	}
	return json
}

// scanLeaves is the internal recursive worker for Search.
// It appends to `all` every scalar leaf whose String() contains keyword.
//
// Parameters:
//   - `all`: A slice of `Context` that accumulates the results. It is initially empty
//     and is populated with matching `Context` objects found during the traversal.
//   - `node`: The `Context` representing the current JSON element being processed.
//     It acts as the starting point for the search in this recursive descent.
//   - `keyword`: The keyword to search for within the string representation of each leaf node.
//
// Returns:
//   - A slice of `Context` containing all the results that match the specified keyword.
//     The slice is accumulated during the recursive descent, and all matches, including
//     those found in nested objects and arrays, are added to the result.
//
// Example Usage:
//
//	json := `{
//	  "store": {
//	    "book": [
//	      { "category": "fiction", "author": "J.K. Rowling", "title": "Harry Potter" },
//	      { "category": "science", "author": "Stephen Hawking", "title": "A Brief History of Time" }
//	    ],
//	    "music": [
//	      { "artist": "The Beatles", "album": "Abbey Road" },
//	      { "artist": "Pink Floyd", "album": "The Wall" }
//	    ]
//	  }
//	}`
//
//	parent := fj.Get(json, "store")
//	results := scanLeaves(nil, parent, "Harry")
//
//	// `results` will contain:
//	// ["Harry Potter"]
//	// The function searches for the "Harry" keyword in the store and collects all matches
//	// found within the nested book array in the store object.
//
// Notes:
//   - The function leverages recursive descent to explore nested JSON objects and arrays,
//     ensuring that all levels of the structure are searched for matches.
//   - If the `parent` element is an object or array, it will iterate over its elements and
//     perform recursive descent for each of them.
//   - The search is performed on the string representation of each leaf node using `node.String()`.
//   - The `keyword` is checked for emptiness using `strutil.IsEmpty()`. If the keyword is empty,
//     all leaf nodes will be considered matches and added to the result.
func scanLeaves(all []Context, node Context, keyword string) []Context {
	if node.IsArray() || node.IsObject() {
		node.Foreach(func(_, child Context) bool {
			all = scanLeaves(all, child, keyword)
			return true
		})
		return all
	}
	if !node.Exists() {
		return all
	}
	if strutil.IsEmpty(keyword) || strings.Contains(node.String(), keyword) {
		all = append(all, node)
	}
	return all
}

// scanLeavesMatch is the internal recursive worker for SearchMatch.
//
// Parameters:
//   - `all`: The slice of Context values to append matches to.
//   - `node`: The current Context node to search.
//   - `pattern`: The wildcard pattern to match against.
//
// Returns:
//   - A slice of `Context` containing all the matches found.
//
// Example Usage:
//
//	scanLeavesMatch(nil, node, "*.name")
//
// Notes:
//   - The function leverages recursive descent to explore nested JSON objects and arrays,
//     ensuring that all levels of the structure are searched for matches.
//   - If the `parent` element is an object or array, it will iterate over its elements and
//     perform recursive descent for each of them.
//   - The search is performed on the values of the JSON elements using `node.String()`.
//   - The `value` is checked for equality using `node.String() == value`.
func scanLeavesMatch(all []Context, node Context, pattern string) []Context {
	if node.IsArray() || node.IsObject() {
		node.Foreach(func(_, child Context) bool {
			all = scanLeavesMatch(all, child, pattern)
			return true
		})
		return all
	}
	if !node.Exists() {
		return all
	}
	if match.Match(node.String(), pattern) {
		all = append(all, node)
	}
	return all
}

// scanByKey is the internal recursive worker for SearchByKey.
//
// Parameters:
//   - `all`: A slice of `Context` that accumulates the results. It is initially empty
//     and is populated with matching `Context` objects found during the traversal.
//   - `node`: The `Context` representing the current JSON element being processed.
//     It acts as the starting point for the search in this recursive descent.
//   - `keySet`: A map of strings representing the keys to search for within the JSON structure.
//
// Returns:
//   - A slice of `Context` containing all the results that match the specified keys.
//     The slice is accumulated during the recursive descent, and all matches, including
//     those found in nested objects and arrays, are added to the result.
//
// Example Usage:
//
//	json := `{
//	  "store": {
//	    "book": [
//	      { "category": "fiction", "author": "J.K. Rowling", "title": "Harry Potter" },
//	      { "category": "science", "author": "Stephen Hawking", "title": "A Brief History of Time" }
//	    ],
//	    "music": [
//	      { "artist": "The Beatles", "album": "Abbey Road" },
//	      { "artist": "Pink Floyd", "album": "The Wall" }
//	    ]
//	  }
//	}`
//
//	parent := fj.Get(json, "store")
//	keySet := map[string]struct{}{"author": {}, "album": {}}
//	results := scanByKey(nil, parent, keySet)
//
//	// `results` will contain:
//	// ["J.K. Rowling", "Stephen Hawking", "The Beatles", "Pink Floyd"]
//	// The function searches for the "author" and "album" keys in the store and collects all matches
//	// found within the nested book and music arrays in the store object.
//
// Notes:
//   - The function leverages recursive descent to explore nested JSON objects and arrays,
//     ensuring that all levels of the structure are searched for matches.
//   - If the `parent` element is an object or array, it will iterate over its elements and
//     perform recursive descent for each of them.
//   - The search is performed on the keys of the JSON elements using `key.String()`.
//   - The `keySet` is checked for the presence of each key using `keySet[key.String()]`.
func scanByKey(all []Context, node Context, keySet map[string]struct{}) []Context {
	if node.IsObject() {
		node.Foreach(func(key, val Context) bool {
			if _, ok := keySet[key.String()]; ok {
				all = append(all, val)
			}
			// Recurse into value regardless of whether the key matched.
			if val.IsObject() || val.IsArray() {
				all = scanByKey(all, val, keySet)
			}
			return true
		})
		return all
	}
	if node.IsArray() {
		node.Foreach(func(_, child Context) bool {
			all = scanByKey(all, child, keySet)
			return true
		})
	}
	return all
}

// scanByKeyPattern is the internal recursive worker for SearchByKeyPattern.
//
// Parameters:
//   - `all`: The slice of Context values to append matches to.
//   - `node`: The current Context node to search.
//   - `keyPattern`: The wildcard pattern to match against.
//
// Returns:
//   - A slice of `Context` containing all the matches found.
//
// Example Usage:
//
//	scanByKeyPattern(nil, node, "*.name")
//
// Notes:
//   - The function leverages recursive descent to explore nested JSON objects and arrays,
//     ensuring that all levels of the structure are searched for matches.
//   - If the `parent` element is an object or array, it will iterate over its elements and
//     perform recursive descent for each of them.
//   - The search is performed on the values of the JSON elements using `node.String()`.
//   - The `value` is checked for equality using `node.String() == value`.
func scanByKeyPattern(all []Context, node Context, keyPattern string) []Context {
	if node.IsObject() {
		node.Foreach(func(key, val Context) bool {
			if match.Match(key.String(), keyPattern) {
				all = append(all, val)
			}
			if val.IsObject() || val.IsArray() {
				all = scanByKeyPattern(all, val, keyPattern)
			}
			return true
		})
		return all
	}
	if node.IsArray() {
		node.Foreach(func(_, child Context) bool {
			all = scanByKeyPattern(all, child, keyPattern)
			return true
		})
	}
	return all
}

// scanPath is the depth-first worker for FindPath.
// Returns the first matching path and a bool indicating whether it was found.
//
// Parameters:
//   - `node`: The `Context` representing the current JSON element being processed.
//     It acts as the starting point for the search in this recursive descent.
//   - `value`: The value to search for within the JSON structure.
//   - `prefix`: The prefix to use for the path.
//
// Returns:
//   - A string representing the path to the first matching value.
//   - A boolean indicating whether a matching value was found.
//
// Example Usage:
//
//	json := `{
//	  "store": {
//	    "book": [
//	      { "category": "fiction", "author": "J.K. Rowling", "title": "Harry Potter" },
//	      { "category": "science", "author": "Stephen Hawking", "title": "A Brief History of Time" }
//	    ],
//	    "music": [
//	      { "artist": "The Beatles", "album": "Abbey Road" },
//	      { "artist": "Pink Floyd", "album": "The Wall" }
//	    ]
//	  }
//	}`
//
//	parent := fj.Get(json, "store")
//	path, found := scanPath(parent, "Harry Potter", "")
//
//	// `path` will be "book.0.title"
//	// `found` will be true
//	// The function searches for the "Harry Potter" value in the store and returns the first path found.
//
// Notes:
//   - The function leverages recursive descent to explore nested JSON objects and arrays,
//     ensuring that all levels of the structure are searched for matches.
//   - If the `parent` element is an object or array, it will iterate over its elements and
//     perform recursive descent for each of them.
//   - The search is performed on the values of the JSON elements using `node.String()`.
//   - The `value` is checked for equality using `node.String() == value`.
func scanPath(node Context, value, prefix string) (string, bool) {
	if node.IsObject() {
		var found string
		var ok bool
		node.Foreach(func(key, child Context) bool {
			p := joinPath(prefix, key.String())
			if child.IsObject() || child.IsArray() {
				found, ok = scanPath(child, value, p)
			} else if child.Exists() && child.String() == value {
				found, ok = p, true
			}
			return !ok
		})
		return found, ok
	}
	if node.IsArray() {
		var found string
		var ok bool
		idx := 0
		node.Foreach(func(_, child Context) bool {
			p := joinPath(prefix, itoa(idx))
			if child.IsObject() || child.IsArray() {
				found, ok = scanPath(child, value, p)
			} else if child.Exists() && child.String() == value {
				found, ok = p, true
			}
			idx++
			return !ok
		})
		return found, ok
	}
	return "", false
}

// scanPathMatch is the depth-first worker for FindPathMatch.
//
// Parameters:
//   - `node`: The current Context node to search.
//   - `pattern`: The wildcard pattern to match against.
//   - `prefix`: The prefix to prepend to the path.
//
// Returns:
//   - The first matching path and a boolean indicating whether it was found.
//
// Example Usage:
//
//	json := `{
//	  "store": {
//	    "book": [
//	      { "category": "fiction", "author": "J.K. Rowling", "title": "Harry Potter" },
//	      { "category": "science", "author": "Stephen Hawking", "title": "A Brief History of Time" }
//	    ]
//	  }
//	}`
//
//	parent := fj.Get(json, "store")
//	path, found := scanPathMatch(parent, "*.title", "")
//
//	// `path` will be "book.0.title"
//	// `found` will be true
//	// The function searches for the "Harry Potter" value in the store and returns the first path found.
//
// Notes:
//   - The function leverages recursive descent to explore nested JSON objects and arrays,
//     ensuring that all levels of the structure are searched for matches.
//   - If the `parent` element is an object or array, it will iterate over its elements and
//     perform recursive descent for each of them.
//   - The search is performed on the values of the JSON elements using `node.String()`.
//   - The `value` is checked for equality using `node.String() == value`.
func scanPathMatch(node Context, pattern, prefix string) (string, bool) {
	if node.IsObject() {
		var found string
		var ok bool
		node.Foreach(func(key, child Context) bool {
			p := joinPath(prefix, key.String())
			if child.IsObject() || child.IsArray() {
				found, ok = scanPathMatch(child, pattern, p)
			} else if child.Exists() && match.Match(child.String(), pattern) {
				found, ok = p, true
			}
			return !ok
		})
		return found, ok
	}
	if node.IsArray() {
		var found string
		var ok bool
		idx := 0
		node.Foreach(func(_, child Context) bool {
			p := joinPath(prefix, itoa(idx))
			if child.IsObject() || child.IsArray() {
				found, ok = scanPathMatch(child, pattern, p)
			} else if child.Exists() && match.Match(child.String(), pattern) {
				found, ok = p, true
			}
			idx++
			return !ok
		})
		return found, ok
	}
	return "", false
}

// scanPaths is the depth-first worker for FindPaths.
//
// Parameters:
//   - `all`: A slice of `string` that accumulates the results. It is initially empty
//     and is populated with matching `string` representations of paths found during the traversal.
//   - `node`: The `Context` representing the current JSON element being processed.
//     It acts as the starting point for the search in this recursive descent.
//   - `value`: The value to search for within the JSON structure.
//   - `prefix`: The prefix to use for the path.
//
// Returns:
//   - A slice of `string` containing all the paths that match the specified value.
//     The slice is accumulated during the recursive descent, and all matches, including
//     those found in nested objects and arrays, are added to the result.
//
// Example Usage:
//
//	json := `{
//	  "store": {
//	    "book": [
//	      { "category": "fiction", "author": "J.K. Rowling", "title": "Harry Potter" },
//	      { "category": "science", "author": "Stephen Hawking", "title": "A Brief History of Time" }
//	    ],
//	    "music": [
//	      { "artist": "The Beatles", "album": "Abbey Road" },
//	      { "artist": "Pink Floyd", "album": "The Wall" }
//	    ]
//	  }
//	}`
//
//	parent := fj.Get(json, "store")
//	paths := scanPaths(nil, parent, "Harry Potter", "")
//
//	// `paths` will contain:
//	// ["book.0.title"]
//	// The function searches for the "Harry Potter" value in the store and collects all paths found.
//
// Notes:
//   - The function leverages recursive descent to explore nested JSON objects and arrays,
//     ensuring that all levels of the structure are searched for matches.
//   - If the `parent` element is an object or array, it will iterate over its elements and
//     perform recursive descent for each of them.
//   - The search is performed on the values of the JSON elements using `node.String()`.
//   - The `value` is checked for equality using `node.String() == value`.
func scanPaths(all []string, node Context, value, prefix string) []string {
	if node.IsObject() {
		node.Foreach(func(key, child Context) bool {
			p := joinPath(prefix, key.String())
			if child.IsObject() || child.IsArray() {
				all = scanPaths(all, child, value, p)
			} else if child.Exists() && child.String() == value {
				all = append(all, p)
			}
			return true
		})
		return all
	}
	if node.IsArray() {
		idx := 0
		node.Foreach(func(_, child Context) bool {
			p := joinPath(prefix, itoa(idx))
			if child.IsObject() || child.IsArray() {
				all = scanPaths(all, child, value, p)
			} else if child.Exists() && child.String() == value {
				all = append(all, p)
			}
			idx++
			return true
		})
	}
	return all
}

// scanPathsMatch is the depth-first worker for FindPathsMatch.
//
// Parameters:
//   - `all`: The slice of paths to append matches to.
//   - `node`: The current Context node to search.
//   - `pattern`: The wildcard pattern to match against.
//   - `prefix`: The prefix to prepend to the path.
//
// Returns:
//   - A slice of strings representing the paths to the matching values.
//
// Example Usage:
//
//	json := `{
//	  "store": {
//	    "book": [
//	      { "category": "fiction", "author": "J.K. Rowling", "title": "Harry Potter" },
//	      { "category": "science", "author": "Stephen Hawking", "title": "A Brief History of Time" }
//	    ]
//	  }
//	}`
//
//	parent := fj.Get(json, "store")
//	paths := scanPathsMatch(nil, parent, "*.title", "")
//
//	// `paths` will contain:
//	// ["book.0.title"]
//	// The function searches for the "Harry Potter" value in the store and collects all paths found.
//
// Notes:
//   - The function leverages recursive descent to explore nested JSON objects and arrays,
//     ensuring that all levels of the structure are searched for matches.
//   - If the `parent` element is an object or array, it will iterate over its elements and
//     perform recursive descent for each of them.
//   - The search is performed on the values of the JSON elements using `node.String()`.
//   - The `value` is checked for equality using `node.String() == value`.
func scanPathsMatch(all []string, node Context, pattern, prefix string) []string {
	if node.IsObject() {
		node.Foreach(func(key, child Context) bool {
			p := joinPath(prefix, key.String())
			if child.IsObject() || child.IsArray() {
				all = scanPathsMatch(all, child, pattern, p)
			} else if child.Exists() && match.Match(child.String(), pattern) {
				all = append(all, p)
			}
			return true
		})
		return all
	}
	if node.IsArray() {
		idx := 0
		node.Foreach(func(_, child Context) bool {
			p := joinPath(prefix, itoa(idx))
			if child.IsObject() || child.IsArray() {
				all = scanPathsMatch(all, child, pattern, p)
			} else if child.Exists() && match.Match(child.String(), pattern) {
				all = append(all, p)
			}
			idx++
			return true
		})
	}
	return all
}

// sortField extracts the value used as sort key for a Context element.
// If keyField is non-empty, it gets that sub-field; otherwise it uses the element itself.
//
// Parameters:
//   - `item`: The Context element to extract the sort key from.
//   - `keyField`: The field to extract the sort key from.
//
// Returns:
//   - The sort key.
//
// Example Usage:
//
//	json := `{
//	  "store": {
//	    "book": [
//	      { "category": "fiction", "author": "J.K. Rowling", "title": "Harry Potter" },
//	      { "category": "science", "author": "Stephen Hawking", "title": "A Brief History of Time" }
//	    ]
//	  }
//	}`
//
//	parent := fj.Get(json, "store")
//	paths := scanPathsMatch(nil, parent, "*.title", "")
//
//	// `paths` will contain:
//	// ["book.0.title"]
//	// The function searches for the "Harry Potter" value in the store and collects all paths found.
//
// Notes:
//   - The function leverages recursive descent to explore nested JSON objects and arrays,
//     ensuring that all levels of the structure are searched for matches.
//   - If the `parent` element is an object or array, it will iterate over its elements and
//     perform recursive descent for each of them.
//   - The search is performed on the values of the JSON elements using `node.String()`.
//   - The `value` is checked for equality using `node.String() == value`.
func sortField(item Context, keyField string) Context {
	if strutil.IsEmpty(keyField) {
		return item
	}
	return item.Get(keyField)
}

// sortOrdered returns true when a should come before b in ascending sort order.
// Numeric values are compared as float64 via conv.Float64.
// All other values fall back to string comparison via conv.String.
//
// Parameters:
//   - `a`: The first Context element to compare.
//   - `b`: The second Context element to compare.
//
// Returns:
//   - `true` if `a` should come before `b` in ascending sort order, `false` otherwise.
//
// Example Usage:
//
//	json := `{
//	  "store": {
//	    "book": [
//	      { "category": "fiction", "author": "J.K. Rowling", "title": "Harry Potter" },
//	      { "category": "science", "author": "Stephen Hawking", "title": "A Brief History of Time" }
//	    ]
//	  }
//	}`
//
//	parent := fj.Get(json, "store")
//	paths := scanPathsMatch(nil, parent, "*.title", "")
//
//	// `paths` will contain:
//	// ["book.0.title"]
//	// The function searches for the "Harry Potter" value in the store and collects all paths found.
//
// Notes:
//   - The function leverages recursive descent to explore nested JSON objects and arrays,
//     ensuring that all levels of the structure are searched for matches.
//   - If the `parent` element is an object or array, it will iterate over its elements and
//     perform recursive descent for each of them.
//   - The search is performed on the values of the JSON elements using `node.String()`.
//   - The `value` is checked for equality using `node.String() == value`.
func sortOrdered(a, b Context) bool {
	if a.kind == Number || b.kind == Number {
		fa, errA := conv.Float64(a.Value())
		fb, errB := conv.Float64(b.Value())
		if errA == nil && errB == nil {
			return fa < fb
		}
	}
	sa, _ := conv.String(a.Value())
	sb, _ := conv.String(b.Value())
	return sa < sb
}

// joinPath concatenates a dot-notation prefix with a segment, inserting "." only
// when the prefix is non-empty.
//
// Parameters:
//   - `prefix`: The prefix to use for the path.
//   - `segment`: The segment to append to the prefix.
//
// Returns:
//   - A string representing the combined path.
//
// Example Usage:
//
//	path := joinPath("store.book", "0.title")
//
// `path` will be "store.book.0.title"
//
// Notes:
//   - The function leverages recursive descent to explore nested JSON objects and arrays,
//     ensuring that all levels of the structure are searched for matches.
//   - If the `parent` element is an object or array, it will iterate over its elements and
//     perform recursive descent for each of them.
//   - The search is performed on the values of the JSON elements using `node.String()`.
//   - The `value` is checked for equality using `node.String() == value`.
func joinPath(prefix, segment string) string {
	if strutil.IsEmpty(prefix) {
		return segment
	}
	return prefix + "." + segment
}

// itoa converts a non-negative integer to its decimal string representation without
// importing the strconv package.
//
// Parameters:
//   - `n`: The non-negative integer to convert.
//
// Returns:
//   - A string representation of the integer.
//
// Example Usage:
//
//	path := itoa(123)
//
// `path` will be "123"
//
// Notes:
//   - The function leverages recursive descent to explore nested JSON objects and arrays,
//     ensuring that all levels of the structure are searched for matches.
//   - If the `parent` element is an object or array, it will iterate over its elements and
//     perform recursive descent for each of them.
//   - The search is performed on the values of the JSON elements using `node.String()`.
//   - The `value` is checked for equality using `node.String() == value`.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := [20]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[pos:])
}

// scanInt is an internal helper shared by Sum, Min, Max, and Avg.
// It visits every Context returned by path (treating a JSON array result as a
// sequence of individual values) and calls fn for each numeric one.
//
// Parameters:
//   - `json`: The JSON string to parse.
//   - `path`: The path to evaluate against the JSON.
//   - `fn`: The function to call for each numeric value found.
//
// Returns:
//   - None
//
// Example Usage:
//
//	scanInt(json, "scores", func(n int) { total += n })
//
// Notes:
//   - The function leverages recursive descent to explore nested JSON objects and arrays,
//     ensuring that all levels of the structure are searched for matches.
//   - If the `parent` element is an object or array, it will iterate over its elements and
//     perform recursive descent for each of them.
//   - The search is performed on the values of the JSON elements using `node.String()`.
//   - The `value` is checked for equality using `node.String() == value`.
func scanInt(json, path string, fn func(int)) {
	ctx := Get(json, path)
	if !ctx.Exists() {
		return
	}
	if ctx.IsArray() {
		ctx.Foreach(func(_, item Context) bool {
			if item.kind == Number {
				fn(item.Int())
			}
			return true
		})
		return
	}
	if ctx.kind == Number {
		fn(ctx.Int())
	}
}

// scanInt8 is an internal helper shared by Sum, Min, Max, and Avg.
// It visits every Context returned by path (treating a JSON array result as a
// sequence of individual values) and calls fn for each numeric one.
//
// Parameters:
//   - `json`: The JSON string to parse.
//   - `path`: The path to evaluate against the JSON.
//   - `fn`: The function to call for each numeric value found.
//
// Returns:
//   - None
//
// Example Usage:
//
//	scanInt8(json, "scores", func(n int8) { total += n })
//
// Notes:
//   - The function leverages recursive descent to explore nested JSON objects and arrays,
//     ensuring that all levels of the structure are searched for matches.
//   - If the `parent` element is an object or array, it will iterate over its elements and
//     perform recursive descent for each of them.
//   - The search is performed on the values of the JSON elements using `node.String()`.
//   - The `value` is checked for equality using `node.String() == value`.
func scanInt8(json, path string, fn func(int8)) {
	ctx := Get(json, path)
	if !ctx.Exists() {
		return
	}
	if ctx.IsArray() {
		ctx.Foreach(func(_, item Context) bool {
			if item.kind == Number {
				fn(item.Int8())
			}
			return true
		})
		return
	}
	if ctx.kind == Number {
		fn(ctx.Int8())
	}
}

// scanInt16 is an internal helper shared by Sum, Min, Max, and Avg.
// It visits every Context returned by path (treating a JSON array result as a
// sequence of individual values) and calls fn for each numeric one.
//
// Parameters:
//   - `json`: The JSON string to parse.
//   - `path`: The path to evaluate against the JSON.
//   - `fn`: The function to call for each numeric value found.
//
// Returns:
//   - None
//
// Example Usage:
//
//	scanInt16(json, "scores", func(n int16) { total += n })
//
// Notes:
//   - The function leverages recursive descent to explore nested JSON objects and arrays,
//     ensuring that all levels of the structure are searched for matches.
//   - If the `parent` element is an object or array, it will iterate over its elements and
//     perform recursive descent for each of them.
//   - The search is performed on the values of the JSON elements using `node.String()`.
//   - The `value` is checked for equality using `node.String() == value`.
func scanInt16(json, path string, fn func(int16)) {
	ctx := Get(json, path)
	if !ctx.Exists() {
		return
	}
	if ctx.IsArray() {
		ctx.Foreach(func(_, item Context) bool {
			if item.kind == Number {
				fn(item.Int16())
			}
			return true
		})
		return
	}
	if ctx.kind == Number {
		fn(ctx.Int16())
	}
}

// scanInt32 is an internal helper shared by Sum, Min, Max, and Avg.
// It visits every Context returned by path (treating a JSON array result as a
// sequence of individual values) and calls fn for each numeric one.
//
// Parameters:
//   - `json`: The JSON string to parse.
//   - `path`: The path to evaluate against the JSON.
//   - `fn`: The function to call for each numeric value found.
//
// Returns:
//   - None
//
// Example Usage:
//
//	scanInt32(json, "scores", func(n int32) { total += n })
//
// Notes:
//   - The function leverages recursive descent to explore nested JSON objects and arrays,
//     ensuring that all levels of the structure are searched for matches.
//   - If the `parent` element is an object or array, it will iterate over its elements and
//     perform recursive descent for each of them.
//   - The search is performed on the values of the JSON elements using `node.String()`.
//   - The `value` is checked for equality using `node.String() == value`.
func scanInt32(json, path string, fn func(int32)) {
	ctx := Get(json, path)
	if !ctx.Exists() {
		return
	}
	if ctx.IsArray() {
		ctx.Foreach(func(_, item Context) bool {
			if item.kind == Number {
				fn(item.Int32())
			}
			return true
		})
		return
	}
	if ctx.kind == Number {
		fn(ctx.Int32())
	}
}

// scanInt64 is an internal helper shared by Sum, Min, Max, and Avg.
// It visits every Context returned by path (treating a JSON array result as a
// sequence of individual values) and calls fn for each numeric one.
//
// Parameters:
//   - `json`: The JSON string to parse.
//   - `path`: The path to evaluate against the JSON.
//   - `fn`: The function to call for each numeric value found.
//
// Returns:
//   - None
//
// Example Usage:
//
//	scanInt64(json, "scores", func(n int64) { total += n })
//
// Notes:
//   - The function leverages recursive descent to explore nested JSON objects and arrays,
//     ensuring that all levels of the structure are searched for matches.
//   - If the `parent` element is an object or array, it will iterate over its elements and
//     perform recursive descent for each of them.
//   - The search is performed on the values of the JSON elements using `node.String()`.
//   - The `value` is checked for equality using `node.String() == value`.
func scanInt64(json, path string, fn func(int64)) {
	ctx := Get(json, path)
	if !ctx.Exists() {
		return
	}
	if ctx.IsArray() {
		ctx.Foreach(func(_, item Context) bool {
			if item.kind == Number {
				fn(item.Int64())
			}
			return true
		})
		return
	}
	if ctx.kind == Number {
		fn(ctx.Int64())
	}
}

// scanUint is an internal helper shared by Sum, Min, Max, and Avg.
// It visits every Context returned by path (treating a JSON array result as a
// sequence of individual values) and calls fn for each numeric one.
//
// Parameters:
//   - `json`: The JSON string to parse.
//   - `path`: The path to evaluate against the JSON.
//   - `fn`: The function to call for each numeric value found.
//
// Returns:
//   - None
//
// Example Usage:
//
//	scanUint(json, "scores", func(n uint) { total += n })
//
// Notes:
//   - The function leverages recursive descent to explore nested JSON objects and arrays,
//     ensuring that all levels of the structure are searched for matches.
//   - If the `parent` element is an object or array, it will iterate over its elements and
//     perform recursive descent for each of them.
//   - The search is performed on the values of the JSON elements using `node.String()`.
//   - The `value` is checked for equality using `node.String() == value`.
func scanUint(json, path string, fn func(uint)) {
	ctx := Get(json, path)
	if !ctx.Exists() {
		return
	}
	if ctx.IsArray() {
		ctx.Foreach(func(_, item Context) bool {
			if item.kind == Number {
				fn(item.Uint())
			}
			return true
		})
		return
	}
	if ctx.kind == Number {
		fn(ctx.Uint())
	}
}

// scanUint8 is an internal helper shared by Sum, Min, Max, and Avg.
// It visits every Context returned by path (treating a JSON array result as a
// sequence of individual values) and calls fn for each numeric one.
//
// Parameters:
//   - `json`: The JSON string to parse.
//   - `path`: The path to evaluate against the JSON.
//   - `fn`: The function to call for each numeric value found.
//
// Returns:
//   - None
//
// Example Usage:
//
//	scanUint8(json, "scores", func(n uint8) { total += n })
//
// Notes:
//   - The function leverages recursive descent to explore nested JSON objects and arrays,
//     ensuring that all levels of the structure are searched for matches.
//   - If the `parent` element is an object or array, it will iterate over its elements and
//     perform recursive descent for each of them.
//   - The search is performed on the values of the JSON elements using `node.String()`.
//   - The `value` is checked for equality using `node.String() == value`.
func scanUint8(json, path string, fn func(uint8)) {
	ctx := Get(json, path)
	if !ctx.Exists() {
		return
	}
	if ctx.IsArray() {
		ctx.Foreach(func(_, item Context) bool {
			if item.kind == Number {
				fn(item.Uint8())
			}
			return true
		})
		return
	}
	if ctx.kind == Number {
		fn(ctx.Uint8())
	}
}

// scanUint16 is an internal helper shared by Sum, Min, Max, and Avg.
// It visits every Context returned by path (treating a JSON array result as a
// sequence of individual values) and calls fn for each numeric one.
//
// Parameters:
//   - `json`: The JSON string to parse.
//   - `path`: The path to evaluate against the JSON.
//   - `fn`: The function to call for each numeric value found.
//
// Returns:
//   - None
//
// Example Usage:
//
//	scanUint16(json, "scores", func(n uint16) { total += n })
//
// Notes:
//   - The function leverages recursive descent to explore nested JSON objects and arrays,
//     ensuring that all levels of the structure are searched for matches.
//   - If the `parent` element is an object or array, it will iterate over its elements and
//     perform recursive descent for each of them.
//   - The search is performed on the values of the JSON elements using `node.String()`.
//   - The `value` is checked for equality using `node.String() == value`.
func scanUint16(json, path string, fn func(uint16)) {
	ctx := Get(json, path)
	if !ctx.Exists() {
		return
	}
	if ctx.IsArray() {
		ctx.Foreach(func(_, item Context) bool {
			if item.kind == Number {
				fn(item.Uint16())
			}
			return true
		})
		return
	}
	if ctx.kind == Number {
		fn(ctx.Uint16())
	}
}

// scanUint32 is an internal helper shared by Sum, Min, Max, and Avg.
// It visits every Context returned by path (treating a JSON array result as a
// sequence of individual values) and calls fn for each numeric one.
//
// Parameters:
//   - `json`: The JSON string to parse.
//   - `path`: The path to evaluate against the JSON.
//   - `fn`: The function to call for each numeric value found.
//
// Returns:
//   - None
//
// Example Usage:
//
//	scanUint32(json, "scores", func(n uint32) { total += n })
//
// Notes:
//   - The function leverages recursive descent to explore nested JSON objects and arrays,
//     ensuring that all levels of the structure are searched for matches.
//   - If the `parent` element is an object or array, it will iterate over its elements and
//     perform recursive descent for each of them.
//   - The search is performed on the values of the JSON elements using `node.String()`.
//   - The `value` is checked for equality using `node.String() == value`.
func scanUint32(json, path string, fn func(uint32)) {
	ctx := Get(json, path)
	if !ctx.Exists() {
		return
	}
	if ctx.IsArray() {
		ctx.Foreach(func(_, item Context) bool {
			if item.kind == Number {
				fn(item.Uint32())
			}
			return true
		})
		return
	}
	if ctx.kind == Number {
		fn(ctx.Uint32())
	}
}

// scanUint64 is an internal helper shared by Sum, Min, Max, and Avg.
// It visits every Context returned by path (treating a JSON array result as a
// sequence of individual values) and calls fn for each numeric one.
//
// Parameters:
//   - `json`: The JSON string to parse.
//   - `path`: The path to evaluate against the JSON.
//   - `fn`: The function to call for each numeric value found.
//
// Returns:
//   - None
//
// Example Usage:
//
//	scanUint64(json, "scores", func(n uint64) { total += n })
//
// Notes:
//   - The function leverages recursive descent to explore nested JSON objects and arrays,
//     ensuring that all levels of the structure are searched for matches.
//   - If the `parent` element is an object or array, it will iterate over its elements and
//     perform recursive descent for each of them.
//   - The search is performed on the values of the JSON elements using `node.String()`.
//   - The `value` is checked for equality using `node.String() == value`.
func scanUint64(json, path string, fn func(uint64)) {
	ctx := Get(json, path)
	if !ctx.Exists() {
		return
	}
	if ctx.IsArray() {
		ctx.Foreach(func(_, item Context) bool {
			if item.kind == Number {
				fn(item.Uint64())
			}
			return true
		})
		return
	}
	if ctx.kind == Number {
		fn(ctx.Uint64())
	}
}

// scanFloat32 is an internal helper shared by Sum, Min, Max, and Avg.
// It visits every Context returned by path (treating a JSON array result as a
// sequence of individual values) and calls fn for each numeric one.
//
// Parameters:
//   - `json`: The JSON string to parse.
//   - `path`: The path to evaluate against the JSON.
//   - `fn`: The function to call for each numeric value found.
//
// Returns:
//   - None
//
// Example Usage:
//
//	scanFloat32(json, "scores", func(n float32) { total += n })
//
// Notes:
//   - The function leverages recursive descent to explore nested JSON objects and arrays,
//     ensuring that all levels of the structure are searched for matches.
//   - If the `parent` element is an object or array, it will iterate over its elements and
//     perform recursive descent for each of them.
//   - The search is performed on the values of the JSON elements using `node.String()`.
//   - The `value` is checked for equality using `node.String() == value`.
func scanFloat32(json, path string, fn func(float32)) {
	ctx := Get(json, path)
	if !ctx.Exists() {
		return
	}
	if ctx.IsArray() {
		ctx.Foreach(func(_, item Context) bool {
			if item.kind == Number {
				fn(item.Float32())
			}
			return true
		})
		return
	}
	if ctx.kind == Number {
		fn(ctx.Float32())
	}
}

// scanFloat64 is an internal helper shared by Sum, Min, Max, and Avg.
// It visits every Context returned by path (treating a JSON array result as a
// sequence of individual values) and calls fn for each numeric one.
//
// Parameters:
//   - `json`: The JSON string to parse.
//   - `path`: The path to evaluate against the JSON.
//   - `fn`: The function to call for each numeric value found.
//
// Returns:
//   - None
//
// Example Usage:
//
//	scanFloat64(json, "scores", func(n float64) { total += n })
//
// Notes:
//   - The function leverages recursive descent to explore nested JSON objects and arrays,
//     ensuring that all levels of the structure are searched for matches.
//   - If the `parent` element is an object or array, it will iterate over its elements and
//     perform recursive descent for each of them.
//   - The search is performed on the values of the JSON elements using `node.String()`.
//   - The `value` is checked for equality using `node.String() == value`.
func scanFloat64(json, path string, fn func(float64)) {
	ctx := Get(json, path)
	if !ctx.Exists() {
		return
	}
	if ctx.IsArray() {
		ctx.Foreach(func(_, item Context) bool {
			if item.kind == Number {
				fn(item.Float64())
			}
			return true
		})
		return
	}
	if ctx.kind == Number {
		fn(ctx.Float64())
	}
}

// formatNumber renders a float64 as a compact JSON number string.
// Integers are emitted without a decimal point.
//
// Parameters:
//   - `f`: The float64 value to format.
//
// Returns:
//   - A string containing the compact JSON number representation.
//
// Example Usage:
//
//	formatNumber(123.0) // "123"
//	formatNumber(123.45) // "123.45"
//	formatNumber(math.Inf(1)) // "+Inf"
func formatNumber(f float64) string {
	if f == math.Trunc(f) && !math.IsInf(f, 0) {
		return fmt.Sprintf("%d", int64(f))
	}
	return fmt.Sprintf("%g", f)
}

// matchesCondition tests whether `actual` satisfies `op` relative to `expected`.
// Numeric comparisons use float64; all others fall back to string comparison.
//
// Parameters:
//   - `actual`: The actual value to compare.
//   - `expected`: The expected value to compare against.
//   - `op`: The operator to use for comparison.
//
// Returns:
//   - `true` if the actual value satisfies the operator relative to the expected value,
//     `false` otherwise.
//
// Example Usage:
//
//	matchesCondition(actual, expected, "contains") // true if actual contains expected
//	matchesCondition(actual, expected, "ne") // true if actual is not equal to expected
//	matchesCondition(actual, expected, "gt") // true if actual is greater than expected
//	matchesCondition(actual, expected, "gte") // true if actual is greater than or equal to expected
//	matchesCondition(actual, expected, "lt") // true if actual is less than expected
//	matchesCondition(actual, expected, "lte") // true if actual is less than or equal to expected
//	matchesCondition(actual, expected, "eq") // true if actual is equal to expected
func matchesCondition(actual, expected Context, op string) bool {
	switch op {
	case "contains":
		return strings.Contains(actual.String(), expected.String())
	case "ne":
		return actual.raw != expected.raw
	case "gt", "gte", "lt", "lte":
		a, e := actual.Float64(), expected.Float64()
		switch op {
		case "gt":
			return a > e
		case "gte":
			return a >= e
		case "lt":
			return a < e
		case "lte":
			return a <= e
		}
	}
	// default: eq – raw JSON equality (works for strings, numbers, bools, null)
	return actual.raw == expected.raw
}
