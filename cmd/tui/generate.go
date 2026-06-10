package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/senforsce/tndr-tui/internal/tuigen"
)

// runGenerate implements the generate subcommand.
// It processes .t2 files and generates corresponding Go source files.
func runGenerate(args []string) error {
	verbose := false
	var paths []string

	// Parse arguments
	for _, arg := range args {
		if arg == "-v" || arg == "--verbose" {
			verbose = true
		} else {
			paths = append(paths, arg)
		}
	}

	// Default to current directory if no paths specified
	if len(paths) == 0 {
		paths = []string{"."}
	}

	// Collect all .t2 files
	files, err := collectT2Files(paths)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return fmt.Errorf("no .t2 files found")
	}

	if verbose {
		fmt.Printf("Found %d .t2 file(s)\n", len(files))
	}

	// Process each file
	var errorCount int
	for _, inputPath := range files {
		outputPath := outputFileName(inputPath)

		if verbose {
			fmt.Printf("Processing %s -> %s\n", inputPath, outputPath)
		}

		if err := generateFile(inputPath, outputPath); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", inputPath, err)
			errorCount++
			continue
		}
	}

	if errorCount > 0 {
		return fmt.Errorf("%d file(s) had errors", errorCount)
	}

	if verbose {
		fmt.Printf("Successfully generated %d file(s)\n", len(files))
	}

	return nil
}

// collectT2Files finds all .t2 files from the given paths.
// Supports:
//   - Direct file paths: "header.t2"
//   - Directory paths: "./components"
//   - Recursive pattern: "./..."
func collectT2Files(paths []string) ([]string, error) {
	var files []string

	for _, path := range paths {
		// Handle ./... recursive pattern
		if before, ok := strings.CutSuffix(path, "/..."); ok {
			root := before
			if root == "." || root == "" {
				root = "."
			}

			err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if !d.IsDir() && strings.HasSuffix(p, ".t2") {
					files = append(files, p)
				}
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("walking %s: %w", root, err)
			}
			continue
		}

		// Check if path exists
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("stat %s: %w", path, err)
		}

		if info.IsDir() {
			// Collect all .t2 files in directory (non-recursive)
			entries, err := os.ReadDir(path)
			if err != nil {
				return nil, fmt.Errorf("reading directory %s: %w", path, err)
			}
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".t2") {
					files = append(files, filepath.Join(path, entry.Name()))
				}
			}
		} else if strings.HasSuffix(path, ".t2") {
			files = append(files, path)
		}
	}

	return files, nil
}

// outputFileName converts a .t2 filename to its output .go filename.
// Examples:
//
//	header.t2     -> header_t2.go
//	my-app.t2     -> my_app_t2.go
//	components.t2 -> components_t2.go
func outputFileName(inputPath string) string {
	dir := filepath.Dir(inputPath)
	base := filepath.Base(inputPath)

	// Remove .t2 extension
	name := strings.TrimSuffix(base, ".t2")

	// Replace hyphens with underscores (Go doesn't like hyphens in filenames)
	name = strings.ReplaceAll(name, "-", "_")

	// Add _t2.go suffix
	output := name + "_t2.go"

	return filepath.Join(dir, output)
}

// generateFile parses a .t2 file and generates the corresponding Go file.
func generateFile(inputPath, outputPath string) error {
	// Read source file
	source, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	// Get just the filename for error messages and header comment
	filename := filepath.Base(inputPath)

	// Parse source
	lexer := tuigen.NewLexer(filename, string(source))
	parser := tuigen.NewParser(lexer)

	file, err := parser.ParseFile()
	if err != nil {
		return err
	}

	// Analyze (validates and adds missing imports)
	analyzer := tuigen.NewAnalyzer()
	if err := analyzer.Analyze(file); err != nil {
		return err
	}

	// Generate Go code
	generator := tuigen.NewGenerator()
	output, err := generator.Generate(file, filename)
	if err != nil {
		return fmt.Errorf("generating code: %w", err)
	}

	// Write output file
	if err := os.WriteFile(outputPath, output, 0o644); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	return nil
}
