/*
Copyright © 2025 Maxim Dribny <mdribnyi@gmail.com>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	excludePatterns []string
	//TODO: finish the include pattern logic and uncomment
	//includePatterns []string
	outputFile      string
	copyToClipboard bool
)

type filter struct {
	excludeDirs map[string]struct{}
	excludeExts map[string]struct{}
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
		// Using RunE to return an error, which Cobra will print nicely.
		// This is modern Go practice for CLIs

		// 1. Determine Starting Path
		startPath := "."
		if len(args) > 0 {
			startPath = args[0]
		}
		// Clean the path to handle things like "path/.." or "./path" consistently

		return nil
	},
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
	// init() is a special Go function that runs before main()
	// This is the standard place to set up flags in Cobra

	// Here we define our flags and bind them to the variables we declared above.
	// The Flags().StringSliceP() function creates a flag that can accept multiple string values.
	// "exclude" is the long name, "e" is the short name.
	// The empty slice is the default value.
	// The final string is the help message.
	rootCmd.Flags().StringSliceVarP(&excludePatterns, "exclude", "e", []string{}, "Patterns to exclude (files/dirs, e.g., .git, *.log)")
	rootCmd.Flags().StringSliceVarP(&includePatterns, "include", "i", []string{}, "Patterns to include (whitelist, e.g., *.go, *.md)")

	// A simple string flag for the output file path
	rootCmd.Flags().StringVarP(&outputFile, "out", "o", "", "Output to a file instead of the console")

	// A boolean flag. If the user includes '--copy' or `-c`, copyToClipboard will be true.
	rootCmd.Flags().BoolVarP(&copyToClipboard, "copy", "c", false, "Copy the output to the system clipboard")
}
