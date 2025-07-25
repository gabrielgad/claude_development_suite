# CW (Claude Worktree) - User Guide

## Overview

`cw` is a Fish shell function that streamlines the creation and management of git worktrees with Claude Code sessions. It enables parallel development by providing isolated environments for different features, bugs, or experiments.

## Installation

The `cw` function is installed at `~/.config/fish/functions/cw.fish`. It should be available immediately in any new Fish shell session.

## Commands

### `cw` or `cw make`
Creates a new git worktree and launches Claude Code.

**Interactive Prompts:**
1. **Directory name**: Name for the worktree directory
2. **Branch name**: Git branch name (e.g., `feature/auth-system`)
3. **Base branch**: Branch to base the new branch on (defaults to current)
4. **Dependencies**: Whether to install npm dependencies
5. **Environment**: Whether to copy `.env` file
6. **Claude prompt**: Optional initial prompt for Claude

**Example:**
```bash
cd /path/to/your/project
cw make
# or just
cw
```

### `cw list`
Lists all current git worktrees.

**Example:**
```bash
cw list
# Output:
# Current worktrees:
# /home/user/project           abc123 [main]
# /home/user/project-auth      def456 [feature/auth]
# /home/user/project-api       ghi789 [feature/api-v2]
```

### `cw remove`
Removes a git worktree with confirmation.

**Interactive Flow:**
1. Shows current worktrees
2. Prompts for worktree path to remove
3. Asks for confirmation
4. Removes the worktree

**Example:**
```bash
cw remove
# Current worktrees:
# /home/user/project           abc123 [main]
# /home/user/project-auth      def456 [feature/auth]
# 
# Enter path of worktree to remove: /home/user/project-auth
# Are you sure you want to remove /home/user/project-auth? (y/N): y
# Worktree removed successfully
```

### `cw prune`
Cleans up stale worktree references.

**Example:**
```bash
cw prune
# Pruned stale worktree references
```

### `cw help`
Shows usage information and available commands.

## Workflow Examples

### Feature Development

```bash
# Start working on user authentication
cd ~/projects/webapp
cw make

# Prompts:
# Enter directory name for the worktree: auth-feature
# Enter branch name: feature/user-authentication  
# Base branch (default: main): [press enter]
# Install npm dependencies? (y/N): y
# Copy .env file? (y/N): y
# Enter initial prompt for Claude: Implement JWT authentication system

# Result: New worktree at ~/projects/webapp-auth-feature with Claude running
```

### Bug Fix

```bash
# Quick hotfix for production issue
cd ~/projects/webapp
cw make

# Prompts:
# Enter directory name for the worktree: cache-fix
# Enter branch name: hotfix/cache-invalidation
# Base branch (default: main): [press enter]
# Install npm dependencies? (y/N): n  # Skip for quick fix
# Copy .env file? (y/N): y
# Enter initial prompt for Claude: Fix cache invalidation bug in user service

# Result: Ready to work on hotfix in isolated environment
```

### Experiment/Prototype

```bash
# Try out a new approach
cd ~/projects/webapp
cw make

# Prompts:
# Enter directory name: experiment-graphql
# Enter branch name: experiment/graphql-migration
# Base branch: feature/api-v2  # Base on existing work
# ...

# Result: Safe environment to experiment without affecting main work
```

## Directory Structure

When you create a worktree, `cw` generates this structure:

```
/parent/directory/
├── project/                    # Main repository
│   ├── .git/                  # Git database (shared)
│   ├── src/
│   └── package.json
├── project-auth-feature/       # Worktree for auth feature
│   ├── src/                   # Same files, different branch
│   ├── package.json
│   ├── .env                   # Copied from main
│   └── node_modules/          # Separate dependencies
└── project-api-refactor/       # Another worktree
    ├── src/
    └── ...
```

## Features

### Smart Environment Setup

- **Dependency Management**: Automatically offers to run `npm install`
- **Environment Variables**: Copies `.env` files from main repository
- **Directory Naming**: Creates descriptive directory names based on your input

### Git Integration

- **Branch Creation**: Creates new branches based on specified base branch
- **Worktree Management**: Uses git's native worktree functionality
- **Shared History**: All worktrees share the same git history and remotes

### Claude Integration

- **Automatic Launch**: Starts Claude Code in the new worktree directory
- **Context Preservation**: Claude has full access to the project codebase
- **Initial Prompts**: Optional initial context for Claude

### Safety Features

- **Validation**: Prevents creating worktrees with existing names
- **Confirmation**: Asks for confirmation before destructive operations
- **Error Handling**: Graceful error messages for common issues

## Configuration

### Fish Function Location
```
~/.config/fish/functions/cw.fish
```

### Customization

You can modify the `cw` function to suit your workflow:

```fish
# Example: Add custom base directories
function cw --description "Manage git worktrees and Claude Code sessions"
    # Add your customizations here
    # ...
end
```

## Troubleshooting

### Common Issues

**"Not in a git repository"**
- Ensure you're in a git repository before running `cw`
- Check that `.git` directory exists

**"Directory already exists"**
- Choose a different directory name
- Remove existing directory if it's stale

**"Failed to create worktree"**
- Check git status for uncommitted changes
- Ensure branch name doesn't already exist
- Verify you have write permissions

**"Claude command not found"**
- Ensure Claude Code is installed and in your PATH
- Try running `claude --version` to verify installation

### Debug Mode

For detailed debugging, you can modify the function to add verbose output:

```fish
# Add this to see all git commands
git worktree add "$worktree_path" -b "$branch_name" "$base_branch" --verbose
```

## Best Practices

### Naming Conventions

- **Directory names**: Use descriptive, lowercase names (`auth-feature`, `bug-cache-fix`)
- **Branch names**: Follow your team's convention (`feature/`, `hotfix/`, `experiment/`)

### Resource Management

- **Dependencies**: Only install dependencies when needed to save time and disk space
- **Cleanup**: Regularly use `cw remove` and `cw prune` to clean up finished work
- **Environment files**: Be careful with secrets in copied `.env` files

### Parallel Development

- **Port conflicts**: Use different ports if running multiple dev servers
- **Database connections**: Be aware of shared database connections
- **File locks**: Some tools may conflict across worktrees

### Integration with CM

The `cw` tool works seamlessly with `cm` (Claude Manager):

1. Create worktrees with `cw`
2. Monitor all sessions with `cm`
3. Switch between contexts efficiently

## Advanced Usage

### Scripted Worktree Creation

```bash
# Create multiple worktrees for related features
for feature in auth api frontend; do
    # You'd need to automate the prompts for this
    echo "Creating worktree for $feature"
done
```

### Custom Base Branches

```bash
# Create feature branch based on another feature
cw make
# When prompted for base branch, enter: feature/existing-work
```

### Integration with IDE

Many IDEs can work with multiple worktrees:

- **VS Code**: Open each worktree in a separate window
- **IntelliJ**: Create separate projects for each worktree
- **Vim**: Use different sessions or tabs