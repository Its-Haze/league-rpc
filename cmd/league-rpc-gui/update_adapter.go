package main

import (
	"context"

	"github.com/its-haze/league-rpc/internal/app"
	"github.com/its-haze/league-rpc/internal/updates"
)

// updateAdapter fits *updates.Coordinator to internal/app's AppUpdater
// interface, converting Status to UpdateStatus at the boundary.
type updateAdapter struct{ c *updates.Coordinator }

func (a updateAdapter) Run(ctx context.Context)              { a.c.Run(ctx) }
func (a updateAdapter) Current() app.UpdateStatus            { return convertStatus(a.c.Status()) }
func (a updateAdapter) Download(ctx context.Context) error   { return a.c.Download(ctx) }
func (a updateAdapter) Restart(ctx context.Context) error    { return a.c.Restart(ctx) }
func (a updateAdapter) Changelog(ctx context.Context) string { return a.c.Changelog(ctx) }

func (a updateAdapter) Check(ctx context.Context) (app.UpdateStatus, error) {
	s, err := a.c.Check(ctx)
	return convertStatus(s), err
}

func (a updateAdapter) OnChange(fn func(app.UpdateStatus)) {
	a.c.OnChange(func(s updates.Status) { fn(convertStatus(s)) })
}

func convertStatus(s updates.Status) app.UpdateStatus {
	return app.UpdateStatus{
		Available:   s.Available,
		Version:     s.Version,
		Notes:       s.Notes,
		LastError:   s.LastError,
		Downloading: s.Downloading,
		Ready:       s.Ready,
	}
}
