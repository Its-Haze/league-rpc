package state

import (
	"testing"
	"time"

	"github.com/its-haze/league-rpc/pkg/types"
	"github.com/rs/zerolog"
)

func TestManager_ClearInGameData_ResetsMatchDataButNotOtherState(t *testing.T) {
	m := NewManager(zerolog.Nop())
	m.UpdateChampion("Chogath", "Cho'Gath", "Battlecast Cho'Gath", "", 1)
	m.UpdateGameStart(1000)
	m.UpdateInGameStats(3, 2, 5, 120, 9, 1500)
	m.UpdateQueue("Ranked Solo/Duo", "RANKED_SOLO_5x5", "", 420, true)

	m.ClearInGameData()

	got := m.Get()
	if got.ChampionID != "" || got.ChampionName != "" || got.SkinName != "" || got.SkinID != 0 || got.GameStartTime != 0 {
		t.Errorf("ClearInGameData() left champion fields set: %+v", got)
	}
	if got.Kills != 0 || got.Deaths != 0 || got.Assists != 0 || got.CreepScore != 0 || got.Level != 0 || got.Gold != 0 {
		t.Errorf("ClearInGameData() left KDA/CS/level/gold set: %+v", got)
	}
	if got.QueueName != "Ranked Solo/Duo" {
		t.Errorf("ClearInGameData() must not touch unrelated fields, got QueueName = %q", got.QueueName)
	}
}

func TestManager_UpdateGameFlowPhase_EnteringInProgressClearsPreviousMatchData(t *testing.T) {
	m := NewManager(zerolog.Nop())
	m.UpdateGameFlowPhase(string(types.GameFlowInProgress))
	m.UpdateChampion("Ahri", "Ahri", "", "", 0)
	m.UpdateInGameStats(0, 1, 0, 120, 0, 0)

	// Match ends, then a brand new match starts: entering InProgress again
	m.UpdateGameFlowPhase(string(types.GameFlowEndOfGame))
	m.UpdateGameFlowPhase(string(types.GameFlowChampSelect))
	m.UpdateGameFlowPhase(string(types.GameFlowInProgress))

	got := m.Get()
	if got.ChampionID != "" {
		t.Errorf("ChampionID = %q, want cleared on entering a new match", got.ChampionID)
	}
	if got.Kills != 0 || got.Deaths != 0 || got.CreepScore != 0 {
		t.Errorf("KDA/CS not cleared on entering a new match: %+v", got)
	}
}

func TestManager_UpdateGameFlowPhase_StayingInProgressDoesNotClearData(t *testing.T) {
	// A reconnect blip re-fetches the current phase unconditionally; if it's
	// still InProgress, an already-resolved mid-match champion must survive.
	m := NewManager(zerolog.Nop())
	m.UpdateGameFlowPhase(string(types.GameFlowInProgress))
	m.UpdateChampion("Ahri", "Ahri", "", "", 0)

	m.UpdateGameFlowPhase(string(types.GameFlowInProgress))

	if got := m.Get().ChampionID; got != "Ahri" {
		t.Errorf("ChampionID = %q, want Ahri preserved across a same-phase update", got)
	}
}

func TestManager_UpdateGameFlowPhase_EnteringWatchingClearsPreviousMatchData(t *testing.T) {
	m := NewManager(zerolog.Nop())
	m.UpdateGameFlowPhase(string(types.GameFlowInProgress))
	m.UpdateChampion("Ahri", "Ahri", "", "", 0)
	m.UpdateInGameStats(5, 1, 3, 80, 0, 0)

	m.UpdateGameFlowPhase(string(types.GameFlowEndOfGame))
	m.UpdateGameFlowPhase(string(types.GameFlowWatching))

	got := m.Get()
	if got.ChampionID != "" || got.Kills != 0 {
		t.Errorf("entering Watching must clear last match's champion/KDA: %+v", got)
	}
}

