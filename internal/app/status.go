package app

import (
	"context"
	"sync"
	"time"

	"github.com/its-haze/league-rpc/internal/discord"
	"github.com/its-haze/league-rpc/internal/state"
)

// defaultStatusPollInterval is how often the bridge re-checks the connection
// accessors, which change without a state.Manager update.
const defaultStatusPollInterval = time.Second

// Connections reports the live connection state shown in the status view.
type Connections interface {
	DiscordConnected() bool
	LCUConnected() bool
	LeagueProcessDetected() bool
}

// PresenceProbe exposes the presence the Updater last transmitted, so the
// status view shows what Discord has rather than a recomputation.
type PresenceProbe interface {
	LastSent() discord.LastSent
}

// StatusSnapshot is the status read model the GUI renders. It is assembled in
// one place and pushed on change so screens never poll the daemon.
type StatusSnapshot struct {
	LeagueProcess    bool             `json:"league_process"`
	LCUConnected     bool             `json:"lcu_connected"`
	DiscordConnected bool             `json:"discord_connected"`
	Paused           bool             `json:"paused"`
	GameFlowPhase    string           `json:"gameflow_phase"`
	Presence         *discord.RPCData `json:"presence"`
	PresenceCleared  bool             `json:"presence_cleared"`
}

// equal reports whether two snapshots would render identically.
func (s StatusSnapshot) equal(o StatusSnapshot) bool {
	if s.LeagueProcess != o.LeagueProcess ||
		s.LCUConnected != o.LCUConnected ||
		s.DiscordConnected != o.DiscordConnected ||
		s.Paused != o.Paused ||
		s.GameFlowPhase != o.GameFlowPhase ||
		s.PresenceCleared != o.PresenceCleared {
		return false
	}
	if (s.Presence == nil) != (o.Presence == nil) {
		return false
	}
	if s.Presence != nil && !s.Presence.Equals(o.Presence) {
		return false
	}
	return true
}

// statusBridge assembles StatusSnapshot from the state manager, the supervisor
// accessors, and the pause flag, and calls onChange when it differs.
type statusBridge struct {
	conns  Connections
	probe  PresenceProbe
	pauser Pauser
	states <-chan *state.State

	pollInterval time.Duration

	mu       sync.Mutex
	phase    string
	last     StatusSnapshot
	haveLast bool
	onChange func(StatusSnapshot)
}

func newStatusBridge(conns Connections, probe PresenceProbe, pauser Pauser, states <-chan *state.State) *statusBridge {
	return &statusBridge{
		conns:        conns,
		probe:        probe,
		pauser:       pauser,
		states:       states,
		pollInterval: defaultStatusPollInterval,
	}
}

// setOnChange registers the callback run whenever the snapshot changes.
func (b *statusBridge) setOnChange(fn func(StatusSnapshot)) {
	b.mu.Lock()
	b.onChange = fn
	b.mu.Unlock()
}

// snapshot assembles the current status from all sources.
func (b *statusBridge) snapshot() StatusSnapshot {
	b.mu.Lock()
	phase := b.phase
	b.mu.Unlock()

	ls := b.probe.LastSent()
	return StatusSnapshot{
		LeagueProcess:    b.conns.LeagueProcessDetected(),
		LCUConnected:     b.conns.LCUConnected(),
		DiscordConnected: b.conns.DiscordConnected(),
		Paused:           b.pauser.IsPaused(),
		GameFlowPhase:    phase,
		Presence:         ls.Data,
		PresenceCleared:  ls.Cleared,
	}
}

// run watches state changes and polls the connection accessors until ctx is
// canceled, emitting on every change.
func (b *statusBridge) run(ctx context.Context) {
	ticker := time.NewTicker(b.pollInterval)
	defer ticker.Stop()

	b.emitIfChanged()
	for {
		select {
		case <-ctx.Done():
			return
		case st, ok := <-b.states:
			if !ok {
				return
			}
			b.mu.Lock()
			b.phase = string(st.GameFlowPhase)
			b.mu.Unlock()
			b.emitIfChanged()
		case <-ticker.C:
			b.emitIfChanged()
		}
	}
}

// emitIfChanged assembles a fresh snapshot and calls onChange only if it
// differs from the last one emitted.
func (b *statusBridge) emitIfChanged() {
	next := b.snapshot()

	b.mu.Lock()
	if b.haveLast && b.last.equal(next) {
		b.mu.Unlock()
		return
	}
	b.last = next
	b.haveLast = true
	cb := b.onChange
	b.mu.Unlock()

	if cb != nil {
		cb(next)
	}
}
