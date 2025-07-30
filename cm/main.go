package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
)

const VERSION = "2.0.0-web"

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

// PTYSession manages a pseudoterminal session
type PTYSession struct {
	ID      string
	PTY     *os.File
	Cmd     *exec.Cmd
	Session *Session
	Clients map[*websocket.Conn]bool
	mu      sync.RWMutex
}

// SessionManager manages all active sessions
type SessionManager struct {
	sessions map[string]*PTYSession
	mu       sync.RWMutex
}

var (
	sessionManager = &SessionManager{
		sessions: make(map[string]*PTYSession),
	}
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for development
		},
	}
)

func main() {
	var (
		serve   = flag.Bool("serve", false, "Start web server mode")
		port    = flag.Int("port", 8080, "Web server port")
		version = flag.Bool("version", false, "Show version")
	)
	flag.Parse()

	if *version {
		fmt.Printf("Claude Manager v%s (Web Terminal Edition)\n", VERSION)
		return
	}

	if *serve {
		log.Printf("Starting Claude Manager Web Server on port %d", *port)
		startWebServer(*port)
	} else {
		// Default behavior - start web server
		log.Printf("Starting Claude Manager Web Server on port %d", *port)
		log.Printf("Open http://localhost:%d in your browser", *port)
		startWebServer(*port)
	}
}

func startWebServer(port int) {
	// Create web directory structure
	err := ensureWebDirectory()
	if err != nil {
		log.Fatalf("Failed to create web directory: %v", err)
	}

	// Set up HTTP routes
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/favicon.ico", handleFavicon)
	http.HandleFunc("/api/sessions", handleSessions)
	http.HandleFunc("/api/sessions/create", handleCreateSession)
	http.HandleFunc("/api/sessions/kill", handleKillSession)
	http.HandleFunc("/api/directories", handleDirectories)
	http.HandleFunc("/api/git-repos", handleGitRepos)
	http.HandleFunc("/ws/", handleWebSocket)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))

	// Create server with graceful shutdown
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: nil,
	}

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Close all PTY sessions
		sessionManager.mu.Lock()
		for _, session := range sessionManager.sessions {
			session.cleanup()
		}
		sessionManager.mu.Unlock()

		server.Shutdown(ctx)
	}()

	log.Printf("🚀 Claude Manager Web Server started")
	log.Printf("📡 Server: http://localhost:%d", port)
	log.Printf("🌐 Open your browser to manage Claude sessions")
	log.Printf("Press Ctrl+C to stop")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}

func ensureWebDirectory() error {
	webDir := "web"
	staticDir := filepath.Join(webDir, "static")

	// Create directories
	if err := os.MkdirAll(staticDir, 0755); err != nil {
		return err
	}

	// Create HTML template
	htmlTemplate := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Claude Manager - Web Terminal</title>
    <link rel="stylesheet" href="/static/app.css">
    <script src="https://unpkg.com/xterm@5.3.0/lib/xterm.js"></script>
    <link rel="stylesheet" href="https://unpkg.com/xterm@5.3.0/css/xterm.css" />
</head>
<body>
    <div id="app">
        <header>
            <h1>🚀 Claude Manager</h1>
            <div class="controls">
                <button id="new-session">New Session</button>
                <button id="refresh">Refresh</button>
            </div>
        </header>
        
        <div class="main-content">
            <div class="sidebar">
                <h3>Sessions</h3>
                <div id="sessions-list">
                    <div class="no-sessions">No active sessions</div>
                </div>
            </div>
            
            <div class="terminal-area">
                <div id="terminal-tabs"></div>
                <div id="terminal-container">
                    <div class="welcome">
                        <h2>Welcome to Claude Manager</h2>
                        <p>Create a new session to get started with Claude Code in your browser.</p>
                        <button class="create-session-btn" onclick="app.showSessionCreator()">Create First Session</button>
                    </div>
                    
                    <div id="session-creator" class="session-creator" style="display: none;">
                        <div class="creator-content">
                            <h3>Create New Session</h3>
                            
                            <div class="form-group">
                                <label>Session Name:</label>
                                <input type="text" id="session-name" placeholder="e.g., auth-feature, api-refactor">
                            </div>
                            
                            <div class="form-group">
                                <label>Repository Location:</label>
                                <div class="directory-browser">
                                    <div class="current-path">
                                        <span id="current-path-text">Loading...</span>
                                        <button type="button" id="go-home" class="path-btn">🏠 Home</button>
                                        <button type="button" id="go-root" class="path-btn">📁 Root</button>
                                    </div>
                                    <div class="directory-list" id="directory-list">
                                        <div class="loading">Loading directories...</div>
                                    </div>
                                    <div class="selected-repo">
                                        <strong>Selected:</strong> <span id="selected-repo-path">None selected</span>
                                    </div>
                                </div>
                            </div>
                            
                            <div class="form-group">
                                <label>
                                    <input type="checkbox" id="use-worktree" checked> 
                                    Create Git Worktree (Recommended)
                                </label>
                                <small>Creates an isolated copy for parallel development</small>
                            </div>
                            
                            <div id="worktree-options" class="form-group">
                                <label>Branch Name:</label>
                                <input type="text" id="branch-name" placeholder="feature/session-name (auto-generated)">
                                
                                <label>Base Branch:</label>
                                <input type="text" id="base-branch" placeholder="main">
                            </div>
                            
                            <div class="creator-actions">
                                <button class="btn-primary" onclick="app.createSessionFromForm()">Create Session</button>
                                <button class="btn-secondary" onclick="app.hideSessionCreator()">Cancel</button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
    
    <script src="/static/app.js"></script>
</body>
</html>`

	cssTemplate := `* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
    background: #1e1e1e;
    color: #ffffff;
    height: 100vh;
    overflow: hidden;
}

