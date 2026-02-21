package logs

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"
)

const (
	defaultInputCursor = "_"
)

// EnterAltScreen switches the terminal to an alternate screen buffer.
func EnterAltScreen() (int, error) {
	return writeANSI("\033[?1049h")
}

// ExitAltScreen returns the terminal to the main screen buffer.
func ExitAltScreen() (int, error) {
	return writeANSI("\033[?1049l")
}

// HideCursor hides the terminal cursor.
func HideCursor() (int, error) {
	return writeANSI("\033[?25l")
}

// ShowCursor shows the terminal cursor.
func ShowCursor() (int, error) {
	return writeANSI("\033[?25h")
}

// MoveTo moves the cursor to a 1-based row/column position.
func MoveTo(row, col int) (int, error) {
	return writeANSI(fmt.Sprintf("\033[%d;%dH", maxOne(row), maxOne(col)))
}

// ClearScreen clears the full terminal viewport.
func ClearScreen() (int, error) {
	return writeANSI("\033[2J")
}

// ClearLine clears the current line and returns the cursor to column 1.
func ClearLine() (int, error) {
	return writeANSI("\033[2K\r")
}

// WriteAt moves to row/col and writes a formatted message.
// Color output is controlled by Config.NoColor.
func WriteAt(row, col int, color, format string, v ...any) (int, error) {
	n, err := MoveTo(row, col)
	if err != nil {
		return n, err
	}
	text := fmt.Sprintf(format, v...)
	cfg := Configured()
	m, err := fmt.Fprint(os.Stdout, colorize(color, text, cfg.NoColor))
	return n + m, err
}

// Clip truncates s to width runes.
func Clip(width int, s string) string {
	if width <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= width {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	n := 0
	for _, r := range s {
		if n >= width {
			break
		}
		b.WriteRune(r)
		n++
	}
	return b.String()
}

// PadLeft left-pads s with spaces up to width runes.
// If s is wider than width, it is clipped.
func PadLeft(width int, s string) string {
	s = Clip(width, s)
	return strings.Repeat(" ", max(width-utf8.RuneCountInString(s), 0)) + s
}

// PadRight right-pads s with spaces up to width runes.
// If s is wider than width, it is clipped.
func PadRight(width int, s string) string {
	s = Clip(width, s)
	return s + strings.Repeat(" ", max(width-utf8.RuneCountInString(s), 0))
}

// Center centers s within width runes.
// If s is wider than width, it is clipped.
func Center(width int, s string) string {
	s = Clip(width, s)
	pad := max(width-utf8.RuneCountInString(s), 0)
	left := pad / 2
	right := pad - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

// MenuItem writes a compact menu entry.
// Selected entries are rendered with title color; others use menu color.
func MenuItem(index int, label string, selected bool) (int, error) {
	cfg := Configured()
	color := cfg.Colors.menu()
	prefix := " "
	if selected {
		color = cfg.Colors.title()
		prefix = ">"
	}
	return printfColorf(color, "%s %2d) %s", prefix, index, label)
}

// KeyHint writes a keyboard hint using prompt and data colors.
func KeyHint(key, desc string) (int, error) {
	cfg := Configured()
	keyText := colorize(cfg.Colors.prompt(), key, cfg.NoColor)
	descText := colorize(cfg.Colors.data(), desc, cfg.NoColor)
	return fmt.Fprintf(os.Stdout, "[%s] %s", keyText, descText)
}

// Field writes a key/value pair using prompt and data colors.
func Field(label string, value any) (int, error) {
	cfg := Configured()
	labelText := colorize(cfg.Colors.prompt(), label, cfg.NoColor)
	valueText := colorize(cfg.Colors.data(), fmt.Sprint(value), cfg.NoColor)
	return fmt.Fprintf(os.Stdout, "%s: %s", labelText, valueText)
}

// StatusInfo writes an info-status message.
func StatusInfo(msg string) (int, error) {
	return printfColorf(Configured().Colors.level("info"), "%s", msg)
}

// StatusWarn writes a warning-status message.
func StatusWarn(msg string) (int, error) {
	return printfColorf(Configured().Colors.level("warn"), "%s", msg)
}

// StatusError writes an error-status message.
func StatusError(msg string) (int, error) {
	return printfColorf(Configured().Colors.level("error"), "%s", msg)
}

// InputLine writes a compact prompt/value input row.
// If active, a lightweight cursor marker is appended.
func InputLine(prefix, value string, active bool) (int, error) {
	cfg := Configured()
	prefixText := colorize(cfg.Colors.prompt(), prefix, cfg.NoColor)
	valueText := colorize(cfg.Colors.data(), value, cfg.NoColor)
	if !active {
		return fmt.Fprintf(os.Stdout, "%s%s", prefixText, valueText)
	}
	cursor := colorize(cfg.Colors.prompt(), defaultInputCursor, cfg.NoColor)
	return fmt.Fprintf(os.Stdout, "%s%s%s", prefixText, valueText, cursor)
}

// BeginFrame switches to alt-screen, hides the cursor, clears the frame,
// and positions the cursor at 1,1.
func BeginFrame() error {
	if _, err := EnterAltScreen(); err != nil {
		return err
	}
	if _, err := HideCursor(); err != nil {
		return err
	}
	if _, err := ClearScreen(); err != nil {
		return err
	}
	_, err := MoveTo(1, 1)
	return err
}

// EndFrame restores the cursor and returns to the main screen.
func EndFrame() error {
	if _, err := ShowCursor(); err != nil {
		return err
	}
	_, err := ExitAltScreen()
	return err
}

func writeANSI(seq string) (int, error) {
	return fmt.Fprint(os.Stdout, seq)
}

func maxOne(n int) int {
	if n < 1 {
		return 1
	}
	return n
}
