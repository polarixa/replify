package encoding

import (
	"bytes"
	stdencoding "encoding"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// compactJSON processes a source byte slice and removes unwanted characters, returning a new cleaned-up byte slice.
//
// This function processes the input `src` byte slice and appends characters to the `dst` byte slice based on certain criteria.
// It specifically filters out characters that are not printable (i.e., characters with ASCII values greater than `' '`).
// Additionally, it handles quoted substrings, ensuring that characters inside properly escaped double quotes are preserved.
// If an unescaped double quote is encountered, it will stop processing further characters.
//
// Parameters:
//   - `dst`: The destination slice of bytes where the cleaned characters will be appended.
//   - `src`: The source slice of bytes to process, which may contain unwanted characters and quoted substrings.
//
// Returns:
//   - A new byte slice (`dst`) with unwanted characters removed. The cleaned-up version of `src`.
//
// Example:
//
//	src := []byte(`hello "world" 1234`)
//	dst := compactJSON([]byte{}, src)
//	// dst will be []byte{'h', 'e', 'l', 'l', 'o', ' ', '"', 'w', 'o', 'r', 'l', 'd', '"', ' ', '1', '2', '3', '4'},
//	// as the function preserves only printable characters and properly handles quoted substrings.
//
// Notes:
//   - This function skips characters that are not printable (less than or equal to ASCII ' ').
//   - When encountering a double quote (`"`), the function ensures that it correctly handles escaped quotes, skipping characters
//     until a valid closing quote is found. If an odd number of backslashes precede the closing quote, it breaks the loop to avoid
//     incorrect parsing of the quotes.
func compactJSON(dst, src []byte) []byte {
	dst = dst[:0] // Reset destination slice to an empty state
	for i := 0; i < len(src); i++ {
		if src[i] > ' ' { // Only include characters that are printable
			dst = append(dst, src[i])
			if src[i] == '"' { // Handle quoted substring (double quotes)
				for i = i + 1; i < len(src); i++ {
					dst = append(dst, src[i])
					if src[i] == '"' {
						// Search backwards for the last non-escaped backslash
						j := i - 1
						for ; ; j-- {
							if src[j] != '\\' {
								break
							}
						}
						// If the number of consecutive backslashes is odd, break the loop
						if (j-i)%2 != 0 {
							break
						}
					}
				}
			}
		}
	}
	return dst
}

// isNaNOrInf checks if a byte slice represents a special numeric value: NaN (Not-a-Number) or Infinity.
//
// This function inspects the first character of the input byte slice to determine if it represents
// either NaN or Infinity, including variations such as `Inf`, `+Inf`, `inf`, `NaN`, and `nan`.
// The function returns `true` if the input matches any of these special values.
//
// Parameters:
//   - `src`: A byte slice to inspect, typically a numeric string.
//
// Returns:
//   - `true` if the byte slice represents NaN or Infinity, `false` otherwise.
//
// Example:
//
//	src1 := []byte("Inf")
//	src2 := []byte("NaN")
//	src3 := []byte("+Inf")
//	src4 := []byte("infinity")
//	result1 := isNaNOrInf(src1) // result1 will be true
//	result2 := isNaNOrInf(src2) // result2 will be true
//	result3 := isNaNOrInf(src3) // result3 will be true
//	result4 := isNaNOrInf(src4) // result4 will be false
//
// Notes:
//   - The function only inspects the first character (or first two characters for lowercase `nan`) to make a quick determination.
//   - It supports the variations `Inf`, `+Inf`, `inf`, `NaN`, and `nan` as valid representations.
func isNaNOrInf(src []byte) bool {
	if len(src) == 0 {
		return false
	}
	return src[0] == 'i' || //Inf
		src[0] == 'I' || // inf
		src[0] == '+' || // +Inf
		src[0] == 'N' || // Nan
		(src[0] == 'n' && len(src) > 1 && src[1] != 'u') // nan
}

// getJSONType identifies the JSON type of a given byte slice based on its first character.
//
// This function analyzes the first character in the byte slice `v` to determine which JSON data type it represents.
// Based on the initial character, it categorizes the input as one of the following types:
// `jsonNull`, `jsonFalse`, `jsonTrue`, `jsonString`, `jNumber`, or `jsonJson` (indicating either a JSON object or array).
//
// Parameters:
//   - `v`: A byte slice representing a JSON value.
//
// Returns:
//   - A `jsonType` value that represents the JSON type of `v`, based on its first character.
//
// Example:
//
//	value1 := []byte(`"hello"`)
//	value2 := []byte("false")
//	value3 := []byte("123")
//	value4 := []byte("null")
//	value5 := []byte("[1, 2, 3]")
//	result1 := getJSONType(value1) // result1 will be jsonString
//	result2 := getJSONType(value2) // result2 will be jsonFalse
//	result3 := getJSONType(value3) // result3 will be jNumber
//	result4 := getJSONType(value4) // result4 will be jsonNull
//	result5 := getJSONType(value5) // result5 will be jsonJson
//
// Notes:
//   - If the byte slice is empty, the function returns `jsonNull`.
//   - The function uses the initial character of `v` to distinguish types, assuming `true`, `false`, and `null` are valid JSON values.
func getJSONType(v []byte) jsonType {
	if len(v) == 0 {
		return jsonNull
	}
	switch v[0] {
	case '"':
		return jsonString
	case 'f':
		return jsonFalse
	case 't':
		return jsonTrue
	case 'n':
		return jsonNull
	case '[', '{':
		return jsonJson
	default:
		return jNumber
	}
}

// unescapeJSONString extracts a JSON string from a byte slice, handling escaped characters when present.
//
// This function takes a JSON byte slice representing a string value and extracts the unescaped content.
// It iterates through the byte slice, checking for either escaped characters or the closing double quote (`"`).
// If an escape character (`\`) is detected, `unescapeJSONString` uses JSON unmarshalling to handle any escape sequences.
// If no escape character is encountered, it returns the substring between the opening and closing quotes.
//
// Parameters:
//   - `s`: A byte slice containing a JSON string, including the enclosing double quotes and potentially escaped characters.
//
// Returns:
//   - A byte slice with the unescaped content of the JSON string if valid, or `nil` if an error occurs.
//
// Example:
//
//	s := []byte(`"Hello, world!"`)
//	result := unescapeJSONString(s)
//	// result will be []byte{'H', 'e', 'l', 'l', 'o', ',', ' ', 'w', 'o', 'r', 'l', 'd', '!'}
//	s := []byte(`"Line1\nLine2"`)
//	result := unescapeJSONString(s)
//	// result will be []byte{'L', 'i', 'n', 'e', '1', '\n', 'L', 'i', 'n', 'e', '2'}
//
// Notes:
//   - If an escape sequence is encountered (`\`), JSON unmarshalling is used to correctly interpret it.
//   - If the byte slice does not contain a properly closed string, the function returns `nil`.
func unescapeJSONString(s []byte) []byte {
	for i := 1; i < len(s); i++ {
		if s[i] == '\\' {
			var str string
			if err := json.Unmarshal(s, &str); err != nil {
				return nil
			}
			return []byte(str)
		}
		if s[i] == '"' {
			return s[1:i]
		}
	}
	return nil
}

// sortObjectMembers sorts JSON key-value pairs in a stable order, preserving original formatting,
// and returns the updated buffer containing the sorted pairs.
//
// This function takes the JSON data, a buffer to store formatted values, and a list of key-value pairs (`pairs`).
// If there are no pairs to sort, it directly returns the buffer. Otherwise, it initializes a `byKeyVal` struct
// with the JSON data, buffer, and pairs, and sorts them by key (and by value if keys are identical) using
// `sort.Stable`. After sorting, it constructs a new byte slice with the sorted pairs in order, each followed
// by a comma and newline, and replaces the original content in `buf`.
//
// Parameters:
//   - `json`: The original JSON data as a byte slice.
//   - `buf`: A byte slice that holds formatted key-value pairs and can be modified in-place.
//   - `pairs`: A slice of `pair` structs representing the key-value pairs in the JSON data.
//
// Returns:
//   - A byte slice containing the buffer with sorted key-value pairs, in stable order.
//
// Example:
//
//	json := []byte(`{"b":2, "a":1}`)
//	pairs := []pair{ /* initialized with positions of keys and values */ }
//	buf := make([]byte, len(json))
//	result := sortObjectMembers(json, buf, pairs)
//	// result will contain sorted pairs by key.
//
// Notes:
//   - If `pairs` is empty, `buf` is returned unchanged.
//   - `sort.Stable` is used to ensure that pairs with identical keys maintain their original relative order.
//   - If `byKeyVal` marks pairs as unsorted, it skips replacing the original buffer.
func sortObjectMembers(json, buf []byte, pairs []kvPairs) []byte {
	if len(pairs) == 0 {
		return buf
	}
	_valStart := pairs[0].valueStart
	_valEnd := pairs[len(pairs)-1].valueEnd
	_keyVal := kvSorter{false, json, buf, pairs}
	sort.Stable(&_keyVal)
	if !_keyVal.sorted {
		return buf
	}
	n := make([]byte, 0, _valEnd-_valStart)
	for i, p := range pairs {
		n = append(n, buf[p.valueStart:p.valueEnd]...)
		if i < len(pairs)-1 {
			n = append(n, ',')
			n = append(n, '\n')
		}
	}
	return append(buf[:_valStart], n...)
}

// appendJSONString appends a JSON string value from the input JSON byte slice (`json`) to a buffer (`buf`),
// handling any escaped characters within the string, and returns the updated buffer and indices.
//
// This function begins at a given index `i` within a JSON byte slice `json`, assuming the current character is
// the start of a JSON string (`"`). It appends the entire string (from opening to closing quote) to the buffer `buf`,
// handling escaped quotes within the string. If an escape sequence (`\`) precedes a closing quote, it continues searching
// until it finds an unescaped closing quote, marking the end of the string.
//
// Parameters:
//   - `buf`: The destination byte slice to which the JSON string value is appended.
//   - `json`: The source JSON byte slice containing the entire JSON structure.
//   - `i`: The starting index in `json`, pointing to the beginning of the JSON string (initial quote).
//   - `nl`: The current newline position (used for pretty-printing in larger context).
//
// Returns:
//   - `buf`: The updated buffer, containing the appended JSON string.
//   - `i`: The updated index in `json` after the end of the string (right after the closing quote).
//   - `nl`: The unchanged newline position (for tracking in pretty-printing).
//   - `true`: A boolean flag indicating the function processed a string (for handling in other contexts).
//
// Example:
//
//	json := []byte(`"example \"string\" value"`)
//	buf := []byte{}
//	buf, i, nl, processed := appendJSONString(buf, json, 0, 0)
//	// buf will contain `example \"string\" value`, i will point to the next index after the closing quote,
//	// nl remains the same, and processed is true.
//
// Notes:
//   - The function counts consecutive backslashes before each closing quote to determine if it is escaped.
//   - It appends the entire string (including quotes) to `buf` for easy integration in pretty-printing or formatting routines.
func appendJSONString(buf, json []byte, i, nl int) ([]byte, int, int, bool) {
	s := i
	i++
	for ; i < len(json); i++ {
		if json[i] == '"' {
			var sc int
			for j := i - 1; j > s; j-- {
				if json[j] == '\\' {
					sc++
				} else {
					break
				}
			}
			if sc%2 == 1 {
				continue
			}
			i++
			break
		}
	}
	return append(buf, json[s:i]...), i, nl, true
}

// appendJSONNumber appends a JSON number value from the input JSON byte slice (`json`) to a buffer (`buf`),
// and returns the updated buffer and indices.
//
// This function starts at a given index `i` within a JSON byte slice `json` (assuming the current character is the start
// of a JSON number). It appends the entire number to the buffer `buf` and handles all characters up to the next
// non-number character, such as spaces, commas, colons, or closing brackets/braces.
//
// Parameters:
//   - `buf`: The destination byte slice to which the JSON number value will be appended.
//   - `json`: The source JSON byte slice containing the entire JSON structure.
//   - `i`: The starting index in `json`, pointing to the first character of the JSON number.
//   - `nl`: The current newline position (used for pretty-printing in larger context).
//
// Returns:
//   - `buf`: The updated buffer, containing the appended JSON number value.
//   - `i`: The updated index in `json` after the number (right after the last character of the number).
//   - `nl`: The unchanged newline position (for tracking in pretty-printing).
//   - `true`: A boolean flag indicating that a number was processed successfully.
//
// Example:
//
//	json := []byte(`12345`)
//	buf := []byte{}
//	buf, i, nl, processed := appendJSONNumber(buf, json, 0, 0)
//	// buf will contain `12345`, i will point to the next index after the number,
//	// nl remains unchanged, and processed will be true.
//
// Notes:
//   - The function scans for all characters that are part of the number (digits, decimal point, etc.) until it
//     encounters a character that is not part of a valid number, such as a space, comma, colon, or closing bracket/braces.
//   - It assumes that the number is well-formed and does not handle error cases like invalid numbers.
func appendJSONNumber(buf, json []byte, i, nl int) ([]byte, int, int, bool) {
	s := i // Record the start index of the number
	i++    // Move past the initial digit (or minus sign if present)
	for ; i < len(json); i++ {
		// Break the loop if a non-number character is encountered (e.g., space, comma, colon, bracket, or brace)
		if json[i] <= ' ' || json[i] == ',' || json[i] == ':' || json[i] == ']' || json[i] == '}' {
			break
		}
	}
	// Append the number from the start index `s` to the updated index `i` (excluding non-number characters)
	return append(buf, json[s:i]...), i, nl, true
}

// appendJSONContainer processes the next JSON object or array in the input JSON byte slice (`json`)
// and appends it to the buffer (`buf`), while handling pretty-printing, sorting of object keys, and enforcing width constraints.
//
// This function handles the parsing and formatting of JSON objects (`{}`) and arrays (`[]`), adding appropriate indentation,
// newlines, and sorting of keys (if specified). It ensures that the resulting object or array is correctly formatted and
// inserted into the buffer, maintaining the structure and respecting pretty-printing preferences.
//
// It also handles the optional width limit (to control the length of single-line arrays) and can process objects with sorted keys,
// ensuring proper formatting of both simple and complex JSON objects.
//
// Parameters:
//   - `buf`: The destination byte slice to which the processed JSON object or array will be appended.
//   - `json`: The source JSON byte slice containing the entire JSON structure.
//   - `i`: The starting index in `json` from where the next object or array should be processed.
//   - `open`: The opening byte (either '{' for an object or '[' for an array).
//   - `close`: The closing byte (either '}' for an object or ']' for an array).
//   - `pretty`: A boolean flag indicating whether pretty-printing should be applied (i.e., adding newlines and indentation).
//   - `width`: The width used for pretty-printing (influences line breaks for large arrays).
//   - `prefix`: A string prefix used for leading indentation (used in pretty-printing).
//   - `indent`: The string used for indentation in pretty-printing.
//   - `sortKeys`: A boolean flag indicating whether the keys in objects should be sorted.
//   - `tabs`: The number of tabs for indentation (used in pretty-printing).
//   - `nl`: The current newline position, used for managing where to insert newlines during pretty-printing.
//   - `max`: The maximum number of characters to pretty-print before breaking into a new line (relevant for width-based formatting).
//
// Returns:
//   - `buf`: The updated buffer containing the appended JSON object or array (pretty-printed if the `pretty` flag is true).
//   - `i`: The updated index in `json`, pointing to the position after the processed object or array.
//   - `nl`: The updated newline position, adjusted for pretty-printing.
//   - `true` or `false`: A boolean flag indicating whether the processing was successful. The function returns `false` if
//     there is an issue (for example, exceeding the width limit or malformed data).
//
// Example usage:
//
//	json := []byte(`{ "key1": 123, "key2": "value", "key3": [1, 2, 3] }`)
//	buf := []byte{}
//	buf, i, nl, processed := appendJSONContainer(buf, json, 0, '{', '}', true, 80, "", "  ", false, 0, 0, -1)
//	// buf will contain the pretty-printed JSON object, i will point to the next index,
//	// nl will be adjusted for newlines, and processed will be true.
//
// Notes:
//   - This function handles both JSON objects and arrays, depending on the value of `open` and `close` (either '{', '}' for objects
//     or '[' and ']' for arrays).
//   - Pretty-printing is applied based on the `pretty` flag, including indentation and line breaks.
//   - If `sortKeys` is set to true, the keys in the JSON object will be sorted lexicographically before being appended to the buffer.
//   - The `max` value helps control the number of characters in a single line for arrays, ensuring that arrays are properly wrapped into multiple lines if necessary.
//   - The function can also handle arrays and objects with nested structures, ensuring the formatting remains correct throughout.
func appendJSONContainer(buf, json []byte, i int, open, close byte, pretty bool, width int, prefix, indent string, sortKeys bool, tabs, nl, max int) ([]byte, int, int, bool) {
	var ok bool
	if width > 0 {
		if pretty && open == '[' && max == -1 {
			// here we try to create a single line array
			max := width - (len(buf) - nl)
			if max > 3 {
				s1, s2 := len(buf), i
				buf, i, _, ok = appendJSONContainer(buf, json, i, '[', ']', false, width, prefix, "", sortKeys, 0, 0, max)
				if ok && len(buf)-s1 <= max {
					return buf, i, nl, true
				}
				buf = buf[:s1]
				i = s2
			}
		} else if max != -1 && open == '{' {
			return buf, i, nl, false
		}
	}
	buf = append(buf, open)
	i++
	var pairs []kvPairs
	if open == '{' && sortKeys {
		pairs = make([]kvPairs, 0, 8)
	}
	var n int
	for ; i < len(json); i++ {
		if json[i] <= ' ' {
			continue
		}
		if json[i] == close {
			if pretty {
				if open == '{' && sortKeys {
					buf = sortObjectMembers(json, buf, pairs)
				}
				if n > 0 {
					nl = len(buf)
					if buf[nl-1] == ' ' {
						buf[nl-1] = '\n'
					} else {
						buf = append(buf, '\n')
					}
				}
				if buf[len(buf)-1] != open {
					buf = appendIndent(buf, prefix, indent, tabs)
				}
			}
			buf = append(buf, close)
			return buf, i + 1, nl, open != '{'
		}
		if open == '[' || json[i] == '"' {
			if n > 0 {
				buf = append(buf, ',')
				if width != -1 && open == '[' {
					buf = append(buf, ' ')
				}
			}
			var p kvPairs
			if pretty {
				nl = len(buf)
				if buf[nl-1] == ' ' {
					buf[nl-1] = '\n'
				} else {
					buf = append(buf, '\n')
				}
				if open == '{' && sortKeys {
					p.keyStart = i
					p.valueStart = len(buf)
				}
				buf = appendIndent(buf, prefix, indent, tabs+1)
			}
			if open == '{' {
				buf, i, nl, _ = appendJSONString(buf, json, i, nl)
				if sortKeys {
					p.keyEnd = i
				}
				buf = append(buf, ':')
				if pretty {
					buf = append(buf, ' ')
				}
			}
			buf, i, nl, ok = appendJSONValue(buf, json, i, pretty, width, prefix, indent, sortKeys, tabs+1, nl, max)
			if max != -1 && !ok {
				return buf, i, nl, false
			}
			if pretty && open == '{' && sortKeys {
				p.valueEnd = len(buf)
				if p.keyStart > p.keyEnd || p.valueStart > p.valueEnd {
					// bad data. disable sorting
					sortKeys = false
				} else {
					pairs = append(pairs, p)
				}
			}
			i--
			n++
		}
	}
	return buf, i, nl, open != '{'
}

// appendJSONByte appends a byte `c` to the destination byte slice `dst`,
// handling special control characters by escaping them in Unicode format.
// If `c` is a control character (ASCII value less than 32) and not one of the
// common whitespace characters (`\r`, `\n`, `\t`, `\v`), it appends
// the Unicode escape sequence `\u00XX` to `dst`, where `XX` is the hexadecimal
// representation of `c`. Otherwise, it appends `c` directly to `dst`.
func appendJSONByte(dst []byte, c byte) []byte {
	if c < ' ' && (c != '\r' && c != '\n' && c != '\t' && c != '\v') {
		dst = append(dst, "\\u00"...)
		dst = append(dst, hexDigit((c>>4)&0xF))
		return append(dst, hexDigit(c&0xF))
	}
	return append(dst, c)
}

// appendJSONValue processes the next JSON value in the input JSON byte slice (`json`) and appends it to the buffer (`buf`),
// while handling different types of JSON values (strings, numbers, objects, arrays, and literals).
// It returns the updated buffer and indices, as well as a boolean flag indicating whether a value was processed.
//
// This function is responsible for recognizing the type of the next JSON value and delegating the task of appending that value
// to the appropriate helper function (such as `appendPrettyString` for strings, `appendPrettyNumber` for numbers, and others
// for objects, arrays, and literals). It processes the JSON byte slice one element at a time and handles all value types
// correctly, ensuring that each value is pretty-printed if required.
//
// Parameters:
//   - `buf`: The destination byte slice to which the processed JSON value will be appended.
//   - `json`: The source JSON byte slice containing the entire JSON structure.
//   - `i`: The starting index in `json` from where the next value should be processed.
//   - `pretty`: A boolean flag indicating whether pretty-printing should be applied (i.e., adding newlines and indentation).
//   - `width`: The width used for pretty-printing (not used in this function, but passed for consistency in pretty-printing logic).
//   - `prefix`: A string prefix (used in pretty-printing to add leading indentation, not used here).
//   - `indent`: The string used for indentation (not used here but part of the pretty-printing configuration).
//   - `sortKeys`: A boolean flag indicating whether the keys in objects should be sorted (not used here but passed for consistency).
//   - `tabs`: The number of tabs for indentation (not used here but passed for consistency in pretty-printing logic).
//   - `nl`: The current newline position (used for pretty-printing, ensuring line breaks are maintained correctly).
//   - `max`: The maximum number of characters to pretty-print before breaking into a new line (not used in this function).
//
// Returns:
//   - `buf`: The updated buffer, containing the appended JSON value (pretty-printed if the `pretty` flag is true).
//   - `i`: The updated index in `json`, pointing to the position after the processed value.
//   - `nl`: The unchanged newline position (used for pretty-printing in the larger context).
//   - `true`: A boolean flag indicating that a JSON value was successfully processed.
//
// Example usage:
//
//	json := []byte(`{ "key1": 123, "key2": "value", "key3": [1, 2, 3] }`)
//	buf := []byte{}
//	buf, i, nl, processed := appendJSONValue(buf, json, 0, true, 0, "", "  ", false, 0, 0, 0)
//	// buf will contain the pretty-printed JSON value, i will point to the next index,
//	// nl remains unchanged, and processed will be true.
//
// Notes:
//   - This function processes and appends various JSON data types, including strings, numbers, objects, arrays,
//     and literals (`true`, `false`, `null`).
//   - The function assumes the JSON is valid and well-formed; it does not handle parsing errors for invalid JSON.
func appendJSONValue(buf, json []byte, i int, pretty bool, width int, prefix, indent string, sortKeys bool, tabs, nl, max int) ([]byte, int, int, bool) {
	for ; i < len(json); i++ {
		if json[i] <= ' ' {
			continue
		}
		if json[i] == '"' {
			return appendJSONString(buf, json, i, nl)
		}
		if (json[i] >= '0' && json[i] <= '9') || json[i] == '-' || isNaNOrInf(json[i:]) {
			return appendJSONNumber(buf, json, i, nl)
		}
		if json[i] == '{' {
			return appendJSONContainer(buf, json, i, '{', '}', pretty, width, prefix, indent, sortKeys, tabs, nl, max)
		}
		if json[i] == '[' {
			return appendJSONContainer(buf, json, i, '[', ']', pretty, width, prefix, indent, sortKeys, tabs, nl, max)
		}
		switch json[i] {
		case 't':
			return append(buf, 't', 'r', 'u', 'e'), i + 4, nl, true
		case 'f':
			return append(buf, 'f', 'a', 'l', 's', 'e'), i + 5, nl, true
		case 'n':
			return append(buf, 'n', 'u', 'l', 'l'), i + 4, nl, true
		}
	}
	return buf, i, nl, true
}

// appendIndent appends indentation to the provided buffer (`buf`) based on the specified `prefix`, `indent`,
// and the number of `tabs` to insert.
//
// This function adds a specific number of tab or space-based indents to the `buf`, depending on the `indent` value.
// If the `indent` string contains exactly two spaces, it will append spaces for each tab; otherwise, it uses
// the provided `indent` string (which can represent any indenting character, such as tabs or custom strings).
// Additionally, if a `prefix` is provided, it will be prepended to the buffer before any indentation is added.
//
// Parameters:
//   - `buf`: The byte slice to which the indentations are appended.
//   - `prefix`: A string (byte slice) that will be prepended to `buf` before any indentation (optional).
//   - `indent`: The string to use for indentation, typically consisting of spaces or tabs (e.g., `"\t"` or `"  "`).
//   - `tabs`: The number of times the `indent` should be repeated to represent the desired level of indentation.
//
// Returns:
//   - The updated `buf` with the appropriate amount of indentation based on `tabs` and `indent`.
//
// Example:
//
//	buf := []byte{}
//	prefix := "  "
//	indent := "\t"
//	tabs := 3
//	buf = appendIndent(buf, prefix, indent, tabs)
//	// buf will be `{"  "\t\t\t` (prefix followed by 3 tab characters).
//
// Notes:
//   - If the `indent` string is exactly two spaces (`"  "`), the function will append two spaces for each `tab`.
//   - If the `indent` string is anything else, it will be appended `tabs` times.
func appendIndent(buf []byte, prefix, indent string, tabs int) []byte {
	if len(prefix) != 0 { // Append prefix if it's not an empty string
		buf = append(buf, prefix...)
	}
	// Check if the indent is exactly two spaces and append spaces for each tab
	if len(indent) == 2 && indent[0] == ' ' && indent[1] == ' ' {
		for i := 0; i < tabs; i++ {
			buf = append(buf, ' ', ' ')
		}
	} else {
		// Otherwise, append the custom indent string for each tab
		for i := 0; i < tabs; i++ {
			buf = append(buf, indent...)
		}
	}
	return buf
}

// hexDigit converts a numeric value to its corresponding hexadecimal character.
//
// This function takes a single byte `p`, which represents a numeric value, and converts it to its
// hexadecimal equivalent as a byte. The function assumes that the input `p` is in the range of 0 to 15
// (i.e., it represents a single hexadecimal digit).
//
// Parameters:
//   - `p`: A byte representing a numeric value between 0 and 15 (inclusive).
//
// Returns:
//   - A byte representing the corresponding hexadecimal character.
//   - If `p` is less than 10, it returns the ASCII character for the corresponding digit (0-9).
//   - If `p` is 10 or greater, it returns the lowercase ASCII character for the corresponding letter (a-f).
//
// Example:
//
//	hexDigit(0)  // returns '0'
//	hexDigit(9)  // returns '9'
//	hexDigit(10) // returns 'a'
//	hexDigit(15) // returns 'f'
//
// Notes:
//   - The function works only for values between 0 and 15 (inclusive).
//   - For input values greater than 15, the behavior is not defined and may lead to unexpected results.
func hexDigit(p byte) byte {
	// If p is less than 10, return the corresponding digit character ('0' to '9')
	switch {
	case p < 10:
		return p + '0' // Add ASCII value of '0' to convert to character
	default:
		// If p is 10 or greater, return the corresponding letter character ('a' to 'f')
		// Add ASCII value of 'a' to get 'a' to 'f'
		return (p - 10) + 'a'
	}
}

// sanitizeJSON processes a source byte slice (source) and removes or replaces comment sections,
// formatting them into a cleaned-up destination byte slice (destination).
// It handles both single-line (`//`) and multi-line (`/* */`) comments and strips them out,
// replacing them with spaces or newlines as appropriate.
//
// Parameters:
//   - `source`: The input byte slice containing the source code to process.
//   - `destination`: The output byte slice that will contain the cleaned code (without comments).
//     It is assumed to be initialized as an empty slice.
//
// Returns:
//   - A byte slice representing the cleaned source code with comments removed or replaced.
//   - Single-line comments (`//`) are replaced with spaces.
//   - Multi-line comments (`/* */`) are replaced with spaces and preserved newlines for line breaks.
//
// Example:
//
//	source := []byte("int x = 10; // initialize x\n/* multi-line\n comment */")
//	destination := sanitizeJSON(source, []byte{})
//	// destination will contain: "int x = 10;    \n   "
//
// Notes:
//   - The function handles the following types of comments:
//   - Single-line comments starting with `//` and ending with the newline (`\n`).
//   - Multi-line comments enclosed in `/* */`, even if they span multiple lines.
//   - Strings inside double quotes (`"`) and special characters like `}` or `]` are preserved as is,
//     with careful handling of quotes and escape sequences within the string.
func sanitizeJSON(source, destination []byte) []byte {
	destination = destination[:0]
	for i := 0; i < len(source); i++ {
		if source[i] == '/' {
			if i < len(source)-1 {
				if source[i+1] == '/' {
					destination = append(destination, ' ', ' ')
					i += 2
					for ; i < len(source); i++ {
						if source[i] == '\n' {
							destination = append(destination, '\n')
							break
						} else if source[i] == '\t' || source[i] == '\r' {
							destination = append(destination, source[i])
						} else {
							destination = append(destination, ' ')
						}
					}
					continue
				}
				if source[i+1] == '*' {
					destination = append(destination, ' ', ' ')
					i += 2
					for ; i < len(source)-1; i++ {
						if source[i] == '*' && source[i+1] == '/' {
							destination = append(destination, ' ', ' ')
							i++
							break
						} else if source[i] == '\n' || source[i] == '\t' ||
							source[i] == '\r' {
							destination = append(destination, source[i])
						} else {
							destination = append(destination, ' ')
						}
					}
					continue
				}
			}
		}
		destination = append(destination, source[i])
		if source[i] == '"' {
			for i = i + 1; i < len(source); i++ {
				destination = append(destination, source[i])
				if source[i] == '"' {
					j := i - 1
					for ; ; j-- {
						if source[j] != '\\' {
							break
						}
					}
					if (j-i)%2 != 0 {
						break
					}
				}
			}
		} else if source[i] == '}' || source[i] == ']' {
			for j := len(destination) - 2; j >= 0; j-- {
				if destination[j] <= ' ' {
					continue
				}
				if destination[j] == ',' {
					destination[j] = ' '
				}
				break
			}
		}
	}
	return destination
}

// Pre-computed interface types used for custom-marshaler detection.
var (
	// jsonMarshalerType is the reflect.Type of the json.Marshaler interface.
	jsonMarshalerType = reflect.TypeFor[json.Marshaler]()

	// textMarshalerType is the reflect.Type of the encoding.TextMarshaler interface.
	textMarshalerType = reflect.TypeFor[stdencoding.TextMarshaler]()

	// complexTypeCache memoises containsComplex results keyed by reflect.Type
	// to avoid repeated full-type-tree walks.
	complexTypeCache sync.Map
)

// marshalJSONCommon is the single implementation shared by [JSON] and [JSONE].
//
// Parameters:
//   - data       - the value to encode.
//   - pretty     - when true the output is indented with 4 spaces.
//   - errorOnNil - when true a nil top-level value returns [ErrNilInterface]
//     instead of an empty string.
//
// Returns:
//   - string - the JSON representation.
//   - error  - non-nil on encoding failure.
func marshalJSONCommon(data any, pretty, errorOnNil bool) (string, error) {
	if data == nil {
		if errorOnNil {
			return "", ErrNilInterface
		}
		return "", nil
	}

	// Fast-path well-known types that do not need reflection.
	switch v := data.(type) {
	case string:
		return marshalStringToken(v)
	case *string:
		if v == nil {
			return "null", nil
		}
		return marshalStringToken(*v)
	case json.RawMessage:
		return marshalRawMessageToken(v, pretty)
	case *json.RawMessage:
		if v == nil {
			return "null", nil
		}
		return marshalRawMessageToken(*v, pretty)
	}

	rv, ok := unwrapInterfaces(reflect.ValueOf(data))
	if !ok {
		if errorOnNil {
			return "", ErrNilInterface
		}
		return "", nil
	}

	if isNilValue(rv) {
		return "null", nil
	}

	// Scalars (bool, int*, uint*, float*, complex*, string) are handled
	// without falling into the heavier marshal path.
	if token, handled, err := scalarJSONToken(rv); handled {
		return token, err
	}

	return safeMarshalJSONString(rv.Interface(), pretty)
}

// marshalStringToken JSON-encodes a Go string value.
//
// The result includes the surrounding double-quote characters and all
// necessary escape sequences as required by the JSON specification.
//
// Parameters:
//   - s - the raw Go string to encode.
//
// Returns:
//   - string - the quoted, escaped JSON string (e.g. `"hello\nworld"`).
//   - error  - non-nil if json.Marshal fails (extremely rare for strings).
//
// Example:
//
//	marshalStringToken("hi")        // `"hi"`, nil
//	marshalStringToken("a\tb")      // `"a\tb"`, nil
//	marshalStringToken("")          // `""`, nil
func marshalStringToken(s string) (string, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// marshalRawMessageToken validates and returns a [json.RawMessage] as a string,
// optionally pretty-printing it.
//
// Parameters:
//   - msg    - the raw JSON bytes to validate and return.
//   - pretty - when true the bytes are re-indented with 4 spaces.
//
// Returns:
//   - string - the (optionally indented) JSON string.
//   - error  - [ErrInvalidRawMessage] if msg is not valid JSON,
//     or a json.Indent error on pretty-print failure.
//
// Example:
//
//	marshalRawMessageToken(json.RawMessage(`{"a":1}`), false) // `{"a":1}`, nil
//	marshalRawMessageToken(json.RawMessage(`{"a":1}`), true)  // "{\n    \"a\": 1\n}", nil
//	marshalRawMessageToken(json.RawMessage(`{bad}`),  false)  // "", ErrInvalidRawMessage
//	marshalRawMessageToken(nil, false)                        // "null", nil
func marshalRawMessageToken(msg json.RawMessage, pretty bool) (string, error) {
	if msg == nil {
		return "null", nil
	}
	if !json.Valid(msg) {
		return "", ErrInvalidRawMessage
	}
	if !pretty {
		return string(msg), nil
	}

	var buf bytes.Buffer
	if err := json.Indent(&buf, msg, "", "    "); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// unwrapInterfaces peels away one or more consecutive interface layers from v.
//
// Unwrapping stops as soon as a non-interface kind is reached. If any
// intermediate interface value is nil, or if v is itself invalid, the
// function signals failure via the boolean return.
//
// Parameters:
//   - v - the reflect.Value to unwrap; may be of any kind.
//
// Returns:
//   - reflect.Value - the innermost non-interface value.
//   - bool          - true on success; false when v is invalid or a nil
//     interface is encountered during unwrapping.
//
// Example:
//
//	var i any = 42
//	rv, ok := unwrapInterfaces(reflect.ValueOf(&i).Elem()) // reflect.Value(42), true
//
//	var nilI any
//	rv, ok := unwrapInterfaces(reflect.ValueOf(&nilI).Elem()) // zero, false
func unwrapInterfaces(v reflect.Value) (reflect.Value, bool) {
	for v.IsValid() && v.Kind() == reflect.Interface {
		if v.IsNil() {
			return reflect.Value{}, false
		}
		v = v.Elem()
	}
	if !v.IsValid() {
		return reflect.Value{}, false
	}
	return v, true
}

// isNilValue reports whether v holds a nil pointer, map, slice, func, or
// channel. All other kinds (including non-nilable scalars and structs) return
// false.
//
// Parameters:
//   - v - any valid reflect.Value.
//
// Returns:
//   - bool - true when v is a nilable kind whose current value is nil.
//
// Example:
//
//	isNilValue(reflect.ValueOf((*int)(nil)))    // true
//	isNilValue(reflect.ValueOf([]int(nil)))     // true
//	isNilValue(reflect.ValueOf([]int{}))        // false
//	isNilValue(reflect.ValueOf(0))              // false
func isNilValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Func, reflect.Chan:
		return v.IsNil()
	default:
		return false
	}
}

// scalarJSONToken converts a scalar reflect.Value to its JSON token string
// without allocating a full marshal round-trip.
//
// Handled kinds: String, Bool, Int*, Uint*, Uintptr, Float32/64,
// Complex64/128. For all other kinds the function returns ("", false, nil)
// to tell the caller to fall back to the full marshal path.
//
// Parameters:
//   - v - a reflect.Value of any kind.
//
// Returns:
//   - token   - the JSON token string (e.g. "true", "42", `"hello"`).
//   - handled - true when the kind was recognised and token is valid.
//   - err     - non-nil when the kind was recognised but encoding failed
//     (e.g. [ErrNonFiniteFloat] for NaN/Inf).
//
// Example:
//
//	scalarJSONToken(reflect.ValueOf(true))        // "true",    true,  nil
//	scalarJSONToken(reflect.ValueOf(-7))          // "-7",      true,  nil
//	scalarJSONToken(reflect.ValueOf(math.NaN()))  // "",        true,  ErrNonFiniteFloat
//	scalarJSONToken(reflect.ValueOf([]int{1}))    // "",        false, nil
func scalarJSONToken(v reflect.Value) (string, bool, error) {
	switch v.Kind() {
	case reflect.String:
		b, err := json.Marshal(v.String())
		return string(b), true, err

	case reflect.Bool:
		return strconv.FormatBool(v.Bool()), true, nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), true, nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10), true, nil

	case reflect.Uintptr:
		// Rendered as a quoted hex string so it is always a valid JSON value.
		return fmt.Sprintf("%q", fmt.Sprintf("0x%x", v.Uint())), true, nil

	case reflect.Float32, reflect.Float64:
		token, _, err := formatFloatToken(v.Float(), v.Kind() == reflect.Float32)
		return token, true, err

	case reflect.Complex64, reflect.Complex128:
		token, err := encodeComplexJSONToken(v)
		return token, true, err

	default:
		return "", false, nil
	}
}

