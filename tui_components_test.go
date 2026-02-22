package logs

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestTUIMenuRendersAllItems(t *testing.T) {
	Configure(Config{
		NoColor: false,
		Colors: ConsoleColors{
			Menu:  StyleColor256(14),
			Title: StyleColor256(15),
		},
		TUI: TUIConfig{
			MenuSelectedPrefix:   ">",
			MenuUnselectedPrefix: " ",
			MenuIndexWidth:       2,
		},
	})
	t.Cleanup(func() { Configure(DefaultConfig()) })

	out := captureStdout(t, func() {
		NewTUI().Menu(&MenuParams{
			Items: []MenuEntry{
				{Label: "alpha", Selected: true},
				{Label: "beta", Selected: false},
			},
		})
	})

	// Both items should appear
	if !strings.Contains(out, "alpha") {
		t.Fatalf("expected 'alpha' in output: %q", out)
	}
	if !strings.Contains(out, "beta") {
		t.Fatalf("expected 'beta' in output: %q", out)
	}
	// Selected item uses title color
	if !strings.Contains(out, "\x1b[38;5;15m") {
		t.Fatalf("expected title color for selected item: %q", out)
	}
	// Unselected item uses menu color
	if !strings.Contains(out, "\x1b[38;5;14m") {
		t.Fatalf("expected menu color for unselected item: %q", out)
	}
	// Item numbering should be present
	plain := StripANSI(out)
	if !strings.Contains(plain, "1)") {
		t.Fatalf("expected '1)' in plain output: %q", plain)
	}
	if !strings.Contains(plain, "2)") {
		t.Fatalf("expected '2)' in plain output: %q", plain)
	}
}

func TestTUIMenuNoColor(t *testing.T) {
	Configure(Config{
		NoColor: true,
		Colors: ConsoleColors{
			Menu:  StyleColor256(14),
			Title: StyleColor256(15),
		},
	})
	t.Cleanup(func() { Configure(DefaultConfig()) })

	out := captureStdout(t, func() {
		NewTUI().Menu(&MenuParams{
			Items: []MenuEntry{
				{Label: "item", Selected: false},
			},
		})
	})

	if strings.Contains(out, "\x1b[") {
		t.Fatalf("expected no ANSI escapes with NoColor=true: %q", out)
	}
	if !strings.Contains(out, "item") {
		t.Fatalf("expected 'item' in output: %q", out)
	}
}

func TestTUIMenuTitleUsesTitleColor(t *testing.T) {
	Configure(Config{
		NoColor: false,
		Colors: ConsoleColors{
			Title: StyleColor256(15),
		},
	})
	t.Cleanup(func() { Configure(DefaultConfig()) })

	out := captureStdout(t, func() {
		NewTUI().MenuTitle(&TitleParams{Text: "Main Menu"})
	})

	if !strings.Contains(out, "\x1b[38;5;15m") {
		t.Fatalf("expected title color escape in output: %q", out)
	}
	if !strings.Contains(out, "Main Menu") {
		t.Fatalf("expected title text in output: %q", out)
	}
}

func TestTUISelectorRendersLabelAndCurrentItem(t *testing.T) {
	Configure(Config{
		NoColor: false,
		Colors: ConsoleColors{
			Prompt: StyleColor256(10),
			Data:   StyleColor256(7),
		},
	})
	t.Cleanup(func() { Configure(DefaultConfig()) })

	out := captureStdout(t, func() {
		NewTUI().Selector(&SelectorParams{
			Label:   "mode",
			Items:   []string{"a", "b", "c"},
			Current: 1, // "b"
		})
	})

	plain := StripANSI(out)
	if !strings.Contains(plain, "< b >") {
		t.Fatalf("expected '< b >' in plain output: %q", plain)
	}
	if !strings.Contains(plain, "mode") {
		t.Fatalf("expected label in plain output: %q", plain)
	}
	// data color should be present for current item
	if !strings.Contains(out, "\x1b[38;5;7m") {
		t.Fatalf("expected data color in output: %q", out)
	}
}

