package logs

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

// fileConfig is the TOML-decodable shape of Config.
//
// Fields that require code — Writer, ConfigureZerolog, ConfigureConsole,
// ConfigureLogger — cannot be expressed in a file and must be set on the
// returned Config programmatically before calling Configure.
type fileConfig struct {
	Level      string      `toml:"level"`
	Timestamp  bool        `toml:"timestamp"`
	Caller     bool        `toml:"caller"`
	Stack      bool        `toml:"stack"`
	TimeFormat string      `toml:"time_format"`
	NoColor    bool        `toml:"no_color"`
	Bypass     bool        `toml:"bypass"`
	Colors     colorConfig `toml:"colors"`
	TUI        []tuiConfig `toml:"tui"`
	Files      []LogFile   `toml:"files"`
}

// colorConfig is the [colors] section of the TOML file.
// Each field is a 256-color palette index (0–255). Omit a field to inherit
// the level color. Menu/CLI helpers read this same map for `menu`, `title`,
// `prompt`, `data`, and `divider`. Use StyleColor256(n) in code for the same
// palette.
type colorConfig struct {
	Trace      *int `toml:"trace"`
	Debug      *int `toml:"debug"`
	Info       *int `toml:"info"`
	Warn       *int `toml:"warn"`
	Error      *int `toml:"error"`
	Fatal      *int `toml:"fatal"`
	Panic      *int `toml:"panic"`
	Message    *int `toml:"message"`
	Timestamp  *int `toml:"timestamp"`
	FieldName  *int `toml:"field_name"`
	FieldValue *int `toml:"field_value"`
	Menu       *int `toml:"menu"`
	Title      *int `toml:"title"`
	Prompt     *int `toml:"prompt"`
	Data       *int `toml:"data"`
	Divider    *int `toml:"divider"`
}

// tuiConfig is the [[tui]] section of the TOML file.
type tuiConfig struct {
	MenuSelectedPrefix   string `toml:"menu_selected_prefix"`
	MenuUnselectedPrefix string `toml:"menu_unselected_prefix"`
	MenuIndexWidth       int    `toml:"menu_index_width"`
	InputCursor          string `toml:"input_cursor"`
	DividerWidth         int    `toml:"divider_width"`
}

// color256 converts a nullable palette index to an ANSI escape string.
// A nil pointer means the field was absent in the file; return empty string
// so ConsoleColors falls back to the level color.
func color256(p *int) string {
	if p == nil {
		return ""
	}
	return StyleColor256(*p)
}

// ConfigFromFile parses a TOML file at path and returns a Config.
//
// Fields absent from the file keep zero values; Configure and normalizeConfig
// will fill them with package defaults (stdout writer, InfoLevel, RFC3339 time
// format, DefaultColors, DefaultTUIConfig).
//
// The returned Config is ready to pass directly to Configure:
//
//	cfg, err := logs.ConfigFromFile("logger.toml")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	logs.Configure(cfg)
//
// To attach a writer or hook before applying:
//
//	cfg.Writer = myWriter
//	cfg.ConfigureLogger = func(l logs.Logger) logs.Logger {
//	    return l.With().Str("service", "api").Logger()
//	}
//	logs.Configure(cfg)
func ConfigFromFile(path string) (Config, error) {
	var fc fileConfig
	if _, err := toml.DecodeFile(path, &fc); err != nil {
		return Config{}, fmt.Errorf("smplog: parse config %q: %w", path, err)
	}

	var level Level
	if fc.Level == "" {
		level = InfoLevel
	} else {
		var err error
		level, err = ParseLevel(fc.Level)
		if err != nil {
			return Config{}, fmt.Errorf("smplog: invalid level %q in %q: %w", fc.Level, path, err)
		}
	}

	return Config{
		Level:      level,
		Timestamp:  fc.Timestamp,
		Caller:     fc.Caller,
		Stack:      fc.Stack,
		TimeFormat: fc.TimeFormat,
		NoColor:    fc.NoColor,
		Bypass:     fc.Bypass,
		Files:      fc.Files,
		Colors: ConsoleColors{
			Trace:      color256(fc.Colors.Trace),
			Debug:      color256(fc.Colors.Debug),
			Info:       color256(fc.Colors.Info),
			Warn:       color256(fc.Colors.Warn),
			Error:      color256(fc.Colors.Error),
			Fatal:      color256(fc.Colors.Fatal),
			Panic:      color256(fc.Colors.Panic),
			Message:    color256(fc.Colors.Message),
			Timestamp:  color256(fc.Colors.Timestamp),
			FieldName:  color256(fc.Colors.FieldName),
			FieldValue: color256(fc.Colors.FieldValue),
			Menu:       color256(fc.Colors.Menu),
			Title:      color256(fc.Colors.Title),
			Prompt:     color256(fc.Colors.Prompt),
			Data:       color256(fc.Colors.Data),
			Divider:    color256(fc.Colors.Divider),
		},
		TUI: parseTUIConfig(fc.TUI),
	}, nil
}

func parseTUIConfig(entries []tuiConfig) TUIConfig {
	if len(entries) == 0 {
		return TUIConfig{}
	}
	// If multiple [[tui]] sections are present, the last one wins.
	last := entries[len(entries)-1]
	return TUIConfig{
		MenuSelectedPrefix:   last.MenuSelectedPrefix,
		MenuUnselectedPrefix: last.MenuUnselectedPrefix,
		MenuIndexWidth:       last.MenuIndexWidth,
		InputCursor:          last.InputCursor,
		DividerWidth:         last.DividerWidth,
	}
}
