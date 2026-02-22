# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
go build ./...          # compile
go test ./...           # run all tests
go test -v ./...        # verbose test output
go test -run TestName ./...  # run a single test
go vet ./...            # static analysis
go mod tidy             # tidy dependencies
```

No Makefile or custom build scripts exist — the standard Go toolchain is the only build system.

## API stability policy (v1)

Treat this module as **v1**.

- No breaking changes to exported API surface.
- No breaking changes to default behavior expected by existing integrations.
- Prefer additive evolution (new APIs/fields/options) over mutation of existing contracts.
- If a breaking change is truly required, do not land it in-place on v1; require explicit maintainer approval and a major version path (`/v2`).

## Architecture

`smplog` is a thin zerolog wrapper (`package logs`, imported as `import logs "github.com/danmuck/smplog"`). It provides two output modes toggled via `Config.Bypass`:

- **Console mode** (default): colorized, human-readable output via `zerolog.ConsoleWriter`
- **Bypass mode**: raw structured JSON from zerolog for production log collectors

It also exposes stdout-first helpers that do not require zerolog events:

- `printf.go`: compact formatted output (`Print`, `Printf`, `Println`, `Colorf`)

TUI components (ANSI terminal control, menu/input/divider renderers) have been extracted to the standalone `github.com/danmuck/tui_go` module.

### Key Design Patterns

**Package-global singleton logger.** A single `*Logger` (`currentLogger`) and `Config` (`currentConfig`) live as package-level vars, protected by `sync.RWMutex` (`stateMu`). `init()` installs a default logger so the package is usable without explicit setup. All convenience functions (`Info`, `Debug`, `Warn`, etc.) dispatch to this global.

**Dual-mode via writer substitution.** The only difference between JSON and console modes is the `io.Writer` passed to `zerolog.New()`: bypass uses the raw writer; console wraps it in a `zerolog.ConsoleWriter`. See `buildLogger()` in `logger.go`.

**Escape-hatch hooks in `Config`.** Three optional `func` fields allow customization without exposing zerolog internals:

- `ConfigureZerolog func()` — called before building the logger (e.g. set global zerolog options)
- `ConfigureConsole func(w *ConsoleWriter)` — called after console writer creation
- `ConfigureLogger func(l Logger) Logger` — called after logger construction (e.g. inject fields)

**zerolog re-exports in `zerolog_api.go`.** All zerolog types (`Logger`, `Event`, `Context`, etc.) and utility functions are re-exported as package-level aliases so callers never import zerolog directly.

**Stdout-first wrappers.** `printf.go` composes existing config/color state (`Configured().NoColor`, `Configured().Colors`) so output follows the same color policy as console logging while remaining independent of zerolog events.

**ANSI colors as raw strings.** `ConsoleColors` struct fields hold raw ANSI escape strings. `colorize(color, text, disabled)` in `colors.go` concatenates `color + text + StyleReset`. No external color library.

### Critical Files

| File             | Purpose                                                                                                                |
| ---------------- | ---------------------------------------------------------------------------------------------------------------------- |
| `logger.go`      | Core: `Config`, `Configure()`, `buildLogger()`, `applyConsoleFormatting()`, all convenience log functions, legacy shim |
| `colors.go`      | `ConsoleColors`, ANSI palette constants, `colorize()`, `Colorize()`, `StyleColor256()`, `StripANSI()`                  |
| `printf.go`      | Stdout wrappers for colored output (no zerolog event required)                                                         |
| `zerolog_api.go` | Re-exports all zerolog types and utility functions                                                                     |
| `logger_test.go` | White-box tests for logging behavior                                                                                   |
| `printf_test.go` | Tests for stdout wrapper color/no-color behavior                                                                       |

### Test Pattern

Tests save global state, configure with a `bytes.Buffer` writer, call a log function, then assert on the captured string. All tests restore global state via `t.Cleanup(func() { Configure(DefaultConfig()) })`.

## Git

- Do not commit changes unless explicitly instructed to do so.
- Include feature size breakpoints in task lists to ask if I would like to commit the changes, giving me time to look them over before I commit them.
- By default, I make all commits

## Context Management

- Use `/compact` frequently during long sessions to reduce token surface and keep context focused.
- Use `/clear` when starting a new logical task or when prior context is no longer relevant.
- All agents (including subagents) should `/compact` after completing major steps.
