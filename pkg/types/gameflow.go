package types

// GameFlowPhase represents the different phases in the League Client
// Based on: /lol-gameflow/v1/gameflow-phase
type GameFlowPhase string

const (
	GameFlowNone                  GameFlowPhase = "None"
	GameFlowLobby                 GameFlowPhase = "Lobby"
	GameFlowMatchmaking           GameFlowPhase = "Matchmaking"
	GameFlowReadyCheck            GameFlowPhase = "ReadyCheck"
	GameFlowChampSelect           GameFlowPhase = "ChampSelect"
	GameFlowGameStart             GameFlowPhase = "GameStart"
	GameFlowFailedToLaunch        GameFlowPhase = "FailedToLaunch"
	GameFlowInProgress            GameFlowPhase = "InProgress"
	GameFlowWatching              GameFlowPhase = "Watching"
	GameFlowReconnect             GameFlowPhase = "Reconnect"
	GameFlowWaitingForStats       GameFlowPhase = "WaitingForStats"
	GameFlowPreEndOfGame          GameFlowPhase = "PreEndOfGame"
	GameFlowEndOfGame             GameFlowPhase = "EndOfGame"
	GameFlowTerminatedInError     GameFlowPhase = "TerminatedInError"
	GameFlowCheckedIntoTournament GameFlowPhase = "CheckedIntoTournament"
)

// IsInGame returns true if the phase represents an active game
func (g GameFlowPhase) IsInGame() bool {
	return g == GameFlowInProgress
}

// IsSpectating returns true if the phase represents spectating someone
// else's game rather than playing in one.
func (g GameFlowPhase) IsSpectating() bool {
	return g == GameFlowWatching
}

// IsInLobby returns true if the phase represents being in a lobby
func (g GameFlowPhase) IsInLobby() bool {
	return g == GameFlowLobby
}

// IsInQueue returns true if the phase represents being in queue
func (g GameFlowPhase) IsInQueue() bool {
	return g == GameFlowMatchmaking || g == GameFlowCheckedIntoTournament
}

// IsInChampSelect returns true if the phase represents champion selection
func (g GameFlowPhase) IsInChampSelect() bool {
	return g == GameFlowChampSelect || g == GameFlowGameStart
}

// IsInClient returns true if the phase represents idle in client
func (g GameFlowPhase) IsInClient() bool {
	return g == GameFlowNone || g == GameFlowWaitingForStats ||
		g == GameFlowPreEndOfGame || g == GameFlowEndOfGame
}
