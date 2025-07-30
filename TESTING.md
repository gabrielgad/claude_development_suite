# Testing Guide - Claude Development Suite

## Overview

This document provides comprehensive testing instructions for the Claude Development Suite web terminal architecture.

## Test Types

### 1. Unit Tests
**Location:** `cm/main_test.go`, `cm/integration_test.go`  
**Purpose:** Test individual functions and components

### 2. Integration Tests  
**Location:** `cm/integration_test.go`  
**Purpose:** Test component interactions and CLI behavior

### 3. Automated Test Suite
**Location:** `cm/test_web_features.sh`  
**Purpose:** Comprehensive automated testing of entire system

## Running Tests

### Quick Test Commands

```bash
# Navigate to project root
cd /home/ggadbois/projects/tooling/claude_development_suite

# 1. Run all Go unit tests
cd cm && go test -v ./...

# 2. Run integration tests only  
cd cm && go test -v -run Integration

# 3. Run automated test suite
cd cm && ./test_web_features.sh

# 4. Build and test binary
cd cm && make build && ./build/cm --version

# 5. Test CW integration
cd cw && ./cw help
```

### Detailed Test Commands

#### Unit Testing
```bash
cd cm

# Run all tests with verbose output
go test -v ./...

# Run specific test functions
go test -v -run TestVersion
go test -v -run TestFlagParsing
go test -v -run TestWebServerStartup

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

#### Integration Testing
```bash
cd cm

# Test binary build process
go test -v -run TestBinaryBuild

# Test binary execution
go test -v -run TestBinaryExecution

# Test web server startup
go test -v -run TestWebServerStartupIntegration

# Test project structure
go test -v -run TestProjectStructure
```

#### Automated Test Suite
```bash
cd cm

# Run full automated test suite
./test_web_features.sh

# The test suite covers:
# - Web Architecture Foundation (6 tests)
# - Build System (5 tests)  
# - CLI Interface (4 tests)
# - Unit Tests (4 tests)
# - CW Integration (4 tests)
# - Documentation (4 tests)
# - Project Structure (4 tests)
# - Integration Verification (3 tests)
```

#### Manual Testing
```bash
# 1. Test binary functionality
cd cm
make build
./build/cm --version
./build/cm --help

# 2. Test web server startup (will run until killed)
./build/cm                    # Default port 8080
./build/cm -port 9000        # Custom port
./build/cm -serve -port 8080 # Explicit serve mode

# 3. Test CW script integration
cd ../cw
./cw help
./cw list
./cw make test-session feature/test main n n "" --no-claude

# 4. Test project structure
ls -la ../                   # Check overall structure
ls -la ../docs/             # Check documentation
ls -la ../cm/backup/        # Check TUI backup
```

## Expected Test Results

### Successful Test Run
```
=== Claude Manager Web Terminal Test Suite ===

1. Testing Web Architecture Foundation
------------------------------------
Testing: Go module exists                         PASSED
Testing: Main.go exists                          PASSED
Testing: Makefile exists                         PASSED
Testing: Web dependencies in go.mod              PASSED
Testing: PTY dependencies in go.mod              PASSED
Testing: Old TUI deps removed                    PASSED

[... all other tests ...]

========================================
Test Results: 34 passed, 0 failed
========================================
All tests passed! Web terminal architecture is ready for Phase 2.
```

### Test Failure Indicators
- **Red "FAILED" messages** in automated test suite
- **Go test failures** with error details
- **Build failures** when running `make build`
- **Missing dependencies** in `go mod verify`

## Troubleshooting

### Common Issues

#### Build Fails
```bash
# Clean and rebuild
make clean
go mod tidy
make build
```

#### Tests Fail Due to Missing Files
```bash
# Ensure you're in the correct directory
cd /home/ggadbois/projects/tooling/claude_development_suite/cm

# Check project structure
ls -la ../
```

#### Go Module Issues
```bash
# Clean module cache
go clean -modcache
go mod download
go mod tidy
```

#### Permission Issues
```bash
# Make scripts executable
chmod +x test_web_features.sh
chmod +x ../cw/cw
```

## Test Coverage

### Current Coverage Areas
- ✅ CLI flag parsing and validation
- ✅ Version information and help output  
- ✅ Build system and binary creation
- ✅ Go module dependencies and structure
- ✅ Project file organization
- ✅ CW script integration
- ✅ Documentation completeness

### Future Coverage Areas (Phase 2)
- 🔄 HTTP server endpoints
- 🔄 WebSocket connection handling
- 🔄 PTY creation and management
- 🔄 Claude process spawning
- 🔄 Session state management
- 🔄 Web frontend functionality

## Continuous Integration

### Pre-commit Testing
```bash
# Run this before every commit
cd cm
./test_web_features.sh && echo "Ready to commit!"
```

### Development Workflow
1. Make changes to code
2. Run unit tests: `go test ./...`
3. Run integration tests: `go test -run Integration`
4. Run full test suite: `./test_web_features.sh`
5. If all pass: commit and push
6. If any fail: fix issues and repeat

## Performance Testing

### Basic Performance Checks
```bash
# Binary size check
ls -lh build/cm

# Startup time check  
time ./build/cm --version

# Memory usage check (when web server implemented)
# ps aux | grep cm
```

### Load Testing (Future)
When web server is implemented:
- WebSocket connection limits
- Concurrent session handling
- PTY resource management
- Memory usage under load

## Security Testing

### Basic Security Checks
```bash
# Check for hardcoded secrets
grep -r "password\|secret\|key" . --exclude-dir=.git

# Check file permissions
find . -type f -executable -ls

# Verify no sensitive data in git
git log --oneline | head -10
```