// safeMarshalJSONString is a panic-safe wrapper around [marshalWithComplexFallback].
//
// Some custom [json.Marshaler] implementations may panic instead of returning
// an error. This function recovers from any such panic and converts it into an
// error wrapping [ErrMarshalPanicRecovered].
//
// Parameters:
//   - v      - the value to marshal; must already be unwrapped from interface.
//   - pretty - when true the output is indented with 4 spaces.
//
// Returns:
//   - string - the JSON string, or "" on panic/error.
//   - error  - non-nil when encoding fails or a panic is caught.
//
// Example:
//
//	safeMarshalJSONString(struct{ X int }{3}, false) // `{"X":3}`, nil
//	// (panic inside a custom marshaler is recovered and returned as an error)
func safeMarshalJSONString(v any, pretty bool) (as string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%w: %v", ErrMarshalPanicRecovered, r)
			as = ""
		}
	}()
	return marshalWithComplexFallback(v, pretty)
}

// marshalWithComplexFallback marshals v to a JSON string, automatically
// switching to the reflection-based complex encoder when the standard library
// refuses the value because it contains complex numbers.
//
// The fast path uses [MarshalJSON] / [MarshalJSONIndent] (stdlib-backed).
// If those return a [*json.UnsupportedTypeError] caused by a complex kind,
// the function retries with [encodeValueString].
//
// Parameters:
//   - v      - the value to marshal.
//   - pretty - when true output is indented with 4 spaces.
//
// Returns:
//   - string - the JSON representation.
//   - error  - non-nil on encoding failure.
//
// Example:
//
//	marshalWithComplexFallback(42, false)                  // "42", nil
//	marshalWithComplexFallback(complex(1, 2), false)       // `{"real":1,"imag":2}`, nil
//	marshalWithComplexFallback(struct{ X int }{7}, true)   // "{\n    \"X\": 7\n}", nil
func marshalWithComplexFallback(v any, pretty bool) (string, error) {
	rv := reflect.ValueOf(v)

	// Skip the stdlib entirely when we already know the type contains complex
	// numbers; it would just return an unsupported-type error.
	if rv.IsValid() && containsComplex(rv.Type()) {
		return encodeValueString(rv, pretty)
	}

	var (
		b   []byte
		err error
	)
	if pretty {
		b, err = MarshalJSONIndent(v, "", "    ")
	} else {
		b, err = MarshalJSON(v)
	}
	if err == nil {
		return string(b), nil
	}

	if !shouldFallbackToComplexEncoder(err) {
		return "", err
	}

	return encodeValueString(rv, pretty)
}