func TestTUISelectorOutOfBoundsCurrentIsEmpty(t *testing.T) {
	Configure(Config{NoColor: true})
	t.Cleanup(func() { Configure(DefaultConfig()) })

	// Should not panic; current="" when index is out of bounds
	out := captureStdout(t, func() {
		NewTUI().Selector(&SelectorParams{
			Label:   "opt",
			Items:   []string{"x"},
			Current: 99,
		})
	})

	// "< %s >" with empty string gives "< >"-style output (space on each side)
	if !strings.Contains(out, "<") || !strings.Contains(out, ">") {
		t.Fatalf("expected selector brackets in out-of-bounds output: %q", out)
	}
}

func TestTUIInputActiveRendersLabelValueCursor(t *testing.T) {
	Configure(Config{
		NoColor: false,
		Colors: ConsoleColors{
			Prompt: StyleColor256(10),
			Data:   StyleColor256(7),
		},
		TUI: TUIConfig{InputCursor: "|"},
	})
	t.Cleanup(func() { Configure(DefaultConfig()) })

	out := captureStdout(t, func() {
		NewTUI().Input(&InputParams{
			Label:  "name",
			Value:  "dan",
			Active: true,
		})
	})

	plain := StripANSI(out)
	if !strings.Contains(plain, "|") {
		t.Fatalf("expected cursor '|' in plain output: %q", plain)
	}
	if !strings.Contains(plain, "name") {
		t.Fatalf("expected label in plain output: %q", plain)
	}
	if !strings.Contains(plain, "dan") {
		t.Fatalf("expected value in plain output: %q", plain)
	}
}

func TestTUIInputInactiveOmitsCursor(t *testing.T) {
	Configure(Config{
		NoColor: true,
		TUI:     TUIConfig{InputCursor: "|"},
	})
	t.Cleanup(func() { Configure(DefaultConfig()) })

	out := captureStdout(t, func() {
		NewTUI().Input(&InputParams{
			Label:  "name",
			Value:  "dan",
			Active: false,
		})
	})

	plain := StripANSI(out)
	if strings.Contains(plain, "|") {
		t.Fatalf("expected no cursor in inactive input: %q", plain)
	}
}

func TestTUIDividerUsesConfigWidth(t *testing.T) {
	Configure(Config{
		NoColor: false,
		Colors: ConsoleColors{
			Divider: StyleColor256(8),
		},
		TUI: TUIConfig{DividerWidth: 40},
	})
	t.Cleanup(func() { Configure(DefaultConfig()) })

	out := captureStdout(t, func() {
		NewTUI().Divider(&DividerParams{})
	})

	plain := strings.TrimRight(StripANSI(out), "\n")
	if utf8.RuneCountInString(plain) != 40 {
		t.Fatalf("expected divider rune count 40, got %d (%q)", utf8.RuneCountInString(plain), plain)
	}
}

func TestTUIDividerParamWidthOverrides(t *testing.T) {
	Configure(Config{
		NoColor: true,
		TUI:     TUIConfig{DividerWidth: 40},
	})
	t.Cleanup(func() { Configure(DefaultConfig()) })

	out := captureStdout(t, func() {
		NewTUI().Divider(&DividerParams{Width: 20})
	})

	plain := strings.TrimRight(StripANSI(out), "\n")
	if utf8.RuneCountInString(plain) != 20 {
		t.Fatalf("expected divider rune count 20, got %d (%q)", utf8.RuneCountInString(plain), plain)
	}
}

func TestTUIDividerCustomRune(t *testing.T) {
	Configure(Config{
		NoColor: true,
		TUI:     TUIConfig{DividerWidth: 10},
	})
	t.Cleanup(func() { Configure(DefaultConfig()) })

	out := captureStdout(t, func() {
		NewTUI().Divider(&DividerParams{Rune: '='})
	})

	plain := strings.TrimRight(StripANSI(out), "\n")
	if plain != strings.Repeat("=", 10) {
		t.Fatalf("expected '=' repeated 10 times, got %q", plain)
	}
}

