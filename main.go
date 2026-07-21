package main

import (
	"log"
	"os"
)

// version, commit and date are injected at release time by goreleaser
// via -ldflags -X main.<name> (see .goreleaser.yaml).
var (
	version    = "dev"
	commit     = "none"
	date       = "unknown"
	binaryName = "consul-catalog-diff"
)

func main() {
	// Parse command line arguments
	config := parseConfig()
	setupLogging(config)

	// Load and parse input file
	operations, err := loadOperations(config.File)
	if err != nil {
		log.Fatalf("[ERROR] Failed to load operations: %v", err)
	}

	// Fetch current state from Consul
	currentState, err := fetchConsulState(config.ConsulAddr, operations)
	if err != nil {
		log.Fatalf("[ERROR] Failed to fetch Consul state: %v", err)
	}

	// Calculate differences
	diff := calculateDiff(operations, currentState)

	// Output results
	outputDiff(diff)

	// Set exit code based on differences
	if diff.HasChanges() {
		os.Exit(1) // Differences found
	}
	os.Exit(0) // No differences
}

func setupLogging(config Config) {
	log.SetFlags(0)
}