// shouldFallbackToComplexEncoder reports whether a json.Marshal error was
// caused by a complex number type that the custom encoder can handle.
//
// It returns true only when err is a [*json.UnsupportedTypeError] whose
// underlying type is complex64, complex128, or a composite type that
// contains one of those kinds.
//
// Parameters:
//   - err - the error returned by json.Marshal or json.MarshalIndent.
//
// Returns:
//   - bool - true when the error is recoverable by [encodeValueString].
//
// Example:
//
//	_, err := json.Marshal(complex(1, 2))
//	shouldFallbackToComplexEncoder(err) // true
//
//	_, err = json.Marshal(make(chan int))
//	shouldFallbackToComplexEncoder(err) // false
func shouldFallbackToComplexEncoder(err error) bool {
	ute, ok := err.(*json.UnsupportedTypeError)
	if !ok || ute.Type == nil {
		return false
	}
	t := ute.Type
	return t.Kind() == reflect.Complex64 || t.Kind() == reflect.Complex128 || containsComplex(t)
}

// encodeValueString encodes a reflect.Value using [encodeValue] and
// optionally pretty-prints the raw JSON bytes.
//
// Parameters:
//   - v      - a reflect.Value to encode (any kind supported by [encodeValue]).
//   - pretty - when true the output is re-indented with 4 spaces via
//     [json.Indent].
//
// Returns:
//   - string - the JSON string.
//   - error  - non-nil when [encodeValue] or [json.Indent] fails.
//
// Example:
//
//	encodeValueString(reflect.ValueOf([]int{1, 2}), false) // "[1,2]", nil
//	encodeValueString(reflect.ValueOf([]int{1, 2}), true)  // "[\n    1,\n    2\n]", nil
func encodeValueString(v reflect.Value, pretty bool) (string, error) {
	b, err := encodeValue(v)
	if err != nil {
		return "", err
	}
	if !pretty {
		return string(b), nil
	}

	var indented bytes.Buffer
	if err := json.Indent(&indented, b, "", "    "); err != nil {
		return "", err
	}
	return indented.String(), nil
}

