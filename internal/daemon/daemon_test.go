package daemon

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/its-haze/league-rpc/internal/config"
	"github.com/its-haze/league-rpc/internal/discord"
	"github.com/its-haze/league-rpc/internal/state"
	"github.com/its-haze/league-rpc/pkg/types"
	"github.com/rs/zerolog"
)

type fakeRunner struct {
	started   atomic.Bool
	stopped   atomic.Bool
	connected atomic.Bool
}

func (r *fakeRunner) Run(ctx context.Context) {
	r.started.Store(true)
	<-ctx.Done()
	r.stopped.Store(true)
}

func (r *fakeRunner) Connected() bool { return r.connected.Load() }

// fakeLCURunner has directly controllable Connected()/LeagueProcessDetected(),
// skipping a real Supervisor's retry/gating behavior.
type fakeLCURunner struct {
	fakeRunner
	connected      atomic.Bool
	leagueDetected atomic.Bool
}

func (r *fakeLCURunner) Connected() bool             { return r.connected.Load() }
func (r *fakeLCURunner) LeagueProcessDetected() bool { return r.leagueDetected.Load() }

// fakeLiveGamePoller stands in for *livegame.Poller; Run/RunSpectating block
// until ctx is canceled, recording which kind was started and how often.
type fakeLiveGamePoller struct {
	mu         sync.Mutex
	starts     int
	spectates  int
	lastMode   types.GameMode
	running    atomic.Bool
	spectating atomic.Bool
}

func (f *fakeLiveGamePoller) Run(ctx context.Context, gameMode types.GameMode) {
	f.mu.Lock()
	f.starts++
	f.lastMode = gameMode
	f.mu.Unlock()

	f.running.Store(true)
	defer f.running.Store(false)
	<-ctx.Done()
}

func (f *fakeLiveGamePoller) RunSpectating(ctx context.Context) {
	f.mu.Lock()
	f.spectates++
	f.mu.Unlock()

	f.spectating.Store(true)
	defer f.spectating.Store(false)
	<-ctx.Done()
}

func (f *fakeLiveGamePoller) startCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.starts
}

func (f *fakeLiveGamePoller) spectateCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.spectates
}

// fakePresenceSender stands in for a real Discord IPC connection; UpdatePresence
// fails while connected is false, like discord.Client when Discord is unreachable.
type fakePresenceSender struct {
	mu        sync.Mutex
	sends     []*discord.RPCData
	cleared   int32
	connected atomic.Bool
}

func (f *fakePresenceSender) UpdatePresence(rpcData *discord.RPCData) error {
	if !f.connected.Load() {
		return errors.New("not connected to Discord RPC")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sends = append(f.sends, rpcData.Copy())
	return nil
}

func (f *fakePresenceSender) ClearPresence() error {
	atomic.AddInt32(&f.cleared, 1)
	return nil
}

func (f *fakePresenceSender) IsConnected() bool { return f.connected.Load() }

func (f *fakePresenceSender) lastSend() *discord.RPCData {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.sends) == 0 {
		return nil
	}
	return f.sends[len(f.sends)-1]
}

func (f *fakePresenceSender) sendCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.sends)
}

func (f *fakePresenceSender) clearCount() int32 {
	return atomic.LoadInt32(&f.cleared)
}

func newTestDaemonDeps() (*discord.Updater, *state.Manager, *fakePresenceSender) {
	logger := zerolog.Nop()
	sender := &fakePresenceSender{}
	sender.connected.Store(true) // Discord connected by default; tests that care override it.
	updater := discord.NewUpdater(sender, config.NewStore(config.DefaultConfig()), logger)
	stateMgr := state.NewManager(logger)
	return updater, stateMgr, sender
}

func TestDaemon_RunStartsBothSupervisorsAndBlocksUntilCanceled(t *testing.T) {
	discordRunner := &fakeRunner{}
	lcuRunner := &fakeLCURunner{}
	updater, stateMgr, _ := newTestDaemonDeps()
	d := New(discordRunner, lcuRunner, updater, stateMgr, &fakeLiveGamePoller{}, zerolog.Nop(), testPollInterval, testPollInterval)

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan struct{})
	go func() {
		d.Run(ctx)
		close(done)
	}()

	waitFor(t, testTimeout, func() bool {
		return discordRunner.started.Load() && lcuRunner.started.Load()
	})

	select {
	case <-done:
		t.Fatal("Run returned before ctx was canceled")
	default:
	}

	cancel()

	select {
	case <-done:
	case <-time.After(testTimeout):
		t.Fatal("Run did not return after ctx cancellation")
	}

	if !discordRunner.stopped.Load() || !lcuRunner.stopped.Load() {
		t.Fatal("expected both supervisors to observe ctx cancellation")
	}
}

