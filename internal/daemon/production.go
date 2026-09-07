package daemon

import (
	"time"

	"github.com/its-haze/league-rpc/internal/discord"
	"github.com/its-haze/league-rpc/internal/lcu"
	"github.com/its-haze/league-rpc/internal/livegame"
)

// Compile-time checks that the real clients satisfy Connector.
var (
	_ Connector      = (*discord.Client)(nil)
	_ Connector      = (*lcu.Client)(nil)
	_ liveGamePoller = (*livegame.Poller)(nil)
)

// Default retry/poll cadences for local process/IPC checks.
const (
	DefaultRetryInterval       = 3 * time.Second
	DefaultConnectPollInterval = 5 * time.Second
	DefaultProcessPollInterval = 5 * time.Second
)

// Default cadences for Daemon's own presence-mode loop.
const (
	DefaultPresencePollInterval = 3 * time.Second
	DefaultPlaceholderInterval  = 5 * time.Second
)

// discordGate is a Gate that waits for League's process and Discord's
// process to both be running before allowing a connect attempt. See ADR-0003.
type discordGate struct {
	checker ProcessChecker
	league  LeagueDetector
}

func (g *discordGate) Ready() (bool, error) {
	if !g.league.LeagueProcessDetected() {
		return false, nil
	}
	return g.checker.IsRunning(discordProcessNames...)
}

// NewDiscordSupervisor builds the production Discord Connection Supervisor.
// league is typically the already-built LCUSupervisor.
func NewDiscordSupervisor(client *discord.Client, checker ProcessChecker, league LeagueDetector) *DiscordSupervisor {
	return newDiscordSupervisor(client, checker, league, DefaultRetryInterval, DefaultConnectPollInterval, DefaultProcessPollInterval,
		WithGate(&discordGate{checker: checker, league: league}))
}

// NewLeagueSupervisor builds the production LCU Connection Supervisor.
func NewLeagueSupervisor(client *lcu.Client, checker ProcessChecker) *LCUSupervisor {
	return NewLCUSupervisor(client, checker, DefaultRetryInterval, DefaultConnectPollInterval, DefaultProcessPollInterval)
}