// containsComplex reports (with memoisation via [complexTypeCache]) whether
// type t is or transitively contains a complex64 or complex128 kind.
//
// Results are cached in complexTypeCache so repeated lookups for the same
// type are O(1).
//
// Parameters:
//   - t - the reflect.Type to inspect; nil returns false.
//
// Returns:
//   - bool - true when t is or contains a complex numeric kind.
//
// Example:
//
//	containsComplex(reflect.TypeOf(complex128(0)))          // true
//	containsComplex(reflect.TypeOf(struct{ C complex64 }{})) // true
//	containsComplex(reflect.TypeOf(42))                      // false
//	containsComplex(nil)                                     // false
func containsComplex(t reflect.Type) bool {
	if t == nil {
		return false
	}
	if cached, ok := complexTypeCache.Load(t); ok {
		return cached.(bool)
	}
	result := typeContainsComplex(t, make(map[reflect.Type]bool))
	complexTypeCache.Store(t, result)
	return result
}

// typeContainsComplex is the recursive worker for [containsComplex].
//
// It walks the full type tree, using seen to prevent infinite loops on
// self-referential struct types (e.g. linked-list nodes).
//
// Parameters:
//   - t    - the reflect.Type to inspect.
//   - seen - a set of already-visited struct types; prevents cycles.
//
// Returns:
//   - bool - true when t or any reachable sub-type is complex64/128.
//
// Example:
//
//	typeContainsComplex(reflect.TypeOf([]complex64{}), map[reflect.Type]bool{}) // true
//	typeContainsComplex(reflect.TypeOf(0),             map[reflect.Type]bool{}) // false
func typeContainsComplex(t reflect.Type, seen map[reflect.Type]bool) bool {
	switch t.Kind() {
	case reflect.Complex64, reflect.Complex128:
		return true

	case reflect.Ptr, reflect.Slice, reflect.Array, reflect.Chan:
		return typeContainsComplex(t.Elem(), seen)

	case reflect.Map:
		return typeContainsComplex(t.Key(), seen) || typeContainsComplex(t.Elem(), seen)

	case reflect.Struct:
		if seen[t] {
			return false
		}
		seen[t] = true
		for i := 0; i < t.NumField(); i++ {
			if typeContainsComplex(t.Field(i).Type, seen) {
				return true
			}
		}
	}
	return false
}

