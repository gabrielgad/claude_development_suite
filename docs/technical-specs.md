# Technical Specifications

## Claude Manager (CM) - Go Implementation

### Dependencies

```go
// go.mod
module github.com/user/claude-manager

go 1.21

require (
    github.com/charmbracelet/bubbletea v0.25.0
    github.com/charmbracelet/lipgloss v0.9.1
    github.com/shirou/gopsutil/v3 v3.23.12
    github.com/fsnotify/fsnotify v1.7.0
    github.com/spf13/cobra v1.8.0
)
```

### Core Data Structures

```go
// Session represents a Claude Code session
type Session struct {
    ID           string            `json:"id"`
    Name         string            `json:"name"`
    Path         string            `json:"path"`
    Branch       string            `json:"branch"`
    PID          int               `json:"pid"`
    Status       SessionStatus     `json:"status"`
    LastSeen     time.Time         `json:"last_seen"`
    LastPrompt   string            `json:"last_prompt"`
    LastResponse string            `json:"last_response"`
    CreatedAt    time.Time         `json:"created_at"`
    Metadata     map[string]string `json:"metadata"`
}

type SessionStatus int

const (
    StatusUnknown SessionStatus = iota
    StatusActive
    StatusIdle
    StatusWorking
    StatusError
    StatusTerminated
)

// Model represents the application state
type Model struct {
    sessions       []Session
    selectedIndex  int
    currentView    ViewMode
    updateChan     chan SessionUpdate
    commandChan    chan Command
    errorChan      chan error
    quitting       bool
    windowWidth    int
    windowHeight   int
}

type ViewMode int

const (
    ViewSessions ViewMode = iota
    ViewDetails
    ViewLogs
)

// SessionUpdate represents state changes
type SessionUpdate struct {
    Type      UpdateType
    SessionID string
    Data      interface{}
    Error     error
}

type UpdateType int

const (
    UpdateSessionAdded UpdateType = iota
    UpdateSessionRemoved
    UpdateSessionStatus
    UpdateSessionResponse
)

// Command represents user actions
type Command struct {
    Type      CommandType
    SessionID string
    Payload   string
}

type CommandType int

const (
    CommandKillSession CommandType = iota
    CommandSendPrompt
    CommandCreateSession
    CommandRefresh
)
```

### Process Discovery Implementation

```go
// ProcessMonitor handles system process monitoring
type ProcessMonitor struct {
    updateChan chan<- SessionUpdate
    interval   time.Duration
    ctx        context.Context
    cancel     context.CancelFunc
}

func NewProcessMonitor(updateChan chan<- SessionUpdate) *ProcessMonitor {
    ctx, cancel := context.WithCancel(context.Background())
    return &ProcessMonitor{
        updateChan: updateChan,
        interval:   2 * time.Second,
        ctx:        ctx,
        cancel:     cancel,
    }
}

func (pm *ProcessMonitor) Start() {
    ticker := time.NewTicker(pm.interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-pm.ctx.Done():
            return
        case <-ticker.C:
            pm.scanProcesses()
        }
    }
}

func (pm *ProcessMonitor) scanProcesses() {
    processes, err := process.Processes()
    if err != nil {
        pm.updateChan <- SessionUpdate{
            Type:  UpdateSessionStatus,
            Error: err,
        }
        return
    }
    
    var claudeSessions []Session
    
    for _, proc := range processes {
        name, err := proc.Name()
        if err != nil {
            continue
        }
        
        if !strings.Contains(strings.ToLower(name), "claude") {
            continue
        }
        
        cwd, err := proc.Cwd()
        if err != nil {
            continue
        }
        
        // Get git branch
        branch := pm.getGitBranch(cwd)
        
        session := Session{
            ID:        fmt.Sprintf("%d", proc.Pid),
            Name:      filepath.Base(cwd),
            Path:      cwd,
            Branch:    branch,
            PID:       int(proc.Pid),
            Status:    StatusActive,
            LastSeen:  time.Now(),
            CreatedAt: time.Now(), // Would be better to get actual process start time
        }
        
        claudeSessions = append(claudeSessions, session)
    }
    
    // Send update
    pm.updateChan <- SessionUpdate{
        Type: UpdateSessionStatus,
        Data: claudeSessions,
    }
}

func (pm *ProcessMonitor) getGitBranch(dir string) string {
    cmd := exec.Command("git", "branch", "--show-current")
    cmd.Dir = dir
    output, err := cmd.Output()
    if err != nil {
        return "unknown"
    }
    return strings.TrimSpace(string(output))
}

func (pm *ProcessMonitor) Stop() {
    pm.cancel()
}
```

### Bubbletea Implementation

