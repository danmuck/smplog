package logs

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestConfigureAppliesConsoleColors verifies configured level/message colors are emitted.
func TestConfigureAppliesConsoleColors(t *testing.T) {
	var out bytes.Buffer
	infoColor := StyleColor256(1)
	messageColor := StyleColor256(2)
	infoANSI := "\x1b[38;5;1m"
	messageANSI := "\x1b[38;5;2m"

	Configure(Config{
		Writer:    &out,
		Level:     InfoLevel,
		Timestamp: false,
		Bypass:    false,
		NoColor:   false,
		Colors: ConsoleColors{
			Info:    infoColor,
			Message: messageColor,
		},
	})
	t.Cleanup(func() { Configure(DefaultConfig()) })

	Info("color-check")

	logLine := out.String()
	if !strings.Contains(logLine, infoANSI) {
		t.Fatalf("expected info color %q in output: %q", infoANSI, logLine)
	}
	if !strings.Contains(logLine, messageANSI) {
		t.Fatalf("expected message color %q in output: %q", messageANSI, logLine)
	}
}

// TestBypassModeEmitsRawJSON verifies bypass mode outputs plain JSON without ANSI escapes.
func TestBypassModeEmitsRawJSON(t *testing.T) {
	var out bytes.Buffer

	Configure(Config{
		Writer:    &out,
		Level:     InfoLevel,
		Timestamp: false,
		Bypass:    true,
	})
	t.Cleanup(func() { Configure(DefaultConfig()) })

	Info("raw-json")

	logLine := strings.TrimSpace(out.String())
	if strings.Contains(logLine, "\x1b[") {
		t.Fatalf("expected no ANSI colors in bypass output: %q", logLine)
	}
	if !json.Valid([]byte(logLine)) {
		t.Fatalf("expected valid JSON in bypass output: %q", logLine)
	}
	if !strings.Contains(logLine, `"message":"raw-json"`) {
		t.Fatalf("expected message field in bypass output: %q", logLine)
	}
}

// TestConfigureLoggerHookIsApplied verifies ConfigureLogger can inject additional fields.
func TestConfigureLoggerHookIsApplied(t *testing.T) {
	var out bytes.Buffer

	Configure(Config{
		Writer:    &out,
		Level:     InfoLevel,
		Timestamp: false,
		Bypass:    true,
		ConfigureLogger: func(l Logger) Logger {
			return l.With().Str("service", "edge").Logger()
		},
	})
	t.Cleanup(func() { Configure(DefaultConfig()) })

	Info("hook-check")

	logLine := out.String()
	if !strings.Contains(logLine, `"service":"edge"`) {
		t.Fatalf("expected custom logger field in output: %q", logLine)
	}
}

// TestNoColorSuppressesANSIInConsoleMode verifies NoColor strips ANSI output in console mode.
func TestNoColorSuppressesANSIInConsoleMode(t *testing.T) {
	var out bytes.Buffer

	Configure(Config{
		Writer:    &out,
		Level:     InfoLevel,
		Timestamp: false,
		Bypass:    false,
		NoColor:   true,
		Colors: ConsoleColors{
			Info:    StyleColor256(1),
			Message: StyleColor256(2),
		},
	})
	t.Cleanup(func() { Configure(DefaultConfig()) })

	Info("plain-console")

	if strings.Contains(out.String(), "\x1b[") {
		t.Fatalf("expected no ANSI when NoColor=true, got %q", out.String())
	}
}

// TestErrorAttachesStructuredErrorField verifies Error emits a structured "error" field in JSON.
func TestErrorAttachesStructuredErrorField(t *testing.T) {
	var out bytes.Buffer

	Configure(Config{
		Writer:    &out,
		Level:     ErrorLevel,
		Timestamp: false,
		Bypass:    true,
	})
	t.Cleanup(func() { Configure(DefaultConfig()) })

	err := errors.New("connection refused")
	Error(err, "database unavailable")

	logLine := strings.TrimSpace(out.String())
	if !json.Valid([]byte(logLine)) {
		t.Fatalf("expected valid JSON, got %q", logLine)
	}
	if !strings.Contains(logLine, `"error":"connection refused"`) {
		t.Fatalf("expected structured error field in output: %q", logLine)
	}
	if !strings.Contains(logLine, `"message":"database unavailable"`) {
		t.Fatalf("expected message field in output: %q", logLine)
	}
}

// TestWriteFileRoutesToNamedFile verifies WriteFile writes JSON to the correct named file.
func TestWriteFileRoutesToNamedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.log")

	originalCfg := Configured()
	t.Cleanup(func() {
		Close()
		Configure(originalCfg)
	})

	Configure(Config{
		Files: []LogFile{{Name: "test", Path: path}},
	})

	WriteFile(At(InfoLevel, "hello-file"), "test")
	WriteFile(Atf(ErrorLevel, "code %d", 42), "test")

	if err := Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 log lines, got %d: %q", len(lines), string(data))
	}
	if !json.Valid([]byte(lines[0])) {
		t.Errorf("line 0 not valid JSON: %q", lines[0])
	}
	if !strings.Contains(lines[0], `"message":"hello-file"`) {
		t.Errorf("expected hello-file in line 0: %q", lines[0])
	}
	if !strings.Contains(lines[1], `"message":"code 42"`) {
		t.Errorf("expected code 42 in line 1: %q", lines[1])
	}
}

// TestWriteFileUnknownNameIsNoop verifies WriteFile with an unknown name does nothing.
func TestWriteFileUnknownNameIsNoop(t *testing.T) {
	WriteFile(At(InfoLevel, "should not panic"), "nonexistent")
}

// TestErrorWithNilErrOmitsErrorField verifies Error with nil err produces no "error" key.
func TestErrorWithNilErrOmitsErrorField(t *testing.T) {
	var out bytes.Buffer

	Configure(Config{
		Writer:    &out,
		Level:     ErrorLevel,
		Timestamp: false,
		Bypass:    true,
	})
	t.Cleanup(func() { Configure(DefaultConfig()) })

	Error(nil, "non-fatal condition")

	logLine := strings.TrimSpace(out.String())
	if strings.Contains(logLine, `"error":`) {
		t.Fatalf("expected no error field for nil err, got %q", logLine)
	}
	if !strings.Contains(logLine, `"message":"non-fatal condition"`) {
		t.Fatalf("expected message field in output: %q", logLine)
	}
}
