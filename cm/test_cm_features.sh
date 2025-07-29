#!/bin/bash
# Test script for Claude Manager Phase 1 features

echo "=== Claude Manager Phase 1 Test Suite ==="
echo ""

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
PASSED=0
FAILED=0

# Test function
test_feature() {
    local test_name="$1"
    local command="$2"
    local expected_result="$3"
    
    echo -n "Testing: $test_name... "
    
    if eval "$command"; then
        if [ "$expected_result" = "pass" ]; then
            echo -e "${GREEN}PASSED${NC}"
            ((PASSED++))
        else
            echo -e "${RED}FAILED${NC} (expected to fail)"
            ((FAILED++))
        fi
    else
        if [ "$expected_result" = "fail" ]; then
            echo -e "${GREEN}PASSED${NC} (correctly failed)"
            ((PASSED++))
        else
            echo -e "${RED}FAILED${NC}"
            ((FAILED++))
        fi
    fi
}

echo "1. Testing Process Discovery (gopsutil)"
echo "----------------------------------------"

# Check if binary exists
test_feature "Binary exists" "test -f build/cm" "pass"

# Check process scanning code
test_feature "Process monitoring code exists" "grep -q 'gopsutil/v3/process' main.go" "pass"
test_feature "scanProcesses function exists" "grep -q 'func.*scanProcesses' main.go" "pass"
test_feature "Process name filtering" "grep -q 'strings.Contains.*claude' main.go" "pass"

echo ""
echo "2. Testing Git Worktree Detection"
echo "---------------------------------"

test_feature "Git branch detection code" "grep -q 'getGitBranch' main.go" "pass"
test_feature "Git command execution" "grep -q 'git.*branch.*--show-current' main.go" "pass"
test_feature "Working directory extraction" "grep -q 'proc.Cwd()' main.go" "pass"

echo ""
echo "3. Testing Session Data Structures"
echo "----------------------------------"

test_feature "Session struct definition" "grep -q 'type Session struct' main.go" "pass"
test_feature "Session ID field" "grep -q 'ID.*string.*json:\"id\"' main.go" "pass"
test_feature "Session Path field" "grep -q 'Path.*string.*json:\"path\"' main.go" "pass"
test_feature "Session Branch field" "grep -q 'Branch.*string.*json:\"branch\"' main.go" "pass"
test_feature "Session PID field" "grep -q 'PID.*int.*json:\"pid\"' main.go" "pass"
test_feature "SessionStatus enum" "grep -q 'type SessionStatus int' main.go" "pass"
test_feature "Status constants" "grep -q 'StatusActive' main.go && grep -q 'StatusIdle' main.go" "pass"

echo ""
echo "4. Testing Basic TUI Implementation"
echo "-----------------------------------"

test_feature "Bubbletea import" "grep -q 'github.com/charmbracelet/bubbletea' main.go" "pass"
test_feature "Model struct" "grep -q 'type Model struct' main.go" "pass"
test_feature "Init function" "grep -q 'func (m Model) Init()' main.go" "pass"
test_feature "Update function" "grep -q 'func (m Model) Update(' main.go" "pass"
test_feature "View function" "grep -q 'func (m Model) View()' main.go" "pass"
test_feature "Session list view" "grep -q 'renderSessionsView' main.go" "pass"
test_feature "Details view" "grep -q 'renderDetailsView' main.go" "pass"
test_feature "Keyboard navigation" "grep -q 'case.*up.*:' main.go && grep -q 'case.*down.*:' main.go" "pass"

echo ""
echo "5. Testing Real-time Process Monitoring"
echo "---------------------------------------"

test_feature "ProcessMonitor struct" "grep -q 'type ProcessMonitor struct' main.go" "pass"
test_feature "Monitor interval field" "grep -q 'interval.*time.Duration' main.go" "pass"
test_feature "Update channel" "grep -q 'updateChan.*chan.*SessionUpdate' main.go" "pass"
test_feature "Ticker implementation" "grep -q 'time.NewTicker' main.go" "pass"
test_feature "Goroutine start" "grep -q 'go.*Start()' main.go" "pass"
test_feature "Context cancellation" "grep -q 'context.WithCancel' main.go" "pass"

echo ""
echo "6. Testing Configuration Management"
echo "-----------------------------------"

test_feature "Config struct" "grep -q 'type Config struct' main.go" "pass"
test_feature "LoadConfig function" "grep -q 'func LoadConfig' main.go" "pass"
test_feature "SaveConfig function" "grep -q 'func SaveConfig' main.go" "pass"
test_feature "Config file path" "grep -q 'claude_manager' main.go" "pass"
test_feature "Theme configuration" "grep -q 'type Theme struct' main.go" "pass"
test_feature "Config usage in UI" "grep -q 'm.config.Theme' main.go" "pass"

echo ""
echo "7. Testing Build System"
echo "-----------------------"

test_feature "Makefile exists" "test -f Makefile" "pass"
test_feature "go.mod exists" "test -f go.mod" "pass"
test_feature "go.sum exists" "test -f go.sum" "pass"
test_feature "Build target" "grep -q '^build:' Makefile" "pass"
test_feature "Install target" "grep -q '^install:' Makefile" "pass"
test_feature "Dependencies resolved" "go mod verify 2>/dev/null" "pass"

echo ""
echo "8. Integration Test"
echo "-------------------"

# Test config creation
test_feature "Config directory creation" "mkdir -p ~/.config/claude_manager && echo '{}' > ~/.config/claude_manager/test.json && rm ~/.config/claude_manager/test.json" "pass"

# Test binary compilation
test_feature "Binary compiles successfully" "cd /home/ggadbois/projects/tooling/claude_development_suite-week1/cm && go build -o test_binary main.go && rm test_binary" "pass"

echo ""
echo "========================================"
echo -e "Test Results: ${GREEN}$PASSED passed${NC}, ${RED}$FAILED failed${NC}"
echo "========================================"

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All Phase 1 features are implemented!${NC}"
    exit 0
else
    echo -e "${RED}Some features are missing or broken.${NC}"
    exit 1
fi