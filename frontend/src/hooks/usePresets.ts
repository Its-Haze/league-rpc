import { GetPresets } from "../../bindings/github.com/its-haze/league-rpc/cmd/league-rpc-gui/guiservice";
import { createExternalStore } from "./createExternalStore";

// Presets are static for the app's lifetime, so this fetches once and shares
// the result, rather than re-fetching on every Advanced screen remount.
const store = createExternalStore<Record<string, string | undefined>>({}, () => {
  GetPresets()
    .then((p) => store.set(p ?? {}))
    .catch(() => {});
});

export function usePresets(): Record<string, string | undefined> {
  return store.useValue();
}
