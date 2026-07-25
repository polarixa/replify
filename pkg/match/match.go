package match

import (
	"unicode/utf8"

	"github.com/polarixa/replify/pkg/strutil"
)

// result represents the outcome of a match operation.
type result int

var (
	// MaxRuneBytes represents the maximum valid UTF-8 encoding of a Unicode code point.
	// It is a byte slice containing the specific byte values [244, 143, 191, 191].
	MaxRuneBytes = [...]byte{244, 143, 191, 191}
)

// Result constants for match outcomes.
const (
	rightNoMatch result = iota
	rightMatch
	rightStop
)

// Match checks if a given string matches a pattern with simple wildcard support.
//
// This function provides a basic matching mechanism where:
//   - '*' matches any sequence of characters, including an empty sequence.
//   - '?' matches any single character.
//
// The function returns `true` if `str` matches the specified `pattern`. If `pattern`
// is just '*', it will automatically match any input string.
//
// Match returns true if str matches pattern. This is a very
// simple wildcard match where '*' matches on any number characters
// and '?' matches on any one character.
//
// pattern:
//
//	{ term }
//
// term:
//
//	'*'         matches any sequence of non-Separator characters
//	'?'         matches any single non-Separator character
//	c           matches character c (c != '*', '?', '\\')
//	'\\' c      matches character c
//
// Parameters:
//   - `str`: The input string to be matched.
//   - `pattern`: The pattern string, which may contain wildcard characters ('*' and '?').
//
// Returns:
//   - A boolean value: `true` if `str` matches `pattern`; otherwise `false`.
//
// Example:
//
//	matched := Match("hello", "h*o") // returns `true`
func Match(str, pattern string) bool {
	if pattern == "*" {
		return true
	}
	// Perform the match using the `match` function with no complexity limit
	return match(str, pattern, 0, nil, -1) == rightMatch
}

// MatchLimit checks if a given string matches a pattern with wildcard support
// within a specified complexity limit.
//
// This function allows matching of `str` against `pattern` using wildcard characters:
//   - '*' matches any sequence of characters (including an empty sequence).
//   - '?' matches any single character.
//
// To prevent excessive computational cost, the function enforces a complexity limit
// specified by `maxComplexity`. If the complexity exceeds `maxComplexity`, the function
// stops matching and indicates that the limit was reached.
//
// Parameters:
//   - `str`: The string to match against the pattern.
//   - `pattern`: The pattern string, potentially containing wildcard characters ('*' and '?').
//   - `maxComplexity`: An integer representing the maximum allowed complexity for the match operation.
//
// Returns:
//   - `matched` (bool): `true` if `str` matches `pattern` within the complexity limit; otherwise `false`.
//   - `stopped` (bool): `true` if the complexity limit was exceeded, causing the match to stop.
//
// Example:
//
//	matched, stopped := MatchLimit("hello", "h*o", 10) // could return `true, false`
func MatchLimit(str, pattern string, maxComplexity int) (matched, stopped bool) {
	if pattern == "*" {
		return true, false
	}
	counter := 0
	r := match(str, pattern, len(str), &counter, maxComplexity)
	if r == rightStop {
		return false, true
	}
	return r == rightMatch, false
}

// WildcardPatternLimits calculates the minimum and maximum possible string values that match a given pattern
// with optional wildcard characters `*` and `?`.
//
// This function interprets a `pattern` string, where:
//   - `*` acts as a wildcard that can match any sequence of characters.
//   - `?` represents any single character.
//
// It returns two strings representing the smallest (`min`) and largest (`max`) possible
// values that match the specified pattern.
// If the pattern begins with `*` or is empty, both `min` and `max` are returned as empty strings.
//
// Parameters:
//   - `pattern`: A string pattern that may contain wildcards `*` and `?`.
//
// Returns:
//   - `min`: The minimum possible string value matching the pattern.
//   - `max`: The maximum possible string value matching the pattern.
//
// Example:
//
//	min, max := WildcardPatternLimits("a?c*") // min could be "a\x00c", and max could be "azzz…"
func WildcardPatternLimits(pattern string) (min, max string) {
	if strutil.IsEmpty(pattern) || pattern[0] == '*' {
		return "", ""
	}
	minVal := make([]byte, 0, len(pattern))
	maxVal := make([]byte, 0, len(pattern))
	var wild bool
	for i := 0; i < len(pattern); i++ {
		if pattern[i] == '*' {
			wild = true
			break
		}
		if pattern[i] == '?' {
			minVal = append(minVal, 0)
			maxVal = append(maxVal, MaxRuneBytes[:]...)
		} else {
			minVal = append(minVal, pattern[i])
			maxVal = append(maxVal, pattern[i])
		}
	}
	if wild {
		r, n := utf8.DecodeLastRune(maxVal)
		if r != utf8.RuneError {
			if r < utf8.MaxRune {
				r++
				if r > 0x7f {
					b := make([]byte, 4)
					nn := utf8.EncodeRune(b, r)
					maxVal = append(maxVal[:len(maxVal)-n], b[:nn]...)
				} else {
					maxVal = append(maxVal[:len(maxVal)-n], byte(r))
				}
			}
		}
	}
	return string(minVal), string(maxVal)
}

