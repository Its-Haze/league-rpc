import { describe, expect, it } from "vitest";
import { formatDiagnostics } from "./diagnostics";

describe("formatDiagnostics", () => {
  it("wraps the text in a Markdown code fence", () => {
    expect(formatDiagnostics("Version: 1.0.0")).toBe("```\nVersion: 1.0.0\n```");
  });

  it("trims surrounding whitespace before fencing", () => {
    expect(formatDiagnostics("\n  Version: 1.0.0\n\n")).toBe("```\nVersion: 1.0.0\n```");
  });
});
