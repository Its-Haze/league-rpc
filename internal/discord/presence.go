package discord

import (
	"strconv"
	"time"

	"github.com/its-haze/league-rpc/internal/config"
	"github.com/its-haze/league-rpc/internal/presence/template"
	"github.com/its-haze/league-rpc/internal/state"
	"github.com/its-haze/league-rpc/pkg/constants"
	"github.com/its-haze/league-rpc/pkg/types"
)

// queueDisplayName resolves the queue name shown as details, falling all
// the way back to "League of Legends" so Discord never gets an empty string.
func queueDisplayName(st *state.State) string {
	if name := st.GetQueueName(); name != "" {
		return name
	}
	if name := FormatGameModeName(st.GameMode); name != "" {
		return name
	}
	return "League of Legends"
}

// BuildInClientPresence builds RPC data for when the player is idle in the client
func BuildInClientPresence(st *state.State, cfg *config.Config) *RPCData {
	emoji := ""
	if cfg.Presence.ShowEmojis {
		emoji = "\U0001F7E2" // green circle
		if st.Availability == types.AvailabilityAway {
			emoji = "\U0001F534" // red circle
		}
	}

	details, stateText := renderPresenceText(cfg, template.ContextInClient, map[string]string{
		"emoji":        emoji,
		"availability": string(st.Availability),
	})

	return &RPCData{
		LargeImage: GetProfileIconURL(st.SummonerIcon),
		LargeText:  "In Client",
		SmallImage: GetLeagueLogoURL(),
		SmallText:  constants.SmallText,
		Details:    details,
		State:      stateText,
		Start:      st.ApplicationStartTime,
	}
}

// BuildInLobbyPresence builds RPC data for when the player is in a lobby
func BuildInLobbyPresence(st *state.State, cfg *config.Config) *RPCData {
	largeImage := GetProfileIconURL(st.SummonerIcon)
	largeText := FormatGameModeName(st.GameMode)
	smallImage := GetMapIconURL(st.MapID)
	smallText := constants.SmallText

	// Handle TFT - use companion instead of profile icon
	if st.GameMode == types.GameModeTFT && st.TFTCompanionIcon != "" {
		largeImage = st.TFTCompanionIcon
		largeText = st.TFTCompanionName
	}

	// Rank known: swap the tooltip for the credit line, small image/text for
	// the rank emblem/tier.
	if cfg.Display.Default.ShowRank {
		rankEmblemURL, rankText := getRankForQueue(st, st.QueueID)
		if rankEmblemURL != "" {
			largeText = constants.SmallText
			smallImage = rankEmblemURL
			smallText = rankText
		}
	}

	// BRAWL/League Classic force the classic icon regardless of rank.
	if isForcedClassicIconMode(st.GameMode) {
		smallImage = GetLeagueLogoURL()
	}

	details, lobbyState := renderPresenceText(cfg, template.ContextLobby, map[string]string{
		"queue":       queueDisplayName(st),
		"players":     strconv.Itoa(st.Players),
		"max_players": strconv.Itoa(st.MaxPlayers),
	})

	return &RPCData{
		LargeImage: largeImage,
		LargeText:  largeText,
		SmallImage: smallImage,
		SmallText:  smallText,
		Details:    details,
		State:      lobbyState,
		Start:      st.ApplicationStartTime,
	}
}

// BuildInCustomLobbyPresence builds RPC data for custom games and the practice tool.
func BuildInCustomLobbyPresence(st *state.State, cfg *config.Config) *RPCData {
	largeImage := GetProfileIconURL(st.SummonerIcon)
	largeText := FormatGameModeName(st.GameMode)
	smallImage := GetMapIconURL(st.MapID)
	smallText := constants.SmallText

	details, lobbyState := renderPresenceText(cfg, template.ContextCustomLobby, map[string]string{
		"queue":       queueDisplayName(st),
		"players":     strconv.Itoa(st.Players),
		"max_players": strconv.Itoa(st.MaxPlayers),
	})

	return &RPCData{
		LargeImage: largeImage,
		LargeText:  largeText,
		SmallImage: smallImage,
		SmallText:  smallText,
		Details:    details,
		State:      lobbyState,
		Start:      st.ApplicationStartTime,
	}
}

