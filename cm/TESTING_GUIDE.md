# Claude Manager Testing Guide

## Quick Testing Options

### 1. Run from Build Directory
```bash
cd /home/ggadbois/projects/tooling/claude_development_suite-week1/cm
./build/cm
```

### 2. Run from Installed Binary
```bash
# If ~/.local/bin is in your PATH
cm

# Or run directly
~/.local/bin/cm
```

### 3. Build and Run in One Command
```bash
cd /home/ggadbois/projects/tooling/claude_development_suite-week1/cm
make run
```

## What You'll See

### Initial Screen
- **Header**: "Claude Manager - Sessions (X active)"
- **Session List**: Shows any detected Claude processes
- **Footer**: Keyboard controls help

### If No Claude Sessions Found
```
Claude Manager - Sessions (0 active)

  No Claude sessions found. Use 'n' to create a new session.

Controls: ↑/↓ navigate, Enter details, r refresh, n new, k kill, q quit
```

### If Claude Sessions Are Running
```
Claude Manager - Sessions (2 active)

● project-main        [main]           Active     /path/to/project
● project-feature     [feature/auth]   Active     /path/to/project-feature

Controls: ↑/↓ navigate, Enter details, r refresh, n new, k kill, q quit
```

## Keyboard Controls

- **↑/↓ or j/k**: Navigate session list
- **Enter**: View session details
- **r**: Refresh session list
- **n**: Create new session (runs `cw make`)
- **k**: Kill selected session
- **q or Ctrl+C**: Quit application

## Testing Scenarios

### Test 1: Basic Launch
```bash
cd /home/ggadbois/projects/tooling/claude_development_suite-week1/cm
./build/cm
```
**Expected**: TUI opens, shows session list (likely empty)

### Test 2: With Existing Claude Processes
```bash
# In another terminal, start a Claude session
cd /some/project
claude "help me with this project" &

# Then run cm
./build/cm
```
**Expected**: Shows the Claude session in the list

### Test 3: Test Session Creation
```bash
# Run cm
./build/cm

# Press 'n' to create new session
# This will try to run: fish -c "cw make"
```
**Expected**: Attempts to create new worktree via cw

### Test 4: Configuration Test
```bash
# Check if config was created
cat ~/.config/claude_manager/config.json

# Run cm to see themed interface
./build/cm
```
**Expected**: Shows configuration file and themed UI

### Test 5: Git Repository Context
```bash
# Run from a git repository
cd /home/ggadbois/projects/tooling/claude_development_suite-week1
../cm/build/cm
```
**Expected**: If Claude is running here, shows branch info

## Troubleshooting

### Issue: "could not open a new TTY"
This is normal in some environments (like headless servers). The app still works.

### Issue: No sessions detected
- Make sure Claude Code processes are actually running
- Try running: `ps aux | grep claude` to verify
- The app looks for processes with "claude" in the name

### Issue: App doesn't start
```bash
# Check if binary exists and is executable
ls -la build/cm

# Try rebuilding
make clean && make

# Check dependencies
go mod verify
```

### Issue: Fish/CW integration doesn't work
- Make sure Fish shell is installed
- Verify cw function is available:
```bash
fish -c "source /path/to/cw.fish && cw help"
```

## Advanced Testing

### Test Real-time Monitoring
1. Start cm: `./build/cm`
2. In another terminal: Start/stop Claude processes
3. Watch cm update in real-time (every 2 seconds)

### Test Configuration Changes
1. Edit `~/.config/claude_manager/config.json`
2. Change colors or monitor interval
3. Restart cm to see changes

### Test Session Management
1. Start cm with existing Claude sessions
2. Use arrow keys to select a session
3. Press Enter to view details
4. Press k to kill a session (be careful!)

## Performance Testing

### Test with Multiple Sessions
```bash
# Start several Claude processes
cd /project1 && claude "task 1" &
cd /project2 && claude "task 2" &
cd /project3 && claude "task 3" &

# Run cm to see all sessions
./build/cm
```

### Monitor Resource Usage
```bash
# In one terminal
./build/cm

# In another terminal
top -p $(pgrep cm)
# Should show low CPU/memory usage
```

## Exit the Application
- Press `q` or `Ctrl+C` to quit cleanly
- The app should restore your terminal state

## Next Steps
Once you've tested basic functionality, you're ready for Phase 2 enhancements!