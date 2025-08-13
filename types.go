package main

// Operation represents a single Consul operation
type Operation struct {
	Node    *NodeOperation    `json:"Node,omitempty"`
	Service *ServiceOperation `json:"Service,omitempty"`
	Check   *CheckOperation   `json:"Check,omitempty"`
}

// NodeOperation represents a node operation
type NodeOperation struct {
	Verb string                 `json:"Verb"`
	Node map[string]interface{} `json:"Node"`
}

// ServiceOperation represents a service operation
type ServiceOperation struct {
	Verb    string                 `json:"Verb"`
	Node    string                 `json:"Node"`
	Service map[string]interface{} `json:"Service"`
}

// CheckOperation represents a check operation
type CheckOperation struct {
	Verb  string                 `json:"Verb"`
	Node  string                 `json:"Node"`
	Check map[string]interface{} `json:"Check"`
}

// ConsulState represents the current state in Consul
type ConsulState struct {
	Nodes    map[string]ConsulNode
	Services map[string][]ConsulService
}

// ConsulNode represents a node in Consul
type ConsulNode struct {
	Node            string            `json:"Node"`
	Address         string            `json:"Address"`
	Datacenter      string            `json:"Datacenter"`
	TaggedAddresses map[string]string `json:"TaggedAddresses"`
	Meta            map[string]string `json:"Meta"`
	CreateIndex     uint64            `json:"CreateIndex"`
	ModifyIndex     uint64            `json:"ModifyIndex"`
}

// ConsulService represents a service in Consul
type ConsulService struct {
	ID                string                 `json:"ID"`
	Service           string                 `json:"Service"`
	Tags              []string               `json:"Tags"`
	Port              int                    `json:"Port"`
	Address           string                 `json:"Address"`
	TaggedAddresses   map[string]interface{} `json:"TaggedAddresses"`
	Meta              map[string]string      `json:"Meta"`
	EnableTagOverride bool                   `json:"EnableTagOverride"`
	CreateIndex       uint64                 `json:"CreateIndex"`
	ModifyIndex       uint64                 `json:"ModifyIndex"`
}

// DiffResult represents the differences found
type DiffResult struct {
	NodeAdditions        []NodeDiff
	NodeModifications    []NodeDiff
	NodeDeletions        []NodeDiff
	ServiceAdditions     []ServiceDiff
	ServiceModifications []ServiceDiff
	ServiceDeletions     []ServiceDiff
}

// NodeDiff represents a node difference
type NodeDiff struct {
	Node     string
	Expected map[string]interface{}
	Current  *ConsulNode
	Fields   []FieldDiff // For modifications
}

// ServiceDiff represents a service difference
type ServiceDiff struct {
	Node      string
	ServiceID string
	Expected  map[string]interface{}
	Current   *ConsulService
	Fields    []FieldDiff // For modifications
}

// FieldDiff represents a field-level difference
type FieldDiff struct {
	Field    string
	Expected interface{}
	Current  interface{}
}

// FormatType represents the input file format
type FormatType int

const (
	UnknownFormat FormatType = iota
	NDJSONTransactionFormat
	JSONTransactionArrayFormat
	JSONCatalogNodeFormat
	JSONCatalogServiceFormat
)

// HasChanges returns true if there are any differences
func (d *DiffResult) HasChanges() bool {
	return len(d.NodeAdditions) > 0 ||
		len(d.NodeModifications) > 0 ||
		len(d.NodeDeletions) > 0 ||
		len(d.ServiceAdditions) > 0 ||
		len(d.ServiceModifications) > 0 ||
		len(d.ServiceDeletions) > 0
}

// TotalChanges returns the total number of changes
func (d *DiffResult) TotalChanges() int {
	return len(d.NodeAdditions) +
		len(d.NodeModifications) +
		len(d.NodeDeletions) +
		len(d.ServiceAdditions) +
		len(d.ServiceModifications) +
		len(d.ServiceDeletions)
}
