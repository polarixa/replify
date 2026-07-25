package fj

import (
	"github.com/polarixa/replify/pkg/encoding"
)

const (
	// Null is a constant representing a JSON null value.
	// In JSON, null is used to represent the absence of a value.
	Null Type = iota
	// False is a constant representing a JSON false boolean value.
	// In JSON, false is a boolean value that represents a negative or off state.
	False
	// Number is a constant representing a JSON number value.
	// In JSON, numbers can be integers or floating-point values.
	Number
	// String is a constant representing a JSON string value.
	// In JSON, strings are sequences of characters enclosed in double quotes.
	String
	// True is a constant representing a JSON true boolean value.
	// In JSON, true is a boolean value that represents a positive or on state.
	True
	// JSON is a constant representing a raw JSON block.
	// This type can be used to represent any valid JSON object or array.
	JSON
)

var (
	// DisableTransformers is a global flag that determines whether transformers should be applied
	// when processing JSON values. If set to true, transformers will not be applied to the JSON values.
	// If set to false, transformers will be applied as expected.
	DisableTransformers = false

	// hexDigits is an array of bytes representing the hexadecimal digits used in JSON encoding.
	// It contains the characters '0' to '9' and 'a' to 'f', which are used for encoding hexadecimal numbers.
	// This is commonly used for encoding special characters or byte sequences in JSON strings (e.g., for Unicode escape sequences).
	hexDigits = [...]byte{
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
		'a', 'b', 'c', 'd', 'e', 'f',
	}
)

