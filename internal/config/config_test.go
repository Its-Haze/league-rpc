package config

import (
	"encoding/json"
	"testing"

	"github.com/its-haze/league-rpc/internal/presence/template"
)

func TestDefaultConfig_ShipsEveryPresenceTemplate(t *testing.T) {
	got := DefaultConfig().Presence.Templates
	for _, ctx := range template.Contexts() {
		pair, ok := got[string(ctx)]
		if !ok {
			t.Fatalf("no default template for context %q", ctx)
		}
		wantD, wantS := template.Default(ctx)
		if pair.Details != wantD || pair.State != wantS {
			t.Errorf("context %q = %+v, want {%q %q}", ctx, pair, wantD, wantS)
		}
	}
}

func TestClamp_BackfillsMissingTemplatesButKeepsUserEdits(t *testing.T) {
	c := DefaultConfig()
	c.Presence.Templates = map[string]TemplatePair{
		"in-game": {Details: "custom {queue}", State: "custom"},
	}
	c.clamp()

	if c.Presence.Templates["in-game"].Details != "custom {queue}" {
		t.Error("clamp overwrote a user-set template")
	}
	if _, ok := c.Presence.Templates["spectating"]; !ok {
		t.Error("clamp did not backfill a missing template")
	}
}

func TestValidate_ReportsEveryProblemAtOnce(t *testing.T) {
	c := DefaultConfig()
	c.DiscordAppID = ""
	c.Theme = "neon"
	c.Behavior.CloseAction = "explode"
	c.Advanced.UpdateInterval = 10
	c.Advanced.StatsPollingInterval = 999999

	err := c.Validate()
	if err == nil {
		t.Fatal("Validate accepted an invalid config")
	}
	for _, want := range []string{"discord_app_id", "theme", "close_action", "update_interval", "stats_polling_interval"} {
		if !contains(err.Error(), want) {
			t.Errorf("error %q missing mention of %q", err, want)
		}
	}
}

func TestValidate_DoesNotMutate(t *testing.T) {
	c := DefaultConfig()
	c.DiscordAppID = ""
	c.Advanced.UpdateInterval = 10

	before, _ := json.Marshal(c)
	_ = c.Validate()
	after, _ := json.Marshal(c)

	if string(before) != string(after) {
		t.Fatalf("Validate mutated the config:\n%s\n%s", before, after)
	}
}

func TestValidate_AcceptsDefault(t *testing.T) {
	if err := DefaultConfig().Validate(); err != nil {
		t.Fatalf("DefaultConfig failed validation: %v", err)
	}
}

func TestClamp_RepairsOutOfBounds(t *testing.T) {
	c := &Config{
		Theme: "bogus",
		Advanced: AdvancedConfig{
			UpdateInterval:       50,
			StatsPollingInterval: 100,
		},
	}
	c.clamp()

	if c.DiscordAppID == "" {
		t.Error("clamp left DiscordAppID empty")
	}
	if c.Theme != ThemeSystem {
		t.Errorf("Theme = %q, want %q", c.Theme, ThemeSystem)
	}
	// A file written before close_action existed has it empty; asking is the
	// only safe repair, since quitting silently would surprise the user.
	if c.Behavior.CloseAction != CloseAsk {
		t.Errorf("CloseAction = %q, want %q", c.Behavior.CloseAction, CloseAsk)
	}
	if c.Advanced.UpdateInterval != DefaultConfig().Advanced.UpdateInterval {
		t.Errorf("UpdateInterval = %d, want default", c.Advanced.UpdateInterval)
	}
	if c.Advanced.StatsPollingInterval != MinStatsPollingInterval {
		t.Errorf("StatsPollingInterval = %d, want %d", c.Advanced.StatsPollingInterval, MinStatsPollingInterval)
	}
	if c.Presence.Templates == nil {
		t.Error("clamp left nested maps nil")
	}
	if err := c.Validate(); err != nil {
		t.Fatalf("config still invalid after clamp: %v", err)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