#app {
    display: flex;
    flex-direction: column;
    height: 100vh;
}

header {
    background: #2d2d2d;
    padding: 1rem;
    border-bottom: 1px solid #404040;
    display: flex;
    justify-content: space-between;
    align-items: center;
}

header h1 {
    color: #00ff00;
    font-size: 1.5rem;
}

.controls button {
    background: #007acc;
    color: white;
    border: none;
    padding: 0.5rem 1rem;
    margin-left: 0.5rem;
    border-radius: 4px;
    cursor: pointer;
    transition: background 0.2s;
}

.controls button:hover {
    background: #005999;
}

.main-content {
    display: flex;
    flex: 1;
    overflow: hidden;
}

.sidebar {
    width: 300px;
    background: #252526;
    border-right: 1px solid #404040;
    padding: 1rem;
    overflow-y: auto;
}

.sidebar h3 {
    color: #cccccc;
    margin-bottom: 1rem;
    font-size: 1.1rem;
}

.session-item {
    background: #2d2d2d;
    padding: 0.75rem;
    margin-bottom: 0.5rem;
    border-radius: 4px;
    cursor: pointer;
    border-left: 3px solid #007acc;
    transition: background 0.2s;
}

.session-item:hover {
    background: #404040;
}

.session-item.active {
    background: #404040;
    border-left-color: #00ff00;
}

.session-name {
    font-weight: bold;
    color: #ffffff;
}

.session-details {
    font-size: 0.85rem;
    color: #cccccc;
    margin-top: 0.25rem;
}

.session-status {
    display: inline-block;
    padding: 0.2rem 0.5rem;
    border-radius: 3px;
    font-size: 0.75rem;
    margin-top: 0.25rem;
}

.status-active {
    background: #4caf50;
    color: white;
}

.status-idle {
    background: #ff9800;
    color: white;
}

.status-starting {
    background: #2196f3;
    color: white;
}

.status-error {
    background: #f44336;
    color: white;
}

.terminal-area {
    flex: 1;
    display: flex;
    flex-direction: column;
}

#terminal-tabs {
    background: #2d2d2d;
    border-bottom: 1px solid #404040;
    padding: 0;
    display: none;
}

.terminal-tab {
    display: inline-block;
    padding: 0.75rem 1rem;
    background: #1e1e1e;
    color: #cccccc;
    border-right: 1px solid #404040;
    cursor: pointer;
    transition: background 0.2s;
}

.terminal-tab.active {
    background: #2d2d2d;
    color: #ffffff;
}

.terminal-tab:hover {
    background: #404040;
}

#terminal-container {
    flex: 1;
    position: relative;
}

.terminal {
    width: 100%;
    height: 100%;
    display: none;
}

.terminal.active {
    display: block;
}

.welcome {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    text-align: center;
}

.welcome h2 {
    color: #00ff00;
    margin-bottom: 1rem;
}

.welcome p {
    color: #cccccc;
    margin-bottom: 2rem;
    max-width: 400px;
}

.create-session-btn {
    background: #00ff00;
    color: #1e1e1e;
    border: none;
    padding: 1rem 2rem;
    border-radius: 6px;
    font-size: 1.1rem;
    font-weight: bold;
    cursor: pointer;
    transition: all 0.2s;
}

.create-session-btn:hover {
    background: #00cc00;
    transform: translateY(-2px);
}

.no-sessions {
    color: #666666;
    font-style: italic;
    text-align: center;
    padding: 2rem 0;
}

.session-actions {
    margin-top: 0.5rem;
}

.kill-btn {
    background: #dc3545;
    color: white;
    border: none;
    padding: 0.3rem 0.6rem;
    border-radius: 3px;
    font-size: 0.75rem;
    cursor: pointer;
}

.kill-btn:hover {
    background: #c82333;
}

/* Session Creator Styles */
.session-creator {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.8);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
}

.creator-content {
    background: #2d2d2d;
    padding: 2rem;
    border-radius: 8px;
    width: 500px;
    max-height: 80vh;
    overflow-y: auto;
}

.creator-content h3 {
    color: #00ff00;
    margin-bottom: 1.5rem;
    text-align: center;
}

