# The Daemon never exits on connection loss: it supervises reconnects forever

**Status**: accepted

The target shape is a Blitz-style background app: launched at Windows startup or manually, living in the system tray, running indefinitely. A user should never need to manually restart it, not because League isn't open yet, not because League or Discord closed mid-session, not because Discord wasn't running when the Daemon started. Only a crash or an explicit tray "Quit" should end the process.

This rules out the one-shot lifecycle both `discord.Client` and `lcu.Client` currently assume: `Connect()` once, run, `Disconnect()` on exit. Instead, each gets wrapped by a Connection Supervisor that retries indefinitely with backoff and treats "the other side closed" as an expected steady-state condition to retry through, not an error that propagates up and unwinds `main()`.

The two supervisors (Discord, LCU) run independently. League being closed must not block waiting for Discord, and vice versa, so presence can resume immediately whenever either side comes back.

**Consequences**: `main()`/the Daemon's top-level goroutine starts the supervisors and then blocks only on an explicit shutdown signal (tray quit, OS shutdown), never on a connection error path. `discord.Client.Connect()`/`lcu.Client.Connect()` keep their existing one-shot-per-attempt signatures; supervising them means calling them in a retry loop from the outside, not rewriting their internals.

No LCU connection means no presence, uniformly, with no grace period, and with exactly one exception: while League's own process is running but the LCU hasn't connected yet, the placeholder rotation from ADR-0001 shows. That single condition, process detected and LCU not yet connected, cleanly separates "genuinely launching" from "not open at all," independent of who started League (the player, or later, the Daemon's own auto-launch).

Once the LCU connects, the placeholder stops and real presence takes over. The moment the LCU disconnects, presence clears immediately and the Daemon goes back to watching for the League process, showing the placeholder again only if it reappears.

A time-based grace period on disconnect was considered and rejected. It trades a guaranteed-correct signal for a heuristic guess about why the LCU dropped. Process detection gives an actually correct signal for the one case, active launch, that legitimately deserves an idle card instead of nothing.
