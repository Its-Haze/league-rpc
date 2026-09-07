package discord

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/its-haze/league-rpc/internal/config"
	"github.com/its-haze/league-rpc/internal/state"
	"github.com/rs/zerolog"
)

// fakePresenceSender stands in for a real Discord IPC connection.
type fakePresenceSender struct {
	mu        sync.Mutex
	sends     []*RPCData
	cleared   int
	connected atomic.Bool
}

func newFakePresenceSender() *fakePresenceSender {
	s := &fakePresenceSender{}
	s.connected.Store(true)
	return s
}

func (f *fakePresenceSender) UpdatePresence(rpcData *RPCData) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sends = append(f.sends, rpcData.Copy())
	return nil
}

func (f *fakePresenceSender) ClearPresence() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.cleared++
	return nil
}

func (f *fakePresenceSender) IsConnected() bool { return f.connected.Load() }

func (f *fakePresenceSender) sendCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.sends)
}

func (f *fakePresenceSender) sendAt(i int) *RPCData {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.sends[i]
}

// newTestState returns a minimal in-client state, good enough to map to a
// non-empty, non-cleared RPCData.
func newTestState() *state.State {
	return state.NewState()
}

// fakeTicker is a controllable Ticker: tests fire ticks by sending on C()
// and observe Reset() calls instead of waiting on real durations.
type fakeTicker struct {
	mu      sync.Mutex
	ch      chan time.Time
	resets  []time.Duration
	stopped bool
}

func newFakeTicker() *fakeTicker {
	return &fakeTicker{ch: make(chan time.Time)}
}

func (f *fakeTicker) C() <-chan time.Time { return f.ch }

func (f *fakeTicker) Reset(d time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.resets = append(f.resets, d)
}

func (f *fakeTicker) Stop() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.stopped = true
}

func (f *fakeTicker) lastReset() time.Duration {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.resets) == 0 {
		return 0
	}
	return f.resets[len(f.resets)-1]
}

func (f *fakeTicker) resetCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.resets)
}

func (f *fakeTicker) tick() {
	f.ch <- time.Now()
}

// fakeClock hands out a single fakeTicker, delivered over a channel so the
// test can grab it without polling/sleeping for Updater.Run's goroutine to start.
type fakeClock struct {
	created chan *fakeTicker
}

func newFakeClock() *fakeClock {
	return &fakeClock{created: make(chan *fakeTicker, 1)}
}

func (c *fakeClock) NewTicker(_ time.Duration) Ticker {
	t := newFakeTicker()
	c.created <- t
	return t
}

func waitForHeartbeat(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("condition not met within %s", timeout)
}

const (
	testHeartbeatInterval = time.Hour // never fires on its own; driven manually via fakeTicker.tick()
	testReclaimInterval   = time.Minute
	testReclaimBurstCount = 2
)

func newHeartbeatTestUpdater(t *testing.T, sender *fakePresenceSender) (*Updater, *fakeClock) {
	t.Helper()
	clock := newFakeClock()
	u := NewUpdater(sender, config.NewStore(config.DefaultConfig()), zerolog.Nop(),
		WithClock(clock),
		WithHeartbeatInterval(testHeartbeatInterval),
		WithReclaimInterval(testReclaimInterval),
		WithReclaimBurstCount(testReclaimBurstCount),
	)

	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	go u.Run(ctx)

	return u, clock
}

func TestUpdater_HeartbeatResendsCurrentPresenceAfterRealSend(t *testing.T) {
	sender := newFakePresenceSender()
	u, clock := newHeartbeatTestUpdater(t, sender)
	ticker := <-clock.created

	st := newTestState()
	u.ImmediateUpdate(st)
	if got := sender.sendCount(); got != 1 {
		t.Fatalf("expected 1 send after ImmediateUpdate, got %d", got)
	}

	ticker.tick()
	waitForHeartbeat(t, time.Second, func() bool { return sender.sendCount() == 2 })
}

func TestUpdater_NoHeartbeatBeforeAnyRealSend(t *testing.T) {
	sender := newFakePresenceSender()
	_, clock := newHeartbeatTestUpdater(t, sender)
	ticker := <-clock.created

	ticker.tick()
	time.Sleep(20 * time.Millisecond)
	if got := sender.sendCount(); got != 0 {
		t.Fatalf("expected no heartbeat sends before any real presence, got %d", got)
	}
}

func TestUpdater_NoHeartbeatWhilePresenceIsCleared(t *testing.T) {
	sender := newFakePresenceSender()
	u, clock := newHeartbeatTestUpdater(t, sender)
	ticker := <-clock.created

	u.ImmediateUpdate(newTestState())
	waitForHeartbeat(t, time.Second, func() bool { return sender.sendCount() == 1 })

	u.ClearPresence()
	ticker.tick()
	time.Sleep(20 * time.Millisecond)

	if got := sender.sendCount(); got != 1 {
		t.Fatalf("expected no heartbeat resends after ClearPresence, got %d", got)
	}
}

