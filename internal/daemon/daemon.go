package daemon

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/its-haze/league-rpc/internal/discord"
	"github.com/its-haze/league-rpc/internal/state"
	"github.com/its-haze/league-rpc/pkg/types"
	"github.com/rs/zerolog"
)

// runner is anything Daemon can start and supervise for the duration of a
// run. *Supervisor and *LCUSupervisor both satisfy this.
type runner interface {
	Run(ctx context.Context)
}

// discordRunner also reports whether Discord is connected, so Daemon can
// resend presence the moment it reconnects. *Supervisor satisfies this.
type discordRunner interface {
	runner
	Connected() bool
}

// lcuRunner also reports LCU connection and League-process state, the
// inputs Daemon needs to pick real presence, placeholder, or cleared.
type lcuRunner interface {
	runner
	Connected() bool
	LeagueProcessDetected() bool
}

// liveGamePoller polls Live Client Data for champion/skin/KDA/timer detail
type liveGamePoller interface {
	Run(ctx context.Context, gameMode types.GameMode)
	RunSpectating(ctx context.Context)
}

// presenceMode is what Daemon currently shows on Discord. See ADR-0002.
type presenceMode int

const (
	modeUnknown presenceMode = iota
	modeConnected
	modePlaceholder
	modeCleared
)

// Daemon owns the Discord and LCU Connection Supervisors and drives Discord
// presence off their state. Run is the single seam here. See ADR-0002.
type Daemon struct {
	discord  discordRunner
	lcu      lcuRunner
	updater  *discord.Updater
	state    *state.Manager
	liveGame liveGamePoller
	logger   zerolog.Logger

	presencePollInterval time.Duration
	placeholderInterval  time.Duration

	// paused is a runtime flag, never persisted to Config. It starts false
	// on every Daemon and clears presence for as long as it is set.
	paused      atomic.Bool
	pauseSignal chan struct{}
}

// New builds a Daemon that drives presence from stateMgr through updater.
func New(
	discordSup discordRunner,
	lcuSup lcuRunner,
	updater *discord.Updater,
	stateMgr *state.Manager,
	liveGame liveGamePoller,
	logger zerolog.Logger,
	presencePollInterval, placeholderInterval time.Duration,
) *Daemon {
	return &Daemon{
		discord:              discordSup,
		lcu:                  lcuSup,
		updater:              updater,
		state:                stateMgr,
		liveGame:             liveGame,
		logger:               logger,
		presencePollInterval: presencePollInterval,
		placeholderInterval:  placeholderInterval,
		pauseSignal:          make(chan struct{}, 1),
	}
}

// SetPaused sets the runtime pause flag and wakes the presence loop so it
// takes effect at once. Paused clears presence; unpausing resumes it.
func (d *Daemon) SetPaused(paused bool) {
	d.paused.Store(paused)
	select {
	case d.pauseSignal <- struct{}{}:
	default:
	}
}

// IsPaused reports the pause flag. A fresh Daemon is always unpaused.
func (d *Daemon) IsPaused() bool { return d.paused.Load() }

// DiscordConnected reports whether Discord IPC is currently reachable.
func (d *Daemon) DiscordConnected() bool { return d.discord.Connected() }

// LCUConnected reports whether the League Client API is currently reachable.
func (d *Daemon) LCUConnected() bool { return d.lcu.Connected() }

// LeagueProcessDetected reports whether League's own process is running.
func (d *Daemon) LeagueProcessDetected() bool { return d.lcu.LeagueProcessDetected() }

// LastSent returns the presence the Updater last pushed to Discord.
func (d *Daemon) LastSent() discord.LastSent { return d.updater.LastSent() }

// SubscribeState returns a fresh channel of state changes for the status
// bridge, independent of the presence loop's own subscription.
func (d *Daemon) SubscribeState() <-chan *state.State { return d.state.Subscribe() }

// Run starts both supervisors and the presence loop, and blocks until ctx
// is canceled; a connection failure on either side never returns early.
func (d *Daemon) Run(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		defer wg.Done()
		d.discord.Run(ctx)
	}()
	go func() {
		defer wg.Done()
		d.lcu.Run(ctx)
	}()
	go func() {
		defer wg.Done()
		d.updater.Run(ctx)
	}()
	go func() {
		defer wg.Done()
		d.presenceLoop(ctx)
	}()

	wg.Wait()
}

