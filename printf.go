package logs

import (
	"fmt"
	"os"
)

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

func printfColorf(color, format string, v ...any) (int, error) {
	cfg := Configured()
	text := fmt.Sprintf(format, v...)
	return fmt.Fprint(os.Stdout, colorize(color, text, cfg.NoColor))
}
