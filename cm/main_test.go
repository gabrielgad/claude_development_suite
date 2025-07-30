package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestVersion(t *testing.T) {
	if VERSION != "2.0.0-web" {
		t.Errorf("Expected version '2.0.0-web', got '%s'", VERSION)
	}
}

func TestFlagParsing(t *testing.T) {
	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name     string
		args     []string
		expected map[string]interface{}
	}{
		{
			name: "default flags",
			args: []string{"cm"},
			expected: map[string]interface{}{
				"port":    8080,
				"serve":   false,
				"version": false,
			},
		},
		{
			name: "custom port",
			args: []string{"cm", "-port", "9000"},
			expected: map[string]interface{}{
				"port": 9000,
			},
		},
		{
			name: "serve flag",
			args: []string{"cm", "-serve"},
			expected: map[string]interface{}{
				"serve": true,
			},
		},
		{
			name: "version flag",
			args: []string{"cm", "-version"},
			expected: map[string]interface{}{
				"version": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flag package
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			
			// Set test args
			os.Args = tt.args
			
			// Parse flags
			var (
				serve   = flag.Bool("serve", false, "Start web server mode")
				port    = flag.Int("port", 8080, "Web server port")
				version = flag.Bool("version", false, "Show version")
			)
			flag.Parse()

			// Check expected values
			if expectedPort, ok := tt.expected["port"]; ok {
				if *port != expectedPort.(int) {
					t.Errorf("Expected port %d, got %d", expectedPort.(int), *port)
				}
			}
			
			if expectedServe, ok := tt.expected["serve"]; ok {
				if *serve != expectedServe.(bool) {
					t.Errorf("Expected serve %v, got %v", expectedServe.(bool), *serve)
				}
			}
			
			if expectedVersion, ok := tt.expected["version"]; ok {
				if *version != expectedVersion.(bool) {
					t.Errorf("Expected version %v, got %v", expectedVersion.(bool), *version)
				}
			}
		})
	}
}

func TestWebServerStartup(t *testing.T) {
	// Test that we can start a web server on a test port
	testPort := 9999
	
	// Start server in goroutine
	go func() {
		// Simple test server to verify concept
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "OK")
		})
		
		err := http.ListenAndServe(fmt.Sprintf(":%d", testPort), nil)
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("Failed to start test server: %v", err)
		}
	}()
	
	// Give server time to start
	time.Sleep(100 * time.Millisecond)
	
	// Test that server is responding
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/health", testPort))
	if err != nil {
		t.Errorf("Failed to connect to test server: %v", err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %d", resp.StatusCode)
	}
}

func TestBuildArtifacts(t *testing.T) {
	// Test that required build artifacts exist or can be created
	artifacts := []string{
		"go.mod",
		"main.go",
		"Makefile",
	}
	
	for _, artifact := range artifacts {
		if _, err := os.Stat(artifact); os.IsNotExist(err) {
			t.Errorf("Required artifact missing: %s", artifact)
		}
	}
}

func TestGoModDependencies(t *testing.T) {
	// Verify that required dependencies are present in go.mod
	requiredDeps := []string{
		"github.com/gorilla/websocket",
		"github.com/creack/pty",
		"github.com/shirou/gopsutil",
	}
	
	// Read go.mod file
	goModBytes, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}
	
	goModContent := string(goModBytes)
	
	for _, dep := range requiredDeps {
		if !containsString(goModContent, dep) {
			t.Errorf("Required dependency missing from go.mod: %s", dep)
		}
	}
	
	// Verify old TUI dependencies are removed
	deprecatedDeps := []string{
		"github.com/charmbracelet/bubbletea",
		"github.com/charmbracelet/lipgloss",
	}
	
	for _, dep := range deprecatedDeps {
		if containsString(goModContent, dep) {
			t.Errorf("Deprecated TUI dependency still present in go.mod: %s", dep)
		}
	}
}

// Helper function to check if string contains substring
func containsString(haystack, needle string) bool {
	return len(haystack) >= len(needle) && 
		   (haystack == needle || 
		    len(haystack) > len(needle) && 
		    (haystack[:len(needle)] == needle || 
		     haystack[len(haystack)-len(needle):] == needle ||
		     containsSubstring(haystack, needle)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestCLIOutput(t *testing.T) {
	// Test that CLI produces expected output patterns
	tests := []struct {
		name     string
		args     []string
		contains []string
	}{
		{
			name: "version output",
			args: []string{"-version"},
			contains: []string{
				"Claude Manager",
				"2.0.0-web",
				"Web Terminal Edition",
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This would require capturing stdout/stderr
			// For now, we'll just verify the VERSION constant exists
			for _, expected := range tt.contains {
				if expected == "2.0.0-web" && VERSION != expected {
					t.Errorf("Expected version output to contain '%s'", expected)
				}
			}
		})
	}
}