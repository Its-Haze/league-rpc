import { describe, expect, it } from "vitest";
import type { StatusSnapshot } from "../../bindings/github.com/its-haze/league-rpc/internal/app/models";
import { summarizeConnection } from "./connectionStatus";

function snapshot(over: Partial<StatusSnapshot> = {}): StatusSnapshot {
  return {
    league_process: true,
    lcu_connected: true,
    discord_connected: true,
    paused: false,
    gameflow_phase: "None",
    presence: null,
    presence_cleared: false,
    ...over,
  };
}

describe("summarizeConnection", () => {
  it("reports connected when everything is up", () => {
    expect(summarizeConnection(snapshot())).toEqual({ label: "Connected", tone: "ok" });
  });

  it("shows a placeholder before the first snapshot lands", () => {
    expect(summarizeConnection(null).tone).toBe("idle");
  });

  it("puts pause ahead of every connection state", () => {
    const s = snapshot({ paused: true, league_process: false, discord_connected: false });
    expect(summarizeConnection(s).label).toBe("Paused");
  });

  it("treats a closed League as idle, not a fault", () => {
    expect(summarizeConnection(snapshot({ league_process: false }))).toEqual({
      label: "League closed",
      tone: "idle",
    });
  });

  it("distinguishes a pending LCU from a missing Discord", () => {
    expect(summarizeConnection(snapshot({ lcu_connected: false })).label).toBe("Connecting");
    expect(summarizeConnection(snapshot({ discord_connected: false })).label).toBe("No Discord");
  });
});
