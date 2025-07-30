class ClaudeManager {
    constructor() {
        this.sessions = new Map();
        this.activeSessions = [];
        this.currentSession = null;
        this.currentTerminal = null;
        this.currentWebSocket = null;
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
            
            if (!branchInput.value || branchInput.placeholder.includes('auto-generated')) {
                branchInput.placeholder = `feature/${cleanName || 'session-name'} (auto-generated)`;
            }
            
            if (e.target.value && cleanName !== e.target.value.toLowerCase()) {
                e.target.style.borderColor = '#ff9800';
                e.target.title = `Will be cleaned to: ${cleanName}`;
            } else {
                e.target.style.borderColor = '#404040';
                e.target.title = '';
            }
        });
        
        document.getElementById('branch-name').addEventListener('input', (e) => {
            const cleanBranch = this.sanitizeForGit(e.target.value);
            
            if (e.target.value && cleanBranch !== e.target.value) {
                e.target.style.borderColor = '#ff9800';
                e.target.title = `Will be cleaned to: ${cleanBranch}`;
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
            
            return `
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
            `;
        }).join('');
    }

    selectSession(sessionId) {
        console.log('Selecting session:', sessionId);
        
        // Update active session UI
        document.querySelectorAll('.session-item').forEach(item => {
            item.classList.remove('active');
        });
        document.querySelector(`[data-session-id="${sessionId}"]`).classList.add('active');
        
        // Open terminal
        this.openTerminal(sessionId);
    }

    openTerminal(sessionId) {
        console.log('Switching to session:', sessionId);
        
        // Find session info
        const session = this.activeSessions.find(s => s.id === sessionId);
        if (!session) {
            console.error('Session not found:', sessionId);
            return;
        }

        // Close current session if different
        if (this.currentSession && this.currentSession !== sessionId) {
            this.disconnectFromCurrentSession();
        }

        // Show terminal area with simple structure
        document.getElementById('welcome-screen').style.display = 'none';
        document.getElementById('terminal-area').style.display = 'flex';
        
        // Update header
        document.getElementById('terminal-title').textContent = `${session.name} - ${session.path}`;
        
        // Use our proven working approach - iframe with the working terminal page
        const terminalDiv = document.getElementById('terminal');
        terminalDiv.innerHTML = `
            <iframe 
                src="/terminal/${sessionId}" 
                style="width: 100%; height: 100%; border: none; background: #1e1e1e;"
                frameborder="0">
            </iframe>
        `;
        
        // Store references (iframe handles its own terminal and websocket)
        this.currentSession = sessionId;
    }

    disconnectFromCurrentSession() {
        // Clear the iframe (it will handle its own cleanup)
        const terminalDiv = document.getElementById('terminal');
        if (terminalDiv) {
            terminalDiv.innerHTML = '';
        }
    }

    showWelcome() {
        this.disconnectFromCurrentSession();
        
        document.getElementById('terminal-area').style.display = 'none';
        document.getElementById('welcome-screen').style.display = 'flex';
        
        this.currentSession = null;
        
        // Clear active session UI
        document.querySelectorAll('.session-item').forEach(item => {
            item.classList.remove('active');
        });
    }

    // Session control methods
    sendFeedback() {
        if (!this.currentSession) return;
        
        const feedback = window.prompt('Enter feedback for this Claude session:');
        if (feedback) {
            // Send message to iframe terminal
            const iframe = document.querySelector('#terminal iframe');
            if (iframe && iframe.contentWindow) {
                iframe.contentWindow.postMessage({
                    type: 'sendInput',
                    data: `[FEEDBACK] ${feedback}\n`
                }, '*');
            }
        }
    }

    continuePrompt() {
        if (!this.currentSession) return;
        
        const userPrompt = window.prompt('Enter additional prompt to continue conversation:');
        if (userPrompt) {
            // Send message to iframe terminal
            const iframe = document.querySelector('#terminal iframe');
            if (iframe && iframe.contentWindow) {
                iframe.contentWindow.postMessage({
                    type: 'sendInput',
                    data: `${userPrompt}\n`
                }, '*');
            }
        }
    }

    pauseSession() {
        if (!this.currentSession) return;
        
        // Send Ctrl+C to iframe terminal
        const iframe = document.querySelector('#terminal iframe');
        if (iframe && iframe.contentWindow) {
            iframe.contentWindow.postMessage({
                type: 'sendInput',
                data: '\x03'
            }, '*');
        }
    }

    closeTerminal() {
        if (this.currentWebSocket) {
            this.currentWebSocket.close();
            this.currentWebSocket = null;
        }
        
        document.getElementById('terminal-area').style.display = 'none';
        document.getElementById('welcome-screen').style.display = 'flex';
        
        this.currentSession = null;
        this.currentTerminal = null;
        
        document.querySelectorAll('.session-item').forEach(item => {
            item.classList.remove('active');
        });
    }

    connectWebSocket(sessionId, terminal) {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws/${sessionId}`;
        
        const ws = new WebSocket(wsUrl);
        this.currentWebSocket = ws;
        
        ws.onopen = () => {
            terminal.write('\r\n✅ Connected!\r\n');
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
            terminal.write('\r\n⚠️ Disconnected\r\n');
        };

        ws.onerror = (error) => {
            terminal.write('\r\n❌ Error\r\n');
        };
    }

    showSessionCreator() {
        document.getElementById('session-creator').style.display = 'flex';
        document.getElementById('session-name').focus();
        this.loadDirectories();
    }

    hideSessionCreator() {
        document.getElementById('session-creator').style.display = 'none';
        document.getElementById('session-name').value = '';
        document.getElementById('branch-name').value = '';
        document.getElementById('base-branch').value = '';
        document.getElementById('use-worktree').checked = true;
        this.selectedRepoPath = '';
        document.getElementById('selected-repo-path').textContent = 'None selected';
    }

    async loadDirectories(path = '') {
        try {
            const url = path ? `/api/directories?path=${encodeURIComponent(path)}` : '/api/directories';
            const response = await fetch(url);
            const data = await response.json();
            
            this.currentPath = data.currentPath;
            this.renderDirectories(data.items);
            document.getElementById('current-path-text').textContent = this.currentPath;
        } catch (error) {
            console.error('Failed to load directories:', error);
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
            
            return `
                <div class="${classes.join(' ')}" data-path="${item.path}" onclick="app.handleDirectoryClick('${item.path}', ${item.isGitRepo})">
                    <span class="directory-icon">${icon}</span>
                    <span class="directory-name">${item.name}</span>
                    ${gitIndicator}
                </div>
            `;
        }).join('');
    }

    handleDirectoryClick(path, isGitRepo) {
        if (isGitRepo) {
            this.selectedRepoPath = path;
            document.getElementById('selected-repo-path').textContent = path;
            
            document.querySelectorAll('.directory-item').forEach(item => {
                item.classList.remove('selected');
            });
            document.querySelector(`[data-path="${path}"]`).classList.add('selected');
        } else {
            this.loadDirectories(path);
        }
    }

    async navigateToHome() {
        this.loadDirectories();
    }

    async navigateToRoot() {
        this.loadDirectories('/');
    }

    sanitizeForGit(input) {
        if (!input) return 'unnamed';
        
        let result = input.toLowerCase();
        result = result.replace(/[^a-z0-9\-_./]/g, '-');
        result = result.replace(/-+/g, '-');
        result = result.replace(/^[-.]|[-.]$/g, '');
        
        if (!result) {
            result = 'branch-' + Date.now();
        }
        
        if (result.length > 50) {
            result = result.substring(0, 50).replace(/-+$/, '');
        }
        
        return result;
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
            alert('Please select a repository');
            return;
        }

        try {
            const response = await fetch('/api/sessions/create', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    name,
                    repoPath,
                    branchName: branchName || `feature/${name}`,
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

    async killSession(sessionId, event) {
        event.stopPropagation();
        
        if (!confirm('Kill this session?')) {
            return;
        }

        try {
            const response = await fetch('/api/sessions/kill', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ sessionId })
            });

            if (response.ok) {
                await this.loadSessions();
                
                if (this.currentSession === sessionId) {
                    this.closeTerminal();
                }
            } else {
                alert('Failed to kill session');
            }
        } catch (error) {
            console.error('Failed to kill session:', error);
            alert('Failed to kill session');
        }
    }

    startSessionPolling() {
        setInterval(() => {
            this.loadSessions();
        }, 5000);
    }
}

// Initialize app
const app = new ClaudeManager();