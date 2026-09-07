package app

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/its-haze/league-rpc/internal/config"
	"github.com/its-haze/league-rpc/internal/discord"
	"github.com/its-haze/league-rpc/internal/state"
	"github.com/its-haze/league-rpc/pkg/types"
)

type fakeConns struct {
	discord, lcu, league atomic.Bool
}

func (f *fakeConns) DiscordConnected() bool      { return f.discord.Load() }
func (f *fakeConns) LCUConnected() bool          { return f.lcu.Load() }
func (f *fakeConns) LeagueProcessDetected() bool { return f.league.Load() }

type fakeProbe struct {
	mu sync.Mutex
	ls discord.LastSent
}

func (f *fakeProbe) set(ls discord.LastSent) {
	f.mu.Lock()
	f.ls = ls
	f.mu.Unlock()
}

func (f *fakeProbe) LastSent() discord.LastSent {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.ls
}

func TestStatusBridge_SnapshotAssemblesFromEverySource(t *testing.T) {
	conns := &fakeConns{}
	conns.discord.Store(true)
	conns.league.Store(true)
	probe := &fakeProbe{}
	probe.set(discord.LastSent{Data: &discord.RPCData{Details: "Ranked Solo/Duo", State: "In Game"}})
	pauser := &fakePauser{paused: true}

	b := newStatusBridge(conns, probe, pauser, nil)
	b.phase = string(types.GameFlowInProgress)

	got := b.snapshot()
	want := StatusSnapshot{
		LeagueProcess:    true,
		LCUConnected:     false,
		DiscordConnected: true,
		Paused:           true,
		GameFlowPhase:    "InProgress",
		Presence:         &discord.RPCData{Details: "Ranked Solo/Duo", State: "In Game"},
	}
	if !got.equal(want) {
		t.Fatalf("snapshot = %+v, want %+v", got, want)
	}
}

func TestStatusBridge_EmitsOnChangeNotOnNoOp(t *testing.T) {
	conns := &fakeConns{}
	probe := &fakeProbe{}
	b := newStatusBridge(conns, probe, &fakePauser{}, nil)

	var mu sync.Mutex
	var seen []StatusSnapshot
	b.onChange = func(s StatusSnapshot) {
		mu.Lock()
		seen = append(seen, s)
		mu.Unlock()
	}

	b.emitIfChanged() // first ever: always emits
	b.emitIfChanged() // identical: must not emit

	mu.Lock()
	if len(seen) != 1 {
		mu.Unlock()
		t.Fatalf("after two identical assemblies, emitted %d times, want 1", len(seen))
	}
	mu.Unlock()

	conns.discord.Store(true)
	b.emitIfChanged() // changed: emits

	mu.Lock()
	defer mu.Unlock()
	if len(seen) != 2 {
		t.Fatalf("after a real change, emitted %d times, want 2", len(seen))
	}
	if !seen[1].DiscordConnected {
		t.Fatalf("second emit did not carry the Discord change: %+v", seen[1])
	}
}

func TestStatusBridge_RunPicksUpPhaseFromStateChannel(t *testing.T) {
	conns := &fakeConns{}
	states := make(chan *state.State, 1)
	b := newStatusBridge(conns, &fakeProbe{}, &fakePauser{}, states)
	b.pollInterval = time.Millisecond

	got := make(chan StatusSnapshot, 8)
	b.onChange = func(s StatusSnapshot) { got <- s }

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go b.run(ctx)

	st := state.NewState()
	st.GameFlowPhase = types.GameFlowChampSelect
	states <- st

	waitForSnapshot(t, got, func(s StatusSnapshot) bool {
		return s.GameFlowPhase == "ChampSelect"
	})
}

func TestApp_GetStatusReflectsSourcesAndZeroWithoutWiring(t *testing.T) {
	store := config.NewStore(config.DefaultConfig())

	plain := New(store, &fakePauser{})
	if plain.GetStatus() != (StatusSnapshot{}) {
		t.Fatalf("GetStatus without WithStatus = %+v, want zero", plain.GetStatus())
	}

	conns := &fakeConns{}
	conns.lcu.Store(true)
	wired := New(store, &fakePauser{paused: true},
		WithStatus(conns, &fakeProbe{}, nil))

	got := wired.GetStatus()
	if !got.LCUConnected || !got.Paused {
		t.Fatalf("GetStatus = %+v, want LCUConnected and Paused set", got)
	}
}

func waitForSnapshot(t *testing.T, ch <-chan StatusSnapshot, cond func(StatusSnapshot) bool) {
	t.Helper()
	deadline := time.After(2 * time.Second)
	for {
		select {
		case s := <-ch:
			if cond(s) {
				return
			}
		case <-deadline:
			t.Fatal("no matching status snapshot within 2s")
		}
	}
}
