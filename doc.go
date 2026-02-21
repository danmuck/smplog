// Package logs provides a thin zerolog wrapper with:
//   - Structured JSON output in bypass mode (Config.Bypass = true).
//   - Human-readable, colorized console output in wrapper mode (default).
//   - Menu/CLI stdout print helpers (Menu, Title, Prompt, Data, Divider)
//     that reuse the same Config.Colors and NoColor settings.
//   - A package-global singleton logger configured via Configure and Config.
//
// Quick start:
//
//	logs.Info("server started")
//	logs.Error(err, "handler failed")
//
// For production JSON output:
//
//	logs.Configure(logs.Config{Bypass: true, Level: logs.InfoLevel})
//
// To inject permanent context fields:
//
//	logs.Configure(logs.Config{
//	    ConfigureLogger: func(l logs.Logger) logs.Logger {
//	        return l.With().Str("service", "api").Logger()
//	    },
//	})
package logs