.form-group {
    margin-bottom: 1rem;
}

.form-group label {
    display: block;
    color: #cccccc;
    margin-bottom: 0.5rem;
    font-size: 0.9rem;
}

.form-group input[type="text"],
.form-group select {
    width: 100%;
    padding: 0.75rem;
    background: #1e1e1e;
    border: 1px solid #404040;
    border-radius: 4px;
    color: #ffffff;
    font-size: 0.9rem;
}

.form-group input[type="text"]:focus,
.form-group select:focus {
    outline: none;
    border-color: #007acc;
}

.form-group input[type="checkbox"] {
    margin-right: 0.5rem;
}

.form-group small {
    display: block;
    color: #888888;
    font-size: 0.8rem;
    margin-top: 0.25rem;
}

.creator-actions {
    display: flex;
    gap: 1rem;
    margin-top: 2rem;
    justify-content: center;
}

.btn-primary {
    background: #00ff00;
    color: #1e1e1e;
    border: none;
    padding: 0.75rem 1.5rem;
    border-radius: 4px;
    font-weight: bold;
    cursor: pointer;
    transition: background 0.2s;
}

.btn-primary:hover {
    background: #00cc00;
}

.btn-secondary {
    background: #404040;
    color: #ffffff;
    border: none;
    padding: 0.75rem 1.5rem;
    border-radius: 4px;
    cursor: pointer;
    transition: background 0.2s;
}

.btn-secondary:hover {
    background: #555555;
}

.repo-item {
    display: flex;
    align-items: center;
    padding: 0.5rem;
    border-bottom: 1px solid #404040;
}

.repo-name {
    font-weight: bold;
    color: #ffffff;
}

.repo-path {
    font-size: 0.8rem;
    color: #888888;
    margin-left: 0.5rem;
}

.repo-branch {
    font-size: 0.8rem;
    color: #00ff00;
    margin-left: auto;
}

/* Directory Browser Styles */
.directory-browser {
    border: 1px solid #404040;
    border-radius: 4px;
    background: #1e1e1e;
    overflow: hidden;
}

.current-path {
    background: #2d2d2d;
    padding: 0.75rem;
    border-bottom: 1px solid #404040;
    display: flex;
    align-items: center;
    gap: 0.5rem;
}

.current-path span {
    flex: 1;
    color: #cccccc;
    font-family: monospace;
    font-size: 0.9rem;
}

.path-btn {
    background: #404040;
    color: #cccccc;
    border: none;
    padding: 0.25rem 0.5rem;
    border-radius: 3px;
    font-size: 0.8rem;
    cursor: pointer;
    transition: background 0.2s;
}

.path-btn:hover {
    background: #555555;
}

.directory-list {
    max-height: 200px;
    overflow-y: auto;
    background: #1e1e1e;
}

.directory-item {
    display: flex;
    align-items: center;
    padding: 0.5rem 0.75rem;
    cursor: pointer;
    border-bottom: 1px solid #2d2d2d;
    transition: background 0.2s;
}

.directory-item:hover {
    background: #2d2d2d;
}

.directory-item.selected {
    background: #007acc;
    color: #ffffff;
}

.directory-item.git-repo {
    background: rgba(0, 255, 0, 0.1);
    border-left: 3px solid #00ff00;
}

.directory-item.git-repo.selected {
    background: #007acc;
}

.directory-icon {
    margin-right: 0.5rem;
    font-size: 1rem;
}

.directory-name {
    flex: 1;
    color: #cccccc;
}

.git-indicator {
    font-size: 0.8rem;
    color: #00ff00;
    margin-left: 0.5rem;
}

.selected-repo {
    padding: 0.75rem;
    background: #2d2d2d;
    border-top: 1px solid #404040;
    color: #cccccc;
    font-size: 0.9rem;
}

.selected-repo span {
    color: #00ff00;
    font-family: monospace;
}

