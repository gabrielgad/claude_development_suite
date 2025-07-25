# CM (Claude Manager)

A terminal user interface (TUI) for monitoring and managing multiple Claude Code sessions across git worktrees.

## Features

- **Real-time Session Monitoring**: Automatically discovers and tracks Claude Code processes
- **Git Integration**: Shows git branch and worktree information for each session
- **Interactive UI**: Navigate sessions with keyboard shortcuts
- **Session Management**: View details, kill sessions, create new ones
- **Performance Optimized**: Built in Go for efficiency and low resource usage

## Installation

### Prerequisites

- Go 1.21 or later
- Git command-line tools
- Claude Code installed and accessible

### Build from Source

```bash
# Clone or navigate to the cm directory
cd cm

# Install dependencies
make deps

# Build the application
make build

# Install to your local bin (recommended)
make install

# Or install system-wide (requires sudo)
make install-system
```

### Verify Installation

```bash
cm --help
```

## Usage

### Basic Commands

```bash
# Start the Claude Manager TUI
cm

# The interface will show all running Claude Code sessions
```

### Keyboard Shortcuts

In the TUI interface:

- **↑/↓ or j/k**: Navigate between sessions
- **Enter**: Toggle between sessions list and details view
- **r**: Refresh session list
- **n**: Create new session (launches `cw make`)
- **k**: Kill selected session
- **q or Ctrl+C**: Quit

### Interface Overview

```
┌─ Claude Manager - Sessions (3 active) ────────────────────────────┐
│                                                                   │
│ ● auth-feature        [feature/auth]     Active     /path/to/auth │
│ ● api-refactor        [api-v2]          Active     /path/to/api   │
│ ● main                [main]            Active     /path/to/main  │
│                                                                   │
│ Controls: ↑/↓ navigate, Enter details, r refresh, n new, q quit  │
└───────────────────────────────────────────────────────────────────┘
```

## Development

### Project Structure

```
cm/
├── main.go           # Main application code
├── go.mod           # Go module definition
├── go.sum           # Go module checksums
├── Makefile         # Build automation
└── README.md        # This file
```

### Building

```bash
# Development build (with debug info)
make build-dev

# Production build (optimized)
make build

# Run directly
make run

# Format code
make fmt

# Run tests
make test
```

### Cross-Platform Builds

```bash
# Build for all platforms
make build-all

# Build for specific platform
make build-linux
make build-macos
make build-windows
```

## Configuration

Claude Manager uses minimal configuration. Future versions will support:

- Custom refresh intervals
- Theme customization
- Session filtering options

Configuration will be stored in `~/.config/claude_manager/config.json`.

## Integration with CW

Claude Manager works seamlessly with the `cw` (Claude Worktree) tool:

1. Use `cw make` to create new worktrees with Claude sessions
2. Use `cm` to monitor all active sessions
3. Use `cm` to create new sessions (internally calls `cw make`)

## Troubleshooting

### Common Issues

**"No Claude sessions found"**
- Ensure Claude Code is running
- Check that processes contain "claude" in their name
- Verify you have permission to inspect processes

**"Failed to create new session"**
- Ensure `cw` command is available in your Fish shell
- Check that you're in a git repository when creating sessions
- Verify git worktree functionality

**Performance Issues**
- Claude Manager is designed to handle 50+ sessions efficiently
- If experiencing slowdown, try reducing the number of active sessions
- Check system resources (CPU, memory)

### Debug Mode

For debugging, you can run with verbose output:

```bash
# Build development version
make build-dev

# Run with debug info (logs to stderr)
./build/cm 2> debug.log
```

## Contributing

See the main project documentation for contribution guidelines.

## License

[To be determined - see main project]

## Related Tools

- **cw (Claude Worktree)**: Fish shell function for creating git worktrees with Claude sessions
- **Claude Code**: The AI-powered development environment this tool manages

## Roadmap

### Current Version (v0.1.0)
- ✅ Basic process monitoring
- ✅ Session list interface
- ✅ Session details view
- ✅ Basic session management

### Planned Features
- [ ] Session logs view
- [ ] Command sending to Claude sessions
- [ ] Session approval workflow
- [ ] Configuration management
- [ ] Session persistence
- [ ] Performance metrics
- [ ] Plugin system