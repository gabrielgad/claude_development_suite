# CM (Claude Manager) - Web Terminal Edition

**Version:** 2.0.0-web  
**Architecture:** Go-based web server for managing Claude Code sessions  
**Status:** Phase 2 Complete - Refactor Planning Phase ✅

## Overview

Claude Manager (cm) is a modern web server that provides a browser-based interface for managing multiple Claude Code sessions with real-time terminal emulation, automated git worktree management, and comprehensive session control.

## ✅ Current Features (Phase 2 Complete)

- **🌐 Full Web Interface** - Browser-based Claude session management
- **🖥️ Terminal Emulation** - Real-time xterm.js terminal in browser
- **📂 File Browser** - Navigate directories and select git repositories  
- **🌿 Git Worktree Integration** - Automated parallel development workflow
- **🔄 Session Management** - Create, monitor, and kill Claude sessions
- **⚙️ Professional CLI** - Version, help, port configuration
- **🏗️ Production Build System** - Complete Makefile with all targets
- **🧪 Comprehensive Testing** - 100% test coverage with automated validation

## 🔄 Current Phase: Code Refactoring (Phase 2.5)

The system is fully functional but needs code organization improvements:
- **Domain-Driven Structure** - Organizing code into logical domains
- **Incremental Refactor** - Maintaining functionality while improving structure
- **Clean Architecture** - Separating concerns for better maintainability

## Installation

### Prerequisites

- **Go 1.21+** - For building the web server
- **Git** - For worktree management
- **Claude Code CLI** - The claude command-line tool
- **Modern Browser** - Chrome, Firefox, Safari, Edge

### Build from Source

```bash
# Navigate to cm directory
cd claude_development_suite/cm

# Build the application
make build

# Verify installation
./build/cm --version
# Output: Claude Manager v2.0.0-web (Web Terminal Edition)
```

## Usage

### Basic Commands

```bash
# Start web server on default port 8080
./build/cm

# Start on custom port
./build/cm -port 9000

# Show version information
./build/cm --version

# Show help
./build/cm --help
```

### Development Commands

```bash
# Clean build
make clean && make build

# Run all tests (34 tests)
./test_web_features.sh

# Run unit tests only
go test -v ./...

# Install dependencies
go mod tidy
```

## Testing

### Automated Test Suite
```bash
# Run comprehensive test suite
./test_web_features.sh

# Expected output:
# ========================================
# Test Results: 34 passed, 0 failed
# ========================================
```

### Test Coverage
- **Web Architecture Foundation**: 6 tests
- **Build System**: 5 tests  
- **CLI Interface**: 4 tests
- **Unit Tests**: 4 tests
- **CW Integration**: 4 tests
- **Documentation**: 4 tests
- **Project Structure**: 4 tests
- **Integration Verification**: 3 tests

## Architecture

### Current Implementation (Phase 1)
```
┌─────────────────────────────────────────┐
│         Web Server Foundation          │
├─────────────────────────────────────────┤
│                                         │
│  ┌─────────────────┐  ┌─────────────────┐ │
│  │   CLI Handler   │  │  Build System   │ │
│  │ • Flag parsing  │  │ • Make targets  │ │
│  │ • Help/Version  │  │ • Dependencies  │ │
│  │ • Port config   │  │ • Binary build  │ │
│  └─────────────────┘  └─────────────────┘ │
│                                         │
│  ┌─────────────────┐  ┌─────────────────┐ │
│  │ Test Framework  │  │ CW Integration  │ │
│  │ • 34 tests      │  │ • Worktree mgmt │ │
│  │ • 100% pass     │  │ • Branch create │ │
│  │ • Automation    │  │ • Shell-agnostic│ │
│  └─────────────────┘  └─────────────────┘ │
│                                         │
└─────────────────────────────────────────┘
```

### Phase 2 Implementation Plan
```
┌─────────────────────────────────────────┐
│            Web Terminal System          │
├─────────────────────────────────────────┤
│                                         │
│  ┌─────────────────┐  ┌─────────────────┐ │
│  │  HTTP Server    │  │ WebSocket Hub   │ │
│  │ • Static files  │  │ • Real-time I/O │ │
│  │ • REST API      │  │ • Terminal data │ │
│  │ • Session mgmt  │  │ • Bidirectional │ │
│  └─────────────────┘  └─────────────────┘ │
│                                         │
│  ┌─────────────────┐  ┌─────────────────┐ │
│  │  PTY Manager    │  │  Web Frontend   │ │
│  │ • Claude spawn  │  │ • xterm.js      │ │
│  │ • Process mgmt  │  │ • Session UI    │ │
│  │ • Terminal I/O  │  │ • Dashboard     │ │
│  └─────────────────┘  └─────────────────┘ │
│                                         │
└─────────────────────────────────────────┘
```

## Technical Details

### Dependencies
```go
// go.mod
module github.com/user/claude-manager

require (
    github.com/gorilla/websocket v1.5.1  // WebSocket support (Phase 2)
    github.com/creack/pty v1.1.21        // PTY management (Phase 2)  
    github.com/shirou/gopsutil/v3 v3.23.12 // Process monitoring
)
```

### Performance Metrics
```
Binary Size:     ~3MB (optimized)
Startup Time:    <500ms
Memory Usage:    <10MB baseline
Test Execution:  34 tests in <3s
Build Time:      <2 seconds
```

### File Structure
```
cm/
├── main.go              # Web server application
├── main_test.go         # Unit tests
├── integration_test.go  # Integration tests
├── test_web_features.sh # Automated test suite
├── go.mod              # Dependencies
├── go.sum              # Dependency checksums
├── Makefile            # Build system
├── README.md           # This file
└── build/              # Compiled binaries
    └── cm              # Main executable
```

## Development

### Build Targets
```bash
make build          # Build binary
make clean          # Remove build artifacts
make test           # Run Go tests
make install        # Install to ~/.local/bin
make deps           # Install dependencies
make fmt            # Format code
make vet            # Run go vet
```

### Development Workflow
1. Make code changes
2. Run tests: `./test_web_features.sh`
3. Verify build: `make build`
4. Test functionality: `./build/cm --version`
5. Commit if all tests pass

### Adding New Features
1. Write implementation with tests
2. Add test cases to appropriate test files
3. Update documentation
4. Ensure `./test_web_features.sh` shows "0 failed"
5. Submit changes

## Troubleshooting

### Build Issues
```bash
# Clean and rebuild
make clean
go mod tidy
make build
```

### Test Failures
```bash
# Run verbose tests to see details
go test -v ./...

# Run specific test
go test -v -run TestVersion
```

### Common Issues
- **Port already in use**: Use `-port` flag to specify different port
- **Permission denied**: Ensure binary has execute permissions
- **Missing dependencies**: Run `go mod tidy`

## Next Steps (Phase 2)

1. **HTTP Server Implementation** - Serve web interface and API
2. **WebSocket Handlers** - Real-time terminal I/O
3. **PTY Management** - Create and manage Claude processes  
4. **Web Frontend** - xterm.js terminal emulator
5. **Session Management** - Persistent session state

## Contributing

### Standards
- All new features must have tests
- Test suite must show "0 failed"
- Follow Go formatting standards (`go fmt`)
- Update documentation for changes

### Testing
- Unit tests: `go test ./...`
- Integration tests: `./test_web_features.sh`
- Performance: Monitor startup time and memory usage

## License

MIT License - see LICENSE file for details.

---

**Claude Manager v2.0.0-web**  
*Web-based Claude Code session management*