.loading {
    padding: 1rem;
    text-align: center;
    color: #888888;
    font-style: italic;
}`

	jsTemplate := `class ClaudeManager {
    constructor() {
        this.sessions = new Map();
        this.activeSessions = [];
        this.currentSession = null;
        this.currentPath = '';
        this.selectedRepoPath = '';
        this.init();
    }

    async init() {
        this.bindEvents();
        await this.loadSessions();
        this.startSessionPolling();
    }

    bindEvents() {
        document.getElementById('new-session').addEventListener('click', () => this.showSessionCreator());
        document.getElementById('refresh').addEventListener('click', () => this.loadSessions());
        
        // Session creator events
        document.getElementById('use-worktree').addEventListener('change', (e) => {
            document.getElementById('worktree-options').style.display = e.target.checked ? 'block' : 'none';
        });
        
        document.getElementById('session-name').addEventListener('input', (e) => {
            const cleanName = this.sanitizeForGit(e.target.value);
            const branchInput = document.getElementById('branch-name');
            
            // Update placeholder with sanitized name
            if (!branchInput.value || branchInput.placeholder.includes('auto-generated')) {
                branchInput.placeholder = ` + "`" + `feature/${cleanName || 'session-name'} (auto-generated)` + "`" + `;
            }
            
            // Show user if their input was changed
            if (e.target.value && cleanName !== e.target.value.toLowerCase()) {
                e.target.style.borderColor = '#ff9800';
                e.target.title = ` + "`" + `Will be cleaned to: ${cleanName}` + "`" + `;
            } else {
                e.target.style.borderColor = '#404040';
                e.target.title = '';
            }
        });
        
        document.getElementById('branch-name').addEventListener('input', (e) => {
            const cleanBranch = this.sanitizeForGit(e.target.value);
            
            // Show user if their input was changed
            if (e.target.value && cleanBranch !== e.target.value) {
                e.target.style.borderColor = '#ff9800';
                e.target.title = ` + "`" + `Will be cleaned to: ${cleanBranch}` + "`" + `;
            } else {
                e.target.style.borderColor = '#404040';
                e.target.title = '';
            }
        });

        // Directory browser events
        document.getElementById('go-home').addEventListener('click', () => this.navigateToHome());
        document.getElementById('go-root').addEventListener('click', () => this.navigateToRoot());
    }

    async loadSessions() {
        try {
            const response = await fetch('/api/sessions');
            const sessions = await response.json();
            this.activeSessions = sessions || [];
            this.renderSessions();
        } catch (error) {
            console.error('Failed to load sessions:', error);
        }
    }

    renderSessions() {
        const sessionsList = document.getElementById('sessions-list');
        
        if (this.activeSessions.length === 0) {
            sessionsList.innerHTML = '<div class="no-sessions">No active sessions</div>';
            return;
        }

        sessionsList.innerHTML = this.activeSessions.map(session => {
            let statusClass = 'status-idle';
            if (session.status === 'active') statusClass = 'status-active';
            else if (session.status === 'starting') statusClass = 'status-starting';
            else if (session.status === 'error') statusClass = 'status-error';
            
            return ` + "`" + `
                <div class="session-item" data-session-id="${session.id}" onclick="app.selectSession('${session.id}')">
                    <div class="session-name">${session.name}</div>
                    <div class="session-details">
                        <div>${session.path}</div>
                        <div>Branch: ${session.branch}</div>
                        <span class="session-status ${statusClass}">${session.status}</span>
                    </div>
                    <div class="session-actions">
                        <button class="kill-btn" onclick="app.killSession('${session.id}', event)">Kill</button>
                    </div>
                </div>
            ` + "`" + `;
        }).join('');
    }

    async loadDirectories(path = '') {
        try {
            const url = path ? ` + "`" + `/api/directories?path=${encodeURIComponent(path)}` + "`" + ` : '/api/directories';
            const response = await fetch(url);
            const data = await response.json();
            
            this.currentPath = data.currentPath;
            this.renderDirectories(data.items);
            document.getElementById('current-path-text').textContent = this.currentPath;
        } catch (error) {
            console.error('Failed to load directories:', error);
            document.getElementById('directory-list').innerHTML = '<div class="loading">Error loading directories</div>';
        }
    }

    renderDirectories(items) {
        const container = document.getElementById('directory-list');
        
        if (items.length === 0) {
            container.innerHTML = '<div class="loading">No directories found</div>';
            return;
        }

        container.innerHTML = items.map(item => {
            const icon = item.name === '..' ? '⬆️' : (item.isGitRepo ? '📦' : '📁');
            const gitIndicator = item.isGitRepo ? '<span class="git-indicator">GIT</span>' : '';
            const classes = ['directory-item'];
            
            if (item.isGitRepo) classes.push('git-repo');
            if (item.path === this.selectedRepoPath) classes.push('selected');
            
            return ` + "`" + `
                <div class="${classes.join(' ')}" data-path="${item.path}" onclick="app.handleDirectoryClick('${item.path}', ${item.isGitRepo})">
                    <span class="directory-icon">${icon}</span>
                    <span class="directory-name">${item.name}</span>
                    ${gitIndicator}
                </div>
            ` + "`" + `;
        }).join('');
    }

    handleDirectoryClick(path, isGitRepo) {
        if (isGitRepo) {
            // Select this git repository
            this.selectedRepoPath = path;
            document.getElementById('selected-repo-path').textContent = path;
            
            // Update selection visual
            document.querySelectorAll('.directory-item').forEach(item => {
                item.classList.remove('selected');
            });
            document.querySelector(` + "`" + `[data-path="${path}"]` + "`" + `).classList.add('selected');
        } else {
            // Navigate to this directory
            this.loadDirectories(path);
        }
    }

    async navigateToHome() {
        this.loadDirectories(); // Empty path defaults to home
    }

    async navigateToRoot() {
        this.loadDirectories('/');
    }

    // Client-side sanitization to match server-side logic
    sanitizeForGit(input) {
        if (!input) return 'unnamed';
        
        // Convert to lowercase
        let result = input.toLowerCase();
        
        // Replace spaces and invalid characters with hyphens
        result = result.replace(/[^a-z0-9\-_./]/g, '-');
        
        // Replace multiple consecutive hyphens with single hyphen
        result = result.replace(/-+/g, '-');
        
        // Remove leading/trailing hyphens and dots
        result = result.replace(/^[-.]|[-.]$/g, '');
        
        // Ensure it's not empty after cleaning
        if (!result) {
            result = 'branch-' + Date.now();
        }
        
        // Ensure max length
        if (result.length > 50) {
            result = result.substring(0, 50).replace(/-+$/, '');
        }
        
        return result;
    }

    showSessionCreator() {
        document.getElementById('session-creator').style.display = 'flex';
        document.getElementById('session-name').focus();
        // Load directory browser starting from home
        this.loadDirectories();
    }

    hideSessionCreator() {
        document.getElementById('session-creator').style.display = 'none';
        // Reset form
        document.getElementById('session-name').value = '';
        document.getElementById('branch-name').value = '';
        document.getElementById('base-branch').value = '';
        document.getElementById('use-worktree').checked = true;
        this.selectedRepoPath = '';
        document.getElementById('selected-repo-path').textContent = 'None selected';
    }

    async createSessionFromForm() {
        const name = document.getElementById('session-name').value.trim();
        const repoPath = this.selectedRepoPath;
        const branchName = document.getElementById('branch-name').value.trim();
        const baseBranch = document.getElementById('base-branch').value.trim() || 'main';
        const useWorktree = document.getElementById('use-worktree').checked;

        if (!name) {
            alert('Please enter a session name');
            return;
        }

        if (!repoPath) {
            alert('Please select a repository from the directory browser');
            return;
        }

        try {
            const response = await fetch('/api/sessions/create', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    name,
                    repoPath,
                    branchName: branchName || ` + "`" + `feature/${name}` + "`" + `,
                    baseBranch,
                    useWorktree
                })
            });

            if (response.ok) {
                const session = await response.json();
                this.hideSessionCreator();
                await this.loadSessions();
                this.selectSession(session.id);
            } else {
                const errorText = await response.text();
                alert('Failed to create session: ' + errorText);
            }
        } catch (error) {
            console.error('Failed to create session:', error);
            alert('Failed to create session: ' + error.message);
        }
    }

    // Legacy method for compatibility
    async createSession() {
        this.showSessionCreator();
    }

    async killSession(sessionId, event) {
        event.stopPropagation();
        
        if (!confirm('Are you sure you want to kill this session?')) {
            return;
        }

        try {
            const response = await fetch('/api/sessions/kill', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ sessionId })
            });

            if (response.ok) {
                this.sessions.delete(sessionId);
                await this.loadSessions();
                
                if (this.currentSession === sessionId) {
                    this.currentSession = null;
                    this.showWelcome();
                }
            } else {
                alert('Failed to kill session');
            }
        } catch (error) {
            console.error('Failed to kill session:', error);
            alert('Failed to kill session');
        }
    }

    selectSession(sessionId) {
        this.currentSession = sessionId;
        
        // Update UI
        document.querySelectorAll('.session-item').forEach(item => {
            item.classList.remove('active');
        });
        document.querySelector(` + "`" + `[data-session-id="${sessionId}"]` + "`" + `).classList.add('active');
        
        this.connectToSession(sessionId);
    }

    connectToSession(sessionId) {
        // Hide welcome screen
        const welcome = document.querySelector('.welcome');
        if (welcome) welcome.style.display = 'none';

        // Create or show terminal
        let terminalDiv = document.getElementById(` + "`" + `terminal-${sessionId}` + "`" + `);
        if (!terminalDiv) {
            terminalDiv = document.createElement('div');
            terminalDiv.id = ` + "`" + `terminal-${sessionId}` + "`" + `;
            terminalDiv.className = 'terminal';
            document.getElementById('terminal-container').appendChild(terminalDiv);

            // Create xterm terminal
            const terminal = new Terminal({
                cursorBlink: true,
                theme: {
                    background: '#1e1e1e',
                    foreground: '#ffffff',
                    cursor: '#00ff00'
                }
            });

            terminal.open(terminalDiv);
            this.sessions.set(sessionId, { terminal, div: terminalDiv });

            // Connect WebSocket
            this.connectWebSocket(sessionId, terminal);
        }

        // Show active terminal
        document.querySelectorAll('.terminal').forEach(t => t.classList.remove('active'));
        terminalDiv.classList.add('active');
        
        // Focus terminal
        const session = this.sessions.get(sessionId);
        if (session) {
            session.terminal.focus();
        }
    }

    connectWebSocket(sessionId, terminal) {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = ` + "`" + `${protocol}//${window.location.host}/ws/${sessionId}` + "`" + `;
        
        const ws = new WebSocket(wsUrl);
        
        ws.onopen = () => {
            console.log(` + "`" + `Connected to session ${sessionId}` + "`" + `);
            
            // Send terminal input to WebSocket
            terminal.onData(data => {
                if (ws.readyState === WebSocket.OPEN) {
                    ws.send(data);
                }
            });
        };

        ws.onmessage = (event) => {
            terminal.write(event.data);
        };

        ws.onclose = () => {
            console.log(` + "`" + `Disconnected from session ${sessionId}` + "`" + `);
            terminal.write('\r\n[Connection closed]\r\n');
        };

        ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            terminal.write('\r\n[Connection error]\r\n');
        };

        // Store WebSocket reference
        const session = this.sessions.get(sessionId);
        if (session) {
            session.websocket = ws;
        }
    }

    showWelcome() {
        document.querySelectorAll('.terminal').forEach(t => t.classList.remove('active'));
        const welcome = document.querySelector('.welcome');
        if (welcome) welcome.style.display = 'flex';
    }

    startSessionPolling() {
        setInterval(() => {
            this.loadSessions();
        }, 5000); // Poll every 5 seconds
    }
}

