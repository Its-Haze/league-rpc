import { AppShell } from "./components/shell/AppShell";
import { useAppliedTheme } from "./hooks/useAppliedTheme";
import { useSettings } from "./hooks/useSettings";
import type { ThemeSetting } from "./lib/theme";

export default function App() {
  const { cfg, error, applyPatch } = useSettings();
  useAppliedTheme(cfg?.theme ?? "system");

  function handleThemeChange(theme: ThemeSetting) {
    void applyPatch({ theme });
  }

  return (
    <AppShell
      theme={cfg?.theme ?? "system"}
      onThemeChange={handleThemeChange}
      themeDisabled={!cfg}
      error={error}
    />
  );
}