// BuildInQueuePresence builds RPC data for when the player is in matchmaking queue
func BuildInQueuePresence(st *state.State, cfg *config.Config) *RPCData {
	largeImage := GetProfileIconURL(st.SummonerIcon)
	largeText := FormatGameModeName(st.GameMode)
	smallImage := GetMapIconURL(st.MapID)
	smallText := constants.SmallText

	// Rank known: swap the tooltip for the credit line, small image/text for
	// the rank emblem/tier.
	if cfg.Display.Default.ShowRank {
		rankEmblemURL, rankText := getRankForQueue(st, st.QueueID)
		if rankEmblemURL != "" {
			largeText = constants.SmallText
			smallImage = rankEmblemURL
			smallText = rankText
		}
	}

	// BRAWL/League Classic force the classic icon regardless of rank.
	if isForcedClassicIconMode(st.GameMode) {
		smallImage = GetLeagueLogoURL()
	}

	details, queueState := renderPresenceText(cfg, template.ContextQueue, map[string]string{
		"queue": queueDisplayName(st),
	})

	return &RPCData{
		LargeImage: largeImage,
		LargeText:  largeText,
		SmallImage: smallImage,
		SmallText:  smallText,
		Details:    details,
		State:      queueState,
		Start:      time.Now().Unix(), // Start timer from now for queue time
	}
}

// BuildInChampSelectPresence builds RPC data for champion selection phase
func BuildInChampSelectPresence(st *state.State, cfg *config.Config) *RPCData {
	largeImage := GetProfileIconURL(st.SummonerIcon)
	largeText := FormatGameModeName(st.GameMode)
	smallImage := GetMapIconURL(st.MapID)
	smallText := constants.SmallText

	// Rank known: swap the tooltip for the credit line, small image/text for
	// the rank emblem/tier.
	if cfg.Display.Default.ShowRank {
		rankEmblemURL, rankText := getRankForQueue(st, st.QueueID)
		if rankEmblemURL != "" {
			largeText = constants.SmallText
			smallImage = rankEmblemURL
			smallText = rankText
		}
	}

	// BRAWL/League Classic force the classic icon regardless of rank.
	if isForcedClassicIconMode(st.GameMode) {
		smallImage = GetLeagueLogoURL()
	}

	details, stateText := renderPresenceText(cfg, template.ContextChampSelect, map[string]string{
		"queue": queueDisplayName(st),
		"mode":  FormatGameModeName(st.GameMode),
	})

	return &RPCData{
		LargeImage: largeImage,
		LargeText:  largeText,
		SmallImage: smallImage,
		SmallText:  smallText,
		Details:    details,
		State:      stateText,
		Start:      time.Now().Unix(), // Start timer from now
	}
}

// BuildInGamePresence builds RPC data for when the player is in an active game
func BuildInGamePresence(st *state.State, cfg *config.Config) *RPCData {
	// Large image: Champion skin
	largeImage := GetChampionSkinURL(st.ChampionID, st.SkinID)
	largeText := FormatSkinName(st.ChampionName, st.SkinName, st.ChromaName)

	// Small image: Rank emblem or League logo
	smallImage := GetLeagueLogoURL()
	smallText := constants.SmallText

	// Show rank emblem if enabled. Unlike Lobby/Queue/ChampSelect, largeText
	// stays the skin name here rather than swapping to the credit line.
	if cfg.Display.Default.ShowRank {
		rankEmblemURL, rankText := getRankForQueue(st, st.QueueID)
		if rankEmblemURL != "" {
			smallImage = rankEmblemURL
			smallText = rankText
		}
	}

	if isForcedClassicIconMode(st.GameMode) {
		smallImage = GetLeagueLogoURL()
	}

	// Mode-specific stat line, empty when stats are hidden. Arena and Swarm
	// show level and gold; everything else shows KDA and CS.
	stats := ""
	if cfg.Display.Default.ShowStats {
		switch st.GameMode {
		case types.GameModeArena:
			stats = FormatArenaStats(st.Kills, st.Deaths, st.Assists, st.Level, st.Gold)
		case types.GameModeSwarm:
			stats = FormatSwarmStats(st.CreepScore, st.Level, st.Gold)
		default:
			stats = FormatKDA(st.Kills, st.Deaths, st.Assists, st.CreepScore)
		}
	}

	details, gameState := renderPresenceText(cfg, template.ContextInGame, map[string]string{
		"queue":    queueDisplayName(st),
		"mode":     FormatGameModeName(st.GameMode),
		"stats":    stats,
		"champion": st.ChampionName,
		"skin":     FormatSkinName(st.ChampionName, st.SkinName, st.ChromaName),
	})

	// Loading screen: GameStartTime isn't resolved yet, so fall back to now.
	start := st.GameStartTime
	if start == 0 {
		start = time.Now().Unix()
	}

	return &RPCData{
		LargeImage: largeImage,
		LargeText:  largeText,
		SmallImage: smallImage,
		SmallText:  smallText,
		Details:    details,
		State:      gameState,
		Start:      start,
	}
}

