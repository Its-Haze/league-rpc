---
name: release
description: Cut a League RPC release end to end (version number, release notes, commit, push, tag). Use whenever the user asks to cut, ship, or do a release, wants to tag a new version, or asks what the next release should look like after finishing a feature or fix. Also trigger after wrapping up a chunk of work when the user says something like "let's release this" or "ready to ship". Do not trigger for routine commits that aren't meant to become a tagged release.
---

# League RPC release

Cuts a release from whatever is currently committed and uncommitted in the
working tree, using the session's own knowledge of what was just built.
There is one maintainer and no pull request workflow: commits go straight to
master.

Stop and wait for explicit approval at the point marked below. Nothing gets
committed, pushed, or tagged before the user has seen and approved the
version number and the release notes together.

## Process

1. **Work out what's actually in this release.** You almost certainly
   already know, from the conversation, what was built or fixed. Confirm it
   against the repo rather than relying on memory alone:
   `git tag --sort=-v:refname | head -1` for the last release, then `git
   status` for anything still uncommitted and `git log <last tag>..HEAD
   --stat` plus the relevant diffs for anything already committed since. If
   there is uncommitted work that belongs in this release, it still needs to
   be committed before the release notes step, since the notes describe
   what shipped, not what merely exists in the tree.

2. **Propose a version number.** Tags in this repo are `vX.Y.Z`
   (`internal/version/version.go` reports whatever gets injected at build
   time). Infer the bump from what actually changed, the same way you'd
   read a diff to write a commit message, not by defaulting to a patch bump
   out of habit:
   - Patch: bug fixes, wording, packaging or pipeline fixes, anything that
     doesn't add or change user-facing behavior.
   - Minor: a new feature or a visible improvement that doesn't break
     anything already there.
   - Major: something that breaks an existing workflow or removes a
     capability. Rare for a single-maintainer desktop app; don't reach for
     it without a real reason.

3. **Draft the release notes.** Follow the `rpc-release-notes` skill's process
   (`.claude/skills/rpc-release-notes/SKILL.md`) to write
   `.github/release-notes/vX.Y.Z.md`: read the actual diffs, filter to what
   a user of the app would notice, group into Highlights and Bug Fixes,
   skip emoji by default. Don't duplicate that skill's instructions here;
   read and follow them.

4. **Present and stop.** Show the user the proposed version number and the
   full drafted release notes together, and wait. This is the one point in
   the process where you need their word before continuing. If they ask for
   changes, revise and present again.

5. **Commit, once approved.** Stage what belongs in the release and commit
   directly to master. The commit message is a single line: what changed,
   stated plainly, no body paragraph, no bullet list. This is the same rule
   the user's global git settings already apply everywhere, so nothing
   special is needed here beyond following it.

6. **Push master.**

7. **Ask before tagging.** Once master is pushed, ask the user whether to
   tag and release now. If yes, create an annotated tag (`git tag -a vX.Y.Z
   -m vX.Y.Z`) and push it. Pushing the tag triggers
   `.github/workflows/release.yml`, which builds, signs, and publishes the
   GitHub release on its own.

8. **Do not watch the workflow run.** The user monitors the GitHub Actions
   run themselves. Don't poll `gh run view`, don't wait for it to finish,
   don't report back on its status unless asked. Once the tag is pushed,
   this skill's job is done.
