import { useState } from "react";
import { CheckForUpdates } from "../../bindings/github.com/its-haze/league-rpc/cmd/league-rpc-gui/guiservice";

export interface UseCheckForUpdatesResult {
  checking: boolean;
  result: string | null;
  check: () => Promise<void>;
}

// The manual "Check for updates" action, shared by the About and Behavior
// screens so the checking/result state and its wording live in one place.
export function useCheckForUpdates(): UseCheckForUpdatesResult {
  const [checking, setChecking] = useState(false);
  const [result, setResult] = useState<string | null>(null);

  async function check() {
    setChecking(true);
    setResult(null);
    try {
      const status = await CheckForUpdates();
      setResult(status.available ? `Version ${status.version} is available.` : "You're up to date.");
    } catch (e) {
      setResult(String(e));
    } finally {
      setChecking(false);
    }
  }

  return { checking, result, check };
}
