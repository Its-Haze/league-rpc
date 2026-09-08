package daemon

import (
	"github.com/its-haze/league-rpc/internal/championdata"
	"github.com/its-haze/league-rpc/internal/config"
	"github.com/its-haze/league-rpc/internal/discord"
	"github.com/its-haze/league-rpc/internal/lcu"
	"github.com/its-haze/league-rpc/internal/livegame"
	"github.com/its-haze/league-rpc/internal/process"
	"github.com/its-haze/league-rpc/internal/state"
	"github.com/rs/zerolog"
)

// Wire builds a fully connected Daemon from a config Store and logger. The
// GUI app and the headless launcher both call this so they run the same graph.
func Wire(store *config.Store, logger zerolog.Logger) *Daemon {
	stateMgr := state.NewManager(logger)
	discordClient := discord.NewClient(store, logger)
	liveGameClient := livegame.NewClient(livegame.NewProductionHTTPDoer())
	lcuClient := lcu.NewClient(stateMgr, store, logger, liveGameClient)
	updater := discord.NewUpdater(discordClient, store, logger)
	checker := process.NewChecker()

	lcuSup := NewLeagueSupervisor(lcuClient, checker)
	discordSup := NewDiscordSupervisor(discordClient, checker, lcuSup)

	championResolver := championdata.NewResolver(championdata.NewProductionHTTPDoer())
	liveGamePoller := livegame.NewPoller(liveGameClient, championResolver, stateMgr, store, logger)

	return New(discordSup, lcuSup, updater, stateMgr, liveGamePoller, logger,
		DefaultPresencePollInterval, DefaultPlaceholderInterval)
}
