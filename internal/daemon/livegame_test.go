package daemon

import (
	"context"
	"testing"
	"time"

	"github.com/its-haze/league-rpc/internal/state"
	"github.com/its-haze/league-rpc/pkg/types"
	"github.com/rs/zerolog"
)

func TestDaemon_LiveGamePoller_StartsOnlyOnGameFlowInProgressAndStopsOtherwise(t *testing.T) {
	discordRunner := &fakeRunner{}
	lcuRunner := &fakeLCURunner{}
	updater, stateMgr, _ := newTestDaemonDeps()
	poller := &fakeLiveGamePoller{}
	d := New(discordRunner, lcuRunner, updater, stateMgr, poller, zerolog.Nop(), testPollInterval, testPollInterval)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go d.Run(ctx)

	// LCU connects while still in champ select: the poller must not start.
	champSelect := state.NewState()
	champSelect.GameFlowPhase = types.GameFlowChampSelect
	stateMgr.Update(champSelect)
	lcuRunner.connected.Store(true)

	time.Sleep(10 * testPollInterval)
	if poller.startCount() != 0 {
		t.Fatalf("startCount = %d, want 0 before GameFlowInProgress", poller.startCount())
	}

	// Phase moves to InProgress: the poller must start with the current game mode.
	inProgress := state.NewState()
	inProgress.GameFlowPhase = types.GameFlowInProgress
	inProgress.GameMode = types.GameModeARAM
	stateMgr.Update(inProgress)

	waitFor(t, testTimeout, func() bool { return poller.running.Load() })
	if got := poller.startCount(); got != 1 {
		t.Fatalf("startCount = %d, want 1", got)
	}
	if poller.lastMode != types.GameModeARAM {
		t.Fatalf("lastMode = %q, want ARAM", poller.lastMode)
	}

	// Phase leaves InProgress: the poller must be canceled (Run returns).
	postGame := state.NewState()
	postGame.GameFlowPhase = types.GameFlowEndOfGame
	stateMgr.Update(postGame)

	waitFor(t, testTimeout, func() bool { return !poller.running.Load() })
}

func TestDaemon_LiveGamePoller_SpectatingStartsSpectatePollerNotPlayPoller(t *testing.T) {
	discordRunner := &fakeRunner{}
	lcuRunner := &fakeLCURunner{}
	updater, stateMgr, _ := newTestDaemonDeps()
	poller := &fakeLiveGamePoller{}
	d := New(discordRunner, lcuRunner, updater, stateMgr, poller, zerolog.Nop(), testPollInterval, testPollInterval)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go d.Run(ctx)

	lcuRunner.connected.Store(true)
	watching := state.NewState()
	watching.GameFlowPhase = types.GameFlowWatching
	stateMgr.Update(watching)

	waitFor(t, testTimeout, func() bool { return poller.spectating.Load() })
	if poller.startCount() != 0 {
		t.Fatalf("startCount = %d, want 0: spectating must not start the play poller", poller.startCount())
	}
	if got := poller.spectateCount(); got != 1 {
		t.Fatalf("spectateCount = %d, want 1", got)
	}

	// Leaving Watching cancels the spectate poller.
	postGame := state.NewState()
	postGame.GameFlowPhase = types.GameFlowEndOfGame
	stateMgr.Update(postGame)
	waitFor(t, testTimeout, func() bool { return !poller.spectating.Load() })
}

func TestDaemon_LiveGamePoller_SwapsBetweenPlayAndSpectate(t *testing.T) {
	discordRunner := &fakeRunner{}
	lcuRunner := &fakeLCURunner{}
	updater, stateMgr, _ := newTestDaemonDeps()
	poller := &fakeLiveGamePoller{}
	d := New(discordRunner, lcuRunner, updater, stateMgr, poller, zerolog.Nop(), testPollInterval, testPollInterval)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go d.Run(ctx)
	lcuRunner.connected.Store(true)

	stateMgr.UpdateGameFlowPhase(string(types.GameFlowInProgress))
	waitFor(t, testTimeout, func() bool { return poller.running.Load() })

	stateMgr.UpdateGameFlowPhase(string(types.GameFlowWatching))
	waitFor(t, testTimeout, func() bool { return poller.spectating.Load() && !poller.running.Load() })
}

func TestDaemon_LiveGamePoller_ClearsPreviousMatchChampionBeforeNextMatchStarts(t *testing.T) {
	discordRunner := &fakeRunner{}
	lcuRunner := &fakeLCURunner{}
	updater, stateMgr, _ := newTestDaemonDeps()
	poller := &fakeLiveGamePoller{}
	d := New(discordRunner, lcuRunner, updater, stateMgr, poller, zerolog.Nop(), testPollInterval, testPollInterval)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go d.Run(ctx)
	lcuRunner.connected.Store(true)

	// First match: champion resolves and is stored, then the match ends.
	stateMgr.UpdateGameFlowPhase(string(types.GameFlowInProgress))
	waitFor(t, testTimeout, func() bool { return poller.running.Load() })
	stateMgr.UpdateChampion("DrMundo", "Dr. Mundo", "Corporate Mundo", "Rose Quartz", 3)

	stateMgr.UpdateGameFlowPhase(string(types.GameFlowEndOfGame))
	waitFor(t, testTimeout, func() bool { return !poller.running.Load() })

	if got := stateMgr.Get().ChampionID; got != "DrMundo" {
		t.Fatalf("precondition failed: ChampionID = %q, want DrMundo still set after match end", got)
	}

	// Second match starts: the previous match's champion must be cleared
	// before the new poller has a chance to resolve anything.
	stateMgr.UpdateGameFlowPhase(string(types.GameFlowInProgress))

	waitFor(t, testTimeout, func() bool { return stateMgr.Get().ChampionID == "" })
}

func TestDaemon_LiveGamePoller_StopsWhenLCUDisconnects(t *testing.T) {
	discordRunner := &fakeRunner{}
	lcuRunner := &fakeLCURunner{}
	updater, stateMgr, _ := newTestDaemonDeps()
	poller := &fakeLiveGamePoller{}
	d := New(discordRunner, lcuRunner, updater, stateMgr, poller, zerolog.Nop(), testPollInterval, testPollInterval)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go d.Run(ctx)

	inProgress := state.NewState()
	inProgress.GameFlowPhase = types.GameFlowInProgress
	stateMgr.Update(inProgress)
	lcuRunner.connected.Store(true)

	waitFor(t, testTimeout, func() bool { return poller.running.Load() })

	lcuRunner.connected.Store(false)

	waitFor(t, testTimeout, func() bool { return !poller.running.Load() })
}
