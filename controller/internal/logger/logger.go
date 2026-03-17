package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

var Log zerolog.Logger

// Init sets up the global logger. Call once at the top of main().
func Init() {
	zerolog.TimeFieldFormat = time.RFC3339
	Log = zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("service", "controller").
		Logger()
}

// AppLogger returns a sub-logger pre-populated with the app name.
// Use this in handlers and reconcile to avoid repeating Str("app", ...) everywhere.
func AppLogger(appName string) zerolog.Logger {
	return Log.With().Str("app", appName).Logger()
}
