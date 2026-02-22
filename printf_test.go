package logs

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	defer r.Close()

	os.Stdout = w
	defer func() { os.Stdout = old }()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	var out bytes.Buffer
	if _, err := io.Copy(&out, r); err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	return out.String()
}

func TestColorfAppliesANSIColor(t *testing.T) {
	Configure(Config{
		NoColor: false,
		Colors: ConsoleColors{
			Menu: StyleColor256(14),
		},
	})
	t.Cleanup(func() { Configure(DefaultConfig()) })

	out := captureStdout(t, func() {
		if _, err := Colorf(StyleColor256(14), "1) status"); err != nil {
			t.Fatalf("colorf: %v", err)
		}
	})

	if !strings.Contains(out, "\x1b[38;5;14m") {
		t.Fatalf("expected ANSI color in output: %q", out)
	}
	if !strings.Contains(out, "1) status") {
		t.Fatalf("expected message in output: %q", out)
	}
}

func TestColorfRespectsNoColor(t *testing.T) {
	Configure(Config{
		NoColor: true,
		Colors: ConsoleColors{
			Prompt: StyleColor256(10),
		},
	})
	t.Cleanup(func() { Configure(DefaultConfig()) })

	out := captureStdout(t, func() {
		if _, err := Colorf(StyleColor256(10), "select> "); err != nil {
			t.Fatalf("colorf: %v", err)
		}
	})

	if strings.Contains(out, "\x1b[") {
		t.Fatalf("expected plain output with NoColor=true: %q", out)
	}
	if out != "select> " {
		t.Fatalf("unexpected output: %q", out)
	}
}
