package lcu

import (
	"context"
	"fmt"
	"sync/atomic"

	lcu "github.com/its-haze/lcu-gopher"
	"github.com/its-haze/league-rpc/internal/config"
	"github.com/its-haze/league-rpc/internal/state"
	"github.com/rs/zerolog"
)

// GameModeFetcher recovers the current match's raw gameMode from Live
type GameModeFetcher interface {
	GameMode(ctx context.Context) (string, error)
}

// Client wraps the lcu-gopher client with application-specific logic
type Client struct {
	lcu      *lcu.Client
	state    *state.Manager
	store    *config.Store
	logger   zerolog.Logger
	liveGame GameModeFetcher

	connected atomic.Bool
}

// NewClient creates a new LCU client wrapper. liveGame recovers gameMode
// from Live Client Data when a (re)connect finds no lobby to query.
func NewClient(stateMgr *state.Manager, store *config.Store, logger zerolog.Logger, liveGame GameModeFetcher) *Client {
	return &Client{
		state:    stateMgr,
		store:    store,
		logger:   logger,
		liveGame: liveGame,
	}
}

// Connect establishes connection to the League Client
func (c *Client) Connect() error {
	c.logger.Info().Msg("Connecting to League Client...")

	// Configure lcu-gopher client. Start from DefaultConfig so Logger is
	// non-nil: lcu-gopher panics on a nil Logger while awaiting connection.
	cfg := c.store.Load()
	lcuConfig := lcu.DefaultConfig()
	lcuConfig.AwaitConnection = true // Wait for LCU to start
	lcuConfig.Debug = cfg.Advanced.DebugMode

	// Create LCU client
	var err error
	c.lcu, err = lcu.NewClient(lcuConfig)
	if err != nil {
		return fmt.Errorf("failed to create LCU client: %w", err)
	}

	// Connect to LCU
	if err := c.lcu.Connect(); err != nil {
		return fmt.Errorf("failed to connect to LCU: %w", err)
	}

	c.logger.Info().Msg("Successfully connected to League Client")

	// Gather initial data
	if err := c.gatherInitialData(); err != nil {
		c.logger.Warn().Err(err).Msg("Failed to gather initial data, will retry on events")
	}

	// Subscribe to LCU events
	if err := c.subscribeToEvents(); err != nil {
		return fmt.Errorf("failed to subscribe to events: %w", err)
	}

	c.logger.Info().Msg("LeagueRPC is ready")
	c.connected.Store(true)

	return nil
}

// Disconnect closes the connection to the League Client
func (c *Client) Disconnect() error {
	c.logger.Info().Msg("Disconnecting from League Client...")

	c.connected.Store(false)

	if c.lcu != nil {
		c.lcu.Disconnect()
	}

	return nil
}

// GetLCU returns the underlying lcu-gopher client
// Useful for making custom API calls
func (c *Client) GetLCU() *lcu.Client {
	return c.lcu
}

// IsConnected reports whether Connect() has succeeded and Disconnect() has
// not since been called. lcu-gopher has no live disconnect signal to detect League closing mid-session.
func (c *Client) IsConnected() bool {
	return c.connected.Load()
}
