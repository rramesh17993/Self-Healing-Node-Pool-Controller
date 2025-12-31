package remediation

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Executor handles node remediation actions.
type Executor struct {
	Client     client.Client
	KubeClient kubernetes.Interface
}

// CordonNode marks the node as unschedulable.
func (e *Executor) CordonNode(ctx context.Context, nodeName string) error {
	patch := []byte(`{"spec":{"unschedulable":true}}`)
	// We use the KubeClient here for direct patch, or client.Client if preferred.
	// Using client.Client for consistency with controller-runtime.
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeName,
		},
	}
	if err := e.Client.Patch(ctx, node, client.RawPatch(types.MergePatchType, patch)); err != nil {
		return fmt.Errorf("failed to cordon node %s: %w", nodeName, err)
	}
	return nil
}

// DrainNode safely evicts all pods from a node.
// This is a simplified implementation. Production grade would handle PDBs and timeouts more robustly.
func (e *Executor) DrainNode(ctx context.Context, nodeName string) error {
	// Safety: Ensure we don't block indefinitely.
	// In production, this timeout should be configurable via the policy.
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	// 1. List pods
	pods := &corev1.PodList{}
	if err := e.Client.List(ctx, pods, client.MatchingFields{"spec.nodeName": nodeName}); err != nil {
		return fmt.Errorf("failed to list pods on node %s: %w", nodeName, err)
	}

	for _, pod := range pods.Items {
		// Skip DaemonSets and Static Pods
		if isDaemonSet(&pod) || isStaticPod(&pod) {
			continue
		}

		// Evict
		eviction := &policyv1.Eviction{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pod.Name,
				Namespace: pod.Namespace,
			},
		}
		if err := e.KubeClient.PolicyV1().Evictions(eviction.Namespace).Evict(ctx, eviction); err != nil {
			return fmt.Errorf("failed to evict pod %s/%s: %w", pod.Namespace, pod.Name, err)
		}
	}

	return nil
}

func isDaemonSet(pod *corev1.Pod) bool {
	for _, owner := range pod.OwnerReferences {
		if owner.Kind == "DaemonSet" {
			return true
		}
	}
	return false
}

func isStaticPod(pod *corev1.Pod) bool {
	// Static pods usually don't have owner refs or are mirrored.
	// Check for source annotation usually present on static pods.
	if _, ok := pod.Annotations["kubernetes.io/config.mirror"]; ok {
		return true
	}
	return false
}
