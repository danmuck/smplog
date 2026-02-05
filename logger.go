package logs

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	INACTIVE    Level = iota // no logging enabled
	ERROR                    // errors only
	WARN                     // warnings and errors
	INFO                     // informational, warning, and errors
	DEBUG                    // include debug info
	DIAGNOSTICS              // all logging enabled
)

// Level represents logger verbosity.
type Level int

// Log is message-only (safe)
func Log(msg string) {
	cfg := ConfigSnapshot()
	log := &ServiceLog{
		ts:  time.Now(),
		msg: msg,
	}
	for key, value := range cfg.Services {
		if strings.Contains(msg, value) {
			logger.set(key, log)
		}
	}
	Print(StyleWhite, "[logs]", "%s", msg)
}

// Logf writes a formatted application log message.
func Logf(format string, v ...any) {
	Log(fmt.Sprintf(format, v...))
}

// Fatal prints a fatal message and exits with status code 1.
func Fatal(msg string) {
	Print(StyleColor256(196), "[fatal]", "%s", msg)
	os.Exit(1)
}

// Fatalf prints a formatted fatal message and exits with status code 1.
func Fatalf(format string, v ...any) { Fatal(fmt.Sprintf(format, v...)) }

// Err prints an error message when current mode allows it.
func Err(msg string) {
	if !enabled(ERROR) {
		return
	}
	Print(StyleRed, "[error]", "%s", msg)
}

// Errf prints a formatted error message when current mode allows it.
func Errf(format string, v ...any) { Err(fmt.Sprintf(format, v...)) }

// Warn prints a warning message when current mode allows it.
func Warn(msg string) {
	if !enabled(WARN) {
		return
	}
	Print(StyleYellow, "[warn]", "%s", msg)
}

// Warnf prints a formatted warning message when current mode allows it.
func Warnf(format string, v ...any) { Warn(fmt.Sprintf(format, v...)) }

// Info prints an informational message when current mode allows it.
func Info(msg string) {
	if !enabled(INFO) {
		return
	}
	Print(StyleBlue, "[info]", "%s", msg)
}

// Infof prints a formatted informational message when current mode allows it.
func Infof(format string, v ...any) { Info(fmt.Sprintf(format, v...)) }

// Debug prints a debug message when current mode allows it.
func Debug(msg string) {
	if !enabled(DEBUG) {
		return
	}
	Print(StyleGreen, "[debug]", "%s", msg)
}

// Debugf prints a formatted debug message when current mode allows it.
func Debugf(format string, v ...any) { Debug(fmt.Sprintf(format, v...)) }

// Dev prints a development-only message without level filtering.
func Dev(msg string) {
	Print(StyleColor256(89), "[dev_]", "%s", msg)
}

// Devf prints a formatted development-only message without level filtering.
func Devf(format string, v ...any) { Dev(fmt.Sprintf(format, v...)) }

// Init prints an initialization message for diagnostics mode.
func Init(msg string) {
	if !enabled(DIAGNOSTICS) {
		return
	}
	Print(StyleBlack, "[init]", "%s", msg)
}

// Initf prints a formatted initialization message for diagnostics mode.
func Initf(format string, v ...any) { Init(fmt.Sprintf(format, v...)) }

// MsgSuccess prints a message in success style.
func MsgSuccess(msg string) { Print(StyleColor256(66), "[MDS]", "%s", msg) }

// MsgSuccessf prints a formatted message in success style.
func MsgSuccessf(format string, v ...any) { MsgSuccess(fmt.Sprintf(format, v...)) }

// MsgFailure prints a message in failure style.
func MsgFailure(msg string) { Print(StyleColor256(130), "[MDF]", "%s", msg) }

// MsgFailuref prints a formatted message in failure style.
func MsgFailuref(format string, v ...any) { MsgFailure(fmt.Sprintf(format, v...)) }

// Print renders a colored log line.
func Print(C, T, format string, v ...any) {
	fmt.Println(ColorText(C, String(T, format, v...)))
}

// String builds a single rendered log line.
func String(T, format string, v ...any) string {
	cfg := ConfigSnapshot()
	msg := fmt.Sprintf(format, v...)
	ts := time.Now().Format(time.Stamp)
	if !cfg.Trace {
		if cfg.EnableTimestamp {
			return fmt.Sprintf("%s %s", ts, msg)
		}
		return msg
	}
	_, file, line, ok := runtime.Caller(3)
	if ok {
		path := file
		if cfg.ProjectRoot != "" {
			path = TrimToProjectRootWidth(cfg.ProjectRoot, file, cfg.PathWidth)
		}
		tag := CenterTag(T, 9)                                 // padded 9 and centered tag
		lineStr := fmt.Sprintf(":%4d", line)                   // pad line number
		prefix := fmt.Sprintf("%s[%s%s] ", tag, path, lineStr) // final prefix
		if cfg.EnableTimestamp {
			return fmt.Sprintf("%s %s%s", ts, prefix, msg)
		}
		return fmt.Sprintf("%s%s", prefix, msg)
	}
	return msg
}

func enabled(level Level) bool {
	cfg := ConfigSnapshot()
	return cfg.Mode >= level
}
