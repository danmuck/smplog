package logs

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// Config controls smplog and zerolog behavior.
type Config struct {
	// Writer is the final output destination.
	Writer io.Writer
	// Level is the per-logger threshold. Does not affect the zerolog global level;
	// call SetGlobalLevel explicitly if a process-wide ceiling is needed.
	Level Level
	// Timestamp appends a timestamp field to every log entry.
	Timestamp bool
	// Caller appends a caller field to every log entry.
	Caller bool
	// Stack appends stack traces when Stack() is used on events.
	Stack bool
	// TimeFormat controls timestamp rendering in console mode.
	// For bypass/JSON mode set zerolog.TimeFieldFormat via SetTimeFieldFormat.
	TimeFormat string
	// NoColor disables ANSI color output in console mode.
	NoColor bool
	// Bypass disables the console wrapper and emits raw zerolog JSON.
	Bypass bool
	// Colors controls per-level ANSI colors in console mode.
	Colors ConsoleColors
	// ConfigureZerolog is called before the logger is built.
	// Use it to set process-wide zerolog options (e.g. SetTimeFieldFormat).
	ConfigureZerolog func()
	// ConfigureConsole is called after the ConsoleWriter is created.
	// Use it to override individual formatter functions.
	ConfigureConsole func(w *ConsoleWriter)
	// ConfigureLogger is called after the logger is built.
	// Use it to inject permanent context fields (e.g. service name).
	ConfigureLogger func(l Logger) Logger
}

var (
	// stateMu guards currentConfig and currentLogger.
	stateMu       sync.RWMutex
	currentConfig Config
	currentLogger *Logger
)

func init() {
	Configure(DefaultConfig())
}

// DefaultConfig returns a console-mode config suitable for local development.
func DefaultConfig() Config {
	return Config{
		Writer:     os.Stdout,
		Level:      InfoLevel,
		Timestamp:  true,
		TimeFormat: time.RFC3339,
		NoColor:    false,
		Bypass:     false,
		Colors:     DefaultColors(),
	}
}

// Configure applies cfg and atomically replaces the package-global logger.
func Configure(cfg Config) {
	stateMu.Lock()
	defer stateMu.Unlock()
	currentConfig = normalizeConfig(cfg)
	logger := buildLogger(currentConfig)
	currentLogger = &logger
}

// Configured returns a snapshot of the currently active config.
func Configured() Config {
	stateMu.RLock()
	defer stateMu.RUnlock()
	return currentConfig
}

// SetBypass toggles bypass (JSON) mode without rebuilding the full config.
func SetBypass(enabled bool) {
	cfg := Configured()
	cfg.Bypass = enabled
	Configure(cfg)
}

// SetColors replaces the console color palette.
func SetColors(colors ConsoleColors) {
	cfg := Configured()
	cfg.Colors = colors
	Configure(cfg)
}

// SetLevel updates the per-logger level threshold.
func SetLevel(level Level) {
	cfg := Configured()
	cfg.Level = level
	Configure(cfg)
}

// SetLogger replaces the package-global logger directly, bypassing Configure.
func SetLogger(l Logger) {
	stateMu.Lock()
	defer stateMu.Unlock()
	currentLogger = &l
}

// Zerolog returns the active package-global logger.
func Zerolog() *Logger {
	stateMu.RLock()
	defer stateMu.RUnlock()
	return currentLogger
}

// With returns a context builder on the active logger for adding permanent fields.
func With() Context {
	return Zerolog().With()
}

// AtLevel returns a level-scoped event from the active logger.
func AtLevel(level Level) *Event {
	return Zerolog().WithLevel(zerolog.Level(level))
}

// normalizeConfig replaces zero-value fields with defaults.
func normalizeConfig(cfg Config) Config {
	if cfg.Writer == nil {
		cfg.Writer = os.Stdout
	}
	if cfg.TimeFormat == "" {
		cfg.TimeFormat = time.RFC3339
	}
	if cfg.Colors == (ConsoleColors{}) {
		cfg.Colors = DefaultColors()
	}
	return cfg
}

