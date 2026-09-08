# Discord's connection is gated by League's process, not independent of it

**Status**: accepted

ADR-0002 established that the Discord and LCU Connection Supervisors run independently: League being closed must not block waiting for Discord, and vice versa. That held for LCU, but for Discord it meant the Daemon would connect to Discord and publish presence the moment Discord itself was running, regardless of whether League was open at all. Closing League left the last real presence sitting on the user's Discord profile indefinitely, since nothing ever told the Discord connection to go away.

There is no presence worth showing without League running, so there is no reason to hold a Discord IPC connection open without it either. `DiscordSupervisor` now takes a `LeagueDetector` (satisfied by `LCUSupervisor`) and requires `LeagueProcessDetected()` to be true both before attempting `Connect()` and for its gated `IsConnected()` to report true. This reuses the Supervisor's existing retry-on-disconnect loop rather than adding separate teardown logic: the moment League's process disappears, the gated `IsConnected()` flips false, `Supervisor.Run()` disconnects and starts polling the Gate again, and the connection only comes back once League's process reappears.

**Consequences**: the Daemon's presence loop logs once (not on every poll tick) whenever League is running but Discord isn't yet reachable, so a user can tell the app is waiting rather than stuck. This supersedes ADR-0002's "the two supervisors run independently" line for the Discord side specifically; the LCU Supervisor is unaffected and still starts connecting the moment League's process is detected, independent of Discord.

A gate that additionally accepted the Riot Client launcher (not just League specifically) was considered, so a "Starting Riot Client..." placeholder could show on Discord before League itself launched. Rejected: it would put a League-branded Discord card on screen while someone launches an unrelated Riot game, which is the exact problem being fixed here.
