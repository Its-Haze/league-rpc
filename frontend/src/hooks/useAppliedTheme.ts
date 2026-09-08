import { useEffect, useState } from "react";
import { resolveTheme } from "../lib/theme";

function prefersDarkNow(): boolean {
  return window.matchMedia?.("(prefers-color-scheme: dark)").matches ?? false;
}

// Resolves Config.theme against the live OS preference and writes the result
// to <html data-theme="light"|"dark">, which tokens.css switches on.
export function useAppliedTheme(setting: string): "light" | "dark" {
  const [prefersDark, setPrefersDark] = useState(prefersDarkNow);

  useEffect(() => {
    const media = window.matchMedia("(prefers-color-scheme: dark)");
    const onChange = (e: MediaQueryListEvent) => setPrefersDark(e.matches);
    media.addEventListener("change", onChange);
    return () => media.removeEventListener("change", onChange);
  }, []);

  const resolved = resolveTheme(setting, prefersDark);

  useEffect(() => {
    document.documentElement.dataset.theme = resolved;
  }, [resolved]);

  return resolved;
}
