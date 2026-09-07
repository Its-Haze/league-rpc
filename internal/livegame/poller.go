package livegame

import (
	"context"
	"time"

	"github.com/its-haze/league-rpc/internal/championdata"
	"github.com/its-haze/league-rpc/internal/config"
	"github.com/its-haze/league-rpc/pkg/types"
	"github.com/rs/zerolog"
)

// ChampionResolver resolves a raw champion/skin triple into display data.
// *championdata.Resolver satisfies this.
type ChampionResolver interface {
	Resolve(ctx context.Context, rawChampionName, rawSkinName string, skinID int) (championdata.Resolution, error)
}

// StateUpdater is the subset of state.Manager the poller writes into.
// *state.Manager satisfies this.
type StateUpdater interface {
	UpdateChampion(championID, championName, skinName, chromaName string, skinID int)
	UpdateInGameStats(kills, deaths, assists, cs, level, gold int)
	UpdateGameStart(start int64)
	UpdateSpectatedGame(rawGameMode string, mapNumber int)
}

// Poller polls Live Client Data for one match: champion/skin once, KDA/CS/timer
// every tick, paced by the live config's StatsPollingInterval.
type Poller struct {
	client     *Client
	resolver   ChampionResolver
	state      StateUpdater
	store      *config.Store
	cfgUpdates <-chan *config.Config
	logger     zerolog.Logger
}

// NewPoller builds a Poller. The tick interval is read from store on every
// loop and reset in place when StatsPollingInterval changes.
func NewPoller(client *Client, resolver ChampionResolver, state StateUpdater, store *config.Store, logger zerolog.Logger) *Poller {
	return &Poller{
		client:     client,
		resolver:   resolver,
		state:      state,
		store:      store,
		cfgUpdates: store.Subscribe(),
		logger:     logger,
	}
}

// tickInterval is the current Live Client Data poll cadence.
func (p *Poller) tickInterval() time.Duration {
	return time.Duration(p.store.Load().Advanced.StatsPollingInterval) * time.Millisecond
}

// Run polls until ctx is canceled, routing to the TFT or non-TFT tick logic
// based on gameMode (fixed for the lifetime of one match).
func (p *Poller) Run(ctx context.Context, gameMode types.GameMode) {
	if gameMode == types.GameModeTFT {
		p.runTFT(ctx)
		return
	}
	p.runNormal(ctx, gameMode)
}

// RunSpectating polls until ctx is canceled, driving presence for a game
// being spectated rather than played. See runSpectating.
func (p *Poller) RunSpectating(ctx context.Context) {
	p.runSpectating(ctx)
}

// isLevelGoldMode reports whether a mode shows a level+gold stat line
// instead of KDA+CS, so the poller knows to also read those each tick.
func isLevelGoldMode(gameMode types.GameMode) bool {
	return gameMode == types.GameModeArena || gameMode == types.GameModeSwarm
}

