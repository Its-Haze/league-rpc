# Project Setup Summary

This document summarizes the initial project setup completed on 2025-10-19.

## What Was Done

### 1. Project Initialization
- Created new `league-rpc` directory separate from reference projects
- Initialized Go module: `github.com/its-haze/league-rpc`
- Set up clean directory structure following Go best practices

### 2. Dependencies Installed

All core dependencies researched and installed (see `go.mod` for pinned versions):

```
github.com/its-haze/lcu-gopher    # LCU API integration
github.com/hugolgst/rich-go       # Discord RPC
github.com/shirou/gopsutil/v4     # Process detection
github.com/rs/zerolog             # High-performance logging
```

**Why these dependencies?**
- **lcu-gopher**: Custom library specifically for League Client API
- **rich-go**: The most widely adopted Go Discord RPC library
- **gopsutil**: Cross-platform process detection, actively maintained
- **zerolog**: Fastest Go logging library with zero allocations

See [DEPENDENCIES.md](DEPENDENCIES.md) for detailed rationale.

### 3. Project Structure Created

```
league-rpc/
├── cmd/
│   └── league-rpc/          # Application entry point (to be created)
├── internal/
│   ├── app/                 # Wails app bindings (to be created)
│   ├── lcu/                 # LCU client wrapper (to be created)
│   ├── discord/             # Discord RPC integration (to be created)
│   ├── config/              # Configuration system (done)
│   ├── state/               # Application state (to be created)
│   └── process/             # Process detection (to be created)
├── pkg/
│   ├── types/               # Shared types (to be created)
│   └── constants/           # Constants (done)
├── assets/                  # Icons and images (to be added)
├── go.mod                   # Module definition
├── README.md                # Project documentation
├── DEPENDENCIES.md          # Dependency rationale
└── PROJECT_SETUP.md         # This file
```

### 4. Configuration System

Created a complete configuration system:

**Files:**
- `internal/config/config.go`: Config struct with defaults
- `internal/config/loader.go`: Load/save JSON config from disk

**Features:**
- JSON-based configuration (NO CLI flags!)
- Stored in `%APPDATA%\league-rpc\config.json`
- All settings accessible through GUI
- Validation and default values
- Discord App ID presets (League of Legends, League of Kittens, League of Linux)

**Configuration Options:**
```go
type Config struct {
    DiscordAppID     string // Discord Application ID
    ShowStats        bool   // Show KDA/CS
    ShowRank         bool   // Show rank/LP
    ShowEmojis       bool   // Show online/away emoji
    ShowInClient     bool   // Show RPC when idle
    AutoLaunchLeague bool   // Auto-launch League
    LeaguePath       string // Custom League path
    UpdateInterval   int    // RPC throttle (ms)
    DebugMode        bool   // Debug logging
}
```

### 5. Constants Package

Created `pkg/constants/constants.go` with:
- Discord App IDs
- LCU API endpoints
- In-game API URLs
- Asset URLs (Community Dragon, Data Dragon)
- Process names
- Game flow phases
- Queue type IDs

### 6. Documentation

Created comprehensive documentation:
- **README.md**: Project overview, tech stack, structure, development status
- **DEPENDENCIES.md**: Detailed dependency choices and rationale
- **CLAUDE.md**: Updated with correct project structure and guidelines
- **PROJECT_SETUP.md**: This summary document

## Architecture Decisions

### No CLI Flags
All configuration is done through a GUI instead of command-line flags. This makes the application more user-friendly and accessible.

### Windows Only
Targeting Windows exclusively due to Riot Vanguard breaking Linux support. This simplifies development and testing.

### Clean Code Focus
Emphasis on:
- Clear separation of concerns
- Single responsibility principle
- Descriptive naming
- Proper error handling
- Well-documented APIs
- Structured logging

### Wails-Ready Structure
Project structure designed from the start to integrate seamlessly with Wails v3:
- `internal/app/` will contain Wails bindings
- `frontend/` will be created by `wails3 init`
- Backend logic separate from UI concerns

## Next Steps

### Immediate (Backend Core)
1. **LCU Client Wrapper** (`internal/lcu/`)
   - Wrap lcu-gopher with app-specific logic
   - Event handlers for gameflow, summoner, lobby, ranked
   - In-game stats polling (127.0.0.1:2999)

2. **Discord RPC Integration** (`internal/discord/`)
   - Wrapper around rich-go
   - Presence state management (in-client, lobby, in-game, etc.)
   - Update throttling (1.5s delay)

3. **State Management** (`internal/state/`)
   - Thread-safe application state
   - Channel-based event propagation
   - State synchronization between LCU and Discord

4. **Process Detection** (`internal/process/`)
   - Detect Discord running
   - Detect League Client running
   - Auto-launch League functionality

### Medium Term (GUI)
5. **Wails Setup**
   - Run `wails3 init` in frontend directory
   - Set up React + TypeScript
   - Install shadcn/ui components

6. **Frontend Development**
   - Settings panel
   - Status display
   - System tray integration
   - Wails bindings for backend functions

### Long Term (Polish)
7. **Testing**
   - Unit tests for business logic
   - Integration tests
   - Manual testing with League Client

8. **Distribution**
   - Build pipeline
   - Update checker
   - Installer creation
   - GitHub releases

## Development Commands

```bash
# Navigate to project
cd league-rpc

# Install dependencies
go mod download

# Run tests (when created)
go test ./...

# Format code
go fmt ./...

# Lint (requires golangci-lint)
golangci-lint run

# Build (standalone, before Wails)
go build -o league-rpc.exe ./cmd/league-rpc

# Wails development (after Wails setup)
wails3 dev

# Wails production build
wails3 build
```

## Reference projects

`../lcu-gopher/` is a separate, actively-used Go library for LCU API access. **DO NOT MODIFY** it unless we need to add functionality to the library itself.

## Design Principles

1. **User-Friendly**: GUI over CLI, clear settings, intuitive interface
2. **Performance**: Fast startup, minimal resource usage, efficient updates
3. **Maintainable**: Clean code, good documentation, proper testing
4. **Extensible**: Easy to add new features, modular design
5. **Robust**: Proper error handling, graceful degradation, logging

## Success Criteria

The project will be considered successful when:
- [ ] Full Discord Rich Presence coverage for every phase of a League session is implemented
- [ ] GUI allows easy configuration of all settings
- [ ] Startup is fast and memory usage stays low
- [ ] Code is clean, well-documented, and tested
- [ ] Single executable runs on Windows without dependencies
- [ ] System tray integration works smoothly
- [ ] Auto-update checking is implemented

## Contact & Resources

- **CLAUDE.md**: Comprehensive guide for future Claude Code instances
- **README.md**: User-facing documentation
- **DEPENDENCIES.md**: Dependency rationale and licenses

---

**Setup completed**: 2025-10-19
**Next session**: Start implementing LCU client wrapper and Discord RPC integration
