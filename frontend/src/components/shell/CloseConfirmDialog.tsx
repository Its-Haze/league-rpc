import { useState } from "react";
import { ResolveClose } from "../../../bindings/github.com/its-haze/league-rpc/cmd/league-rpc-gui/guiservice";
import { useSettings } from "../../hooks/useSettings";
import { withCloseAction, type CloseAction } from "../../lib/behaviorPatch";
import { Button, Dialog } from "../ui";

export interface CloseConfirmDialogProps {
  open: boolean;
  onDismiss: () => void;
}

// Replaces the OS message box that used to appear on close: the window stays
// visible behind this, and the backend waits for ResolveClose either way.
export function CloseConfirmDialog({ open, onDismiss }: CloseConfirmDialogProps) {
  const { cfg, applyPatch } = useSettings();
  const [remember, setRemember] = useState(false);

  function dismiss() {
    setRemember(false);
    onDismiss();
  }

  async function answer(action: CloseAction) {
    // Persist first: "quit" tears the process down moments later.
    if (remember && cfg) await applyPatch(withCloseAction(cfg, action));
    dismiss();
    await ResolveClose(action);
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(next) => !next && dismiss()}
      title="Close League RPC?"
      description="Presence keeps updating while the window is hidden. Quitting stops it until you launch the app again."
    >
      <label className="text-muted flex items-center gap-2 text-sm">
        <input
          type="checkbox"
          checked={remember}
          onChange={(e) => setRemember(e.target.checked)}
          className="accent-accent size-3.5"
        />
        Remember my choice, don&rsquo;t ask again
      </label>

      <div className="mt-6 flex justify-end gap-2">
        <Button variant="ghost" onClick={dismiss}>
          Cancel
        </Button>
        <Button variant="danger" onClick={() => void answer("quit")}>
          Quit
        </Button>
        {/* autoFocus so Enter takes the safe answer, not Radix's first tabbable. */}
        <Button variant="primary" autoFocus onClick={() => void answer("tray")}>
          Hide to tray
        </Button>
      </div>
    </Dialog>
  );
}
