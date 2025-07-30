package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBinaryBuild(t *testing.T) {
	// Test that the binary can be built successfully
	cmd := exec.Command("make", "clean")
	if err := cmd.Run(); err != nil {
		t.Errorf("Failed to clean build: %v", err)
	}
	
	cmd = exec.Command("make", "build")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Failed to build binary: %v\nOutput: %s", err, string(output))
	}
	
	// Verify binary exists
	binaryPath := filepath.Join("build", "cm")
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Errorf("Binary not created at expected path: %s", binaryPath)
	}
	
	// Verify binary is executable
	info, err := os.Stat(binaryPath)
	if err != nil {
		t.Errorf("Cannot stat binary: %v", err)
	}
	
	mode := info.Mode()
	if mode&0111 == 0 {
		t.Errorf("Binary is not executable: %v", mode)
	}
}

func TestBinaryExecution(t *testing.T) {
	// Ensure binary is built
	binaryPath := filepath.Join("build", "cm")
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skip("Binary not found, skipping execution test")
	}
	
	tests := []struct {
		name     string
		args     []string
		expectSuccess bool
		contains []string
		timeout  time.Duration
	}{
		{
			name:          "version flag",
			args:          []string{"--version"},
			expectSuccess: true,
			contains:      []string{"Claude Manager", "2.0.0-web", "Web Terminal Edition"},
			timeout:       5 * time.Second,
		},
		{
			name:          "help flag",
			args:          []string{"--help"},
			expectSuccess: true,
			contains:      []string{"Usage", "port", "serve", "version"},
			timeout:       5 * time.Second,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			
			// Set timeout
			done := make(chan error, 1)
			var output []byte
			
			go func() {
				var err error
				output, err = cmd.CombinedOutput()
				done <- err
			}()
			
			select {
			case err := <-done:
				if tt.expectSuccess && err != nil {
					t.Errorf("Command failed unexpectedly: %v\nOutput: %s", err, string(output))
				}
				if !tt.expectSuccess && err == nil {
					t.Errorf("Command succeeded when it should have failed")
				}
				
				outputStr := string(output)
				for _, expected := range tt.contains {
					if !strings.Contains(outputStr, expected) {
						t.Errorf("Output missing expected text '%s'\nFull output: %s", expected, outputStr)
					}
				}
				
			case <-time.After(tt.timeout):
				cmd.Process.Kill()
				t.Errorf("Command timed out after %v", tt.timeout)
			}
		})
	}
}

func TestWebServerStartupIntegration(t *testing.T) {
	// Test that the binary can start a web server (but don't let it run forever)
	binaryPath := filepath.Join("build", "cm")
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skip("Binary not found, skipping web server test")
	}
	
	// Start the server with a custom port
	testPort := "9998"
	cmd := exec.Command(binaryPath, "-port", testPort)
	
	if err := cmd.Start(); err != nil {
		t.Errorf("Failed to start web server: %v", err)
		return
	}
	
	// Make sure to clean up
	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()
	
	// Give it a moment to start
	time.Sleep(500 * time.Millisecond)
	
	// For now, just verify the process started
	// In the future, we could make HTTP requests to test endpoints
	if cmd.Process == nil {
		t.Error("Process should be running")
	}
	
	// Verify process is still running
	if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
		t.Error("Process exited unexpectedly")
	}
}

func TestCWIntegration(t *testing.T) {
	// Test that CW script exists and is executable
	cwPath := filepath.Join("..", "cw", "cw")
	
	if _, err := os.Stat(cwPath); os.IsNotExist(err) {
		t.Errorf("CW script not found at: %s", cwPath)
		return
	}
	
	// Test CW help command
	cmd := exec.Command(cwPath, "help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("CW help command failed: %v\nOutput: %s", err, string(output))
		return
	}
	
	outputStr := string(output)
	expectedStrings := []string{
		"cw - Claude Worktree Manager",
		"USAGE:",
		"make",
		"list",
		"help",
		"--no-claude",
	}
	
	for _, expected := range expectedStrings {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("CW help output missing expected text '%s'\nFull output: %s", expected, outputStr)
		}
	}
}

func TestProjectStructure(t *testing.T) {
	// Test that the project has the expected structure
	expectedFiles := []string{
		"README.md",
		"cm/main.go",
		"cm/go.mod",  
		"cm/Makefile",
		"cw/cw",
		"cw/cw.fish",
		"docs/web-terminal-architecture.md",
	}
	
	projectRoot := ".."
	
	for _, file := range expectedFiles {
		fullPath := filepath.Join(projectRoot, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Expected project file missing: %s", file)
		}
	}
}

func TestMakefileTargets(t *testing.T) {
	// Test that Makefile has expected targets
	makefileContent, err := os.ReadFile("Makefile")
	if err != nil {
		t.Fatalf("Failed to read Makefile: %v", err)
	}
	
	content := string(makefileContent)
	expectedTargets := []string{
		"build:",
		"clean:",
		"install:",
		"test:",
	}
	
	for _, target := range expectedTargets {
		if !strings.Contains(content, target) {
			t.Errorf("Makefile missing expected target: %s", target)
		}
	}
}