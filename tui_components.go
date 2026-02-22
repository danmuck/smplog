package logs

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"
)

// MenuEntry is a single item in a Menu component.
type MenuEntry struct {
	Label    string
	Selected bool
}

// MenuParams configures TUI.Menu.
type MenuParams struct {
	Items []MenuEntry
	Width int // 0 = TUIConfig.MaxWidth
}

// TitleParams configures TUI.MenuTitle.
type TitleParams struct {
	Text  string
	Width int
}

// SelectorParams configures TUI.Selector.
type SelectorParams struct {
	Label   string
	Items   []string
	Current int // 0-based index of current selection
	Width   int
}

// InputParams configures TUI.Input.
type InputParams struct {
	Label  string
	Value  string
	Active bool // if true, appends configured cursor character
	Width  int
}

// DividerParams configures TUI.Divider.
type DividerParams struct {
	Rune  rune // 0 = '-'
	Width int  // 0 = TUIConfig.DividerWidth
}

// TUI is a stateless component renderer. Construct with NewTUI.
type TUI struct{}

// NewTUI returns a new TUI renderer.
func NewTUI() TUI { return TUI{} }

// effectiveWidth resolves the layout width to apply.
// Priority: paramWidth → cfg.TUI.MaxWidth → 0 (unconstrained).
func effectiveWidth(paramWidth int, cfg Config) int {
	if paramWidth > 0 {
		return paramWidth
	}
	if cfg.TUI.MaxWidth > 0 {
		return cfg.TUI.MaxWidth
	}
	return 0
}

// writeComponent is the single choke-point for all single-color component output.
// It clips, colorizes, optionally centers, then writes a line to stdout.
func writeComponent(cfg Config, color, plainContent string, width int) (int, error) {
	if width > 0 {
		plainContent = Clip(width, plainContent)
	}
	colored := colorize(color, plainContent, cfg.NoColor)
	return writeComposite(cfg, colored, utf8.RuneCountInString(plainContent))
}

// writeComposite is the centering choke-point for multi-color component lines.
// line must already be fully colorized. plainWidth is the visible rune count of
// line (without ANSI escape bytes) and is used for centering math.
func writeComposite(cfg Config, line string, plainWidth int) (int, error) {
	var output string
	if cfg.TUI.Centered && cfg.TUI.MaxWidth > 0 {
		pad := max(cfg.TUI.MaxWidth-plainWidth, 0)
		left := pad / 2
		right := pad - left
		output = strings.Repeat(" ", left) + line + strings.Repeat(" ", right)
	} else {
		output = line
	}
	return fmt.Fprintln(os.Stdout, output)
}

// Menu renders a list of MenuEntry items to stdout.
// Selected entries use title color; others use menu color.
// When centered, all items share the same left margin so their prefix markers
// stay visually aligned as a block.
func (TUI) Menu(p *MenuParams) {
	cfg := Configured()
	cfg.TUI = normalizeTUIConfig(cfg.TUI)
	width := effectiveWidth(p.Width, cfg)

	// Pre-compute plain strings and find the widest one so every item can be
	// right-padded to the same width before centering.
	type row struct {
		plain string
		color string
	}
	rows := make([]row, len(p.Items))
	blockWidth := 0
	for i, entry := range p.Items {
		color := cfg.Colors.menu()
		prefix := cfg.TUI.MenuUnselectedPrefix
		if entry.Selected {
			color = cfg.Colors.title()
			prefix = cfg.TUI.MenuSelectedPrefix
		}
		plain := fmt.Sprintf("%s %*d) %s", prefix, cfg.TUI.MenuIndexWidth, i+1, entry.Label)
		rows[i] = row{plain: plain, color: color}
		if n := utf8.RuneCountInString(plain); n > blockWidth {
			blockWidth = n
		}
	}

	for _, r := range rows {
		// Pad to blockWidth so writeComponent uses a consistent centering anchor.
		padded := PadRight(blockWidth, r.plain)
		writeComponent(cfg, r.color, padded, width) //nolint:errcheck
	}
}

