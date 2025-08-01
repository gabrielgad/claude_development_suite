package terminal

import (
	"io"
	"log"
	"sync"
)

// Manager manages all active PTY sessions
type Manager struct {
	sessions map[string]*PTYSession
	mu       sync.RWMutex
}

// NewManager creates a new terminal manager
func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]*PTYSession),
	}
}

// Add adds a new PTY session to the manager
func (m *Manager) Add(session *PTYSession) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[session.ID] = session
	
	// Start reading from PTY and broadcasting to clients
	go m.readPTYOutput(session)
}

// Get retrieves a PTY session by ID
func (m *Manager) Get(id string) (*PTYSession, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	session, exists := m.sessions[id]
	return session, exists
}

// Remove removes a PTY session from the manager
func (m *Manager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if session, exists := m.sessions[id]; exists {
		session.Cleanup()
		delete(m.sessions, id)
	}
}

// List returns all PTY sessions
func (m *Manager) List() []*PTYSession {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	sessions := make([]*PTYSession, 0, len(m.sessions))
	for _, session := range m.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// CleanupAll cleans up all PTY sessions
func (m *Manager) CleanupAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for id, session := range m.sessions {
		session.Cleanup()
		delete(m.sessions, id)
	}
}

// readPTYOutput continuously reads from PTY and broadcasts to clients
func (m *Manager) readPTYOutput(session *PTYSession) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("PTY reader panic for session %s: %v", session.ID, r)
		}
	}()

	buffer := make([]byte, 1024)
	for {
		n, err := session.PTY.Read(buffer)
		if err != nil {
			if err == io.EOF {
				log.Printf("PTY session %s ended", session.ID)
			} else {
				log.Printf("PTY read error for session %s: %v", session.ID, err)
			}
			m.Remove(session.ID)
			return
		}

		if n > 0 {
			session.BroadcastToClients(buffer[:n])
		}
	}
}