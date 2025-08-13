package main

import (
	"fmt"
)

// extractNodeInfo extracts node information from the operation
func extractNodeInfo(nodeOp map[string]interface{}) (string, map[string]interface{}) {
	nodeName := ""
	if name, ok := nodeOp["Node"].(string); ok {
		nodeName = name
	}
	return nodeName, nodeOp
}

// extractServiceInfo extracts service information from the operation
func extractServiceInfo(serviceOp *ServiceOperation) (string, string, map[string]interface{}) {
	nodeName := serviceOp.Node
	serviceID := ""

	if svc, ok := serviceOp.Service["ID"].(string); ok {
		serviceID = svc
	} else if svc, ok := serviceOp.Service["Service"].(string); ok {
		serviceID = svc
	}

	return nodeName, serviceID, serviceOp.Service
}

// groupOperationsByTarget groups operations by their target (node/service)
func groupOperationsByTarget(operations []Operation) (map[string]*NodeOperation, map[string]*ServiceOperation) {
	nodes := make(map[string]*NodeOperation)
	services := make(map[string]*ServiceOperation)

	for _, op := range operations {
		if op.Node != nil {
			nodeName, _ := extractNodeInfo(op.Node.Node)
			if nodeName != "" {
				nodes[nodeName] = op.Node
			}
		}

		if op.Service != nil {
			nodeName, serviceID, _ := extractServiceInfo(op.Service)
			key := fmt.Sprintf("%s/%s", nodeName, serviceID)
			services[key] = op.Service
		}
	}

	return nodes, services
}

// normalizeValue converts interface{} values to comparable types
func normalizeValue(v interface{}) interface{} {
	switch val := v.(type) {
	case float64:
		// JSON numbers are parsed as float64, but we might want int
		if val == float64(int(val)) {
			return int(val)
		}
		return val
	case []interface{}:
		// Normalize arrays
		return normalizeArray(val)
	case map[string]interface{}:
		// Normalize nested maps
		return normalizeMap(val)
	default:
		return v
	}
}

// normalizeArray normalizes an array of values
func normalizeArray(arr []interface{}) []interface{} {
	result := make([]interface{}, len(arr))
	for i, item := range arr {
		result[i] = normalizeValue(item)
	}
	return result
}

// normalizeMap normalizes a map of values
func normalizeMap(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		result[k] = normalizeValue(v)
	}
	return result
}

// compareNodeFields compares node fields and returns differences
func compareNodeFields(expected map[string]interface{}, current ConsulNode) []FieldDiff {
	var diffs []FieldDiff

	// Compare Address
	if addr, ok := expected["Address"].(string); ok && addr != current.Address {
		diffs = append(diffs, FieldDiff{
			Field:    "Address",
			Expected: addr,
			Current:  current.Address,
		})
	}

	// Compare Datacenter
	if dc, ok := expected["Datacenter"].(string); ok && dc != current.Datacenter {
		diffs = append(diffs, FieldDiff{
			Field:    "Datacenter",
			Expected: dc,
			Current:  current.Datacenter,
		})
	}

	// Compare Meta
	diffs = append(diffs, compareNodeMeta(expected, current)...)

	// Compare TaggedAddresses
	diffs = append(diffs, compareNodeTaggedAddresses(expected, current)...)

	return diffs
}

// compareNodeMeta compares node metadata fields
func compareNodeMeta(expected map[string]interface{}, current ConsulNode) []FieldDiff {
	var diffs []FieldDiff

	expectedMeta, ok := expected["Meta"].(map[string]interface{})
	if !ok {
		return diffs
	}

	for key, expectedVal := range expectedMeta {
		currentVal := current.Meta[key]
		if fmt.Sprint(expectedVal) != currentVal {
			diffs = append(diffs, FieldDiff{
				Field:    fmt.Sprintf("Meta.%s", key),
				Expected: expectedVal,
				Current:  currentVal,
			})
		}
	}

	return diffs
}

// compareNodeTaggedAddresses compares node tagged addresses
func compareNodeTaggedAddresses(expected map[string]interface{}, current ConsulNode) []FieldDiff {
	var diffs []FieldDiff

	expectedTA, ok := expected["TaggedAddresses"].(map[string]interface{})
	if !ok {
		return diffs
	}

	for key, expectedVal := range expectedTA {
		currentVal := current.TaggedAddresses[key]
		if fmt.Sprint(expectedVal) != currentVal {
			diffs = append(diffs, FieldDiff{
				Field:    fmt.Sprintf("TaggedAddresses.%s", key),
				Expected: expectedVal,
				Current:  currentVal,
			})
		}
	}

	return diffs
}

// compareServiceFields compares service fields and returns differences
func compareServiceFields(expected map[string]interface{}, current ConsulService) []FieldDiff {
	var diffs []FieldDiff

	// Compare Service name
	if svc, ok := expected["Service"].(string); ok && svc != current.Service {
		diffs = append(diffs, FieldDiff{
			Field:    "Service",
			Expected: svc,
			Current:  current.Service,
		})
	}

	// Compare Port
	if port, ok := normalizeValue(expected["Port"]).(int); ok && port != current.Port {
		diffs = append(diffs, FieldDiff{
			Field:    "Port",
			Expected: port,
			Current:  current.Port,
		})
	}

	// Compare Address
	if addr, ok := expected["Address"].(string); ok && addr != current.Address {
		diffs = append(diffs, FieldDiff{
			Field:    "Address",
			Expected: addr,
			Current:  current.Address,
		})
	}

	// Compare Tags
	if tagDiff := compareServiceTags(expected, current); tagDiff != nil {
		diffs = append(diffs, *tagDiff)
	}

	// Compare Meta
	diffs = append(diffs, compareServiceMeta(expected, current)...)

	return diffs
}

// compareServiceMeta compares service metadata fields
func compareServiceMeta(expected map[string]interface{}, current ConsulService) []FieldDiff {
	var diffs []FieldDiff

	expectedMeta, ok := expected["Meta"].(map[string]interface{})
	if !ok {
		return diffs
	}

	for key, expectedVal := range expectedMeta {
		currentVal := current.Meta[key]
		if fmt.Sprint(expectedVal) != currentVal {
			diffs = append(diffs, FieldDiff{
				Field:    fmt.Sprintf("Meta.%s", key),
				Expected: expectedVal,
				Current:  currentVal,
			})
		}
	}

	return diffs
}

// compareServiceTags compares service tags and returns a FieldDiff if different
func compareServiceTags(expected map[string]interface{}, current ConsulService) *FieldDiff {
	expectedTags, ok := expected["Tags"].([]interface{})
	if !ok {
		return nil
	}

	expectedTagStrs := make([]string, len(expectedTags))
	for i, tag := range expectedTags {
		expectedTagStrs[i] = fmt.Sprint(tag)
	}

	if !stringSlicesEqual(expectedTagStrs, current.Tags) {
		return &FieldDiff{
			Field:    "Tags",
			Expected: expectedTagStrs,
			Current:  current.Tags,
		}
	}

	return nil
}

// stringSlicesEqual compares two string slices
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	// Create maps for comparison (order doesn't matter for tags)
	aMap := make(map[string]bool)
	for _, s := range a {
		aMap[s] = true
	}
	for _, s := range b {
		if !aMap[s] {
			return false
		}
	}
	return true
}
