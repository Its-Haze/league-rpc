# Release signing

Every tagged release publishes four assets: the NSIS installer (`*-setup.exe`,
what a person downloads and runs), the raw binary (`internal/updates.ReleaseAsset`,
named `.bin` on purpose so it doesn't look double-clickable next to the
installer), `SHA256SUMS` (its sha256 digest), and `SHA256SUMS.sig` (an ed25519
signature over that digest). `internal/updates` verifies both before
installing an update; see
[ADR-0005](adr/0005-in-app-update-via-signed-binary-swap.md). Only the
installer and the raw binary are meant for anyone to touch directly, and only
the installer is meant for a human — the release notes say so, and the raw
binary's `.bin` extension means Windows won't just run it if someone
double-clicks it anyway.

The workflow (`.github/workflows/release.yml`) runs on every `v*` tag push:
a `build` job produces the binary, a `sign` job (restricted to the
`release-signing` environment) hashes and signs it with `cmd/sign-release`,
and a `release` job publishes everything to GitHub Releases. A tag containing
a hyphen (`v1.2.3-rc1`) is published as a GitHub prerelease, which the app's
`Prerelease: false` config never picks up, so it is safe to use for a dry run.

## Generating the keypair

```
openssl genpkey -algorithm ed25519 -out update-private.pem
openssl pkey -in update-private.pem -pubout -out update-public.pem
```

Commit `update-public.pem` to `internal/updates/keys/update-public.pem`.

In the repo's Settings > Environments, create `release-signing` (add any
required reviewers you want), then add `update-private.pem`'s contents as an
**environment secret** named `UPDATE_SIGNING_KEY` on that environment, not a
plain repository secret — an environment secret is only exposed to a job
that declares `environment: release-signing`, which is what actually keeps
the key out of the `build` and `release` jobs. Delete the local private key
file once it's stored. Never commit it.

## Dry run

Push a tag with a hyphen, e.g. `v0.0.1-rc1`, and watch the workflow. It
produces a real GitHub prerelease with all three assets; verify by hand that
`sha256sum` matches `SHA256SUMS` and that the signature in `SHA256SUMS.sig`
verifies against `internal/updates/keys/update-public.pem`. Because the app
only checks `/releases/latest` (which excludes prereleases), no running
client will pick this release up.

## Rotating the key

1. Generate a new keypair as above.
2. Ship a build whose public-key check accepts both the old and new key
   (temporarily embed both PEMs and try each in turn) so users on the old
   key can still verify the release that switches them over.
3. Update the `UPDATE_SIGNING_KEY` secret to the new private key and start
   signing releases with it.
4. Once telemetry or a reasonable waiting period shows no clients still
   need the old key, ship a build that trusts only the new one.
