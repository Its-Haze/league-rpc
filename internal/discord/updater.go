package discord

import (
	"context"
	"sync"
	"time"

	"github.com/its-haze/league-rpc/internal/config"
	"github.com/its-haze/league-rpc/internal/state"
	"github.com/rs/zerolog"
)

// presenceSender is what Updater needs from a Discord RPC client, letting
// tests substitute a fake for *Client.
type presenceSender interface {
	UpdatePresence(rpcData *RPCData) error
	ClearPresence() error
	IsConnected() bool
}

var _ presenceSender = (*Client)(nil)

// Default heartbeat/reclaim cadence. Tighter than the reference's 5s/1.5s/4
// since League's native presence was observed outlasting a 5s heartbeat.
const (
	DefaultHeartbeatInterval = 2 * time.Second
	DefaultReclaimInterval   = 750 * time.Millisecond
	DefaultReclaimBurstCount = 8
)

// zeroWidthSpace forces a resend's Details to differ from what was last
// transmitted; Discord ignores a byte-identical SetActivity call. See ADR-0001.
const zeroWidthSpace = "​"

// UpdaterOption configures a Updater.
type UpdaterOption func(*Updater)

// WithClock overrides the Clock used for the heartbeat/reclaim ticker.
// Tests inject a fake; production uses the default real clock.
func WithClock(c Clock) UpdaterOption {
	return func(u *Updater) { u.clock = c }
}

// WithHeartbeatInterval overrides the cadence at which the current presence
// is resent once no reclaim burst is active.
func WithHeartbeatInterval(d time.Duration) UpdaterOption {
	return func(u *Updater) { u.heartbeatInterval = d }
}

// WithReclaimInterval overrides the faster cadence used for the burst of
// resends right after a real presence change.
func WithReclaimInterval(d time.Duration) UpdaterOption {
	return func(u *Updater) { u.reclaimInterval = d }
}

// WithReclaimBurstCount overrides how many reclaim-cadence ticks follow a
// real presence change before settling back to the heartbeat cadence.
func WithReclaimBurstCount(n int) UpdaterOption {
	return func(u *Updater) { u.reclaimBurstCount = n }
}

// Updater handles throttled Discord RPC updates and owns the heartbeat/
// reclaim cadence that defends them per ADR-0001.
type Updater struct {
	mu              sync.Mutex
	timer           *time.Timer
	previousRPCData *RPCData
	client          presenceSender
	store           *config.Store
	logger          zerolog.Logger

	// Track previous state for comparison
	previousState *state.State

	clock             Clock
	ticker            Ticker
	heartbeatInterval time.Duration
	reclaimInterval   time.Duration
	reclaimBurstCount int
	reclaimRemaining  int
	// heartbeating is true only while previousRPCData holds a real (not
	// placeholder, not cleared) presence that's worth defending. See ADR-0001.
	heartbeating bool
	// lastSentDetails is the Details string of the last payload actually
	lastSentDetails string
	// lastSent mirrors what Discord currently shows: the last payload sent, or
	// a cleared marker. The GUI preview reads this, never a recomputation.
	lastSent LastSent
}

// LastSent is the presence the Updater most recently pushed to Discord.
type LastSent struct {
	Data    *RPCData `json:"data"`
	Cleared bool     `json:"cleared"`
}

// NewUpdater creates a new RPC updater
func NewUpdater(client presenceSender, store *config.Store, logger zerolog.Logger, opts ...UpdaterOption) *Updater {
	u := &Updater{
		client:            client,
		store:             store,
		logger:            logger,
		clock:             NewRealClock(),
		heartbeatInterval: DefaultHeartbeatInterval,
		reclaimInterval:   DefaultReclaimInterval,
		reclaimBurstCount: DefaultReclaimBurstCount,
	}
	for _, opt := range opts {
		opt(u)
	}
	u.ticker = u.clock.NewTicker(u.heartbeatInterval)
	return u
}

// ConfigChanges reports when the live config changed, so the daemon can
// re-send presence at once instead of waiting for the next poll tick.
func (u *Updater) ConfigChanges() <-chan *config.Config {
	return u.store.Subscribe()
}

