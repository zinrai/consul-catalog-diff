package main

import (
	"testing"
)

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected FormatType
	}{
		{
			name: "NDJSON Transaction format",
			input: `{"Node":{"Verb":"set","Node":{"Node":"web-001","Address":"10.0.0.1"}}}
{"Service":{"Verb":"set","Node":"web-001","Service":{"ID":"nginx","Port":80}}}`,
			expected: NDJSONTransactionFormat,
		},
		{
			name: "JSON Transaction Array format",
			input: `[
				{"Node":{"Verb":"set","Node":{"Node":"web-001","Address":"10.0.0.1"}}},
				{"Service":{"Verb":"set","Node":"web-001","Service":{"ID":"nginx","Port":80}}}
			]`,
			expected: JSONTransactionArrayFormat,
		},
		{
			name:     "Empty input",
			input:    "",
			expected: UnknownFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectFormat([]byte(tt.input))
			if result != tt.expected {
				t.Errorf("detectFormat() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCompareNodeFields(t *testing.T) {
	tests := []struct {
		name     string
		expected map[string]interface{}
		current  ConsulNode
		wantDiff int
	}{
		{
			name: "No differences",
			expected: map[string]interface{}{
				"Address":    "10.0.0.1",
				"Datacenter": "dc1",
			},
			current: ConsulNode{
				Address:    "10.0.0.1",
				Datacenter: "dc1",
			},
			wantDiff: 0,
		},
		{
			name: "Address different",
			expected: map[string]interface{}{
				"Address": "10.0.0.2",
			},
			current: ConsulNode{
				Address: "10.0.0.1",
			},
			wantDiff: 1,
		},
		{
			name: "Meta differences",
			expected: map[string]interface{}{
				"Meta": map[string]interface{}{
					"type":     "web",
					"location": "rack-1",
				},
			},
			current: ConsulNode{
				Meta: map[string]string{
					"type":     "web",
					"location": "rack-2",
				},
			},
			wantDiff: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diffs := compareNodeFields(tt.expected, tt.current)
			if len(diffs) != tt.wantDiff {
				t.Errorf("compareNodeFields() returned %d diffs, want %d", len(diffs), tt.wantDiff)
			}
		})
	}
}

func TestStringSlicesEqual(t *testing.T) {
	tests := []struct {
		name string
		a    []string
		b    []string
		want bool
	}{
		{
			name: "Equal slices same order",
			a:    []string{"tag1", "tag2"},
			b:    []string{"tag1", "tag2"},
			want: true,
		},
		{
			name: "Equal slices different order",
			a:    []string{"tag2", "tag1"},
			b:    []string{"tag1", "tag2"},
			want: true,
		},
		{
			name: "Different slices",
			a:    []string{"tag1", "tag2"},
			b:    []string{"tag1", "tag3"},
			want: false,
		},
		{
			name: "Different lengths",
			a:    []string{"tag1"},
			b:    []string{"tag1", "tag2"},
			want: false,
		},
		{
			name: "Empty slices",
			a:    []string{},
			b:    []string{},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stringSlicesEqual(tt.a, tt.b); got != tt.want {
				t.Errorf("stringSlicesEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiffResultHasChanges(t *testing.T) {
	tests := []struct {
		name string
		diff DiffResult
		want bool
	}{
		{
			name: "No changes",
			diff: DiffResult{},
			want: false,
		},
		{
			name: "Node addition",
			diff: DiffResult{
				NodeAdditions: []NodeDiff{{Node: "test"}},
			},
			want: true,
		},
		{
			name: "Service modification",
			diff: DiffResult{
				ServiceModifications: []ServiceDiff{{Node: "test", ServiceID: "web"}},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.diff.HasChanges(); got != tt.want {
				t.Errorf("HasChanges() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseNDJSON(t *testing.T) {
	input := `{"Node":{"Verb":"set","Node":{"Node":"web-001","Address":"10.0.0.1"}}}
{"Service":{"Verb":"set","Node":"web-001","Service":{"ID":"nginx","Port":80}}}`

	ops, err := parseNDJSON([]byte(input))
	if err != nil {
		t.Fatalf("parseNDJSON() error = %v", err)
	}

	if len(ops) != 2 {
		t.Errorf("parseNDJSON() returned %d operations, want 2", len(ops))
	}

	// Check first operation is Node
	if ops[0].Node == nil {
		t.Error("First operation should be a Node operation")
	}

	// Check second operation is Service
	if ops[1].Service == nil {
		t.Error("Second operation should be a Service operation")
	}
}
