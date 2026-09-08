import { describe, expect, it } from "vitest";
import { DefaultConfig } from "./testFixtures";
import { withCloseAction, withLaunchAtStartup } from "./behaviorPatch";

describe("withLaunchAtStartup", () => {
  it("sets the flag and keeps sibling behavior fields", () => {
    const cfg = DefaultConfig();
    const patch = withLaunchAtStartup(cfg, true);
    expect(patch.behavior).toEqual({ ...cfg.behavior, launch_at_startup: true });
  });
});

describe("withCloseAction", () => {
  it("sets the action and keeps sibling behavior fields", () => {
    const cfg = DefaultConfig();
    const patch = withCloseAction(cfg, "quit");
    expect(patch.behavior).toEqual({ ...cfg.behavior, close_action: "quit" });
  });
});
