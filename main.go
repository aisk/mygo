package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aisk/mygo/transpiler"
	"golang.org/x/term"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: mygo [options] [files...|folders...]\n")
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  mygo file1.mygo file2.mygo    # Transpile specific files\n")
		fmt.Fprintf(os.Stderr, "  mygo ./folder                 # Transpile all .mygo files in folder\n")
		fmt.Fprintf(os.Stderr, "  mygo ./...                    # Transpile all .mygo files recursively\n")
		fmt.Fprintf(os.Stderr, "  mygo                          # Transpile file from stdin\n")
	}
	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		// Check if stdin is terminal
		if term.IsTerminal(int(os.Stdin.Fd())) {
			// stdin is terminal - print help and exit 0
			flag.Usage()
			return
		}

		// stdin is redirected - process it
		if err := transpiler.Transpile(os.Stdin, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Process arguments: files or folders
	for _, arg := range args {
		if err := processPath(arg); err != nil {
			fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", arg, err)
			os.Exit(1)
		}
	}
}

func processPath(path string) error {
	// Handle Go's ... style recursive pattern
	if strings.HasSuffix(path, "/...") || strings.HasSuffix(path, "\\...") || path == "..." {
		var basePath string
		if path == "..." {
			basePath = "."
		} else {
			basePath = strings.TrimSuffix(path, "/...")
			basePath = strings.TrimSuffix(basePath, "\\...")
		}
		return filepath.Walk(basePath, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(filePath, ".mygo") {
				return transpileFile(filePath)
			}
			return nil
		})
	}

	// Check if path is a directory
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		// Process directory non-recursively (only immediate files)
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".mygo") {
				filePath := filepath.Join(path, entry.Name())
				if err := transpileFile(filePath); err != nil {
					return err
				}
			}
		}
		return nil
	}

	// Process individual file
	if !strings.HasSuffix(path, ".mygo") {
		return fmt.Errorf("file must have .mygo extension: %s", path)
	}
	return transpileFile(path)
}

func transpileFile(inputPath string) error {
	// Generate output path by replacing .mygo with .go
	outputPath := strings.TrimSuffix(inputPath, ".mygo") + ".go"

	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inputFile.Close()

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	if err := transpiler.Transpile(inputFile, outputFile); err != nil {
		return fmt.Errorf("transpilation failed: %w", err)
	}

	fmt.Printf("Transpiled: %s -> %s\n", inputPath, outputPath)
	return nil
}
