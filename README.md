# wintree 🌲

> Tired of using complicated external tree packages with too many options and complicated syntax? Try wintree - a fast, simple and easy to use alternative to built-in and third party tree commands for all operating systems.

[![GitHub release (latest by date)](https://img.shields.io/github/v/release/maxdribny/wintree)](https://github.com/maxdribny/wintree/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/maxdribny/wintree)](https://goreportcard.com/report/github.com/maxdribny/wintree)
![License](https://img.shields.io/badge/license-MIT-blue.svg)

`wintree` is a modern, cross-platform command-line tool for generating directory tree listings. Written in Go, it provides a single, dependency-free binary that runs on Windows, macOS, and Linux. It's designed to be fast, intuitive, and powerful, focusing on the most essential features like advanced filtering and easy output redirection.

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

1.  Go to the [**Latest Release**](https://github.com/maxdribny/wintree/releases/latest) page.
2.  Download the archive for your operating system and architecture (e.g., `wintree_Windows_x86_64.zip` or `wintree_Darwin_arm64.tar.gz`).
3.  Extract the `wintree` (or `wintree.exe`) binary.
4.  Move the binary to a location in your system's `PATH` (e.g., `/usr/local/bin` on macOS/Linux or `C:\Windows\System32` on Windows) to make it runnable from anywhere.

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

| Flag              | Shorthand | Description                                                       | Example                                 |
|-------------------|-----------|-------------------------------------------------------------------|-----------------------------------------|
| `--exclude <str>` | `-e`      | Exclude directories or extensions. Can be used multiple times.    | `-e .git -e .log`                        |
| `--include <glob>`| `-i`      | Whitelist files using glob patterns. Can be used multiple times.  | `-i "*.go" -i "Makefile"`               |
| `--out <file>`    | `-o`      | Write the output to the specified file instead of the console.    | `-o my_tree.txt`                        |
| `--copy`          | `-c`      | Copy the final output tree to the system clipboard.               | `-c`                                    |
| `--help`          | `-h`      | Show the help message.                                            | `--help`                                |

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

## Contributing

Contributions are welcome! If you find a bug or have a feature request, please open an issue. If you'd like to contribute code, please feel free to fork the repository and open a pull request.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

