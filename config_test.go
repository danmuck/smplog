package logs

import (
	"os"
	"path/filepath"
	"testing"
)

// writeTOML writes content to a temp file and returns its path.
func writeTOML(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "*.toml")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

// TestConfigFromFileBasicFields verifies level, flags, and time_format are parsed.
func TestConfigFromFileBasicFields(t *testing.T) {
	path := writeTOML(t, `
level       = "debug"
timestamp   = true
caller      = true
stack       = false
time_format = "15:04:05"
no_color    = true
bypass      = true
`)

	cfg, err := ConfigFromFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Level != DebugLevel {
		t.Errorf("level: got %v, want %v", cfg.Level, DebugLevel)
	}
	if !cfg.Timestamp {
		t.Error("timestamp: expected true")
	}
	if !cfg.Caller {
		t.Error("caller: expected true")
	}
	if cfg.Stack {
		t.Error("stack: expected false")
	}
	if cfg.TimeFormat != "15:04:05" {
		t.Errorf("time_format: got %q, want %q", cfg.TimeFormat, "15:04:05")
	}
	if !cfg.NoColor {
		t.Error("no_color: expected true")
	}
	if !cfg.Bypass {
		t.Error("bypass: expected true")
	}
}

// TestConfigFromFileColors verifies the [colors] section converts palette
// indexes to ANSI escape sequences via StyleColor256.
func TestConfigFromFileColors(t *testing.T) {
	path := writeTOML(t, `
[colors]
info        = 4
error       = 1
field_name  = 6
field_value = 7
`)

	cfg, err := ConfigFromFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Colors.Info != StyleColor256(4) {
		t.Errorf("colors.info: got %q, want %q", cfg.Colors.Info, StyleColor256(4))
	}
	if cfg.Colors.Error != StyleColor256(1) {
		t.Errorf("colors.error: got %q, want %q", cfg.Colors.Error, StyleColor256(1))
	}
	if cfg.Colors.FieldName != StyleColor256(6) {
		t.Errorf("colors.field_name: got %q, want %q", cfg.Colors.FieldName, StyleColor256(6))
	}
	if cfg.Colors.FieldValue != StyleColor256(7) {
		t.Errorf("colors.field_value: got %q, want %q", cfg.Colors.FieldValue, StyleColor256(7))
	}
}

// TestConfigFromFileOmittedColorIsEmpty verifies omitted color fields produce
// empty strings so ConsoleColors falls back to the level color.
func TestConfigFromFileOmittedColorIsEmpty(t *testing.T) {
	path := writeTOML(t, ``)

	cfg, err := ConfigFromFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Colors.Info != "" {
		t.Errorf("expected empty info color for omitted field, got %q", cfg.Colors.Info)
	}
}

// TestConfigFromFileDefaultsOnEmptyFile verifies an empty file returns InfoLevel.
func TestConfigFromFileDefaultsOnEmptyFile(t *testing.T) {
	path := writeTOML(t, "")

	cfg, err := ConfigFromFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Level != InfoLevel {
		t.Errorf("level: got %v, want InfoLevel", cfg.Level)
	}
}

// TestConfigFromFileInvalidLevel verifies an unrecognised level returns an error.
func TestConfigFromFileInvalidLevel(t *testing.T) {
	path := writeTOML(t, `level = "verbose"`)

	_, err := ConfigFromFile(path)
	if err == nil {
		t.Fatal("expected error for invalid level, got nil")
	}
}

// TestConfigFromFileMissingFile verifies a missing path returns an error.
func TestConfigFromFileMissingFile(t *testing.T) {
	_, err := ConfigFromFile(filepath.Join(t.TempDir(), "nonexistent.toml"))
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

// TestConfigFromFileLogFiles verifies [[files]] entries are parsed into Config.Files.
func TestConfigFromFileLogFiles(t *testing.T) {
	path := writeTOML(t, `
[[files]]
name = "dev"
path = "logs/dev.log"

[[files]]
name = "errors"
path = "logs/errors.log"
`)

	cfg, err := ConfigFromFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(cfg.Files))
	}
	if cfg.Files[0].Name != "dev" || cfg.Files[0].Path != "logs/dev.log" {
		t.Errorf("files[0]: got %+v", cfg.Files[0])
	}
	if cfg.Files[1].Name != "errors" || cfg.Files[1].Path != "logs/errors.log" {
		t.Errorf("files[1]: got %+v", cfg.Files[1])
	}
}
