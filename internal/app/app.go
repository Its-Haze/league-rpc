// Package app is the seam the GUI binds to. Its methods are plain Go so they
// can be exposed as Wails bindings without this package importing Wails.
package app

import (
	"context"
	"fmt"
	"sort"

	"github.com/its-haze/league-rpc/internal/config"
	"github.com/its-haze/league-rpc/internal/discord"
	"github.com/its-haze/league-rpc/internal/logging"
	"github.com/its-haze/league-rpc/internal/presence/template"
	"github.com/its-haze/league-rpc/internal/state"
	"github.com/its-haze/league-rpc/internal/version"
	"github.com/its-haze/league-rpc/pkg/types"
)

// Pauser is the runtime pause control the daemon exposes. Kept as a local
// interface so this package stays free of a daemon import.
type Pauser interface {
	SetPaused(bool)
	IsPaused() bool
}

// UpdateStatus is the App Update state the GUI renders. Distinct from Updater
// (internal/discord), which only concerns Discord presence sends.
type UpdateStatus struct {
	Available   bool   `json:"available"`
	Version     string `json:"version"`
	Notes       string `json:"notes"`
	LastError   string `json:"last_error,omitempty"`
	Downloading bool   `json:"downloading"`
	Ready       bool   `json:"ready"`
}

// AppUpdater is the in-app self-update surface the GUI drives, over app-local
// types so this package need not import the Wails updater it wraps.
type AppUpdater interface {
	// Run drives the launch check and the periodic re-check until ctx is
	// canceled. A dev build returns immediately and never checks.
	Run(ctx context.Context)
	// OnChange registers the callback fired whenever the status changes.
	OnChange(func(UpdateStatus))
	// Current returns the last known status without checking again.
	Current() UpdateStatus
	// Check runs a check now, for the manual "Check for updates" action.
	Check(ctx context.Context) (UpdateStatus, error)
	// Download re-checks, then downloads, verifies, and swaps the binary.
	// Used only for the manual "Retry" action; it otherwise runs on its own.
	Download(ctx context.Context) error
	// Restart relaunches into the swapped binary. Only valid after Download.
	Restart(ctx context.Context) error
	// Changelog returns the latest release's notes as Markdown, or a fixed
	// placeholder when GitHub can't be reached.
	Changelog(ctx context.Context) string
}

// AppNameLookup resolves a Discord Application ID to its public display
type AppNameLookup interface {
	Name(ctx context.Context, appID string) (string, error)
}

// App exposes the settings surface to the frontend. Everything runtime-visible
// goes through the config.Store it holds; the daemon reads the same store.
type App struct {
	store  *config.Store
	pauser Pauser

	status    *statusBridge
	updater   AppUpdater
	appLookup AppNameLookup

	logs       *logging.Ring
	logDir     string
	openFolder func(string) error
}

// Option configures optional App wiring the settings surface does not need.
type Option func(*App)

// WithStatus wires the status bridge over conns, probe, and state changes.
func WithStatus(conns Connections, probe PresenceProbe, states <-chan *state.State) Option {
	return func(a *App) {
		a.status = newStatusBridge(conns, probe, a.pauser, states)
	}
}

// WithUpdater wires the App Update surface. Omitted entirely in a headless build.
func WithUpdater(u AppUpdater) Option {
	return func(a *App) { a.updater = u }
}

// WithAppNameLookup wires the Discord Application ID name lookup.
func WithAppNameLookup(l AppNameLookup) Option {
	return func(a *App) { a.appLookup = l }
}

