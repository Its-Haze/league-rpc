package discord

import (
	"testing"
	"time"

	"github.com/its-haze/league-rpc/internal/config"
	"github.com/its-haze/league-rpc/internal/state"
	"github.com/its-haze/league-rpc/pkg/constants"
	"github.com/its-haze/league-rpc/pkg/types"
)

func TestFormatSkinName(t *testing.T) {
	tests := []struct {
		name                                     string
		championName, skinName, chromaName, want string
	}{
		{"skin found: skin name alone, no champion prefix", "Qiyana", "Battle Boss Qiyana", "", "Battle Boss Qiyana"},
		{"chroma found: skin name plus chroma, no champion prefix", "Chogath", "Battlecast Cho'Gath", "Emerald", "Battlecast Cho'Gath (Emerald)"},
		{"default skin: falls back to champion name", "Ahri", "", "", "Ahri"},
		{"skin literally \"default\": falls back to champion name", "Ahri", "default", "", "Ahri"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatSkinName(tt.championName, tt.skinName, tt.chromaName); got != tt.want {
				t.Errorf("FormatSkinName(%q, %q, %q) = %q, want %q", tt.championName, tt.skinName, tt.chromaName, got, tt.want)
			}
		})
	}
}

func TestFormatGameModeName(t *testing.T) {
	tests := []struct {
		mode types.GameMode
		want string
	}{
		{types.GameModeClassic, "Summoner's Rift"},
		{types.GameModeARAM, "Howling Abyss (ARAM)"},
		{types.GameModeBrawl, "Brawl"},
		{types.GameModeArena, "Arena"},
		{types.GameModeSwarm, "Swarm"},
		{types.GameModeUltimateSpellbook, "Ultimate Spellbook"},
		{"SOME_UNKNOWN_MODE", "SOME_UNKNOWN_MODE"},
	}
	for _, tt := range tests {
		if got := FormatGameModeName(tt.mode); got != tt.want {
			t.Errorf("FormatGameModeName(%q) = %q, want %q", tt.mode, got, tt.want)
		}
	}
}

