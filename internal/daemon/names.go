package daemon

import "github.com/its-haze/league-rpc/pkg/constants"

// discordProcessNames are the process names checked to decide whether
// Discord is running, before attempting an IPC connect. Forks are included
// because arRPC serves the same pipe, so the connect works identically.
var discordProcessNames = []string{
	constants.DiscordProcessName,
	constants.DiscordPTBProcessName,
	constants.DiscordCanaryProcessName,
	constants.DiscordDevelopmentProcessName,
	constants.VesktopProcessName,
	constants.LegcordProcessName,
	constants.ArmCordProcessName,
}

// leagueProcessNames are the process names checked to decide whether League
var leagueProcessNames = []string{
	constants.LeagueClientProcessName,
	constants.LeagueClientUxProcessName,
}