// Global functions
function createSession() {
    app.createSession();
}

// Initialize app
const app = new ClaudeManager();`

	// Write files
	files := map[string]string{
		filepath.Join(webDir, "index.html"): htmlTemplate,
		filepath.Join(staticDir, "app.css"): cssTemplate,
		filepath.Join(staticDir, "app.js"):  jsTemplate,
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return err
		}
	}

	return nil
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "web/index.html")
}

func handleFavicon(w http.ResponseWriter, r *http.Request) {
	// Simple SVG favicon
	favicon := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16">
		<rect width="16" height="16" fill="#00ff00"/>
		<text x="2" y="12" font-family="monospace" font-size="10" fill="#000">CM</text>
	</svg>`
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Write([]byte(favicon))
}

func handleSessions(w http.ResponseWriter, r *http.Request) {
	sessionManager.mu.RLock()
	sessions := make([]*Session, 0, len(sessionManager.sessions))
	for _, ptySession := range sessionManager.sessions {
		sessions = append(sessions, ptySession.Session)
	}
	sessionManager.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

func handleCreateSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name        string `json:"name"`
		RepoPath    string `json:"repoPath"`
		BranchName  string `json:"branchName"`
		BaseBranch  string `json:"baseBranch"`
		UseWorktree bool   `json:"useWorktree"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var workingPath string
	var err error

	if req.UseWorktree && isGitRepository(req.RepoPath) {
		// Create worktree: projects/repo-name-sessionname
		workingPath, err = createWorktreeForSession(req.RepoPath, req.Name, req.BranchName, req.BaseBranch)
		if err != nil {
			log.Printf("Failed to create worktree: %v", err)
			http.Error(w, fmt.Sprintf("Failed to create worktree: %v", err), http.StatusInternalServerError)
			return
		}
	} else {
		// Use existing directory
		workingPath = req.RepoPath
	}

	session, err := createPTYSession(req.Name, workingPath)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		// Clean up worktree if we created one
		if req.UseWorktree && workingPath != req.RepoPath {
			cleanupWorktree(workingPath)
		}
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session.Session)
}

func createWorktreeForSession(repoPath, sessionName, branchName, baseBranch string) (string, error) {
	if baseBranch == "" {
		baseBranch = "main" // Default base branch
	}

	// Sanitize inputs for git compatibility
	cleanSessionName := sanitizeForGit(sessionName)
	cleanBranchName := branchName
	if cleanBranchName == "" {
		cleanBranchName = fmt.Sprintf("feature/%s", cleanSessionName)
	} else {
		cleanBranchName = sanitizeForGit(cleanBranchName)
	}

	// Create worktree directory name: repo-sessionname (cleaned)
	repoName := filepath.Base(repoPath)
	cleanRepoName := sanitizeForPath(repoName)
	worktreePath := filepath.Join(filepath.Dir(repoPath), fmt.Sprintf("%s-%s", cleanRepoName, cleanSessionName))

	// Check if worktree already exists
	if _, err := os.Stat(worktreePath); err == nil {
		return "", fmt.Errorf("worktree directory already exists: %s", worktreePath)
	}

	// Create the worktree with new branch
	cmd := exec.Command("git", "worktree", "add", "-b", cleanBranchName, worktreePath, baseBranch)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git worktree add failed: %v\nOutput: %s", err, string(output))
	}

	log.Printf("Created worktree: %s (branch: %s, base: %s)", worktreePath, cleanBranchName, baseBranch)
	log.Printf("Sanitized: session '%s' -> '%s', branch '%s' -> '%s'", sessionName, cleanSessionName, branchName, cleanBranchName)
	return worktreePath, nil
}

// sanitizeForGit cleans strings for git branch names and other git identifiers
func sanitizeForGit(input string) string {
	if input == "" {
		return "unnamed"
	}

	// Convert to lowercase
	result := strings.ToLower(input)

	// Replace spaces and invalid characters with hyphens
	invalidChars := regexp.MustCompile(`[^a-z0-9\-_./]`)
	result = invalidChars.ReplaceAllString(result, "-")

	// Replace multiple consecutive hyphens with single hyphen
	multiHyphens := regexp.MustCompile(`-+`)
	result = multiHyphens.ReplaceAllString(result, "-")

	// Remove leading/trailing hyphens and dots
	result = strings.Trim(result, "-.")

	// Ensure it's not empty after cleaning
	if result == "" {
		result = "branch-" + fmt.Sprintf("%d", time.Now().Unix())
	}

	// Ensure max length (git has a limit)
	if len(result) > 50 {
		result = result[:50]
		result = strings.TrimSuffix(result, "-")
	}

	return result
}

// sanitizeForPath cleans strings for filesystem paths
func sanitizeForPath(input string) string {
	if input == "" {
		return "unnamed"
	}

	// Replace invalid filesystem characters with hyphens
	invalidChars := regexp.MustCompile(`[^a-zA-Z0-9\-_.]`)
	result := invalidChars.ReplaceAllString(input, "-")

	// Replace multiple consecutive hyphens with single hyphen
	multiHyphens := regexp.MustCompile(`-+`)
	result = multiHyphens.ReplaceAllString(result, "-")

	// Remove leading/trailing hyphens
	result = strings.Trim(result, "-")

	// Ensure it's not empty after cleaning
	if result == "" {
		result = "directory-" + fmt.Sprintf("%d", time.Now().Unix())
	}

	return result
}

func cleanupWorktree(worktreePath string) {
	// Find the main repo to run git worktree remove from
	cmd := exec.Command("git", "worktree", "remove", worktreePath, "--force")
	if err := cmd.Run(); err != nil {
		log.Printf("Failed to cleanup worktree %s: %v", worktreePath, err)
		// Fallback: remove directory manually
		os.RemoveAll(worktreePath)
	}
}

func handleKillSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SessionID string `json:"sessionId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	sessionManager.mu.Lock()
	ptySession, exists := sessionManager.sessions[req.SessionID]
	if exists {
		delete(sessionManager.sessions, req.SessionID)
	}
	sessionManager.mu.Unlock()

	if exists {
		ptySession.cleanup()
	}

	w.WriteHeader(http.StatusOK)
}

