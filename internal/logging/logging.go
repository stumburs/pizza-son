// Package logging provides a small wrapper around the standard logger so
// routine diagnostics can be hidden unless LOG_LEVEL=debug is set.
package logging

import (
	"log"
	"os"
	"strings"
)

var debugEnabled = func() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("LOG_LEVEL"))) {
	case "debug", "trace":
		return true
	default:
		return false
	}
}()

// DebugEnabled reports whether debug logging is turned on.
func DebugEnabled() bool {
	return debugEnabled
}

// Debugf logs only when LOG_LEVEL is debug or trace.
func Debugf(format string, args ...any) {
	if debugEnabled {
		log.Printf(format, args...)
	}
}

// Debugln logs only when LOG_LEVEL is debug or trace.
func Debugln(args ...any) {
	if debugEnabled {
		log.Println(args...)
	}
}