var (
	// defaultStyle defines the default styling rules for different JSON elements.
	// Each style consists of a pair of ANSI escape codes: a start and end sequence.
	// These styles are applied to highlight keys, strings, numbers, booleans, nulls,
	// escape sequences, and brackets in colored terminal output.
	//
	// Fields:
	//   - Key: Styling for JSON keys.
	//   - String: Styling for JSON string values.
	//   - Number: Styling for JSON numbers.
	//   - True: Styling for the boolean value `true`.
	//   - False: Styling for the boolean value `false`.
	//   - Null: Styling for the JSON value `null`.
	//   - Escape: Styling for escape sequences in strings.
	//   - Brackets: Styling for JSON brackets (e.g., `{}`, `[]`).
	//   - Append: A custom append function for byte slices.
	defaultStyle = &encoding.Style{
		Key:      [2]string{"\033[1;34m", "\033[0m"},
		String:   [2]string{"\033[1;32m", "\033[0m"},
		Number:   [2]string{"\033[1;33m", "\033[0m"},
		True:     [2]string{"\033[1;35m", "\033[0m"},
		False:    [2]string{"\033[1;35m", "\033[0m"},
		Null:     [2]string{"\033[1;35m", "\033[0m"},
		Escape:   [2]string{"\033[1;31m", "\033[0m"},
		Brackets: [2]string{"\033[1;37m", "\033[0m"},
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// DarkStyle uses darker tones for styling.
	DarkStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;25m", "\033[0m"},  // Dark blue
		String:   [2]string{"\033[38;5;34m", "\033[0m"},  // Dark green
		Number:   [2]string{"\033[38;5;178m", "\033[0m"}, // Dark yellow
		True:     [2]string{"\033[38;5;127m", "\033[0m"}, // Dark magenta
		False:    [2]string{"\033[38;5;127m", "\033[0m"}, // Dark magenta
		Null:     [2]string{"\033[38;5;127m", "\033[0m"}, // Dark magenta
		Escape:   [2]string{"\033[38;5;124m", "\033[0m"}, // Dark red
		Brackets: [2]string{"\033[38;5;245m", "\033[0m"}, // Gray
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// NeonStyle is a vibrant style using neon-like colors.
	NeonStyle = &encoding.Style{
		Key:      [2]string{"\033[1;96m", "\033[0m"}, // Bright cyan
		String:   [2]string{"\033[1;92m", "\033[0m"}, // Bright green
		Number:   [2]string{"\033[1;93m", "\033[0m"}, // Bright yellow
		True:     [2]string{"\033[1;95m", "\033[0m"}, // Bright magenta
		False:    [2]string{"\033[1;95m", "\033[0m"}, // Bright magenta
		Null:     [2]string{"\033[1;95m", "\033[0m"}, // Bright magenta
		Escape:   [2]string{"\033[1;91m", "\033[0m"}, // Bright red
		Brackets: [2]string{"\033[1;97m", "\033[0m"}, // Bright white
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// PastelStyle applies softer colors for a subdued look.
	PastelStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;152m", "\033[0m"}, // Soft teal
		String:   [2]string{"\033[38;5;121m", "\033[0m"}, // Soft green
		Number:   [2]string{"\033[38;5;180m", "\033[0m"}, // Soft yellow
		True:     [2]string{"\033[38;5;139m", "\033[0m"}, // Soft magenta
		False:    [2]string{"\033[38;5;139m", "\033[0m"}, // Soft magenta
		Null:     [2]string{"\033[38;5;139m", "\033[0m"}, // Soft magenta
		Escape:   [2]string{"\033[38;5;167m", "\033[0m"}, // Soft red
		Brackets: [2]string{"\033[38;5;253m", "\033[0m"}, // Light gray
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// HighContrastStyle uses bold and contrasting colors for better visibility.
	HighContrastStyle = &encoding.Style{
		Key:      [2]string{"\033[1;37;44m", "\033[0m"}, // White on blue background
		String:   [2]string{"\033[1;37;42m", "\033[0m"}, // White on green background
		Number:   [2]string{"\033[1;37;43m", "\033[0m"}, // White on yellow background
		True:     [2]string{"\033[1;37;45m", "\033[0m"}, // White on magenta background
		False:    [2]string{"\033[1;37;45m", "\033[0m"}, // White on magenta background
		Null:     [2]string{"\033[1;37;45m", "\033[0m"}, // White on magenta background
		Escape:   [2]string{"\033[1;37;41m", "\033[0m"}, // White on red background
		Brackets: [2]string{"\033[1;30;47m", "\033[0m"}, // Black on white background
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// VintageStyle uses muted tones reminiscent of old terminal displays.
	VintageStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;94m", "\033[0m"},  // Faded blue
		String:   [2]string{"\033[38;5;130m", "\033[0m"}, // Faded orange
		Number:   [2]string{"\033[38;5;136m", "\033[0m"}, // Faded yellow
		True:     [2]string{"\033[38;5;95m", "\033[0m"},  // Faded magenta
		False:    [2]string{"\033[38;5;95m", "\033[0m"},  // Faded magenta
		Null:     [2]string{"\033[38;5;95m", "\033[0m"},  // Faded magenta
		Escape:   [2]string{"\033[38;5;124m", "\033[0m"}, // Dark red
		Brackets: [2]string{"\033[38;5;242m", "\033[0m"}, // Gray
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// CyberpunkStyle mimics a futuristic neon cyberpunk aesthetic.
	CyberpunkStyle = &encoding.Style{
		Key:      [2]string{"\033[1;35;45m", "\033[0m"}, // Magenta with magenta background
		String:   [2]string{"\033[1;36;46m", "\033[0m"}, // Cyan with cyan background
		Number:   [2]string{"\033[1;33;43m", "\033[0m"}, // Yellow with yellow background
		True:     [2]string{"\033[1;32;42m", "\033[0m"}, // Green with green background
		False:    [2]string{"\033[1;31;41m", "\033[0m"}, // Red with red background
		Null:     [2]string{"\033[1;37;40m", "\033[0m"}, // White with black background
		Escape:   [2]string{"\033[1;31;41m", "\033[0m"}, // Red with red background
		Brackets: [2]string{"\033[1;30;47m", "\033[0m"}, // Black with white background
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// OceanStyle is inspired by oceanic hues and soft contrasts.
	OceanStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;27m", "\033[0m"},  // Deep blue
		String:   [2]string{"\033[38;5;45m", "\033[0m"},  // Aqua
		Number:   [2]string{"\033[38;5;33m", "\033[0m"},  // Sea green
		True:     [2]string{"\033[38;5;77m", "\033[0m"},  // Light turquoise
		False:    [2]string{"\033[38;5;77m", "\033[0m"},  // Light turquoise
		Null:     [2]string{"\033[38;5;67m", "\033[0m"},  // Dull aqua
		Escape:   [2]string{"\033[38;5;196m", "\033[0m"}, // Coral red
		Brackets: [2]string{"\033[38;5;15m", "\033[0m"},  // White
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// FieryStyle uses intense warm colors like flames.
	FieryStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;166m", "\033[0m"}, // Burnt orange
		String:   [2]string{"\033[38;5;202m", "\033[0m"}, // Orange
		Number:   [2]string{"\033[38;5;220m", "\033[0m"}, // Yellow-orange
		True:     [2]string{"\033[38;5;214m", "\033[0m"}, // Bright orange
		False:    [2]string{"\033[38;5;160m", "\033[0m"}, // Red-orange
		Null:     [2]string{"\033[38;5;196m", "\033[0m"}, // Red
		Escape:   [2]string{"\033[38;5;124m", "\033[0m"}, // Dark red
		Brackets: [2]string{"\033[38;5;244m", "\033[0m"}, // Light gray
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// GalaxyStyle uses space-themed colors with a starry effect.
	GalaxyStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;57m", "\033[0m"},  // Dark blue
		String:   [2]string{"\033[38;5;93m", "\033[0m"},  // Soft purple
		Number:   [2]string{"\033[38;5;141m", "\033[0m"}, // Violet
		True:     [2]string{"\033[38;5;219m", "\033[0m"}, // Light pink
		False:    [2]string{"\033[38;5;219m", "\033[0m"}, // Light pink
		Null:     [2]string{"\033[38;5;250m", "\033[0m"}, // Light gray
		Escape:   [2]string{"\033[38;5;129m", "\033[0m"}, // Purple
		Brackets: [2]string{"\033[38;5;244m", "\033[0m"}, // Gray
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// SunsetStyle mimics the colors of a sunset, using warm hues and deep purples.
	SunsetStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;214m", "\033[0m"}, // Warm orange
		String:   [2]string{"\033[38;5;213m", "\033[0m"}, // Soft pink
		Number:   [2]string{"\033[38;5;178m", "\033[0m"}, // Soft yellow
		True:     [2]string{"\033[38;5;229m", "\033[0m"}, // Light peach
		False:    [2]string{"\033[38;5;160m", "\033[0m"}, // Deep red
		Null:     [2]string{"\033[38;5;236m", "\033[0m"}, // Dark gray
		Escape:   [2]string{"\033[38;5;202m", "\033[0m"}, // Orange-red
		Brackets: [2]string{"\033[38;5;15m", "\033[0m"},  // White
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// JungleStyle draws inspiration from a dense jungle with deep greens and browns.
	JungleStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;22m", "\033[0m"},  // Deep green
		String:   [2]string{"\033[38;5;28m", "\033[0m"},  // Lush green
		Number:   [2]string{"\033[38;5;130m", "\033[0m"}, // Earthy brown
		True:     [2]string{"\033[38;5;46m", "\033[0m"},  // Forest green
		False:    [2]string{"\033[38;5;166m", "\033[0m"}, // Brownish red
		Null:     [2]string{"\033[38;5;143m", "\033[0m"}, // Muted yellow
		Escape:   [2]string{"\033[38;5;94m", "\033[0m"},  // Brownish orange
		Brackets: [2]string{"\033[38;5;23m", "\033[0m"},  // Dark green
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// MonochromeStyle uses different shades of black and white for a simple, high-contrast theme.
	MonochromeStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;235m", "\033[0m"}, // Dark gray
		String:   [2]string{"\033[38;5;255m", "\033[0m"}, // White
		Number:   [2]string{"\033[38;5;240m", "\033[0m"}, // Light gray
		True:     [2]string{"\033[38;5;255m", "\033[0m"}, // White
		False:    [2]string{"\033[38;5;232m", "\033[0m"}, // Black
		Null:     [2]string{"\033[38;5;243m", "\033[0m"}, // Light gray
		Escape:   [2]string{"\033[38;5;237m", "\033[0m"}, // Dark gray
		Brackets: [2]string{"\033[38;5;255m", "\033[0m"}, // White
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// ForestStyle uses deep greens and browns to create a natural, earthy look.
	ForestStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;28m", "\033[0m"},  // Dark green
		String:   [2]string{"\033[38;5;35m", "\033[0m"},  // Moss green
		Number:   [2]string{"\033[38;5;130m", "\033[0m"}, // Wood brown
		True:     [2]string{"\033[38;5;46m", "\033[0m"},  // Bright green
		False:    [2]string{"\033[38;5;88m", "\033[0m"},  // Olive
		Null:     [2]string{"\033[38;5;102m", "\033[0m"}, // Light green
		Escape:   [2]string{"\033[38;5;136m", "\033[0m"}, // Earthy red
		Brackets: [2]string{"\033[38;5;24m", "\033[0m"},  // Dark green
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// IceStyle brings a cool, frosty aesthetic with blues and whites.
	IceStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;63m", "\033[0m"},  // Icy blue
		String:   [2]string{"\033[38;5;159m", "\033[0m"}, // Frosty white
		Number:   [2]string{"\033[38;5;81m", "\033[0m"},  // Cool turquoise
		True:     [2]string{"\033[38;5;39m", "\033[0m"},  // Light blue
		False:    [2]string{"\033[38;5;35m", "\033[0m"},  // Soft blue
		Null:     [2]string{"\033[38;5;15m", "\033[0m"},  // White
		Escape:   [2]string{"\033[38;5;44m", "\033[0m"},  // Pale blue
		Brackets: [2]string{"\033[38;5;66m", "\033[0m"},  // Ice gray
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// RetroStyle brings back the vibrant colors from older computer systems and arcade games.
	RetroStyle = &encoding.Style{
		Key:      [2]string{"\033[1;38;5;208m", "\033[0m"}, // Bright orange
		String:   [2]string{"\033[1;38;5;119m", "\033[0m"}, // Light green
		Number:   [2]string{"\033[1;38;5;220m", "\033[0m"}, // Bright yellow
		True:     [2]string{"\033[1;38;5;51m", "\033[0m"},  // Neon green
		False:    [2]string{"\033[1;38;5;160m", "\033[0m"}, // Red
		Null:     [2]string{"\033[1;38;5;232m", "\033[0m"}, // Dark gray
		Escape:   [2]string{"\033[1;38;5;161m", "\033[0m"}, // Dark red
		Brackets: [2]string{"\033[1;38;5;227m", "\033[0m"}, // Light yellow
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// AutumnStyle uses rich oranges, reds, and browns, evoking the colors of fall.
	AutumnStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;130m", "\033[0m"}, // Autumn orange
		String:   [2]string{"\033[38;5;214m", "\033[0m"}, // Autumn yellow
		Number:   [2]string{"\033[38;5;52m", "\033[0m"},  // Warm brown
		True:     [2]string{"\033[38;5;166m", "\033[0m"}, // Warm red
		False:    [2]string{"\033[38;5;88m", "\033[0m"},  // Brownish red
		Null:     [2]string{"\033[38;5;240m", "\033[0m"}, // Muted brown
		Escape:   [2]string{"\033[38;5;166m", "\033[0m"}, // Red
		Brackets: [2]string{"\033[38;5;54m", "\033[0m"},  // Dark brown
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// GothicStyle uses darker colors with a moody atmosphere, ideal for dark themes.
	GothicStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;235m", "\033[0m"}, // Dark gray
		String:   [2]string{"\033[38;5;244m", "\033[0m"}, // Light gray
		Number:   [2]string{"\033[38;5;8m", "\033[0m"},   // Dark black
		True:     [2]string{"\033[38;5;68m", "\033[0m"},  // Muted green
		False:    [2]string{"\033[38;5;61m", "\033[0m"},  // Muted red
		Null:     [2]string{"\033[38;5;235m", "\033[0m"}, // Dark gray
		Escape:   [2]string{"\033[38;5;16m", "\033[0m"},  // Darkest gray
		Brackets: [2]string{"\033[38;5;232m", "\033[0m"}, // Almost black
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// VaporWaveStyle embraces the retro aesthetics of vapor-wave, with bright neon and pastel colors.
	VaporWaveStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;219m", "\033[0m"}, // Neon pink
		String:   [2]string{"\033[38;5;189m", "\033[0m"}, // Pastel purple
		Number:   [2]string{"\033[38;5;204m", "\033[0m"}, // Pastel pink
		True:     [2]string{"\033[38;5;207m", "\033[0m"}, // Pastel magenta
		False:    [2]string{"\033[38;5;142m", "\033[0m"}, // Pastel cyan
		Null:     [2]string{"\033[38;5;255m", "\033[0m"}, // White
		Escape:   [2]string{"\033[38;5;129m", "\033[0m"}, // Neon pink
		Brackets: [2]string{"\033[38;5;155m", "\033[0m"}, // Light pink
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// VampireStyle brings dark and sinister colors, with a touch of red for a spooky theme.
	VampireStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;88m", "\033[0m"},  // Blood red
		String:   [2]string{"\033[38;5;124m", "\033[0m"}, // Dark red
		Number:   [2]string{"\033[38;5;16m", "\033[0m"},  // Black
		True:     [2]string{"\033[38;5;160m", "\033[0m"}, // Vampire red
		False:    [2]string{"\033[38;5;88m", "\033[0m"},  // Blood red
		Null:     [2]string{"\033[38;5;16m", "\033[0m"},  // Black
		Escape:   [2]string{"\033[38;5;0m", "\033[0m"},   // Very dark black
		Brackets: [2]string{"\033[38;5;16m", "\033[0m"},  // Black
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// CarnivalStyle is inspired by a fun, bright carnival atmosphere, full of vivid, exciting colors.
	CarnivalStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;220m", "\033[0m"}, // Bright yellow
		String:   [2]string{"\033[38;5;204m", "\033[0m"}, // Bright pink
		Number:   [2]string{"\033[38;5;202m", "\033[0m"}, // Bright orange
		True:     [2]string{"\033[38;5;46m", "\033[0m"},  // Bright green
		False:    [2]string{"\033[38;5;160m", "\033[0m"}, // Bright red
		Null:     [2]string{"\033[38;5;214m", "\033[0m"}, // Bright light yellow
		Escape:   [2]string{"\033[38;5;213m", "\033[0m"}, // Bright purple
		Brackets: [2]string{"\033[38;5;33m", "\033[0m"},  // Deep bright blue
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// SteampunkStyle has a vintage industrial look with brass and copper colors.
	SteampunkStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;136m", "\033[0m"}, // Brass yellow
		String:   [2]string{"\033[38;5;214m", "\033[0m"}, // Copper orange
		Number:   [2]string{"\033[38;5;130m", "\033[0m"}, // Earthy brown
		True:     [2]string{"\033[38;5;184m", "\033[0m"}, // Light copper
		False:    [2]string{"\033[38;5;52m", "\033[0m"},  // Dark brown
		Null:     [2]string{"\033[38;5;94m", "\033[0m"},  // Muted gray
		Escape:   [2]string{"\033[38;5;124m", "\033[0m"}, // Dark red
		Brackets: [2]string{"\033[38;5;250m", "\033[0m"}, // Light gray
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// WoodlandStyle blends earthy tones with deep forest greens and browns.
	WoodlandStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;22m", "\033[0m"},  // Deep green
		String:   [2]string{"\033[38;5;36m", "\033[0m"},  // Moss green
		Number:   [2]string{"\033[38;5;130m", "\033[0m"}, // Earthy brown
		True:     [2]string{"\033[38;5;46m", "\033[0m"},  // Fresh green
		False:    [2]string{"\033[38;5;160m", "\033[0m"}, // Dark red
		Null:     [2]string{"\033[38;5;143m", "\033[0m"}, // Muted yellow
		Escape:   [2]string{"\033[38;5;94m", "\033[0m"},  // Brownish orange
		Brackets: [2]string{"\033[38;5;23m", "\033[0m"},  // Dark green
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// CandyStyle is bright, with pastel hues that resemble candy colors.
	CandyStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;218m", "\033[0m"}, // Pink
		String:   [2]string{"\033[38;5;226m", "\033[0m"}, // Yellow
		Number:   [2]string{"\033[38;5;222m", "\033[0m"}, // Light lime
		True:     [2]string{"\033[38;5;45m", "\033[0m"},  // Aqua
		False:    [2]string{"\033[38;5;51m", "\033[0m"},  // Light teal
		Null:     [2]string{"\033[38;5;255m", "\033[0m"}, // White
		Escape:   [2]string{"\033[38;5;196m", "\033[0m"}, // Red
		Brackets: [2]string{"\033[38;5;226m", "\033[0m"}, // Bright yellow
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// TwilightStyle brings in dusky, cool tones reminiscent of dusk.
	TwilightStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;54m", "\033[0m"},  // Dark purple
		String:   [2]string{"\033[38;5;123m", "\033[0m"}, // Light violet
		Number:   [2]string{"\033[38;5;39m", "\033[0m"},  // Deep blue
		True:     [2]string{"\033[38;5;108m", "\033[0m"}, // Soft green
		False:    [2]string{"\033[38;5;166m", "\033[0m"}, // Soft red
		Null:     [2]string{"\033[38;5;242m", "\033[0m"}, // Pale gray
		Escape:   [2]string{"\033[38;5;124m", "\033[0m"}, // Dark red
		Brackets: [2]string{"\033[38;5;239m", "\033[0m"}, // Dark gray
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// EarthStyle reflects natural earthy colors with muted greens and browns.
	EarthStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;22m", "\033[0m"},  // Deep green
		String:   [2]string{"\033[38;5;130m", "\033[0m"}, // Brown
		Number:   [2]string{"\033[38;5;52m", "\033[0m"},  // Olive
		True:     [2]string{"\033[38;5;46m", "\033[0m"},  // Forest green
		False:    [2]string{"\033[38;5;208m", "\033[0m"}, // Orange
		Null:     [2]string{"\033[38;5;130m", "\033[0m"}, // Brown
		Escape:   [2]string{"\033[38;5;94m", "\033[0m"},  // Brownish orange
		Brackets: [2]string{"\033[38;5;24m", "\033[0m"},  // Dark green
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// ElectricStyle uses electric, bright neon colors for a futuristic vibe.
	ElectricStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;51m", "\033[0m"},  // Neon green
		String:   [2]string{"\033[38;5;87m", "\033[0m"},  // Electric yellow
		Number:   [2]string{"\033[38;5;93m", "\033[0m"},  // Bright cyan
		True:     [2]string{"\033[38;5;39m", "\033[0m"},  // Bright blue
		False:    [2]string{"\033[38;5;160m", "\033[0m"}, // Electric red
		Null:     [2]string{"\033[38;5;255m", "\033[0m"}, // White
		Escape:   [2]string{"\033[38;5;196m", "\033[0m"}, // Red
		Brackets: [2]string{"\033[38;5;227m", "\033[0m"}, // Light yellow
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// WitchingHourStyle combines deep purples with dark greens for a magical look.
	WitchingHourStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;128m", "\033[0m"}, // Purple
		String:   [2]string{"\033[38;5;88m", "\033[0m"},  // Dark green
		Number:   [2]string{"\033[38;5;91m", "\033[0m"},  // Dark violet
		True:     [2]string{"\033[38;5;24m", "\033[0m"},  // Green
		False:    [2]string{"\033[38;5;231m", "\033[0m"}, // Light gray
		Null:     [2]string{"\033[38;5;234m", "\033[0m"}, // Dark gray
		Escape:   [2]string{"\033[38;5;29m", "\033[0m"},  // Midnight blue
		Brackets: [2]string{"\033[38;5;102m", "\033[0m"}, // Greenish gray
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// MidnightStyle gives a mysterious and dark aesthetic, like a quiet midnight scene.
	MidnightStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;17m", "\033[0m"},  // Deep blackish-blue
		String:   [2]string{"\033[38;5;73m", "\033[0m"},  // Teal
		Number:   [2]string{"\033[38;5;135m", "\033[0m"}, // Dark purple
		True:     [2]string{"\033[38;5;11m", "\033[0m"},  // Bright yellow
		False:    [2]string{"\033[38;5;124m", "\033[0m"}, // Dark red
		Null:     [2]string{"\033[38;5;244m", "\033[0m"}, // Light gray
		Escape:   [2]string{"\033[38;5;17m", "\033[0m"},  // Deep blackish-blue
		Brackets: [2]string{"\033[38;5;233m", "\033[0m"}, // Very dark gray
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// RetroFutureStyle combines retro tones with a futuristic neon palette for a vintage-tech feel.
	RetroFutureStyle = &encoding.Style{
		Key:      [2]string{"\033[1;38;5;214m", "\033[0m"}, // Bright orange
		String:   [2]string{"\033[1;38;5;189m", "\033[0m"}, // Pastel purple
		Number:   [2]string{"\033[1;38;5;220m", "\033[0m"}, // Bright yellow
		True:     [2]string{"\033[1;38;5;49m", "\033[0m"},  // Electric green
		False:    [2]string{"\033[1;38;5;231m", "\033[0m"}, // Light gray
		Null:     [2]string{"\033[1;38;5;250m", "\033[0m"}, // Soft white
		Escape:   [2]string{"\033[1;38;5;135m", "\033[0m"}, // Light purple
		Brackets: [2]string{"\033[1;38;5;254m", "\033[0m"}, // Off-white
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// ForestMistStyle invokes the serene and cool vibes of a misty forest.
	ForestMistStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;22m", "\033[0m"},  // Deep green
		String:   [2]string{"\033[38;5;118m", "\033[0m"}, // Sage green
		Number:   [2]string{"\033[38;5;140m", "\033[0m"}, // Earthy yellow
		True:     [2]string{"\033[38;5;45m", "\033[0m"},  // Aqua
		False:    [2]string{"\033[38;5;56m", "\033[0m"},  // Dark turquoise
		Null:     [2]string{"\033[38;5;242m", "\033[0m"}, // Misty gray
		Escape:   [2]string{"\033[38;5;235m", "\033[0m"}, // Dark gray
		Brackets: [2]string{"\033[38;5;249m", "\033[0m"}, // Light gray
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// PrismStyle offers a colorful, dazzling light prism effect for a modern, energetic look.
	PrismStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;39m", "\033[0m"},  // Light blue
		String:   [2]string{"\033[38;5;129m", "\033[0m"}, // Electric purple
		Number:   [2]string{"\033[38;5;166m", "\033[0m"}, // Bright orange
		True:     [2]string{"\033[38;5;51m", "\033[0m"},  // Neon green
		False:    [2]string{"\033[38;5;161m", "\033[0m"}, // Neon pink
		Null:     [2]string{"\033[38;5;250m", "\033[0m"}, // Light gray
		Escape:   [2]string{"\033[38;5;33m", "\033[0m"},  // Electric cyan
		Brackets: [2]string{"\033[38;5;15m", "\033[0m"},  // White
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// SpringStyle brings the fresh, light colors of spring to life.
	SpringStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;82m", "\033[0m"},  // Grass green
		String:   [2]string{"\033[38;5;214m", "\033[0m"}, // Spring yellow
		Number:   [2]string{"\033[38;5;190m", "\033[0m"}, // Soft peach
		True:     [2]string{"\033[38;5;87m", "\033[0m"},  // Light yellow-green
		False:    [2]string{"\033[38;5;161m", "\033[0m"}, // Soft red
		Null:     [2]string{"\033[38;5;249m", "\033[0m"}, // Light gray
		Escape:   [2]string{"\033[38;5;33m", "\033[0m"},  // Mint green
		Brackets: [2]string{"\033[38;5;255m", "\033[0m"}, // White
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// DesertStyle evokes the warmth and serenity of a desert landscape.
	DesertStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;95m", "\033[0m"},  // Sand beige
		String:   [2]string{"\033[38;5;130m", "\033[0m"}, // Desert orange
		Number:   [2]string{"\033[38;5;137m", "\033[0m"}, // Dusty yellow
		True:     [2]string{"\033[38;5;216m", "\033[0m"}, // Sandy yellow
		False:    [2]string{"\033[38;5;197m", "\033[0m"}, // Sunburn red
		Null:     [2]string{"\033[38;5;248m", "\033[0m"}, // Light brown
		Escape:   [2]string{"\033[38;5;167m", "\033[0m"}, // Soft red
		Brackets: [2]string{"\033[38;5;230m", "\033[0m"}, // Soft beige
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// SolarFlareStyle uses vibrant oranges and fiery reds, inspired by the intense heat of the sun.
	SolarFlareStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;214m", "\033[0m"}, // Fiery orange
		String:   [2]string{"\033[38;5;196m", "\033[0m"}, // Bright red
		Number:   [2]string{"\033[38;5;226m", "\033[0m"}, // Yellow
		True:     [2]string{"\033[38;5;220m", "\033[0m"}, // Light orange
		False:    [2]string{"\033[38;5;161m", "\033[0m"}, // Soft red
		Null:     [2]string{"\033[38;5;255m", "\033[0m"}, // White
		Escape:   [2]string{"\033[38;5;203m", "\033[0m"}, // Red-orange
		Brackets: [2]string{"\033[38;5;208m", "\033[0m"}, // Dark orange
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// IceQueenStyle reflects a cool, frosty appearance with icy blues and silvers.
	IceQueenStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;81m", "\033[0m"},  // Ice blue
		String:   [2]string{"\033[38;5;153m", "\033[0m"}, // Frosty lavender
		Number:   [2]string{"\033[38;5;77m", "\033[0m"},  // Arctic cyan
		True:     [2]string{"\033[38;5;45m", "\033[0m"},  // Winter green
		False:    [2]string{"\033[38;5;250m", "\033[0m"}, // Light gray
		Null:     [2]string{"\033[38;5;15m", "\033[0m"},  // White
		Escape:   [2]string{"\033[38;5;59m", "\033[0m"},  // Icy blue
		Brackets: [2]string{"\033[38;5;96m", "\033[0m"},  // Frosty teal
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// ForestGroveStyle brings earthy tones with a dense forest theme.
	ForestGroveStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;28m", "\033[0m"},  // Dark green
		String:   [2]string{"\033[38;5;48m", "\033[0m"},  // Olive green
		Number:   [2]string{"\033[38;5;130m", "\033[0m"}, // Earthy brown
		True:     [2]string{"\033[38;5;46m", "\033[0m"},  // Forest green
		False:    [2]string{"\033[38;5;238m", "\033[0m"}, // Dark gray
		Null:     [2]string{"\033[38;5;144m", "\033[0m"}, // Brownish yellow
		Escape:   [2]string{"\033[38;5;94m", "\033[0m"},  // Brown
		Brackets: [2]string{"\033[38;5;36m", "\033[0m"},  // Pine green
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// AutumnLeavesStyle uses warm, fall-inspired hues like browns, reds, and golden yellows.
	AutumnLeavesStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;130m", "\033[0m"}, // Pumpkin orange
		String:   [2]string{"\033[38;5;172m", "\033[0m"}, // Golden yellow
		Number:   [2]string{"\033[38;5;52m", "\033[0m"},  // Earthy brown
		True:     [2]string{"\033[38;5;214m", "\033[0m"}, // Rusty red
		False:    [2]string{"\033[38;5;136m", "\033[0m"}, // Deep maroon
		Null:     [2]string{"\033[38;5;240m", "\033[0m"}, // Light gray
		Escape:   [2]string{"\033[38;5;217m", "\033[0m"}, // Light brown
		Brackets: [2]string{"\033[38;5;95m", "\033[0m"},  // Chestnut brown
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// VaporStyle uses pastel tones and calming shades of pink, purple, and blue.
	VaporStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;189m", "\033[0m"}, // Soft pink
		String:   [2]string{"\033[38;5;153m", "\033[0m"}, // Light lavender
		Number:   [2]string{"\033[38;5;81m", "\033[0m"},  // Cool blue
		True:     [2]string{"\033[38;5;159m", "\033[0m"}, // Soft green
		False:    [2]string{"\033[38;5;129m", "\033[0m"}, // Light red
		Null:     [2]string{"\033[38;5;255m", "\033[0m"}, // White
		Escape:   [2]string{"\033[38;5;144m", "\033[0m"}, // Mint green
		Brackets: [2]string{"\033[38;5;113m", "\033[0m"}, // Light teal
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// SunsetBoulevardStyle mimics the stunning colors of a sunset, featuring warm oranges, pinks, and purples.
	SunsetBoulevardStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;214m", "\033[0m"}, // Sunset orange
		String:   [2]string{"\033[38;5;205m", "\033[0m"}, // Pinkish red
		Number:   [2]string{"\033[38;5;93m", "\033[0m"},  // Light purple
		True:     [2]string{"\033[38;5;226m", "\033[0m"}, // Bright yellow
		False:    [2]string{"\033[38;5;160m", "\033[0m"}, // Coral red
		Null:     [2]string{"\033[38;5;255m", "\033[0m"}, // White
		Escape:   [2]string{"\033[38;5;133m", "\033[0m"}, // Lavender pink
		Brackets: [2]string{"\033[38;5;57m", "\033[0m"},  // Purple
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// NeonCityStyle is bold and energetic, with electrifying neons of blue, pink, and green.
	NeonCityStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;51m", "\033[0m"},  // Neon green
		String:   [2]string{"\033[38;5;197m", "\033[0m"}, // Neon pink
		Number:   [2]string{"\033[38;5;32m", "\033[0m"},  // Neon cyan
		True:     [2]string{"\033[38;5;82m", "\033[0m"},  // Bright green
		False:    [2]string{"\033[38;5;130m", "\033[0m"}, // Bright yellow
		Null:     [2]string{"\033[38;5;7m", "\033[0m"},   // Light gray
		Escape:   [2]string{"\033[38;5;21m", "\033[0m"},  // Electric blue
		Brackets: [2]string{"\033[38;5;93m", "\033[0m"},  // Neon purple
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// MoonlitNightStyle gives a serene and calm atmosphere with cool blues and soft silvers.
	MoonlitNightStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;18m", "\033[0m"},  // Deep midnight blue
		String:   [2]string{"\033[38;5;153m", "\033[0m"}, // Soft lavender
		Number:   [2]string{"\033[38;5;110m", "\033[0m"}, // Soft teal
		True:     [2]string{"\033[38;5;48m", "\033[0m"},  // Pale green
		False:    [2]string{"\033[38;5;238m", "\033[0m"}, // Dark gray
		Null:     [2]string{"\033[38;5;15m", "\033[0m"},  // Moonlit white
		Escape:   [2]string{"\033[38;5;73m", "\033[0m"},  // Silver
		Brackets: [2]string{"\033[38;5;99m", "\033[0m"},  // Light silver
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// CandyShopStyle features bright, sugary tones of pinks, blues, and yellows for a fun and sweet theme.
	CandyShopStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;219m", "\033[0m"}, // Candy pink
		String:   [2]string{"\033[38;5;186m", "\033[0m"}, // Light yellow
		Number:   [2]string{"\033[38;5;81m", "\033[0m"},  // Light teal
		True:     [2]string{"\033[38;5;112m", "\033[0m"}, // Lime green
		False:    [2]string{"\033[38;5;196m", "\033[0m"}, // Red
		Null:     [2]string{"\033[38;5;255m", "\033[0m"}, // White
		Escape:   [2]string{"\033[38;5;222m", "\033[0m"}, // Pale yellow
		Brackets: [2]string{"\033[38;5;174m", "\033[0m"}, // Bright pink
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// UnderwaterStyle is inspired by the deep ocean, featuring calming blues and aquatic greens.
	UnderwaterStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;32m", "\033[0m"},  // Deep ocean green
		String:   [2]string{"\033[38;5;51m", "\033[0m"},  // Sea blue
		Number:   [2]string{"\033[38;5;39m", "\033[0m"},  // Ocean teal
		True:     [2]string{"\033[38;5;33m", "\033[0m"},  // Seafoam green
		False:    [2]string{"\033[38;5;236m", "\033[0m"}, // Deep sea black
		Null:     [2]string{"\033[38;5;15m", "\033[0m"},  // White
		Escape:   [2]string{"\033[38;5;73m", "\033[0m"},  // Coral pink
		Brackets: [2]string{"\033[38;5;45m", "\033[0m"},  // Greenish blue
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// OceanBreezeStyle reflects the calm and refreshing hues of the ocean.
	OceanBreezeStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;33m", "\033[0m"},  // Ocean blue
		String:   [2]string{"\033[38;5;75m", "\033[0m"},  // Seafoam green
		Number:   [2]string{"\033[38;5;51m", "\033[0m"},  // Aqua blue
		True:     [2]string{"\033[38;5;44m", "\033[0m"},  // Sea green
		False:    [2]string{"\033[38;5;61m", "\033[0m"},  // Oceanic cyan
		Null:     [2]string{"\033[38;5;250m", "\033[0m"}, // Light gray
		Escape:   [2]string{"\033[38;5;32m", "\033[0m"},  // Teal
		Brackets: [2]string{"\033[38;5;244m", "\033[0m"}, // Very light gray
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// CandyPopStyle brings a playful and sweet color palette, like a candy store.
	CandyPopStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;201m", "\033[0m"}, // Bubblegum pink
		String:   [2]string{"\033[38;5;207m", "\033[0m"}, // Candy pink
		Number:   [2]string{"\033[38;5;220m", "\033[0m"}, // Lemon yellow
		True:     [2]string{"\033[38;5;51m", "\033[0m"},  // Lime green
		False:    [2]string{"\033[38;5;160m", "\033[0m"}, // Strawberry red
		Null:     [2]string{"\033[38;5;248m", "\033[0m"}, // Pale gray
		Escape:   [2]string{"\033[38;5;33m", "\033[0m"},  // Mint green
		Brackets: [2]string{"\033[38;5;255m", "\033[0m"}, // White
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// NoirStyle gives a film-noir inspired look with dark, moody colors.
	NoirStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;16m", "\033[0m"},  // Jet black
		String:   [2]string{"\033[38;5;249m", "\033[0m"}, // Dark gray
		Number:   [2]string{"\033[38;5;235m", "\033[0m"}, // Charcoal gray
		True:     [2]string{"\033[38;5;231m", "\033[0m"}, // Off-white
		False:    [2]string{"\033[38;5;242m", "\033[0m"}, // Soft gray
		Null:     [2]string{"\033[38;5;233m", "\033[0m"}, // Dim gray
		Escape:   [2]string{"\033[38;5;245m", "\033[0m"}, // Light gray
		Brackets: [2]string{"\033[38;5;255m", "\033[0m"}, // White
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// GalacticStyle evokes the mysterious vastness of outer space with deep, cosmic hues.
	GalacticStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;54m", "\033[0m"},  // Cosmic blue
		String:   [2]string{"\033[38;5;92m", "\033[0m"},  // Starry green
		Number:   [2]string{"\033[38;5;129m", "\033[0m"}, // Galactic purple
		True:     [2]string{"\033[38;5;106m", "\033[0m"}, // Astral blue
		False:    [2]string{"\033[38;5;166m", "\033[0m"}, // Red nova
		Null:     [2]string{"\033[38;5;233m", "\033[0m"}, // Deep space gray
		Escape:   [2]string{"\033[38;5;25m", "\033[0m"},  // Nebula cyan
		Brackets: [2]string{"\033[38;5;247m", "\033[0m"}, // Light cosmic gray
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// VintagePastelStyle offers a retro aesthetic with soft, pastel tones for a gentle, nostalgic atmosphere.
	VintagePastelStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;213m", "\033[0m"}, // Soft pink
		String:   [2]string{"\033[38;5;187m", "\033[0m"}, // Lavender
		Number:   [2]string{"\033[38;5;153m", "\033[0m"}, // Pastel mint
		True:     [2]string{"\033[38;5;79m", "\033[0m"},  // Pastel green
		False:    [2]string{"\033[38;5;226m", "\033[0m"}, // Soft yellow
		Null:     [2]string{"\033[38;5;248m", "\033[0m"}, // Light beige
		Escape:   [2]string{"\033[38;5;229m", "\033[0m"}, // Pale peach
		Brackets: [2]string{"\033[38;5;245m", "\033[0m"}, // Light gray
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// VintageFilmStyle is inspired by the golden era of cinema, featuring muted golds, sepias, and classic black.
	VintageFilmStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;220m", "\033[0m"}, // Classic gold
		String:   [2]string{"\033[38;5;138m", "\033[0m"}, // Vintage brown
		Number:   [2]string{"\033[38;5;239m", "\033[0m"}, // Film grain gray
		True:     [2]string{"\033[38;5;220m", "\033[0m"}, // Film gold
		False:    [2]string{"\033[38;5;58m", "\033[0m"},  // Dark sepia
		Null:     [2]string{"\033[38;5;255m", "\033[0m"}, // Crisp white
		Escape:   [2]string{"\033[38;5;167m", "\033[0m"}, // Soft red tint
		Brackets: [2]string{"\033[38;5;59m", "\033[0m"},  // Blackish gray
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// FireworksStyle captures the excitement of a night sky lit up by colorful fireworks, featuring bold reds, yellows, and purples.
	FireworksStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;196m", "\033[0m"}, // Firework red
		String:   [2]string{"\033[38;5;226m", "\033[0m"}, // Bright yellow
		Number:   [2]string{"\033[38;5;57m", "\033[0m"},  // Purple sky
		True:     [2]string{"\033[38;5;196m", "\033[0m"}, // Firecracker red
		False:    [2]string{"\033[38;5;11m", "\033[0m"},  // Sparkling yellow
		Null:     [2]string{"\033[38;5;15m", "\033[0m"},  // White like stars
		Escape:   [2]string{"\033[38;5;201m", "\033[0m"}, // Neon orange
		Brackets: [2]string{"\033[38;5;93m", "\033[0m"},  // Bright purple
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// ArcticSnowStyle brings the cool, crisp whites and icy blues of the arctic tundra into the design.
	ArcticSnowStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;15m", "\033[0m"},  // Bright white snow
		String:   [2]string{"\033[38;5;45m", "\033[0m"},  // Icy blue
		Number:   [2]string{"\033[38;5;33m", "\033[0m"},  // Frosted teal
		True:     [2]string{"\033[38;5;8m", "\033[0m"},   // Cold gray
		False:    [2]string{"\033[38;5;7m", "\033[0m"},   // Light gray
		Null:     [2]string{"\033[38;5;255m", "\033[0m"}, // Snow white
		Escape:   [2]string{"\033[38;5;153m", "\033[0m"}, // Ice blue
		Brackets: [2]string{"\033[38;5;75m", "\033[0m"},  // Arctic blue
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// ElectricVibeStyle takes on high-energy neon tones with a touch of electric brightness.
	ElectricVibeStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;27m", "\033[0m"},  // Electric blue
		String:   [2]string{"\033[38;5;51m", "\033[0m"},  // Neon cyan
		Number:   [2]string{"\033[38;5;51m", "\033[0m"},  // Fluorescent teal
		True:     [2]string{"\033[38;5;15m", "\033[0m"},  // Bright white
		False:    [2]string{"\033[38;5;196m", "\033[0m"}, // Fiery red
		Null:     [2]string{"\033[38;5;15m", "\033[0m"},  // Pure white
		Escape:   [2]string{"\033[38;5;57m", "\033[0m"},  // Electric purple
		Brackets: [2]string{"\033[38;5;99m", "\033[0m"},  // Neon pink
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// DesertSunsetStyle brings warm and deep hues inspired by the desert landscape at sunset.
	DesertSunsetStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;216m", "\033[0m"}, // Golden yellow
		String:   [2]string{"\033[38;5;215m", "\033[0m"}, // Sunset orange
		Number:   [2]string{"\033[38;5;166m", "\033[0m"}, // Red clay
		True:     [2]string{"\033[38;5;130m", "\033[0m"}, // Sandy brown
		False:    [2]string{"\033[38;5;52m", "\033[0m"},  // Desert cactus green
		Null:     [2]string{"\033[38;5;15m", "\033[0m"},  // Desert white
		Escape:   [2]string{"\033[38;5;58m", "\033[0m"},  // Dusty purple
		Brackets: [2]string{"\033[38;5;94m", "\033[0m"},  // Desert tan
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// PastelDreamStyle evokes a dreamy, soft pastel palette perfect for relaxed and whimsical visuals.
	PastelDreamStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;226m", "\033[0m"}, // Soft yellow
		String:   [2]string{"\033[38;5;193m", "\033[0m"}, // Pastel pink
		Number:   [2]string{"\033[38;5;153m", "\033[0m"}, // Mint green
		True:     [2]string{"\033[38;5;118m", "\033[0m"}, // Pastel teal
		False:    [2]string{"\033[38;5;189m", "\033[0m"}, // Light lavender
		Null:     [2]string{"\033[38;5;248m", "\033[0m"}, // Soft gray
		Escape:   [2]string{"\033[38;5;41m", "\033[0m"},  // Soft turquoise
		Brackets: [2]string{"\033[38;5;245m", "\033[0m"}, // Light gray
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}

	// TropicalVibeStyle draws inspiration from lush tropical jungles, with bright and vibrant greens and yellows.
	TropicalVibeStyle = &encoding.Style{
		Key:      [2]string{"\033[38;5;46m", "\033[0m"},  // Leaf green
		String:   [2]string{"\033[38;5;220m", "\033[0m"}, // Yellow
		Number:   [2]string{"\033[38;5;118m", "\033[0m"}, // Jungle green
		True:     [2]string{"\033[38;5;105m", "\033[0m"}, // Lime green
		False:    [2]string{"\033[38;5;55m", "\033[0m"},  // Tropical cyan
		Null:     [2]string{"\033[38;5;250m", "\033[0m"}, // Light gray
		Escape:   [2]string{"\033[38;5;39m", "\033[0m"},  // Vibrant turquoise
		Brackets: [2]string{"\033[38;5;43m", "\033[0m"},  // Palm green
		Append:   func(dst []byte, c byte) []byte { return append(dst, c) },
	}
)
