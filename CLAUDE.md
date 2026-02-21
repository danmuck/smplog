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

### Key Design Patterns

**Package-global singleton logger.** A single `*Logger` (`currentLogger`) and `Config` (`currentConfig`) live as package-level vars, protected by `sync.RWMutex` (`stateMu`). `init()` installs a default logger so the package is usable without explicit setup. All convenience functions (`Info`, `Debug`, `Warn`, etc.) dispatch to this global.

**Dual-mode via writer substitution.** The only difference between JSON and console modes is the `io.Writer` passed to `zerolog.New()`: bypass uses the raw writer; console wraps it in a `zerolog.ConsoleWriter`. See `buildLogger()` in `logger.go`.

**Escape-hatch hooks in `Config`.** Three optional `func` fields allow customization without exposing zerolog internals:
- `ConfigureZerolog func()` — called before building the logger (e.g. set global zerolog options)
- `ConfigureConsole func(w *ConsoleWriter)` — called after console writer creation
- `ConfigureLogger func(l Logger) Logger` — called after logger construction (e.g. inject fields)

**zerolog re-exports in `zerolog_api.go`.** All zerolog types (`Logger`, `Event`, `Context`, etc.) and utility functions are re-exported as package-level aliases so callers never import zerolog directly.

**Legacy compatibility shim.** Integer mode constants (`INACTIVE=0` through `DIAGNOSTICS=5`) and functions (`SetMode`, `Dev`, `Init`, `Print`, `String`, `MsgSuccess`, `MsgFailure`) coexist with the modern zerolog-style API. `modeToLevel()` translates between them. `TRACE` (bool) and `LOGGER_enable_timestamp` are legacy package-level vars that affect `String()` behavior.

**ANSI colors as raw strings.** `ConsoleColors` struct fields hold raw ANSI escape strings. `colorize(color, text, disabled)` in `colors.go` concatenates `color + text + StyleReset`. No external color library.

### Critical Files

| File | Purpose |
|---|---|
| `logger.go` | Core: `Config`, `Configure()`, `buildLogger()`, `applyConsoleFormatting()`, all convenience log functions, legacy shim |
| `colors.go` | `ConsoleColors`, ANSI palette constants, `colorize()`, `ColorText()`, `StyleColor256()`, `StripANSI()`, `CenterTag()`, `FormatPath()` |
| `zerolog_api.go` | Re-exports all zerolog types and utility functions |
| `logger_test.go` | White-box tests (same package `logs`); uses `bytes.Buffer` as writer to capture and assert on output |

### Test Pattern

Tests save global state, configure with a `bytes.Buffer` writer, call a log function, then assert on the captured string. All tests restore global state via `t.Cleanup(func() { Configure(DefaultConfig()) })`.
