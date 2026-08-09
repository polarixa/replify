package strutil

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
	"unsafe"

	"github.com/polarixa/replify/pkg/truncate"
)

var (
	// Len is an alias for the utf8.RuneCountInString function, which returns the number of runes
	// (Unicode code points) in the given string. This function treats erroneous and short
	// encodings as single runes of width 1 byte. It is useful for determining the
	// length of a string in terms of its Unicode characters, rather than bytes,
	// allowing for accurate character counting in UTF-8 encoded strings.
	Len = utf8.RuneCountInString

	// RegexpDupSpaces is a precompiled regular expression that matches one or more consecutive
	// whitespace characters (including spaces, tabs, and newlines). This can be used for tasks
	// such as normalizing whitespace in strings by replacing multiple whitespace characters
	// with a single space, or for validating string formats where excessive whitespace should
	// be trimmed or removed.
	RegexpDupSpaces = regexp.MustCompile(`\s+`)

	// MaxRuneBytes represents the maximum valid UTF-8 encoding of a Unicode code point.
	// It is a byte slice containing the specific byte values [244, 143, 191, 191].
	MaxRuneBytes = [...]byte{244, 143, 191, 191}

	// lineBreakReplacer replaces every line-breaking sequence with a single space.
	// \r\n is listed first so it is consumed as one unit before \r or \n are tried.
	lineBreakReplacer = strings.NewReplacer(
		"\r\n", " ",
		"\r", " ",
		"\n", " ",
		"\t", " ",
		"\v", " ",
		"\f", " ",
		"\u0085", " ", // NEL - Next Line
		"\u2028", " ", // LS  - Line Separator
		"\u2029", " ", // PS  - Paragraph Separator
	)
)

// IsEmpty checks if the provided string is empty or consists solely of whitespace characters.
//
// The function trims leading and trailing whitespace from the input string `s` using
// strings.TrimSpace. It then evaluates the length of the trimmed string. If the length is
// zero, it indicates that the original string was either empty or contained only whitespace,
// and the function returns true. Otherwise, it returns false.
//
// Parameters:
//   - `s`: A string that needs to be checked for emptiness.
//
// Returns:
//
//	A boolean value:
//	 - true if the string is empty or contains only whitespace characters;
//	 - false if the string contains any non-whitespace characters.
//
// Example:
//
//	result := IsEmpty("   ") // result will be true
//	result = IsEmpty("Hello") // result will be false
func IsEmpty(s string) bool {
	trimmed := strings.TrimSpace(s)
	return len(trimmed) == 0
}

// IsEmptyPtr checks if the provided string pointer is nil or points to an empty string.
//
// Parameters:
//   - `s`: A pointer to a string that needs to be checked for emptiness.
//
// Returns:
//
//	A boolean value:
//	 - true if the string pointer is nil or points to an empty string;
//	 - false if the string pointer points to a non-empty string.
//
// Example:
//
//	result := IsEmptyPtr(nil) // result will be true
//	result = IsEmptyPtr(&"   ") // result will be true
//	result = IsEmptyPtr(&"Hello") // result will be false
func IsEmptyPtr(s *string) bool {
	if s == nil {
		return true
	}
	return IsEmpty(*s)
}

// IsAnyEmpty checks if any of the provided strings are empty.
//
// This function takes a variadic number of string arguments and iterates through
// each string to check if it is empty (i.e., has a length of zero or consists
// solely of whitespace characters). If any string in the provided list is found
// to be empty, the function returns `true`. If all strings are non-empty, it
// returns `false`.
//
// Parameters:
//   - `strings`: A variadic parameter that allows passing multiple strings to be checked.
//
// Returns:
//   - `true` if at least one of the provided strings is empty; `false` if all
//     strings are non-empty.
//
// Example:
//
//	result := IsAnyEmpty("hello", "", "world") // result will be true because one of the strings is empty.
//
// Notes:
//   - The function utilizes the IsEmpty helper function to determine if a string
//     is considered empty, which may include strings that are only whitespace.
func IsAnyEmpty(strings ...string) bool {
	for _, s := range strings {
		if IsEmpty(s) {
			return true
		}
	}
	return false
}

// IsAnyEmptyPtr checks if any of the provided string pointers are nil or point to an empty string.
//
// This function takes a variadic number of string pointer arguments and iterates through
// each string pointer to check if it is nil or points to an empty string (i.e., has a length of zero or consists
// solely of whitespace characters). If any string pointer in the provided list is found
// to be nil or points to an empty string, the function returns `true`. If all string pointers are non-nil and point to non-empty strings, it
// returns `false`.
//
// Parameters:
//   - `strings`: A variadic parameter that allows passing multiple string pointers to be checked.
//
// Returns:
//   - `true` if at least one of the provided string pointers is nil or points to an empty string; `false` if all
//     string pointers are non-nil and point to non-empty strings.
//
// Example:
//
//	result := IsAnyEmptyPtr(&"hello", nil, &"world") // result will be true because one of the string pointers is nil.
//
// Notes:
//   - The function utilizes the IsEmptyPtr helper function to determine if a string pointer
//     is considered empty, which may include strings that are only whitespace.
func IsAnyEmptyPtr(strings ...*string) bool {
	for _, s := range strings {
		if IsEmptyPtr(s) {
			return true
		}
	}
	return false
}

// IsNotEmpty checks if the provided string is not empty or does not consist solely of whitespace characters.
//
// This function leverages the IsEmpty function to determine whether the input string `s`
// is empty or contains only whitespace. It returns the negation of the result from IsEmpty.
// If IsEmpty returns true (indicating the string is empty or whitespace), IsNotEmpty will return false,
// and vice versa.
//
// Parameters:
//   - `s`: A string that needs to be checked for non-emptiness.
//
// Returns:
//
//		 A boolean value:
//	  - true if the string contains at least one non-whitespace character;
//	  - false if the string is empty or contains only whitespace characters.
//
// Example:
//
//	result := IsNotEmpty("Hello") // result will be true
//	result = IsNotEmpty("   ") // result will be false
func IsNotEmpty(s string) bool {
	return !IsEmpty(s)
}

// IsNotEmptyPtr checks if the provided string pointer is not nil and points to a non-empty string.
//
// Parameters:
//   - `s`: A pointer to a string that needs to be checked for non-emptiness.
//
// Returns:
//
//	A boolean value:
//	 - true if the string pointer is not nil and points to a non-empty string;
//	 - false if the string pointer is nil or points to an empty string.
//
// Example:
//
//	result := IsNotEmptyPtr(nil) // result will be false
//	result = IsNotEmptyPtr(&"") // result will be false
//	result = IsNotEmptyPtr(&"Hello") // result will be true
func IsNotEmptyPtr(s *string) bool {
	if s == nil {
		return false
	}
	return IsNotEmpty(*s)
}

// IsNoneEmpty checks if all provided strings are non-empty.
//
// This function takes a variadic number of string arguments and iterates through
// each string to verify that none of them are empty (i.e., have a length of zero
// or consist solely of whitespace characters). If any string in the provided list
// is found to be empty, the function immediately returns `false`. If all strings
// are non-empty, it returns `true`.
//
// Parameters:
//   - `strings`: A variadic parameter that allows passing multiple strings to be checked.
//
// Returns:
//   - `true` if all of the provided strings are non-empty; `false` if at least one
//     string is empty.
//
// Example:
//
//	result := IsNoneEmpty("hello", "world", "!") // result will be true because all strings are non-empty.
//	result2 := IsNoneEmpty("hello", "", "world") // result2 will be false because one of the strings is empty.
//
// Notes:
//   - The function utilizes the IsEmpty helper function to determine if a string
//     is considered empty, which may include strings that are only whitespace.
func IsNoneEmpty(strings ...string) bool {
	for _, s := range strings {
		if IsEmpty(s) {
			return false
		}
	}
	return true
}

// IsNoneEmptyPtr checks if all provided string pointers are non-nil and point to non-empty strings.
//
// This function takes a variadic number of string pointer arguments and iterates through
// each string pointer to verify that none of them are nil or point to an empty string (i.e., have a length of zero
// or consist solely of whitespace characters). If any string pointer in the provided list
// is found to be nil or points to an empty string, the function immediately returns `false`. If all string pointers
// are non-nil and point to non-empty strings, it returns `true`.
//
// Parameters:
//   - `strings`: A variadic parameter that allows passing multiple string pointers to be checked.
//
// Returns:
//   - `true` if all of the provided string pointers are non-nil and point to non-empty strings; `false` if at least one
//     string pointer is nil or points to an empty string.
//
// Example:
//
//	result := IsNoneEmptyPtr(&"hello", &"world", &"!") // result will be true because all string pointers are non-nil and point to non-empty strings.
//	result2 := IsNoneEmptyPtr(&"hello", nil, &"world") // result2 will be false because one of the string pointers is nil.
//	result3 := IsNoneEmptyPtr(&"hello", &"", &"world") // result3 will be false because one of the string pointers points to an empty string.
//
// Notes:
//   - The function utilizes the IsEmptyPtr helper function to determine if a string pointer
//     is considered empty, which may include strings that are only whitespace.
func IsNoneEmptyPtr(strings ...*string) bool {
	for _, s := range strings {
		if IsEmptyPtr(s) {
			return false
		}
	}
	return true
}

// IsBlank checks if a string is blank (empty or contains only whitespace).
//
// This function determines if the input string `s` is considered blank. A string
// is considered blank if it is either an empty string or consists solely of
// whitespace characters (spaces, tabs, newlines, etc.).
//
// The function first checks if the string is empty. If it is, it returns `true`.
// If the string is not empty, it uses a regular expression to check if the
// string contains only whitespace characters. If the string matches this
// condition, it also returns `true`. If neither condition is met, the function
// returns `false`, indicating that the string contains non-whitespace characters.
//
// Parameters:
//   - `s`: The input string to check for blankness.
//
// Returns:
//   - `true` if the string is blank (empty or contains only whitespace);
//     `false` otherwise.
//
// Example:
//
//	result1 := IsBlank("") // result1 will be true because the string is empty.
//	result2 := IsBlank("   ") // result2 will be true because the string contains only spaces.
//	result3 := IsBlank("Hello") // result3 will be false because the string contains non-whitespace characters.
//
// Notes:
//   - The function uses a regular expression to match strings that consist entirely
//     of whitespace. The regex `^\s+$` matches strings that contain one or more
//     whitespace characters from the start to the end of the string.
func IsBlank(s string) bool {
	if s == "" {
		return true
	}
	if regexp.MustCompile(`^\s+$`).MatchString(s) {
		return true
	}
	return false
}

// IsBlankPtr checks if a string pointer is nil or points to a blank string (empty or contains only whitespace).
//
// This function determines if the input string pointer `s` is considered blank. A string pointer
// is considered blank if it is either nil or points to an empty string or a string that consists solely of
// whitespace characters (spaces, tabs, newlines, etc.).
//
// The function first checks if the string pointer is nil. If it is, it returns `true`.
// If the string pointer is not nil, it uses the `IsBlank` helper function to check if the
// string it points to is blank. If the string is blank, the function also returns `true`.
// If neither condition is met, the function returns `false`, indicating that the string pointer
// is not nil and points to a non-blank string.
//
// Parameters:
//   - `s`: The input string pointer to check for blankness.
//
// Returns:
//   - `true` if the string pointer is nil or points to a blank string (empty or contains only whitespace);
//     `false` otherwise.
//
// Example:
//
//	result1 := IsBlankPtr(nil) // result1 will be true because the string pointer is nil.
//	result2 := IsBlankPtr(&"") // result2 will be true because the string pointer points to an empty string.
//	result3 := IsBlankPtr(&"   ") // result3 will be true because the string pointer points to a string with only spaces.
//	result4 := IsBlankPtr(&"Hello") // result4 will be false because the string pointer points to a non-blank string.
//
// Notes:
//   - This function provides a convenient way to check for blank string pointers by combining
//     the nil check with the `IsBlank` function.
func IsBlankPtr(s *string) bool {
	if s == nil {
		return true
	}
	return IsBlank(*s)
}

// IsNotBlank checks if a string is not blank (not empty and contains non-whitespace characters).
//
// This function serves as a logical negation of the `IsBlank` function. It checks
// if the input string `s` contains any non-whitespace characters. A string is
// considered not blank if it is neither empty nor consists solely of whitespace
// characters. This is determined by calling the `IsBlank` function and
// negating its result.
//
// Parameters:
//   - `s`: The input string to check for non-blankness.
//
// Returns:
//   - `true` if the string is not blank (contains at least one non-whitespace character);
//     `false` if the string is blank (empty or contains only whitespace).
//
// Example:
//
//	result1 := IsNotBlank("Hello") // result1 will be true because the string contains non-whitespace characters.
//	result2 := IsNotBlank("   ") // result2 will be false because the string contains only spaces.
//	result3 := IsNotBlank("") // result3 will be false because the string is empty.
//
// Notes:
//   - This function provides a convenient way to check for meaningful content
//     in a string by confirming that it is not blank.
func IsNotBlank(s string) bool {
	return !IsBlank(s)
}

// IsNotBlankPtr checks if a string pointer is not nil and points to a non-blank string.
//
// This function serves as a logical negation of the `IsBlankPtr` function. It checks
// if the input string pointer `s` points to a non-blank string (i.e., a string that is neither empty nor consists solely of whitespace characters).
// This is determined by first checking if the string pointer is nil (in which case it returns `false`),
// and then calling the `IsNotBlank` helper function on the dereferenced string.
//
// Parameters:
//   - `s`: The input string pointer to check for non-blankness.
//
// Returns:
//   - `true` if the string pointer is not nil and points to a non-blank string;
//     `false` if the string pointer is nil or points to a blank string.
//
// Example:
//
//	result1 := IsNotBlankPtr(&"Hello") // result1 will be true because the string pointer is not nil and points to a non-blank string.
//	result2 := IsNotBlankPtr(&"   ") // result2 will be false because the string pointer points to a blank string.
//	result3 := IsNotBlankPtr(nil) // result3 will be false because the string pointer is nil.
//
// Notes:
//   - This function provides a convenient way to check for non-blank string pointers by combining
//     the nil check with the `IsNotBlank` function.
func IsNotBlankPtr(s *string) bool {
	if s == nil {
		return false
	}
	return IsNotBlank(*s)
}

// IsAnyBlank checks if any of the provided strings are blank (empty or containing only whitespace).
//
// This function iterates through a variadic list of strings and determines if at least
// one of them is considered blank. A string is considered blank if it is either empty
// or consists solely of whitespace characters. The function returns `true` as soon as
// it finds a blank string; if none of the strings are blank, it returns `false`.
//
// Parameters:
//   - `strings`: A variadic parameter that accepts one or more strings to check for blankness.
//
// Returns:
//   - `true` if at least one of the provided strings is blank;
//     `false` if none of the strings are blank.
//
// Example:
//
//	result1 := IsAnyBlank("Hello", "World", " ") // result1 will be true because the third string is blank (contains only a space).
//	result2 := IsAnyBlank("Hello", "World") // result2 will be false because both strings are not blank.
//
// Notes:
//   - This function is useful for validating input or ensuring that required fields
//     are not left blank in forms or data processing.
func IsAnyBlank(strings ...string) bool {
	for _, s := range strings {
		if IsBlank(s) {
			return true
		}
	}
	return false
}

// IsAnyBlankPtr checks if any of the provided string pointers are nil or point to a blank string.
//
// This function iterates through a variadic list of string pointers and determines if at least
// one of them is considered blank. A string pointer is considered blank if it is either nil
// or points to an empty string or a string that consists solely of whitespace characters.
// The function returns `true` as soon as it finds a blank string pointer; if none of the
// string pointers are blank, it returns `false`.
//
// Parameters:
//   - `strings`: A variadic parameter that accepts one or more string pointers to check for blankness.
//
// Returns:
//   - `true` if at least one of the provided string pointers is nil or points to a blank string;
//     `false` if none of the string pointers are nil or point to a blank string.
//
// Example:
//
//	result1 := IsAnyBlankPtr(&"Hello", &"World", nil) // result1 will be true because the third element is a nil pointer.
//	result2 := IsAnyBlankPtr(&"Hello", &"World") // result2 will be false because both string pointers point to non-blank strings.
//
// Notes:
//   - This function provides a convenient way to check for blank string pointers by combining
//     the nil check with the `IsBlankPtr` function.
func IsAnyBlankPtr(strings ...*string) bool {
	for _, s := range strings {
		if IsBlankPtr(s) {
			return true
		}
	}
	return false
}

// IsNoneBlank checks if none of the provided strings are blank (empty or containing only whitespace).
//
// This function iterates through a variadic list of strings and returns `false` if any of them
// is blank. A string is considered blank if it is either empty or consists solely of whitespace
// characters. If all strings are non-blank, it returns `true`.
//
// Parameters:
//   - `strings`: A variadic parameter that accepts one or more strings to check for blankness.
//
// Returns:
//   - `true` if none of the provided strings are blank;
//     `false` if at least one of the strings is blank.
//
// Example:
//
//	result1 := IsNoneBlank("Hello", "World", " ") // result1 will be false because the third string is blank (contains only a space).
//	result2 := IsNoneBlank("Hello", "World") // result2 will be true because both strings are non-blank.
//
// Notes:
//   - This function is useful for validating input or ensuring that all required fields
//     contain meaningful data in forms or data processing.
func IsNoneBlank(strings ...string) bool {
	for _, s := range strings {
		if IsBlank(s) {
			return false
		}
	}
	return true
}

