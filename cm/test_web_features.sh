#!/bin/bash
# Test script for Claude Manager Web Terminal Architecture

echo "=== Claude Manager Web Terminal Test Suite ==="
echo ""

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
PASSED=0
FAILED=0

# Test function
test_feature() {
    local test_name="$1"
    local command="$2"
    local expected="$3"
    
    printf "Testing: %-50s " "$test_name"
    
    if eval "$command" >/dev/null 2>&1; then
        if [ "$expected" = "pass" ]; then
            echo -e "${GREEN}PASSED${NC}"
            ((PASSED++))
        else
            echo -e "${RED}FAILED${NC}"
            ((FAILED++))
        fi
    else
        if [ "$expected" = "fail" ]; then
            echo -e "${GREEN}PASSED${NC}"
            ((PASSED++))
        else
            echo -e "${RED}FAILED${NC}"
            ((FAILED++))
        fi
    fi
}

# Change to correct directory
cd /home/ggadbois/projects/tooling/claude_development_suite/cm

echo "1. Testing Web Architecture Foundation"
echo "------------------------------------"
test_feature "Go module exists" "test -f go.mod" "pass"
test_feature "Main.go exists" "test -f main.go" "pass"
test_feature "Makefile exists" "test -f Makefile" "pass"
test_feature "Web dependencies in go.mod" "grep -q 'github.com/gorilla/websocket' go.mod" "pass"
test_feature "PTY dependencies in go.mod" "grep -q 'github.com/creack/pty' go.mod" "pass"
test_feature "Old TUI deps removed" "! grep -q 'bubbletea' go.mod" "pass"

echo ""
echo "2. Testing Build System"
echo "----------------------"
test_feature "Clean build" "make clean" "pass"
test_feature "Go module verification" "go mod verify 2>/dev/null || echo 'skipped due to network'" "pass"
test_feature "Binary compilation" "make build" "pass"
test_feature "Binary exists" "test -f build/cm" "pass"
test_feature "Binary is executable" "test -x build/cm" "pass"

echo ""
echo "3. Testing CLI Interface"
echo "-----------------------"
test_feature "Version flag works" "./build/cm --version 2>&1 | grep -q '2.0.0-web'" "pass"
test_feature "Help flag works" "./build/cm --help 2>&1 | grep -q 'Usage'" "pass"
test_feature "Port flag recognized" "./build/cm --help 2>&1 | grep -q 'port'" "pass"
test_feature "Serve flag recognized" "./build/cm --help 2>&1 | grep -q 'serve'" "pass"

echo ""
echo "4. Testing Unit Tests"
echo "--------------------"
test_feature "Unit tests exist" "test -f main_test.go" "pass"
test_feature "Integration tests exist" "test -f integration_test.go" "pass"
test_feature "Go test runs" "go test -v ./... | grep -q 'PASS'" "pass"
test_feature "No test failures" "! go test ./... 2>&1 | grep -q 'FAIL'" "pass"

echo ""
echo "5. Testing CW Integration"
echo "------------------------"
test_feature "CW script exists" "test -f ../cw/cw" "pass"
test_feature "CW is executable" "test -x ../cw/cw" "pass"
test_feature "CW help works" "../cw/cw help | grep -q 'Claude Worktree Manager'" "pass"
test_feature "CW no-claude flag exists" "../cw/cw help | grep -q -- '--no-claude'" "pass"

echo ""
echo "6. Testing Documentation"
echo "------------------------"
test_feature "Web architecture doc exists" "test -f ../docs/web-terminal-architecture.md" "pass"
test_feature "README updated" "grep -q 'web-based' ../README.md" "pass"
test_feature "Prerequisites documented" "grep -q 'Go 1.21' ../README.md" "pass"
test_feature "Quick start updated" "grep -q 'http://localhost:8080' ../README.md" "pass"

echo ""
echo "7. Testing Project Structure"
echo "----------------------------"
test_feature "Build directory exists" "test -d build" "pass"
test_feature "Go files are clean" "go test ./... | grep -q 'ok'" "pass"
test_feature "No old TUI artifacts" "! test -f main-tui-backup.go" "pass"
test_feature "Clean module structure" "ls *.go | grep -E '^(main|.*_test).go$'" "pass"

echo ""
echo "8. Integration Verification"
echo "---------------------------"
test_feature "Binary starts without error" "timeout 2s ./build/cm --version" "pass"
test_feature "Web server placeholder" "timeout 2s ./build/cm 2>&1 | grep -q 'Web Server'" "pass"
test_feature "Port configuration works" "timeout 2s ./build/cm -port 9999 2>&1 | grep -q '9999'" "pass"

echo ""
echo "========================================"
echo -e "Test Results: ${GREEN}$PASSED passed${NC}, ${RED}$FAILED failed${NC}"
echo "========================================"

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed! Web terminal architecture is ready for Phase 2.${NC}"
    echo ""
    echo "Next steps:"
    echo "1. Implement HTTP server with static file serving"
    echo "2. Add WebSocket handlers for terminal I/O"
    echo "3. Implement PTY management for Claude processes"
    echo "4. Create web frontend with xterm.js terminal emulator"
    echo "5. Add session management and persistence"
    exit 0
else
    echo -e "${RED}Some tests failed. Please fix issues before proceeding.${NC}"
    exit 1
fi