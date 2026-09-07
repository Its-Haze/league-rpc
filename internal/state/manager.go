package state

import (
	"sync"

	"github.com/its-haze/league-rpc/pkg/types"
	"github.com/rs/zerolog"
)

// Manager manages the application state with thread-safe operations
type Manager struct {
	mu      sync.RWMutex
	current *State
	logger  zerolog.Logger

	// Channel for state change notifications
	updates chan *State

	// Fan-out channels handed to Subscribe callers. Each gets its own copy of
	// every change; a slow reader only ever misses intermediate states.
	subs []chan *State
}

// NewManager creates a new state manager
func NewManager(logger zerolog.Logger) *Manager {
	return &Manager{
		current: NewState(),
		logger:  logger,
		updates: make(chan *State, 100), // Buffered channel for updates
	}
}

// Get returns a copy of the current state (thread-safe read)
func (m *Manager) Get() *State {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.current.Copy()
}

// Update updates the current state and notifies listeners if changed
func (m *Manager) Update(newState *State) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if state actually changed
	if m.current.Equals(newState) {
		m.logger.Debug().Msg("State unchanged, skipping update")
		return
	}

	m.logger.Debug().
		Str("phase", string(newState.GameFlowPhase)).
		Str("queue", newState.QueueName).
		Msg("State updated")

	m.current = newState.Copy()
	m.broadcast()
}

// UpdateField updates a specific field and notifies if changed
func (m *Manager) UpdateField(updateFunc func(*State)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create a copy to modify
	newState := m.current.Copy()
	updateFunc(newState)

	// Check if changed
	if m.current.Equals(newState) {
		return
	}

	m.current = newState
	m.broadcast()
}

// Updates returns the channel for receiving state updates
func (m *Manager) Updates() <-chan *State {
	return m.updates
}

// Subscribe returns a fresh channel that receives a copy of the state on every
// change, independent of Updates() and of any other subscriber.
func (m *Manager) Subscribe() <-chan *State {
	m.mu.Lock()
	defer m.mu.Unlock()
	ch := make(chan *State, 1)
	m.subs = append(m.subs, ch)
	return ch
}

// broadcast pushes the current state to the shared Updates() channel and to
// every Subscribe() channel. Callers must hold m.mu.
func (m *Manager) broadcast() {
	select {
	case m.updates <- m.current.Copy():
	default:
		m.logger.Warn().Msg("State update channel full, dropping update")
	}
	for _, ch := range m.subs {
		// Coalesce: drop a stale pending value, then push the latest. Both
		// sends are non-blocking so a slow subscriber never stalls a writer.
		select {
		case <-ch:
		default:
		}
		select {
		case ch <- m.current.Copy():
		default:
		}
	}
}

// Close closes the state manager and its update channels
func (m *Manager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	close(m.updates)
	for _, ch := range m.subs {
		close(ch)
	}
	m.subs = nil
}

// Specific update methods for common operations

// UpdateSummonerIcon updates the summoner icon
func (m *Manager) UpdateSummonerIcon(icon int) {
	m.UpdateField(func(s *State) {
		s.SummonerIcon = icon
	})
}

// UpdateAvailability updates the player's availability status
func (m *Manager) UpdateAvailability(availability string) {
	m.UpdateField(func(s *State) {
		switch availability {
		case "online":
			s.Availability = "Online"
		case "away":
			s.Availability = "Away"
		case "dnd":
			s.Availability = "dnd"
		default:
			s.Availability = "Online"
		}
	})
}

// UpdateGameFlowPhase sets the phase, clearing last match's in-game data
// when entering the InProgress/Watching cluster from outside it.
func (m *Manager) UpdateGameFlowPhase(phase string) {
	m.UpdateField(func(s *State) {
		newPhase := types.GameFlowPhase(phase)
		isGameLike := newPhase == types.GameFlowInProgress || newPhase == types.GameFlowWatching
		wasGameLike := s.GameFlowPhase == types.GameFlowInProgress || s.GameFlowPhase == types.GameFlowWatching
		if isGameLike && !wasGameLike {
			resetInGameData(s)
		}
		s.GameFlowPhase = newPhase
	})
}

// UpdateLobby updates lobby information
func (m *Manager) UpdateLobby(lobbyID string, players, maxPlayers int, isCustom, isPractice bool) {
	m.UpdateField(func(s *State) {
		s.LobbyID = lobbyID
		s.Players = players
		s.MaxPlayers = maxPlayers
		s.IsCustom = isCustom
		s.IsPractice = isPractice
	})
}

// UpdateQueue updates queue information
func (m *Manager) UpdateQueue(name, queueType, description string, queueID int, isRanked bool) {
	m.UpdateField(func(s *State) {
		s.QueueName = name
		s.QueueType = queueType
		s.QueueDescription = description
		s.QueueID = types.QueueID(queueID)
		s.IsRanked = isRanked
	})
}

// UpdateTFTCompanion updates TFT companion data
func (m *Manager) UpdateTFTCompanion(id int, icon, name, description string) {
	m.UpdateField(func(s *State) {
		s.TFTCompanionID = id
		s.TFTCompanionIcon = icon
		s.TFTCompanionName = name
		s.TFTCompanionDescription = description
	})
}

// UpdateInGameStats updates in-game statistics (KDA, CS, etc.)
func (m *Manager) UpdateInGameStats(kills, deaths, assists, cs, level, gold int) {
	m.UpdateField(func(s *State) {
		s.Kills = kills
		s.Deaths = deaths
		s.Assists = assists
		s.CreepScore = cs
		s.Level = level
		s.Gold = gold
	})
}

// UpdateChampion updates champion/skin information. championID is Data
// Dragon's URL-safe id (feeds asset URLs); championName is its display name.
func (m *Manager) UpdateChampion(championID, championName, skinName, chromaName string, skinID int) {
	m.UpdateField(func(s *State) {
		s.ChampionID = championID
		s.ChampionName = championName
		s.SkinName = skinName
		s.ChromaName = chromaName
		s.SkinID = skinID
	})
}

// UpdateGameStart records the match's real start time, so presence's
// elapsed timer counts up accurately instead of resetting on every poll.
func (m *Manager) UpdateGameStart(start int64) {
	m.UpdateField(func(s *State) {
		s.GameStartTime = start
	})
}

// UpdateSpectatedGame records the game mode and map of a game being
// spectated, both read from Live Client Data since there is no lobby.
func (m *Manager) UpdateSpectatedGame(rawGameMode string, mapNumber int) {
	m.UpdateField(func(s *State) {
		if rawGameMode != "" {
			s.GameMode = types.GameMode(rawGameMode)
		}
		if mapNumber != 0 {
			s.MapID = types.MapID(mapNumber)
		}
	})
}

// UpdateApplicationStartTime records the League client's own start time.
func (m *Manager) UpdateApplicationStartTime(start int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.current.ApplicationStartTime == start {
		return
	}
	m.current.ApplicationStartTime = start
	m.broadcast()
}

// ClearInGameData resets last match's champion/skin/chroma, start time, and
func (m *Manager) ClearInGameData() {
	m.UpdateField(resetInGameData)
}

// resetInGameData zeroes the fields that are only meaningful for one match,
// so a new match never inherits data left over from the previous one.
func resetInGameData(s *State) {
	s.ChampionID = ""
	s.ChampionName = ""
	s.SkinName = ""
	s.ChromaName = ""
	s.SkinID = 0
	s.GameStartTime = 0
	s.Kills = 0
	s.Deaths = 0
	s.Assists = 0
	s.CreepScore = 0
	s.Level = 0
	s.Gold = 0
}
