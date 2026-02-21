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

- `printf.go`: compact formatted output (`Menu`, `Title`, `Prompt`, `Data`, `Divider`)
- `tui_engine.go`: ANSI terminal control + component helpers (`MoveTo`, `WriteAt`, `MenuItem`, `Field`, `BeginFrame`/`EndFrame`)

### Key Design Patterns

**Package-global singleton logger.** A single `*Logger` (`currentLogger`) and `Config` (`currentConfig`) live as package-level vars, protected by `sync.RWMutex` (`stateMu`). `init()` installs a default logger so the package is usable without explicit setup. All convenience functions (`Info`, `Debug`, `Warn`, etc.) dispatch to this global.

**Dual-mode via writer substitution.** The only difference between JSON and console modes is the `io.Writer` passed to `zerolog.New()`: bypass uses the raw writer; console wraps it in a `zerolog.ConsoleWriter`. See `buildLogger()` in `logger.go`.

**Escape-hatch hooks in `Config`.** Three optional `func` fields allow customization without exposing zerolog internals:
- `ConfigureZerolog func()` — called before building the logger (e.g. set global zerolog options)
- `ConfigureConsole func(w *ConsoleWriter)` — called after console writer creation
- `ConfigureLogger func(l Logger) Logger` — called after logger construction (e.g. inject fields)

**zerolog re-exports in `zerolog_api.go`.** All zerolog types (`Logger`, `Event`, `Context`, etc.) and utility functions are re-exported as package-level aliases so callers never import zerolog directly.

**Stdout-first wrappers.** `printf.go` and `tui_engine.go` compose existing config/color state (`Configured().NoColor`, `Configured().Colors`) so menu/TUI rendering follows the same color policy as console logging while remaining independent of zerolog events.

**ANSI colors as raw strings.** `ConsoleColors` struct fields hold raw ANSI escape strings. `colorize(color, text, disabled)` in `colors.go` concatenates `color + text + StyleReset`. No external color library.

### Critical Files

| File | Purpose |
|---|---|
| `logger.go` | Core: `Config`, `Configure()`, `buildLogger()`, `applyConsoleFormatting()`, all convenience log functions, legacy shim |
| `colors.go` | `ConsoleColors`, ANSI palette constants, `colorize()`, `StyleColor256()`, `StripANSI()` |
| `printf.go` | Stdout wrappers for menu-style colored output (no zerolog event required) |
| `tui_engine.go` | Compact terminal control/layout/component helpers for component-style TUIs |
| `zerolog_api.go` | Re-exports all zerolog types and utility functions |
| `logger_test.go` | White-box tests for logging behavior |
| `printf_test.go` | Tests for stdout wrapper color/no-color behavior |
| `tui_engine_test.go` | Tests for ANSI control, layout helpers, and component wrappers |

### Test Pattern

Tests save global state, configure with a `bytes.Buffer` writer, call a log function, then assert on the captured string. All tests restore global state via `t.Cleanup(func() { Configure(DefaultConfig()) })`.
