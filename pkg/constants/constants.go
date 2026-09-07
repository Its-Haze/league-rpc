package constants

import (
	"fmt"

	"github.com/its-haze/league-rpc/internal/version"
)

const (
	// Application Info
	AppName = "League RPC"

	// Discord App IDs
	DiscordAppIDDefault = "1194034071588851783" // League of Legends
	DiscordAppIDKittens = "1230607224296968303" // League of Kittens
	DiscordAppIDLinux   = "1185274747836174377" // League of Linux

	// LCU API Endpoints (for reference)
	EndpointSummoner             = "/lol-summoner/v1/current-summoner"
	EndpointGameflow             = "/lol-gameflow/v1/gameflow-phase"
	EndpointChampSelect          = "/lol-champ-select/v1/session"
	EndpointLobby                = "/lol-lobby/v2/lobby"
	EndpointRankedStats          = "/lol-ranked/v1/current-ranked-stats"
	EndpointChatStatus           = "/lol-chat/v1/me"
	EndpointTFTCompanion         = "/lol-cosmetics/v1/inventories/tft/companions"
	EndpointGameQueues           = "/lol-game-queues/v1/queues"
	EndpointGameflowMetadata     = "/lol-gameflow/v1/gameflow-metadata/player-status"
	EndpointApplicationStartTime = "/telemetry/v1/application-start-time"

	// In-Game API
	InGameAPIURL        = "https://127.0.0.1:2999"
	InGameStatsEndpoint = "/liveclientdata/allgamedata"

	// Asset URLs
	CommunityDragonBaseURL = "https://raw.communitydragon.org"
	DataDragonBaseURL      = "https://ddragon.leagueoflegends.com"

	// Update Throttle
	DefaultUpdateInterval = 1500 // milliseconds

	// Process Names
	DiscordProcessName        = "Discord.exe"
	LeagueClientProcessName   = "LeagueClient.exe"
	LeagueClientUxProcessName = "LeagueClientUx.exe"
	RiotClientProcessName     = "RiotClientServices.exe"
)

// SmallText is the small-icon hover tooltip shown on every presence state.
// Built at load time so the tooltip reports the running build, not a literal.
var SmallText = fmt.Sprintf("its-haze/league-rpc @Github.com (%s)", version.Version())

// GameFlowPhases represents the different phases in the League Client
var GameFlowPhases = struct {
	None              string
	Lobby             string
	Matchmaking       string
	ReadyCheck        string
	ChampSelect       string
	GameStart         string
	FailedToLaunch    string
	InProgress        string
	Reconnect         string
	WaitingForStats   string
	PreEndOfGame      string
	EndOfGame         string
	TerminatedInError string
}{
	None:              "None",
	Lobby:             "Lobby",
	Matchmaking:       "Matchmaking",
	ReadyCheck:        "ReadyCheck",
	ChampSelect:       "ChampSelect",
	GameStart:         "GameStart",
	FailedToLaunch:    "FailedToLaunch",
	InProgress:        "InProgress",
	Reconnect:         "Reconnect",
	WaitingForStats:   "WaitingForStats",
	PreEndOfGame:      "PreEndOfGame",
	EndOfGame:         "EndOfGame",
	TerminatedInError: "TerminatedInError",
}

// QueueTypes represents different queue types
var QueueTypes = struct {
	SoloQ int
	Flex  int
	ARAM  int
	TFT   int
	Arena int
}{
	SoloQ: 420,
	Flex:  440,
	ARAM:  450,
	TFT:   1090,
	Arena: 1700,
}
