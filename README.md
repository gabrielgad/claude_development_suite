# Claude Development Suite

A modern web-based development environment for managing multiple Claude Code sessions with automated git worktree management and real-time terminal interfaces.

## 🚀 Current Status: Phase 2 Complete

**Version:** 2.0.0-web (Web Terminal Edition)  
**Architecture:** Full web-based terminal with file browser and git worktree integration  
**Status:** ✅ Production ready with complete parallel development workflow
**Binary:** `claude-manager`, `cm`, `claude-web`, `cmgr` (multiple aliases)

## 🏗️ Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                      Web Browser                                │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │ Session Manager │  │  Terminal UI    │  │ Permission Gate │  │
│  │ • Create/Kill   │  │ • xterm.js      │  │ • Approve/Deny  │  │
│  │ • Session List  │  │ • Live I/O      │  │ • File Changes  │  │
│  │ • Git Status    │  │ • Colors/Cursor │  │ • User Prompts  │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
└─────────────────────────┬───────────────────────────────────────┘
                          │ WebSocket (bidirectional)
┌─────────────────────────▼───────────────────────────────────────┐
│                    Go Web Server                                │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │   PTY Manager   │  │ Session Store   │  │ WebSocket Hub   │  │
│  │ • Create PTY    │  │ • Session State │  │ • Client Mgmt   │  │
│  │ • Process Spawn │  │ • Git Worktrees │  │ • Broadcast     │  │
│  │ • I/O Bridge    │  │ • Metadata      │  │ • Real-time     │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
└─────────────────────────┬───────────────────────────────────────┘
                          │ Standard PTY interface
┌─────────────────────────▼───────────────────────────────────────┐
│                   Claude Processes                              │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │ Session 1       │  │ Session 2       │  │ Session 3       │  │
│  │ ├─ Claude PTY   │  │ ├─ Claude PTY   │  │ ├─ Claude PTY   │  │
│  │ ├─ Git Worktree │  │ ├─ Git Worktree │  │ ├─ Git Worktree │  │
│  │ └─ Working Dir  │  │ └─ Working Dir  │  │ └─ Working Dir  │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## 📦 Prerequisites

- **Go 1.21+** - For building the web server
- **Modern Browser** - Chrome, Firefox, Safari, Edge
- **Git** - For worktree management  
- **Claude Code CLI** - The claude command-line tool

## ⚡ Quick Start

### 1. Build the Web Server
```bash
cd claude_development_suite/cm
make build
```

### 2. Run Tests (Optional)
```bash
# Run comprehensive test suite
./test_web_features.sh
# Expected: "34 passed, 0 failed"

# Or just unit tests  
go test -v ./...
```

### 3. Start the Server
```bash
./build/cm
# Server starts on http://localhost:8080
```

### 4. Verify Installation
```bash
./build/cm --version
# Output: Claude Manager v2.0.0-web (Web Terminal Edition)

./build/cm --help
# Shows all available options
```

## 🎯 What's Working Now

### ✅ Complete Implementation (Phase 2)

- **🌐 Full HTTP Server** - REST API + WebSocket handlers + static file serving  
- **💻 Professional File Browser** - Navigate any directory, git repo detection
- **🌳 Automated Git Worktrees** - Create `repo-name-sessionname` with feature branches
- **🖥️ Real-Time Terminal** - xterm.js with bidirectional WebSocket I/O
- **🎛️ Session Management** - Create, monitor, kill Claude sessions through web UI
- **🧹 Input Sanitization** - Git-compatible name cleaning with real-time feedback
- **⚙️ Professional Tooling** - Clean build system, multiple binary aliases
- **📚 Complete Documentation** - Architecture guides, testing procedures, usage docs

### 🎯 Ready for Production Use

- **Parallel Development** - Multiple isolated Claude sessions working simultaneously
- **Universal File Access** - Browse and select repositories anywhere on the system  
- **Git Integration** - Automatic worktree/branch creation for feature development
- **Web-Based UI** - Professional browser interface, no terminal dependencies
- **Cross-Platform** - Works on any system with Go and a modern browser

## 🚀 Usage Commands

### Basic Commands
```bash
# Start web server (default port 8080)
./build/cm

# Start on custom port
./build/cm -port 9000

# Show version
./build/cm --version

# Show help
./build/cm --help
```