// IsNoneBlankPtr checks if none of the provided string pointers are nil or point to a blank string.
//
// This function iterates through a variadic list of string pointers and returns `false` if any of them
// is blank. A string pointer is considered blank if it is either nil or points to an empty string or a string that consists solely of whitespace characters.
// If all string pointers point to non-blank strings, it returns `true`.
//
// Parameters:
//   - `strings`: A variadic parameter that accepts one or more string pointers to check for blankness.
//
// Returns:
//   - `true` if none of the provided string pointers are nil or point to a blank string;
//     `false` if at least one of the string pointers is nil or points to a blank string.
//
// Example:
//
//	result1 := IsNoneBlankPtr(&"Hello", &"World", &" ") // result1 will be false because the third element points to a blank string.
//	result2 := IsNoneBlankPtr(&"Hello", &"World") // result2 will be true because both string pointers point to non-blank strings.
//
// Notes:
//   - This function provides a convenient way to check for non-blank string pointers by combining
//     the nil check with the `IsNoneBlank` function.
func IsNoneBlankPtr(strings ...*string) bool {
	for _, s := range strings {
		if IsBlankPtr(s) {
			return false
		}
	}
	return true
}

// IsNumeric checks if the provided string contains only numeric digits (0-9).
//
// This function iterates through each character of the input string and verifies if each
// character is a digit using the `unicode.IsDigit` function. If it encounters any character
// that is not a digit, it returns `false`. If all characters are digits, it returns `true`.
//
// Parameters:
//   - `str`: The input string to be checked for numeric characters.
//
// Returns:
//   - `true` if the string contains only numeric digits;
//     `false` if the string contains any non-numeric characters.
//
// Example:
//
//	result1 := IsNumeric("12345") // result1 will be true because the string contains only digits.
//	result2 := IsNumeric("123A45") // result2 will be false because the string contains a non-digit character ('A').
//
// Notes:
//   - This function is useful for validating numeric input, such as in forms or parsing
//     data that should be strictly numeric.
func IsNumeric(str string) bool {
	for _, c := range str {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

// IsNumericPtr checks if the provided string pointer points to a string containing only numeric digits (0-9).
//
// This function first checks if the string pointer is nil. If it is, the function returns `false`.
// If the pointer is not nil, it dereferences the pointer to get the actual string and then uses
// the `IsNumeric` function to check if that string contains only numeric digits.
//
// Parameters:
//   - `str`: The input string pointer to be checked for numeric characters.
//
// Returns:
//   - `true` if the string pointer is not nil and the pointed-to string contains only numeric digits;
//     `false` if the string pointer is nil or the pointed-to string contains any non-numeric characters.
//
// Example:
//
//	result1 := IsNumericPtr(&"12345") // result1 will be true because the string pointer points to a string containing only digits.
//	result2 := IsNumericPtr(&"123A45") // result2 will be false because the string pointer points to a string containing a non-digit character ('A').
//	result3 := IsNumericPtr(nil) // result3 will be false because the string pointer is nil.
//
// Notes:
//   - This function provides a convenient way to check for numeric string pointers by combining
//     the nil check with the `IsNumeric` function.
func IsNumericPtr(str *string) bool {
	if str == nil {
		return false
	}
	return IsNumeric(*str)
}

// IsNumericSpace checks if the provided string contains only numeric digits and whitespace characters.
//
// This function iterates through each character of the input string and verifies if each character
// is either a digit (0-9) or a whitespace character (spaces, tabs, etc.) using the `unicode.IsDigit`
// and `unicode.IsSpace` functions. If it encounters any character that is neither a digit nor a
// whitespace, it returns `false`. If all characters are valid, it returns `true`.
//
// Parameters:
//   - `str`: The input string to be checked for numeric digits and whitespace.
//
// Returns:
//   - `true` if the string contains only digits and whitespace;
//     `false` if the string contains any non-numeric and non-whitespace characters.
//
// Example:
//
//	result1 := IsNumericSpace("123 456") // result1 will be true because the string contains only digits and a space.
//	result2 := IsNumericSpace("123A456") // result2 will be false because the string contains a non-numeric character ('A').
//
// Notes:
//   - This function is useful for validating input that should be a number but may also include
//     spaces, such as in user forms or data processing where formatting is flexible.
func IsNumericSpace(str string) bool {
	for _, c := range str {
		if !unicode.IsDigit(c) && !unicode.IsSpace(c) {
			return false
		}
	}
	return true
}

// IsNumericSpacePtr checks if the provided string pointer points to a string containing only numeric digits and whitespace characters.
//
// This function first checks if the string pointer is nil. If it is, the function returns `false`.
// If the pointer is not nil, it dereferences the pointer to get the actual string and then uses
// the `IsNumericSpace` function to check if that string contains only numeric digits and whitespace.
//
// Parameters:
//   - `str`: The input string pointer to be checked for numeric digits and whitespace.
//
// Returns:
//   - `true` if the string pointer is not nil and the pointed-to string contains only numeric digits and whitespace;
//     `false` if the string pointer is nil or the pointed-to string contains any non-numeric and non-whitespace characters.
//
// Example:
//
//	result1 := IsNumericSpacePtr(&"123 456") // result1 will be true because the string pointer points to a string containing only digits and a space.
//	result2 := IsNumericSpacePtr(&"123A456") // result2 will be false because the string pointer points to a string containing a non-numeric character ('A').
//	result3 := IsNumericSpacePtr(nil) // result3 will be false because the string pointer is nil.
//
// Notes:
//   - This function provides a convenient way to check for numeric space string pointers by combining
//     the nil check with the `IsNumericSpace` function.
func IsNumericSpacePtr(str *string) bool {
	if str == nil {
		return false
	}
	return IsNumericSpace(*str)
}

// IsWhitespace checks if the provided string contains only whitespace characters.
//
// This function iterates through each character of the input string and checks if each character
// is a whitespace character (spaces, tabs, newlines, etc.) using the `unicode.IsSpace` function.
// If it encounters any character that is not a whitespace, it returns `false`. If all characters
// are whitespace, it returns `true`.
//
// Parameters:
//   - `str`: The input string to be checked for whitespace.
//
// Returns:
//   - `true` if the string contains only whitespace characters;
//     `false` if the string contains any non-whitespace characters.
//
// Example:
//
//	result1 := IsWhitespace("    ") // result1 will be true because the string contains only spaces.
//	result2 := IsWhitespace("Hello") // result2 will be false because the string contains non-whitespace characters.
//
// Notes:
//   - This function is useful for determining if a string is blank in terms of visible content,
//     which can be important in user input validation or string processing tasks.
func IsWhitespace(str string) bool {
	for _, c := range str {
		if !unicode.IsSpace(c) {
			return false
		}
	}
	return true
}

// IsWhitespacePtr checks if the provided string pointer points to a string containing only whitespace characters.
//
// This function first checks if the string pointer is nil. If it is, the function returns `false`.
// If the pointer is not nil, it dereferences the pointer to get the actual string and then uses
// the `IsWhitespace` function to check if that string contains only whitespace characters.
//
// Parameters:
//   - `str`: The input string pointer to be checked for whitespace.
//
// Returns:
//   - `true` if the string pointer is not nil and the pointed-to string contains only whitespace characters;
//     `false` if the string pointer is nil or the pointed-to string contains any non-whitespace characters.
//
// Example:
//
//	result1 := IsWhitespacePtr(&"    ") // result1 will be true because the string pointer points to a string containing only spaces.
//	result2 := IsWhitespacePtr(&"Hello") // result2 will be false because the string pointer points to a string containing non-whitespace characters.
//	result3 := IsWhitespacePtr(nil) // result3 will be false because the string pointer is nil.
//
// Notes:
//   - This function provides a convenient way to check for whitespace string pointers by combining
//     the nil check with the `IsWhitespace` function.
func IsWhitespacePtr(str *string) bool {
	if str == nil {
		return false
	}
	return IsWhitespace(*str)
}

// TrimWhitespace removes extra whitespace from the input string,
// replacing any sequence of whitespace characters with a single space.
//
// This function first checks if the input string `s` is empty or consists solely of whitespace
// using the IsEmpty function. If so, it returns an empty string. If the string contains
// non-whitespace characters, it utilizes a precompiled regular expression (regexpDupSpaces)
// to identify and replace all sequences of whitespace characters (including spaces, tabs, and
// newlines) with a single space. This helps to normalize whitespace in the string.
//
// Parameters:
// - `s`: The input string from which duplicate whitespace needs to be removed.
//
// Returns:
//   - A string with all sequences of whitespace characters replaced by a single space.
//     If the input string is empty or only contains whitespace, an empty string is returned.
//
// Example:
//
//	result := TrimWhitespace("This   is  an example.\n\nThis is another line.") // result will be "This is an example. This is another line."
func TrimWhitespace(s string) string {
	if IsEmpty(s) {
		return ""
	}
	// Use a regular expression to replace all sequences of whitespace characters with a single space.
	s = RegexpDupSpaces.ReplaceAllString(s, " ")
	return s
}

// CleanSpaces removes leading and trailing whitespace characters from a given string and replaces sequences of whitespace characters with a single space.
// It first checks if the input string is empty or consists solely of whitespace characters. If so, it returns an empty string.
// Otherwise, it calls TrimWhitespace to replace all sequences of whitespace characters with a single space, effectively removing duplicates.
// Finally, it trims the leading and trailing whitespace characters from the resulting string using strings.TrimSpace and returns the cleaned string.
func CleanSpaces(s string) string {
	if IsEmpty(s) {
		return ""
	}
	return strings.TrimSpace(TrimWhitespace(s))
}

// Trim removes leading and trailing whitespace characters from a string.
// The function iteratively checks and removes spaces (or any character less than or equal to a space)
// from both the left (beginning) and right (end) of the string.
//
// Parameters:
//   - s: A string that may contain leading and trailing whitespace characters that need to be removed.
//
// Returns:
//   - A new string with leading and trailing whitespace removed. The function does not modify the original string,
//     as strings in Go are immutable.
//
// Example Usage:
//
//	str := "  hello world  "
//	trimmed := Trim(str)
//	// trimmed: "hello world" (leading and trailing spaces removed)
//
//	str = "\n\n   Trim me   \t\n"
//	trimmed = Trim(str)
//	// trimmed: "Trim me" (leading and trailing spaces and newline characters removed)
//
// Details:
//
//   - The function works by iteratively removing any characters less than or equal to a space (ASCII 32) from the
//     left side of the string until no such characters remain. It then performs the same operation on the right side of
//     the string until no whitespace characters are left.
//
//   - The function uses a `goto` mechanism to handle the removal in a loop, which ensures all leading and trailing
//     spaces (or any whitespace characters) are removed without additional checks for length or condition evaluation
//     in every iteration.
//
//   - The trimmed result string will not contain leading or trailing whitespace characters after the function completes.
//
//   - The function returns an unchanged string if no whitespace is present.
func Trim(s string) string {
	if IsEmpty(s) {
		return s
	}
left:
	if len(s) > 0 && s[0] <= ' ' {
		s = s[1:]
		goto left
	}
right:
	if len(s) > 0 && s[len(s)-1] <= ' ' {
		s = s[:len(s)-1]
		goto right
	}
	return s
}

// TrimLeft removes all leading occurrences of the specified cutset string from the
// input string s. It repeatedly strips the cutset from the left side until no more
// leading matches remain.
//
// Parameters:
//   - s: The input string to trim.
//   - cutset: The substring to remove from the left side. If cutset is empty,
//     the original string is returned unchanged.
//
// Returns:
//   - A new string with all leading occurrences of cutset removed.
//
// Example:
//
//	result := TrimLeft("###hello", "#")   // result will be "hello"
//	result = TrimLeft("ababHi", "ab")     // result will be "Hi"
//	result = TrimLeft("hello", "#")       // result will be "hello" (no change)
func TrimLeft(s string, cutset string) string {
	if IsEmpty(cutset) {
		return s
	}
	for strings.HasPrefix(s, cutset) {
		s = s[len(cutset):]
	}
	return s
}

// TrimRight removes all trailing occurrences of the specified cutset string from the
// input string s. It repeatedly strips the cutset from the right side until no more
// trailing matches remain.
//
// Parameters:
//   - s: The input string to trim.
//   - cutset: The substring to remove from the right side. If cutset is empty,
//     the original string is returned unchanged.
//
// Returns:
//   - A new string with all trailing occurrences of cutset removed.
//
// Example:
//
//	result := TrimRight("hello###", "#")   // result will be "hello"
//	result = TrimRight("Hiabab", "ab")     // result will be "Hi"
//	result = TrimRight("hello", "#")       // result will be "hello" (no change)
func TrimRight(s string, cutset string) string {
	if IsEmpty(cutset) {
		return s
	}
	for strings.HasSuffix(s, cutset) {
		s = s[:len(s)-len(cutset)]
	}
	return s
}

// Quote formats a string argument for safe output, escaping any special characters
// and enclosing the result in double quotes.
//
// This function uses the fmt.Sprintf function with the %#q format verb to create a quoted
// string representation of the input argument `arg`. The output will escape any special
// characters (such as newlines or tabs) in the string, ensuring that it is suitable for
// safe display or logging. The resulting string will be surrounded by double quotes,
// making it clear where the string begins and ends.
//
// Parameters:
//   - `arg`: The input string to be formatted.
//
// Returns:
//   - A string that represents the input `arg` as a quoted string with special characters
//     escaped. This can be useful for creating safe outputs in logs or console displays.
//
// Example:
//
//	formatted := Quote("Hello, world!\nNew line here.") // formatted will be "\"Hello, world!\\nNew line here.\""
func Quote(arg string) string {
	return fmt.Sprintf("%#q", arg)
}

// TrimPrefixAll returns a new string with all occurrences of prefix at the start of s removed.
// If prefix is the empty string, this function returns s.
func TrimPrefixAll(s string, prefix string) string {
	if IsEmpty(prefix) {
		return s
	}
	for strings.HasPrefix(s, prefix) {
		s = s[len(prefix):]
	}
	return s
}

// TrimPrefixN returns a new string with up to n occurrences of prefix at the start of s removed.
// If prefix is the empty string, this function returns s.
// If n is negative, returns TrimPrefixAll(s, prefix).
func TrimPrefixN(s string, prefix string, n int) string {
	if n < 0 {
		return TrimPrefixAll(s, prefix)
	}
	if IsEmpty(prefix) {
		return s
	}
	for n > 0 && strings.HasPrefix(s, prefix) {
		s = s[len(prefix):]
		n--
	}
	return s
}

// TrimSuffixAll returns a new string with all occurrences of suffix at the end of s removed.
// If suffix is the empty string, this function returns s.
func TrimSuffixAll(s string, suffix string) string {
	if IsEmpty(suffix) {
		return s
	}
	for strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}
	return s
}

// TrimSuffixN returns a new string with up to n occurrences of suffix at the end of s removed.
// If suffix is the empty string, this function returns s.
// If n is negative, returns TrimSuffixAll(s, suffix).
func TrimSuffixN(s string, suffix string, n int) string {
	if n < 0 {
		return TrimSuffixAll(s, suffix)
	}
	if IsEmpty(suffix) {
		return s
	}
	for n > 0 && strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
		n--
	}
	return s
}

// TrimSequenceAll returns a new string with all occurrences of sequence at the start and end of s removed.
// If sequence is the empty string, this function returns s.
func TrimSequenceAll(s string, sequence string) string {
	return TrimSuffixAll(TrimPrefixAll(s, sequence), sequence)
}

// ReplaceAllStrings takes a slice of strings and replaces all occurrences of a specified
// substring (old) with a new substring (new) in each string of the slice.
//
// This function creates a new slice of strings, where each string is the result of
// replacing all instances of the old substring with the new substring in the corresponding
// string from the input slice. The original slice remains unchanged.
//
// Parameters:
//   - `ss`: A slice of strings in which the replacements will be made.
//   - `old`: The substring to be replaced.
//   - `new`: The substring to replace the old substring with.
//
// Returns:
//   - A new slice of strings with all occurrences of `old` replaced by `new` in each string
//     from the input slice.
//
// Example:
//
//	input := []string{"hello world", "world peace", "goodbye world"}
//	output := ReplaceAllStrings(input, "world", "universe") // output will be []string{"hello universe", "universe peace", "goodbye universe"}
func ReplaceAllStrings(ss []string, old string, new string) []string {
	values := make([]string, len(ss))
	for i, s := range ss {
		values[i] = strings.ReplaceAll(s, old, new)
	}
	return values
}

// Slash is like strings.Join(elems, "/"), except that all leading and trailing occurrences of '/'
// between elems are trimmed before they are joined together. Non-trailing leading slashes in the
// first element as well as non-leading trailing slashes in the last element are kept.
func Slash(elems ...string) string {
	return Join(elems, "/")
}

