# Dependencies

This document explains the dependencies chosen for this project and why they were selected.

## Core Dependencies

### 1. lcu-gopher
**Repository**: `github.com/its-haze/lcu-gopher`

**Purpose**: League Client Update (LCU) API integration

**Why chosen**:
- Custom-built library specifically for League Client API interaction
- Handles both HTTP requests and WebSocket events
- Built-in authentication and credential detection
- Cross-platform support (Windows, macOS, Linux)
- Uses WAMP protocol for WebSocket communication
- Thread-safe with proper mutex handling
- Already includes all necessary LCU types

**Key Features**:
- Automatic credential extraction from lockfile or process
- Event subscription system for real-time updates
- Helper methods for common operations (GetCurrentSummoner, GetChampSelectSession, etc.)
- Configurable timeouts and polling intervals

---

### 2. Discord IPC client (internal)
**Location**: `internal/discord/ipc`

**Purpose**: Discord Rich Presence transport

**Why built in-house rather than using a third-party library**:
- The Discord RPC libraries available for Go all swallow named-pipe write errors internally (print-and-discard) instead of returning them, so nothing above them can ever detect a broken connection that isn't also a process crash
- One candidate is GPL-3.0 licensed, which would impose copyleft obligations on any distributed build of this app
- The remaining candidates are small, largely unmaintained forks of each other with the same underlying defect
- The actual wire protocol (dial the named pipe, send/receive length-prefixed JSON frames) is small enough that owning it outright costs less than working around a third-party library's error-swallowing

**Key Features**:
- Each connection is instance-scoped, not a package-level singleton, so reconnecting needs no separate state-reset step
- Send failures return real errors immediately, so the app can detect a dropped connection (e.g. Discord's in-app reload) without waiting on a process-liveness check

---

### 3. gopsutil/v4
**Repository**: `github.com/shirou/gopsutil/v4`

**Purpose**: Process detection and system monitoring

**Why chosen**:
- Cross-platform process management
- Actively maintained
- No CGO dependencies (pure Go)
- Comprehensive process information (name, PID, command line, etc.)

**Use cases**:
- Detect if Discord is running
- Find Discord process name
- Detect League Client process
- Monitor process status

**Key Features**:
- Process listing and filtering
- Process name and command-line argument access
- Memory and CPU usage (if needed for debugging)
- Platform-specific information through Ex structs

---

### 4. zerolog
**Repository**: `github.com/rs/zerolog`

**Purpose**: High-performance structured logging

**Why chosen**:
- One of the fastest Go logging libraries available
- Zero-allocation design for minimal overhead
- JSON-based structured logging
- Beautiful console output with color support
- Simple, intuitive API
- No external dependencies (besides stdlib)

**Alternatives considered**:
- `uber-go/zap`: Similar performance, but more complex API
- `logrus`: Slower performance, in maintenance mode
- `log/slog`: Good stdlib option, but zerolog is faster

**Key Features**:
- Leveled logging (debug, info, warn, error, fatal)
- Structured fields (key-value pairs)
- Context-based logging
- Multiple output targets
- Caller information and stack traces
- Conditional logging with sampling

---

### 5. lumberjack
**Repository**: `gopkg.in/natefinch/lumberjack.v2`

**Purpose**: Log file rotation

**Why chosen**:
- The GUI process has no attached console, so `zerolog` needs a file sink; lumberjack adds size-capped rotation with no extra moving parts
- Small, single-purpose, no dependencies of its own

**Key Features**:
- Size-capped rotation with a configurable number of retained backups
- Drop-in `io.Writer`, so it composes with `zerolog`'s multi-writer alongside the in-memory log ring the Help screen reads from

---

## Frontend Dependencies

### Wails v3
**Repository**: `github.com/wailsapp/wails/v3`

**Purpose**: Desktop application framework hosting the GUI and the daemon in one process

**Why chosen**:
- Native Go + Web frontend combination, so the existing daemon and `internal/app` binding seam need no rewrite
- System tray, native window menus/dialogs, and a built-in signed-update service, all needed by this project
- Auto-generated TypeScript bindings for every `internal/app` method
- No Electron overhead

---

### React + TypeScript
**Repository**: `https://react.dev/`

**Purpose**: Frontend UI framework

**Why chosen**:
- TypeScript-first, and the framework Wails' binding generation and examples are built around

---

### Tailwind CSS + Radix UI primitives (not shadcn/ui)
**Repositories**: `https://tailwindcss.com/`, `@radix-ui/react-*`

**Purpose**: Styling and accessible unstyled component behavior

**Why chosen over shadcn/ui**:
- shadcn/ui is copy-paste boilerplate on top of the same Radix primitives; for an app this size (a handful of screens), building the small component set directly on Radix costs little extra and avoids inheriting shadcn's now-widely-recognized default look
- Radix supplies the accessible behavior (dialog, select, switch, tabs, label) that would otherwise need to be hand-rolled

---

### Supporting frontend libraries
- **lucide-react**: icon set used throughout the shell and screens
- **dompurify**: sanitizes the changelog markdown rendered in About before it hits the DOM
- **@fontsource-variable/inter**: bundles the Inter variable font locally, no external font request
- **Vitest + React Testing Library**: unit tests for template-preview rendering, config-form state, and patch-builder logic

---

## Dependency Philosophy

Our dependency selection follows these principles:

1. **Performance First**: Choose libraries with minimal overhead and optimal performance
2. **Active Maintenance**: Prefer libraries with recent updates and active communities
3. **Type Safety**: Prioritize libraries with good type definitions and Go idioms
4. **Minimal Dependencies**: Avoid heavy dependency trees
5. **Cross-Platform**: Support Windows primarily, with potential for future expansion
6. **Production-Ready**: Use battle-tested libraries with proven track records

---

## Dependency Update Policy

- **Regular Reviews**: Check for updates monthly
- **Security Patches**: Apply immediately
- **Breaking Changes**: Evaluate carefully, update when beneficial
- **Version Pinning**: Use go.mod for reproducible builds

---

## Dependency count

See `go.mod` for the current, authoritative list of direct and indirect dependencies.

---

## License Compatibility

All dependencies use permissive licenses (MIT, BSD, Apache-2.0, or ISC), compatible with this project's MIT license. See each dependency's repository for its exact license.
