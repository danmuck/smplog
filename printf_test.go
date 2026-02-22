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

func TestMenufAppliesConfiguredColor(t *testing.T) {
	Configure(Config{
		NoColor: false,
		Colors: ConsoleColors{
			Menu: StyleColor256(14),
		},
	})
	t.Cleanup(func() { Configure(DefaultConfig()) })

	out := captureStdout(t, func() {
		if _, err := Menuf("1) status"); err != nil {
			t.Fatalf("menuf: %v", err)
		}
	})

	if !strings.Contains(out, "\x1b[38;5;14m") {
		t.Fatalf("expected menu ANSI color in output: %q", out)
	}
	if !strings.Contains(out, "1) status") {
		t.Fatalf("expected message in output: %q", out)
	}
}

func TestTitleFallsBackToInfoColor(t *testing.T) {
	Configure(Config{
		NoColor: false,
		Colors: ConsoleColors{
			Info: StyleColor256(4),
		},
	})
	t.Cleanup(func() { Configure(DefaultConfig()) })

	out := captureStdout(t, func() {
		if _, err := Title("Main Menu"); err != nil {
			t.Fatalf("title: %v", err)
		}
	})

	if !strings.Contains(out, "\x1b[38;5;4m") {
		t.Fatalf("expected fallback info ANSI color in output: %q", out)
	}
}

func TestPromptfRespectsNoColor(t *testing.T) {
	Configure(Config{
		NoColor: true,
		Colors: ConsoleColors{
			Prompt: StyleColor256(10),
		},
	})
	t.Cleanup(func() { Configure(DefaultConfig()) })

	out := captureStdout(t, func() {
		if _, err := Promptf("select> "); err != nil {
			t.Fatalf("promptf: %v", err)
		}
	})

	if strings.Contains(out, "\x1b[") {
		t.Fatalf("expected plain output with NoColor=true: %q", out)
	}
	if out != "select> " {
		t.Fatalf("unexpected prompt output: %q", out)
	}
}

func TestDividerUsesDefaultWidth(t *testing.T) {
	Configure(Config{
		NoColor: false,
		Colors: ConsoleColors{
			Divider: StyleColor256(8),
		},
		TUI: TUIConfig{
			DividerWidth: 72,
		},
	})
	t.Cleanup(func() { Configure(DefaultConfig()) })

	out := captureStdout(t, func() {
		if _, err := Divider(0); err != nil {
			t.Fatalf("divider: %v", err)
		}
	})

	plain := StripANSI(out)
	count := strings.Count(plain, "-")
	if count != 72 {
		t.Fatalf("expected %d dashes in divider, got %d (%q)", 72, count, plain)
	}
	if !strings.Contains(out, "\x1b[38;5;8m") {
		t.Fatalf("expected divider ANSI color in output: %q", out)
	}
}
