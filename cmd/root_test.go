package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExpandBraces(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected []string
	}{
		{
			name:     "no braces",
			pattern:  "*.go",
			expected: []string{"*.go"},
		},
		{
			name:     "simple brace expansion",
			pattern:  "*.{go,js}",
			expected: []string{"*.go", "*.js"},
		},
		{
			name:     "multiple extensions",
			pattern:  "*.{go,js,py,java}",
			expected: []string{"*.go", "*.js", "*.py", "*.java"},
		},
		{
			name:     "with spaces",
			pattern:  "*.{go, js, py}",
			expected: []string{"*.go", "*.js", "*.py"},
		},
		{
			name:     "directory names",
			pattern:  "{src,test,docs}",
			expected: []string{"src", "test", "docs"},
		},
		{
			name:     "complex pattern",
			pattern:  "test*.{go,js}",
			expected: []string{"test*.go", "test*.js"},
		},
		{
			name:     "empty braces",
			pattern:  "*.{}",
			expected: []string{"*."},
		},
		{
			name:     "single option",
			pattern:  "*.{go}",
			expected: []string{"*.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandBraces(tt.pattern)
			if len(result) != len(tt.expected) {
				t.Errorf("expandBraces(%q) returned %d items, expected %d", tt.pattern, len(result), len(tt.expected))
				return
			}
			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("expandBraces(%q) returned %q, expected %q", tt.pattern, result[i], expected)
				}
			}
		})
	}
}

func TestProcessFilters(t *testing.T) {
	tests := []struct {
		name            string
		exclude         []string
		include         []string
		expectedExclude []string
		expectedInclude []string
	}{
		{
			name:            "no patterns",
			exclude:         []string{},
			include:         []string{},
			expectedExclude: []string{},
			expectedInclude: []string{},
		},
		{
			name:            "simple patterns",
			exclude:         []string{"*.log"},
			include:         []string{"*.go"},
			expectedExclude: []string{"*.log"},
			expectedInclude: []string{"*.go"},
		},
		{
			name:            "brace expansion in exclude",
			exclude:         []string{"*.{log,tmp}"},
			include:         []string{"*.go"},
			expectedExclude: []string{"*.log", "*.tmp"},
			expectedInclude: []string{"*.go"},
		},
		{
			name:            "brace expansion in include",
			exclude:         []string{"*.log"},
			include:         []string{"*.{go,js}"},
			expectedExclude: []string{"*.log"},
			expectedInclude: []string{"*.go", "*.js"},
		},
		{
			name:            "multiple patterns with braces",
			exclude:         []string{"*.{log,tmp}", "node_modules"},
			include:         []string{"*.{go,js,py}", "*.md"},
			expectedExclude: []string{"*.log", "*.tmp", "node_modules"},
			expectedInclude: []string{"*.go", "*.js", "*.py", "*.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processFilters(tt.exclude, tt.include)

			if len(result.excludeGlobs) != len(tt.expectedExclude) {
				t.Errorf("processFilters() excludeGlobs length = %d, expected %d",
					len(result.excludeGlobs), len(tt.expectedExclude))
			}

			if len(result.includeGlobs) != len(tt.expectedInclude) {
				t.Errorf("processFilters() includeGlobs length = %d, expected %d",
					len(result.includeGlobs), len(tt.expectedInclude))
			}

			for i, expected := range tt.expectedExclude {
				if i >= len(result.excludeGlobs) || result.excludeGlobs[i] != expected {
					t.Errorf("processFilters() excludeGlobs[%d] = %q, expected %q", i, result.excludeGlobs[i], expected)
				}
			}

			for i, expected := range tt.expectedInclude {
				if i >= len(result.includeGlobs) || result.includeGlobs[i] != expected {
					t.Errorf("processFilters() includeGlobs[%d] = %q, expected %q", i, result.includeGlobs[i], expected)
				}
			}
		})
	}
}

func TestBuildTreeOutput(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir, err := os.MkdirTemp("", "wintree_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	// Create test files
	testFiles := []string{
		"file1.go",
		"file2.js",
		"subdir/file3.py",
		"subdir/nested/file4.txt",
		"another/file5.md",
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(fullPath, []byte("test content"), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Test building tree output
	paths := make([]string, len(testFiles))
	for i, file := range testFiles {
		paths[i] = filepath.Join(tempDir, file)
	}

	output := buildTreeOutput(tempDir, paths)

	// Check that output contains expected elements
	expectedElements := []string{
		filepath.Base(tempDir), // root directory name
		"file1.go",
		"file2.js",
		"subdir",
		"file3.py",
		"nested",
		"file4.txt",
		"another",
		"file5.md",
	}

	for _, element := range expectedElements {
		if !strings.Contains(output, element) {
			t.Errorf("buildTreeOutput() missing expected element: %q", element)
		}
	}

	// Check tree characters are present
	if !strings.Contains(output, "├──") && !strings.Contains(output, "└──") {
		t.Error("buildTreeOutput() missing tree structure containers")
	}
}

func TestRootCmd(t *testing.T) {
	// Setup test directory
	tempDir, err := os.MkdirTemp("", "root_cmd_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFiles := []string{
		"file1.go",
		"file2.js",
		"file3.log",
		"subdir/file4.py",
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(fullPath, []byte("test content"), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Test cases for rootCmd
	tests := []struct {
		name           string
		args           []string
		excludeFlags   []string
		includeFlags   []string
		outputFile     string
		clipboard      bool
		showPatterns   bool
		expectError    bool
		expectedOutput string
	}{
		{
			name:         "show patterns help",
			showPatterns: true,
			expectError:  false,
		},
		{
			name:        "basic tree",
			args:        []string{tempDir},
			expectError: false,
		},
		{
			name:         "include only go files",
			args:         []string{tempDir},
			includeFlags: []string{"*.go"},
			expectError:  false,
		},
		{
			name:         "exclude log files",
			args:         []string{tempDir},
			excludeFlags: []string{"*.log"},
			expectError:  false,
		},
		{
			name:         "brace expansion",
			args:         []string{tempDir},
			includeFlags: []string{"*.{go,js}"},
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags before each test
			excludePatterns = []string{}
			includePatterns = []string{}
			outputFile = ""
			copyToClipboard = false
			showPatterns = false
			showVersion = false
			useSmartDefaults = false

			// Set up flags
			excludePatterns = tt.excludeFlags
			includePatterns = tt.includeFlags
			outputFile = tt.outputFile
			copyToClipboard = tt.clipboard
			showPatterns = tt.showPatterns

			// Capture stdout for verification
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Execute command
			err := rootCmd.RunE(rootCmd, tt.args)

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			io.Copy(&buf, r)

			// Check error
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// For show patterns, verify help text is printed
			if tt.showPatterns {
				if !strings.Contains(buf.String(), "GLOB PATTERN GUIDE") {
					t.Error("expected pattern help but didn't get it")
				}
			}
		})
	}
}
