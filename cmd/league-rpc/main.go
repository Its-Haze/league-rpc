// Command league-rpc runs the always-on background daemon. See ADR-0001/0002.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/its-haze/league-rpc/internal/config"
	"github.com/its-haze/league-rpc/internal/daemon"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg, err := config.LoadOrCreate()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load configuration")
	}

	logger := newLogger(cfg.Advanced.DebugMode)
	store := config.NewStore(cfg)
	d := daemon.Wire(store, logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Info().Msg("League RPC starting")
	d.Run(ctx)
	logger.Info().Msg("League RPC stopped")
}

func newLogger(debug bool) zerolog.Logger {
	level := zerolog.InfoLevel
	if debug {
		level = zerolog.DebugLevel
	}
	return zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).
		Level(level).
		With().Timestamp().Logger()
}
