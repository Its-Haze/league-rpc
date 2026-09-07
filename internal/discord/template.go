package discord

import (
	"github.com/its-haze/league-rpc/internal/config"
	"github.com/its-haze/league-rpc/internal/presence/template"
)

// renderPresenceText renders ctx's details and state through the template
// engine, using the user's config override or the built-in default per line.
func renderPresenceText(cfg *config.Config, ctx template.Context, data map[string]string) (details, state string) {
	p := cfg.Presence.Templates[string(ctx)]
	details, state, _ = template.RenderPair(ctx, p.Details, p.State, data)
	return details, state
}
