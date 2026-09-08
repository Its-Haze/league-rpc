// Package updates checks GitHub Releases on a timer and auto-downloads a
// newer release once found; restarting into it is still the user's call.
package updates

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"
	wupdater "github.com/wailsapp/wails/v3/pkg/updater"
)

// engine is the slice of the Wails updater Coordinator drives. *wupdater.Updater
// satisfies it; tests supply a fake.
type engine interface {
	Check(ctx context.Context) (*wupdater.Release, error)
	DownloadAndInstall(ctx context.Context) error
	Restart(ctx context.Context) error
}

// Status is what the GUI renders for the App Update state. Available flips true
// only while a check has found a strictly newer release.
type Status struct {
	Available bool   `json:"available"`
	Version   string `json:"version"`
	Notes     string `json:"notes"`
	LastError string `json:"last_error,omitempty"`
	// Downloading is true while the background auto-download is in flight.
	Downloading bool `json:"downloading"`
	// Ready is true once the release is downloaded, verified, and staged;
	// only Restart is left.
	Ready bool `json:"ready"`
}

// Coordinator owns the check cadence and the download/restart actions; a dev
// build disables it entirely, since there is no released version to compare.
type Coordinator struct {
	eng       engine
	changelog *changelogFetcher
	log       zerolog.Logger
	dev       bool
	interval  time.Duration

	mu          sync.Mutex
	status      Status
	onChange    func(Status)
	runCtx      context.Context
	downloading bool
	ready       bool
	// failCount tracks consecutive auto-download failures; see doDownload.
	failCount int
}

// New builds a Coordinator. dev=true disables every check/download action.
// doer backs the changelog fetch and is injected so tests never hit GitHub.
func New(eng engine, doer HTTPDoer, dev bool, log zerolog.Logger) *Coordinator {
	return &Coordinator{
		eng:       eng,
		changelog: newChangelogFetcher(doer, RepoSlug),
		log:       log,
		dev:       dev,
		interval:  CheckInterval,
	}
}

// OnChange registers the callback fired whenever the status changes. The GUI
// adapter forwards it to a frontend event.
func (c *Coordinator) OnChange(fn func(Status)) {
	c.mu.Lock()
	c.onChange = fn
	c.mu.Unlock()
}

// Status returns the last known App Update status.
func (c *Coordinator) Status() Status {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.status
}

// Run does a launch check, then re-checks every CheckInterval, feeding ctx
// to any auto-download a check kicks off. A dev build returns immediately.
func (c *Coordinator) Run(ctx context.Context) {
	if c.dev {
		c.log.Info().Msg("dev build: App Update checks disabled")
		return
	}
	c.mu.Lock()
	c.runCtx = ctx
	c.mu.Unlock()

	c.checkNow(ctx)

	t := time.NewTicker(c.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			c.checkNow(ctx)
		}
	}
}

// Check runs one check now and reports the resulting status. Used by the
// manual "Check for updates" action.
func (c *Coordinator) Check(ctx context.Context) (Status, error) {
	if c.dev {
		return c.Status(), nil
	}
	return c.checkNow(ctx)
}

// checkNow runs one check and, unless a download is already running or
// staged, kicks off the auto-download in the background without waiting.
func (c *Coordinator) checkNow(ctx context.Context) (Status, error) {
	s, err := c.check(ctx)
	if err != nil || !s.Available {
		return s, err
	}

	c.mu.Lock()
	alreadyHandled := c.ready || c.downloading
	bg := c.runCtx
	c.mu.Unlock()
	if bg == nil {
		bg = context.Background()
	}

	if !alreadyHandled {
		go c.autoDownload(bg)
	}
	return s, err
}

// check performs one provider round trip; an error keeps any prior available
// release (recording the message), and a clean "no update" clears it.
func (c *Coordinator) check(ctx context.Context) (Status, error) {
	rel, err := c.eng.Check(ctx)

	c.mu.Lock()
	switch {
	case err != nil:
		c.status.LastError = err.Error()
		c.log.Warn().Err(err).Msg("App Update check failed")
	case rel == nil:
		c.status = Status{}
		c.downloading = false
		c.ready = false
		c.failCount = 0
	default:
		// Downloading/Ready carry forward: a re-check mid-download must not
		// erase that progress.
		c.status = Status{
			Available:   true,
			Version:     rel.Version,
			Notes:       rel.Notes,
			Downloading: c.downloading,
			Ready:       c.ready,
		}
		c.log.Info().Str("version", rel.Version).Msg("App Update available")
	}
	s := c.status
	cb := c.onChange
	c.mu.Unlock()

	if cb != nil {
		cb(s)
	}
	return s, err
}

// Download re-checks, then downloads, verifies, and swaps the release. This
// is the manual "Retry" action; auto-download otherwise runs on its own.
func (c *Coordinator) Download(ctx context.Context) error {
	if c.dev {
		return errors.New("updates: a dev build cannot self-update")
	}
	s, err := c.check(ctx)
	if err != nil {
		return fmt.Errorf("updates: re-check before download: %w", err)
	}
	if !s.Available {
		return errors.New("updates: no update available")
	}
	return c.doDownload(ctx, true)
}

// autoDownload runs the background download that follows a check finding a
// new release, skipping it if one is already in flight or already staged.
func (c *Coordinator) autoDownload(ctx context.Context) {
	c.mu.Lock()
	already := c.ready || c.downloading
	c.mu.Unlock()
	if already {
		return
	}
	_ = c.doDownload(ctx, false)
}

// doDownload downloads, verifies, and swaps the release. surfaceImmediately
// controls whether a failure reaches Status.LastError right away or waits.
func (c *Coordinator) doDownload(ctx context.Context, surfaceImmediately bool) error {
	c.mu.Lock()
	if c.ready {
		c.mu.Unlock()
		return nil
	}
	c.downloading = true
	c.status.Downloading = true
	s := c.status
	cb := c.onChange
	c.mu.Unlock()
	if cb != nil {
		cb(s)
	}

	err := c.eng.DownloadAndInstall(ctx)

	c.mu.Lock()
	c.downloading = false
	c.status.Downloading = false
	if err != nil {
		c.failCount++
		c.log.Warn().Err(err).Int("consecutive_failures", c.failCount).
			Msg("App Update download failed")
		if surfaceImmediately || c.failCount >= 2 {
			c.status.LastError = err.Error()
		}
	} else {
		c.failCount = 0
		c.ready = true
		c.status.Ready = true
		c.status.LastError = ""
	}
	s = c.status
	cb = c.onChange
	c.mu.Unlock()
	if cb != nil {
		cb(s)
	}

	if err != nil {
		return fmt.Errorf("updates: download and install: %w", err)
	}
	return nil
}

// Restart relaunches into the freshly swapped binary. Only valid after a
// successful Download.
func (c *Coordinator) Restart(ctx context.Context) error {
	if c.dev {
		return errors.New("updates: a dev build cannot self-update")
	}
	return c.eng.Restart(ctx)
}

// Changelog returns the latest release's notes as Markdown, or the literal
// "changelog unavailable" when GitHub can't be reached.
func (c *Coordinator) Changelog(ctx context.Context) string {
	body, err := c.changelog.fetch(ctx)
	if err != nil {
		c.log.Debug().Err(err).Msg("changelog fetch failed")
		return ChangelogUnavailable
	}
	return body
}
