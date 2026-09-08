package app

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/its-haze/league-rpc/internal/config"
	"github.com/its-haze/league-rpc/internal/version"
)

// fakePauser records the last SetPaused value and reports it from IsPaused.
type fakePauser struct{ paused bool }

func (f *fakePauser) SetPaused(p bool) { f.paused = p }
func (f *fakePauser) IsPaused() bool   { return f.paused }

// fakeUpdater is a scripted AppUpdater.
type fakeUpdater struct {
	current     UpdateStatus
	checkErr    error
	downloadErr error
	restartErr  error
	changelog   string
	ran         bool
	onChange    func(UpdateStatus)
}

func (f *fakeUpdater) Run(ctx context.Context) {
	f.ran = true
	<-ctx.Done()
}
func (f *fakeUpdater) OnChange(fn func(UpdateStatus)) { f.onChange = fn }
func (f *fakeUpdater) Current() UpdateStatus          { return f.current }
func (f *fakeUpdater) Check(context.Context) (UpdateStatus, error) {
	return f.current, f.checkErr
}
func (f *fakeUpdater) Download(context.Context) error   { return f.downloadErr }
func (f *fakeUpdater) Restart(context.Context) error    { return f.restartErr }
func (f *fakeUpdater) Changelog(context.Context) string { return f.changelog }

func TestApp_GetVersion_DevPlaceholderWithoutInjection(t *testing.T) {
	a := New(config.NewStore(config.DefaultConfig()), &fakePauser{})
	if got := a.GetVersion(); got != version.DevPlaceholder {
		t.Fatalf("GetVersion() = %q, want %q", got, version.DevPlaceholder)
	}
}

func TestApp_UpdateMethods_NilUpdaterDegradeGracefully(t *testing.T) {
	a := New(config.NewStore(config.DefaultConfig()), &fakePauser{})

	if got := a.GetUpdateStatus(); got.Available {
		t.Fatalf("GetUpdateStatus() = %+v, want empty", got)
	}
	if _, err := a.CheckForUpdates(context.Background()); err != nil {
		t.Fatalf("CheckForUpdates without an updater errored: %v", err)
	}
	if err := a.RetryUpdate(context.Background()); err == nil {
		t.Fatal("RetryUpdate without an updater should error")
	}
	if err := a.RestartForUpdate(context.Background()); err == nil {
		t.Fatal("RestartForUpdate without an updater should error")
	}
	if got := a.GetChangelog(context.Background()); got != "changelog unavailable" {
		t.Fatalf("GetChangelog() = %q, want the fallback", got)
	}
}

func TestApp_UpdateMethods_DelegateToUpdater(t *testing.T) {
	u := &fakeUpdater{
		current:   UpdateStatus{Available: true, Version: "2.0.0"},
		changelog: "## notes",
	}
	a := New(config.NewStore(config.DefaultConfig()), &fakePauser{}, WithUpdater(u))

	if got := a.GetUpdateStatus(); got != u.current {
		t.Fatalf("GetUpdateStatus() = %+v, want %+v", got, u.current)
	}
	if got, err := a.CheckForUpdates(context.Background()); err != nil || got != u.current {
		t.Fatalf("CheckForUpdates() = (%+v, %v), want (%+v, nil)", got, err, u.current)
	}
	if got := a.GetChangelog(context.Background()); got != "## notes" {
		t.Fatalf("GetChangelog() = %q, want %q", got, "## notes")
	}

	u.downloadErr = errors.New("boom")
	if err := a.RetryUpdate(context.Background()); err == nil {
		t.Fatal("RetryUpdate did not propagate the updater's error")
	}
	u.downloadErr = nil
	if err := a.RetryUpdate(context.Background()); err != nil {
		t.Fatalf("RetryUpdate: %v", err)
	}
	if err := a.RestartForUpdate(context.Background()); err != nil {
		t.Fatalf("RestartForUpdate: %v", err)
	}

	var seen UpdateStatus
	a.OnUpdateChange(func(s UpdateStatus) { seen = s })
	u.onChange(UpdateStatus{Available: true, Version: "3.0.0"})
	if seen.Version != "3.0.0" {
		t.Fatalf("OnUpdateChange did not wire through to the updater: %+v", seen)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { a.RunUpdates(ctx); close(done) }()
	cancel()
	<-done
	if !u.ran {
		t.Fatal("RunUpdates did not call the updater's Run")
	}
}

func TestApp_ApplySettingsPersistsAndGetReflects(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())
	store := config.NewStore(config.DefaultConfig())
	a := New(store, &fakePauser{})

	cfg := a.GetSettings()
	cfg.Display.Default.ShowStats = false
	cfg.Advanced.UpdateInterval = 2500

	if err := a.ApplySettings(cfg); err != nil {
		t.Fatalf("ApplySettings failed: %v", err)
	}

	got := a.GetSettings()
	if got.Display.Default.ShowStats || got.Advanced.UpdateInterval != 2500 {
		t.Fatalf("GetSettings did not reflect the change: %+v", got)
	}
}

