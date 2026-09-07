package lcu

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	lcu "github.com/its-haze/lcu-gopher"
	"github.com/its-haze/league-rpc/internal/discord"
	"github.com/its-haze/league-rpc/internal/state"
	"github.com/its-haze/league-rpc/pkg/constants"
	"github.com/its-haze/league-rpc/pkg/types"
)

// gatherInitialData fetches all initial state from LCU when first connecting.
func (c *Client) gatherInitialData() error {
	c.logger.Debug().Msg("Gathering initial data from League Client...")

	// Get when the League client itself started, so the elapsed timer counts
	// from client launch rather than daemon start.
	if err := c.fetchApplicationStartTime(); err != nil {
		c.logger.Debug().Err(err).Msg("Failed to fetch application start time")
	}

	// Get current summoner info
	if err := c.fetchSummonerInfo(); err != nil {
		c.logger.Warn().Err(err).Msg("Failed to fetch summoner info")
	}

	// Get chat status (online/away)
	if err := c.fetchChatStatus(); err != nil {
		c.logger.Warn().Err(err).Msg("Failed to fetch chat status")
	}

	// Get TFT companion data
	if err := c.fetchTFTCompanion(); err != nil {
		c.logger.Debug().Err(err).Msg("Failed to fetch TFT companion (may not be selected)")
	}

	// Get ranked stats
	if err := c.fetchRankedStats(); err != nil {
		c.logger.Warn().Err(err).Msg("Failed to fetch ranked stats")
	}

	// Get current gameflow phase
	if err := c.fetchGameFlowPhase(); err != nil {
		c.logger.Warn().Err(err).Msg("Failed to fetch gameflow phase")
	}

	// /lol-lobby/v2/lobby only exists for a party lobby; fall back to the
	// gameflow metadata endpoint, then to Live Client Data, when it 404s.
	if err := c.fetchLobbyInfo(); err != nil {
		c.logger.Debug().Err(err).Msg("No party lobby, trying gameflow metadata")
		if err := c.fetchLobbyFromGameflowMetadata(); err != nil {
			c.logger.Debug().Err(err).Msg("No gameflow lobby status either")
			c.recoverGameModeFromLiveClientData()
		}
	}

	c.logger.Debug().Msg("Initial data gathering complete")
	return nil
}

