package utils

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Utils Suite")
}

var _ = Describe("Node Utils", func() {
	Describe("IsNodeReady", func() {
		It("should return true for a ready node", func() {
			node := &corev1.Node{
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{
						{
							Type:   corev1.NodeReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
			}

			Expect(IsNodeReady(node)).To(BeTrue())
		})

		It("should return false for a not ready node", func() {
			node := &corev1.Node{
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{
						{
							Type:   corev1.NodeReady,
							Status: corev1.ConditionFalse,
						},
					},
				},
			}

			Expect(IsNodeReady(node)).To(BeFalse())
		})

		It("should return false for a node with unknown ready status", func() {
			node := &corev1.Node{
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{
						{
							Type:   corev1.NodeReady,
							Status: corev1.ConditionUnknown,
						},
					},
				},
			}

			Expect(IsNodeReady(node)).To(BeFalse())
		})

		It("should return false for a node without ready condition", func() {
			node := &corev1.Node{
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{
						{
							Type:   corev1.NodeMemoryPressure,
							Status: corev1.ConditionFalse,
						},
					},
				},
			}

			Expect(IsNodeReady(node)).To(BeFalse())
		})

		It("should return false for a nil node", func() {
			Expect(IsNodeReady(nil)).To(BeFalse())
		})

		It("should return true for a ready node with multiple conditions", func() {
			node := &corev1.Node{
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{
						{
							Type:   corev1.NodeMemoryPressure,
							Status: corev1.ConditionFalse,
						},
						{
							Type:   corev1.NodeReady,
							Status: corev1.ConditionTrue,
						},
						{
							Type:   corev1.NodeDiskPressure,
							Status: corev1.ConditionFalse,
						},
					},
				},
			}

			Expect(IsNodeReady(node)).To(BeTrue())
		})
	})

	Describe("FilterReadyNodes", func() {
		var readyNode, notReadyNode, unknownNode corev1.Node

		BeforeEach(func() {
			readyNode = corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "ready-node"},
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{
						{
							Type:   corev1.NodeReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
			}

			notReadyNode = corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "not-ready-node"},
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{
						{
							Type:   corev1.NodeReady,
							Status: corev1.ConditionFalse,
						},
					},
				},
			}

			unknownNode = corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "unknown-node"},
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{
						{
							Type:   corev1.NodeReady,
							Status: corev1.ConditionUnknown,
						},
					},
				},
			}
		})

		It("should filter only ready nodes from mixed list", func() {
			nodes := []corev1.Node{readyNode, notReadyNode, unknownNode}

			filtered := FilterReadyNodes(nodes)

			Expect(filtered).To(HaveLen(1))
			Expect(filtered[0].Name).To(Equal("ready-node"))
		})

		It("should return empty slice when no nodes are ready", func() {
			nodes := []corev1.Node{notReadyNode, unknownNode}

			filtered := FilterReadyNodes(nodes)

			Expect(filtered).To(BeEmpty())
		})

		It("should return all nodes when all are ready", func() {
			anotherReadyNode := readyNode
			anotherReadyNode.Name = "another-ready-node"
			nodes := []corev1.Node{readyNode, anotherReadyNode}

			filtered := FilterReadyNodes(nodes)

			Expect(filtered).To(HaveLen(2))
		})

		It("should handle empty node list", func() {
			nodes := []corev1.Node{}

			filtered := FilterReadyNodes(nodes)

			Expect(filtered).To(BeEmpty())
		})
	})
})
