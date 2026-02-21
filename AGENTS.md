# AGENTS.md

Guidance for coding agents working in this repository.

## Project overview

`smplog` is a small Go logging package (`package logs`) that wraps `zerolog` with:

- console mode: human-readable logs via `zerolog.ConsoleWriter`
- bypass mode: raw structured JSON for log collectors
- config-file loading (`smplog.config.toml`) plus programmatic hooks
- optional named file sinks (`WriteFile` + `Config.Files`)

This package is intentionally thin. Prefer explicit behavior over abstraction.

## API stability policy (v1)

This repository is considered **v1**.

- Do not introduce breaking changes to exported APIs.
- Do not introduce breaking changes to default runtime behavior.
- Evolve additively: new functions/options are acceptable when existing call sites keep working unchanged.
- If a true breaking change is unavoidable, require explicit maintainer approval and a major-version path (`/v2`) instead of changing v1 behavior in place.

## Commands

```bash
go build ./...
go test ./...
go test -v ./...
go test -run TestName ./...
go test -race ./...
go vet ./...
go mod tidy
```

## Repo map

- `logger.go`: core config/state management, logger construction, top-level log functions, file sink lifecycle
- `config.go`: TOML decoding (`ConfigFromFile`) into runtime `Config`
- `colors.go`: ANSI palette/types and formatting helpers
- `printf.go`: stdout-first formatting wrappers for menu/CLI output (`Menu`, `Title`, `Prompt`, `Data`, `Divider`)
- `zerolog_api.go`: re-exports of zerolog types/helpers
- `logger_test.go`: behavior tests for output modes, hooks, and file routing
- `config_test.go`: TOML parsing tests
- `printf_test.go`: behavior tests for stdout formatting wrappers and color/no-color behavior
- `smplog.config.toml`: example/default config file used at init
- `doc.go`, `README.md`: package-facing docs

## Architecture and invariants

1. Package-global runtime state:
- `currentConfig` and `currentLogger` are guarded by `stateMu`.
- `openFiles` is guarded by `filesMu`.

2. Startup behavior:
- `init()` attempts `ConfigFromFile("smplog.config.toml")`.
- On parse/read failure, falls back to `DefaultConfig()`.
- `Configure(cfg)` is always called once during init.

3. Configure semantics:
- `Configure` closes/replaces file handles via `applyFiles`.
- `normalizeConfig` fills zero-value config fields.
- `buildLogger` is the single constructor path for logger behavior.

4. Output mode contract:
- `Bypass=true` must emit raw JSON (no console formatting, no ANSI).
- `Bypass=false` must use `ConsoleWriter` and apply configured formatting/colors.

5. Hooks contract:
- `ConfigureZerolog` runs before logger creation.
- `ConfigureConsole` runs after console writer creation.
- `ConfigureLogger` runs after logger/context construction.

6. File sink contract:
- `WriteFile(fn, name)` is a no-op for unknown names.
- File sink entries always log JSON with timestamps.
- `Close()` closes all open file sinks and returns joined errors.

## Testing expectations

1. Always run `go test ./...` after behavior changes.
2. Add/adjust tests in the same package (`package logs`) for new behavior.
3. Preserve global state in tests with cleanup (`Configure(DefaultConfig())` or saved config restoration).
4. Avoid `t.Parallel()` in tests mutating package-global logger state.
5. For output assertions:
- console mode: assert human-readable output and ANSI behavior as needed
- bypass mode: assert valid JSON and structured fields

## Editing guidance

1. Keep functions focused and explicit; avoid hidden side effects.
2. Do not bypass mutex-protected state paths when changing globals.
3. Keep mode behavior localized to writer selection in `buildLogger`.
4. Preserve v1 compatibility: no breaking public API or behavior changes.
5. If introducing/removing APIs, sync `README.md`, `doc.go`, and this file.

## Documentation note

`README.md` and `CLAUDE.md` may reference legacy helper APIs that are not present in current code. Prefer the implementation in `logger.go`/`zerolog_api.go` as source of truth, and update stale docs when touching related behavior.