// Join concatenates a slice of strings into a single string, separating each element
// with a specified separator. The function handles various cases of input size and optimizes
// memory allocation based on expected lengths.
//
// Parameters:
//   - `elems`: A slice of strings to be concatenated.
//   - `separator`: A string used to separate the elements in the final concatenated string.
//
// Returns:
//   - A single string resulting from the concatenation of the input strings, with the specified
//     separator inserted between each element. If the slice is empty, it returns an empty string.
//     If there is only one element in the slice, it returns that element without any separators.
//
// The function performs the following steps:
//  1. Checks if the input slice is empty; if so, it returns an empty string.
//  2. If the slice contains a single element, it returns that element directly.
//  3. A `strings.Builder` is used to efficiently build the output string, with an initial capacity
//     that is calculated based on the number of elements and their average length.
//  4. Each element is appended to the builder, with the specified separator added between them.
//  5. The function also trims any leading or trailing occurrences of the separator from each element
//     to avoid duplicate separators in the output.
//
// Example:
//
//	elems := []string{"apple", "banana", "cherry"}
//	separator := ", "
//	result := Join(elems, separator) // result will be "apple, banana, cherry"
func Join(elems []string, separator string) string {
	if len(elems) == 0 {
		return ""
	}
	if len(elems) == 1 {
		return elems[0]
	}
	var sb strings.Builder
	const maxGuess = 100
	const guessAverageElementLen = 5
	if len(elems) <= maxGuess {
		sb.Grow((len(elems)-1)*len(separator) + len(elems)*guessAverageElementLen)
	} else {
		sb.Grow((len(elems)-1)*len(separator) + maxGuess*guessAverageElementLen)
	}
	t := TrimSuffixAll(elems[0], separator) + separator
	for _, element := range elems[1 : len(elems)-1] {
		sb.WriteString(t)
		t = TrimSequenceAll(element, separator) + separator
	}
	sb.WriteString(t)
	t = TrimPrefixAll(elems[len(elems)-1], separator)
	sb.WriteString(t)
	return sb.String()
}

// Reverse returns a new string that is the reverse of the input string s.
// This function handles multi-byte Unicode characters correctly by operating on runes,
// ensuring that each character is reversed without corrupting the character encoding.
//
// Parameters:
//   - `s`: A string to be reversed.
//
// Returns:
//   - A new string that contains the characters of the input string in reverse order.
//     If the input string has fewer than two characters (i.e., is empty or a single character),
//     it returns the input string as-is.
//
// The function works as follows:
//  1. It checks the length of the input string using utf8.RuneCountInString. If the string has
//     fewer than two characters, it returns the original string.
//  2. The input string is converted to a slice of runes to correctly handle multi-byte characters.
//  3. A buffer of runes is created to hold the reversed characters.
//  4. A loop iterates over the original rune slice from the end to the beginning, copying each
//     character into the buffer in reverse order.
//  5. Finally, the function converts the buffer back to a string and returns it.
//
// Example:
//
//	original := "hello"
//	reversed := Reverse(original) // reversed will be "olleh"
func Reverse(s string) string {
	if utf8.RuneCountInString(s) < 2 {
		return s
	}
	r := []rune(s)
	buffer := make([]rune, len(r))
	for i, j := len(r)-1, 0; i >= 0; i-- {
		buffer[j] = r[i]
		j++
	}
	return string(buffer)
}

// Hash256 computes the SHA256 hash of the input string s and returns it as a hexadecimal string.
//
// This function performs the following steps:
//  1. It checks if the input string is empty or consists only of whitespace characters using the IsEmpty function.
//     If the string is empty, it returns the original string.
//  2. It creates a new SHA256 hash using the sha256.New() function.
//  3. The input string is converted to a byte slice and written to the hash. If an error occurs during this process,
//     the function returns an empty string.
//  4. Once the string has been written to the hash, it calculates the final hash value using the Sum method.
//  5. The hash value is then formatted as a hexadecimal string using fmt.Sprintf and returned.
//
// Parameters:
//   - `s`: The input string to be hashed.
//
// Returns:
//   - A string representing the SHA256 hash of the input string in hexadecimal format.
//     If the input string is empty or if an error occurs during hashing, an empty string is returned.
//
// Example:
//
//	input := "hello"
//	hashValue := Hash256(input) // hashValue will contain the SHA256 hash of "hello" in hexadecimal format.
//
// Notes:
//   - This function is suitable for generating hash values for strings that can be used for comparisons,
//     checksums, or other cryptographic purposes. However, if the input string is empty, it returns the empty
//     string as a direct response.
func Hash256(s string) string {
	// Check if the input string is empty or consists solely of whitespace characters
	if IsEmpty(s) {
		return s
	}
	// Create a new SHA256 hash
	h := sha256.New()
	// Write the input string to the hash
	if _, err := h.Write([]byte(s)); err != nil {
		// If an error occurs during the write process, return an empty string
		return ""
	}
	// Sum the hash to get the final hash value
	sum := h.Sum(nil)
	// Format the hash value as a hexadecimal string and return it
	return fmt.Sprintf("%x", sum)
}

// OnlyLetters returns a new string containing only the letters from the original string, excluding all non-letter characters such as numbers, spaces, and special characters.
// This function iterates through each character in the input string, checks if it is a letter using the unicode.IsLetter function, and appends it to a slice of runes if it is.
// The function returns a string created from the slice of letters.
func OnlyLetters(sequence string) string {
	// Check if the input string is empty or consists solely of whitespace characters
	if IsEmpty(sequence) {
		return ""
	}
	// Check if the input string has no runes (e.g., it contains only whitespace characters)
	if utf8.RuneCountInString(sequence) == 0 {
		return ""
	}
	// Initialize a slice to store the letters found in the input string
	var letters []rune
	// Iterate through each character in the input string
	for _, r := range sequence {
		// Check if the current character is a letter
		if unicode.IsLetter(r) {
			// If it is a letter, append it to the slice of letters
			letters = append(letters, r)
		}
	}
	// Convert the slice of letters back into a string and return it
	return string(letters)
}

// OnlyDigits returns a new string containing only the digits from the original string, excluding all non-digit characters such as letters, spaces, and special characters.
// This function first checks if the input string is empty or consists solely of whitespace characters. If so, it returns an empty string.
// If the input string is not empty, it uses a regular expression to replace all non-digit characters with an empty string, effectively removing them.
// The function returns the resulting string, which contains only the digits from the original string.
func OnlyDigits(sequence string) string {
	if IsEmpty(sequence) {
		return ""
	}
	if utf8.RuneCountInString(sequence) > 0 {
		re, _ := regexp.Compile(`[\D]`)
		sequence = re.ReplaceAllString(sequence, "")
	}
	return sequence
}

// Indent takes a string `s` and a string `left`, and indents every line in `s` by prefixing it with `left`.
// Empty lines are also indented.
//
// Parameters:
//   - `s`: The input string whose lines will be indented. It may contain multiple lines separated by newline characters (`\n`).
//   - `left`: The string that will be used as the indentation prefix. This string is prepended to every line of `s`, including empty lines.
//
// Behavior:
//   - The function works by replacing each newline character (`\n`) in `s` with a newline followed by the indentation string `left`.
//   - It also adds `left` to the beginning of the string, ensuring the first line is indented.
//   - Empty lines, if present, are preserved and indented like non-empty lines.
//
// Returns:
//   - A new string where every line of the input `s` has been indented by the string `left`.
//
// Example:
//
// Input:
//
//	s = "Hello\nWorld\n\nThis is a test"
//	left = ">>> "
//
// Output:
//
//	">>> Hello\n>>> World\n>>> \n>>> This is a test"
//
// In this example, each line of the input, including the empty line, is prefixed with ">>> ".
func Indent(s string, left string) string {
	return left + strings.Replace(s, "\n", "\n"+left, -1)
}

// RemoveAccents removes accents and diacritics from the input string s,
// converting special characters into their basic ASCII equivalents.
//
// This function processes each rune in the input string and uses the
// normalizeRune function to convert accented characters to their unaccented
// counterparts. The results are collected in a strings.Builder for efficient
// string concatenation.
//
// Parameters:
//   - `s`: The input string from which accents and diacritics are to be removed.
//
// Returns:
//   - A new string that contains the same characters as the input string,
//     but with all accents and diacritics removed. Characters that do not
//     have a corresponding unaccented equivalent are returned as they are.
//
// Example:
//
//	input := "Café naïve"
//	output := RemoveAccents(input) // output will be "Cafe naive"
//
// Notes:
//   - This function is useful for normalizing strings for comparison,
//     searching, or displaying in a consistent format. It relies on the
//     normalizeRune function to perform the actual character conversion.
func RemoveAccents(s string) string {
	var buff strings.Builder
	buff.Grow(len(s))
	for _, r := range s {
		buff.WriteString(normalizeRune(r))
	}
	return buff.String()
}

// Slugify converts a string to a slug which is useful in URLs, filenames.
// It removes accents, converts to lower case, remove the characters which
// are not letters or numbers and replaces spaces with "-".
//
// Example:
//
//	strutil.Slugify("'We löve Motörhead'") //Output: we-love-motorhead
//
// Normalzation is done with strutil.ReplaceAccents function using a rune replacement map
// You can use the following code for better normalization before strutil.Slugify()
//
//	str := "'We löve Motörhead'"
//	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
//	str = transform.String(t, str) //We love Motorhead
//
// Slugify doesn't support transliteration. You should use a transliteration
// library before Slugify like github.com/rainycape/unidecode
//
// Example:
//
//	import "github.com/rainycape/unidecode"
//
//	str := unidecode.Unidecode("你好, world!")
//	strutil.Slugify(str) //Output: ni-hao-world
func Slugify(s string) string {
	return SlugifySpec(s, "-")
}

// SlugifySpec converts a string to a slug with the delimiter.
// It removes accents, converts string to lower case, remove the characters
// which are not letters or numbers and replaces spaces with the delimiter.
//
// Example:
//
//	strutil.SlugifySpec("'We löve Motörhead'", "-") //Output: we-love-motorhead
//
// SlugifySpec doesn't support transliteration. You should use a transliteration
// library before SlugifySpec like github.com/rainycape/unidecode
//
// Example:
//
//	import "github.com/rainycape/unidecode"
//
//	str := unidecode.Unidecode("你好, world!")
//	strutil.SlugifySpec(str, "-") //Output: ni-hao-world
func SlugifySpec(str string, delimiter string) string {
	str = RemoveAccents(str)
	delBytes := []byte(delimiter)
	n := make([]byte, 0, len(str))
	isPrevSpace := false
	for _, r := range str {
		if r >= 'A' && r <= 'Z' {
			r -= 'A' - 'a'
		}
		//replace non-alpha chars with delimiter
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			n = append(n, byte(int8(r)))
			isPrevSpace = false
		case !isPrevSpace:
			if len(n) > 0 {
				n = append(n, delBytes...)
			}
			fallthrough
		default:
			isPrevSpace = true
		}
	}
	ln := len(n)
	ld := len(delimiter)
	if ln >= ld && string(n[ln-ld:]) == delimiter {
		n = n[:ln-ld]
	}
	return string(n)
}

// ToSnakeCase converts the input string s to snake_case format,
// where all characters are lowercase and spaces are replaced with underscores.
//
// This function first trims any leading or trailing whitespace from the input
// string and then converts all characters to lowercase. It subsequently
// replaces all spaces in the string with underscores to achieve the desired
// snake_case format.
//
// Parameters:
//   - `s`: The input string to be converted to snake_case.
//
// Returns:
//   - A new string formatted in snake_case. If the input string is empty or
//     contains only whitespace, the function will return an empty string.
//
// Example:
//
//	input := "Hello World"
//	output := ToSnakeCase(input) // output will be "hello_world"
//
// Notes:
//   - This function is useful for generating variable names, file names,
//     or other identifiers that conform to snake_case naming conventions.
func ToSnakeCase(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	return strings.ReplaceAll(s, " ", "_")
}

// ToCamelCase converts the input string s to CamelCase format,
// where the first letter of each word is capitalized and all spaces
// are removed.
//
// This function first trims any leading or trailing whitespace from the input
// string. It then iterates over each character in the string, capitalizing the
// first character of each word (defined as a sequence of characters following
// a space) while removing all spaces from the final result. The first character
// of the string remains unchanged unless it follows a space.
//
// Parameters:
//   - `s`: The input string to be converted to CamelCase.
//
// Returns:
//   - A new string formatted in CamelCase. If the input string has fewer than
//     two characters, it returns the original string unchanged. If the input
//     string contains only spaces, it returns an empty string.
//
// Example:
//
//	input := "hello world"
//	output := ToCamelCase(input) // output will be "HelloWorld"
//
// Notes:
//   - This function is useful for generating variable names or identifiers that
//     conform to CamelCase naming conventions.
func ToCamelCase(s string) string {
	s = strings.TrimSpace(s)
	if Len(s) < 2 {
		return s
	}
	var buff strings.Builder
	var prev string
	for _, r := range s {
		c := string(r)
		if c != " " {
			if prev == " " {
				c = strings.ToUpper(c)
			}
			buff.WriteString(c)
		}
		prev = c
	}
	return buff.String()
}

// SplitCamelCase splits a CamelCase string into its component words.
//
// This function takes a string in CamelCase format and separates it into
// individual words based on transitions between upper and lower case letters,
// as well as transitions between letters and digits. It handles the following cases:
//
//   - A transition from a lowercase letter to an uppercase letter indicates the
//     start of a new word.
//   - A transition from an uppercase letter to a lowercase letter indicates
//     the continuation of a word, unless preceded by a digit.
//   - A digit following a letter also indicates a split between words.
//
// The function also trims any leading or trailing whitespace from the input string.
// If the input string has fewer than two characters, it returns a slice containing
// the original string.
//
// Parameters:
//   - `s`: The input CamelCase string to be split into words.
//
// Returns:
//   - A slice of strings containing the individual words extracted from the
//     input string.
//
// Example:
//
//	input := "CamelCaseString123"
//	output := SplitCamelCase(input) // output will be []string{"Camel", "Case", "String", "123"}
//
// Notes:
//   - This function is useful for parsing identifiers or names that follow
//     CamelCase conventions, making them easier to read and understand.
func SplitCamelCase(s string) []string {
	s = strings.TrimSpace(s)
	if Len(s) < 2 {
		return []string{s}
	}
	var prev rune
	var start int
	words := []string{}
	runes := []rune(s)
	for i, r := range runes {
		if i != 0 {
			switch {
			case unicode.IsDigit(r) && unicode.IsLetter(prev):
				fallthrough
			case unicode.IsUpper(r) && unicode.IsLower(prev):
				words = append(words, string(runes[start:i]))
				start = i
			case unicode.IsLower(r) && unicode.IsUpper(prev) && start != i-1:
				words = append(words, string(runes[start:i-1]))
				start = i - 1
			}
		}
		prev = r
	}
	if start < len(runes) {
		words = append(words, string(runes[start:]))
	}
	return words
}

// RemovePrefixes removes specified prefixes from the start of a given string.
//
// This function checks the input string `s` and removes any prefixes provided
// in the `prefix` variadic parameter. If the string is empty or if no prefixes
// are provided, the original string is returned unchanged. The function will
// attempt to remove each specified prefix in the order they are provided.
//
// Parameters:
//   - `s`: The input string from which prefixes will be removed.
//   - `prefix`: A variadic parameter that takes one or more prefixes to be removed
//     from the beginning of the string.
//
// Returns:
//   - A string with the specified prefixes removed. If no prefixes are matched,
//     or if the string is empty, the original string is returned.
//
// Example:
//
//	input := "prefix_example"
//	output := RemovePrefixes(input, "prefix_", "test_") // output will be "example"
//
// Notes:
//   - This function is useful for cleaning up strings by removing unwanted or
//     redundant prefixes in various contexts.
func RemovePrefixes(s string, prefixes ...string) string {
	if IsEmpty(s) {
		return s
	}
	if len(prefixes) == 0 {
		return s
	}
	for _, v := range prefixes {
		s = strings.TrimPrefix(s, v)
	}
	return s
}

// Abbreviate abbreviates a string by shortening it to a specified `maxWidth` and adding ellipses if necessary.
//
// This function shortens a given string `str` to the length specified by `maxWidth`.
// If the string is already shorter than `maxWidth`, it is returned as is. If the
// string needs to be shortened, an ellipsis ("...") is added at the end of the abbreviated
// string. The abbreviation starts from the beginning of the string.
//
// Parameters:
//   - `str`: The input string to abbreviate.
//   - `maxWidth`: The maximum allowed length for the abbreviated string, including the ellipsis.
//
// Returns:
//   - A string shortened to `maxWidth` characters with an ellipsis if abbreviation is required.
//     If the input string is shorter than or equal to `maxWidth`, the original string is returned.
//
// Example:
//
//	input := "This is a long string."
//	output := Abbreviate(input, 10) // output will be "This is..."
func Abbreviate(str string, maxWidth int) string {
	if IsEmpty(str) {
		return str
	}
	return AbbreviateWithOffset(str, 0, maxWidth)
}

