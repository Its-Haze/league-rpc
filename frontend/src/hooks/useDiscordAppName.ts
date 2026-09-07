import { useEffect, useState } from "react";
import { GetApplicationName } from "../../bindings/github.com/its-haze/league-rpc/cmd/league-rpc-gui/guiservice";

// Cached by id across every caller: Home and Advanced can both be resolving
// the same app id at once, and the id rarely changes.
const cache = new Map<string, string | null>();

// Resolves a Discord Application ID to its public display name via
export function useDiscordAppName(appId: string): string | null {
  const [name, setName] = useState<string | null>(cache.get(appId) ?? null);

  useEffect(() => {
    if (!appId) {
      setName(null);
      return;
    }
    const cached = cache.get(appId);
    if (cached !== undefined) {
      setName(cached);
      return;
    }

    let cancelled = false;
    GetApplicationName(appId)
      .then((n) => {
        cache.set(appId, n);
        if (!cancelled) setName(n);
      })
      .catch(() => {
        cache.set(appId, null);
        if (!cancelled) setName(null);
      });
    return () => {
      cancelled = true;
    };
  }, [appId]);

  return name;
}
