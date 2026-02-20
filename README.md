# smplog

`smplog` is a thin wrapper around [`zerolog`](https://github.com/rs/zerolog) that supports:

- structured JSON logs for machine ingestion
- colored terminal logs for local development
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