// encodeValue is the core recursive encoder used by the complex-number fallback
// path and all composite-type helpers.
//
// It first unwraps any interface/pointer indirection layers (returning "null"
// for nil pointers or interfaces), then checks for custom marshalers via
// [marshalViaCustomMarshaler], and finally dispatches by kind:
//
//   - complex64/128  → [encodeComplexBytes]
//   - struct         → [encodeStruct]
//   - slice/array    → [encodeSlice]
//   - map            → [encodeMap]
//   - everything else → json.Marshal (stdlib)
//
// Parameters:
//   - v - the reflect.Value to encode; any kind is accepted.
//
// Returns:
//   - []byte - the raw JSON bytes.
//   - error  - non-nil when encoding fails.
//
// Example:
//
//	encodeValue(reflect.ValueOf(42))              // []byte("42"), nil
//	encodeValue(reflect.ValueOf((*int)(nil)))     // []byte("null"), nil
//	encodeValue(reflect.ValueOf(complex(3, 4)))   // []byte(`{"real":3,"imag":4}`), nil
func encodeValue(v reflect.Value) ([]byte, error) {
	// Unwrap interfaces and pointers in a single unified loop.
	for v.IsValid() {
		switch v.Kind() {
		case reflect.Interface:
			if v.IsNil() {
				return []byte("null"), nil
			}
			v = v.Elem()
			continue
		case reflect.Ptr:
			if v.IsNil() {
				return []byte("null"), nil
			}
			v = v.Elem()
			continue
		}
		break
	}
	if !v.IsValid() {
		return []byte("null"), nil
	}

	// Custom marshalers take precedence over built-in logic.
	if b, ok, err := marshalViaCustomMarshaler(v); ok {
		return b, err
	}

	switch v.Kind() {
	case reflect.Complex64, reflect.Complex128:
		return encodeComplexBytes(v)

	case reflect.Struct:
		return encodeStruct(v)

	case reflect.Slice:
		if v.IsNil() {
			return []byte("null"), nil
		}
		return encodeSlice(v)

	case reflect.Array:
		return encodeSlice(v)

	case reflect.Map:
		if v.IsNil() {
			return []byte("null"), nil
		}
		return encodeMap(v)

	default:
		return json.Marshal(v.Interface())
	}
}

