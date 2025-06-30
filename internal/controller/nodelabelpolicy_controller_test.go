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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	nlpv1alpha1 "github.com/jivvon/node-label-controller/api/v1alpha1"
	"github.com/jivvon/node-label-controller/internal/constants"
	"github.com/jivvon/node-label-controller/internal/controller/handlers"
	"github.com/jivvon/node-label-controller/internal/external/k8s"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("NodeLabelPolicy Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name: resourceName,
		}
		nodelabelpolicy := &nlpv1alpha1.NodeLabelPolicy{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind NodeLabelPolicy")
			err := k8sClient.Get(ctx, typeNamespacedName, nodelabelpolicy)
			if err != nil && errors.IsNotFound(err) {
				resource := &nlpv1alpha1.NodeLabelPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name: resourceName,
					},
					Spec: nlpv1alpha1.NodeLabelPolicySpec{
						Strategy: nlpv1alpha1.NodeLabelPolicyStrategy{
							Type:  "oldest",
							Count: 1,
						},
						Labels: map[string]string{
							"test-label": "test-value",
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &nlpv1alpha1.NodeLabelPolicy{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance NodeLabelPolicy")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")

			client := k8s.NewClient(k8sClient)
			handler := handlers.NewNodeLabelPolicyHandler(client)

			controllerReconciler := NewNodeLabelPolicyReconciler(
				client,
				handler,
				k8sClient.Scheme(),
			)

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("When deleting a NodeLabelPolicy resource", func() {
		const resourceName = "test-delete-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name: resourceName,
		}
		nodelabelpolicy := &nlpv1alpha1.NodeLabelPolicy{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind NodeLabelPolicy")
			err := k8sClient.Get(ctx, typeNamespacedName, nodelabelpolicy)
			if err != nil && errors.IsNotFound(err) {
				resource := &nlpv1alpha1.NodeLabelPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name: resourceName,
					},
					Spec: nlpv1alpha1.NodeLabelPolicySpec{
						Strategy: nlpv1alpha1.NodeLabelPolicyStrategy{
							Type:  "oldest",
							Count: 1,
						},
						Labels: map[string]string{
							"test-label":    "test-value",
							"managed-label": "managed-value",
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &nlpv1alpha1.NodeLabelPolicy{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			if err == nil {
				By("Cleanup the specific resource instance NodeLabelPolicy")
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			}
		})

		It("should cleanup labels when NodeLabelPolicy is deleted", func() {
			By("First reconciling to apply labels and add finalizer")

			client := k8s.NewClient(k8sClient)
			handler := handlers.NewNodeLabelPolicyHandler(client)

			controllerReconciler := NewNodeLabelPolicyReconciler(
				client,
				handler,
				k8sClient.Scheme(),
			)

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			// check finalizer
			resource := &nlpv1alpha1.NodeLabelPolicy{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			Expect(resource.ObjectMeta.Finalizers).To(ContainElement(constants.FinalizerName))

			// check node label
			nodeList := &corev1.NodeList{}
			Expect(k8sClient.List(ctx, nodeList)).To(Succeed())

			hasLabeledNode := false
			for _, node := range nodeList.Items {
				if node.Labels != nil && node.Labels["test-label"] == "test-value" {
					hasLabeledNode = true
					break
				}
			}
			Expect(hasLabeledNode).To(BeTrue(), "At least one node should have the test label")

			By("Deleting the NodeLabelPolicy")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())

			By("Reconciling during deletion to trigger cleanup")
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Verifying that labels are cleaned up from all nodes")
			Expect(k8sClient.List(ctx, nodeList)).To(Succeed())

			for _, node := range nodeList.Items {
				if node.Labels != nil {
					Expect(node.Labels).NotTo(HaveKey("test-label"), "Node %s should not have test-label", node.Name)
					Expect(node.Labels).NotTo(HaveKey("managed-label"), "Node %s should not have managed-label", node.Name)
					Expect(node.Labels).NotTo(HaveKey(fmt.Sprintf("%s.test-delete-resource/managed-by", constants.ManagedByLabelPrefix)), "Node %s should not have managed-by label", node.Name)
				}
			}
		})
	})
})
