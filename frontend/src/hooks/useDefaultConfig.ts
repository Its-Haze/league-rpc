import { GetDefaultConfig } from "../../bindings/github.com/its-haze/league-rpc/cmd/league-rpc-gui/guiservice";
import type { Config } from "../../bindings/github.com/its-haze/league-rpc/internal/config/models";
import { createExternalStore } from "./createExternalStore";

// The built-in default settings tree, fetched once: it never changes at
// runtime, so every "reset to default" control can compare against it.
const store = createExternalStore<Config | null>(null, () => {
  GetDefaultConfig()
    .then((c) => store.set(c))
    .catch(() => {});
});

export function useDefaultConfig(): Config | null {
  return store.useValue();
}
