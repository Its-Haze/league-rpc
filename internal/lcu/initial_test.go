package lcu

import (
	"context"
	"errors"
	"testing"

	"github.com/its-haze/league-rpc/internal/state"
	"github.com/its-haze/league-rpc/pkg/types"
	"github.com/rs/zerolog"
)

// fakeGameModeFetcher lets tests control what Live Client Data reports
// without a real League client.
type fakeGameModeFetcher struct {
	mode string
	err  error
}

func (f *fakeGameModeFetcher) GameMode(ctx context.Context) (string, error) {
	return f.mode, f.err
}

func newTestClient(liveGame GameModeFetcher) (*Client, *state.Manager) {
	stateMgr := state.NewManager(zerolog.Nop())
	c := &Client{state: stateMgr, logger: zerolog.Nop(), liveGame: liveGame}
	return c, stateMgr
}

func TestRecoverGameModeFromLiveClientData_PracticeTool(t *testing.T) {
	c, stateMgr := newTestClient(&fakeGameModeFetcher{mode: "PRACTICETOOL"})
	stateMgr.UpdateGameFlowPhase(string(types.GameFlowInProgress))

	c.recoverGameModeFromLiveClientData()

	got := stateMgr.Get()
	if got.GameMode != "PRACTICETOOL" {
		t.Errorf("GameMode = %q, want PRACTICETOOL", got.GameMode)
	}
	if got.QueueName != "Practice Tool" {
		t.Errorf("QueueName = %q, want Practice Tool", got.QueueName)
	}
	if !got.IsPractice {
		t.Error("IsPractice = false, want true")
	}
}

func TestRecoverGameModeFromLiveClientData_NonPracticeDoesNotSetQueueName(t *testing.T) {
	c, stateMgr := newTestClient(&fakeGameModeFetcher{mode: "CLASSIC"})
	stateMgr.UpdateGameFlowPhase(string(types.GameFlowInProgress))

	c.recoverGameModeFromLiveClientData()

	got := stateMgr.Get()
	if got.GameMode != "CLASSIC" {
		t.Errorf("GameMode = %q, want CLASSIC", got.GameMode)
	}
	if got.QueueName != "" || got.IsPractice {
		t.Errorf("expected no queue name/practice flag for a non-practice mode, got %+v", got)
	}
}

func TestRecoverGameModeFromLiveClientData_NoopWhenNotInProgress(t *testing.T) {
	c, stateMgr := newTestClient(&fakeGameModeFetcher{mode: "PRACTICETOOL"})
	stateMgr.UpdateGameFlowPhase(string(types.GameFlowChampSelect))

	c.recoverGameModeFromLiveClientData()

	if got := stateMgr.Get().GameMode; got != types.GameModeClassic {
		t.Errorf("GameMode = %q, want unchanged default (not in progress yet)", got)
	}
}

func TestRecoverGameModeFromLiveClientData_NoopOnFetchError(t *testing.T) {
	c, stateMgr := newTestClient(&fakeGameModeFetcher{err: errors.New("live client unreachable")})
	stateMgr.UpdateGameFlowPhase(string(types.GameFlowInProgress))

	c.recoverGameModeFromLiveClientData()

	if got := stateMgr.Get().GameMode; got != types.GameModeClassic {
		t.Errorf("GameMode = %q, want unchanged default (fetch failed)", got)
	}
}

func TestRecoverGameModeFromLiveClientData_NoopWhenFetcherNil(t *testing.T) {
	c, stateMgr := newTestClient(nil)
	stateMgr.UpdateGameFlowPhase(string(types.GameFlowInProgress))

	c.recoverGameModeFromLiveClientData() // must not panic on a nil liveGame
}

func TestCustomLobbyDefaults(t *testing.T) {
	tests := []struct {
		name       string
		queueID    int
		isPractice bool
		wantName   string
		wantMode   types.GameMode
		wantMap    types.MapID
		wantMax    int
	}{
		{"practice tool by queue id", queueIDPracticeTool, false, "Practice Tool", "PRACTICETOOL", types.MapSummonersRift, 1},
		{"practice tool by flag", 0, true, "Practice Tool", "PRACTICETOOL", types.MapSummonersRift, 1},
		{"custom aram", queueIDARAMCustomDraft, false, "Custom ARAM", "ARAM", types.MapHowlingAbyss, 0},
		{"generic custom", queueIDCustomGameDraft, false, "Custom Game", "PRACTICETOOL", types.MapSummonersRift, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, mode, mapID, maxPlayers := customLobbyDefaults(tt.queueID, tt.isPractice)
			if name != tt.wantName || mode != tt.wantMode || mapID != tt.wantMap || maxPlayers != tt.wantMax {
				t.Errorf("customLobbyDefaults(%d, %v) = %q/%q/%d/%d, want %q/%q/%d/%d",
					tt.queueID, tt.isPractice, name, mode, mapID, maxPlayers,
					tt.wantName, tt.wantMode, tt.wantMap, tt.wantMax)
			}
		})
	}
}
