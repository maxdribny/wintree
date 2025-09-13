# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`wintree` is a cross-platform command-line tool for generating directory tree listings, written in Go. It's designed as a fast, simple alternative to built-in tree commands with advanced filtering capabilities.

## Common Development Commands

### Building
- **Local development build**: `.\build-local.ps1`
  - Builds for Windows with version info from git
  - Installs to `$env:USERPROFILE\bin\wintree.exe`
- **Release build**: `.\build-release.ps1 -Version v1.0.0`
  - Builds for all platforms (Windows, macOS, Linux)
  - Creates zip archives in `dist/` directory
- **Manual build**: `go build .`

### Testing
- **Run all tests**: `go test -v ./...`
- **Run with coverage**: `go test -coverprofile=coverage.out ./...`

### Running
- **Basic usage**: `go run . [path]`
- **With flags**: `go run . --include "*.go" --exclude ".git"`

## Code Architecture

### Core Structure
- **`main.go`**: Entry point that calls `cmd.Execute()`
- **`cmd/root.go`**: Contains all CLI logic using Cobra framework
  - Command definitions and flag handling
  - Core file filtering and tree building logic
  - Smart defaults for different project types

### Key Components

#### File Filtering System (`cmd/root.go`)
- `filter` struct: Holds include/exclude glob patterns
- `findMatchingFiles()`: Walks directory tree applying filters
- `processFilters()`: Expands brace patterns (e.g., `*.{go,js}`)
- Supports depth limiting with `--depth` flag

#### Tree Building (`buildTreeOutput()`)
- Converts file list to hierarchical tree representation
- Uses Unicode box drawing characters (`├──`, `└──`, `│`)
- Maintains proper indentation for nested directories

#### Smart Defaults System
- `detectProjectType()`: Analyzes directory for project indicators
- `getSmartDefaults()`: Returns exclusion patterns for detected project type
- Supports Go, Node.js, Python, Rust, Java, and other common project types

#### Pattern Expansion
- `expandBraces()`: Handles brace expansion like `*.{go,js}` → `["*.go", "*.js"]`
- Uses regex to find and expand brace patterns

### Dependencies
- **Cobra**: CLI framework for command structure and flag handling
- **atotto/clipboard**: System clipboard integration for `--copy` flag

### Testing Structure
All tests are in `cmd/` package:
- `root_test.go`: Unit tests for core functionality
- `integration_test.go`: End-to-end CLI tests
- `benchmark_test.go`: Performance benchmarks

Tests cover file filtering, pattern expansion, tree building, and CLI integration.