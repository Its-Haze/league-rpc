import { Events } from "@wailsio/runtime";
import { GetUpdateStatus } from "../../bindings/github.com/its-haze/league-rpc/cmd/league-rpc-gui/guiservice";
import type { UpdateStatus } from "../../bindings/github.com/its-haze/league-rpc/internal/app/models";
import { createExternalStore } from "./createExternalStore";

// Also emitted directly by main.go whenever the App Update coordinator's
// status changes (launch check, periodic check, or a download settling).
export const UPDATE_CHANGED_EVENT = "update:changed";

// Shared across the sidebar badge and the About screen's banner, so both
// read the exact same status instead of keeping their own copies in sync.
const store = createExternalStore<UpdateStatus | null>(null, () => {
  GetUpdateStatus()
    .then((s) => store.set(s))
    .catch(() => {});

  Events.On(UPDATE_CHANGED_EVENT, (ev: { data: UpdateStatus }) => store.set(ev.data));
});

// The live App Update status: loaded once, kept in sync with update:changed.
export function useUpdateStatus(): UpdateStatus | null {
  return store.useValue();
}
