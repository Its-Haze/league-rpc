# In-app updates: signed binary swap from GitHub Releases

**Status**: accepted

The app ships as a portable single `.exe` and updates itself through the Wails v3 updater service: on launch and every few hours it checks GitHub Releases, and when the user clicks Update it downloads the new binary, verifies an ed25519 signature against a `SHA256SUMS` sidecar, swaps the executable via a helper process, and restarts. Users never visit GitHub or run an installer.

**Considered options**: an NSIS/MSI installer with silent-reinstall updates (better first-run integration, but in-place update means re-running the installer unattended and fighting UAC); MSIX/winget (clean OS-managed updates, but needs a Store-style package and an OS code-signing certificate up front). Rejected for v1 because the portable-exe plus binary-swap path is the least infrastructure for exactly the "update from inside the app" requirement.

**Consequences**: the release artifact layout is now load-bearing: every release must publish the `.exe` plus a `SHA256SUMS` sidecar signed with the project's ed25519 key, produced by the tagged CI build. The private key lives in a GitHub Actions secret behind a protected environment; the public key is compiled into the app. Key rotation requires shipping a build that trusts both keys before retiring the old one. This is separate from OS code signing: without a code-signing certificate, Windows SmartScreen still warns on first run and after each update. Downloads are click-initiated, never silent, so a presence app does not spend a user's bandwidth on updates they did not ask for.
