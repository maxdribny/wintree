# wintree 🌲

[![GitHub release (latest by date)](https://img.shields.io/github/v/release/maxdribny/wintree)](https://github.com/maxdribny/wintree/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/maxdribny/wintree)](https://goreportcard.com/report/github.com/maxdribny/wintree)
![License](https://img.shields.io/badge/license-MIT-blue.svg)

Tired of using complicated external tree packages with too many options and complicated syntax? Try wintree - a fast, simple and easy to use alternative to built-in and third party tree commands for all operating systems.

> `wintree` is a simple, cross-platform command-line tool for generating directory tree listings. Written in Go, it provides a single, dependency-free binary that runs on Windows, macOS, and Linux. It's designed to be fast, intuitive, and simple, focusing on the most essential features like advanced filtering and easy output redirection.

## Key Features

* **⚡ Fast & Performant:** Built with Go and efficiently skips large, unwanted directories like `.git` and `node_modules`.
* **💻 Cross-Platform:** A single binary for Windows, macOS, and Linux.
* **✅ Simple Syntax:** Intuitive flags that can be used in any order.
* **🔍 Powerful Filtering:**
  * **Exclude:** Easily ignore specific directories (`node_modules`), file extensions (`.log`), or patterns.
  * **Include:** Use glob patterns (`*.go`, `*config*`) to create a whitelist of files to display.
* **📋 Flexible Output:**
  * Print to the console (default).
  * Save the output to a file (`--out`).
  * Copy the output directly to the system clipboard (`--copy`).

## Installation

There are two easy ways to install `wintree`.

### 1. From a Release (Recommended)

This is the easiest way to get `wintree` without needing Go installed.

1. Go to the [**Latest Release**](https://github.com/maxdribny/wintree/releases/latest) page.
2. Download the archive for your operating system and architecture (e.g., `wintree_Windows_x86_64.zip` or `wintree_Darwin_arm64.tar.gz`).
3. Extract the `wintree` (or `wintree.exe`) binary.
4. Move the binary to a location in your system's `PATH` (e.g., `/usr/local/bin` on macOS/Linux or `C:\Windows\System32` on Windows) to make it runnable from anywhere.

### 2. From Source (Requires Go)

If you have Go installed, you can build and install `wintree` with a single command:

```sh
go install [github.com/maxdribny/wintree@latest](https://github.com/maxdribny/wintree@latest)
```

## Usage

```bash
wintree [path] [flags]
```

- `[path]`: The directory to start the tree from. Defaults to the current directory (`.`) if omitted.

### Flags

| Flag               | Shorthand | Description                                                      | Example                   |
| ------------------ | --------- | ---------------------------------------------------------------- | ------------------------- |
| `--exclude <str>`  | `-e`      | Exclude directories or extensions. Can be used multiple times.   | `-e .git -e .log`         |
| `--include <glob>` | `-i`      | Whitelist files using glob patterns. Can be used multiple times. | `-i "*.go" -i "Makefile"` |
| `--out <file>`     | `-o`      | Write the output to the specified file instead of the console.   | `-o my_tree.txt`          |
| `--copy`           | `-c`      | Copy the final output tree to the system clipboard.              | `-c`                      |
| `--smart-defaults` | `-s`      | Apply smart defaults based on detected project type.             | `-s`                      |
| `--show-patterns`  | `-p`      | Show a guide for using glob patterns.                            | `-p`                      |
| `--full-path`      | `-f`      | Show the full directory path above the tree output.              | `-f`                      |
| `--depth <int>`    | `-d`      | Set maximum depth of directory tree (-1 for unlimited).          | `-d 3`                    |
| `--version`        | `-v`      | Show version information.                                        | `-v`                      |
| `--help`           | `-h`      | Show the help message.                                           | `--help`                  |

## Examples

### Basic Usage

Generate a tree from the current directory.

```bash
wintree
```

### Excluding Directories and Extensions

List all files in your project, but ignore the `.git` directory and all `.tmp` files.

```bash
wintree . --exclude .git --exclude .tmp
```

### Including Specific Files (Whitelist)

Only show Go source files and markdown files.

```bash
wintree --include "*.go" --include "*.md"
```

### Combining Include and Exclude

Show all Go files, but ignore test files.

```bash
wintree --include "*.go" --exclude "*_test.go"
```

### Show Full Directory Path

Display the absolute path of the directory being visualized above the tree output.

```bash
wintree --full-path
```

### Control Tree Depth

Limit how deep the tree goes into subdirectories.

```bash
# Show only immediate children.
wintree --depth 1

# Show up to 3 levels deep.
wintree --depth 3

# Show entire tree (unlimited depth).
wintree --depth -1
```

### Saving to a File

Generate a tree of your `src` folder and save it to `docs/directory-structure.txt`.

```bash
wintree ./src --out docs/directory-structure.txt
```

### Copying to Clipboard

Generate a tree of the current directory, excluding `node_modules`, and copy it to the clipboard to easily paste into a document or message.

```bash
wintree -e node_modules -c
```

### Smart Defaults

Apply intelligent filtering based on the detected project type. This automatically excludes common build artifacts, dependency directories, and temporary files.

```bash
wintree --smart-defaults
```

**Supported project types:**

- **Go**: Excludes `vendor`, `*.exe`, `*.dll`, `*.so`, `*.dylib`
- **Node.js/JavaScript**: Excludes `node_modules`, `dist`, `build`, `.next`, `coverage`, `*.log`
- **Python**: Excludes `__pycache__`, `*.pyc`, `venv`, `.env`, `pip-log.txt`
- **Rust**: Excludes `target`, `Cargo.lock`
- **Java**: Excludes `target`, `build`, `*.class`, `*.jar`
- **And many more...**

All project types automatically exclude `.git`, `.DS_Store`, and `Thumbs.db`.

### Show Pattern Help

Get a comprehensive guide on using glob patterns:

```bash
wintree --show-patterns
```

## Building From Source

If you want to contribute to development:

### Clone the repository:

```bash
git clone https://github.com/maxdribny/wintree.git
```

### Navigate to the directory:

```bash
cd wintree
```

### Build the binary:

```bash
go build .
```

## Developers

This section covers the development workflow for contributing to wintree.

### Prerequisites

- Go 1.21 or later
- Git
- PowerShell (windows) or Bash (macOS/Linux)

### Development Workflow

#### 1. Setup

```bash
# Clone the repository
git clone https://github.com/maxdribny/wintree.git
cd wintree

# Create a feature branch
git checkout -b feature/your-feature-name
```

### 2. Make Your Changes

The project structure:

- ```cmd/``` - Core application logic
  - ```root.go``` - Main command implementation
  - ```root_test.go``` - Unit tests
  - ```integration_test.go``` - Integration tests
  - ```benchmark_test.go``` - Performance benchmarks
- ```main.go``` - Entry point

### 3. Run Tests

Always run tests before committing:

```bash
# Run all tests
go test -v ./...


# Run specific test
go test -v -run TestBuildTreeOutput_WithFullPath ./cmd


# Run with coverage (PowerShell)
go test "-coverprofile=coverage.out" ./cmd

# Run with coverage (macOS/Linux)
go test -coverprofile=coverage.out ./cmd
go tool voer -html=coverage.out # View in browser
```

#### 4. Run Benchmarks

Track performance of your changes:

```bash
# Run All Benchmarks - Windows (PowerShell)
./run-benchmarks.ps1

# macOS/Linux
go test -bench=. -benchmem -count=5 ./cmd
```

Compare performance before and after changes in your code:

```bash
# Before changes - Windows (PowerShell)
go test "-bench=." -count=10 ./cmd > benchmarks/before.txt

# macOS/Linux
go test -bench=. -count=10 ./cmd > benchmarks/before.txt

# After changes - Windows (PowerShell)
go test "-bench=." -count=10 ./cmd > benchmarks/after.txt

# macOS/Linux
go test -bench=. -count=10 ./cmd > benchmarks/after.txt


# Install benchstat and compare
go install golang.org/x/perf/cmd/benchstat@latest
benchstat benchmarks/before.txt benchmarks/after.txt
```

### 5. Build and Test Locally

```bash
# Build for local testing
go build .
./wintree --help

# Or use the build script (Windows only)
./build-local.ps1


# Test your changes manually
./wintree --full-path --depth 3
./wintree -f -e node_modules -i "*.go"
```

### 6. Format and Lint

```bash
# Format code
go fmt ./...


# Run go vet
go vet ./...
```

### 7. Commit and Push

```bash
# Stage changes
git add -a

# Commit with descriptive message
git commit -m "feat: add full-path flag to display absolute directory path"

# Push to your fork
git push origin feature/your-feature-name
```

### 8. Create Pull Request

1. Go to [GitHub - maxdribny/wintree](https://github.com/maxdribny/wintree)

2. Click "New Pull Request"

3. "Select your feature branch"

4. Provide a clear description of changes

5. Reference any related issues

## Release Process

For mainters creating a new release:

```bash
# 1. Ensure all tests pass
go test -v ./...

# 2. Update version in code if needed

# 3. Build release artifacts
./build-release.ps1 -Version v[version]


# 4. Create and push tag
git tag v0.4.0
git push origin v0.4.0

# 5. Create github release
# - Upload the .zip files from dist/
# - Add release notes
```

## Benchmark Tracking

We track performance over time. When making changes:

1. Run benchmarks before your changes

2. Make your changes

3. Run benchmarks after

4. Include benchmark comparison in PR if significant changes

Benchmark results are stored in ```benchmarks/``` for historical comparison.

## Contributing

Contributions are welcome! If you find a bug or have a feature request, please open an issue. If you'd like to contribute code, please feel free to fork the repository and open a pull request.

Code Style Guidelines:

- Use meaningful variable names

- Add comments for complex logic

- Keep functions focused and small

- Write tests for new functionality

- Include documentation for user-facing changes in PRs

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
