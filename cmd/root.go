/*
Copyright © 2025 Maxim Dribny <mdribnyi@gmail.com>
*/
package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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
	excludeDirs  map[string]struct{}
	excludeExts  map[string]struct{}
	includeGlobs []string
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "wintree [path]",
	Short: "A modern, cross-platform tree command.",
	Long: `wintree is a powerful, intuitive, and easy-to-use alternative to the
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

		filter := processFilters(excludePatterns, includePatterns)

		// 3. Walk Directory
		var output strings.Builder

		output.WriteString(filepath.Base(startPath) + "\n")

		// filepath.WalkDir takes a root path and a callback
		// which is executed for every file and directory it finds
		walkErr := filepath.WalkDir(startPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if path == startPath {
				return nil
			}

			// --- Filtering --- //

			// Check if the directory should be excluded
			if d.IsDir() {
				dirName := d.Name()
				if _, shouldExclude := filter.excludeDirs[dirName]; shouldExclude {
					// fs.SkipDir is a special error flag that tells WalkDir to
					// not visit any of the files in a v efficient way
					return fs.SkipDir
				}
			}

			// Check if the file extension should be excluded
			extension := filepath.Ext(d.Name())
			if _, shouldExclude := filter.excludeExts[extension]; shouldExclude && extension != "" {
				// We just return nil to skip the single file and continue the walk
				return nil
			}

			// --- Formatting Logic --- //

			relativePath, err := filepath.Rel(startPath, path)
			if err != nil {
				return err
			}

			depth := strings.Count(relativePath, string(filepath.Separator))

			for i := 0; i < depth; i++ {
				output.WriteString("|   ")
			}
			// Last branch in the tree
			output.WriteString("|-- ")
			output.WriteString(d.Name() + "\n")

			return nil
		})

		if walkErr != nil {
			return fmt.Errorf("error walking directory: %w", walkErr)
		}

		// 4. Handle Final Output
		finalOutput := output.String()

		// If --copy flag is set
		if copyToClipboard {
			if err := clipboard.WriteAll(finalOutput); err != nil {
				return fmt.Errorf("failed to copy to clipboard: %w", err)
			}
			fmt.Println("Output copied to clipboard.")
		}

		// If --out flag is set
		if outputFile != "" {
			// os.WriteFile is a convenient way to write data to a file
			// 0644 is a standard file permission setting (readable by all, writable by owner)
			if err := os.WriteFile(outputFile, []byte(finalOutput), 0644); err != nil {
				return fmt.Errorf("failed to write to output file: %w", err)
			}
			fmt.Printf("Output written to %s\n", outputFile)
		}

		// If neither flag, print to standard output
		if !copyToClipboard && outputFile == "" {
			fmt.Print(finalOutput)
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
	var matchingPaths = []string
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

		matchingPaths = append(matchingPaths, path)
		return nil
	})

	return matchingPaths, walkErr
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
