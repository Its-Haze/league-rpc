import type { StatusSnapshot } from "../../bindings/github.com/its-haze/league-rpc/internal/app/models";

export type ConnectionTone = "ok" | "warn" | "idle";

export interface ConnectionSummary {
  label: string;
  tone: ConnectionTone;
}

// Collapses the status snapshot into one sidebar line. Order is precedence:
// an explicit pause outranks everything, and no League means nothing else runs.
export function summarizeConnection(status: StatusSnapshot | null): ConnectionSummary {
  if (!status) return { label: "Starting up", tone: "idle" };
  if (status.paused) return { label: "Paused", tone: "warn" };
  if (!status.league_process) return { label: "League closed", tone: "idle" };
  // A stalled LCU looks identical to a slow start-up for the first while,
  // so the label only hardens once the daemon says it has waited long enough.
  if (!status.lcu_connected) {
    if (status.lcu_stalled) return { label: "Can't reach League", tone: "warn" };
    return { label: "Connecting", tone: "warn" };
  }
  if (!status.discord_connected) return { label: "No Discord", tone: "warn" };
  return { label: "Connected", tone: "ok" };
}
