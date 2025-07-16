package handlers

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	nlpv1alpha1 "github.com/jivvon/node-label-controller/api/v1alpha1"
	"github.com/jivvon/node-label-controller/internal/constants"
	"github.com/jivvon/node-label-controller/internal/external/k8s"
	"github.com/jivvon/node-label-controller/internal/utils"
)

const (
	managedByLabelValue = "true"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

// NodeLabelPolicyHandler handles the business logic for NodeLabelPolicy reconciliation
//
//counterfeiter:generate . NodeLabelPolicyHandler
type NodeLabelPolicyHandler interface {
	// SelectNodes selects nodes based on the given strategy
	SelectNodes(ctx context.Context, nodes []corev1.Node, strategy nlpv1alpha1.NodeLabelPolicyStrategy) ([]corev1.Node, error)

	// ApplyLabelsToNode applies labels to a specific node
	ApplyLabelsToNode(ctx context.Context, node *corev1.Node, labels map[string]string, managedByLabelKey string) error

	// RemoveLabelsFromUnselectedNodes removes labels from nodes that are not selected
	RemoveLabelsFromUnselectedNodes(ctx context.Context, allNodes []corev1.Node, selectedNodes []corev1.Node, managedByLabelKey string, policyLabels map[string]string) error

	// CleanupLabelsFromAllNodes removes all labels related to a policy from all nodes
	CleanupLabelsFromAllNodes(ctx context.Context, policyName string, policyLabels map[string]string) error

	// UpdatePolicyStatus updates the status of a NodeLabelPolicy
	UpdatePolicyStatus(ctx context.Context, policy *nlpv1alpha1.NodeLabelPolicy, selectedNodeNames []string) error
}

type nodeLabelPolicyHandler struct {
	client k8s.Client
}

// NewNodeLabelPolicyHandler creates a new NodeLabelPolicyHandler
func NewNodeLabelPolicyHandler(client k8s.Client) NodeLabelPolicyHandler {
	return &nodeLabelPolicyHandler{
		client: client,
	}
}

// SelectNodes selects nodes based on the given strategy
func (h *nodeLabelPolicyHandler) SelectNodes(ctx context.Context, nodes []corev1.Node, strategy nlpv1alpha1.NodeLabelPolicyStrategy) ([]corev1.Node, error) {
	if len(nodes) == 0 {
		return []corev1.Node{}, nil
	}

	// Filter to only include Ready nodes
	readyNodes := utils.FilterReadyNodes(nodes)
	if len(readyNodes) == 0 {
		return []corev1.Node{}, nil
	}

	nodeCopies := make([]corev1.Node, len(readyNodes))
	copy(nodeCopies, readyNodes)

	switch strategy.Type {
	case "oldest":
		sort.Slice(nodeCopies, func(i, j int) bool {
			return nodeCopies[i].CreationTimestamp.Before(&nodeCopies[j].CreationTimestamp)
		})
	case "newest":
		sort.Slice(nodeCopies, func(i, j int) bool {
			return nodeCopies[j].CreationTimestamp.Before(&nodeCopies[i].CreationTimestamp)
		})
	case "random":
		rand.Shuffle(len(nodeCopies), func(i, j int) {
			nodeCopies[i], nodeCopies[j] = nodeCopies[j], nodeCopies[i]
		})
	default:
		return nil, fmt.Errorf("unsupported strategy type: %s", strategy.Type)
	}

	count := int(strategy.Count)
	if count > len(nodeCopies) {
		count = len(nodeCopies)
	}

	return nodeCopies[:count], nil
}

// ApplyLabelsToNode applies labels to a specific node
func (h *nodeLabelPolicyHandler) ApplyLabelsToNode(ctx context.Context, node *corev1.Node, labels map[string]string, managedByLabelKey string) error {
	if node.Labels == nil {
		node.Labels = make(map[string]string)
	}

	for key, value := range labels {
		node.Labels[key] = value
	}

	node.Labels[managedByLabelKey] = managedByLabelValue

	if err := h.client.Update(ctx, node); err != nil {
		return fmt.Errorf("failed to update node %s: %w", node.Name, err)
	}

	return nil
}

// RemoveLabelsFromUnselectedNodes removes labels from nodes that are not selected
func (h *nodeLabelPolicyHandler) RemoveLabelsFromUnselectedNodes(ctx context.Context, allNodes []corev1.Node, selectedNodes []corev1.Node, managedByLabelKey string, policyLabels map[string]string) error {
	selectedNodeNames := make(map[string]bool)
	for _, node := range selectedNodes {
		selectedNodeNames[node.Name] = true
	}

	for _, node := range allNodes {
		if selectedNodeNames[node.Name] {
			continue
		}

		if node.Labels != nil && node.Labels[managedByLabelKey] == managedByLabelValue {
			nodeCopy := node.DeepCopy()

			delete(nodeCopy.Labels, managedByLabelKey)

			for key := range policyLabels {
				delete(nodeCopy.Labels, key)
			}

			if err := h.client.Update(ctx, nodeCopy); err != nil {
				return fmt.Errorf("failed to remove labels from node %s: %w", node.Name, err)
			}
		}
	}

	return nil
}

// CleanupLabelsFromAllNodes removes all labels related to a policy from all nodes
// If policyLabels is nil, only managed-by and policy-prefix labels are removed
func (h *nodeLabelPolicyHandler) CleanupLabelsFromAllNodes(ctx context.Context, policyName string, policyLabels map[string]string) error {
	nodeList := &corev1.NodeList{}
	if err := h.client.List(ctx, nodeList); err != nil {
		return fmt.Errorf("failed to list nodes: %w", err)
	}

	managedByLabelKey := fmt.Sprintf("%s.%s/managed-by", constants.ManagedByLabelPrefix, policyName)
	policyLabelPrefix := fmt.Sprintf("%s.%s/", constants.ManagedByLabelPrefix, policyName)

	for _, node := range nodeList.Items {
		if node.Labels != nil && node.Labels[managedByLabelKey] == managedByLabelValue {
			nodeCopy := node.DeepCopy()

			// Remove the managed-by label
			delete(nodeCopy.Labels, managedByLabelKey)

			// Remove any labels with policy-specific prefix
			for key := range nodeCopy.Labels {
				if strings.HasPrefix(key, policyLabelPrefix) {
					delete(nodeCopy.Labels, key)
				}
			}

			// Remove policy-specific labels (safe with nil map)
			for key := range policyLabels {
				delete(nodeCopy.Labels, key)
			}

			if err := h.client.Update(ctx, nodeCopy); err != nil {
				return fmt.Errorf("failed to cleanup labels from node %s: %w", node.Name, err)
			}
		}
	}

	return nil
}

// UpdatePolicyStatus updates the status of a NodeLabelPolicy
func (h *nodeLabelPolicyHandler) UpdatePolicyStatus(ctx context.Context, policy *nlpv1alpha1.NodeLabelPolicy, selectedNodeNames []string) error {
	policy.Status.SelectedNodes = selectedNodeNames
	policy.Status.LastReconcileTime = &metav1.Time{Time: metav1.Now().Time}

	if err := h.client.Status().Update(ctx, policy); err != nil {
		return fmt.Errorf("failed to update NodeLabelPolicy status: %w", err)
	}

	return nil
}
