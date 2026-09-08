import { GetPreviewAssets } from "../../bindings/github.com/its-haze/league-rpc/cmd/league-rpc-gui/guiservice";
import type { PreviewAssets } from "../../bindings/github.com/its-haze/league-rpc/internal/app/models";
import { createExternalStore } from "./createExternalStore";

// Static for the app's lifetime, so fetched once and shared, same reasoning
// as usePresets.
const store = createExternalStore<PreviewAssets | null>(null, () => {
  GetPreviewAssets()
    .then((a) => store.set(a))
    .catch(() => {});
});

export function usePreviewAssets(): PreviewAssets | null {
  return store.useValue();
}
