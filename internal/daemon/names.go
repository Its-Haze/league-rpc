package daemon

import "github.com/its-haze/league-rpc/pkg/constants"

// discordProcessNames are the process names checked to decide whether
// Discord is running, before attempting an IPC connect.
var discordProcessNames = []string{
	constants.DiscordProcessName,
}

// leagueProcessNames are the process names checked to decide whether League
var leagueProcessNames = []string{
	constants.LeagueClientProcessName,
	constants.LeagueClientUxProcessName,
}
