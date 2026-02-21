# smplog

`smplog` is a thin wrapper around [`zerolog`](https://github.com/rs/zerolog) that supports:

- structured JSON logs for machine ingestion
- colored terminal logs for local development
- lightweight stdout formatting helpers for CLI/menu UIs (`Menu`, `Title`, `Prompt`, `Data`, `Divider`)
- compatibility helpers for older call sites (`SetMode`, `Logf`, `Print`, etc.)

## Install

```bash
go get github.com/danmuck/smplog
```

## Quick start

```go
package main

import logs "github.com/danmuck/smplog"

func main() {
	// Human-readable console output with colors.
	logs.Configure(logs.Config{
		Level:     logs.InfoLevel,
		Timestamp: true,
		Bypass:    false,
		NoColor:   false,
	})
	logs.Info("console message")

	// Structured JSON output (recommended for production collectors).
	logs.SetBypass(true)
	logs.Info("json message")
}
```

## Configuration notes

- `Bypass=true`: writes raw structured JSON from zerolog.
- `Bypass=false`: writes formatted console logs.
- `NoColor=true`: disables ANSI colors when console formatting is enabled.
- `SetMode(...)`: maps legacy mode constants (`INACTIVE`, `ERROR`, `INFO`, `WARN`, `DEBUG`, `DIAGNOSTICS`) to zerolog levels.

## Menu/CLI print helpers

The package also includes stdout wrappers that reuse the same `Config.Colors` and `NoColor` settings, without using `zerolog` events:

```go
logs.Configure(logs.Config{
	NoColor: false,
	Colors:  logs.DefaultColors(),
})

logs.Title("Ghost Control Plane\n")
logs.Divider(48)
logs.Menu("1) Inventory\n")
logs.Menu("2) Start service\n")
logs.Prompt("Select option > ")
```

## Compact TUI engine helpers

For component-style TUIs, `tui_engine.go` adds ANSI control and positional rendering helpers that compose the same color config:

```go
_ = logs.BeginFrame()
defer logs.EndFrame()

logs.WriteAt(1, 1, logs.Configured().Colors.title(), "Ghost\n")
logs.WriteAt(3, 2, logs.Configured().Colors.menu(), logs.PadRight(24, "1) Inventory"))
logs.WriteAt(4, 2, logs.Configured().Colors.menu(), logs.PadRight(24, "2) Services"))
logs.WriteAt(6, 2, logs.Configured().Colors.prompt(), "Select > ")
```

TUI defaults can be set in TOML with `[[tui]]`:

```toml
[[tui]]
menu_selected_prefix   = ">"
menu_unselected_prefix = " "
menu_index_width       = 2
input_cursor           = "_"
divider_width          = 64
```
