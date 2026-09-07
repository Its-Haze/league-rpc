package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStore_LoadReturnsInitial(t *testing.T) {
	init := DefaultConfig()
	s := NewStore(init)
	if s.Load() != init {
		t.Fatal("Load did not return the initial config pointer")
	}
}

func TestStore_ApplySwapsPersistsAndNotifies(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())
	s := NewStore(DefaultConfig())
	sub := s.Subscribe()

	next := *DefaultConfig()
	next.Display.Default.ShowRank = false
	next.Advanced.UpdateInterval = 2000

	if err := s.Apply(next); err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}

	if s.Load().Display.Default.ShowRank || s.Load().Advanced.UpdateInterval != 2000 {
		t.Fatalf("Load did not reflect applied config: %+v", s.Load())
	}

	select {
	case got := <-sub:
		if got.Advanced.UpdateInterval != 2000 {
			t.Fatalf("subscriber got stale config: %+v", got)
		}
	default:
		t.Fatal("subscriber was not notified")
	}

	path, _ := GetConfigPath()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("config file not written: %v", err)
	}
	if dir := filepath.Dir(path); dir == "" {
		t.Fatal("empty config dir")
	}
}

func TestStore_ApplyRejectsInvalidAndKeepsOld(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())
	orig := DefaultConfig()
	s := NewStore(orig)
	sub := s.Subscribe()

	bad := *DefaultConfig()
	bad.Advanced.UpdateInterval = 10 // below MinUpdateInterval

	if err := s.Apply(bad); err == nil {
		t.Fatal("Apply accepted an out-of-range UpdateInterval")
	}
	if s.Load() != orig {
		t.Fatal("Load changed after a rejected Apply")
	}
	select {
	case <-sub:
		t.Fatal("subscriber notified on a rejected Apply")
	default:
	}
}

func TestStore_SubscribeCoalescesForSlowSubscriber(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())
	s := NewStore(DefaultConfig())
	sub := s.Subscribe()

	for _, iv := range []int{1000, 2000, 3000} {
		next := *DefaultConfig()
		next.Advanced.UpdateInterval = iv
		if err := s.Apply(next); err != nil {
			t.Fatalf("Apply(%d) failed: %v", iv, err)
		}
	}

	got := <-sub
	if got.Advanced.UpdateInterval != 3000 {
		t.Fatalf("coalesced subscriber got %d, want the latest 3000", got.Advanced.UpdateInterval)
	}
	select {
	case extra := <-sub:
		t.Fatalf("expected only the latest value buffered, also got %d", extra.Advanced.UpdateInterval)
	default:
	}
}
