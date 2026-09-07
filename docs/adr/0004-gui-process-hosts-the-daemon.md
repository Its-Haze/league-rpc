# The GUI process hosts the Daemon: one binary, no IPC

**Status**: accepted

The settings-and-status GUI (Wails v3) and the always-on Daemon run in the same process. The GUI binary calls `daemon.Wire(store, logger)` in a goroutine on launch and runs the Wails event loop on the main thread; closing the window hides it, and only the tray "Quit" ends the process. `internal/app` stays the binding seam and does not import Wails.

**Considered options**: a separate headless daemon process with the GUI as a second process talking over a local socket or named pipe. Rejected: `config.Store` already gives every in-process reader a lock-free live snapshot plus a `Subscribe()` channel, and the Daemon already exposes the connection state the GUI needs through plain accessors. A split would mean rebuilding both of those over IPC, plus a lifecycle contract for "GUI running, daemon not" and version-skew between the two binaries, for no benefit on a single-user desktop app.

**Consequences**: the daemon packages must stay importable and side-effect-free at package scope (already true). Logging can no longer assume a console: a Wails process on Windows has none, so the file sink plus in-memory ring buffer is mandatory, not a nicety. The existing `cmd/league-rpc` headless entry point stays as a dev/debug build, not a shipped artifact. A GUI crash takes the Daemon down with it, which is acceptable because a crashed presence updater is not worth keeping alive headless with no way to see or fix it.