func TestQueueDisplayName(t *testing.T) {
	tests := []struct {
		name string
		st   *state.State
		want string
	}{
		{
			name: "detailed description wins",
			st: &state.State{
				QueueDetailedDescription: "Ranked Solo/Duo",
				QueueName:                "5v5 Ranked Solo games",
			},
			want: "Ranked Solo/Duo",
		},
		{
			name: "falls back to queue name",
			st:   &state.State{QueueName: "Custom Game"},
			want: "Custom Game",
		},
		{
			name: "falls back to game mode display name",
			st:   &state.State{GameMode: types.GameModeARAM},
			want: "Howling Abyss (ARAM)",
		},
		{
			name: "final catch-all",
			st:   &state.State{},
			want: "League of Legends",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := queueDisplayName(tt.st); got != tt.want {
				t.Errorf("queueDisplayName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildInGamePresence_StatePrefixedWithInGame(t *testing.T) {
	cfg := config.DefaultConfig()
	st := state.NewState()
	st.Kills, st.Deaths, st.Assists, st.CreepScore = 3, 2, 5, 120

	cfg.Display.Default.ShowStats = true
	got := BuildInGamePresence(st, cfg)
	if want := "In Game · 3/2/5 · 120cs"; got.State != want {
		t.Errorf("State (stats shown) = %q, want %q", got.State, want)
	}

	cfg.Display.Default.ShowStats = false
	got = BuildInGamePresence(st, cfg)
	if want := "In Game"; got.State != want {
		t.Errorf("State (stats hidden) = %q, want %q", got.State, want)
	}
}

func TestBuildInGamePresence_ArenaAndSwarmShowLevelAndGold(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Display.Default.ShowStats = true

	arena := state.NewState()
	arena.GameMode = types.GameModeArena
	arena.Kills, arena.Deaths, arena.Assists = 4, 1, 7
	arena.Level, arena.Gold = 14, 8200
	if got := BuildInGamePresence(arena, cfg).State; got != "In Game · 4/1/7 · lvl: 14 · gold: 8200" {
		t.Errorf("Arena State = %q", got)
	}

	swarm := state.NewState()
	swarm.GameMode = types.GameModeSwarm
	swarm.CreepScore = 240
	swarm.Level, swarm.Gold = 9, 1500
	if got := BuildInGamePresence(swarm, cfg).State; got != "In Game · 240cs · lvl: 9 · gold: 1500" {
		t.Errorf("Swarm State = %q", got)
	}

	// A normal mode still shows KDA + CS, unchanged.
	sr := state.NewState()
	sr.GameMode = types.GameModeClassic
	sr.Kills, sr.Deaths, sr.Assists, sr.CreepScore = 3, 2, 5, 120
	if got := BuildInGamePresence(sr, cfg).State; got != "In Game · 3/2/5 · 120cs" {
		t.Errorf("Classic State = %q", got)
	}
}

func TestBuildSpectatingPresence(t *testing.T) {
	cfg := config.DefaultConfig()

	t.Run("champion resolved: skin tile as large image", func(t *testing.T) {
		st := state.NewState()
		st.GameFlowPhase = types.GameFlowWatching
		st.GameMode = types.GameModeARAM
		st.ChampionID = "Chogath"
		st.SkinID = 1
		st.GameStartTime = 5000

		got := BuildSpectatingPresence(st, cfg)
		if got.LargeText != "Spectating" || got.State != "Spectating" {
			t.Errorf("LargeText/State = %q/%q, want Spectating/Spectating", got.LargeText, got.State)
		}
		if got.Details != "Howling Abyss (ARAM)" {
			t.Errorf("Details = %q, want the game mode display name", got.Details)
		}
		if got.LargeImage != GetChampionSkinURL("Chogath", 1) {
			t.Errorf("LargeImage = %q, want the champion skin tile", got.LargeImage)
		}
		if got.Start != 5000 {
			t.Errorf("Start = %d, want 5000 (from GameStartTime)", got.Start)
		}
	})

	t.Run("champion unresolved: map icon as large image", func(t *testing.T) {
		st := state.NewState()
		st.GameFlowPhase = types.GameFlowWatching
		st.GameMode = types.GameModeARAM
		st.MapID = types.MapHowlingAbyss

		got := BuildSpectatingPresence(st, cfg)
		if got.LargeImage != GetMapIconURL(types.MapHowlingAbyss) {
			t.Errorf("LargeImage = %q, want the map icon", got.LargeImage)
		}
	})

	t.Run("nothing resolved: league logo as large image", func(t *testing.T) {
		st := state.NewState()
		st.GameFlowPhase = types.GameFlowWatching
		st.GameMode = ""
		st.MapID = 0

		got := BuildSpectatingPresence(st, cfg)
		if got.LargeImage != GetLeagueLogoLargeURL() {
			t.Errorf("LargeImage = %q, want the league logo", got.LargeImage)
		}
		if got.Details == "" {
			t.Error("Details must never be empty (Discord rejects it)")
		}
	})
}

func TestMapStateToPresence_WatchingRoutesToSpectating(t *testing.T) {
	cfg := config.DefaultConfig()
	st := state.NewState()
	st.GameFlowPhase = types.GameFlowWatching
	st.GameMode = types.GameModeClassic

	got := MapStateToPresence(st, cfg)
	if got.State != "Spectating" {
		t.Errorf("State = %q, want Spectating", got.State)
	}
}

func TestBuildTFTInGamePresence_UsesResolvedGameStartTime(t *testing.T) {
	cfg := config.DefaultConfig()
	st := state.NewState()
	st.GameStartTime = 1000

	got := BuildTFTInGamePresence(st, cfg)
	if got.Start != 1000 {
		t.Errorf("Start = %d, want 1000 (the poller-supplied GameStartTime)", got.Start)
	}
}

func TestBuildTFTInGamePresence_FallsBackToNowWhenGameStartTimeUnresolved(t *testing.T) {
	cfg := config.DefaultConfig()
	st := state.NewState()
	st.GameStartTime = 0

	before := time.Now().Unix()
	got := BuildTFTInGamePresence(st, cfg)
	if got.Start < before {
		t.Errorf("Start = %d, want >= %d (fallback to time.Now())", got.Start, before)
	}
}

func TestBuildInCustomLobbyPresence_DetailsIsQueueNameStateIsFixed(t *testing.T) {
	cfg := config.DefaultConfig()
	st := state.NewState()
	st.QueueName = "Practice Tool"
	st.IsPractice = true
	st.Players, st.MaxPlayers = 1, 1

	got := BuildInCustomLobbyPresence(st, cfg)
	if got.Details != "Practice Tool" {
		t.Errorf("Details = %q, want %q", got.Details, "Practice Tool")
	}
	// Default state stays the fixed "In Lobby", unlike matchmaking lobbies:
	// a solo practice tool session showing "(1/1)" isn't useful information.
	if want := "In Lobby"; got.State != want {
		t.Errorf("State = %q, want %q", got.State, want)
	}
}

func TestBuildInLobbyPresence_StateShowsPlayerCount(t *testing.T) {
	cfg := config.DefaultConfig()
	st := state.NewState()
	st.QueueID = types.QueueSoloQ
	st.Players, st.MaxPlayers = 3, 5

	got := BuildInLobbyPresence(st, cfg)
	if want := "In Lobby (3/5)"; got.State != want {
		t.Errorf("State = %q, want %q", got.State, want)
	}
}

func TestBuildInQueuePresence_StateIsPlainInQueueRegardlessOfClash(t *testing.T) {
	cfg := config.DefaultConfig()
	st := state.NewState()
	st.GameFlowPhase = types.GameFlowCheckedIntoTournament

	got := BuildInQueuePresence(st, cfg)
	if want := "In Queue"; got.State != want {
		t.Errorf("State = %q, want %q (Clash no longer gets a special-cased state)", got.State, want)
	}
}

func TestBuildTFTInGamePresence_StateShowsLevel(t *testing.T) {
	cfg := config.DefaultConfig()
	st := state.NewState()
	st.Level = 7

	got := BuildTFTInGamePresence(st, cfg)
	if want := "In Game · lvl: 7"; got.State != want {
		t.Errorf("State = %q, want %q", got.State, want)
	}
}

func TestBuildInLobbyPresence_RankKnownSwapsLargeTextToCredit(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Display.Default.ShowRank = true
	st := state.NewState()
	st.QueueID = types.QueueSoloQ
	st.SummonerRank = types.RankedStats{Tier: types.TierGold, Division: types.DivisionIV, LeaguePoints: 40}

	got := BuildInLobbyPresence(st, cfg)
	if got.LargeText != constants.SmallText {
		t.Errorf("LargeText = %q, want the credit line %q", got.LargeText, constants.SmallText)
	}
	if got.SmallText != "Gold IV: 40 LP" {
		t.Errorf("SmallText = %q, want %q", got.SmallText, "Gold IV: 40 LP")
	}
}

func TestMapStateToPresence_InProgressWithUnresolvedChampionFallsBackToChampSelect(t *testing.T) {
	cfg := config.DefaultConfig()
	st := state.NewState()
	st.GameFlowPhase = types.GameFlowInProgress
	st.GameMode = types.GameModeClassic
	// ChampionID left empty: this match's poller hasn't resolved a champion
	// yet (e.g. right after entering InProgress clears the previous match's data).

	got := MapStateToPresence(st, cfg)
	want := BuildInChampSelectPresence(st, cfg)
	if got.State != want.State || got.Details != want.Details {
		t.Errorf("MapStateToPresence() = %+v, want champ-select presence %+v", got, want)
	}
}

func TestMapStateToPresence_InProgressWithResolvedChampionUsesInGamePresence(t *testing.T) {
	cfg := config.DefaultConfig()
	st := state.NewState()
	st.GameFlowPhase = types.GameFlowInProgress
	st.GameMode = types.GameModeClassic
	st.ChampionID = "Chogath"
	st.ChampionName = "Cho'Gath"

	got := MapStateToPresence(st, cfg)
	if got.LargeText != "Cho'Gath" {
		t.Errorf("LargeText = %q, want the resolved champion's display name", got.LargeText)
	}
}

func TestBuildInGamePresence_RankKnownDoesNotSwapLargeText(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Display.Default.ShowRank = true
	st := state.NewState()
	st.ChampionName = "Ahri"
	st.QueueID = types.QueueSoloQ
	st.SummonerRank = types.RankedStats{Tier: types.TierGold, Division: types.DivisionIV, LeaguePoints: 40}

	got := BuildInGamePresence(st, cfg)
	if got.LargeText == constants.SmallText {
		t.Errorf("LargeText should stay the skin name, not the credit line, got %q", got.LargeText)
	}
}