```go
// Main application model
func (m Model) Init() tea.Cmd {
    return tea.Batch(
        tea.EnterAltScreen,
        m.startProcessMonitor(),
    )
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        return m.handleKeyPress(msg)
    case tea.WindowSizeMsg:
        m.windowWidth = msg.Width
        m.windowHeight = msg.Height
        return m, nil
    case SessionUpdate:
        return m.handleSessionUpdate(msg)
    case tea.QuitMsg:
        return m, tea.Quit
    }
    return m, nil
}

func (m Model) View() string {
    if m.quitting {
        return "Goodbye!\n"
    }
    
    switch m.currentView {
    case ViewSessions:
        return m.renderSessionsView()
    case ViewDetails:
        return m.renderDetailsView()
    case ViewLogs:
        return m.renderLogsView()
    default:
        return "Unknown view"
    }
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    case "q", "ctrl+c":
        m.quitting = true
        return m, tea.Quit
    case "r":
        return m, m.refreshSessions()
    case "n":
        return m, m.createNewSession()
    case "k":
        return m, m.killSelectedSession()
    case "j", "down":
        if m.selectedIndex < len(m.sessions)-1 {
            m.selectedIndex++
        }
    case "k", "up":
        if m.selectedIndex > 0 {
            m.selectedIndex--
        }
    case "enter":
        // Toggle to details view
        if m.currentView == ViewSessions {
            m.currentView = ViewDetails
        } else {
            m.currentView = ViewSessions
        }
    }
    return m, nil
}

func (m Model) renderSessionsView() string {
    var b strings.Builder
    
    // Header
    b.WriteString(lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("#00ff00")).
        Render("Claude Manager - Sessions"))
    b.WriteString("\n\n")
    
    // Sessions list
    for i, session := range m.sessions {
        style := lipgloss.NewStyle()
        if i == m.selectedIndex {
            style = style.Background(lipgloss.Color("#444444"))
        }
        
        statusColor := "#ff0000" // red for inactive
        if session.Status == StatusActive {
            statusColor = "#00ff00" // green for active
        }
        
        line := fmt.Sprintf("● %s [%s] %s %s",
            session.Name,
            session.Branch,
            lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor)).Render(session.Status.String()),
            session.Path,
        )
        
        b.WriteString(style.Render(line))
        b.WriteString("\n")
    }
    
    // Footer
    b.WriteString("\n")
    b.WriteString("Controls: ↑/↓ navigate, Enter details, r refresh, n new, k kill, q quit")
    
    return b.String()
}

func (m Model) renderDetailsView() string {
    if len(m.sessions) == 0 || m.selectedIndex >= len(m.sessions) {
        return "No session selected"
    }
    
    session := m.sessions[m.selectedIndex]
    
    var b strings.Builder
    
    // Header
    b.WriteString(lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("#00ff00")).
        Render(fmt.Sprintf("Session Details: %s", session.Name)))
    b.WriteString("\n\n")
    
    // Session details
    b.WriteString(fmt.Sprintf("Path: %s\n", session.Path))
    b.WriteString(fmt.Sprintf("Branch: %s\n", session.Branch))
    b.WriteString(fmt.Sprintf("PID: %d\n", session.PID))
    b.WriteString(fmt.Sprintf("Status: %s\n", session.Status.String()))
    b.WriteString(fmt.Sprintf("Last Seen: %s\n", session.LastSeen.Format("2006-01-02 15:04:05")))
    b.WriteString(fmt.Sprintf("Created: %s\n", session.CreatedAt.Format("2006-01-02 15:04:05")))
    
    if session.LastPrompt != "" {
        b.WriteString(fmt.Sprintf("\nLast Prompt: %s\n", session.LastPrompt))
    }
    
    // Footer
    b.WriteString("\nPress Enter to return to sessions list")
    
    return b.String()
}

// Command handlers
func (m Model) startProcessMonitor() tea.Cmd {
    return func() tea.Msg {
        monitor := NewProcessMonitor(m.updateChan)
        go monitor.Start()
        return nil
    }
}

func (m Model) refreshSessions() tea.Cmd {
    return func() tea.Msg {
        // Trigger immediate scan
        return SessionUpdate{Type: UpdateSessionStatus}
    }
}

func (m Model) createNewSession() tea.Cmd {
    return func() tea.Msg {
        // Execute cw command
        cmd := exec.Command("fish", "-c", "cw make")
        err := cmd.Start()
        if err != nil {
            return SessionUpdate{Error: err}
        }
        return nil
    }
}

func (m Model) killSelectedSession() tea.Cmd {
    if len(m.sessions) == 0 || m.selectedIndex >= len(m.sessions) {
        return nil
    }
    
    session := m.sessions[m.selectedIndex]
    
    return func() tea.Msg {
        proc, err := process.NewProcess(int32(session.PID))
        if err != nil {
            return SessionUpdate{Error: err}
        }
        
        err = proc.Terminate()
        if err != nil {
            return SessionUpdate{Error: err}
        }
        
        return SessionUpdate{
            Type:      UpdateSessionRemoved,
            SessionID: session.ID,
        }
    }
}

func (m Model) handleSessionUpdate(update SessionUpdate) (tea.Model, tea.Cmd) {
    switch update.Type {
    case UpdateSessionStatus:
        if sessions, ok := update.Data.([]Session); ok {
            m.sessions = sessions
            // Adjust selected index if needed
            if m.selectedIndex >= len(m.sessions) {
                m.selectedIndex = len(m.sessions) - 1
            }
            if m.selectedIndex < 0 {
                m.selectedIndex = 0
            }
        }
    case UpdateSessionRemoved:
        // Remove session from list
        for i, session := range m.sessions {
            if session.ID == update.SessionID {
                m.sessions = append(m.sessions[:i], m.sessions[i+1:]...)
                if m.selectedIndex >= len(m.sessions) {
                    m.selectedIndex = len(m.sessions) - 1
                }
                break
            }
        }
    }
    
    if update.Error != nil {
        // Handle error (could show in status bar)
        // For now, just ignore
    }
    
    return m, nil
}

// SessionStatus string representation
func (s SessionStatus) String() string {
    switch s {
    case StatusActive:
        return "Active"
    case StatusIdle:
        return "Idle"
    case StatusWorking:
        return "Working"
    case StatusError:
        return "Error"
    case StatusTerminated:
        return "Terminated"
    default:
        return "Unknown"
    }
}
```

