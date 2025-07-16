package utils

import (
	corev1 "k8s.io/api/core/v1"
)

// IsNodeReady checks if a node is in Ready state
// Returns true if the node has a Ready condition with status True
func IsNodeReady(node *corev1.Node) bool {
	if node == nil {
		return false
	}

	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			return condition.Status == corev1.ConditionTrue
		}
	}

	// If no Ready condition is found, consider the node as not ready
	return false
}

// FilterReadyNodes filters a slice of nodes to include only Ready nodes
func FilterReadyNodes(nodes []corev1.Node) []corev1.Node {
	var readyNodes []corev1.Node

	for _, node := range nodes {
		if IsNodeReady(&node) {
			readyNodes = append(readyNodes, node)
		}
	}

	return readyNodes
}
