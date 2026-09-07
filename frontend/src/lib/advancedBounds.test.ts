import { describe, expect, it } from "vitest";
import {
  CUSTOM_PRESET_VALUE,
  STATS_POLLING_INTERVAL_BOUNDS,
  UPDATE_INTERVAL_BOUNDS,
  clampToBounds,
  formatIntervalSeconds,
  isCustomAppId,
  presetNameFor,
  presetSelectValue,
  resolveSelectValue,
} from "./advancedBounds";

describe("clampToBounds", () => {
  it("passes values already in range through unchanged", () => {
    expect(clampToBounds(1500, UPDATE_INTERVAL_BOUNDS)).toBe(1500);
  });

  it("clamps below the minimum", () => {
    expect(clampToBounds(10, UPDATE_INTERVAL_BOUNDS)).toBe(500);
  });

  it("clamps above the maximum", () => {
    expect(clampToBounds(999999, STATS_POLLING_INTERVAL_BOUNDS)).toBe(30000);
  });

  it("falls back to the minimum for NaN", () => {
    expect(clampToBounds(NaN, UPDATE_INTERVAL_BOUNDS)).toBe(500);
  });
});

describe("formatIntervalSeconds", () => {
  it("formats a whole-second value with no decimal", () => {
    expect(formatIntervalSeconds(10000)).toBe("10s");
  });

  it("formats a fractional value to one decimal place", () => {
    expect(formatIntervalSeconds(1500)).toBe("1.5s");
  });

  it("rounds to the nearest tenth of a second", () => {
    expect(formatIntervalSeconds(1540)).toBe("1.5s");
    expect(formatIntervalSeconds(1560)).toBe("1.6s");
  });
});

describe("presetNameFor", () => {
  const presets = { "League of Legends": "111", "League of Kittens": "222" };

  it("finds the preset matching the id", () => {
    expect(presetNameFor("222", presets)).toBe("League of Kittens");
  });

  it("returns null for a custom id", () => {
    expect(presetNameFor("999", presets)).toBeNull();
  });
});

describe("presetSelectValue", () => {
  const presets = { "League of Legends": "111" };

  it("returns the preset name on a match", () => {
    expect(presetSelectValue("111", presets)).toBe("League of Legends");
  });

  it("returns the custom sentinel otherwise", () => {
    expect(presetSelectValue("777", presets)).toBe(CUSTOM_PRESET_VALUE);
  });
});

describe("resolveSelectValue", () => {
  const presets = { "League of Kittens": "222" };

  it("shows the matching preset when not editing custom", () => {
    expect(resolveSelectValue("222", presets, false)).toBe("League of Kittens");
  });

  it("shows Custom while editing, even if the id still matches a preset", () => {
    // Regression: switching to Custom right after picking a preset must not
    // immediately snap back to that preset's name.
    expect(resolveSelectValue("222", presets, true)).toBe(CUSTOM_PRESET_VALUE);
  });

  it("shows Custom for an unmatched id regardless of editingCustom", () => {
    expect(resolveSelectValue("999", presets, false)).toBe(CUSTOM_PRESET_VALUE);
    expect(resolveSelectValue("999", presets, true)).toBe(CUSTOM_PRESET_VALUE);
  });
});

describe("isCustomAppId", () => {
  const presets = { "League of Kittens": "222" };

  it("is false for a preset id", () => {
    expect(isCustomAppId("222", presets)).toBe(false);
  });

  it("is true for anything else", () => {
    expect(isCustomAppId("999", presets)).toBe(true);
  });
});
