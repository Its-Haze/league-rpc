package daemon

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

// fakeLeagueDetector lets tests control LeagueProcessDetected() directly.
type fakeLeagueDetector struct {
	detected atomic.Bool
}

func (f *fakeLeagueDetector) LeagueProcessDetected() bool { return f.detected.Load() }

func TestDiscordSupervisor_DetectsDisconnectWhenDiscordProcessDisappears(t *testing.T) {
	// Mirrors the real discord.Client: IsConnected() never flips back to false on its own.
	connector := &fakeConnector{}
	checker := &fakeProcessChecker{running: true}
	league := &fakeLeagueDetector{}
	league.detected.Store(true)

	var connectCount atomic.Int32
	// Gate mirrors production wiring: Connect() only fires while both are up.
	ds := newDiscordSupervisor(connector, checker, league, testRetryInterval, testPollInterval, testPollInterval,
		WithGate(&discordGate{checker: checker, league: league}),
		WithOnConnect(func() { connectCount.Add(1) }))

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go ds.Run(ctx)

	waitFor(t, testTimeout, func() bool { return connectCount.Load() == 1 })
	waitFor(t, testTimeout, ds.Connected)

	if !connector.IsConnected() {
		t.Fatal("underlying connector should still report connected, as the real discord.Client would")
	}

	checker.running = false // Discord's process is killed; connector.IsConnected() never flips
	waitFor(t, testTimeout, func() bool { return !ds.Connected() })

	checker.running = true // Discord relaunches
	waitFor(t, testTimeout, func() bool { return connectCount.Load() >= 2 })
	waitFor(t, testTimeout, ds.Connected)
}

func TestDiscordSupervisor_NeverConnectsWhileLeagueNotDetected(t *testing.T) {
	connector := &fakeConnector{}
	checker := &fakeProcessChecker{running: true}
	league := &fakeLeagueDetector{} // League not running

	ds := newDiscordSupervisor(connector, checker, league, testRetryInterval, testPollInterval, testPollInterval,
		WithGate(&discordGate{checker: checker, league: league}))

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go ds.Run(ctx)

	time.Sleep(10 * testRetryInterval)
	if connector.connectCalls.Load() != 0 {
		t.Fatalf("expected no Connect() attempts while League isn't running, got %d", connector.connectCalls.Load())
	}

	league.detected.Store(true) // League launches
	waitFor(t, testTimeout, ds.Connected)
}

func TestDiscordSupervisor_DisconnectsWhenLeagueProcessDisappears(t *testing.T) {
	// This is the actual bug: Discord stayed connected and kept showing a
	// stale presence after League closed, since nothing tied the two together.
	connector := &fakeConnector{}
	checker := &fakeProcessChecker{running: true}
	league := &fakeLeagueDetector{}
	league.detected.Store(true)

	var disconnectCount atomic.Int32
	ds := newDiscordSupervisor(connector, checker, league, testRetryInterval, testPollInterval, testPollInterval,
		WithGate(&discordGate{checker: checker, league: league}),
		WithOnDisconnect(func() { disconnectCount.Add(1) }))

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go ds.Run(ctx)

	waitFor(t, testTimeout, ds.Connected)
	if !connector.IsConnected() {
		t.Fatal("underlying connector should still report connected, as the real discord.Client would")
	}

	league.detected.Store(false) // League closes; connector.IsConnected() never flips on its own
	waitFor(t, testTimeout, func() bool { return !ds.Connected() })
	waitFor(t, testTimeout, func() bool { return disconnectCount.Load() >= 1 })
}

func TestDiscordSupervisor_ChecksErrorLeavesLastKnownState(t *testing.T) {
	connector := &fakeConnector{}
	checker := &fakeProcessChecker{running: true}
	league := &fakeLeagueDetector{}
	league.detected.Store(true)

	ds := newDiscordSupervisor(connector, checker, league, testRetryInterval, testPollInterval, testPollInterval)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go ds.Run(ctx)

	waitFor(t, testTimeout, ds.Connected)

	checker.err = errors.New("boom")
	waitFor(t, testTimeout, ds.Connected) // stays connected across transient checker errors
}