// MenuTitle renders a title string to stdout using the title color.
func (TUI) MenuTitle(p *TitleParams) {
	cfg := Configured()
	width := effectiveWidth(p.Width, cfg)
	writeComponent(cfg, cfg.Colors.title(), p.Text, width) //nolint:errcheck
}

// Selector renders a "label: < current >" selector row to stdout.
// It uses prompt color for the label and data color for the current item.
func (TUI) Selector(p *SelectorParams) {
	cfg := Configured()

	var current string
	if p.Current >= 0 && p.Current < len(p.Items) {
		current = p.Items[p.Current]
	}

	label := p.Label
	width := effectiveWidth(p.Width, cfg)
	if width > 0 {
		// Reserve space for ": < " (4) + " >" (2) = 6 chars around current
		labelMax := max(width-utf8.RuneCountInString(current)-6, 0)
		label = Clip(labelMax, label)
		remaining := max(width-utf8.RuneCountInString(label)-4, 0) // ": < " = 4
		current = Clip(remaining-2, current)                        // " >" = 2
	}

	plainWidth := utf8.RuneCountInString(label) + utf8.RuneCountInString(current) + 6 // ": < " + " >"
	labelText := colorize(cfg.Colors.prompt(), label, cfg.NoColor)
	currentText := colorize(cfg.Colors.data(), current, cfg.NoColor)
	line := fmt.Sprintf("%s: < %s >", labelText, currentText)
	writeComposite(cfg, line, plainWidth) //nolint:errcheck
}

// Input renders a "label: value[cursor]" input row to stdout.
// It uses prompt color for the label and data color for the value.
// When Active is true, the configured cursor character is appended.
func (TUI) Input(p *InputParams) {
	cfg := Configured()
	cfg.TUI = normalizeTUIConfig(cfg.TUI)

	value := p.Value
	width := effectiveWidth(p.Width, cfg)
	if width > 0 {
		labelRunes := utf8.RuneCountInString(p.Label) + 2 // ": " = 2
		cursorRunes := 0
		if p.Active {
			cursorRunes = utf8.RuneCountInString(cfg.TUI.InputCursor)
		}
		value = Clip(max(width-labelRunes-cursorRunes, 0), value)
	}

	cursorStr := ""
	if p.Active {
		cursorStr = cfg.TUI.InputCursor
	}
	plainWidth := utf8.RuneCountInString(p.Label) + 2 + utf8.RuneCountInString(value) + utf8.RuneCountInString(cursorStr)

	labelText := colorize(cfg.Colors.prompt(), p.Label, cfg.NoColor)
	valueText := colorize(cfg.Colors.data(), value, cfg.NoColor)
	var line string
	if p.Active {
		cursorText := colorize(cfg.Colors.prompt(), cfg.TUI.InputCursor, cfg.NoColor)
		line = fmt.Sprintf("%s: %s%s", labelText, valueText, cursorText)
	} else {
		line = fmt.Sprintf("%s: %s", labelText, valueText)
	}
	writeComposite(cfg, line, plainWidth) //nolint:errcheck
}

// Divider renders a horizontal divider line to stdout using the divider color.
func (TUI) Divider(p *DividerParams) {
	cfg := Configured()
	cfg.TUI = normalizeTUIConfig(cfg.TUI)

	r := p.Rune
	if r == 0 {
		r = '-'
	}

	width := p.Width
	if width <= 0 {
		width = effectiveWidth(0, cfg)
	}
	if width <= 0 {
		width = cfg.TUI.DividerWidth
	}

	plain := strings.Repeat(string(r), width)
	writeComponent(cfg, cfg.Colors.divider(), plain, 0) //nolint:errcheck
}

// Refresh clears the screen and moves the cursor to position (1, 1).
func (TUI) Refresh() error {
	if _, err := ClearScreen(); err != nil {
		return err
	}
	_, err := MoveTo(1, 1)
	return err
}