// AbbreviateWithOffset abbreviates a string by shortening it to a specified `maxWidth` starting at a given offset.
//
// This function shortens the input string `str` to the length specified by `maxWidth`, starting from the specified
// `offset`. An ellipsis ("...") is added at the beginning or end of the abbreviated string depending on the offset.
// If the string is already shorter than `maxWidth`, it is returned as is. If the offset is past the length of the string,
// the function adjusts the offset to ensure the abbreviation can be made.
//
// Parameters:
//   - `str`: The input string to abbreviate.
//   - `offset`: The index from which the abbreviation should start.
//   - `maxWidth`: The maximum allowed length for the abbreviated string, including the ellipsis.
//
// Returns:
//   - A string shortened to `maxWidth` characters with an ellipsis. If the input string is shorter
//     than or equal to `maxWidth`, or the offset is invalid, the original string is returned.
//
// Example:
//
//	input := "This is a long string."
//	output := AbbreviateWithOffset(input, 5, 10) // output will be "...s is a..."
//
// Notes:
//   - The function ensures that the abbreviation makes sense even if the offset or maxWidth
//     values are adjusted to fit the input string's length.
func AbbreviateWithOffset(str string, offset int, maxWidth int) string {
	size := len(str)
	if IsEmpty(str) || maxWidth < 4 || size <= maxWidth {
		return str
	}
	if offset > size {
		offset = size
	}
	if size-offset < maxWidth-3 {
		offset = size - (maxWidth - 3)
	}
	abbrevMarker := "..."
	if offset <= 4 {
		return str[0:maxWidth-3] + abbrevMarker
	}
	if maxWidth < 7 {
		return str
	}
	if offset+maxWidth-3 < size {
		return abbrevMarker + Abbreviate(str[offset:], maxWidth-3)
	}
	return abbrevMarker + str[size-(maxWidth-3):]
}

// EndsWith checks if the given string `str` ends with the specified suffix `suffix`.
//
// This function performs a case-sensitive comparison to determine if `str` ends with `suffix`.
// If either the string or the suffix is empty, it will return true only if both are empty.
// It uses the internal helper function `internalEndsWith` to handle the logic.
//
// Parameters:
//   - `str`: The input string to check.
//   - `suffix`: The suffix to check for at the end of the string.
//
// Returns:
//   - `true` if `str` ends with the specified `suffix`, `false` otherwise.
//
// Example:
//
//	input := "hello world"
//	suffix := "world"
//	result := EndsWith(input, suffix) // result will be true
func EndsWith(str string, suffix string) bool {
	return internalEndsWith(str, suffix, false)
}

// EndsWithIgnoreCase performs a case-insensitive check to determine if a string ends with a specified suffix.
//
// This function converts both `str` and `suffix` to lowercase before checking if `str` ends with `suffix`.
// Like `EndsWith`, it will return true if both the string and the suffix are empty.
// It uses the internal helper function `internalEndsWith` to handle the case-insensitive comparison.
//
// Parameters:
//   - `str`: The input string to check.
//   - `suffix`: The suffix to check for at the end of the string.
//
// Returns:
//   - `true` if `str` ends with the specified `suffix` (case-insensitively), `false` otherwise.
//
// Example:
//
//	input := "hello world"
//	suffix := "WORLD"
//	result := EndsWithIgnoreCase(input, suffix) // result will be true
func EndsWithIgnoreCase(str string, suffix string) bool {
	return internalEndsWith(str, suffix, true)
}

// AppendIfMissing appends a specified suffix to a string if it is not already present.
//
// This function checks if the input string `str` ends with a specified `suffix` or any of the additional
// suffixes provided in the variadic `suffixes` parameter. If none of the specified suffixes are present,
// it appends the main suffix to `str`. The comparison is case-sensitive.
//
// Parameters:
//   - `str`: The input string to check and potentially modify.
//   - `suffix`: The main suffix to append if `str` does not already end with it.
//   - `suffixes`: An optional variadic parameter specifying additional suffixes to check against.
//
// Returns:
//   - The original string if it already ends with the specified suffix (or one of the provided suffixes).
//     Otherwise, it returns the string with the `suffix` appended.
//
// Example:
//
//	input := "example"
//	suffix := "txt"
//	result := AppendIfMissing(input, suffix) // result will be "exampletxt" if input does not already end with "txt".
//
// Notes:
//   - This function is case-sensitive.
func AppendIfMissing(str string, suffix string, suffixes ...string) string {
	return internalAppendIfMissing(str, suffix, false, suffixes...)
}

// AppendIfMissingIgnoreCase appends a specified suffix to a string if it is not already present,
// using case-insensitive comparison.
//
// This function checks if the input string `str` ends with a specified `suffix` or any of the additional
// suffixes provided in the variadic `suffixes` parameter. If none of the specified suffixes are present,
// it appends the main suffix to `str`. The comparison is case-insensitive.
//
// Parameters:
//   - `str`: The input string to check and potentially modify.
//   - `suffix`: The main suffix to append if `str` does not already end with it.
//   - `suffixes`: An optional variadic parameter specifying additional suffixes to check against.
//
// Returns:
//   - The original string if it already ends with the specified suffix (or one of the provided suffixes).
//     Otherwise, it returns the string with the `suffix` appended.
//
// Example:
//
//	input := "example"
//	suffix := "Txt"
//	result := AppendIfMissingIgnoreCase(input, suffix) // result will be "exampleTxt" if input does not already end with "Txt" or any suffix provided.
//
// Notes:
//   - This function is case-insensitive, so "Txt" and "txt" will be considered equivalent.
func AppendIfMissingIgnoreCase(str string, suffix string, suffixes ...string) string {
	return internalAppendIfMissing(str, suffix, true, suffixes...)
}

// Capitalize capitalizes the first letter of the given string.
//
// This function converts the first character of the input string `str` to its uppercase form,
// while leaving the rest of the string unchanged. If the input string is empty, it returns the string as is.
//
// Parameters:
//   - `str`: The input string that needs to have its first letter capitalized.
//
// Returns:
//   - A new string with the first character in uppercase and the rest of the string unchanged.
//     If the input string is empty, the original string is returned.
//
// Example:
//
//	input := "hello"
//	output := Capitalize(input) // output will be "Hello"
//
// Notes:
//   - The function only modifies the first character. If the first character is already uppercase,
//     or if it's not a letter, the string is returned unchanged.
func Capitalize(str string) string {
	if IsEmpty(str) {
		return str
	}
	for i, c := range str {
		return string(unicode.ToUpper(c)) + str[i+1:]
	}
	return ""
}

// Title converts the first letter of each word in the given string to title case.
// A word is defined as a sequence of characters separated by whitespace or punctuation.
//
// This function iterates through each character in the input string `str` and checks if it is
// the first character of a word. If it is, the character is converted to title case using
// `unicode.ToTitle`. The function uses a helper function `isSeparator` to determine if the
// previous character is a separator (whitespace or punctuation).
//
// Parameters:
//   - `str`: The input string to be converted to title case.
//
// Returns:
//   - A new string with the first letter of each word converted to title case.
//
// Example:
//
//	input := "hello world! this is a test."
//	output := Title(input) // output will be "Hello World! This Is A Test."
func Title(str string) string {
	if IsEmpty(str) {
		return str
	}
	prev := ' '
	return strings.Map(
		func(r rune) rune {
			if isSeparator(prev) {
				prev = r
				return unicode.ToTitle(r)
			}
			prev = r
			return r
		},
		str)
}

// Chomp removes the trailing newline or carriage return characters from the given string.
//
// This function trims newline (`\n`), carriage return (`\r`), or the combination of both (`\r\n`) from the
// end of the input string `str`. If the string consists solely of one newline or carriage return character,
// it returns an empty string. If the string does not end with any of these characters, the original string
// is returned unchanged.
//
// Parameters:
//   - `str`: The input string from which trailing newline or carriage return characters will be removed.
//
// Returns:
//   - A new string with the trailing newline or carriage return characters removed, or the original string
//     if none are found.
//
// Example:
//
//	input := "hello\n"
//	output := Chomp(input) // output will be "hello"
//
//	input := "world\r\n"
//	output := Chomp(input) // output will be "world"
//
//	input := "test"
//	output := Chomp(input) // output will be "test" (no change)
//
// Notes:
//   - This function removes at most one newline or carriage return sequence from the end of the string.
//   - It handles the common line-ending conventions: `\n` (Unix), `\r\n` (Windows), and `\r` (legacy Mac).
func Chomp(str string) string {
	if IsEmpty(str) {
		return str
	}
	if len(str) == 1 && (str[0:1] == "\n" || str[0:1] == "\r") {
		return ""
	}
	if EndsWith(str, "\r\n") {
		return str[:len(str)-2]
	} else if EndsWith(str, "\n") || EndsWith(str, "\r") {
		return str[:len(str)-1]
	} else {
		return str
	}
}

// Chop removes the last character or the last newline sequence from the given string.
//
// This function works similarly to `Chomp` in that it removes newline sequences (`\n`, `\r`, `\r\n`),
// but if the input string does not end with a newline, it removes the last character instead.
// If the string is empty, it returns the string unchanged.
//
// Parameters:
//   - `str`: The input string from which the last character or newline sequence will be removed.
//
// Returns:
//   - A new string with either the trailing newline sequence or the last character removed, or
//     the original string if it is empty.
//
// Example:
//
//	input := "hello\n"
//	output := Chop(input) // output will be "hello" (removes the newline)
//
//	input := "world"
//	output := Chop(input) // output will be "worl" (removes the last character)
//
// Notes:
//   - `Chop` first attempts to remove a newline sequence using `Chomp`. If a newline is found and removed,
//     it returns the result. Otherwise, it removes the last character of the string.
//   - If the string contains only one character, `Chop` will return an empty string.
func Chop(str string) string {
	if IsEmpty(str) {
		return str
	}
	sc := Chomp(str)
	if len(str) > len(sc) {
		return sc
	}
	return str[0 : len(str)-1]
}

// Contains checks if the given string contains a specified substring.
//
// This function checks if the input string `str` contains the substring `search`.
// It performs a case-sensitive comparison to determine if the substring is present.
//
// Parameters:
//   - `str`: The input string in which to search for the substring.
//   - `search`: The substring to search for within `str`.
//
// Returns:
//   - `true` if the `search` substring is found within `str`, `false` otherwise.
//
// Example:
//
//	input := "hello world"
//	search := "world"
//	result := Contains(input, search) // result will be true as "world" is present in "hello world".
func Contains(str string, search string) bool {
	return strings.Contains(str, search)
}

// ContainsAny checks if the given string contains any of the specified substrings.
//
// This function checks if the input string `str` contains any of the substrings provided
// in the variadic `search` parameter. It returns `true` as soon as one of the substrings is found.
// The comparison is case-sensitive.
//
// Parameters:
//   - `str`: The input string in which to search for substrings.
//   - `search`: A variadic parameter that accepts one or more substrings to search for.
//
// Returns:
//   - `true` if any of the specified substrings are found within `str`, `false` otherwise.
//
// Example:
//
//	input := "hello world"
//	search := []string{"foo", "world"}
//	result := ContainsAny(input, search...) // result will be true because "world" is found in "hello world".
//
// Notes:
//   - The function iterates over the `search` substrings and checks each one individually.
//   - The search is case-sensitive.
func ContainsAny(str string, search ...string) bool {
	for _, s := range search {
		if Contains(str, (string)(s)) {
			return true
		}
	}
	return false
}

// ContainsIgnoreCase checks if the string contains the specified substring,
// ignoring case differences.
//
// This function converts both the input string `str` and the `search` substring
// to lowercase before performing the check, ensuring that the comparison is
// case-insensitive.
//
// Parameters:
//   - `str`: The input string in which to search for the substring.
//   - `search`: The substring to search for within `str`.
//
// Returns:
//   - `true` if the `search` substring is found within `str` (case-insensitive),
//     `false` otherwise.
//
// Example:
//
//	input := "Hello World"
//	search := "world"
//	result := ContainsIgnoreCase(input, search) // result will be true as "world" is found in "Hello World" ignoring case.
func ContainsIgnoreCase(str string, search string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(search))
}

// ContainsNone checks if the string contains none of the specified substrings.
//
// This function checks the input string `str` to ensure that it does not contain
// any of the substrings provided in the variadic `search` parameter. It returns
// `true` if none of the substrings are present and `false` if at least one is found.
//
// Parameters:
//   - `str`: The input string to check against the substrings.
//   - `search`: A variadic parameter that accepts one or more substrings to check for.
//
// Returns:
//   - `true` if `str` does not contain any of the specified substrings,
//     `false` otherwise.
//
// Example:
//
//	input := "hello world"
//	search := []string{"foo", "bar"}
//	result := ContainsNone(input, search...) // result will be true because neither "foo" nor "bar" are present in "hello world".
func ContainsNone(str string, search ...string) bool {
	return !ContainsAny(str, search...)
}

// ContainsAnyCharacter checks if the string contains any of the characters
// specified in the `search` string.
//
// This function iterates through each character in the `search` string and checks
// if any of those characters are present in the input string `str`.
//
// Parameters:
//   - `str`: The input string in which to search for characters.
//   - `search`: A string containing characters to search for within `str`.
//
// Returns:
//   - `true` if any character from `search` is found in `str`, `false` otherwise.
//
// Example:
//
//	input := "hello"
//	search := "aeiou"
//	result := ContainsAnyCharacter(input, search) // result will be true because "hello" contains the character 'e'.
func ContainsAnyCharacter(str string, search string) bool {
	for _, c := range search {
		if Contains(str, (string)(c)) {
			return true
		}
	}
	return false
}

// ContainsNoneCharacter checks if the string contains none of the characters
// specified in the `search` string.
//
// This function calls `ContainsAnyCharacter` to determine if any character
// from `search` is present in the input string `str`. It returns `true` if
// none of the characters are found and `false` if at least one is found.
//
// Parameters:
//   - `str`: The input string to check against the characters.
//   - `search`: A string containing characters to check for within `str`.
//
// Returns:
//   - `true` if `str` does not contain any characters from `search`,
//     `false` otherwise.
//
// Example:
//
//	input := "hello"
//	search := "xyz"
//	result := ContainsNoneCharacter(input, search) // result will be true because "hello" contains none of the characters 'x', 'y', or 'z'.
func ContainsNoneCharacter(str string, search string) bool {
	return !ContainsAnyCharacter(str, search)
}

// ContainsOnly checks if a string contains only the specified characters.
//
// This function evaluates the input string `str` and verifies that every character
// in the string is present in the provided variadic parameter `search`. If the input
// string is empty, the function returns `false`. If all characters in the string
// are found in the `search` slice, it returns `true`; otherwise, it returns `false`.
//
// Parameters:
//   - `str`: The input string to be checked.
//   - `search`: A variadic parameter that accepts one or more characters to check
//     against the characters in `str`.
//
// Returns:
//   - `true` if all characters in `str` are present in `search`; `false` otherwise.
//
// Example:
//
//	inputStr := "abc"
//	allowedChars := []string{"a", "b", "c", "d"}
//	result := ContainsOnly(inputStr, allowedChars...) // result will be true because "abc" contains only the allowed characters.
//
// Notes:
//   - The function performs a character-by-character check and is case-sensitive.
func ContainsOnly(str string, search ...string) bool {
	if IsEmpty(str) {
		return false
	}
	for _, c := range str {
		if !stringInSlice((string)(c), search) {
			return false
		}
	}
	return true
}

// IsAllLowerCase checks if the string contains only lowercase characters.
func IsAllLowerCase(str string) bool {
	if IsEmpty(str) {
		return false
	}
	for _, c := range str {
		if !unicode.IsLower(c) {
			return false
		}
	}
	return true
}

// IsAllUpperCase checks if the string contains only uppercase characters.
func IsAllUpperCase(str string) bool {
	if IsEmpty(str) {
		return false
	}
	for _, c := range str {
		if !unicode.IsUpper(c) {
			return false
		}
	}
	return true
}

// IsAlpha checks if the string contains only Unicode letters.
func IsAlpha(str string) bool {
	if IsEmpty(str) {
		return false
	}
	for _, c := range str {
		if !unicode.IsLetter(c) {
			return false
		}
	}
	return true
}

