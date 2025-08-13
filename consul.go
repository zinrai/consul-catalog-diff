package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// fetchConsulState fetches the current state from Consul based on operations
func fetchConsulState(consulAddr string, operations []Operation) (*ConsulState, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	state := &ConsulState{
		Nodes:    make(map[string]ConsulNode),
		Services: make(map[string][]ConsulService),
	}

	// Group operations by target to minimize API calls
	nodeOps, serviceOps := groupOperationsByTarget(operations)

	// Fetch nodes that are referenced in operations
	for nodeName := range nodeOps {
		log.Printf("[INFO] Fetching node: %s", nodeName)
		node, err := fetchNode(client, consulAddr, nodeName)
		if err != nil {
			if isNotFoundError(err) {
				log.Printf("[INFO] Node %s not found in Consul", nodeName)
				continue
			}
			return nil, fmt.Errorf("failed to fetch node %s: %w", nodeName, err)
		}
		state.Nodes[nodeName] = *node
	}

	// Fetch services that are referenced in operations
	processedNodes := make(map[string]bool)
	for key := range serviceOps {
		nodeName := getNodeFromServiceKey(key)
		if processedNodes[nodeName] {
			continue
		}
		processedNodes[nodeName] = true

		log.Printf("[INFO] Fetching services for node: %s", nodeName)
		services, err := fetchNodeServices(client, consulAddr, nodeName)
		if err != nil {
			if isNotFoundError(err) {
				log.Printf("[INFO] Node %s not found in Consul", nodeName)
				continue
			}
			return nil, fmt.Errorf("failed to fetch services for node %s: %w", nodeName, err)
		}
		state.Services[nodeName] = services
	}

	return state, nil
}

// fetchNode fetches a single node from Consul
func fetchNode(client *http.Client, consulAddr, nodeName string) (*ConsulNode, error) {
	// First, try to get the node from the nodes list
	u, err := url.Parse(fmt.Sprintf("%s/v1/catalog/nodes", consulAddr))
	if err != nil {
		return nil, fmt.Errorf("invalid consul address: %w", err)
	}

	resp, err := client.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch nodes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("consul returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var nodes []ConsulNode
	if err := json.Unmarshal(body, &nodes); err != nil {
		return nil, fmt.Errorf("failed to parse nodes: %w", err)
	}

	// Find the specific node
	for _, node := range nodes {
		if node.Node == nodeName {
			return &node, nil
		}
	}

	return nil, &notFoundError{resource: "node", name: nodeName}
}

// fetchNodeServices fetches services for a specific node
func fetchNodeServices(client *http.Client, consulAddr, nodeName string) ([]ConsulService, error) {
	u, err := url.Parse(fmt.Sprintf("%s/v1/catalog/node/%s", consulAddr, nodeName))
	if err != nil {
		return nil, fmt.Errorf("invalid consul address: %w", err)
	}

	resp, err := client.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch node services: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, &notFoundError{resource: "node", name: nodeName}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("consul returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var nodeData struct {
		Services map[string]ConsulService `json:"Services"`
	}

	if err := json.Unmarshal(body, &nodeData); err != nil {
		return nil, fmt.Errorf("failed to parse node services: %w", err)
	}

	// Convert map to slice
	var services []ConsulService
	for _, svc := range nodeData.Services {
		services = append(services, svc)
	}

	return services, nil
}

// getNodeFromServiceKey extracts node name from service key
func getNodeFromServiceKey(key string) string {
	idx := strings.Index(key, "/")
	if idx > 0 {
		return key[:idx]
	}
	return key
}

// notFoundError represents a resource not found error
type notFoundError struct {
	resource string
	name     string
}

func (e *notFoundError) Error() string {
	return fmt.Sprintf("%s %s not found", e.resource, e.name)
}

// isNotFoundError checks if an error is a not found error
func isNotFoundError(err error) bool {
	_, ok := err.(*notFoundError)
	return ok
}
