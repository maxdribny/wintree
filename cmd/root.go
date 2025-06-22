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
)

var (
	excludePatterns []string
	includePatterns []string
	outputFile      string
	copyToClipboard bool
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

func processFilters(exclude, include []string) filter {
	return filter{
		excludeGlobs: exclude,
		includeGlobs: include,
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
	var output strings.Builder
	output.WriteString(filepath.Base(root) + "\n")

	nodes := make(map[string]struct{})
	for _, path := range paths {
		nodes[path] = struct{}{}
		// Add all parent directories of the path to the nodes map as well
		dir := filepath.Dir(path)
		for dir != root && dir != "." {
			nodes[dir] = struct{}{}
			dir = filepath.Dir(dir)
		}
	}

	sortedNodes := make([]string, 0, len(nodes))
	for nodePath := range nodes {
		sortedNodes = append(sortedNodes, nodePath)
	}
	sort.Strings(sortedNodes)

	// A map to track which directory levels have more items, for drawing the tree with '|'
	lastInDir := make(map[int]bool)
	for i, path := range sortedNodes {
		relPath, _ := filepath.Rel(root, path)
		depth := strings.Count(relPath, string(filepath.Separator))

		// Check if this is the last entry at its depth or in its parent directory
		isLast := true
		if i+1 < len(sortedNodes) {
			nextRelPath, _ := filepath.Rel(root, sortedNodes[i+1])
			// If the next item has the same directory prefix, this one isn't the last.
			if strings.HasPrefix(nextRelPath, filepath.Dir(relPath)+string(filepath.Separator)) {
				isLast = false
			}
		}

		// Print indentation
		for j := 0; j < depth; j++ {
			if lastInDir[j] {
				output.WriteString("    ") // Parent was the last, so no vertical line
			} else {
				output.WriteString("│   ")
			}
		}

		// Print branch prefix
		if isLast {
			output.WriteString("└── ")
			lastInDir[depth] = true
		} else {
			output.WriteString("├── ")
			lastInDir[depth] = false
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
}