// IsAlphanumeric checks if a string contains only alphanumeric characters.
//
// This function evaluates the input string `str` and determines if it consists solely
// of letters (both uppercase and lowercase) and digits. If the string is empty,
// the function returns `false`. The check is performed character by character,
// returning `false` as soon as a non-alphanumeric character is found; if all
// characters are alphanumeric, it returns `true`.
//
// Parameters:
//   - `str`: The input string to be checked for alphanumeric content.
//
// Returns:
//   - `true` if the string contains only letters and digits; `false` otherwise.
//
// Example:
//
//	inputStr := "Hello123"
//	result := IsAlphanumeric(inputStr) // result will be true because "Hello123" consists only of letters and digits.
//
// Notes:
//   - This function is case-sensitive; both uppercase and lowercase letters are accepted.
//   - Special characters and whitespace will cause the function to return false.
func IsAlphanumeric(str string) bool {
	if IsEmpty(str) {
		return false
	}
	for _, c := range str {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

// IsAlphaSpace checks if a string contains only alphabetic characters and spaces.
//
// This function evaluates the input string `str` to determine if it consists solely
// of letters (both uppercase and lowercase) and whitespace characters. If the string
// is empty, the function returns `false`. The check is performed character by character,
// returning `false` as soon as a non-alphabetic and non-space character is found;
// if all characters are either alphabetic or spaces, it returns `true`.
//
// Parameters:
//   - `str`: The input string to be checked for alphabetic characters and spaces.
//
// Returns:
//   - `true` if the string contains only letters and spaces; `false` otherwise.
//
// Example:
//
//	inputStr := "Hello World"
//	result := IsAlphaSpace(inputStr) // result will be true because "Hello World" consists only of letters and spaces.
//
// Notes:
//   - This function is case-sensitive; both uppercase and lowercase letters are accepted.
//   - Any numeric characters, punctuation, or special characters will cause the function
//     to return false.
func IsAlphaSpace(str string) bool {
	if IsEmpty(str) {
		return false
	}
	for _, c := range str {
		if !unicode.IsLetter(c) && !unicode.IsSpace(c) {
			return false
		}
	}
	return true
}

// IsAlphanumericSpace checks if a string contains only alphanumeric characters and spaces.
//
// This function evaluates the input string `str` to determine if it consists solely
// of letters (both uppercase and lowercase), digits, and whitespace characters. If the
// string is empty, the function returns `false`. The check is performed character by
// character, returning `false` as soon as a non-alphanumeric and non-space character
// is found; if all characters are either alphanumeric or spaces, it returns `true`.
//
// Parameters:
//   - `str`: The input string to be checked for alphanumeric characters and spaces.
//
// Returns:
//   - `true` if the string contains only letters, digits, and spaces; `false` otherwise.
//
// Example:
//
//	inputStr := "Hello123 World"
//	result := IsAlphanumericSpace(inputStr) // result will be true because "Hello123 World" consists only of letters, digits, and spaces.
//
// Notes:
//   - This function is case-sensitive; both uppercase and lowercase letters are accepted.
//   - Any special characters or punctuation will cause the function to return false.
func IsAlphanumericSpace(str string) bool {
	if IsEmpty(str) {
		return false
	}
	for _, c := range str {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) && !unicode.IsSpace(c) {
			return false
		}
	}
	return true
}

// JoinBool concatenates a slice of boolean values into a single string,
// with each value separated by the specified separator.
//
// This function converts each boolean in the input slice `a` to its string representation
// ("true" or "false") using strconv.FormatBool, and then joins them together using the
// specified separator `sep`.
//
// Parameters:
//   - `a`: A slice of boolean values to be joined.
//   - `sep`: A string used to separate the boolean values in the output.
//
// Returns:
//   - A string that contains the boolean values joined together, separated by `sep`.
//
// Example:
//
//	bools := []bool{true, false, true}
//	result := JoinBool(bools, ", ") // result will be "true, false, true"
func JoinBool(a []bool, sep string) string {
	strs := make([]string, len(a))
	for idx, i := range a {
		strs[idx] = strconv.FormatBool(i)
	}
	return Join(strs, sep)
}

// JoinFloat64 concatenates a slice of float64 values into a single string,
// using the default format ('G') and bit size (32) for float conversion.
//
// This function converts each float64 in the input slice `a` to its string representation
// using strconv.FormatFloat with default formatting options, and then joins them together
// using the specified separator `sep`.
//
// Parameters:
//   - `a`: A slice of float64 values to be joined.
//   - `sep`: A string used to separate the float64 values in the output.
//
// Returns:
//   - A string that contains the float64 values joined together, separated by `sep`.
//
// Example:
//
//	floats := []float64{1.23, 4.56, 7.89}
//	result := JoinFloat64(floats, ", ") // result will be "1.23, 4.56, 7.89"
func JoinFloat64(a []float64, sep string) string {
	return JoinFloat64WithFormatAndPrecision(a, 'G', 32, sep)
}

// JoinFloat64WithFormatAndPrecision concatenates a slice of float64 values into a single string,
// allowing for custom formatting and precision during the conversion.
//
// This function converts each float64 in the input slice `a` to its string representation
// using strconv.FormatFloat with the specified format `fmt` and precision `precision`,
// and then joins them together using the specified separator `sep`.
//
// Parameters:
//   - `a`: A slice of float64 values to be joined.
//   - `fmt`: A byte that specifies the format for float conversion ('b', 'e', 'E', 'f', 'g', 'G').
//   - `precision`: An integer specifying the precision for float conversion (bit size).
//   - `sep`: A string used to separate the float64 values in the output.
//
// Returns:
//   - A string that contains the float64 values joined together, separated by `sep`.
//
// Example:
//
//	floats := []float64{1.2345, 6.7890}
//	result := JoinFloat64WithFormatAndPrecision(floats, 'f', 2, ", ") // result will be "1.23, 6.79"
func JoinFloat64WithFormatAndPrecision(a []float64, fmt byte, precision int, sep string) string {
	list := make([]string, len(a))
	for idx, i := range a {
		list[idx] = strconv.FormatFloat(i, fmt, -1, precision)
	}
	return Join(list, sep)
}

// JoinInt concatenates a slice of integers into a single string,
// with each integer separated by the specified separator.
//
// This function converts each integer in the input slice `a` to its string representation
// using strconv.Itoa, and then joins them together using the specified separator `sep`.
//
// Parameters:
//   - `a`: A slice of integers to be joined.
//   - `sep`: A string used to separate the integer values in the output.
//
// Returns:
//   - A string that contains the integer values joined together, separated by `sep`.
//
// Example:
//
//	ints := []int{1, 2, 3}
//	result := JoinInt(ints, ", ") // result will be "1, 2, 3"
func JoinInt(a []int, sep string) string {
	list := make([]string, len(a))
	for idx, i := range a {
		list[idx] = strconv.Itoa(i)
	}
	return Join(list, sep)
}

// JoinInt64 concatenates a slice of int64 values into a single string,
// with each int64 separated by the specified separator.
//
// This function converts each int64 in the input slice `a` to its string representation
// using strconv.FormatInt, and then joins them together using the specified separator `sep`.
//
// Parameters:
//   - `a`: A slice of int64 values to be joined.
//   - `sep`: A string used to separate the int64 values in the output.
//
// Returns:
//   - A string that contains the int64 values joined together, separated by `sep`.
//
// Example:
//
//	int64s := []int64{100, 200, 300}
//	result := JoinInt64(int64s, ", ") // result will be "100, 200, 300"
func JoinInt64(a []int64, sep string) string {
	list := make([]string, len(a))
	for idx, i := range a {
		list[idx] = strconv.FormatInt(i, 10)
	}
	return Join(list, sep)
}

// JoinUint64 concatenates a slice of uint64 values into a single string,
// with each uint64 separated by the specified separator.
//
// This function converts each uint64 in the input slice `ints` to its string representation
// using strconv.FormatUint, and then joins them together using the specified separator `sep`.
//
// Parameters:
//   - `ints`: A slice of uint64 values to be joined.
//   - `sep`: A string used to separate the uint64 values in the output.
//
// Returns:
//   - A string that contains the uint64 values joined together, separated by `sep`.
//
// Example:
//
//	uint64s := []uint64{1000, 2000, 3000}
//	result := JoinUint64(uint64s, ", ") // result will be "1000, 2000, 3000"
func JoinUint64(ints []uint64, sep string) string {
	list := make([]string, len(ints))
	for idx, i := range ints {
		list[idx] = strconv.FormatUint(i, 10)
	}
	return Join(list, sep)
}

// Mid extracts a substring of a specified size from the middle of a given string, starting at a specified position.
//
// This function returns a substring from the input string `str`, starting from the specified `pos`
// and extending for the specified number of `size` characters. If the `size` is negative, it returns an empty string.
// If the starting position is outside the bounds of the string, or if the input string is empty,
// an empty string is returned. If the specified `size` extends beyond the end of the string,
// the substring from the starting position to the end of the string is returned.
//
// Parameters:
//   - `str`: The input string from which to extract the substring.
//   - `pos`: The starting position in the string to begin extraction.
//   - `size`: The number of characters to extract from the string.
//
// Returns:
//   - A substring of the input string starting from `pos` and having length `size`.
//     If the starting position is invalid or the size is negative, returns an empty string.
//
// Example:
//
//	input := "Hello, World!"
//	result := Mid(input, 7, 5) // result will be "World"
func Mid(str string, pos int, size int) string {
	if str == "" || size < 0 || pos > len(str) {
		return ""
	}
	if pos < 0 {
		pos = 0
	}
	if len(str) <= pos+size {
		return str[pos:]
	}
	return str[pos : pos+size]
}

// Overlay replaces a portion of the input string with another string, starting from a specified start index
// and ending at a specified end index.
//
// This function takes the input string `str` and overlays it with the string `overlay`, replacing the
// portion of `str` that starts at index `start` and ends at index `end`. If the start index is out of bounds,
// it is adjusted to the nearest valid position. If the end index is out of bounds, it is set to the length
// of the string. If the start index is greater than the end index, the indices are swapped.
//
// Parameters:
//   - `str`: The original string to be modified.
//   - `overlay`: The string that will replace the specified portion of `str`.
//   - `start`: The starting index from which to begin the overlay.
//   - `end`: The ending index up to which the overlay will occur.
//
// Returns:
//   - A new string that results from overlaying `overlay` onto `str` between the specified indices.
//     If both `start` and `end` are out of bounds, returns the original string without modification.
//
// Example:
//
//	original := "Hello, World!"
//	newString := Overlay(original, "Gopher", 7, 12) // newString will be "Hello, Gopher!"
func Overlay(str string, overlay string, start int, end int) string {
	strLen := len(str)
	if start < 0 {
		start = 0
	}
	if start > strLen {
		start = strLen
	}
	if end < 0 {
		end = 0
	}
	if end > strLen {
		end = strLen
	}
	if start > end {
		start, end = end, start
	}
	return str[:start] + overlay + str[end:]
}

// Truncate truncates the input string s to the specified length.
// If the input string is longer than the specified length, it is truncated to the specified length.
// Otherwise, the original string is returned.
//
// Parameters:
//   - `s`: The input string to be truncated.
//   - `length`: The maximum length of the truncated string.
//
// Returns:
//   - A new string that is the truncated version of the input string.
//
// Example:
//
//	original := "Hello, World!"
//	newString := Truncate(original, 5) // newString will be "Hello..."
func Truncate(s string, length int) string {
	if IsEmpty(s) {
		return s
	}
	if length < 0 {
		return s
	}
	strategy := truncate.NewCutEllipsisStrategy()
	return strategy.Truncate(s, length)
}

// Remove removes all occurrences of a specified substring from the source string.
//
// This function searches for all instances of the substring `remove` within the input string `str`
// and removes them. If the input string is empty, the function returns it unchanged. The function
// utilizes a loop to find and remove each occurrence of the specified substring until none remain.
//
// Parameters:
//   - `str`: The source string from which the substring will be removed.
//   - `remove`: The substring to be removed from the source string.
//
// Returns:
//   - A new string with all occurrences of `remove` removed from `str`.
//     If the `remove` substring is not found, or if `str` is empty, the original string is returned unchanged.
//
// Example:
//
//	input := "Hello, world! Goodbye, world!"
//	result := Remove(input, "world") // result will be "Hello, ! Goodbye, !"
func Remove(str string, remove string) string {
	if IsEmpty(str) {
		return str
	}
	index := strings.Index(str, remove)
	for index > -1 {
		str = str[:index] + str[index+len(remove):]
		index = strings.Index(str, remove)
	}
	return str
}

// RemoveEnd removes a specified substring from the end of a source string,
// if the substring is found there. If the substring is not present at the end,
// the original source string is returned unchanged.
//
// Parameters:
//   - `str`: The source string from which the substring will be removed.
//   - `remove`: The substring to be removed from the end of the source string.
//
// Returns:
//   - A new string with the specified `remove` substring removed from the end
//     of `str`. If `str` is empty or `remove` is empty, the original string is returned.
//     If the substring is not found at the end, the original string is returned unchanged.
//
// Example:
//
//	input := "example.txt"
//	result := RemoveEnd(input, ".txt") // result will be "example" since ".txt" is at the end of the string.
func RemoveEnd(str string, remove string) string {
	if IsEmpty(str) || IsEmpty(remove) {
		return str
	}
	if EndsWith(str, remove) {
		return str[:len(str)-len(remove)]
	}
	return str
}

// RemoveEndIgnoreCase removes a specified substring from the end of a source string
// in a case-insensitive manner. If the substring is found at the end, it is removed;
// otherwise, the original string is returned unchanged.
//
// Parameters:
//   - `str`: The source string from which the substring will be removed.
//   - `remove`: The substring to be removed from the end of the source string.
//
// Returns:
//   - A new string with the specified `remove` substring removed from the end
//     of `str`. If `str` is empty or `remove` is empty, the original string is returned.
//     If the substring is not found at the end (case insensitive), the original string is returned unchanged.
//
// Example:
//
//	input := "example.TXT"
//	result := RemoveEndIgnoreCase(input, ".txt") // result will be "example" since ".txt" is found at the end of the string ignoring case.
func RemoveEndIgnoreCase(str string, remove string) string {
	if IsEmpty(str) || IsEmpty(remove) {
		return str
	}
	if EndsWithIgnoreCase(str, remove) {
		return str[:len(str)-len(remove)]
	}
	return str
}

// RemovePattern removes all substrings from the source string that match the specified
// regular expression pattern. If no matches are found, the original string is returned unchanged.
//
// Parameters:
//   - `str`: The source string from which substrings matching the pattern will be removed.
//   - `pattern`: A regular expression pattern that defines the substrings to be removed.
//
// Returns:
//   - A new string with all occurrences of substrings matching the given `pattern` removed.
//     If no matches are found, the original string is returned unchanged.
//
// Example:
//
//	input := "abc123xyz"
//	pattern := "[0-9]+" // matches all digits
//	result := RemovePattern(input, pattern) // result will be "abcxyz" since all digits are removed from the string.
func RemovePattern(str string, pattern string) string {
	return regexp.MustCompile(pattern).ReplaceAllString(str, "")
}

// RemoveStart removes a substring only if it is at the beginning of a source string,
// otherwise returns the source string
func RemoveStart(str string, remove string) string {
	if IsEmpty(str) || IsEmpty(remove) {
		return str
	}
	if StartsWith(str, remove) {
		return str[len(remove)+1:]
	}
	return str
}

// RemoveStartIgnoreCase removes a specified substring from the start of a string,
// ignoring case differences.
//
// This function checks if the input string `str` starts with the specified `remove`
// substring in a case-insensitive manner. If `str` starts with `remove`, it removes
// that substring from the start of `str` and returns the modified string. If `str`
// does not start with `remove`, or if either `str` or `remove` is empty, the function
// returns the original string unchanged.
//
// Parameters:
//   - `str`: The source string from which to remove the specified substring.
//   - `remove`: The substring to be removed from the start of `str`.
//
// Returns:
//   - The modified string with the `remove` substring removed from the start if it
//     was found; otherwise, returns the original string unchanged.
//
// Example:
//
//	input := "Hello, World!"
//	remove := "hello"
//	result := RemoveStartIgnoreCase(input, remove) // result will be "Hello, World!" since input does not start with "hello" (case-insensitive).
//
//	input2 := "Hello, World!"
//	remove2 := "Hello"
//	result2 := RemoveStartIgnoreCase(input2, remove2) // result2 will be ", World!" since input2 starts with "Hello" (case-insensitive).
//
// Notes:
//   - This function is case-insensitive, so it will remove the substring regardless of
//     the casing in `str` and `remove`.
func RemoveStartIgnoreCase(str string, remove string) string {
	if IsEmpty(str) || IsEmpty(remove) {
		return str
	}
	if StartsWithIgnoreCase(str, remove) {
		return str[len(remove)+1:]
	}
	return str
}

// StartsWith checks if a string starts with a specified prefix.
//
// This function determines whether the input string `str` begins with the specified
// `prefix` in a case-sensitive manner.
//
// Parameters:
//   - `str`: The source string to be checked for the specified prefix.
//   - `prefix`: The prefix string to look for at the start of `str`.
//
// Returns:
//   - `true` if `str` starts with `prefix` (case-sensitive).
//   - `false` if `str` does not start with `prefix`, or if either `str` or `prefix` is empty
//     (in which case it returns `true` only if both are empty).
//
// Example:
//
//	str := "Hello, World!"
//	prefix := "Hello"
//	result := StartsWith(str, prefix) // result will be true since str starts with the prefix "Hello".
//
// Notes:
//   - This function is case-sensitive; for case-insensitive checks, use a different
//     function that leverages the `internalStartsWith` with the `ignoreCase` flag set to true.
func StartsWith(str string, prefix string) bool {
	return internalStartsWith(str, prefix, false)
}

// StartsWithIgnoreCase checks if a string starts with a specified prefix,
// ignoring case differences.
//
// This function determines whether the input string `str` begins with the specified
// `prefix` in a case-insensitive manner.
//
// Parameters:
//   - `str`: The source string to be checked for the specified prefix.
//   - `prefix`: The prefix string to look for at the start of `str`.
//
// Returns:
//   - `true` if `str` starts with `prefix`, ignoring case differences.
//   - `false` if `str` does not start with `prefix`, or if either `str` or `prefix` is empty
//     (in which case it returns `true` only if both are empty).
//
// Example:
//
//	str := "Hello, World!"
//	prefix := "hello"
//	result := StartsWithIgnoreCase(str, prefix) // result will be true since str starts with the prefix "hello", ignoring case.
//
// Notes:
//   - This function is case-insensitive; if a case-sensitive check is required,
//     use the `StartsWith` function instead.
func StartsWithIgnoreCase(str string, prefix string) bool {
	return internalStartsWith(str, prefix, true)
}

// Repeat returns a new string consisting of the specified string repeated a given number of times.
//
// This function takes an input string `str` and an integer `repeat`, and constructs a new
// string that contains `str` concatenated `repeat` times. If `repeat` is less than or equal to
// zero, an empty string is returned.
//
// Parameters:
//   - `str`: The input string to be repeated.
//   - `repeat`: The number of times to repeat the string.
//
// Returns:
//   - A new string that contains `str` repeated `repeat` times. If `repeat` is 0 or negative,
//     an empty string is returned.
//
// Example:
//
//	input := "abc"
//	result := Repeat(input, 3) // result will be "abcabcabc".
//
//	input2 := "xyz"
//	result2 := Repeat(input2, 0) // result2 will be "" (empty string).
func Repeat(str string, repeat int) string {
	if IsEmpty(str) || repeat <= 0 {
		return str // return empty string
	}
	buff := ""
	for repeat > 0 {
		repeat = repeat - 1
		buff += str
	}
	return buff
}

// RepeatWithSeparator returns a new string consisting of the specified string repeated a given number of times,
// with a separator placed between each repetition.
//
// This function takes an input string `str`, a separator string `sep`, and an integer `repeat`,
// and constructs a new string that contains `str` concatenated `repeat` times, separated by `sep`.
// If `repeat` is less than or equal to zero, an empty string is returned.
//
// Parameters:
//   - `str`: The input string to be repeated.
//   - `sep`: The separator string to be placed between repetitions of `str`.
//   - `repeat`: The number of times to repeat the string.
//
// Returns:
//   - A new string that contains `str` repeated `repeat` times, with `sep` separating each occurrence.
//     If `repeat` is 0 or negative, an empty string is returned.
//
// Example:
//
//	input := "abc"
//	sep := "-"
//	result := RepeatWithSeparator(input, sep, 3) // result will be "abc-abc-abc".
//
//	input2 := "hello"
//	sep2 := ", "
//	result2 := RepeatWithSeparator(input2, sep2, 0) // result2 will be "" (empty string).
func RepeatWithSeparator(str string, sep string, repeat int) string {
	if IsEmpty(str) || repeat <= 0 {
		return str // return empty string
	}
	buff := ""
	for repeat > 0 {
		repeat = repeat - 1
		buff += str
		if repeat > 0 {
			buff += sep
		}
	}
	return buff
}

// ReverseDelimited reverses the order of substrings in a string that are separated by a specified delimiter.
//
// This function takes an input string `str` and a delimiter `del`, splits the string into
// substrings using the delimiter, reverses the order of these substrings, and then joins them
// back together using the same delimiter. If the input string is empty or does not contain the
// delimiter, it returns the original string unchanged.
//
// Parameters:
//   - `str`: The input string to be reversed, containing substrings separated by the specified delimiter.
//   - `del`: The delimiter used to split the input string into substrings.
//
// Returns:
//   - A new string with the order of substrings reversed. If the input string is empty or the delimiter
//     is not found, the original string is returned unchanged.
//
// Example:
//
//	input := "apple,banana,cherry"
//	delimiter := ","
//	result := ReverseDelimited(input, delimiter) // result will be "cherry,banana,apple".
//	input2 := "one|two|three"
//	delimiter2 := "|"
//	result2 := ReverseDelimited(input2, delimiter2) // result2 will be "three|two|one".
//
// Notes:
//   - The function modifies the order of the substrings but retains the original delimiter.
func ReverseDelimited(str string, del string) string {
	s := strings.Split(str, del)
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return Join(s, del)
}

// Strip removes whitespace from both the start and end of a string.
//
// This function takes an input string `str` and uses a regular expression to
// remove any whitespace characters (spaces, tabs, newlines, etc.) that appear
// at the beginning and the end of the string. If the input string is empty,
// it returns the string unchanged.
//
// Parameters:
//   - `str`: The input string from which leading and trailing whitespace will be removed.
//
// Returns:
//   - A new string with leading and trailing whitespace removed. If there is no
//     whitespace or if the string is empty, the original string is returned unchanged.
//
// Example:
//
//	input := "   Hello, World!   "
//	result := Strip(input) // result will be "Hello, World!".
func Strip(str string) string {
	return regexp.MustCompile(`^\s+|\s+$`).ReplaceAllString(str, "")
}

// StripEnd removes whitespace from the end of a string.
//
// This function takes an input string `str` and uses a regular expression to
// remove any whitespace characters that appear at the end of the string.
// If the input string is empty, it returns the string unchanged.
//
// Parameters:
//   - `str`: The input string from which trailing whitespace will be removed.
//
// Returns:
//   - A new string with trailing whitespace removed. If there is no trailing
//     whitespace or if the string is empty, the original string is returned unchanged.
//
// Example:
//
//	input := "Hello, World!   "
//	result := StripEnd(input) // result will be "Hello, World!".
func StripEnd(str string) string {
	return regexp.MustCompile(`\s+$`).ReplaceAllString(str, "")
}

// StripStart removes whitespace from the start of a string.
//
// This function takes an input string `str` and uses a regular expression to
// remove any whitespace characters that appear at the beginning of the string.
// If the input string is empty, it returns the string unchanged.
//
// Parameters:
//   - `str`: The input string from which leading whitespace will be removed.
//
// Returns:
//   - A new string with leading whitespace removed. If there is no leading
//     whitespace or if the string is empty, the original string is returned unchanged.
//
// Example:
//
//	input := "   Hello, World!"
//	result := StripStart(input) // result will be "Hello, World!".
func StripStart(str string) string {
	return regexp.MustCompile(`^\s+`).ReplaceAllString(str, "")
}

// SwapCase returns a new string with all uppercase letters converted to lowercase
// and all lowercase letters converted to uppercase.
//
// This function iterates through each character in the input string `str` and checks
// whether the character is uppercase or lowercase. If the character is lowercase,
// it converts it to uppercase; if it is uppercase, it converts it to lowercase.
// All non-alphabetic characters remain unchanged.
//
// Parameters:
//   - `str`: The input string whose letter case will be swapped.
//
// Returns:
//   - A new string with the cases of the letters swapped. If the input string is
//     empty or contains no alphabetic characters, it returns the original string
//     unchanged.
//
// Example:
//
//	input := "Hello, World!"
//	result := SwapCase(input) // result will be "hELLO, wORLD!".
//
// Notes:
//   - This function utilizes the `unicode` package to check the case of each character
//     and assumes that the functions `UpperCase` and `LowerCase` are defined to
//     convert characters to their respective cases.
func SwapCase(str string) string {
	buff := ""
	for _, c := range str {
		cs := (string)(c)
		if unicode.IsLower(c) {
			buff += strings.ToUpper(cs)
		} else if unicode.IsUpper(c) {
			buff += strings.ToLower(cs)
		} else {
			buff += cs
		}
	}
	return buff
}

// UnCapitalize changes the first letter of a string to lowercase.
//
// This function checks if the input string `str` is empty. If it is, the function
// returns the original string unchanged. If the first character of the string is
// uppercase, it converts that character to lowercase while leaving the rest of the
// string intact.
//
// Parameters:
//   - `str`: The input string to un-capitalize.
//
// Returns:
//   - A new string with the first character converted to lowercase if it was
//     originally uppercase. If the string is empty or if the first character is
//     already lowercase, the original string is returned unchanged.
//
// Example:
//
//	input := "Hello"
//	result := UnCapitalize(input) // result will be "hello".
//
// Notes:
//   - This function assumes that the input string contains at least one character
//     if it is not empty.
func UnCapitalize(str string) string {
	if IsEmpty(str) {
		return str
	}
	r := []rune(str)
	if unicode.IsUpper(r[0]) {
		return string(unicode.ToLower(r[0])) + string(r[1:])
	}
	return str
}

// Wrap wraps a given string with another string.
//
// This function checks if the input string `str` is empty. If it is, the function
// returns the original string unchanged. If not, it concatenates the `wrapWith`
// string to both the beginning and the end of `str`, effectively wrapping it.
//
// Parameters:
//   - `str`: The input string to be wrapped.
//   - `wrapWith`: The string to wrap around `str`.

// Returns:
//   - A new string that consists of `wrapWith` followed by `str` and then
//     another `wrapWith`. If the input string is empty, the original string is
//     returned unchanged.
//
// Example:
//
//	input := "Hello"
//	wrapWith := "***"
//	result := Wrap(input, wrapWith) // result will be "***Hello***".
//
// Notes:
//   - This function does not modify the original string; it creates and returns
//     a new string instead.
func Wrap(str string, wrapWith string) string {
	if IsEmpty(str) {
		return str
	}
	return wrapWith + str + wrapWith
}

// PrependIfMissing prepends a specified prefix to the start of a string if the string does not already start with that prefix
// or any additional prefixes provided.
//
// This function checks if the input string `str` starts with the specified `prefix`. If it does not, it will check
// against any prefixes provided in the variadic `prefixes` parameter. If `str` does not start with any of the
// specified prefixes, the function prepends the `prefix` to `str` and returns the modified string. If `prefix`
// is empty or if `str` already starts with `prefix` or any of the provided prefixes, the original string is returned.
//
// Parameters:
//   - `str`: The input string to check and potentially modify.
//   - `prefix`: The main prefix to prepend if `str` does not already start with it.
//   - `prefixes`: An optional variadic parameter specifying additional prefixes to check against.
//
// Returns:
//   - The original string if it already starts with the specified prefix (or one of the provided prefixes).
//     Otherwise, it returns the string with the `prefix` prepended.
//
// Example:
//
//	input := "example"
//	prefix := "test_"
//	result := PrependIfMissing(input, prefix) // result will be "test_example" if input does not already start with "test_".
//
// Notes:
//   - This function is case-sensitive.
func PrependIfMissing(str string, prefix string, prefixes ...string) string {
	return prependIfMissing(str, prefix, false, prefixes...)
}

// PrependIfMissingIgnoreCase prepends a specified prefix to the start of a string if the string does not already start
// (case-insensitively) with that prefix or any additional prefixes provided.
//
// This function performs the same checks as `PrependIfMissing`, but it ignores case when checking for the presence
// of `prefix` and the additional prefixes. If `str` does not start with `prefix` or any of the provided prefixes,
// the function prepends `prefix` to `str` and returns the modified string. If `prefix` is empty or if `str`
// already starts with `prefix` or any of the provided prefixes (ignoring case), the original string is returned.
//
// Parameters:
//   - `str`: The input string to check and potentially modify.
//   - `prefix`: The main prefix to prepend if `str` does not already start with it.
//   - `prefixes`: An optional variadic parameter specifying additional prefixes to check against.
//
// Returns:
//   - The original string if it already starts with the specified prefix (or one of the provided prefixes,
//     ignoring case). Otherwise, it returns the string with the `prefix` prepended.
//
// Example:
//
//	input := "example"
//	prefix := "Test_"
//	result := PrependIfMissingIgnoreCase(input, prefix) // result will be "Test_example" if input does not already start with "Test_" or any prefix provided.
//
// Notes:
//   - This function is case-insensitive, so "Test_" and "test_" will be considered equivalent.
func PrependIfMissingIgnoreCase(str string, prefix string, prefixes ...string) string {
	return prependIfMissing(str, prefix, true, prefixes...)
}

// ContainsAllWords checks if the given string contains all specified words.
//
// This function verifies if the input string `str` includes all words
// provided in the variadic parameter `words`. It uses a regular
// expression to match whole words only (considering word boundaries).
//
// Parameters:
//   - `str`: The input string to search within.
//   - `words`: A variadic parameter containing words to check for in the input string.
//
// Returns:
//   - `true` if the input string contains all specified words; otherwise, returns `false`.
//   - Returns `false` if `str` is empty or if no words are provided.
//
// Example:
//
//	str := "The quick brown fox"
//	words := []string{"quick", "fox"}
//	result := ContainsAllWords(str, words...) // result will be true
func ContainsAllWords(str string, words ...string) bool {
	if str == "" || len(words) == 0 {
		return false
	}
	found := 0
	for _, word := range words {
		if regexp.MustCompile(`.*\b` + word + `\b.*`).MatchString(str) {
			found++
		}
	}
	return found == len(words)
}

// Initials extracts the initial letters from each word in the given string.
//
// This function returns the first letter of the string and the first letters
// of all words following defined delimiters as a new string. The case of
// the letters is not changed in the output.
//
// Parameters:
//   - `str`: The input string from which to extract initials.
//
// Returns:
//   - A string containing the initials of the words in the input string.
//   - Returns an empty string if the input string is empty.
//
// Example:
//
//	str := "Hello World"
//	result := Initials(str) // result will be "HW"
func Initials(str string) string {
	return InitialsDelimited(str, nil...)
}

// InitialsDelimited extracts the initial letters from each word in the given string,
// considering specified delimiters.
//
// This function returns the first letter of the string and the first letters
// of all words after the defined delimiters as a new string. The case of
// the letters is not changed in the output.
//
// Parameters:
//   - `str`: The input string from which to extract initials.
//   - `delimiters`: A variadic parameter that specifies delimiters to consider
//     when determining word boundaries. If none are provided, whitespace is used.
//
// Returns:
//   - A string containing the initials of the words in the input string.
//   - Returns an empty string if the input string is empty or if no delimiters
//     are provided.
//
// Example:
//
//	str := "Hello, World! How are you?"
//	delimiters := []string{",", "!", "?"}
//	result := InitialsDelimited(str, delimiters...) // result will be "HWH"
func InitialsDelimited(str string, delimiters ...string) string {
	if IsEmpty(str) || (delimiters != nil && len(delimiters) == 0) {
		return str
	}
	strLen := len(str)
	buff := make([]rune, strLen/2+1)
	count := 0
	lastWasGap := true
	for _, ch := range str {
		if isDelimiter(ch, delimiters...) {
			lastWasGap = true
		} else if lastWasGap {
			buff[count] = ch
			count++
			lastWasGap = false
		}
	}
	return string(buff[:count])
}

// IsAllEmpty checks if all of the provided strings are empty or consist solely of whitespace characters.
//
// This function takes a variadic number of string arguments and iterates through
// each string to verify that all of them are empty (i.e., have a length of zero
// or consist solely of whitespace characters). If any string in the provided list
// is found to be non-empty, the function immediately returns false. If all strings
// are empty, it returns true.
//
// Parameters:
//   - ss: A variadic parameter that allows passing multiple strings to be checked.
//
// Returns:
//   - true if all of the provided strings are empty; false if at least one
//     string is non-empty.
//
// Example:
//
//	result := IsAllEmpty("", "   ", "")    // result will be true because all strings are empty or whitespace.
//	result2 := IsAllEmpty("", "hello", "") // result2 will be false because one string is non-empty.
func IsAllEmpty(ss ...string) bool {
	for _, s := range ss {
		if !IsEmpty(s) {
			return false
		}
	}
	return true
}

// StartsWithAny checks if the string s starts with any of the specified prefixes.
//
// This function iterates through each prefix and returns true as soon as a match
// is found. If none of the prefixes match, it returns false.
//
// Parameters:
//   - s: The input string to check.
//   - prefixes: A variadic list of prefix strings to test.
//
// Returns:
//   - true if s starts with at least one of the provided prefixes; false otherwise.
//
// Example:
//
//	result := StartsWithAny("Hello", "Hi", "He")  // result will be true
//	result = StartsWithAny("Hello", "Foo", "Bar") // result will be false
func StartsWithAny(s string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}

// EndsWithAny checks if the string s ends with any of the specified suffixes.
//
// This function iterates through each suffix and returns true as soon as a match
// is found. If none of the suffixes match, it returns false.
//
// Parameters:
//   - s: The input string to check.
//   - suffixes: A variadic list of suffix strings to test.
//
// Returns:
//   - true if s ends with at least one of the provided suffixes; false otherwise.
//
// Example:
//
//	result := EndsWithAny("Hello", "lo", "llo")   // result will be true
//	result = EndsWithAny("Hello", "Foo", "Bar")   // result will be false
func EndsWithAny(s string, suffixes ...string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(s, suffix) {
			return true
		}
	}
	return false
}

