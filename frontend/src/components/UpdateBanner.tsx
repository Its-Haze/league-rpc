import { useEffect, useState } from "react";
import { Events } from "@wailsio/runtime";
import DOMPurify from "dompurify";
import {
  GetChangelog,
  RestartForUpdate,
  RetryUpdate,
} from "../../bindings/github.com/its-haze/league-rpc/cmd/league-rpc-gui/guiservice";
import { useUpdateStatus } from "../hooks/useUpdateStatus";
import { handleExternalLinkClick } from "../lib/links";
import { renderMarkdown } from "../lib/markdown";

// Wails' own updater engine emits this directly on the same event bus, purely
// for a live byte count; everything else comes from our own update:changed.
const EVENT_DOWNLOAD_PROGRESS = "wails:updater:download-progress";

interface Progress {
  written: number;
  total: number;
}

export default function UpdateBanner() {
  const status = useUpdateStatus();
  const [dismissedVersion, setDismissedVersion] = useState<string | null>(null);
  const [progress, setProgress] = useState<Progress | null>(null);
  const [retrying, setRetrying] = useState(false);
  const [restartError, setRestartError] = useState<string | null>(null);
  const [changelog, setChangelog] = useState<string | null>(null);
  const [showChangelog, setShowChangelog] = useState(false);

  useEffect(() => {
    // A dismissal lasts only until the next status update, background poll
    // or manual check alike.
    setDismissedVersion(null);
  }, [status]);

  useEffect(() => {
    return Events.On(EVENT_DOWNLOAD_PROGRESS, (ev: { data: Progress }) => setProgress(ev.data));
  }, []);

  if (!status?.available || status.version === dismissedVersion) {
    return null;
  }

  async function handleRetry() {
    setRetrying(true);
    try {
      await RetryUpdate();
    } finally {
      setRetrying(false);
    }
  }

  async function handleRestart() {
    setRestartError(null);
    try {
      await RestartForUpdate();
    } catch (e) {
      setRestartError(String(e));
    }
  }

  async function handleToggleChangelog() {
    const next = !showChangelog;
    setShowChangelog(next);
    if (next && changelog === null) {
      setChangelog(await GetChangelog());
    }
  }

  return (
    <div className="border-accent bg-surface-raised flex flex-col gap-3 rounded-lg border p-4 text-sm">
      <div className="flex items-center justify-between gap-4">
        <div>
          <strong>League RPC {status.version}</strong> is available.
          {restartError ? (
            <span className="text-danger ml-2">restart failed: {restartError}</span>
          ) : status.ready ? (
            <span className="text-ok ml-2">ready to install</span>
          ) : status.last_error ? (
            <span className="text-danger ml-2">{status.last_error}</span>
          ) : (
            <span className="text-muted ml-2">
              downloading{progress ? progressPercent(progress) : "..."}
            </span>
          )}
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={handleToggleChangelog}
            className="text-muted hover:text-text press"
          >
            {showChangelog ? "Hide changelog" : "Changelog"}
          </button>
          {status.ready ? (
            <button
              onClick={handleRestart}
              className="bg-accent text-accent-text press rounded-sm px-3 py-1"
            >
              Install & Restart
            </button>
          ) : (
            status.last_error && (
              <button
                onClick={handleRetry}
                disabled={retrying}
                className="bg-accent text-accent-text press rounded-sm px-3 py-1 disabled:opacity-65"
              >
                {retrying ? "Retrying..." : "Retry"}
              </button>
            )
          )}
          <button
            onClick={() => setDismissedVersion(status.version)}
            aria-label="Dismiss"
            className="text-muted hover:text-text press px-2 py-1 text-base"
          >
            {"×"}
          </button>
        </div>
      </div>

      {showChangelog && (
        <div
          className="changelog border-border text-sm border-t pt-3"
          onClick={handleExternalLinkClick}
          dangerouslySetInnerHTML={{
            __html: DOMPurify.sanitize(renderMarkdown(changelog ?? "Loading changelog...")),
          }}
        />
      )}
    </div>
  );
}

function progressPercent(p: Progress): string {
  if (p.total <= 0) return "";
  return ` ${Math.round((p.written / p.total) * 100)}%`;
}
