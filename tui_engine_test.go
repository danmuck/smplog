package logs

import (
	"strings"
	"testing"
)

func TestWriteAtMovesCursorAndColorizes(t *testing.T) {
	Configure(Config{
		NoColor: false,
		Colors: ConsoleColors{
			Menu: StyleColor256(14),
		},
	})
	t.Cleanup(func() { Configure(DefaultConfig()) })

	out := captureStdout(t, func() {
		if _, err := WriteAt(3, 5, Configured().Colors.menu(), "node:%d", 7); err != nil {
			t.Fatalf("writeat: %v", err)
		}
	})

	if !strings.Contains(out, "\x1b[3;5H") {
		t.Fatalf("expected cursor move sequence in output: %q", out)
	}
	if !strings.Contains(out, "\x1b[38;5;14m") {
		t.Fatalf("expected color sequence in output: %q", out)
	}
	if !strings.Contains(out, "node:7") {
		t.Fatalf("expected payload in output: %q", out)
	}
}

func TestWriteAtRespectsNoColor(t *testing.T) {
	Configure(Config{
		NoColor: true,
		Colors: ConsoleColors{
			Menu: StyleColor256(14),
		},
	})
	t.Cleanup(func() { Configure(DefaultConfig()) })

	out := captureStdout(t, func() {
		if _, err := WriteAt(1, 1, Configured().Colors.menu(), "plain"); err != nil {
			t.Fatalf("writeat: %v", err)
		}
	})

	if strings.Contains(out, "\x1b[38;5;14m") {
		t.Fatalf("expected no color sequence with NoColor=true: %q", out)
	}
	if !strings.Contains(out, "\x1b[1;1Hplain") {
		t.Fatalf("expected move+payload output: %q", out)
	}
}

func TestClipPadCenter(t *testing.T) {
	if got := Clip(4, "abcdef"); got != "abcd" {
		t.Fatalf("clip: got %q want %q", got, "abcd")
	}
	if got := PadLeft(6, "xy"); got != "    xy" {
		t.Fatalf("padleft: got %q want %q", got, "    xy")
	}
	if got := PadRight(6, "xy"); got != "xy    " {
		t.Fatalf("padright: got %q want %q", got, "xy    ")
	}
	if got := Center(7, "abc"); got != "  abc  " {
		t.Fatalf("center: got %q want %q", got, "  abc  ")
	}
}

func TestMenuItemSelectionUsesTitleColor(t *testing.T) {
	Configure(Config{
		NoColor: false,
		Colors: ConsoleColors{
			Menu:  StyleColor256(14),
			Title: StyleColor256(15),
		},
		TUI: TUIConfig{
			MenuSelectedPrefix:   ">>",
			MenuUnselectedPrefix: "..",
			MenuIndexWidth:       4,
			InputCursor:          "|",
			DividerWidth:         72,
		},
	})
	t.Cleanup(func() { Configure(DefaultConfig()) })

	out := captureStdout(t, func() {
		if _, err := MenuItem(2, "services", true); err != nil {
			t.Fatalf("menuitem: %v", err)
		}
	})

	if !strings.Contains(out, "\x1b[38;5;15m") {
		t.Fatalf("expected selected title color in output: %q", out)
	}
	if !strings.Contains(out, ">>    2) services") {
		t.Fatalf("expected selected item payload in output: %q", out)
	}
}

func TestBeginEndFrameWritesExpectedSequences(t *testing.T) {
	out := captureStdout(t, func() {
		if err := BeginFrame(); err != nil {
			t.Fatalf("begin frame: %v", err)
		}
		if err := EndFrame(); err != nil {
			t.Fatalf("end frame: %v", err)
		}
	})

	required := []string{
		"\x1b[?1049h",
		"\x1b[?25l",
		"\x1b[2J",
		"\x1b[1;1H",
		"\x1b[?25h",
		"\x1b[?1049l",
	}
	for _, seq := range required {
		if !strings.Contains(out, seq) {
			t.Fatalf("expected sequence %q in output %q", seq, out)
		}
	}
}

func TestKeyHintFieldAndInputLineNoColor(t *testing.T) {
	Configure(Config{
		NoColor: true,
		Colors: ConsoleColors{
			Prompt: StyleColor256(10),
			Data:   StyleColor256(7),
		},
		TUI: TUIConfig{
			InputCursor: "|",
		},
	})
	t.Cleanup(func() { Configure(DefaultConfig()) })

	out := captureStdout(t, func() {
		if _, err := KeyHint("q", "quit"); err != nil {
			t.Fatalf("keyhint: %v", err)
		}
		if _, err := Field("mode", "diag"); err != nil {
			t.Fatalf("field: %v", err)
		}
		if _, err := InputLine("select> ", "2", true); err != nil {
			t.Fatalf("inputline: %v", err)
		}
	})

	if strings.Contains(out, "\x1b[") {
		t.Fatalf("expected no ANSI when NoColor=true: %q", out)
	}
	if !strings.Contains(out, "[q] quit") {
		t.Fatalf("expected key hint payload in output: %q", out)
	}
	if !strings.Contains(out, "mode: diag") {
		t.Fatalf("expected field payload in output: %q", out)
	}
	if !strings.Contains(out, "select> 2|") {
		t.Fatalf("expected active input payload in output: %q", out)
	}
}
