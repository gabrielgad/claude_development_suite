package main

import (
	"context"
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

	// Verify binary exists (correct name: claude-manager)
	binaryPath := filepath.Join("build", "claude-manager")
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Errorf("Binary not created at expected path: %s", binaryPath)
	}

	// Verify binary is executable
	info, err := os.Stat(binaryPath)
	if err != nil {
		t.Errorf("Cannot stat binary: %v", err)
		return
	}

	mode := info.Mode()
	if mode&0111 == 0 {
		t.Errorf("Binary is not executable: %v", mode)
	}
}

func TestBinaryExecution(t *testing.T) {
	// Ensure binary is built
	binaryPath := filepath.Join("build", "claude-manager")
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skip("Binary not found, skipping execution test")
	}

	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "version flag",
			args: []string{"--version"},
			want: "2.0.0-web",
		},
		{
			name: "help flag",
			args: []string{"--help"},
			want: "Usage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := contextWithTimeout(t, 5*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, binaryPath, tt.args...)
			output, err := cmd.CombinedOutput()
			
			if err != nil && ctx.Err() == nil {
				// Only fail if it's not a timeout
				t.Errorf("Command failed: %v\nOutput: %s", err, string(output))
				return
			}

			if !strings.Contains(string(output), tt.want) {
				t.Errorf("Expected output to contain %q, got: %s", tt.want, string(output))
			}
		})
	}
}

func TestModuleStructure(t *testing.T) {
	// Test that go.mod exists and has correct dependencies
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		t.Error("go.mod file missing")
	}

	// Read go.mod
	content, err := os.ReadFile("go.mod")
	if err != nil {
		t.Errorf("Failed to read go.mod: %v", err)
		return
	}

	goMod := string(content)
	
	// Check for required dependencies
	requiredDeps := []string{
		"github.com/gorilla/websocket",
		"github.com/creack/pty",
	}

	for _, dep := range requiredDeps {
		if !strings.Contains(goMod, dep) {
			t.Errorf("Missing required dependency: %s", dep)
		}
	}
}

// Helper function for context with timeout
func contextWithTimeout(t *testing.T, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}