type DirectoryInfo struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	IsDir     bool   `json:"isDir"`
	IsGitRepo bool   `json:"isGitRepo"`
}

func handleDirectories(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		// Start from user's home directory
		if homeDir, err := os.UserHomeDir(); err == nil {
			path = homeDir
		} else {
			path = "/"
		}
	}

	// Clean and validate path
	path = filepath.Clean(path)

	entries, err := os.ReadDir(path)
	if err != nil {
		log.Printf("Cannot read directory %s: %v", path, err)
		http.Error(w, fmt.Sprintf("Cannot read directory: %v", err), http.StatusBadRequest)
		return
	}

	var items []DirectoryInfo

	// Add parent directory if not at filesystem root
	if path != "/" {
		parentPath := filepath.Dir(path)
		items = append(items, DirectoryInfo{
			Name:  "..",
			Path:  parentPath,
			IsDir: true,
		})
	}

	// Add all directories (and optionally files)
	for _, entry := range entries {
		fullPath := filepath.Join(path, entry.Name())

		// Skip hidden files/directories unless specifically requested
		if strings.HasPrefix(entry.Name(), ".") && entry.Name() != ".." {
			continue
		}

		if entry.IsDir() {
			isGitRepo := isGitRepository(fullPath)

			items = append(items, DirectoryInfo{
				Name:      entry.Name(),
				Path:      fullPath,
				IsDir:     true,
				IsGitRepo: isGitRepo,
			})
		}
	}

	response := struct {
		CurrentPath string          `json:"currentPath"`
		Items       []DirectoryInfo `json:"items"`
	}{
		CurrentPath: path,
		Items:       items,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type GitRepoInfo struct {
	Name          string   `json:"name"`
	Path          string   `json:"path"`
	CurrentBranch string   `json:"currentBranch"`
	Worktrees     []string `json:"worktrees"`
}

func handleGitRepos(w http.ResponseWriter, r *http.Request) {
	searchPath := r.URL.Query().Get("path")
	if searchPath == "" {
		// Default to user's home directory
		if homeDir, err := os.UserHomeDir(); err == nil {
			searchPath = homeDir
		} else {
			searchPath = "/"
		}
	}

	repos := findGitRepositories(searchPath)

	var gitRepos []GitRepoInfo
	for _, repoPath := range repos {
		branch := getGitBranch(repoPath)
		worktrees := getGitWorktrees(repoPath)

		gitRepos = append(gitRepos, GitRepoInfo{
			Name:          filepath.Base(repoPath),
			Path:          repoPath,
			CurrentBranch: branch,
			Worktrees:     worktrees,
		})
	}

	response := struct {
		SearchPath string        `json:"searchPath"`
		Repos      []GitRepoInfo `json:"repos"`
	}{
		SearchPath: searchPath,
		Repos:      gitRepos,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func isGitRepository(path string) bool {
	gitDir := filepath.Join(path, ".git")
	if info, err := os.Stat(gitDir); err == nil {
		return info.IsDir()
	}
	return false
}

func findGitRepositories(baseDir string) []string {
	var repos []string

	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if info.IsDir() && info.Name() == ".git" {
			repoPath := filepath.Dir(path)
			repos = append(repos, repoPath)
			return filepath.SkipDir // Don't descend into .git directories
		}

		// Skip deep nesting
		depth := strings.Count(strings.TrimPrefix(path, baseDir), string(filepath.Separator))
		if depth > 3 {
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		log.Printf("Error walking directory %s: %v", baseDir, err)
	}

	return repos
}

func getGitWorktrees(repoPath string) []string {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return []string{}
	}

	var worktrees []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "worktree ") {
			path := strings.TrimPrefix(line, "worktree ")
			worktrees = append(worktrees, path)
		}
	}

	return worktrees
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Path[len("/ws/"):]
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	sessionManager.mu.Lock()
	ptySession, exists := sessionManager.sessions[sessionID]
	sessionManager.mu.Unlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Add client to session
	ptySession.mu.Lock()
	ptySession.Clients[conn] = true
	ptySession.mu.Unlock()

	// Remove client when done
	defer func() {
		ptySession.mu.Lock()
		delete(ptySession.Clients, conn)
		ptySession.mu.Unlock()
	}()

	// Handle WebSocket messages (terminal input)
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				break
			}

			// Write to PTY
			if ptySession.PTY != nil {
				ptySession.PTY.Write(message)
			}
		}
	}()

	// Keep connection alive
	select {}
}

