package session

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// Session represents a Claude Code session
type Session struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	Branch   string    `json:"branch"`
	PID      int       `json:"pid"`
	Status   string    `json:"status"`
	Created  time.Time `json:"created"`
	LastSeen time.Time `json:"last_seen"`
}

// CreateRequest represents a session creation request
type CreateRequest struct {
	Name        string `json:"name"`
	RepoPath    string `json:"repoPath"`
	BranchName  string `json:"branchName"`
	BaseBranch  string `json:"baseBranch"`
	UseWorktree bool   `json:"useWorktree"`
}

// KillRequest represents a session kill request
type KillRequest struct {
	SessionID string `json:"sessionId"`
}

// NewSession creates a new session with defaults
func NewSession(name, path, branch string) *Session {
	return &Session{
		ID:       generateSessionID(),
		Name:     name,
		Path:     path,
		Branch:   branch,
		Status:   "starting",
		Created:  time.Now(),
		LastSeen: time.Now(),
	}
}

// UpdateLastSeen updates the last seen timestamp
func (s *Session) UpdateLastSeen() {
	s.LastSeen = time.Now()
}

// SetStatus updates the session status
func (s *Session) SetStatus(status string) {
	s.Status = status
	s.UpdateLastSeen()
}

// generateSessionID generates a unique session ID
func generateSessionID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if crypto/rand fails
		return fmt.Sprintf("session-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}