package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// setupTestDirectory creates a test directory structure
func setupTestDirectory(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "wintree_integration_test")
	if err != nil {
		t.Fatal(err)
	}

	// Create a complex directory structure
	structure := map[string]string{
		"main.go":                       "package main",
		"README.md":                     "# Test Project",
		"go.mod":                        "module test",
		".gitignore":                    "*.log",
		"src/app.go":                    "package src",
		"src/utils.js":                  "// utils",
		"src/styles.css":                "body {}",
		"tests/main_test.go":            "package tests",
		"tests/app_test.js":             "// test",
		"docs/api.md":                   "# API",
		"docs/guide.txt":                "Guide",
		"logs/app.log":                  "log entry",
		"logs/error.log":                "error",
		"node_modules/package/index.js": "// package",
		"build/output.bin":              "binary",
		"temp/cache.tmp":                "cache",
		".vscode/settings.json":         "{}",
		"hidden/.secret":                "secret",
	}

	for filePath, content := range structure {
		fullPath := filepath.Join(tempDir, filePath)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	return tempDir
}

func TestFindMatchingFiles_IncludeMode(t *testing.T) {
	testDir := setupTestDirectory(t)
	defer os.RemoveAll(testDir)

	// Save original maxDepth and restore it after test
	originalMaxDepth := maxDepth
	defer func() { maxDepth = originalMaxDepth }()

	maxDepth = -1

	tests := []struct {
		name            string
		includePatterns []string
		expectedFiles   []string
		unexpectedFiles []string
	}{
		{
			name:            "include go files",
			includePatterns: []string{"*.go"},
			expectedFiles:   []string{"main.go", "app.go", "main_test.go"},
			unexpectedFiles: []string{"utils.js", "README.md"},
		},
		{
			name:            "include with brace expansion",
			includePatterns: []string{"*.{go,js}"},
			expectedFiles:   []string{"main.go", "app.go", "utils.js", "main_test.go", "app_test.js", "index.js"},
			unexpectedFiles: []string{"README.md", "styles.css"},
		},
		{
			name:            "include markdown files",
			includePatterns: []string{"*.md"},
			expectedFiles:   []string{"README.md", "api.md"},
			unexpectedFiles: []string{"main.go", "guide.txt"},
		},
		{
			name:            "include directory",
			includePatterns: []string{"docs"},
			expectedFiles:   []string{"api.md", "guide.txt"},
			unexpectedFiles: []string{"main.go", "app.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters := processFilters([]string{}, tt.includePatterns)
			matchingFiles, err := findMatchingFiles(testDir, filters)
			if err != nil {
				t.Fatalf("findMatchingFiles() error = %v", err)
			}

			// Convert to set of basenames for easier checking
			fileNames := make(map[string]bool)
			for _, file := range matchingFiles {
				fileNames[filepath.Base(file)] = true
			}

			for _, expected := range tt.expectedFiles {
				if !fileNames[expected] {
					t.Errorf("Expected file %q not found in results", expected)
				}
			}

			for _, unexpected := range tt.unexpectedFiles {
				if fileNames[unexpected] {
					t.Errorf("Unexpected file %q found in results", unexpected)
				}
			}
		})
	}
}

func TestFindMatchingFiles_ExcludeMode(t *testing.T) {
	testDir := setupTestDirectory(t)
	defer os.RemoveAll(testDir)

	tests := []struct {
		name            string
		excludePatterns []string
		expectedFiles   []string
		unexpectedFiles []string
	}{
		{
			name:            "exclude log files",
			excludePatterns: []string{"*.log"},
			expectedFiles:   []string{"main.go", "README.md"},
			unexpectedFiles: []string{"app.log", "error.log"},
		},
		{
			name:            "exclude with brace expansion",
			excludePatterns: []string{"*.{log,tmp}"},
			expectedFiles:   []string{"main.go", "README.md"},
			unexpectedFiles: []string{"app.log", "cache.tmp"},
		},
		{
			name:            "exclude directories",
			excludePatterns: []string{"node_modules", "build"},
			expectedFiles:   []string{"main.go", "README.md"},
			unexpectedFiles: []string{"index.js", "output.bin"},
		},
		{
			name:            "exclude hidden files",
			excludePatterns: []string{".*"},
			expectedFiles:   []string{"main.go", "README.md"},
			unexpectedFiles: []string{".gitignore", "settings.json", ".secret"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters := processFilters(tt.excludePatterns, []string{})
			matchingFiles, err := findMatchingFiles(testDir, filters)
			if err != nil {
				t.Fatalf("findMatchingFiles() error = %v", err)
			}

			// Convert to set of basenames for easier checking
			fileNames := make(map[string]bool)
			for _, file := range matchingFiles {
				fileNames[filepath.Base(file)] = true
			}

			for _, expected := range tt.expectedFiles {
				if !fileNames[expected] {
					t.Errorf("Expected file %q not found in results", expected)
				}
			}

			for _, unexpected := range tt.unexpectedFiles {
				if fileNames[unexpected] {
					t.Errorf("Unexpected file %q found in results", unexpected)
				}
			}
		})
	}
}

