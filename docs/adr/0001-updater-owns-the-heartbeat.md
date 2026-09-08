# Updater owns debounce, heartbeat, and reclaim as one goroutine

**Status**: accepted

League's own client silently overwrites our Discord presence if we stop sending updates. Fighting that requires three behaviors working together: an initial debounced send, a periodic heartbeat resend so our presence keeps winning over League's, and a faster "reclaim burst" right after any real change, all serialized so they never race each other on the same Discord connection.

We're consolidating all of that into the existing `discord.Updater` (`internal/discord/updater.go`), which already debounces `State` changes behind a mutex. The debounce timer becomes a repeating ticker instead of a one-shot: after the first real send, `Updater` keeps resending the current presence on a heartbeat cadence, with a shorter interval for a few cycles right after a real change (the "reclaim burst"), until the next real change resets it.

`discord.Client.UpdatePresence`/`ClearPresence` must only ever be called through `Updater`. No separate goroutine touches the Discord connection directly, so `Updater`'s existing `sync.Mutex` is sufficient to serialize debounce, heartbeat, and reclaim against each other.

Considered running debounce, heartbeat, and reclaim as separate goroutines, each owning its own timer, for a cleaner separation of concerns. Rejected it: a single goroutine driven by one ticker covers all three with one clock and one lock, and splitting them apart would only add coordination overhead between goroutines that all touch the same Discord connection anyway.