// marshalViaCustomMarshaler checks whether v (or its address, when
// addressable) implements [json.Marshaler] or [encoding.TextMarshaler] and
// invokes the appropriate method.
//
// Priority order:
//  1. json.Marshaler - value receiver
//  2. json.Marshaler - pointer receiver (requires v.CanAddr())
//  3. encoding.TextMarshaler - value receiver (output is JSON-quoted)
//  4. encoding.TextMarshaler - pointer receiver (requires v.CanAddr())
//
// For TextMarshaler the raw text bytes are wrapped in a JSON string via
// json.Marshal so the result is always valid JSON.
//
// Parameters:
//   - v - the reflect.Value to inspect and possibly marshal.
//
// Returns:
//   - []byte - the marshalled bytes when a marshaler was found.
//   - bool   - true when a marshaler was found (regardless of error).
//   - error  - non-nil when the marshaler returned an error.
//
// Example:
//
//	// Given a type implementing json.Marshaler:
//	marshalViaCustomMarshaler(reflect.ValueOf(myType{})) // <bytes>, true, nil
//
//	// Given a plain int:
//	marshalViaCustomMarshaler(reflect.ValueOf(42)) // nil, false, nil
func marshalViaCustomMarshaler(v reflect.Value) ([]byte, bool, error) {
	t := v.Type()

	// json.Marshaler - value receiver.
	if t.Implements(jsonMarshalerType) {
		b, err := json.Marshal(v.Interface())
		return b, true, err
	}
	// json.Marshaler - pointer receiver.
	if v.CanAddr() && v.Addr().Type().Implements(jsonMarshalerType) {
		b, err := json.Marshal(v.Addr().Interface())
		return b, true, err
	}

	// encoding.TextMarshaler - value receiver.
	if t.Implements(textMarshalerType) {
		b, err := v.Interface().(stdencoding.TextMarshaler).MarshalText()
		if err != nil {
			return nil, true, err
		}
		quoted, qerr := json.Marshal(string(b))
		return quoted, true, qerr
	}
	// encoding.TextMarshaler - pointer receiver.
	if v.CanAddr() && v.Addr().Type().Implements(textMarshalerType) {
		b, err := v.Addr().Interface().(stdencoding.TextMarshaler).MarshalText()
		if err != nil {
			return nil, true, err
		}
		quoted, qerr := json.Marshal(string(b))
		return quoted, true, qerr
	}

	return nil, false, nil
}

