package main

import (
	"testing"

	"github.com/its-haze/league-rpc/internal/app"
	"github.com/its-haze/league-rpc/internal/config"
)

func TestGUIService_SettingsRoundTrip(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())
	store := config.NewStore(config.DefaultConfig())
	svc := newGUIService(app.New(store, &fakePause{}))

	if len(svc.GetPresets()) == 0 {
		t.Fatal("GetPresets returned nothing")
	}

	cfg := svc.GetSettings()
	cfg.Presence.ShowEmojis = !cfg.Presence.ShowEmojis
	cfg.Advanced.UpdateInterval = 2222
	want := cfg.Presence.ShowEmojis

	if err := svc.ApplySettings(cfg); err != nil {
		t.Fatalf("ApplySettings: %v", err)
	}

	got := svc.GetSettings()
	if got.Presence.ShowEmojis != want || got.Advanced.UpdateInterval != 2222 {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
}

func TestGUIService_ApplyRejectsInvalid(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())
	svc := newGUIService(app.New(config.NewStore(config.DefaultConfig()), &fakePause{}))

	cfg := svc.GetSettings()
	cfg.DiscordAppID = ""

	if err := svc.ApplySettings(cfg); err == nil {
		t.Fatal("ApplySettings accepted an empty DiscordAppID")
	}
	if svc.GetSettings().DiscordAppID == "" {
		t.Fatal("rejected ApplySettings still mutated live settings")
	}
}
