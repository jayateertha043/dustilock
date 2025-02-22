package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	"github.com/jayateertha043/dustilock/analysis"
	"log"
	"os"
	"path/filepath"
)

var excludedDirectories = map[string]bool{
	".git":         true,
	"node_modules": true,
}

func main() {
	parser := argparse.NewParser("DustiLock", "a tool to find which of your dependencies is susceptible to Dependency Confusion attack")
	recursive := parser.Flag("r", "recursive", &argparse.Options{Help: "scan all files recursively"})
	audit := parser.Flag("a", "audit", &argparse.Options{Help: "audit only mode"})
	customDir := parser.String("p", "path", &argparse.Options{Help: "the path to scan. default is cwd"})
	err := parser.Parse(os.Args)

	if err != nil {
		fmt.Print(parser.Usage(err))
	}

	auditValue := *audit
	customDirValue := *customDir
	recursiveValue := *recursive

	// -------------

	fmt.Println("DustiLock started")
	scanDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	if customDirValue != "" {
		scanDir, _ = filepath.Abs(customDirValue)
	}

	fmt.Printf("[*] Scanning directory \"%v\" (recursive=%v) ...\n", scanDir, recursiveValue)
	var hasAnyPackageAvailableForRegistration bool

	if recursiveValue {
		hasAnyPackageAvailableForRegistration, err = analysis.AnalyzeDirectoryRecursive(scanDir, excludedDirectories)
	} else {
		hasAnyPackageAvailableForRegistration, err = analysis.AnalyzeDirectory(scanDir)
	}

	if err != nil {
		log.Fatal(err)
	}

	if hasAnyPackageAvailableForRegistration {
		_, _ = fmt.Fprintln(os.Stderr, "[!] One or more packages are available for public registration")
		if !auditValue {
			os.Exit(1)
		}
	} else {
		fmt.Printf("[*] Finished scanning directory \"%v\"\n", scanDir)
	}
}
