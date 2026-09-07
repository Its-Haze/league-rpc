import { GetConfigBounds } from "../../bindings/github.com/its-haze/league-rpc/cmd/league-rpc-gui/guiservice";
import { UPDATE_INTERVAL_BOUNDS, STATS_POLLING_INTERVAL_BOUNDS, type Bounds } from "../lib/advancedBounds";
import { createExternalStore } from "./createExternalStore";

export interface ConfigBoundsResult {
  updateInterval: Bounds;
  statsPollingInterval: Bounds;
}

// Read from the backend so a bound tuned in internal/config can't drift from
// the value this screen actually clamps and validates against.
const store = createExternalStore<ConfigBoundsResult>(
  { updateInterval: UPDATE_INTERVAL_BOUNDS, statsPollingInterval: STATS_POLLING_INTERVAL_BOUNDS },
  () => {
    GetConfigBounds()
      .then((b) =>
        store.set({
          updateInterval: { min: b.update_interval_min, max: b.update_interval_max },
          statsPollingInterval: { min: b.stats_polling_interval_min, max: b.stats_polling_interval_max },
        }),
      )
      .catch(() => {});
  },
);

export function useConfigBounds(): ConfigBoundsResult {
  return store.useValue();
}