func TestApp_ApplySettingsRejectsInvalid(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())
	a := New(config.NewStore(config.DefaultConfig()), &fakePauser{})

	cfg := a.GetSettings()
	cfg.DiscordAppID = ""

	if err := a.ApplySettings(cfg); err == nil {
		t.Fatal("ApplySettings accepted an empty DiscordAppID")
	}
	if a.GetSettings().DiscordAppID == "" {
		t.Fatal("rejected ApplySettings still mutated live settings")
	}
}

func TestApp_RenderTemplatePreview_UsesSampleDataAndDefaults(t *testing.T) {
	a := New(config.NewStore(config.DefaultConfig()), &fakePauser{})

	dt, st := config.DefaultConfig().Presence.Templates["in-game"].Details, config.DefaultConfig().Presence.Templates["in-game"].State
	got, err := a.RenderTemplatePreview("in-game", config.TemplatePair{Details: dt, State: st}, nil)
	if err != nil {
		t.Fatalf("RenderTemplatePreview: %v", err)
	}
	if got.Details != "Ranked Solo/Duo" || got.State != "In Game · 3/2/5 · 120cs" {
		t.Fatalf("preview = %+v, want the sample-data render", got)
	}
	if len(got.Warnings) != 0 {
		t.Fatalf("warnings = %v, want none", got.Warnings)
	}
}

func TestApp_RenderTemplatePreview_ReportsUnknownTokens(t *testing.T) {
	a := New(config.NewStore(config.DefaultConfig()), &fakePauser{})

	got, err := a.RenderTemplatePreview("in-client",
		config.TemplatePair{Details: "{availability} {foo}", State: "{bar}"},
		map[string]string{"availability": "Online"})
	if err != nil {
		t.Fatalf("RenderTemplatePreview: %v", err)
	}
	if got.Details != "Online {foo}" || got.State != "{bar}" {
		t.Fatalf("preview = %+v, want unknown tokens left literal", got)
	}
	if want := []string{"unknown token {bar}", "unknown token {foo}"}; !reflect.DeepEqual(got.Warnings, want) {
		t.Fatalf("warnings = %v, want %v", got.Warnings, want)
	}
}

func TestApp_RenderTemplatePreview_RejectsUnknownContext(t *testing.T) {
	a := New(config.NewStore(config.DefaultConfig()), &fakePauser{})
	if _, err := a.RenderTemplatePreview("bogus", config.TemplatePair{}, nil); err == nil {
		t.Fatal("accepted an unknown presence context")
	}
}

// A blank line in the preview must resolve to the same default the daemon uses,
// so the settings screen can't show something Discord won't.
func TestApp_RenderTemplatePreview_BlankLineMatchesRuntimeDefault(t *testing.T) {
	a := New(config.NewStore(config.DefaultConfig()), &fakePauser{})

	got, err := a.RenderTemplatePreview("in-game", config.TemplatePair{Details: "", State: "custom state"}, nil)
	if err != nil {
		t.Fatalf("RenderTemplatePreview: %v", err)
	}
	if got.Details != "Ranked Solo/Duo" {
		t.Fatalf("blank Details previewed as %q, want the built-in default", got.Details)
	}
	if got.State != "custom state" {
		t.Fatalf("State = %q, want the override", got.State)
	}
}

// A non-nil but empty sample map (what the JS boundary produces from {}) must
// still fall back to representative data, not render every token empty.
func TestApp_RenderTemplatePreview_EmptyMapUsesSampleData(t *testing.T) {
	a := New(config.NewStore(config.DefaultConfig()), &fakePauser{})

	got, err := a.RenderTemplatePreview("champ-select", config.TemplatePair{}, map[string]string{})
	if err != nil {
		t.Fatalf("RenderTemplatePreview: %v", err)
	}
	if got.Details == "" {
		t.Fatal("empty sample map produced a blank preview")
	}
}

func TestApp_GetDisplayPreview_ShowsStatsWhenEnabled(t *testing.T) {
	a := New(config.NewStore(config.DefaultConfig()), &fakePauser{})

	got, err := a.GetDisplayPreview("in-game", config.TemplatePair{}, true, true)
	if err != nil {
		t.Fatalf("GetDisplayPreview: %v", err)
	}
	if got.State == "" || !strings.Contains(got.State, "3/2/5") {
		t.Fatalf("State = %q, want the sample KDA included", got.State)
	}
}

func TestApp_GetDisplayPreview_HidesStatsWhenDisabled(t *testing.T) {
	a := New(config.NewStore(config.DefaultConfig()), &fakePauser{})

	got, err := a.GetDisplayPreview("in-game", config.TemplatePair{}, false, true)
	if err != nil {
		t.Fatalf("GetDisplayPreview: %v", err)
	}
	if strings.Contains(got.State, "3/2/5") {
		t.Fatalf("State = %q, stats should be hidden", got.State)
	}
}