func TestDaemon_RunNeverExitsWhileBothSupervisorsFailRepeatedly(t *testing.T) {
	discordSup := NewSupervisor(&fakeConnector{connectFailures: 1 << 30}, testRetryInterval, testPollInterval)
	lcuSup := NewLCUSupervisor(&fakeConnector{connectFailures: 1 << 30}, &fakeProcessChecker{}, testRetryInterval, testPollInterval, testPollInterval)
	updater, stateMgr, _ := newTestDaemonDeps()
	d := New(discordSup, lcuSup, updater, stateMgr, &fakeLiveGamePoller{}, zerolog.Nop(), testPollInterval, testPollInterval)

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan struct{})
	go func() {
		d.Run(ctx)
		close(done)
	}()

	time.Sleep(20 * testRetryInterval)
	select {
	case <-done:
		t.Fatal("Run returned despite both supervisors failing to connect")
	default:
	}

	cancel()
	select {
	case <-done:
	case <-time.After(testTimeout):
		t.Fatal("Run did not return after ctx cancellation")
	}
}

func TestDaemon_OneSupervisorFailingDoesNotBlockTheOther(t *testing.T) {
	stuck := NewSupervisor(&fakeConnector{connectFailures: 1 << 30}, testRetryInterval, testPollInterval)

	var healthyConnected atomic.Bool
	healthy := NewLCUSupervisor(&fakeConnector{}, &fakeProcessChecker{}, testRetryInterval, testPollInterval, testPollInterval,
		WithOnConnect(func() { healthyConnected.Store(true) }))

	updater, stateMgr, _ := newTestDaemonDeps()
	d := New(stuck, healthy, updater, stateMgr, &fakeLiveGamePoller{}, zerolog.Nop(), testPollInterval, testPollInterval)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go d.Run(ctx)

	waitFor(t, testTimeout, healthyConnected.Load)
}

func TestDaemon_RealPresenceOnceBothConnected(t *testing.T) {
	discordRunner := &fakeRunner{}
	discordRunner.connected.Store(true)
	lcuRunner := &fakeLCURunner{}
	updater, stateMgr, sender := newTestDaemonDeps()
	d := New(discordRunner, lcuRunner, updater, stateMgr, &fakeLiveGamePoller{}, zerolog.Nop(), testPollInterval, testPollInterval)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go d.Run(ctx)

	lcuRunner.connected.Store(true)

	waitFor(t, testTimeout, func() bool {
		last := sender.lastSend()
		return last != nil && last.State == "In Client"
	})
}

func TestDaemon_PresenceFollowsStateChangesOnceConnected(t *testing.T) {
	discordRunner := &fakeRunner{}
	discordRunner.connected.Store(true)
	lcuRunner := &fakeLCURunner{}
	updater, stateMgr, sender := newTestDaemonDeps()
	// UpdateInterval controls DelayUpdate's debounce; keep it short for the test.
	cfg := config.DefaultConfig()
	cfg.Advanced.UpdateInterval = 1
	updater = discord.NewUpdater(sender, config.NewStore(cfg), zerolog.Nop())
	d := New(discordRunner, lcuRunner, updater, stateMgr, &fakeLiveGamePoller{}, zerolog.Nop(), testPollInterval, testPollInterval)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go d.Run(ctx)

	lcuRunner.connected.Store(true)
	waitFor(t, testTimeout, func() bool { return sender.sendCount() > 0 })

	newState := state.NewState()
	newState.GameFlowPhase = types.GameFlowLobby
	newState.Players = 1
	newState.MaxPlayers = 5
	stateMgr.Update(newState)

	waitFor(t, testTimeout, func() bool {
		last := sender.lastSend()
		return last != nil && last.Details != "" && last.State == "In Lobby (1/5)"
	})
}

func TestDaemon_ResendsPresenceOnConfigChangeWhileConnected(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())
	discordRunner := &fakeRunner{}
	discordRunner.connected.Store(true)
	lcuRunner := &fakeLCURunner{}
	_, stateMgr, sender := newTestDaemonDeps()

	store := config.NewStore(config.DefaultConfig())
	updater := discord.NewUpdater(sender, store, zerolog.Nop())
	d := New(discordRunner, lcuRunner, updater, stateMgr, &fakeLiveGamePoller{}, zerolog.Nop(), testPollInterval, testPollInterval)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go d.Run(ctx)

	lcuRunner.connected.Store(true)
	waitFor(t, testTimeout, func() bool { return sender.sendCount() > 0 })

	// A config change alone, no state change, must trigger an immediate resend.
	before := sender.sendCount()
	next := *config.DefaultConfig()
	next.Display.Default.ShowRank = !next.Display.Default.ShowRank
	if err := store.Apply(next); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	waitFor(t, testTimeout, func() bool { return sender.sendCount() > before })
}

