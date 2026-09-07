import { Bug, LifeBuoy, MessageSquarePlus, PlayCircle, ScrollText } from "lucide-react";
import type { ReactNode } from "react";
import { useState } from "react";
import {
  GetDiagnostics,
  OpenLogsFolder,
} from "../../../bindings/github.com/its-haze/league-rpc/cmd/league-rpc-gui/guiservice";
import { useSettings } from "../../hooks/useSettings";
import { formatDiagnostics } from "../../lib/diagnostics";
import { BUG_REPORT_URL, DISCORD_COMMUNITY_URL, FEATURE_REQUEST_URL, openExternal } from "../../lib/links";
import { DiscordIcon } from "../icons";
import { Button, SettingsCard } from "../ui";
import { LogViewer } from "./help/LogViewer";

// The Help section: log viewer, logs folder, diagnostics, and where to go
// for support or to file an issue.
export function HelpScreen() {
  const { cfg, applyPatch } = useSettings();
  const [copyResult, setCopyResult] = useState<string | null>(null);
  const [folderError, setFolderError] = useState<string | null>(null);

  async function handleOpenFolder() {
    setFolderError(null);
    try {
      await OpenLogsFolder();
    } catch (e) {
      setFolderError(String(e));
    }
  }

  async function handleCopyDiagnostics() {
    try {
      const diag = await GetDiagnostics();
      await navigator.clipboard.writeText(formatDiagnostics(diag));
      setCopyResult("Copied to clipboard.");
    } catch (e) {
      setCopyResult(String(e));
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-xl font-semibold">Help</h1>

      <SettingsCard
        icon={ScrollText}
        title="Logs"
        description="What League RPC has been doing. Copy diagnostics grabs your version and settings too, which is what a bug report needs."
        action={
          <>
            <Button variant="secondary" onClick={handleOpenFolder}>
              Open logs folder
            </Button>
            <Button variant="secondary" onClick={handleCopyDiagnostics}>
              Copy diagnostics
            </Button>
          </>
        }
      >
        <LogViewer />
        {folderError && <p className="text-danger text-xs">{folderError}</p>}
        {copyResult && <p className="text-muted text-xs">{copyResult}</p>}
      </SettingsCard>

      <SettingsCard
        icon={LifeBuoy}
        title="Get help"
        description="Ask a question, report something broken, or suggest what to build next."
      >
        <HelpLink href={DISCORD_COMMUNITY_URL} icon={<DiscordIcon className="size-4" />}>
          Join the Discord community
        </HelpLink>
        <HelpLink href={BUG_REPORT_URL} icon={<Bug className="size-4" />}>
          Report a bug
        </HelpLink>
        <HelpLink href={FEATURE_REQUEST_URL} icon={<MessageSquarePlus className="size-4" />}>
          Request a feature
        </HelpLink>
      </SettingsCard>

      {cfg && (
        <SettingsCard
          icon={PlayCircle}
          title="First-run walkthrough"
          description="Replay the setup steps shown on first launch. Your current settings stay as they are."
          action={
            <Button variant="secondary" onClick={() => void applyPatch({ onboarding_complete: false })}>
              Replay walkthrough
            </Button>
          }
        />
      )}
    </div>
  );
}

// Mirrors the sidebar footer's link treatment (icon + label, hover row) so
// external links look consistent across the app.
function HelpLink({ href, icon, children }: { href: string; icon: ReactNode; children: ReactNode }) {
  return (
    <a
      href={href}
      onClick={(e) => {
        e.preventDefault();
        openExternal(href);
      }}
      className="text-muted hover:bg-surface-raised hover:text-text -mx-3 flex items-center gap-2 rounded-sm px-3 py-2 text-sm font-medium transition-colors"
    >
      {icon}
      {children}
    </a>
  );
}