func TestManager_UpdateGameFlowPhase_InProgressToWatchingDoesNotDoubleClear(t *testing.T) {
	// InProgress -> Watching stays within the game-like cluster, so a champion
	// resolved for one shouldn't be wiped just because the phase flipped.
	m := NewManager(zerolog.Nop())
	m.UpdateGameFlowPhase(string(types.GameFlowInProgress))
	m.UpdateChampion("Ahri", "Ahri", "", "", 0)

	m.UpdateGameFlowPhase(string(types.GameFlowWatching))

	if got := m.Get().ChampionID; got != "Ahri" {
		t.Errorf("ChampionID = %q, want Ahri kept across InProgress->Watching", got)
	}
}

func TestManager_UpdateApplicationStartTime_NotifiesEvenThoughEqualsIgnoresIt(t *testing.T) {
	m := NewManager(zerolog.Nop())
	// Drain the buffered channel of any startup notification.
	select {
	case <-m.Updates():
	default:
	}

	m.UpdateApplicationStartTime(1_700_000_000)

	if got := m.Get().ApplicationStartTime; got != 1_700_000_000 {
		t.Fatalf("ApplicationStartTime = %d, want 1700000000", got)
	}
	select {
	case st := <-m.Updates():
		if st.ApplicationStartTime != 1_700_000_000 {
			t.Errorf("notified state ApplicationStartTime = %d", st.ApplicationStartTime)
		}
	default:
		t.Error("UpdateApplicationStartTime did not notify listeners")
	}

	// Same value again: no-op, no notification.
	m.UpdateApplicationStartTime(1_700_000_000)
	select {
	case <-m.Updates():
		t.Error("re-setting the same start time should not notify")
	default:
	}
}

func TestManager_UpdateSpectatedGame_SetsModeAndMap(t *testing.T) {
	m := NewManager(zerolog.Nop())
	m.UpdateSpectatedGame("ARAM", 12)

	got := m.Get()
	if got.GameMode != types.GameModeARAM || got.MapID != types.MapID(12) {
		t.Errorf("UpdateSpectatedGame() = mode %q map %d, want ARAM/12", got.GameMode, got.MapID)
	}

	// Zero values are treated as "unknown" and left alone.
	m.UpdateSpectatedGame("", 0)
	if got := m.Get(); got.GameMode != types.GameModeARAM || got.MapID != types.MapID(12) {
		t.Errorf("empty UpdateSpectatedGame() overwrote known values: %+v", got)
	}
}

func TestManager_Subscribe_DeliversToEachSubscriberAndUpdates(t *testing.T) {
	m := NewManager(zerolog.Nop())
	subA := m.Subscribe()
	subB := m.Subscribe()

	m.UpdateGameFlowPhase(string(types.GameFlowLobby))

	for name, ch := range map[string]<-chan *State{"A": subA, "B": subB} {
		select {
		case st := <-ch:
			if st.GameFlowPhase != types.GameFlowLobby {
				t.Errorf("subscriber %s got phase %q, want Lobby", name, st.GameFlowPhase)
			}
		case <-time.After(time.Second):
			t.Errorf("subscriber %s received nothing", name)
		}
	}

	// The legacy Updates() channel still fires for the same change.
	select {
	case st := <-m.Updates():
		if st.GameFlowPhase != types.GameFlowLobby {
			t.Errorf("Updates() got phase %q, want Lobby", st.GameFlowPhase)
		}
	default:
		t.Error("Updates() channel did not receive the change")
	}
}

func TestManager_Subscribe_CoalescesForASlowSubscriber(t *testing.T) {
	m := NewManager(zerolog.Nop())
	sub := m.Subscribe()

	m.UpdateInGameStats(1, 0, 0, 10, 1, 100)
	m.UpdateInGameStats(2, 0, 0, 20, 2, 200)
	m.UpdateInGameStats(3, 0, 0, 30, 3, 300)

	st := <-sub
	if st.Kills != 3 {
		t.Fatalf("expected the latest coalesced state (Kills=3), got Kills=%d", st.Kills)
	}
	select {
	case extra := <-sub:
		t.Fatalf("expected only the latest state buffered, also got Kills=%d", extra.Kills)
	default:
	}
}

func TestState_Equals_DetectsTFTCompanionDescriptionChange(t *testing.T) {
	a := NewState()
	a.TFTCompanionDescription = "one"
	b := a.Copy()
	b.TFTCompanionDescription = "two"

	if a.Equals(b) {
		t.Error("Equals() = true for states differing only in TFTCompanionDescription")
	}
}
