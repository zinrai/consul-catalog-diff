package main

import (
	"log"
)

// calculateDiff calculates differences between expected operations and current state
func calculateDiff(operations []Operation, currentState *ConsulState) *DiffResult {
	result := &DiffResult{}

	// Process each operation
	for _, op := range operations {
		if op.Node != nil {
			processNodeOperation(op.Node, currentState, result)
		}
		if op.Service != nil {
			processServiceOperation(op.Service, currentState, result)
		}
		// Note: Check operations not implemented yet
	}

	return result
}

// processNodeOperation processes a single node operation
func processNodeOperation(nodeOp *NodeOperation, state *ConsulState, result *DiffResult) {
	nodeName, nodeData := extractNodeInfo(nodeOp.Node)
	if nodeName == "" {
		log.Printf("[WARN] Node operation missing node name")
		return
	}

	currentNode, exists := state.Nodes[nodeName]

	switch nodeOp.Verb {
	case "set", "cas":
		if !exists {
			// Node doesn't exist - addition
			result.NodeAdditions = append(result.NodeAdditions, NodeDiff{
				Node:     nodeName,
				Expected: nodeData,
				Current:  nil,
			})
		} else {
			// Node exists - check for modifications
			diffs := compareNodeFields(nodeData, currentNode)
			if len(diffs) > 0 {
				result.NodeModifications = append(result.NodeModifications, NodeDiff{
					Node:     nodeName,
					Expected: nodeData,
					Current:  &currentNode,
					Fields:   diffs,
				})
			}
		}

	case "delete":
		if exists {
			// Node exists and should be deleted
			result.NodeDeletions = append(result.NodeDeletions, NodeDiff{
				Node:     nodeName,
				Expected: nodeData,
				Current:  &currentNode,
			})
		}
		// If node doesn't exist, nothing to delete (already in desired state)
	}
}

// processServiceOperation processes a single service operation
func processServiceOperation(serviceOp *ServiceOperation, state *ConsulState, result *DiffResult) {
	nodeName, serviceID, serviceData := extractServiceInfo(serviceOp)
	if nodeName == "" || serviceID == "" {
		log.Printf("[WARN] Service operation missing node name or service ID")
		return
	}

	// Find current service
	currentService := findCurrentService(state, nodeName, serviceID)

	switch serviceOp.Verb {
	case "set", "cas":
		processServiceSetOperation(currentService, nodeName, serviceID, serviceData, result)
	case "delete":
		processServiceDeleteOperation(currentService, nodeName, serviceID, serviceData, result)
	}
}

// findCurrentService finds a service in the current state
func findCurrentService(state *ConsulState, nodeName, serviceID string) *ConsulService {
	services, ok := state.Services[nodeName]
	if !ok {
		return nil
	}

	for _, svc := range services {
		if svc.ID == serviceID {
			return &svc
		}
	}

	return nil
}

// processServiceSetOperation processes set/cas operations for services
func processServiceSetOperation(currentService *ConsulService, nodeName, serviceID string, serviceData map[string]interface{}, result *DiffResult) {
	if currentService == nil {
		// Service doesn't exist - addition
		result.ServiceAdditions = append(result.ServiceAdditions, ServiceDiff{
			Node:      nodeName,
			ServiceID: serviceID,
			Expected:  serviceData,
			Current:   nil,
		})
		return
	}

	// Service exists - check for modifications
	diffs := compareServiceFields(serviceData, *currentService)
	if len(diffs) > 0 {
		result.ServiceModifications = append(result.ServiceModifications, ServiceDiff{
			Node:      nodeName,
			ServiceID: serviceID,
			Expected:  serviceData,
			Current:   currentService,
			Fields:    diffs,
		})
	}
}

// processServiceDeleteOperation processes delete operations for services
func processServiceDeleteOperation(currentService *ConsulService, nodeName, serviceID string, serviceData map[string]interface{}, result *DiffResult) {
	if currentService == nil {
		// Service doesn't exist, nothing to delete (already in desired state)
		return
	}

	// Service exists and should be deleted
	result.ServiceDeletions = append(result.ServiceDeletions, ServiceDiff{
		Node:      nodeName,
		ServiceID: serviceID,
		Expected:  serviceData,
		Current:   currentService,
	})
}
