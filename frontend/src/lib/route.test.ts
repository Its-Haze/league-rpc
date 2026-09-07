import { describe, expect, it } from "vitest";
import { initialRoute, isSection, routeReducer, SECTIONS } from "./route";

describe("routeReducer", () => {
  it("starts on home", () => {
    expect(initialRoute.section).toBe("home");
  });

  it("navigates to the requested section", () => {
    const next = routeReducer(initialRoute, { type: "navigate", section: "advanced" });
    expect(next.section).toBe("advanced");
  });

  it("returns the same state object when navigating to the current section", () => {
    const state = { section: "display" as const };
    const next = routeReducer(state, { type: "navigate", section: "display" });
    expect(next).toBe(state);
  });

  it("covers every sidebar section", () => {
    for (const section of SECTIONS) {
      expect(routeReducer(initialRoute, { type: "navigate", section }).section).toBe(section);
    }
  });
});

describe("isSection", () => {
  it("accepts every known section", () => {
    for (const section of SECTIONS) {
      expect(isSection(section)).toBe(true);
    }
  });

  it("rejects unknown values", () => {
    expect(isSection("settings")).toBe(false);
    expect(isSection("")).toBe(false);
  });
});
