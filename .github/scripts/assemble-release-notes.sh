#!/usr/bin/env bash
set -euo pipefail

: "${GITHUB_REF_NAME:?}"
: "${GITHUB_REPOSITORY:?}"

notes_file=".github/release-notes/${GITHUB_REF_NAME}.md"
if [[ ! -f "$notes_file" ]]; then
  echo "::error::$notes_file is missing. Copy .github/release-notes/TEMPLATE.md to it and fill in what changed before tagging."
  exit 1
fi

installer=$(basename dist/*-setup.exe)
installer_url="https://github.com/$GITHUB_REPOSITORY/releases/download/$GITHUB_REF_NAME/$installer"
# Display name only, no version in it; installer_url above still points at
# the real versioned asset.
installer_display="league-rpc-setup.exe"

previous_tag=$(gh release list --repo "$GITHUB_REPOSITORY" --limit 1 --json tagName -q '.[0].tagName' 2>/dev/null || true)
if [[ -n "$previous_tag" ]]; then
  changelog_line="**Full Changelog**: https://github.com/$GITHUB_REPOSITORY/compare/$previous_tag...$GITHUB_REF_NAME"
else
  changelog_line="**Full Changelog**: https://github.com/$GITHUB_REPOSITORY/commits/$GITHUB_REF_NAME"
fi

cat >header.md <<HEADER
## Welcome to Release $GITHUB_REF_NAME

Got questions? Join Discord: https://discord.haze.sh

### Download

Download and run **[$installer_display]($installer_url)** to install League RPC.

Already have it installed? You don't need this. League RPC checks for updates on its own, and you can install them right from the About screen.

---

<!-- changelog:start -->
HEADER

cat >footer.md <<FOOTER
<!-- changelog:end -->

---

<details>
<summary>🛠️ Installation</summary>

1. Download **$installer_display** from the Assets section below, or the link above.
2. Run it and go through the setup like any other installer.
3. League RPC starts automatically and lives in your system tray from then on. Click the tray icon any time to open the window.

</details>

<details>
<summary>🛡️ Safety notice</summary>

League RPC is open-source and not yet signed by a certificate authority, so
Windows SmartScreen or your antivirus may flag it as unrecognized. That's
expected for a small open-source tool, not a sign anything is wrong.

- If SmartScreen blocks it: click "More info" then "Run anyway".
- If your antivirus blocks it: add an exclusion, or whitelist it.
- Prefer to check first? The source is public: https://github.com/$GITHUB_REPOSITORY

</details>

---

$changelog_line
FOOTER

cat header.md "$notes_file" footer.md >assembled-notes.md
rm -f header.md footer.md
