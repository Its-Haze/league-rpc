package lcu

import (
	"encoding/json"
	"fmt"
	"strings"

	lcu "github.com/its-haze/lcu-gopher"
	"github.com/its-haze/league-rpc/internal/discord"
	"github.com/its-haze/league-rpc/internal/state"
	"github.com/its-haze/league-rpc/pkg/constants"
	"github.com/its-haze/league-rpc/pkg/types"
)

// subscribeToEvents subscribes to all LCU WebSocket events
func (c *Client) subscribeToEvents() error {
	c.logger.Info().Msg("Subscribing to LCU events...")

	// Subscribe to summoner icon changes
	if err := c.lcu.Subscribe(constants.EndpointSummoner, c.handleSummonerUpdate, "Update"); err != nil {
		return fmt.Errorf("subscribe to summoner: %w", err)
	}

	// Subscribe to chat status changes (online/away)
	if err := c.lcu.Subscribe(constants.EndpointChatStatus, c.handleChatStatusUpdate, "Update"); err != nil {
		return fmt.Errorf("subscribe to chat status: %w", err)
	}

	// Subscribe to gameflow phase changes (most important!)
	if err := c.lcu.Subscribe(constants.EndpointGameflow, c.handleGameFlowPhaseUpdate, "Update"); err != nil {
		return fmt.Errorf("subscribe to gameflow: %w", err)
	}

	// Subscribe to lobby events (create, update, delete)
	if err := c.lcu.Subscribe(constants.EndpointLobby, c.handleLobbyUpdate, "Create", "Update", "Delete"); err != nil {
		return fmt.Errorf("subscribe to lobby: %w", err)
	}

	// Subscribe to ranked stats changes
	if err := c.lcu.Subscribe(constants.EndpointRankedStats, c.handleRankedStatsUpdate, "Update"); err != nil {
		return fmt.Errorf("subscribe to ranked stats: %w", err)
	}

	// Subscribe to TFT companion changes
	if err := c.lcu.Subscribe(constants.EndpointTFTCompanion, c.handleTFTCompanionUpdate, "Update"); err != nil {
		return fmt.Errorf("subscribe to TFT companion: %w", err)
	}

	c.logger.Info().Msg("Successfully subscribed to all LCU events")
	return nil
}

// handleSummonerUpdate handles updates to the summoner profile
func (c *Client) handleSummonerUpdate(event *lcu.Event) {
	c.logger.Debug().Msg("Summoner update received")

	data, ok := event.Data.(map[string]interface{})
	if !ok {
		c.logger.Warn().Msg("Invalid summoner data format")
		return
	}

	// Extract profile icon ID
	iconID, ok := data["profileIconId"].(float64)
	if !ok {
		c.logger.Warn().Msg("Profile icon ID not found in summoner data")
		return
	}

	currentState := c.state.Get()
	if currentState.SummonerIcon == int(iconID) {
		c.logger.Debug().Msg("Summoner icon unchanged, skipping update")
		return
	}

	c.state.UpdateSummonerIcon(int(iconID))
	c.logger.Info().Int("icon_id", int(iconID)).Msg("Summoner icon updated")
}

// handleChatStatusUpdate handles online/away status changes
func (c *Client) handleChatStatusUpdate(event *lcu.Event) {
	c.logger.Debug().Msg("Chat status update received")

	data, ok := event.Data.(map[string]interface{})
	if !ok {
		c.logger.Warn().Msg("Invalid chat status data format")
		return
	}

	// Extract availability
	availability, ok := data["availability"].(string)
	if !ok {
		c.logger.Warn().Msg("Availability not found in chat status data")
		return
	}

	// Ignore DND status
	if availability == "dnd" {
		c.logger.Debug().Msg("Ignoring DND status")
		return
	}

	currentState := c.state.Get()
	if string(currentState.Availability) == availability {
		c.logger.Debug().Msg("Availability unchanged, skipping update")
		return
	}

	c.state.UpdateAvailability(availability)
	c.logger.Debug().Str("availability", availability).Msg("Availability updated")
}

// handleGameFlowPhaseUpdate handles game flow phase transitions
// This is the MOST IMPORTANT event as it drives the entire RPC state machine
func (c *Client) handleGameFlowPhaseUpdate(event *lcu.Event) {
	phase, ok := event.Data.(string)
	if !ok {
		c.logger.Warn().Msg("Invalid gameflow phase data format")
		return
	}

	c.logger.Info().Str("phase", phase).Msg("GameFlow phase changed")

	currentState := c.state.Get()
	if string(currentState.GameFlowPhase) == phase {
		c.logger.Debug().Msg("GameFlow phase unchanged, skipping update")
		return
	}

	c.state.UpdateGameFlowPhase(phase)

	// Warn on phases that indicate a game-side problem; the ordinary
	// transitions are already covered by the "GameFlow phase changed" log above.
	switch types.GameFlowPhase(phase) {
	case types.GameFlowFailedToLaunch:
		c.logger.Warn().Msg("League failed to launch - this is a game issue, not LeagueRPC")
	case types.GameFlowTerminatedInError:
		c.logger.Warn().Msg("Game terminated in error - this is a League client issue")
	}
}

