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

	"github.com/spf13/cobra"
)

var (
	excludePatterns []string
	includePatterns []string
	outputFile      string
	copyToClipboard bool
)

type filter struct {
	excludeDirs  map[string]struct{}
	excludeExts  map[string]struct{}
	includeGlobs []string
}

// rootCmd represents the base command when called without any subcommands
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
			fmt.Println("No files found matching include patterns.")
			return nil
		}

		// 3. Build the tree output from the list of files
		finalOuput := buildTreeOutput(startPath, matchingFiles, filters)

		// 4. Handle final output
		if copyToClipboard {
			if err := os.WriteFile(outputFile, []byte(finalOuput), 0644); err != nil {
				return fmt.Errorf("failed to write to output file: %w", err)
			}
			fmt.Println("Output copied to clipboard.")
		}
		if outputFile != "" {
			if err := os.WriteFile(outputFile, []byte(finalOuput), 0644); err != nil {
				return fmt.Errorf("failed to write to output file: %w", err)
			}
			fmt.Printf("Output written to %s\n", outputFile)
		}
		if !copyToClipboard && outputFile == "" {
			fmt.Print(finalOuput)
		}

		return nil
	},
}

// processFilters takes the raw string slices from the flags and organizes them into maps for fast lookups
func processFilters(exclude, include []string) filter {
	f := filter{
		excludeDirs:  make(map[string]struct{}),
		excludeExts:  make(map[string]struct{}),
		includeGlobs: include,
	}

	for _, pattern := range exclude {
		if strings.HasPrefix(pattern, ".") && !strings.Contains(pattern, string(filepath.Separator)) {
			f.excludeExts[pattern] = struct{}{}
		} else {
			f.excludeDirs[pattern] = struct{}{}
		}
	}
	return f
}

func findMatchingFiles(root string, f filter) ([]string, error) {
	var matchingPaths []string
	isIncludeMode := len(f.includeGlobs) > 0

	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Exclusion logic - always applied first
		if d.IsDir() {
			if _, shouldExclude := f.excludeDirs[d.Name()]; shouldExclude && path != root {
				return fs.SkipDir
			}
		}
		if ext := filepath.Ext(d.Name()); ext != "" {
			if _, shouldExclude := f.excludeExts[ext]; shouldExclude {
				return nil
			}
		}

		// If it's a directory, we don't add it to the list. We let the walk continue to find files inside it.
		if d.IsDir() {
			return nil
		}

		if isIncludeMode {
			match := false
			for _, pattern := range f.includeGlobs {
				if matched, _ := filepath.Match(pattern, d.Name()); matched {
					match = true
					break
				}
			}
			if !match {
				return nil // File does not match any include pattern, so skip it.
			}
		}

		// File has passed all filters, so we add it to the list.
		matchingPaths = append(matchingPaths, path)
		return nil
	})

	return matchingPaths, walkErr
}

// buildTreeOutput takes a list of file paths and constructs the visual tree string.
func buildTreeOutput(root string, paths []string, f filter) string {
	var output strings.Builder
	output.WriteString(filepath.Base(root) + "\n")

	// Use a map to store all unique nodes (files and their parent dirs) that need to be printed.
	nodes := make(map[string]struct{})
	for _, path := range paths {
		nodes[path] = struct{}{}
		// Add all parent directories of the path to the nodes map as well
		dir := filepath.Dir(path)
		for dir != root && dir != "." {
			// Stop if we hit an excluded directory
			if _, isExcluded := f.excludeDirs[filepath.Base(dir)]; isExcluded {
				break
			}
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
			// If the next item has the same directory prefix, this one isnt the last.
			if strings.HasPrefix(nextRelPath, filepath.Dir(relPath)+string(filepath.Separator)) {
				isLast = false
			}
		}

		// Print indentation
		for j := 0; j < depth; j++ {
			if lastInDir[j] {
				output.WriteString("    ") // Parent was the last, so no vertical line
			} else {
				output.WriteString("|   ")
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
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringSliceVarP(&excludePatterns, "exclude", "e", []string{}, "Patterns to exclude (dirs like 'node_modules' or exts like '.log')")
	rootCmd.Flags().StringSliceVarP(&includePatterns, "include", "i", []string{}, "Patterns to include (whitelist, e.g., *.go, *.md)")
	rootCmd.Flags().StringVarP(&outputFile, "out", "o", "", "Output to a file instead of the console")
	rootCmd.Flags().BoolVarP(&copyToClipboard, "copy", "c", false, "Copy the output to the system clipboard")
}
