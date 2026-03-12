package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	photoorg "github.com/buckelij/sif.io/photo_org"
)

func main() {
	var (
		dir    string
		dryRun bool
		apply  bool
	)

	flag.StringVar(&dir, "dir", ".", "Top-level directory to scan")
	flag.BoolVar(&dryRun, "dry-run", true, "Preview changes without moving files")
	flag.BoolVar(&apply, "apply", false, "Apply changes (equivalent to -dry-run=false)")
	flag.Parse()

	if apply {
		dryRun = false
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error resolving directory: %v\n", err)
		os.Exit(1)
	}

	info, err := os.Stat(absDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading directory %s: %v\n", absDir, err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "not a directory: %s\n", absDir)
		os.Exit(1)
	}

	result, err := photoorg.Organize(absDir, &photoorg.Options{DryRun: dryRun})
	if err != nil {
		fmt.Fprintf(os.Stderr, "organize failed: %v\n", err)
		os.Exit(1)
	}

	mode := "DRY-RUN"
	if !dryRun {
		mode = "APPLY"
	}

	fmt.Printf("mode=%s dir=%s\n", mode, absDir)
	for _, action := range result.Actions {
		state := "PLANNED"
		if action.Moved {
			state = "MOVED"
		}
		fmt.Printf("[%s] %s -> %s\n", state, action.Source, action.Destination)
	}

	fmt.Printf("processed=%d moved=%d undated=%d\n", result.ProcessedFile, result.MovedCount, result.UndatedCount)
}