// fetchApplicationStartTime records the epoch time the League client
// process started, so the presence timer counts from client launch.
func (c *Client) fetchApplicationStartTime() error {
	resp, err := c.lcu.Get(constants.EndpointApplicationStartTime)
	if err != nil {
		return fmt.Errorf("get application start time: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("application start time status %d", resp.StatusCode)
	}

	var startTime int64
	if err := json.NewDecoder(resp.Body).Decode(&startTime); err != nil {
		return fmt.Errorf("decode application start time: %w", err)
	}
	if startTime <= 0 {
		return fmt.Errorf("application start time not a positive epoch: %d", startTime)
	}

	c.state.UpdateApplicationStartTime(startTime)
	c.logger.Debug().Int64("start_time", startTime).Msg("Updated application start time")
	return nil
}

// fetchLobbyFromGameflowMetadata reads the current lobby's queue from the
// gameflow metadata endpoint, which stays populated past the lobby phase.
func (c *Client) fetchLobbyFromGameflowMetadata() error {
	resp, err := c.lcu.Get(constants.EndpointGameflowMetadata)
	if err != nil {
		return fmt.Errorf("get gameflow metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gameflow metadata status %d", resp.StatusCode)
	}

	// memberSummonerIds is decoded as raw elements: only its length matters,
	// and the element type has changed across client versions.
	var meta struct {
		CurrentLobbyStatus *struct {
			QueueID           int               `json:"queueId"`
			LobbyID           string            `json:"lobbyId"`
			MemberSummonerIds []json.RawMessage `json:"memberSummonerIds"`
			IsCustom          bool              `json:"isCustom"`
			IsPracticeTool    bool              `json:"isPracticeTool"`
		} `json:"currentLobbyStatus"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return fmt.Errorf("decode gameflow metadata: %w", err)
	}
	if meta.CurrentLobbyStatus == nil {
		return fmt.Errorf("no currentLobbyStatus in gameflow metadata")
	}

	ls := meta.CurrentLobbyStatus
	c.resolveLobbyFromMetadata(ls.LobbyID, ls.QueueID, len(ls.MemberSummonerIds), ls.IsCustom, ls.IsPracticeTool)
	return nil
}

// resolveLobbyFromMetadata writes lobby/queue state, filling name/map/mode
// from custom-lobby defaults or /lol-game-queues, like updateLobbyState.
func (c *Client) resolveLobbyFromMetadata(lobbyID string, queueID, players int, isCustom, isPractice bool) {
	if isCustomOrPracticeQueue(queueID, isCustom, isPractice) {
		name, mode, mapID, maxPlayers := customLobbyDefaults(queueID, isPractice)
		if name == "Practice Tool" {
			isPractice = true
		} else {
			isCustom = true
		}
		c.state.UpdateField(func(s *state.State) {
			s.LobbyID = lobbyID
			s.QueueID = types.QueueID(queueID)
			s.Players = players
			if maxPlayers > 0 {
				s.MaxPlayers = maxPlayers
			}
			s.IsCustom = isCustom
			s.IsPractice = isPractice
			s.QueueName = name
			s.QueueDetailedDescription = ""
			s.GameMode = mode
			s.MapID = mapID
		})
		return
	}

	info, err := c.fetchQueueInfo(queueID)
	if err != nil {
		c.logger.Warn().Err(err).Int("queue_id", queueID).Msg("Failed to fetch queue info from metadata path, keeping what we know")
		c.state.UpdateField(func(s *state.State) {
			s.LobbyID = lobbyID
			s.QueueID = types.QueueID(queueID)
			s.Players = players
			s.IsCustom = isCustom
			s.IsPractice = isPractice
		})
		return
	}

	c.state.UpdateField(func(s *state.State) {
		s.LobbyID = lobbyID
		s.QueueID = types.QueueID(queueID)
		s.Players = players
		s.IsCustom = isCustom
		s.IsPractice = isPractice
		s.QueueName = info.Name
		s.QueueType = info.Type
		s.QueueDescription = info.Description
		s.QueueDetailedDescription = info.DetailedDescription
		s.IsRanked = info.IsRanked
		if info.MaxPlayers > 0 {
			s.MaxPlayers = info.MaxPlayers
		}
		if info.MapID != 0 {
			s.MapID = types.MapID(info.MapID)
		}
		if info.GameMode != "" {
			s.GameMode = types.GameMode(info.GameMode)
		}
	})
}

// customLobbyDefaults returns name/mode/map/maxPlayers for a custom or
// practice lobby, matching the reference's gather_base_data special-casing.
func customLobbyDefaults(queueID int, isPractice bool) (name string, mode types.GameMode, mapID types.MapID, maxPlayers int) {
	switch {
	case queueID == queueIDPracticeTool || isPractice:
		return "Practice Tool", "PRACTICETOOL", types.MapSummonersRift, 1
	case isARAMCustomQueue(queueID):
		return "Custom ARAM", "ARAM", types.MapHowlingAbyss, 0
	default:
		return "Custom Game", "PRACTICETOOL", types.MapSummonersRift, 0
	}
}

// recoverGameModeFromLiveClientData sets GameMode (and, for Practice Tool,
func (c *Client) recoverGameModeFromLiveClientData() {
	if c.liveGame == nil || c.state.Get().GameFlowPhase != types.GameFlowInProgress {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rawMode, err := c.liveGame.GameMode(ctx)
	if err != nil || rawMode == "" {
		c.logger.Debug().Err(err).Msg("Could not recover game mode from Live Client Data")
		return
	}

	isPractice := rawMode == "PRACTICETOOL"

	c.state.UpdateField(func(s *state.State) {
		s.GameMode = types.GameMode(rawMode)
		if isPractice {
			s.QueueName = "Practice Tool"
			s.IsPractice = true
		}
	})

	c.logger.Info().Str("game_mode", rawMode).Bool("is_practice", isPractice).
		Msg("Recovered game mode from Live Client Data (no lobby available)")
}

// fetchSummonerInfo gets the current summoner's profile
func (c *Client) fetchSummonerInfo() error {
	summoner, err := c.lcu.GetCurrentSummoner()
	if err != nil {
		return fmt.Errorf("get current summoner: %w", err)
	}

	c.state.UpdateSummonerIcon(summoner.ProfileIconID)
	c.logger.Debug().Int("icon_id", summoner.ProfileIconID).Msg("Updated summoner icon")

	return nil
}

// fetchChatStatus gets the current online/away status
func (c *Client) fetchChatStatus() error {
	resp, err := c.lcu.Get(constants.EndpointChatStatus)
	if err != nil {
		return fmt.Errorf("get chat status: %w", err)
	}
	defer resp.Body.Close()

	var chatStatus struct {
		Availability string `json:"availability"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&chatStatus); err != nil {
		return fmt.Errorf("decode chat status: %w", err)
	}

	c.state.UpdateAvailability(chatStatus.Availability)
	c.logger.Debug().Str("availability", chatStatus.Availability).Msg("Updated availability")

	return nil
}

// fetchTFTCompanion gets the selected TFT companion
func (c *Client) fetchTFTCompanion() error {
	resp, err := c.lcu.Get(constants.EndpointTFTCompanion)
	if err != nil {
		return fmt.Errorf("get TFT companion: %w", err)
	}
	defer resp.Body.Close()

	var companionData struct {
		SelectedLoadoutItem struct {
			ItemID       int    `json:"itemId"`
			LoadoutsIcon string `json:"loadoutsIcon"`
			Name         string `json:"name"`
			Description  string `json:"description"`
		} `json:"selectedLoadoutItem"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&companionData); err != nil {
		return fmt.Errorf("decode TFT companion: %w", err)
	}

	item := companionData.SelectedLoadoutItem

	// Extract icon path from full path
	// Example: "ASSETS/Characters/TFT/..." → "characters/tft/..."
	fullPath := item.LoadoutsIcon
	parts := strings.Split(fullPath, "ASSETS/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid icon path format: %s", fullPath)
	}
	finalPath := strings.ToLower(parts[1])

	iconURL := discord.GetTFTCompanionURL(finalPath)

	c.state.UpdateTFTCompanion(item.ItemID, iconURL, item.Name, item.Description)
	c.logger.Debug().Str("name", item.Name).Msg("Updated TFT companion")

	return nil
}

// fetchRankedStats gets the current ranked statistics
func (c *Client) fetchRankedStats() error {
	resp, err := c.lcu.Get(constants.EndpointRankedStats)
	if err != nil {
		return fmt.Errorf("get ranked stats: %w", err)
	}
	defer resp.Body.Close()

	var rankedData struct {
		Queues []struct {
			QueueType    string `json:"queueType"`
			Tier         string `json:"tier"`
			Division     string `json:"division"`
			LeaguePoints int    `json:"leaguePoints"`
			RatedTier    string `json:"ratedTier"`   // For Arena
			RatedRating  int    `json:"ratedRating"` // For Arena
		} `json:"queues"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rankedData); err != nil {
		return fmt.Errorf("decode ranked stats: %w", err)
	}

	// Parse ranked stats for each queue
	currentState := c.state.Get()
	for _, queue := range rankedData.Queues {
		switch queue.QueueType {
		case "RANKED_SOLO_5x5":
			currentState.SummonerRank = types.RankedStats{
				Tier:         types.Tier(queue.Tier),
				Division:     types.Division(queue.Division),
				LeaguePoints: queue.LeaguePoints,
			}

		case "RANKED_FLEX_SR":
			currentState.SummonerRankFlex = types.RankedStats{
				Tier:         types.Tier(queue.Tier),
				Division:     types.Division(queue.Division),
				LeaguePoints: queue.LeaguePoints,
			}

		case "RANKED_TFT":
			currentState.TFTRank = types.RankedStats{
				Tier:         types.Tier(queue.Tier),
				Division:     types.Division(queue.Division),
				LeaguePoints: queue.LeaguePoints,
			}

		case "CHERRY":
			currentState.ArenaRank = types.ArenaStats{
				RatedTier:   types.ArenaRatedTier(queue.RatedTier),
				Tier:        types.MapRatedTierToTier(types.ArenaRatedTier(queue.RatedTier)),
				RatedRating: queue.RatedRating,
			}
		}
	}

	c.state.Update(currentState)
	c.logger.Debug().Msg("Updated ranked stats")

	return nil
}

// fetchGameFlowPhase gets the current game flow phase
func (c *Client) fetchGameFlowPhase() error {
	resp, err := c.lcu.Get(constants.EndpointGameflow)
	if err != nil {
		return fmt.Errorf("get gameflow phase: %w", err)
	}
	defer resp.Body.Close()

	var phase string
	if err := json.NewDecoder(resp.Body).Decode(&phase); err != nil {
		return fmt.Errorf("decode gameflow phase: %w", err)
	}

	c.state.UpdateGameFlowPhase(phase)
	c.logger.Debug().Str("phase", phase).Msg("Updated gameflow phase")

	return nil
}

// fetchLobbyInfo gets current lobby information (if in lobby). The LCU
func (c *Client) fetchLobbyInfo() error {
	resp, err := c.lcu.Get(constants.EndpointLobby)
	if err != nil {
		return fmt.Errorf("get lobby: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("not in a lobby (status %d)", resp.StatusCode)
	}

	var lobbyData lcu.Lobby
	if err := json.NewDecoder(resp.Body).Decode(&lobbyData); err != nil {
		return fmt.Errorf("decode lobby: %w", err)
	}

	// Update lobby state
	c.updateLobbyState(&lobbyData)

	return nil
}
