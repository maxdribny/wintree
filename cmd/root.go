/*
Copyright © 2025 Maxim Dribny <mdribnyi@gmail.com>
*/
package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"

	"regexp"
)

// Version information
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

var (
	excludePatterns  []string
	includePatterns  []string
	outputFile       string
	copyToClipboard  bool
	showPatterns     bool
	showVersion      bool
	useSmartDefaults bool
	maxDepth         int
)

type filter struct {
	excludeGlobs []string
	includeGlobs []string
}

var rootCmd = &cobra.Command{
	Use:   "wintree [path]",
	Short: "A modern, cross-platform tree command.",
	Long: `wintree is a simple, intuitive, and easy-to-use alternative to the
built-in tree commands on Windows and other operating systems.

It allows for advanced filtering with inclusion and exclusion patterns
and can output to the terminal, a file, or the system clipboard.`,
	Args: cobra.MaximumNArgs(1), // We expect at most one argument: the path.
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if the user wants version info
		if showVersion {
			printVersionInfo()
			return nil
		}

		// Check if the user wants pattern help
		if showPatterns {
			printPatternHelp()
			return nil
		}

		// 1. Setup - Find Start Path
		startPath := "."
		if len(args) > 0 {
			startPath = args[0]
		}
		startPath, err := filepath.Abs(startPath)
		if err != nil {
			return fmt.Errorf("invalid starting path: %w", err)
		}

		// Apply smart defaults if requested
		if useSmartDefaults {
			applySmartDefaults(startPath)
		}

		filters := processFilters(excludePatterns, includePatterns)

		// 2. Find all matching files
		matchingFiles, err := findMatchingFiles(startPath, filters)
		if err != nil {
			return fmt.Errorf("error finding files: %w", err)
		}

		// If in include mode and no files were found, nothing to do
		if len(filters.includeGlobs) > 0 && len(matchingFiles) == 0 {
			fmt.Println("No files found matching the given patterns.")
			return nil
		}

		// 3. Build the tree output from the list of files
		finalOutput := buildTreeOutput(startPath, matchingFiles)

		// 4. Handle final output
		if copyToClipboard {
			if err := clipboard.WriteAll(finalOutput); err != nil {
				return fmt.Errorf("failed to copy to clipboard: %w", err)
			}
			fmt.Println("Output copied to clipboard.")
		}
		if outputFile != "" {
			if err := os.WriteFile(outputFile, []byte(finalOutput), 0644); err != nil {
				return fmt.Errorf("failed to write to output file: %w", err)
			}
			fmt.Printf("Output written to %s\n", outputFile)
		}
		if !copyToClipboard && outputFile == "" {
			fmt.Print(finalOutput)
		}

		return nil
	},
}

// expandBraces expands brace patterns like "*.{go,js}" into ["*.go", "*.js"]
func expandBraces(pattern string) []string {
	braceRegex := regexp.MustCompile(`\{([^}]*)\}`)
	matches := braceRegex.FindStringSubmatch(pattern)

	if len(matches) < 2 {
		return []string{pattern} // No braces found
	}

	// Handle empty braces case - "{}"
	if matches[1] == "" {
		return []string{strings.Replace(pattern, matches[0], "", 1)}
	}

	options := strings.Split(matches[1], ",")
	var expanded []string

	for _, option := range options {
		newPattern := strings.Replace(pattern, matches[0], strings.TrimSpace(option), 1)
		expanded = append(expanded, newPattern)
	}

	return expanded
}

func processFilters(exclude, include []string) filter {
	var expandedExclude, expandedInclude []string

	// Expand braces for exclude patterns
	for _, pattern := range exclude {
		expandedExclude = append(expandedExclude, expandBraces(pattern)...)
	}

	// Expand braces for include patterns
	for _, pattern := range include {
		expandedInclude = append(expandedInclude, expandBraces(pattern)...)
	}

	return filter{
		excludeGlobs: expandedExclude,
		includeGlobs: expandedInclude,
	}
}

