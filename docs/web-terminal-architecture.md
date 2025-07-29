# Web Terminal Architecture

## Overview

The Claude Development Suite now uses a **web-based terminal interface** that provides direct PTY (pseudoterminal) communication with Claude processes. This eliminates terminal ownership conflicts and provides a rich, browser-based development experience.

## Architecture Diagram

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

## Key Components

### 1. Web Frontend
- **Session Dashboard**: Visual session management
- **Terminal Emulator**: xterm.js for true terminal experience  
- **Permission UI**: Approve/deny Claude requests
- **Real-time Updates**: Live session status and logs

### 2. Go Backend
- **HTTP Server**: Serves web UI and API endpoints
- **WebSocket Handler**: Bidirectional terminal I/O
- **PTY Manager**: Creates and manages pseudoterminals
- **Session Store**: Tracks session state and metadata

### 3. Claude Integration
- **Direct PTY**: Claude gets real terminal interface
- **Git Worktrees**: Each session in isolated worktree
- **Permission Hooks**: Intercept and handle user prompts

## Benefits

✅ **No Terminal Conflicts**: Web server owns PTY, no ownership issues
✅ **Rich UI**: Browser-based dashboard with full session control
✅ **True Terminal**: Claude gets real PTY with colors, cursor control
✅ **Multi-Session**: Handle multiple Claude instances simultaneously  
✅ **Permission Control**: Centralized approval system
✅ **Cross-Platform**: Works everywhere with just a browser
✅ **Session Persistence**: Sessions survive browser refresh
✅ **No Dependencies**: No tmux/zellij/multiplexer requirements

## User Workflow

1. **Start Server**: `./cm serve` starts web server on localhost:8080
2. **Open Browser**: Navigate to dashboard, see session list
3. **Create Session**: Click "New Session", creates git worktree + PTY
4. **Interact**: Full terminal in browser, direct Claude communication
5. **Manage**: Pause, resume, kill sessions from dashboard
6. **Permissions**: Handle Claude requests via web UI

## Technical Stack

- **Backend**: Go with Gorilla WebSocket, PTY library
- **Frontend**: HTML/CSS/JS with xterm.js terminal emulator
- **Communication**: WebSocket for real-time bidirectional I/O
- **Session Storage**: In-memory with optional persistence
- **Git Integration**: CW script for worktree management

This architecture provides the best of both worlds: the power of command-line tools with the convenience of modern web interfaces.