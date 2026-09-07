import type { MouseEvent } from "react";
import { Browser } from "@wailsio/runtime";

// External links the Help section points at. The issue templates these
// target are added under .github/ISSUE_TEMPLATE/ separately.
export const DISCORD_COMMUNITY_URL = "https://discord.haze.sh";
export const DISCORD_DEVELOPER_PORTAL_URL = "https://discord.com/developers/applications";
export const GITHUB_REPO_URL = "https://github.com/its-haze/league-rpc";
export const GITHUB_PROFILE_URL = "https://github.com/its-haze";
export const AUTHOR_WEBSITE_URL = "https://haze.sh";
export const BUG_REPORT_URL = "https://github.com/its-haze/league-rpc/issues/new?template=bug_report.md";
export const FEATURE_REQUEST_URL =
  "https://github.com/its-haze/league-rpc/issues/new?template=feature_request.md";

// The webview intercepts plain <a target="_blank"> navigation into its own
export function openExternal(url: string) {
  void Browser.OpenURL(url);
}

// Click handler for containers of externally-rendered HTML (e.g. markdown
// changelogs), where links can't be given their own onClick prop.
export function handleExternalLinkClick(e: MouseEvent<HTMLElement>) {
  const anchor = (e.target as HTMLElement).closest("a");
  if (anchor?.href) {
    e.preventDefault();
    openExternal(anchor.href);
  }
}