// match performs a pattern matching operation on a given `text` string using a `pattern`
// that can include special wildcard characters '*' and '?'.
//
// This function evaluates if `text` matches `pattern`, where:
//   - '?' represents any single character.
//   - '*' matches any sequence of characters (including an empty sequence).
//
// It also allows for a limit on the number of pattern comparisons based on `maxComplexity` to
// prevent excessive matching complexity. If the complexity exceeds `maxComplexity * lenString`,
// the function returns early.
//
// Parameters:
//   - `text`: The string to be matched against the `pattern`.
//   - `pattern`: The pattern string, potentially containing wildcard characters `*` and `?`.
//   - `lenString`: The length of the original `text`, used to calculate the complexity limit.
//   - `counter`: A pointer to an integer counter that tracks the number of recursive calls.
//   - `maxComplexity`: An integer specifying the maximum allowed complexity based on `lenString`.
//
// Returns:
//   - A `result` value representing the match outcome:
//   - `rightMatch` if the `text` matches the `pattern` according to the rules.
//   - `rightNoMatch` if the `text` does not match the `pattern`.
//   - `rightStop` if the complexity exceeds the limit set by `maxComplexity`.
//
// Example:
//
//	result := match("hello", "h?llo*", len("hello"), &counter, 10) // could return `rightMatch`
func match(text, pattern string, lenString int, counter *int, maxComplexity int) result {
	if maxComplexity > -1 {
		if *counter > lenString*maxComplexity {
			return rightStop
		}
		*counter++
	}

	for len(pattern) > 0 {
		var wild bool
		pc, ps := rune(pattern[0]), 1
		if pc > 0x7f {
			pc, ps = utf8.DecodeRuneInString(pattern)
		}
		var sc rune
		var ss int
		if len(text) > 0 {
			sc, ss = rune(text[0]), 1
			if sc > 0x7f {
				sc, ss = utf8.DecodeRuneInString(text)
			}
		}
		switch pc {
		case '?':
			if ss == 0 {
				return rightNoMatch
			}
		case '*':
			// Ignore repeating stars.
			for len(pattern) > 1 && pattern[1] == '*' {
				pattern = pattern[1:]
			}
			// If this star is the last character then it must be a match.
			if len(pattern) == 1 {
				return rightMatch
			}
			// Match and trim any non-wildcard suffix characters.
			var ok bool
			text, pattern, ok = wildcardSuffixMatch(text, pattern)
			if !ok {
				return rightNoMatch
			}
			// Check for single star again.
			if len(pattern) == 1 {
				return rightMatch
			}
			// Perform recursive wildcard search.
			r := match(text, pattern[1:], lenString, counter, maxComplexity)
			if r != rightNoMatch {
				return r
			}
			if len(text) == 0 {
				return rightNoMatch
			}
			wild = true
		default:
			if ss == 0 {
				return rightNoMatch
			}
			if pc == '\\' {
				pattern = pattern[ps:]
				pc, ps = utf8.DecodeRuneInString(pattern)
				if ps == 0 {
					return rightNoMatch
				}
			}
			if sc != pc {
				return rightNoMatch
			}
		}
		text = text[ss:]
		if !wild {
			pattern = pattern[ps:]
		}
	}
	if len(text) == 0 {
		return rightMatch
	}
	return rightNoMatch
}

// wildcardSuffixMatch checks if a given text string matches a pattern with potential wildcard support,
// and returns the remaining unmatched portions of text and pattern, along with a boolean indicating
// if a match was successful.
//
// This function iterates over the `text` and `pattern` from the end, checking if characters match
// with special handling for wildcards:
//   - A '*' character in the `pattern` (not preceded by an escape character `\`) acts as a wildcard
//     that matches any sequence of characters in `text`.
//   - A '?' character in the `pattern` matches any single character in `text`, unless escaped with `\`.
//
// Parameters:
//   - `text`: The string to match against the pattern.
//   - `pattern`: The pattern string, potentially containing wildcard characters (`*` and `?`).
//
// Returns:
//   - A modified `text` and `pattern` string, with any matched characters removed from the end.
//   - A boolean value:
//   - true if the `text` and `pattern` match according to the wildcard rules.
//   - false if there was no match.
//
// Example:
//
//	trimmedText, trimmedPattern, isMatch := wildcardSuffixMatch("hello world", "*world") // returns "hello ", "", true
func wildcardSuffixMatch(text, pattern string) (string, string, bool) {
	// It's expected that the pattern has at least two bytes and the first byte
	// is a wildcard star '*'
	match := true
	for len(text) > 0 && len(pattern) > 1 {
		pc, ps := utf8.DecodeLastRuneInString(pattern)
		var esc bool
		for i := 0; ; i++ {
			if pattern[len(pattern)-ps-i-1] != '\\' {
				if i&1 == 1 {
					esc = true
					ps++
				}
				break
			}
		}
		if pc == '*' && !esc {
			match = true
			break
		}
		sc, ss := utf8.DecodeLastRuneInString(text)
		if !((pc == '?' && !esc) || pc == sc) {
			match = false
			break
		}
		text = text[:len(text)-ss]
		pattern = pattern[:len(pattern)-ps]
	}
	return text, pattern, match
}
