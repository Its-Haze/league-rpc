package lcu

import (
	"fmt"

	"github.com/rs/zerolog"
)

// gopherLogger forwards lcu-gopher's own logging into the app's logger, so a
// connection that never comes up explains itself in the copied diagnostics.
type gopherLogger struct {
	logger zerolog.Logger
}

// Info is unused by lcu-gopher today, and is routine when it is used, so it
// lands at debug level rather than competing with the app's own lifecycle log.
func (g gopherLogger) Info(endpoint, msg string, args ...interface{}) {
	g.logger.Debug().Str("lcu", endpoint).Msg(formatGopher(msg, args...))
}

func (g gopherLogger) Error(endpoint, msg string, args ...interface{}) {
	g.logger.Warn().Str("lcu", endpoint).Msg(formatGopher(msg, args...))
}

func (g gopherLogger) Debug(endpoint, msg string, args ...interface{}) {
	g.logger.Debug().Str("lcu", endpoint).Msg(formatGopher(msg, args...))
}

// formatGopher applies lcu-gopher's printf-style arguments, leaving a message
// with no arguments untouched so stray verbs in it survive.
func formatGopher(msg string, args ...interface{}) string {
	if len(args) == 0 {
		return msg
	}
	return fmt.Sprintf(msg, args...)
}
