package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
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
	http.HandleFunc("/terminal/", handleTerminal)
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
	templatesDir := filepath.Join(webDir, "templates")

	// Create directories
	if err := os.MkdirAll(staticDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return err
	}

	// Verify that external files exist, create them if they don't
	files := map[string]string{
		filepath.Join(templatesDir, "index.html"): "",
		filepath.Join(staticDir, "app.css"):       "",
		filepath.Join(staticDir, "app.js"):        "",
	}

	for path := range files {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("required web file missing: %s", path)
		}
	}

	return nil
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "web/templates/index.html")
}

func handleTerminal(w http.ResponseWriter, r *http.Request) {
	log.Printf("Terminal request: %s", r.URL.Path)
	
	// Extract session ID from URL path
	sessionID := r.URL.Path[len("/terminal/"):]
	log.Printf("Extracted session ID: %s", sessionID)
	
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	// Find session info
	sessionManager.mu.RLock()
	ptySession, exists := sessionManager.sessions[sessionID]
	sessionManager.mu.RUnlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Prepare template data
	data := struct {
		SessionID   string
		SessionName string
		SessionPath string
	}{
		SessionID:   sessionID,
		SessionName: ptySession.Session.Name,
		SessionPath: ptySession.Session.Path,
	}

	// Parse and execute template
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <title>{{.SessionName}} - Terminal</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/xterm@5.3.0/css/xterm.css" />
    <style>
        body {
            margin: 0;
            padding: 20px;
            background: #1e1e1e;
            color: white;
            font-family: Arial, sans-serif;
        }
        #terminal {
            margin: 20px 0;
            border: 1px solid #444;
        }
        .header {
            background: #2d2d2d;
            padding: 10px;
            border-radius: 4px 4px 0 0;
            border: 1px solid #444;
            border-bottom: none;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .header h3 {
            margin: 0;
            color: #00ff00;
        }
        .back-btn {
            background: #007acc;
            color: white;
            border: none;
            padding: 5px 10px;
            border-radius: 4px;
            cursor: pointer;
        }
        .back-btn:hover {
            background: #005999;
        }
    </style>
</head>
<body>
    <div class="header">
        <h3>{{.SessionName}} - {{.SessionPath}}</h3>
        <button class="back-btn" onclick="window.location.href='/'">Back to Manager</button>
    </div>
    <div id="terminal"></div>
    
    <script src="https://cdn.jsdelivr.net/npm/xterm@5.3.0/lib/xterm.js"></script>
    <script>
        let terminal;
        let websocket;

        // Initialize terminal - exact same as our working test
        document.addEventListener('DOMContentLoaded', function() {
            terminal = new Terminal({
                cursorBlink: true,
                theme: {
                    background: '#1e1e1e',
                    foreground: '#ffffff',
                    cursor: '#00ff00',
                    black: '#000000',
                    red: '#cd3131',
                    green: '#0dbc79',
                    yellow: '#e5e510',
                    blue: '#2472c8',
                    magenta: '#bc3fbc',
                    cyan: '#11a8cd',
                    white: '#e5e5e5',
                    brightBlack: '#666666',
                    brightRed: '#f14c4c',
                    brightGreen: '#23d18b',
                    brightYellow: '#f5f543',
                    brightBlue: '#3b8eea',
                    brightMagenta: '#d670d6',
                    brightCyan: '#29b8db',
                    brightWhite: '#e5e5e5'
                },
                fontSize: 14,
                fontFamily: 'Consolas, "Liberation Mono", Menlo, Courier, monospace'
            });

            terminal.open(document.getElementById('terminal'));
            terminal.write('Terminal initialized successfully!\r\n');
            terminal.write('Connecting to Claude session...\r\n');
            
            // Connect WebSocket - same as working test
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = protocol + '//' + window.location.host + '/ws/{{.SessionID}}';
            
            websocket = new WebSocket(wsUrl);
            
            websocket.onopen = () => {
                terminal.write('\r\n✅ Connected to Claude session!\r\n');
                
                // Send terminal input to WebSocket
                terminal.onData(data => {
                    if (websocket.readyState === WebSocket.OPEN) {
                        websocket.send(data);
                    }
                });
            };
            
            websocket.onmessage = (event) => {
                terminal.write(event.data);
            };
            
            websocket.onclose = () => {
                terminal.write('\r\n⚠️ Connection closed\r\n');
            };
            
            websocket.onerror = (error) => {
                terminal.write('\r\n❌ Connection error\r\n');
            };
            
            // Focus terminal
            setTimeout(() => terminal.focus(), 100);
        });

        // Listen for messages from parent window (for control buttons)
        window.addEventListener('message', function(event) {
            if (event.data.type === 'sendInput' && websocket && websocket.readyState === WebSocket.OPEN) {
                websocket.send(event.data.data);
                terminal.write(event.data.data);
            }
        });
    </script>
</body>
</html>`

	// Execute template
	t, err := template.New("terminal").Parse(tmpl)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, data)
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

	// Check if branch already exists
	checkCmd := exec.Command("git", "branch", "--list", cleanBranchName)
	checkCmd.Dir = repoPath
	checkOutput, _ := checkCmd.Output()
	
	var cmd *exec.Cmd
	if len(strings.TrimSpace(string(checkOutput))) > 0 {
		// Branch exists, use existing branch
		log.Printf("Branch '%s' already exists, using existing branch", cleanBranchName)
		cmd = exec.Command("git", "worktree", "add", worktreePath, cleanBranchName)
	} else {
		// Branch doesn't exist, create new branch
		cmd = exec.Command("git", "worktree", "add", "-b", cleanBranchName, worktreePath, baseBranch)
	}
	
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
			// Start interactive Claude session
			log.Printf("Starting Claude session for %s", sessionID)
			cmd = exec.Command("claude")
		} else {
			// Fallback to bash shell for testing
			log.Printf("Claude not found, starting bash shell for session %s", sessionID)
			cmd = exec.Command("bash", "-i") // Interactive bash
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

		// Send welcome message and test commands to terminal
		go func() {
			time.Sleep(100 * time.Millisecond) // Give PTY time to initialize
			welcome := fmt.Sprintf("\r\n\033[32m🚀 Claude Manager Session Started\033[0m\r\n")
			welcome += fmt.Sprintf("\033[90mSession: %s\033[0m\r\n", session.Name)
			welcome += fmt.Sprintf("\033[90mDirectory: %s\033[0m\r\n", session.Path)
			welcome += fmt.Sprintf("\033[90mBranch: %s\033[0m\r\n", session.Branch)
			welcome += "\r\n"
			ptySession.PTY.Write([]byte(welcome))
			
			// Send a test command to trigger shell output
			time.Sleep(200 * time.Millisecond)
			if strings.Contains(cmd.Path, "bash") {
				// For bash, send a simple command to get a prompt
				ptySession.PTY.Write([]byte("echo 'Terminal ready. Type commands:'\r\n"))
				time.Sleep(100 * time.Millisecond)
				ptySession.PTY.Write([]byte("pwd\r\n")) // Show current directory
			}
		}()

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
