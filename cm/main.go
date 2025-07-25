package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shirou/gopsutil/v3/process"
)

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

// Model represents the application state
type Model struct {
	sessions      []Session
	selectedIndex int
	currentView   ViewMode
	updateChan    chan SessionUpdate
	quitting      bool
	windowWidth   int
	windowHeight  int
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

		// Get process creation time
		createTime, _ := proc.CreateTime()
		createdAt := time.Unix(createTime/1000, 0)

		session := Session{
			ID:        fmt.Sprintf("%d", proc.Pid),
			Name:      filepath.Base(cwd),
			Path:      cwd,
			Branch:    branch,
			PID:       int(proc.Pid),
			Status:    StatusActive,
			LastSeen:  time.Now(),
			CreatedAt: createdAt,
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

// Bubbletea implementation
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
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00ff00")).
		Padding(0, 1)

	b.WriteString(headerStyle.Render("Claude Manager - Sessions"))
	b.WriteString(fmt.Sprintf(" (%d active)\n\n", len(m.sessions)))

	// Sessions list
	for i, session := range m.sessions {
		style := lipgloss.NewStyle().Padding(0, 1)
		if i == m.selectedIndex {
			style = style.Background(lipgloss.Color("#444444"))
		}

		statusColor := "#ff0000" // red for inactive
		if session.Status == StatusActive {
			statusColor = "#00ff00" // green for active
		}

		line := fmt.Sprintf("● %-20s [%-15s] %s %s",
			session.Name,
			session.Branch,
			lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor)).Render(fmt.Sprintf("%-10s", session.Status.String())),
			session.Path,
		)

		b.WriteString(style.Render(line))
		b.WriteString("\n")
	}

	if len(m.sessions) == 0 {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Render("  No Claude sessions found. Use 'n' to create a new session."))
		b.WriteString("\n")
	}

	// Footer
	b.WriteString("\n")
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Padding(0, 1)

	b.WriteString(footerStyle.Render("Controls: ↑/↓ navigate, Enter details, r refresh, n new, k kill, q quit"))

	return b.String()
}

func (m Model) renderDetailsView() string {
	if len(m.sessions) == 0 || m.selectedIndex >= len(m.sessions) {
		return "No session selected\n\nPress Enter to return to sessions list"
	}

	session := m.sessions[m.selectedIndex]

	var b strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00ff00")).
		Padding(0, 1)

	b.WriteString(headerStyle.Render(fmt.Sprintf("Session Details: %s", session.Name)))
	b.WriteString("\n\n")

	// Session details
	detailStyle := lipgloss.NewStyle().Padding(0, 1)

	details := []string{
		fmt.Sprintf("Path: %s", session.Path),
		fmt.Sprintf("Branch: %s", session.Branch),
		fmt.Sprintf("PID: %d", session.PID),
		fmt.Sprintf("Status: %s", session.Status.String()),
		fmt.Sprintf("Last Seen: %s", session.LastSeen.Format("2006-01-02 15:04:05")),
		fmt.Sprintf("Created: %s", session.CreatedAt.Format("2006-01-02 15:04:05")),
	}

	for _, detail := range details {
		b.WriteString(detailStyle.Render(detail))
		b.WriteString("\n")
	}

	if session.LastPrompt != "" {
		b.WriteString("\n")
		b.WriteString(detailStyle.Render(fmt.Sprintf("Last Prompt: %s", session.LastPrompt)))
		b.WriteString("\n")
	}

	// Footer
	b.WriteString("\n")
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Padding(0, 1)

	b.WriteString(footerStyle.Render("Press Enter to return to sessions list"))

	return b.String()
}

// Command handlers
func (m Model) startProcessMonitor() tea.Cmd {
	return func() tea.Msg {
		monitor := NewProcessMonitor(m.updateChan)
		go monitor.Start()
		return SessionUpdate{Type: UpdateSessionStatus} // Trigger initial scan
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
		// Execute cw command in background
		cmd := exec.Command("fish", "-c", "cw make")
		err := cmd.Start()
		if err != nil {
			return SessionUpdate{Error: fmt.Errorf("failed to create new session: %w", err)}
		}
		return SessionUpdate{Type: UpdateSessionStatus} // Refresh after creation
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
			return SessionUpdate{Error: fmt.Errorf("failed to find process: %w", err)}
		}

		err = proc.Terminate()
		if err != nil {
			return SessionUpdate{Error: fmt.Errorf("failed to terminate process: %w", err)}
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
			if m.selectedIndex >= len(m.sessions) && len(m.sessions) > 0 {
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
				if m.selectedIndex >= len(m.sessions) && len(m.sessions) > 0 {
					m.selectedIndex = len(m.sessions) - 1
				}
				if len(m.sessions) == 0 {
					m.selectedIndex = 0
				}
				break
			}
		}
	}

	if update.Error != nil {
		// For now, just log errors
		// In the future, we could show them in a status bar
		log.Printf("Error: %v", update.Error)
	}

	return m, nil
}

func main() {
	// Create channels for communication
	updateChan := make(chan SessionUpdate, 100)

	// Initialize model
	model := Model{
		sessions:    make([]Session, 0),
		currentView: ViewSessions,
		updateChan:  updateChan,
	}

	// Create program
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Start the program
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}