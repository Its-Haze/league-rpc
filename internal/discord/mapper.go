package discord

import (
	"github.com/its-haze/league-rpc/internal/config"
	"github.com/its-haze/league-rpc/internal/state"
	"github.com/its-haze/league-rpc/pkg/types"
)

// MapStateToPresence converts application state to Discord RPC data
// This is the main routing function that determines which presence builder to use
func MapStateToPresence(st *state.State, cfg *config.Config) *RPCData {
	if st == nil {
		return &RPCData{}
	}

	// Route to appropriate builder based on game flow phase
	switch st.GameFlowPhase {
	case types.GameFlowInProgress:
		// In game - check if TFT or regular game
		if st.GameMode == types.GameModeTFT {
			return BuildTFTInGamePresence(st, cfg)
		}
		// Champion/skin/chroma resolve once per match and ChampionID is
		if st.ChampionID == "" {
			return BuildInChampSelectPresence(st, cfg)
		}
		return BuildInGamePresence(st, cfg)

	case types.GameFlowWatching:
		// Spectating someone else's game.
		return BuildSpectatingPresence(st, cfg)

	case types.GameFlowChampSelect, types.GameFlowGameStart:
		// Champion selection phase
		return BuildInChampSelectPresence(st, cfg)

	case types.GameFlowMatchmaking, types.GameFlowCheckedIntoTournament:
		// In queue
		return BuildInQueuePresence(st, cfg)

	case types.GameFlowReadyCheck:
		// Ready check - keep showing queue state
		return BuildInQueuePresence(st, cfg)

	case types.GameFlowLobby:
		// In lobby - check if custom game / practice tool or matchmaking
		if st.IsCustom || st.IsPractice {
			return BuildInCustomLobbyPresence(st, cfg)
		}
		return BuildInLobbyPresence(st, cfg)

	case types.GameFlowNone, types.GameFlowWaitingForStats,
		types.GameFlowPreEndOfGame, types.GameFlowEndOfGame:
		// Idle in client or post-game
		return BuildInClientPresence(st, cfg)

	default:
		// Unknown phase - default to in client
		return BuildInClientPresence(st, cfg)
	}
}

// ShouldClearPresence returns true if presence should be cleared instead of updated
// This happens when the user has --hide-in-client enabled and is idle
func ShouldClearPresence(st *state.State, cfg *config.Config) bool {
	if !cfg.Presence.ShowInClient && st.GameFlowPhase.IsInClient() {
		return true
	}
	return false
}