// RemoveAll removes all occurrences of the specified substring from the input string.
//
// This function replaces every instance of the remove substring with an empty string,
// effectively deleting all occurrences from the original string.
//
// Parameters:
//   - s: The input string from which substrings will be removed.
//   - remove: The substring to remove from s.
//
// Returns:
//   - A new string with all occurrences of remove deleted.
//
// Example:
//
//	result := RemoveAll("a-b-c", "-")         // result will be "abc"
//	result = RemoveAll("hello world", "o")    // result will be "hell wrld"
func RemoveAll(s string, remove string) string {
	return strings.ReplaceAll(s, remove, "")
}

// Replace replaces the first n non-overlapping occurrences of old with new in the string s.
// If n < 0, there is no limit on the number of replacements.
//
// Parameters:
//   - s: The input string in which replacements will be made.
//   - old: The substring to be replaced.
//   - new: The replacement substring.
//   - n: The maximum number of replacements to perform. Use -1 for unlimited.
//
// Returns:
//   - A new string with up to n replacements applied.
//
// Example:
//
//	result := Replace("hello", "l", "L", 1) // result will be "heLlo"
//	result = Replace("hello", "l", "L", -1) // result will be "heLLo"
func Replace(s string, old string, new string, n int) string {
	return strings.Replace(s, old, new, n)
}

// ReplaceAll replaces all non-overlapping occurrences of old with new in the string s.
//
// Parameters:
//   - s: The input string in which replacements will be made.
//   - old: The substring to be replaced.
//   - new: The replacement substring.
//
// Returns:
//   - A new string with all occurrences of old replaced by new.
//
// Example:
//
//	result := ReplaceAll("hello", "l", "L") // result will be "heLLo"
func ReplaceAll(s string, old string, new string) string {
	return strings.ReplaceAll(s, old, new)
}

