# Claude Development Suite

A web-based development environment for managing multiple Claude Code sessions with git worktrees and real-time terminal interfaces.

## Overview

This suite provides:

1. **Web Dashboard** - Browser-based session management and monitoring
2. **Terminal Interface** - Full terminal emulator (xterm.js) for direct Claude interaction  
3. **Git Integration** - Automated worktree creation and branch management
4. **Permission System** - Centralized approval/denial of Claude requests

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                   Claude Development Suite                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────┐              ┌─────────────────────────────┐ │
│  │   cw (Fish)     │              │    cm (Go TUI)              │ │
│  │                 │              │                             │ │
│  │ • Create        │─────────────▶│ • Monitor Sessions         │ │
│  │   worktrees     │              │ • Real-time Status         │ │
│  │ • Launch Claude │              │ • Send Commands            │ │
│  │ • Manage deps   │              │ • Kill Sessions            │ │
│  └─────────────────┘              │ • Approve/Deny Changes     │ │
│                                   └─────────────────────────────┘ │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│                        Git Worktrees                            │
│                                                                 │
│  project/          project-auth/       project-api/            │
│  ├── .git/         ├── files          ├── files                │
│  ├── main files    ├── auth branch    ├── api branch           │
│  └── ...           └── Claude PID     └── Claude PID           │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## Prerequisites

- **Go 1.21+** - For building the web server
- **Modern Browser** - Chrome, Firefox, Safari, Edge
- **Git** - For worktree management  
- **Claude Code CLI** - The claude command-line tool

## Quick Start

### 1. Build the Web Server
```bash
cd cm
make build
```

### 2. Start the Server
```bash
./build/cm serve
# Server starts on http://localhost:8080
```

### 3. Open Dashboard
```bash
# Open browser to http://localhost:8080
# See session dashboard with create/manage options
```

### 4. Create Claude Session
```bash
# Click "New Session" in web interface
# Creates git worktree + starts Claude in web terminal
# Interact directly through browser terminal
```

## Components

- [`cw/`](./cw/) - Claude Worktree Fish shell function
- [`cm/`](./cm/) - Claude Manager Go TUI application
- [`docs/`](./docs/) - Detailed documentation and guides
- [`examples/`](./examples/) - Usage examples and workflows

## Benefits

- **Parallel Development**: Work on multiple features simultaneously without context switching
- **Isolated Environments**: Each Claude session operates in its own worktree
- **Real-time Monitoring**: Track all Claude sessions from a central dashboard
- **Performance**: Go-based TUI for efficient resource usage
- **Integration**: Seamless workflow with git worktrees and Claude Code

## Documentation

- [Architecture Overview](./docs/architecture.md)
- [CW (Claude Worktree) Guide](./docs/cw-guide.md)
- [CM (Claude Manager) Guide](./docs/cm-guide.md)
- [Development Workflows](./docs/workflows.md)
- [Technical Specifications](./docs/technical-specs.md)

## Status

- ✅ `cw` - Completed and functional
- 🚧 `cm` - In development
- 📋 Documentation - In progress