// handleLobbyUpdate handles lobby create/update/delete events
func (c *Client) handleLobbyUpdate(event *lcu.Event) {
	c.logger.Debug().Str("event_type", event.EventType).Msg("Lobby event received")

	// Handle lobby deletion
	if event.EventType == "Delete" {
		c.logger.Info().Msg("Left lobby")
		// Clear lobby data
		c.state.UpdateLobby("", 0, 0, false, false)
		return
	}

	// Parse lobby data
	data, ok := event.Data.(map[string]interface{})
	if !ok {
		c.logger.Warn().Msg("Invalid lobby data format")
		return
	}

	// Convert to lcu.Lobby struct for easier parsing
	jsonData, _ := json.Marshal(data)
	var lobbyData lcu.Lobby
	if err := json.Unmarshal(jsonData, &lobbyData); err != nil {
		c.logger.Warn().Err(err).Msg("Failed to parse lobby data")
		return
	}

	c.updateLobbyState(&lobbyData)
	c.logger.Debug().Str("lobby_id", lobbyData.PartyId).Int("players", len(lobbyData.Members)).Msg("Lobby updated")
}

// Custom/practice game queue IDs, matching the reference implementation.
const (
	queueIDCustomGame           = 3100
	queueIDCustomGameDraft      = 3110
	queueIDCustomGameARAM       = 3120
	queueIDCustomGameTournament = 3130
	queueIDPracticeTool         = 3140
	queueIDARAMCustomBlind      = 3200
	queueIDARAMCustomDraft      = 3210
	queueIDARAMCustomAllRandom  = 3220
	queueIDARAMCustomTournament = 3230
	queueIDARAMCustomMayhem     = 3270
)

// isCustomOrPracticeQueue reports whether queueID (or the lobby's own custom/
// practice flags) means "skip the /lol-game-queues lookup" per the reference.
func isCustomOrPracticeQueue(queueID int, isCustom, isPractice bool) bool {
	switch queueID {
	case queueIDCustomGame, queueIDCustomGameDraft, queueIDCustomGameARAM, queueIDCustomGameTournament,
		queueIDPracticeTool, queueIDARAMCustomBlind, queueIDARAMCustomDraft, queueIDARAMCustomAllRandom,
		queueIDARAMCustomTournament, queueIDARAMCustomMayhem:
		return true
	default:
		return isCustom || isPractice
	}
}

func isARAMCustomQueue(queueID int) bool {
	switch queueID {
	case queueIDARAMCustomBlind, queueIDARAMCustomDraft, queueIDARAMCustomAllRandom,
		queueIDARAMCustomTournament, queueIDARAMCustomMayhem:
		return true
	default:
		return false
	}
}

// updateLobbyState updates the state manager with lobby information.
func (c *Client) updateLobbyState(lobby *lcu.Lobby) {
	lobbyID := lobby.PartyId
	players := len(lobby.Members)
	maxPlayers := lobby.GameConfig.MaxLobbySize
	queueID := lobby.GameConfig.QueueId
	gameMode := types.GameMode(lobby.GameConfig.GameMode)
	mapID := types.MapID(lobby.GameConfig.MapId)

	// isCustom comes straight from the API field, not a heuristic; practice
	// tool is the one gameMode that isn't otherwise flagged as custom.
	isCustom := lobby.GameConfig.IsCustom
	isPractice := false
	if lobby.GameConfig.GameMode == "PRACTICETOOL" {
		isPractice = true
		maxPlayers = 1
	}

	var (
		queueName, queueType, queueDescription, queueDetailedDescription string
		isRanked                                                         bool
		queueInfoResolved                                                bool
	)

	switch {
	case isCustomOrPracticeQueue(queueID, isCustom, isPractice):
		switch {
		case queueID == queueIDPracticeTool || isPractice:
			queueName = "Practice Tool"
			isPractice = true
		case isARAMCustomQueue(queueID):
			queueName = "Custom ARAM"
			isCustom = true
		default:
			queueName = "Custom Game"
			isCustom = true
		}
		queueInfoResolved = true

	default:
		if info, err := c.fetchQueueInfo(queueID); err != nil {
			c.logger.Warn().Err(err).Int("queue_id", queueID).Msg("Failed to fetch queue info, keeping last known values")
		} else {
			queueName = info.Name
			queueType = info.Type
			queueDescription = info.Description
			queueDetailedDescription = info.DetailedDescription
			isRanked = info.IsRanked
			queueInfoResolved = true
		}
	}

	c.logger.Debug().
		Str("game_mode", string(gameMode)).
		Str("queue_name", queueName).
		Int("queue_id", queueID).
		Bool("is_custom", isCustom).
		Bool("is_practice", isPractice).
		Msg("Lobby state resolved")

	c.state.UpdateField(func(s *state.State) {
		s.LobbyID = lobbyID
		s.Players = players
		s.MaxPlayers = maxPlayers
		s.IsCustom = isCustom
		s.IsPractice = isPractice
		s.QueueID = types.QueueID(queueID)
		s.GameMode = gameMode
		s.MapID = mapID

		if queueInfoResolved {
			s.QueueName = queueName
			s.QueueType = queueType
			s.QueueDescription = queueDescription
			s.QueueDetailedDescription = queueDetailedDescription
			s.IsRanked = isRanked
		}
	})
}

