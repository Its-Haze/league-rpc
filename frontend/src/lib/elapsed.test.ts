import { describe, expect, it } from "vitest";
import { formatElapsed } from "./elapsed";

describe("formatElapsed", () => {
  it("formats under an hour as M:SS", () => {
    const start = 1000;
    expect(formatElapsed(start, (start + 125) * 1000)).toBe("2:05");
  });

  it("formats an hour or more as H:MM:SS", () => {
    const start = 1000;
    expect(formatElapsed(start, (start + 3725) * 1000)).toBe("1:02:05");
  });

  it("pads minutes and seconds under 10", () => {
    const start = 1000;
    expect(formatElapsed(start, (start + 65) * 1000)).toBe("1:05");
  });

  it("returns null for a non-positive start", () => {
    expect(formatElapsed(0, Date.now())).toBeNull();
    expect(formatElapsed(-5, Date.now())).toBeNull();
  });

  it("returns null for a start in the future", () => {
    const start = 10_000;
    expect(formatElapsed(start, (start - 5) * 1000)).toBeNull();
  });

  it("returns 0:00 right at the start", () => {
    const start = 1000;
    expect(formatElapsed(start, start * 1000)).toBe("0:00");
  });
});