func TestFindMatchingFiles_CombinedMode(t *testing.T) {
	testDir := setupTestDirectory(t)
	defer os.RemoveAll(testDir)

	// Test include + exclude combination
	filters := processFilters(
		[]string{"*_test.go"}, // exclude test files
		[]string{"*.go"},      // include only go files
	)

	matchingFiles, err := findMatchingFiles(testDir, filters)
	if err != nil {
		t.Fatalf("findMatchingFiles() error = %v", err)
	}

	fileNames := make(map[string]bool)
	for _, file := range matchingFiles {
		fileNames[filepath.Base(file)] = true
	}

	// Should include go files but not test files
	if !fileNames["main.go"] {
		t.Error("Expected main.go to be included")
	}
	if !fileNames["app.go"] {
		t.Error("Expected app.go to be included")
	}
	if fileNames["main_test.go"] {
		t.Error("main_test.go should be excluded")
	}
	if fileNames["utils.js"] {
		t.Error("utils.js should not be included")
	}
}

func TestEdgeCases(t *testing.T) {
	t.Run("empty directory", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "empty_test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)

		filters := processFilters([]string{}, []string{})
		matchingFiles, err := findMatchingFiles(tempDir, filters)
		if err != nil {
			t.Fatalf("findMatchingFiles() error = %v", err)
		}

		if len(matchingFiles) != 0 {
			t.Errorf("Expected 0 files in empty directory, got %d", len(matchingFiles))
		}
	})

	t.Run("nonexistent directory", func(t *testing.T) {
		_, err := findMatchingFiles("/nonexistent/path", filter{})
		if err == nil {
			t.Error("Expected error for nonexistent directory")
		}
	})

	t.Run("include mode with no matches", func(t *testing.T) {
		testDir := setupTestDirectory(t)
		defer os.RemoveAll(testDir)

		filters := processFilters([]string{}, []string{"*.nonexistent"})
		matchingFiles, err := findMatchingFiles(testDir, filters)
		if err != nil {
			t.Fatalf("findMatchingFiles() error = %v", err)
		}

		if len(matchingFiles) != 0 {
			t.Errorf("Expected 0 files for nonexistent pattern, got %d", len(matchingFiles))
		}
	})
}

func TestFindMatchingFiles_DepthMode(t *testing.T) {
	testDir := setupTestDirectory(t)
	defer os.RemoveAll(testDir)

	tests := []struct {
		name          string
		depth         int
		expectedFiles []string
		unexpected    []string
	}{
		{
			name:          "depth 0",
			depth:         0,
			expectedFiles: []string{"main.go", "README.md", "go.mod", ".gitignore"},
			unexpected:    []string{"app.go", "main_test.go", "api.md"},
		},
		{
			name:          "depth 1",
			depth:         1,
			expectedFiles: []string{"main.go", "app.go", "utils.js", "main_test.go", "api.md", "settings.json"},
			unexpected:    []string{"index.js"}, // This is at depth 2
		},
		{
			name:          "unlimited depth",
			depth:         -1, // -1 means unlimited
			expectedFiles: []string{"main.go", "app.go", "index.js", "output.bin"},
			unexpected:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the global maxDepth for the test
			maxDepth = tt.depth

			filters := processFilters([]string{}, []string{}) // No filters, just depth
			matchingFiles, err := findMatchingFiles(testDir, filters)
			if err != nil {
				t.Fatalf("findMatchingFiles() with depth %d error = %v", tt.depth, err)
			}

			// Convert to set of basenames for easier checking
			fileNames := make(map[string]bool)
			for _, file := range matchingFiles {
				fileNames[filepath.Base(file)] = true
			}

			for _, expected := range tt.expectedFiles {
				if !fileNames[expected] {
					t.Errorf("Depth %d: Expected file %q not found", tt.depth, expected)
				}
			}

			for _, unexpectedFile := range tt.unexpected {
				if fileNames[unexpectedFile] {
					t.Errorf("Depth %d: Unexpected file %q found", tt.depth, unexpectedFile)
				}
			}
		})
	}
	// Reset maxDepth after tests to avoid affecting other tests
	maxDepth = -1
}
