package logs

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	defaultProjectRoot = "smplog"
	defaultPathWidth   = 32
)

var (
	configMu      sync.RWMutex
	currentConfig = DefaultConfig()
)

// Config defines logger behavior and routing.
type Config struct {
	Mode            Level
	Trace           bool
	EnableTimestamp bool
	ProjectRoot     string
	PathWidth       int
	Services        map[string]string
}

// DefaultConfig returns a safe baseline configuration.
func DefaultConfig() Config {
	return Config{
		Mode:            WARN,
		Trace:           false,
		EnableTimestamp: false,
		ProjectRoot:     defaultProjectRoot,
		PathWidth:       defaultPathWidth,
		Services: map[string]string{
			"api":     "api",
			"users":   "users",
			"metrics": "metrics",
			"auth":    "auth",
		},
	}
}

// Configure replaces the active runtime logger configuration.
func Configure(cfg Config) {
	configMu.Lock()
	defer configMu.Unlock()

	currentConfig = normalizeConfig(cfg)
	logger.reset()
}

// ConfigSnapshot returns a copy of the current logger configuration.
func ConfigSnapshot() Config {
	configMu.RLock()
	defer configMu.RUnlock()
	return cloneConfig(currentConfig)
}

// LoadConfigFile parses TOML from path and applies the configuration.
func LoadConfigFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open logger config: %w", err)
	}
	defer f.Close()

	cfg, err := parseTOMLConfig(f)
	if err != nil {
		return err
	}
	Configure(cfg)
	return nil
}

// MustLoadConfigFile parses TOML and applies config, panicking on error.
func MustLoadConfigFile(path string) {
	if err := LoadConfigFile(path); err != nil {
		panic(err)
	}
}

// ParseLevel converts string names into a logger level.
func ParseLevel(value string) (Level, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "inactive":
		return INACTIVE, nil
	case "error":
		return ERROR, nil
	case "warn", "warning":
		return WARN, nil
	case "info":
		return INFO, nil
	case "debug":
		return DEBUG, nil
	case "diagnostics":
		return DIAGNOSTICS, nil
	default:
		return INACTIVE, fmt.Errorf("invalid logger mode %q", value)
	}
}

// String returns the lowercase text form of a logger level.
func (l Level) String() string {
	switch l {
	case INACTIVE:
		return "inactive"
	case ERROR:
		return "error"
	case WARN:
		return "warn"
	case INFO:
		return "info"
	case DEBUG:
		return "debug"
	case DIAGNOSTICS:
		return "diagnostics"
	default:
		return fmt.Sprintf("level(%d)", int(l))
	}
}

func cloneConfig(cfg Config) Config {
	cloned := cfg
	cloned.Services = make(map[string]string, len(cfg.Services))
	for k, v := range cfg.Services {
		cloned.Services[k] = v
	}
	return cloned
}

func normalizeConfig(cfg Config) Config {
	if cfg.PathWidth <= 0 {
		cfg.PathWidth = defaultPathWidth
	}
	if strings.TrimSpace(cfg.ProjectRoot) == "" {
		cfg.ProjectRoot = defaultProjectRoot
	}
	if cfg.Services == nil {
		cfg.Services = map[string]string{}
	}
	return cloneConfig(cfg)
}

func parseTOMLConfig(r io.Reader) (Config, error) {
	cfg := DefaultConfig()
	scanner := bufio.NewScanner(r)
	section := ""
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := stripComments(scanner.Text())
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.TrimSpace(line[1 : len(line)-1])
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return Config{}, fmt.Errorf("parse logger config line %d: invalid key/value", lineNum)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch section {
		case "":
			if err := setConfigValue(&cfg, key, value); err != nil {
				return Config{}, fmt.Errorf("parse logger config line %d: %w", lineNum, err)
			}
		case "services":
			svcValue, err := parseString(value)
			if err != nil {
				return Config{}, fmt.Errorf("parse logger config line %d: %w", lineNum, err)
			}
			cfg.Services[key] = svcValue
		default:
			return Config{}, fmt.Errorf("parse logger config line %d: unsupported section [%s]", lineNum, section)
		}
	}

	if err := scanner.Err(); err != nil {
		return Config{}, fmt.Errorf("read logger config: %w", err)
	}

	return normalizeConfig(cfg), nil
}

func setConfigValue(cfg *Config, key, value string) error {
	switch key {
	case "mode":
		if unquoted, err := parseString(value); err == nil {
			level, levelErr := ParseLevel(unquoted)
			if levelErr != nil {
				return levelErr
			}
			cfg.Mode = level
			return nil
		}

		i, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("mode must be a string or integer")
		}
		cfg.Mode = Level(i)
		return nil
	case "trace":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("trace must be a boolean")
		}
		cfg.Trace = b
		return nil
	case "enable_timestamp":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("enable_timestamp must be a boolean")
		}
		cfg.EnableTimestamp = b
		return nil
	case "project_root":
		s, err := parseString(value)
		if err != nil {
			return fmt.Errorf("project_root must be a string")
		}
		cfg.ProjectRoot = s
		return nil
	case "path_width":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("path_width must be an integer")
		}
		cfg.PathWidth = n
		return nil
	default:
		return fmt.Errorf("unsupported key %q", key)
	}
}

func parseString(value string) (string, error) {
	if len(value) < 2 {
		return "", fmt.Errorf("invalid string")
	}
	if value[0] != '"' || value[len(value)-1] != '"' {
		return "", fmt.Errorf("value must be quoted")
	}
	unquoted, err := strconv.Unquote(value)
	if err != nil {
		return "", fmt.Errorf("invalid quoted string")
	}
	return unquoted, nil
}

func stripComments(s string) string {
	inQuotes := false
	escaped := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if ch == '"' {
			inQuotes = !inQuotes
			continue
		}
		if ch == '#' && !inQuotes {
			return s[:i]
		}
	}
	return s
}
