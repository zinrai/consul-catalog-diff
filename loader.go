package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

// loadOperations loads operations from a file
func loadOperations(filename string) ([]Operation, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read file content
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Detect format and parse
	format := detectFormat(data)
	log.Printf("[INFO] Detected format: %s", formatString(format))

	switch format {
	case NDJSONTransactionFormat:
		return parseNDJSON(data)
	case NDJSONBatchFormat:
		return parseBatchNDJSON(data)
	case JSONTransactionArrayFormat:
		return parseTransactionArrayJSON(data)
	case JSONCatalogNodeFormat, JSONCatalogServiceFormat:
		return nil, fmt.Errorf("catalog format not yet supported for diff operations")
	default:
		return nil, fmt.Errorf("unable to detect file format")
	}
}

// detectFormat detects the format of the input data
func detectFormat(data []byte) FormatType {
	// Try to detect NDJSON first
	if format := detectNDJSONFormat(data); format != UnknownFormat {
		return format
	}

	// Try to parse as single JSON
	return detectSingleJSONFormat(data)
}

// detectNDJSONFormat checks if data is newline-delimited JSON objects.
// Two line shapes are recognised:
//   - batch envelopes {"batch":N,"operations":[...]} (consul-catalog-sync -payload)
//   - bare operations  {"Node":{"Verb":...}} / {"Service":{"Verb":...}}
//
// A leading '[' means a JSON array, which single-JSON detection handles instead.
func detectNDJSONFormat(data []byte) FormatType {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || trimmed[0] == '[' {
		return UnknownFormat
	}

	lines := bytes.Split(trimmed, []byte("\n"))
	sawBatch := false
	sawOperation := false

	for _, line := range lines {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}

		var obj map[string]interface{}
		if err := json.Unmarshal(line, &obj); err != nil {
			return UnknownFormat
		}

		switch {
		case hasOperationsArray(obj):
			sawBatch = true
		case containsVerb(obj):
			sawOperation = true
		default:
			return UnknownFormat
		}
	}

	switch {
	case sawBatch && !sawOperation:
		return NDJSONBatchFormat
	case sawOperation && !sawBatch:
		return NDJSONTransactionFormat
	default:
		return UnknownFormat
	}
}

// hasOperationsArray reports whether obj is a batch envelope carrying an
// "operations" array (the per-batch object consul-catalog-sync -payload emits).
func hasOperationsArray(obj map[string]interface{}) bool {
	_, ok := obj["operations"].([]interface{})
	return ok
}

// detectSingleJSONFormat checks single JSON format
func detectSingleJSONFormat(data []byte) FormatType {
	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return UnknownFormat
	}

	switch v := jsonData.(type) {
	case map[string]interface{}:
		return detectJSONObjectFormat(v)
	case []interface{}:
		return detectJSONArrayFormat(v)
	default:
		return UnknownFormat
	}
}

// detectJSONObjectFormat detects format for JSON object
func detectJSONObjectFormat(v map[string]interface{}) FormatType {
	// Check for node format
	if _, hasNode := v["Node"]; hasNode {
		return JSONCatalogNodeFormat
	}

	return UnknownFormat
}

// detectJSONArrayFormat detects format for JSON array
func detectJSONArrayFormat(v []interface{}) FormatType {
	if len(v) == 0 {
		return UnknownFormat
	}

	first, ok := v[0].(map[string]interface{})
	if !ok {
		return UnknownFormat
	}

	// Check if it's an array of operations (Transaction format)
	if containsVerb(first) {
		return JSONTransactionArrayFormat
	}

	// Check for catalog node format
	if _, hasAddress := first["Address"]; hasAddress {
		return JSONCatalogNodeFormat
	}

	// Check for catalog service format
	if _, hasService := first["Service"]; hasService {
		return JSONCatalogServiceFormat
	}

	return UnknownFormat
}

// parseNDJSON parses NDJSON format
func parseNDJSON(data []byte) ([]Operation, error) {
	var operations []Operation
	scanner := bufio.NewScanner(bytes.NewReader(data))

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}

		var op Operation
		if err := json.Unmarshal(line, &op); err != nil {
			return nil, fmt.Errorf("failed to parse line %d: %w", lineNum, err)
		}
		operations = append(operations, op)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read NDJSON: %w", err)
	}

	return operations, nil
}

// parseBatchNDJSON flattens batch-enveloped NDJSON, where each line is a
// {"operations":[...]} object as written by consul-catalog-sync -payload.
func parseBatchNDJSON(data []byte) ([]Operation, error) {
	var operations []Operation
	scanner := bufio.NewScanner(bytes.NewReader(data))
	// A batch line carries up to 64 operations, so raise the line limit
	// well above bufio's default 64 KiB token size.
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}

		var batch struct {
			Operations []Operation `json:"operations"`
		}
		if err := json.Unmarshal(line, &batch); err != nil {
			return nil, fmt.Errorf("failed to parse batch line %d: %w", lineNum, err)
		}
		operations = append(operations, batch.Operations...)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read batch NDJSON: %w", err)
	}

	return operations, nil
}

// parseTransactionArrayJSON parses JSON array of transaction operations
func parseTransactionArrayJSON(data []byte) ([]Operation, error) {
	var operations []Operation

	if err := json.Unmarshal(data, &operations); err != nil {
		return nil, fmt.Errorf("failed to parse transaction array JSON: %w", err)
	}

	return operations, nil
}

// containsVerb checks if the object contains a Verb field
func containsVerb(obj map[string]interface{}) bool {
	for _, v := range obj {
		if nested, ok := v.(map[string]interface{}); ok {
			if _, hasVerb := nested["Verb"]; hasVerb {
				return true
			}
		}
	}
	return false
}

// formatString returns a human-readable format name
func formatString(format FormatType) string {
	switch format {
	case NDJSONTransactionFormat:
		return "NDJSON Transaction format"
	case NDJSONBatchFormat:
		return "NDJSON batch format"
	case JSONTransactionArrayFormat:
		return "JSON Transaction Array format"
	case JSONCatalogNodeFormat:
		return "JSON Catalog Node format"
	case JSONCatalogServiceFormat:
		return "JSON Catalog Service format"
	default:
		return "Unknown format"
	}
}