// New builds an App over store and the daemon's pause control.
func New(store *config.Store, pauser Pauser, opts ...Option) *App {
	a := &App{store: store, pauser: pauser, openFolder: defaultOpenFolder}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// GetStatus returns the current status snapshot. Zero value until WithStatus
// is wired.
func (a *App) GetStatus() StatusSnapshot {
	if a.status == nil {
		return StatusSnapshot{}
	}
	return a.status.snapshot()
}

// OnStatusChange registers the callback fired whenever the status snapshot
// changes. The GUI adapter forwards it to a frontend event.
func (a *App) OnStatusChange(fn func(StatusSnapshot)) {
	if a.status == nil {
		return
	}
	a.status.setOnChange(fn)
}

// RunStatus drives the status bridge until ctx is canceled. Without WithStatus
// it just blocks until then. The GUI adapter runs this in a goroutine.
func (a *App) RunStatus(ctx context.Context) {
	if a.status == nil {
		<-ctx.Done()
		return
	}
	a.status.run(ctx)
}

// GetVersion returns the running build's version: the release tag injected at
// build time, or a clear placeholder for a dev build.
func (a *App) GetVersion() string {
	return version.Version()
}

// OnUpdateChange registers the callback fired whenever the App Update status
// changes. No-op if the updater was never wired (headless build).
func (a *App) OnUpdateChange(fn func(UpdateStatus)) {
	if a.updater == nil {
		return
	}
	a.updater.OnChange(fn)
}

// RunUpdates drives the App Update launch and periodic checks until ctx is
// canceled. Without WithUpdater it just blocks until then.
func (a *App) RunUpdates(ctx context.Context) {
	if a.updater == nil {
		<-ctx.Done()
		return
	}
	a.updater.Run(ctx)
}

// GetUpdateStatus returns the last known App Update status without checking
// again. Zero value until WithUpdater is wired.
func (a *App) GetUpdateStatus() UpdateStatus {
	if a.updater == nil {
		return UpdateStatus{}
	}
	return a.updater.Current()
}

// CheckForUpdates runs the manual "Check for updates" action.
func (a *App) CheckForUpdates(ctx context.Context) (UpdateStatus, error) {
	if a.updater == nil {
		return UpdateStatus{}, nil
	}
	return a.updater.Check(ctx)
}

// RetryUpdate downloads, verifies, and swaps the pending release by hand.
func (a *App) RetryUpdate(ctx context.Context) error {
	if a.updater == nil {
		return fmt.Errorf("no update available")
	}
	return a.updater.Download(ctx)
}

// RestartForUpdate relaunches into the freshly swapped binary. Only valid
// once the release is downloaded, verified, and staged (UpdateStatus.Ready).
func (a *App) RestartForUpdate(ctx context.Context) error {
	if a.updater == nil {
		return fmt.Errorf("no update staged")
	}
	return a.updater.Restart(ctx)
}

// GetChangelog returns the latest release's notes as Markdown, or a fixed
// placeholder when GitHub can't be reached.
func (a *App) GetChangelog(ctx context.Context) string {
	if a.updater == nil {
		return "changelog unavailable"
	}
	return a.updater.Changelog(ctx)
}

// GetSettings returns the current settings as a value copy for the frontend.
func (a *App) GetSettings() config.Config {
	return *a.store.Load()
}

// GetDefaultConfig returns the built-in default settings tree, so the GUI can
// offer a "reset to default" action next to each setting.
func (a *App) GetDefaultConfig() config.Config {
	return *config.DefaultConfig()
}

// ApplySettings validates cfg, persists it, and swaps it in live. A returned
// error means nothing changed and the message is meant for the user.
func (a *App) ApplySettings(cfg config.Config) error {
	return a.store.Apply(cfg)
}

// GetPresets returns the named Discord Application ID choices for the picker.
func (a *App) GetPresets() map[string]string {
	return config.DiscordAppIDPresets()
}

// GetApplicationName resolves appID to its public Discord Application name,
// so a custom ID can be shown by name instead of just its digits.
func (a *App) GetApplicationName(ctx context.Context, appID string) (string, error) {
	if a.appLookup == nil {
		return "", fmt.Errorf("application name lookup unavailable")
	}
	return a.appLookup.Name(ctx, appID)
}

// ConfigBounds is the min/max the Advanced screen clamps its numeric fields
// to, read from config so the UI can't drift from what Validate() accepts.
type ConfigBounds struct {
	UpdateIntervalMin       int `json:"update_interval_min"`
	UpdateIntervalMax       int `json:"update_interval_max"`
	StatsPollingIntervalMin int `json:"stats_polling_interval_min"`
	StatsPollingIntervalMax int `json:"stats_polling_interval_max"`
}

// GetConfigBounds returns the numeric bounds the Advanced screen clamps to.
func (a *App) GetConfigBounds() ConfigBounds {
	return ConfigBounds{
		UpdateIntervalMin:       config.MinUpdateInterval,
		UpdateIntervalMax:       config.MaxUpdateInterval,
		StatsPollingIntervalMin: config.MinStatsPollingInterval,
		StatsPollingIntervalMax: config.MaxStatsPollingInterval,
	}
}

// SetPaused toggles the runtime pause flag. Paused clears Discord presence
// immediately; it never persists to Config and resets to unpaused on restart.
func (a *App) SetPaused(paused bool) {
	a.pauser.SetPaused(paused)
}

// IsPaused reports the runtime pause flag.
func (a *App) IsPaused() bool {
	return a.pauser.IsPaused()
}

// SubscribeSettings returns a channel that receives the new settings on every
// successful ApplySettings. The GUI adapter forwards these to the frontend.
func (a *App) SubscribeSettings() <-chan *config.Config {
	return a.store.Subscribe()
}

// GetTemplateTokens returns the {token} names valid for ctx, so the Display
// screen can show a reference next to each editor. Nil for an unknown ctx.
func (a *App) GetTemplateTokens(ctx string) []string {
	return template.KnownTokens(template.Context(ctx))
}

// TemplatePreview is the rendered result of one presence template pair against
// sample data, plus any unknown-token warnings the engine reported.
type TemplatePreview struct {
	Details  string   `json:"details"`
	State    string   `json:"state"`
	Warnings []string `json:"warnings"`
}

// RenderTemplatePreview renders tmpl for ctx the same way the daemon will: a
// blank line falls back to the default, an empty sample to the built-in sample.
func (a *App) RenderTemplatePreview(ctx string, tmpl config.TemplatePair, sample map[string]string) (TemplatePreview, error) {
	tctx := template.Context(ctx)
	if !template.IsContext(tctx) {
		return TemplatePreview{}, fmt.Errorf("unknown presence context %q", ctx)
	}
	if len(sample) == 0 {
		sample = template.SampleData(tctx)
	}
	return renderTemplatePair(tctx, tmpl, sample), nil
}

// GetDisplayPreview renders ctx's template against sample data with the
func (a *App) GetDisplayPreview(ctx string, tmpl config.TemplatePair, showStats bool, showEmojis bool) (TemplatePreview, error) {
	tctx := template.Context(ctx)
	if !template.IsContext(tctx) {
		return TemplatePreview{}, fmt.Errorf("unknown presence context %q", ctx)
	}
	sample := template.SampleData(tctx)
	if !showStats {
		sample["stats"] = ""
	}
	if !showEmojis {
		sample["emoji"] = ""
	}
	return renderTemplatePair(tctx, tmpl, sample), nil
}

// PreviewAssets are static sample image URLs the Display screen's live
// preview uses to approximate a real presence card.
type PreviewAssets struct {
	ChampionSkinURL string `json:"champion_skin_url"`
	TFTCompanionURL string `json:"tft_companion_url"`
	RankEmblemURL   string `json:"rank_emblem_url"`
	LeagueLogoURL   string `json:"league_logo_url"`
	ProfileIconURL  string `json:"profile_icon_url"`
	MapIconURL      string `json:"map_icon_url"`
}

// Giftwrapped Poro (Snowdown 2013), used as the sample profile icon below.
const previewProfileIconID = 588

// Ossia, the default starter Little Legend, used as the sample TFT companion below.
const previewTFTCompanionIconPath = "loadouts/companions/tooltip_akalidragon_ossia_tier1.png"

// GetPreviewAssets returns those sample image URLs.
func (a *App) GetPreviewAssets() PreviewAssets {
	return PreviewAssets{
		ChampionSkinURL: discord.GetChampionSkinURL("Chogath", 1),
		TFTCompanionURL: discord.GetTFTCompanionURL(previewTFTCompanionIconPath),
		RankEmblemURL:   discord.GetRankEmblemURL(types.TierGold),
		LeagueLogoURL:   discord.GetLeagueLogoURL(),
		ProfileIconURL:  discord.GetProfileIconURL(previewProfileIconID),
		MapIconURL:      discord.GetMapIconURL(types.MapSummonersRift),
	}
}

// renderTemplatePair renders tmpl for tctx against sample and collects any
// unknown-token warnings, shared by RenderTemplatePreview and GetDisplayPreview.
func renderTemplatePair(tctx template.Context, tmpl config.TemplatePair, sample map[string]string) TemplatePreview {
	details, state, unknown := template.RenderPair(tctx, tmpl.Details, tmpl.State, sample)

	var warnings []string
	for _, name := range unknown {
		warnings = append(warnings, fmt.Sprintf("unknown token {%s}", name))
	}
	sort.Strings(warnings)

	return TemplatePreview{Details: details, State: state, Warnings: warnings}
}