// parseJSONTag parses the `json:"…"` struct tag on sf and returns the
// resolved field name together with encoding options.
//
// Behaviour matches the encoding/json package:
//   - No tag       → use the field name, no omitempty, not skipped.
//   - Tag "-"      → field is skipped entirely.
//   - Tag "name"   → use "name".
//   - Tag ",opts"  → use the field name, apply opts.
//   - Tag "name,opts" → use "name", apply opts.
//
// Parameters:
//   - sf - the reflect.StructField whose tag is to be parsed.
//
// Returns:
//   - name      - the JSON key for this field.
//   - omitempty - true when the "omitempty" option is present.
//   - skip      - true when the tag value is exactly "-".
//
// Example:
//
//	// Field tagged `json:"id,omitempty"`
//	parseJSONTag(sf) // "id", true, false
//
//	// Field tagged `json:"-"`
//	parseJSONTag(sf) // "", false, true
//
//	// Field with no json tag
//	parseJSONTag(sf) // "<FieldName>", false, false
func parseJSONTag(sf reflect.StructField) (name string, omitempty, skip bool) {
	tag := sf.Tag.Get("json")
	if tag == "" {
		return sf.Name, false, false
	}
	if tag == "-" {
		return "", false, true
	}

	before, after, hasSep := strings.Cut(tag, ",")
	if !hasSep {
		return tag, false, false
	}

	name = before
	if name == "" {
		name = sf.Name
	}
	omitempty = strings.Contains(after, "omitempty")
	return name, omitempty, false
}

// shouldPromoteAnonymousStruct reports whether an anonymous (embedded) struct
// field should have its own fields inlined ("promoted") into the parent JSON
// object rather than being nested under a key.
//
// Promotion is performed when all of the following are true:
//   - sf.Anonymous is true.
//   - The json tag is not "-".
//   - The json tag does not provide an explicit name (only options are allowed,
//     e.g. ",omitempty").
//   - The concrete value reached after pointer dereferences is a struct.
//
// Parameters:
//   - sf - the reflect.StructField describing the embedded field.
//   - fv - the reflect.Value of that field in the parent struct.
//
// Returns:
//   - reflect.Value - the dereferenced struct value to inline, when promotion
//     applies.
//   - bool          - true when the field should be promoted.
//
// Example:
//
//	type Inner struct{ X int }
//	type Outer struct{ Inner }      // promoted  → true
//	type Outer2 struct {
//	    Inner `json:"inner"` }      // explicit name → false
//	type Outer3 struct {
//	    Inner `json:",omitempty"` } // options only → true (promoted)
func shouldPromoteAnonymousStruct(sf reflect.StructField, fv reflect.Value) (reflect.Value, bool) {
	if !sf.Anonymous {
		return reflect.Value{}, false
	}

	tag := sf.Tag.Get("json")
	if tag == "-" {
		return reflect.Value{}, false
	}
	// Text before the first comma means an explicit name → no promotion.
	if idx := strings.IndexByte(tag, ','); idx > 0 {
		return reflect.Value{}, false
	}

	w := fv
	for w.IsValid() && w.Kind() == reflect.Ptr {
		if w.IsNil() {
			return reflect.Value{}, false
		}
		w = w.Elem()
	}
	if !w.IsValid() || w.Kind() != reflect.Struct {
		return reflect.Value{}, false
	}

	return w, true
}

// isEmptyJSONValue reports whether v is considered "empty" for the purpose of
// the "omitempty" json tag option.
//
// The definition matches the encoding/json package:
//   - false, 0, 0.0, 0+0i, nil pointer/interface, empty string/slice/map/array.
//
// Parameters:
//   - v - any valid reflect.Value.
//
// Returns:
//   - bool - true when v should be omitted.
//
// Example:
//
//	isEmptyJSONValue(reflect.ValueOf(""))          // true
//	isEmptyJSONValue(reflect.ValueOf("x"))         // false
//	isEmptyJSONValue(reflect.ValueOf(0))           // true
//	isEmptyJSONValue(reflect.ValueOf(1))           // false
//	isEmptyJSONValue(reflect.ValueOf([]int(nil)))  // true
//	isEmptyJSONValue(reflect.ValueOf([]int{}))     // true
func isEmptyJSONValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Complex64, reflect.Complex128:
		return v.Complex() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

// writeStructFields appends JSON key-value pairs for all exported fields of the struct value v into buf.
//
// The first pointer is a shared "is this the very first field written" flag that controls comma insertion; callers must initialise it to true.
// Anonymous (embedded) struct fields are promoted recursively via [shouldPromoteAnonymousStruct].
//
// Parameters:
//   - v     - a reflect.Value of kind Struct.
//   - buf   - destination buffer; key-value pairs are appended in field order.
//   - first - pointer to a bool tracking whether any field has been written
//     yet (used for comma separation).
//
// Returns:
//   - error - non-nil when [encodeValue] or json.Marshal (for the key) fails.
//
// Example:
//
//	var buf bytes.Buffer
//	first := true
//	writeStructFields(reflect.ValueOf(struct{ A int }{1}), &buf, &first)
//	// buf.String() == `"A":1`
func writeStructFields(v reflect.Value, buf *bytes.Buffer, first *bool) error {
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if !sf.IsExported() {
			continue
		}

		fv := v.Field(i)

		if promoted, ok := shouldPromoteAnonymousStruct(sf, fv); ok {
			if err := writeStructFields(promoted, buf, first); err != nil {
				return err
			}
			continue
		}

		name, omitempty, skip := parseJSONTag(sf)
		if skip {
			continue
		}
		if omitempty && isEmptyJSONValue(fv) {
			continue
		}

		if !*first {
			buf.WriteByte(',')
		}
		*first = false

		keyBytes, err := json.Marshal(name)
		if err != nil {
			return err
		}
		buf.Write(keyBytes)
		buf.WriteByte(':')

		valueBytes, err := encodeValue(fv)
		if err != nil {
			return err
		}
		buf.Write(valueBytes)
	}

	return nil
}

// encodeStruct encodes v (which must be of kind Struct) into a JSON object.
//
// Exported fields are written in declaration order. Anonymous embedded fields
// are promoted (inlined) according to standard encoding/json rules. Fields
// tagged with `json:"-"` are omitted; fields tagged with `json:",omitempty"`
// are omitted when their value is empty per [isEmptyJSONValue].
//
// Parameters:
//   - v - a reflect.Value of kind Struct.
//
// Returns:
//   - []byte - the JSON object bytes, e.g. `{"Name":"Alice","Age":30}`.
//   - error  - non-nil when any field value cannot be encoded.
//
// Example:
//
//	type Point struct{ X, Y int }
//	encodeStruct(reflect.ValueOf(Point{1, 2})) // []byte(`{"X":1,"Y":2}`), nil
func encodeStruct(v reflect.Value) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	first := true
	if err := writeStructFields(v, &buf, &first); err != nil {
		return nil, err
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// encodeSlice encodes v (a slice or array) into a JSON array.
//
// A nil slice must be detected by the caller before invoking this function;
// by the time encodeSlice is called the value is assumed to be non-nil.
// Each element is encoded recursively via [encodeValue].
//
// Parameters:
//   - v - a reflect.Value of kind Slice or Array.
//
// Returns:
//   - []byte - the JSON array bytes, e.g. `[1,2,3]`.
//   - error  - non-nil when any element cannot be encoded.
//
// Example:
//
//	encodeSlice(reflect.ValueOf([]int{1, 2, 3}))    // []byte(`[1,2,3]`), nil
//	encodeSlice(reflect.ValueOf([2]string{"a","b"})) // []byte(`["a","b"]`), nil
func encodeSlice(v reflect.Value) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('[')
	for i := 0; i < v.Len(); i++ {
		if i > 0 {
			buf.WriteByte(',')
		}
		b, err := encodeValue(v.Index(i))
		if err != nil {
			return nil, err
		}
		buf.Write(b)
	}
	buf.WriteByte(']')
	return buf.Bytes(), nil
}