// ReplaceIgnoreCase replaces the first occurrence of old with new in the string s,
// performing a case-insensitive match. If old is not found (case-insensitively),
// the original string is returned unchanged.
//
// Parameters:
//   - s: The input string in which the replacement will be made.
//   - old: The substring to search for (case-insensitive).
//   - new: The replacement substring.
//
// Returns:
//   - A new string with the first case-insensitive match of old replaced by new.
//
// Example:
//
//	result := ReplaceIgnoreCase("Hello", "hello", "Hi") // result will be "Hi"
//	result = ReplaceIgnoreCase("WORLD", "world", "Go")  // result will be "Go"
//	result = ReplaceIgnoreCase("abc", "xyz", "123")     // result will be "abc" (no match)
func ReplaceIgnoreCase(s string, old string, new string) string {
	lower := strings.ToLower(s)
	oldLower := strings.ToLower(old)
	idx := strings.Index(lower, oldLower)
	if idx < 0 {
		return s
	}
	return s[:idx] + new + s[idx+len(old):]
}

// PadLeft pads the string s on the left side with the pad string until the total
// rune length reaches at least size. If the string is already longer than or equal
// to size, it is returned unchanged.
//
// Parameters:
//   - s: The input string to pad.
//   - size: The desired minimum rune length of the result.
//   - pad: The string to use as padding. If empty, the original string is returned.
//
// Returns:
//   - A new string padded on the left to the specified size.
//
// Example:
//
//	result := PadLeft("5", 3, "0")     // result will be "005"
//	result = PadLeft("hello", 3, "0")  // result will be "hello" (already >= 3)
//	result = PadLeft("1", 5, "ab")     // result will be "abab1"
func PadLeft(s string, size int, pad string) string {
	if IsEmpty(pad) {
		return s
	}
	sLen := utf8.RuneCountInString(s)
	if sLen >= size {
		return s
	}
	padLen := utf8.RuneCountInString(pad)
	var sb strings.Builder
	remaining := size - sLen
	sb.Grow(len(s) + remaining*len(pad)/padLen + len(pad))
	for remaining > 0 {
		if padLen <= remaining {
			sb.WriteString(pad)
			remaining -= padLen
		} else {
			sb.WriteString(string([]rune(pad)[:remaining]))
			remaining = 0
		}
	}
	sb.WriteString(s)
	return sb.String()
}

// PadRight pads the string s on the right side with the pad string until the total
// rune length reaches at least size. If the string is already longer than or equal
// to size, it is returned unchanged.
//
// Parameters:
//   - s: The input string to pad.
//   - size: The desired minimum rune length of the result.
//   - pad: The string to use as padding. If empty, the original string is returned.
//
// Returns:
//   - A new string padded on the right to the specified size.
//
// Example:
//
//	result := PadRight("5", 3, "0")     // result will be "500"
//	result = PadRight("hello", 3, "0")  // result will be "hello" (already >= 3)
//	result = PadRight("1", 5, "ab")     // result will be "1abab"
func PadRight(s string, size int, pad string) string {
	if IsEmpty(pad) {
		return s
	}
	sLen := utf8.RuneCountInString(s)
	if sLen >= size {
		return s
	}
	padLen := utf8.RuneCountInString(pad)
	var sb strings.Builder
	remaining := size - sLen
	sb.Grow(len(s) + remaining*len(pad)/padLen + len(pad))
	sb.WriteString(s)
	for remaining > 0 {
		if padLen <= remaining {
			sb.WriteString(pad)
			remaining -= padLen
		} else {
			sb.WriteString(string([]rune(pad)[:remaining]))
			remaining = 0
		}
	}
	return sb.String()
}

// Center centres the string s within a field of the specified size by padding equally
// on both sides with the pad string. If the padding does not divide evenly, the extra
// character is placed on the right side.
//
// Parameters:
//   - s: The input string to centre.
//   - size: The desired total rune length of the result.
//   - pad: The string to use as padding. If empty, the original string is returned.
//
// Returns:
//   - A new string centred within the specified size.
//
// Example:
//
//	result := Center("Hi", 6, " ")  // result will be "  Hi  "
//	result = Center("Hi", 7, "-")   // result will be "--Hi---"
//	result = Center("Hello", 3, " ") // result will be "Hello" (already >= 3)
func Center(s string, size int, pad string) string {
	if IsEmpty(pad) {
		return s
	}
	sLen := utf8.RuneCountInString(s)
	if sLen >= size {
		return s
	}
	total := size - sLen
	leftPad := total / 2
	rightPad := total - leftPad
	return PadLeft(s, sLen+leftPad, pad)[:0] +
		padRepeat(pad, leftPad) + s + padRepeat(pad, rightPad)
}

// Unwrap removes the specified wrapper string from both the beginning and the end
// of the input string s. Both the leading and trailing wrapper must be present for
// the unwrapping to occur; otherwise the original string is returned unchanged.
//
// Parameters:
//   - s: The input string to unwrap.
//   - wrapper: The string to remove from both ends.
//
// Returns:
//   - A new string with the wrapper removed from both sides, or the original
//     string if it is not wrapped.
//
// Example:
//
//	result := Unwrap("***hello***", "***")  // result will be "hello"
//	result = Unwrap("'hello'", "'")         // result will be "hello"
//	result = Unwrap("hello***", "***")      // result will be "hello***" (not wrapped on both sides)
func Unwrap(s string, wrapper string) string {
	if IsEmpty(wrapper) {
		return s
	}
	if strings.HasPrefix(s, wrapper) && strings.HasSuffix(s, wrapper) {
		wLen := len(wrapper)
		if len(s) >= 2*wLen {
			return s[wLen : len(s)-wLen]
		}
	}
	return s
}

// DefaultIfEmpty returns the provided default value if the input string s is empty
// or consists solely of whitespace characters. Otherwise, it returns s unchanged.
//
// Parameters:
//   - s: The input string to check.
//   - defaultValue: The fallback value to return if s is empty.
//
// Returns:
//   - s if it is non-empty, or defaultValue otherwise.
//
// Example:
//
//	result := DefaultIfEmpty("", "default")      // result will be "default"
//	result = DefaultIfEmpty("value", "default")  // result will be "value"
func DefaultIfEmpty(s string, defaultValue string) string {
	if IsEmpty(s) {
		return defaultValue
	}
	return s
}

// DefaultIfBlank returns the provided default value if the input string s is blank
// (empty or contains only whitespace characters). Otherwise, it returns s unchanged.
//
// Parameters:
//   - s: The input string to check.
//   - defaultValue: The fallback value to return if s is blank.
//
// Returns:
//   - s if it is non-blank, or defaultValue otherwise.
//
// Example:
//
//	result := DefaultIfBlank("   ", "default")    // result will be "default"
//	result = DefaultIfBlank("value", "default")   // result will be "value"
func DefaultIfBlank(s string, defaultValue string) string {
	if IsBlank(s) {
		return defaultValue
	}
	return s
}

// Equals checks if two strings are exactly equal.
//
// Parameters:
//   - a: The first string.
//   - b: The second string.
//
// Returns:
//   - true if a and b are identical; false otherwise.
//
// Example:
//
//	result := Equals("hello", "hello") // result will be true
//	result = Equals("hello", "world")  // result will be false
func Equals(a string, b string) bool {
	return a == b
}

// EqualsIgnoreCase checks if two strings are equal, ignoring case differences.
// Both strings are converted to lowercase before comparison.
//
// Parameters:
//   - a: The first string.
//   - b: The second string.
//
// Returns:
//   - true if a and b are equal when compared case-insensitively; false otherwise.
//
// Example:
//
//	result := EqualsIgnoreCase("Hello", "hello") // result will be true
//	result = EqualsIgnoreCase("Hello", "world")  // result will be false
func EqualsIgnoreCase(a string, b string) bool {
	return strings.EqualFold(a, b)
}

// TailLines returns the last `num` lines from a slice of strings.
//
// If `num` is less than or equal to 0 or the slice is empty, it returns an empty slice.
// If `num` is greater than the number of lines in the slice, it returns all lines.
//
// Parameters:
//   - lines: A slice of strings representing the lines.
//   - num: The number of lines to return from the end of the slice.
//
// Returns:
//   - A slice containing the last `num` lines from the input slice.
//
// Example:
//
//	lines := []string{"line1", "line2", "line3", "line4", "line5"}
//	result := TailLines(lines, 3) // result will be []string{"line3", "line4", "line5"}
func TailLines(lines []string, num int) []string {
	if num <= 0 || len(lines) == 0 {
		return []string{}
	}
	start := len(lines) - num
	if start < 0 {
		start = 0
	}
	return lines[start:]
}

// RemoveDoubleQuotes removes all double quotes from a string.
//
// Parameters:
//   - str: The input string.
//
// Returns:
//   - The string with all double quotes removed.
//
// Example:
//
//	result := RemoveDoubleQuotes("\"hello\"") // result will be "hello"
func RemoveDoubleQuotes(str string) string {
	if IsEmpty(str) {
		return str
	}
	return strings.ReplaceAll(str, `"`, "")
}

// StripNonWhitespace removes all non-whitespace characters from the input string, leaving only whitespace characters.
// The function iterates over each character in the input string and appends only whitespace characters (' ', '\t', '\n', '\r')
// to a new string. All non-whitespace characters are ignored and not included in the result.
//
// Parameters:
//   - s: A string that may contain a mixture of whitespace and non-whitespace characters.
//
// Returns:
//   - A new string consisting only of whitespace characters from the original string. If there are no whitespace characters
//     in the input string, it returns an empty string.
//
// Example Usage:
//
//	str := "  \t\n   abc  "
//	result := StripNonWhitespace(str)
//	// result: "     " (all non-whitespace characters are removed)
//
//	str = "hello"
//	result = StripNonWhitespace(str)
//	// result: "" (no whitespace characters, returns an empty string)
//
// Details:
//
//   - The function iterates through each character in the input string `s` and skips any non-whitespace character.
//
//   - It appends each whitespace character to a new byte slice `s2`, which is later converted to a string and returned.
//
//   - If the input string contains no whitespace characters, the function returns an empty string.
//
//   - This function may not be very efficient for long strings, as it performs an inner loop on each non-whitespace character.
func StripNonWhitespace(s string) string {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case ' ', '\t', '\n', '\r':
			continue
		default:
			var s2 []byte
			for i := 0; i < len(s); i++ {
				switch s[i] {
				case ' ', '\t', '\n', '\r':
					s2 = append(s2, s[i])
				}
			}
			return string(s2)
		}
	}
	return s
}

