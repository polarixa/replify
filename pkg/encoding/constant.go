package encoding

import "regexp"

// Styles
var (
	// TerminalStyle is for terminals
	TerminalStyle *Style

	// VSCodeDarkStyle is for VS Code dark theme
	VSCodeDarkStyle *Style

	// DraculaStyle is for Dracula theme
	DraculaStyle *Style

	// MonokaiStyle is for Monokai theme
	MonokaiStyle *Style

	// SolarizedDarkStyle is for Solarized Dark theme
	SolarizedDarkStyle *Style

	// MinimalGrayStyle is a minimal gray style
	MinimalGrayStyle *Style
)

// DefaultOptionsConfig is a pre-configured default set of options for pretty-printing JSON.
// This configuration uses a width of 80, an empty prefix, two-space indentation, and does not sort keys.
// It is used when no custom options are provided in the PrettyOptions function.
var DefaultOptionsConfig = &OptionsConfig{Width: 80, Prefix: "", Indent: "  ", SortKeys: false}

// normalizeTrailingCommaRe matches a comma followed by optional whitespace then } or ],
// which is the pattern produced by trailing-comma JSON artifacts.
var normalizeTrailingCommaRe = regexp.MustCompile(`,\s*([}\]])`)

// Toggle to choose how to handle NaN/±Inf floats in *safe* variants.
// When true: produce "null" (JSON-safe). When false: treat as error.
const floatsUseNullForNonFinite = true

// Constants for sorting by key or by value
// Used in the byKeyVal struct's isLess method.
const (
	sKey sortCriteria = 0
	sVal sortCriteria = 1
)

// Constants representing different JSON types.
// These constants are used to identify the type of a JSON value
// based on its first character.
const (
	jsonNull   jsonType = iota // Represents a JSON null value
	jsonFalse                  // Represents a JSON false boolean
	jNumber                    // Represents a JSON number
	jsonString                 // Represents a JSON string
	jsonTrue                   // Represents a JSON true boolean
	jsonJson                   // Represents a JSON object or array
)

// binarySniffLen is the number of leading bytes inspected when performing
// content-based binary detection. This mirrors the sample size used by
// net/http.DetectContentType and common gitattributes-style heuristics.
const binarySniffLen = 512
