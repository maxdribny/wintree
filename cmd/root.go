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

var (
	excludePatterns []string
	includePatterns []string
	outputFile      string
	copyToClipboard bool
	showPatterns    bool
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
	isIncludeMode := len(f.includeGlobs) > 0

	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
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
		if !isIncludeMode {
			if !d.IsDir() {
				matchingPaths = append(matchingPaths, path)
			}
			return nil
		}

		// In include mode, we must match files or directories explicitly.
		// Case 1: A directory is an exact match for an include pattern.
		// If so, we do a sub-walk and add all its files.
		if d.IsDir() {
			for _, pattern := range f.includeGlobs {
				// Don't glob, use exact match for directory inclusion.
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
					// We've processed this directory, so skip it in the main walk.
					return fs.SkipDir
				}
			}
		}

		// Case 2: A file matches a glob-style include pattern.
		if !d.IsDir() {
			for _, pattern := range f.includeGlobs {
				if matched, _ := filepath.Match(pattern, d.Name()); matched {
					matchingPaths = append(matchingPaths, path)
					break
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
	rootCmd.Flags().BoolVarP(&showPatterns, "show-patterns", "p", false, "Show a guide for using glob pattenrns")
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