func TestDaemon_ClearsPresenceImmediatelyWhenDisconnectedAndNoLeagueProcess(t *testing.T) {
	discordRunner := &fakeRunner{}
	discordRunner.connected.Store(true)
	lcuRunner := &fakeLCURunner{}
	updater, stateMgr, sender := newTestDaemonDeps()
	d := New(discordRunner, lcuRunner, updater, stateMgr, &fakeLiveGamePoller{}, zerolog.Nop(), testPollInterval, testPollInterval)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go d.Run(ctx)

	waitFor(t, testTimeout, func() bool { return sender.clearCount() > 0 })

	// Connect, then disconnect with no League process detected. Must clear again.
	lcuRunner.connected.Store(true)
	waitFor(t, testTimeout, func() bool { return sender.sendCount() > 0 })

	before := sender.clearCount()
	lcuRunner.connected.Store(false)
	waitFor(t, testTimeout, func() bool { return sender.clearCount() > before })
}

func TestDaemon_ShowsRotatingPlaceholderWhileLeagueLaunchingAndStopsOnConnect(t *testing.T) {
	discordRunner := &fakeRunner{}
	discordRunner.connected.Store(true)
	lcuRunner := &fakeLCURunner{}
	updater, stateMgr, sender := newTestDaemonDeps()
	d := New(discordRunner, lcuRunner, updater, stateMgr, &fakeLiveGamePoller{}, zerolog.Nop(), testPollInterval, testPollInterval)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go d.Run(ctx)

	lcuRunner.leagueDetected.Store(true)

	waitFor(t, testTimeout, func() bool {
		last := sender.lastSend()
		return last != nil && last.Details == "Launching League..."
	})

	// Expect more than one send while still launching (rotation).
	waitFor(t, testTimeout, func() bool { return sender.sendCount() >= 2 })

	// The instant LCU connects, real presence takes over.
	lcuRunner.connected.Store(true)
	waitFor(t, testTimeout, func() bool {
		last := sender.lastSend()
		return last != nil && last.Details != "Launching League..."
	})
}

func TestDaemon_NoWarnLogWhenPlaceholderRacesAheadOfDiscordConnecting(t *testing.T) {
	discordRunner := &fakeRunner{} // starts disconnected
	lcuRunner := &fakeLCURunner{}
	_, stateMgr, sender := newTestDaemonDeps()
	writer := &countingWriter{}
	updater := discord.NewUpdater(sender, config.NewStore(config.DefaultConfig()), zerolog.New(writer))
	d := New(discordRunner, lcuRunner, updater, stateMgr, &fakeLiveGamePoller{}, zerolog.Nop(), testPollInterval, testPollInterval)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go d.Run(ctx)

	lcuRunner.leagueDetected.Store(true) // League launches; Discord hasn't connected yet
	time.Sleep(10 * testPollInterval)
	if sender.sendCount() != 0 {
		t.Fatalf("expected no placeholder sends while Discord isn't connected, got %d", sender.sendCount())
	}
	if got := writer.n.Load(); got != 0 {
		t.Fatalf("expected no warning logs while Discord isn't connected, got %d", got)
	}

	discordRunner.connected.Store(true) // Discord catches up
	waitFor(t, testTimeout, func() bool { return sender.sendCount() > 0 })
}

func TestDaemon_ResendsPresenceOnceDiscordConnectsAfterLCU(t *testing.T) {
	discordRunner := &fakeRunner{} // starts disconnected
	lcuRunner := &fakeLCURunner{}
	updater, stateMgr, sender := newTestDaemonDeps()
	sender.connected.Store(false) // Discord not reachable yet, like the real Client
	d := New(discordRunner, lcuRunner, updater, stateMgr, &fakeLiveGamePoller{}, zerolog.Nop(), testPollInterval, testPollInterval)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go d.Run(ctx)

	// LCU connects while Discord is still unreachable, so the send fails,
	// the same as a real disconnected discord.Client.
	lcuRunner.connected.Store(true)
	time.Sleep(10 * testPollInterval)
	if sender.sendCount() != 0 {
		t.Fatalf("expected no successful sends while Discord isn't connected, got %d", sender.sendCount())
	}

	// Discord connects. With no further state change, Daemon must still
	// resend the current presence rather than leaving Discord blank forever.
	sender.connected.Store(true)
	discordRunner.connected.Store(true)

	waitFor(t, testTimeout, func() bool {
		last := sender.lastSend()
		return last != nil && last.State == "In Client"
	})
}

