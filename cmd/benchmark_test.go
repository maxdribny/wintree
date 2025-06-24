package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkExpandBraces(b *testing.B) {
	pattern := "*.{go,js,py,java,cpp,c,h,hpp}"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		expandBraces(pattern)
	}
}

func BenchmarkProcessFilters(b *testing.B) {
	exclude := []string{"*.{log,tmp,bak}", "node_modules", ".git"}
	include := []string{"*.{go,js,py,java}", "*.md"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processFilters(exclude, include)
	}
}

func BenchmarkFindMatchingFiles(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "benchmark_test")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	for i := 0; i < 100; i++ {
		for _, ext := range []string{".go", ".js", ".py", ".md", ".log"} {
			dir := filepath.Join(tempDir, "subdir", "nested")
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				b.Fatal(err)
			}
			file := filepath.Join(dir, fmt.Sprintf("file%d%s", i, ext))
			err = os.WriteFile(file, []byte("content"), 0644)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
	filters := processFilters([]string{"*.log"}, []string{"*.go"})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := findMatchingFiles(tempDir, filters)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBuildTreeOutput(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "tree_benchmark")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	var paths []string
	for i := 0; i < 50; i++ {
		for j := 0; j < 5; j++ {
			dir := filepath.Join(tempDir, fmt.Sprintf("dir%d", i))
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				b.Fatal(err)
			}
			file := filepath.Join(dir, fmt.Sprintf("file%d.go", j))
			err = os.WriteFile(file, []byte("content"), 0644)
			if err != nil {
				b.Fatal(err)
			}
			paths = append(paths, file)
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buildTreeOutput(tempDir, paths)
	}
}
