import { useEffect, useState } from "react";
import { Events } from "@wailsio/runtime";

const CLOSE_REQUESTED_EVENT = "window:close-requested";

export interface UseCloseRequestResult {
  /** True while the window close is cancelled and waiting on an answer. */
  pending: boolean;
  /** Dismiss the dialog, leaving the window open. */
  dismiss: () => void;
}

// Raised by the backend when a close arrives and close_action is "ask". The
// window stays up until the dialog answers, so nothing here has to restore it.
export function useCloseRequest(): UseCloseRequestResult {
  const [pending, setPending] = useState(false);

  useEffect(() => Events.On(CLOSE_REQUESTED_EVENT, () => setPending(true)), []);

  return { pending, dismiss: () => setPending(false) };
}