### Configuration Management

```go
type Config struct {
    MonitorInterval time.Duration `json:"monitor_interval"`
    MaxSessions     int           `json:"max_sessions"`
    LogLevel        string        `json:"log_level"`
    Theme           Theme         `json:"theme"`
}

type Theme struct {
    ActiveColor   string `json:"active_color"`
    InactiveColor string `json:"inactive_color"`
    SelectedColor string `json:"selected_color"`
}

func LoadConfig() (*Config, error) {
    configPath := filepath.Join(os.Getenv("HOME"), ".config", "claude_manager", "config.json")
    
    // Default config
    config := &Config{
        MonitorInterval: 2 * time.Second,
        MaxSessions:     50,
        LogLevel:        "info",
        Theme: Theme{
            ActiveColor:   "#00ff00",
            InactiveColor: "#ff0000",
            SelectedColor: "#444444",
        },
    }
    
    if _, err := os.Stat(configPath); os.IsNotExist(err) {
        // Create default config
        return config, SaveConfig(config)
    }
    
    data, err := os.ReadFile(configPath)
    if err != nil {
        return config, err
    }
    
    err = json.Unmarshal(data, config)
    return config, err
}

func SaveConfig(config *Config) error {
    configDir := filepath.Join(os.Getenv("HOME"), ".config", "claude_manager")
    os.MkdirAll(configDir, 0755)
    
    configPath := filepath.Join(configDir, "config.json")
    data, err := json.MarshalIndent(config, "", "  ")
    if err != nil {
        return err
    }
    
    return os.WriteFile(configPath, data, 0644)
}
```

### Main Application Entry Point

```go
func main() {
    config, err := LoadConfig()
    if err != nil {
        log.Fatal(err)
    }
    
    // Create channels for communication
    updateChan := make(chan SessionUpdate, 100)
    commandChan := make(chan Command, 10)
    errorChan := make(chan error, 10)
    
    // Initialize model
    model := Model{
        sessions:    make([]Session, 0),
        currentView: ViewSessions,
        updateChan:  updateChan,
        commandChan: commandChan,
        errorChan:   errorChan,
    }
    
    // Create program
    p := tea.NewProgram(model, tea.WithAltScreen())
    
    // Start the program
    if _, err := p.Run(); err != nil {
        log.Fatal(err)
    }
}
```

### Build and Installation

```bash
# Build the application
go build -o cm main.go

# Install to local bin
cp cm ~/.local/bin/

# Or install system-wide
sudo cp cm /usr/local/bin/
```

### Performance Considerations

- **Memory Usage**: Target <50MB for typical usage (10-20 sessions)
- **CPU Usage**: <5% when idle, <15% during active monitoring
- **Update Frequency**: 2-second intervals for process monitoring
- **Concurrent Goroutines**: 3-5 background workers maximum
- **Channel Buffers**: Sized to prevent blocking (100 for updates, 10 for commands)

### Testing Strategy

```go
// Unit tests for core functionality
func TestProcessDiscovery(t *testing.T) {
    // Test process scanning and Claude detection
}

func TestSessionManagement(t *testing.T) {
    // Test session creation, updates, removal
}

func TestGitIntegration(t *testing.T) {
    // Test git worktree detection and branch parsing
}

// Integration tests
func TestEndToEnd(t *testing.T) {
    // Test full workflow with mock processes
}
```