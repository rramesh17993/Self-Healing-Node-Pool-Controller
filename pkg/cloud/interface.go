package cloud

import "context"

// Provider defines the interface for cloud-specific node operations.
// Implementations of this interface (e.g., for AWS, GCP, Azure) are responsible for
// interacting with the respective cloud APIs to perform actions like node replacement
// and pool size querying.
type Provider interface {
	// ReplaceNode triggers the cloud provider to replace the instance.
	ReplaceNode(ctx context.Context, nodeID string) error

	// GetNodePoolSize returns the current size and health of the node pool.
	GetNodePoolSize(ctx context.Context, poolID string) (int, error)
}