// Run drives the heartbeat/reclaim ticker until ctx is canceled. Daemon
// starts this alongside the Discord/LCU Connection Supervisors.
func (u *Updater) Run(ctx context.Context) {
	defer u.ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-u.ticker.C():
			u.heartbeat()
		}
	}
}

// heartbeat fires on every ticker tick: it advances/settles the reclaim
// burst, then resends the current presence if one is active. See ADR-0001.
func (u *Updater) heartbeat() {
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.reclaimRemaining > 0 {
		u.reclaimRemaining--
		if u.reclaimRemaining == 0 {
			u.ticker.Reset(u.heartbeatInterval)
		}
	}

	if !u.heartbeating || u.previousRPCData == nil {
		return
	}

	// Discord's Connection Supervisor may have dropped the connection since
	// the last real update; skip silently rather than warn on the expected error.
	if !u.client.IsConnected() {
		return
	}

	resend := u.previousRPCData.Copy()
	if resend.Details == u.lastSentDetails {
		resend.Details += zeroWidthSpace
	}

	if err := u.client.UpdatePresence(resend); err != nil {
		u.logger.Warn().Err(err).Msg("Failed to resend Discord presence (heartbeat)")
		return
	}
	u.lastSentDetails = resend.Details
	// Record the canonical payload, not resend: its Details may carry a
	// zero-width space only there to force Discord to accept the resend.
	u.recordSent(u.previousRPCData)
}

// LastSent returns what Discord currently shows: the last payload the Updater
// transmitted, or a cleared marker. Safe for concurrent reads.
func (u *Updater) LastSent() LastSent {
	u.mu.Lock()
	defer u.mu.Unlock()
	return LastSent{Data: u.lastSent.Data.Copy(), Cleared: u.lastSent.Cleared}
}

// recordSent / recordCleared are the single places lastSent is written, so a
// new send path can't forget to. Callers must hold u.mu.
func (u *Updater) recordSent(rpc *RPCData) { u.lastSent = LastSent{Data: rpc.Copy()} }
func (u *Updater) recordCleared()          { u.lastSent = LastSent{Cleared: true} }

// startReclaimBurst marks the current presence as worth defending and resets
// the ticker to the faster reclaim cadence. Callers must hold u.mu.
func (u *Updater) startReclaimBurst() {
	u.heartbeating = true
	u.reclaimRemaining = u.reclaimBurstCount
	if u.reclaimBurstCount > 0 {
		u.ticker.Reset(u.reclaimInterval)
	}
}

// stopHeartbeat marks there's no real presence left to defend and settles
// the ticker back to the normal heartbeat cadence. Callers must hold u.mu.
func (u *Updater) stopHeartbeat() {
	u.heartbeating = false
	if u.reclaimRemaining > 0 {
		u.reclaimRemaining = 0
		u.ticker.Reset(u.heartbeatInterval)
	}
}

// DelayUpdate schedules an RPC update after a delay
// If called multiple times rapidly, only the last call will execute (debouncing)
func (u *Updater) DelayUpdate(newState *state.State) {
	u.mu.Lock()
	defer u.mu.Unlock()

	// Check if state actually changed
	if u.previousState != nil && u.previousState.Equals(newState) {
		u.logger.Debug().Msg("State unchanged, skipping RPC update")
		return
	}

	// Cancel previous timer if exists
	if u.timer != nil {
		u.timer.Stop()
	}

	// Create copy of state for the timer closure
	stateCopy := newState.Copy()

	// Calculate delay from config (convert ms to duration)
	delay := time.Duration(u.store.Load().Advanced.UpdateInterval) * time.Millisecond

	// Start new timer
	u.timer = time.AfterFunc(delay, func() {
		u.executeUpdate(stateCopy)
	})

	u.logger.Debug().
		Dur("delay", delay).
		Str("phase", string(newState.GameFlowPhase)).
		Msg("RPC update scheduled")
}

