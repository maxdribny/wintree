**Project:** `wintree`
**Author:** Maxim Dribny
**Language:** Go
**Description:** `wintree` is a command-line tool for generating directory tree listings. It is designed to be a fast, simple, and cross-platform alternative to the built-in `tree` command on various operating systems.
**Key Features:**
*   **Cross-Platform:** Single binary for Windows, macOS, and Linux.
*   **Fast and Performant:** Efficiently skips large directories like `.git` and `node_modules`.
*   **Powerful Filtering:** Supports excluding and including files and directories using glob patterns.
*   **Flexible Output:** Can print to the console, save to a file, or copy to the clipboard.
*   **Smart Defaults:** Automatically applies common ignore patterns based on the detected project type (e.g., Go, Node.js, Python).
**Dependencies:**
*   `github.com/spf13/cobra` for command-line argument parsing.
*   `github.com/atotto/clipboard` for clipboard functionality.
