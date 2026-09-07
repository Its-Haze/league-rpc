package main

import (
	"context"

	"github.com/its-haze/league-rpc/internal/app"
	"github.com/its-haze/league-rpc/internal/config"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// configChangedEvent is emitted to the frontend whenever the live config
// changes, carrying the new settings so screens don't have to re-fetch.
const configChangedEvent = "settings:changed"

// statusChangedEvent carries a StatusSnapshot to the frontend whenever the
// connection state, phase, or last-sent presence changes.
const statusChangedEvent = "status:changed"

// updateChangedEvent carries an app.UpdateStatus to the frontend whenever the
// App Update coordinator's launch or periodic check finds something new.
const updateChangedEvent = "update:changed"

// logLineEvent carries one new log line to the frontend as it is written.
const logLineEvent = "log:line"

// closeRequestedEvent asks the frontend to raise the close confirmation
// dialog. Only emitted while close_action is "ask".
const closeRequestedEvent = "window:close-requested"

// navigateAboutEvent asks the frontend to switch to the About screen. Emitted
// when the user clicks the update-available toast notification.
const navigateAboutEvent = "navigate:about"

// updateReadyNotificationID identifies the toast shown once a new version is
// first discovered, distinguishing its click handler from future toasts.
const updateReadyNotificationID = "update-ready"

// guiService adapts internal/app (plain Go, Wails-free) to a Wails service.
type guiService struct {
	app *app.App
	// pauseHook, when set, handles pause changes so the tray checkbox stays
	// in sync. Nil until the tray is wired; falls back to app then.
	pauseHook func(bool)
	// closeHook applies the close dialog's answer. Nil until the tray is wired.
	closeHook func(string)
}

func newGUIService(a *app.App) *guiService {
	return &guiService{app: a}
}

// GetSettings returns the current settings tree.
func (s *guiService) GetSettings() config.Config {
	return s.app.GetSettings()
}

// GetDefaultConfig returns the built-in default settings tree, for the "reset
// to default" action next to each setting.
func (s *guiService) GetDefaultConfig() config.Config {
	return s.app.GetDefaultConfig()
}

// ApplySettings validates and persists cfg, swapping it in live. A returned
// error means nothing changed and the message is meant for the user.
func (s *guiService) ApplySettings(cfg config.Config) error {
	return s.app.ApplySettings(cfg)
}

// GetPresets returns the named Discord Application ID choices.
func (s *guiService) GetPresets() map[string]string {
	return s.app.GetPresets()
}

// GetApplicationName resolves appID to its public Discord Application name.
func (s *guiService) GetApplicationName(ctx context.Context, appID string) (string, error) {
	return s.app.GetApplicationName(ctx, appID)
}

// RenderTemplatePreview renders a presence template pair for ctx against sample
// data so the settings screen can preview an edit before it is saved.
func (s *guiService) RenderTemplatePreview(ctx string, tmpl config.TemplatePair, sample map[string]string) (app.TemplatePreview, error) {
	return s.app.RenderTemplatePreview(ctx, tmpl, sample)
}

// GetDisplayPreview renders ctx's template with showStats/showEmojis
// honored, for the Display screen's live preview of the current settings.
func (s *guiService) GetDisplayPreview(ctx string, tmpl config.TemplatePair, showStats bool, showEmojis bool) (app.TemplatePreview, error) {
	return s.app.GetDisplayPreview(ctx, tmpl, showStats, showEmojis)
}

// GetPreviewAssets returns sample image URLs for the Display screen's preview.
func (s *guiService) GetPreviewAssets() app.PreviewAssets {
	return s.app.GetPreviewAssets()
}

// GetStatus returns the current status snapshot for the frontend.
func (s *guiService) GetStatus() app.StatusSnapshot {
	return s.app.GetStatus()
}

// SetPaused toggles the runtime pause flag from the frontend.
func (s *guiService) SetPaused(paused bool) {
	if s.pauseHook != nil {
		s.pauseHook(paused)
		return
	}
	s.app.SetPaused(paused)
}

// ResolveClose applies the close dialog's answer: config.CloseQuit exits,
// anything else hides to the tray.
func (s *guiService) ResolveClose(action string) {
	if s.closeHook != nil {
		s.closeHook(action)
	}
}

// IsPaused reports the runtime pause flag to the frontend.
func (s *guiService) IsPaused() bool {
	return s.app.IsPaused()
}

// GetVersion returns the running build's version, or a dev placeholder.
func (s *guiService) GetVersion() string {
	return s.app.GetVersion()
}

// GetUpdateStatus returns the last known App Update status.
func (s *guiService) GetUpdateStatus() app.UpdateStatus {
	return s.app.GetUpdateStatus()
}

// CheckForUpdates runs the manual "Check for updates" action.
func (s *guiService) CheckForUpdates(ctx context.Context) (app.UpdateStatus, error) {
	return s.app.CheckForUpdates(ctx)
}

// RetryUpdate downloads, verifies, and swaps the pending release by hand.
func (s *guiService) RetryUpdate(ctx context.Context) error {
	return s.app.RetryUpdate(ctx)
}

// RestartForUpdate relaunches into the freshly swapped binary.
func (s *guiService) RestartForUpdate(ctx context.Context) error {
	return s.app.RestartForUpdate(ctx)
}

// GetChangelog returns the latest release's notes as Markdown, or a fixed
// placeholder when GitHub can't be reached.
func (s *guiService) GetChangelog(ctx context.Context) string {
	return s.app.GetChangelog(ctx)
}

// GetConfigBounds returns the numeric bounds the Advanced screen clamps to.
func (s *guiService) GetConfigBounds() app.ConfigBounds {
	return s.app.GetConfigBounds()
}

// GetTemplateTokens returns the {token} names valid for ctx.
func (s *guiService) GetTemplateTokens(ctx string) []string {
	return s.app.GetTemplateTokens(ctx)
}

// GetRecentLogs returns the buffered log lines, oldest first.
func (s *guiService) GetRecentLogs() []string {
	return s.app.GetRecentLogs()
}

// OpenLogsFolder opens the logs directory in Explorer.
func (s *guiService) OpenLogsFolder() error {
	return s.app.OpenLogsFolder()
}

// GetDiagnostics returns a paste-ready Markdown block for bug reports.
func (s *guiService) GetDiagnostics() string {
	return s.app.GetDiagnostics()
}

// publishConfigChanges forwards every config.Store update to the frontend as
// a Wails event until ctx is canceled.
func (s *guiService) publishConfigChanges(ctx context.Context, wailsApp *application.App) {
	updates := s.app.SubscribeSettings()
	for {
		select {
		case <-ctx.Done():
			return
		case cfg, ok := <-updates:
			if !ok {
				return
			}
			wailsApp.Event.Emit(configChangedEvent, cfg)
		}
	}
}

// publishLogLines forwards every new log line to the frontend as a Wails
// event until ctx is canceled. No-op if logging was never wired.
func (s *guiService) publishLogLines(ctx context.Context, wailsApp *application.App) {
	lines := s.app.SubscribeLogs()
	if lines == nil {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case line, ok := <-lines:
			if !ok {
				return
			}
			wailsApp.Event.Emit(logLineEvent, line)
		}
	}
}
