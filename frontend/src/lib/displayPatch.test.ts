import { describe, expect, it } from "vitest";
import { DefaultConfig } from "./testFixtures";
import { withShowEmojis, withShowInClient, withShowRank, withShowStats } from "./displayPatch";

describe("withShowRank", () => {
  it("sets show_rank and keeps sibling display fields", () => {
    const cfg = DefaultConfig();
    const patch = withShowRank(cfg, false);
    expect(patch.display).toEqual({
      ...cfg.display,
      default: { ...cfg.display.default, show_rank: false },
    });
  });
});

describe("withShowStats", () => {
  it("sets show_stats and keeps sibling display fields", () => {
    const cfg = DefaultConfig();
    const patch = withShowStats(cfg, false);
    expect(patch.display).toEqual({
      ...cfg.display,
      default: { ...cfg.display.default, show_stats: false },
    });
  });
});

describe("withShowEmojis", () => {
  it("sets show_emojis and keeps sibling presence fields", () => {
    const cfg = DefaultConfig();
    const patch = withShowEmojis(cfg, false);
    expect(patch.presence).toEqual({ ...cfg.presence, show_emojis: false });
  });
});

describe("withShowInClient", () => {
  it("sets show_in_client and keeps sibling presence fields", () => {
    const cfg = DefaultConfig();
    const patch = withShowInClient(cfg, false);
    expect(patch.presence).toEqual({ ...cfg.presence, show_in_client: false });
  });
});