func TestApp_GetDisplayPreview_HidesEmojiWhenDisabled(t *testing.T) {
	a := New(config.NewStore(config.DefaultConfig()), &fakePauser{})

	got, err := a.GetDisplayPreview("in-client", config.TemplatePair{}, true, false)
	if err != nil {
		t.Fatalf("GetDisplayPreview: %v", err)
	}
	if strings.Contains(got.Details, "\U0001F7E2") {
		t.Fatalf("Details = %q, emoji should be hidden", got.Details)
	}
}

func TestApp_GetDisplayPreview_RejectsUnknownContext(t *testing.T) {
	a := New(config.NewStore(config.DefaultConfig()), &fakePauser{})
	if _, err := a.GetDisplayPreview("bogus", config.TemplatePair{}, true, true); err == nil {
		t.Fatal("accepted an unknown presence context")
	}
}

func TestApp_GetPreviewAssets_ReturnsNonEmptyURLs(t *testing.T) {
	a := New(config.NewStore(config.DefaultConfig()), &fakePauser{})

	got := a.GetPreviewAssets()
	if got.ChampionSkinURL == "" || got.TFTCompanionURL == "" || got.RankEmblemURL == "" || got.LeagueLogoURL == "" {
		t.Fatalf("GetPreviewAssets() = %+v, want every URL populated", got)
	}
}

func TestApp_GetTemplateTokens(t *testing.T) {
	a := New(config.NewStore(config.DefaultConfig()), &fakePauser{})

	got := a.GetTemplateTokens("in-client")
	want := []string{"emoji", "availability"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetTemplateTokens(in-client) = %v, want %v", got, want)
	}
	if got := a.GetTemplateTokens("bogus"); got != nil {
		t.Errorf("GetTemplateTokens(bogus) = %v, want nil", got)
	}
}

func TestApp_GetConfigBounds_MatchesConfigPackage(t *testing.T) {
	a := New(config.NewStore(config.DefaultConfig()), &fakePauser{})

	want := ConfigBounds{
		UpdateIntervalMin:       config.MinUpdateInterval,
		UpdateIntervalMax:       config.MaxUpdateInterval,
		StatsPollingIntervalMin: config.MinStatsPollingInterval,
		StatsPollingIntervalMax: config.MaxStatsPollingInterval,
	}
	if got := a.GetConfigBounds(); got != want {
		t.Errorf("GetConfigBounds() = %+v, want %+v", got, want)
	}
}

func TestApp_GetDefaultConfig_MatchesConfigDefault(t *testing.T) {
	a := New(config.NewStore(config.DefaultConfig()), &fakePauser{})
	got := a.GetDefaultConfig()
	want := *config.DefaultConfig()
	if got.DiscordAppID != want.DiscordAppID || got.Display.Default != want.Display.Default {
		t.Fatalf("GetDefaultConfig() = %+v, want %+v", got, want)
	}
}

func TestApp_GetPresetsNonEmpty(t *testing.T) {
	a := New(config.NewStore(config.DefaultConfig()), &fakePauser{})
	if len(a.GetPresets()) == 0 {
		t.Fatal("GetPresets returned nothing")
	}
}

type fakeAppLookup struct {
	name string
	err  error
}

func (f *fakeAppLookup) Name(context.Context, string) (string, error) { return f.name, f.err }

func TestApp_GetApplicationName_NilLookupErrors(t *testing.T) {
	a := New(config.NewStore(config.DefaultConfig()), &fakePauser{})
	if _, err := a.GetApplicationName(context.Background(), "123"); err == nil {
		t.Fatal("GetApplicationName without WithAppNameLookup should error")
	}
}

func TestApp_GetApplicationName_DelegatesToLookup(t *testing.T) {
	a := New(config.NewStore(config.DefaultConfig()), &fakePauser{},
		WithAppNameLookup(&fakeAppLookup{name: "Jungle diff"}))

	got, err := a.GetApplicationName(context.Background(), "123")
	if err != nil {
		t.Fatalf("GetApplicationName: %v", err)
	}
	if got != "Jungle diff" {
		t.Errorf("GetApplicationName() = %q, want %q", got, "Jungle diff")
	}
}

func TestApp_GetApplicationName_PropagatesLookupError(t *testing.T) {
	a := New(config.NewStore(config.DefaultConfig()), &fakePauser{},
		WithAppNameLookup(&fakeAppLookup{err: errors.New("not found")}))

	if _, err := a.GetApplicationName(context.Background(), "bogus"); err == nil {
		t.Fatal("GetApplicationName did not propagate the lookup error")
	}
}

func TestApp_PauseDelegatesToPauser(t *testing.T) {
	p := &fakePauser{}
	a := New(config.NewStore(config.DefaultConfig()), p)

	if a.IsPaused() {
		t.Fatal("expected unpaused initially")
	}
	a.SetPaused(true)
	if !p.paused || !a.IsPaused() {
		t.Fatal("SetPaused(true) did not reach the pauser")
	}
	a.SetPaused(false)
	if p.paused || a.IsPaused() {
		t.Fatal("SetPaused(false) did not reach the pauser")
	}
}