// BuildTFTInGamePresence builds RPC data for TFT games
func BuildTFTInGamePresence(st *state.State, cfg *config.Config) *RPCData {
	largeImage := st.TFTCompanionIcon
	largeText := st.TFTCompanionName
	smallImage := GetLeagueLogoURL()
	smallText := constants.SmallText

	// Use TFT rank if available
	if cfg.Display.Default.ShowRank && !st.TFTRank.IsEmpty() {
		smallImage = GetRankEmblemURL(st.TFTRank.Tier)
		smallText = st.TFTRank.String()
	}

	details, gameState := renderPresenceText(cfg, template.ContextTFTInGame, map[string]string{
		"queue": queueDisplayName(st),
		"level": strconv.Itoa(st.Level),
	})

	// Loading screen: GameStartTime isn't resolved yet, so fall back to now.
	start := st.GameStartTime
	if start == 0 {
		start = time.Now().Unix()
	}

	return &RPCData{
		LargeImage: largeImage,
		LargeText:  largeText,
		SmallImage: smallImage,
		SmallText:  smallText,
		Details:    details,
		State:      gameState,
		Start:      start,
	}
}

// BuildSpectatingPresence builds RPC data for when the player is spectating
func BuildSpectatingPresence(st *state.State, cfg *config.Config) *RPCData {
	largeImage := GetLeagueLogoLargeURL()
	switch {
	case st.ChampionID != "":
		largeImage = GetChampionSkinURL(st.ChampionID, st.SkinID)
	case st.MapID != 0:
		largeImage = GetMapIconURL(st.MapID)
	}

	// Loading screen: GameStartTime isn't resolved yet, so fall back to now.
	start := st.GameStartTime
	if start == 0 {
		start = time.Now().Unix()
	}

	details, stateText := renderPresenceText(cfg, template.ContextSpectating, map[string]string{
		"mode":  FormatGameModeName(st.GameMode),
		"queue": queueDisplayName(st),
	})
	// Discord rejects an empty details string; before the first spectate poll
	// resolves the mode there may be nothing to show.
	if details == "" {
		details = "League of Legends"
	}

	return &RPCData{
		LargeImage: largeImage,
		LargeText:  "Spectating",
		SmallImage: GetLeagueLogoURL(),
		SmallText:  constants.SmallText,
		Details:    details,
		State:      stateText,
		Start:      start,
	}
}

// BuildLaunchingPresence builds the "launching" placeholder, with a
// randomly picked animated skin as its large image on every call.
func BuildLaunchingPresence(start int64) *RPCData {
	return &RPCData{
		LargeImage: GetLaunchingPlaceholderImageURL(),
		LargeText:  "LeagueRPC",
		SmallImage: GetLeagueLogoURL(),
		SmallText:  constants.SmallText,
		Details:    "Launching League...",
		State:      "LeagueRPC",
		Start:      start,
	}
}

// isForcedClassicIconMode reports whether gameMode forces the classic icon
// as the small image regardless of rank.
func isForcedClassicIconMode(gameMode types.GameMode) bool {
	switch gameMode {
	case types.GameModeBrawl, "JADE", "KIWI_JADE":
		return true
	default:
		return false
	}
}

// getRankForQueue returns the appropriate rank emblem and text for the current queue
func getRankForQueue(st *state.State, queueID types.QueueID) (string, string) {
	switch queueID {
	case types.QueueSoloQ:
		if !st.SummonerRank.IsEmpty() {
			return GetRankEmblemURL(st.SummonerRank.Tier), st.SummonerRank.String()
		}

	case types.QueueFlex:
		if !st.SummonerRankFlex.IsEmpty() {
			return GetRankEmblemURL(st.SummonerRankFlex.Tier), st.SummonerRankFlex.String()
		}

	case types.QueueTFT:
		if !st.TFTRank.IsEmpty() {
			return GetRankEmblemURL(st.TFTRank.Tier), st.TFTRank.String()
		}

	case types.QueueArena:
		if !st.ArenaRank.IsEmpty() {
			return GetArenaEmblemURL(st.ArenaRank.Tier), st.ArenaRank.String()
		}
	}

	// No rank for this queue or unranked
	return "", ""
}
