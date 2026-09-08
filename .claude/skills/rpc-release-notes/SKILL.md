---
name: rpc-release-notes
description: Draft the curated changes file for a League RPC release (.github/release-notes/vX.Y.Z.md). Use when cutting a release, preparing to tag a version, or asked to write/draft release notes or a changelog for this project.
---

# League RPC release notes

Produces exactly one file: `.github/release-notes/vX.Y.Z.md`, the curated
list of what changed in a release. That file is the only input the release
workflow needs from you: `.github/scripts/assemble-release-notes.sh` wraps
it in the rest of the GitHub release (download link, install steps, Discord
link, safety notice) automatically, and the app's own update banner pulls
this exact file's content back out of the published release via the
`<!-- changelog:start/end -->` markers. Never write the surrounding
boilerplate yourself; it would end up duplicated in the GitHub page and
leaking into the in-app changelog view, where a user already has the app
installed and running and doesn't need to be told how to install it.

Read `.github/release-notes/TEMPLATE.md` for the exact section shape and the
emoji policy before writing anything.

## Process

1. **Find the range.** The previous tag (`git tag --sort=-v:refname | head -1`,
   or `git describe --tags --abbrev=0` from the commit being released) through
   the commit you're releasing. If there's no previous tag, this is the first
   release, so cover everything.
2. **Read the actual changes**, not just commit subject lines: `git log
   <prev>..HEAD --stat` to see what moved, then read the diffs that matter.
   A commit message can undersell or oversell what happened; the diff is the
   truth.
3. **Filter hard.** Only a user of the app would notice or care about this
   list. Cut internal refactors, test-only changes, CI/pipeline/release
   tooling changes, doc-only commits, and dependency bumps with no visible
   effect, unless one of those commits also happens to fix something a user
   experienced, in which case describe the user-visible fix, not the
   mechanism.
4. **Group into two sections**, each optional. Omit a section entirely
   rather than leaving it empty or writing "no bug fixes this release":
   - `### Highlights`: features and improvements, each its own `#### Title`,
     then one line on why someone would care, then a bullet list of the
     specifics. Don't write a paragraph per change, and don't pad the
     bullets. Three honest ones beat five with filler.
   - `### Bug Fixes`: a terse bullet list, one line per fix.
5. **Match the tone**: direct and specific, not promotional. State what
   changed and what it does for the person using the app. No filler
   ("we're excited to..."), no vague credit ("various improvements"), no
   restating the obvious. Follow the global writing-style rules (plain
   words, active voice, no em dashes, cut hedging and puffery).
6. **Emoji: one per `####` heading, and nowhere else.** Never on a bullet.
   A row of emoji-headed bullets reads as generated, not written. Pick one
   that fits the feature; skip it rather than reach for something generic.
7. **Write the file**, show the user the draft, and get their sign-off
   before anything gets tagged. This is public-facing copy; a human read
   before it ships matters even when a model drafted it.
