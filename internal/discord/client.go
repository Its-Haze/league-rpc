package discord

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/its-haze/league-rpc/internal/config"
	"github.com/its-haze/league-rpc/internal/discord/ipc"
	"github.com/rs/zerolog"
)

const (
	opcodeHandshake = 0
	opcodeFrame     = 1
)

// ipcConn is what Client needs from a Discord IPC transport, letting tests
// substitute a fake for *ipc.Conn.
type ipcConn interface {
	Send(opcode int32, payload []byte) ([]byte, error)
	Close() error
}

// Client is a Discord IPC client, owning its own named-pipe connection so
// Connect/Disconnect can be called repeatedly with no shared global state.
type Client struct {
	logger zerolog.Logger
	store  *config.Store
	dial   func() (ipcConn, error)

	mu             sync.Mutex
	conn           ipcConn
	connected      bool
	connectedAppID string // app ID the live connection was opened with
}

// NewClient creates a new Discord RPC client
func NewClient(store *config.Store, logger zerolog.Logger) *Client {
	return &Client{
		logger: logger,
		store:  store,
		dial:   dialIPC,
	}
}

// dialIPC adapts ipc.Dial's concrete return type to the ipcConn interface.
func dialIPC() (ipcConn, error) {
	return ipc.Dial()
}

// Connect dials Discord's IPC pipe and performs the handshake. The app ID is
// read fresh from the Store so a changed setting takes effect on reconnect.
func (c *Client) Connect() error {
	appID := c.store.Load().DiscordAppID
	c.logger.Info().Str("app_id", appID).Msg("Connecting to Discord RPC...")

	conn, err := c.dial()
	if err != nil {
		return fmt.Errorf("failed to connect to Discord RPC: %w", err)
	}

	payload, err := json.Marshal(handshake{V: "1", ClientID: appID})
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to build Discord handshake: %w", err)
	}
	if _, err := conn.Send(opcodeHandshake, payload); err != nil {
		conn.Close()
		return fmt.Errorf("failed to send Discord handshake: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.connected = true
	c.connectedAppID = appID
	c.mu.Unlock()

	c.logger.Info().Msg("Successfully connected to Discord RPC")
	return nil
}

// Disconnect closes the Discord RPC connection.
func (c *Client) Disconnect() error {
	c.mu.Lock()
	conn := c.conn
	wasConnected := c.connected
	c.conn = nil
	c.connected = false
	c.mu.Unlock()

	if !wasConnected {
		return nil
	}

	c.logger.Info().Msg("Disconnecting from Discord RPC...")
	if conn != nil {
		conn.Close()
	}
	c.logger.Info().Msg("Disconnected from Discord RPC")
	return nil
}

// UpdatePresence updates the Discord Rich Presence with the provided data.
func (c *Client) UpdatePresence(rpcData *RPCData) error {
	if rpcData == nil {
		return fmt.Errorf("rpc data is nil")
	}

	c.logger.Debug().
		Str("details", rpcData.Details).
		Str("state", rpcData.State).
		Msg("Updating Discord presence")

	if err := c.setActivity(&activityPayload{
		Details: rpcData.Details,
		State:   rpcData.State,
		Assets: &activityAssets{
			LargeImage: rpcData.LargeImage,
			LargeText:  rpcData.LargeText,
			SmallImage: rpcData.SmallImage,
			SmallText:  rpcData.SmallText,
		},
		Timestamps: timestampFor(rpcData.Start),
	}); err != nil {
		return fmt.Errorf("failed to set Discord activity: %w", err)
	}

	c.logger.Debug().Msg("Discord presence updated successfully")
	return nil
}

// ClearPresence clears the Discord Rich Presence.
func (c *Client) ClearPresence() error {
	if !c.IsConnected() {
		return nil
	}

	c.logger.Debug().Msg("Clearing Discord presence")

	// Discord clears only on a null activity; an empty object still registers
	// one, which renders as the bare app name with a timer.
	if err := c.setActivity(nil); err != nil {
		return fmt.Errorf("failed to clear Discord activity: %w", err)
	}
	return nil
}

// setActivity sends a SET_ACTIVITY frame. It marks the client disconnected
// on any transport failure, so IsConnected() reflects reality immediately.
func (c *Client) setActivity(activity *activityPayload) error {
	c.mu.Lock()
	conn := c.conn
	connected := c.connected
	c.mu.Unlock()

	if !connected || conn == nil {
		return fmt.Errorf("not connected to Discord RPC")
	}

	nonce, err := newNonce()
	if err != nil {
		return err
	}

	payload, err := json.Marshal(activityFrame{
		Cmd: "SET_ACTIVITY",
		Args: activityArgs{
			Pid:      os.Getpid(),
			Activity: activity,
		},
		Nonce: nonce,
	})
	if err != nil {
		return err
	}

	if _, err := conn.Send(opcodeFrame, payload); err != nil {
		c.mu.Lock()
		c.connected = false
		c.conn = nil
		c.mu.Unlock()
		return err
	}
	return nil
}

// IsConnected returns whether the client is connected to Discord. A live
func (c *Client) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected && c.connectedAppID == c.store.Load().DiscordAppID
}

// timestampFor builds the "Elapsed" timer timestamp, or nil if unset.
func timestampFor(start int64) *activityTimestamp {
	if start <= 0 {
		return nil
	}
	return &activityTimestamp{Start: start}
}
