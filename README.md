# Claude Development Suite

A comprehensive toolset for managing multiple Claude Code sessions and git worktrees for parallel development workflows.

## Overview

This suite consists of two main tools:

1. **`cw` (Claude Worktree)** - Fish shell function for creating isolated git worktrees with Claude Code sessions
2. **`cm` (Claude Manager)** - Go-based TUI for monitoring and managing multiple Claude Code sessions

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

## Quick Start

### 1. Install `cw` (Claude Worktree)
```bash
# The cw function is already installed in your Fish config
cw help
```

### 2. Create a new worktree
```bash
# Navigate to your main project
cd /path/to/your/project

# Create a new worktree for feature development
cw make
# or just
cw
```

### 3. Install `cm` (Claude Manager)
```bash
# Build and install the Go binary
cd claude_manager
go build -o ~/.local/bin/cm
```

### 4. Monitor your sessions
```bash
cm
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