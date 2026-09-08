import { describe, expect, it } from "vitest";
import { isThemeSetting, resolveTheme } from "./theme";

describe("resolveTheme", () => {
  it("light and dark win outright, regardless of the OS preference", () => {
    expect(resolveTheme("light", true)).toBe("light");
    expect(resolveTheme("light", false)).toBe("light");
    expect(resolveTheme("dark", true)).toBe("dark");
    expect(resolveTheme("dark", false)).toBe("dark");
  });

  it("system follows the OS preference", () => {
    expect(resolveTheme("system", true)).toBe("dark");
    expect(resolveTheme("system", false)).toBe("light");
  });

  it("treats an unrecognized setting as system", () => {
    expect(resolveTheme("", true)).toBe("dark");
    expect(resolveTheme("sepia", false)).toBe("light");
  });
});

describe("isThemeSetting", () => {
  it("accepts the three known settings", () => {
    expect(isThemeSetting("system")).toBe(true);
    expect(isThemeSetting("light")).toBe(true);
    expect(isThemeSetting("dark")).toBe(true);
  });

  it("rejects anything else", () => {
    expect(isThemeSetting("sepia")).toBe(false);
    expect(isThemeSetting("")).toBe(false);
  });
});
