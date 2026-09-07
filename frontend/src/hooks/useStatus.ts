import { Events } from "@wailsio/runtime";
import { GetStatus } from "../../bindings/github.com/its-haze/league-rpc/cmd/league-rpc-gui/guiservice";
import type { StatusSnapshot } from "../../bindings/github.com/its-haze/league-rpc/internal/app/models";
import { createExternalStore } from "./createExternalStore";

const STATUS_CHANGED_EVENT = "status:changed";

// Shared across screens (Behavior's pause toggle, Home's onboarding step), so
// a per-call fetch+listener would mean duplicate round trips for the same data.
const store = createExternalStore<StatusSnapshot | null>(null, () => {
  GetStatus()
    .then((s) => store.set(s))
    .catch(() => {});

  Events.On(STATUS_CHANGED_EVENT, (ev: { data: StatusSnapshot }) => store.set(ev.data));
});

// The live status snapshot: loaded once, kept in sync with status:changed.
export function useStatus(): StatusSnapshot | null {
  return store.useValue();
}