### Development Commands
```bash
# Clean build
make clean && make build

# Run all tests
./test_web_features.sh

# Run unit tests only
go test -v ./...

# Test CW integration
cd ../cw && ./cw help
```

## 📁 Project Structure

```
claude_development_suite/
├── README.md                 # This file
├── TESTING.md               # Comprehensive testing guide
├── cm/                      # Claude Manager (Web Server)
│   ├── main.go             # Web server application
│   ├── main_test.go        # Unit tests
│   ├── integration_test.go # Integration tests
│   ├── test_web_features.sh # Automated test suite
│   ├── go.mod              # Go dependencies
│   ├── Makefile            # Build system
│   └── build/              # Compiled binaries
├── cw/                     # Claude Worktree Manager
│   ├── cw                  # Shell-agnostic script
│   └── cw.fish            # Fish shell version
└── docs/                   # Architecture documentation
    └── web-terminal-architecture.md
```

## 🧪 Testing

### Run All Tests
```bash
cd cm
./test_web_features.sh
```

### Expected Output
```
=== Claude Manager Web Terminal Test Suite ===
[... all tests ...]
========================================
Test Results: 34 passed, 0 failed
========================================
All tests passed! Web terminal architecture is ready for Phase 2.
```

See [TESTING.md](./TESTING.md) for detailed testing instructions.

## 🔧 Technical Details

### Dependencies
- **WebSocket:** `github.com/gorilla/websocket v1.5.1`
- **PTY:** `github.com/creack/pty v1.1.21`  
- **Process Monitoring:** `github.com/shirou/gopsutil/v3 v3.23.12`

### Architecture Benefits
- ✅ **No Terminal Conflicts** - Web server owns PTY, no ownership issues
- ✅ **Rich UI Capabilities** - Browser-based with full session control
- ✅ **True Terminal Experience** - Claude gets real PTY with colors, cursor control
- ✅ **Multi-Session Support** - Handle multiple Claude instances simultaneously  
- ✅ **Cross-Platform** - Works everywhere with just browser + Go binary
- ✅ **No External Dependencies** - No tmux/zellij requirements

### Performance
- **Binary Size:** ~3MB optimized
- **Startup Time:** <500ms
- **Memory Usage:** <50MB baseline
- **Test Suite:** 34 tests in <3 seconds

## 🚦 Development Status

| Component | Status | Implementation |
|-----------|--------|---------------|
| HTTP Server | ✅ Complete | REST API + Static Files |
| WebSocket I/O | ✅ Complete | Real-time Terminal |
| PTY Management | ✅ Complete | Claude Process Control |
| Web Frontend | ✅ Complete | xterm.js + File Browser |
| File Browser | ✅ Complete | Universal Directory Navigation |
| Git Worktrees | ✅ Complete | Automated Parallel Development |
| Input Sanitization | ✅ Complete | Git-compatible Names |
| Session Management | ✅ Complete | Create/Monitor/Kill |
| Build System | ✅ Complete | Professional Tooling |
| Documentation | ✅ Complete | Complete Guides |

## 🎯 Usage Example

```bash
# 1. Start the server
claude-manager
# Server: http://localhost:8080

# 2. Open browser and create session
# - Navigate to any git repository using file browser  
# - Enter session name: "auth-feature"
# - Creates: repo-name-auth-feature/ directory
# - Creates: feature/auth-feature branch
# - Spawns: Claude in isolated worktree

# 3. Parallel development
# - Create multiple sessions simultaneously
# - Each gets its own worktree and branch
# - Work on different features in parallel
# - No conflicts between sessions
```

## 🤝 Contributing

### Development Workflow
1. Make changes to code
2. Run tests: `./test_web_features.sh`
3. Verify build: `make build`
4. Test functionality: `./build/cm --version`
5. If all tests pass: commit and push

### Testing Standards
- All new features must have tests
- Test suite must show "X passed, 0 failed"
- Unit tests must pass: `go test ./...`
- Integration tests must pass

## 📄 License

MIT License - see LICENSE file for details.

## 📞 Support

- **Documentation:** [docs/web-terminal-architecture.md](./docs/web-terminal-architecture.md)
- **Testing Guide:** [TESTING.md](./TESTING.md)  
- **Issues:** Please report via GitHub issues

---

**Claude Development Suite v2.0.0-web**  
*Modern web-based development environment for Claude Code sessions*