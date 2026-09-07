import { useState } from "react";
import { Events } from "@wailsio/runtime";
import {
  ApplySettings,
  GetSettings,
} from "../../bindings/github.com/its-haze/league-rpc/cmd/league-rpc-gui/guiservice";
import type { Config } from "../../bindings/github.com/its-haze/league-rpc/internal/config/models";
import { createExternalStore } from "./createExternalStore";

const CONFIG_CHANGED_EVENT = "settings:changed";

export interface UseSettingsResult {
  cfg: Config | null;
  error: string | null;
  saving: boolean;
  /** Shallow-merges patch onto the current config and persists it. Nested
   * groups (display, presence, ...) must be spread by the caller. */
  applyPatch: (patch: Partial<Config>) => Promise<void>;
}

interface SettingsState {
  cfg: Config | null;
  error: string | null;
}

// Shared across every screen, not one store per caller: otherwise two
// screens mounted at once could each hold a stale copy and clobber writes.
const store = createExternalStore<SettingsState>({ cfg: null, error: null }, () => {
  GetSettings()
    .then((c) => store.set({ ...store.get(), cfg: c }))
    .catch((e: unknown) => store.set({ ...store.get(), error: String(e) }));

  Events.On(CONFIG_CHANGED_EVENT, (ev: { data: Config }) => store.set({ cfg: ev.data, error: null }));
});

// The live settings tree every settings screen binds to: loaded once,
// applied through the daemon, and kept in sync with settings:changed.
export function useSettings(): UseSettingsResult {
  const snapshot = store.useValue();
  const [saving, setSaving] = useState(false);

  async function applyPatch(patch: Partial<Config>) {
    const current = store.get().cfg;
    if (!current) return;
    const next: Config = { ...current, ...patch };
    store.set({ ...store.get(), cfg: next });
    setSaving(true);
    try {
      await ApplySettings(next);
      store.set({ ...store.get(), error: null });
    } catch (e) {
      // Refetch rather than revert to the pre-patch snapshot: another
      // screen's write may have landed since, and reverting would drop it.
      try {
        store.set({ cfg: await GetSettings(), error: String(e) });
      } catch {
        store.set({ ...store.get(), error: String(e) });
      }
    } finally {
      setSaving(false);
    }
  }

  return { cfg: snapshot.cfg, error: snapshot.error, saving, applyPatch };
}