// encodeMap encodes v (a map) into a JSON object whose keys are sorted
// lexicographically to guarantee deterministic output.
//
// A nil map must be detected by the caller before invoking this function.
// Map keys are converted to strings via [mapKeyString]; values are encoded
// recursively via [encodeValue].
//
// Parameters:
//   - v - a reflect.Value of kind Map (non-nil).
//
// Returns:
//   - []byte - the JSON object bytes with sorted keys, e.g. `{"a":1,"b":2}`.
//   - error  - non-nil when a key or value cannot be encoded.
//
// Example:
//
//	m := map[string]int{"b": 2, "a": 1}
//	encodeMap(reflect.ValueOf(m)) // []byte(`{"a":1,"b":2}`), nil  (sorted)
func encodeMap(v reflect.Value) ([]byte, error) {
	entries := make([]mapEntry, 0, v.Len())

	for _, key := range v.MapKeys() {
		s, err := mapKeyString(key)
		if err != nil {
			return nil, err
		}
		entries = append(entries, mapEntry{key: s, val: v.MapIndex(key)})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].key < entries[j].key
	})

	var buf bytes.Buffer
	buf.WriteByte('{')

	for i, entry := range entries {
		if i > 0 {
			buf.WriteByte(',')
		}
		keyBytes, err := json.Marshal(entry.key)
		if err != nil {
			return nil, err
		}
		buf.Write(keyBytes)
		buf.WriteByte(':')

		valueBytes, err := encodeValue(entry.val)
		if err != nil {
			return nil, err
		}
		buf.Write(valueBytes)
	}

	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// mapKeyString converts a map key reflect.Value to its string representation
// for use as a JSON object key.
//
// Conversion priority:
//  1. [encoding.TextMarshaler] (value or pointer receiver) - highest priority,
//     matches stdlib behavior.
//  2. string
//  3. bool
//  4. Signed integer kinds (int, int8, int16, int32, int64)
//  5. Unsigned integer kinds (uint, uint8, uint16, uint32, uint64, uintptr)
//  6. float32 / float64
//  7. All other kinds → [*json.UnsupportedTypeError]
//
// Parameters:
//   - v  - a reflect.Value that is a key extracted from a map.
//
// Returns:
//   - string - the string form of the key.
//   - error  - non-nil ([*json.UnsupportedTypeError]) for unsupported kinds.
//
// Example:
//
//	mapKeyString(reflect.ValueOf("name"))  // "name", nil
//	mapKeyString(reflect.ValueOf(42))      // "42",   nil
//	mapKeyString(reflect.ValueOf(true))    // "true", nil
//	mapKeyString(reflect.ValueOf(3.14))    // "3.14", nil
func mapKeyString(v reflect.Value) (string, error) {
	// TextMarshaler has highest priority (matches stdlib behaviour).
	if b, ok, err := marshalTextValue(v); ok {
		return string(b), err
	}

	switch v.Kind() {
	case reflect.String:
		return v.String(), nil

	case reflect.Bool:
		return strconv.FormatBool(v.Bool()), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(v.Uint(), 10), nil

	case reflect.Float32:
		return strconv.FormatFloat(v.Float(), 'g', -1, 32), nil

	case reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'g', -1, 64), nil

	default:
		return "", &json.UnsupportedTypeError{Type: v.Type()}
	}
}

// marshalTextValue invokes [encoding.TextMarshaler] on v, or on &v when v is
// addressable, and returns the raw text bytes produced by MarshalText.
//
// Unlike [marshalViaCustomMarshaler], this function does NOT JSON-quote the
// result; callers that need a JSON string must quote it themselves.
//
// Parameters:
//   - v - the reflect.Value to inspect.
//
// Returns:
//   - []byte - the raw text bytes when the interface is satisfied.
//   - bool   - true when [encoding.TextMarshaler] was found and invoked.
//   - error  - non-nil when MarshalText returned an error.
//
// Example:
//
//	// net.IP implements encoding.TextMarshaler:
//	ip := net.ParseIP("127.0.0.1")
//	b, ok, err := marshalTextValue(reflect.ValueOf(ip)) // []byte("127.0.0.1"), true, nil
//
//	// Plain int does not:
//	b, ok, err := marshalTextValue(reflect.ValueOf(42))  // nil, false, nil
func marshalTextValue(v reflect.Value) ([]byte, bool, error) {
	t := v.Type()

	if t.Implements(textMarshalerType) {
		b, err := v.Interface().(stdencoding.TextMarshaler).MarshalText()
		return b, true, err
	}
	if v.CanAddr() && v.Addr().Type().Implements(textMarshalerType) {
		b, err := v.Addr().Interface().(stdencoding.TextMarshaler).MarshalText()
		return b, true, err
	}

	return nil, false, nil
}

// formatFloatToken converts a float64 value to its JSON token string.
//
// Special cases:
//   - NaN or ±Inf: returns "null" when [floatsUseNullForNonFinite] is true,
//     otherwise returns [ErrNonFiniteFloat].
//   - is32 = true selects 32-bit shortest-representation formatting
//     (strconv bitSize 32) to faithfully round-trip float32 values.
//
// Parameters:
//   - f    - the float64 (or widened float32) value to format.
//   - is32 - true when the original value was a float32.
//
// Returns:
//   - string - the JSON number token (e.g. "3.14", "1e+100"), or "null" for
//     non-finite values when floatsUseNullForNonFinite is true.
//   - bool   - always true (signals to callers that the kind was handled).
//   - error  - [ErrNonFiniteFloat] for NaN/Inf when floatsUseNullForNonFinite
//     is false.
//
// Example:
//
//	formatFloatToken(3.14, false)        // "3.14",  true, nil
//	formatFloatToken(1e308, false)       // "1e+308", true, nil
//	formatFloatToken(math.NaN(), false)  // "",      true, ErrNonFiniteFloat
//	formatFloatToken(math.NaN(), false)  // "null",  true, nil  (if floatsUseNullForNonFinite)
func formatFloatToken(f float64, is32 bool) (string, bool, error) {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		if floatsUseNullForNonFinite {
			return "null", true, nil
		}
		return "", true, ErrNonFiniteFloat
	}

	bitSize := 64
	if is32 {
		bitSize = 32
	}
	return strconv.FormatFloat(f, 'g', -1, bitSize), true, nil
}

// encodeComplexJSONToken encodes a complex reflect.Value as a JSON object of
// the form {"real":<r>,"imag":<i>}.
//
// Both the real and imaginary parts are formatted via [formatFloatToken] with
// the precision appropriate for the kind (complex64 uses 32-bit precision,
// complex128 uses 64-bit).
//
// Parameters:
//   - v - a reflect.Value of kind Complex64 or Complex128.
//
// Returns:
//   - string - e.g. `{"real":1,"imag":-2.5}`.
//   - error  - [ErrNonFiniteFloat] when either part is NaN or ±Inf and
//     [floatsUseNullForNonFinite] is false.
//
// Example:
//
//	encodeComplexJSONToken(reflect.ValueOf(complex(1, -2)))   // `{"real":1,"imag":-2}`, nil
//	encodeComplexJSONToken(reflect.ValueOf(complex64(0.5+1i))) // `{"real":0.5,"imag":1}`, nil
func encodeComplexJSONToken(v reflect.Value) (string, error) {
	r, i := complexParts(v)
	is32 := v.Kind() == reflect.Complex64

	rs, _, err := formatFloatToken(r, is32)
	if err != nil {
		return "", err
	}
	is, _, err := formatFloatToken(i, is32)
	if err != nil {
		return "", err
	}

	return `{"real":` + rs + `,"imag":` + is + `}`, nil
}

// encodeComplexBytes is the []byte-returning variant of [encodeComplexJSONToken].
//
// Parameters:
//   - v - a reflect.Value of kind Complex64 or Complex128.
//
// Returns:
//   - []byte - the JSON object bytes, e.g. []byte(`{"real":1,"imag":2}`).
//   - error  - non-nil when [encodeComplexJSONToken] fails.
//
// Example:
//
//	encodeComplexBytes(reflect.ValueOf(complex(3, 4))) // []byte(`{"real":3,"imag":4}`), nil
func encodeComplexBytes(v reflect.Value) ([]byte, error) {
	token, err := encodeComplexJSONToken(v)
	if err != nil {
		return nil, err
	}
	return []byte(token), nil
}

// complexParts extracts the real and imaginary components from a complex
// reflect.Value as float64 values.
//
// v.Complex() always returns complex128; for complex64 values Go widens them
// automatically, so no separate branch is needed.
//
// Parameters:
//   - v - a reflect.Value of kind Complex64 or Complex128.
//
// Returns:
//   - r - the real part as float64.
//   - i - the imaginary part as float64.
//
// Example:
//
//	complexParts(reflect.ValueOf(complex(1.5, -2.0))) // r=1.5, i=-2.0
//	complexParts(reflect.ValueOf(complex64(0+1i)))    // r=0,   i=1
func complexParts(v reflect.Value) (r, i float64) {
	c := v.Complex()
	return real(c), imag(c)
}
