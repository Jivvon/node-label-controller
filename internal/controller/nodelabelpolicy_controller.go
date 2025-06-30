/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	nlpv1alpha1 "github.com/jivvon/node-label-controller/api/v1alpha1"
	"github.com/jivvon/node-label-controller/internal/constants"
	"github.com/jivvon/node-label-controller/internal/controller/handlers"
	"github.com/jivvon/node-label-controller/internal/external/k8s"
)

type NodeLabelPolicyReconciler struct {
	client  k8s.Client
	handler handlers.NodeLabelPolicyHandler
	Scheme  *runtime.Scheme
}

// +kubebuilder:rbac:groups=nlp.lento.dev,resources=nodelabelpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nlp.lento.dev,resources=nodelabelpolicies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=nlp.lento.dev,resources=nodelabelpolicies/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch;update;patch

func (r *NodeLabelPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	nodeLabelPolicy := &nlpv1alpha1.NodeLabelPolicy{}
	if err := r.client.Get(ctx, req.NamespacedName, nodeLabelPolicy); err != nil {
		if errors.IsNotFound(err) {
			log.Info("NodeLabelPolicy not found, cleaning up labels from all nodes", "policyName", req.Name)
			if err := r.handler.CleanupLabelsFromAllNodes(ctx, req.Name); err != nil {
				log.Error(err, "Failed to cleanup labels from nodes", "policyName", req.Name)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get NodeLabelPolicy")
		return ctrl.Result{}, err
	}

	finalizerName := constants.FinalizerName

	if nodeLabelPolicy.ObjectMeta.DeletionTimestamp.IsZero() {
		if !containsString(nodeLabelPolicy.ObjectMeta.Finalizers, finalizerName) {
			nodeLabelPolicy.ObjectMeta.Finalizers = append(nodeLabelPolicy.ObjectMeta.Finalizers, finalizerName)
			if err := r.client.Update(ctx, nodeLabelPolicy); err != nil {
				log.Error(err, "Failed to add finalizer")
				return ctrl.Result{}, err
			}
		}
	} else {
		if containsString(nodeLabelPolicy.ObjectMeta.Finalizers, finalizerName) {
			if err := r.handler.CleanupLabelsFromAllNodes(ctx, nodeLabelPolicy.Name); err != nil {
				log.Error(err, "Failed to cleanup labels during deletion")
				return ctrl.Result{}, err
			}
			nodeLabelPolicy.ObjectMeta.Finalizers = removeString(nodeLabelPolicy.ObjectMeta.Finalizers, finalizerName)
			if err := r.client.Update(ctx, nodeLabelPolicy); err != nil {
				log.Error(err, "Failed to remove finalizer")
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	log.Info("Reconciling NodeLabelPolicy", "policyName", nodeLabelPolicy.Name, "strategy", nodeLabelPolicy.Spec.Strategy)

	nodeList := &corev1.NodeList{}
	if err := r.client.List(ctx, nodeList); err != nil {
		log.Error(err, "Failed to list nodes")
		return ctrl.Result{}, err
	}

	selectedNodes, err := r.handler.SelectNodes(ctx, nodeList.Items, nodeLabelPolicy.Spec.Strategy)
	if err != nil {
		log.Error(err, "Failed to select nodes", "strategy", nodeLabelPolicy.Spec.Strategy)
		return ctrl.Result{}, err
	}

	log.V(4).Info("Node selection details",
		"strategy", nodeLabelPolicy.Spec.Strategy.Type,
		"count", nodeLabelPolicy.Spec.Strategy.Count,
		"totalNodes", len(nodeList.Items),
		"selectedNodes", len(selectedNodes))

	for i, node := range selectedNodes {
		log.V(4).Info("Selected node",
			"index", i,
			"nodeName", node.Name,
			"creationTimestamp", node.CreationTimestamp.Time.Format("2006-01-02T15:04:05Z"))
	}

	managedByLabelKey := fmt.Sprintf("%s.%s/managed-by", constants.ManagedByLabelPrefix, nodeLabelPolicy.Name)

	for _, node := range selectedNodes {
		if err := r.handler.ApplyLabelsToNode(ctx, &node, nodeLabelPolicy.Spec.Labels, managedByLabelKey); err != nil {
			log.Error(err, "Failed to apply labels to node", "nodeName", node.Name)
			return ctrl.Result{}, err
		}
	}

	if err := r.handler.RemoveLabelsFromUnselectedNodes(ctx, nodeList.Items, selectedNodes, managedByLabelKey, nodeLabelPolicy.Spec.Labels); err != nil {
		log.Error(err, "Failed to remove labels from unselected nodes")
		return ctrl.Result{}, err
	}

	selectedNodeNames := make([]string, len(selectedNodes))
	for i, node := range selectedNodes {
		selectedNodeNames[i] = node.Name
	}

	if err := r.handler.UpdatePolicyStatus(ctx, nodeLabelPolicy, selectedNodeNames); err != nil {
		log.Error(err, "Failed to update NodeLabelPolicy status")
		return ctrl.Result{}, err
	}

	log.Info("Successfully reconciled NodeLabelPolicy", "policyName", nodeLabelPolicy.Name, "selectedNodes", selectedNodeNames)

	return ctrl.Result{RequeueAfter: constants.ReconcileInterval}, nil
}

func (r *NodeLabelPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&nlpv1alpha1.NodeLabelPolicy{}).
		Watches(&corev1.Node{}, handler.EnqueueRequestsFromMapFunc(r.nodeToNodeLabelPolicy)).
		Named("nodelabelpolicy").
		Complete(r)
}

func (r *NodeLabelPolicyReconciler) nodeToNodeLabelPolicy(ctx context.Context, obj client.Object) []reconcile.Request {
	nodeLabelPolicyList := &nlpv1alpha1.NodeLabelPolicyList{}
	if err := r.client.List(ctx, nodeLabelPolicyList); err != nil {
		return []reconcile.Request{}
	}

	var requests []reconcile.Request
	for _, policy := range nodeLabelPolicyList.Items {
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name: policy.Name,
			},
		})
	}
	return requests
}

func NewNodeLabelPolicyReconciler(client k8s.Client, handler handlers.NodeLabelPolicyHandler, scheme *runtime.Scheme) *NodeLabelPolicyReconciler {
	return &NodeLabelPolicyReconciler{
		client:  client,
		handler: handler,
		Scheme:  scheme,
	}
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) []string {
	result := []string{}
	for _, item := range slice {
		if item != s {
			result = append(result, item)
		}
	}
	return result
}