// findMatchingFiles handles directory-based includes and file-based glob includes.
func findMatchingFiles(root string, f filter) ([]string, error) {
	var matchingPaths []string

	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Depth check (before exclusion / inclusion)
		if d.IsDir() && path != root {
			relPath, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}

			// Depth 0 is the root's immediate children
			depth := strings.Count(relPath, string(filepath.Separator))

			// if maxdepth is set and the current depth exceeds it, skip this directory
			if maxDepth != -1 && depth > maxDepth {
				return fs.SkipDir
			}
		}

		// --- Exclusion Logic (runs first) ---
		entryName := d.Name()
		for _, pattern := range f.excludeGlobs {
			matched, _ := filepath.Match(pattern, entryName)
			if matched {
				if d.IsDir() {
					if path == root {
						return nil
					}
					return fs.SkipDir
				}
				return nil
			}
		}

		// If not in include mode, add all non-directory files.
		if len(f.includeGlobs) == 0 && !d.IsDir() {
			// Also check depth for files when not in include mode.
			relPath, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			depth := strings.Count(relPath, string(filepath.Separator))
			if maxDepth == -1 || depth < maxDepth+1 {
				matchingPaths = append(matchingPaths, path)
			}
		}

		// In include mode, we must match files or directories explicitly.
		if len(f.includeGlobs) > 0 {
			// Case 1: A directory is an exact match for an include pattern.
			// If so, we do a sub-walk and add all its files.
			if d.IsDir() {
				for _, pattern := range f.includeGlobs {
					if d.Name() == pattern {
						// This directory is explicitly included. Walk it and add all files within.
						subWalkErr := filepath.WalkDir(path, func(subPath string, subD fs.DirEntry, _ error) error {
							if !subD.IsDir() {
								// Check if this sub-file is excluded.
								isExcluded := false
								for _, excludePattern := range f.excludeGlobs {
									if matched, _ := filepath.Match(excludePattern, subD.Name()); matched {
										isExcluded = true
										break
									}
								}
								if !isExcluded {
									matchingPaths = append(matchingPaths, subPath)
								}
							}
							return nil
						})
						if subWalkErr != nil {
							return subWalkErr
						}
						// We've processed this directory, so skip it in the main walk to avoid duplication.
						return fs.SkipDir
					}
				}
			} else { // Case 2: It's a file, check if it matches a glob-style include pattern.
				for _, pattern := range f.includeGlobs {
					if matched, _ := filepath.Match(pattern, d.Name()); matched {
						// Also check depth for files when in include mode.
						relPath, err := filepath.Rel(root, path)
						if err != nil {
							return err
						}
						depth := strings.Count(relPath, string(filepath.Separator))
						if maxDepth == -1 || depth < maxDepth+1 {
							matchingPaths = append(matchingPaths, path)
							break // Found a match, no need to check other patterns
						}
					}
				}
			}
		}

		return nil
	})

	return matchingPaths, walkErr
}

// Construct the tree output as a string
func buildTreeOutput(root string, paths []string) string {
	if len(paths) == 0 {
		return filepath.Base(root) + "\n"
	}

	// Initialize a map to hold all nodes (directories and files)
	nodes := make(map[string]bool)
	// Always include the root directory
	nodes[root] = true

	// Add all files and their parent directories to the nodes map
	for _, path := range paths {
		// Skip paths outside the root directory
		if !strings.HasPrefix(path, root) {
			continue
		}

		// Add the file itself
		nodes[path] = true

		// Add all parent directories of the path to the nodes map as well
		dir := filepath.Dir(path)
		for dir != root && dir != "." && dir != "/" {
			nodes[dir] = true
			dir = filepath.Dir(dir)
		}
	}

	// Convert to a sorted slice for consistent output
	sortedNodes := make([]string, 0, len(nodes))
	for nodePath := range nodes {
		sortedNodes = append(sortedNodes, nodePath)
	}
	sort.Strings(sortedNodes)

	// Generate the tree output
	var output strings.Builder
	// Start with the root directory name
	output.WriteString(filepath.Base(root) + "\n")

	// A map to track which directory levels have more items, for drawing the tree with '|'
	lastInDir := make(map[int]bool)

	// Process each node (skipping the root which we already output)
	for i := 1; i < len(sortedNodes); i++ {
		path := sortedNodes[i]
		// Get relative path from root
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			continue
		}

		// Calculate the depth of this node relative to root
		parts := strings.Split(relPath, string(filepath.Separator))
		depth := len(parts) - 1

		// Check if this is the last entry at its depth or in its parent directory
		isLast := true
		if i < len(sortedNodes)-1 {
			nextPath := sortedNodes[i+1]
			nextRelPath, err := filepath.Rel(root, nextPath)
			if err == nil {
				nextParts := strings.Split(nextRelPath, string(filepath.Separator))
				nextDepth := len(nextParts) - 1

				// If next item is at same depth and has same parent, this isn't last
				if nextDepth == depth && len(parts) > 1 && len(nextParts) > 1 &&
					parts[0] == nextParts[0] {
					isLast = false
				}
			}
		}
		lastInDir[depth] = isLast

		// Print indentation
		for j := 0; j < depth; j++ {
			if lastInDir[j] {
				output.WriteString("    ")
			} else {
				output.WriteString("│   ")
			}
		}

		// Print branch prefix
		if isLast {
			output.WriteString("└── ")
		} else {
			output.WriteString("├── ")
		}

		output.WriteString(filepath.Base(path) + "\n")
	}

	return output.String()
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringSliceVarP(&excludePatterns, "exclude", "e", []string{}, "Glob patterns to exclude (e.g., .git, *.log, node_modules)")
	rootCmd.Flags().StringSliceVarP(&includePatterns, "include", "i", []string{}, "Glob patterns to include (e.g., .git, *.go, *.md)")
	rootCmd.Flags().StringVarP(&outputFile, "out", "o", "", "Output to a file instead of the console")
	rootCmd.Flags().BoolVarP(&copyToClipboard, "copy", "c", false, "Copy the output to the system clipboard")
	rootCmd.Flags().BoolVarP(&showPatterns, "show-patterns", "p", false, "Show a guide for using glob patterns")
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "Show version information")
	rootCmd.Flags().BoolVarP(&useSmartDefaults, "smart-defaults", "s", false, "Apply smart defaults based on detected project type")
	rootCmd.Flags().IntVarP(&maxDepth, "depth", "d", 1, "Set the maximum depth of the directory tree to display (-1 for unlimited). (Default = 1)")
}

