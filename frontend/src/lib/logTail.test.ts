import { describe, expect, it } from "vitest";
import { MAX_TAIL_LINES, appendLine, appendLines, isScrolledToBottom } from "./logTail";

describe("appendLine", () => {
  it("appends a line", () => {
    expect(appendLine(["a"], "b")).toEqual(["a", "b"]);
  });

  it("drops the oldest lines once past the cap", () => {
    const lines = Array.from({ length: MAX_TAIL_LINES }, (_, i) => String(i));
    const next = appendLine(lines, "new");
    expect(next.length).toBe(MAX_TAIL_LINES);
    expect(next[0]).toBe("1");
    expect(next[next.length - 1]).toBe("new");
  });
});

describe("appendLines", () => {
  it("appends a batch in one pass", () => {
    expect(appendLines(["a"], ["b", "c"])).toEqual(["a", "b", "c"]);
  });

  it("is a no-op for an empty batch", () => {
    const lines = ["a", "b"];
    expect(appendLines(lines, [])).toBe(lines);
  });

  it("drops the oldest lines once the batch pushes past the cap", () => {
    const lines = Array.from({ length: MAX_TAIL_LINES - 1 }, (_, i) => String(i));
    const next = appendLines(lines, ["x", "y", "z"]);
    expect(next.length).toBe(MAX_TAIL_LINES);
    expect(next[next.length - 1]).toBe("z");
  });
});

describe("isScrolledToBottom", () => {
  it("is true when scrolled all the way down", () => {
    expect(isScrolledToBottom(100, 50, 150)).toBe(true);
  });

  it("is true within tolerance", () => {
    expect(isScrolledToBottom(80, 50, 150, 24)).toBe(true);
  });

  it("is false once scrolled up past tolerance", () => {
    expect(isScrolledToBottom(0, 50, 150, 24)).toBe(false);
  });
});