func TestTUIWidthClampsTruncates(t *testing.T) {
	Configure(Config{
		NoColor: true,
		Colors:  ConsoleColors{Title: StyleColor256(15)},
	})
	t.Cleanup(func() { Configure(DefaultConfig()) })

	out := captureStdout(t, func() {
		NewTUI().MenuTitle(&TitleParams{Text: "Hello World", Width: 5})
	})

	plain := strings.TrimRight(StripANSI(out), "\n")
	if utf8.RuneCountInString(plain) != 5 {
		t.Fatalf("expected 5 runes after clipping, got %d (%q)", utf8.RuneCountInString(plain), plain)
	}
	if plain != "Hello" {
		t.Fatalf("expected 'Hello', got %q", plain)
	}
}

func TestTUICenteringPadsContent(t *testing.T) {
	Configure(Config{
		NoColor: true,
		TUI:     TUIConfig{MaxWidth: 20, Centered: true},
	})
	t.Cleanup(func() { Configure(DefaultConfig()) })

	out := captureStdout(t, func() {
		NewTUI().MenuTitle(&TitleParams{Text: "Hi"})
	})

	// Strip trailing newline for analysis
	line := strings.TrimRight(StripANSI(out), "\n")
	total := utf8.RuneCountInString(line)
	if total != 20 {
		t.Fatalf("expected total visible width 20, got %d (%q)", total, line)
	}
	if !strings.HasPrefix(line, " ") {
		t.Fatalf("expected leading spaces for centering: %q", line)
	}
	if !strings.HasSuffix(line, " ") {
		t.Fatalf("expected trailing spaces for centering: %q", line)
	}
}

func TestTUICenteringRequiresMaxWidth(t *testing.T) {
	Configure(Config{
		NoColor: true,
		TUI:     TUIConfig{MaxWidth: 0, Centered: true},
	})
	t.Cleanup(func() { Configure(DefaultConfig()) })

	out := captureStdout(t, func() {
		NewTUI().MenuTitle(&TitleParams{Text: "Hi"})
	})

	line := strings.TrimRight(StripANSI(out), "\n")
	// Without MaxWidth, no padding should be added
	if line != "Hi" {
		t.Fatalf("expected 'Hi' without padding when MaxWidth=0, got %q", line)
	}
}

func TestTUISelectorCenteringPadsContent(t *testing.T) {
	Configure(Config{
		NoColor: true,
		TUI:     TUIConfig{MaxWidth: 30, Centered: true},
	})
	t.Cleanup(func() { Configure(DefaultConfig()) })

	out := captureStdout(t, func() {
		NewTUI().Selector(&SelectorParams{
			Label:   "x",
			Items:   []string{"y"},
			Current: 0,
		})
	})

	// "x: < y >" = 9 runes; padded to 30 total
	line := strings.TrimRight(StripANSI(out), "\n")
	total := utf8.RuneCountInString(line)
	if total != 30 {
		t.Fatalf("expected total visible width 30, got %d (%q)", total, line)
	}
	if !strings.HasPrefix(line, " ") {
		t.Fatalf("expected leading spaces for centering: %q", line)
	}
}

func TestTUIInputCenteringPadsContent(t *testing.T) {
	Configure(Config{
		NoColor: true,
		TUI:     TUIConfig{MaxWidth: 30, Centered: true, InputCursor: "|"},
	})
	t.Cleanup(func() { Configure(DefaultConfig()) })

	out := captureStdout(t, func() {
		NewTUI().Input(&InputParams{
			Label:  "name",
			Value:  "dan",
			Active: true,
		})
	})

	// "name: dan|" = 10 runes; padded to 30 total
	line := strings.TrimRight(StripANSI(out), "\n")
	total := utf8.RuneCountInString(line)
	if total != 30 {
		t.Fatalf("expected total visible width 30, got %d (%q)", total, line)
	}
	if !strings.HasPrefix(line, " ") {
		t.Fatalf("expected leading spaces for centering: %q", line)
	}
}

