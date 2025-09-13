# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Project Overview

`wintree` is a cross-platform command-line directory tree generator written in Go. It's designed as a simple, fast alternative to the built-in `tree` command with advanced filtering capabilities via glob patterns.

## Architecture

### Core Components

1. **Command Structure** (`cmd/`)
   - `root.go`: Main command implementation using Cobra framework
   - Pattern expansion and filtering logic (brace expansion, glob matching)
   - Tree output generation with Unicode box-drawing characters
   - Smart defaults detection for different project types

2. **Entry Point** (`main.go`)
   - Simple entry point that delegates to `cmd.Execute()`

3. **Dependencies** (via `go.mod`)
   - `github.com/spf13/cobra`: Command-line argument parsing
   - `github.com/atotto/clipboard`: Cross-platform clipboard support

### Key Functions

- `expandBraces()`: Expands patterns like `*.{go,js}` into multiple patterns
- `processFilters()`: Processes include/exclude patterns with brace expansion
- `findMatchingFiles()`: Walks directory tree applying filters and depth limits
- `buildTreeOutput()`: Generates the tree visualization from matched paths
- `applySmartDefaults()`: Detects project type and applies appropriate exclusions

## Development Commands

### Building

```bash
# Build the binary
go build .

# Build with version information
go build -ldflags "-X github.com/maxdribny/wintree/cmd.Version=v1.0.0 -X github.com/maxdribny/wintree/cmd.Commit=$(git rev-parse HEAD) -X github.com/maxdribny/wintree/cmd.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" .

# Install locally
go install .
```

### Testing

```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# View coverage report in browser
go tool cover -html=coverage.out -o coverage.html

# Run specific test
go test -v -run TestExpandBraces ./cmd

# Run benchmarks
go test -bench=. ./cmd

# Run integration tests
go test -v -run TestFindMatchingFiles ./cmd
```

### Linting & Formatting

```bash
# Install golangci-lint if not present
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run --timeout=5m

# Format code
go fmt ./...

# Tidy dependencies
go mod tidy

# Verify module dependencies
go mod verify
```

### Security Scanning

```bash
# Install security tools
go install github.com/securego/gosec/v2/cmd/gosec@latest
go install golang.org/x/vuln/cmd/govulncheck@latest

# Run gosec security scanner
gosec ./...

# Check for vulnerable dependencies
govulncheck ./...
```

## CI/CD Pipeline

The project uses GitHub Actions for CI/CD with two workflows:

1. **Security Scan & CI** (`security-scan.yml`)
   - Runs on every push and PR to main
   - Security vulnerability scanning (gosec, govulncheck, Nancy)
   - Code quality checks (golangci-lint)
   - Test coverage with Codecov integration
   - Multi-platform build matrix (Ubuntu, Windows, macOS)

2. **CodeQL Analysis** (`codeql.yml`)
   - Weekly scheduled scans
   - Advanced security and quality analysis

## Testing Strategy

### Unit Tests (`cmd/root_test.go`)
- Pattern expansion logic
- Filter processing
- Tree output generation

### Benchmark Tests (`cmd/benchmark_test.go`)
- Performance testing for critical functions
- File system operations benchmarking

### Integration Tests (`cmd/integration_test.go`)
- Complex directory structure scenarios
- Include/exclude pattern combinations
- End-to-end workflow testing

## Common Development Tasks

### Adding a New Feature

1. Implement feature in `cmd/root.go`
2. Add corresponding tests in `cmd/root_test.go`
3. Update README.md with usage examples
4. Run tests and linter before committing

### Debugging

```bash
# Run with verbose output
go run . --help

# Test specific patterns
go run . --include "*.go" --exclude "*_test.go"

# Test clipboard functionality
go run . --copy

# Test smart defaults
go run . --smart-defaults
```

### Release Process

The project uses GitHub releases. To create a new release:

1. Update version in code if version constants are used
2. Create and push a git tag
3. GitHub Actions will build cross-platform binaries
4. Upload binaries to the release

## Project-Specific Patterns

### Error Handling
- Use `fmt.Errorf()` with wrapped errors for context
- Return errors from functions rather than panicking
- Check all file system operations for errors

### File System Operations
- Always use `filepath` package for path operations (cross-platform)
- Handle both Unix and Windows path separators
- Use `os.MkdirAll()` for creating nested directories

### Testing Patterns
- Create temporary directories with `os.MkdirTemp()` for tests
- Always defer cleanup with `os.RemoveAll()`
- Test both positive and negative cases
- Use table-driven tests for multiple scenarios

## Performance Considerations

- Efficient directory walking with `filepath.WalkDir`
- Early termination with `fs.SkipDir` for excluded directories
- Minimal memory allocation in hot paths
- Pre-compiled glob patterns for repeated matching

## Known Limitations

- Maximum depth flag defaults to 1 (configurable with `-d`)
- Clipboard functionality requires system clipboard access
- Large directories may take time to process without feedback