// runNormal resolves the local player's riotId, then champion/skin/chroma
func (p *Poller) runNormal(ctx context.Context, gameMode types.GameMode) {
	var (
		riotID   string
		resolved bool
	)
	wantLevelGold := isLevelGoldMode(gameMode)

	tick := func() {
		if ctx.Err() != nil {
			return
		}

		if riotID == "" {
			ap, err := p.client.ActivePlayer(ctx)
			if err != nil || ap.RiotID == "" {
				return
			}
			riotID = ap.RiotID
			p.logger.Debug().Str("riotID", riotID).Msg("Live game: resolved local player")
		}

		if !resolved {
			allData, err := p.client.AllGameData(ctx)
			if err != nil {
				p.logger.Debug().Err(err).Msg("Live game: allgamedata poll failed, will retry")
			} else if player, ok := FindPlayer(allData.AllPlayers, riotID); !ok {
				p.logger.Debug().Str("riotID", riotID).Msg("Live game: local player not found yet in allgamedata")
			} else {
				res, rerr := p.resolver.Resolve(ctx, player.RawChampionName, player.RawSkinName, player.SkinID)
				if rerr != nil {
					p.logger.Debug().Err(rerr).
						Str("rawChampionName", player.RawChampionName).
						Str("rawSkinName", player.RawSkinName).
						Int("skinID", player.SkinID).
						Msg("Live game: champion/skin not resolved yet")
				} else {
					resolved = true
					p.state.UpdateChampion(res.ID, res.Name, res.SkinName, res.ChromaName, res.BaseSkinID)
					p.logger.Info().
						Str("championID", res.ID).
						Str("championName", res.Name).
						Str("skinName", res.SkinName).
						Str("chromaName", res.ChromaName).
						Int("baseSkinID", res.BaseSkinID).
						Int("rawSkinID", player.SkinID).
						Msg("Live game: champion/skin/chroma resolved")
				}
			}
		}

		scores, err := p.client.PlayerScores(ctx, riotID)
		if err != nil {
			return
		}
		stats, err := p.client.GameStats(ctx)
		if err != nil {
			return
		}

		level, gold := 0, 0
		if wantLevelGold {
			ap, aerr := p.client.ActivePlayer(ctx)
			if aerr != nil {
				return
			}
			level, gold = ap.Level, int(ap.CurrentGold)
		}

		p.state.UpdateInGameStats(scores.Kills, scores.Deaths, scores.Assists, scores.CreepScore, level, gold)
		p.state.UpdateGameStart(time.Now().Unix() - int64(stats.GameTime))
	}

	p.loop(ctx, tick)
}

// runSpectating polls allgamedata for the spectated game's mode, map and
// clock, and resolves the leading player's champion once. No local player.
func (p *Poller) runSpectating(ctx context.Context) {
	var resolved bool

	tick := func() {
		if ctx.Err() != nil {
			return
		}

		allData, err := p.client.AllGameData(ctx)
		if err != nil {
			p.logger.Debug().Err(err).Msg("Spectate: allgamedata poll failed, will retry")
			return
		}

		p.state.UpdateSpectatedGame(allData.GameData.GameMode, allData.GameData.MapNumber)
		p.state.UpdateGameStart(time.Now().Unix() - int64(allData.GameData.GameTime))

		if resolved || len(allData.AllPlayers) == 0 {
			return
		}
		first := allData.AllPlayers[0]
		res, rerr := p.resolver.Resolve(ctx, first.RawChampionName, first.RawSkinName, first.SkinID)
		if rerr != nil {
			p.logger.Debug().Err(rerr).
				Str("rawChampionName", first.RawChampionName).
				Msg("Spectate: leading player champion not resolved yet")
			return
		}
		resolved = true
		p.state.UpdateChampion(res.ID, res.Name, res.SkinName, res.ChromaName, res.BaseSkinID)
		p.logger.Info().
			Str("championID", res.ID).
			Str("championName", res.Name).
			Msg("Spectate: leading player champion/skin resolved")
	}

	p.loop(ctx, tick)
}

// runTFT independently polls activeplayer for level and gamestats for the
// timer; no riotId/allgamedata/championdata resolution is involved.
func (p *Poller) runTFT(ctx context.Context) {
	tick := func() {
		if ctx.Err() != nil {
			return
		}
		ap, err := p.client.ActivePlayer(ctx)
		if err != nil {
			return
		}
		p.state.UpdateInGameStats(0, 0, 0, 0, ap.Level, 0)

		if stats, err := p.client.GameStats(ctx); err == nil {
			p.state.UpdateGameStart(time.Now().Unix() - int64(stats.GameTime))
		}
	}

	p.loop(ctx, tick)
}

func (p *Poller) loop(ctx context.Context, tick func()) {
	interval := p.tickInterval()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	tick()
	for {
		select {
		case <-ctx.Done():
			return
		case <-p.cfgUpdates:
			if d := p.tickInterval(); d != interval {
				interval = d
				ticker.Reset(d)
			}
		case <-ticker.C:
			tick()
		}
	}
}
