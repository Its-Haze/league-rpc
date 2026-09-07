// Package logging builds the GUI process's logger: a zerolog logger fanned
package logging

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/its-haze/league-rpc/internal/config"
	"github.com/rs/zerolog"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// LogFileName is the active log file inside the logs directory.
const LogFileName = "league-rpc.log"

// Options tunes the logger. The zero value is usable: it produces an
// info-level logger with default sizes.
type Options struct {
	Debug      bool // follows config Advanced.DebugMode
	RingSize   int  // recent lines kept in memory; 0 -> DefaultRingSize
	MaxSizeMB  int  // rotate the file past this size; 0 -> 5
	MaxBackups int  // rotated files retained; 0 -> 2
}

// Sink is the assembled logging pipeline. Hold on to Ring to read recent
// lines and subscribe to new ones; call Close on shutdown to flush the file.
type Sink struct {
	Logger zerolog.Logger
	Ring   *Ring
	file   *lumberjack.Logger
}

// LogDir returns %APPDATA%\league-rpc\logs, creating it if missing.
func LogDir() (string, error) {
	base, err := config.GetConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "logs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create logs directory: %w", err)
	}
	return dir, nil
}

// New builds the Sink: a logger writing to both the rotating file and the
// ring buffer, at debug or info level per opts.
func New(opts Options) (*Sink, error) {
	dir, err := LogDir()
	if err != nil {
		return nil, err
	}

	maxSize := opts.MaxSizeMB
	if maxSize == 0 {
		maxSize = 5
	}
	maxBackups := opts.MaxBackups
	if maxBackups == 0 {
		maxBackups = 2
	}

	file := &lumberjack.Logger{
		Filename:   filepath.Join(dir, LogFileName),
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		Compress:   false,
	}
	ring := NewRing(opts.RingSize)

	// Left unset on the logger itself so the level can change live: with no
	// per-logger level, zerolog filters against the process-wide global level.
	SetDebug(opts.Debug)
	logger := zerolog.New(zerolog.MultiLevelWriter(file, ring)).
		With().Timestamp().Logger()

	return &Sink{Logger: logger, Ring: ring, file: file}, nil
}

// SetDebug updates the process-wide log level immediately, so a config
// change to Advanced.DebugMode takes effect without restarting.
func SetDebug(debug bool) {
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

// Close flushes and closes the underlying file.
func (s *Sink) Close() error {
	if s.file == nil {
		return nil
	}
	return s.file.Close()
}