func createPTYSession(name, path string) (*PTYSession, error) {
	sessionID := fmt.Sprintf("session_%d", time.Now().Unix())

	// Check if directory exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory does not exist: %s", path)
	}

	// Get git branch first (faster than starting claude)
	branch := getGitBranch(path)

	// Create PTY session first, then start command
	session := &Session{
		ID:       sessionID,
		Name:     name,
		Path:     path,
		Branch:   branch,
		PID:      0, // Will be set after command starts
		Status:   "starting",
		Created:  time.Now(),
		LastSeen: time.Now(),
	}

	ptySession := &PTYSession{
		ID:      sessionID,
		Session: session,
		Clients: make(map[*websocket.Conn]bool),
	}

	// Store session immediately so UI can show it
	sessionManager.mu.Lock()
	sessionManager.sessions[sessionID] = ptySession
	sessionManager.mu.Unlock()

	// Start Claude Code with PTY in background
	go func() {
		// Try different shell first, then Claude
		var cmd *exec.Cmd

		// Check if claude command exists
		if _, err := exec.LookPath("claude"); err == nil {
			cmd = exec.Command("claude", "--help")
		} else {
			// Fallback to bash shell for testing
			cmd = exec.Command("bash")
		}

		cmd.Dir = path
		cmd.Env = append(os.Environ(), "TERM=xterm-256color")

		// Create PTY
		ptyFile, err := pty.Start(cmd)
		if err != nil {
			log.Printf("Failed to start PTY for session %s: %v", sessionID, err)
			session.Status = "error"
			return
		}

		// Update session with PTY info
		ptySession.PTY = ptyFile
		ptySession.Cmd = cmd
		session.PID = cmd.Process.Pid
		session.Status = "active"

		// Start output forwarder
		go ptySession.forwardOutput()

		// Monitor process
		go ptySession.monitorProcess()

		log.Printf("Started session %s with PID %d", sessionID, cmd.Process.Pid)
	}()

	return ptySession, nil
}

