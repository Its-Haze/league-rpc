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
  if (!status.lcu_connected) return { label: "Connecting", tone: "warn" };
  if (!status.discord_connected) return { label: "No Discord", tone: "warn" };
  return { label: "Connected", tone: "ok" };
}
