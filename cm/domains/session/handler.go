package session

import (
	"encoding/json"
	"net/http"
)

// Handler handles HTTP requests for sessions
type Handler struct {
	sessionManager *Manager
}

// NewHandler creates a new session handler
func NewHandler(sessionManager *Manager) *Handler {
	return &Handler{
		sessionManager: sessionManager,
	}
}

// HandleSessions handles GET /api/sessions
func (h *Handler) HandleSessions(w http.ResponseWriter, r *http.Request) {
	sessions := h.sessionManager.List()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

// HandleCreateSession handles POST /api/sessions/create
func (h *Handler) HandleCreateSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" {
		http.Error(w, "Session name is required", http.StatusBadRequest)
		return
	}
	if req.RepoPath == "" {
		http.Error(w, "Repository path is required", http.StatusBadRequest)
		return
	}

	// Create new session
	session := NewSession(req.Name, req.RepoPath, req.BranchName)
	h.sessionManager.Add(session)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

// HandleKillSession handles POST /api/sessions/kill
func (h *Handler) HandleKillSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req KillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	// Remove session
	h.sessionManager.Remove(req.SessionID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}