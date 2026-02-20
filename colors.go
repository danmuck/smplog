package logs

import (
	"fmt"
	"regexp"
	"strings"
)

// Palette structure:
//   - 0..7     : base colors (black, red, green, yellow, blue, magenta, cyan, white)
//   - 8..15    : bright versions of 0..7
//   - 16..231  : color cube, index formula:
//     N = 16 + (36 * r) + (6 * g) + b, where r,g,b in [0..5]
//     channel levels are approximately: 0, 95, 135, 175, 215, 255
//   - 232..255 : grayscale, 24 shades
//     N = 232 + k, k in [0..23]
const (
	// ANSI-256 base palette indexes.
	Black   = 0
	Red     = 1
	Green   = 2
	Yellow  = 3
	Blue    = 4
	Magenta = 5
	Cyan    = 6
	White   = 7

	BrightBlack   = 8
	BrightRed     = 9
	BrightGreen   = 10
	BrightYellow  = 11
	BrightBlue    = 12
	BrightMagenta = 13
	BrightCyan    = 14
	BrightWhite   = 15
)

var (
	// StyleReset clears all ANSI attributes.
	StyleReset = sgr(0)

	// Text attributes.
	StyleBold      = sgr(1)
	StyleDim       = sgr(2)
	StyleItalic    = sgr(3)
	StyleUnderline = sgr(4)
	StyleBlink     = sgr(5)
	StyleReverse   = sgr(7)
	StyleHidden    = sgr(8)
	StyleStrike    = sgr(9)
)

// ansiPattern matches CSI Select Graphic Rendition (SGR) sequences,
// e.g. "\x1b[31m" (red) and "\x1b[0m" (reset).
var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// ConsoleColors defines ANSI colors applied by the console wrapper.
type ConsoleColors struct {
	Trace      string
	Debug      string
	Info       string
	Warn       string
	Error      string
	Fatal      string
	Panic      string
	Message    string
	Timestamp  string
	FieldName  string
	FieldValue string
}

// DefaultColors returns the default level-based color palette.
func DefaultColors() ConsoleColors {
	return ConsoleColors{
		Trace:      StyleColor256(BrightBlack),
		Debug:      StyleColor256(Green),
		Info:       StyleColor256(Blue),
		Warn:       StyleColor256(Yellow),
		Error:      StyleColor256(Red),
		Fatal:      StyleColor256(196),
		Panic:      StyleColor256(Magenta),
		Timestamp:  StyleColor256(BrightBlack),
		FieldName:  StyleColor256(Cyan),
		FieldValue: StyleColor256(White),
	}
}

// NoColors returns a palette with no color overrides.
func NoColors() ConsoleColors {
	return ConsoleColors{}
}

// level returns the configured color for a level name string.
func (c ConsoleColors) level(level string) string {
	switch strings.ToLower(level) {
	case "trace":
		return c.Trace
	case "debug":
		return c.Debug
	case "info":
		return c.Info
	case "warn", "warning":
		return c.Warn
	case "error":
		return c.Error
	case "fatal":
		return c.Fatal
	case "panic":
		return c.Panic
	default:
		return ""
	}
}

// colorize wraps text with a style prefix and a trailing reset sequence.
// It is a no-op when color output is disabled or style/text is empty.
func colorize(color, text string, disabled bool) string {
	if disabled || color == "" || text == "" {
		return text
	}
	return fmt.Sprintf("%s%s%s", color, text, StyleReset)
}

// sgr returns a Select Graphic Rendition control sequence for a numeric code.
// Example: sgr(1) => bold ("\x1b[1m"), sgr(0) => reset ("\x1b[0m").
func sgr(n int) string {
	return fmt.Sprintf("\033[%dm", n)
}

// StyleColor256 returns a 256-color ANSI foreground escape sequence.
func StyleColor256(n int) string {
	return fmt.Sprintf("\033[38;5;%dm", n)
}

// BgColor256 returns a 256-color ANSI background escape sequence.
func BgColor256(n int) string {
	return fmt.Sprintf("\033[48;5;%dm", n)
}

// StripANSI removes ANSI escape sequences from s.
func StripANSI(s string) string {
	return ansiPattern.ReplaceAllString(s, "")
}