func TestUpdater_NoHeartbeatWhileDisconnected(t *testing.T) {
	// Regression: the Discord Connection Supervisor can drop the connection
	sender := newFakePresenceSender()
	u, clock := newHeartbeatTestUpdater(t, sender)
	ticker := <-clock.created

	u.ImmediateUpdate(newTestState())
	waitForHeartbeat(t, time.Second, func() bool { return sender.sendCount() == 1 })

	sender.connected.Store(false)
	ticker.tick()
	time.Sleep(20 * time.Millisecond)
	if got := sender.sendCount(); got != 1 {
		t.Fatalf("expected no heartbeat resend while disconnected, got %d sends", got)
	}

	sender.connected.Store(true)
	ticker.tick()
	waitForHeartbeat(t, time.Second, func() bool { return sender.sendCount() == 2 })
}

func TestUpdater_NoHeartbeatWhileShowingPlaceholder(t *testing.T) {
	sender := newFakePresenceSender()
	u, clock := newHeartbeatTestUpdater(t, sender)
	ticker := <-clock.created

	u.UpdatePlaceholder(BuildLaunchingPresence(0))
	if got := sender.sendCount(); got != 1 {
		t.Fatalf("expected 1 send after UpdatePlaceholder, got %d", got)
	}

	ticker.tick()
	time.Sleep(20 * time.Millisecond)
	if got := sender.sendCount(); got != 1 {
		t.Fatalf("expected no heartbeat resend of the placeholder, got %d", got)
	}
}

func TestUpdater_ReclaimBurstThenSettlesBackToHeartbeatCadence(t *testing.T) {
	sender := newFakePresenceSender()
	u, clock := newHeartbeatTestUpdater(t, sender)
	ticker := <-clock.created

	u.ImmediateUpdate(newTestState())
	waitForHeartbeat(t, time.Second, func() bool { return ticker.lastReset() == testReclaimInterval })

	for range testReclaimBurstCount {
		ticker.tick()
	}
	waitForHeartbeat(t, time.Second, func() bool { return sender.sendCount() == 1+testReclaimBurstCount })
	waitForHeartbeat(t, time.Second, func() bool { return ticker.lastReset() == testHeartbeatInterval })

	// Further ticks stay at the settled heartbeat cadence: no more Reset calls.
	resetsAfterBurst := ticker.resetCount()
	ticker.tick()
	waitForHeartbeat(t, time.Second, func() bool { return sender.sendCount() == 2+testReclaimBurstCount })
	if got := ticker.resetCount(); got != resetsAfterBurst {
		t.Fatalf("expected no further Reset calls once settled, resets went from %d to %d", resetsAfterBurst, got)
	}
}

func TestUpdater_EachRealChangeRestartsTheReclaimBurst(t *testing.T) {
	sender := newFakePresenceSender()
	u, clock := newHeartbeatTestUpdater(t, sender)
	ticker := <-clock.created

	u.ImmediateUpdate(newTestState())
	waitForHeartbeat(t, time.Second, func() bool { return ticker.lastReset() == testReclaimInterval })

	ticker.tick() // consume one reclaim tick, one left in the burst
	waitForHeartbeat(t, time.Second, func() bool { return sender.sendCount() == 2 })

	changed := newTestState()
	changed.GameFlowPhase = "Lobby"
	changed.Players = 1
	changed.MaxPlayers = 5
	u.ImmediateUpdate(changed)
	sendsAfterSecondChange := sender.sendCount()

	resetsAfterSecondChange := ticker.resetCount()
	if ticker.lastReset() != testReclaimInterval {
		t.Fatalf("expected reclaim burst to restart after a second real change, last reset = %s", ticker.lastReset())
	}

	// The full burst count must be available again, not just the 1 tick left over.
	for range testReclaimBurstCount {
		ticker.tick()
	}
	waitForHeartbeat(t, time.Second, func() bool { return sender.sendCount() == sendsAfterSecondChange+testReclaimBurstCount })
	if ticker.resetCount() <= resetsAfterSecondChange {
		t.Fatalf("expected the burst to settle back to heartbeat cadence again")
	}
}

// TestUpdater_HeartbeatTogglesZeroWidthSpaceOnUnchangedResends guards the fix
func TestUpdater_HeartbeatTogglesZeroWidthSpaceOnUnchangedResends(t *testing.T) {
	sender := newFakePresenceSender()
	u, clock := newHeartbeatTestUpdater(t, sender)
	ticker := <-clock.created

	u.ImmediateUpdate(newTestState())
	waitForHeartbeat(t, time.Second, func() bool { return sender.sendCount() == 1 })
	firstDetails := sender.sendAt(0).Details

	ticker.tick()
	waitForHeartbeat(t, time.Second, func() bool { return sender.sendCount() == 2 })
	secondDetails := sender.sendAt(1).Details
	if secondDetails == firstDetails {
		t.Fatalf("expected the first resend to differ from the real send (zero-width-space toggle), both were %q", firstDetails)
	}
	if secondDetails != firstDetails+zeroWidthSpace {
		t.Fatalf("expected the first resend to be the original details plus a zero-width space, got %q", secondDetails)
	}

	ticker.tick()
	waitForHeartbeat(t, time.Second, func() bool { return sender.sendCount() == 3 })
	thirdDetails := sender.sendAt(2).Details
	if thirdDetails != firstDetails {
		t.Fatalf("expected the toggle to alternate back to the original details, got %q", thirdDetails)
	}
}
