package state

import (
	"time"

	"github.com/its-haze/league-rpc/pkg/types"
)

// State represents the complete application state.
type State struct {
	// Player Status
	Availability         types.Availability `json:"availability"`
	SummonerIcon         int                `json:"summoner_icon"`
	ApplicationStartTime int64              `json:"application_start_time"`

	// Game Phase
	GameMode      types.GameMode      `json:"game_mode"`
	GameFlowPhase types.GameFlowPhase `json:"gameflow_phase"`

	// Queue Information
	QueueName                string        `json:"queue_name"`
	QueueType                string        `json:"queue_type"`
	QueueID                  types.QueueID `json:"queue_id"`
	QueueDescription         string        `json:"queue_description"`
	QueueDetailedDescription string        `json:"queue_detailed_description"`
	IsRanked                 bool          `json:"is_ranked"`

	// Lobby Information
	LobbyID    string `json:"lobby_id"`
	Players    int    `json:"players"`
	MaxPlayers int    `json:"max_players"`
	IsCustom   bool   `json:"is_custom"`
	IsPractice bool   `json:"is_practice"`

	// Map Data
	MapID types.MapID `json:"map_id"`

	// Ranked Stats
	SummonerRank     types.RankedStats `json:"summoner_rank"`      // SoloQ
	SummonerRankFlex types.RankedStats `json:"summoner_rank_flex"` // Flex
	ArenaRank        types.ArenaStats  `json:"arena_rank"`
	TFTRank          types.RankedStats `json:"tft_rank"`

	// TFT Companion Data
	TFTCompanionID          int    `json:"tft_companion_id"`
	TFTCompanionIcon        string `json:"tft_companion_icon"`
	TFTCompanionName        string `json:"tft_companion_name"`
	TFTCompanionDescription string `json:"tft_companion_description"`

	// In-Game Data (populated during InProgress phase)
	ChampionID    string `json:"champion_id"`   // Data Dragon's URL-safe id, e.g. "Chogath"
	ChampionName  string `json:"champion_name"` // Data Dragon's display name, e.g. "Cho'Gath"
	SkinName      string `json:"skin_name"`
	ChromaName    string `json:"chroma_name"`
	SkinID        int    `json:"skin_id"`
	Level         int    `json:"level"`
	Gold          int    `json:"gold"`
	Kills         int    `json:"kills"`
	Deaths        int    `json:"deaths"`
	Assists       int    `json:"assists"`
	CreepScore    int    `json:"creep_score"`
	GameStartTime int64  `json:"game_start_time"` // Unix time the match actually started, from live gameTime
}

// NewState creates a new State with default values
func NewState() *State {
	return &State{
		Availability:         types.AvailabilityOnline,
		ApplicationStartTime: time.Now().Unix(),
		GameFlowPhase:        types.GameFlowNone,
		GameMode:             types.GameModeClassic,
	}
}

// Equals compares two states for equality
// Used for detecting changes before triggering RPC updates
func (s *State) Equals(other *State) bool {
	if other == nil {
		return false
	}

	// Compare all fields
	return s.Availability == other.Availability &&
		s.SummonerIcon == other.SummonerIcon &&
		s.GameMode == other.GameMode &&
		s.GameFlowPhase == other.GameFlowPhase &&
		s.QueueName == other.QueueName &&
		s.QueueType == other.QueueType &&
		s.QueueID == other.QueueID &&
		s.QueueDescription == other.QueueDescription &&
		s.QueueDetailedDescription == other.QueueDetailedDescription &&
		s.IsRanked == other.IsRanked &&
		s.LobbyID == other.LobbyID &&
		s.Players == other.Players &&
		s.MaxPlayers == other.MaxPlayers &&
		s.IsCustom == other.IsCustom &&
		s.IsPractice == other.IsPractice &&
		s.MapID == other.MapID &&
		s.SummonerRank == other.SummonerRank &&
		s.SummonerRankFlex == other.SummonerRankFlex &&
		s.ArenaRank == other.ArenaRank &&
		s.TFTRank == other.TFTRank &&
		s.TFTCompanionID == other.TFTCompanionID &&
		s.TFTCompanionIcon == other.TFTCompanionIcon &&
		s.TFTCompanionName == other.TFTCompanionName &&
		s.TFTCompanionDescription == other.TFTCompanionDescription &&
		s.ChampionID == other.ChampionID &&
		s.ChampionName == other.ChampionName &&
		s.SkinName == other.SkinName &&
		s.ChromaName == other.ChromaName &&
		s.SkinID == other.SkinID &&
		s.Level == other.Level &&
		s.Gold == other.Gold &&
		s.Kills == other.Kills &&
		s.Deaths == other.Deaths &&
		s.Assists == other.Assists &&
		s.CreepScore == other.CreepScore &&
		s.GameStartTime == other.GameStartTime
}

// Copy creates a deep copy of the state
func (s *State) Copy() *State {
	if s == nil {
		return nil
	}

	copy := *s
	return &copy
}

// GetQueueName prefers the detailed queue description, falling back to the
// plain queue name; callers add their own final fallback if both are empty.
func (s *State) GetQueueName() string {
	if s.QueueDetailedDescription != "" {
		return s.QueueDetailedDescription
	}
	return s.QueueName
}
