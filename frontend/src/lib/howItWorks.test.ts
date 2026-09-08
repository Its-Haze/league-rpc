import { describe, expect, it } from "vitest";
import { howItWorksCopy } from "./howItWorks";

const COMBINATIONS = [
  { launchAtStartup: true, notifyUpdates: true },
  { launchAtStartup: true, notifyUpdates: false },
  { launchAtStartup: false, notifyUpdates: true },
  { launchAtStartup: false, notifyUpdates: false },
];

describe("howItWorksCopy", () => {
  it("always describes the same four stages", () => {
    for (const inputs of COMBINATIONS) {
      const { steps } = howItWorksCopy(inputs);
      expect(steps.map((s) => s.title)).toEqual([
        "It starts",
        "It waits",
        "League opens, it connects",
        "League closes, it sleeps",
      ]);
      expect(steps.every((s) => s.body.length > 0)).toBe(true);
    }
  });

  it("says the app starts itself when launch at startup is on", () => {
    const { steps } = howItWorksCopy({ launchAtStartup: true, notifyUpdates: true });
    expect(steps[0].body).toContain("starts with Windows");
  });

  it("says the user starts it, at any time, when launch at startup is off", () => {
    const { steps } = howItWorksCopy({ launchAtStartup: false, notifyUpdates: true });
    expect(steps[0].body).toContain("You start League RPC yourself");
    expect(steps[0].body).toContain("mid-game");
  });

  it("only the first stage changes with the startup toggle", () => {
    const on = howItWorksCopy({ launchAtStartup: true, notifyUpdates: true }).steps;
    const off = howItWorksCopy({ launchAtStartup: false, notifyUpdates: true }).steps;
    expect(on.slice(1)).toEqual(off.slice(1));
  });

  it("points at toasts when update notifications are on", () => {
    expect(howItWorksCopy({ launchAtStartup: true, notifyUpdates: true }).updates).toContain("toast");
  });

  it("points at the Discord server when update notifications are off", () => {
    expect(howItWorksCopy({ launchAtStartup: true, notifyUpdates: false }).updates).toContain("Discord server");
  });
});
