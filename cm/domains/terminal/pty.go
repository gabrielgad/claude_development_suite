package terminal

import (
	"os"
	"os/exec"
	"sync"

	"github.com/gorilla/websocket"
	
	"github.com/user/claude-manager/domains/session"
)

// PTYSession manages a pseudoterminal session
type PTYSession struct {
	ID      string
	PTY     *os.File
	Cmd     *exec.Cmd
	Session *session.Session
	Clients map[*websocket.Conn]bool
	Mu      sync.RWMutex
}

// NewPTYSession creates a new PTY session
func NewPTYSession(id string, session *session.Session) (*PTYSession, error) {
	return &PTYSession{
		ID:      id,
		Session: session,
		Clients: make(map[*websocket.Conn]bool),
	}, nil
}

// AddClient adds a websocket client to the PTY session
func (ps *PTYSession) AddClient(conn *websocket.Conn) {
	ps.Mu.Lock()
	defer ps.Mu.Unlock()
	ps.Clients[conn] = true
}

// RemoveClient removes a websocket client from the PTY session
func (ps *PTYSession) RemoveClient(conn *websocket.Conn) {
	ps.Mu.Lock()
	defer ps.Mu.Unlock()
	delete(ps.Clients, conn)
}

// BroadcastToClients sends data to all connected clients
func (ps *PTYSession) BroadcastToClients(data []byte) {
	ps.Mu.RLock()
	defer ps.Mu.RUnlock()

	for client := range ps.Clients {
		if err := client.WriteMessage(websocket.TextMessage, data); err != nil {
			// Remove client on error
			delete(ps.Clients, client)
			client.Close()
		}
	}
}

// WriteInput writes input to the PTY
func (ps *PTYSession) WriteInput(data []byte) error {
	_, err := ps.PTY.Write(data)
	return err
}

// Cleanup closes the PTY and terminates the command
func (ps *PTYSession) Cleanup() {
	ps.Mu.Lock()
	defer ps.Mu.Unlock()

	// Close all websocket connections
	for client := range ps.Clients {
		client.Close()
	}
	ps.Clients = make(map[*websocket.Conn]bool)

	// Close PTY
	if ps.PTY != nil {
		ps.PTY.Close()
	}

	// Terminate command
	if ps.Cmd != nil && ps.Cmd.Process != nil {
		ps.Cmd.Process.Kill()
		ps.Cmd.Wait()
	}
}

// GetClientCount returns the number of connected clients
func (ps *PTYSession) GetClientCount() int {
	ps.Mu.RLock()
	defer ps.Mu.RUnlock()
	return len(ps.Clients)
}