func printPatternHelp() {
	fmt.Print(`
GLOB PATTERN GUIDE
==================

Glob patterns are simple wildcard patterns used to match file and directory names.
Here are the most common patterns you can use with wintree:

BASIC PATTERNS:
┌─────────────┬─────────────────────────────────────────────────────────┐
│ Pattern     │ Description                                             │
├─────────────┼─────────────────────────────────────────────────────────┤
│ *           │ Matches any number of characters (except path separator)│
│ ?           │ Matches exactly one character                           │
│ [abc]       │ Matches any one of the characters a, b, or c            │
│ [a-z]       │ Matches any character from a to z                       │
│ [!abc]      │ Matches any character except a, b, or c                 │
│ **          │ Not supported                                           │
│ *.ext       │ Matches all files ending with .ext                      │
│ file*       │ Matches all files starting with 'file'                  │
│ *file*      │ Matches all files containing 'file'                     │
│ file?.txt   │ Matches file1.txt, fileA.txt, etc.                      │
│ [0-9]*      │ Matches files starting with a digit                     │
│ *.[ch]      │ Matches files ending with .c or .h                      │
│ *.{go,js}   │ Expands to *.go and *.js (Now supported!)               │
│ dir/*       │ Matches all files in 'dir' directory                    │
│ dir/**      │ Not supported; use --include "dir" for directories      │
└─────────────┴─────────────────────────────────────────────────────────┘

COMMON USE CASES:

📁 Exclude common directories:
   --exclude .git --exclude node_modules --exclude .vscode

📄 Include only specific file types:
   --include "*.go" --include "*.md" --include "*.txt"

🚫 Exclude log and temporary files:
   --exclude "*.log" --exclude "*.tmp" --exclude "*.temp"

💻 Include only source code files:
   --include "*.go" --include "*.js" --include "*.py" --include "*.java"

🔧 Exclude build and cache directories:
   --exclude target --exclude build --exclude dist --exclude .cache

EXAMPLES:

1. Show only Go files:
   wintree --include "*.go"

2. Exclude git and node_modules directories:
   wintree --exclude .git --exclude node_modules

3. Show only documentation files:
   wintree --include "*.md" --include "*.txt" --include "*.rst"

4. Exclude all hidden files and directories (starting with .):
   wintree --exclude ".*"

5. Include files starting with 'test':
   wintree --include "test*"

6. Include files ending with specific extensions:
   wintree --include "*.{go,js,py}"

7. Include files with a single character suffix:
   wintree --include "file?.txt"

8. Include files starting with a digit:
   wintree --include "[0-9]*"

9. Include C and header files:
   wintree --include "*.[ch]"

TIPS:
• You can use multiple --include and --exclude flags
• Patterns are case-sensitive on Linux/Mac, case-insensitive on Windows
• Directory names are matched exactly (no glob patterns for directories)
• File names support full glob pattern matching
• Exclusions are processed before inclusions
• Curly brace expansion (*.{go,js}) is supported

`)
}