func getGitBranch(dir string) string {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

func (pts *PTYSession) forwardOutput() {
	buffer := make([]byte, 1024)
	for {
		n, err := pts.PTY.Read(buffer)
		if err != nil {
			break
		}

		data := buffer[:n]

		// Convert to UTF-8 safe string, replacing invalid bytes
		output := string(data)
		if !isValidUTF8(data) {
			// Convert invalid UTF-8 to safe representation
			output = strings.ToValidUTF8(string(data), "�")
		}

		pts.mu.RLock()
		for client := range pts.Clients {
			if err := client.WriteMessage(websocket.TextMessage, []byte(output)); err != nil {
				// Client disconnected, will be cleaned up elsewhere
				continue
			}
		}
		pts.mu.RUnlock()
	}
}

func isValidUTF8(data []byte) bool {
	return utf8.Valid(data)
}

func (pts *PTYSession) monitorProcess() {
	pts.Cmd.Wait()
	pts.cleanup()
}

func (pts *PTYSession) cleanup() {
	if pts.PTY != nil {
		pts.PTY.Close()
	}
	if pts.Cmd != nil && pts.Cmd.Process != nil {
		pts.Cmd.Process.Kill()
	}

	pts.mu.Lock()
	for client := range pts.Clients {
		client.Close()
	}
	pts.mu.Unlock()
}
