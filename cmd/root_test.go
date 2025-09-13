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
			if _, err := io.Copy(&buf, r); err != nil {
				t.Errorf("failed to copy output to buffer: %v", err)
			}

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

func TestBuildTreeOutput_WithFullPath(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir, err := os.MkdirTemp("", "wintree_test_fullpath")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFiles := []string{
		"file1.go",
		"subdir/file2.js",
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

	// Get full paths for the files
	paths := make([]string, len(testFiles))
	for i, file := range testFiles {
		paths[i] = filepath.Join(tempDir, file)
	}

	// Test with showFullPath = false (default behaviour)
	t.Run("without full path", func(t *testing.T) {
		showFullPath = false
		output := buildTreeOutput(tempDir, paths)

		// Should NOT contain the full path
		if strings.HasPrefix(output, tempDir) {
			t.Error("buildTreeOutput() should not start with the full path when showFullPath=false")
		}

		// Should start with base directory name
		if !strings.HasPrefix(output, filepath.Base(tempDir)) {
			t.Errorf("buildTreeOutput() should start with base directory name %q", filepath.Base(tempDir))
		}
	})

	// Test with showFullPath = true
	t.Run("with full path", func(t *testing.T) {
		showFullPath = true

		defer func() {
			showFullPath = false
		}()

		output := buildTreeOutput(tempDir, paths)

		// Should contain the full path as first line
		lines := strings.Split(output, "\n")
		if len(lines) < 2 {
			t.Errorf("Output should have at least two lines")
		}

		if lines[0] != tempDir {
			t.Errorf("First line should be full path: got %q, want %q", lines[0], tempDir)
		}

		if lines[1] != filepath.Base(tempDir) {
			t.Errorf("Second line should be base path: got %q, want %q", lines[1], filepath.Base(tempDir))
		}
	})

	// Test with empty directory
	t.Run("empty directory", func(t *testing.T) {
		showFullPath = true

		defer func() {
			showFullPath = false
		}()

		output := buildTreeOutput(tempDir, []string{})

		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) != 2 {
			t.Fatalf("Empty directory output should have exactly two lines, got %d", len(lines))
		}

		if lines[0] != tempDir {
			t.Errorf("First line should be full path: got %q, want %q", lines[0], tempDir)
		}

		if lines[1] != filepath.Base(tempDir) {
			t.Errorf("Second line should be base path: got %q, want %q", lines[1], filepath.Base(tempDir))
		}
	})
}