// printVersionInfo displays version information
// printVersionInfo displays version information
func printVersionInfo() {
	if Version == "dev" {
		// For dev builds, try to get version from git describe
		fmt.Printf("wintree %s (development build)\n", Version)
	} else {
		// For release builds, just show the version
		fmt.Printf("wintree %s\n", Version)
	}
}

// detectProjectType analyzes the directory to determine project type
func detectProjectType(path string) string {
	files, err := os.ReadDir(path)
	if err != nil {
		return "unknown"
	}

	for _, file := range files {
		name := file.Name()
		switch name {
		case "go.mod", "go.sum":
			return "go"
		case "package.json", "node_modules":
			return "node"
		case "requirements.txt", "pyproject.toml", "setup.py", "Pipfile":
			return "python"
		case "Cargo.toml":
			return "rust"
		case "pom.xml", "build.gradle":
			return "java"
		case "Gemfile":
			return "ruby"
		case "composer.json":
			return "php"
		case "package.swift", "Package.swift":
			return "swift"
		case "pubspec.yaml":
			return "dart"
		case "mix.exs":
			return "elixir"
		case ".gitignore":
			// Continue checking for more specific indicators
			continue
		}
	}

	// Check for common framework files
	for _, file := range files {
		name := file.Name()
		switch name {
		case "angular.json":
			return "angular"
		case "next.config.js", "next.config.mjs":
			return "nextjs"
		case "nuxt.config.js", "nuxt.config.ts":
			return "nuxtjs"
		case "gatsby-config.js":
			return "gatsby"
		case "svelte.config.js":
			return "svelte"
		case "vite.config.js", "vite.config.ts":
			return "vite"
		case "webpack.config.js":
			return "webpack"
		}
	}

	return "unknown"
}

// getSmartDefaults returns smart exclusion patterns based on project type
func getSmartDefaults(projectType string) []string {
	commonDefaults := []string{".git", ".DS_Store", "Thumbs.db"}

	switch projectType {
	case "go":
		return append(commonDefaults, "vendor", "*.exe", "*.dll", "*.so", "*.dylib")
	case "node", "angular", "nextjs", "nuxtjs", "gatsby", "svelte", "vite", "webpack":
		return append(commonDefaults, "node_modules", "dist", "build", ".next", ".nuxt", "coverage", "*.log")
	case "python":
		return append(commonDefaults, "__pycache__", "*.pyc", "*.pyo", "*.pyd", ".Python", "env", "venv", ".env", ".venv", "pip-log.txt", "pip-delete-this-directory.txt", ".coverage")
	case "rust":
		return append(commonDefaults, "target", "Cargo.lock")
	case "java":
		return append(commonDefaults, "target", "build", "*.class", "*.jar", "*.war", "*.ear")
	case "ruby":
		return append(commonDefaults, "vendor", ".bundle", "*.gem")
	case "php":
		return append(commonDefaults, "vendor", "composer.lock")
	case "swift":
		return append(commonDefaults, ".build", "Packages", "*.xcodeproj", "*.xcworkspace")
	case "dart":
		return append(commonDefaults, ".dart_tool", "build", ".packages")
	case "elixir":
		return append(commonDefaults, "_build", "deps", "*.beam")
	default:
		return append(commonDefaults, "node_modules", "target", "build", "dist", "vendor", "*.log", "*.tmp")
	}
}

// applySmartDefaults applies smart exclusion patterns based on detected project type
func applySmartDefaults(path string) {
	projectType := detectProjectType(path)
	smartDefaults := getSmartDefaults(projectType)

	// Only add defaults that aren't already in excludePatterns
	for _, defaultPattern := range smartDefaults {
		alreadyExists := false
		for _, existing := range excludePatterns {
			if existing == defaultPattern {
				alreadyExists = true
				break
			}
		}
		if !alreadyExists {
			excludePatterns = append(excludePatterns, defaultPattern)
		}
	}

	fmt.Printf("🧠 Smart defaults applied for %s project\n", projectType)
	fmt.Printf("   Excluding: %s\n", strings.Join(smartDefaults, ", "))
	fmt.Println()
}