// UnsafeStr converts a byte slice to a string without copying.
//
// Parameters:
//   - b: The byte slice to convert.
//
// Returns:
//   - A string representation of the byte slice.
//
// Notes:
//   - This function uses unsafe operations to directly reinterpret the byte slice's underlying data
//     structure as a string. This allows efficient conversion without copying the data.
//   - The resulting string must be treated with care. Modifying the original byte slice
//     after conversion will affect the string and can lead to undefined behavior.
//
// Example:
//
//	result := UnsafeStr([]byte("hello")) // result will be "hello"
func UnsafeStr(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// SafeStr converts a byte slice to a string safely.
//
// Parameters:
//   - b: The byte slice to convert.
//
// Returns:
//   - A string representation of the byte slice.
//
// Notes:
//   - This function use safe operations to convert a byte slice to a string.
//   - It is safer than UnsafeStr because it does not use unsafe operations.
//   - It is slower than UnsafeStr because it copies the data.
//
// Example:
//
//	result := SafeStr([]byte("hello")) // result will be "hello"
func SafeStr(b []byte) string {
	if b == nil {
		return ""
	}
	return string(b)
}

// SafeBytes converts a string to a byte slice safely.
//
// Parameters:
//   - s: The string to convert.
//
// Returns:
//   - A byte slice representation of the string.
//
// Notes:
//   - This function use safe operations to convert a string to a byte slice.
//   - It is safer than UnsafeBytes because it does not use unsafe operations.
//   - It is slower than UnsafeBytes because it copies the data.
//
// Example:
//
//	result := SafeBytes("hello") // result will be []byte("hello")
func SafeBytes(s string) []byte {
	if IsEmpty(s) {
		return nil
	}
	return []byte(s)
}

// isDelimiter checks if a character is a delimiter.
//
// This helper function determines whether a given character `c` is a
// delimiter. If no delimiters are provided, it checks if the character
// is whitespace. If delimiters are specified, it checks if the character
// matches any of those delimiters.
//
// Parameters:
//   - `c`: The character to check.
//   - `delimiters`: A variadic parameter that specifies delimiters to check against.
//
// Returns:
//   - `true` if the character is a delimiter; otherwise, returns `false`.
func isDelimiter(c rune, delimiters ...string) bool {
	if delimiters == nil {
		return unicode.IsSpace(c)
	}
	cs := string(c)
	for _, delimiter := range delimiters {
		if cs == delimiter {
			return true
		}
	}
	return false
}

// prependIfMissing is an internal function that performs the logic of prepending a prefix to a string
// if the string does not already start with the specified prefix or any of the additional prefixes,
// with an option for case-insensitive comparison.
//
// Parameters:
//   - `str`: The input string to check and potentially modify.
//   - `prefix`: The main prefix to prepend if `str` does not already start with it.
//   - `ignoreCase`: A boolean flag indicating whether to ignore case when checking for prefixes.
//   - `prefixes`: An optional variadic parameter specifying additional prefixes to check against.
//
// Returns:
//   - The original string if it already starts with the specified prefix (or one of the provided prefixes).
//     Otherwise, it returns the string with the `prefix` prepended.
func prependIfMissing(str string, prefix string, ignoreCase bool, prefixes ...string) string {
	if IsEmpty(prefix) || internalStartsWith(str, prefix, ignoreCase) {
		return str
	}
	for _, pref := range prefixes {
		if pref == "" || internalStartsWith(str, pref, ignoreCase) {
			return str
		}
	}
	return prefix + str
}

// internalStartsWith checks if a given string starts with a specified prefix.
//
// This function determines whether the input string `str` begins with the specified
// `prefix`. It can perform the check in a case-sensitive or case-insensitive manner
// based on the `ignoreCase` parameter.
//
// Parameters:
//   - `str`: The source string to be checked for the specified prefix.
//   - `prefix`: The prefix string to look for at the start of `str`.
//   - `ignoreCase`: A boolean indicating whether the comparison should be case-sensitive
//     (false) or case-insensitive (true).
//
// Returns:
//   - `true` if `str` starts with `prefix` (case-sensitive or case-insensitive depending
//     on the `ignoreCase` flag).
//   - `false` if `str` does not start with `prefix`, or if either `str` or `prefix` is empty
//     (in which case it returns `true` only if both are empty).
//
// Example:
//
//	str := "Hello, World!"
//	prefix := "Hello"
//	result := internalStartsWith(str, prefix, false) // result will be true since str starts with the prefix "Hello".
//
// Notes:
//   - If both `str` and `prefix` are empty, the function returns true.
//   - If `prefix` is longer than `str`, the function will return false.
func internalStartsWith(str string, prefix string, ignoreCase bool) bool {
	if str == "" || prefix == "" {
		return (str == "" && prefix == "")
	}
	if utf8.RuneCountInString(prefix) > utf8.RuneCountInString(str) {
		return false
	}
	if ignoreCase {
		return strings.HasPrefix(strings.ToLower(str), strings.ToLower(prefix))
	}
	return strings.HasPrefix(str, prefix)
}

// internalEndsWith is a helper function that checks if the given string `str` ends with the specified `suffix`.
//
// This function provides both case-sensitive and case-insensitive checks based on the `ignoreCase` flag.
// It first checks if either the string or suffix is empty, returning true if both are empty. Then, it ensures
// that the `suffix` is not longer than `str`. If `ignoreCase` is true, it performs a case-insensitive comparison;
// otherwise, it uses a case-sensitive comparison.
//
// Parameters:
//   - `str`: The input string to check.
//   - `suffix`: The suffix to check for at the end of the string.
//   - `ignoreCase`: A boolean flag indicating whether the comparison should be case-insensitive.
//
// Returns:
//   - `true` if `str` ends with `suffix` based on the case-sensitivity specified by `ignoreCase`, `false` otherwise.
func internalEndsWith(str string, suffix string, ignoreCase bool) bool {
	if str == "" || suffix == "" {
		return (str == "" && suffix == "")
	}
	if utf8.RuneCountInString(suffix) > utf8.RuneCountInString(str) {
		return false
	}
	if ignoreCase {
		return strings.HasSuffix(strings.ToLower(str), strings.ToLower(suffix))
	}
	return strings.HasSuffix(str, suffix)
}

// internalAppendIfMissing appends a specified suffix to a string if it is not already present.
//
// This function checks if the input string `str` ends with a specified `suffix` or any of the additional
// suffixes provided in the variadic `suffixes` parameter. If none of the specified suffixes are present,
// it appends the first suffix to `str`. The comparison can be case-sensitive or case-insensitive based on
// the `ignoreCase` flag.
//
// Parameters:
//   - `str`: The input string to check and potentially modify.
//   - `suffix`: The main suffix to append if `str` does not already end with it.
//   - `ignoreCase`: A boolean flag that determines whether the comparison should be case-insensitive.
//   - `suffixes`: An optional variadic parameter specifying additional suffixes to check against.
//
// Returns:
//   - The original string if it already ends with the specified suffix (or one of the provided suffixes).
//     Otherwise, it returns the string with the `suffix` appended.
//
// Example:
//
//	input := "example"
//	suffix := ".txt"
//	result := internalAppendIfMissing(input, suffix, false) // result will be "example.txt" if input does not already end with ".txt".
//
// Notes:
//   - This function is case-insensitive if `ignoreCase` is set to true.
//   - It checks both the `suffix` and any additional suffixes in the order they are provided.
func internalAppendIfMissing(str string, suffix string, ignoreCase bool, suffixes ...string) string {
	if str == "" || IsEmpty(str) {
		return str
	}
	if ignoreCase {
		if EndsWithIgnoreCase(str, suffix) {
			return str
		}

		for _, suffix := range suffixes {
			if EndsWithIgnoreCase(str, (string)(suffix)) {
				return str
			}
		}
	} else {
		if EndsWith(str, suffix) {
			return str
		}

		for _, suffix := range suffixes {
			if EndsWith(str, (string)(suffix)) {
				return str
			}
		}
	}
	return str + suffix
}

// stringInSlice checks if a given string is present in a slice of strings.
//
// This function iterates through the provided slice `list` and compares each
// element with the specified string `a`. If a match is found, it returns
// `true`. If the iteration completes without finding a match, it returns `false`.
//
// Parameters:
//   - `a`: The string to search for within the slice.
//   - `list`: A slice of strings to search through.
//
// Returns:
//   - `true` if the string `a` is found in the `list`, `false` otherwise.
//
// Example:
//
//	inputStr := "apple"
//	strList := []string{"banana", "orange", "apple", "grape"}
//	result := stringInSlice(inputStr, strList) // result will be true because "apple" is present in the strList.
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// normalizeRune takes a rune as input and returns its normalized string representation.
//
// The function handles a wide variety of special characters and diacritics, replacing them with their
// more common ASCII equivalents. The specific transformations are defined in a switch statement that
// maps each rune to its normalized string value. This includes:
//
//   - Ligatures (e.g., 'Æ' to "AE", 'Œ' to "OE")
//   - Special characters (e.g., 'ß' to "ss")
//   - Accented characters (e.g., 'é' to "e", 'ç' to "c")
//   - Various other characters with diacritics are converted to their closest ASCII equivalent.
//
// If the input rune does not match any of the defined cases, the function returns the rune as a string,
// preserving its original representation.
//
// Parameters:
//   - `r`: A rune representing a Unicode character that may need normalization.
//
// Returns:
//   - A string that represents the normalized form of the input rune. This string is typically ASCII,
//     making it easier to handle in contexts where only basic characters are expected.
//
// Example:
//
//	normalized := normalizeRune('É')
//	fmt.Println(normalized) // Output: "E"
//
// Notes:
//   - This function is particularly useful for text processing, sanitization, or when preparing data
//     for systems that require ASCII-compatible strings.
//
// Replacements are taken from
//   - https://github.com/sindresorhus/slugify/blob/master/replacements.js
//   - https://github.com/cocur/slugify/tree/master/Resources/rules
func normalizeRune(r rune) string {
	switch r {
	case 'Ə':
		return "E"
	case 'Æ':
		return "AE"
	case 'ß':
		return "ss"
	case 'æ':
		return "ae"
	case 'Ø':
		return "OE"
	case 'ø':
		return "oe"
	case 'Å':
		return "a"
	case 'å':
		return "a"
	case 'É':
		return "E"
	case 'é':
		return "e"
	case 'Č':
		return "C"
	case 'Ć':
		return "C"
	case 'Ž':
		return "Z"
	case 'Š':
		return "S"
	case 'Đ':
		return "D"
	case 'č':
		return "c"
	case 'ć':
		return "c"
	case 'ž':
		return "z"
	case 'š':
		return "s"
	case 'đ':
		return "d"
	case 'Ď':
		return "D"
	case 'Ě':
		return "E"
	case 'Ň':
		return "N"
	case 'Ř':
		return "R"
	case 'Ť':
		return "T"
	case 'Ů':
		return "U"
	case 'ď':
		return "d"
	case 'ě':
		return "e"
	case 'ň':
		return "n"
	case 'ř':
		return "r"
	case 'ť':
		return "t"
	case 'ů':
		return "u"
	case 'ǽ':
		return "ae"
	case 'À':
		return "A"
	case 'Á':
		return "A"
	case 'Â':
		return "A"
	case 'Ã':
		return "A"
	case 'Ǻ':
		return "A"
	case 'Ă':
		return "A"
	case 'Ǎ':
		return "A"
	case 'Ǽ':
		return "AE"
	case 'à':
		return "a"
	case 'á':
		return "a"
	case 'â':
		return "a"
	case 'ã':
		return "a"
	case 'ǻ':
		return "a"
	case 'ă':
		return "a"
	case 'ǎ':
		return "a"
	case 'ª':
		return "a"
	case '@':
		return "at"
	case 'Ĉ':
		return "C"
	case 'Ċ':
		return "C"
	case 'Ç':
		return "C"
	case 'ç':
		return "c"
	case 'ĉ':
		return "c"
	case 'ċ':
		return "c"
	case '©':
		return "c"
	case 'Ð':
		return "Dj"
	case 'ð':
		return "dj"
	case 'È':
		return "E"
	case 'Ê':
		return "E"
	case 'Ë':
		return "E"
	case 'Ĕ':
		return "E"
	case 'Ė':
		return "E"
	case 'è':
		return "e"
	case 'ê':
		return "e"
	case 'ë':
		return "e"
	case 'ĕ':
		return "e"
	case 'ė':
		return "e"
	case 'ƒ':
		return "f"
	case 'Ĝ':
		return "G"
	case 'Ġ':
		return "G"
	case 'ĝ':
		return "g"
	case 'ġ':
		return "g"
	case 'Ĥ':
		return "H"
	case 'Ħ':
		return "H"
	case 'ĥ':
		return "h"
	case 'ħ':
		return "h"
	case 'Ì':
		return "I"
	case 'Í':
		return "I"
	case 'Î':
		return "I"
	case 'Ï':
		return "I"
	case 'Ĩ':
		return "I"
	case 'Ĭ':
		return "I"
	case 'Ǐ':
		return "I"
	case 'Į':
		return "I"
	case 'Ĳ':
		return "IJ"
	case 'ì':
		return "i"
	case 'í':
		return "i"
	case 'î':
		return "i"
	case 'ï':
		return "i"
	case 'ĩ':
		return "i"
	case 'ĭ':
		return "i"
	case 'ǐ':
		return "i"
	case 'į':
		return "i"
	case 'ĳ':
		return "ij"
	case 'Ĵ':
		return "J"
	case 'ĵ':
		return "j"
	case 'Ĺ':
		return "L"
	case 'Ľ':
		return "L"
	case 'Ŀ':
		return "L"
	case 'ĺ':
		return "l"
	case 'ľ':
		return "l"
	case 'ŀ':
		return "l"
	case 'Ñ':
		return "N"
	case 'ñ':
		return "n"
	case 'ŉ':
		return "n"
	case 'Ò':
		return "O"
	case 'Ó':
		return "O"
	case 'Ô':
		return "O"
	case 'Õ':
		return "O"
	case 'Ō':
		return "O"
	case 'Ŏ':
		return "O"
	case 'Ǒ':
		return "O"
	case 'Ő':
		return "O"
	case 'Ơ':
		return "O"
	case 'Ǿ':
		return "O"
	case 'Œ':
		return "OE"
	case 'ò':
		return "o"
	case 'ó':
		return "o"
	case 'ô':
		return "o"
	case 'õ':
		return "o"
	case 'ō':
		return "o"
	case 'ŏ':
		return "o"
	case 'ǒ':
		return "o"
	case 'ő':
		return "o"
	case 'ơ':
		return "o"
	case 'ǿ':
		return "o"
	case 'º':
		return "o"
	case 'œ':
		return "oe"
	case 'Ŕ':
		return "R"
	case 'Ŗ':
		return "R"
	case 'ŕ':
		return "r"
	case 'ŗ':
		return "r"
	case 'Ŝ':
		return "S"
	case 'Ș':
		return "S"
	case 'ŝ':
		return "s"
	case 'ș':
		return "s"
	case 'ſ':
		return "s"
	case 'Ţ':
		return "T"
	case 'Ț':
		return "T"
	case 'Ŧ':
		return "T"
	case 'Þ':
		return "TH"
	case 'ţ':
		return "t"
	case 'ț':
		return "t"
	case 'ŧ':
		return "t"
	case 'þ':
		return "th"
	case 'Ù':
		return "U"
	case 'Ú':
		return "U"
	case 'Û':
		return "U"
	case 'Ü':
		return "U"
	case 'Ũ':
		return "U"
	case 'Ŭ':
		return "U"
	case 'Ű':
		return "U"
	case 'Ų':
		return "U"
	case 'Ư':
		return "U"
	case 'Ǔ':
		return "U"
	case 'Ǖ':
		return "U"
	case 'Ǘ':
		return "U"
	case 'Ǚ':
		return "U"
	case 'Ǜ':
		return "U"
	case 'ù':
		return "u"
	case 'ú':
		return "u"
	case 'û':
		return "u"
	case 'ü':
		return "u"
	case 'ũ':
		return "u"
	case 'ŭ':
		return "u"
	case 'ű':
		return "u"
	case 'ų':
		return "u"
	case 'ư':
		return "u"
	case 'ǔ':
		return "u"
	case 'ǖ':
		return "u"
	case 'ǘ':
		return "u"
	case 'ǚ':
		return "u"
	case 'ǜ':
		return "u"
	case 'Ŵ':
		return "W"
	case 'ŵ':
		return "w"
	case 'Ý':
		return "Y"
	case 'Ÿ':
		return "Y"
	case 'Ŷ':
		return "Y"
	case 'ý':
		return "y"
	case 'ÿ':
		return "y"
	case 'ŷ':
		return "y"
	case 'Ä':
		return "A"
	case 'Ö':
		return "O"
	case 'ä':
		return "a"
	case 'ö':
		return "o"
	case 'Ā':
		return "A"
	case 'Ē':
		return "E"
	case 'Ģ':
		return "G"
	case 'Ī':
		return "I"
	case 'Ķ':
		return "K"
	case 'Ļ':
		return "L"
	case 'Ņ':
		return "N"
	case 'Ū':
		return "U"
	case 'ā':
		return "a"
	case 'ē':
		return "e"
	case 'ģ':
		return "g"
	case 'ī':
		return "i"
	case 'ķ':
		return "k"
	case 'ļ':
		return "l"
	case 'ņ':
		return "n"
	case 'ū':
		return "u"
	case 'Ą':
		return "A"
	case 'Ę':
		return "E"
	case 'ą':
		return "a"
	case 'ę':
		return "e"
	case 'Ł':
		return "L"
	case 'Ń':
		return "N"
	case 'Ś':
		return "S"
	case 'Ź':
		return "Z"
	case 'Ż':
		return "Z"
	case 'ł':
		return "l"
	case 'ń':
		return "n"
	case 'ś':
		return "s"
	case 'ź':
		return "z"
	case 'ż':
		return "z"
	case 'ş':
		return "s"
	case 'Ş':
		return "S"
	case 'Ğ':
		return "G"
	case 'İ':
		return "I"
	case 'ğ':
		return "g"
	case 'ı':
		return "i"
	case 'ạ':
		return "a"
	case 'Ạ':
		return "A"
	case 'ả':
		return "a"
	case 'Ả':
		return "A"
	case 'ấ':
		return "a"
	case 'Ấ':
		return "A"
	case 'ầ':
		return "a"
	case 'Ầ':
		return "A"
	case 'ẩ':
		return "a"
	case 'Ẩ':
		return "A"
	case 'ẫ':
		return "a"
	case 'Ẫ':
		return "A"
	case 'ậ':
		return "a"
	case 'Ậ':
		return "A"
	case 'ắ':
		return "a"
	case 'Ắ':
		return "A"
	case 'ằ':
		return "a"
	case 'Ằ':
		return "A"
	case 'ẳ':
		return "a"
	case 'Ẳ':
		return "A"
	case 'ẵ':
		return "a"
	case 'Ẵ':
		return "A"
	case 'ặ':
		return "a"
	case 'Ặ':
		return "A"
	case 'ẹ':
		return "e"
	case 'Ẹ':
		return "E"
	case 'ẻ':
		return "e"
	case 'Ẻ':
		return "E"
	case 'ẽ':
		return "e"
	case 'Ẽ':
		return "E"
	case 'ế':
		return "e"
	case 'Ế':
		return "E"
	case 'ề':
		return "e"
	case 'Ề':
		return "E"
	case 'ể':
		return "e"
	case 'Ể':
		return "E"
	case 'ễ':
		return "e"
	case 'Ễ':
		return "E"
	case 'ệ':
		return "e"
	case 'Ệ':
		return "E"
	case 'ỉ':
		return "i"
	case 'Ỉ':
		return "I"
	case 'ị':
		return "i"
	case 'Ị':
		return "I"
	case 'ọ':
		return "o"
	case 'Ọ':
		return "O"
	case 'ỏ':
		return "o"
	case 'Ỏ':
		return "O"
	case 'ố':
		return "o"
	case 'Ố':
		return "O"
	case 'ồ':
		return "o"
	case 'Ồ':
		return "O"
	case 'ổ':
		return "o"
	case 'Ổ':
		return "O"
	case 'ỗ':
		return "o"
	case 'Ỗ':
		return "O"
	case 'ộ':
		return "o"
	case 'Ộ':
		return "O"
	case 'ớ':
		return "o"
	case 'Ớ':
		return "O"
	case 'ờ':
		return "o"
	case 'Ờ':
		return "O"
	case 'ở':
		return "o"
	case 'Ở':
		return "O"
	case 'ỡ':
		return "o"
	case 'Ỡ':
		return "O"
	case 'ợ':
		return "o"
	case 'Ợ':
		return "O"
	case 'ụ':
		return "u"
	case 'Ụ':
		return "U"
	case 'ủ':
		return "u"
	case 'Ủ':
		return "U"
	case 'ứ':
		return "u"
	case 'Ứ':
		return "U"
	case 'ừ':
		return "u"
	case 'Ừ':
		return "U"
	case 'ử':
		return "u"
	case 'Ử':
		return "U"
	case 'ữ':
		return "u"
	case 'Ữ':
		return "U"
	case 'ự':
		return "u"
	case 'Ự':
		return "U"
	case 'ỳ':
		return "y"
	case 'Ỳ':
		return "Y"
	case 'ỵ':
		return "y"
	case 'Ỵ':
		return "Y"
	case 'ỷ':
		return "y"
	case 'Ỷ':
		return "Y"
	case 'ỹ':
		return "y"
	case 'Ỹ':
		return "Y"
	}
	return string(r)
}

// isSeparator checks if a rune is considered a separator.
// This function defines separators as any character that is not an ASCII alphanumeric
// or underscore. For non-ASCII characters, it treats letters and digits as non-separators,
// and considers spaces as separators.
func isSeparator(r rune) bool {
	if r <= 0x7F {
		switch {
		case '0' <= r && r <= '9':
			return false
		case 'a' <= r && r <= 'z':
			return false
		case 'A' <= r && r <= 'Z':
			return false
		case r == '_':
			return false
		}
		return true
	}
	// Letters and digits are not separators
	if unicode.IsLetter(r) || unicode.IsDigit(r) {
		return false
	}
	// Otherwise, all we can do for now is treat spaces as separators.
	return unicode.IsSpace(r)
}

// padRepeat is a helper that builds a padding string of exactly n runes
// by repeating the pad string and truncating if necessary.
func padRepeat(pad string, n int) string {
	if n <= 0 {
		return ""
	}
	padLen := utf8.RuneCountInString(pad)
	var sb strings.Builder
	sb.Grow(n * len(pad) / padLen)
	remaining := n
	for remaining > 0 {
		if padLen <= remaining {
			sb.WriteString(pad)
			remaining -= padLen
		} else {
			sb.WriteString(string([]rune(pad)[:remaining]))
			remaining = 0
		}
	}
	return sb.String()
}

// Oneline returns str with every line-breaking sequence replaced by a single
// space, guaranteeing the result can be rendered on a single line.
//
// The following sequences are treated as line breaks:
//
//   - \r\n  (CRLF)
//   - \r    (CR)
//   - \n    (LF)
//   - \t    (horizontal tab)
//   - \v    (vertical tab)
//   - \f    (form feed)
//   - U+0085 (NEL - Next Line)
//   - U+2028 (LS  - Line Separator)
//   - U+2029 (PS  - Paragraph Separator)
//
// \r\n is matched as a single unit so it produces exactly one space rather
// than two. Word boundaries at line breaks are preserved by the inserted space.
// An empty string is returned unchanged.
//
// Parameters:
//   - str: The input string to sanitize.
//
// Returns:
//   - A new string guaranteed to contain no line-breaking characters.
//
// Example:
//
//	Oneline("hello\nworld")      // "hello world"
//	Oneline("foo\r\nbar")        // "foo bar"
//	Oneline("a\tb")              // "a b"
//	Oneline("")                  // ""
//	Oneline("no breaks here")    // "no breaks here"
func Oneline(str string) string {
	if str == "" {
		return ""
	}
	return lineBreakReplacer.Replace(str)
}
