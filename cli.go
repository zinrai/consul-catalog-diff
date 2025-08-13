package main

import (
	"flag"
	"fmt"
	"os"
)

// Config holds command-line configuration
type Config struct {
	File       string
	ConsulAddr string
}

func parseConfig() Config {
	var config Config

	flag.StringVar(&config.File, "file", "", "JSON/NDJSON file containing expected operations (required)")
	flag.StringVar(&config.ConsulAddr, "consul-addr", "http://127.0.0.1:8500", "Consul HTTP address")

	// Handle special flags before parsing
	if handleSpecialFlags() {
		os.Exit(0)
	}

	flag.Parse()

	// Validate required flags
	if config.File == "" {
		fmt.Fprintf(os.Stderr, "Error: -file flag is required\n\n")
		showUsage()
		os.Exit(2)
	}

	return config
}

// handleSpecialFlags handles version and help flags
func handleSpecialFlags() bool {
	if len(os.Args) <= 1 {
		return false
	}

	switch os.Args[1] {
	case "-version", "--version":
		fmt.Printf("%s version %s\n", binaryName, version)
		return true
	case "-help", "--help", "-h":
		showUsage()
		return true
	}

	return false
}

func showUsage() {
	fmt.Fprintf(os.Stderr, "%s - Detect differences between JSON operations and Consul Catalog\n\n", binaryName)
	fmt.Fprintf(os.Stderr, "Version: %s\n\n", version)
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  %s -file <path> [options]\n\n", binaryName)
	fmt.Fprintf(os.Stderr, "Required flags:\n")
	fmt.Fprintf(os.Stderr, "  -file        Path to JSON/NDJSON file containing expected operations\n\n")
	fmt.Fprintf(os.Stderr, "Optional flags:\n")
	fmt.Fprintf(os.Stderr, "  -consul-addr Consul HTTP address (default: http://127.0.0.1:8500)\n")
	fmt.Fprintf(os.Stderr, "  -version     Show version\n")
	fmt.Fprintf(os.Stderr, "  -help        Show this help message\n")
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  # Check differences from file\n")
	fmt.Fprintf(os.Stderr, "  %s -file operations.json -consul-addr http://consul:8500\n\n", binaryName)
	fmt.Fprintf(os.Stderr, "  # Use process substitution\n")
	fmt.Fprintf(os.Stderr, "  %s -file <(consul-catalog-sync -payload) -consul-addr http://consul:8500\n\n", binaryName)
	fmt.Fprintf(os.Stderr, "Exit codes:\n")
	fmt.Fprintf(os.Stderr, "  0 - No differences found\n")
	fmt.Fprintf(os.Stderr, "  1 - Differences found\n")
	fmt.Fprintf(os.Stderr, "  2 - Error occurred\n")
}
