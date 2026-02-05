package logs

import (
	"strings"
	"testing"
)

func TestParseTOMLConfig(t *testing.T) {
	in := `
mode = "debug"
trace = true
enable_timestamp = true
project_root = "billing"
path_width = 40

[services]
api = "api"
payments = "payments"
`

	cfg, err := parseTOMLConfig(strings.NewReader(in))
	if err != nil {
		t.Fatalf("parseTOMLConfig returned error: %v", err)
	}

	if cfg.Mode != DEBUG {
		t.Fatalf("mode: want %v got %v", DEBUG, cfg.Mode)
	}
	if !cfg.Trace {
		t.Fatalf("trace: want true got false")
	}
	if !cfg.EnableTimestamp {
		t.Fatalf("enable_timestamp: want true got false")
	}
	if cfg.ProjectRoot != "billing" {
		t.Fatalf("project_root: want billing got %q", cfg.ProjectRoot)
	}
	if cfg.PathWidth != 40 {
		t.Fatalf("path_width: want 40 got %d", cfg.PathWidth)
	}
	if cfg.Services["payments"] != "payments" {
		t.Fatalf("service override not parsed")
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		in   string
		want Level
	}{
		{in: "inactive", want: INACTIVE},
		{in: "error", want: ERROR},
		{in: "warn", want: WARN},
		{in: "info", want: INFO},
		{in: "debug", want: DEBUG},
		{in: "diagnostics", want: DIAGNOSTICS},
	}

	for _, tt := range tests {
		got, err := ParseLevel(tt.in)
		if err != nil {
			t.Fatalf("ParseLevel(%q) returned error: %v", tt.in, err)
		}
		if got != tt.want {
			t.Fatalf("ParseLevel(%q): want %v got %v", tt.in, tt.want, got)
		}
	}
}
