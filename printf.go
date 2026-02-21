package logs

import (
	"fmt"
	"os"
	"strings"
)

const defaultDividerWidth = 64

// Print writes msg to stdout.
func Print(msg string) (int, error) {
	return fmt.Fprint(os.Stdout, msg)
}

// Printf writes a formatted message to stdout.
func Printf(format string, v ...any) (int, error) {
	return fmt.Fprintf(os.Stdout, format, v...)
}

// Println writes msg and a trailing newline to stdout.
func Println(msg string) (int, error) {
	return fmt.Fprintln(os.Stdout, msg)
}

// Colorf writes a formatted message to stdout with an ANSI style prefix.
// When Config.NoColor is enabled, output is written without ANSI escapes.
func Colorf(color, format string, v ...any) (int, error) {
	return printfColorf(color, format, v...)
}

// Menu writes msg using the configured menu color.
func Menu(msg string) (int, error) {
	return Menuf("%s", msg)
}

// Menuf writes a formatted menu message with Config.Colors.Menu.
func Menuf(format string, v ...any) (int, error) {
	return printfColorf(Configured().Colors.menu(), format, v...)
}

// Title writes msg using the configured title color.
func Title(msg string) (int, error) {
	return Titlef("%s", msg)
}

// Titlef writes a formatted title message with Config.Colors.Title.
func Titlef(format string, v ...any) (int, error) {
	return printfColorf(Configured().Colors.title(), format, v...)
}

// Prompt writes msg using the configured prompt color.
func Prompt(msg string) (int, error) {
	return Promptf("%s", msg)
}

// Promptf writes a formatted prompt message with Config.Colors.Prompt.
func Promptf(format string, v ...any) (int, error) {
	return printfColorf(Configured().Colors.prompt(), format, v...)
}

// Data writes msg using the configured data color.
func Data(msg string) (int, error) {
	return Dataf("%s", msg)
}

// Dataf writes a formatted data message with Config.Colors.Data.
func Dataf(format string, v ...any) (int, error) {
	return printfColorf(Configured().Colors.data(), format, v...)
}

// DataKV writes a key/value pair using the configured data color.
func DataKV(key string, value any) (int, error) {
	return Dataf("%s: %v", key, value)
}

// Divider writes a horizontal divider using '-' and Config.Colors.Divider.
// If width <= 0, a default width is used.
func Divider(width int) (int, error) {
	return DividerRune(width, '-')
}

// DividerRune writes a horizontal divider using r and Config.Colors.Divider.
// If width <= 0, a default width is used.
func DividerRune(width int, r rune) (int, error) {
	if width <= 0 {
		width = defaultDividerWidth
	}
	if r == 0 {
		r = '-'
	}
	line := strings.Repeat(string(r), width)
	return printfColorf(Configured().Colors.divider(), "%s", line)
}

func printfColorf(color, format string, v ...any) (int, error) {
	cfg := Configured()
	text := fmt.Sprintf(format, v...)
	return fmt.Fprint(os.Stdout, colorize(color, text, cfg.NoColor))
}