// queueInfo is the subset of /lol-game-queues/v1/queues/{id} used for display.
type queueInfo struct {
	Name                string `json:"name"`
	Type                string `json:"type"`
	IsRanked            bool   `json:"isRanked"`
	Description         string `json:"description"`
	DetailedDescription string `json:"detailedDescription"`
	MapID               int    `json:"mapId"`
	GameMode            string `json:"gameMode"`
	MaxPlayers          int    `json:"maximumParticipantListSize"`
}

// fetchQueueInfo looks up a queue's display name/description by ID.
func (c *Client) fetchQueueInfo(queueID int) (*queueInfo, error) {
	resp, err := c.lcu.Get(fmt.Sprintf("%s/%d", constants.EndpointGameQueues, queueID))
	if err != nil {
		return nil, fmt.Errorf("get queue info: %w", err)
	}
	defer resp.Body.Close()

	var info queueInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decode queue info: %w", err)
	}
	return &info, nil
}

// handleRankedStatsUpdate handles ranked stats changes
func (c *Client) handleRankedStatsUpdate(event *lcu.Event) {
	c.logger.Debug().Msg("Ranked stats update received")

	data, ok := event.Data.(map[string]interface{})
	if !ok {
		c.logger.Warn().Msg("Invalid ranked stats data format")
		return
	}

	// Parse queues array
	queuesData, ok := data["queues"].([]interface{})
	if !ok {
		c.logger.Warn().Msg("Queues not found in ranked stats data")
		return
	}

	c.state.UpdateField(func(s *state.State) {
		for _, queueInterface := range queuesData {
			queue, ok := queueInterface.(map[string]interface{})
			if !ok {
				continue
			}

			queueType, _ := queue["queueType"].(string)
			tier, _ := queue["tier"].(string)
			division, _ := queue["division"].(string)
			leaguePoints, _ := queue["leaguePoints"].(float64)
			ratedTier, _ := queue["ratedTier"].(string)
			ratedRating, _ := queue["ratedRating"].(float64)

			switch queueType {
			case "RANKED_SOLO_5x5":
				s.SummonerRank = types.RankedStats{
					Tier:         types.Tier(tier),
					Division:     types.Division(division),
					LeaguePoints: int(leaguePoints),
				}
				c.logger.Info().Str("rank", s.SummonerRank.String()).Msg("SoloQ rank updated")

			case "RANKED_FLEX_SR":
				s.SummonerRankFlex = types.RankedStats{
					Tier:         types.Tier(tier),
					Division:     types.Division(division),
					LeaguePoints: int(leaguePoints),
				}
				c.logger.Info().Str("rank", s.SummonerRankFlex.String()).Msg("Flex rank updated")

			case "RANKED_TFT":
				s.TFTRank = types.RankedStats{
					Tier:         types.Tier(tier),
					Division:     types.Division(division),
					LeaguePoints: int(leaguePoints),
				}
				c.logger.Info().Str("rank", s.TFTRank.String()).Msg("TFT rank updated")

			case "CHERRY":
				s.ArenaRank = types.ArenaStats{
					RatedTier:   types.ArenaRatedTier(ratedTier),
					Tier:        types.MapRatedTierToTier(types.ArenaRatedTier(ratedTier)),
					RatedRating: int(ratedRating),
				}
				c.logger.Info().Str("rank", s.ArenaRank.String()).Msg("Arena rank updated")
			}
		}
	})
}

// handleTFTCompanionUpdate handles TFT companion selection changes
func (c *Client) handleTFTCompanionUpdate(event *lcu.Event) {
	c.logger.Debug().Msg("TFT companion update received")

	data, ok := event.Data.(map[string]interface{})
	if !ok {
		c.logger.Warn().Msg("Invalid TFT companion data format")
		return
	}

	// Extract selected loadout item
	selectedLoadoutItem, ok := data["selectedLoadoutItem"].(map[string]interface{})
	if !ok {
		c.logger.Debug().Msg("No TFT companion selected")
		return
	}

	itemID, _ := selectedLoadoutItem["itemId"].(float64)
	loadoutsIcon, _ := selectedLoadoutItem["loadoutsIcon"].(string)
	name, _ := selectedLoadoutItem["name"].(string)
	description, _ := selectedLoadoutItem["description"].(string)

	// Extract icon path from full path
	fullPath := loadoutsIcon
	parts := strings.Split(fullPath, "ASSETS/")
	if len(parts) < 2 {
		c.logger.Warn().Str("path", fullPath).Msg("Invalid TFT companion icon path format")
		return
	}
	finalPath := strings.ToLower(parts[1])

	iconURL := discord.GetTFTCompanionURL(finalPath)

	c.state.UpdateTFTCompanion(int(itemID), iconURL, name, description)
	c.logger.Info().Str("companion", name).Msg("TFT companion updated")
}
