package daemon

import (
	"testing"

	"github.com/its-haze/league-rpc/internal/process"
	"github.com/its-haze/league-rpc/pkg/constants"
)

// TestLeagueProcessNames_ExcludesRiotClient guards against showing League as
func TestLeagueProcessNames_ExcludesRiotClient(t *testing.T) {
	for _, name := range leagueProcessNames {
		if name == constants.RiotClientProcessName {
			t.Fatalf("leagueProcessNames must not include the Riot Client (%s), it outlives League itself", name)
		}
	}
}

// TestDiscordProcessNames_MatchForksAsTaskManagerReportsThem covers the bug
// where a Vesktop user's gate never opened, keeping the app off a live pipe.
func TestDiscordProcessNames_MatchForksAsTaskManagerReportsThem(t *testing.T) {
	// Lowercased the way Process Explorer reported them on a real machine.
	for _, running := range []string{"vesktop.exe", "discordptb.exe", "discordcanary.exe", "legcord.exe"} {
		checker := process.NewCheckerWithLister(fakeNameLister{running})
		got, err := checker.IsRunning(discordProcessNames...)
		if err != nil {
			t.Fatalf("IsRunning(%s): %v", running, err)
		}
		if !got {
			t.Errorf("%s must count as Discord: it serves the RPC pipe via arRPC", running)
		}
	}
}

type fakeNameLister []string

func (f fakeNameLister) ProcessNames() ([]string, error) { return f, nil }
