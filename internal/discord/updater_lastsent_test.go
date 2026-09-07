package discord

import (
	"sync"
	"testing"

	"github.com/its-haze/league-rpc/internal/config"
	"github.com/its-haze/league-rpc/internal/state"
	"github.com/rs/zerolog"
)

func newLastSentTestUpdater(sender *fakePresenceSender) *Updater {
	return NewUpdater(sender, config.NewStore(config.DefaultConfig()), zerolog.Nop())
}

func TestUpdater_LastSent_ZeroBeforeAnySend(t *testing.T) {
	u := newLastSentTestUpdater(newFakePresenceSender())

	got := u.LastSent()
	if got.Data != nil || got.Cleared {
		t.Fatalf("LastSent before any send = %+v, want zero", got)
	}
}

func TestUpdater_LastSent_ReflectsLastRealSend(t *testing.T) {
	u := newLastSentTestUpdater(newFakePresenceSender())

	u.ImmediateUpdate(newTestState())

	got := u.LastSent()
	if got.Cleared {
		t.Fatal("LastSent marked cleared after a real send")
	}
	if got.Data == nil || got.Data.State != "In Client" {
		t.Fatalf("LastSent.Data = %+v, want the in-client payload", got.Data)
	}

	// The caller must not be able to mutate the Updater's copy.
	got.Data.State = "tampered"
	if again := u.LastSent(); again.Data.State != "In Client" {
		t.Fatalf("LastSent returned a shared pointer: second read = %q", again.Data.State)
	}
}

func TestUpdater_LastSent_MarksClear(t *testing.T) {
	u := newLastSentTestUpdater(newFakePresenceSender())

	u.ImmediateUpdate(newTestState())
	u.ClearPresence()

	got := u.LastSent()
	if !got.Cleared || got.Data != nil {
		t.Fatalf("LastSent after ClearPresence = %+v, want {Cleared:true Data:nil}", got)
	}
}

func TestUpdater_LastSent_MarksClearFromHideInClient(t *testing.T) {
	sender := newFakePresenceSender()
	cfg := config.DefaultConfig()
	cfg.Presence.ShowInClient = false
	u := NewUpdater(sender, config.NewStore(cfg), zerolog.Nop())

	// An in-client state with hide-in-client on takes executeUpdate's clear branch.
	u.ImmediateUpdate(state.NewState())

	got := u.LastSent()
	if !got.Cleared || got.Data != nil {
		t.Fatalf("LastSent = %+v, want the clear marker", got)
	}
}

func TestUpdater_LastSent_TracksPlaceholder(t *testing.T) {
	u := newLastSentTestUpdater(newFakePresenceSender())

	u.UpdatePlaceholder(BuildLaunchingPresence(0))
	if got := u.LastSent(); got.Cleared || got.Data == nil || got.Data.Details != "Launching League..." {
		t.Fatalf("LastSent after placeholder = %+v", got.Data)
	}
}

// Run with -race: LastSent must be safe to read while sends mutate the state.
func TestUpdater_LastSent_ConcurrentReadIsSafe(t *testing.T) {
	u := newLastSentTestUpdater(newFakePresenceSender())

	var wg sync.WaitGroup
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 200 {
				_ = u.LastSent()
			}
		}()
	}
	for range 200 {
		u.ImmediateUpdate(newTestState())
		u.ClearPresence()
	}
	wg.Wait()
}
