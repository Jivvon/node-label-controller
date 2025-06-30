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

package handlers

import (
	"context"
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	nlpv1alpha1 "github.com/jivvon/node-label-controller/api/v1alpha1"
	"github.com/jivvon/node-label-controller/internal/constants"
)

func TestHandlers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Handlers Suite")
}

type mockClient struct {
}

func (m *mockClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	return nil
}

func (m *mockClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	return nil
}

func (m *mockClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return nil
}

func (m *mockClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	return nil
}

func (m *mockClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	return nil
}

func (m *mockClient) List(ctx context.Context, obj client.ObjectList, opts ...client.ListOption) error {
	return nil
}

func (m *mockClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	return nil
}

func (m *mockClient) RESTMapper() meta.RESTMapper {
	return nil
}

func (m *mockClient) Scheme() *runtime.Scheme {
	return nil
}

func (m *mockClient) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	return schema.GroupVersionKind{}, nil
}

func (m *mockClient) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	return false, nil
}

func (m *mockClient) Status() client.StatusWriter {
	return &mockStatusWriter{}
}

func (m *mockClient) SubResource(subResource string) client.SubResourceClient {
	return nil
}

type mockStatusWriter struct{}

func (m *mockStatusWriter) Create(ctx context.Context, obj client.Object, subResource client.Object, opts ...client.SubResourceCreateOption) error {
	return nil
}

func (m *mockStatusWriter) Update(ctx context.Context, obj client.Object, opts ...client.SubResourceUpdateOption) error {
	return nil
}

func (m *mockStatusWriter) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.SubResourcePatchOption) error {
	return nil
}

var _ = Describe("NodeLabelPolicyHandler", func() {
	var (
		handler NodeLabelPolicyHandler
		ctx     context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		mockK8sClient := &mockClient{}
		handler = NewNodeLabelPolicyHandler(mockK8sClient)
	})

	Describe("SelectNodes", func() {
		var nodes []corev1.Node

		BeforeEach(func() {
			now := metav1.Now()
			oldTime := metav1.Time{Time: now.Add(-24 * time.Hour)}
			newTime := metav1.Time{Time: now.Add(24 * time.Hour)}

			nodes = []corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "node-old",
						CreationTimestamp: oldTime,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "node-new",
						CreationTimestamp: newTime,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "node-middle",
						CreationTimestamp: now,
					},
				},
			}
		})

		It("should select oldest nodes", func() {
			strategy := nlpv1alpha1.NodeLabelPolicyStrategy{
				Type:  "oldest",
				Count: 2,
			}

			selected, err := handler.SelectNodes(ctx, nodes, strategy)
			Expect(err).NotTo(HaveOccurred())
			Expect(selected).To(HaveLen(2))
			Expect(selected[0].Name).To(Equal("node-old"))
			Expect(selected[1].Name).To(Equal("node-middle"))
		})

		It("should select newest nodes", func() {
			strategy := nlpv1alpha1.NodeLabelPolicyStrategy{
				Type:  "newest",
				Count: 2,
			}

			selected, err := handler.SelectNodes(ctx, nodes, strategy)
			Expect(err).NotTo(HaveOccurred())
			Expect(selected).To(HaveLen(2))
			Expect(selected[0].Name).To(Equal("node-new"))
			Expect(selected[1].Name).To(Equal("node-middle"))
		})

		It("should return error for unsupported strategy", func() {
			strategy := nlpv1alpha1.NodeLabelPolicyStrategy{
				Type:  "unsupported",
				Count: 1,
			}

			_, err := handler.SelectNodes(ctx, nodes, strategy)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unsupported strategy type"))
		})

		It("should handle empty node list", func() {
			strategy := nlpv1alpha1.NodeLabelPolicyStrategy{
				Type:  "oldest",
				Count: 1,
			}

			selected, err := handler.SelectNodes(ctx, []corev1.Node{}, strategy)
			Expect(err).NotTo(HaveOccurred())
			Expect(selected).To(HaveLen(0))
		})

		It("should limit selection to available nodes", func() {
			strategy := nlpv1alpha1.NodeLabelPolicyStrategy{
				Type:  "oldest",
				Count: 5,
			}

			selected, err := handler.SelectNodes(ctx, nodes, strategy)
			Expect(err).NotTo(HaveOccurred())
			Expect(selected).To(HaveLen(3))
		})
	})

	Describe("ApplyLabelsToNode", func() {
		var node *corev1.Node
		var labels map[string]string

		BeforeEach(func() {
			node = &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test-node",
					Labels: make(map[string]string),
				},
			}
			labels = map[string]string{
				"environment": "production",
				"workload":    "monitoring",
			}
		})

		It("should apply labels to node", func() {
			managedByLabelKey := fmt.Sprintf("%s.test/managed-by", constants.ManagedByLabelPrefix)
			err := handler.ApplyLabelsToNode(ctx, node, labels, managedByLabelKey)
			Expect(err).NotTo(HaveOccurred())

			Expect(node.Labels["environment"]).To(Equal("production"))
			Expect(node.Labels["workload"]).To(Equal("monitoring"))
			Expect(node.Labels[managedByLabelKey]).To(Equal("true"))
		})

		It("should handle nil labels map", func() {
			node.Labels = nil
			managedByLabelKey := fmt.Sprintf("%s.test/managed-by", constants.ManagedByLabelPrefix)
			err := handler.ApplyLabelsToNode(ctx, node, labels, managedByLabelKey)
			Expect(err).NotTo(HaveOccurred())

			Expect(node.Labels).NotTo(BeNil())
			Expect(node.Labels["environment"]).To(Equal("production"))
		})
	})
})