// buildLogger constructs a zerolog logger from cfg.
// It does not mutate any zerolog package-level globals; all configuration is
// scoped to the returned logger instance. To set process-wide zerolog options
// use cfg.ConfigureZerolog.
func buildLogger(cfg Config) Logger {
	if cfg.ConfigureZerolog != nil {
		cfg.ConfigureZerolog()
	}

	writer := cfg.Writer
	if !cfg.Bypass {
		console := ConsoleWriter{
			Out:        cfg.Writer,
			NoColor:    cfg.NoColor,
			TimeFormat: cfg.TimeFormat,
		}
		applyConsoleFormatting(&console, cfg)
		if cfg.ConfigureConsole != nil {
			cfg.ConfigureConsole(&console)
		}
		writer = console
	}

	logger := zerolog.New(writer).Level(cfg.Level)
	ctx := logger.With()
	if cfg.Timestamp {
		ctx = ctx.Timestamp()
	}
	if cfg.Caller {
		ctx = ctx.Caller()
	}
	if cfg.Stack {
		ctx = ctx.Stack()
	}
	logger = ctx.Logger()

	if cfg.ConfigureLogger != nil {
		logger = cfg.ConfigureLogger(logger)
	}

	return logger
}

// applyConsoleFormatting wires ANSI color transforms onto the ConsoleWriter.
func applyConsoleFormatting(console *ConsoleWriter, cfg Config) {
	console.FormatPrepare = func(evt map[string]any) error {
		level := strings.ToLower(fmt.Sprint(evt[zerolog.LevelFieldName]))
		if raw, ok := evt[zerolog.LevelFieldName]; ok {
			evt[zerolog.LevelFieldName] = colorize(
				cfg.Colors.level(level),
				strings.ToUpper(fmt.Sprint(raw)),
				cfg.NoColor,
			)
		}
		if raw, ok := evt[zerolog.MessageFieldName]; ok {
			msgColor := cfg.Colors.Message
			if msgColor == "" {
				msgColor = cfg.Colors.level(level)
			}
			evt[zerolog.MessageFieldName] = colorize(
				msgColor,
				fmt.Sprint(raw),
				cfg.NoColor,
			)
		}
		return nil
	}
	console.FormatTimestamp = func(i any) string {
		if i == nil {
			return ""
		}
		return colorize(cfg.Colors.Timestamp, fmt.Sprint(i), cfg.NoColor)
	}
	console.FormatLevel = func(i any) string {
		return fmt.Sprint(i)
	}
	console.FormatFieldName = func(i any) string {
		return colorize(cfg.Colors.FieldName, fmt.Sprint(i), cfg.NoColor)
	}
	console.FormatFieldValue = func(i any) string {
		return colorize(cfg.Colors.FieldValue, fmt.Sprint(i), cfg.NoColor)
	}
	console.FormatErrFieldName = console.FormatFieldName
	console.FormatErrFieldValue = console.FormatFieldValue
}

// Trace logs a message at trace level.
func Trace(msg string) { Zerolog().Trace().Msg(msg) }

// Tracef logs a formatted message at trace level.
func Tracef(format string, v ...any) { Zerolog().Trace().Msgf(format, v...) }

// Debug logs a message at debug level.
func Debug(msg string) { Zerolog().Debug().Msg(msg) }

// Debugf logs a formatted message at debug level.
func Debugf(format string, v ...any) { Zerolog().Debug().Msgf(format, v...) }

// Info logs a message at info level.
func Info(msg string) { Zerolog().Info().Msg(msg) }

// Infof logs a formatted message at info level.
func Infof(format string, v ...any) { Zerolog().Info().Msgf(format, v...) }

// Warn logs a message at warn level.
func Warn(msg string) { Zerolog().Warn().Msg(msg) }

// Warnf logs a formatted message at warn level.
func Warnf(format string, v ...any) { Zerolog().Warn().Msgf(format, v...) }

// Error logs a message at error level with a structured error field.
// If err is nil zerolog omits the error field.
func Error(err error, msg string) { Zerolog().Error().Err(err).Msg(msg) }

// Errorf logs a formatted message at error level with a structured error field.
// If err is nil zerolog omits the error field.
func Errorf(err error, format string, v ...any) { Zerolog().Error().Err(err).Msgf(format, v...) }

// Fatal logs a message at fatal level with a structured error field, then exits.
// If err is nil zerolog omits the error field.
func Fatal(err error, msg string) { Zerolog().Fatal().Err(err).Msg(msg) }

// Fatalf logs a formatted message at fatal level with a structured error field, then exits.
// If err is nil zerolog omits the error field.
func Fatalf(err error, format string, v ...any) { Zerolog().Fatal().Err(err).Msgf(format, v...) }
