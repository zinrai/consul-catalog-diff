package main

import (
	"fmt"
	"sort"
	"strings"
)

// outputDiff outputs the diff results
func outputDiff(diff *DiffResult) {
	if !diff.HasChanges() {
		fmt.Println("No differences found")
		return
	}

	fmt.Printf("=== Consul Catalog Diff Report ===\n")
	fmt.Printf("Total changes: %d\n\n", diff.TotalChanges())

	// Output node changes
	if len(diff.NodeAdditions) > 0 || len(diff.NodeModifications) > 0 || len(diff.NodeDeletions) > 0 {
		fmt.Println("NODE CHANGES:")
		outputNodeDiffs(diff)
		fmt.Println()
	}

	// Output service changes
	if len(diff.ServiceAdditions) > 0 || len(diff.ServiceModifications) > 0 || len(diff.ServiceDeletions) > 0 {
		fmt.Println("SERVICE CHANGES:")
		outputServiceDiffs(diff)
		fmt.Println()
	}
}

// outputNodeDiffs outputs node differences
func outputNodeDiffs(diff *DiffResult) {
	// Additions
	if len(diff.NodeAdditions) > 0 {
		outputNodeAdditions(diff.NodeAdditions)
	}

	// Modifications
	if len(diff.NodeModifications) > 0 {
		outputNodeModifications(diff.NodeModifications)
	}

	// Deletions
	if len(diff.NodeDeletions) > 0 {
		outputNodeDeletions(diff.NodeDeletions)
	}
}

// outputNodeAdditions outputs node additions
func outputNodeAdditions(additions []NodeDiff) {
	fmt.Printf("  Additions (%d):\n", len(additions))
	for _, add := range additions {
		fmt.Printf("    + %s", add.Node)
		if addr, ok := add.Expected["Address"].(string); ok {
			fmt.Printf(" [%s]", addr)
		}
		fmt.Println()
		outputNodeDetails(add.Expected, "      ")
	}
}

// outputNodeModifications outputs node modifications
func outputNodeModifications(modifications []NodeDiff) {
	fmt.Printf("  Modifications (%d):\n", len(modifications))
	for _, mod := range modifications {
		fmt.Printf("    ~ %s\n", mod.Node)
		for _, field := range mod.Fields {
			fmt.Printf("      - %s: %v -> %v\n", field.Field, field.Current, field.Expected)
		}
	}
}

// outputNodeDeletions outputs node deletions
func outputNodeDeletions(deletions []NodeDiff) {
	fmt.Printf("  Deletions (%d):\n", len(deletions))
	for _, del := range deletions {
		fmt.Printf("    - %s", del.Node)
		if del.Current != nil {
			fmt.Printf(" [%s]", del.Current.Address)
		}
		fmt.Println()
	}
}

// outputServiceDiffs outputs service differences
func outputServiceDiffs(diff *DiffResult) {
	// Additions
	if len(diff.ServiceAdditions) > 0 {
		outputServiceAdditions(diff.ServiceAdditions)
	}

	// Modifications
	if len(diff.ServiceModifications) > 0 {
		outputServiceModifications(diff.ServiceModifications)
	}

	// Deletions
	if len(diff.ServiceDeletions) > 0 {
		outputServiceDeletions(diff.ServiceDeletions)
	}
}

// outputServiceAdditions outputs service additions
func outputServiceAdditions(additions []ServiceDiff) {
	fmt.Printf("  Additions (%d):\n", len(additions))
	for _, add := range additions {
		fmt.Printf("    + %s/%s", add.Node, add.ServiceID)
		outputServiceSummary(add.Expected, add.ServiceID)
		fmt.Println()
		outputServiceDetails(add.Expected, "      ")
	}
}

// outputServiceModifications outputs service modifications
func outputServiceModifications(modifications []ServiceDiff) {
	fmt.Printf("  Modifications (%d):\n", len(modifications))
	for _, mod := range modifications {
		fmt.Printf("    ~ %s/%s\n", mod.Node, mod.ServiceID)
		for _, field := range mod.Fields {
			fmt.Printf("      - %s: %v -> %v\n", field.Field, field.Current, field.Expected)
		}
	}
}

// outputServiceDeletions outputs service deletions
func outputServiceDeletions(deletions []ServiceDiff) {
	fmt.Printf("  Deletions (%d):\n", len(deletions))
	for _, del := range deletions {
		fmt.Printf("    - %s/%s", del.Node, del.ServiceID)
		if del.Current != nil && del.Current.Service != del.ServiceID {
			fmt.Printf(" (service: %s)", del.Current.Service)
		}
		fmt.Println()
	}
}

// outputServiceSummary outputs a brief summary of service info
func outputServiceSummary(serviceData map[string]interface{}, serviceID string) {
	if svc, ok := serviceData["Service"].(string); ok && svc != serviceID {
		fmt.Printf(" (service: %s)", svc)
	}
	if port, ok := serviceData["Port"]; ok {
		fmt.Printf(" port:%v", port)
	}
}

// outputNodeDetails outputs detailed node information
func outputNodeDetails(nodeData map[string]interface{}, indent string) {
	// Sort keys for consistent output
	keys := make([]string, 0, len(nodeData))
	for k := range nodeData {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		if key == "Node" {
			continue // Already shown
		}
		fmt.Printf("%s%s: %v\n", indent, key, nodeData[key])
	}
}

// outputServiceDetails outputs detailed service information
func outputServiceDetails(serviceData map[string]interface{}, indent string) {
	// Sort keys for consistent output
	keys := make([]string, 0, len(serviceData))
	for k := range serviceData {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		if key == "ID" {
			continue // Already shown
		}

		value := serviceData[key]
		if key == "Tags" {
			outputTagsField(value, indent, key)
			continue
		}

		fmt.Printf("%s%s: %v\n", indent, key, value)
	}
}

// outputTagsField outputs tags field with special formatting
func outputTagsField(value interface{}, indent, key string) {
	arr, ok := value.([]interface{})
	if !ok || len(arr) == 0 {
		fmt.Printf("%s%s: %v\n", indent, key, value)
		return
	}

	fmt.Printf("%s%s: [%s]\n", indent, key, formatStringArray(arr))
}

// formatStringArray formats an array of interfaces as strings
func formatStringArray(arr []interface{}) string {
	strs := make([]string, len(arr))
	for i, v := range arr {
		strs[i] = fmt.Sprint(v)
	}
	return strings.Join(strs, ", ")
}