// executeUpdate actually performs the Discord RPC update
func (u *Updater) executeUpdate(st *state.State) {
	u.mu.Lock()
	defer u.mu.Unlock()

	// Store as previous state
	u.previousState = st.Copy()

	cfg := u.store.Load()

	// Map state to RPC data
	rpcData := MapStateToPresence(st, cfg)

	// Check if we should clear presence instead
	if ShouldClearPresence(st, cfg) {
		u.logger.Debug().Msg("Clearing presence (hide-in-client enabled)")
		if err := u.client.ClearPresence(); err != nil {
			u.logger.Warn().Err(err).Msg("Failed to clear Discord presence")
		}
		u.previousRPCData = nil
		u.stopHeartbeat()
		u.recordCleared()
		return
	}

	// Check if RPC data actually changed
	if u.previousRPCData != nil && u.previousRPCData.Equals(rpcData) {
		u.logger.Debug().Msg("RPC data unchanged, skipping Discord API call")
		return
	}

	// Update Discord
	u.logger.Debug().
		Str("details", rpcData.Details).
		Str("state", rpcData.State).
		Msg("Updating Discord RPC")

	if err := u.client.UpdatePresence(rpcData); err != nil {
		u.logger.Warn().Err(err).Msg("Failed to update Discord presence")
		return
	}

	// Store as previous RPC data
	u.previousRPCData = rpcData.Copy()
	u.lastSentDetails = rpcData.Details
	u.recordSent(rpcData)
	u.startReclaimBurst()
}

// ImmediateUpdate performs an immediate RPC update without delay
// Useful for initial connection or manual refresh
func (u *Updater) ImmediateUpdate(st *state.State) {
	u.mu.Lock()
	defer u.mu.Unlock()

	// Cancel any pending timer
	if u.timer != nil {
		u.timer.Stop()
		u.timer = nil
	}

	cfg := u.store.Load()

	// Map state to RPC data
	rpcData := MapStateToPresence(st, cfg)

	// Check if we should clear presence
	if ShouldClearPresence(st, cfg) {
		u.logger.Debug().Msg("Clearing presence immediately")
		if err := u.client.ClearPresence(); err != nil {
			u.logger.Warn().Err(err).Msg("Failed to clear Discord presence")
		}
		u.previousRPCData = nil
		u.previousState = st.Copy()
		u.stopHeartbeat()
		u.recordCleared()
		return
	}

	// Update Discord
	u.logger.Debug().
		Str("details", rpcData.Details).
		Str("state", rpcData.State).
		Msg("Updating Discord RPC (immediate)")

	if err := u.client.UpdatePresence(rpcData); err != nil {
		u.logger.Warn().Err(err).Msg("Failed to update Discord presence")
		return
	}

	// Store state and RPC data
	u.previousState = st.Copy()
	u.previousRPCData = rpcData.Copy()
	u.lastSentDetails = rpcData.Details
	u.recordSent(rpcData)
	u.startReclaimBurst()
}

// UpdatePlaceholder sends rpcData as-is, bypassing MapStateToPresence, for
// the "Launching League..." rotation shown before there's any state to map.
func (u *Updater) UpdatePlaceholder(rpcData *RPCData) {
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.timer != nil {
		u.timer.Stop()
		u.timer = nil
	}

	if err := u.client.UpdatePresence(rpcData); err != nil {
		u.logger.Warn().Err(err).Msg("Failed to update Discord placeholder presence")
		return
	}

	// Clear previousState so the first real state isn't skipped as "unchanged".
	u.previousState = nil
	u.previousRPCData = rpcData.Copy()
	u.recordSent(rpcData)
	u.stopHeartbeat()
}

// ClearPresence clears the Discord presence
func (u *Updater) ClearPresence() {
	u.mu.Lock()
	defer u.mu.Unlock()

	// Cancel any pending timer
	if u.timer != nil {
		u.timer.Stop()
		u.timer = nil
	}

	if err := u.client.ClearPresence(); err != nil {
		u.logger.Warn().Err(err).Msg("Failed to clear Discord presence")
	}

	u.previousRPCData = nil
	u.previousState = nil
	u.recordCleared()
	u.stopHeartbeat()
}
