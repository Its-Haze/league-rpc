# League RPC (Go rewrite)

Discord Rich Presence for League of Legends, driven by the LCU (League Client Update) API.

## Language

**State**:
The application's live snapshot of what the summoner is doing right now (client phase, queue, champion/skin, ranked stats).
_Avoid_: ClientData, session

**GameFlowPhase**:
The LCU's own phase string for what the client is doing (`None`, `Lobby`, `ChampSelect`, `InProgress`, `Watching`, ...). Values must mirror the LCU API's raw strings exactly: this is not a value we get to invent, it's a contract with Riot's client.
_Avoid_: GameState, Phase

**Watching**:
The `GameFlowPhase` value meaning the summoner is spectating a game, not playing in one. Presence for this phase shows the spectated match's leading player's champion/skin (or the map icon, for modes without champions) rather than the summoner's own KDA.

**TFT queue variants**:
Ranked TFT (queue ID 1100), Double Up (1160), and Hyper Roll (1130) are each distinct from normal/unranked TFT (1090) and must not share one `QueueID` constant: ranked-stat lookups need the specific ID, not just "is this TFT."

**Updater**:
The single component responsible for every Discord presence send. It debounces rapid `State` changes into one send, then keeps resending periodically (the "heartbeat") for as long as that presence is current, with a faster resend burst (the "reclaim") right after each real change. It exists because Discord lets League's own native Rich Presence integration silently overwrite ours if we stop sending. See [ADR-0001](./docs/adr/0001-updater-owns-the-heartbeat.md).
_Avoid_: RPCUpdater, presence manager

**Live Client Data**:
The unauthenticated local API at `127.0.0.1:2999`, distinct from the LCU: fixed port, no lockfile auth, and only reachable while a match is actually in progress. Used for in-game KDA/champion/gold polling. Owned by its own package (`internal/livegame`), not `internal/lcu`.
_Avoid_: LCU, game data (bare)

**Champion Data**:
Static champion/skin/chroma lookups from Data Dragon and Community Dragon: plain HTTP, no League process required at all. Owned by its own package (`internal/championdata`), not `internal/lcu`.
_Avoid_: LCU, ddragon (as a stand-in for the whole concept)

**League Classic**:
Display name for the `JADE` game mode (raw LCU/game-data identifier), played on the Classic Rift map (map ID 453).

**ARAM: Mayhem**:
Display name for the `KIWI` game mode.

**ARAM: Mayhem Classic-ish**:
Display name for the `KIWI_JADE` game mode: ARAM Mayhem played on the League Classic map. The "-ish" is intentional, not a typo to silently fix; it's the current player-facing string.

**Locale**:
The League client's own display language (e.g. `en_US`), read by inspecting the running League process's `--locale=` command-line argument. This is auto-detected from the OS process, never a user-facing setting in `Config`.
_Avoid_: language, config.Locale

**Daemon**:
The long-running background process itself, launched at Windows startup or manually. It lives in the system tray and keeps running until it crashes or the user explicitly quits it, from the tray or through the `Close Action`. Opening the GUI window never starts or stops it. See [ADR-0002](./docs/adr/0002-daemon-never-exits-on-connection-loss.md).
_Avoid_: app, process, service

**Connection Supervisor**:
The component that owns retrying a connection (to Discord, or to the LCU) forever. A closed League client or a closed Discord client is expected: it resumes the retry loop, never exits the Daemon. No LCU connection means no Discord presence, cleared immediately with no grace period. The one exception: while League's process is running but the LCU hasn't connected yet, the placeholder rotation shows instead (see `Updater`). The Discord Connection Supervisor is additionally gated on League's own process: it never connects to Discord, and disconnects if already connected, whenever League isn't running. See [ADR-0003](./docs/adr/0003-discord-connection-gated-by-league.md).
_Avoid_: reconnect logic, watchdog

**GUI**:
The Wails-hosted settings-and-status window. It runs in the same process as the Daemon, not a separate one; there is no GUI-to-Daemon IPC. See [ADR-0004](./docs/adr/0004-gui-process-hosts-the-daemon.md).
_Avoid_: frontend (for the process), app

**Close Action**:
What the window's close button does, held in `Config` as `behavior.close_action`. `ask` raises an in-app confirmation, `tray` hides the window silently, `quit` stops the Daemon. Closing the window never ends the Daemon under `ask` or `tray`.
_Avoid_: close behavior, exit mode

**Tray**:
The system-tray icon the Daemon owns. Left-click shows the `GUI`. Right-click menu is `Open`, `Pause presence`, `Quit`. Apart from `Quit` and a `quit` `Close Action`, nothing stops the Daemon.

**Pause**:
A runtime-only flag, never written to `Config`. While paused, Discord presence is cleared immediately, the same as League not running. It resets to unpaused whenever the Daemon restarts, so a forgotten pause can't outlive a session.

**Startup Launch**:
The `HKCU\...\CurrentVersion\Run` registration that makes Windows start the app at logon. It is reconciled to `Config`'s launch-at-startup setting on every launch. A run started this way opens hidden to the `Tray`, not with the `GUI` shown.
_Avoid_: autostart, Task Scheduler (it is not used)

**Presence Template**:
A user-editable `details` and `state` string pair, one pair per context: in-client, champ select, in-game, spectating (the `Watching` phase). Templates use plain `{token}` substitution, no logic. One Go engine renders them for real sends and for the `GUI` preview; there is no second implementation. A token with no data at send time resolves to empty and surrounding whitespace and separators collapse.
_Avoid_: format string, text/template (it is not Go's template package)

**Status Snapshot**:
The read model the `GUI` renders: the three connection states (League process, LCU, Discord), the current `GameFlowPhase`, and the presence the `Updater` last sent. Assembled from `State` and the supervisors' accessors and pushed to the `GUI` on change. The preview is always the last-sent presence, never a recomputation, so it cannot drift from what Discord actually shows.

**Onboarding**:
The full-screen flow shown in place of the normal `GUI` (sidebar and sections) until `Config`'s `onboarding_complete` is set. Cannot be skipped or dismissed early; the only way through is completing every screen, though it can be replayed later by clearing the flag from the GUI's help section. Four screens: an intro, a settings screen for `show_rank`/`show_stats`/`show_emojis` shown beside fixed, hardcoded demo presence cards, a start-with-Windows recommendation, and a closing screen. The demo presence cards are fabricated display data, not a `Status Snapshot`: unlike the Status Snapshot preview, they are not derived from anything ever sent to Discord, because no real presence may exist yet on first launch.
_Avoid_: walkthrough, first-run wizard, tour

**App Update**:
The in-app self-update flow: check GitHub Releases, download the new binary, verify its ed25519 signature against the `SHA256SUMS` sidecar, swap it, restart. Distinct from `Updater`, which is only about Discord presence sends. Downloads happen on user click, never silently in the background. See [ADR-0005](./docs/adr/0005-in-app-update-via-signed-binary-swap.md).
_Avoid_: Updater, upgrader
