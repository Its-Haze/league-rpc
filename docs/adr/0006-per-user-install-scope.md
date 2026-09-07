# Windows installs go to the per-user scope

**Status**: accepted

The NSIS installer built by `task package` installs to `$LOCALAPPDATA\Programs\League RPC` and requests no elevation. Machine-wide installs into `$PROGRAMFILES64` remain possible with `task package INSTALL_SCOPE=machine`, but nothing in the release pipeline builds one.

This amends [0005](0005-in-app-update-via-signed-binary-swap.md), which rejected an installer for v1. Both artifacts now ship: the installer is how a person gets the app the first time, and the signed binary swap is still how it updates itself afterwards. The installer is deliberately not part of the update path, and the updater's asset matcher skips any filename containing `-installer.`.

**Why per-user**: the update flow overwrites the running executable in place. Under `$PROGRAMFILES64` that write needs elevation, so every update would raise a UAC prompt or fail outright, which defeats the point of updating from inside the app. The rest of the app is already per-user anyway: settings live under `%APPDATA%\league-rpc` and the launch-at-login entry is written to `HKCU\...\CurrentVersion\Run`. A machine-wide install would put one per-machine file among per-user state.

**Considered options**: machine scope with an elevated update helper (a second binary, a UAC prompt per update, and a signing story we do not have yet); machine scope with an ACL loosened on the install directory (a writable directory under `$PROGRAMFILES64` is a local privilege escalation, so no).

**Consequences**: each Windows account that wants League RPC installs its own copy, and an admin cannot deploy it once for every user on a shared machine. Add/Remove Programs lists it under the current user only. The installer refuses to run while the app is running, since overwriting a locked executable would leave a half-updated install and there is no way to ask a tray app to quit from outside. Uninstall deletes the `Run` value it may have left behind, but keeps settings and logs so a reinstall picks up where the user left off.