// countingWriter counts Write calls, standing in for a log sink so tests can
// assert on log volume without parsing log lines.
type countingWriter struct {
	n atomic.Int32
}

func (w *countingWriter) Write(p []byte) (int, error) {
	w.n.Add(1)
	return len(p), nil
}

func TestDaemon_LogsOnceWhileWaitingForDiscordThenAgainOnNextEdge(t *testing.T) {
	discordRunner := &fakeRunner{} // starts disconnected
	lcuRunner := &fakeLCURunner{}
	updater, stateMgr, _ := newTestDaemonDeps()
	writer := &countingWriter{}
	logger := zerolog.New(writer)
	d := New(discordRunner, lcuRunner, updater, stateMgr, &fakeLiveGamePoller{}, logger, testPollInterval, testPollInterval)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go d.Run(ctx)

	lcuRunner.leagueDetected.Store(true) // League up, Discord still unreachable
	waitFor(t, testTimeout, func() bool { return writer.n.Load() >= 1 })

	time.Sleep(10 * testPollInterval)
	if got := writer.n.Load(); got != 1 {
		t.Fatalf("expected exactly one log line while waiting, got %d", got)
	}

	discordRunner.connected.Store(true) // Discord connects; no new log on this edge
	time.Sleep(10 * testPollInterval)
	if got := writer.n.Load(); got != 1 {
		t.Fatalf("expected no additional log line once Discord connects, got %d", got)
	}

	discordRunner.connected.Store(false) // Discord drops again while League still up
	waitFor(t, testTimeout, func() bool { return writer.n.Load() >= 2 })
}

func TestDaemon_PauseFlagDefaultsUnpaused(t *testing.T) {
	updater, stateMgr, _ := newTestDaemonDeps()
	d := New(&fakeRunner{}, &fakeLCURunner{}, updater, stateMgr, &fakeLiveGamePoller{}, zerolog.Nop(), testPollInterval, testPollInterval)

	if d.IsPaused() {
		t.Fatal("a fresh Daemon must start unpaused")
	}
}

func TestDaemon_PauseClearsPresenceAndUnpauseResumes(t *testing.T) {
	discordRunner := &fakeRunner{}
	discordRunner.connected.Store(true)
	lcuRunner := &fakeLCURunner{}
	updater, stateMgr, sender := newTestDaemonDeps()
	d := New(discordRunner, lcuRunner, updater, stateMgr, &fakeLiveGamePoller{}, zerolog.Nop(), testPollInterval, testPollInterval)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go d.Run(ctx)

	lcuRunner.connected.Store(true)
	waitFor(t, testTimeout, func() bool {
		last := sender.lastSend()
		return last != nil && last.State == "In Client"
	})

	// Pausing must clear presence at once, on the League-not-running path.
	before := sender.clearCount()
	d.SetPaused(true)
	waitFor(t, testTimeout, func() bool { return sender.clearCount() > before })

	if !d.IsPaused() {
		t.Fatal("IsPaused should report true after SetPaused(true)")
	}

	// A state change while paused must not push presence.
	sendsWhilePaused := sender.sendCount()
	st := state.NewState()
	st.GameFlowPhase = types.GameFlowLobby
	st.Players = 2
	st.MaxPlayers = 5
	stateMgr.Update(st)
	time.Sleep(10 * testPollInterval)
	if sender.sendCount() != sendsWhilePaused {
		t.Fatalf("presence sent while paused: before=%d after=%d", sendsWhilePaused, sender.sendCount())
	}

	// Unpausing resumes normal presence.
	resumeBefore := sender.sendCount()
	d.SetPaused(false)
	waitFor(t, testTimeout, func() bool { return sender.sendCount() > resumeBefore })
}

func TestDaemon_NoPlaceholderWhenLeagueProcessNotDetected(t *testing.T) {
	discordRunner := &fakeRunner{}
	lcuRunner := &fakeLCURunner{}
	updater, stateMgr, sender := newTestDaemonDeps()
	d := New(discordRunner, lcuRunner, updater, stateMgr, &fakeLiveGamePoller{}, zerolog.Nop(), testPollInterval, testPollInterval)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go d.Run(ctx)

	time.Sleep(10 * testPollInterval)

	if sender.sendCount() != 0 {
		t.Fatalf("expected no presence sends while LCU disconnected and no League process detected, got %d", sender.sendCount())
	}
}
