import { describe, expect, it } from "vitest";
import { PRESENCE_CONTEXTS, isPresenceContext } from "./presenceContexts";

describe("isPresenceContext", () => {
  it("accepts every known context", () => {
    for (const ctx of PRESENCE_CONTEXTS) {
      expect(isPresenceContext(ctx)).toBe(true);
    }
  });

  it("rejects unknown values", () => {
    expect(isPresenceContext("bogus")).toBe(false);
    expect(isPresenceContext("")).toBe(false);
  });
});
