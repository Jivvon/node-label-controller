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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NodeLabelPolicyStrategy defines the strategy for selecting nodes
type NodeLabelPolicyStrategy struct {
	// Type specifies the selection strategy type
	// +kubebuilder:validation:Enum=oldest;newest;random
	Type string `json:"type"`

	// Count specifies the number of nodes to select
	// +kubebuilder:validation:Minimum=1
	Count int32 `json:"count"`
}

// NodeLabelPolicySpec defines the desired state of NodeLabelPolicy.
type NodeLabelPolicySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Strategy defines how to select nodes for label application
	Strategy NodeLabelPolicyStrategy `json:"strategy"`

	// Labels defines the labels to be applied to selected nodes
	Labels map[string]string `json:"labels"`
}

// NodeLabelPolicyStatus defines the observed state of NodeLabelPolicy.
type NodeLabelPolicyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// SelectedNodes contains the list of node names that currently have this policy's labels
	SelectedNodes []string `json:"selectedNodes,omitempty"`

	// LastReconcileTime is the timestamp of the last successful reconciliation
	LastReconcileTime *metav1.Time `json:"lastReconcileTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status

// NodeLabelPolicy is the Schema for the nodelabelpolicies API.
type NodeLabelPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeLabelPolicySpec   `json:"spec,omitempty"`
	Status NodeLabelPolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NodeLabelPolicyList contains a list of NodeLabelPolicy.
type NodeLabelPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeLabelPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NodeLabelPolicy{}, &NodeLabelPolicyList{})
}