// TestTUIComponentsVisualExample renders all components to stdout so you can
// inspect them visually. Run with:
//
//	go test -v -run TestTUIComponentsVisualExample ./...
//
// The test renders two back-to-back scenes:
//   - Scene 1: centered layout (MaxWidth=48, Centered=true)
//   - Scene 2: left-aligned layout (no centering) with different palette
func TestTUIComponentsVisualExample(t *testing.T) {
	tui := NewTUI()

	// ── Scene 1: centered, 48-wide ─────────────────────────────────────────
	t.Log("Scene 1: Centered layout (MaxWidth=48)")
	Configure(Config{
		NoColor: false,
		Colors: ConsoleColors{
			Title:   StyleColor256(15), // bright white
			Menu:    StyleColor256(14), // bright cyan
			Prompt:  StyleColor256(10), // bright green
			Data:    StyleColor256(7),  // white
			Divider: StyleColor256(8),  // gray
		},
		TUI: TUIConfig{
			MaxWidth:             48,
			Centered:             true,
			MenuSelectedPrefix:   ">",
			MenuUnselectedPrefix: " ",
			MenuIndexWidth:       2,
			InputCursor:          "_",
			DividerWidth:         48,
		},
	})

	tui.Divider(&DividerParams{})
	tui.MenuTitle(&TitleParams{Text: "Main Menu"})
	tui.Divider(&DividerParams{})
	tui.Menu(&MenuParams{Items: []MenuEntry{
		{Label: "Status", Selected: true},
		{Label: "Settings", Selected: false},
		{Label: "Logs", Selected: false},
		{Label: "Quit", Selected: false},
	}})
	tui.Divider(&DividerParams{})
	tui.Selector(&SelectorParams{
		Label:   "mode",
		Items:   []string{"debug", "verbose", "silent"},
		Current: 1,
	})
	tui.Input(&InputParams{Label: "filter", Value: "error", Active: false})
	tui.Input(&InputParams{Label: "output", Value: "stdout", Active: true})
	tui.Divider(&DividerParams{})

	// ── Scene 2: left-aligned, different palette ────────────────────────────
	t.Log("Scene 2: Left-aligned layout (no centering)")
	Configure(Config{
		NoColor: false,
		Colors: ConsoleColors{
			Title:   StyleColor256(11), // bright yellow
			Menu:    StyleColor256(12), // bright blue
			Prompt:  StyleColor256(13), // bright magenta
			Data:    StyleColor256(7),  // white
			Divider: StyleColor256(6),  // cyan
		},
		TUI: TUIConfig{
			MaxWidth:             0,
			Centered:             false,
			MenuSelectedPrefix:   "▶",
			MenuUnselectedPrefix: " ",
			MenuIndexWidth:       2,
			InputCursor:          "█",
			DividerWidth:         40,
		},
	})

	tui.Divider(&DividerParams{Rune: '='})
	tui.MenuTitle(&TitleParams{Text: "Settings"})
	tui.Divider(&DividerParams{Rune: '-'})
	tui.Menu(&MenuParams{Items: []MenuEntry{
		{Label: "Network", Selected: false},
		{Label: "Storage", Selected: true},
		{Label: "Security", Selected: false},
	}})
	tui.Divider(&DividerParams{Rune: '-'})
	tui.Selector(&SelectorParams{
		Label:   "theme",
		Items:   []string{"dark", "light", "system"},
		Current: 0,
	})
	tui.Input(&InputParams{Label: "alias", Value: "dev-box", Active: true})
	tui.Divider(&DividerParams{Rune: '='})

	t.Cleanup(func() { Configure(DefaultConfig()) })
}

func TestTUIRefreshWritesClearAndMoveTo(t *testing.T) {
	out := captureStdout(t, func() {
		if err := NewTUI().Refresh(); err != nil {
			t.Fatalf("refresh: %v", err)
		}
	})

	if !strings.Contains(out, "\x1b[2J") {
		t.Fatalf("expected clear screen sequence in output: %q", out)
	}
	if !strings.Contains(out, "\x1b[1;1H") {
		t.Fatalf("expected move-to-1,1 sequence in output: %q", out)
	}
}
