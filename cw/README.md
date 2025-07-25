# CW (Claude Worktree)

A Fish shell function for creating and managing git worktrees with Claude Code sessions.

## Installation

Copy the function to your Fish functions directory:

```bash
cp cw.fish ~/.config/fish/functions/
```

## Quick Start

```bash
# Create a new worktree and start Claude
cw

# List all worktrees  
cw list

# Remove a worktree
cw remove

# Show help
cw help
```

## Features

- Interactive worktree creation with prompts
- Automatic environment setup (dependencies, .env files)
- Integrated Claude Code launching
- Safe worktree management with confirmations
- Git integration and validation

## Documentation

See [../docs/cw-guide.md](../docs/cw-guide.md) for detailed usage instructions.