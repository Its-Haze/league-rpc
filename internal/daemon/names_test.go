package daemon

import (
	"testing"

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
