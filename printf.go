package logs

import (
	"fmt"
	"io"
	"os"
)

// Fprint writes msg to w.
func Fprint(w io.Writer, msg string) (int, error) {
	return fmt.Fprint(w, msg)
}

// Fprintf writes a formatted message to w.
func Fprintf(w io.Writer, format string, v ...any) (int, error) {
	return fmt.Fprintf(w, format, v...)
}

// Fprintln writes msg and a trailing newline to w.
func Fprintln(w io.Writer, msg string) (int, error) {
	return fmt.Fprintln(w, msg)
}

// Fcolorf writes a formatted message to w with an ANSI style prefix.
// When Config.NoColor is enabled, output is written without ANSI escapes.
func Fcolorf(w io.Writer, color, format string, v ...any) (int, error) {
	cfg := Configured()
	text := fmt.Sprintf(format, v...)
	return fmt.Fprint(w, colorize(color, text, cfg.NoColor))
}

// Print writes msg to stdout.
func Print(msg string) (int, error) {
	return Fprint(os.Stdout, msg)
}

// Printf writes a formatted message to stdout.
func Printf(format string, v ...any) (int, error) {
	return Fprintf(os.Stdout, format, v...)
}

// Println writes msg and a trailing newline to stdout.
func Println(msg string) (int, error) {
	return Fprintln(os.Stdout, msg)
}

// Colorf writes a formatted message to stdout with an ANSI style prefix.
// When Config.NoColor is enabled, output is written without ANSI escapes.
func Colorf(color, format string, v ...any) (int, error) {
	return Fcolorf(os.Stdout, color, format, v...)
}