// presenceLoop shows real presence once the LCU connects, a rotating
// placeholder while League is launching, or clears presence. See ADR-0002.
func (d *Daemon) presenceLoop(ctx context.Context) {
	ticker := time.NewTicker(d.presencePollInterval)
	defer ticker.Stop()

	cfgUpdates := d.updater.ConfigChanges()

	mode := modeUnknown
	discordConnected := false
	waitingForDiscordLogged := false

	var placeholderTicker *time.Ticker
	var placeholderC <-chan time.Time
	var placeholderStart int64

	stopPlaceholder := func() {
		if placeholderTicker != nil {
			placeholderTicker.Stop()
			placeholderTicker = nil
			placeholderC = nil
		}
	}
	defer stopPlaceholder()

	sendPlaceholder := func() {
		// Skip while Discord isn't connected yet; the reconnect handler below resends once it is.
		if !d.discord.Connected() {
			return
		}
		d.updater.UpdatePlaceholder(discord.BuildLaunchingPresence(placeholderStart))
	}

	// The live game poller runs in one of two modes while mode == modeConnected:
	// "playing" during GameFlowInProgress, "spectating" during GameFlowWatching.
	const (
		liveNone = iota
		livePlaying
		liveSpectating
	)
	var liveGameCancel context.CancelFunc
	liveGameKind := liveNone

	var stopLiveGame func()
	stopLiveGame = func() {
		if liveGameKind == liveNone {
			return
		}
		liveGameKind = liveNone
		liveGameCancel()
		liveGameCancel = nil
	}
	startPlaying := func(gameMode types.GameMode) {
		if liveGameKind == livePlaying {
			return
		}
		stopLiveGame()
		liveGameKind = livePlaying
		var gameCtx context.Context
		gameCtx, liveGameCancel = context.WithCancel(ctx)
		go d.liveGame.Run(gameCtx, gameMode)
	}
	startSpectating := func() {
		if liveGameKind == liveSpectating {
			return
		}
		stopLiveGame()
		liveGameKind = liveSpectating
		var gameCtx context.Context
		gameCtx, liveGameCancel = context.WithCancel(ctx)
		go d.liveGame.RunSpectating(gameCtx)
	}
	defer stopLiveGame()

	// reconcile picks the presence mode from the current connection and pause
	// state. It runs on every poll tick and on any pause-flag change.
	reconcile := func() {
		// Pause is a runtime flag: while set, hold presence cleared the same
		// way League-not-running does, and skip the rest of the decision.
		if d.paused.Load() {
			stopPlaceholder()
			stopLiveGame()
			if mode != modeCleared {
				d.updater.ClearPresence()
				mode = modeCleared
			}
			return
		}

		// Discord can connect after the LCU already has; resend the
		// current mode's presence instead of waiting for a state change.
		nowConnected := d.discord.Connected()
		if nowConnected && !discordConnected {
			switch mode {
			case modeConnected:
				d.updater.ImmediateUpdate(d.state.Get())
			case modePlaceholder:
				sendPlaceholder()
			}
		}
		discordConnected = nowConnected

		// League is up but Discord isn't reachable yet: log it once per
		// edge, not on every poll tick, so it isn't spammy.
		if d.lcu.LeagueProcessDetected() && !nowConnected {
			if !waitingForDiscordLogged {
				d.logger.Info().Msg("League is running, waiting for Discord to be reachable")
				waitingForDiscordLogged = true
			}
		} else {
			waitingForDiscordLogged = false
		}

		switch {
		case d.lcu.Connected():
			st := d.state.Get()
			if mode != modeConnected {
				stopPlaceholder()
				// Skip while Discord isn't connected yet; the reconnect handler above resends once it is.
				if d.discord.Connected() {
					d.updater.ImmediateUpdate(st)
				}
				mode = modeConnected
			}

			switch st.GameFlowPhase {
			case types.GameFlowInProgress:
				startPlaying(st.GameMode)
			case types.GameFlowWatching:
				startSpectating()
			default:
				stopLiveGame()
			}

		case d.lcu.LeagueProcessDetected():
			if mode != modePlaceholder {
				mode = modePlaceholder
				placeholderStart = time.Now().Unix()
				sendPlaceholder()
				placeholderTicker = time.NewTicker(d.placeholderInterval)
				placeholderC = placeholderTicker.C
			}

		default:
			if mode != modeCleared {
				stopPlaceholder()
				d.updater.ClearPresence()
				mode = modeCleared
			}
		}

		// The live game poller only ever runs while mode == modeConnected
		if mode != modeConnected {
			stopLiveGame()
		}
	}

	for {
		select {
		case <-ctx.Done():
			return

		case st, ok := <-d.state.Updates():
			if !ok {
				return
			}
			if mode == modeConnected {
				d.updater.DelayUpdate(st)
			}

		case <-cfgUpdates:
			// Display settings may have changed; reflect them now instead of
			// waiting for the next real state change or poll tick.
			if mode == modeConnected {
				d.updater.ImmediateUpdate(d.state.Get())
			}

		case <-placeholderC:
			sendPlaceholder()

		case <-d.pauseSignal:
			reconcile()

		case <-ticker.C:
			reconcile()
		}
	}
}
